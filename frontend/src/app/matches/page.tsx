'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation'; // For App Router

interface MatchListItem {
  id: string;
  match_name: string;
  upload_date: string;
  analytics_status: string;
  home_team?: string;
  away_team?: string;
  // Add any other fields your Go backend's MatchListItem might have
  competition?: string;
  season?: string;
}

export default function MatchesPage() {
  const [matches, setMatches] = useState<MatchListItem[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchMatches = async () => {
      setIsLoading(true);
      setError(null);
      try {
        // Assuming the Go backend is on the same origin or proxied via Next.js rewrites
        // Adjust '/api/v1/matches' if your setup is different (e.g., http://localhost:8000/api/v1/matches)
        const response = await fetch('/api/v1/matches');

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ message: response.statusText }));
          throw new Error(`Failed to fetch matches: ${errorData.message || response.statusText}`);
        }
        const data: MatchListItem[] = await response.json();
        setMatches(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : String(err));
        console.error("Error fetching matches:", err);
      } finally {
        setIsLoading(false);
      }
    };
    fetchMatches();
  }, []); // Empty dependency array means this runs once on mount

  const getStatusColor = (status: string): string => {
    status = status.toLowerCase();
    if (status.includes('processed') || status.includes('completed')) {
      return 'text-green-600 bg-green-100 px-2 py-0.5 rounded-full text-xs font-medium';
    }
    if (status.includes('pending') || status.includes('processing')) {
      return 'text-yellow-600 bg-yellow-100 px-2 py-0.5 rounded-full text-xs font-medium';
    }
    if (status.includes('error') || status.includes('failed')) {
      return 'text-red-600 bg-red-100 px-2 py-0.5 rounded-full text-xs font-medium';
    }
    return 'text-gray-600 bg-gray-100 px-2 py-0.5 rounded-full text-xs font-medium'; // Default/unknown
  };

  const handleMatchClick = (matchId: string) => {
    router.push(`/dashboard/${matchId}`);
  };

  if (isLoading) {
    return <div className="container mx-auto p-4 text-center">Loading matches...</div>;
  }

  if (error) {
    return <div className="container mx-auto p-4 text-center text-red-500">Error: {error}</div>;
  }

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-8 text-center text-slate-800">Available Matches</h1>
      {matches.length === 0 ? (
        <p className="text-center text-gray-500">No matches found.</p>
      ) : (
        <ul className="space-y-4">
          {matches.map((match) => (
            <li
              key={match.id}
              onClick={() => handleMatchClick(match.id)}
              className="bg-white shadow-lg rounded-lg p-6 cursor-pointer hover:shadow-xl transition-shadow duration-200 ease-in-out border border-slate-200"
            >
              <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
                <div className="mb-2 sm:mb-0">
                    <h3 className="text-xl font-semibold text-blue-700 hover:text-blue-800">{match.match_name}</h3>
                    {match.home_team && match.away_team && (
                        <p className="text-sm text-slate-600">{match.home_team} vs {match.away_team}</p>
                    )}
                    {match.competition && (
                        <p className="text-sm text-slate-500">{match.competition} - {match.season}</p>
                    )}
                </div>
                <div className="flex flex-col items-start sm:items-end space-y-1">
                    <span className={getStatusColor(match.analytics_status)}>
                        {match.analytics_status.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                    </span>
                    <p className="text-xs text-slate-400">
                        Uploaded: {new Date(match.upload_date).toLocaleDateString()} {new Date(match.upload_date).toLocaleTimeString()}
                    </p>
                </div>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
