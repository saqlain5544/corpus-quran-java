/**
 * Database seed script — populates quran.db from all raw data files.
 * Run with: npx tsx prisma/seed.ts
 */
import { PrismaClient } from "../src/generated/prisma";
import { XMLParser } from "fast-xml-parser";
import fs from "fs";
import path from "path";
import { parse } from "papaparse";

const prisma = new PrismaClient();
const DATA_DIR = path.resolve(process.cwd(), "..", "data");

const SURAH_NAMES: Record<number, { englishName: string; englishNameTranslation: string; revelationType: "Meccan" | "Medinan"; numberOfAyahs: number }> = {
  1: { englishName: "Al-Faatiha", englishNameTranslation: "The Opening", revelationType: "Meccan", numberOfAyahs: 7 },
  2: { englishName: "Al-Baqara", englishNameTranslation: "The Cow", revelationType: "Medinan", numberOfAyahs: 286 },
  3: { englishName: "Aal-i-Imraan", englishNameTranslation: "The Family of Imraan", revelationType: "Medinan", numberOfAyahs: 200 },
  4: { englishName: "An-Nisaa", englishNameTranslation: "The Women", revelationType: "Medinan", numberOfAyahs: 176 },
  5: { englishName: "Al-Maaida", englishNameTranslation: "The Table Spread", revelationType: "Medinan", numberOfAyahs: 120 },
  6: { englishName: "Al-An'aam", englishNameTranslation: "The Cattle", revelationType: "Meccan", numberOfAyahs: 165 },
  7: { englishName: "Al-A'raaf", englishNameTranslation: "The Heights", revelationType: "Meccan", numberOfAyahs: 206 },
  8: { englishName: "Al-Anfaal", englishNameTranslation: "The Spoils of War", revelationType: "Medinan", numberOfAyahs: 75 },
  9: { englishName: "At-Tawba", englishNameTranslation: "The Repentance", revelationType: "Medinan", numberOfAyahs: 129 },
  10: { englishName: "Yunus", englishNameTranslation: "Jonah", revelationType: "Meccan", numberOfAyahs: 109 },
  11: { englishName: "Hud", englishNameTranslation: "Hud", revelationType: "Meccan", numberOfAyahs: 123 },
  12: { englishName: "Yusuf", englishNameTranslation: "Joseph", revelationType: "Meccan", numberOfAyahs: 111 },
  13: { englishName: "Ar-Ra'd", englishNameTranslation: "The Thunder", revelationType: "Medinan", numberOfAyahs: 43 },
  14: { englishName: "Ibrahim", englishNameTranslation: "Abraham", revelationType: "Meccan", numberOfAyahs: 52 },
  15: { englishName: "Al-Hijr", englishNameTranslation: "The Rocky Tract", revelationType: "Meccan", numberOfAyahs: 99 },
  16: { englishName: "An-Nahl", englishNameTranslation: "The Bee", revelationType: "Meccan", numberOfAyahs: 128 },
  17: { englishName: "Al-Israa", englishNameTranslation: "The Night Journey", revelationType: "Meccan", numberOfAyahs: 111 },
  18: { englishName: "Al-Kahf", englishNameTranslation: "The Cave", revelationType: "Meccan", numberOfAyahs: 110 },
  19: { englishName: "Maryam", englishNameTranslation: "Mary", revelationType: "Meccan", numberOfAyahs: 98 },
  20: { englishName: "Taa-Haa", englishNameTranslation: "Taa-Haa", revelationType: "Meccan", numberOfAyahs: 135 },
  21: { englishName: "Al-Anbiyaa", englishNameTranslation: "The Prophets", revelationType: "Meccan", numberOfAyahs: 112 },
  22: { englishName: "Al-Hajj", englishNameTranslation: "The Pilgrimage", revelationType: "Medinan", numberOfAyahs: 78 },
  23: { englishName: "Al-Muminoon", englishNameTranslation: "The Believers", revelationType: "Meccan", numberOfAyahs: 118 },
  24: { englishName: "An-Noor", englishNameTranslation: "The Light", revelationType: "Medinan", numberOfAyahs: 64 },
  25: { englishName: "Al-Furqaan", englishNameTranslation: "The Criterion", revelationType: "Meccan", numberOfAyahs: 77 },
  26: { englishName: "Ash-Shu'araa", englishNameTranslation: "The Poets", revelationType: "Meccan", numberOfAyahs: 227 },
  27: { englishName: "An-Naml", englishNameTranslation: "The Ant", revelationType: "Meccan", numberOfAyahs: 93 },
  28: { englishName: "Al-Qasas", englishNameTranslation: "The Stories", revelationType: "Meccan", numberOfAyahs: 88 },
  29: { englishName: "Al-Ankaboot", englishNameTranslation: "The Spider", revelationType: "Meccan", numberOfAyahs: 69 },
  30: { englishName: "Ar-Room", englishNameTranslation: "The Romans", revelationType: "Meccan", numberOfAyahs: 60 },
  31: { englishName: "Luqman", englishNameTranslation: "Luqman", revelationType: "Meccan", numberOfAyahs: 34 },
  32: { englishName: "As-Sajda", englishNameTranslation: "The Prostration", revelationType: "Meccan", numberOfAyahs: 30 },
  33: { englishName: "Al-Ahzaab", englishNameTranslation: "The Clans", revelationType: "Medinan", numberOfAyahs: 73 },
  34: { englishName: "Saba", englishNameTranslation: "Sheba", revelationType: "Meccan", numberOfAyahs: 54 },
  35: { englishName: "Faatir", englishNameTranslation: "The Originator", revelationType: "Meccan", numberOfAyahs: 45 },
  36: { englishName: "Yaseen", englishNameTranslation: "Yaseen", revelationType: "Meccan", numberOfAyahs: 83 },
  37: { englishName: "As-Saaffaat", englishNameTranslation: "Those drawn up in Ranks", revelationType: "Meccan", numberOfAyahs: 182 },
  38: { englishName: "Saad", englishNameTranslation: "The Letter Saad", revelationType: "Meccan", numberOfAyahs: 88 },
  39: { englishName: "Az-Zumar", englishNameTranslation: "The Groups", revelationType: "Meccan", numberOfAyahs: 75 },
  40: { englishName: "Ghaafir", englishNameTranslation: "The Forgiver", revelationType: "Meccan", numberOfAyahs: 85 },
  41: { englishName: "Fussilat", englishNameTranslation: "Explained in Detail", revelationType: "Meccan", numberOfAyahs: 54 },
  42: { englishName: "Ash-Shooraa", englishNameTranslation: "Consultation", revelationType: "Meccan", numberOfAyahs: 53 },
  43: { englishName: "Az-Zukhruf", englishNameTranslation: "Ornaments of Gold", revelationType: "Meccan", numberOfAyahs: 89 },
  44: { englishName: "Ad-Dukhaan", englishNameTranslation: "The Smoke", revelationType: "Meccan", numberOfAyahs: 59 },
  45: { englishName: "Al-Jaathiya", englishNameTranslation: "Crouching", revelationType: "Meccan", numberOfAyahs: 37 },
  46: { englishName: "Al-Ahqaaf", englishNameTranslation: "The Dunes", revelationType: "Meccan", numberOfAyahs: 35 },
  47: { englishName: "Muhammad", englishNameTranslation: "Muhammad", revelationType: "Medinan", numberOfAyahs: 38 },
  48: { englishName: "Al-Fath", englishNameTranslation: "The Victory", revelationType: "Medinan", numberOfAyahs: 29 },
  49: { englishName: "Al-Hujuraat", englishNameTranslation: "The Inner Apartments", revelationType: "Medinan", numberOfAyahs: 18 },
  50: { englishName: "Qaaf", englishNameTranslation: "The Letter Qaaf", revelationType: "Meccan", numberOfAyahs: 45 },
  51: { englishName: "Adh-Dhaariyat", englishNameTranslation: "The Winnowing Winds", revelationType: "Meccan", numberOfAyahs: 60 },
  52: { englishName: "At-Toor", englishNameTranslation: "The Mount", revelationType: "Meccan", numberOfAyahs: 49 },
  53: { englishName: "An-Najm", englishNameTranslation: "The Star", revelationType: "Meccan", numberOfAyahs: 62 },
  54: { englishName: "Al-Qamar", englishNameTranslation: "The Moon", revelationType: "Meccan", numberOfAyahs: 55 },
  55: { englishName: "Ar-Rahmaan", englishNameTranslation: "The Beneficent", revelationType: "Medinan", numberOfAyahs: 78 },
  56: { englishName: "Al-Waaqia", englishNameTranslation: "The Inevitable", revelationType: "Meccan", numberOfAyahs: 96 },
  57: { englishName: "Al-Hadeed", englishNameTranslation: "The Iron", revelationType: "Medinan", numberOfAyahs: 29 },
  58: { englishName: "Al-Mujaadila", englishNameTranslation: "The Pleading Woman", revelationType: "Medinan", numberOfAyahs: 22 },
  59: { englishName: "Al-Hashr", englishNameTranslation: "The Exile", revelationType: "Medinan", numberOfAyahs: 24 },
  60: { englishName: "Al-Mumtahana", englishNameTranslation: "She That is to be Examined", revelationType: "Medinan", numberOfAyahs: 13 },
  61: { englishName: "As-Saff", englishNameTranslation: "The Ranks", revelationType: "Medinan", numberOfAyahs: 14 },
  62: { englishName: "Al-Jumu'a", englishNameTranslation: "Friday", revelationType: "Medinan", numberOfAyahs: 11 },
  63: { englishName: "Al-Munaafiqoon", englishNameTranslation: "The Hypocrites", revelationType: "Medinan", numberOfAyahs: 11 },
  64: { englishName: "At-Taghaabun", englishNameTranslation: "Mutual Disillusion", revelationType: "Medinan", numberOfAyahs: 18 },
  65: { englishName: "At-Talaaq", englishNameTranslation: "The Divorce", revelationType: "Medinan", numberOfAyahs: 12 },
  66: { englishName: "At-Tahreem", englishNameTranslation: "The Prohibition", revelationType: "Medinan", numberOfAyahs: 12 },
  67: { englishName: "Al-Mulk", englishNameTranslation: "The Sovereignty", revelationType: "Meccan", numberOfAyahs: 30 },
  68: { englishName: "Al-Qalam", englishNameTranslation: "The Pen", revelationType: "Meccan", numberOfAyahs: 52 },
  69: { englishName: "Al-Haaqqa", englishNameTranslation: "The Reality", revelationType: "Meccan", numberOfAyahs: 52 },
  70: { englishName: "Al-Ma'aarij", englishNameTranslation: "The Ascending Stairways", revelationType: "Meccan", numberOfAyahs: 44 },
  71: { englishName: "Nooh", englishNameTranslation: "Noah", revelationType: "Meccan", numberOfAyahs: 28 },
  72: { englishName: "Al-Jinn", englishNameTranslation: "The Jinn", revelationType: "Meccan", numberOfAyahs: 28 },
  73: { englishName: "Al-Muzzammil", englishNameTranslation: "The Enshrouded One", revelationType: "Meccan", numberOfAyahs: 20 },
  74: { englishName: "Al-Muddaththir", englishNameTranslation: "The Cloaked One", revelationType: "Meccan", numberOfAyahs: 56 },
  75: { englishName: "Al-Qiyaama", englishNameTranslation: "The Resurrection", revelationType: "Meccan", numberOfAyahs: 40 },
  76: { englishName: "Al-Insaan", englishNameTranslation: "Man", revelationType: "Medinan", numberOfAyahs: 31 },
  77: { englishName: "Al-Mursalaat", englishNameTranslation: "The Emissaries", revelationType: "Meccan", numberOfAyahs: 50 },
  78: { englishName: "An-Naba", englishNameTranslation: "The Announcement", revelationType: "Meccan", numberOfAyahs: 40 },
  79: { englishName: "An-Naazi'aat", englishNameTranslation: "Those who Drag Forth", revelationType: "Meccan", numberOfAyahs: 46 },
  80: { englishName: "Abasa", englishNameTranslation: "He Frowned", revelationType: "Meccan", numberOfAyahs: 42 },
  81: { englishName: "At-Takweer", englishNameTranslation: "The Overthrowing", revelationType: "Meccan", numberOfAyahs: 29 },
  82: { englishName: "Al-Infitaar", englishNameTranslation: "The Cleaving", revelationType: "Meccan", numberOfAyahs: 19 },
  83: { englishName: "Al-Mutaffifeen", englishNameTranslation: "Defrauding", revelationType: "Meccan", numberOfAyahs: 36 },
  84: { englishName: "Al-Inshiqaaq", englishNameTranslation: "The Splitting Open", revelationType: "Meccan", numberOfAyahs: 25 },
  85: { englishName: "Al-Burooj", englishNameTranslation: "The Constellations", revelationType: "Meccan", numberOfAyahs: 22 },
  86: { englishName: "At-Taariq", englishNameTranslation: "The Morning Star", revelationType: "Meccan", numberOfAyahs: 17 },
  87: { englishName: "Al-A'laa", englishNameTranslation: "The Most High", revelationType: "Meccan", numberOfAyahs: 19 },
  88: { englishName: "Al-Ghaashiya", englishNameTranslation: "The Overwhelming", revelationType: "Meccan", numberOfAyahs: 26 },
  89: { englishName: "Al-Fajr", englishNameTranslation: "The Dawn", revelationType: "Meccan", numberOfAyahs: 30 },
  90: { englishName: "Al-Balad", englishNameTranslation: "The City", revelationType: "Meccan", numberOfAyahs: 20 },
  91: { englishName: "Ash-Shams", englishNameTranslation: "The Sun", revelationType: "Meccan", numberOfAyahs: 15 },
  92: { englishName: "Al-Lail", englishNameTranslation: "The Night", revelationType: "Meccan", numberOfAyahs: 21 },
  93: { englishName: "Ad-Dhuhaa", englishNameTranslation: "The Morning Hours", revelationType: "Meccan", numberOfAyahs: 11 },
  94: { englishName: "Ash-Sharh", englishNameTranslation: "The Consolation", revelationType: "Meccan", numberOfAyahs: 8 },
  95: { englishName: "At-Teen", englishNameTranslation: "The Fig", revelationType: "Meccan", numberOfAyahs: 8 },
  96: { englishName: "Al-Alaq", englishNameTranslation: "The Clot", revelationType: "Meccan", numberOfAyahs: 19 },
  97: { englishName: "Al-Qadr", englishNameTranslation: "The Power, Fate", revelationType: "Meccan", numberOfAyahs: 5 },
  98: { englishName: "Al-Bayyina", englishNameTranslation: "The Evidence", revelationType: "Medinan", numberOfAyahs: 8 },
  99: { englishName: "Az-Zalzala", englishNameTranslation: "The Earthquake", revelationType: "Medinan", numberOfAyahs: 8 },
  100: { englishName: "Al-Aadiyaat", englishNameTranslation: "The Courser", revelationType: "Meccan", numberOfAyahs: 11 },
  101: { englishName: "Al-Qaari'a", englishNameTranslation: "The Calamity", revelationType: "Meccan", numberOfAyahs: 11 },
  102: { englishName: "At-Takaathur", englishNameTranslation: "Competition", revelationType: "Meccan", numberOfAyahs: 8 },
  103: { englishName: "Al-Asr", englishNameTranslation: "The Declining Day", revelationType: "Meccan", numberOfAyahs: 3 },
  104: { englishName: "Al-Humaza", englishNameTranslation: "The Traducer", revelationType: "Meccan", numberOfAyahs: 9 },
  105: { englishName: "Al-Feel", englishNameTranslation: "The Elephant", revelationType: "Meccan", numberOfAyahs: 5 },
  106: { englishName: "Quraish", englishNameTranslation: "Quraish", revelationType: "Meccan", numberOfAyahs: 4 },
  107: { englishName: "Al-Maa'oon", englishNameTranslation: "Small Kindnesses", revelationType: "Meccan", numberOfAyahs: 7 },
  108: { englishName: "Al-Kawthar", englishNameTranslation: "Abundance", revelationType: "Meccan", numberOfAyahs: 3 },
  109: { englishName: "Al-Kaafiroon", englishNameTranslation: "The Disbelievers", revelationType: "Meccan", numberOfAyahs: 6 },
  110: { englishName: "An-Nasr", englishNameTranslation: "Divine Support", revelationType: "Medinan", numberOfAyahs: 3 },
  111: { englishName: "Al-Masad", englishNameTranslation: "The Palm Fibre", revelationType: "Meccan", numberOfAyahs: 5 },
  112: { englishName: "Al-Ikhlaas", englishNameTranslation: "Sincerity", revelationType: "Meccan", numberOfAyahs: 4 },
  113: { englishName: "Al-Falaq", englishNameTranslation: "The Dawn", revelationType: "Meccan", numberOfAyahs: 5 },
  114: { englishName: "An-Naas", englishNameTranslation: "Mankind", revelationType: "Meccan", numberOfAyahs: 6 },
};

function parseMorphologyTag(morphTag: string): { lemma?: string; root?: string } {
  const parts = morphTag.split(/\s+/).map((p) => p.trim()).filter(Boolean);
  const result: { lemma?: string; root?: string } = {};
  for (const part of parts) {
    if (part.startsWith("LEM:")) result.lemma = part.slice(4).replace(/[{}]/g, "");
    else if (part.startsWith("ROOT:")) result.root = part.slice(5);
  }
  return result;
}

async function main() {
  console.log("Seeding database...");

  // 1. Surahs
  console.log("  Surahs...");
  for (let i = 1; i <= 114; i++) {
    const m = SURAH_NAMES[i];
    await prisma.surah.create({
      data: { id: i, name: m.englishName, englishName: m.englishName, englishTranslation: m.englishNameTranslation, revelationType: m.revelationType, verseCount: m.numberOfAyahs },
    });
  }

  // 2. Verses from quran-uthmani.xml
  console.log("  Verses...");
  const quranXml = fs.readFileSync(path.join(DATA_DIR, "quran", "quran-uthmani.xml"), "utf-8");
  const xmlParser = new XMLParser({ ignoreAttributes: false, attributeNamePrefix: "@_" });
  const quranParsed = xmlParser.parse(quranXml);
  const suras = quranParsed.quran.sura as Array<Record<string, unknown>>;

  for (const sura of suras) {
    const surahId = Number(sura["@_index"]);
    const ayas = (Array.isArray(sura.aya) ? sura.aya : [sura.aya]) as Array<Record<string, unknown>>;
    for (const aya of ayas) {
      if (!aya) continue;
      await prisma.verse.create({
        data: {
          surahId,
          number: Number(aya["@_index"]),
          text: String(aya["@_text"] || ""),
        },
      });
    }
  }

  // 3. Words + Segments from MASAQ.csv
  console.log("  Words & segments...");
  const csv = fs.readFileSync(path.join(DATA_DIR, "morphology", "MASAQ.csv"), "utf-8");
  const { data: masaRows } = parse<Record<string, string>>(csv, { header: true, skipEmptyLines: true });

  // 3a. Root enrichment from corpus-roots.json (inverted: word→root)
  const rootRefsJson = JSON.parse(fs.readFileSync(path.join(DATA_DIR, "morphology", "corpus-roots.json"), "utf-8"));
  const wordRootLookup = new Map<string, string>();
  for (const [root, refs] of Object.entries(rootRefsJson)) {
    for (const ref of refs as string[]) wordRootLookup.set(ref, root);
  }

  // 3b. Seed words and segments
  // Group by word key
  const wordGroups = new Map<string, { surah: number; verse: number; word: number; token: string; without: string; punct: string; segments: Array<Record<string, unknown>> }>();
  for (const row of masaRows) {
    const s = parseInt(row["Sura_No"], 10);
    const v = parseInt(row["Verse_No"], 10);
    const w = parseInt(row["Word_No"], 10);
    const segNo = parseInt(row["Segment_No"], 10);
    const key = `${s}:${v}:${w}`;

    if (!wordGroups.has(key)) {
      wordGroups.set(key, {
        surah: s, verse: v, word: w,
        token: (row["Word"] || "").replace(/^"(.*)"$/, "$1"),
        without: row["Without_Diacritics"] || "",
        punct: row["Punctuation_Mark"] || "",
        segments: [],
      });
    }

    const segData: Record<string, unknown> = {
      number: segNo,
      text: (row["Segmented_Word"] && row["Segmented_Word"] !== "(null)") ? row["Segmented_Word"] : "",
      partOfSpeech: row["Morph_Tag"] || "UNKNOWN",
      morphType: row["Morph_Type"] || "",
      syntacticRole: row["Syntactic_Role"] || undefined,
      gloss: (row["Gloss"] && row["Gloss"] !== "(null)") ? row["Gloss"] : undefined,
      possessiveConstruct: row["Possessive_Construct"] || undefined,
      caseField: row["Case_Mood"] || undefined,
      caseMoodMarker: row["Case_Mood_Marker"] || undefined,
      invariableDeclinable: row["Invariable_Declinable"] || undefined,
      phrase: row["Phrase"] || undefined,
      phrasalFunction: row["Phrasal_Function"] || undefined,
      punctuationMark: row["Punctuation_Mark"] || undefined,
    };
    wordGroups.get(key)!.segments.push(segData);
  }

  // Batch insert: first get verse IDs
  const verseLookup = new Map<string, number>();
  const allVerses = await prisma.verse.findMany({ select: { id: true, surahId: true, number: true } });
  for (const v of allVerses) verseLookup.set(`${v.surahId}:${v.number}`, v.id);

  let wordCount = 0;
  let segCount = 0;

  // Process in batches of 500 for performance
  const wordEntries = [...wordGroups.entries()];
  for (let i = 0; i < wordEntries.length; i += 500) {
    const batch = wordEntries.slice(i, i + 500);
    await prisma.$transaction(async (tx) => {
      const seenSegments = new Map<string, boolean>();
      for (const [, wg] of batch) {
        const verseId = verseLookup.get(`${wg.surah}:${wg.verse}`);
        if (!verseId) continue;

        const wordRoot = wordRootLookup.get(`${wg.surah}:${wg.verse}:${wg.word}`) || undefined;

        const word = await tx.word.create({
          data: {
            surahId: wg.surah,
            verseId,
            number: wg.word,
            token: wg.token,
            withoutDiacritics: wg.without,
            translation: wg.segments.map((s) => s.gloss || s.text).filter(Boolean).join(" "),
            punctuationMark: wg.punct || undefined,
          },
        });

        for (const seg of wg.segments) {
          // Deduplicate: skip segments with same number (MASAQ can have dupes)
          const segKey = `${word.id}:${seg.number}`;
          if (seenSegments.has(segKey)) continue;
          seenSegments.set(segKey, true);

          await tx.segment.create({
            data: {
              wordId: word.id,
              number: seg.number as number,
              text: seg.text as string,
              partOfSpeech: seg.partOfSpeech as string,
              morphType: seg.morphType as string,
              lemma: (seg.lemma as string) || undefined,
              root: (seg.root as string) || wordRoot || undefined,
              syntacticRole: seg.syntacticRole as string | undefined,
              gloss: seg.gloss as string | undefined,
              possessiveConstruct: seg.possessiveConstruct as string | undefined,
              caseField: seg.caseField as string | undefined,
              caseMoodMarker: seg.caseMoodMarker as string | undefined,
              invariableDeclinable: seg.invariableDeclinable as string | undefined,
              phrase: seg.phrase as string | undefined,
              phrasalFunction: seg.phrasalFunction as string | undefined,
              punctuationMark: seg.punctuationMark as string | undefined,
            },
          });
          segCount++;
        }
        wordCount++;
      }
    });
    if (wordCount % 5000 === 0) console.log(`    ${wordCount} words, ${segCount} segments...`);
  }

  console.log(`    Done: ${wordCount} words, ${segCount} segments`);

  // 4. Root references
  console.log("  Root references...");
  const rootsJson = JSON.parse(fs.readFileSync(path.join(DATA_DIR, "morphology", "corpus-roots.json"), "utf-8"));
  for (const [root, refs] of Object.entries(rootRefsJson)) {
    for (const ref of refs as string[]) {
      const parts = ref.split(":");
      await prisma.rootReference.create({
        data: {
          root,
          surahId: parseInt(parts[0], 10),
          verseNumber: parseInt(parts[1], 10),
          wordNumber: parseInt(parts[2], 10),
        },
      });
    }
  }

  // 5. Root meanings
  console.log("  Root meanings...");
  const rootCsv = fs.readFileSync(path.join(DATA_DIR, "morphology", "root_meaning.csv"), "utf-8");
  const { data: rootRows } = parse<Record<string, string>>(rootCsv, { header: true, skipEmptyLines: true });
  for (const row of rootRows) {
    const r = row["root"]?.trim();
    if (!r) continue;
    await prisma.rootMeaning.create({
      data: {
        id: parseInt(row["id"] || "0", 10),
        root: r,
        rootArabic: row["root_arabic"] || "",
        generalMeaning: row["general_meaning"] || "",
        shadesOfMeaning: row["shades_of_meaning"] || "[]",
        hadithExamples: row["hadith_examples"] || "[]",
        classicalSources: row["classical_sources"] || "[]",
        idiomsCustoms: row["idioms_customs"] || "[]",
        wordByWordShades: row["word_by_word_shades"] || "[]",
        pos: row["pos"] || undefined,
        modelUsed: row["model_used"] || undefined,
      },
    });
  }

  console.log("\n✅ Database seeded successfully!");
  console.log(`  - 114 surahs`);
  console.log(`  - ${allVerses.length} verses`);
  console.log(`  - ${wordCount} words`);
  console.log(`  - ${segCount} segments`);
  console.log(`  - Root references: ${Object.keys(rootRefsJson).length} roots`);
  console.log(`  - Root meanings: ${rootRows.length} entries`);
}

main()
  .catch(console.error)
  .finally(() => prisma.$disconnect());
