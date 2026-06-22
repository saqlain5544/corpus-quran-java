import fs from "node:fs/promises";
import path from "node:path";

const QURAN_HTML_DIR = path.resolve(process.cwd(), "..", "quran-html");

export const SURAH_NAMES_INTERNAL: Record<number, string> = {
  1: "الفاتحة",
  2: "البقرة",
  3: "آل-عمران",
  4: "النساء",
  5: "المائدة",
  6: "الأنعام",
  7: "الأعراف",
  8: "الأنفال",
  9: "التوبة",
  10: "يونس",
  11: "هود",
  12: "يوسف",
  13: "الرعد",
  14: "ابراهيم",
  15: "الحجر",
  16: "النحل",
  17: "الإسراء",
  18: "الكهف",
  19: "مريم",
  20: "طه",
  21: "الأنبياء",
  22: "الحج",
  23: "المؤمنون",
  24: "النور",
  25: "الفرقان",
  26: "الشعراء",
  27: "النمل",
  28: "القصص",
  29: "العنكبوت",
  30: "الروم",
  31: "لقمان",
  32: "السجدة",
  33: "الأحزاب",
  34: "سبإ",
  35: "فاطر",
  36: "يس",
  37: "الصافات",
  38: "ص",
  39: "الزمر",
  40: "غافر",
  41: "فصلت",
  42: "الشورى",
  43: "الزخرف",
  44: "الدخان",
  45: "الجاثية",
  46: "الأحقاف",
  47: "محمد",
  48: "الفتح",
  49: "الحجرات",
  50: "ق",
  51: "الذاريات",
  52: "الطور",
  53: "النجم",
  54: "القمر",
  55: "الرحمن",
  56: "الواقعة",
  57: "الحديد",
  58: "المجادلة",
  59: "الحشر",
  60: "الممتحنة",
  61: "الصف",
  62: "الجمعة",
  63: "المنافقون",
  64: "التغابن",
  65: "الطلاق",
  66: "التحريم",
  67: "الملك",
  68: "القلم",
  69: "الحاقة",
  70: "المعارج",
  71: "نوح",
  72: "الجن",
  73: "المزمل",
  74: "المدثر",
  75: "القيامة",
  76: "الانسان",
  77: "المرسلات",
  78: "النبإ",
  79: "النازعات",
  80: "عبس",
  81: "التكوير",
  82: "الإنفطار",
  83: "المطففين",
  84: "الإنشقاق",
  85: "البروج",
  86: "الطارق",
  87: "الأعلى",
  88: "الغاشية",
  89: "الفجر",
  90: "البلد",
  91: "الشمس",
  92: "الليل",
  93: "الضحى",
  94: "الشرح",
  95: "التين",
  96: "العلق",
  97: "القدر",
  98: "البينة",
  99: "الزلزلة",
  100: "العاديات",
  101: "القارعة",
  102: "التكاثر",
  103: "العصر",
  104: "الهمزة",
  105: "الفيل",
  106: "قريش",
  107: "الماعون",
  108: "الكوثر",
  109: "الكافرون",
  110: "النصر",
  111: "المسد",
  112: "الإخلاص",
  113: "الفلق",
  114: "الناس",
};

/**
 * Read a surah's HTML file from the sibling `quran-html/` directory.
 * Returns the raw HTML string (full document) or null if not found.
 */
export async function readSurahHtml(surahNumber: number): Promise<string | null> {
  const name = SURAH_NAMES_INTERNAL[surahNumber];
  if (!name) return null;

  const filename = `${String(surahNumber).padStart(3, "0")}-${name}.html`;
  const filePath = path.join(QURAN_HTML_DIR, filename);

  try {
    return await fs.readFile(filePath, "utf-8");
  } catch {
    return null;
  }
}

/**
 * Extract the contents of the <main class="quran">...</main> block from
 * a full surah HTML document. Returns null if no <main> is found.
 * Tolerant of newlines/whitespace between `<main` and its attributes.
 */
export function extractMain(html: string): string | null {
  const openMatch = html.match(/<main\b[^>]*class="quran"[^>]*>/);
  if (!openMatch) return null;
  const open = openMatch.index!;
  const openEnd = open + openMatch[0].length;
  const close = html.indexOf("</main>", openEnd);
  if (close === -1) return null;
  return html.slice(open, close + "</main>".length);
}

/**
 * Read the surah's <main> HTML — the body the reader renders directly.
 * Returns null if either the file or the <main> block is missing.
 */
export async function readSurahMain(surahNumber: number): Promise<string | null> {
  const html = await readSurahHtml(surahNumber);
  if (!html) return null;
  return extractMain(html);
}

/**
 * Extract <title>...</title> contents from the surah HTML for metadata.
 */
export function extractTitle(html: string): string | null {
  const m = html.match(/<title>\s*([\s\S]*?)\s*<\/title>/);
  return m ? m[1].trim() : null;
}