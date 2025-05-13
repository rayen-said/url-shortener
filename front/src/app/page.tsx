// src/app/page.tsx
import UrlShortenerForm from '../components/UrlShortenerForm';

export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center bg-gradient-to-br from-slate-900 to-slate-800 p-8 text-white">
      <div className="w-full max-w-xl rounded-xl bg-slate-800/70 p-8 shadow-2xl backdrop-blur-md">
        <h1 className="mb-8 text-center text-4xl font-bold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-sky-400 to-blue-500">
          URL Shortener
        </h1>

        <UrlShortenerForm />
      </div>
      <footer className="mt-12 text-sm text-slate-500">
        <p>Powered by Next.js & Go</p>
      </footer>
    </main>
  );
}
