/**
 * Arabic text utilities used by the surah reader UI.
 */

// ── Number conversion ──────────────────────────────────────────

/** Convert a number to its Arabic-Indic representation (e.g. 12 → ١٢). */
export function toArabicNumeral(n: number): string {
  const arabicDigits = "٠١٢٣٤٥٦٧٨٩";
  return String(n).split("").map((d) => arabicDigits[parseInt(d, 10)] || d).join("");
}

// ── Pause / word classification ─────────────────────────────────
//
// The old regex-based isPauseOnly was buggy:
//   - Couldn't handle words split across whitespace (أَيْنَ + مَا = أَيْنَمَا)
//   - Anything lacking Arabic letters was labeled a "pause mark"
// The replacement is in lib/tokenizer.ts: a simple whitespace-split
// tokenizer with index-based alignment to the DB word list.

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

// ── Buckwalter transliteration → Arabic script ────────────────
//
// The DB stores Quranic roots in Buckwalter transliteration (ASCII letters
// like "$Am" for ش م). For display in tooltip/modal we want the Arabic
// script the user actually sees on tanzil.net.
//
// Standard Buckwalter mapping (the 28 Arabic letters + diacritics):

const BUCKWALTER_MAP: Record<string, string> = {
  "'": "ء", "|": "ء",           // hamza
  ">": "أ", "&": "أ",            // alef + hamza above
  "<": "إ", "}": "إ",            // alef + hamza below (and alif wasla in some dialects)
  "A": "آ",                       // alef + madda
  "p": "ة",                       // ta marbuta
  "H": "ح",                       // ha (voiceless pharyngeal fricative)
  "x": "خ",                       // kha
  "$": "ش", "c": "ش",            // shin (both spellings seen in the wild)
  "*": "ذ",                       // dhal (alternate)
  "v": "ث",                       // tha
  "Y": "ى",                       // alif maqsura
  "b": "ب", "t": "ت", "j": "ج", "d": "د", "r": "ر", "z": "ز",
  "s": "س", "S": "ص", "D": "ض", "T": "ط", "Z": "ظ",
  "E": "ع", "g": "غ", "f": "ف", "q": "ق", "k": "ك", "l": "ل",
  "m": "م", "n": "ن", "h": "ه", "w": "و", "y": "ي",
};

/**
 * Convert a Buckwalter transliteration string to its Arabic-script
 * equivalent. If the string is already in Arabic script (no Latin
 * letters mapped), it is returned unchanged. Unknown Latin chars
 * (typically English, e.g. in translations) pass through as-is.
 */
export function buckwalterToArabic(s: string | null | undefined): string {
  if (!s) return "";
  let out = "";
  for (const ch of s) {
    out += BUCKWALTER_MAP[ch] ?? ch;
  }
  return out;
}

// ── Normalization ──────────────────────────────────────────────

/** Normalize Arabic text for search (strip diacritics, tatweel, etc.) */
export function normalizeArabic(text: string): string {
  return text
    .replace(/[\u064B-\u065F\u0610-\u061A\u06D6-\u06ED]/g, "") // diacritics
    .replace(/[\u0640]/g, "")                                    // tatweel
    .replace(/[\u0670]/g, "")                                    // dagger alif
    .replace(/ٱ/g, "ا")                                          // alif wasla
    .trim();
}

/** Check if a string contains Arabic characters. */
export function isArabic(text: string): boolean {
  return /[\u0600-\u06FF\u0750-\u077F]/.test(text);
}

// ── Translation helper ─────────────────────────────────────────

/**
 * Extract a clean English gloss from a word's segments for display.
 * Deduplicates segment glosses and filters out noise tokens.
 */
export function getWordTranslation(w: {
  translation: string;
  segments: Array<{ gloss?: string }>;
}): string {
  const unique = [...new Set(
    w.segments
      .map((s) => s.gloss)
      .filter((g): g is string => !!g && g !== "null" && g !== "undefined")
  )];
  return unique.length > 0 ? unique.join(" ") : w.translation;
}