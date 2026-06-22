/**
 * Shared Quran types — single source of truth for morphological data.
 * Used by data layer, parser, and all UI components.
 */

// ── Morphological segment ───────────────────────────────────────

export interface Segment {
  segmentNumber: number;
  text: string;
  /** Part of speech (PV, IV, CV, NOUN, PRON_3MP, REL_PRON, etc.) */
  partOfSpeech: string;
  /** Morpheme slot: Prefix, Stem, Suffix */
  morphType: string;
  lemma?: string;
  root?: string;
  gender?: string;
  /** Grammatical case: NOMINATIVE, ACCUSATIVE, GENITIVE, INVARIABLE */
  case?: string;
  /** Syntactic role: SUBJ, OBJ, ADJ, GEN_CONS, PREP_OBJ, AGNT, etc. */
  syntacticRole?: string;
  gloss?: string;
  /** CONSTRUCT (mudaf) or NOT_CONSTRUCT */
  possessiveConstruct?: string;
  caseMoodMarker?: string;
  invariableDeclinable?: string;
  /** PHRASE / NON_GOV_REL_CLS / etc. */
  phrase?: string;
  /** PRED (predicate) etc. */
  phrasalFunction?: string;
  punctuationMark?: string;
}

// ── Word with morphology ────────────────────────────────────────

export interface WordData {
  /** DB row id — used as `data-word-id` on the rendered span so the
   *  client can fetch detailed morphology without re-querying. */
  id: number;
  wordNumber: number;
  token: string;
  withoutDiacritics: string;
  translation: string;
  punctuationMark?: string;
  segments: Segment[];
}

// ── Verse & Surah data ──────────────────────────────────────────

export interface VerseData {
  number: number;
  text: string;
  surahNumber: number;
  words?: WordData[];
}

export interface SurahData {
  number: number;
  name: string;
  englishName: string;
  englishNameTranslation: string;
  revelationType: string;
  numberOfAyahs: number;
  verses: VerseData[];
  morphology?: Record<number, WordData[]> | null;
}

// ── UI helper types ─────────────────────────────────────────────

/** Compact word info shown in the hover tooltip */
export interface TooltipWord {
  token: string;
  gloss: string;
  root?: string;
  role: string;
  pos: string;
  case?: string;
  lemma?: string;
}

// ── Backward-compat aliases ─────────────────────────────────────
// Older code referenced "MorphWord" / "MorphSegment" — these are the same as
// WordData / Segment, just aliased for semantic clarity in the parser context.

/** @deprecated Use WordData instead. */
export type MorphWord = WordData;
/** @deprecated Use Segment instead. */
export type MorphSegment = Segment;