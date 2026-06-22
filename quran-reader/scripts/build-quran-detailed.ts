/**
 * Build data/new/quran-detailed.xml — amalgamates data sources:
 *   quran-uthmani.xml   → surah structure + verse text
 *   MASAQ.csv           → word + segment morphology
 *   corpus-roots.json   → root (Buckwalter) per word
 *   root_meaning.csv    → root Arabic script
 *
 * Run: npx tsx scripts/build-quran-detailed.ts
 */
import fs from "fs";
import path from "path";
import { parse } from "papaparse";
import { XMLParser } from "fast-xml-parser";

const DATA_DIR = path.resolve(process.cwd(), "..", "data");
const OUT_PATH = path.join(DATA_DIR, "new", "quran-detailed.xml");

// ── Surah metadata ──────────────────────────────────────────────

const SURAH_META: Record<number, { englishName: string; englishTranslation: string; revelationType: string; verseCount: number }> = {
  1:  { englishName: "Al-Faatiha",     englishTranslation: "The Opening",          revelationType: "Meccan",  verseCount: 7 },
  2:  { englishName: "Al-Baqara",      englishTranslation: "The Cow",              revelationType: "Medinan", verseCount: 286 },
  3:  { englishName: "Aal-i-Imraan",   englishTranslation: "The Family of Imraan", revelationType: "Medinan", verseCount: 200 },
  4:  { englishName: "An-Nisaa",       englishTranslation: "The Women",            revelationType: "Medinan", verseCount: 176 },
  5:  { englishName: "Al-Maaida",      englishTranslation: "The Table Spread",     revelationType: "Medinan", verseCount: 120 },
  6:  { englishName: "Al-An'aam",      englishTranslation: "The Cattle",           revelationType: "Meccan",  verseCount: 165 },
  7:  { englishName: "Al-A'raaf",      englishTranslation: "The Heights",          revelationType: "Meccan",  verseCount: 206 },
  8:  { englishName: "Al-Anfaal",      englishTranslation: "The Spoils of War",    revelationType: "Medinan", verseCount: 75 },
  9:  { englishName: "At-Tawba",       englishTranslation: "The Repentance",       revelationType: "Medinan", verseCount: 129 },
  10: { englishName: "Yunus",          englishTranslation: "Jonah",                revelationType: "Meccan",  verseCount: 109 },
  11: { englishName: "Hud",            englishTranslation: "Hud",                  revelationType: "Meccan",  verseCount: 123 },
  12: { englishName: "Yusuf",          englishTranslation: "Joseph",               revelationType: "Meccan",  verseCount: 111 },
  13: { englishName: "Ar-Ra'd",        englishTranslation: "The Thunder",          revelationType: "Medinan", verseCount: 43 },
  14: { englishName: "Ibrahim",        englishTranslation: "Abraham",              revelationType: "Meccan",  verseCount: 52 },
  15: { englishName: "Al-Hijr",        englishTranslation: "The Rocky Tract",      revelationType: "Meccan",  verseCount: 99 },
  16: { englishName: "An-Nahl",        englishTranslation: "The Bee",              revelationType: "Meccan",  verseCount: 128 },
  17: { englishName: "Al-Israa",       englishTranslation: "The Night Journey",    revelationType: "Meccan",  verseCount: 111 },
  18: { englishName: "Al-Kahf",        englishTranslation: "The Cave",             revelationType: "Meccan",  verseCount: 110 },
  19: { englishName: "Maryam",         englishTranslation: "Mary",                 revelationType: "Meccan",  verseCount: 98 },
  20: { englishName: "Taa-Haa",        englishTranslation: "Taa-Haa",              revelationType: "Meccan",  verseCount: 135 },
  21: { englishName: "Al-Anbiyaa",     englishTranslation: "The Prophets",         revelationType: "Meccan",  verseCount: 112 },
  22: { englishName: "Al-Hajj",        englishTranslation: "The Pilgrimage",       revelationType: "Medinan", verseCount: 78 },
  23: { englishName: "Al-Muminoon",    englishTranslation: "The Believers",        revelationType: "Meccan",  verseCount: 118 },
  24: { englishName: "An-Noor",        englishTranslation: "The Light",            revelationType: "Medinan", verseCount: 64 },
  25: { englishName: "Al-Furqaan",     englishTranslation: "The Criterion",        revelationType: "Meccan",  verseCount: 77 },
  26: { englishName: "Ash-Shu'araa",   englishTranslation: "The Poets",            revelationType: "Meccan",  verseCount: 227 },
  27: { englishName: "An-Naml",        englishTranslation: "The Ant",              revelationType: "Meccan",  verseCount: 93 },
  28: { englishName: "Al-Qasas",       englishTranslation: "The Stories",          revelationType: "Meccan",  verseCount: 88 },
  29: { englishName: "Al-Ankaboot",    englishTranslation: "The Spider",           revelationType: "Meccan",  verseCount: 69 },
  30: { englishName: "Ar-Room",        englishTranslation: "The Romans",           revelationType: "Meccan",  verseCount: 60 },
  31: { englishName: "Luqman",         englishTranslation: "Luqman",               revelationType: "Meccan",  verseCount: 34 },
  32: { englishName: "As-Sajda",       englishTranslation: "The Prostration",      revelationType: "Meccan",  verseCount: 30 },
  33: { englishName: "Al-Ahzaab",      englishTranslation: "The Clans",            revelationType: "Medinan", verseCount: 73 },
  34: { englishName: "Saba",           englishTranslation: "Sheba",                revelationType: "Meccan",  verseCount: 54 },
  35: { englishName: "Faatir",         englishTranslation: "The Originator",       revelationType: "Meccan",  verseCount: 45 },
  36: { englishName: "Yaseen",         englishTranslation: "Yaseen",               revelationType: "Meccan",  verseCount: 83 },
  37: { englishName: "As-Saaffaat",    englishTranslation: "Those drawn up in Ranks", revelationType: "Meccan", verseCount: 182 },
  38: { englishName: "Saad",           englishTranslation: "The Letter Saad",      revelationType: "Meccan",  verseCount: 88 },
  39: { englishName: "Az-Zumar",       englishTranslation: "The Groups",           revelationType: "Meccan",  verseCount: 75 },
  40: { englishName: "Ghaafir",        englishTranslation: "The Forgiver",         revelationType: "Meccan",  verseCount: 85 },
  41: { englishName: "Fussilat",       englishTranslation: "Explained in Detail",  revelationType: "Meccan",  verseCount: 54 },
  42: { englishName: "Ash-Shooraa",    englishTranslation: "Consultation",         revelationType: "Meccan",  verseCount: 53 },
  43: { englishName: "Az-Zukhruf",     englishTranslation: "Ornaments of Gold",    revelationType: "Meccan",  verseCount: 89 },
  44: { englishName: "Ad-Dukhaan",     englishTranslation: "The Smoke",            revelationType: "Meccan",  verseCount: 59 },
  45: { englishName: "Al-Jaathiya",    englishTranslation: "Crouching",            revelationType: "Meccan",  verseCount: 37 },
  46: { englishName: "Al-Ahqaaf",      englishTranslation: "The Dunes",            revelationType: "Meccan",  verseCount: 35 },
  47: { englishName: "Muhammad",       englishTranslation: "Muhammad",             revelationType: "Medinan", verseCount: 38 },
  48: { englishName: "Al-Fath",        englishTranslation: "The Victory",          revelationType: "Medinan", verseCount: 29 },
  49: { englishName: "Al-Hujuraat",    englishTranslation: "The Inner Apartments",  revelationType: "Medinan", verseCount: 18 },
  50: { englishName: "Qaaf",           englishTranslation: "The Letter Qaaf",      revelationType: "Meccan",  verseCount: 45 },
  51: { englishName: "Adh-Dhaariyat",  englishTranslation: "The Winnowing Winds",  revelationType: "Meccan",  verseCount: 60 },
  52: { englishName: "At-Toor",        englishTranslation: "The Mount",            revelationType: "Meccan",  verseCount: 49 },
  53: { englishName: "An-Najm",        englishTranslation: "The Star",             revelationType: "Meccan",  verseCount: 62 },
  54: { englishName: "Al-Qamar",       englishTranslation: "The Moon",             revelationType: "Meccan",  verseCount: 55 },
  55: { englishName: "Ar-Rahmaan",     englishTranslation: "The Beneficent",       revelationType: "Medinan", verseCount: 78 },
  56: { englishName: "Al-Waaqia",      englishTranslation: "The Inevitable",       revelationType: "Meccan",  verseCount: 96 },
  57: { englishName: "Al-Hadeed",      englishTranslation: "The Iron",             revelationType: "Medinan", verseCount: 29 },
  58: { englishName: "Al-Mujaadila",   englishTranslation: "The Pleading Woman",   revelationType: "Medinan", verseCount: 22 },
  59: { englishName: "Al-Hashr",       englishTranslation: "The Exile",            revelationType: "Medinan", verseCount: 24 },
  60: { englishName: "Al-Mumtahana",   englishTranslation: "She That is to be Examined", revelationType: "Medinan", verseCount: 13 },
  61: { englishName: "As-Saff",        englishTranslation: "The Ranks",            revelationType: "Medinan", verseCount: 14 },
  62: { englishName: "Al-Jumu'a",      englishTranslation: "Friday",               revelationType: "Medinan", verseCount: 11 },
  63: { englishName: "Al-Munaafiqoon", englishTranslation: "The Hypocrites",       revelationType: "Medinan", verseCount: 11 },
  64: { englishName: "At-Taghaabun",   englishTranslation: "Mutual Disillusion",   revelationType: "Medinan", verseCount: 18 },
  65: { englishName: "At-Talaaq",      englishTranslation: "The Divorce",          revelationType: "Medinan", verseCount: 12 },
  66: { englishName: "At-Tahreem",     englishTranslation: "The Prohibition",      revelationType: "Medinan", verseCount: 12 },
  67: { englishName: "Al-Mulk",        englishTranslation: "The Sovereignty",      revelationType: "Meccan",  verseCount: 30 },
  68: { englishName: "Al-Qalam",       englishTranslation: "The Pen",              revelationType: "Meccan",  verseCount: 52 },
  69: { englishName: "Al-Haaqqa",      englishTranslation: "The Reality",          revelationType: "Meccan",  verseCount: 52 },
  70: { englishName: "Al-Ma'aarij",    englishTranslation: "The Ascending Stairways", revelationType: "Meccan", verseCount: 44 },
  71: { englishName: "Nooh",           englishTranslation: "Noah",                 revelationType: "Meccan",  verseCount: 28 },
  72: { englishName: "Al-Jinn",        englishTranslation: "The Jinn",             revelationType: "Meccan",  verseCount: 28 },
  73: { englishName: "Al-Muzzammil",   englishTranslation: "The Enshrouded One",   revelationType: "Meccan",  verseCount: 20 },
  74: { englishName: "Al-Muddaththir", englishTranslation: "The Cloaked One",      revelationType: "Meccan",  verseCount: 56 },
  75: { englishName: "Al-Qiyaama",     englishTranslation: "The Resurrection",     revelationType: "Meccan",  verseCount: 40 },
  76: { englishName: "Al-Insaan",      englishTranslation: "Man",                  revelationType: "Medinan", verseCount: 31 },
  77: { englishName: "Al-Mursalaat",   englishTranslation: "The Emissaries",       revelationType: "Meccan",  verseCount: 50 },
  78: { englishName: "An-Naba",        englishTranslation: "The Announcement",     revelationType: "Meccan",  verseCount: 40 },
  79: { englishName: "An-Naazi'aat",   englishTranslation: "Those who Drag Forth", revelationType: "Meccan",  verseCount: 46 },
  80: { englishName: "Abasa",          englishTranslation: "He Frowned",           revelationType: "Meccan",  verseCount: 42 },
  81: { englishName: "At-Takweer",     englishTranslation: "The Overthrowing",     revelationType: "Meccan",  verseCount: 29 },
  82: { englishName: "Al-Infitaar",    englishTranslation: "The Cleaving",         revelationType: "Meccan",  verseCount: 19 },
  83: { englishName: "Al-Mutaffifeen", englishTranslation: "Defrauding",            revelationType: "Meccan",  verseCount: 36 },
  84: { englishName: "Al-Inshiqaaq",   englishTranslation: "The Splitting Open",    revelationType: "Meccan",  verseCount: 25 },
  85: { englishName: "Al-Burooj",      englishTranslation: "The Constellations",    revelationType: "Meccan",  verseCount: 22 },
  86: { englishName: "At-Taariq",      englishTranslation: "The Morning Star",      revelationType: "Meccan",  verseCount: 17 },
  87: { englishName: "Al-A'laa",       englishTranslation: "The Most High",         revelationType: "Meccan",  verseCount: 19 },
  88: { englishName: "Al-Ghaashiya",   englishTranslation: "The Overwhelming",      revelationType: "Meccan",  verseCount: 26 },
  89: { englishName: "Al-Fajr",        englishTranslation: "The Dawn",              revelationType: "Meccan",  verseCount: 30 },
  90: { englishName: "Al-Balad",       englishTranslation: "The City",              revelationType: "Meccan",  verseCount: 20 },
  91: { englishName: "Ash-Shams",      englishTranslation: "The Sun",               revelationType: "Meccan",  verseCount: 15 },
  92: { englishName: "Al-Lail",        englishTranslation: "The Night",             revelationType: "Meccan",  verseCount: 21 },
  93: { englishName: "Ad-Dhuhaa",      englishTranslation: "The Morning Hours",     revelationType: "Meccan",  verseCount: 11 },
  94: { englishName: "Ash-Sharh",      englishTranslation: "The Consolation",       revelationType: "Meccan",  verseCount: 8 },
  95: { englishName: "At-Teen",        englishTranslation: "The Fig",               revelationType: "Meccan",  verseCount: 8 },
  96: { englishName: "Al-Alaq",        englishTranslation: "The Clot",              revelationType: "Meccan",  verseCount: 19 },
  97: { englishName: "Al-Qadr",        englishTranslation: "The Power, Fate",       revelationType: "Meccan",  verseCount: 5 },
  98: { englishName: "Al-Bayyina",     englishTranslation: "The Evidence",          revelationType: "Medinan", verseCount: 8 },
  99: { englishName: "Az-Zalzala",     englishTranslation: "The Earthquake",        revelationType: "Medinan", verseCount: 8 },
  100:{ englishName: "Al-Aadiyaat",    englishTranslation: "The Courser",           revelationType: "Meccan",  verseCount: 11 },
  101:{ englishName: "Al-Qaari'a",     englishTranslation: "The Calamity",          revelationType: "Meccan",  verseCount: 11 },
  102:{ englishName: "At-Takaathur",   englishTranslation: "Competition",           revelationType: "Meccan",  verseCount: 8 },
  103:{ englishName: "Al-Asr",         englishTranslation: "The Declining Day",     revelationType: "Meccan",  verseCount: 3 },
  104:{ englishName: "Al-Humaza",      englishTranslation: "The Traducer",          revelationType: "Meccan",  verseCount: 9 },
  105:{ englishName: "Al-Feel",        englishTranslation: "The Elephant",          revelationType: "Meccan",  verseCount: 5 },
  106:{ englishName: "Quraish",        englishTranslation: "Quraish",               revelationType: "Meccan",  verseCount: 4 },
  107:{ englishName: "Al-Maa'oon",     englishTranslation: "Small Kindnesses",      revelationType: "Meccan",  verseCount: 7 },
  108:{ englishName: "Al-Kawthar",     englishTranslation: "Abundance",             revelationType: "Meccan",  verseCount: 3 },
  109:{ englishName: "Al-Kaafiroon",   englishTranslation: "The Disbelievers",      revelationType: "Meccan",  verseCount: 6 },
  110:{ englishName: "An-Nasr",        englishTranslation: "Divine Support",        revelationType: "Medinan", verseCount: 3 },
  111:{ englishName: "Al-Masad",       englishTranslation: "The Palm Fibre",        revelationType: "Meccan",  verseCount: 5 },
  112:{ englishName: "Al-Ikhlaas",     englishTranslation: "Sincerity",             revelationType: "Meccan",  verseCount: 4 },
  113:{ englishName: "Al-Falaq",       englishTranslation: "The Dawn",              revelationType: "Meccan",  verseCount: 5 },
  114:{ englishName: "An-Naas",        englishTranslation: "Mankind",               revelationType: "Meccan",  verseCount: 6 },
};

// ── Helpers ─────────────────────────────────────────────────────

function esc(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

function opt(s: string | undefined | null): string {
  if (!s || s === "(null)" || s === "null" || s === "None") return "";
  return esc(s.trim());
}

function attr(name: string, value: string): string {
  return value ? ` ${name}="${value}"` : "";
}

// ── Main ────────────────────────────────────────────────────────

async function main() {
  console.log("Building quran-detailed.xml...");

  // 1. Load root_meaning.csv → Buckwalter root → Arabic root
  console.log("  Loading root_meaning.csv...");
  const rootCsv = fs.readFileSync(path.join(DATA_DIR, "morphology", "root_meaning.csv"), "utf-8");
  const { data: rootRows } = parse<Record<string, string>>(rootCsv, { header: true, skipEmptyLines: true });
  const rootArabicMap = new Map<string, string>(); // Buckwalter → Arabic
  for (const row of rootRows) {
    const r = row["root"]?.trim();
    const ar = row["root_arabic"]?.trim();
    if (r && ar) rootArabicMap.set(r, ar);
  }
  console.log(`    ${rootArabicMap.size} root Buckwalter→Arabic mappings`);

  // 2. Load corpus-roots.json → invert to word→root lookup
  console.log("  Loading corpus-roots.json...");
  const rootsJson = JSON.parse(fs.readFileSync(path.join(DATA_DIR, "morphology", "corpus-roots.json"), "utf-8"));
  // Invert: "surah:verse:word" → root (Buckwalter)
  const wordRootMap = new Map<string, string>();
  for (const [root, refs] of Object.entries(rootsJson)) {
    for (const ref of refs as string[]) {
      wordRootMap.set(ref, root);
    }
  }
  console.log(`    ${wordRootMap.size} word→root mappings from ${Object.keys(rootsJson).length} roots`);

  // 3. Load MASAQ.csv → segment-level morphology
  console.log("  Loading MASAQ.csv...");
  const masaCsv = fs.readFileSync(path.join(DATA_DIR, "morphology", "MASAQ.csv"), "utf-8");
  const { data: masaData } = parse<Record<string, string>>(masaCsv, { header: true, skipEmptyLines: true });

  // Group segments by word key: surah:verse:word → segments[]
  const wordSegmentMap = new Map<string, {
    word: string; withoutDiacritics: string; gloss: string; punctuationMark: string;
    segments: SegmentData[];
  }>();
  for (const row of masaData) {
    const s = parseInt(row["Sura_No"], 10);
    const v = parseInt(row["Verse_No"], 10);
    const w = parseInt(row["Word_No"], 10);
    const segNo = parseInt(row["Segment_No"], 10);
    const key = `${s}:${v}:${w}`;

    if (!wordSegmentMap.has(key)) {
      wordSegmentMap.set(key, {
        word: row["Word"] || "",
        withoutDiacritics: row["Without_Diacritics"] || "",
        gloss: row["Gloss"] || "",
        punctuationMark: row["Punctuation_Mark"] || "",
        segments: [],
      });
    }

    wordSegmentMap.get(key)!.segments.push({
      number: segNo,
      text: opt(row["Segmented_Word"]),
      morphTag: opt(row["Morph_Tag"]),
      morphType: opt(row["Morph_Type"]),
      syntacticRole: opt(row["Syntactic_Role"]),
      possessiveConstruct: opt(row["Possessive_Construct"]),
      caseMood: opt(row["Case_Mood"]),
      caseMoodMarker: opt(row["Case_Mood_Marker"]),
      invariableDeclinable: opt(row["Invariable_Declinable"]),
      phrase: opt(row["Phrase"]),
      phrasalFunction: opt(row["Phrasal_Function"]),
      gloss: opt(row["Gloss"]),
    });
  }
  console.log(`    ${wordSegmentMap.size} words with ${masaData.length} total segments`);

  // 4. Load quran-uthmani.xml → verse structure + text
  console.log("  Loading quran-uthmani.xml...");
  const quranXml = fs.readFileSync(path.join(DATA_DIR, "quran", "quran-uthmani.xml"), "utf-8");
  const xmlParser = new XMLParser({ ignoreAttributes: false, attributeNamePrefix: "@_" });
  const quranParsed = xmlParser.parse(quranXml);
  const suras = quranParsed.quran.sura as Array<Record<string, unknown>>;

  // 5. Write the XML
  console.log("  Writing quran-detailed.xml...");
  const out = fs.createWriteStream(OUT_PATH, { encoding: "utf-8" });

  out.write('<?xml version="1.0" encoding="UTF-8"?>\n');
  out.write('<quran>\n');

  for (const sura of suras) {
    const surahId = Number(sura["@_index"]);
    const surahName = String(sura["@_name"] || "");
    const meta = SURAH_META[surahId] || { englishName: surahName, englishTranslation: "", revelationType: "Meccan", verseCount: 0 };
    const ayas = (Array.isArray(sura.aya) ? sura.aya : [sura.aya]) as Array<Record<string, unknown>>;

    out.write(`  <surah index="${surahId}" name="${esc(surahName)}" englishName="${esc(meta.englishName)}" englishTranslation="${esc(meta.englishTranslation)}" revelationType="${meta.revelationType}" verseCount="${meta.verseCount}">\n`);

    for (const aya of ayas) {
      if (!aya) continue;
      const verseNum = Number(aya["@_index"]);
      const verseText = String(aya["@_text"] || "");

      out.write(`    <verse index="${verseNum}" text="${esc(verseText)}">\n`);

      // Get words for this verse from MASAQ
      // Collect all word keys for this verse
      const wordKeys: string[] = [];
      for (const key of wordSegmentMap.keys()) {
        const [ks, kv] = key.split(":").map(Number);
        if (ks === surahId && kv === verseNum) wordKeys.push(key);
      }
      wordKeys.sort((a, b) => {
        const wa = parseInt(a.split(":")[2], 10);
        const wb = parseInt(b.split(":")[2], 10);
        return wa - wb;
      });

      for (const wk of wordKeys) {
        const wData = wordSegmentMap.get(wk)!;
        const wNum = parseInt(wk.split(":")[2], 10);

        // Get root for this word from corpus-roots.json (inverted lookup)
        const rootBw = wordRootMap.get(`${surahId}:${verseNum}:${wNum}`) || "";
        const rootAr = rootBw ? (rootArabicMap.get(rootBw) || "") : "";

        // Deduplicate segments by number
        const seenSegs = new Set<number>();
        const uniqueSegs = wData.segments.filter((s) => {
          if (seenSegs.has(s.number)) return false;
          seenSegs.add(s.number);
          return true;
        });

        out.write(`      <word index="${wNum}" token="${esc(wData.word)}" withoutDiacritics="${esc(wData.withoutDiacritics)}" translation="${esc(wData.gloss)}"${attr("rootBuckwalter", rootBw)}${attr("rootArabic", rootAr)}${attr("punctuationMark", wData.punctuationMark)}>\n`);

        for (const seg of uniqueSegs) {
          const attrs = [
            attr("index", String(seg.number)),
            attr("text", seg.text),
            attr("morphTag", seg.morphTag),
            attr("morphType", seg.morphType),
            attr("syntacticRole", seg.syntacticRole),
            attr("possessiveConstruct", seg.possessiveConstruct),
            attr("caseMood", seg.caseMood),
            attr("caseMoodMarker", seg.caseMoodMarker),
            attr("invariableDeclinable", seg.invariableDeclinable),
            attr("phrase", seg.phrase),
            attr("phrasalFunction", seg.phrasalFunction),
            attr("gloss", seg.gloss),
          ].join("");
          out.write(`        <segment${attrs} />\n`);
        }

        out.write(`      </word>\n`);
      }

      out.write(`    </verse>\n`);
    }

    out.write(`  </surah>\n`);
  }

  out.write('</quran>\n');

  // Wait for stream to finish
  await new Promise<void>((resolve, reject) => {
    out.end(() => resolve());
    out.on("error", reject);
  });

  // Stats
  const stats = fs.statSync(OUT_PATH);
  console.log(`\n✅ quran-detailed.xml created!`);
  console.log(`  Path: ${OUT_PATH}`);
  console.log(`  Size: ${(stats.size / 1024 / 1024).toFixed(1)} MB`);
}

interface SegmentData {
  number: number;
  text: string;
  morphTag: string;
  morphType: string;
  syntacticRole: string;
  possessiveConstruct: string;
  caseMood: string;
  caseMoodMarker: string;
  invariableDeclinable: string;
  phrase: string;
  phrasalFunction: string;
  gloss: string;
}

main().catch(console.error);
