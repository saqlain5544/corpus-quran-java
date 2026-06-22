/**
 * Unit test: verify lock/unlock preserves page state.
 * Run: npx tsx scripts/test-scroll-lock.ts
 *
 * Mocks the minimal DOM API the module uses (body.style, innerWidth,
 * clientWidth) — no jsdom needed.
 */
/* eslint-disable no-console */

// ── Minimal DOM mock ────────────────────────────────────────────

function makeBody(): HTMLBodyElement {
  const style: Record<string, string> = {};
  return {
    style: new Proxy(style, {
      set: (target, prop: string, value: string) => {
        target[prop] = value;
        return true;
      },
      get: (target, prop: string) => target[prop] ?? "",
    }),
  } as unknown as HTMLBodyElement;
}

let mockBody = makeBody();
const mockDocElement: { clientWidth: number } = { clientWidth: 1007 };
let mockInnerWidth = 1024; // 1024 - 1007 = 17px scrollbar

// The module reads `document.body`, `document.documentElement`, and
// `window.innerWidth`. We need a real DOM-like object as `document`,
// plus `window` exposing `innerWidth`.
const mockDocument = {
  get body() { return mockBody; },
  get documentElement() { return mockDocElement; },
};
(globalThis as unknown as { window: object }).window = {
  get innerWidth() { return mockInnerWidth; },
  get document() { return mockDocument; },
};
(globalThis as unknown as { document: object }).document = mockDocument;

// ── Run tests ───────────────────────────────────────────────────

async function main() {
  const mod = await import("../src/lib/scroll-lock");
  const { lockBodyScroll, unlockBodyScroll, _resetScrollLockForTesting } = mod;

  let passed = 0;
  let failed = 0;
  const assert = (cond: boolean, label: string) => {
    if (cond) { passed++; console.log(`  ✓ ${label}`); }
    else { failed++; console.log(`  ✗ ${label}`); }
  };

  const snapshot = () => ({
    overflow: mockBody.style.overflow,
    paddingRight: mockBody.style.paddingRight,
    position: mockBody.style.position,
    top: mockBody.style.top,
    width: mockBody.style.width,
  });

  // ── Test 1: single lock/unlock cycle ──────────────────────────

  console.log("\n── Test 1: single lock/unlock cycle ──");
  _resetScrollLockForTesting();
  mockBody = makeBody();
  const before = snapshot();
  lockBodyScroll();
  const locked = snapshot();
  unlockBodyScroll();
  const after = snapshot();

  assert(locked.overflow === "hidden", "overflow becomes hidden");
  assert(locked.paddingRight === "17px",
    `paddingRight compensates for 17px scrollbar (got: "${locked.paddingRight}")`);
  assert(locked.position === "", "position NOT set to fixed (page stays in flow)");
  assert(locked.top === "", "top NOT set to negative (no scrollTo needed)");
  assert(after.overflow === before.overflow,
    `overflow restored to original (got: "${after.overflow}")`);
  assert(after.paddingRight === before.paddingRight,
    `paddingRight restored to original (got: "${after.paddingRight}")`);

  // ── Test 2: nested modals ─────────────────────────────────────

  console.log("\n── Test 2: nested modals ──");
  _resetScrollLockForTesting();
  mockBody = makeBody();
  lockBodyScroll();   // Modal A
  lockBodyScroll();   // Modal B (over A)
  const nestedLocked = snapshot();
  unlockBodyScroll(); // Close B — should NOT release
  const stillLocked = snapshot();
  unlockBodyScroll(); // Close A — should release
  const unlocked = snapshot();

  assert(nestedLocked.overflow === "hidden", "nested: still hidden while both open");
  assert(stillLocked.overflow === "hidden", "after closing B, body stays locked (A still open)");
  assert(unlocked.overflow === "", "after closing A, body is unlocked");

  // ── Test 3: no scrollbar (mobile / short content) ───────────

  console.log("\n── Test 3: no scrollbar (mobile / short content) ──");
  _resetScrollLockForTesting();
  mockBody = makeBody();
  mockInnerWidth = 800;
  mockDocElement.clientWidth = 800;
  lockBodyScroll();
  const mobileLocked = snapshot();
  unlockBodyScroll();
  const mobileAfter = snapshot();

  assert(mobileLocked.overflow === "hidden", "overflow still hidden on mobile");
  assert(mobileLocked.paddingRight === "" || mobileLocked.paddingRight === "0px",
    `no padding-right when no scrollbar (got: "${mobileLocked.paddingRight}")`);
  assert(mobileAfter.overflow === "", "mobile: cleaned up correctly");

  // ── Test 4: pre-existing inline styles preserved ─────────────

  console.log("\n── Test 4: pre-existing inline styles preserved ──");
  _resetScrollLockForTesting();
  mockBody = makeBody();
  mockBody.style.paddingRight = "20px"; // page already had padding
  lockBodyScroll();
  unlockBodyScroll();
  assert(mockBody.style.paddingRight === "20px",
    "pre-existing padding-right preserved after unlock");

  // ── Done ─────────────────────────────────────────────────────

  console.log(`\n── Results: ${passed} passed, ${failed} failed ──`);
  process.exit(failed > 0 ? 1 : 0);
}

main().catch((e) => { console.error(e); process.exit(1); });
