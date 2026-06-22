/**
 * Arabic text alignment for Quran rendering.
 *
 * Strategy: align by INDEX. The verse text (split on whitespace, dropping
 * pure-pause tokens) and the DB word list are 99.6% 1-to-1 aligned. The
 * remaining 0.4% fall into two patterns:
 *
 *   1. Compound SPLIT (text has more tokens than DB)
 *      e.g. text "أَيْنَ مَا"  ↔  DB "أَيْنَمَا"
 *           text "وَحَيْثُ مَا"  ↔  DB "وَحَيْثُمَا"
 *           text "مَا دُمْتُ"   ↔  DB "مَا دُمْتُ"
 *
 *   2. Compound JOIN (DB has more words than text)
 *      e.g. text "يَـٰقَوْمِ"  ↔  DB "يَا" + "قَوْمِ"
 *
 * The align() function below tries three strategies in order:
 *   (a) direct 1-to-1 match (with edit-distance tolerance for diacritic noise)
 *   (b) compound JOIN  — text token contains DB[i] + DB[i+1]
 *   (c) compound SPLIT — accumulate multiple text tokens until they form DB[i]
 *   (d) fallback — use the next text token anyway so the page still renders
 *
 * No complex state machine, no per-character classification. Just
 * whitespace-split, normalize, and walk both lists in lockstep.
 */

// ── Diacritic / Alef variant stripping ─────────────────────────

/** Strip Quranic diacritics (tashkeel) and tatweel. */
export function stripDiacritics(s: string): string {
  return s.replace(/[\u064B-\u065F\u0610-\u061A\u0640]/g, "");
}

/**
 * Canonical form for matching. Convert all alef variants to one
 * letter so that "الرَّحْمَـٰنِ" matches "الرَّحْمَنِ" and
 * "الصِّرَاطَ" matches "ٱلصِّرَٰطَ".
 *
 *   ٰ (dagger alif U+0670) → ا (alef U+0627)
 *   ٱ (alef wasla U+0671) → ا
 *   ى (alef maqsura U+0649) → ي (yeh U+064A)
 *   آ (alef madda U+0622) → ا
 *
 * Dagger alif is converted (NOT stripped) because in words like
 * "الرَّحْمَـٰنِ" the dagger alif represents a real long alif
 * that the Hafs reading pronounces.
 */
export function normalize(s: string): string {
  return stripDiacritics(s)
    .replace(/\u0670/g, "\u0627")
    .replace(/\u0671/g, "\u0627")
    .replace(/\u0649/g, "\u064A")
    .replace(/\u0622/g, "\u0627");
}

// ── Levenshtein distance (with max cutoff) ──────────────────────

function levenshtein(a: string, b: string, maxDist = 1): number {
  if (Math.abs(a.length - b.length) > maxDist) return maxDist + 1;
  if (a.length === 0) return b.length;
  if (b.length === 0) return a.length;

  const m = a.length, n = b.length;
  // Use a single rolling row
  let prev = new Array<number>(n + 1);
  let curr = new Array<number>(n + 1);
  for (let j = 0; j <= n; j++) prev[j] = j;

  for (let i = 1; i <= m; i++) {
    curr[0] = i;
    let rowMin = curr[0];
    for (let j = 1; j <= n; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      curr[j] = Math.min(prev[j] + 1, curr[j - 1] + 1, prev[j - 1] + cost);
      if (curr[j] < rowMin) rowMin = curr[j];
    }
    if (rowMin > maxDist) return maxDist + 1;
    [prev, curr] = [curr, prev];
  }
  return prev[n];
}

// ── Token classification ──────────────────────────────────────────

const PAUSE_RANGE = [0x06d6, 0x06ed] as const;
const AYAH_OPEN = 0xfd3f;
const AYAH_CLOSE = 0xfd3e;

function isPauseChar(cp: number): boolean {
  return cp >= PAUSE_RANGE[0] && cp <= PAUSE_RANGE[1];
}
function isPauseToken(tok: string): boolean {
  if (tok.length === 0) return false;
  for (let i = 0; i < tok.length; i++) {
    if (!isPauseChar(tok.codePointAt(i)!)) return false;
  }
  return true;
}

export type QuranTokenType = "word" | "pause" | "ayah";

export interface QuranToken {
  type: QuranTokenType;
  text: string;
}

/**
 * Split verse text by whitespace, classify each token as:
 *   word   — anything containing letters (including pause marks embedded in words)
 *   pause  — a token consisting ONLY of Quranic pause marks (ۚ ۖ ۗ …)
 *   ayah   — a token that is ﴾N﴾ (an end-of-verse marker; rarely present in verse.text)
 */
export function tokenize(text: string): QuranToken[] {
  const out: QuranToken[] = [];
  for (const raw of text.split(/\s+/)) {
    if (!raw) continue;
    if (isPauseToken(raw)) {
      out.push({ type: "pause", text: raw });
    } else if (raw.includes("\uFD3F") || raw.includes("\uFD3E")) {
      out.push({ type: "ayah", text: raw });
    } else {
      out.push({ type: "word", text: raw });
    }
  }
  return out;
}

// ── Alignment ───────────────────────────────────────────────────

/** A DB word with its identifying metadata. */
export interface DBWord {
  number: number;
  token: string;
}

/** One alignment entry: how many text word tokens map to one DB word. */
export interface AlignmentEntry {
  /** DB word index in the input list. */
  dbIndex: number;
  /** The DB word record. */
  dbWord: DBWord;
  /** Which text token indices this DB word consumes (1 or more). */
  textIndices: number[];
  /** How the text token(s) were matched:
   *   "direct"  — single text token, 1-to-1
   *   "join"    — text token split: this DB word + next share one text token
   *   "split"   — text split: multiple text tokens form this DB word
   *   "fallback"— no match found; used the next text token to keep alignment moving
   */
  match: "direct" | "join" | "split" | "fallback";
}

/**
 * Align DB words to text word tokens. The result is a list with one
 * entry per DB word, in order. Each entry records which text indices
 * it consumed and how the match was made.
 *
 * The text-token list should already be filtered to `type === "word"`
 * (passes via {@link tokenize} + filter); pause and ayah tokens are not
 * consumed by alignment.
 */
export function align(
  dbWords: DBWord[],
  textWordTokens: string[],
  maxDist = 1
): AlignmentEntry[] {
  if (dbWords.length === 0) return [];

  const dbNorms = dbWords.map((w) => normalize(w.token));
  const textNorms = textWordTokens.map(normalize);

  const out: AlignmentEntry[] = [];
  let ti = 0;

  for (let wi = 0; wi < dbWords.length; wi++) {
    const dbNorm = dbNorms[wi];

    // Bail if text exhausted — render nothing for the remaining DB words.
    if (ti >= textWordTokens.length) {
      out.push({ dbIndex: wi, dbWord: dbWords[wi], textIndices: [], match: "fallback" });
      continue;
    }

    const twNorm = textNorms[ti];

    // Strategy (a): direct 1-to-1 match
    if (dbNorm === twNorm || levenshtein(dbNorm, twNorm, maxDist) <= maxDist) {
      out.push({ dbIndex: wi, dbWord: dbWords[wi], textIndices: [ti], match: "direct" });
      ti += 1;
      continue;
    }

    // Strategy (b): compound JOIN — text token contains DB[i] + DB[i+1]
    // (e.g., text "يَـٰقَوْمِ" → DB "يَا" + "قَوْمِ")
    if (
      wi + 1 < dbWords.length &&
      twNorm.length > dbNorm.length &&
      twNorm.startsWith(dbNorm)
    ) {
      const rest = twNorm.slice(dbNorm.length);
      const nextDbNorm = dbNorms[wi + 1];
      if (levenshtein(rest, nextDbNorm, maxDist) <= maxDist) {
        // Mark current DB aligned to text[ti]
        out.push({ dbIndex: wi, dbWord: dbWords[wi], textIndices: [ti], match: "join" });
        // Mark NEXT DB also aligned to text[ti] (shared token)
        out.push({
          dbIndex: wi + 1,
          dbWord: dbWords[wi + 1],
          textIndices: [ti],
          match: "join",
        });
        ti += 1;
        // Skip the next iteration since we already recorded DB[i+1]
        wi += 1;
        continue;
      }
    }

    // Strategy (c): compound SPLIT — accumulate multiple text tokens
    // until they form DB[i]
    let accumulated = "";
    const consumed: number[] = [];
    const tiStart = ti;
    while (ti < textWordTokens.length && consumed.length < 5) {
      accumulated += textNorms[ti];
      consumed.push(ti);
      ti += 1;
      if (
        dbNorm === accumulated ||
        levenshtein(dbNorm, accumulated, maxDist) <= maxDist
      ) {
        break;
      }
      if (accumulated.length > dbNorm.length + maxDist + 2) break;
    }
    if (
      consumed.length > 0 &&
      (dbNorm === accumulated ||
        levenshtein(dbNorm, accumulated, maxDist) <= maxDist)
    ) {
      out.push({
        dbIndex: wi,
        dbWord: dbWords[wi],
        textIndices: consumed,
        match: "split",
      });
      continue;
    }

    // Strategy (d): fallback — use next text token anyway
    out.push({
      dbIndex: wi,
      dbWord: dbWords[wi],
      textIndices: [tiStart],
      match: "fallback",
    });
    ti = tiStart + 1;
  }

  return out;
}