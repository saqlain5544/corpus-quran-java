// surah-side-panel.js — fixed left panel showing the full analysis
// for the clicked word. Per plan.md §Surah-Page Column1: "don't
// show concordance data here just all the masaq, and root
// meanings data. don't use multiple columns, use stacked data."
//
// The panel renders a single stacked column with this structure:
//
//   ┌──────────────────────────┐
//   │ Word  9:5:18            │  ← word + ref
//   │ Root   (arabic/buckwalter)│
//   │ Lemma  (from masaq)      │
//   │ Gloss  english gloss     │
//   │ ─ Masaq Tags ─           │
//   │ seg 1: prefix/stem/suffix│  ← full MASAQ data, no repetition
//   │ seg 2: ...                │
//   │ ─ Lemma frequency ─      │  ← first block
//   │ lemma × 74  (current)     │
//   │ lemma × 5                │
//   │ ─ Full Meaning (JSONL) ─  │  ← second block
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

  var cache = {};
  var currentLemma = "";

  document.addEventListener("click", function (e) {
    var w = e.target.closest('[data-component="word"]');
    if (!w) return;
    var root = w.dataset.root;
    var s = w.dataset.surah, a = w.dataset.ayah, wn = w.dataset.word;
    var key = s + ":" + a + ":" + wn;

    if (cache[key]) {
      content.innerHTML = cache[key];
      panel.classList.add("is-open");
      return;
    }

    content.innerHTML = '<p class="ssp-loading">Loading…</p>';
    panel.classList.add("is-open");

    var wordData = null;

    fetch("/api/word?s=" + encodeURIComponent(s) +
          "&a=" + encodeURIComponent(a) +
          "&w=" + encodeURIComponent(wn))
      .then(function (r) { return r.ok ? r.json() : Promise.reject(r.status); })
      .then(function (wd) {
        wordData = wd;
        currentLemma = wd.lemma || "";
        if (root) {
          return fetch("/api/root/" + encodeURIComponent(root) + "/summary")
            .then(function (r) { return r.ok ? r.json() : null; });
        }
        return null;
      })
      .then(function (rootData) {
        renderPanel(wordData, rootData);
        cache[key] = content.innerHTML;
      })
      .catch(function () { content.innerHTML = '<p class="ssp-empty">Failed to load.</p>'; });
  });

  // ── Render helpers ────────────────────────────────────────────
  function esc(s) {
    if (s == null) return "";
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  // Render a labeled definition-list row: <dt>label</dt><dd>value</dd>.
  // `valueHtml` is inserted as raw HTML (caller's responsibility to
  // escape). `attrs` is an optional object of extra <dd> attributes.
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

    // ── 6. Lemma frequency block (first) ───────────────────────
    // Per the user's request: "I want lemma frequency first as a
    // block then full jsonl block after that". So this section
    // comes BEFORE the full meaning data.
    if (rootData && rootData.lemmas && rootData.lemmas.length > 0) {
      html += section("Lemma frequency");
      // Sort: current lemma first, then by frequency desc.
      var lemmas = rootData.lemmas.slice();
      lemmas.sort(function (a, b) {
        if (currentLemma && a.arabic === currentLemma) return -1;
        if (currentLemma && b.arabic === currentLemma) return 1;
        return b.occurrences - a.occurrences;
      });
      html += '<ol class="ssp-freq-list">';
      for (var j = 0; j < lemmas.length; j++) {
        var lm = lemmas[j];
        var isCur = currentLemma && lm.arabic === currentLemma;
        var cls = "ssp-freq-row" + (isCur ? " is-current" : "");
        html += '<li class="' + cls + '">';
        html += '<span class="ssp-freq-arabic" lang="ar" dir="rtl">' +
          esc(lm.arabic) + '</span>';
        html += '<span class="ssp-freq-count">' + lm.occurrences + '×</span>';
        html += '</li>';
      }
      html += '</ol>';
    }

    // ── 7. Full JSONL block (meanings, lexicons, examples) ──────
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