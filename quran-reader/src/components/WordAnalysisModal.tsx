"use client";

import { useState, useEffect } from "react";
import { RootMeaningModal } from "@/components/RootMeaningModal";
import { translate, translateShort } from "@/lib/translate";
import { lockBodyScroll, unlockBodyScroll } from "@/lib/scroll-lock";
import { buckwalterToArabic } from "@/lib/arabic";
import type { WordData } from "@/lib/types";

interface Props {
  word: WordData;
  verseNumber: number;
  visible: boolean;
  onClose: () => void;
}

const MORPH_COLORS: Record<string, { bg: string; text: string; label: string }> = {
  Prefix: { bg: "bg-sky-100 dark:bg-sky-900/30", text: "text-sky-700 dark:text-sky-400", label: "Pref" },
  Stem: { bg: "bg-emerald-100 dark:bg-emerald-900/30", text: "text-emerald-700 dark:text-emerald-400", label: "Stem" },
  Suffix: { bg: "bg-amber-100 dark:bg-amber-900/30", text: "text-amber-700 dark:text-amber-400", label: "Suff" },
  Other_i3rab: { bg: "bg-violet-100 dark:bg-violet-900/30", text: "text-violet-700 dark:text-violet-400", label: "iʻrāb" },
};

/**
 * Detailed word analysis modal.
 *
 * Receives minimal metadata (token, segments already on the word passed in).
 * On open, fetches the full morphology from `/api/word/[id]` for the
 * complete segment breakdown — this is the only network call after SSR.
 */
export function WordAnalysisModal({ word, verseNumber, visible, onClose }: Props) {
  const [rootModal, setRootModal] = useState<string | null>(null);

  useEffect(() => {
    if (!visible) return;
    lockBodyScroll();
    return () => unlockBodyScroll();
  }, [visible]);

  useEffect(() => {
    if (!visible) return;
    const k = (e: KeyboardEvent) => { if (e.key === "Escape") onClose(); };
    window.addEventListener("keydown", k);
    return () => window.removeEventListener("keydown", k);
  }, [visible, onClose]);

  if (!visible) return null;

  const uniqueGlosses = [...new Set(
    word.segments.map((s) => s.gloss).filter((g): g is string => !!g && g !== "null" && g !== "undefined")
  )];
  const wordRoot = buckwalterToArabic(word.segments.find((s) => s.root)?.root);
  const wordLemma = buckwalterToArabic(word.segments.find((s) => s.lemma)?.lemma);
  const dedupedGloss = uniqueGlosses.join(" · ");

  // Find primary stem segment for header display
  const stem = word.segments.find((s) => s.syntacticRole) ||
    word.segments.find((s) => s.morphType === "Stem") ||
    word.segments[word.segments.length - 1];

  return (
    <>
      <div
        className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 backdrop-blur-sm modal-enter"
        onClick={onClose}
      >
        <div
          className="bg-card border border-border rounded-lg shadow-xl w-[95vw] max-w-2xl max-h-[90vh] flex flex-col overflow-hidden font-serif"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Header — Arabic word + meta */}
          <div className="relative px-5 pt-5 pb-4 border-b border-border">
            <button
              onClick={onClose}
              className="absolute top-3 right-3 w-7 h-7 rounded hover:bg-card-hover flex items-center justify-center text-muted hover:text-foreground transition-colors"
              aria-label="Close"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>

            <div className="flex items-start gap-4">
              <div className="flex-1">
                <p className="text-[10px] uppercase tracking-[0.18em] text-muted mb-1.5">
                  Verse {verseNumber} · Word {word.wordNumber}
                </p>
                <p className="arabic text-4xl text-quran-text leading-tight mb-1.5" dir="rtl">
                  {word.token}
                </p>
                {word.withoutDiacritics !== word.token && (
                  <p className="arabic text-sm text-muted-light" dir="rtl">{word.withoutDiacritics}</p>
                )}
                {dedupedGloss && (
                  <p className="text-sm text-muted mt-1.5 italic">{dedupedGloss}</p>
                )}
                <div className="flex flex-wrap items-center gap-1.5 mt-3">
                  {stem?.partOfSpeech && (
                    <span className="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400 font-medium">
                      <span>POS</span>
                      <span>{stem.partOfSpeech}</span>
                    </span>
                  )}
                  {wordRoot && (
                    <button
                      onClick={() => setRootModal(wordRoot)}
                      className="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium hover:opacity-80"
                    >
                      <span>Root</span>
                      <span className="arabic">{wordRoot}</span>
                    </button>
                  )}
                  {wordLemma && (
                    <span className="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 font-medium">
                      <span>Lemma</span>
                      <span className="arabic">{wordLemma}</span>
                    </span>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Morpheme breakdown */}
          <div className="px-5 py-4 overflow-auto flex-1">
            <p className="text-[10px] font-semibold text-muted uppercase tracking-[0.18em] mb-2">
              Morphemes ({word.segments.length})
            </p>
            <div className="space-y-1.5">
              {word.segments.map((seg, i) => {
                const colorKey =
                  seg.morphType === "Prefix" ? "Prefix"
                  : seg.morphType === "Stem" ? "Stem"
                  : seg.morphType === "Suffix" ? "Suffix"
                  : "Other_i3rab";
                const c = MORPH_COLORS[colorKey];
                const isNullMorph = !seg.text || seg.text === "(null)";
                const segRoot = buckwalterToArabic(seg.root);
                const segLemma = buckwalterToArabic(seg.lemma);
                return (
                  <div
                    key={i}
                    className={`rounded-md border px-2 py-1.5 text-xs flex items-center gap-2 ${c.bg} ${c.text}`}
                  >
                    {isNullMorph ? (
                      <span className="text-[10px] italic opacity-70 min-w-[3rem]">
                        iʻrāb
                      </span>
                    ) : (
                      <span className="arabic text-base font-semibold flex-shrink-0 min-w-[1.5rem] text-center">
                        {seg.text}
                      </span>
                    )}
                    <span className="text-[10px] truncate flex-1 opacity-80">{seg.partOfSpeech}</span>
                    <span className="text-[9px] opacity-70 flex-shrink-0">{seg.morphType}</span>
                    {segRoot && (
                      <span className="arabic text-[10px] opacity-80">ج:{segRoot}</span>
                    )}
                    {segLemma && (
                      <span className="arabic text-[10px] opacity-80">ل:{segLemma}</span>
                    )}
                  </div>
                );
              })}
            </div>

            {/* Per-segment detail rows */}
            <p className="text-[10px] font-semibold text-muted uppercase tracking-[0.18em] mt-4 mb-2">
              Per-segment details
            </p>
            <div className="space-y-1.5">
              {word.segments.map((seg, i) => {
                const segRoot = buckwalterToArabic(seg.root);
                const segLemma = buckwalterToArabic(seg.lemma);
                return (
                  <div key={i} className="rounded-md border border-border/50 p-2 text-xs">
                    {seg.gloss && (
                      <div className="italic text-muted-foreground mb-1">{seg.gloss}</div>
                    )}
                    <div className="grid grid-cols-2 gap-x-3 gap-y-0.5">
                      {seg.syntacticRole && (
                        <Field label="Role" value={seg.syntacticRole} sub={translate(seg.syntacticRole)} />
                      )}
                      {seg.partOfSpeech && (
                        <Field label="POS" value={seg.partOfSpeech} sub={translate(seg.partOfSpeech)} />
                      )}
                      {seg.case && (
                        <Field label="Case" value={seg.case} sub={translate(seg.case)} />
                      )}
                      {seg.gender && <Field label="Gender" value={seg.gender} />}
                      {segRoot && <Field label="Root" value={segRoot} isArabic />}
                      {segLemma && <Field label="Lemma" value={segLemma} isArabic />}
                      {seg.possessiveConstruct && (
                        <Field label="Construct" value={seg.possessiveConstruct} />
                      )}
                      {seg.invariableDeclinable && (
                        <Field label="Decl." value={seg.invariableDeclinable} />
                      )}
                      {seg.phrase && (
                        <Field label="Phrase" value={seg.phrase} sub={seg.phrasalFunction} />
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Footer */}
          <div className="px-5 py-2 border-t border-border bg-card-hover text-[10px] text-muted flex items-center justify-between font-serif">
            <span>
              Click <span className="font-mono text-foreground/70">Root</span> chip above for full root meaning
            </span>
            <span>
              <kbd className="px-1.5 py-0.5 bg-background border border-border rounded">Esc</kbd> to close
            </span>
          </div>
        </div>
      </div>
      {rootModal && (
        <RootMeaningModal
          root={rootModal}
          visible={true}
          onClose={() => setRootModal(null)}
        />
      )}
    </>
  );
}

function Field({
  label,
  value,
  sub,
  isArabic,
}: {
  label: string;
  value: string;
  sub?: string;
  isArabic?: boolean;
}) {
  return (
    <div className="flex items-baseline gap-1.5">
      <span className="text-muted w-14 flex-shrink-0">{label}</span>
      <span className="font-medium text-foreground truncate">{value}</span>
      {isArabic ? (
        <span className="arabic text-base text-muted">{value}</span>
      ) : sub ? (
        <span className="text-muted-light truncate">{sub}</span>
      ) : null}
    </div>
  );
}