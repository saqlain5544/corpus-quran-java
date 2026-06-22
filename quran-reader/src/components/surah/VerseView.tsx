"use client";

import { useState, useMemo } from "react";
import { BookmarkButton } from "@/components/BookmarkButton";
import { buckwalterToArabic, toArabicNumeral } from "@/lib/arabic";
import { align, tokenize } from "@/lib/tokenizer";
import type { WordData, VerseData } from "@/lib/types";

interface Props {
  verse: VerseData;
  verseIndex: number;
  words: WordData[];
  fontSize: number;
  showTranslation: boolean;
  surahName: string;
  onWordEnter: (anchorEl: HTMLElement) => void;
  onWordLeave: () => void;
  onWordClick: (verseNumber: number, word: WordData) => void;
}

/**
 * Renders a single ayah in mushaf / tanzil.net style.
 *
 * SSR-only data flow:
 *   1. Tokenize verse.text (whitespace-split, classify as word/pause/ayah)
 *   2. Align text tokens with DB words (index-based with split/join strategies)
 *   3. Render each token with full metadata as `data-*` attributes on the span
 *      — the raw DB text is the visible content, untouched
 *
 * Client side (only event handlers) reads the `data-*` attributes directly
 * to construct the tooltip — no extra server call, no client-side
 * tokenization.
 *
 * Each word span carries:
 *   data-word-id        DB row id (for /api/word/[id] on click)
 *   data-word-number    position within the verse
 *   data-token          Arabic text (with diacritics)
 *   data-translation    English gloss
 *   data-gloss          stem-segment gloss
 *   data-pos            part of speech (NOUN_ABSTRACT, PV, …)
 *   data-role           syntactic role (SUBJ, OBJ, ADJ, …)
 *   data-case           grammatical case (NOMINATIVE, GENITIVE, …)
 *   data-gender         M / F
 *   data-root           Arabic root letters
 *   data-lemma          Arabic lemma
 */
export function VerseView({
  verse,
  verseIndex,
  words,
  fontSize,
  showTranslation,
  surahName,
  onWordEnter,
  onWordLeave,
  onWordClick,
}: Props) {
  const [hoveredTokenIdx, setHoveredTokenIdx] = useState<number | null>(null);

  // Tokenize + align once per (verse, words) — runs on server during SSR.
  const { tokens, alignment } = useMemo(() => {
    const toks = tokenize(verse.text);
    const textWords = toks
      .filter((t) => t.type === "word")
      .map((t) => t.text);
    const dbWords = words.map((w) => ({ number: w.wordNumber, token: w.token }));
    return { tokens: toks, alignment: align(dbWords, textWords) };
  }, [verse.text, words]);

  // Map text token index → DB word (used for metadata)
  const wordByTextIdx = useMemo(() => {
    const m = new Map<number, WordData>();
    alignment.forEach((entry) => {
      const w = words[entry.dbIndex];
      if (w) for (const ti of entry.textIndices) m.set(ti, w);
    });
    return m;
  }, [alignment, words]);

  return (
    <article
      id={`verse-${verse.number}`}
      className="verse-row group relative"
      style={{ animationDelay: `${Math.min(verseIndex * 20, 400)}ms` }}
    >
      {/* Floating bookmark — appears on verse hover */}
      <div className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity z-10">
        <BookmarkButton
          surahNumber={verse.surahNumber}
          verseNumber={verse.number}
          surahName={surahName}
        />
      </div>

      {/* Mushaf line: raw DB text rendered with metadata. No client-side
          text mutation. Each word has full data-* attrs; pause/ayah marks
          have just class names. */}
      <p
        className="aya arabic verse-text text-quran-text"
        dir="rtl"
        style={{
          fontSize: `${fontSize}px`,
          lineHeight: 2.35,
          letterSpacing: 0,
          textAlign: "justify",
          textJustify: "inter-word",
          wordSpacing: "normal",
        }}
      >
        {tokens.map((tok, idx) => {
          if (tok.type === "pause") {
            return (
              <span
                key={idx}
                className="sign verse-word select-none"
                aria-hidden="true"
                data-sign
              >
                {tok.text}
              </span>
            );
          }
          if (tok.type === "ayah") {
            return (
              <span
                key={idx}
                className="ayaNumber verse-word select-none"
                aria-label={`Verse ${verse.number}`}
                data-ayah={verse.number}
              >
                {tok.text}
              </span>
            );
          }

          // Word token
          const word = wordByTextIdx.get(idx);
          if (!word) {
            // No DB match (alignment gap). Render plain so the page still
            // shows the raw verse text without a hole.
            return (
              <span key={idx} className="arabic verse-word">
                {tok.text}
              </span>
            );
          }

          const attrs = wordMetadataAttrs(word);

          return (
            <span
              key={idx}
              {...attrs}
              className={`word-clickable verse-word ${
                hoveredTokenIdx === idx ? "word-selected" : ""
              }`}
              onMouseEnter={(e) => {
                setHoveredTokenIdx(idx);
                onWordEnter(e.currentTarget);
              }}
              onMouseLeave={() => {
                setHoveredTokenIdx(null);
                onWordLeave();
              }}
              onClick={() => onWordClick(verse.number, word)}
            >
              {tok.text}
            </span>
          );
        })}

        {/* Auto-append ﴾N﴾ if the verse.text doesn't already contain one. */}
        {!verse.text.includes("\uFD3F") && (
          <span
            className="ayaNumber verse-word select-none"
            aria-label={`Verse ${verse.number}`}
            data-ayah={verse.number}
          >
            ﴿{toArabicNumeral(verse.number)}﴾
          </span>
        )}
      </p>

      {showTranslation && (
        <Translation words={words} verseNumber={verse.number} />
      )}
    </article>
  );
}

// ── Metadata extraction ──────────────────────────────────────────

/**
 * Build the data-* attribute object for a word span.
 *
 * Stems carry the linguistic content (POS, role, case, gender, root,
 * lemma, gloss); if a word has no Stem segment, we fall back to the
 * last segment. Buckwalter root/lemma are converted to Arabic script
 * (e.g. "EAlm" → "علم") via `buckwalterToArabic`.
 */
export function wordMetadataAttrs(word: WordData): Record<string, string> {
  const stem =
    word.segments.find((s) => s.morphType === "Stem") ||
    word.segments.find((s) => s.syntacticRole) ||
    word.segments[word.segments.length - 1];

  const root = stem?.root ? buckwalterToArabic(stem.root) : "";
  const lemma = stem?.lemma ? buckwalterToArabic(stem.lemma) : "";

  // Compact, lowercased keys so the DOM stays tidy
  return {
    "data-word-id": String(word.id ?? ""),
    "data-word-number": String(word.wordNumber ?? ""),
    "data-token": word.token ?? "",
    "data-translation": word.translation ?? "",
    "data-gloss": stem?.gloss ?? "",
    "data-pos": stem?.partOfSpeech ?? "",
    "data-role": stem?.syntacticRole ?? "",
    "data-case": stem?.case ?? "",
    "data-gender": stem?.gender ?? "",
    "data-root": root,
    "data-lemma": lemma,
  };
}

// ── Translation panel ──────────────────────────────────────────

function Translation({ words, verseNumber }: { words: WordData[]; verseNumber: number }) {
  return (
    <div className="mt-3 pt-3 border-t border-dashed border-border/70">
      <p className="text-sm text-muted leading-relaxed italic font-serif">
        <span className="verse-number-ar mr-2">{toArabicNumeral(verseNumber)}.</span>
        {words.map((w) => w.translation).join(" ")}
      </p>
    </div>
  );
}