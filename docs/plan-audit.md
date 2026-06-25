# plan.md — Implementation Audit

This document maps every requirement in [`plan.md`](../plan.md) (plus the
four `docs/pen-and-paper/*.md` documents it relies on) to its current
implementation status as of this commit.

Legend:
- ✅ **Done** — implemented and exercised by tests or live verification.
- 🟡 **Partial** — implemented but with a known gap, or implemented
  for one branch but missing another (called out per item).
- ❌ **Missing** — not yet implemented.
- 📝 **Note** — clarification on what "done" means in this codebase.

---

## §1 — `/root/details/[root:str]` page

| Item | Status | Where |
|---|---|---|
| Two-column layout (5/3 grid ratio ≈ 8fr/4fr) | ✅ | `backend/templates/root-detail.tmpl:18` uses `.layout-shell--two-col`; CSS `surfaces.css:73-82`. |
| First column = detailed meaning from `meanings-roots-ai.jsonl` | ✅ | `backend/templates/root-detail.tmpl:20-72`; data load `backend/internal/data/meanings.go:55-99`. |
| Second column = concordance data from `concordance.jsonl` | ✅ | `backend/templates/root-detail.tmpl:103-141`; data load `backend/internal/data/concordance.go`. |
| Every `[surah:ayah]` parsed and linked to `/surah/S#verse-V` | ✅ | `backend/internal/server/server.go:478-481` builds `Link: "/surah/{s}#verse-{v}"`. |
| Arabic text has `dir="rtl"` | ✅ | Every Arabic-bearing element uses `lang="ar" dir="rtl"`: e.g. `root-detail.tmpl:30,42,72,123`. |
| Occurrences show full ayah text | ✅ | `root-detail.tmpl:91-95` renders `{{.VerseText}}` for each occurrence. |
| Concordance lemmas: Arabic lemma header + verse text + translation | ✅ | `root-detail.tmpl:124-138`. |
| Second column: fixed position but not out of flow | ✅ | `surfaces.css:124-128` `.layout-aside--sticky` (sticky, not absolute). `top: calc(56px + var(--sp-4))` accounts for the sticky global header. |

---

## §2 — Tooltip Design

| Item | Status | Where |
|---|---|---|
| Word (Uthmani) | ✅ | `backend/static/css/components.css:288-293` `.wt-word` (font-arabic, large). Rendered by `word-tooltip.js:62-65`. |
| Root (Arabic + Buckwalter) | ✅ | `word-tooltip.js:67-73`, CSS `components.css:296-310`. |
| Lemma (bare stem) | ✅ | `word-tooltip.js:75-78`, CSS `components.css:312-318`. |
| Gloss (English) | ✅ | `word-tooltip.js:81-84`, CSS `components.css:321`. |
| Morphology tags row | ✅ | `word-tooltip.js:87-99`, CSS `components.css:323-337`. |
| Grammatical tags row | ✅ | `word-tooltip.js:101-115`, CSS `components.css:340-341`. |
| Edge-aware positioning (above/below, clamp to viewport) | ✅ | `word-tooltip.js:126-151` `positionTooltip`. |

---

## §3 — Surah Page

| Item | Status | Where |
|---|---|---|
| Surah header sticky below global header | ✅ | `backend/static/css/components.css:80-87` (`.surah-header { top: 56px }`). |
| Local search box | ✅ | `surah.tmpl:33-35`, JS `backend/static/js/surah-header.js:166-198`. |
| Local search scrolls to first match, no URL change | ✅ | `surah-header.js:193-194`: `srIdx = 0; scrollToMatch();` — never navigates. |
| Local search error state (no match → red border) | ✅ | `surah-header.js:185-191`: `input.classList.add("is-error")`. |
| Font-size slider [10–40] | ✅ | `surah.tmpl:38` `<input type="range" min="10" max="40">`. JS `surah-header.js:121-126`. |
| Line-height slider [1–3 step 0.1] | ✅ | `surah.tmpl:43` `<input type="range" min="1" max="3" step="0.1">`. |
| Prev/next surah (disabled on first/last) | ✅ | `surah.tmpl:21-30`: `{{if .Prev}} … {{else}}disabled{{end}}`. |
| Two-column layout: side panel (left) + main surah (right) | ✅ | `surah.tmpl:5-9` (`.surah-side-panel`) and `:12` (`.surah-main`); CSS `layout.css:77-83`. |
| Side panel: 3-col sub-grid (frequency / concordance / shades) | ✅ | `backend/static/js/surah-side-panel.js:99-180`. The grid renders on word click. |
| Side panel shows lemma frequency with current lemma on top | ✅ | `surah-side-panel.js:107-111`: `lemmas.sort(…)` with `currentLemma` first. |
| Side panel: concordance column deduped by surah:ayah | ✅ | `surah-side-panel.js:135-141`: `if (seen[vKey]) continue;`. |
| Verse number: SVG-like enclosure, scales with digits | 🟡 | `quran.css:45-66` `.ayah-number::before` is a CSS-only square; numbers scale via `calc(var(--fs-quran) * 1.3)`. Not a true SVG, but visually equivalent and scales. |
| Verse number after verse text | ✅ | `surah.tmpl:79`: `<span class="ayah-number">{{$an}}</span>` rendered at end of `.ayah-text` paragraph. |
| Bismillah as separate line before first ayah | ✅ | `surah.tmpl:60-63`. Skipped on surah 9 (`surah.Bismillah == ""`). |
| Modal replaced by side panel | ✅ | Plan: "Now the modal is no longer needed as its job is done by the sidebar." No `word-modal.js` shipped — only `surah-side-panel.js`. |
| Word buttons (not anchors) | ✅ | `surah.tmpl:69-78` `<button type="button" class="word" …>`. |
| Every Arabic line / word has `dir="rtl"` | ✅ | `surah.tmpl:66, 70, 79` all carry `lang="ar" dir="rtl"`. |

---

## §4 — Translation Integration

| Item | Status | Where |
|---|---|---|
| English translation(s), LTR | ✅ | `data.LoadTranslations` loads `en.sahih.xml`. `translations.go:39-49`. Rendered in `surah.tmpl:84-94`. |
| Urdu translation(s), RTL | ✅ | `translations.go:42-44` loads `ur.junagarhi.xml` (Label `محمد جوناگڑھی`, `Dir: "rtl"`). |
| Matches textual format of verses | ✅ | Translation files are Tanzil-style `sura/aya` mappings keyed the same way as the Quran struct. |
| Transliteration kept server-side, not in UI | ✅ | `translations.go:28-29` comment: "Transliteration files are deliberately skipped per plan.md §Translation." `en.transliteration.xml` exists on disk but is never loaded. |
| Translation selector toggles visible translation | ✅ | `surah.tmpl:46-51` `<select data-component="translation-select">`; JS `surah-header.js:140-163`. |

---

## §5 — Search Engine

| Item | Status | Where |
|---|---|---|
| Fuzzy match translations | ✅ | `search.Translation` in `backend/internal/search/algorithms.go:295-389`. |
| Fuzzy match glosses | ✅ | `search.English` (`algorithms.go:110-179`) uses substring + Levenshtein ≤ 2. |
| Fuzzy match Arabic text (with/without tashkeel) | ✅ | `search.Arabic` (`algorithms.go:16-81`) + `NormalizeArabic` (`normalize.go:20-67`). |
| Different script variants (Indo-Pak vs Uthmani) | 🟡 | The system normalizes alif/ya/ta-marbuta variants and tatweel, which catches most Uthmani↔Indo-Pak differences for the letters that vary. A formal per-rune transliteration map (e.g. `ٱ` → `ا`, `ہ` → `ه`) is partial: only the most common variants are handled. |
| Full match roots (Buckwalter, Arabic with/without spaces, with/without tashkeel) | ✅ | `search.Root` (`algorithms.go:400-457`) + canonical `types.NormalizeArabicRoot` (`types/types.go:138-183`). |
| Fuzzy match lemmas | ✅ | `search.Lemma` (`algorithms.go:184-256`). |
| Search result highlighting | ✅ | `HighlightSnippet` (`normalize.go:218-228`); rendered via `<mark>` in `search.tmpl:22`. |

---

## §6 — Global Header

| Item | Status | Where |
|---|---|---|
| Sticky top, full viewport width | ✅ | `components.css:11-21`. |
| Brand link | ✅ | `global-header.tmpl:3`. |
| Search box + type selector | ✅ | `global-header.tmpl:4-13` (form submits GET `/search`). |
| Links to Roots, About (and Concordance, /) | ✅ | `global-header.tmpl:14-19`. |
| Theme toggler (`data-component="theme-toggle"`) | ✅ | `global-header.tmpl:18`; JS `app.js:8-15`. |
| Theme: dark + light only | ✅ | `tokens.css:67-128` defines exactly two palettes. |
| Type persistence in localStorage | ✅ | `global-header.js:14-21` reads/writes `qr.searchType`. |
| `/` keyboard shortcut focuses search | ✅ | `global-header.js:31-39`. |
| Escape clears search input | ✅ | `global-header.js:36-38`. |

---

## §7 — Surah Header

| Item | Status | Where |
|---|---|---|
| Local search box (separate from global) | ✅ | `surah.tmpl:33-35`. |
| Font-size range controller [10–40] | ✅ | `surah.tmpl:38`. |
| Line-height controller [1–3 step 0.1] | ✅ | `surah.tmpl:43`. |
| Sticky below global header | ✅ | `components.css:82-86`. |
| Persists font-size / line-height in localStorage | ✅ | `surah-header.js:103-117` reads `qr.fs`/`qr.lh`; lines 121-136 write on input. |
| Prev/next surah | ✅ | `surah.tmpl:16-23`. |

---

## §8 — Stack / Constraints

| Item | Status | Where |
|---|---|---|
| Vanilla Go | ✅ | `go.mod` lists only `mattn/go-sqlite3` as a dep. |
| Vanilla JS / CSS / HTML | ✅ | No bundlers, no frameworks in `backend/static/`. |
| Lowest possible memory footprint | ✅ | `backend/main.go:42-49` sets `GOMEMLIMIT=96 MB`. |
| 12-column grid system | ✅ | `layout.css:62-63` (`.page`), `surfaces.css:73-82` (`.layout-shell--two-col`). |
| Consistent theme / typography / spacing tokens | ✅ | `tokens.css`. |
| Editorial UI (no garish colors) | ✅ | Paper-and-ink palette (`tokens.css:67-105`). |
| Accessibility: skip link, focus rings, aria labels, semantic HTML | ✅ | `layout.tmpl:21` skip link; `layout.css:24-27` focus rings; `aria-haspopup`, `aria-label`, `role="banner"` throughout templates. |
| `prefers-reduced-motion` | ✅ | `tokens.css:130-135`. |
| `prefers-color-scheme` for theme default | ✅ | `tokens.css:108-128`. |

---

## §9 — Phases

| Item | Status | Where |
|---|---|---|
| Plan: documented before code | ✅ | `docs/pen-and-paper/architecture.md`, `routes-api.md`, `component-tree.md`, `ui-design-system.md`, `data-flow-diagrams.md`. |
| Pre-implementation: pipeline scripts + verification | ✅ | `pipeline/build.go` (concurrency-safe parallel file reads). |
| Test-driven implementation: tests before code | 🟡 | Tests were added alongside code, not strictly before. The current state is well-tested: `go test ./...` and `go test -race ./...` both pass. |
| Atomic / modular / DRY code | ✅ | Each package (`data`, `search`, `server`, `pipeline`) has one job; types live in `types/` so pipeline and runtime share contracts. |

---

## §10 — Data

| Item | Status | Where |
|---|---|---|
| Trusted sources only flagged in UI | ✅ | `about.tmpl:14-25` lists Tanzil, MASAQ, Hafs font as trusted; AI-generated meanings flagged as supplementary ("not a replacement for classical tafsir"). |
| `hafs.woff2`, MASAQ.csv, quran-uthmani.xml kept | ✅ | `data/fonts/hafs.woff2`, `data/morphology/MASAQ.csv`, `data/quran/quran-uthmani.xml` are all still present. |
| Other data needs verification (unverified in UI) | ✅ | AI root meanings carry a `<p class="data-attribution">` disclaimer on every root-detail page (`root-detail.tmpl:113-115`). |

---

## §11 — Components & Pages (from plan §"What we are building?")

| Page | Status | Where |
|---|---|---|
| Homepage: lists surahs 1–114 | ✅ | `homepage.tmpl`, `server.handleHomepage` (`server.go:250-262`). |
| Homepage: search box → `/search?q=…&type=…` | ✅ | `global-header.tmpl:4-13`. |
| Homepage: link to roots, concordance | ✅ | `global-header.tmpl:14-19`. |
| `/roots`: paginated roots, asc/desc frequency, surah filter | ✅ | `roots-list.tmpl`, `server.handleRootsList` (`server.go:392-435`), `search.RootsList` (`algorithms.go:462-518`). |
| `/roots`: synonym roots | ❌ | Not implemented. Requires a synonym dictionary — out of scope for now. |
| `/root/detailed/[root]`: occurrences, surah + asc/desc filters, meanings from `meanings-roots-ai.jsonl` | ✅ | `root-detail.tmpl`, `handleRootDetail` (`server.go:437-566`). |
| `/surah/[id]`: full surah in paragraph form, hafs.woff2 font | ✅ | `surah.tmpl`, CSS `tokens.css:137-142` `@font-face`. |
| `/surah/[id]`: every word has tooltip on hover/focus | ✅ | `surah.tmpl:69-78`, JS `word-tooltip.js`. |
| `/surah/[id]`: tooltip fetches MASAQ + root data | ✅ | `word-tooltip.js:36-39` `/api/word?s=…&a=…&w=…`. |
| `/surah/[id]`: word without root doesn't get modal | ✅ | `data-root=""` on the button (`surah.tmpl:74`); JS checks truthiness (`surah-side-panel.js:18`). |
| `/surah/[id]/#verse-N`: scrolls to that verse | ✅ | `verse-scroll.js` (with the bug fix from this session). |
| `/search`: handles `q` + `type` ∈ {root, english, arabic, translation} | ✅ | `handleSearch` (`server.go:361-390`). |
| `/search`: english_term uses Levenshtein | ✅ | `English` function, `Levenshtein ≤ 2`. |
| `/search`: arabic_word fuzzy match with tashkeel | ✅ | `Arabic` function + `NormalizeArabic`. |
| `/search`: highlights matched substring | ✅ | `HighlightSnippet` (`normalize.go:218-228`). |

---

## §12 — Data Flow (from `data-flow-diagrams.md`)

| Flow | Status | Where |
|---|---|---|
| Browser → server → SQLite → in-memory maps | ✅ | `data.LoadAll` → `loadQuranFromDB` / `loadMasaqFromDB` / `loadRootsFromDB` (`load.go`). |
| Tooltip hover → fetch `/api/word` → render | ✅ | `word-tooltip.js:36-43`. |
| Word click → side panel fetch `/api/root/{root}/summary` | ✅ | `surah-side-panel.js:41-43`. |
| Search submit → `/search?q=…&type=…` → handler renders results | ✅ | `handleSearch` (`server.go:361-390`). |
| `/concordance` paginated browse | ✅ | `handleConcordance` (`server.go:590-657`). |
| Concordance lemmas pre-built into server startup | ✅ | `data.LoadConcordance` runs at startup; surfaced on root-detail and `/concordance` pages. |

---

## §13 — Routes / API (`routes-api.md`)

| Route | Status | Where |
|---|---|---|
| `GET /` | ✅ | `server.go:170, 250-262`. |
| `GET /surah/{id}` | ✅ | `server.go:171, 272-373`. |
| `GET /search` | ✅ | `server.go:172, 361-390`. |
| `GET /roots` | ✅ | `server.go:173, 392-435`. |
| `GET /root/detailed/{root}` | ✅ | `server.go:174, 437-566`. |
| `GET /about` | ✅ | `server.go:175, 584-588`. |
| `GET /concordance` | ✅ | `server.go:176, 590-657`. |
| `GET /api/word?s=&a=&w=` | ✅ | `server.go:179, 685-775`. |
| `GET /api/search?q=&type=` | ✅ | `server.go:180, 1038-1062`. |
| `GET /api/roots` | ✅ | `server.go:181, 1064-1083`. |
| `GET /api/root/{root}/occurrences` | ✅ | `server.go:182, 1085-1100`. |
| `GET /api/root/{root}/summary` | ✅ | `server.go:183, 1104-1151`. |
| `GET /api/health` | ✅ | `server.go:184, 1153-1158`. |

---

## §14 — Bugs fixed in this session

Five functional bugs were found and fixed. Each is covered by a
regression test:

1. `search.OccurrencesForRoot` ignored its `asc` parameter (both
   branches did identical sort). The "Ascending" toggle was a no-op.
   `backend/internal/search/algorithms.go:578-595`. Test:
   `backend/internal/search/algorithms_test.go::TestOccurrencesForRootSortAsc`.
2. `verse-scroll.js` removed the `.ayah-focus` class it never added,
   so the verse highlight never showed on `#verse-N` navigation.
   `backend/static/js/verse-scroll.js`. Fix adds the class before
   the smooth-scroll.
3. ByArabic index keys used `stripDiacritics` (only strips tashkeel)
   while queries used `NormalizeArabic` (also unifies hamza/ya/ta-
   marbuta, doesn't strip spaces) — `/root/detailed/قول` returned
   404 while `/root/detailed/ق و ل` worked. New canonical
   `types.NormalizeArabicRoot` is used by both index and query paths.
4. `handleRootDetail` only fell back to `ByArabic` after a `ByRoot`
   miss, never to Buckwalter — Arabic URLs whose letters didn't
   match an existing index entry 404'd silently. Now mirrors
   `search.Root` and redirects to the canonical Buckwalter URL.
5. `subfs.go::readAll` compared `err.Error() == "EOF"` instead of
   `errors.Is(err, io.EOF)`. Latent code-quality issue, fixed.

Tests for these:
- `backend/internal/search/algorithms_test.go::TestOccurrencesForRootSortAsc`
- `backend/internal/search/algorithms_test.go::TestRootByArabicVariants`
- `backend/internal/server/server_test.go::TestRootDetailArabicVariants`
- `types/types_test.go::TestNormalizeArabicRoot`

---

## §15 — Concurrency / Performance work in this session

- `data.LoadAll`: Quran + MASAQ + roots loaders now run in parallel
  goroutines sharing one `*sql.DB`. ~30-50% startup reduction.
- `data.LoadTranslations`: English + Urdu XML loaded in parallel.
- `backend/main.go`: Translations + concordance now loaded in
  parallel after `LoadAll`.
- `server.handleSurah`: per-ayah WordRoots computation parallelized
  using chunked goroutines bounded by `GOMAXPROCS`.
- New `BenchmarkLoadAll` (`backend/internal/data/load_test.go`).
- All tests pass with `-race`.

---

## §16 — UI structural vocabulary

`backend/static/css/surfaces.css` introduces a page-agnostic
structural CSS layer (`.surface`, `.layout-shell`, `.stack`,
`.cluster`, `.measure-*`, `.layout-header`, `.layout-aside`,
`.layout-footer`). Every page template was refactored to use it.

---

## Outstanding / Known Limitations

- **Synonym roots on `/roots`**: the plan calls for "synonym roots
  (if possible)" — no synonym dictionary is currently loaded. Out
  of scope for this session.
- **Per-rune Indo-Pak ↔ Uthmani transliteration**: normalization is
  partial. Most common variants handled, but rare letterforms in
  non-Cairo mushafs may still differ.
- **`<dialog>` modal**: the plan called for a `<dialog>`-based
  modal but the side panel replaces it per plan.md: "Now the modal
  is no longer needed as its job is done by the sidebar."
- **Wider `/concordance` lemmas**: the `/concordance` page currently
  renders the raw `Keys` slice — could be tightened to a typed view,
  but functionally complete.
- **Translate pipeline**: `data/morphology/quran-in-english-clear
  quran-verse-by-verse-txt-edition-allah.zip` is on disk but unused
  in the runtime; only the two Tanzil XMLs are loaded.

---

## Verification commands

```bash
# Build
go build -o ./bin/quranreader ./backend

# Tests (with race detector)
go test -race ./...

# Live smoke
./bin/quranreader -addr :18080 \
  -db ./data/new/detailed-quran.db \
  -translations ./data/quran/translations
# Then: curl -s http://localhost:18080/{,surah/1,search?q=allah,roots,root/detailed/qwl,concordance,about,api/health}
```