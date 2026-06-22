"use client";

import Link from "next/link";

interface Props {
  surahNumber: number;
}

/**
 * Bottom navigation: previous / all surahs / next. Editorial — hairline
 * top border, small text, no card background.
 */
export function FooterNav({ surahNumber }: Props) {
  return (
    <footer className="mt-8 border-t border-border">
      <div className="max-w-5xl mx-auto px-4 py-4 flex flex-wrap items-center justify-between gap-3 text-[11px] text-muted">
        {surahNumber > 1 ? (
          <Link
            href={`/surah/${surahNumber - 1}`}
            className="flex items-center gap-1.5 hover:text-foreground transition-colors"
          >
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
            <span>Previous</span>
          </Link>
        ) : <span />}
        <Link
          href="/"
          className="hover:text-foreground transition-colors"
        >
          All Surahs
        </Link>
        {surahNumber < 114 ? (
          <Link
            href={`/surah/${surahNumber + 1}`}
            className="flex items-center gap-1.5 hover:text-foreground transition-colors"
          >
            <span>Next</span>
            <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </Link>
        ) : <span />}
      </div>
    </footer>
  );
}
