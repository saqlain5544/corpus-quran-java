"use client";

import { useLayoutEffect, useRef, useState } from "react";

export interface FloatingPosition {
  /** Viewport-coordinate Y of the floating element's top edge. */
  top: number;
  /** Viewport-coordinate X of the floating element's left edge. */
  left: number;
  /** X coordinate of the arrow tip, relative to the floating element's left edge. */
  arrowX: number;
  /** Whether the floating element sits above or below the anchor. */
  placement: "above" | "below";
}

interface UsePositionOptions {
  gap?: number;
  viewportPadding?: number;
}

/**
 * Compute and track the position of a floating element (e.g. tooltip)
 * anchored to another DOM element. Both are expected to use
 * `position: fixed`, so the returned coordinates are in viewport space.
 *
 * Implementation notes:
 * - `useLayoutEffect` (not `useEffect`) so position is set **before** the
 *   browser paints — prevents the "flash at top-left" you get when a
 *   fixed element first renders without a position.
 * - Re-measures on `scroll` (capture-phase, so we hear scrolls on any
 *   ancestor container) and `resize`.
 * - Picks `placement: "above"` when the anchor is high enough on screen,
 *   `"below"` otherwise.
 * - Horizontally clamps to the viewport so the floating element never
 *   overflows the left/right edges; computes `arrowX` so the arrow
 *   stays pointing at the anchor's center even after clamping.
 * - Skips the `setPos` call when nothing changed, to avoid re-renders
 *   during scroll storms.
 */
export function usePosition(
  anchorEl: HTMLElement | null,
  floatingEl: HTMLElement | null,
  { gap = 8, viewportPadding = 8 }: UsePositionOptions = {}
): FloatingPosition | null {
  const [pos, setPos] = useState<FloatingPosition | null>(null);
  const posRef = useRef<FloatingPosition | null>(null);

  useLayoutEffect(() => {
    if (!anchorEl || !floatingEl) {
      posRef.current = null;
      setPos(null);
      return;
    }

    const measure = () => {
      if (!anchorEl || !floatingEl) return;

      const a = anchorEl.getBoundingClientRect();
      const f = floatingEl.getBoundingClientRect();
      const vw = window.innerWidth;
      const vh = window.innerHeight;

      const spaceAbove = a.top;
      const spaceBelow = vh - a.bottom;
      const placement: "above" | "below" =
        spaceAbove >= f.height + gap ? "above" : "below";

      const top =
        placement === "above"
          ? a.top - f.height - gap
          : a.bottom + gap;

      const anchorCenterX = a.left + a.width / 2;
      // Clamp the tooltip horizontally so it never overflows the viewport.
      // If the tooltip is wider than the viewport, prefer centering it
      // and let the caller add their own horizontal scroll if needed.
      let left = anchorCenterX - f.width / 2;
      if (f.width >= vw - viewportPadding * 2) {
        left = viewportPadding;
      } else {
        left = Math.max(viewportPadding, Math.min(left, vw - f.width - viewportPadding));
      }

      const arrowX = anchorCenterX - left;

      const next: FloatingPosition = { top, left, arrowX, placement };
      const cur = posRef.current;
      if (
        cur &&
        cur.top === next.top &&
        cur.left === next.left &&
        cur.arrowX === next.arrowX &&
        cur.placement === next.placement
      ) {
        return;
      }
      posRef.current = next;
      setPos(next);
    };

    measure();

    const onScrollOrResize = () => measure();
    window.addEventListener("scroll", onScrollOrResize, { passive: true, capture: true });
    window.addEventListener("resize", onScrollOrResize);

    return () => {
      window.removeEventListener("scroll", onScrollOrResize, { capture: true });
      window.removeEventListener("resize", onScrollOrResize);
    };
  }, [anchorEl, floatingEl, gap, viewportPadding]);

  return pos;
}