"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

interface SearchResult {
  surahNumber: number;
  surahName: string;
  verseNumber: number;
  verseText: string;
  matchContext?: string;
  relevance: number;
}

const SEARCH_TYPES = [
  { id: "text", label: "Arabic Text", icon: "T" },
  { id: "root", label: "Arabic Root", icon: "ع" },
  { id: "gloss", label: "English Gloss", icon: "EN" },
] as const;

type SearchType = (typeof SEARCH_TYPES)[number]["id"];

export function SearchResults({
  initialQuery,
  initialType,
}: {
  initialQuery?: string;
  initialType?: SearchType;
}) {
  const router = useRouter();
  const [query, setQuery] = useState(initialQuery || "");
  const [searchType, setSearchType] = useState<SearchType>(initialType || "text");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [searched, setSearched] = useState(false);

  const doSearch = useCallback(async () => {
    if (!query.trim()) return;
    setLoading(true);
    setSearched(true);

    try {
      const res = await fetch(
        `/api/search?q=${encodeURIComponent(query.trim())}&type=${searchType}&limit=50`
      );
      const data = await res.json();
      setResults(data.results || []);
    } catch {
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, [query, searchType]);

  useEffect(() => {
    if (initialQuery) {
      doSearch();
    }
  }, [initialQuery, doSearch]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!query.trim()) return;
    router.push(`/search?q=${encodeURIComponent(query.trim())}&type=${searchType}`);
    doSearch();
  };

  return (
    <div className="max-w-4xl mx-auto px-4 py-8 w-full">
      {/* Search header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-foreground mb-1">Search</h1>
        <p className="text-sm text-muted">Search Arabic text, roots, or English glosses</p>
      </div>

      {/* Search form */}
      <form onSubmit={handleSubmit} className="space-y-4 mb-8">
        <div className="relative">
          <svg className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search the Quran..."
            className="w-full pl-10 pr-4 py-3 rounded-xl border border-border bg-card text-foreground placeholder:text-muted focus:outline-none focus:ring-2 focus:ring-accent/30 focus:border-accent transition-all text-base shadow-sm"
            autoFocus
          />
        </div>

        {/* Search type tabs */}
        <div className="flex items-center gap-1 bg-card border border-border rounded-xl p-1">
          {SEARCH_TYPES.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setSearchType(t.id)}
              className={`flex-1 flex items-center justify-center gap-2 text-sm py-2 px-3 rounded-lg transition-all ${
                searchType === t.id
                  ? "bg-accent text-white shadow-sm"
                  : "text-muted hover:text-foreground hover:bg-card-hover"
              }`}
            >
              <span className="font-bold">{t.icon}</span>
              <span>{t.label}</span>
            </button>
          ))}
        </div>
      </form>

      {/* Loading */}
      {loading && (
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-2 border-accent border-t-transparent rounded-full animate-spin" />
        </div>
      )}

      {/* No query state */}
      {!loading && !searched && !query && (
        <div className="text-center py-16 border border-dashed border-border rounded-2xl">
          <svg className="w-12 h-12 mx-auto text-muted-light mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <p className="text-sm text-muted">Enter a search term above</p>
          <p className="text-xs text-muted-light mt-1">Try a word like "mercy" or "raHma"</p>
        </div>
      )}

      {/* No results */}
      {!loading && searched && results.length === 0 && (
        <div className="text-center py-16 border border-dashed border-border rounded-2xl">
          <svg className="w-12 h-12 mx-auto text-muted-light mb-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm text-muted">No results found for "{query}"</p>
          <p className="text-xs text-muted-light mt-1">Try different terms or search type</p>
        </div>
      )}

      {/* Results */}
      {!loading && results.length > 0 && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <p className="text-sm text-muted">
              <span className="font-semibold text-foreground">{results.length}</span> result{results.length !== 1 ? "s" : ""} for "{query}"
            </p>
            <p className="text-xs text-muted-light">sorted by relevance</p>
          </div>
          <div className="space-y-2">
            {results.map((result, i) => (
              <Link
                key={`${result.surahNumber}:${result.verseNumber}-${i}`}
                href={`/surah/${result.surahNumber}#verse-${result.verseNumber}`}
                className="block p-4 rounded-xl bg-card hover:bg-card-hover border border-border transition-all hover:shadow-md hover:border-accent group"
              >
                <div className="flex items-center gap-2 mb-2">
                  <span className="text-xs font-semibold text-accent bg-accent-light px-2 py-0.5 rounded">
                    {result.surahName}
                  </span>
                  <span className="text-xs text-muted font-mono">
                    {result.surahNumber}:{result.verseNumber}
                  </span>
                  {result.matchContext && (
                    <span className="text-xs text-muted truncate flex-1 ml-2 hidden sm:block italic">
                      {result.matchContext}
                    </span>
                  )}
                </div>
                <p className="arabic text-xl text-quran-text leading-relaxed group-hover:text-accent transition-colors">
                  {result.verseText}
                </p>
              </Link>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}