/**
 * Small general-purpose utilities used across the app.
 * Larger domain-specific helpers live in their own modules:
 *   - lib/arabic.ts    — Arabic text helpers
 *   - lib/quran-meta.ts — Quran page/juz lookups
 *   - lib/scroll-lock.ts — body scroll lock
 *   - lib/translate.ts — POS/role translation lookups
 */

import { normalizeArabic } from "./arabic";

/** Format surah:verse reference (e.g., "2:255") */
export function formatReference(surahNumber: number, verseNumber: number): string {
  return `${surahNumber}:${verseNumber}`;
}

/** Parse surah:verse reference string. */
export function parseReference(ref: string): { surahNumber: number; verseNumber: number } | null {
  const match = ref.match(/^(\d+):(\d+)$/);
  if (!match) return null;
  return {
    surahNumber: parseInt(match[1], 10),
    verseNumber: parseInt(match[2], 10),
  };
}

/** Re-export for backward compatibility — prefer importing from `@/lib/arabic`. */
export { normalizeArabic };