// surah-side-panel.js — left panel showing the full analysis
// for the clicked word. Concordance/lemma-frequency data is NOT
// shown here (per user instruction). The panel renders a single
// stacked column with this structure:
//
//   ┌──────────────────────────┐
//   │ Word  9:5:18            │  ← word + ref
//   │ [MorphSummary chips]     │  ← MorphTag × MorphType at top
//   │ Root   (arabic/buckwalter)│
//   │ Lemma  (from masaq)      │
//   │ Gloss  english gloss     │
//   │ ─ Masaq Tags ─           │
//   │ seg 1: prefix/stem/suffix│  ← full MASAQ data per segment
//   │ seg 2: ...                │
//   │ ─ Full Meaning (JSONL) ─  │  ← meaning, lexicons, examples
//   │ meaning_en (full)         │
//   │ meaning_ar (full)         │
//   │ Ibn Fāris: ...            │
//   │ al-Rāghib: ...            │
//   │ Quran example 2:255       │
//   │ Hadith: ...               │
//   │ Full root detail →        │
//   └──────────────────────────┘
(function () {
  "use strict";

  var panel = document.getElementById("surah-side-panel");
  if (!panel) return;
  var content = document.getElementById("ssp-content");
  if (!content) return;

  // Make the panel a proper dialog for screen readers.
  panel.setAttribute("role", "dialog");
  panel.setAttribute("aria-modal", "true");
  panel.setAttribute("aria-labelledby", "ssp-header-label");

  var cache = {};
  var currentLemma = "";
  // Monotonic counter — only the latest click is allowed to render.
  // Without this, clicking word B while word A's root fetch is
  // still in flight lets word A's stale rootData land in B's
  // panel (and pollutes the cache under B's key).
  var requestId = 0;
  var lastFocused = null;

  // ── Close ────────────────────────────────────────────────────
  // Escape closes; clicking outside the panel (on the surah body)
  // also closes. The visible close button below is the primary
  // affordance for touch users.
  // Mobile dimming — when the side panel opens on narrow viewports
  // it transforms into a bottom-sheet. To make the panel feel
  // modal-ish without becoming a full-screen overlay, dim the
  // underlying .surah-main via a class. The class is a no-op on
  // desktop (the CSS only applies it inside the mobile @media
  // query). */
  var surahMain = document.querySelector(".surah-main");

  function closePanel() {
    panel.classList.remove("is-open");
    if (surahMain) surahMain.classList.remove("is-dimmed");
    if (lastFocused && typeof lastFocused.focus === "function") {
      lastFocused.focus();
      lastFocused = null;
    }
  }

  function openPanel() {
    panel.classList.add("is-open");
    if (surahMain) surahMain.classList.add("is-dimmed");
  }

  // Add a visible close button to the header on first render.
  var header = panel.querySelector(".ssp-header");
  if (header && !header.querySelector(".ssp-close")) {
    var btn = document.createElement("button");
    btn.type = "button";
    btn.className = "ssp-close";
    btn.setAttribute("aria-label", "Close word analysis");
    btn.textContent = "✕";
    btn.addEventListener("click", closePanel);
    header.appendChild(btn);
  }

  // Escape closes.
  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && panel.classList.contains("is-open")) {
      e.preventDefault();
      closePanel();
    }
  });

  // Focus trap: when the panel is open, Tab cycles between its
  // focusable descendants (close button + any links inside content).
  panel.addEventListener("keydown", function (e) {
    if (e.key !== "Tab" || !panel.classList.contains("is-open")) return;
    var focusables = panel.querySelectorAll(
      'button:not([disabled]), a[href], [tabindex]:not([tabindex="-1"])'
    );
    if (focusables.length === 0) return;
    var first = focusables[0], last = focusables[focusables.length - 1];
    var active = document.activeElement;
    if (e.shiftKey && active === first) {
      e.preventDefault();
      last.focus();
    } else if (!e.shiftKey && active === last) {
      e.preventDefault();
      first.focus();
    }
  });

  document.addEventListener("click", function (e) {
    // Click outside the panel while it's open → close. Skip clicks
    // on other words (those are handled by the next branch) and on
    // the panel itself.
    if (panel.classList.contains("is-open")
        && !panel.contains(e.target)
        && !QR.dom.closestWord(e.target)) {
      closePanel();
      return;
    }

    var w = QR.dom.closestWord(e.target);
    if (!w) return;
    var root = w.dataset.root;
    var s = w.dataset.surah, a = w.dataset.ayah, wn = w.dataset.word;
    var key = s + ":" + a + ":" + wn;

    // Track the previously focused element so closePanel() can
    // restore focus to the word the user clicked (the trigger).
    lastFocused = w;

    if (cache[key]) {
      content.innerHTML = cache[key];
      openPanel();
      // Move focus to the close button so keyboard users land
      // inside the dialog and Esc / Tab work as expected.
      var cb = panel.querySelector(".ssp-close");
      if (cb) cb.focus();
      return;
    }

    content.innerHTML = '<p class="ssp-loading">Loading…</p>';
    openPanel();
    var cb = panel.querySelector(".ssp-close");
    if (cb) cb.focus();

    // Stamp this request; older fetches discard their results.
    var myRequest = ++requestId;
    var wordData = null;

    fetch("/api/word?s=" + encodeURIComponent(s) +
          "&a=" + encodeURIComponent(a) +
          "&w=" + encodeURIComponent(wn))
      .then(function (r) { return r.ok ? r.json() : Promise.reject(r.status); })
      .then(function (wd) {
        if (myRequest !== requestId) return; // stale
        wordData = wd;
        currentLemma = wd.lemma || "";
        if (root) {
          return fetch("/api/root/" + encodeURIComponent(root) + "/summary")
            .then(function (r) { return r.ok ? r.json() : null; });
        }
        return null;
      })
      .then(function (rootData) {
        if (myRequest !== requestId) return; // stale
        renderPanel(wordData, rootData);
        cache[key] = content.innerHTML;
      })
      .catch(function () {
        if (myRequest !== requestId) return; // stale
        content.innerHTML = '<p class="ssp-empty">Failed to load.</p>';
      });
  });

  // ── Render helpers ────────────────────────────────────────────

  // Render a labeled definition-list row: <dt>label</dt><dd>value</dd>.
  // `valueHtml` is inserted as raw HTML (caller's responsibility to
  // escape). `attrs` is an optional object of extra <dd> attributes.
  // Uses QR.dom.escape so the single escape helper in dom.js is the
  // canonical implementation — no duplicate escape logic anywhere.
  var esc = QR.dom.escape;
  function dl(label, valueHtml, attrs) {
    var attrStr = "";
    if (attrs) {
      for (var k in attrs) {
        attrStr += " " + k + '="' + esc(attrs[k]) + '"';
      }
    }
    return '<dl class="ssp-dl"><dt>' + esc(label) + '</dt><dd' +
      attrStr + '>' + valueHtml + '</dd></dl>';
  }

  // Render one MASAQ segment as a labelled row: the segmented word
  // + every non-empty field grouped by category. We previously
  // dumped every field into one long " · "-separated list, which
  // read as visual noise ("Stem · OTHER · ACRON · NOMINATIVE ·
  // SUKUN · INVAR" — six tags for one word, with multiple fields
  // describing the same thing). The grouped layout shows:
  //   POS:    MorphType · MorphTag  (e.g. "Stem · OTHER")
  //   Grammar: SyntacticRole, CaseMood(+marker), Phrase + Func.
  //   Decl.:   Invariable, PossessiveConstruct
  // …so the reader can scan by category rather than parse a flat
  // list. Duplicates are also suppressed (e.g. PREP/·PREP).
  function renderSegment(s) {
    var arabic = esc(s.SegmentedWord || s.Word || "");
    var out = '<div class="ssp-seg">';
    out += '<span class="ssp-seg-text" lang="ar" dir="rtl">' + arabic + '</span>';

    // Category 1 — Morphology (POS / segment type)
    var morphBits = [];
    if (s.MorphType) morphBits.push(s.MorphType);
    if (s.MorphTag) morphBits.push(s.MorphTag);
    if (morphBits.length) {
      out += ' <span class="ssp-tag ssp-tag--pos">' + esc(morphBits.join(' · ')) + '</span>';
    }

    // Category 2 — Grammar (SyntacticRole + Case + Phrase)
    var gramBits = [];
    if (s.SyntacticRole && s.SyntacticRole !== s.MorphTag) {
      gramBits.push(s.SyntacticRole);
    }
    if (s.CaseMood) {
      // CaseMoodMarker is the actual vowel (KASRA/DHAMMA/FATHA) —
      // surface it next to the abstract CaseMood tag.
      gramBits.push(s.CaseMoodMarker && s.CaseMoodMarker !== s.CaseMood
        ? s.CaseMood + ' ' + s.CaseMoodMarker
        : s.CaseMood);
    }
    if (s.Phrase) gramBits.push(s.Phrase);
    if (s.PhrasalFunction) gramBits.push(s.PhrasalFunction);
    if (gramBits.length) {
      out += ' <span class="ssp-tag ssp-tag--grammar">' + esc(gramBits.join(' · ')) + '</span>';
    }

    // Category 3 — Declension state (when non-default)
    var declBits = [];
    if (s.InvariableDeclinable) declBits.push(s.InvariableDeclinable);
    if (s.PossessiveConstruct && s.PossessiveConstruct !== "NOT_CONSTRUCT") {
      declBits.push(s.PossessiveConstruct);
    }
    if (declBits.length) {
      out += ' <span class="ssp-tag ssp-tag--decl">' + esc(declBits.join(' · ')) + '</span>';
    }

    out += '</div>';
    return out;
  }

  // Render a labeled section heading — gold uppercase caption with
  // a hairline rule below, matching the manuscript ink accent.
  function section(label) {
    return '<h3 class="ssp-section">' + esc(label) + '</h3>';
  }

  // Render a compact morphology summary for the current word's segments:
  // shows each unique (MorphTag, MorphType) pair as a chip with count.
  // For a single-segment word this is one chip; for prefixed words
  // (e.g. "وَقَالُوا") it shows each segment's pattern distinctly.
  function renderMorphSummary(segs) {
    if (!segs || segs.length === 0) return '';
    // Count unique (MorphTag, MorphType) combos.
    var counts = {};
    for (var i = 0; i < segs.length; i++) {
      var s = segs[i];
      var key = (s.MorphTag || '') + '|' + (s.MorphType || '');
      if (!counts[key]) counts[key] = { tag: s.MorphTag || '', type: s.MorphType || '', count: 0 };
      counts[key].count++;
    }
    var keys = Object.keys(counts);
    if (keys.length === 0) return '';

    var html = '<div class="ssp-morph-summary">';
    for (var k = 0; k < keys.length; k++) {
      var e = counts[keys[k]];
      var label = e.tag || '—';
      if (e.type) label = e.type + ' · ' + label;
      html += '<span class="ssp-morph-chip">';
      html += '<span class="ssp-morph-chip-label">' + esc(label) + '</span>';
      if (e.count > 1) {
        html += '<span class="ssp-morph-chip-count">' + e.count + '×</span>';
      }
      html += '</span>';
    }
    html += '</div>';
    return html;
  }

  function renderPanel(wordData, rootData) {
    var html = '';
    var segs = wordData.segments || [];

    // ── 1. Word + reference ──────────────────────────────────────
    html += '<div class="ssp-word">';
    html += '<span class="ssp-word-text" lang="ar" dir="rtl">' +
      esc(wordData.word_text) + '</span>';
    html += '<span class="ssp-word-ref">' +
      esc(wordData.surah + ":" + wordData.ayah + ":" + wordData.word) + '</span>';
    html += '</div>';

    // ── 1b. Morphology summary (top of panel) ─────────────────────
    // Show aggregated MorphTag × MorphType chips so the reader immediately
    // sees the morphological pattern before the per-segment breakdown.
    html += renderMorphSummary(segs);

    // ── 2. Root (Arabic / Buckwalter) ───────────────────────────
    if (wordData.root && wordData.root.buckwalter) {
      html += dl("Root",
        esc(wordData.root.arabic || "") +
        ' <span class="ssp-buckwalter">(' + esc(wordData.root.buckwalter) + ')</span>',
        {"lang": "ar", "dir": "rtl"}
      );
    }

    // ── 3. Lemma (bare stem, from MASAQ) ─────────────────────────
    if (wordData.lemma) {
      html += dl("Lemma", esc(wordData.lemma), {"lang": "ar", "dir": "rtl"});
    }

    // ── 4. Gloss + Translation ─────────────────────────────────
    if (wordData.gloss) {
      html += '<p class="ssp-gloss">' + esc(wordData.gloss) + '</p>';
    }
    if (wordData.translation && wordData.translation !== wordData.gloss) {
      html += dl("Trans.", esc(wordData.translation));
    }

    // ── 5. Full MASAQ data per segment ─────────────────────────
    // No "Func." line — the SyntacticRole + CaseMood + Phrase
    // fields appear inline with each segment below, so the data is
    // presented once in full rather than split between a header
    // summary and a per-segment breakdown.
    if (segs.length > 0) {
      html += section("Masaq");
      html += '<div class="ssp-segs">';
      for (var i = 0; i < segs.length; i++) {
        html += renderSegment(segs[i]);
      }
      html += '</div>';
    }

    // ── 6. Full JSONL block (meanings, lexicons, examples) ──────
    if (rootData) {
      html += section("Meaning (from AI source)");

      if (rootData.meaning_en) {
        html += '<p class="ssp-meaning-en">' + esc(rootData.meaning_en) + '</p>';
      }
      if (rootData.meaning_ar) {
        html += '<p class="ssp-meaning-ar" lang="ar" dir="rtl">' +
          esc(rootData.meaning_ar) + '</p>';
      }
      if (rootData.core_semantic) {
        html += '<p class="ssp-meaning-core"><strong>Core semantic field:</strong> ' +
          esc(rootData.core_semantic) + '</p>';
      }
      if (rootData.ibn_faris) {
        html += '<p class="ssp-lex"><strong>Ibn Fāris:</strong> ' +
          esc(rootData.ibn_faris) + '</p>';
      }
      if (rootData.al_raghib) {
        html += '<p class="ssp-lex"><strong>Al-Rāghib:</strong> ' +
          esc(rootData.al_raghib) + '</p>';
      }

      // Quran examples
      var qex = rootData.quran_examples || [];
      if (qex.length > 0) {
        html += '<h4 class="ssp-h4">Quran examples</h4>';
        html += '<ol class="ssp-examples">';
        for (var k = 0; k < qex.length; k++) {
          var ex = qex[k];
          var ref = ex.Ref || "";
          html += '<li class="ssp-example">';
          if (ref) {
            var parts = ref.split(":");
            var href = "/surah/" + parts[0] + "#verse-" + parts[1];
            html += '<a class="ssp-example-ref" href="' + esc(href) + '">' +
              esc(ref) + '</a>';
          }
          if (ex.Arabic) {
            html += '<p class="ssp-example-ar" lang="ar" dir="rtl">' +
              esc(ex.Arabic) + '</p>';
          }
          if (ex.English) {
            html += '<p class="ssp-example-en">' + esc(ex.English) + '</p>';
          }
          if (ex.Context) {
            html += '<p class="ssp-example-ctx">' + esc(ex.Context) + '</p>';
          }
          html += '</li>';
        }
        html += '</ol>';
      }

      // Hadith examples
      var had = rootData.hadith || [];
      if (had.length > 0) {
        html += '<h4 class="ssp-h4">Hadith references</h4>';
        html += '<ol class="ssp-hadith">';
        for (var m = 0; m < had.length; m++) {
          var h = had[m];
          html += '<li class="ssp-hadith-item">';
          if (h.Arabic) {
            html += '<p class="ssp-hadith-ar" lang="ar" dir="rtl">' +
              esc(h.Arabic) + '</p>';
          }
          if (h.English) {
            html += '<p class="ssp-hadith-en">' + esc(h.English) + '</p>';
          }
          if (h.Source) {
            html += '<p class="ssp-hadith-src">' + esc(h.Source) + '</p>';
          }
          html += '</li>';
        }
        html += '</ol>';
      }

      html += '<a class="ssp-link" href="/root/detailed/' +
        esc(rootData.root) + '">Full root detail →</a>';
    } else if (!rootData && !(wordData.root && wordData.root.buckwalter)) {
      html += '<p class="ssp-empty">No trilateral root.</p>';
    }

    content.innerHTML = html;
  }
})();