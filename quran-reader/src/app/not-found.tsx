export default function NotFound() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-4">
      <div className="text-center max-w-md">
        <div className="w-16 h-16 rounded-full bg-accent-light flex items-center justify-center mx-auto mb-4">
          <span className="text-2xl">🔍</span>
        </div>
        <h1 className="text-xl font-semibold mb-2">Surah Not Found</h1>
        <p className="text-sm text-muted mb-6">
          The surah you are looking for does not exist. The Quran has 114 surahs.
        </p>
        <a
          href="/"
          className="inline-flex px-4 py-2 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors"
        >
          Back to all surahs
        </a>
      </div>
    </div>
  );
}
