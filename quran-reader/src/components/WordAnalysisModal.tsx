"use client";

import { useEffect, useState } from "react";
import { RootMeaningModal } from "@/components/RootMeaningModal";
import { translate, translateShort } from "@/lib/translate";
import { lockBodyScroll, unlockBodyScroll } from "@/lib/scroll-lock";
import { buckwalterToArabic } from "@/lib/arabic";
import type { WordData, Segment } from "@/lib/types";

interface Props {
  word: WordData;
  verseNumber: number;
  visible: boolean;
  onClose: () => void;
}

type SegmentLike = Pick<
  Segment,
  "text" | "partOfSpeech" | "morphType" | "gloss" | "case" | "gender" | "syntacticRole" | "lemma" | "root" | "possessiveConstruct" | "invariableDeclinable" | "phrase" | "phrasalFunction"
>;

type MorphColor = "sky" | "emerald" | "amber" | "violet";

const MORPH_COLORS: Record<MorphColor, { bg: string; text: string; label: string }> = {
  sky: { bg: "bg-sky-100 dark:bg-sky-900/30", text: "text-sky-700 dark:text-sky-400", label: "Pref" },
  emerald: { bg: "bg-emerald-100 dark:bg-emerald-900/30", text: "text-emerald-700 dark:text-emerald-400", label: "Stem" },
  amber: { bg: "bg-amber-100 dark:bg-amber-900/30", text: "text-amber-700 dark:text-amber-400", label: "Suff" },
  violet: { bg: "bg-violet-100 dark:bg-violet-900/30", text: "text-violet-700 dark:text-violet-400", label: "iʻrāb" },
};

function colorFor(seg: SegmentLike): MorphColor {
  if (seg.morphType === "Prefix") return "sky";
  if (seg.morphType === "Stem") return "emerald";
  if (seg.morphType === "Suffix") return "amber";
  return "violet";
}

/**
 * Detailed word analysis modal.
 *
 * Composed of small, single-purpose pieces:
 *   - `ModalHeader`  — Arabic word + chips
 *   - `MorphemesList` — prefix/stem/suffix breakdown (flex-wrap chips)
 *   - `SegmentDetailsList` — per-segment metadata (semantic <dl>, wraps)
 *   - `ModalFooter`  — keyboard hint
 *
 * Each piece is responsible for its own layout so that long morphology
 * codes (e.g. "IMPERF_PREF", "INVARIABLE", Arabic roots) wrap onto the
 * next line instead of being clipped.
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

  const uniqueGlosses = [
    ...new Set(
      word.segments
        .map((s) => s.gloss)
        .filter((g): g is string => !!g && g !== "null" && g !== "undefined")
    ),
  ];
  const dedupedGloss = uniqueGlosses.join(" · ");

  const stem =
    word.segments.find((s) => s.syntacticRole) ||
    word.segments.find((s) => s.morphType === "Stem") ||
    word.segments[word.segments.length - 1];

  const wordRoot = buckwalterToArabic(word.segments.find((s) => s.root)?.root);
  const wordLemma = buckwalterToArabic(word.segments.find((s) => s.lemma)?.lemma);

  return (
    <>
      <div
        role="dialog"
        aria-modal="true"
        aria-label={`Word analysis: ${word.token}`}
        className="fixed inset-0 z-50 flex items-center justify-center bg-foreground/20 backdrop-blur-sm modal-enter p-4"
        onClick={onClose}
      >
        <div
          className="bg-card border border-border rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col overflow-hidden font-serif"
          onClick={(e) => e.stopPropagation()}
        >
          <ModalHeader
            word={word}
            verseNumber={verseNumber}
            dedupedGloss={dedupedGloss}
            stemPartOfSpeech={stem?.partOfSpeech}
            wordRoot={wordRoot}
            wordLemma={wordLemma}
            onClose={onClose}
            onRootClick={setRootModal}
          />

          <div className="px-5 py-4 overflow-auto flex-1 space-y-5">
            <MorphemesList segments={word.segments} />
            <SegmentDetailsList segments={word.segments} />
          </div>

          <ModalFooter />
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

// ── Sub-components ──────────────────────────────────────────────

function ModalHeader({
  word,
  verseNumber,
  dedupedGloss,
  stemPartOfSpeech,
  wordRoot,
  wordLemma,
  onClose,
  onRootClick,
}: {
  word: WordData;
  verseNumber: number;
  dedupedGloss: string;
  stemPartOfSpeech?: string;
  wordRoot: string;
  wordLemma: string;
  onClose: () => void;
  onRootClick: (root: string) => void;
}) {
  return (
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

      <p className="text-[10px] uppercase tracking-[0.18em] text-muted font-serif">
        Verse {verseNumber} · Word {word.wordNumber}
      </p>
      <p
        className="arabic text-4xl text-quran-text leading-tight mt-1 mb-1 break-words"
        dir="rtl"
      >
        {word.token}
      </p>
      {word.withoutDiacritics !== word.token && (
        <p className="arabic text-sm text-muted-light break-words" dir="rtl">
          {word.withoutDiacritics}
        </p>
      )}
      {dedupedGloss && (
        <p className="text-sm text-muted italic mt-1.5 break-words">
          {dedupedGloss}
        </p>
      )}

      <div className="flex flex-wrap items-center gap-1.5 mt-3">
        {stemPartOfSpeech && (
          <Pill label="POS" value={stemPartOfSpeech} color="sky" />
        )}
        {wordRoot && (
          <button
            onClick={() => onRootClick(wordRoot)}
            className="inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full bg-emerald-100 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 font-medium hover:opacity-80"
            title="Open root meaning"
          >
            <span>Root</span>
            <span className="arabic" dir="rtl">{wordRoot}</span>
          </button>
        )}
        {wordLemma && (
          <Pill label="Lemma" value={wordLemma} color="amber" isArabic />
        )}
      </div>
    </div>
  );
}

function MorphemesList({ segments }: { segments: SegmentLike[] }) {
  return (
    <section>
      <SectionTitle>Morphemes ({segments.length})</SectionTitle>
      <div className="space-y-1.5">
        {segments.map((seg, i) => (
          <MorphemeChip key={i} seg={seg} />
        ))}
      </div>
    </section>
  );
}

function MorphemeChip({ seg }: { seg: SegmentLike }) {
  const c = MORPH_COLORS[colorFor(seg)];
  const isNullMorph = !seg.text || seg.text === "(null)";
  const segRoot = buckwalterToArabic(seg.root);
  const segLemma = buckwalterToArabic(seg.lemma);

  return (
    <div
      className={`rounded-md border px-2 py-1.5 text-xs flex flex-wrap items-baseline gap-x-2 gap-y-1 ${c.bg} ${c.text}`}
    >
      {isNullMorph ? (
        <span className="text-[10px] italic opacity-70 min-w-[3rem]">iʻrāb</span>
      ) : (
        <span className="arabic text-base font-semibold break-words" dir="rtl">
          {seg.text}
        </span>
      )}
      <span className="text-[10px] opacity-80 flex-1 min-w-[6rem] break-words">
        {seg.partOfSpeech}
      </span>
      <span className="text-[9px] opacity-70 flex-shrink-0">{seg.morphType}</span>
      {segRoot && (
        <span className="arabic text-[10px] opacity-80" dir="rtl">
          ج:{segRoot}
        </span>
      )}
      {segLemma && (
        <span className="arabic text-[10px] opacity-80" dir="rtl">
          ل:{segLemma}
        </span>
      )}
    </div>
  );
}

function SegmentDetailsList({ segments }: { segments: SegmentLike[] }) {
  return (
    <section>
      <SectionTitle>Per-segment details</SectionTitle>
      <div className="space-y-2">
        {segments.map((seg, i) => (
          <SegmentDetail key={i} seg={seg} />
        ))}
      </div>
    </section>
  );
}

/**
 * One per-segment detail card.
 *
 * Uses semantic `<dl>` / `<dt>` / `<dd>` so screen readers announce
 * the label/value relationship. Layout is a two-column grid where the
 * first column auto-sizes to the widest label and the second column
 * takes the rest — and `break-words` allows the value to wrap when it
 * is wider than the cell.
 */
function SegmentDetail({ seg }: { seg: SegmentLike }) {
  const segRoot = buckwalterToArabic(seg.root);
  const segLemma = buckwalterToArabic(seg.lemma);

  const rows: Array<[string, string, { isArabic?: boolean; sub?: string }]> = [];
  if (seg.syntacticRole)
    rows.push(["Role", seg.syntacticRole, { sub: translate(seg.syntacticRole) }]);
  if (seg.partOfSpeech)
    rows.push(["POS", seg.partOfSpeech, { sub: translate(seg.partOfSpeech) }]);
  if (seg.case) rows.push(["Case", seg.case, { sub: translate(seg.case) }]);
  if (seg.gender) rows.push(["Gender", seg.gender, {}]);
  if (segRoot) rows.push(["Root", segRoot, { isArabic: true }]);
  if (segLemma) rows.push(["Lemma", segLemma, { isArabic: true }]);
  if (seg.possessiveConstruct) rows.push(["Construct", seg.possessiveConstruct, {}]);
  if (seg.invariableDeclinable) rows.push(["Decl.", seg.invariableDeclinable, {}]);
  if (seg.phrase) rows.push(["Phrase", seg.phrase, { sub: seg.phrasalFunction }]);

  return (
    <div className="rounded-md border border-border/50 p-2.5 text-xs space-y-1.5">
      {seg.gloss && (
        <p className="italic text-muted-foreground break-words">{seg.gloss}</p>
      )}
      {rows.length > 0 && (
        <dl className="grid grid-cols-[max-content_1fr] gap-x-3 gap-y-1 items-baseline">
          {rows.map(([label, value, opts]) => (
            <DetailRow
              key={label}
              label={label}
              value={value}
              isArabic={opts.isArabic}
              sub={opts.sub}
            />
          ))}
        </dl>
      )}
    </div>
  );
}

function DetailRow({
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
    <>
      <dt className="text-muted font-medium">{label}</dt>
      <dd className="text-foreground break-words min-w-0">
        <span className={isArabic ? "arabic" : ""} dir={isArabic ? "rtl" : undefined}>
          {value}
        </span>
        {sub && !isArabic && <span className="text-muted-light"> — {sub}</span>}
      </dd>
    </>
  );
}

function ModalFooter() {
  return (
    <div className="px-5 py-2 border-t border-border bg-card-hover text-[10px] text-muted flex items-center justify-between font-serif">
      <span>
        Click <span className="font-mono text-foreground/70">Root</span> chip above for full root meaning
      </span>
      <span>
        <kbd className="px-1.5 py-0.5 bg-background border border-border rounded">Esc</kbd> to close
      </span>
    </div>
  );
}

function SectionTitle({ children }: { children: React.ReactNode }) {
  return (
    <p className="text-[10px] font-semibold text-muted uppercase tracking-[0.18em] mb-2 font-serif">
      {children}
    </p>
  );
}

function Pill({
  label,
  value,
  color,
  isArabic,
}: {
  label: string;
  value: string;
  color: "sky" | "amber";
  isArabic?: boolean;
}) {
  const cls =
    color === "sky"
      ? "bg-sky-100 dark:bg-sky-900/30 text-sky-700 dark:text-sky-400"
      : "bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400";
  return (
    <span
      className={`inline-flex items-center gap-1.5 text-xs px-2.5 py-1 rounded-full font-medium ${cls}`}
    >
      <span>{label}</span>
      <span className={isArabic ? "arabic" : ""} dir={isArabic ? "rtl" : undefined}>
        {value}
      </span>
    </span>
  );
}