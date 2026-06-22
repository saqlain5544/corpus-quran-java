"use client";

import Link from "next/link";

interface Props {
  surahNumber: number;
  englishName: string;
  fontSize: number;
  showTranslation: boolean;
  onFontSizeChange: (delta: number) => void;
  onTranslationToggle: () => void;
}

/**
 * Sticky top bar: back link, surah name, prev/next nav, font size,
 * translation toggle. Editorial — single compact row.
 */
export function SurahHeader({
  surahNumber,
  englishName,
  fontSize,
  showTranslation,
  onFontSizeChange,
  onTranslationToggle,
}: Props) {
  return (
    <header className="sticky top-0 z-30 glass-strong">
      <div className="max-w-5xl mx-auto px-4">
        <div className="flex items-center justify-between h-10">
          <div className="flex items-center gap-2 min-w-0">
            <Link
              href="/"
              className="flex items-center gap-1 text-[11px] text-muted hover:text-foreground transition-colors"
            >
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              <span className="hidden sm:inline">Surahs</span>
            </Link>
            <span className="text-border-strong">/</span>
            <span className="text-[11px] font-medium text-foreground truncate font-serif">
              {englishName}
            </span>
          </div>

          <div className="flex items-center gap-1">
            {surahNumber > 1 && (
              <Link
                href={`/surah/${surahNumber - 1}`}
                className="w-6 h-6 rounded flex items-center justify-center text-muted hover:text-foreground hover:bg-card-hover transition-colors"
                title="Previous Surah"
                aria-label="Previous Surah"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
                </svg>
              </Link>
            )}
            {surahNumber < 114 && (
              <Link
                href={`/surah/${surahNumber + 1}`}
                className="w-6 h-6 rounded flex items-center justify-center text-muted hover:text-foreground hover:bg-card-hover transition-colors"
                title="Next Surah"
                aria-label="Next Surah"
              >
                <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </Link>
            )}

            <span className="w-px h-3.5 bg-border mx-1" />

            <div className="flex items-center bg-card-hover/60 rounded p-0.5">
              <button
                onClick={() => onFontSizeChange(-2)}
                className="w-5 h-5 rounded flex items-center justify-center text-muted hover:text-foreground transition-colors text-[10px] font-bold"
                title="Decrease font size"
                aria-label="Decrease font size"
              >
                A−
              </button>
              <span className="text-[10px] text-muted font-mono w-5 text-center tabular-nums">{fontSize}</span>
              <button
                onClick={() => onFontSizeChange(2)}
                className="w-5 h-5 rounded flex items-center justify-center text-muted hover:text-foreground transition-colors text-[10px] font-bold"
                title="Increase font size"
                aria-label="Increase font size"
              >
                A+
              </button>
            </div>

            <span className="w-px h-3.5 bg-border mx-1" />

            <button
              onClick={onTranslationToggle}
              className={`text-[10px] px-2 py-0.5 rounded transition-all ${
                showTranslation
                  ? "bg-accent text-white"
                  : "text-muted hover:text-foreground hover:bg-card-hover"
              }`}
              aria-pressed={showTranslation}
              aria-label="Toggle English translation"
            >
              EN
            </button>
          </div>
        </div>
      </div>
    </header>
  );
}
