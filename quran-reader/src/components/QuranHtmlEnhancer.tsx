"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { WordTooltip } from "@/components/WordTooltip";
import { WordAnalysisModal } from "@/components/WordAnalysisModal";
import { buckwalterToArabic } from "@/lib/arabic";
import type { Segment, WordData } from "@/lib/types";

/** Raw shape of each entry in a span's `data-morphology` JSON attribute. */
interface RawMorph {
  ID?: string;
  Sura_No?: string;
  Verse_No?: string;
  Word_No?: string;
  Segment_No?: string;
  Word?: string;
  Without_Diacritics?: string;
  Segmented_Word?: string;
  Morph_Tag?: string;
  Morph_Type?: string;
  Lemma?: string;
  Root?: string;
  Gender?: string;
  Case_Mood?: string;
  Case_Mood_Marker?: string;
  Syntactic_Role?: string;
  Gloss?: string;
  Possessive_Construct?: string;
  Invariable_Declinable?: string;
  Phrase?: string;
  Phrasal_Function?: string;
  Punctuation_Mark?: string;
  [k: string]: unknown;
}

function parseMorph(span: HTMLElement): RawMorph[] | null {
  const raw = span.getAttribute("data-morphology");
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    return Array.isArray(parsed) ? (parsed as RawMorph[]) : null;
  } catch {
    return null;
  }
}

function toSegment(r: RawMorph): Segment {
  return {
    segmentNumber: parseInt(r.Segment_No ?? "0", 10) || 0,
    text: r.Segmented_Word ?? r.Word ?? "",
    partOfSpeech: r.Morph_Tag ?? "",
    morphType: r.Morph_Type ?? "",
    lemma: r.Lemma || undefined,
    root: r.Root || undefined,
    gender: r.Gender || undefined,
    case: r.Case_Mood || undefined,
    syntacticRole: r.Syntactic_Role || undefined,
    gloss: r.Gloss || undefined,
    possessiveConstruct: r.Possessive_Construct || undefined,
    caseMoodMarker: r.Case_Mood_Marker || undefined,
    invariableDeclinable: r.Invariable_Declinable || undefined,
    phrase: r.Phrase || undefined,
    phrasalFunction: r.Phrasal_Function || undefined,
    punctuationMark: r.Punctuation_Mark || undefined,
  };
}

/** Stem segment is the linguistic head; fall back to the last segment. */
function pickPrimary(segs: RawMorph[]): RawMorph | null {
  return (
    segs.find((s) => s.Morph_Type === "Stem") ??
    segs.find((s) => s.Syntactic_Role) ??
    segs[segs.length - 1] ??
    null
  );
}

/** Fill in the data-* attrs that WordTooltip reads, derived from data-morphology. */
function enrichSpan(span: HTMLElement): void {
  if (span.dataset.token) return;
  const segs = parseMorph(span);
  if (!segs || segs.length === 0) return;

  const primary = pickPrimary(segs);
  const uniqueGlosses = [
    ...new Set(
      segs.map((s) => s.Gloss).filter((g): g is string => !!g && g !== "null")
    ),
  ];

  span.dataset.token = span.textContent?.trim() ?? "";
  span.dataset.translation = uniqueGlosses.join(" ");
  span.dataset.gloss = primary?.Gloss ?? "";
  span.dataset.pos = primary?.Morph_Tag ?? "";
  span.dataset.role = primary?.Syntactic_Role ?? "";
  span.dataset.case = primary?.Case_Mood ?? "";
  span.dataset.gender = primary?.Gender ?? "";
  span.dataset.root = buckwalterToArabic(primary?.Root ?? "");
  span.dataset.lemma = buckwalterToArabic(primary?.Lemma ?? "");
}

function buildWordData(span: HTMLElement): WordData | null {
  const segs = parseMorph(span);
  if (!segs || segs.length === 0) return null;

  const uniqueGlosses = [
    ...new Set(
      segs.map((s) => s.Gloss).filter((g): g is string => !!g && g !== "null")
    ),
  ];

  return {
    // Global word index across Quran — used for display only, not as DB id.
    id: parseInt(span.dataset.index ?? "0", 10) || 0,
    wordNumber: parseInt(span.dataset.word ?? "0", 10) || 0,
    token: span.textContent?.trim() ?? "",
    withoutDiacritics: segs[0]?.Without_Diacritics ?? "",
    translation: uniqueGlosses.join(" "),
    punctuationMark: "",
    segments: segs.map(toSegment),
  };
}

/**
 * Wraps the injected surah HTML body, attaches delegated event listeners
 * to all `.quran-word` spans, and renders the tooltip + analysis modal.
 *
 * - Hovering a word → fills data-* attrs (from data-morphology) and shows
 *   the existing WordTooltip.
 * - Clicking a word → builds a WordData from the span and opens the
 *   existing WordAnalysisModal.
 *
 * The HTML comes from the sibling `quran-html/` directory via the server
 * component; this client component never re-parses the structure, just
 * decorates spans that are already in the DOM.
 */
export function QuranHtmlEnhancer({ html }: { html: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [hoveredEl, setHoveredEl] = useState<HTMLElement | null>(null);
  const [clicked, setClicked] = useState<{ word: WordData; verseNumber: number } | null>(null);
  const leaveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const cancelLeave = useCallback(() => {
    if (leaveTimer.current) {
      clearTimeout(leaveTimer.current);
      leaveTimer.current = null;
    }
  }, []);

  const scheduleLeave = useCallback(() => {
    cancelLeave();
    leaveTimer.current = setTimeout(() => setHoveredEl(null), 120);
  }, [cancelLeave]);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const onOver = (e: Event) => {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      const word = target.closest?.(".quran-word") as HTMLElement | null;
      if (!word || !container.contains(word)) return;
      cancelLeave();
      enrichSpan(word);
      setHoveredEl(word);
    };

    const onOut = (e: Event) => {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      const word = target.closest?.(".quran-word");
      if (!word) return;
      // Don't leave if we're entering a child (e.g., tooltip).
      const related = (e as MouseEvent).relatedTarget as Node | null;
      if (related && word.contains(related)) return;
      scheduleLeave();
    };

    const onClick = (e: Event) => {
      const target = e.target as HTMLElement | null;
      if (!target) return;
      const word = target.closest?.(".quran-word") as HTMLElement | null;
      if (!word || !container.contains(word)) return;
      const wordData = buildWordData(word);
      if (!wordData) return;
      const verseNumber = parseInt(word.dataset.ayah ?? "0", 10) || 0;
      e.preventDefault();
      cancelLeave();
      setHoveredEl(null);
      setClicked({ word: wordData, verseNumber });
    };

    container.addEventListener("mouseover", onOver);
    container.addEventListener("mouseout", onOut);
    container.addEventListener("click", onClick);

    return () => {
      container.removeEventListener("mouseover", onOver);
      container.removeEventListener("mouseout", onOut);
      container.removeEventListener("click", onClick);
      cancelLeave();
    };
  }, [cancelLeave, scheduleLeave]);

  return (
    <>
      <div
        ref={containerRef}
        className="quran-html-body flex-1 max-w-3xl mx-auto px-4 sm:px-6 py-6 w-full"
        dir="rtl"
        lang="ar"
        dangerouslySetInnerHTML={{ __html: html }}
      />
      <WordTooltip
        anchorEl={hoveredEl}
        onLeave={scheduleLeave}
        onTooltipEnter={cancelLeave}
      />
      {clicked && (
        <WordAnalysisModal
          word={clicked.word}
          verseNumber={clicked.verseNumber}
          visible={true}
          onClose={() => setClicked(null)}
        />
      )}
    </>
  );
}