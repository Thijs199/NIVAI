'use client';

import { useState, FormEvent, ChangeEvent } from 'react';

export default function UploadPage() {
  const [title, setTitle] = useState<string>(''); // "title" for the backend
  const [videoFile, setVideoFile] = useState<File | null>(null);
  const [trackingFile, setTrackingFile] = useState<File | null>(null);
  const [eventFile, setEventFile] = useState<File | null>(null);

  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [message, setMessage] = useState<string | null>(null);
  const [messageType, setMessageType] = useState<'success' | 'error' | null>(null);

  const handleFileChange =
    (setter: (file: File | null) => void) => (event: ChangeEvent<HTMLInputElement>) => {
      if (event.target.files && event.target.files[0]) {
        setter(event.target.files[0]);
      } else {
        setter(null);
      }
    };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsLoading(true);
    setMessage(null);
    setMessageType(null);

    if (!title.trim()) {
      setMessage('Match Name is required.');
      setMessageType('error');
      setIsLoading(false);
      return;
    }
    if (!trackingFile) {
      setMessage('Tracking Data File is required.');
      setMessageType('error');
      setIsLoading(false);
      return;
    }
    if (!eventFile) {
      setMessage('Event Data File is required.');
      setMessageType('error');
      setIsLoading(false);
      return;
    }

    const formData = new FormData();
    formData.append('title', title); // Backend expects "title"

    if (videoFile) {
      formData.append('video_file', videoFile);
    }
    // Tracking and event files are checked for null above, so they should exist.
    formData.append('tracking_file', trackingFile!);
    formData.append('event_file', eventFile!);

    // Example of other potential metadata fields (if backend supports them on this endpoint)
    // formData.append("description", "Match description placeholder");
    // formData.append("match_id", "client-generated-match-id"); // Go backend generates videoID which is used as match_id for processing

    try {
      const response = await fetch('/api/v1/videos', {
        // Assuming Go backend is served on the same origin or proxied
        method: 'POST',
        body: formData,
      });

      const data = await response.json();

      if (response.ok) {
        setMessage(
          `Match uploaded successfully! Video ID: ${data.video_id}. Processing initiated.`
        );
        setMessageType('success');
        // Reset form
        setTitle('');
        setVideoFile(null);
        setTrackingFile(null);
        setEventFile(null);
        // Clear file input elements visually
        const fileInputs = document.querySelectorAll<HTMLInputElement>('input[type="file"]');
        fileInputs.forEach((input) => (input.value = ''));
      } else {
        setMessage(`Upload failed: ${data.error || data.message || response.statusText}`);
        setMessageType('error');
      }
    } catch (error) {
      console.error('Upload error:', error);
      setMessage(
        `An unexpected error occurred during upload: ${error instanceof Error ? error.message : String(error)}`
      );
      setMessageType('error');
    } finally {
      setIsLoading(false);
    }
  };

  // Basic styling using Tailwind CSS classes (assuming Tailwind is configured)
  // If not, these will be unstyled but functional.
  const inputClass =
    'mt-1 block w-full px-3 py-2 bg-white border border-slate-300 rounded-md text-sm shadow-sm placeholder-slate-400 focus:outline-none focus:border-sky-500 focus:ring-1 focus:ring-sky-500 disabled:bg-slate-50 disabled:text-slate-500 disabled:border-slate-200 disabled:shadow-none';
  const labelClass = 'block text-sm font-medium text-slate-700';
  const buttonClass =
    'mt-4 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:bg-slate-400';
  const messageBoxBaseClass = 'mt-4 p-3 rounded-md text-sm';
  const successBoxClass = `${messageBoxBaseClass} bg-green-100 text-green-700 border border-green-300`;
  const errorBoxClass = `${messageBoxBaseClass} bg-red-100 text-red-700 border border-red-300`;

  return (
    <div className="container mx-auto p-4 max-w-lg">
      <h1 className="text-2xl font-bold mb-6 text-center text-slate-800">Upload New Match Data</h1>

      <form onSubmit={handleSubmit} className="space-y-6 bg-white shadow-md rounded-lg p-6">
        <div>
          <label htmlFor="title" className={labelClass}>
            Match Name:
          </label>
          <input
            type="text"
            id="title"
            name="title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className={inputClass}
            required
            disabled={isLoading}
          />
        </div>

        <div>
          <label htmlFor="video_file" className={labelClass}>
            Video File (Optional):
          </label>
          <input
            type="file"
            id="video_file"
            name="video_file"
            accept="video/*"
            onChange={handleFileChange(setVideoFile)}
            className={inputClass}
            disabled={isLoading}
          />
        </div>

        <div>
          <label htmlFor="tracking_file" className={labelClass}>
            Tracking Data (e.g., .gzip, .parquet):
          </label>
          <input
            type="file"
            id="tracking_file"
            name="tracking_file"
            accept=".gzip,.parquet,.gz" // Adjust accept types as needed
            onChange={handleFileChange(setTrackingFile)}
            className={inputClass}
            required
            disabled={isLoading}
          />
        </div>

        <div>
          <label htmlFor="event_file" className={labelClass}>
            Event Data (e.g., .gzip, .parquet):
          </label>
          <input
            type="file"
            id="event_file"
            name="event_file"
            accept=".gzip,.parquet,.gz" // Adjust accept types as needed
            onChange={handleFileChange(setEventFile)}
            className={inputClass}
            required
            disabled={isLoading}
          />
        </div>

        {/* Example of other text input if needed for backend
        <div>
          <label htmlFor="description" className={labelClass}>Description (Optional):</label>
          <input
            type="text"
            id="description"
            name="description"
            // value={description}
            // onChange={(e) => setDescription(e.target.value)}
            className={inputClass}
            disabled={isLoading}
          />
        </div>
        */}

        <button type="submit" disabled={isLoading} className={buttonClass}>
          {isLoading ? 'Uploading...' : 'Upload Match Data'}
        </button>
      </form>

      {message && (
        <div className={messageType === 'error' ? errorBoxClass : successBoxClass}>{message}</div>
      )}
    </div>
  );
}
