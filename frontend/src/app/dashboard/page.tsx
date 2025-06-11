/**
 * Dashboard is the main hub for football analytics visualizations and metrics.
 * Provides an overview of key insights and access to detailed analysis features.
 *
 * @returns The dashboard page component
 */
export default function Dashboard() {
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
                <dd className="mt-1 text-3xl font-semibold text-gray-900">36.2 km/h</dd>
              </dl>
            </div>
            <div className="bg-gray-50 px-4 py-2">
              <div className="text-sm flex justify-between">
                <span className="font-medium text-blue-700">View player</span>
                <span className="text-gray-600">J. Timber - Ajax</span>
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
                Player Performance
              </h3>
              <div className="space-y-4">
                {/* Player 1 */}
                <div className="flex items-center">
                  <div className="flex-shrink-0 h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                    <span className="text-xs font-medium">AM</span>
                  </div>
                  <div className="ml-4 flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-gray-900">Antony Matheus</h4>
                      <span className="text-sm text-gray-500">Ajax</span>
                    </div>
                    <div className="mt-1 w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '87%' }}></div>
                    </div>
                  </div>
                  <div className="ml-2">
                    <span className="text-sm font-medium text-gray-900">8.7</span>
                  </div>
                </div>

                {/* Player 2 */}
                <div className="flex items-center">
                  <div className="flex-shrink-0 h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                    <span className="text-xs font-medium">JT</span>
                  </div>
                  <div className="ml-4 flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-gray-900">Jurriën Timber</h4>
                      <span className="text-sm text-gray-500">Ajax</span>
                    </div>
                    <div className="mt-1 w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '82%' }}></div>
                    </div>
                  </div>
                  <div className="ml-2">
                    <span className="text-sm font-medium text-gray-900">8.2</span>
                  </div>
                </div>

                {/* Player 3 */}
                <div className="flex items-center">
                  <div className="flex-shrink-0 h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                    <span className="text-xs font-medium">CV</span>
                  </div>
                  <div className="ml-4 flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-gray-900">Cody Vipko</h4>
                      <span className="text-sm text-gray-500">PSV</span>
                    </div>
                    <div className="mt-1 w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '78%' }}></div>
                    </div>
                  </div>
                  <div className="ml-2">
                    <span className="text-sm font-medium text-gray-900">7.8</span>
                  </div>
                </div>

                {/* Player 4 */}
                <div className="flex items-center">
                  <div className="flex-shrink-0 h-10 w-10 rounded-full bg-gray-200 flex items-center justify-center">
                    <span className="text-xs font-medium">DK</span>
                  </div>
                  <div className="ml-4 flex-1">
                    <div className="flex items-center justify-between">
                      <h4 className="text-sm font-medium text-gray-900">Daley Klaassen</h4>
                      <span className="text-sm text-gray-500">Ajax</span>
                    </div>
                    <div className="mt-1 w-full bg-gray-200 rounded-full h-2">
                      <div className="bg-blue-600 h-2 rounded-full" style={{ width: '75%' }}></div>
                    </div>
                  </div>
                  <div className="ml-2">
                    <span className="text-sm font-medium text-gray-900">7.5</span>
                  </div>
                </div>
              </div>
              <div className="mt-4 flex justify-end">
                <button className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-blue-700 hover:bg-blue-50">
                  View All Players
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Match Analysis & Video Highlights Row */}
        <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
          {/* Recent Matches */}
          <div className="bg-white overflow-hidden shadow rounded-lg lg:col-span-2">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-3">Recent Matches</h3>
              <div className="space-y-3">
                {/* Match 1 */}
                <div className="border border-gray-200 rounded-md p-4 hover:border-blue-500 transition-colors duration-200 cursor-pointer">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <span className="text-sm font-medium">May 14</span>
                      <span className="text-gray-500">|</span>
                      <span className="font-medium">Ajax 3 - 1 PSV</span>
                    </div>
                    <div>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        Complete
                      </span>
                    </div>
                  </div>
                  <div className="mt-2 flex items-center justify-between">
                    <div className="text-sm text-gray-500">
                      Possession: 58% | Shots: 12 | Passes: 472
                    </div>
                    <button className="inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
                      View Analysis
                    </button>
                  </div>
                </div>

                {/* Match 2 */}
                <div className="border border-gray-200 rounded-md p-4 hover:border-blue-500 transition-colors duration-200 cursor-pointer">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <span className="text-sm font-medium">May 7</span>
                      <span className="text-gray-500">|</span>
                      <span className="font-medium">Feyenoord 2 - 0 AZ</span>
                    </div>
                    <div>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                        Complete
                      </span>
                    </div>
                  </div>
                  <div className="mt-2 flex items-center justify-between">
                    <div className="text-sm text-gray-500">
                      Possession: 62% | Shots: 14 | Passes: 508
                    </div>
                    <button className="inline-flex items-center text-sm font-medium text-blue-600 hover:text-blue-500">
                      View Analysis
                    </button>
                  </div>
                </div>

                {/* Match 3 */}
                <div className="border border-gray-200 rounded-md p-4 hover:border-blue-500 transition-colors duration-200 cursor-pointer">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4">
                      <span className="text-sm font-medium">May 1</span>
                      <span className="text-gray-500">|</span>
                      <span className="font-medium">FC Utrecht 1 - 3 FC Twente</span>
                    </div>
                    <div>
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
                        Processing
                      </span>
                    </div>
                  </div>
                  <div className="mt-2 flex items-center justify-between">
                    <div className="text-sm text-gray-500">
                      Possession: 45% | Shots: 8 | Passes: 346
                    </div>
                    <button className="inline-flex items-center text-sm font-medium text-gray-400 cursor-not-allowed">
                      Analysis Pending
                    </button>
                  </div>
                </div>
              </div>
              <div className="mt-4 flex justify-end">
                <button className="inline-flex items-center px-3 py-1.5 border border-transparent text-sm font-medium rounded-md text-blue-700 hover:bg-blue-50">
                  View All Matches
                </button>
              </div>
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
