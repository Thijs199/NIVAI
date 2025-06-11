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
    mock_path_exists.return_value = True
    payload = {
        "tracking_data_path": "/fake/tracking.gzip", # Updated extension
        "event_data_path": "/fake/events.gzip",   # Updated extension
        "match_id": "test_match_01"
    }
    response = client.post("/process-match", json=payload)

    assert response.status_code == 202 # Accepted
    json_response = response.json()
    assert json_response["match_id"] == "test_match_01"
    assert "Match processing started" in json_response["message"]

    mock_bg_task.assert_called_once_with(
        "test_match_01",
        Path("/fake/tracking.gzip"), # Updated extension
        Path("/fake/events.gzip")   # Updated extension
    )
    assert "test_match_01" in processed_match_data_cache
    assert processed_match_data_cache["test_match_01"]["status"] == "pending"

def test_process_match_missing_tracking_file(mock_path_exists):
    def side_effect_func_missing_tracking(path_obj=None):
        if path_obj is None:
            return False
        path_str = str(path_obj)
        if path_str == "/fake/tracking.gzip": # Updated extension
            return False
        elif path_str == "/fake/events.gzip": # Updated extension
            return True
        return False
    mock_path_exists.side_effect = side_effect_func_missing_tracking
    payload = {
        "tracking_data_path": "/fake/tracking.gzip", # Updated extension
        "event_data_path": "/fake/events.gzip",   # Updated extension
    }
    response = client.post("/process-match", json=payload)
    assert response.status_code == 404
    assert "Tracking data file not found" in response.json()["detail"]

def test_process_match_missing_event_file(mock_path_exists):
    mock_path_exists.reset_mock(return_value=True, side_effect=None)

    # Simulate that tracking.gzip exists and events.gzip does not
    mock_path_exists.side_effect = [
        True,  # First call to path.exists() (for tracking.gzip)
        False  # Second call to path.exists() (for events.gzip)
    ]

    payload = {
        "tracking_data_path": "/fake/tracking.gzip", # Ensure .gzip
        "event_data_path": "/fake/events.gzip",   # Ensure .gzip
    }
    response = client.post("/process-match", json=payload)
    assert response.status_code == 404  # Correctly indented
    assert "Event data file not found" in response.json()["detail"] # Correctly indented

def test_process_match_invalid_request_body():
    response = client.post("/process-match", json={"tracking_data_path": "path"}) # Missing event_data_path
    assert response.status_code == 422 # Unprocessable Entity

# --- Tests for /match/{match_id}/status ---
@patch('python_api.src.api.main._process_match_data_background', new_callable=MagicMock) # Keep it from running
def test_get_match_status_pending(mock_bg_task, mock_path_exists):
    mock_path_exists.return_value = True
    match_id = "test_status_pending"
    client.post("/process-match", json={
        "tracking_data_path": "/fake/track.gzip",  # Updated extension
        "event_data_path": "/fake/event.gzip", # Updated extension
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
@patch('python_api.src.api.main.generate_player_time_series') # Patched in main's namespace
def test_get_player_details_success(mock_main_generate_ts): # Renamed mock argument
    match_id = "test_p_details_ok"
    player_id = "p1"

    mock_ts_data = [{"timestamp_ms": 0, "speed_kmh": 5.0}]
    mock_main_generate_ts.return_value = mock_ts_data

    cached_enriched_df = get_dummy_tracking_df().copy()
    if 'speed_kmh' not in cached_enriched_df.columns: # ensure conceptual completeness
        cached_enriched_df['speed_kmh'] = 0.0

    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": cached_enriched_df
    }

    response = client.get(f"/match/{match_id}/player/{player_id}/details")
    assert response.status_code == 200
    assert response.json() == {"match_id": match_id, "player_id": player_id, "time_series": mock_ts_data}

    expected_player_df_slice = cached_enriched_df[cached_enriched_df['player_id'] == player_id]
    pd.testing.assert_frame_equal(mock_main_generate_ts.call_args[0][0], expected_player_df_slice, check_dtype=False)


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
@patch('python_api.src.api.main.generate_team_intervals') # Patched in main's namespace
def test_get_team_summary_over_time_success(mock_main_generate_intervals): # Renamed mock argument
    match_id = "test_t_intervals_ok"
    team_id = "tA"

    cached_enriched_df = get_dummy_tracking_df().copy()
    # Add columns that enrich_tracking_data would add, and generate_team_intervals expects
    if 'time_s' not in cached_enriched_df.columns:
        cached_enriched_df['time_s'] = cached_enriched_df['timestamp_ms'] / 1000.0
    if 'relative_time_s' not in cached_enriched_df.columns: # Added by enrich_tracking_data
        cached_enriched_df['relative_time_s'] = cached_enriched_df.groupby('player_id', group_keys=False)['time_s'].transform(lambda x: x - x.min()) # Apply transformation
    if 'distance_covered_m' not in cached_enriched_df.columns:
        cached_enriched_df['distance_covered_m'] = 0.1
    if 'speed_m_s' not in cached_enriched_df.columns:
        cached_enriched_df['speed_m_s'] = 1.0
    if 'speed_kmh' not in cached_enriched_df.columns:
        cached_enriched_df['speed_kmh'] = cached_enriched_df['speed_m_s'] * 3.6
    if 'is_running' not in cached_enriched_df.columns:
        cached_enriched_df['is_running'] = cached_enriched_df['speed_kmh'] > 5.0 # Example threshold
    if 'is_sprinting' not in cached_enriched_df.columns:
        cached_enriched_df['is_sprinting'] = cached_enriched_df['speed_kmh'] > 7.0 # Example threshold
    if 'is_high_intensity_running' not in cached_enriched_df.columns: # CRITICAL
        cached_enriched_df['is_high_intensity_running'] = cached_enriched_df['speed_kmh'] > 6.0 # Example threshold

    processed_match_data_cache[match_id] = {
        "status": "processed",
        "enriched_tracking_df": cached_enriched_df,
        "player_to_team_map": {'p1': 'tA', 'p2': 'tB'}
    }

    mock_interval_data_list = [{"interval_start_time_s": 0, "total_distance_m": 500}]
    mock_main_generate_intervals.return_value = pd.DataFrame(mock_interval_data_list)

    response = client.get(f"/match/{match_id}/team/{team_id}/summary-over-time")
    assert response.status_code == 200
    assert response.json() == {"match_id": match_id, "team_id": team_id, "intervals": mock_interval_data_list}

    expected_team_df_slice = cached_enriched_df[cached_enriched_df['team_id'] == team_id]
    pd.testing.assert_frame_equal(mock_main_generate_intervals.call_args[0][0], expected_team_df_slice, check_dtype=False)


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
@pytest.mark.asyncio
@patch('python_api.src.api.main.load_tracking_data')      # Patched in the correct module
@patch('python_api.src.api.main.load_event_data')        # Patched in the correct module
@patch('python_api.src.api.main.enrich_tracking_data')   # Patched in the correct module
@patch('python_api.src.api.main.generate_all_player_summaries') # Patched in the correct module
@patch('python_api.src.api.main.generate_team_summaries') # Patched in the correct module
async def test_process_match_data_background_flow(
    mock_main_gen_team_sum, mock_main_gen_player_sum, mock_main_enrich, mock_main_load_event, mock_main_load_tracking
):
    match_id = "bg_test_match"
    tracking_path = Path("/fake/tracking.gzip") # Updated extension
    event_path = Path("/fake/events.gzip")     # Updated extension

    dummy_tracking = get_dummy_tracking_df()
    dummy_events = get_dummy_event_df()
    dummy_enriched = dummy_tracking.copy(); dummy_enriched['speed_kmh'] = 10.0 # enrich_tracking_data adds more
    # For more accurate testing, dummy_enriched should fully mock the output of your actual enrich_tracking_data
    dummy_player_summaries = {"p1": {"total_distance_m": 120}}
    dummy_team_summaries = {"tA": {"total_distance_m": 120}}

    mock_main_load_tracking.return_value = dummy_tracking
    mock_main_load_event.return_value = dummy_events
    mock_main_enrich.return_value = dummy_enriched
    mock_main_gen_player_sum.return_value = dummy_player_summaries
    mock_main_gen_team_sum.return_value = dummy_team_summaries

    await _process_match_data_background(match_id, tracking_path, event_path)

    assert match_id in processed_match_data_cache
    cache_item = processed_match_data_cache[match_id]
    assert cache_item["status"] == "processed"
    pd.testing.assert_frame_equal(cache_item["enriched_tracking_df"], dummy_enriched)
    assert cache_item["player_summaries"] == dummy_player_summaries
    assert cache_item["team_summaries"] == dummy_team_summaries
    assert "player_to_team_map" in cache_item

    mock_main_load_tracking.assert_called_once_with(tracking_path)
    mock_main_load_event.assert_called_once_with(event_path)
    mock_main_enrich.assert_called_once()
    # Ensure DataFrame passed to enrich is the one from load_tracking_data
    pd.testing.assert_frame_equal(mock_main_enrich.call_args[0][0], dummy_tracking)

    mock_main_gen_player_sum.assert_called_once()
    # Ensure DataFrame passed to gen_player_sum is the one from enrich
    pd.testing.assert_frame_equal(mock_main_gen_player_sum.call_args[0][0], dummy_enriched)

    mock_main_gen_team_sum.assert_called_once()
    assert mock_main_gen_team_sum.call_args[0][0] == dummy_player_summaries
    assert isinstance(mock_main_gen_team_sum.call_args[0][1], dict)


@pytest.mark.asyncio
@patch('python_api.src.api.main.load_tracking_data') # Target where it's used
async def test_process_match_data_background_tracking_load_fails(mock_main_load_tracking):
    match_id = "bg_test_load_fail"
    tracking_path = Path("/fake/tracking_fails.gzip") # Updated extension
    event_path = Path("/fake/events_ok.gzip")     # Updated extension

    mock_main_load_tracking.return_value = pd.DataFrame() # Empty DataFrame indicates load failure

    # We also need to mock load_event_data for this test, otherwise it will try to load real file
    with patch('python_api.src.api.main.load_event_data') as mock_main_load_event_ignored:
        mock_main_load_event_ignored.return_value = get_dummy_event_df() # Return some valid event data
        await _process_match_data_background(match_id, tracking_path, event_path)

    assert match_id in processed_match_data_cache
    cache_item = processed_match_data_cache[match_id]
    assert cache_item["status"] == "error"
    assert "Tracking data loading failed" in cache_item["message"]
