import pandas as pd
import numpy as np

# Speed and Intensity Thresholds
DEFAULT_HIGH_SPEED_THRESHOLD_KMH = 19.8  # km/h for running
DEFAULT_SPRINT_SPEED_THRESHOLD_KMH = 25.2 # km/h for sprinting
ACCELERATION_THRESHOLD_MS2 = 0.5 # m/s^2
DECELERATION_THRESHOLD_MS2 = -0.5 # m/s^2 (negative for deceleration)
TIME_WINDOW_FOR_ACCEL_SEC = 0.5 # seconds, window to check for speed change (Note: This constant is defined but not explicitly used in the functions described below, may be for a different type of acc/dec calculation)

# Unit conversions
KMH_TO_MS = 1000 / 3600
MS_TO_KMH = 3600 / 1000

# Expected columns (for reference, not strict enforcement in this file)
# player_id, team_id, timestamp_ms, x, y, smooth_x_speed, smooth_y_speed

# --- Helper Functions ---

def calculate_speed_kmh(tracking_df: pd.DataFrame) -> pd.DataFrame:
    """
    Calculates the magnitude of the speed vector from smooth_x_speed and smooth_y_speed,
    converts it to km/h, and adds it as a new column 'speed_kmh'.
    It also calculates 'speed_ms'.
    """
    if 'smooth_x_speed' not in tracking_df.columns or 'smooth_y_speed' not in tracking_df.columns:
        raise ValueError("DataFrame must contain 'smooth_x_speed' and 'smooth_y_speed' columns.")

    tracking_df['speed_ms'] = np.sqrt(tracking_df['smooth_x_speed']**2 + tracking_df['smooth_y_speed']**2)
    tracking_df['speed_kmh'] = tracking_df['speed_ms'] * MS_TO_KMH
    return tracking_df

def calculate_distance_covered(tracking_df: pd.DataFrame) -> pd.DataFrame:
    """
    Calculates the distance covered between consecutive timestamps for each player.
    Assumes tracking_df is sorted by player_id and then by timestamp_ms.
    Adds 'distance_covered_m' column.
    """
    if not {'x', 'y', 'player_id', 'timestamp_ms'}.issubset(tracking_df.columns):
        raise ValueError("DataFrame must contain 'x', 'y', 'player_id', and 'timestamp_ms' columns.")

    # Ensure data is sorted for correct diff operation per player
    tracking_df = tracking_df.sort_values(by=['player_id', 'timestamp_ms'])

    # Calculate distance: sqrt((x2-x1)^2 + (y2-y1)^2)
    # .diff() calculates difference with the PREVIOUS row.
    # Within a group (player_id), the first row will have NaNs for diffs.
    tracking_df['delta_x'] = tracking_df.groupby('player_id')['x'].diff()
    tracking_df['delta_y'] = tracking_df.groupby('player_id')['y'].diff()

    tracking_df['distance_covered_m'] = np.sqrt(tracking_df['delta_x']**2 + tracking_df['delta_y']**2)
    # Fill NaN for the first record of each player with 0 distance
    tracking_df['distance_covered_m'] = tracking_df['distance_covered_m'].fillna(0)

    tracking_df = tracking_df.drop(columns=['delta_x', 'delta_y'])
    return tracking_df

def calculate_acceleration(tracking_df: pd.DataFrame) -> pd.DataFrame:
    """
    Calculates acceleration based on changes in speed_ms and timestamp_ms.
    Assumes tracking_df is sorted by player_id and then by timestamp_ms.
    Adds 'acceleration_ms2' column.
    """
    if 'speed_ms' not in tracking_df.columns or 'timestamp_ms' not in tracking_df.columns:
        raise ValueError("DataFrame must contain 'speed_ms' and 'timestamp_ms' columns. Run calculate_speed_kmh first.")
    if 'player_id' not in tracking_df.columns:
        raise ValueError("DataFrame must contain 'player_id' for correct grouping.")

    tracking_df = tracking_df.sort_values(by=['player_id', 'timestamp_ms'])

    tracking_df['time_s'] = tracking_df['timestamp_ms'] / 1000

    tracking_df['delta_speed_ms'] = tracking_df.groupby('player_id')['speed_ms'].diff()
    tracking_df['delta_time_s'] = tracking_df.groupby('player_id')['time_s'].diff()

    # Acceleration = delta_speed / delta_time
    # Handle division by zero if delta_time_s is 0 (consecutive timestamps are identical)
    tracking_df['acceleration_ms2'] = tracking_df['delta_speed_ms'].divide(tracking_df['delta_time_s']).fillna(0)
    # Replace inf with 0 if delta_time_s was 0 but delta_speed_ms was not.
    tracking_df.replace([np.inf, -np.inf], 0, inplace=True)

    tracking_df = tracking_df.drop(columns=['delta_speed_ms', 'delta_time_s']) # keep 'time_s' for now
    return tracking_df

# --- Enrichment Function ---

def enrich_tracking_data(tracking_df: pd.DataFrame,
                         high_speed_threshold_kmh: float = DEFAULT_HIGH_SPEED_THRESHOLD_KMH,
                         sprint_speed_threshold_kmh: float = DEFAULT_SPRINT_SPEED_THRESHOLD_KMH) -> pd.DataFrame:
    """
    Enriches the tracking DataFrame with calculated metrics like speed, distance, acceleration,
    and flags for high-intensity running and sprinting.

    Args:
        tracking_df (pd.DataFrame): Raw tracking data. Must include 'player_id',
                                    'timestamp_ms', 'x', 'y', 'smooth_x_speed', 'smooth_y_speed'.
        high_speed_threshold_kmh (float): Threshold for high-intensity running.
        sprint_speed_threshold_kmh (float): Threshold for sprinting.

    Returns:
        pd.DataFrame: Enriched tracking data.
    """
    if tracking_df.empty:
        # Return an empty DataFrame with expected enriched columns if input is empty
        return pd.DataFrame(columns=[
            'player_id', 'team_id', 'timestamp_ms', 'x', 'y',
            'smooth_x_speed', 'smooth_y_speed', 'speed_ms', 'speed_kmh',
            'distance_covered_m', 'time_s', 'acceleration_ms2',
            'is_sprinting', 'is_high_intensity_running'
        ])

    # Calculate speed in m/s and km/h
    tracking_df = calculate_speed_kmh(tracking_df.copy()) # Use .copy() to avoid SettingWithCopyWarning

    # Calculate distance covered between points
    tracking_df = calculate_distance_covered(tracking_df)

    # Calculate acceleration
    tracking_df = calculate_acceleration(tracking_df)

    # Add boolean flags
    tracking_df['is_sprinting'] = tracking_df['speed_kmh'] > sprint_speed_threshold_kmh
    tracking_df['is_high_intensity_running'] = (tracking_df['speed_kmh'] > high_speed_threshold_kmh) & \
                                               (tracking_df['speed_kmh'] <= sprint_speed_threshold_kmh)
    return tracking_df

# --- High-Intensity Running and Sprinting Stats ---

def calculate_high_intensity_running_stats(player_tracking_data: pd.DataFrame,
                                           high_speed_threshold_kmh: float = DEFAULT_HIGH_SPEED_THRESHOLD_KMH,
                                           sprint_speed_threshold_kmh: float = DEFAULT_SPRINT_SPEED_THRESHOLD_KMH) -> pd.Series:
    """
    Calculates total high-intensity running distance and total sprint distance for a single player.

    Args:
        player_tracking_data (pd.DataFrame): Enriched tracking data for a single player.
                                             Must have 'speed_kmh', 'distance_covered_m',
                                             'is_sprinting', 'is_high_intensity_running' columns.
        high_speed_threshold_kmh (float): Threshold for high-intensity running. (Used if flags not pre-calculated)
        sprint_speed_threshold_kmh (float): Threshold for sprinting. (Used if flags not pre-calculated)

    Returns:
        pd.Series: Contains 'total_high_intensity_running_distance_m' and 'total_sprint_distance_m'.
    """
    if not {'speed_kmh', 'distance_covered_m'}.issubset(player_tracking_data.columns):
        raise ValueError("Input DataFrame must contain 'speed_kmh' and 'distance_covered_m'. Consider running enrich_tracking_data first.")

    # Use pre-calculated boolean flags if available, otherwise calculate them
    if 'is_high_intensity_running' not in player_tracking_data.columns:
        is_hir = (player_tracking_data['speed_kmh'] > high_speed_threshold_kmh) & \
                 (player_tracking_data['speed_kmh'] <= sprint_speed_threshold_kmh)
    else:
        is_hir = player_tracking_data['is_high_intensity_running']

    if 'is_sprinting' not in player_tracking_data.columns:
        is_sprint = player_tracking_data['speed_kmh'] > sprint_speed_threshold_kmh
    else:
        is_sprint = player_tracking_data['is_sprinting']

    total_high_intensity_running_distance_m = player_tracking_data.loc[is_hir, 'distance_covered_m'].sum()
    total_sprint_distance_m = player_tracking_data.loc[is_sprint, 'distance_covered_m'].sum()

    return pd.Series({
        'total_high_intensity_running_distance_m': total_high_intensity_running_distance_m,
        'total_sprint_distance_m': total_sprint_distance_m
    })

# --- Accelerations and Decelerations Stats ---

def count_accelerations_decelerations(player_tracking_data: pd.DataFrame,
                                      acceleration_threshold_ms2: float = ACCELERATION_THRESHOLD_MS2,
                                      deceleration_threshold_ms2: float = DECELERATION_THRESHOLD_MS2) -> pd.Series:
    """
    Counts the number of significant accelerations and decelerations for a single player.

    Args:
        player_tracking_data (pd.DataFrame): Enriched tracking data for a single player.
                                             Must have 'acceleration_ms2' column.
        acceleration_threshold_ms2 (float): Minimum positive acceleration to count.
        deceleration_threshold_ms2 (float): Maximum negative acceleration (closest to zero) to count as deceleration.

    Returns:
        pd.Series: Contains 'num_accelerations' and 'num_decelerations'.
    """
    if 'acceleration_ms2' not in player_tracking_data.columns:
        # Attempt to calculate acceleration if not present
        # This assumes 'speed_ms' and 'timestamp_ms' are present, or calculate_acceleration will handle it
        # This is a fallback, ideally enrich_tracking_data is called first.
        if {'speed_ms', 'timestamp_ms', 'player_id'}.issubset(player_tracking_data.columns):
            player_tracking_data = calculate_acceleration(player_tracking_data.copy())
        else:
            raise ValueError("Input DataFrame must contain 'acceleration_ms2'. Consider running enrich_tracking_data first.")

    num_accelerations = (player_tracking_data['acceleration_ms2'] > acceleration_threshold_ms2).sum()
    num_decelerations = (player_tracking_data['acceleration_ms2'] < deceleration_threshold_ms2).sum()

    return pd.Series({
        'num_accelerations': num_accelerations,
        'num_decelerations': num_decelerations
    })

# --- Main Calculation and Aggregation Functions ---

def calculate_player_summary_stats(player_enriched_data: pd.DataFrame,
                                   high_speed_threshold_kmh: float = DEFAULT_HIGH_SPEED_THRESHOLD_KMH,
                                   sprint_speed_threshold_kmh: float = DEFAULT_SPRINT_SPEED_THRESHOLD_KMH,
                                   acceleration_threshold_ms2: float = ACCELERATION_THRESHOLD_MS2,
                                   deceleration_threshold_ms2: float = DECELERATION_THRESHOLD_MS2) -> pd.Series:
    """
    Calculates summary statistics for a single player from their enriched tracking data.

    Args:
        player_enriched_data (pd.DataFrame): Enriched tracking data for one player.
        high_speed_threshold_kmh (float): Threshold for high-intensity running.
        sprint_speed_threshold_kmh (float): Threshold for sprinting.
        acceleration_threshold_ms2 (float): Threshold for counting an acceleration.
        deceleration_threshold_ms2 (float): Threshold for counting a deceleration.

    Returns:
        pd.Series: Summary statistics for the player.
    """
    if player_enriched_data.empty:
        return pd.Series({
            'total_distance_m': 0,
            'total_high_intensity_running_distance_m': 0,
            'total_sprint_distance_m': 0,
            'num_accelerations': 0,
            'num_decelerations': 0,
            'avg_speed_kmh': 0,
            'max_speed_kmh': 0,
            'duration_minutes': 0
        })

    total_distance_m = player_enriched_data['distance_covered_m'].sum()

    intensity_stats = calculate_high_intensity_running_stats(
        player_enriched_data,
        high_speed_threshold_kmh,
        sprint_speed_threshold_kmh
    )

    accel_decel_stats = count_accelerations_decelerations(
        player_enriched_data,
        acceleration_threshold_ms2,
        deceleration_threshold_ms2
    )

    avg_speed_kmh = player_enriched_data['speed_kmh'].mean()
    max_speed_kmh = player_enriched_data['speed_kmh'].max()

    duration_seconds = (player_enriched_data['timestamp_ms'].max() - player_enriched_data['timestamp_ms'].min()) / 1000
    duration_minutes = duration_seconds / 60

    summary = pd.Series({
        'total_distance_m': total_distance_m,
        'avg_speed_kmh': avg_speed_kmh,
        'max_speed_kmh': max_speed_kmh,
        'duration_minutes': duration_minutes
    }).append(intensity_stats).append(accel_decel_stats)

    return summary

def aggregate_stats_by_interval(player_enriched_data: pd.DataFrame,
                                time_interval_minutes: int = 5,
                                high_speed_threshold_kmh: float = DEFAULT_HIGH_SPEED_THRESHOLD_KMH,
                                sprint_speed_threshold_kmh: float = DEFAULT_SPRINT_SPEED_THRESHOLD_KMH,
                                acceleration_threshold_ms2: float = ACCELERATION_THRESHOLD_MS2,
                                deceleration_threshold_ms2: float = DECELERATION_THRESHOLD_MS2) -> pd.DataFrame:
    """
    Aggregates player statistics into time intervals.

    Args:
        player_enriched_data (pd.DataFrame): Enriched tracking data for a single player.
        time_interval_minutes (int): Duration of each interval in minutes.
        Other thresholds: Passed to helper functions.

    Returns:
        pd.DataFrame: DataFrame where each row is an interval with aggregated stats.
    """
    if player_enriched_data.empty or 'timestamp_ms' not in player_enriched_data.columns:
        return pd.DataFrame(columns=[
            'interval_start_time_s', 'interval_end_time_s',
            'distance_m', 'high_intensity_running_distance_m', 'sprint_distance_m',
            'num_accelerations', 'num_decelerations', 'avg_speed_kmh'
        ])

    # Convert timestamp to seconds and make it the index for resampling
    data = player_enriched_data.copy()
    data['time_s'] = data['timestamp_ms'] / 1000

    # Determine the actual start time for relative intervals
    min_time_s = data['time_s'].min()
    data['relative_time_s'] = data['time_s'] - min_time_s

    # Create time bins
    interval_seconds = time_interval_minutes * 60
    max_relative_time = data['relative_time_s'].max()
    bins = np.arange(0, max_relative_time + interval_seconds, interval_seconds)

    data['time_interval_group'] = pd.cut(data['relative_time_s'], bins=bins, right=False, include_lowest=True)

    def aggregate_group(group):
        if group.empty:
            return pd.Series({
                'distance_m': 0, 'high_intensity_running_distance_m': 0, 'sprint_distance_m': 0,
                'num_accelerations': 0, 'num_decelerations': 0, 'avg_speed_kmh': 0
            })

        intensity_stats = calculate_high_intensity_running_stats(
            group, high_speed_threshold_kmh, sprint_speed_threshold_kmh
        )
        accel_decel_stats = count_accelerations_decelerations(
            group, acceleration_threshold_ms2, deceleration_threshold_ms2
        )

        return pd.Series({
            'distance_m': group['distance_covered_m'].sum(),
            'high_intensity_running_distance_m': intensity_stats['total_high_intensity_running_distance_m'],
            'sprint_distance_m': intensity_stats['total_sprint_distance_m'],
            'num_accelerations': accel_decel_stats['num_accelerations'],
            'num_decelerations': accel_decel_stats['num_decelerations'],
            'avg_speed_kmh': group['speed_kmh'].mean() if not group['speed_kmh'].empty else 0
        })

    interval_stats = data.groupby('time_interval_group', observed=False).apply(aggregate_group)

    # Add interval start and end times for clarity
    interval_stats['interval_start_time_s'] = [interval.left + min_time_s for interval in interval_stats.index]
    interval_stats['interval_end_time_s'] = [interval.right + min_time_s for interval in interval_stats.index]

    interval_stats = interval_stats.reset_index(drop=True)
    # Reorder columns
    cols = ['interval_start_time_s', 'interval_end_time_s', 'distance_m',
            'high_intensity_running_distance_m', 'sprint_distance_m',
            'num_accelerations', 'num_decelerations', 'avg_speed_kmh']
    return interval_stats[cols].fillna(0)


# --- Top-Level Functions for API Consumption ---

def generate_all_player_summaries(enriched_tracking_df: pd.DataFrame) -> dict:
    """
    Generates summary statistics for all players in the provided enriched tracking data.

    Args:
        enriched_tracking_df (pd.DataFrame): Enriched tracking data for all players.
                                           Must include 'player_id'.

    Returns:
        dict: Keys are player_ids, values are their summary stats (pd.Series).
    """
    if 'player_id' not in enriched_tracking_df.columns:
        raise ValueError("Input DataFrame must contain 'player_id' column.")
    if enriched_tracking_df.empty:
        return {}

    all_summaries = {}
    for player_id, player_data in enriched_tracking_df.groupby('player_id'):
        all_summaries[player_id] = calculate_player_summary_stats(player_data)
    return all_summaries

def generate_team_summaries(all_player_summaries_dict: dict, player_to_team_map: dict) -> dict:
    """
    Aggregates player summary statistics to team totals.

    Args:
        all_player_summaries_dict (dict): Output from generate_all_player_summaries.
                                          Keys are player_ids, values are pd.Series of stats.
        player_to_team_map (dict): Mapping from player_id to team_id.
                                   Example: {player1_id: teamA_id, player2_id: teamA_id, ...}

    Returns:
        dict: Keys are team_ids, values are their aggregated summary stats (pd.Series).
    """
    team_summaries = {}
    if not all_player_summaries_dict:
        return {}

    # Initialize team stats structure from the first player's summary
    # This assumes all player summaries have the same stats columns
    first_player_stats = next(iter(all_player_summaries_dict.values()))
    stats_cols = first_player_stats.index

    # Create a temporary DataFrame for easier aggregation
    player_stats_list = []
    for player_id, stats in all_player_summaries_dict.items():
        team_id = player_to_team_map.get(player_id, "UnknownTeam") # Handle players not in map
        player_stats_list.append(stats.rename(player_id).to_frame().T.assign(team_id=team_id))

    if not player_stats_list:
        return {}

    all_player_stats_df = pd.concat(player_stats_list)

    # Sum most stats, average the averages, max of maxes
    # Define aggregation functions for each stat type
    agg_funcs = {}
    for col in stats_cols:
        if 'avg_' in col or 'duration_' in col : # Average of averages or durations might not be ideal, sum duration
            agg_funcs[col] = 'mean'
        elif 'max_' in col:
            agg_funcs[col] = 'max'
        else: # Default to sum for totals and counts
            agg_funcs[col] = 'sum'

    # Correct specific aggregations
    if 'duration_minutes' in agg_funcs: agg_funcs['duration_minutes'] = 'sum' # Total duration for a team

    team_summaries_df = all_player_stats_df.groupby('team_id').agg(agg_funcs)

    # Convert back to dictionary of Series
    for team_id in team_summaries_df.index:
        team_summaries[team_id] = team_summaries_df.loc[team_id]

    return team_summaries


def generate_player_time_series(player_enriched_data: pd.DataFrame) -> list:
    """
    Formats enriched tracking data for a single player into a time-series list of dictionaries.

    Args:
        player_enriched_data (pd.DataFrame): Enriched data for one player.

    Returns:
        list: List of dictionaries, each representing a time point.
    """
    if player_enriched_data.empty:
        return []

    relevant_cols = ['timestamp_ms', 'x', 'y', 'speed_kmh', 'distance_covered_m',
                     'is_sprinting', 'is_high_intensity_running', 'acceleration_ms2', 'time_s']
    # Ensure only existing columns are selected
    cols_to_select = [col for col in relevant_cols if col in player_enriched_data.columns]

    return player_enriched_data[cols_to_select].to_dict(orient='records')


def generate_team_intervals(enriched_tracking_df_for_team: pd.DataFrame,
                            time_interval_minutes: int = 5) -> pd.DataFrame:
    """
    Aggregates data for all players in a team into time intervals.
    Calculates total distance, HIR dist, sprint dist, acc, dec, and avg speed for the team per interval.

    Args:
        enriched_tracking_df_for_team (pd.DataFrame): Enriched tracking data for all players in a single team.
        time_interval_minutes (int): Duration of each interval.

    Returns:
        pd.DataFrame: DataFrame where each row is an interval with aggregated team stats.
    """
    if enriched_tracking_df_for_team.empty or 'timestamp_ms' not in enriched_tracking_df_for_team.columns:
        return pd.DataFrame(columns=[
            'interval_start_time_s', 'interval_end_time_s',
            'total_distance_m', 'total_high_intensity_running_distance_m', 'total_sprint_distance_m',
            'total_num_accelerations', 'total_num_decelerations', 'avg_team_speed_kmh'
        ])

    data = enriched_tracking_df_for_team.copy()
    data['time_s'] = data['timestamp_ms'] / 1000 # Ensure time_s is present

    min_time_s = data['time_s'].min()
    data['relative_time_s'] = data['time_s'] - min_time_s

    interval_seconds = time_interval_minutes * 60
    max_relative_time = data['relative_time_s'].max()
    bins = np.arange(0, max_relative_time + interval_seconds, interval_seconds)

    data['time_interval_group'] = pd.cut(data['relative_time_s'], bins=bins, right=False, include_lowest=True)

    def aggregate_team_group(group):
        if group.empty:
            return pd.Series({
                'total_distance_m': 0, 'total_high_intensity_running_distance_m': 0, 'total_sprint_distance_m': 0,
                'total_num_accelerations': 0, 'total_num_decelerations': 0, 'avg_team_speed_kmh': 0
            })

        # Calculate HIR and sprint distances for the group
        # These flags should already be in 'group' from enrich_tracking_data
        hir_dist = group.loc[group['is_high_intensity_running'], 'distance_covered_m'].sum()
        sprint_dist = group.loc[group['is_sprinting'], 'distance_covered_m'].sum()

        # Count accelerations/decelerations for the group
        num_accel = (group['acceleration_ms2'] > ACCELERATION_THRESHOLD_MS2).sum()
        num_decel = (group['acceleration_ms2'] < DECELERATION_THRESHOLD_MS2).sum()

        return pd.Series({
            'total_distance_m': group['distance_covered_m'].sum(),
            'total_high_intensity_running_distance_m': hir_dist,
            'total_sprint_distance_m': sprint_dist,
            'total_num_accelerations': num_accel,
            'total_num_decelerations': num_decel,
            'avg_team_speed_kmh': group['speed_kmh'].mean() if not group['speed_kmh'].empty else 0
        })

    team_interval_stats = data.groupby('time_interval_group', observed=False).apply(aggregate_team_group)

    team_interval_stats['interval_start_time_s'] = [interval.left + min_time_s for interval in team_interval_stats.index]
    team_interval_stats['interval_end_time_s'] = [interval.right + min_time_s for interval in team_interval_stats.index]

    team_interval_stats = team_interval_stats.reset_index(drop=True)
    cols = ['interval_start_time_s', 'interval_end_time_s', 'total_distance_m',
            'total_high_intensity_running_distance_m', 'total_sprint_distance_m',
            'total_num_accelerations', 'total_num_decelerations', 'avg_team_speed_kmh']
    return team_interval_stats[cols].fillna(0)

# Example Usage (for testing, typically called from elsewhere)
if __name__ == '__main__':
    # Create a dummy DataFrame for testing
    # This would typically come from data_loader.py
    num_rows = 2000
    data = {
        'player_id': ['player1'] * (num_rows // 2) + ['player2'] * (num_rows // 2),
        'team_id': ['teamA'] * (num_rows // 2) + ['teamB'] * (num_rows // 2),
        'timestamp_ms': np.concatenate([np.arange(0, (num_rows // 2) * 100, 100), np.arange(0, (num_rows // 2) * 100, 100)]), # 10Hz data
        'x': np.random.rand(num_rows) * 100,
        'y': np.random.rand(num_rows) * 50,
        'smooth_x_speed': np.random.randn(num_rows) * 3, # m/s
        'smooth_y_speed': np.random.randn(num_rows) * 1  # m/s
    }
    sample_tracking_df = pd.DataFrame(data)

    # Test enrich_tracking_data
    print("Enriching data...")
    enriched_df = enrich_tracking_data(sample_tracking_df.copy())
    print("Enriched DataFrame head:\n", enriched_df.head())
    print("\nEnriched DataFrame info:\n")
    enriched_df.info()
    print("\nNaN check in enriched_df:\n", enriched_df.isnull().sum())

    if not enriched_df.empty:
        # Test calculate_player_summary_stats for player1
        player1_data = enriched_df[enriched_df['player_id'] == 'player1']
        if not player1_data.empty:
            print("\nCalculating summary for Player 1...")
            player1_summary = calculate_player_summary_stats(player1_data)
            print("Player 1 Summary:\n", player1_summary)
        else:
            print("\nNo data for Player 1 to summarize.")

        # Test aggregate_stats_by_interval for player1
        if not player1_data.empty:
            print("\nCalculating intervals for Player 1...")
            player1_intervals = aggregate_stats_by_interval(player1_data, time_interval_minutes=1)
            print("Player 1 Intervals (1 min):\n", player1_intervals)
        else:
            print("\nNo data for Player 1 to aggregate into intervals.")

        # Test generate_all_player_summaries
        print("\nGenerating all player summaries...")
        all_summaries = generate_all_player_summaries(enriched_df)
        print("All Player Summaries:")
        for p_id, summary in all_summaries.items():
            print(f"Player {p_id}:\n{summary}\n")

        # Test generate_team_summaries
        # Create a dummy player_to_team_map
        # In a real scenario, this map would come from game metadata or player profiles
        player_team_map = enriched_df[['player_id', 'team_id']].drop_duplicates().set_index('player_id')['team_id'].to_dict()
        print("\nPlayer to Team Map:\n", player_team_map)

        print("\nGenerating team summaries...")
        team_summaries = generate_team_summaries(all_summaries, player_team_map)
        print("Team Summaries:")
        for t_id, summary in team_summaries.items():
            print(f"Team {t_id}:\n{summary}\n")

        # Test generate_player_time_series for player1
        if not player1_data.empty:
            print("\nGenerating time series for Player 1...")
            player1_ts = generate_player_time_series(player1_data)
            print(f"Player 1 Time Series (first 3 records):\n {player1_ts[:3]}")
        else:
            print("\nNo data for Player 1 for time series.")

        # Test generate_team_intervals for teamA
        team_a_data = enriched_df[enriched_df['team_id'] == 'teamA']
        if not team_a_data.empty:
            print("\nGenerating team intervals for Team A...")
            team_a_intervals = generate_team_intervals(team_a_data, time_interval_minutes=1)
            print("Team A Intervals (1 min):\n", team_a_intervals)
        else:
            print("\nNo data for Team A to aggregate into intervals.")
    else:
        print("\nEnriched DataFrame is empty, skipping further tests.")
