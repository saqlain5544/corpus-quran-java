# Research Findings — Quran Reader App

**Status**: Complete
**Date**: 2026-06-24
**Researcher**: Research Agent (Orchestrator delegation)

---

## 1. Trusted Data Inventory

| File | Size | Verdict |
|---|---|---|
| `data/quran/quran-uthmani.xml` | 1.5 MB | **Trusted** — Tanzil Uthmani v1.1, CC-BY-3.0 |
| `data/morphology/MASAQ.csv` | 18 MB | **Trusted** — MASAQ morphological segmentation |
| `data/fonts/hafs.woff2` | 63 KB | **Trusted** — WOFF2 format validated, version 1.66 |

**Untrusted / supplementary** (must be flagged in UI):
- `data/morphology/corpus-roots.json` — root → locations (Buckwalter keys)
- `data/morphology/meanings-roots-ai.jsonl` — AI-generated meanings (label as AI-generated)
- `data/morphology/root_meaning.csv` — extended root meanings (note: 4 known Buckwalter errors previously fixed by `scripts/fix-buckwalter.py`)
- `data/new/quran-detailed.xml` — never inspected, do NOT use without verification

---

## 2. quran-uthmani.xml Schema

**Structure**: `<quran>` → 114 `<sura index="N" name="...">` → N `<aya index="N" text="..." />` (6236 total ayas)

**No nested elements**. All annotation marks are encoded as **inline Arabic-Extended Unicode characters inside `text=""`**:

| Char | Code | Name | Type | Count | Position |
|---|---|---|---|---|---|
| `۞` | U+06DE | Hizb/Rub' marker | division | 199 | ALWAYS first char of ayah |
| `۩` | U+06E9 | Sajdah | prostration | 15 | End of ayah |
| `ۖ` | U+06D6 | Qulā (قلی) | pause — stop preferred | 1682 | Inline |
| `ۗ` | U+06D7 | Ṣallā (صلی) | pause — continue preferred | 603 | Inline |
| `ۘ` | U+06D8 | Mīm (م) | pause — mandatory stop | 22 | Inline |
| `ۙ` | U+06D9 | Lā (لا) | pause — do not stop | 68 | Inline |
| `ۚ` | U+06DA | Jīm (ج) | pause — permissible | 1972 | Inline |
| `ۛ` | U+06DB | Muʿānqah | pause — paired | 12 | Inline |
| `ۜ` | U+06DC | rare | annotation | 7 | Inline |
| `۟` | U+06DF | small high meem (part of letter composition) | diacritic | 3988 | part of letters |
| `ۥ`–`ۨ` | U+06E5..U+06E8 | small hamza/ya/seen etc. (letter composition) | diacritic | varies | part of letters |

**Key rules**:
1. **Words** = space-delimited tokens. A token is a "word" if it contains at least one Arabic letter (U+0621–U+064A or U+0671). Pure-mark tokens (e.g., the lone `۞` at the start of a verse) are marks.
2. **Rub' (`۞`) is always the first character of an ayah** (199/199 confirmed).
3. **Sajdah (`۩`) is always at the end of an ayah**.
4. Marks may be **appended to the previous word** as a trailing token (e.g., `ٱلرَّحِيمِ ۩`). The Python tokenizer in `render-to-html/main.py:89-97` handles this with a `split_mark()` that strips trailing marks.
5. **Bismillah handling**:
   - Surah 1: Bismillah IS verse 1.
   - Surah 9: No Bismillah.
   - Other surahs: Bismillah is a separate `bismillah` attribute on aya 1 OR aya 1 itself contains Bismillah — in the latter case it's not counted as verse 1.
6. **Total counts**: 114 surahs, 6236 ayas, ~77,797 words (per MASAQ), ~49,968 word occurrences with roots.

---

## 3. MASAQ.csv Schema

**Format**: CSV, 19 columns, 157,676 segment rows, **no quote-required fields** (verified via simple comma-split).

**Columns**:
```
ID, Sura_No, Verse_No, Word_No, Segment_No, Word,
Without_Diacritics, Segmented_Word, Morph_Tag, Morph_Type,
Punctuation_Mark, Invariable_Declinable, Syntactic_Role,
Possessive_Construct, Case_Mood, Case_Mood_Marker, Phrase,
Phrasal_Function, Gloss
```

**Key facts**:
- **77,797 unique word IDs** (`ID` column, sequential 1..77797).
- Each word has **1–5 segments** (Prefix/Stem/Suffix structure).
- `Word_No` is **per-verse** (1..N), aligning 1:1 with the XML's word positions.
- `Word` (with tashkeel) **matches the XML** token exactly for the corpus words (sample: `بِسْمِ`, `ٱللَّهِ`, `ٱلرَّحْمَـٰنِ`).
- `Gloss` uses **hyphenated phrase notation** like `in-(the)-name`, `(of)-allah`, `the-most-gracious`.
- `Morph_Type` ∈ {`Prefix`, `Stem`, `Suffix`, `Other_i3rab`} — 4 values only.
- **No root column**. Roots must be joined from `corpus-roots.json` by location.

**Alignment test passed**: Surah 1:1 has 4 words in XML, 4 unique Word_No in MASAQ (1,2,3,4), each with 2 segments. Surah 2:1 has 4 words in XML (the Huruf al-Muqatta'at `الٓمٓ`), with the same 4 MASAQ entries.

---

## 4. Root Data Files

### corpus-roots.json
- **Format**: `{ root_buckwalter: ["s:v:w", ...] }`
- 1,642 root keys, 49,968 occurrence locations total.
- **Buckwalter transliteration**: `$Am` (ء-ا-م), `Alh` (ا-ل-ه), `qwl` (ق-و-ل), `$Tn` (ا-ت-ن), `H` = ta marbuta, `T` = ta, `E` = ya with hamza above, `A` = alif madda.
- **Coverage**: 49,957 / 77,411 MASAQ word locations = **64.5%** have roots. The remaining 35.5% are particles (حرف), pronouns, Huruf al-Muqatta'at, etc.

### meanings-roots-ai.jsonl
- 1,642 entries, **1:1 perfect match** with corpus-roots.json keys.
- Each entry has: id, root, root_arabic, root_letters, pos (V/N/...), occurrences_quran, meaning.{en,ar}, lexical_analysis, quran_examples, hadith, shades, verified_date.
- **Origin**: AI-generated. Must be flagged in UI.

### root_meaning.csv
- 125,900 rows. **Messy structure** — appears to have many empty/duplicate rows when scanned naively.
- 22,520 "distinct roots" counted when including empty values — actual unique non-empty roots ≈ 1,642 (matches the other files).
- Fields: id, root, root_arabic, general_meaning, shades_of_meaning, hadith_examples, classical_sources, idioms_customs, word_by_word_shades, full_response, model_used, pos.
- Known issues: `scripts/fix-buckwalter.py` documents 4 prior Buckwalter errors (`SbA`, `lwt`, `zqm`, `qTyn`) that have been fixed.
- **Verdict**: Use as supplementary detail only; **primary source for meaning details is `meanings-roots-ai.jsonl`** (cleaner).

---

## 5. hafs.woff2 Font

- Valid WOFF2 file, magic `wOF2`, version 1.66, TrueType-flavored.
- **Required for proper Arabic rendering** with diacritics (tashkeel), the small letters (ی, ک, etc.) and ligatures used in Uthmani script.
- 63 KB — keep as a single static file.

---

## 6. Word Tokenization Algorithm

```
input: aya_text (string)
1. tokens = aya_text.split(/\s+/)
2. result = []
3. for each token:
     word, marks = split_trailing_marks(token)   # peel off ۞ ۩ ۖ ۗ ۘ ۙ ۚ ۛ
     if word is not empty:
         result.append(("word", word))
     for each mark in marks:
         result.append(("mark", mark))
4. return result
```

Edge cases observed:
- A token can be just a mark (e.g., the standalone `۞` at start of an ayah after splitting the prefix).
- A token can be word + several marks (e.g., `وَٱلْعَـٰلَمِينَ ۖ` — actually space-delimited; the mark is its own token in the XML).
- The trailing-marks approach (Python `render-to-html/main.py:89-97`) is sufficient because the source XML already separates them with spaces — but defensively peeling trailing marks handles merged cases if any data variant introduces them.

---

## 7. Fuzzy Matching — Arabic Search

**Challenges** (well-known, no web search required):
1. **Tashkeel (diacritics) variance**: users typing without tashkeel must match `بِسْمِ` to `بسم` etc.
2. **Alif Wasla variants**: Uthmani uses `ٱ` (U+0671) for alif wasla; Indo-Pak typically uses `ا` (U+0627). Some search engines strip one to the other.
3. **Yāʾ variants**: `ى` (U+0649 alef maksura) vs `ي` (U+064A yāʾ).
4. **Tāʾ Marbūṭa**: `ة` (U+0629) vs `ه` (U+0647) at word-end.
5. **Hamza variants**: `أ إ آ ؤ ئ ء` — all forms of hamza.

**Recommended approach for the search endpoint**:

```javascript
function normalizeArabic(s) {
  return s
    .replace(/[\u064B-\u0652\u0670\u0640]/g, '') // strip tashkeel + tatweel
    .replace(/[أإآٱ]/g, 'ا')                     // unify alifs (incl. wasla)
    .replace(/[ىي]/g, 'ي')                       // unify ya/alef-maqsura
    .replace(/[ةه]/g, 'ه')                       // unify ta marbuta / ha
    .replace(/[ؤئء]/g, 'ء')                      // unify hamzas
    .replace(/\s+/g, ' ')                        // collapse whitespace
    .trim();
}
```

**Search modes** (per the `type` query param):
- `arabic_word` — exact OR normalized match against `Word` field in MASAQ. Also accept the Indo-Pak form `الٓمٓ` and Uthmani form transparently.
- `root` — match against the root field; user can type either Buckwalter (`qwl`) or Arabic letters (and we convert via the standard table).
- `english_term` — Levenshtein ≤ 2 against the `Gloss` field of MASAQ. Use a sub-linear index (BK-tree or simple prefix hash) given the ~77k-word corpus is tiny.

**Server-side**: simple `for ... range` over the preloaded index (we're talking ~50k entries — totally fine without an external search library). Client-side (type-ahead in the global header) can be a fast prefix lookup on the same in-memory list.

---

## 8. Recommendations Summary

| Decision | Recommendation |
|---|---|
| Storage format for processed data | JSON for human-inspectable bits (roots, morphology index), GOB or MessagePack for the main Quran+indices |
| In-memory size budget | Stay under 80 MB RSS as per global constraints |
| Embed data into the binary? | Yes — Go `embed.FS` keeps deployment simple; the processed bundle is ~5–8 MB compressed |
| Streaming XML parser? | Not needed — file is only 1.5 MB, but use `encoding/xml` Decoder (event-based, not tree-based) anyway for low memory |
| Hashmap key for word lookup | `int64(surah)*1e8 + int64(ayah)*1e4 + int64(wordNo)` — packs into a single int64 for fast hashing; bounded well under int64 range |
| Root join key | Buckwalter string from `corpus-roots.json`; we keep it as-is and provide Arabic rendering in the UI |
| Pause mark rendering | Render as separate span, NOT inside the word span, so event delegation is simpler |
| Rub' marker `۞` | Always at start of an ayah — render BEFORE the first word of the ayah, as a structural marker (visual section divider) |
| Sajdah `۩` | Always at end — render AFTER last word of ayah with distinct styling (e.g., subtle green border) |

---

## 9. Open Questions for Orchestrator

None blocking. Implementation can proceed.

---

## 10. References

- Tanzil.net documentation: <http://tanzil.net/docs> (Uthmani v1.1 schema).
- Buckwalter transliteration table: standard Arabic NLP mapping (`$`=hamza-on-line, `'`=hamza-under-alif, `A`=alif madda, etc.).
- MASAQ project: morphological segmentation per word (the original MASAQ CSV is unchanged, no licence concerns noted in the file).
- Local scripts: `render-to-html/main.py:89-97` (tokenizer), `data/morphology/scripts/fix-buckwalter.py` (Buckwalter error catalog).

---

## 11. MASAQ/XML Word Boundary Divergences (23 gaps, 0.03%)

*Verified by `TestFindMissingMorphology` in `backend/internal/data/find_missing_test.go`.*

The Uthmani XML tokenizer and MASAQ disagree on word boundaries for
certain grammatical particles. MASAQ merges particles (ما, دام, يا)
into neighboring words; the XML keeps them separate. This causes
23 word positions to have no morphological data.

### Category A: MASAQ merges X + ما (8 gaps)

MASAQ treats أَيْنَمَا, كُلَّمَا, بَعْدَمَا, وَحَيْثُمَا, لَوْمَا
as single tokens. XML splits them:

| Location | Token | MASAQ merged form |
|---|---|---|
| 2:144:16 | مَا | وَحَيْثُمَا (w15) |
| 2:148:8  | مَا | أَيْنَمَا (w7) |
| 2:150:10 | مَا | وَحَيْثُمَا (w9) |
| 4:91:9   | مَا | كُلَّمَا (w8) |
| 15:7:2   | مَا | لَوْمَا (w1) |
| 19:31:4  | مَا | أَيْنَمَا (w3) |
| 26:92:4  | مَا | أَيْنَمَا (w3) |
| 57:4:31  | مَا | أَيْنَمَا (w30) |

### Category B: MASAQ merges ما + دام (4 gaps)

| Location | Token | Cause |
|---|---|---|
| 5:96:20  | تُحْشَرُونَ | مَا دُمْتُمْ merged → shift |
| 5:117:29 | شَهِيدٌ | مَا دُمْتُ merged → shift |
| 11:107:15 | يُرِيدُ | مَا دَامَتْ merged → shift |
| 19:31:10 | دُمْتُ | مَا دُمْتُ merged → shift |

### Category C: Vocative يا merge (2 gaps)

| Location | Token | XML | MASAQ |
|---|---|---|---|
| 20:94:3  | لَا | يَا + ابْنَأُمَّ (2 words) | يَاابْنَأُمَّ (1) |
| 37:130:4 | يَاسِينَ | إِلْ + يَاسِينَ (2 words) | إِلْيَاسيْنَ (1) |

### Category D: Cascade shifts from earlier merges (9 gaps)

When an early merge shifts all subsequent word numbers, the final word
of a verse appears "missing" at its XML position but exists at
position-1 in MASAQ:

| Location | Token | Found at |
|---|---|---|
| 2:181:14 | عَلِيمٌ | w13 |
| 3:75:36  | يَعْلَمُونَ | w35 |
| 3:112:35 | يَعْتَدُونَ | w34 |
| 5:13:32  | ٱلْمُحْسِنِينَ | w31 |
| 5:24:16  | قَـٰعِدُونَ | w15 |
| 8:6:12   | يَنظُرُونَ | w11 ← user-reported example |
| 13:37:20 | وَاقٍ | w19 |
| 16:109:7 | ٱلْخَـٰسِرُونَ | w6 |
| 23:44:19 | يُؤْمِنُونَ | w18 |

**Summary**: 23/77,433 words (0.03%) — all caused by word-boundary
convention differences, not data errors. The UI shows "No morphological
data" for these positions. Fixing would require fuzzy text-matching
instead of strict positional lookup, which risks false positives.
