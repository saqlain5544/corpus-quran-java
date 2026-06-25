# UI Design System — Quran Reader App

Typography-first, editorial design. Inspired by print typography (Manshūr / Mushaf tradition adapted for screen). **No garish colors** — restrained, ink-and-paper palette.

## 1. Design Tokens (CSS Custom Properties)

```css
/* tokens.css — :root and [data-theme] blocks */

:root {
  /* ── Typography ────────────────────────────── */
  --font-arabic:    "Hafs", "Amiri Quran", "Amiri", "Scheherazade New", "Noto Naskh Arabic", serif;
  --font-latin:     "Iowan Old Style", "Charter", "Georgia", serif;
  --font-mono:      ui-monospace, "SF Mono", "Menlo", monospace;

  --fs-base:        16px;     /* user-adjustable on surah pages */
  --fs-scale:       1.25;     /* modular scale ratio */
  --lh-base:        1.6;      /* user-adjustable */

  --fs-xs:          calc(var(--fs-base) * 0.75);
  --fs-sm:          calc(var(--fs-base) * 0.875);
  --fs-md:          var(--fs-base);
  --fs-lg:          calc(var(--fs-base) * var(--fs-scale));
  --fs-xl:          calc(var(--fs-lg)  * var(--fs-scale));
  --fs-2xl:         calc(var(--fs-xl)  * var(--fs-scale));
  --fs-3xl:         calc(var(--fs-2xl) * var(--fs-scale));
  --fs-arabic-md:   calc(var(--fs-base) * 1.75);  /* Quran body */
  --fs-arabic-lg:   calc(var(--fs-arabic-md) * 1.15);

  /* ── Spacing (4px base) ─────────────────────── */
  --sp-1:           0.25rem;
  --sp-2:           0.5rem;
  --sp-3:           0.75rem;
  --sp-4:           1rem;
  --sp-6:           1.5rem;
  --sp-8:           2rem;
  --sp-12:          3rem;
  --sp-16:          4rem;

  /* ── Layout ────────────────────────────────── */
  --measure-narrow: 32rem;   /* ~70ch */
  --measure-base:   42rem;   /* ~90ch */
  --measure-wide:   60rem;
  --measure-quran:  48rem;   /* Quran body column */

  --grid-cols:      12;
  --grid-gap:       var(--sp-4);
  --grid-max:       80rem;

  /* ── Z-index scale (strict) ────────────────── */
  --z-base:         1;
  --z-sticky:       100;     /* global header */
  --z-tooltip:      200;
  --z-modal:        300;
  --z-modal-fade:   250;

  /* ── Motion ────────────────────────────────── */
  --ease-out:       cubic-bezier(.2, .8, .2, 1);
  --dur-fast:       120ms;
  --dur-base:       200ms;
  --dur-slow:       320ms;

  /* ── Borders ───────────────────────────────── */
  --radius-sm:      2px;
  --radius-md:      4px;
  --radius-lg:      8px;

  --border-w:       1px;
}

/* Light theme — paper-and-ink */
[data-theme="light"] {
  --bg:             #faf8f4;      /* warm paper */
  --bg-elev:        #ffffff;
  --bg-sunken:      #f0ebe1;
  --ink:            #1a1612;      /* deep ink */
  --ink-muted:      #5c544a;
  --ink-soft:       #8a8278;
  --rule:           #d8d2c5;      /* hairline */
  --accent:         #8a5a2b;      /* manuscript gold */
  --accent-soft:    #c9a36b;
  --error:          #a93226;
  --success:        #2d6a4f;
  --mark-quran-rub: #8a5a2b;
  --mark-sajdah:    #2d6a4f;
  --mark-pause:     #8a8278;
  --word-hover-bg:  rgba(138, 90, 43, 0.08);
  --tooltip-bg:     #ffffff;
  --modal-backdrop: rgba(26, 22, 18, 0.4);
}

/* Dark theme — night reading */
[data-theme="dark"] {
  --bg:             #14110d;
  --bg-elev:        #1e1a14;
  --bg-sunken:      #0f0d09;
  --ink:            #ece5d6;
  --ink-muted:      #a89e8b;
  --ink-soft:       #6b6358;
  --rule:           #2a2520;
  --accent:         #d4a86a;
  --accent-soft:    #8a5a2b;
  --error:          #d97a6c;
  --success:        #7fbf9e;
  --mark-quran-rub: #d4a86a;
  --mark-sajdah:    #7fbf9e;
  --mark-pause:     #6b6358;
  --word-hover-bg:  rgba(212, 168, 106, 0.10);
  --tooltip-bg:     #1e1a14;
  --modal-backdrop: rgba(0, 0, 0, 0.6);
}

@media (prefers-color-scheme: dark) {
  :root:not([data-theme]) {
    /* mirrors [data-theme="dark"] block */
    --bg: #14110d; --ink: #ece5d6; /* … same as dark … */
  }
}
```

## 2. Grid System (12 columns)

```css
.grid {
  display: grid;
  grid-template-columns: repeat(var(--grid-cols), minmax(0, 1fr));
  gap: var(--grid-gap);
}
.col-12 { grid-column: span 12; }
.col-9  { grid-column: span 9; }
.col-8  { grid-column: span 8; }
.col-6  { grid-column: span 6; }
.col-4  { grid-column: span 4; }
.col-3  { grid-column: span 3; }

@media (max-width: 768px) {
  .col-md-12 { grid-column: span 12; }
  .col-md-6  { grid-column: span 6; }
}
```

**Usage policy**: grid is for top-level layout (header, homepage columns, roots list grid). Components **inside** a content area use their own internal layouts (flex, intrinsic sizing).

## 3. Typography Rules

| Surface | Font | Size | Line Height | Notes |
|---|---|---|---|---|
| Quran body (`.ayah-text`) | `var(--font-arabic)` | `var(--fs-arabic-md)` | 2.0 | Generous leading for diacritics |
| Quran word hover state | `var(--font-arabic)` | inherit | inherit | Only color changes |
| Ayah number marker | `var(--font-arabic)` | `var(--fs-md)` | 1 | Inline at ayah start, gold accent |
| Page titles (h1) | `var(--font-latin)` | `var(--fs-2xl)` | 1.2 | Latin script, English UI |
| Body text (English) | `var(--font-latin)` | `var(--fs-md)` | `var(--lh-base)` | |
| Captions / meta | `var(--font-latin)` | `var(--fs-sm)` | 1.4 | `--ink-muted` color |

**Mixed-direction handling**:
- `<html lang="ar" dir="rtl">` — base RTL.
- Sections that are mostly English (e.g., search summary) wrap in `<div dir="ltr">` to override.
- The bismillah verse and Quran body use `text-align: justify` with `text-justify: inter-word`.

## 4. Component Tokens (semantic classes)

### Global Header

```css
.global-header {
  position: sticky;
  top: 0;
  z-index: var(--z-sticky);
  background: var(--bg-elev);
  border-bottom: var(--border-w) solid var(--rule);
}
.gh-inner {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: var(--sp-4);
  max-width: var(--grid-max);
  margin: 0 auto;
  padding: var(--sp-3) var(--sp-4);
}
.gh-brand  { font: var(--fs-lg)/1 var(--font-latin); font-weight: 600; letter-spacing: -0.01em; text-decoration: none; color: var(--ink); }
.gh-search { display: flex; gap: var(--sp-2); }
.gh-search input,
.gh-search select,
.gh-search button {
  padding: var(--sp-2) var(--sp-3);
  background: var(--bg);
  color: var(--ink);
  border: var(--border-w) solid var(--rule);
  border-radius: var(--radius-md);
  font: inherit;
}
.gh-search input:focus { outline: 2px solid var(--accent); outline-offset: 1px; }
.gh-search button { background: var(--accent); color: var(--bg); border-color: var(--accent); }
.gh-nav { display: flex; gap: var(--sp-3); align-items: center; }
.gh-theme { background: none; border: none; cursor: pointer; color: var(--ink); font-size: var(--fs-lg); }
```

### Surah Paragraph

```css
.ayah {
  display: grid;
  grid-template-columns: 4rem 1fr;
  gap: var(--sp-4);
  align-items: baseline;
  margin: var(--sp-8) 0;
  padding: 0 var(--sp-4);
}
.ayah-number {
  font: 600 var(--fs-md)/1 var(--font-arabic);
  color: var(--accent);
  text-align: center;
}
.ayah-text {
  font: var(--fs-arabic-md)/2 var(--font-arabic);
  text-align: justify;
  text-justify: inter-word;
  hyphens: none;
  word-spacing: 0.05em;
}
.ayah-text .word {
  display: inline-block;
  padding: 0 var(--sp-1);
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  cursor: pointer;
  border-radius: var(--radius-sm);
  transition: background var(--dur-fast) var(--ease-out);
}
.ayah-text .word:hover,
.ayah-text .word:focus-visible {
  background: var(--word-hover-bg);
  outline: none;
}
.ayah-text .mark { padding: 0 var(--sp-1); user-select: none; }
.ayah-text .mark.division-marker { color: var(--mark-quran-rub); font-weight: 600; }
.ayah-text .mark.sajdah         { color: var(--mark-sajdah); font-weight: 600; }
.ayah-text .mark.waqf           { color: var(--mark-pause); }
```

### Tooltip

```css
.word-tooltip {
  position: fixed; /* JS-positioned */
  z-index: var(--z-tooltip);
  background: var(--tooltip-bg);
  color: var(--ink);
  border: var(--border-w) solid var(--rule);
  border-radius: var(--radius-md);
  padding: var(--sp-3);
  font: var(--fs-sm)/1.4 var(--font-latin);
  box-shadow: 0 6px 24px rgba(0,0,0,0.10);
  max-width: 24rem;
  pointer-events: none; /* so hover on word doesn't get stolen */
  transition: opacity var(--dur-fast) var(--ease-out);
}
.word-tooltip[hidden] { opacity: 0; pointer-events: none; }
.word-tooltip dl { display: grid; grid-template-columns: max-content 1fr; gap: var(--sp-1) var(--sp-2); }
.word-tooltip dt { color: var(--ink-muted); }
.word-tooltip dd { margin: 0; }
```

### Modal

```css
.word-modal {
  border: none;
  padding: 0;
  background: transparent;
  max-width: var(--measure-base);
  width: 92vw;
}
.word-modal::backdrop {
  background: var(--modal-backdrop);
  z-index: var(--z-modal-fade);
}
.word-modal > .modal-shell {
  background: var(--bg-elev);
  color: var(--ink);
  border-radius: var(--radius-lg);
  padding: var(--sp-6);
  z-index: var(--z-modal);
  box-shadow: 0 24px 64px rgba(0,0,0,0.25);
  display: grid;
  gap: var(--sp-4);
}
```

### Search Results

```css
.search-result { padding: var(--sp-4) 0; border-bottom: var(--border-w) solid var(--rule); }
.search-result mark { background: var(--accent-soft); color: var(--bg); padding: 0 var(--sp-1); border-radius: var(--radius-sm); }
```

## 5. Accessibility Checklist

- [ ] All interactive elements have visible focus rings (`outline: 2px solid var(--accent)`).
- [ ] Modal traps focus; closes on Escape; restores focus on close.
- [ ] Tooltip uses `role="tooltip"`, aria-describedby from the word button.
- [ ] Color contrast ≥ 4.5:1 for body text, ≥ 3:1 for large text (tested against both themes).
- [ ] Skip link to main content: `<a class="skip" href="#main">Skip to content</a>`.
- [ ] `dir="rtl"` on `<html>`; explicit `dir="ltr"` on Latin-only sections.
- [ ] All buttons have `aria-label` if icon-only.
- [ ] Quran body sets `lang="ar"` on the `<p class="ayah-text">`.
- [ ] Form fields have associated `<label>` elements (or `aria-label`).
- [ ] `prefers-reduced-motion` respected — disable transitions when set.

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after { transition-duration: 0.01ms !important; animation-duration: 0.01ms !important; }
}
```

## 6. Responsive Breakpoints

| Breakpoint | Behavior |
|---|---|
| ≥ 1024px | 12-col grid active; Quran body at `var(--measure-quran)` width, centered. |
| 768–1023px | Quran body fills the column; SurahHeader controls stack. |
| < 768px | Global header collapses to 2 rows (brand + theme; search full-width below). Font-size slider 14–28. |

## 7. Font Loading

```css
@font-face {
  font-family: "Hafs";
  src: url("/static/fonts/hafs.woff2") format("woff2");
  font-display: swap;
  unicode-range: U+0600-06FF, U+0750-077F, U+08A0-08FF, U+FB50-FDFF, U+FE70-FEFF;
}
```

The preload in `<head>` ensures hafs.woff2 starts loading in parallel with CSS — critical for first paint of Quran body text.

## 8. State Persistence

| State | Storage | Restored on |
|---|---|---|
| Theme | `localStorage["qr.theme"]` | every page load (before paint, via inline script) |
| Font size / line height | `localStorage["qr.reader"]` | surah pages only |
| Last searched query | session-only | n/a |

Inline theme init script in `<head>` (before stylesheet) prevents flash-of-wrong-theme:

```html
<script>
  try {
    const t = localStorage.getItem("qr.theme") ||
              (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
    document.documentElement.setAttribute("data-theme", t);
  } catch (e) {}
</script>
```

## 9. Iconography

No icon font. Unicode glyphs where possible:
- Theme toggle: ◐ (U+25D0)
- Close (modal): × (U+00D7)
- Dropdown caret: ▾ (U+25BE)

This keeps the static payload minimal (no icon font file).

