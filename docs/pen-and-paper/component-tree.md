# Component Tree — Quran Reader App UI

Atomic components are the smallest user-facing blocks. Each has a clear data contract (props) and behavior. The tree shows the nesting for each page template.

## Page Templates (Top Level)

```
layout.tmpl            (shared shell — html, head, body, global header slot)
├── homepage.tmpl      (slot: <main id="homepage">)
├── surah.tmpl         (slot: <main id="surah-view">)
├── search.tmpl        (slot: <main id="search-view">)
├── roots-list.tmpl    (slot: <main id="roots-view">)
├── root-detail.tmpl   (slot: <main id="root-detail-view">)
├── about.tmpl         (slot: <main id="about">)
└── error.tmpl         (404 / 500)
```

## Shared Shell (`layout.tmpl`)

```
<html lang="ar" dir="rtl" data-theme="dark|light">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width,initial-scale=1">
    <link rel="preload" href="/static/fonts/hafs.woff2" as="font" type="font/woff2" crossorigin>
    <link rel="stylesheet" href="/static/css/tokens.css">
    <link rel="stylesheet" href="/static/css/layout.css">
    <link rel="stylesheet" href="/static/css/quran.css">
    <link rel="stylesheet" href="/static/css/components.css">
    <title>{{.Title}} — Quran Reader</title>
  </head>
  <body class="app">
    {{template "global-header" .}}
    <main class="page">
      {{block "content" .}}{{end}}
    </main>
    {{template "word-tooltip" .}}     <!-- empty tooltip stub, populated by JS -->
    {{template "word-modal" .}}       <!-- empty modal stub, populated by JS -->
    <script src="/static/js/app.js" defer></script>
  </body>
</html>
```

## Global Header (sticky)

```
<header class="global-header" data-sticky="top">
  <div class="gh-inner">
    <a class="gh-brand" href="/">Quran Reader</a>
    <form class="gh-search" role="search" data-component="global-search">
      <input type="search" name="q" placeholder="Search…" aria-label="Search">
      <select name="type" aria-label="Search type">
        <option value="root">root</option>
        <option value="english">english</option>
        <option value="arabic">arabic</option>
      </select>
      <button type="submit">Search</button>
    </form>
    <nav class="gh-nav">
      <a href="/roots">Roots</a>
      <a href="/about">About</a>
      <button class="gh-theme" data-component="theme-toggle" aria-label="Toggle theme">◐</button>
    </nav>
  </div>
</header>
```

- **State**: theme (localStorage), search input value
- **Behavior**: Enter in search → `GET /search?q=...&type=...`. Theme toggle flips `data-theme` on `<html>`.

## Surah Page

```
surah.tmpl
├── SurahHeader (sticky below global header)
│   ├── <h1>{{.Surah.Name}} <small>{{.Surah.NameLatin}}</small></h1>
│   ├── Local search box (data-component="surah-local-search")
│   ├── Font-size slider (data-component="font-size", min=10 max=40 step=1 value=16)
│   ├── Line-height slider (data-component="line-height", min=1 max=3 step=0.1 value=1.6)
│   └── Prev / Next surah buttons
├── Bismillah (if present, centered, large, gold accent)
└── SurahBody (paragraph flow)
    └── for each ayah:
        ├── <span class="ayah-anchor" id="verse-{n}" aria-label="Ayah {n}"></span>
        ├── <span class="ayah-number">{n}</span>
        └── <p class="ayah-text">
              for each token (WordToken):
                ├── if mark: <span class="mark {kind}">{value}</span>
                └── if word:
                    <button class="word"
                       data-surah="{s}" data-ayah="{a}" data-word="{w}"
                       data-component="word"
                       aria-haspopup="dialog">{{text}}</button>
```

**Word buttons vs anchors**: a `<button>` is used (not `<a>`) because they trigger in-page popovers/modals, not navigation. They have `aria-haspopup="dialog"` to announce the modal relationship to assistive tech.

## Word Tooltip (popover)

```
<div class="word-tooltip" role="tooltip" hidden data-component="word-tooltip"
     data-state="loading|ready|empty">
  <!-- populated by word-tooltip.js -->
</div>
```

- Position: viewport-edge-aware; centered above the word, flipped below if no space.
- Content: glosses from MASAQ (one per segment), root link if present.
- Triggered by `mouseenter` / `focus`; dismissed on `mouseleave` / `Escape`.

## Word Modal (dialog)

```
<dialog class="word-modal" data-component="word-modal">
  <header>
    <h2>Word analysis</h2>
    <button class="close" data-action="close" aria-label="Close">×</button>
  </header>
  <div class="word-modal-body">
    <!-- populated: arabic word, segments table, root link, occurrences link -->
  </div>
</dialog>
```

- Triggered by click on `.word` button.
- Loaded with `fetch /api/word?s=..&a=..&w=..`.
- Closes on `Escape`, on backdrop click, on close button.
- When opened, focus moves into the dialog; trap focus inside.

## Search Results Page

```
search.tmpl
└── <main>
    <header class="search-summary">
      <h1>Results for "{q}"</h1>
      <span class="search-meta">{N} results in {type}</span>
    </header>
    <ol class="search-results">
      {{range .Results}}
        <li class="search-result" data-kind="{{.Kind}}">
          <a href="{{.Link}}">{{.Title}}</a>
          <p class="search-snippet">{{.Snippet}}</p>
          <small class="search-meta">{{.Meta}}</small>
        </li>
      {{end}}
    </ol>
```

`Snippet` has the matched substring wrapped in `<mark>` for highlighting (escaped properly).

## Roots List Page

```
roots-list.tmpl
└── <main>
    <header class="roots-header">
      <h1>Roots</h1>
      <div class="roots-controls">
        <label>Sort:
          <select name="sort">
            <option value="desc">Most frequent first</option>
            <option value="asc">Least frequent first</option>
            <option value="alpha">Alphabetical (Arabic)</option>
          </select>
        </label>
        <label>Surah:
          <input type="number" name="surah" min="1" max="114" placeholder="Any">
        </label>
      </div>
    </header>
    <ol class="roots-list">
      {{range .Roots}}
        <li class="root-row">
          <a href="/root/{{.Buckwalter}}" class="root-link">
            <span class="root-arabic">{{.Arabic}}</span>
            <span class="root-buckwalter">{{.Buckwalter}}</span>
          </a>
          <span class="root-count">{{.Occurrences}}×</span>
          <p class="root-meaning">{{truncate .MeaningEN 120}}</p>
        </li>
      {{end}}
    </ol>
    {{template "pagination" .}}
```

Pagination = `?page=N` (default 50 per page), simple Prev/Next links.

## Root Detail Page

```
root-detail.tmpl
└── <main>
    <header class="root-header">
      <h1><span class="root-arabic">{{.Root.Arabic}}</span>
          <small>({{.Root.Buckwalter}})</small></h1>
      <span class="root-pos">{{.Root.POS}}</span>
      <span class="root-count">{{.Root.Occurrences}} occurrences</span>
    </header>
    <section class="root-meaning">
      <h2>Meaning</h2>
      <p>{{.Root.MeaningEN}}</p>
      <p lang="ar">{{.Root.MeaningAR}}</p>
    </section>
    <section class="root-occurrences">
      <h2>Occurrences</h2>
      <ol class="root-occ-list">
        {{range .Occurrences}}
          <li>
            <a href="/surah/{{.Surah}}/{{.Ayah}}">Surah {{.Surah}}:{{.Ayah}}</a>
            <span class="occ-word">{{.Word}}</span>
          </li>
        {{end}}
      </ol>
    </section>
    <p class="data-attribution">
      Meaning generated by AI — <a href="/about">learn more</a>.
    </p>
```

## Error Pages

```
error.tmpl
└── <main class="error-page">
      <h1>{status}</h1>
      <p>{message}</p>
      <a href="/">← Back to home</a>
```

## Component → JS Module Mapping

| Component (data-component) | JS Module | Trigger |
|---|---|---|
| `global-search` | `app.js` → submits to `/search` | form submit |
| `theme-toggle` | `app.js` | click |
| `surah-local-search` | `surah-header.js` | Enter → scroll to first ayah containing word |
| `font-size`, `line-height` | `surah-header.js` | input → apply CSS variable on `<html>` |
| `word` (every word button) | `word-tooltip.js`, `word-modal.js` | hover/focus, click |
| `word-modal` (close) | `word-modal.js` | click, Escape, backdrop |
| `pagination` | `app.js` | click → reload with `?page=N` |

## JavaScript Modules (vanilla, no bundler)

```
static/js/
├── app.js              — router, layout glue, theme, global search submit
├── global-header.js    — search box helpers, theme toggle
├── surah-header.js     — local search, font-size, line-height, prev/next
├── word-tooltip.js     — event delegation, fetch, positioning
├── word-modal.js       — event delegation, fetch, focus trap
├── verse-scroll.js     — on load, scroll to #verse-N if present
└── search.js           — input highlight (only when needed; result highlighting is server-side)
```

Each module is **event-delegated** on a parent container (no per-element listeners), uses `fetch` + DOM API, cleans up on close.

