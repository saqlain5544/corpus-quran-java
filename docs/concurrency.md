# Concurrency Model

The server is single-process Go. Every HTTP request runs in its own
goroutine (Go's `net/http` does this automatically). Beyond that, the
code uses `sync` primitives to parallelize CPU-bound and I/O-bound
work where it matters. This document explains where, why, and how.

## Rules of thumb

1. **No shared mutable state across goroutines** — the existing
   server code is read-heavy: search indexes (`roots.ByLoc`,
   `masaq.ByWord`, `concordance.ByRoot`) and the Quran struct are
   loaded once at startup and read-only for the lifetime of the
   process. Every parallel section below respects that.
2. **Bounded parallelism** — we use `runtime.GOMAXPROCS` (or a
   bounded semaphore) so the goroutine count stays proportional
   to cores, not to data size. Cold-cache startup is dominated
   by I/O, not goroutine scheduling.
3. **`go test -race ./...` is green** — all parallel sections have
   been exercised by the race detector. If you add a new one, run
   the suite with `-race` before merging.

## Where parallelism is used

### Startup (`backend/main.go`)

Three independent I/O paths run concurrently:

| Path | What it does | Goroutines |
|---|---|---|
| `data.LoadAll` | Three SQL queries (Quran / MASAQ / roots) against the same `*sql.DB`. | 3 |
| `data.LoadTranslations` | Two XML files (en, ur). | 2 |
| `data.LoadConcordance` | One JSONL file. | 1 |

Total wall-clock startup ≈ `max(LoadAll, LoadTranslations, LoadConcordance)`
rather than the sum.

```go
loadWg.Add(2)
go func() { defer loadWg.Done(); /* LoadTranslations */ }()
go func() { defer loadWg.Done(); /* LoadConcordance  */ }()
loadWg.Wait()
```

Each sub-loader writes to its own result struct, so the only shared
state is the `sync.WaitGroup` and the result variables.

### `data.LoadAll` (`backend/internal/data/load.go`)

`loadQuranFromDB` / `loadMasaqFromDB` / `loadRootsFromDB` run in
parallel goroutines that share one `*sql.DB`. SQLite's `?mode=ro`
connection pool handles the concurrent readers cleanly. Results
travel through `chan qResult` / `chan mResult` / `chan rResult`
(buffered size 1, so the goroutine exits immediately after writing
even if the receiver hasn't read yet).

### `data.LoadTranslations` (`backend/internal/data/translations.go`)

English and Urdu XML files decode in parallel goroutines. Each file
is independent — no shared state. The two results are merged into
the `Translations` struct after `WaitGroup.Wait()`.

### `server.handleSurah` (`backend/internal/server/server.go`)

Per-ayah `WordRoots` maps are built in parallel chunks. Each chunk
holds a contiguous range of ayah indices, reads from `roots.ByLoc`
(read-only), and writes into a pre-sized slice at unique indices.
No locks needed.

```go
const chunkSize = 16
sem := make(chan struct{}, runtime.GOMAXPROCS(0))
for start := 1; start <= totalAyahs; start += chunkSize {
    end := min(start + chunkSize, totalAyahs)
    wg.Add(1)
    sem <- struct{}{}
    go func(s, e int) {
        defer wg.Done()
        defer func() { <-sem }()
        for an := s; an <= e; an++ {
            ayahEntries[an-1] = buildEntry(an)
        }
    }(start, end)
}
wg.Wait()
```

The semaphore bounds concurrent goroutines to `GOMAXPROCS`. The
indexed slice assignment means there's no shared mutable state — the
race detector stays clean.

### `server.handleRootDetail` (`backend/internal/server/server.go`)

Concordance lemmas are built in parallel chunks of 4 lemmas. Each
lemma is independent (reads `roots.ByRoot` and `translations.Sets`
once captured, writes its own slot in the output slice). The active
translation set is captured once before the goroutines start:

```go
var trans *data.TranslationSet
if s.translations != nil && s.translations.Default < len(s.translations.Sets) {
    trans = &s.translations.Sets[s.translations.Default]
}
// ... snapshot lemmas into ins, parallel-build into outs
```

## Where parallelism is NOT used (and why)

| Path | Reason it's sequential |
|---|---|
| Search algorithms (`search.Arabic`, `search.English`, …) | Per-key work is ~tens of ns (hashmap lookups + substring scan). Goroutine setup (µs) would dominate. |
| Token render per ayah | Each ayah is a few hundred bytes; rendering a whole surah takes <10ms total. Parallel would add overhead. |
| API handlers | They fan out to one fetch each; no CPU-heavy work to parallelize. |
| Template rendering | `html/template` is already efficient and the render path is a single sequential pass. |

## What concurrency DOES NOT cover

- **Per-request goroutine pool** — the stdlib's `net/http` already
  uses one goroutine per request. We don't add a worker pool.
- **Cross-request caching** — every request is independent; we
  don't memoize between requests.
- **Streaming responses** — pages are rendered in full before being
  written to the wire. Acceptable for pages that fit in <100KB.

## Adding new parallel code

If you add a new parallel section, follow these rules:

1. **Capture shared state into locals BEFORE spawning goroutines.**
   Reading from a struct field across goroutines without locking
   works only if the field is never written after startup. The
   server's data fields are safe; the `translations` and
   `concordance` slices are not re-assigned, but pointer aliasing
   in goroutines is still a hazard — prefer local variables.
2. **Use indexed slice assignment** so each goroutine writes to a
   unique slot. Avoid `append` from goroutines.
3. **Bound the goroutine count.** Use `runtime.GOMAXPROCS(0)` or
   a `chan struct{}` semaphore. `for i := 0; i < N; i++ { go … }`
   without a semaphore spawns up to N goroutines.
4. **Run `go test -race ./...`** before merging. The race
   detector catches every misuse listed above.