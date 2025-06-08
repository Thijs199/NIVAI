/**
 * UploadPage allows users to upload match videos and tracking data.
 * Provides form controls for video upload, metadata entry, and upload status tracking.
 *
 * @returns The upload page component
 */
export default function UploadPage() {
  return (
    <div className="bg-gray-50 min-h-screen py-8">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="bg-white shadow overflow-hidden rounded-lg">
          <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
            <h1 className="text-xl font-semibold text-gray-900">
              Upload Match Data
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Upload match videos and tracking data for analysis.
            </p>
          </div>

          {/* Upload Form */}
          <form className="px-4 py-5 sm:p-6 space-y-6">
            {/* Video File Upload */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Video File
              </label>
              <div className="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 border-dashed rounded-md">
                <div className="space-y-1 text-center">
                  <svg
                    className="mx-auto h-12 w-12 text-gray-400"
                    stroke="currentColor"
                    fill="none"
                    viewBox="0 0 48 48"
                    aria-hidden="true"
                  >
                    <path
                      d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
                      strokeWidth={2}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <div className="flex text-sm text-gray-600">
                    <label
                      htmlFor="file-upload"
                      className="relative cursor-pointer bg-white rounded-md font-medium text-blue-600 hover:text-blue-500 focus-within:outline-none"
                    >
                      <span>Upload a video file</span>
                      <input
                        id="file-upload"
                        name="file-upload"
                        type="file"
                        className="sr-only"
                        accept="video/*"
                      />
                    </label>
                    <p className="pl-1">or drag and drop</p>
                  </div>
                  <p className="text-xs text-gray-500">
                    MP4, MOV, or AVI up to 2GB
                  </p>
                </div>
              </div>
            </div>

            {/* Tracking Data Upload */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Tracking Data (Optional)
              </label>
              <div className="mt-1 flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 border-dashed rounded-md">
                <div className="space-y-1 text-center">
                  <svg
                    className="mx-auto h-12 w-12 text-gray-400"
                    stroke="currentColor"
                    fill="none"
                    viewBox="0 0 48 48"
                    aria-hidden="true"
                  >
                    <path
                      d="M8 14v20c0 4 4 4 4 4h24c4 0 4-4 4-4V14m-36 0c0-4 4-4 4-4h24c4 0 4 4 4 4m-36 0l5-6m26 6l-5-6m-21 6h20"
                      strokeWidth={2}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                  <div className="flex text-sm text-gray-600">
                    <label
                      htmlFor="tracking-data-upload"
                      className="relative cursor-pointer bg-white rounded-md font-medium text-blue-600 hover:text-blue-500 focus-within:outline-none"
                    >
                      <span>Upload tracking data</span>
                      <input
                        id="tracking-data-upload"
                        name="tracking-data-upload"
                        type="file"
                        className="sr-only"
                        accept=".csv,.xml,.json"
                      />
                    </label>
                    <p className="pl-1">or drag and drop</p>
                  </div>
                  <p className="text-xs text-gray-500">
                    CSV, XML, or JSON files
                  </p>
                </div>
              </div>
            </div>

            {/* Video Metadata */}
            <div className="space-y-4">
              <div className="grid grid-cols-1 gap-y-6 gap-x-4 sm:grid-cols-6">
                <div className="sm:col-span-3">
                  <label
                    htmlFor="title"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Video Title
                  </label>
                  <div className="mt-1">
                    <input
                      type="text"
                      name="title"
                      id="title"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      placeholder="Ajax vs PSV - May 2023"
                    />
                  </div>
                </div>

                <div className="sm:col-span-3">
                  <label
                    htmlFor="match-id"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Match ID
                  </label>
                  <div className="mt-1">
                    <input
                      type="text"
                      name="match-id"
                      id="match-id"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      placeholder="e.g., KNVB-2023-0542"
                    />
                  </div>
                </div>

                <div className="sm:col-span-3">
                  <label
                    htmlFor="home-team"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Home Team
                  </label>
                  <div className="mt-1">
                    <input
                      type="text"
                      name="home-team"
                      id="home-team"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      placeholder="Ajax"
                    />
                  </div>
                </div>

                <div className="sm:col-span-3">
                  <label
                    htmlFor="away-team"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Away Team
                  </label>
                  <div className="mt-1">
                    <input
                      type="text"
                      name="away-team"
                      id="away-team"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      placeholder="PSV"
                    />
                  </div>
                </div>

                <div className="sm:col-span-3">
                  <label
                    htmlFor="competition"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Competition
                  </label>
                  <div className="mt-1">
                    <select
                      id="competition"
                      name="competition"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    >
                      <option value="">Select Competition</option>
                      <option value="eredivisie">Eredivisie</option>
                      <option value="knvb-cup">KNVB Cup</option>
                      <option value="champions-league">Champions League</option>
                      <option value="europa-league">Europa League</option>
                      <option value="friendly">Friendly</option>
                    </select>
                  </div>
                </div>

                <div className="sm:col-span-3">
                  <label
                    htmlFor="match-date"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Match Date
                  </label>
                  <div className="mt-1">
                    <input
                      type="date"
                      name="match-date"
                      id="match-date"
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                    />
                  </div>
                </div>

                <div className="sm:col-span-6">
                  <label
                    htmlFor="description"
                    className="block text-sm font-medium text-gray-700"
                  >
                    Description
                  </label>
                  <div className="mt-1">
                    <textarea
                      id="description"
                      name="description"
                      rows={3}
                      className="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
                      placeholder="Add any additional information about this match..."
                    />
                  </div>
                </div>
              </div>
            </div>

            <div className="flex justify-end space-x-3">
              <button
                type="button"
                className="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Cancel
              </button>
              <button
                type="submit"
                className="bg-blue-600 py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Upload
              </button>
            </div>
          </form>
        </div>

        {/* Recent Uploads */}
        <div className="mt-8 bg-white shadow overflow-hidden rounded-lg">
          <div className="px-4 py-5 sm:px-6 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">
              Recent Uploads
            </h2>
          </div>
          <div className="divide-y divide-gray-200">
            {/* Upload Item 1 */}
            <div className="px-4 py-4 sm:px-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-blue-600 truncate">
                    Ajax vs PSV - May 14, 2023
                  </p>
                  <div className="mt-1 flex">
                    <p className="text-xs text-gray-500">
                      Uploaded 3 hours ago
                    </p>
                    <p className="ml-2 text-xs text-gray-500">1.2GB</p>
                  </div>
                </div>
                <div className="flex items-center">
                  <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                    Processing Complete
                  </span>
                  <button className="ml-4 text-sm text-blue-600 hover:text-blue-500">
                    View
                  </button>
                </div>
              </div>
              <div className="mt-2 w-full bg-gray-200 rounded-full h-1.5">
                <div
                  className="bg-green-600 h-1.5 rounded-full"
                  style={{ width: "100%" }}
                ></div>
              </div>
            </div>

            {/* Upload Item 2 */}
            <div className="px-4 py-4 sm:px-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-blue-600 truncate">
                    Feyenoord vs AZ - May 7, 2023
                  </p>
                  <div className="mt-1 flex">
                    <p className="text-xs text-gray-500">Uploaded 1 day ago</p>
                    <p className="ml-2 text-xs text-gray-500">890MB</p>
                  </div>
                </div>
                <div className="flex items-center">
                  <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-green-100 text-green-800">
                    Processing Complete
                  </span>
                  <button className="ml-4 text-sm text-blue-600 hover:text-blue-500">
                    View
                  </button>
                </div>
              </div>
              <div className="mt-2 w-full bg-gray-200 rounded-full h-1.5">
                <div
                  className="bg-green-600 h-1.5 rounded-full"
                  style={{ width: "100%" }}
                ></div>
              </div>
            </div>

            {/* Upload Item 3 - In Progress */}
            <div className="px-4 py-4 sm:px-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-blue-600 truncate">
                    FC Utrecht vs FC Twente - Today
                  </p>
                  <div className="mt-1 flex">
                    <p className="text-xs text-gray-500">Uploading now</p>
                    <p className="ml-2 text-xs text-gray-500">945MB</p>
                  </div>
                </div>
                <div className="flex items-center">
                  <span className="px-2 inline-flex text-xs leading-5 font-semibold rounded-full bg-yellow-100 text-yellow-800">
                    Uploading
                  </span>
                  <button className="ml-4 text-sm text-red-600 hover:text-red-500">
                    Cancel
                  </button>
                </div>
              </div>
              <div className="mt-2 w-full bg-gray-200 rounded-full h-1.5">
                <div
                  className="bg-blue-600 h-1.5 rounded-full"
                  style={{ width: "65%" }}
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
