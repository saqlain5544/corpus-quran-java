# Architecture — Quran Reader App (Go + Vanilla JS/CSS/HTML)

## 1. High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│  Browser (no framework, only fetch + DOM APIs)                  │
│  ┌──────────────┐ ┌─────────────┐ ┌───────────────────────┐     │
│  │ GlobalHeader │ │ SurahHeader │ │ SurahView/Roots/Search │     │
│  └──────┬───────┘ └──────┬──────┘ └───────────┬───────────┘     │
│         │                │                     │                 │
│         ▼                ▼                     ▼                 │
│   ┌──────────────────────────────────────────────────────┐      │
│   │  app.js — router, state, event delegation            │      │
│   │  word-tooltip.js word-modal.js verse-scroll.js       │      │
│   └──────────────────────────────────────────────────────┘      │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP (text/html, application/json)
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Go server (stdlib only — net/http, html/template, encoding/xml)│
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ main.go        — flag parsing, server bootstrap, embed.FS  │ │
│  │ router.go      — http.ServeMux wiring                      │ │
│  │ handlers.go    — page handlers (homepage, surah, roots…)   │ │
│  │ api.go         — JSON handlers (/api/word, /api/root/…)    │ │
│  │ data/          — loaded structs (Quran, Masaq, Roots)      │ │
│  │ search/        — fuzzy matchers (root/eng/arabic)          │ │
│  │ templates/     — html/template files (embedded)            │ │
│  │ static/        — CSS, JS, hafs.woff2 (embedded)            │ │
│  └────────────────────────────────────────────────────────────┘ │
└──────────────────────────┬──────────────────────────────────────┘
                           │ init-time read of
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│  Processed data (data/processed/)                                │
│  - quran.gob          (~1.2 MB) — Surah/Ayah/Word tree          │
│  - masaq_index.gob    (~3 MB)   — wordLocKey → []Segment        │
│  - loc_to_root.gob    (~0.5 MB) — wordLocKey → root buckwalter  │
│  - roots.gob          (~1.5 MB) — root → []Loc + arabic + gloss │
│  - meta.json          (~5 KB)   — surah names, counts           │
└─────────────────────────────────────────────────────────────────┘
```

## 2. Package Layout

```
backend/
├── go.mod                          (module: quranreader; stdlib only)
├── main.go                         (server bootstrap, flag, embed.FS)
├── internal/
│   ├── data/
│   │   ├── quran.go                (Surah, Ayah, Word structs + load)
│   │   ├── masaq.go                (MasaqSegment + index loader)
│   │   ├── roots.go                (RootEntry + index loader)
│   │   └── embed.go                (//go:embed of processed/*.gob)
│   ├── search/
│   │   ├── arabic.go               (normalizer + arabic-word search)
│   │   ├── english.go              (Levenshtein + gloss search)
│   │   ├── root.go                 (root search by arabic/buckwalter)
│   │   └── normalize.go            (shared Arabic normalizer)
│   ├── server/
│   │   ├── router.go               (http.ServeMux wiring + helpers)
│   │   ├── handlers_pages.go       (page handlers)
│   │   ├── handlers_api.go         (JSON API handlers)
│   │   ├── handlers_static.go      (static + font)
│   │   ├── middleware.go           (logging, recovery, memory gauge)
│   │   └── render.go               (template.Render + helpers)
│   └── pipeline/
│       └── pipeline_test.go        (round-trip: raw → processed → load)
├── pipeline/
│   ├── build.go                    (CLI tool: reads raw → emits processed/)
│   └── README.md                   (regeneration steps)
├── templates/                      (*.tmpl embedded into binary)
└── static/                         (CSS + JS + hafs.woff2 embedded)
```

## 3. Data Structures (in-memory)

```go
// data/quran.go
type Quran struct {
    Surahs  map[int]*Surah            // 1..114
    Meta    Meta                       // name lists, counts
}

type Surah struct {
    Number     int
    Name       string                   // Arabic
    NameLatin  string                   // "Al-Fatihah"
    Bismillah  string                   // raw text (may be empty for surah 9)
    Ayahs      map[int]*Ayah
}

type Ayah struct {
    Number  int
    Text    string                      // raw text with marks inline
    Words   []WordToken                 // parsed tokens (word or mark)
}

type WordToken struct {
    Kind   string                       // "word" | "mark"
    Value  string                       // text
    Marks  []string                     // trailing marks (only when Kind=="word")
    WordNo int                          // only meaningful when Kind=="word"
}

// data/masaq.go
type MasaqSegment struct {
    ID                  int
    SuraNo              int
    VerseNo             int
    WordNo              int
    SegmentNo           int
    Word                string
    WithoutDiacritics   string
    SegmentedWord       string
    MorphTag            string
    MorphType           string          // Prefix/Stem/Suffix/Other_i3rab
    PunctuationMark     string
    InvariableDeclinable string
    SyntacticRole       string
    PossessiveConstruct string
    CaseMood            string
    CaseMoodMarker      string
    Phrase              string
    PhrasalFunction     string
    Gloss               string
}

type MasaqIndex struct {
    ByWord map[uint64][]MasaqSegment  // locKey(s,v,w) → segments
}

// data/roots.go
type RootEntry struct {
    Buckwalter    string                // "qwl"
    Arabic        string                // "قَوْل"
    Letters       string                // "ق و ل"
    POS           string                // "V" | "N" | "P"
    Occurrences   int
    MeaningEN     string
    MeaningAR     string
    Locations     []string              // ["1:1:1", "1:2:6", ...]
}

type RootsIndex struct {
    ByRoot      map[string]*RootEntry  // buckwalter key → entry
    ByLoc       map[uint64]string      // locKey → buckwalter
    ByArabic    map[string]string      // arabic letters → buckwalter (for reverse lookup)
}
```

## 4. Word-Location Key Encoding

To make hashmap lookups O(1) without allocating strings:

```go
// Max values: surah 1..114, ayah 1..286, word 1..~130
// Encoding: surah * 1_000_000 + ayah * 1_000 + word
// Range: 114 * 1e6 + 286 * 1e3 + 130 = 114,286,130 → fits comfortably in uint32.
// We use uint64 for headroom (future-proofing).
func LocKey(surah, ayah, word int) uint64 {
    return uint64(surah)*1_000_000 + uint64(ayah)*1_000 + uint64(word)
}
```

This is the single key used everywhere: MasaqIndex.ByWord, RootsIndex.ByLoc.

## 5. Memory Budget

| Structure | Estimated size |
|---|---|
| `Quran.Surahs` (114 surahs, names, ayas, raw text, words) | ~1.5 MB |
| `MasaqIndex.ByWord` (77,797 entries × ~250B average) | ~19 MB |
| `RootsIndex.ByRoot` (1,642 entries × ~2 KB with full meanings) | ~3.3 MB |
| `RootsIndex.ByLoc` (49,967 entries × 16B) | ~0.8 MB |
| `RootsIndex.ByArabic` (~1,642 entries × 32B) | ~52 KB |
| Static assets (CSS + JS + hafs.woff2) | ~80 KB |
| HTTP server overhead + Go runtime | ~6 MB |
| **Total** | **~30 MB** |

Comfortably under the 80 MB RSS target. GOB encoding will compress further; we'll measure at integration time.

## 6. Streaming vs In-Memory Tradeoff

The full Quran text is 1.5 MB; loading it all at once is fine. We use Go's `encoding/xml` with a **decoder-based parser** (not a tree-based DOM) for the build pipeline to keep peak memory low during preprocessing, then materialize the in-memory tree only in the final structures.

The Masaq CSV (18 MB) is read once at startup into the hashmap. CSV parser is line-by-line; no full file in memory at once.

## 7. Concurrency Model

- All data structures are **immutable after init** (no mutex needed for read paths).
- HTTP handlers are **stateless** and read-only.
- `http.ServeMux` is goroutine-safe; we register handlers at startup.
- Search handlers iterate over slices/maps without locks — safe due to immutability.

## 8. Deployment

- Single static binary: `quranreader -addr :8080 -data ./data/processed`.
- Or: zero-config `quranreader` reads `./data/processed` relative to working dir.
- Or: data is **embedded** at compile time via `//go:embed` so the binary is fully self-contained.

**Decision**: We embed via `go:embed` for distribution simplicity. The pipeline produces a directory `data/processed/*.gob` and the Go binary embeds it at compile time. A `-dev` flag allows loading from filesystem instead of the embedded blob, for iteration during development.

## 9. Error Handling & Status Pages

- 404 → minimal editorial-styled 404 page.
- 500 → minimal 500 page; no internal details leaked.
- All handlers return a `respondError(w, status, msg)` helper.

## 10. Build Pipeline (high level)

`pipeline/build.go` is a separate `go build`-able CLI that:

1. Parses `quran-uthmani.xml` → emits `quran.gob` (Quran struct).
2. Parses `MASAQ.csv` → emits `masaq_index.gob` (MasaqIndex struct).
3. Parses `corpus-roots.json` + `meanings-roots-ai.jsonl` → emits `roots.gob` (RootsIndex struct).
4. Computes and emits `meta.json` (surah name lists, counts).
5. Writes verification report to `docs/data-verification/build-report.txt`.

The build is reproducible; rerunning produces byte-identical gob files for non-floating-point data (verified in test).

See `docs/pen-and-paper/pipeline-flow.md` for the full flow diagram.

