// For Match List
export interface MatchListItem {
  id: string;
  match_name: string;
  upload_date: string; // Assuming ISO string date
  analytics_status: string; // e.g., "processed", "pending", "error_processing", "unknown"
  home_team?: string;
  away_team?: string;
  competition?: string;
  season?: string;
}

// For Player Summary Stats (from Python API's /match/{matchid}/stats/summary)
export interface PlayerSummaryStats {
  player_id?: string; // Added to identify player if not part of the key
  total_distance_m: number;
  avg_speed_kmh: number;
  max_speed_kmh: number;
  duration_minutes: number;
  total_high_intensity_running_distance_m: number;
  total_sprint_distance_m: number;
  num_accelerations: number;
  num_decelerations: number;
}

// For Overall Match Analytics Summary
export interface MatchAnalyticsSummary {
  match_id: string;
  players: { [playerId: string]: PlayerSummaryStats }; // Player ID -> Stats
  teams: { [teamId: string]: any }; // Team stats (can be more specific if needed)
}
