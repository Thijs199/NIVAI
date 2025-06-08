"use client";

import React, { useState } from 'react';

/**
 * AnalyticsPage provides comprehensive football analytics and visualizations.
 * Offers detailed performance metrics, match statistics, and video analysis tools.
 * 
 * @returns The Analytics page component with interactive data visualizations
 */
export default function AnalyticsPage() {
  const [activeTab, setActiveTab] = useState('performance');
  const [selectedTimeframe, setSelectedTimeframe] = useState('season');
  
  return (
    <div className="container mx-auto p-4">
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Analytics Dashboard</h1>
        <div className="flex items-center space-x-3 mt-4 md:mt-0">
          <select 
            className="bg-white border border-gray-300 rounded-md py-2 px-4 text-sm"
            value={selectedTimeframe}
            onChange={(e) => setSelectedTimeframe(e.target.value)}
          >
            <option value="match">Last Match</option>
            <option value="month">Last Month</option>
            <option value="season">Current Season</option>
            <option value="alltime">All Time</option>
          </select>
          <button className="bg-blue-600 text-white py-2 px-4 rounded-md text-sm hover:bg-blue-700 transition-colors">
            Export Data
          </button>
        </div>
      </div>
      
      {/* Analytics Navigation Tabs */}
      <div className="mb-6 border-b border-gray-200">
        <nav className="flex space-x-8">
          <button 
            onClick={() => setActiveTab('performance')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'performance' 
                ? 'border-blue-500 text-blue-600' 
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Performance Metrics
          </button>
          <button 
            onClick={() => setActiveTab('statistics')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'statistics' 
                ? 'border-blue-500 text-blue-600' 
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Match Statistics
          </button>
          <button 
            onClick={() => setActiveTab('video')}
            className={`py-4 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'video' 
                ? 'border-blue-500 text-blue-600' 
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Video Analysis
          </button>
        </nav>
      </div>
      
      {/* Performance Metrics Content */}
      {activeTab === 'performance' && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-3">Team Performance</h2>
            <div className="space-y-4">
              <div>
                <div className="flex justify-between items-center mb-1">
                  <span className="text-sm font-medium text-gray-700">Possession</span>
                  <span className="text-sm font-medium text-gray-700">58%</                </div>
                <div className="w-full bg-gray-200 rounded-full h-2.5">
                  <div className="bg-blue-600 h-2.5 rounded-full" style={{width: '58%'}}></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-1">
                  <span className="text-sm font-medium text-gray-700">Passing Accuracy</span>
                  <span className="text-sm font-medium text-gray-700">82%</                </div>
                <div className="w-full bg-gray-200 rounded-full h-2.5">
                  <div className="bg-blue-600 h-2.5 rounded-full" style={{width: '82%'}}></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-1">
                  <span className="text-sm font-medium text-gray-700">Shots on Target</span>
                  <span className="text-sm font-medium text-gray-700">65%</                </div>
                <div className="w-full bg-gray-200 rounded-full h-2.5">
                  <div className="bg-blue-600 h-2.5 rounded-full" style={{width: '65%'}}></div>
                </div>
              </div>
              <div>
                <div className="flex justify-between items-center mb-1">
                  <span className="text-sm font-medium text-gray-700">Defensive Actions</span>
                  <span className="text-sm font-medium text-gray-700">74%</                </div>
                <div className="w-full bg-gray-200 rounded-full h-2.5">
                  <div className="bg-blue-600 h-2.5 rounded-full" style={{width: '74%'}}></div>
                </div>
              </div>
            </div>
            <div className="mt-4 pt-4 border-t">
              <button className="text-blue-600 text-sm font-medium hover:text-blue-800">
                View Detailed Performance Report →
              </button>
            </div>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-3">Player Heatmaps</h2>
            <div className="aspect-[4/3] bg-gray-100 rounded relative mb-3">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center">
                  <p className="text-gray-500 text-sm">Interactive player heatmap</p>
                  <p className="text-xs text-gray-400 mt-1">(visualization would render here)</p>
                </div>
              </div>
            </div>
            <div className="flex space-x-2 mb-2">
              <div className="flex-1">
                <select className="w-full text-sm border border-gray-300 rounded p-2">
                  <option>Select Player</option>
                  <option>Antony Matheus</option>
                  <option>Jurriën Timber</option>
                  <option>Cody Vipko</option>
                  <option>Daley Klaassen</option>
                </select>
              </div>
              <div className="flex-1">
                <select className="w-full text-sm border border-gray-300 rounded p-2">
                  <option>All Matches</option>
                  <option>Ajax vs PSV</option>
                  <option>Feyenoord vs AZ</option>
                  <option>FC Utrecht vs FC Twente</option>
                </select>
              </div>
            </div>
            <p className="text-xs text-gray-500">Select a player and match to view movement patterns and positioning analysis</p>
          </div>
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-3">Physical Metrics</h2>
            <div className="space-y-3">
              <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                <div>
                  <p className="font-medium">Total Distance</p>
                  <p className="text-gray-500 text-sm">Team Average</p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-bold">112.4 km</p>
                  <p className="text-xs text-green-600">↑ 3.2% vs Previous</p>
                </div>
              </div>
              <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                <div>
                  <p className="font-medium">Sprint Distance</p>
                  <p className="text-gray-500 text-sm">Team Average</p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-bold">8.7 km</p>
                  <p className="text-xs text-green-600">↑ 1.5% vs Previous</p>
                </div>
              </div>
              <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                <div>
                  <p className="font-medium">High-Intensity Efforts</p>
                  <p className="text-gray-500 text-sm">Team Average</p>
                </div>
                <div className="text-right">
                  <p className="text-xl font-bold">184</p>
                  <p className="text-xs text-red-600">↓ 2.1% vs Previous</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
      
      {/* Match Statistics Content */}
      {activeTab === 'statistics' && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="bg-white rounded-lg shadow p-6 lg:col-span-2">
            <h2 className="text-xl font-semibold mb-4">Match Comparison</h2>
            <div className="flex justify-between items-center mb-6">
              <div className="text-center">
                <div className="w-16 h-16 bg-gray-200 rounded-full mx-auto mb-2"></div>
                <h3 className="font-medium">Ajax</h3>
              </div>
              <div className="text-center text-xl font-bold">3 - 1</div>
              <div className="text-center">
                <div className="w-16 h-16 bg-gray-200 rounded-full mx-auto mb-2"></div>
                <h3 className="font-medium">PSV</h3>
              </div>
            </div>
            
            <div className="space-y-4">
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>58%</span>
                  <span className="font-medium">Possession</span>
                  <span>42%</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '58%'}}></div>
                  <div className="bg-red-600" style={{width: '42%'}}></div>
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>14</span>
                  <span className="font-medium">Shots</span>
                  <span>9</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '60%'}}></div>
                  <div className="bg-red-600" style={{width: '40%'}}></div>
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>6</span>
                  <span className="font-medium">Shots on Target</span>
                  <span>3</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '67%'}}></div>
                  <div className="bg-red-600" style={{width: '33%'}}></div>
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>532</span>
                  <span className="font-medium">Passes</span>
                  <span>418</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '56%'}}></div>
                  <div className="bg-red-600" style={{width: '44%'}}></div>
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>87%</span>
                  <span className="font-medium">Pass Accuracy</span>
                  <span>82%</                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '51.5%'}}></div>
                  <div className="bg-red-600" style={{width: '48.5%'}}></div>
                </div>
              </div>
              
              <div>
                <div className="flex justify-between text-sm text-gray-600 mb-1">
                  <span>8</span>
                  <span className="font-medium">Corners</span>
                  <span>5</span>
                </div>
                <div className="flex h-2 rounded-full overflow-hidden bg-gray-200">
                  <div className="bg-blue-600" style={{width: '61.5%'}}></div>
                  <div className="bg-red-600" style={{width: '38.5%'}}></div>
                </div>
              </div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-3">Key Events</h2>
            <div className="space-y-4">
              <div className="relative pl-6 pb-4 border-l-2 border-blue-500">
                <span className="absolute left-[-8px] top-0 w-4 h-4 rounded-full bg-blue-500"></span>
                <p className="font-medium">Goal - Ajax</p>
                <p className="text-gray-500 text-sm">Antony Matheus (23')</p>
                <p className="text-xs text-gray-400 mt-1">Header from center of the box</p>
              </div>
              
              <div className="relative pl-6 pb-4 border-l-2 border-yellow-500">
                <span className="absolute left-[-8px] top-0 w-4 h-4 rounded-full bg-yellow-500"></span>
                <p className="font-medium">Yellow Card - PSV</p>
                <p className="text-gray-500 text-sm">Ibrahim Sangaré (32')</p>
                <p className="text-xs text-gray-400 mt-1">Tactical foul in midfield</p>
              </div>
              
              <div className="relative pl-6 pb-4 border-l-2 border-blue-500">
                <span className="absolute left-[-8px] top-0 w-4 h-4 rounded-full bg-blue-500"></span>
                <p className="font-medium">Goal - Ajax</p>
                <p className="text-gray-500 text-sm">Daley Klaassen (45+2')</p>
                <p className="text-xs text-gray-400 mt-1">Penalty kick, bottom right corner</p>
              </div>
              
              <div className="relative pl-6 pb-4 border-l-2 border-red-500">
                <span className="absolute left-[-8px] top-0 w-4 h-4 rounded-full bg-red-500"></span>
                <p className="font-medium">Goal - PSV</p>
                <p className="text-gray-500 text-sm">Cody Vipko (58')</p>
                <p className="text-xs text-gray-400 mt-1">Left-footed shot from outside the box</p>
              </div>
              
              <div className="relative pl-6 border-l-2 border-blue-500">
                <span className="absolute left-[-8px] top-0 w-4 h-4 rounded-full bg-blue-500"></span>
                <p className="font-medium">Goal - Ajax</p>
                <p className="text-gray-500 text-sm">Jurriën Timber (76')</p>
                <p className="text-xs text-gray-400 mt-1">Header from set piece</p>
              </div>
            </div>
            <div className="mt-4 pt-4 border-t">
              <button className="text-blue-600 text-sm font-medium hover:text-blue-800">
                View Full Timeline →
              </button>
            </div>
          </div>
        </div>
      )}
      
      {/* Video Analysis Content */}
      {activeTab === 'video' && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bg-white rounded-lg shadow p-6 md:col-span-2">
            <h2 className="text-xl font-semibold mb-3">AI Video Analysis</h2>
            <div className="aspect-video bg-gray-100 rounded relative mb-4">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="text-center">
                  <p className="text-gray-500 text-sm">Video player with AI annotations</p>
                  <p className="text-xs text-gray-400 mt-1">(visualization would render here)</p>
                </div>
              </div>
              <div className="absolute bottom-0 left-0 right-0 h-8 bg-black bg-opacity-50 flex items-center px-2">
                <div className="flex items-center justify-between w-full">
                  <button className="text-white">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clipRule="evenodd" />
                    </svg>
                  </button>
                  <div className="flex-1 mx-2">
                    <div className="h-1 bg-white bg-opacity-50 rounded">
                      <div className="h-1 bg-blue-500 rounded" style={{width: '45%'}}></div>
                    </div>
                  </div>
                  <span className="text-white text-xs">2:15 / 5:30</span>
                </div>
              </div>
            </div>
            <div className="flex space-x-2 mb-4">
              <button className="flex-1 bg-blue-600 text-white py-2 rounded-md text-sm hover:bg-blue-700">
                Toggle Tracking Data
              </button>
              <button className="flex-1 bg-blue-600 text-white py-2 rounded-md text-sm hover:bg-blue-700">
                Show Heatmaps
              </button>
              <button className="flex-1 bg-blue-600 text-white py-2 rounded-md text-sm hover:bg-blue-700">
                Player Spotlight
              </button>
            </div>
            <div className="grid grid-cols-5 gap-2">
              <div className="aspect-video bg-gray-200 rounded cursor-pointer"></div>
              <div className="aspect-video bg-gray-200 rounded cursor-pointer"></div>
              <div className="aspect-video bg-gray-200 rounded cursor-pointer"></div>
              <div className="aspect-video bg-gray-200 rounded cursor-pointer"></div>
              <div className="aspect-video bg-gray-200 rounded cursor-pointer"></div>
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-6">
            <h2 className="text-xl font-semibold mb-3">Video Insights</h2>
            <div className="space-y-4">
              <div className="p-3 bg-blue-50 rounded-lg border border-blue-200">
                <h3 className="font-medium text-blue-800">Defensive Organization</h3>
                <p className="text-sm text-gray-600 mt-1">
                  Ajax's compact defensive shape reduced PSV's attacking effectiveness by 32% compared to their season average.
                </p>
              </div>
              
              <div className="p-3 bg-green-50 rounded-lg border border-green-200">
                <h3 className="font-medium text-green-800">Pressing Patterns</h3>
                <p className="text-sm text-gray-600 mt-1">
                  High pressing intensity in opposition half resulted in 12 ball recoveries leading to 3 scoring opportunities.
                </p>
              </div>
              
              <div className="p-3 bg-purple-50 rounded-lg border border-purple-200">
                <h3 className="font-medium text-purple-800">Set Piece Analysis</h3>
                <p className="text-sm text-gray-600 mt-1">
                  Ajax scored from 2 of 8 corner kick opportunities, significantly above the league average conversion rate.
                </p>
              </div>
              
              <div className="p-3 bg-orange-50 rounded-lg border border-orange-200">
                <h3 className="font-medium text-orange-800">Transition Play</h3>
                <p className="text-sm text-gray-600 mt-1">
                  Counter-attacking speed averaged 3.2 seconds from ball recovery to shot - fastest in the league this season.
                </p>
              </div>
            </div>
            
            <div className="mt-4 pt-4 border-t">
              <button className="text-blue-600 text-sm font-medium hover:text-blue-800">
                Generate Full Analysis Report →
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
