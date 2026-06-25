# Tasks — Quran Reader App (Go + Vanilla)

Phase-ordered. Each task is atomic, testable, and has a clear "done" condition.

---

## Phase 1 — Research ✅ COMPLETE

- [x] Validate trusted data files (quran-uthmani.xml, MASAQ.csv, hafs.woff2)
- [x] Document XML schema, mark conventions, word tokenization
- [x] Document MASAQ schema and alignment with XML
- [x] Document root data files (corpus-roots.json, meanings-roots-ai.jsonl, root_meaning.csv)
- [x] Document Arabic fuzzy matching approach
- [x] Output: `docs/research/data-validation.md`

---

## Phase 2 — Planning ✅ COMPLETE

- [x] `docs/pen-and-paper/architecture.md`
- [x] `docs/pen-and-paper/routes-api.md`
- [x] `docs/pen-and-paper/component-tree.md`
- [x] `docs/pen-and-paper/ui-design-system.md`
- [x] `docs/pen-and-paper/data-flow-diagrams.md`
- [x] `docs/to-dos/tasks.md` (this file)

---

## Phase 3 — Data Pipeline

- [ ] Create `pipeline/build.go` (CLI tool)
- [ ] Implement XML parser (stream-based; Surah/Ayah/Word with mark-tokenization)
- [ ] Implement CSV parser for MASAQ (handle CRLF, comma-in-quoted-gloss if any)
- [ ] Implement root joiner (corpus-roots.json + meanings-roots-ai.jsonl → RootsIndex)
- [ ] Implement gob emitters for quran, masaq_index, roots
- [ ] Implement meta.json emitter
- [ ] Write verification report: 114 surahs, 6236 ayas, 77797 words, root coverage
- [ ] Verify every MASAQ word ID is present in `MasaqIndex.ByWord`
- [ ] Verify every XML word token has a MASAQ entry (or flag the diff)
- [ ] Document regeneration in `data/processed/README.md`

**Done when**: `go run ./pipeline/build.go` produces all 4 files with zero verification failures, build-report.txt matches expected counts.

---

## Phase 4 — Tests (TDD)

- [ ] `data/quran_test.go` — parser, tokenizer, mark handling, Bismillah rules
- [ ] `data/masaq_test.go` — parser, index build, LocKey encoding, segcount per word
- [ ] `data/roots_test.go` — join by location, ByArabic reverse lookup
- [ ] `search/arabic_test.go` — normalizer edge cases, exact + partial match
- [ ] `search/english_test.go` — Levenshtein, substring, short-query skip
- [ ] `search/root_test.go` — buckwalter + arabic input, prefix match
- [ ] `server/handlers_pages_test.go` — 404s, valid pages, headers (using httptest)
- [ ] `server/handlers_api_test.go` — JSON shape, error cases
- [ ] `server/render_test.go` — template execution against fixture data
- [ ] `pipeline/pipeline_test.go` — round-trip raw → processed → load
- [ ] JS tests: tooltip positioning (viewport flip), modal open/close, scroll-to-anchor

**Done when**: `go test ./...` passes with -race and -count=1. JS test suite passes.

---

## Phase 5 — Backend (Go)

- [ ] `go.mod` (stdlib only)
- [ ] `main.go` — flag parsing (`-addr`, `-dev`), server bootstrap, embed.FS
- [ ] `internal/data/embed.go` — embed processed/*.gob + static/ + templates/
- [ ] `internal/data/quran.go` — Quran/Surah/Ayah/Word types + load
- [ ] `internal/data/masaq.go` — MasaqIndex type + load
- [ ] `internal/data/roots.go` — RootsIndex type + load
- [ ] `internal/data/lockey.go` — `LocKey(s,a,w) uint64`
- [ ] `internal/search/normalize.go` — Arabic normalizer
- [ ] `internal/search/arabic.go` — Arabic word search
- [ ] `internal/search/english.go` — English gloss search (Levenshtein)
- [ ] `internal/search/root.go` — root search
- [ ] `internal/server/router.go` — http.ServeMux wiring
- [ ] `internal/server/handlers_pages.go` — handleHomepage, handleSurah, handleSearch, handleRootsList, handleRootDetail, handleAbout
- [ ] `internal/server/handlers_api.go` — handleAPIWord, handleAPISearch, handleAPIRoots, handleAPIRootOccurrences, handleHealth
- [ ] `internal/server/handlers_static.go` — CSS, JS, fonts
- [ ] `internal/server/middleware.go` — Recovery, Logging, MemoryGauge
- [ ] `internal/server/render.go` — template helpers + error page
- [ ] Templates: `layout.tmpl`, `homepage.tmpl`, `surah.tmpl`, `search.tmpl`, `roots-list.tmpl`, `root-detail.tmpl`, `about.tmpl`, `error.tmpl`, partials: `global-header.tmpl`, `word-tooltip.tmpl`, `word-modal.tmpl`
- [ ] Run `go test ./...` until green
- [ ] Run `go vet ./...` clean
- [ ] Build binary: `go build -ldflags='-s -w' -o ./bin/quranreader .`

**Done when**: server starts on `:8080`, all routes return correct responses, all tests pass, binary builds.

---

## Phase 6 — Frontend (Vanilla)

- [ ] `static/css/tokens.css` — design tokens (light + dark themes)
- [ ] `static/css/layout.css` — grid, page shell, responsive
- [ ] `static/css/quran.css` — ayah, word, mark, hafs rendering
- [ ] `static/css/components.css` — header, tooltip, modal, search, roots
- [ ] `static/js/app.js` — router glue, theme persistence, global search submit
- [ ] `static/js/global-header.js` — search submit, theme toggle
- [ ] `static/js/surah-header.js` — local search, font-size, line-height, prev/next
- [ ] `static/js/word-tooltip.js` — event delegation, fetch, edge-aware positioning
- [ ] `static/js/word-modal.js` — fetch, render, focus trap, escape, backdrop close
- [ ] `static/js/verse-scroll.js` — load-time scroll to `#verse-N`, brief highlight
- [ ] Verify keyboard accessibility on every interactive element
- [ ] Verify focus trap in modal manually + via JS test

**Done when**: visual + interaction tests pass, no console errors on page load, no external fonts except hafs.woff2.

---

## Phase 7 — Integration & Debugging

- [ ] Run full server, exercise all routes with `curl`
- [ ] Browser smoke test: load homepage → navigate to surah → hover word → click word → see modal → click root link → load root page → use search → toggle theme
- [ ] Memory profile: `runtime.ReadMemStats` under steady state, confirm < 80 MB RSS
- [ ] Verify mobile responsive: 375×667 viewport renders correctly
- [ ] Verify dark theme + light theme both look editorial (no garish colors)
- [ ] Verify font loading: hafs.woff2 loads, no FOIT
- [ ] Verify accessibility: keyboard nav through all interactive elements
- [ ] Confirm all attributions on `/about` page
- [ ] Write final `README.md` with run/build instructions
- [ ] Tag a final commit

**Done when**: all integration tests pass, memory < 80 MB RSS, README is complete.

---

## Cross-cutting Constraints (Always)

- **Memory**: < 80 MB RSS at steady state.
- **No external dependencies**: only Go stdlib. No npm, no frameworks.
- **Editorial design**: no garish colors; typography-first.
- **Trusted data only**: only quran-uthmani.xml, MASAQ.csv, hafs.woff2 used directly. AI-generated meanings clearly labeled.
- **Atomic, modular, DRY**: small functions, clear single responsibility.
- **Tests first**: every function has at least one happy-path test + one edge case.
- **No secrets in repo**.
- **No commits without explicit user ask**.

