// Client component for the form to use hooks like useState
// We can extract this to its own file: src/components/UrlShortenerForm.tsx
// For now, keeping it here for simplicity of the example.

"use client"; // This directive is crucial for components with interactivity & hooks

import { useState } from 'react';
import { shortenUrlAction } from '../actions/urlshortener';

function UrlShortenerForm() {
  const [longUrl, setLongUrl] = useState('');
  const [shortUrl, setShortUrl] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false); // For basic loading state

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsLoading(true);
    setError('');
    setShortUrl('');

    if (!longUrl.trim()) {
      setError('Please enter a URL.');
      setIsLoading(false);
      return;
    }

    try {
      // Call the Server Action
      const result = await shortenUrlAction(longUrl);

      if (result.error) {
        setError(result.error);
      } else if (result.shortUrl) {
        setShortUrl(result.shortUrl);
      }
    } catch (e) {
      console.error("Client-side error:", e);
      setError('An unexpected error occurred. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div>
        <label htmlFor="longUrl" className="block text-sm font-medium text-slate-300 mb-1">
          Enter Long URL
        </label>
        <input
          type="url"
          id="longUrl"
          name="longUrl"
          value={longUrl}
          onChange={(e) => setLongUrl(e.target.value)}
          placeholder="https://example.com/very-long-url"
          required
          className="w-full px-4 py-3 rounded-md border border-slate-600 bg-slate-700 text-white placeholder-slate-400 focus:border-sky-500 focus:ring-1 focus:ring-sky-500 transition duration-150 ease-in-out"
        />
      </div>

      <button
        type="submit"
        disabled={isLoading}
        className="w-full flex justify-center py-3 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-slate-800 focus:ring-sky-500 disabled:opacity-50 disabled:cursor-not-allowed transition duration-150 ease-in-out"
      >
        {isLoading ? (
          <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        ) : (
          'Shorten URL'
        )}
      </button>

      {error && (
        <p className="mt-2 text-sm text-red-400 bg-red-900/30 p-3 rounded-md text-center">{error}</p>
      )}

      {shortUrl && (
        <div className="mt-6 p-4 border border-dashed border-green-500/50 rounded-md bg-green-900/20">
          <p className="text-sm text-green-300 mb-1">Shortened URL:</p>
          <a
            href={shortUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="block break-all text-lg font-medium text-green-400 hover:text-green-300 hover:underline"
          >
            {shortUrl}
          </a>
           <button
            onClick={() => navigator.clipboard.writeText(shortUrl)}
            className="mt-3 px-3 py-1.5 text-xs font-medium text-sky-300 bg-sky-700/50 hover:bg-sky-600/50 rounded-md transition"
          >
            Copy to Clipboard
          </button>
        </div>
      )}
    </form>
  );}
export default UrlShortenerForm;