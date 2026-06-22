"use client";

import { useRef } from "react";
import { usePosition } from "@/hooks/usePosition";
import { translate, translateShort } from "@/lib/translate";

interface WordTooltipProps {
  /** The `.quran-word` span being hovered — its `data-*` attrs are the
   *  source of truth for the tooltip content. */
  anchorEl: HTMLElement | null;
  onLeave: () => void;
  onTooltipEnter?: () => void;
}

const VIEWPORT_PAD = 8;
const TOOLTIP_GAP = 8;
const COLOR_DOT: Record<string, string> = {
  blue: "bg-blue-500",
  purple: "bg-purple-500",
  emerald: "bg-emerald-500",
  amber: "bg-amber-500",
  teal: "bg-teal-500",
};

/**
 * Floating tooltip that reads word metadata from the hovered span's
 * `data-*` attributes and positions itself next to it.
 *
 * Positioning is delegated to `usePosition` (a layout-effect hook that
 * measures the anchor + the tooltip, picks above/below, clamps to the
 * viewport, and tracks scroll/resize).
 *
 * The tooltip is **always** rendered (so the layout-effect can measure
 * it); visibility is driven by `anchorEl` and `pos` via opacity +
 * visibility. This avoids a mount/unmount flicker on every hover.
 */
export function WordTooltip({ anchorEl, onLeave, onTooltipEnter }: WordTooltipProps) {
  const tipRef = useRef<HTMLDivElement>(null);
  const pos = usePosition(anchorEl, tipRef.current, {
    gap: TOOLTIP_GAP,
    viewportPadding: VIEWPORT_PAD,
  });

  // Pull all metadata off the anchor's data-* attrs. QuranHtmlEnhancer
  // sets these on first mouseover from data-morphology JSON.
  const token = anchorEl?.dataset.token ?? "";
  const translation = anchorEl?.dataset.translation ?? "";
  const gloss = anchorEl?.dataset.gloss ?? "";
  const pos_ = anchorEl?.dataset.pos ?? "";
  const role = anchorEl?.dataset.role ?? "";
  const root = anchorEl?.dataset.root ?? "";
  const lemma = anchorEl?.dataset.lemma ?? "";
  const case_ = anchorEl?.dataset.case ?? "";

  const visible = anchorEl !== null && pos !== null;

  return (
    <div
      ref={tipRef}
      role="tooltip"
      aria-hidden={!visible}
      onMouseEnter={onTooltipEnter}
      onMouseLeave={onLeave}
      className="fixed z-50 pointer-events-auto transition-opacity duration-150"
      style={{
        top: pos?.top ?? -9999,
        left: pos?.left ?? -9999,
        opacity: visible ? 1 : 0,
        visibility: visible ? "visible" : "hidden",
        maxWidth: `min(320px, calc(100vw - ${VIEWPORT_PAD * 2}px))`,
      }}
    >
      {pos && <Arrow placement={pos.placement} anchorX={pos.arrowX} />}

      <div className="bg-card border border-border rounded-xl shadow-xl backdrop-blur-md overflow-hidden">
        {/* Arabic word */}
        <div className="px-5 py-3 bg-quran-bg text-center border-b border-border/50">
          <p className="arabic text-2xl text-quran-text leading-relaxed break-words" dir="rtl">
            {token}
          </p>
          {(translation || gloss) && (
            <p className="text-xs text-muted italic mt-1 break-words">
              {translation || gloss}
            </p>
          )}
        </div>

        {/* Attribute rows — each row wraps if the value is long */}
        <div className="px-4 py-3 space-y-1.5">
          {pos_ && pos_ !== "?" && (
            <Row
              label="POS"
              value={translateShort(pos_)}
              title={translate(pos_)}
              color="blue"
            />
          )}
          {role && role !== "?" && (
            <Row
              label="Role"
              value={translateShort(role)}
              title={translate(role)}
              color="purple"
            />
          )}
          {root && (
            <Row
              label="Root"
              value={root}
              title="Open root meaning"
              color="emerald"
              isArabic
            />
          )}
          {lemma && <Row label="Lemma" value={lemma} color="amber" isArabic />}
          {case_ && (
            <Row
              label="Case"
              value={translateShort(case_)}
              title={translate(case_)}
              color="teal"
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

function Arrow({ placement, anchorX }: { placement: "above" | "below"; anchorX: number }) {
  const isAbove = placement === "above";
  return (
    <div
      aria-hidden
      className={`absolute w-3 h-3 bg-card border-border ${
        isAbove ? "border-r border-b" : "border-l border-t"
      }`}
      style={{
        left: `${anchorX}px`,
        transform: "translateX(-50%) rotate(45deg)",
        [isAbove ? "bottom" : "top"]: "-7px",
      }}
    />
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
  color: keyof typeof COLOR_DOT;
  title?: string;
  isArabic?: boolean;
}) {
  return (
    <div className="flex items-baseline gap-2 text-xs" title={title}>
      <span className={`w-1.5 h-1.5 rounded-full flex-shrink-0 translate-y-[1px] ${COLOR_DOT[color]}`} />
      <span className="text-muted w-10 flex-shrink-0">{label}</span>
      <span
        className={`font-medium text-foreground break-words min-w-0 ${
          isArabic ? "arabic text-base" : ""
        }`}
        dir={isArabic ? "rtl" : undefined}
      >
        {value}
      </span>
    </div>
  );
}