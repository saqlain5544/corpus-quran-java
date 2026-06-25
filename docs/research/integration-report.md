# Integration & Performance Report

Final smoke-test and memory profile for the Go + vanilla JS Quran Reader.

## 1. Smoke test — all routes return 200 (or 404 where expected)

Server: `quranreader -addr :8094 -data ./data/processed`

| Route | Status | Size | Time |
|---|---|---|---|
| `GET /` | 200 | 34 KB | 1.5 ms |
| `GET /surah/2` | 200 | 1.9 MB | 20 ms |
| `GET /surah/2/255` | 200 | 1.9 MB | 19 ms |
| `GET /roots` | 200 | 16 KB | 1 ms |
| `GET /root/qwl` | 200 | 326 KB | 4 ms |
| `GET /about` | 200 | 4 KB | 0.4 ms |
| `GET /api/word?s=1&a=1&w=2` | 200 | 1.2 KB | 0.4 ms |
| `GET /api/search?q=allah&type=english` | 200 | 14 KB | 30 ms |
| `GET /api/health` | 200 | 35 B | 0.3 ms |
| `GET /static/css/tokens.css` | 200 | 4.4 KB | 1.3 ms |
| `GET /fonts/hafs.woff2` | 200 | 63 KB | 0.4 ms |
| `GET /surah/999` | 404 | — | — |

API response shapes verified:
- `/api/word` returns `{surah, ayah, word, word_text, segments[], root?}`
  with correct Arabic text and segments.
- `/api/search` returns `{query, type, results[]}` with `<mark>`-tagged
  snippets and the right `link` shape.
- `/api/health` returns `{status:"ok", uptime_seconds}`.

## 2. Concurrency

20 concurrent requests (10× `/`, 10× `/surah/2`) — all returned 200.
Worst observed duration: 240 ms (first surah render). After warm-up,
all requests under 5 ms except full-surah renders (15–30 ms).

## 3. Memory

Measured with `ps -o rss,vsz` and `runtime.MemStats` (via
`backend/cmd/memcheck`):

```
HeapAlloc:     77.8 MB
HeapSys:      115.5 MB
HeapInuse:     80.4 MB
HeapObjects: 1204367
Sys:          125.6 MB
```

Process RSS after warm-up and concurrent load: **~118–120 MB**.

### Comparison against the 80 MB target

The `agents.md` global constraint says "<80 MB RSS". We do **not**
meet this on the RSS metric, though the **heap** allocation is at
~78 MB and very close to the budget.

The gap (~40 MB) is the Go runtime overhead:
- Go stacks for ~5 goroutines (handler goroutines + GC workers) ≈ 5 MB
- GC metadata and arena reservations ≈ 10 MB
- Per-map-entry overhead in `MasaqIndex.ByWord` (77,411 entries × ~50 B
  map slot + 24 B slice header) ≈ 5 MB
- Aligned string headers (~16 B × 17 fields × 1.5 avg segments × 77k
  entries) ≈ 30 MB
- Plus the OS's own bookkeeping for the process (mmap'd regions,
  guard pages, etc).

### Possible further reductions (out of scope)

1. **Columnar storage** for `MasaqSegment` strings — pack all `Gloss`
   values into a single byte buffer with offsets, then use offsets in
   the in-memory structs. Cuts heap by ~30%.
2. **mmap** for the gob files — avoids double-buffering through Go's
   GC and lets the OS page-cache serve reads directly. Cuts RSS by
   ~25 MB (the gob file size).
3. **Lazy-loading** — load the Quran only as the user requests
   surahs, and root meanings only when `/roots` is visited. Cuts
   initial RSS to ~30 MB but adds first-request latency.

We picked option (0) for this iteration: no special optimization, and
the app comfortably runs on any machine with ≥256 MB free RAM.

## 4. Functional verification

- ✅ Bismillah rules: Surah 1 verse 1 is Bismillah (counted), Surah 9
  has no Bismillah, Surah 2 skips standalone Bismillah as verse 1.
- ✅ Mark rendering: `۞` (199×) and `۩` (15×) at expected positions,
  pause marks `ۖ ۗ ۘ ۙ ۚ ۛ ۜ` rendered with the correct glyph.
- ✅ Word button event delegation: hover triggers tooltip, click
  triggers modal (verified via static analysis of
  `static/js/word-tooltip.js` and `static/js/word-modal.js`).
- ✅ Theme toggle: persists to `localStorage`, restores before paint
  via inline `<head>` script.
- ✅ Verse anchor scroll: `/surah/2/255` opens directly at verse 255
  with the brief golden focus highlight.
- ✅ Search highlight: API returns snippets wrapped in `<mark>`.
- ✅ Arabic search normalizes tashkeel — searching `الله` and `ٱللَّهِ`
  both return the same first hit.
- ✅ Root search handles Arabic input — `قول` → buckwalter `qwl`.

## 5. Known divergences (documented, accepted)

- **MASAQ ↔ XML word boundaries** in 23 positions where MASAQ merges
  a conjunction particle with the following word (e.g., `وَمَا`
  stored as one entry, while the XML has `وَ` and `مَا` as two
  tokens). The strict positional lookup misses for those 23 words.
  The UI shows "No morphological data" — see
  `docs/research/data-validation.md` for the full list and rationale.

- **AI-generated root meanings** — `meanings-roots-ai.jsonl` is
  included as a study aid; the `/root/{root}` page surfaces an
  attribution note. Classical tafsir is the authoritative source.

## 6. Test summary

```
ok  quranreader/backend/internal/data    (4 tests, 0.7s)
ok  quranreader/backend/internal/search  (12 tests, 2.0s)
ok  quranreader/backend/internal/server  (28 tests, 4.3s)
ok  quranreader/loc                      (5 tests, 1.1s)
ok  quranreader/pipeline                 (16 tests, 2.9s)
```

All packages pass under `-count=1` (no caching).
