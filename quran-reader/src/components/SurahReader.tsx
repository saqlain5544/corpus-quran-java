"use client";

import { useState, useCallback, useRef } from "react";
import { SurahHeader, FooterNav, VerseView } from "@/components/surah";
import { WordTooltip } from "@/components/WordTooltip";
import { WordAnalysisModal } from "@/components/WordAnalysisModal";
import { toArabicNumeral } from "@/lib/arabic";
import type { WordData, SurahData } from "@/lib/types";

const FONT_SIZE_MIN = 20;
const FONT_SIZE_MAX = 48;
const FONT_SIZE_STEP = 2;

/**
 * Top-level reader page. Holds UI state (hovered word, clicked word,
 * font size, translation visibility) and composes the surah sections.
 *
 * Editorial layout: sticky compact header, refined title banner with serif
 * type and a hairline divider, the verses flow on the page's grid-paper
 * background, and a small footer.
 *
 * The tooltip reads all its data from the hovered span's `data-*`
 * attributes — this component never receives TooltipWord payloads.
 */
export function SurahReader({ surah }: { surah: SurahData }) {
  const [hoveredEl, setHoveredEl] = useState<HTMLElement | null>(null);
  const [clickedWord, setClickedWord] = useState<{
    word: WordData;
    verseNumber: number;
  } | null>(null);
  const [fontSize, setFontSize] = useState(32);
  const [showTranslation, setShowTranslation] = useState(false);
  const leaveTimer = useRef<ReturnType<typeof setTimeout>>(null);

  const handleWordEnter = useCallback((anchorEl: HTMLElement) => {
    if (leaveTimer.current) {
      clearTimeout(leaveTimer.current);
      leaveTimer.current = null;
    }
    setHoveredEl(anchorEl);
  }, []);

  const handleWordLeave = useCallback(() => {
    leaveTimer.current = setTimeout(() => setHoveredEl(null), 120);
  }, []);

  const handleTooltipEnter = useCallback(() => {
    if (leaveTimer.current) {
      clearTimeout(leaveTimer.current);
      leaveTimer.current = null;
    }
  }, []);

  const handleWordClick = useCallback((verseNumber: number, word: WordData) => {
    setHoveredEl(null);
    setClickedWord({ word, verseNumber });
  }, []);

  const handleFontSizeChange = useCallback((delta: number) => {
    setFontSize((s) => Math.max(FONT_SIZE_MIN, Math.min(FONT_SIZE_MAX, s + delta)));
  }, []);

  return (
    <div className="flex flex-col min-h-screen">
      <SurahHeader
        surahNumber={surah.number}
        englishName={surah.englishName}
        fontSize={fontSize}
        showTranslation={showTranslation}
        onFontSizeChange={handleFontSizeChange}
        onTranslationToggle={() => setShowTranslation((s) => !s)}
      />

      <SurahTitleBanner
        number={surah.number}
        name={surah.name}
        englishName={surah.englishName}
        englishNameTranslation={surah.englishNameTranslation}
        verseCount={surah.verses.length}
        revelationType={surah.revelationType}
      />

      <main className="flex-1 max-w-3xl mx-auto px-4 sm:px-6 py-6 w-full">
        {surah.verses.map((verse, idx) => {
          const words = surah.morphology?.[verse.number] ?? [];
          return (
            <VerseView
              key={verse.number}
              verse={verse}
              verseIndex={idx}
              words={words}
              fontSize={fontSize}
              showTranslation={showTranslation}
              surahName={surah.englishName}
              onWordEnter={handleWordEnter}
              onWordLeave={handleWordLeave}
              onWordClick={handleWordClick}
            />
          );
        })}
      </main>

      <FooterNav surahNumber={surah.number} />

      {/* Overlays */}
      <WordTooltip
        anchorEl={hoveredEl}
        onLeave={handleWordLeave}
        onTooltipEnter={handleTooltipEnter}
      />

      {clickedWord && (
        <WordAnalysisModal
          word={clickedWord.word}
          verseNumber={clickedWord.verseNumber}
          visible={true}
          onClose={() => setClickedWord(null)}
        />
      )}
    </div>
  );
}

// ── Sub-components ──────────────────────────────────────────────

function SurahTitleBanner({
  number,
  name,
  englishName,
  englishNameTranslation,
  verseCount,
  revelationType,
}: {
  number: number;
  name: string;
  englishName: string;
  englishNameTranslation: string;
  verseCount: number;
  revelationType: string;
}) {
  return (
    <header className="border-b border-border">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 py-6 text-center">
        <p className="text-[10px] uppercase tracking-[0.22em] text-muted mb-3 font-serif">
          <span className="text-muted-light">· </span>
          Surah {toArabicNumeral(number)}
          <span className="text-muted-light"> ·</span>
        </p>
        <h1 className="arabic text-4xl sm:text-5xl text-quran-text leading-tight mb-2 surah-title-frame inline-block">
          {name}
        </h1>
        <h2 className="text-lg sm:text-xl font-medium text-foreground mb-1 font-serif">
          {englishName}
        </h2>
        <p className="text-xs text-muted font-serif italic">
          {englishNameTranslation}
          <span className="mx-1.5 text-muted-light">·</span>
          <span className="tabular-nums">{verseCount}</span> verses
          <span className="mx-1.5 text-muted-light">·</span>
          {revelationType}
        </p>
      </div>
    </header>
  );
}
