import { z } from "zod";

// ── Quran Text Schemas ──────────────────────────────────────────

export const VerseSchema = z.object({
  number: z.number().int().positive(),
  text: z.string(),
  surahNumber: z.number().int().min(1).max(114),
});

export const SurahSchema = z.object({
  number: z.number().int().min(1).max(114),
  name: z.string(),
  englishName: z.string(),
  englishNameTranslation: z.string(),
  revelationType: z.enum(["Meccan", "Medinan"]),
  numberOfAyahs: z.number().int().positive(),
  verses: z.array(VerseSchema),
});

export const QuranSchema = z.array(SurahSchema);

// ── Morphology Schemas ──────────────────────────────────────────

export const MorphSegmentSchema = z.object({
  segmentNumber: z.number().int().positive(),
  text: z.string(),
  partOfSpeech: z.string(),
  morphType: z.string(),
  lemma: z.string().optional(),
  root: z.string().optional(),
  gender: z.string().optional(),
  case: z.string().optional(),
  syntacticRole: z.string().optional(),
  gloss: z.string().optional(),
  possessiveConstruct: z.string().optional(),
  caseMoodMarker: z.string().optional(),
  invariableDeclinable: z.string().optional(),
  phrase: z.string().optional(),
  phrasalFunction: z.string().optional(),
  punctuationMark: z.string().optional(),
});

export const WordMorphologySchema = z.object({
  wordNumber: z.number().int().positive(),
  token: z.string(),
  withoutDiacritics: z.string(),
  translation: z.string().optional(),
  punctuationMark: z.string().optional(),
  segments: z.array(MorphSegmentSchema),
});

export const VerseMorphologySchema = z.object({
  surahNumber: z.number().int().min(1).max(114),
  verseNumber: z.number().int().positive(),
  words: z.array(WordMorphologySchema),
});

// ── Root Index Schema ───────────────────────────────────────────

export const RootIndexSchema = z.record(z.string(), z.array(z.string()));

// ── Search Schemas ──────────────────────────────────────────────

export const SearchQuerySchema = z.object({
  q: z.string().min(1).max(200),
  type: z.enum(["text", "root", "gloss"]).default("text"),
  limit: z.number().int().min(1).max(100).default(20),
});

export const SearchResultSchema = z.object({
  surahNumber: z.number().int().min(1).max(114),
  surahName: z.string(),
  verseNumber: z.number().int().positive(),
  verseText: z.string(),
  matchContext: z.string().optional(),
  relevance: z.number(),
});

// ── Derived Types ───────────────────────────────────────────────

export type Verse = z.infer<typeof VerseSchema>;
export type Surah = z.infer<typeof SurahSchema>;
export type Quran = z.infer<typeof QuranSchema>;
export type MorphSegment = z.infer<typeof MorphSegmentSchema>;
export type WordMorphology = z.infer<typeof WordMorphologySchema>;
export type VerseMorphology = z.infer<typeof VerseMorphologySchema>;
export type RootIndex = z.infer<typeof RootIndexSchema>;
export type SearchQuery = z.infer<typeof SearchQuerySchema>;
export type SearchResult = z.infer<typeof SearchResultSchema>;
