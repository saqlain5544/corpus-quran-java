export default function SurahLoading() {
  return (
    <div className="flex flex-col min-h-screen">
      <div className="border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-6">
          <div className="text-center space-y-4 animate-pulse">
            <div className="h-8 w-48 bg-border rounded mx-auto" />
            <div className="h-4 w-64 bg-border rounded mx-auto" />
            <div className="h-4 w-32 bg-border rounded mx-auto" />
          </div>
        </div>
      </div>
      <main className="flex-1 max-w-4xl mx-auto px-4 py-8 w-full space-y-4">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="animate-pulse space-y-3 py-4 border-b border-divider">
            <div className="h-6 w-full bg-border rounded" />
            <div className="h-6 w-3/4 bg-border rounded" />
          </div>
        ))}
      </main>
    </div>
  );
}
