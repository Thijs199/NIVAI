import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Link from "next/link";
import { ReactNode } from "react";

// Initialize Inter font with latin subset
const inter = Inter({ subsets: ["latin"] });

/**
 * Metadata configuration for the application
 * Provides SEO information and browser configuration
 */
export const metadata: Metadata = {
  title: "NIVAI - Football Analytics Platform",
  description:
    "Advanced football tracking data visualization and analysis platform",
  applicationName: "NIVAI Football Analytics",
};

/**
 * RootLayout defines the overall structure of the application.
 * Wraps all pages with consistent navigation and footer.
 *
 * @param {Object} props - Component properties
 * @param {ReactNode} props.children - Child components to render within the layout
 * @returns Root layout component
 */
export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <div className="flex flex-col min-h-screen">
          {/* Header/Navigation */}
          <header className="bg-white shadow">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="flex justify-between h-16">
                <div className="flex">
                  <div className="flex-shrink-0 flex items-center">
                    <Link href="/" className="text-2xl font-bold text-blue-800">
                      NIVAI
                    </Link>
                  </div>
                  <nav className="ml-10 flex items-center space-x-4">
                    <Link
                      href="/dashboard"
                      className="px-3 py-2 text-sm font-medium text-gray-700 hover:text-blue-800"
                    >
                      Dashboard
                    </Link>
                    <Link
                      href="/matches"
                      className="px-3 py-2 text-sm font-medium text-gray-700 hover:text-blue-800"
                    >
                      Matches
                    </Link>
                    <Link
                      href="/teams"
                      className="px-3 py-2 text-sm font-medium text-gray-700 hover:text-blue-800"
                    >
                      Teams
                    </Link>
                    <Link
                      href="/players"
                      className="px-3 py-2 text-sm font-medium text-gray-700 hover:text-blue-800"
                    >
                      Players
                    </Link>
                    <Link
                      href="/analytics"
                      className="px-3 py-2 text-sm font-medium text-gray-700 hover:text-blue-800"
                    >
                      Analytics
                    </Link>
                  </nav>
                </div>
                <div className="flex items-center">
                  <Link
                    href="/upload"
                    className="ml-6 inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-800 hover:bg-blue-700"
                  >
                    Upload
                  </Link>
                  <div className="ml-4 flex items-center">
                    {/* User profile/avatar would go here */}
                    <button className="p-1 rounded-full text-gray-500 hover:text-blue-800 focus:outline-none">
                      <span className="sr-only">View profile</span>
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                        strokeWidth={1.5}
                        stroke="currentColor"
                        className="w-6 h-6"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          d="M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z"
                        />
                      </svg>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </header>

          {/* Main content */}
          <main className="flex-grow">{children}</main>

          {/* Footer */}
          <footer className="bg-gray-800 text-white">
            <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div>
                  <h3 className="text-lg font-semibold mb-4">
                    NIVAI Football Analytics
                  </h3>
                  <p className="text-gray-300 text-sm">
                    Advanced football tracking data visualization and analysis
                    platform.
                  </p>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-4">Quick Links</h3>
                  <ul className="space-y-2 text-sm">
                    <li>
                      <Link
                        href="/about"
                        className="text-gray-300 hover:text-white"
                      >
                        About Us
                      </Link>
                    </li>
                    <li>
                      <Link
                        href="/contact"
                        className="text-gray-300 hover:text-white"
                      >
                        Contact
                      </Link>
                    </li>
                    <li>
                      <Link
                        href="/support"
                        className="text-gray-300 hover:text-white"
                      >
                        Support
                      </Link>
                    </li>
                  </ul>
                </div>
                <div>
                  <h3 className="text-lg font-semibold mb-4">Legal</h3>
                  <ul className="space-y-2 text-sm">
                    <li>
                      <Link
                        href="/privacy"
                        className="text-gray-300 hover:text-white"
                      >
                        Privacy Policy
                      </Link>
                    </li>
                    <li>
                      <Link
                        href="/terms"
                        className="text-gray-300 hover:text-white"
                      >
                        Terms of Service
                      </Link>
                    </li>
                  </ul>
                </div>
              </div>
              <div className="border-t border-gray-700 mt-8 pt-8 text-center text-sm text-gray-400">
                &copy; {new Date().getFullYear()} NIVAI. All rights reserved.
              </div>
            </div>
          </footer>
        </div>
      </body>
    </html>
  );
}
