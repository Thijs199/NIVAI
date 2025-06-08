import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock, ANY
from pathlib import Path
import pandas as pd
import asyncio # For direct async function calls if needed, though TestClient handles app calls

# Import the app instance and cache from your main application file
from python_api.src.api.main import app, processed_match_data_cache, _process_match_data_background
# Import stats_calculator to mock its functions
from python_api.src import stats_calculator
from python_api.src import data_loader


# Initialize the TestClient
client = TestClient(app)

# --- Test Data ---
def get_dummy_tracking_df():
    return pd.DataFrame({
        'player_id': ['p1', 'p1', 'p2', 'p2'],
        'team_id': ['tA', 'tA', 'tB', 'tB'],
        'timestamp_ms': [0, 100, 0, 100],
        'x': [10, 12, 50, 52],
        'y': [20, 22, 60, 62],
        'smooth_x_speed': [1.0, 1.5, -1.0, -1.2],
        'smooth_y_speed': [0.5, 0.7, 0.3, 0.4]
    })

def get_dummy_event_df():
    return pd.DataFrame({
        'event_id': [1, 2], 'event_type': ['PASS', 'SHOT'], 'player_id': ['p1', 'p2'],
        'team_id': ['tA', 'tB'], 'timestamp_ms': [50, 70],
        'start_x': [11, 51], 'start_y': [21, 61]
    })

# --- Fixtures ---
@pytest.fixture(autouse=True)
def clear_cache_and_mocks(monkeypatch):
    """Clears the cache before each test and resets relevant mocks."""
    processed_match_data_cache.clear()
    # Reset any global mocks or state if necessary
    yield # Test runs here
    processed_match_data_cache.clear()

@pytest.fixture
def mock_path_exists(monkeypatch):
    """Fixture to mock Path.exists(). Call with .return_value = True/False in test."""
    mock = MagicMock()
    monkeypatch.setattr(Path, "exists", mock)
    return mock

# --- Tests for /process-match ---
@patch('python_api.src.api.main._process_match_data_background', new_callable=MagicMock)
def test_process_match_success(mock_bg_task, mock_path_exists):
    mock_path_exists.return_value = True # Both tracking and event files exist
    payload = {
        "tracking_data_path": "/fake/tracking.parquet",
        "event_data_path": "/fake/events.parquet",
        "match_id": "test_match_01"
    }
    response = client.post("/process-match", json=payload)

    assert response.status_code == 202 # Accepted
    json_response = response.json()
    assert json_response["match_id"] == "test_match_01"
    assert "Match processing started" in json_response["message"]

    mock_bg_task.assert_called_once_with(
        "test_match_01",
        Path("/fake/tracking.parquet"),
        Path("/fake/events.parquet")
    )
    assert "test_match_01" in processed_match_data_cache
    assert processed_match_data_cache["test_match_01"]["status"] == "pending"

def test_process_match_missing_tracking_file(mock_path_exists):
    mock_path_exists.side_effect = lambda p: str(p) == "/fake/events.parquet" # Only event file exists
    payload = {
        "tracking_data_path": "/fake/tracking.parquet", # This one will "not exist"
        "event_data_path": "/fake/events.parquet",
    }
    response = client.post("/process-match", json=payload)
    assert response.status_code == 404
    assert "Tracking data file not found" in response.json()["detail"]

def test_process_match_missing_event_file(mock_path_exists):
    mock_path_exists.side_effect = lambda p: str(p) == "/fake/tracking.parquet" # Only tracking file exists
    payload = {
        "tracking_data_path": "/fake/tracking.parquet",
        "event_data_path": "/fake/events.parquet", # This one will "not exist"
    }
    response = client.post("/process-match", json=payload)
    assert response.status_code == 404
    assert "Event data file not found" in response.json()["detail"]

def test_process_match_invalid_request_body():
    response = client.post("/process-match", json={"tracking_data_path": "path"}) # Missing event_data_path
    assert response.status_code == 422 # Unprocessable Entity

# --- Tests for /match/{match_id}/status ---
@patch('python_api.src.api.main._process_match_data_background', new_callable=MagicMock) # Keep it from running
def test_get_match_status_pending(mock_bg_task, mock_path_exists):
    mock_path_exists.return_value = True
    match_id = "test_status_pending"
    client.post("/process-match", json={
        "tracking_data_path": "/fake/track.parquet",
        "event_data_path": "/fake/event.parquet",
        "match_id": match_id
    }) # This sets status to "pending"

    response = client.get(f"/match/{match_id}/status")
    assert response.status_code == 200
    assert response.json() == {"status": "pending", "match_id": match_id, "message": None}

def test_get_match_status_processed():
    match_id = "test_status_processed"
    processed_match_data_cache[match_id] = {"status": "processed", "message": "Completed."}
    response = client.get(f"/match/{match_id}/status")
    assert response.status_code == 200
    assert response.json() == {"status": "processed", "match_id": match_id, "message": "Completed."}

def test_get_match_status_error():
    match_id = "test_status_error"
    processed_match_data_cache[match_id] = {"status": "error", "message": "Something went wrong."}
    response = client.get(f"/match/{match_id}/status")
    assert response.status_code == 200
    assert response.json() == {"status": "error", "match_id": match_id, "message": "Something went wrong."}

def test_get_match_status_non_existent():
    response = client.get("/match/non_existent_match/status")
    assert response.status_code == 404

# --- Tests for /match/{match_id}/stats/summary ---
def test_get_match_summary_success():
    match_id = "test_summary_ok"
    dummy_players = {"p1": {"total_distance_m": 100}}
    dummy_teams = {"tA": {"total_distance_m": 100}}
    processed_match_data_cache[match_id] = {
        "status": "processed",
        "player_summaries": dummy_players,
        "team_summaries": dummy_teams
    }
    response = client.get(f"/match/{match_id}/stats/summary")
    assert response.status_code == 200
    assert response.json() == {"match_id": match_id, "players": dummy_players, "teams": dummy_teams}

def test_get_match_summary_not_processed():
    match_id = "test_summary_pending"
    processed_match_data_cache[match_id] = {"status": "pending"}
    response = client.get(f"/match/{match_id}/stats/summary")
    assert response.status_code == 404
    assert "not processed" in response.json()["detail"].lower()

def test_get_match_summary_non_existent():
    response = client.get("/match/summary_non_existent/stats/summary")
    assert response.status_code == 404

# --- Tests for /match/{match_id}/player/{player_id}/details ---
@patch.object(stats_calculator, 'generate_player_time_series')
def test_get_player_details_success(mock_generate_ts):
    match_id = "test_p_details_ok"
    player_id = "p1"
    dummy_enriched_df = get_dummy_tracking_df()[get_dummy_tracking_df()['player_id'] == player_id]
    mock_ts_data = [{"timestamp_ms": 0, "speed_kmh": 5.0}]
    mock_generate_ts.return_value = mock_ts_data

    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": get_dummy_tracking_df() # Full df for general lookup
    }

    response = client.get(f"/match/{match_id}/player/{player_id}/details")
    assert response.status_code == 200
    assert response.json() == {"match_id": match_id, "player_id": player_id, "time_series": mock_ts_data}

    # Check that mock_generate_ts was called with the correct DataFrame slice
    pd.testing.assert_frame_equal(mock_generate_ts.call_args[0][0], dummy_enriched_df, check_dtype=False)


def test_get_player_details_player_not_found():
    match_id = "test_p_not_found"
    player_id = "p_non_existent"
    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": get_dummy_tracking_df()
    }
    response = client.get(f"/match/{match_id}/player/{player_id}/details")
    assert response.status_code == 404
    assert "player id" in response.json()["detail"].lower() and "not found" in response.json()["detail"].lower()

# --- Tests for /match/{match_id}/team/{team_id}/summary-over-time ---
@patch.object(stats_calculator, 'generate_team_intervals')
def test_get_team_summary_over_time_success(mock_generate_intervals):
    match_id = "test_t_intervals_ok"
    team_id = "tA"
    # Simulate the DataFrame that would be filtered for team tA
    team_df_tA = get_dummy_tracking_df()[get_dummy_tracking_df()['team_id'] == team_id]

    mock_interval_data_list = [{"interval_start_time_s": 0, "total_distance_m": 500}]
    # generate_team_intervals in main code returns a DataFrame.
    # The endpoint then calls .to_dict(orient='records').
    mock_generate_intervals.return_value = pd.DataFrame(mock_interval_data_list)

    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": get_dummy_tracking_df(), # Full df
        "player_to_team_map": {'p1': 'tA', 'p2': 'tB'}
    }

    response = client.get(f"/match/{match_id}/team/{team_id}/summary-over-time")
    assert response.status_code == 200
    assert response.json() == {"match_id": match_id, "team_id": team_id, "intervals": mock_interval_data_list}

    # Check that mock_generate_intervals was called with the correct DataFrame slice for team tA
    pd.testing.assert_frame_equal(mock_generate_intervals.call_args[0][0], team_df_tA, check_dtype=False)


def test_get_team_summary_over_time_team_not_found():
    match_id = "test_t_not_found"
    team_id = "t_non_existent"
    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": get_dummy_tracking_df(),
        "player_to_team_map": {'p1': 'tA', 'p2': 'tB'}
    }
    response = client.get(f"/match/{match_id}/team/{team_id}/summary-over-time")
    assert response.status_code == 404
    assert "team id" in response.json()["detail"].lower() and "not found" in response.json()["detail"].lower()


# --- Integration-style test for _process_match_data_background ---
# This uses mocks for data_loader and stats_calculator functions to test the flow.
@pytest.mark.asyncio # Requires pytest-asyncio if not using another async test runner like anyio from FastAPI
@patch.object(data_loader, 'load_tracking_data')
@patch.object(data_loader, 'load_event_data')
@patch.object(stats_calculator, 'enrich_tracking_data')
@patch.object(stats_calculator, 'generate_all_player_summaries')
@patch.object(stats_calculator, 'generate_team_summaries')
async def test_process_match_data_background_flow(
    mock_gen_team_sum, mock_gen_player_sum, mock_enrich, mock_load_event, mock_load_tracking
):
    match_id = "bg_test_match"
    tracking_path = Path("/fake/tracking.parquet") # Path objects
    event_path = Path("/fake/events.parquet")     # Path objects

    dummy_tracking = get_dummy_tracking_df()
    dummy_events = get_dummy_event_df()
    dummy_enriched = dummy_tracking.copy(); dummy_enriched['speed_kmh'] = 10.0
    dummy_player_summaries = {"p1": {"total_distance_m": 120}}
    dummy_team_summaries = {"tA": {"total_distance_m": 120}}

    mock_load_tracking.return_value = dummy_tracking
    mock_load_event.return_value = dummy_events
    mock_enrich.return_value = dummy_enriched
    mock_gen_player_sum.return_value = dummy_player_summaries
    mock_gen_team_sum.return_value = dummy_team_summaries

    await _process_match_data_background(match_id, tracking_path, event_path)

    assert match_id in processed_match_data_cache
    cache_item = processed_match_data_cache[match_id]
    assert cache_item["status"] == "processed"
    pd.testing.assert_frame_equal(cache_item["enriched_tracking_df"], dummy_enriched)
    assert cache_item["player_summaries"] == dummy_player_summaries
    assert cache_item["team_summaries"] == dummy_team_summaries
    assert "player_to_team_map" in cache_item # Check existence, content depends on dummy_enriched

    mock_load_tracking.assert_called_once_with(tracking_path)
    mock_load_event.assert_called_once_with(event_path)
    # For DataFrames, assert_called_once_with checks object identity, not value.
    # To check value: pd.testing.assert_frame_equal(mock_enrich.call_args[0][0], dummy_tracking)
    mock_enrich.assert_called_once()
    pd.testing.assert_frame_equal(mock_enrich.call_args[0][0], dummy_tracking)

    mock_gen_player_sum.assert_called_once()
    pd.testing.assert_frame_equal(mock_gen_player_sum.call_args[0][0], dummy_enriched)

    mock_gen_team_sum.assert_called_once()
    # mock_gen_team_sum is called with (player_summaries, player_to_team_map)
    assert mock_gen_team_sum.call_args[0][0] == dummy_player_summaries
    # player_to_team_map is generated internally, check it's a dict
    assert isinstance(mock_gen_team_sum.call_args[0][1], dict)


@pytest.mark.asyncio
@patch.object(data_loader, 'load_tracking_data')
async def test_process_match_data_background_tracking_load_fails(mock_load_tracking):
    match_id = "bg_test_load_fail"
    tracking_path = Path("/fake/tracking_fails.parquet")
    event_path = Path("/fake/events_ok.parquet")

    mock_load_tracking.return_value = pd.DataFrame() # Empty DataFrame indicates load failure

    await _process_match_data_background(match_id, tracking_path, event_path)

    assert match_id in processed_match_data_cache
    cache_item = processed_match_data_cache[match_id]
    assert cache_item["status"] == "error"
    assert "Tracking data loading failed" in cache_item["message"]
