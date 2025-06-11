import logging
from pathlib import Path
from unittest.mock import patch

import pandas as pd
import pytest

# Import functions and constants to be tested
from python_api.src.data_loader import (EXPECTED_EVENT_COLS,
                                        EXPECTED_TRACKING_COLS,
                                        load_event_data, load_tracking_data)


# Fixture to capture log output
@pytest.fixture
def caplog_fixture(caplog):
    caplog.set_level(logging.INFO)
    return caplog


# --- Tests for load_tracking_data ---


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_tracking_data_success(mock_read_parquet, caplog_fixture):
    dummy_df_content = {
        "timestamp_ms": [0, 100],
        "x": [1, 2],
        "y": [3, 4],
        "player_id": ["p1", "p1"],
        "team_id": ["tA", "tA"],
        "smooth_x_speed": [0.1, 0.2],
        "smooth_y_speed": [0.1, 0.1],
    }
    # Ensure all expected columns are present in the mock_df for this success case
    for col in EXPECTED_TRACKING_COLS:
        if col not in dummy_df_content:
            dummy_df_content[col] = [None] * len(
                dummy_df_content["timestamp_ms"]
            )  # Add dummy data or appropriate type

    mock_df = pd.DataFrame(dummy_df_content)
    mock_read_parquet.return_value = mock_df

    file_path = Path("dummy_tracking.gzip")  # Updated extension
    result_df = load_tracking_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    pd.testing.assert_frame_equal(result_df, mock_df)
    assert f"Successfully loaded tracking data from: {file_path}" in caplog_fixture.text


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_tracking_data_missing_essential_cols(mock_read_parquet, caplog_fixture):
    # Missing 'x' column, which is one of the essential_cols in the function
    dummy_df_content = {"timestamp_ms": [0, 100], "y": [3, 4]}
    mock_df = pd.DataFrame(dummy_df_content)
    mock_read_parquet.return_value = mock_df

    file_path = Path("dummy_tracking_missing_cols.gzip")  # Updated extension
    result_df = load_tracking_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_TRACKING_COLS)
    assert (
        f"Essential columns missing in tracking data: {file_path}"
        in caplog_fixture.text
    )


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_tracking_data_file_not_found(mock_read_parquet, caplog_fixture):
    mock_read_parquet.side_effect = FileNotFoundError("File not found")

    file_path = Path("non_existent_tracking.gzip")  # Updated extension
    result_df = load_tracking_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_TRACKING_COLS)
    assert f"Tracking data file not found: {file_path}" in caplog_fixture.text


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_tracking_data_generic_exception(mock_read_parquet, caplog_fixture):
    mock_read_parquet.side_effect = Exception("Some generic Parquet error")

    file_path = Path("error_tracking.gzip")  # Updated extension
    result_df = load_tracking_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_TRACKING_COLS)
    assert (
        f"Error loading tracking data from {file_path}: Some generic Parquet error"
        in caplog_fixture.text
    )


def test_load_tracking_data_path_conversion():
    # Test that string path is converted to Path object
    # Create a mock DataFrame that includes all expected columns
    mock_df_with_all_cols = pd.DataFrame(columns=EXPECTED_TRACKING_COLS)
    for (
        col
    ) in EXPECTED_TRACKING_COLS:  # Ensure some dummy data for essential check if any
        if col in [
            "timestamp_ms",
            "x",
            "y",
            "player_id",
            "team_id",
            "smooth_x_speed",
            "smooth_y_speed",
        ]:
            mock_df_with_all_cols[col] = [0]  # Minimal data to pass checks
        else:
            mock_df_with_all_cols[col] = [None]

    with patch(
        "python_api.src.data_loader.pd.read_parquet", return_value=mock_df_with_all_cols
    ) as mock_read_parquet_conv:
        load_tracking_data("dummy_tracking_str_path.gzip")  # Updated extension
        mock_read_parquet_conv.assert_called_once_with(
            Path("dummy_tracking_str_path.gzip")
        )  # Updated extension


# --- Tests for load_event_data ---


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_event_data_success(mock_read_parquet, caplog_fixture):
    dummy_df_content = {
        "event_id": [1, 2],
        "event_type": ["PASS", "SHOT"],
        "timestamp_ms": [50, 90],
        "player_id": ["p1", "p1"],
        "team_id": ["tA", "tA"],
        "start_x": [1, 2],
        "start_y": [3, 4],
        "end_x": [5, 6],
        "end_y": [7, 8],
    }
    # Ensure all expected columns are present
    for col in EXPECTED_EVENT_COLS:
        if col not in dummy_df_content:
            dummy_df_content[col] = [None] * len(dummy_df_content["event_id"])

    mock_df = pd.DataFrame(dummy_df_content)
    mock_read_parquet.return_value = mock_df

    file_path = Path("dummy_event.gzip")  # Updated extension
    result_df = load_event_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    pd.testing.assert_frame_equal(result_df, mock_df)
    assert f"Successfully loaded event data from: {file_path}" in caplog_fixture.text


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_event_data_missing_essential_cols(mock_read_parquet, caplog_fixture):
    # Missing 'event_type' column (essential)
    dummy_df_content = {"event_id": [1, 2], "timestamp_ms": [50, 90]}
    mock_df = pd.DataFrame(dummy_df_content)
    mock_read_parquet.return_value = mock_df

    file_path = Path("dummy_event_missing_cols.gzip")  # Updated extension
    result_df = load_event_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_EVENT_COLS)
    assert (
        f"Essential columns missing in event data: {file_path}" in caplog_fixture.text
    )


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_event_data_file_not_found(mock_read_parquet, caplog_fixture):
    mock_read_parquet.side_effect = FileNotFoundError("File not found")

    file_path = Path("non_existent_event.gzip")  # Updated extension
    result_df = load_event_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_EVENT_COLS)
    assert f"Event data file not found: {file_path}" in caplog_fixture.text


@patch("python_api.src.data_loader.pd.read_parquet")
def test_load_event_data_generic_exception(mock_read_parquet, caplog_fixture):
    mock_read_parquet.side_effect = Exception("Some generic Parquet error for event")

    file_path = Path("error_event.gzip")  # Updated extension
    result_df = load_event_data(file_path)

    mock_read_parquet.assert_called_once_with(file_path)
    assert result_df.empty
    assert sorted(list(result_df.columns)) == sorted(EXPECTED_EVENT_COLS)
    assert (
        f"Error loading event data from {file_path}: Some generic Parquet error for event"
        in caplog_fixture.text
    )


def test_load_event_data_path_conversion():
    # Create a mock DataFrame that includes all expected columns
    mock_df_with_all_cols = pd.DataFrame(columns=EXPECTED_EVENT_COLS)
    # Populate essential columns with some data to pass checks
    for col in EXPECTED_EVENT_COLS:
        if col in [
            "event_id",
            "event_type",
            "timestamp_ms",
            "player_id",
            "team_id",
            "start_x",
            "start_y",
            "end_x",
            "end_y",
        ]:
            mock_df_with_all_cols[col] = [0]  # Minimal data to pass checks
        else:
            mock_df_with_all_cols[col] = [None]

    with patch(
        "python_api.src.data_loader.pd.read_parquet", return_value=mock_df_with_all_cols
    ) as mock_read_parquet_conv:
        load_event_data("dummy_event_str_path.gzip")  # Updated extension
        mock_read_parquet_conv.assert_called_once_with(
            Path("dummy_event_str_path.gzip")
        )  # Updated extension
