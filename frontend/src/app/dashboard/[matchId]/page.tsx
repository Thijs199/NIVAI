'use client';

import { useState, useEffect, ChangeEvent } from 'react';
import { useParams } from 'next/navigation'; // To get [matchId] from URL
import {
  LineChart, BarChart, XAxis, YAxis, CartesianGrid, Tooltip, Legend, Line, Bar, ResponsiveContainer
} from 'recharts'; // Assuming recharts is installed

// --- Data Interfaces ---
interface PlayerSummary {
  player_id: string;
  player_name?: string;
  team_id: string;
  total_distance_m: number;
  total_sprint_distance_m?: number;
  num_accelerations?: number;
  num_decelerations?: number;
  avg_speed_kmh?: number;
  max_speed_kmh?: number;
}

interface TeamSummaryStats {
  total_distance_m: number;
  total_sprint_distance_m?: number;
  total_high_intensity_running_distance_m?: number;
  total_num_accelerations?: number;
  total_num_decelerations?: number;
  avg_speed_kmh?: number;
  max_speed_kmh?: number;
  team_name?: string; // Added for displaying team name in dropdown
}

interface MatchAnalyticsSummary {
  match_id: string;
  players: Record<string, PlayerSummary>;
  teams: Record<string, TeamSummaryStats>;
}

interface PlayerTimeSeriesDataPoint {
  timestamp_ms: number;
  time_s: number;
  speed_kmh: number;
  is_sprinting: boolean;
  is_high_intensity_running: boolean;
  acceleration_ms2: number;
  distance_covered_m: number;
}

interface TeamIntervalDataPoint {
  interval_start_time_s: number;
  interval_end_time_s: number;
  total_distance_m: number;
  total_high_intensity_running_distance_m: number;
  total_sprint_distance_m: number;
  total_num_accelerations: number;
  total_num_decelerations: number;
  avg_team_speed_kmh: number;
}


// --- Component ---
export default function MatchDashboardPage() {
  const params = useParams();
  const matchId = params.matchId as string;

  const [summaryData, setSummaryData] = useState<MatchAnalyticsSummary | null>(null);
  const [allPlayersList, setAllPlayersList] = useState<PlayerSummary[]>([]);

  const [selectedPlayerId, setSelectedPlayerId] = useState<string | null>(null);
  const [playerTimeSeries, setPlayerTimeSeries] = useState<PlayerTimeSeriesDataPoint[] | null>(null);
  const [playerImageUrl, setPlayerImageUrl] = useState<string | null>(null);
  const [isLoadingImage, setIsLoadingImage] = useState<boolean>(false);
  const [imageError, setImageError] = useState<string | null>(null);

  const [selectedTeamIdForChart, setSelectedTeamIdForChart] = useState<string | null>(null);
  const [teamIntervalData, setTeamIntervalData] = useState<TeamIntervalDataPoint[] | null>(null);

  const [isLoadingSummary, setIsLoadingSummary] = useState<boolean>(true);
  const [isLoadingPlayer, setIsLoadingPlayer] = useState<boolean>(false); // For time series
  const [isLoadingTeamInterval, setIsLoadingTeamInterval] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null); // General error for summary or critical data

  // Fetch Match Summary
  useEffect(() => {
    if (matchId) {
      setIsLoadingSummary(true);
      setError(null);
      fetch(`/api/v1/analytics/matches/${matchId}`)
        .then(res => {
          if (!res.ok) throw new Error(`Failed to fetch match summary: ${res.status} ${res.statusText}`);
          return res.json();
        })
        .then((data: MatchAnalyticsSummary) => {
          setSummaryData(data);
          const playersArray = data.players ? Object.values(data.players).map(p => ({...p, player_name: p.player_name || p.player_id })) : [];
          setAllPlayersList(playersArray);

          if (data.teams && Object.keys(data.teams).length > 0) {
            const firstTeamId = Object.keys(data.teams)[0];
            setSelectedTeamIdForChart(firstTeamId);
          }
        })
        .catch(err => {
          console.error(err);
          setError(err.message);
        })
        .finally(() => setIsLoadingSummary(false));
    }
  }, [matchId]);

  // Fetch Player Time Series Data AND Player Image
  useEffect(() => {
    if (selectedPlayerId && matchId) {
      // Reset previous player specific data
      setPlayerTimeSeries(null);
      setPlayerImageUrl(null);
      setImageError(null);
      setIsLoadingPlayer(true); // For time series
      setIsLoadingImage(true); // For image

      // Fetch Time Series
      fetch(`/api/v1/analytics/players/${selectedPlayerId}/details?match_id=${matchId}`)
        .then(res => {
          if (!res.ok) throw new Error(`Player details: ${res.status} ${res.statusText}`);
          return res.json();
        })
        .then((data: { time_series: PlayerTimeSeriesDataPoint[] }) => {
            const processedData = data.time_series.map(d => ({...d, time_s: d.timestamp_ms / 1000 }));
            setPlayerTimeSeries(processedData);
        })
        .catch(err => {
          console.error(err);
          setError(prev => prev ? `${prev}\nFailed to load player time series: ${err.message}` : `Failed to load player time series: ${err.message}`);
        })
        .finally(() => setIsLoadingPlayer(false));

      // Fetch Player Image
      const playerData = allPlayersList.find(p => p.player_id === selectedPlayerId);
      const playerName = playerData?.player_name || selectedPlayerId; // Fallback to ID if name not found

      if (playerName) {
        fetch(`/api/v1/analytics/players/image_search?name=${encodeURIComponent(playerName)}`)
          .then(res => {
            if (!res.ok) throw new Error(`Player image: ${res.status} ${res.statusText}`);
            return res.json();
          })
          .then(data => {
            setPlayerImageUrl(data.image_url);
          })
          .catch(err => {
            console.error(err);
            setImageError(`Failed to load player image: ${err.message}`);
          })
          .finally(() => setIsLoadingImage(false));
      } else {
        setIsLoadingImage(false);
        setImageError("Player name not found to search image.");
      }
    } else {
        // Clear player specific data if no player is selected
        setPlayerTimeSeries(null);
        setPlayerImageUrl(null);
        setImageError(null);
    }
  }, [selectedPlayerId, matchId, allPlayersList]); // allPlayersList added as dependency for playerName

  // Fetch Team Interval Data
  useEffect(() => {
    if (selectedTeamIdForChart && matchId) {
      setIsLoadingTeamInterval(true);
      setTeamIntervalData(null); // Clear previous data
      fetch(`/api/v1/analytics/teams/${selectedTeamIdForChart}/summary-over-time?match_id=${matchId}`)
        .then(res => {
          if (!res.ok) throw new Error(`Team interval data: ${res.status} ${res.statusText}`);
          return res.json();
        })
        .then((data: { intervals: TeamIntervalDataPoint[] }) => {
            setTeamIntervalData(data.intervals);
        })
        .catch(err => {
          console.error(err);
          setError(prev => prev ? `${prev}\nFailed to load team interval data: ${err.message}` : `Failed to load team interval data: ${err.message}`);
        })
        .finally(() => setIsLoadingTeamInterval(false));
    }
  }, [selectedTeamIdForChart, matchId]);

  // --- Render Logic ---
  if (isLoadingSummary) return <div className="p-4 text-center">Loading match data...</div>;
  if (error && !summaryData && !isLoadingSummary) return <div className="p-4 text-center text-red-500">Error loading match data: {error}</div>;
  if (!summaryData) return <div className="p-4 text-center">No summary data available for this match.</div>;

  const selectedPlayerData = selectedPlayerId ? allPlayersList.find(p => p.player_id === selectedPlayerId) : null;

  const playerDistanceData = allPlayersList.map(p => ({
    name: p.player_name || p.player_id,
    distance: p.total_distance_m / 1000,
    sprint_distance: (p.total_sprint_distance_m || 0) / 1000,
  }));

  const teamDistanceData = Object.entries(summaryData.teams).map(([teamId, teamStats]) => ({
      name: teamStats.team_name || teamId,
      distance: teamStats.total_distance_m / 1000,
  }));


  return (
    <div className="container mx-auto p-4 space-y-8">
      <h1 className="text-3xl font-bold text-center text-slate-800">Match Dashboard: {summaryData.match_id || matchId}</h1>
      {error && (
        <div className="p-4 my-4 bg-red-100 text-red-700 border-l-4 border-red-500 rounded-md text-sm" role="alert">
          <h3 className="font-bold mb-1">Encountered an error:</h3>
          {/* Split errors by newline if multiple were concatenated, otherwise show as is */}
          <pre className="whitespace-pre-wrap text-xs">{error}</pre>
        </div>
      )}

      {/* --- Team Overview Section --- */}
      <section className="p-6 bg-white shadow-lg rounded-lg">
        <h2 className="text-2xl font-semibold mb-4 text-slate-700">Team Overview</h2>
        <div className="grid md:grid-cols-2 gap-6">
            <div>
                <h3 className="text-xl font-medium mb-2 text-slate-600">Total Distance Run by Team (km)</h3>
                <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={teamDistanceData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis label={{ value: 'Distance (km)', angle: -90, position: 'insideLeft' }} />
                    <Tooltip formatter={(value: number) => `${value.toFixed(2)} km`} />
                    <Legend />
                    <Bar dataKey="distance" fill="#8884d8" name="Total Distance" />
                    </BarChart>
                </ResponsiveContainer>
            </div>
            <div>
                <h3 className="text-xl font-medium mb-2 text-slate-600">Total Distance Run by Player (km)</h3>
                 <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={playerDistanceData} layout="vertical">
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis type="number"  label={{ value: 'Distance (km)', position: 'insideBottom', offset: -5 }}/>
                        <YAxis type="category" dataKey="name" width={100} />
                        <Tooltip formatter={(value: number) => `${value.toFixed(2)} km`} />
                        <Legend />
                        <Bar dataKey="distance" fill="#82ca9d" name="Total Distance" />
                        <Bar dataKey="sprint_distance" fill="#ffc658" name="Sprint Distance" />
                    </BarChart>
                </ResponsiveContainer>
            </div>
        </div>
        <div className="mt-6">
            <h3 className="text-xl font-medium mb-2 text-slate-600">Team Intensity Over Time</h3>
            <select
                onChange={(e: ChangeEvent<HTMLSelectElement>) => setSelectedTeamIdForChart(e.target.value)}
                value={selectedTeamIdForChart || ""}
                className="p-2 border rounded mb-2 bg-white shadow-sm focus:border-indigo-300 focus:ring focus:ring-indigo-200 focus:ring-opacity-50"
            >
                <option value="" disabled>Select a Team</option>
                {Object.entries(summaryData.teams).map(([teamId, teamDetails]) => (
                    <option key={teamId} value={teamId}>{teamDetails.team_name || teamId}</option>
                ))}
            </select>
            {isLoadingTeamInterval && <p className="p-2">Loading team interval data...</p>}
            {teamIntervalData && teamIntervalData.length > 0 && (
                <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={teamIntervalData.map(d => ({...d, interval: `${(d.interval_start_time_s / 60).toFixed(0)}-${(d.interval_end_time_s / 60).toFixed(0)}min`}))}>
                        <CartesianGrid strokeDasharray="3 3" />
                        <XAxis dataKey="interval" />
                        <YAxis yAxisId="left" label={{ value: 'Avg Speed (km/h)', angle: -90, position: 'insideLeft' }} />
                        <YAxis yAxisId="right" orientation="right" label={{ value: 'Distance (m)', angle: -90, position: 'insideRight' }} />
                        <Tooltip />
                        <Legend />
                        <Line yAxisId="left" type="monotone" dataKey="avg_team_speed_kmh" stroke="#8884d8" name="Avg Team Speed" dot={false}/>
                        <Line yAxisId="right" type="monotone" dataKey="total_distance_m" stroke="#82ca9d" name="Distance Covered" dot={false}/>
                    </LineChart>
                </ResponsiveContainer>
            )}
            {!teamIntervalData && !isLoadingTeamInterval && selectedTeamIdForChart && <p className="p-2">No interval data available for {summaryData.teams[selectedTeamIdForChart]?.team_name || selectedTeamIdForChart}.</p>}
        </div>
      </section>

      {/* --- Player Details Section --- */}
      <section className="p-6 bg-white shadow-lg rounded-lg">
        <h2 className="text-2xl font-semibold mb-4 text-slate-700">Player Details</h2>
        <div className="mb-4">
          <label htmlFor="player-select" className="block text-sm font-medium text-gray-700 mr-2">Select Player:</label>
          <select
            id="player-select"
            value={selectedPlayerId || ''}
            onChange={(e: ChangeEvent<HTMLSelectElement>) => setSelectedPlayerId(e.target.value)}
            className="mt-1 block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md bg-white shadow-sm"
          >
            <option value="" disabled>-- Select a Player --</option>
            {allPlayersList.map(player => (
              <option key={player.player_id} value={player.player_id}>
                {player.player_name || player.player_id} (Team: {player.team_id})
              </option>
            ))}
          </select>
        </div>

        {selectedPlayerId && selectedPlayerData && (
          <div className="grid md:grid-cols-3 gap-6 items-start">
            <div className="md:col-span-1 p-4 border rounded bg-slate-50 shadow-sm">
                <h3 className="text-xl font-medium mb-2 text-slate-600">{selectedPlayerData.player_name || selectedPlayerData.player_id}</h3>
                <div className="player-image-section mb-4">
                  {isLoadingImage && <div className="w-32 h-32 bg-gray-200 flex items-center justify-center text-gray-500 rounded-md border animate-pulse">Loading Image...</div>}
                  {imageError && <div className="w-32 h-32 bg-red-100 flex items-center justify-center text-red-500 rounded-md border p-2">{imageError}</div>}
                  {!isLoadingImage && !imageError && playerImageUrl && (
                    <img src={playerImageUrl} alt={`Profile of ${selectedPlayerData.player_name || selectedPlayerData.player_id}`} className="w-32 h-32 object-cover rounded-md border shadow" />
                  )}
                  {!isLoadingImage && !imageError && !playerImageUrl && (
                    <div className="w-32 h-32 bg-gray-200 flex items-center justify-center text-gray-500 rounded-md border">
                      No Image Available
                    </div>
                  )}
                </div>
                <p><strong>Team:</strong> {selectedPlayerData.team_id}</p>
                <p><strong>Total Distance:</strong> {(selectedPlayerData.total_distance_m / 1000).toFixed(2)} km</p>
                {selectedPlayerData.total_sprint_distance_m !== undefined && <p><strong>Sprint Distance:</strong> {(selectedPlayerData.total_sprint_distance_m / 1000).toFixed(2)} km</p>}
                {selectedPlayerData.num_accelerations !== undefined && <p><strong>Accelerations:</strong> {selectedPlayerData.num_accelerations}</p>}
                {selectedPlayerData.num_decelerations !== undefined && <p><strong>Decelerations:</strong> {selectedPlayerData.num_decelerations}</p>}
                {selectedPlayerData.max_speed_kmh !== undefined && <p><strong>Max Speed:</strong> {selectedPlayerData.max_speed_kmh.toFixed(2)} km/h</p>}
            </div>
            <div className="md:col-span-2">
                <h3 className="text-xl font-medium mb-2 text-slate-600">Player Speed Over Time</h3>
                {isLoadingPlayer && <p>Loading player time series data...</p>}
                {playerTimeSeries && playerTimeSeries.length > 0 ? (
                <ResponsiveContainer width="100%" height={300}>
                    <LineChart data={playerTimeSeries}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="time_s" unit="s" label={{ value: 'Time (s)', position: 'insideBottom', offset: -5 }}/>
                    <YAxis label={{ value: 'Speed (km/h)', angle: -90, position: 'insideLeft' }}/>
                    <Tooltip formatter={(value: number, name: string) => name === 'Speed' ? `${value.toFixed(2)} km/h` : value} />
                    <Legend />
                    <Line type="monotone" dataKey="speed_kmh" stroke="#8884d8" name="Speed" dot={false} />
                    </LineChart>
                </ResponsiveContainer>
                ) : (
                    <p>{!isLoadingPlayer && selectedPlayerId ? 'No time series data available for this player.' : ''}</p>
                )}
            </div>
          </div>
        )}
         {!selectedPlayerId && <p className="text-center text-slate-500 py-4">Select a player to view details.</p>}
      </section>
    </div>
  );
}
