"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import { useBookmarks, type Bookmark } from "@/hooks/useBookmarks";

interface BookmarkButtonProps {
  surahNumber: number;
  verseNumber: number;
  surahName: string;
}

export function BookmarkButton({ surahNumber, verseNumber, surahName }: BookmarkButtonProps) {
  const { isBookmarked, toggleBookmark } = useBookmarks();
  const bookmarked = isBookmarked(surahNumber, verseNumber);
  const [animating, setAnimating] = useState(false);

  const handleClick = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setAnimating(true);
      setTimeout(() => setAnimating(false), 300);
      toggleBookmark({ surahNumber, verseNumber, surahName, timestamp: Date.now() });
    },
    [surahNumber, verseNumber, surahName, toggleBookmark]
  );

  return (
    <button
      onClick={handleClick}
      className={`p-1.5 rounded transition-all hover:bg-accent-light ${
        animating ? "bookmark-active" : ""
      } ${
        bookmarked ? "text-amber-500" : "text-muted hover:text-amber-500"
      }`}
      title={bookmarked ? "Remove bookmark" : "Bookmark this verse"}
    >
      <svg
        className="w-4 h-4"
        fill={bookmarked ? "currentColor" : "none"}
        stroke="currentColor"
        viewBox="0 0 24 24"
        strokeWidth={2}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z"
        />
      </svg>
    </button>
  );
}

export function BookmarkList() {
  const { bookmarks, removeBookmark, loaded } = useBookmarks();
  const [open, setOpen] = useState(false);

  if (!loaded || bookmarks.length === 0) return null;

  return (
    <div className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1.5 text-xs text-muted hover:text-accent transition-colors px-2.5 py-1.5 rounded-lg hover:bg-accent-light"
        title="Bookmarks"
      >
        <svg className="w-4 h-4" fill={bookmarks.length > 0 ? "currentColor" : "none"} stroke="currentColor" viewBox="0 0 24 24" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
        </svg>
        <span className="font-medium">{bookmarks.length}</span>
      </button>

      {open && (
        <>
          <div className="fixed inset-0 z-50" onClick={() => setOpen(false)} />
          <div className="absolute right-0 top-full mt-2 w-80 bg-card border border-border rounded-xl shadow-xl z-50 max-h-80 overflow-hidden">
            <div className="px-4 py-3 border-b border-border flex items-center justify-between bg-card-hover">
              <h3 className="font-semibold text-sm text-foreground">Bookmarks</h3>
              <button
                onClick={() => setOpen(false)}
                className="text-xs text-muted hover:text-foreground"
              >
                ✕
              </button>
            </div>
            <div className="overflow-y-auto max-h-64">
              {bookmarks.map((b) => (
                <div
                  key={`${b.surahNumber}:${b.verseNumber}`}
                  className="flex items-center gap-3 px-4 py-2.5 hover:bg-card-hover group border-b border-border/30 last:border-b-0"
                >
                  <Link
                    href={`/surah/${b.surahNumber}#verse-${b.verseNumber}`}
                    className="flex-1 min-w-0"
                    onClick={() => setOpen(false)}
                  >
                    <p className="text-sm font-medium text-foreground truncate">
                      {b.surahName}
                    </p>
                    <p className="text-xs text-muted font-mono">
                      Verse {b.surahNumber}:{b.verseNumber}
                    </p>
                    {b.note && <p className="text-xs text-muted truncate mt-0.5 italic">{b.note}</p>}
                  </Link>
                  <button
                    onClick={() => removeBookmark(b.surahNumber, b.verseNumber)}
                    className="text-muted hover:text-red-500 opacity-0 group-hover:opacity-100 transition-opacity p-1 rounded hover:bg-card"
                    title="Remove bookmark"
                  >
                    <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}