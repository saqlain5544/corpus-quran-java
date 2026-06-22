import { SearchResults } from "@/components/SearchResults";

interface Props {
  searchParams: Promise<{ q?: string; type?: string }>;
}

export default async function SearchPage({ searchParams }: Props) {
  const { q, type } = await searchParams;

  return (
    <div className="flex flex-col min-h-screen">
      <header className="sticky top-0 z-50 bg-background/80 backdrop-blur-md border-b border-border">
        <div className="max-w-4xl mx-auto px-4 py-4 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold text-accent">Al-Quran Al-Kareem</h1>
            <p className="text-xs text-muted mt-0.5">Search</p>
          </div>
        </div>
      </header>

      <main className="flex-1 max-w-4xl mx-auto px-4 py-6 w-full">
        <SearchResults initialQuery={q} initialType={type as "text" | "root" | "gloss" | undefined} />
      </main>
    </div>
  );
}
