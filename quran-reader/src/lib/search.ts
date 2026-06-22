import { normalizeArabic } from "./utils";

/**
 * Search across Quran text (from original XML) and morphology data.
 * Text search queries the XML directly; root/gloss search uses preprocessed morphology.
 */
export function searchQuranText(
  query: string,
  verses: { surahNumber: number; verseNumber: number; text: string }[],
  surahNames: Record<number, string>,
  limit = 20
) {
  const results: Array<{ surahNumber: number; surahName: string; verseNumber: number; verseText: string; relevance: number }> = [];
  const normalizedQuery = normalizeArabic(query).toLowerCase();
  const queryTerms = normalizedQuery.split(/\s+/);

  for (const verse of verses) {
    const normalizedText = normalizeArabic(verse.text);
    let relevance = 0;

    for (const term of queryTerms) {
      if (term.length < 2) continue;
      const escaped = term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      const matches = (normalizedText.match(new RegExp(escaped, "gi")) || []).length;
      relevance += matches;
      if (verse.text.includes(term)) relevance += 2;
    }

    if (relevance > 0) {
      results.push({
        surahNumber: verse.surahNumber,
        surahName: surahNames[verse.surahNumber] || "",
        verseNumber: verse.verseNumber,
        verseText: verse.text,
        relevance,
      });
    }
  }

  results.sort((a, b) => b.relevance - a.relevance);
  return results.slice(0, limit);
}

export function searchQuranByRoot(
  query: string,
  morphology: { surahNumber: number; verseNumber: number; translation: string; roots: string[] }[],
  verses: { surahNumber: number; verseNumber: number; text: string }[],
  surahNames: Record<number, string>,
  limit = 20
) {
  const results: Array<{ surahNumber: number; surahName: string; verseNumber: number; verseText: string; matchContext?: string; relevance: number }> = [];
  const normalizedQuery = query.toLowerCase();

  for (const entry of morphology) {
    if (!entry.roots?.length) continue;
    const found = entry.roots.some((r) => normalizedQuery.includes(r.toLowerCase()) || r.toLowerCase().includes(normalizedQuery));
    if (!found) continue;

    const verse = verses.find((v) => v.surahNumber === entry.surahNumber && v.verseNumber === entry.verseNumber);
    if (!verse) continue;

    results.push({
      surahNumber: entry.surahNumber,
      surahName: surahNames[entry.surahNumber] || "",
      verseNumber: entry.verseNumber,
      verseText: verse.text,
      matchContext: `Root: ${entry.roots.join(", ")}`,
      relevance: 1,
    });
  }

  return results.slice(0, limit);
}

export function searchQuranByGloss(
  query: string,
  morphology: { surahNumber: number; verseNumber: number; translation: string; roots: string[] }[],
  verses: { surahNumber: number; verseNumber: number; text: string }[],
  surahNames: Record<number, string>,
  limit = 20
) {
  const results: Array<{ surahNumber: number; surahName: string; verseNumber: number; verseText: string; matchContext?: string; relevance: number }> = [];
  const qLower = query.toLowerCase();

  for (const entry of morphology) {
    if (!entry.translation.toLowerCase().includes(qLower)) continue;

    const verse = verses.find((v) => v.surahNumber === entry.surahNumber && v.verseNumber === entry.verseNumber);
    if (!verse) continue;

    results.push({
      surahNumber: entry.surahNumber,
      surahName: surahNames[entry.surahNumber] || "",
      verseNumber: entry.verseNumber,
      verseText: verse.text,
      matchContext: entry.translation,
      relevance: 1,
    });
  }

  return results.slice(0, limit);
}
