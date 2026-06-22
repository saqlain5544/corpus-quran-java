"use client";

import { useEffect, useRef, useState } from "react";
import { translate, translateShort } from "@/lib/translate";

interface WordTooltipProps {
  /** The word span being hovered — its data-* attrs are the source of truth. */
  anchorEl: HTMLElement | null;
  onLeave: () => void;
  onTooltipEnter?: () => void;
}

/**
 * Compact floating tooltip.
 *
 * The hovered word span carries all metadata as `data-*` attributes
 * (set by the SSR VerseView). The tooltip reads them directly — no
 * server call, no client-side tokenization, no prop drilling.
 *
 * Required attrs on the anchor element:
 *   data-token        Arabic text with diacritics
 *   data-translation  English translation (Word.translation)
 *   data-gloss        Stem-segment gloss
 *   data-pos          Part of speech
 *   data-role         Syntactic role
 *   data-case         Grammatical case
 *   data-root         Arabic root letters
 *   data-lemma        Arabic lemma
 */
export function WordTooltip({ anchorEl, onLeave, onTooltipEnter }: WordTooltipProps) {
  const tipRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState<{
    top: number;
    left: number;
    above: boolean;
  } | null>(null);

  useEffect(() => {
    if (!anchorEl) {
      setPos(null);
      return;
    }
    const recalc = () => {
      if (!anchorEl || !tipRef.current) return;
      const a = anchorEl.getBoundingClientRect();
      const t = tipRef.current.getBoundingClientRect();
      const gap = 10;
      const vw = window.innerWidth;
      const vh = window.innerHeight;

      const cx = a.left + a.width / 2;
      let left = cx - t.width / 2;
      left = Math.max(8, Math.min(left, vw - t.width - 8));

      const above =
        a.top - gap - t.height >= 0 || a.top - gap > vh - a.bottom - gap;
      const top = above ? a.top - t.height - gap : a.bottom + gap;
      setPos({ top, left, above });
    };

    const raf = requestAnimationFrame(() => {
      requestAnimationFrame(recalc);
    });

    const onScroll = () => recalc();
    window.addEventListener("scroll", onScroll, { passive: true });
    window.addEventListener("resize", onScroll);
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("scroll", onScroll);
      window.removeEventListener("resize", onScroll);
    };
  }, [anchorEl]);

  if (!anchorEl) return null;
  const above = pos?.above ?? true;

  // Read everything from the DOM — no props, no closure.
  const token = anchorEl.dataset.token ?? "";
  const translation = anchorEl.dataset.translation ?? "";
  const gloss = anchorEl.dataset.gloss ?? "";
  const pos_ = anchorEl.dataset.pos ?? "";
  const role = anchorEl.dataset.role ?? "";
  const root = anchorEl.dataset.root ?? "";
  const lemma = anchorEl.dataset.lemma ?? "";
  const case_ = anchorEl.dataset.case ?? "";

  return (
    <div
      ref={tipRef}
      className="fixed z-100 tooltip-enter pointer-events-auto"
      style={{
        top: pos?.top ?? -9999,
        left: pos?.left ?? -9999,
        visibility: pos ? "visible" : "hidden",
        opacity: pos ? 1 : 0,
        transition: "opacity 120ms ease-out",
        maxWidth: "min(320px, calc(100vw - 32px))",
      }}
      onMouseLeave={onLeave}
      onMouseEnter={onTooltipEnter}
    >
      {/* Arrow */}
      <div
        className={`absolute left-1/2 -translate-x-1/2 w-3 h-3 bg-card border-border rotate-45 ${
          above
            ? "-bottom-1.5 border-r border-b border-t-transparent border-l-transparent"
            : "-top-1.5 border-l border-t border-b-transparent border-r-transparent"
        }`}
      />

      {/* Card */}
      <div className="bg-card border border-border rounded-xl shadow-xl backdrop-blur-md overflow-hidden">
        {/* Arabic word */}
        <div className="px-5 py-3 bg-quran-bg text-center border-b border-border/50">
          <p className="arabic text-2xl text-quran-text leading-relaxed" dir="rtl">
            {token}
          </p>
          {(translation || gloss) && (
            <p className="text-xs text-muted italic mt-1 line-clamp-2">
              {translation || gloss}
            </p>
          )}
        </div>

        {/* Attributes */}
        <div className="px-4 py-3 space-y-1.5">
          {pos_ && pos_ !== "?" && (
            <Row
              label="POS"
              value={translateShort(pos_)}
              color="blue"
              title={translate(pos_)}
            />
          )}
          {role && role !== "?" && (
            <Row
              label="Role"
              value={translateShort(role)}
              color="purple"
              title={translate(role)}
            />
          )}
          {root && (
            <Row
              label="Root"
              value={root}
              color="emerald"
              title="Open root meaning"
              isArabic
            />
          )}
          {lemma && (
            <Row label="Lemma" value={lemma} color="amber" isArabic />
          )}
          {case_ && (
            <Row
              label="Case"
              value={translateShort(case_)}
              color="teal"
              title={translate(case_)}
            />
          )}
        </div>

        {/* Hint */}
        <div className="px-4 py-2 border-t border-border/50 bg-card-hover text-[10px] text-muted flex items-center justify-center gap-1.5">
          <svg
            className="w-3 h-3"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"
            />
          </svg>
          Click for full analysis
        </div>
      </div>
    </div>
  );
}

function Row({
  label,
  value,
  color,
  title,
  isArabic,
}: {
  label: string;
  value: string;
  color: "blue" | "purple" | "emerald" | "amber" | "teal";
  title?: string;
  isArabic?: boolean;
}) {
  const colors: Record<string, string> = {
    blue: "bg-blue-500",
    purple: "bg-purple-500",
    emerald: "bg-emerald-500",
    amber: "bg-amber-500",
    teal: "bg-teal-500",
  };
  return (
    <div className="flex items-center gap-2 text-xs" title={title}>
      <span
        className={`w-1.5 h-1.5 rounded-full flex-shrink-0 ${colors[color]}`}
      />
      <span className="text-muted w-10 flex-shrink-0">{label}</span>
      <span
        className={`font-medium text-foreground truncate ${
          isArabic ? "arabic text-base" : ""
        }`}
        dir={isArabic ? "rtl" : undefined}
      >
        {value}
      </span>
    </div>
  );
}