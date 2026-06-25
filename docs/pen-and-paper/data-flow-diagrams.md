# Data Flow Diagrams — Quran Reader App

## 1. Request → Response (Page: Surah View)

```
┌──────────┐  GET /surah/2        ┌──────────────────┐
│ Browser  │ ─────────────────────▶│ Go ServeMux      │
└──────────┘                       └────────┬─────────┘
                                            │
                                            ▼
                              ┌──────────────────────────┐
                              │ handleSurah()            │
                              │  1. parse id from path    │
                              │  2. lookup quran.Surahs   │
                              │  3. compute prev/next     │
                              │  4. render template       │
                              └──────────────┬───────────┘
                                             │
                                             ▼
                              ┌──────────────────────────┐
                              │ templates/surah.tmpl     │
                              │   Execute(w, data)       │
                              │   reads quran, static    │
                              └──────────────┬───────────┘
                                             │
                  ┌──── HTML (one round trip) ┘
                  ▼
┌──────────────────────────────────────────────────────────┐
│ <html>                                                    │
│   <head>… CSS preloads, fonts …</head>                   │
│   <body>                                                   │
│     <header class="global-header">…</header>              │
│     <main class="page">                                   │
│       <section class="ayah" data-ayah="1">                │
│         <p class="ayah-text">                              │
│           <button class="word" data-surah=2 …>ٱلٓمٓ</button>│
│           <button class="word" data-surah=2 …>ذَٰلِكَ</…> │
│           …                                                │
│           <span class="mark division-marker">۞</span>      │
│         </p>                                              │
│       </section>                                          │
│       …                                                   │
│     </main>                                                │
│   </body>                                                  │
│ </html>                                                    │
└──────────────────────────────────────────────────────────┘
```

The Quran body ships **without** the morphology data inlined — we render the word `<button>` only, and fetch the details lazily on hover/click. This keeps the HTML payload small (~10–20 KB per surah vs the ~200 KB it would be with inlined morphology).

## 2. Word Hover → Tooltip

```
user hovers a word
   │
   ▼
[word-tooltip.js] event delegation: mouseenter on .word
   │
   │  1. read data-surah / data-ayah / data-word attrs
   │  2. construct cache key "w:{s}:{a}:{w}"
   │  3. if cache hit → render immediately, position
   │     else:
   ▼
fetch('/api/word?s=2&a=1&w=2', { cache: 'force-cache' })
   │
   ▼
[handleAPIWord]
   │  1. parse query
   │  2. masaq.ByWord[LocKey(s,a,w)]
   │  3. roots.ByLoc[LocKey(s,a,w)] → root key → root entry
   │  4. respondJSON
   ▼
[word-tooltip.js]
   │  5. populate tooltip DOM with:
   │       - Arabic word + transliteration (optional)
   │       - segment glosses (each on its own row)
   │       - root link (if present)
   │  6. position tooltip:
   │       - getBoundingClientRect() of word
   │       - compute tooltip width
   │       - if word.top - tooltip.height > 0 → above
   │         else → below
   │       - center horizontally; clamp to viewport edges
   │  7. unhide (display = block)
   │  8. mouseleave on .word OR Escape → hide
   ▼
user sees tooltip
```

## 3. Word Click → Modal

```
user clicks a word
   │
   ▼
[word-modal.js] event delegation: click on .word
   │
   │  1. preventDefault (don't trigger tooltip hide path)
   │  2. show loading state (spinner) in modal
   │  3. fetch('/api/word?s=…&a=…&w=…')
   ▼
[handleAPIWord] → JSON
   │
   ▼
[word-modal.js]
   │  4. populate modal body:
   │       - Arabic word (large)
   │       - segments table (segmented | tag | gloss)
   │       - root box: arabic + buckwalter + meaning en/ar
   │       - link: "see all occurrences → /root/{buckwalter}"
   │  5. modal.showModal() — native <dialog>
   │  6. focus first interactive element inside modal
   │  7. set up close handlers (Escape, backdrop click, × button)
   │  8. trap focus (Tab cycles within modal)
   ▼
user reads + dismisses modal → focus returns to word button
```

## 4. Search Submission

```
user types in global header search, picks "english", hits Enter
   │
   ▼
[global-header.js] form submit → preventDefault
   │
   │  navigate: location.href = `/search?q=${encodeURIComponent(q)}&type=english`
   ▼
[handleSearch]
   │  1. parse q, type
   │  2. dispatch:
   │       case "root"     → search.Root(q, 50)
   │       case "english"  → search.English(q, 50)
   │       case "arabic"   → search.Arabic(q, 50)
   │  3. results is []SearchResult{
   │       { kind, surah, ayah, word?, root?, text, snippet }
   │     }
   │  4. template render
   ▼
HTML page with results, snippet has <mark>…</mark> around the match
```

### Search Algorithms (detailed)

#### `search.Arabic(q)` — Arabic word match

```
input: q (string, may or may not have tashkeel, may be Indo-Pak form)
limit: 50
output: []SearchResult

1. normQ = normalizeArabic(q)        // strip tashkeel, unify letters
2. results = []
3. for each entry in masaq.ByWord (iterate map values):
     wordNorm = normalizeArabic(entry.Word)
     if wordNorm == normQ:
         exact match → results.append({ kind: "exact", … })
     else if wordNorm.includes(normQ) || normQ.includes(wordNorm):
         partial match → results.append({ kind: "partial", … })
4. cap results at 50
5. return results
```

Cost: O(77k) per search, ~50µs each iteration = ~4 ms worst case. No external library needed.

#### `search.English(q)` — English gloss match

```
input: q (English phrase)
limit: 50
output: []SearchResult

1. normQ = strings.ToLower(strings.TrimSpace(q))
2. if len(normQ) < 2: return []
3. results = []
4. for each entry in masaq.ByWord:
     gloss = strings.ToLower(entry.Gloss)
     if gloss == normQ:
         exact → push
     else if levenshtein(gloss, normQ) <= 2 and len(normQ) <= 12:
         fuzzy → push
     else if strings.Contains(gloss, normQ):
         substring → push
5. sort by (kind priority, levenshtein asc)
6. cap at 50
7. return
```

Levenshtein with early termination (band = 2) is O(|a|·|b|) but bounded by `min(50, 2·|q|+1)·|q|`. Fine for 77k entries — total ~30 ms cold.

#### `search.Root(q)` — root match

```
input: q (Arabic letters or Buckwalter)
limit: 50
output: []SearchResult

1. detect form:
     - if all ASCII: treat as buckwalter
     - else: normalize to buckwalter (arabic → bw map)
2. if exact buckwalter key exists in roots.ByRoot → top result
3. else prefix search:
     for each key in roots.ByRoot:
       if strings.HasPrefix(key, normQ):
         results.push(...)
4. cap 50, sort by occurrences desc
```

## 5. Verse Anchor Scroll (on load)

```
/surah/2/255  → page renders, then:
   │
   ▼
[verse-scroll.js] on DOMContentLoaded
   │
   │  read r.pathname → extract verse number
   │  target = document.getElementById(`verse-255`)
   │  if target:
   │    target.scrollIntoView({ behavior: 'smooth', block: 'start' })
   │    target.classList.add('verse-focus')  // brief highlight
   │    setTimeout(() => target.classList.remove('verse-focus'), 2000)
   ▼
user lands directly at the ayah, with a soft golden ring around it that fades
```

## 6. Build Pipeline Flow

```
[data/processed/]                          [data/] (raw)
   quran.gob              ◀── GOB encode   ─ quran-uthmani.xml
   masaq_index.gob        ◀── GOB encode   ─ MASAQ.csv
   roots.gob              ◀── GOB encode   ─ corpus-roots.json
                                       ─    meanings-roots-ai.jsonl
   meta.json              ◀── JSON write   ─ (computed)

   ┌─────────────────────────┐
   │  pipeline/build.go      │
   │                         │
   │  parse XML ─→ Quran{}   │    in-memory structures built once
   │  parse CSV ─→ MasaqIdx{}│    in-memory structures built once
   │  parse JSON+JSONL ─→    │
   │     RootsIdx{}          │
   │                         │
   │  validate:              │    assertions + counts → stderr
   │    - 114 surahs         │
   │    - 6236 ayas          │
   │    - every word has     │
   │      masaq entries      │
   │    - root join          │
   │      coverage reported  │
   │                         │
   │  emit gob/json files    │
   └─────────────────────────┘
            │
            ▼
   docs/data-verification/
      build-report.txt      ◀── counts + mismatches (if any)
```

## 7. Theme Persistence Flow

```
[app.js] on DOMContentLoaded
   │
   │  read localStorage["qr.theme"]; default → matchMedia
   │  <html data-theme="dark|light">
   │
[global-header.js] theme toggle click
   │
   │  1. toggle html.dataset.theme
   │  2. localStorage["qr.theme"] = new
   │
   ▼
inline <script> in <head> runs FIRST (before paint)
   to prevent flash-of-wrong-theme.
```

## 8. Font Loading Flow

```
<head>
  <link rel="preload" href="/static/fonts/hafs.woff2" as="font" type="font/woff2" crossorigin>
  <link rel="stylesheet" href="/static/css/tokens.css">   <!-- declares @font-face -->
</head>

@font-face { font-display: swap; … }

Result: text is visible immediately with fallback font,
        then re-renders with hafs.woff2 when loaded (~10–50 ms).
        No FOIT (flash of invisible text).
```

