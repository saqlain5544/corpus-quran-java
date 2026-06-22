/**
 * Body scroll lock — prevents the page from scrolling behind a modal.
 *
 * Design: the page is NOT moved out of flow. Instead, `body { overflow: hidden }`
 * blocks scrolling, and a `padding-right` equal to the scrollbar width prevents
 * the visible layout shift that would otherwise occur when the scrollbar
 * disappears. The page's natural scroll position is preserved — no `scrollTo`
 * restoration needed, so opening/closing a modal causes no visible jump.
 *
 * Nested modals are supported via a reference counter; only the first lock
 * applies styles, only the last unlock removes them.
 *
 * Server-safe: bails out when `window` is undefined.
 */

let lockCount = 0;
let savedStyles: { overflow: string; paddingRight: string } | null = null;

export function lockBodyScroll(): void {
  if (typeof window === "undefined") return;
  if (lockCount === 0) {
    // Snapshot the current inline styles so we restore them exactly on unlock.
    savedStyles = {
      overflow: document.body.style.overflow,
      paddingRight: document.body.style.paddingRight,
    };
    // Compensate for the scrollbar that disappears when overflow becomes hidden.
    // innerWidth includes the scrollbar; clientWidth doesn't.
    const scrollbarWidth = window.innerWidth - document.documentElement.clientWidth;
    document.body.style.overflow = "hidden";
    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${scrollbarWidth}px`;
    }
  }
  lockCount++;
}

export function unlockBodyScroll(): void {
  if (typeof window === "undefined") return;
  lockCount = Math.max(0, lockCount - 1);
  if (lockCount === 0 && savedStyles) {
    document.body.style.overflow = savedStyles.overflow;
    document.body.style.paddingRight = savedStyles.paddingRight;
    savedStyles = null;
  }
}

/** Test helper — reset internal state (e.g. for HMR). */
export function _resetScrollLockForTesting(): void {
  lockCount = 0;
  savedStyles = null;
}
