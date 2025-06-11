import logging
import uuid
from pathlib import Path
from typing import Any, Dict
import os
import tempfile
from azure.storage.blob import BlobServiceClient

import pandas as pd
from fastapi import BackgroundTasks, FastAPI, HTTPException

# Data Loading Functions
from ..data_loader import load_event_data, load_tracking_data

# Define environment variable names
STORAGE_TYPE_ENV = "STORAGE_TYPE"
PYTHON_API_DATA_PATH_ENV = "PYTHON_API_DATA_PATH"
AZURE_STORAGE_CONNECTION_STRING_ENV = "AZURE_STORAGE_CONNECTION_STRING"
AZURE_STORAGE_CONTAINER_NAME_ENV = "AZURE_STORAGE_CONTAINER_NAME"
# Stats Calculation Functions
from ..stats_calculator import (  # Constants for thresholds can be imported if needed by API logic directly; For now, they are used within stats_calculator with defaults
    enrich_tracking_data, generate_all_player_summaries,
    generate_player_time_series, generate_team_intervals,
    generate_team_summaries)
# Pydantic Models
from .models import BasicResponse, ProcessMatchRequest, StatusResponse

# Configure basic logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


def _download_blob_to_tempfile(blob_name: str, connection_string: str, container_name: str, logger_instance: logging.Logger) -> Path:
    logger_instance.info(f"Attempting to download blob: {blob_name} from container: {container_name}")
    try:
        blob_service_client = BlobServiceClient.from_connection_string(connection_string)
        blob_client = blob_service_client.get_blob_client(container=container_name, blob=blob_name)

        temp_file = tempfile.NamedTemporaryFile(delete=False, suffix=Path(blob_name).suffix)
        with open(temp_file.name, "wb") as download_file:
            download_stream = blob_client.download_blob()
            download_file.write(download_stream.readall())

        logger_instance.info(f"Successfully downloaded {blob_name} to {temp_file.name}")
        return Path(temp_file.name)
    except Exception as e:
        logger_instance.exception(f"Failed to download blob {blob_name}: {e}")
        raise


app = FastAPI(title="Football Analysis API")

# In-memory cache for processed data
# For production, consider Redis or another distributed cache.
processed_match_data_cache: Dict[str, Dict[str, Any]] = {}

# --- Background Processing Task ---


def _get_player_to_team_map(df: pd.DataFrame) -> Dict[str, str]:
    """Helper to extract player_id to team_id mapping from a DataFrame."""
    if df.empty or "player_id" not in df.columns or "team_id" not in df.columns:
        return {}
    return (
        df[["player_id", "team_id"]]
        .drop_duplicates()
        .set_index("player_id")["team_id"]
        .to_dict()
    )


async def _process_match_data_background(
    match_id: str, tracking_path: Path, event_path: Path
):
    """
    Background task to load, process, and cache match data.
    """
    temp_files_to_clean: list[Path] = []
    logger.info(
        f"[{match_id}] Starting background processing for tracking: {tracking_path}, event: {event_path}"
    )
    try:
        storage_type = os.getenv(STORAGE_TYPE_ENV, "local").lower()
        logger.info(f"[{match_id}] Storage type configured: {storage_type}")

        # Convert input Path objects to string representations for blob names or relative paths
        # These original string paths are what Go backend provides.
        input_tracking_path_str = str(tracking_path)
        input_event_path_str = str(event_path)

        final_tracking_path: Path
        final_event_path: Path

        if storage_type == "azure":
            connection_string = os.getenv(AZURE_STORAGE_CONNECTION_STRING_ENV)
            container_name = os.getenv(AZURE_STORAGE_CONTAINER_NAME_ENV)
            if not connection_string or not container_name:
                logger.error(
                    f"[{match_id}] Azure storage type configured, but connection string or container name is missing."
                )
                processed_match_data_cache[match_id] = {
                    "status": "error",
                    "message": "Azure configuration incomplete.",
                }
                return # Exit if config is bad

            # This download block itself needs error handling
            try:
                logger.info(f"[{match_id}] Downloading tracking data from Azure: {input_tracking_path_str}")
                # Pass the module/instance logger to the helper
                temp_tracking_file = _download_blob_to_tempfile(input_tracking_path_str, connection_string, container_name, logger)
                final_tracking_path = temp_tracking_file
                temp_files_to_clean.append(temp_tracking_file)

                logger.info(f"[{match_id}] Downloading event data from Azure: {input_event_path_str}")
                temp_event_file = _download_blob_to_tempfile(input_event_path_str, connection_string, container_name, logger)
                final_event_path = temp_event_file
                temp_files_to_clean.append(temp_event_file)
            except Exception as e: # Catch exceptions from _download_blob_to_tempfile
                logger.error(f"[{match_id}] Failed to download one or more files from Azure: {e}")
                processed_match_data_cache[match_id] = {
                    "status": "error",
                    "message": f"Azure file download failed: {e}",
                }
                # No return here, finally block will clean up any partially downloaded files.
                raise # Re-raise to be caught by the outer try/except that sets main status

        elif storage_type == "local":
            python_api_data_path_str = os.getenv(PYTHON_API_DATA_PATH_ENV)
            if not python_api_data_path_str:
                logger.error(
                    f"[{match_id}] Local storage type configured, but PYTHON_API_DATA_PATH is not set."
                )
                processed_match_data_cache[match_id] = {
                    "status": "error",
                    "message": "Local storage path configuration missing.",
                }
                return # Exit if config is bad

            base_data_path = Path(python_api_data_path_str)
            # tracking_path and event_path are Path objects representing relative paths.
            final_tracking_path = base_data_path / tracking_path
            final_event_path = base_data_path / event_path

            logger.info(f"[{match_id}] Using local tracking data path: {final_tracking_path}")
            logger.info(f"[{match_id}] Using local event data path: {final_event_path}")

        else: # Invalid storage type
            logger.error(f"[{match_id}] Invalid STORAGE_TYPE: {storage_type}")
            processed_match_data_cache[match_id] = {
                "status": "error",
                "message": f"Invalid storage type: {storage_type}",
            }
            return # Exit if config is bad

        # Load data
        # Note: load_tracking_data/load_event_data are synchronous.
        # For very large files or remote storage, consider running them in a thread pool.
        tracking_df = load_tracking_data(final_tracking_path)
        event_df = load_event_data(
            final_event_path
        )  # Currently not used extensively by stats_calculator

        if tracking_df.empty:
            logger.error(
                f"[{match_id}] Failed to load tracking data or data is empty. Aborting processing."
            )
            processed_match_data_cache[match_id] = {
                "status": "error",
                "message": "Tracking data loading failed.",
            }
            return

        # Enrich tracking data
        logger.info(f"[{match_id}] Enriching tracking data...")
        enriched_df = enrich_tracking_data(tracking_df)  # This can be CPU intensive
        if enriched_df.empty:
            logger.error(
                f"[{match_id}] Enriched tracking data is empty. Aborting processing."
            )
            processed_match_data_cache[match_id] = {
                "status": "error",
                "message": "Data enrichment resulted in empty dataset.",
            }
            return

        logger.info(f"[{match_id}] Generating player summaries...")
        player_summaries = generate_all_player_summaries(enriched_df)

        logger.info(f"[{match_id}] Generating player to team map...")
        # Assuming team_id is present in enriched_df (comes from tracking_df)
        # If not, event_df might be a source for this map.
        player_to_team_map = _get_player_to_team_map(enriched_df)
        if not player_to_team_map:
            logger.warning(
                f"[{match_id}] Could not generate player_to_team_map from tracking data."
            )
            # Potentially load from event_df or use a default if critical

        logger.info(f"[{match_id}] Generating team summaries...")
        team_summaries = generate_team_summaries(player_summaries, player_to_team_map)

        # Store results in cache
        processed_match_data_cache[match_id] = {
            "status": "processed",
            "enriched_tracking_df": enriched_df,
            "player_summaries": player_summaries,
            "team_summaries": team_summaries,
            "player_to_team_map": player_to_team_map,
            "event_df": event_df,  # Store event_df too, might be useful later
        }
        logger.info(f"[{match_id}] Successfully processed and cached data.")

    except Exception as e:
        logger.exception(f"[{match_id}] Error during background processing: {e}")
        processed_match_data_cache[match_id] = {"status": "error", "message": str(e)}
    finally:
        logger.info(f"[{match_id}] Cleaning up temporary files: {temp_files_to_clean}")
        for temp_file_path in temp_files_to_clean:
            if temp_file_path.exists():
                try:
                    os.remove(temp_file_path)
                    logger.info(f"[{match_id}] Removed temporary file: {temp_file_path}")
                except OSError as ose: # More specific exception for os.remove
                    logger.error(f"[{match_id}] Error removing temporary file {temp_file_path}: {ose}")


# --- API Endpoints ---


@app.post("/process-match", response_model=BasicResponse, status_code=202)
async def process_match(
    request: ProcessMatchRequest, background_tasks: BackgroundTasks
):
    """
    Starts background processing for a match given tracking and event data paths.
    """
    match_id = request.match_id or str(uuid.uuid4())

    tracking_file = Path(request.tracking_data_path)
    event_file = Path(request.event_data_path)

    if not tracking_file.exists():
        raise HTTPException(
            status_code=404, detail=f"Tracking data file not found: {tracking_file}"
        )
    if not event_file.exists():
        raise HTTPException(
            status_code=404, detail=f"Event data file not found: {event_file}"
        )

    # Mark as pending before starting task
    processed_match_data_cache[match_id] = {"status": "pending"}

    background_tasks.add_task(
        _process_match_data_background, match_id, tracking_file, event_file
    )

    return BasicResponse(
        message="Match processing started in background.", match_id=match_id
    )


@app.get("/match/{match_id}/status", response_model=StatusResponse)
async def get_match_status(match_id: str):
    """
    Checks the processing status of a match.
    """
    cache_entry = processed_match_data_cache.get(match_id)
    if not cache_entry:
        raise HTTPException(status_code=404, detail="Match ID not found.")

    status = cache_entry.get("status", "unknown")
    message = cache_entry.get(
        "message"
    )  # Return message regardless of status if present

    return StatusResponse(status=status, match_id=match_id, message=message)


@app.get(
    "/match/{match_id}/stats/summary"
)  # Consider a more specific response model later
async def get_match_summary(match_id: str):
    """
    Retrieves overall player and team summary statistics for a processed match.
    """
    cache_entry = processed_match_data_cache.get(match_id)
    if not cache_entry or cache_entry.get("status") != "processed":
        raise HTTPException(
            status_code=404, detail="Match data not processed or match ID not found."
        )

    player_summaries = cache_entry.get("player_summaries", {})
    team_summaries = cache_entry.get("team_summaries", {})

    # Pandas Series/DataFrames in summaries might not be directly JSON serializable by default.
    # stats_calculator currently returns dicts/Series. FastAPI handles basic pd.Series to dict.
    return {"match_id": match_id, "players": player_summaries, "teams": team_summaries}


@app.get(
    "/match/{match_id}/player/{player_id}/details"
)  # Consider a more specific response model
async def get_player_details(match_id: str, player_id: str):
    """
    Retrieves detailed time-series data for a specific player in a match.
    """
    cache_entry = processed_match_data_cache.get(match_id)
    if not cache_entry or cache_entry.get("status") != "processed":
        raise HTTPException(
            status_code=404, detail="Match data not processed or match ID not found."
        )

    enriched_df = cache_entry.get("enriched_tracking_df")
    if enriched_df is None or enriched_df.empty:
        raise HTTPException(
            status_code=404,
            detail="Enriched tracking data not available for this match.",
        )

    player_df = enriched_df[enriched_df["player_id"] == player_id]
    if player_df.empty:
        raise HTTPException(
            status_code=404, detail=f"Player ID {player_id} not found in this match."
        )

    time_series_data = generate_player_time_series(player_df)
    return {
        "match_id": match_id,
        "player_id": player_id,
        "time_series": time_series_data,
    }


@app.get(
    "/match/{match_id}/team/{team_id}/summary-over-time"
)  # Consider a more specific response model
async def get_team_summary_over_time(match_id: str, team_id: str):
    """
    Retrieves time-interval based summary statistics for a specific team in a match.
    """
    cache_entry = processed_match_data_cache.get(match_id)
    if not cache_entry or cache_entry.get("status") != "processed":
        raise HTTPException(
            status_code=404, detail="Match data not processed or match ID not found."
        )

    enriched_df = cache_entry.get("enriched_tracking_df")
    player_to_team_map = cache_entry.get(
        "player_to_team_map", {}
    )  # Should be generated

    if enriched_df is None or enriched_df.empty:
        raise HTTPException(
            status_code=404,
            detail="Enriched tracking data not available for this match.",
        )

    # Filter enriched_df for players belonging to the specified team_id
    # This relies on 'team_id' column being present in enriched_df.
    if "team_id" not in enriched_df.columns:
        # Fallback: Try to map player_ids to team_ids if team_id column is missing from enriched_df
        # This is less ideal; team_id should ideally be part of enriched_df.
        players_in_team = [
            pid for pid, tid in player_to_team_map.items() if tid == team_id
        ]
        if not players_in_team:
            raise HTTPException(
                status_code=404,
                detail=f"Team ID {team_id} not found or no players mapped to it.",
            )
        team_df = enriched_df[enriched_df["player_id"].isin(players_in_team)]
    else:
        team_df = enriched_df[enriched_df["team_id"] == team_id]

    if team_df.empty:
        raise HTTPException(
            status_code=404,
            detail=f"Team ID {team_id} not found or no data for this team.",
        )

    # Default interval of 5 minutes, can be parameterized if needed
    team_interval_data = generate_team_intervals(team_df, time_interval_minutes=5)

    # generate_team_intervals returns a DataFrame. FastAPI will convert to list of dicts.
    return {
        "match_id": match_id,
        "team_id": team_id,
        "intervals": team_interval_data.to_dict(orient="records"),
    }


# Root endpoint
@app.get("/")
async def read_root():
    return {"message": "Welcome to the Football Analysis API!"}


# Placeholder for future endpoints (already existed from previous step)
@app.get("/api/v1/matches")
async def list_matches():
    return {"message": "Endpoint to list matches - TBD"}


if __name__ == "__main__":

    # This is for local development testing.
    # You would typically run this with: uvicorn python_api.src.api.main:app --reload --port 8081
    # Ensure dummy files (e.g., dummy_tracking.parquet, dummy_event.parquet) exist in the root
    # of the `python_api` directory if you want to test the /process-match endpoint locally.
    # Create dummy parquet files for testing if they don't exist (requires pyarrow)
    try:
        import pandas as pd

        # Check if pyarrow is installed, otherwise this block will fail
        pd.DataFrame().to_parquet("dummy.parquet")  # Test if pyarrow is usable

        dummy_tracking_content = {
            "player_id": ["p1", "p1", "p2", "p2"] * 50,
            "team_id": ["tA", "tA", "tB", "tB"] * 50,
            "timestamp_ms": list(range(0, 20000, 100)),  # 200 records, 10Hz
            "x": [i * 0.5 for i in range(200)],
            "y": [i * 0.25 for i in range(200)],
            "smooth_x_speed": [0.1, 0.2, -0.1, 0.3] * 50,
            "smooth_y_speed": [0.05, -0.1, 0.15, 0.05] * 50,
        }
        dummy_tracking_df = pd.DataFrame(dummy_tracking_content)
        # Place it where the API might expect, e.g., inside python_api or a designated data folder
        # For this example, let's assume it's in the root of the project /app/
        # The API paths are relative to where the server is run or absolute.
        # For robustness, paths in ProcessMatchRequest should ideally be absolute or relative to a known data root.
        Path("dummy_tracking.parquet").write_bytes(dummy_tracking_df.to_parquet())

        dummy_event_content = {
            "event_id": [1, 2, 3],
            "event_type": ["PASS", "SHOT", "DRIBBLE"],
            "player_id": ["p1", "p2", "p1"],
            "team_id": ["tA", "tB", "tA"],
            "timestamp_ms": [1000, 5000, 10000],
            "start_x": [10, 50, 30],
            "start_y": [20, 30, 40],
            "end_x": [15, 55, 35],
            "end_y": [22, 32, 42],
        }
        dummy_event_df = pd.DataFrame(dummy_event_content)
        Path("dummy_event.parquet").write_bytes(dummy_event_df.to_parquet())

        logger.info(
            "Dummy parquet files created for local testing (dummy_tracking.parquet, dummy_event.parquet)."
        )
        logger.info(
            "Run: uvicorn python_api.src.api.main:app --reload --port 8081 --app-dir /app"
        )
        logger.info(
            "Then try POSTing to /process-match with paths like '/app/dummy_tracking.parquet'"
        )

    except ImportError:
        logger.warning(
            "pandas or pyarrow not installed. Cannot create dummy files for local testing."
        )
    except Exception as e:
        logger.warning(f"Could not create dummy files: {e}")

    # uvicorn.run("main:app", host="0.0.0.0", port=8081, reload=True, app_dir=str(Path(__file__).parent))
    # The above app_dir might not be correct when running inside the agent's environment.
    # It's better to run uvicorn from the command line specifying the app directory.
    # Example: poetry run uvicorn python_api.src.api.main:app --reload --port 8081 --app-dir /app
    # (if /app is the root of the project where python_api folder is)
    # For the agent, this `if __name__ == "__main__":` block is mostly for code structure reference.
    # The agent will not directly execute this `uvicorn.run` call.
    pass  # End of main block, uvicorn is run externally.
