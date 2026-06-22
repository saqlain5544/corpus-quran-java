/**
 * Barrel export for the @/lib module.
 * Import from "@/lib" to get the public surface.
 */

// Core types
export type {
  Segment,
  WordData,
  VerseData,
  SurahData,
  TooltipWord,
} from "./types";

// Arabic utilities
export {
  toArabicNumeral,
  buckwalterToArabic,
  getWordTranslation,
  normalizeArabic,
  isArabic,
} from "./arabic";

// Tokenizer (index-based alignment)
export {
  tokenize,
  align,
  normalize,
  stripDiacritics,
  type QuranToken,
  type QuranTokenType,
  type DBWord,
  type AlignmentEntry,
} from "./tokenizer";

// Quran metadata
export { getPageNumber, getJuzNumber } from "./quran-meta";

// General utilities
export { formatReference, parseReference } from "./utils";