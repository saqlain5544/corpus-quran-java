import { PrismaClient } from "@/generated/prisma";
import type { WordData } from "@/lib/types";

const prisma = new PrismaClient();

// ── Surahs ──────────────────────────────────────────────────────

/** Lightweight surah list for the homepage */
export async function getSurahList() {
  const surahs = await prisma.surah.findMany({
    orderBy: { id: "asc" },
  });
  return surahs.map((s) => ({
    number: s.id,
    name: s.name,
    englishName: s.englishName,
    englishNameTranslation: s.englishTranslation,
    revelationType: s.revelationType,
    numberOfAyahs: s.verseCount,
  }));
}

/** Full surah with verses */
export async function getSurah(surahNumber: number) {
  const surah = await prisma.surah.findUnique({
    where: { id: surahNumber },
    include: {
      verses: {
        select: { number: true, text: true },
        orderBy: { number: "asc" },
      },
    },
  });
  if (!surah) throw new Error(`Surah ${surahNumber} not found`);
  return {
    number: surah.id,
    name: surah.name,
    englishName: surah.englishName,
    englishNameTranslation: surah.englishTranslation,
    revelationType: surah.revelationType,
    numberOfAyahs: surah.verseCount,
    verses: surah.verses.map((v) => ({
      number: v.number,
      text: v.text,
      surahNumber: surah.id,
    })),
  };
}

export async function getVerse(surahNumber: number, verseNumber: number) {
  const verse = await prisma.verse.findUnique({
    where: { surahId_number: { surahId: surahNumber, number: verseNumber } },
    select: { number: true, text: true, surahId: true },
  });
  if (!verse) return null;
  return { number: verse.number, text: verse.text, surahNumber: verse.surahId };
}

// ── Morphology ──────────────────────────────────────────────────

export async function getSurahMorphology(surahNumber: number): Promise<Record<number, WordData[]>> {
  // Get all verse IDs and numbers for this surah
  const verses = await prisma.verse.findMany({
    where: { surahId: surahNumber },
    select: { id: true, number: true },
  });
  const verseNumById = new Map(verses.map((v) => [v.id, v.number]));
  if (verses.length === 0) return {};

  const words = await prisma.word.findMany({
    where: { surahId: surahNumber },
    include: {
      segments: { orderBy: { number: "asc" } },
    },
    orderBy: [{ verseId: "asc" }, { number: "asc" }],
  });

  const byVerse: Record<number, WordData[]> = {};
  for (const w of words) {
    const vn = verseNumById.get(w.verseId);
    if (!vn) continue;
    if (!byVerse[vn]) byVerse[vn] = [];
    byVerse[vn].push({
      id: w.id,
      wordNumber: w.number,
      token: w.token,
      withoutDiacritics: w.withoutDiacritics,
      translation: w.translation || "",
      punctuationMark: w.punctuationMark || "",
      segments: w.segments.map((s) => ({
        segmentNumber: s.number,
        text: s.text,
        partOfSpeech: s.partOfSpeech,
        morphType: s.morphType,
        lemma: s.lemma ?? undefined,
        root: s.root ?? undefined,
        gender: s.gender ?? undefined,
        case: s.caseField ?? undefined,
        syntacticRole: s.syntacticRole ?? undefined,
        gloss: s.gloss ?? undefined,
        possessiveConstruct: s.possessiveConstruct ?? undefined,
        caseMoodMarker: s.caseMoodMarker ?? undefined,
        invariableDeclinable: s.invariableDeclinable ?? undefined,
        phrase: s.phrase ?? undefined,
        phrasalFunction: s.phrasalFunction ?? undefined,
        punctuationMark: s.punctuationMark ?? undefined,
      })),
    });
  }
  return byVerse;
}

// ── Search ─────────────────────────────────────────────────────

export async function getAllVersesForSearch() {
  const verses = await prisma.verse.findMany({
    select: { surahId: true, number: true, text: true },
    orderBy: [{ surahId: "asc" }, { number: "asc" }],
  });
  return verses.map((v) => ({
    surahNumber: v.surahId,
    verseNumber: v.number,
    text: v.text,
  }));
}

export async function getSearchMorphology() {
  // Use raw query to avoid SQLite parameter limit with nested includes
  const rows = await prisma.$queryRaw<
    Array<{ surahId: number; verseNumber: number; translation: string | null; root: string | null }>
  >`
    SELECT w.surahId, v.number as verseNumber, w.translation, s.root
    FROM word w
    JOIN verse v ON v.id = w.verseId
    LEFT JOIN segment s ON s.wordId = w.id AND s.root IS NOT NULL
  `;

  // Group by verse
  const map = new Map<string, { surahNumber: number; verseNumber: number; translation: string; roots: Set<string> }>();
  for (const r of rows) {
    const key = `${r.surahId}:${r.verseNumber}`;
    if (!map.has(key)) {
      map.set(key, { surahNumber: r.surahId, verseNumber: r.verseNumber, translation: r.translation || "", roots: new Set() });
    }
    if (r.root) map.get(key)!.roots.add(r.root);
  }

  return [...map.values()].map((v) => ({
    surahNumber: v.surahNumber,
    verseNumber: v.verseNumber,
    translation: v.translation,
    roots: [...v.roots],
  }));
}
