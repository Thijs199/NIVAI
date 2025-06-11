/**
 * Dashboard is the main hub for football analytics visualizations and metrics.
 * Provides an overview of key insights and access to detailed analysis features.
 *
 * @returns The dashboard page component
 */

"use client";

import React, { useEffect, useState } from 'react';
// Adjust path if your types/api files are located differently
import { MatchListItem, MatchAnalyticsSummary, PlayerSummaryStats } from '../../types/analytics';
import { fetchMatches, fetchMatchAnalyticsSummary } from '../../lib/api';

export default function Dashboard() {
  const [matches, setMatches] = useState<MatchListItem[]>([]);
  const [recentMatchesForDisplay, setRecentMatchesForDisplay] = useState<MatchListItem[]>([]);
  const [topSpeedRecord, setTopSpeedRecord] = useState<{ value: number; player: string; matchName: string } | null>(null);
  const [latestMatchPlayerPerformance, setLatestMatchPlayerPerformance] = useState<PlayerSummaryStats[]>([]);
  const [isLoadingMatches, setIsLoadingMatches] = useState(true);
  const [isLoadingAnalytics, setIsLoadingAnalytics] = useState(true); // Start true as analytics loads after matches
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function loadDashboardData() {
      setIsLoadingMatches(true);
      setError(null);
      try {
        const fetchedMatches = await fetchMatches();
        const sortedMatches = [...fetchedMatches].sort((a, b) => new Date(b.upload_date).getTime() - new Date(a.upload_date).getTime());
        setMatches(sortedMatches);
        setRecentMatchesForDisplay(sortedMatches.slice(0, 3));
      } catch (err) {
        console.error("Error fetching matches:", err);
        setError(err instanceof Error ? err.message : "Unknown error fetching matches");
      } finally {
        setIsLoadingMatches(false);
      }
    }
    loadDashboardData();
  }, []);

  useEffect(() => {
    async function loadAnalyticsData() {
      if (matches.length === 0) {
        setIsLoadingAnalytics(false);
        return;
      }

      setIsLoadingAnalytics(true);
      // setError(null); // Don't reset error if matches loading failed previously. Only for analytics specific errors.

      let overallTopSpeed = 0;
      let topSpeedPlayerInfo = "N/A";
      let topSpeedMatchName = "N/A";

      const processedMatches = matches.filter(m => m.analytics_status === 'processed' || m.analytics_status === 'completed');

      if (processedMatches.length === 0) {
         setIsLoadingAnalytics(false);
         setTopSpeedRecord(null); // No processed matches, so no top speed.
         setLatestMatchPlayerPerformance([]); // No player performance data.
         return;
      }

      let analyticsErrorOccurred = false;

      // For Top Speed Record - Analyze up to 5 recent processed matches
      for (const match of processedMatches.slice(0, 5)) {
        try {
          const summary = await fetchMatchAnalyticsSummary(match.id);
          if (summary.players) {
            for (const playerId in summary.players) {
              const playerStats = summary.players[playerId];
              if (playerStats.max_speed_kmh > overallTopSpeed) {
                overallTopSpeed = playerStats.max_speed_kmh;
                topSpeedPlayerInfo = `Player ${playerId.substring(0, 6)}`;
                topSpeedMatchName = match.match_name || `Match ${match.id.substring(0,6)}`;
              }
            }
          }
        } catch (err) {
          console.error(`Error fetching analytics for top speed (match ${match.id}):`, err);
          analyticsErrorOccurred = true;
        }
      }
      if (overallTopSpeed > 0) {
        setTopSpeedRecord({ value: overallTopSpeed, player: topSpeedPlayerInfo, matchName: topSpeedMatchName });
      } else if (!analyticsErrorOccurred) { // If no errors but no speed, set to null
         setTopSpeedRecord(null);
      }


      // For Player Performance Widget - Use the most recent processed match
      const latestProcessedMatch = processedMatches[0];
      if (latestProcessedMatch) {
        try {
          const summary = await fetchMatchAnalyticsSummary(latestProcessedMatch.id);
          const playersArray = Object.entries(summary.players).map(([id, stats]) => ({
              player_id: id,
              ...stats
          }));
          const sortedPlayers = playersArray.sort((a, b) => b.max_speed_kmh - a.max_speed_kmh).slice(0, 4);
          setLatestMatchPlayerPerformance(sortedPlayers);
        } catch (err) {
          console.error(`Error fetching analytics for player performance (match ${latestProcessedMatch.id}):`, err);
          setError(err instanceof Error ? err.message : "Unknown error fetching player performance"); // Set specific error for this part
          setLatestMatchPlayerPerformance([]); // Clear previous data on error
        }
      } else {
         setLatestMatchPlayerPerformance([]); // No processed matches for player performance
      }

      if (analyticsErrorOccurred && !error) { // If there was a partial error during top speed scan, set a general message
         setError("Some analytics data could not be loaded.");
      }

      setIsLoadingAnalytics(false);
    }

    if (!isLoadingMatches && matches.length > 0) {
      loadAnalyticsData();
    } else if (!isLoadingMatches && matches.length === 0) {
      setIsLoadingAnalytics(false);
    }
  }, [matches, isLoadingMatches]);

  return (
    <div className="bg-gray-50 min-h-screen">
      {/* Dashboard Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="md:flex md:items-center md:justify-between">
            <div className="flex-1 min-w-0">
              <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                Analytics Dashboard
              </h1>
            </div>
            <div className="mt-4 flex md:mt-0 md:ml-4">
              <select className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                <option value="current">Current Season</option>
                <option value="2022-2023">2022-2023</option>
                <option value="2021-2022">2021-2022</option>
              </select>
            </div>
          </div>
        </div>
      </header>

      {/* Main Dashboard Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Quick Stats Row */}
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
          {/* Stat Card 1 */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <dl>
                <dt className="text-sm font-medium text-gray-500 truncate">Total Matches</dt>
                <dd className="mt-1 text-3xl font-semibold text-gray-900">248</dd>
              </dl>
            </div>
            <div className="bg-gray-50 px-4 py-2">
              <div className="text-sm flex justify-between">
                <span className="font-medium text-blue-700">View all</span>
                <span className="text-green-600">+12% from last season</span>
              </div>
            </div>
          </div>

          {/* Stat Card 2 */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <dl>
                <dt className="text-sm font-medium text-gray-500 truncate">Average Goals</dt>
                <dd className="mt-1 text-3xl font-semibold text-gray-900">2.7</dd>
              </dl>
            </div>
            <div className="bg-gray-50 px-4 py-2">
              <div className="text-sm flex justify-between">
                <span className="font-medium text-blue-700">View details</span>
                <span className="text-red-600">-0.3 from last season</span>
              </div>
            </div>
          </div>

          {/* Stat Card 3 */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <dl>
                <dt className="text-sm font-medium text-gray-500 truncate">Possession Average</dt>
                <dd className="mt-1 text-3xl font-semibold text-gray-900">52%</dd>
              </dl>
            </div>
            <div className="bg-gray-50 px-4 py-2">
              <div className="text-sm flex justify-between">
                <span className="font-medium text-blue-700">View details</span>
                <span className="text-green-600">+2% from last season</span>
              </div>
            </div>
          </div>

          {/* Stat Card 4 */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <dl>
                <dt className="text-sm font-medium text-gray-500 truncate">Top Speed Record</dt>
                {isLoadingAnalytics ? (
                  <dd className="mt-1 text-3xl font-semibold text-gray-400">Loading...</dd>
                ) : topSpeedRecord ? (
                  <dd className="mt-1 text-3xl font-semibold text-gray-900">{topSpeedRecord.value.toFixed(1)} km/h</dd>
                ) : (
                  <dd className="mt-1 text-3xl font-semibold text-gray-400">N/A</dd>
                )}
              </dl>
            </div>
            <div className="bg-gray-50 px-4 py-2">
              <div className="text-sm flex justify-between">
                <span className="font-medium text-blue-700">Player Details</span>
                {topSpeedRecord ? (<span className="text-gray-600 truncate" title={`${topSpeedRecord.player} - ${topSpeedRecord.matchName}`}>{topSpeedRecord.player} - {topSpeedRecord.matchName}</span>) : (<span className="text-gray-400">N/A</span>)}
              </div>
            </div>
          </div>
        </div>

        {/* Analysis Cards Row */}
        <div className="grid grid-cols-1 gap-5 lg:grid-cols-2 mb-8">
          {/* Interactive Pitch Widget */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-3">
                Team Formation Analysis
              </h3>
              <div className="aspect-[4/3] bg-green-100 rounded relative">
                {/* This is where we would render the Pixi.js canvas */}
                <div className="absolute inset-0 flex items-center justify-center">
                  <p className="text-gray-500">
                    Interactive pitch visualization would render here with Pixi.js
                  </p>
                </div>
              </div>
              <div className="mt-3 flex justify-end">
                <button className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-blue-700 hover:bg-blue-50">
                  Explore Formation
                </button>
              </div>
            </div>
          </div>

          {/* Player Performance Widget */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-3">
                Player Performance (Last Processed Match)
              </h3>
              {isLoadingAnalytics && <p className="text-sm text-gray-500">Loading performance data...</p>}
              {!isLoadingAnalytics && error && latestMatchPlayerPerformance.length === 0 && <p className="text-sm text-red-500">Error: {error}</p>}
              {!isLoadingAnalytics && !error && latestMatchPlayerPerformance.length === 0 && <p className="text-sm text-gray-500">No player performance data available for the latest processed match.</p>}
              <div className="space-y-4">
                {latestMatchPlayerPerformance.map(player => (
                  <div key={player.player_id} className="flex items-center">
                    <div className="flex-shrink-0 h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                      <span className="text-xs font-medium text-gray-700">{(player.player_id || 'P').substring(0,2).toUpperCase()}</span>
                    </div>
                    <div className="ml-4 flex-1">
                      <div className="flex items-center justify-between">
                        <h4 className="text-sm font-medium text-gray-900 truncate" title={`Player ${(player.player_id || 'Unknown').substring(0,6)}`}>Player {(player.player_id || 'Unknown').substring(0,6)}</h4>
                      </div>
                      <div className="mt-1 text-xs text-gray-600">
                        Max Speed: <span className="font-semibold">{player.max_speed_kmh.toFixed(1)} km/h</span> |
                        Distance: <span className="font-semibold">{(player.total_distance_m / 1000).toFixed(2)} km</span>
                      </div>
                      <div className="text-xs text-gray-600">
                        Sprints Dist: <span className="font-semibold">{player.total_sprint_distance_m.toFixed(0)}m</span> |
                        Accels: <span className="font-semibold">{player.num_accelerations}</span> |
                        Decels: <span className="font-semibold">{player.num_decelerations}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
              {/* Optional: "View All Players" button if more detailed view is available */}
            </div>
          </div>
        </div>

        {/* Match Analysis & Video Highlights Row */}
        <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
          {/* Recent Matches */}
          <div className="bg-white overflow-hidden shadow rounded-lg lg:col-span-2">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-3">Recent Matches</h3>
              {isLoadingMatches && <p className="text-sm text-gray-500">Loading matches...</p>}
              {error && !isLoadingMatches && <p className="text-sm text-red-500">Error loading matches: {error}</p>}
              {!isLoadingMatches && !error && recentMatchesForDisplay.length === 0 && <p className="text-sm text-gray-500">No matches found.</p>}
              <div className="space-y-3">
                {recentMatchesForDisplay.map(match => (
                  <div key={match.id} className="border border-gray-200 rounded-md p-4 hover:border-blue-500 transition-colors duration-200 cursor-pointer">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-4">
                        <span className="text-sm font-medium">{new Date(match.upload_date).toLocaleDateString()}</span>
                        <span className="text-gray-500">|</span>
                        <span className="font-medium truncate pr-2" title={match.match_name || `${match.home_team || 'Home'} vs ${match.away_team || 'Away'}`}>{match.match_name || `${match.home_team || 'Home'} vs ${match.away_team || 'Away'}`}</span>
                      </div>
                      <div>
                        <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                          match.analytics_status === 'processed' || match.analytics_status === 'completed' ? 'bg-green-100 text-green-800' :
                          match.analytics_status === 'pending' ? 'bg-yellow-100 text-yellow-800' :
                          match.analytics_status && (match.analytics_status.startsWith('error') || match.analytics_status === 'failed') ? 'bg-red-100 text-red-800' :
                          'bg-gray-100 text-gray-800'
                        }`}>
                          {match.analytics_status ? match.analytics_status.replace(/_/g, ' ').toLowerCase() : 'Unknown'}
                        </span>
                      </div>
                    </div>
                    <div className="mt-2 flex items-center justify-between">
                      <div className="text-sm text-gray-500 truncate">
                        Comp: {match.competition || 'N/A'} | Season: {match.season || 'N/A'}
                      </div>
                      <button className="inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
                        View Analysis
                      </button>
                    </div>
                  </div>
                ))}
              </div>
              {/* Potentially add a "View All Matches" button if matches.length > recentMatchesForDisplay.length */}
            </div>
          </div>

          {/* Video Highlights */}
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-3">Video Highlights</h3>
              <div className="space-y-4">
                <div className="aspect-video bg-gray-100 rounded-md relative">
                  <div className="absolute inset-0 flex items-center justify-center">
                    <div className="h-16 w-16 rounded-full bg-blue-600 bg-opacity-75 flex items-center justify-center">
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        strokeWidth={1.5}
                        stroke="currentColor"
                        className="w-8 h-8 text-white"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z"
                        />
                      </svg>
                    </div>
                  </div>
                  <div className="absolute bottom-2 left-2 bg-black bg-opacity-70 px-2 py-1 rounded text-xs text-white">
                    Ajax vs PSV Highlights
                  </div>
                </div>
                <div className="flex overflow-x-auto space-x-3 pb-1">
                  <div className="flex-shrink-0 w-24 h-16 bg-gray-100 rounded cursor-pointer"></div>
                  <div className="flex-shrink-0 w-24 h-16 bg-gray-100 rounded cursor-pointer"></div>
                  <div className="flex-shrink-0 w-24 h-16 bg-gray-100 rounded cursor-pointer"></div>
                </div>
                <div className="mt-2 flex justify-end">
                  <button className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-blue-700 hover:bg-blue-50">
                    More Videos
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
