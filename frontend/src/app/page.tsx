import Image from "next/image";
import Link from "next/link";

/**
 * HomePage is the entry point for the AIFAA football dashboard.
 * Displays a welcome screen and provides navigation to main features.
 *
 * @returns The home page component
 */
export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col">
      {/* Hero Section */}
      <section className="relative flex flex-col items-center justify-center px-6 py-32 bg-gradient-to-br from-blue-900 via-blue-800 to-indigo-900 text-white">
        <div className="absolute inset-0 overflow-hidden opacity-20">
          <div className="absolute inset-y-0 right-0 w-1/2 bg-gradient-to-l from-indigo-500 to-transparent transform rotate-12" />
          <div className="absolute inset-y-0 left-0 w-1/2 bg-gradient-to-r from-blue-500 to-transparent -rotate-12" />
        </div>

        <div className="relative z-10 max-w-4xl mx-auto text-center">
          <h1 className="text-5xl font-bold tracking-tight sm:text-6xl mb-6">
            AIFAA Football Analytics
          </h1>
          <p className="text-xl text-blue-100 mb-10 max-w-2xl mx-auto">
            Advanced football tracking data visualization and analysis platform.
            Gain tactical insights and performance metrics powered by
            cutting-edge technology.
          </p>

          <div className="flex flex-wrap justify-center gap-4 mt-8">
            <Link
              href="/dashboard"
              className="bg-white text-blue-900 hover:bg-blue-50 px-6 py-3 rounded-lg font-medium shadow-lg transition duration-300 focus:outline-none focus:ring-2 focus:ring-blue-300"
            >
              Open Dashboard
            </Link>
            <Link
              href="/upload"
              className="bg-transparent border-2 border-white text-white hover:bg-white/10 px-6 py-3 rounded-lg font-medium transition duration-300 focus:outline-none focus:ring-2 focus:ring-blue-300"
            >
              Upload Match Data
            </Link>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 px-6 bg-gray-50">
        <div className="max-w-7xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12 text-gray-800">
            Powerful Analytics Tools
          </h2>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
            {/* Feature 1 */}
            <div className="bg-white rounded-xl shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300">
              <div className="p-8">
                <div className="flex items-center justify-center h-12 w-12 rounded-md bg-blue-800 text-white mb-5">
                  <Image
                    src="/globe.svg"
                    alt="Performance Analytics"
                    width={24}
                    height={24}
                  />
                </div>
                <h3 className="text-xl font-semibold mb-3 text-gray-800">
                  Performance Analytics
                </h3>
                <p className="text-gray-600">
                  Track player movements, heatmaps, and physical metrics with
                  precise data-driven insights.
                </p>
              </div>
            </div>

            {/* Feature 2 */}
            <div className="bg-white rounded-xl shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300">
              <div className="p-8">
                <div className="flex items-center justify-center h-12 w-12 rounded-md bg-blue-800 text-white mb-5">
                  <Image
                    src="/window.svg"
                    alt="Tactical Analysis"
                    width={24}
                    height={24}
                  />
                </div>
                <h3 className="text-xl font-semibold mb-3 text-gray-800">
                  Tactical Analysis
                </h3>
                <p className="text-gray-600">
                  Visualize team formations, passing networks, and pressing
                  patterns in real-time.
                </p>
              </div>
            </div>

            {/* Feature 3 */}
            <div className="bg-white rounded-xl shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300">
              <div className="p-8">
                <div className="flex items-center justify-center h-12 w-12 rounded-md bg-blue-800 text-white mb-5">
                  <Image
                    src="/file.svg"
                    alt="Video Integration"
                    width={24}
                    height={24}
                  />
                </div>
                <h3 className="text-xl font-semibold mb-3 text-gray-800">
                  Video Integration
                </h3>
                <p className="text-gray-600">
                  Synchronize video footage with tracking data for comprehensive
                  match analysis.
                </p>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Recent Matches Section */}
      <section className="py-20 px-6 bg-white">
        <div className="max-w-7xl mx-auto">
          <h2 className="text-3xl font-bold text-center mb-12 text-gray-800">
            Recent Matches
          </h2>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* This would be populated from an API call in a real application */}
            {["Ajax vs PSV", "Feyenoord vs AZ", "FC Utrecht vs FC Twente"].map(
              (match, index) => (
                <div
                  key={index}
                  className="border border-gray-200 rounded-lg overflow-hidden hover:border-blue-500 transition-colors duration-300 cursor-pointer"
                >
                  <div className="aspect-video bg-gray-100 relative">
                    <div className="absolute inset-0 flex items-center justify-center">
                      <p className="text-gray-400 text-sm">Match Preview</p>
                    </div>
                  </div>
                  <div className="p-4">
                    <h3 className="font-semibold text-lg text-gray-800">
                      {match}
                    </h3>
                    <p className="text-gray-500 text-sm mt-1">
                      Eredivisie • May 15, 2023
                    </p>
                  </div>
                </div>
              )
            )}
          </div>

          <div className="text-center mt-10">
            <Link
              href="/matches"
              className="text-blue-600 hover:text-blue-800 font-medium"
            >
              View All Matches →
            </Link>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-16 px-6 bg-gradient-to-r from-blue-800 to-indigo-900 text-white">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-3xl font-bold mb-6">
            Ready to elevate your football analysis?
          </h2>
          <p className="text-lg mb-8 text-blue-100">
            Get started with AIFAA's advanced football analytics platform today.
          </p>
          <Link
            href="/dashboard"
            className="bg-white text-blue-900 hover:bg-blue-50 px-6 py-3 rounded-lg font-medium shadow-lg inline-block transition duration-300"
          >
            Explore the Dashboard
          </Link>
        </div>
      </section>
    </main>
  );
}
