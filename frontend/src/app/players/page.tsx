import Link from 'next/link';

/**
 * PlayersPage displays comprehensive player statistics and performance metrics.
 * Allows filtering, searching and comparing players across multiple dimensions.
 *
 * @returns The players page component
 */
export default function PlayersPage() {
  // Mock data for player statistics
  const playerData = [
    {
      id: 1,
      name: 'Antony Matheus',
      team: 'Ajax',
      position: 'Winger',
      nationality: 'Brazil',
      age: 23,
      stats: {
        goals: 12,
        assists: 9,
        matches: 28,
        rating: 8.7,
        passAccuracy: 87,
        shotsOnTarget: 64,
        distanceCovered: 312.4,
        topSpeed: 34.8,
      },
    },
    {
      id: 2,
      name: 'Jurriën Timber',
      team: 'Ajax',
      position: 'Defender',
      nationality: 'Netherlands',
      age: 21,
      stats: {
        goals: 4,
        assists: 2,
        matches: 34,
        rating: 8.2,
        passAccuracy: 92,
        shotsOnTarget: 38,
        distanceCovered: 341.2,
        topSpeed: 36.2,
      },
    },
    {
      id: 3,
      name: 'Cody Gakpo',
      team: 'PSV',
      position: 'Winger',
      nationality: 'Netherlands',
      age: 24,
      stats: {
        goals: 19,
        assists: 8,
        matches: 32,
        rating: 7.8,
        passAccuracy: 83,
        shotsOnTarget: 71,
        distanceCovered: 297.5,
        topSpeed: 35.1,
      },
    },
    {
      id: 4,
      name: 'Daley Klaassen',
      team: 'Ajax',
      position: 'Midfielder',
      nationality: 'Netherlands',
      age: 30,
      stats: {
        goals: 8,
        assists: 11,
        matches: 29,
        rating: 7.5,
        passAccuracy: 89,
        shotsOnTarget: 42,
        distanceCovered: 324.8,
        topSpeed: 32.5,
      },
    },
    {
      id: 5,
      name: 'Ibrahim Sangaré',
      team: 'PSV',
      position: 'Midfielder',
      nationality: 'Ivory Coast',
      age: 25,
      stats: {
        goals: 5,
        assists: 4,
        matches: 31,
        rating: 7.4,
        passAccuracy: 86,
        shotsOnTarget: 26,
        distanceCovered: 352.1,
        topSpeed: 33.7,
      },
    },
    {
      id: 6,
      name: 'Orkun Kökçü',
      team: 'Feyenoord',
      position: 'Midfielder',
      nationality: 'Turkey',
      age: 22,
      stats: {
        goals: 11,
        assists: 7,
        matches: 33,
        rating: 7.9,
        passAccuracy: 88,
        shotsOnTarget: 45,
        distanceCovered: 338.6,
        topSpeed: 33.2,
      },
    },
    {
      id: 7,
      name: 'Lutsharel Geertruida',
      team: 'Feyenoord',
      position: 'Defender',
      nationality: 'Netherlands',
      age: 23,
      stats: {
        goals: 3,
        assists: 5,
        matches: 34,
        rating: 7.6,
        passAccuracy: 90,
        shotsOnTarget: 22,
        distanceCovered: 347.9,
        topSpeed: 35.8,
      },
    },
    {
      id: 8,
      name: 'Lars Unnerstall',
      team: 'FC Twente',
      position: 'Goalkeeper',
      nationality: 'Germany',
      age: 32,
      stats: {
        goals: 0,
        assists: 0,
        matches: 34,
        rating: 7.7,
        passAccuracy: 82,
        shotsOnTarget: 0,
        distanceCovered: 187.2,
        topSpeed: 29.8,
      },
    },
  ];

  /**
   * Formats a player's position into a standardized short form and color code
   *
   * @param position - The player's position (e.g., "Defender", "Midfielder")
   * @returns An object containing the short form and CSS color class
   */
  const getPositionDetails = (position: string): { short: string; colorClass: string } => {
    switch (position) {
      case 'Goalkeeper':
        return { short: 'GK', colorClass: 'bg-yellow-100 text-yellow-800' };
      case 'Defender':
        return { short: 'DEF', colorClass: 'bg-blue-100 text-blue-800' };
      case 'Midfielder':
        return { short: 'MID', colorClass: 'bg-green-100 text-green-800' };
      case 'Winger':
        return { short: 'WIN', colorClass: 'bg-purple-100 text-purple-800' };
      case 'Forward':
        return { short: 'FWD', colorClass: 'bg-red-100 text-red-800' };
      default:
        return { short: '---', colorClass: 'bg-gray-100 text-gray-800' };
    }
  };

  return (
    <div className="bg-gray-50 min-h-screen">
      {/* Page Header */}
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="md:flex md:items-center md:justify-between">
            <div className="flex-1 min-w-0">
              <h1 className="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
                Players
              </h1>
            </div>
            <div className="mt-4 flex flex-wrap space-x-3 md:mt-0 md:ml-4">
              <select className="mb-2 md:mb-0 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                <option value="all">All Teams</option>
                <option value="ajax">Ajax</option>
                <option value="psv">PSV</option>
                <option value="feyenoord">Feyenoord</option>
                <option value="fc-twente">FC Twente</option>
              </select>
              <select className="mb-2 md:mb-0 inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                <option value="all">All Positions</option>
                <option value="goalkeeper">Goalkeeper</option>
                <option value="defender">Defender</option>
                <option value="midfielder">Midfielder</option>
                <option value="winger">Winger</option>
                <option value="forward">Forward</option>
              </select>
              <div className="relative">
                <input
                  type="text"
                  placeholder="Search players..."
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                />
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
                    />
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Player Comparison Feature */}
        <div className="bg-white shadow-sm rounded-lg mb-8 overflow-hidden">
          <div className="px-4 py-5 sm:p-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">Compare Players</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              {/* Player 1 Selection */}
              <div>
                <label htmlFor="player1" className="block text-sm font-medium text-gray-700 mb-1">
                  Player 1
                </label>
                <select
                  id="player1"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select player...</option>
                  {playerData.map((player) => (
                    <option key={`p1-${player.id}`} value={player.id}>
                      {player.name} ({player.team})
                    </option>
                  ))}
                </select>
              </div>

              {/* Player 2 Selection */}
              <div>
                <label htmlFor="player2" className="block text-sm font-medium text-gray-700 mb-1">
                  Player 2
                </label>
                <select
                  id="player2"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm text-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select player...</option>
                  {playerData.map((player) => (
                    <option key={`p2-${player.id}`} value={player.id}>
                      {player.name} ({player.team})
                    </option>
                  ))}
                </select>
              </div>

              {/* Compare Button */}
              <div className="flex items-end">
                <button
                  type="button"
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Compare Statistics
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Player Leaderboards */}
        <div className="mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-6">Performance Leaderboards</h2>
          <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-4">
            {/* Goals Leaderboard */}
            <div className="bg-white shadow-sm rounded-lg overflow-hidden">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Goals</h3>
                <ol className="space-y-3">
                  {playerData
                    .sort((a, b) => b.stats.goals - a.stats.goals)
                    .slice(0, 5)
                    .map((player, idx) => (
                      <li key={player.id} className="flex items-center">
                        <span className="font-bold text-gray-600 w-5">{idx + 1}.</span>
                        <div className="flex-1 ml-2">
                          <p className="text-sm font-medium text-gray-900">{player.name}</p>
                          <p className="text-xs text-gray-500">{player.team}</p>
                        </div>
                        <span className="font-semibold text-blue-600">{player.stats.goals}</span>
                      </li>
                    ))}
                </ol>
                <div className="mt-4 text-right">
                  <button className="text-sm font-medium text-blue-600 hover:text-blue-800">
                    View Full Rankings
                  </button>
                </div>
              </div>
            </div>

            {/* Assists Leaderboard */}
            <div className="bg-white shadow-sm rounded-lg overflow-hidden">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Assists</h3>
                <ol className="space-y-3">
                  {playerData
                    .sort((a, b) => b.stats.assists - a.stats.assists)
                    .slice(0, 5)
                    .map((player, idx) => (
                      <li key={player.id} className="flex items-center">
                        <span className="font-bold text-gray-600 w-5">{idx + 1}.</span>
                        <div className="flex-1 ml-2">
                          <p className="text-sm font-medium text-gray-900">{player.name}</p>
                          <p className="text-xs text-gray-500">{player.team}</p>
                        </div>
                        <span className="font-semibold text-blue-600">{player.stats.assists}</span>
                      </li>
                    ))}
                </ol>
                <div className="mt-4 text-right">
                  <button className="text-sm font-medium text-blue-600 hover:text-blue-800">
                    View Full Rankings
                  </button>
                </div>
              </div>
            </div>

            {/* Rating Leaderboard */}
            <div className="bg-white shadow-sm rounded-lg overflow-hidden">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Player Rating</h3>
                <ol className="space-y-3">
                  {playerData
                    .sort((a, b) => b.stats.rating - a.stats.rating)
                    .slice(0, 5)
                    .map((player, idx) => (
                      <li key={player.id} className="flex items-center">
                        <span className="font-bold text-gray-600 w-5">{idx + 1}.</span>
                        <div className="flex-1 ml-2">
                          <p className="text-sm font-medium text-gray-900">{player.name}</p>
                          <p className="text-xs text-gray-500">{player.team}</p>
                        </div>
                        <span className="font-semibold text-blue-600">
                          {player.stats.rating.toFixed(1)}
                        </span>
                      </li>
                    ))}
                </ol>
                <div className="mt-4 text-right">
                  <button className="text-sm font-medium text-blue-600 hover:text-blue-800">
                    View Full Rankings
                  </button>
                </div>
              </div>
            </div>

            {/* Top Speed Leaderboard */}
            <div className="bg-white shadow-sm rounded-lg overflow-hidden">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Top Speed (km/h)</h3>
                <ol className="space-y-3">
                  {playerData
                    .sort((a, b) => b.stats.topSpeed - a.stats.topSpeed)
                    .slice(0, 5)
                    .map((player, idx) => (
                      <li key={player.id} className="flex items-center">
                        <span className="font-bold text-gray-600 w-5">{idx + 1}.</span>
                        <div className="flex-1 ml-2">
                          <p className="text-sm font-medium text-gray-900">{player.name}</p>
                          <p className="text-xs text-gray-500">{player.team}</p>
                        </div>
                        <span className="font-semibold text-blue-600">
                          {player.stats.topSpeed.toFixed(1)}
                        </span>
                      </li>
                    ))}
                </ol>
                <div className="mt-4 text-right">
                  <button className="text-sm font-medium text-blue-600 hover:text-blue-800">
                    View Full Rankings
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Player Listing Table */}
        <div className="mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-6">All Players</h2>

          {/* Table */}
          <div className="bg-white shadow-sm overflow-hidden rounded-lg">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Player
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Position
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Age
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Matches
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Goals
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Assists
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Rating
                    </th>
                    <th
                      scope="col"
                      className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider"
                    >
                      Actions
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {playerData.map((player) => {
                    const position = getPositionDetails(player.position);

                    return (
                      <tr key={player.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4 whitespace-nowrap">
                          <div className="flex items-center">
                            <div className="flex-shrink-0 h-10 w-10 bg-gray-200 rounded-full flex items-center justify-center">
                              <span className="text-xs font-medium">
                                {player.name
                                  .split(' ')
                                  .map((n) => n[0])
                                  .join('')}
                              </span>
                            </div>
                            <div className="ml-4">
                              <div className="text-sm font-medium text-gray-900">
                                <Link
                                  href={`/players/${player.id}`}
                                  className="hover:text-blue-600"
                                >
                                  {player.name}
                                </Link>
                              </div>
                              <div className="text-sm text-gray-500">{player.team}</div>
                            </div>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <span
                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${position.colorClass}`}
                          >
                            {position.short}
                          </span>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center text-sm text-gray-500">
                          {player.age}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center text-sm text-gray-500">
                          {player.stats.matches}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center text-sm text-gray-900 font-medium">
                          {player.stats.goals}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center text-sm text-gray-900 font-medium">
                          {player.stats.assists}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center">
                          <div className="flex items-center justify-center">
                            <span
                              className={`inline-flex items-center justify-center h-8 w-8 rounded-full text-sm font-medium ${
                                player.stats.rating >= 8
                                  ? 'bg-green-100 text-green-800'
                                  : player.stats.rating >= 7
                                    ? 'bg-blue-100 text-blue-800'
                                    : 'bg-gray-100 text-gray-800'
                              }`}
                            >
                              {player.stats.rating.toFixed(1)}
                            </span>
                          </div>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-center text-sm">
                          <Link
                            href={`/players/${player.id}`}
                            className="text-blue-600 hover:text-blue-900 mr-3"
                          >
                            View Profile
                          </Link>
                          <Link
                            href={`/analytics?player=${player.id}`}
                            className="text-blue-600 hover:text-blue-900"
                          >
                            Analytics
                          </Link>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            <div className="bg-white px-4 py-3 border-t border-gray-200 sm:px-6">
              <div className="flex items-center justify-between">
                <div className="flex-1 flex justify-between sm:hidden">
                  <a
                    href="#"
                    className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
                  >
                    Previous
                  </a>
                  <a
                    href="#"
                    className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
                  >
                    Next
                  </a>
                </div>
                <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                  <div>
                    <p className="text-sm text-gray-700">
                      Showing <span className="font-medium">1</span> to{' '}
                      <span className="font-medium">8</span> of{' '}
                      <span className="font-medium">300</span> results
                    </p>
                  </div>
                  <div>
                    <nav
                      className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px"
                      aria-label="Pagination"
                    >
                      <a
                        href="#"
                        className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
                      >
                        <span className="sr-only">Previous</span>
                        <svg
                          className="h-5 w-5"
                          xmlns="http://www.w3.org/2000/svg"
                          viewBox="0 0 20 20"
                          fill="currentColor"
                          aria-hidden="true"
                        >
                          <path
                            fillRule="evenodd"
                            d="M12.707 5.293a1 1 0 010 1.414L9.414 10l3.293 3.293a1 1 0 01-1.414 1.414l-4-4a1 1 0 010-1.414l4-4a1 1 0 011.414 0z"
                            clipRule="evenodd"
                          />
                        </svg>
                      </a>
                      <a
                        href="#"
                        aria-current="page"
                        className="z-10 bg-blue-50 border-blue-500 text-blue-600 relative inline-flex items-center px-4 py-2 border text-sm font-medium"
                      >
                        1
                      </a>
                      <a
                        href="#"
                        className="bg-white border-gray-300 text-gray-500 hover:bg-gray-50 relative inline-flex items-center px-4 py-2 border text-sm font-medium"
                      >
                        2
                      </a>
                      <a
                        href="#"
                        className="bg-white border-gray-300 text-gray-500 hover:bg-gray-50 relative inline-flex items-center px-4 py-2 border text-sm font-medium"
                      >
                        3
                      </a>
                      <span className="relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium text-gray-700">
                        ...
                      </span>
                      <a
                        href="#"
                        className="bg-white border-gray-300 text-gray-500 hover:bg-gray-50 relative inline-flex items-center px-4 py-2 border text-sm font-medium"
                      >
                        38
                      </a>
                      <a
                        href="#"
                        className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
                      >
                        <span className="sr-only">Next</span>
                        <svg
                          className="h-5 w-5"
                          xmlns="http://www.w3.org/2000/svg"
                          viewBox="0 0 20 20"
                          fill="currentColor"
                          aria-hidden="true"
                        >
                          <path
                            fillRule="evenodd"
                            d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z"
                            clipRule="evenodd"
                          />
                        </svg>
                      </a>
                    </nav>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
