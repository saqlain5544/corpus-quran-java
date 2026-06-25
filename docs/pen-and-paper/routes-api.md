# Routes & API — Quran Reader App

## Page Routes (HTML)

| Method | Path | Handler | Template | Notes |
|---|---|---|---|---|
| GET | `/` | `handleHomepage` | `homepage.tmpl` | List 114 surahs + global header. |
| GET | `/surah/{id}` | `handleSurah` | `surah.tmpl` | Full surah. id ∈ [1,114]. |
| GET | `/surah/{id}/{verse}` | `handleSurah` | `surah.tmpl` | Same, with `#verse-{n}` anchor auto-scrolled by client. |
| GET | `/search` | `handleSearch` | `search.tmpl` | `?q=...&type=root\|english\|arabic`. No results → empty state. |
| GET | `/roots` | `handleRootsList` | `roots-list.tmpl` | Paginated. `?page=N&sort=asc\|desc&surah=N`. |
| GET | `/root/{root}` | `handleRootDetail` | `root-detail.tmpl` | `root` is buckwalter or arabic; redirects canonical form. |
| GET | `/about` | `handleAbout` | `about.tmpl` | Static — data sources, attribution. |

## API Routes (JSON)

| Method | Path | Handler | Response | Caching |
|---|---|---|---|---|
| GET | `/api/word?s=N&a=N&w=N` | `handleAPIWord` | `{ word, segments: [...], root: {buckwalter, arabic, meaning_en, meaning_ar} \| null }` | `Cache-Control: public, max-age=3600` (data is static) |
| GET | `/api/search?q=...&type=...` | `handleAPISearch` | `{ query, type, results: [{kind, surah, ayah, word?, root?, text, snippet}] }` | `Cache-Control: no-store` |
| GET | `/api/roots?page=N&sort=...&surah=N` | `handleAPIRoots` | `{ roots: [{buckwalter, arabic, occurrences, meaning_en}], total, page, pageSize }` | 5 min |
| GET | `/api/root/{root}/occurrences` | `handleAPIRootOccurrences` | `{ root, arabic, occurrences: [{surah, ayah, word, snippet}] }` | 1 hour |
| GET | `/api/health` | `handleHealth` | `{ status: "ok", uptime_seconds, mem_alloc_mb }` | no-store |

## Static Routes

| Path | Source |
|---|---|
| `/static/css/{name}.css` | `static/css/*.css` (embedded) |
| `/static/js/{name}.js` | `static/js/*.js` (embedded) |
| `/static/fonts/hafs.woff2` | `data/fonts/hafs.woff2` (embedded) |

## Redirects

- `/index.html` → `/`
- `/surah` → `/` (or list page if we add one)
- `/surah/{id}/verse/{n}` → `/surah/{id}/{n}` (canonical form)

## URL Design Principles

1. **Surah IDs are integers** (1..114) — never slugs. Names are user-facing in the page but URL stays numeric.
2. **Ayah numbers are integers** within the surah path.
3. **Root parameter in URL** is the **buckwalter form** (e.g., `/root/qwl`) because it's ASCII, file-system-safe, and unambiguous. We render the Arabic form on the page.
4. **Search params** (`q`, `type`) are repeatable/encoded; we never use the path for search to keep it bookmarkable.
5. **Trailing slashes**: server-side normalize to non-slash form via 301.

## Handler Pseudocode

```go
// handleSurah — page handler
func handleSurah(w, r) {
    id, err := strconv.Atoi(r.PathValue("id"))
    if err != nil || id < 1 || id > 114 {
        respondError(w, 404, "Surah not found")
        return
    }
    s := quran.Surahs[id]
    if s == nil { respondError(w, 404); return }

    data := struct {
        Surah *Surah
        Prev  *Surah
        Next  *Surah
        VerseAnchor int   // optional
    }{Surah: s}
    if id > 1 { data.Prev = quran.Surahs[id-1] }
    if id < 114 { data.Next = quran.Surahs[id+1] }
    if verseStr := r.PathValue("verse"); verseStr != "" {
        if v, err := strconv.Atoi(verseStr); err == nil {
            data.VerseAnchor = v
        }
    }
    templates["surah.tmpl"].Execute(w, data)
}

// handleAPIWord — JSON API
func handleAPIWord(w, r) {
    s, _ := strconv.Atoi(r.URL.Query().Get("s"))
    a, _ := strconv.Atoi(r.URL.Query().Get("a"))
    wn, _ := strconv.Atoi(r.URL.Query().Get("w"))
    if !validLoc(s, a, wn) { respondJSONError(w, 404); return }

    key := LocKey(s, a, wn)
    segments := masaq.ByWord[key]
    var root *RootEntry
    if rb, ok := roots.ByLoc[key]; ok {
        root = roots.ByRoot[rb]
    }
    respondJSON(w, map[string]any{
        "word":     quran.Surahs[s].Ayahs[a].Words[wn-1].Value,
        "segments": segments,
        "root":     root,
    })
}

// handleAPISearch — fuzzy search
func handleAPISearch(w, r) {
    q := r.URL.Query().Get("q")
    typ := r.URL.Query().Get("type")
    if q == "" { respondJSON(w, map[string]any{"results": []}); return }
    var results []SearchResult
    switch typ {
    case "root":
        results = search.Root(q, 50)
    case "english":
        results = search.English(q, 50)
    case "arabic":
        results = search.Arabic(q, 50)
    default:
        respondJSONError(w, 400, "type must be root|english|arabic"); return
    }
    respondJSON(w, map[string]any{"query": q, "type": typ, "results": results})
}
```

## Middleware Order

```
incoming → RecoveryMiddleware → LoggingMiddleware → MemoryGauge → handler → response
```

- **Recovery**: catches panics, returns 500.
- **Logging**: structured one-line per request via `log/slog`.
- **MemoryGauge**: records `runtime.MemStats` every 5 min, exposed via `/api/health`.

