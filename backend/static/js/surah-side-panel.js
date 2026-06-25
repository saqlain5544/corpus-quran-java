// surah-side-panel.js — fixed left panel with 3-col sub-grid
// showing frequency listing, concordance, and shades of meaning
// for the clicked word's root. Morphology shown above the grid.
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
      .then(function (rootSummary) {
        renderPanel(wordData, rootSummary);
        cache[key] = content.innerHTML;
      })
      .catch(function () { content.innerHTML = '<p class="ssp-empty">Failed to load.</p>'; });
  });

  function renderPanel(wordData, rootData) {
    var html = '';
    var segs = wordData.segments || [];

    // ── Word + morphology (above the grid) ──────────────────────
    html += '<div class="ssp-word">';
    html += '<span class="ssp-word-text" lang="ar" dir="rtl">' +
      esc(wordData.word_text) + '</span>';
    html += '<span class="ssp-word-ref">' +
      esc(wordData.surah + ":" + wordData.ayah + ":" + wordData.word) + '</span>';
    html += '</div>';

    if (wordData.lemma) {
      html += '<dl class="ssp-dl"><dt>Lemma</dt><dd lang="ar" dir="rtl">' +
        esc(wordData.lemma) + '</dd></dl>';
    }
    if (wordData.gloss) {
      html += '<p class="ssp-gloss">' + esc(wordData.gloss) + '</p>';
    }
    if (wordData.translation) {
      html += '<dl class="ssp-dl"><dt>Trans.</dt><dd>' +
        esc(wordData.translation) + '</dd></dl>';
    }
    if (wordData.function) {
      html += '<dl class="ssp-dl"><dt>Func.</dt><dd>' +
        esc(wordData.function) + '</dd></dl>';
    }
    if (segs.length > 0) {
      html += '<div class="ssp-segs">';
      for (var i = 0; i < segs.length; i++) {
        var s = segs[i];
        html += '<div class="ssp-seg">';
        html += '<span class="ssp-seg-text" lang="ar" dir="rtl">' +
          esc(s.SegmentedWord || s.Word) + '</span> ';
        var bits = [];
        if (s.MorphType) bits.push(s.MorphType);
        if (s.MorphTag) bits.push(s.MorphTag);
        if (s.SyntacticRole) bits.push(s.SyntacticRole);
        if (s.CaseMood) bits.push(s.CaseMood);
        html += '<span class="ssp-seg-tags">' + esc(bits.join(' · ')) + '</span>';
        html += '</div>';
      }
      html += '</div>';
    }

    // ── Root 3-col sub-grid ────────────────────────────────────
    if (rootData) {
      html += '<div class="ssp-root-grid">';

      // Sub-col 1: Frequency listing (lemmas sorted, current on top)
      html += '<div class="ssp-col">';
      html += '<div class="ssp-col-head">Frequency</div>';
      var lemmas = rootData.lemmas || [];
      // Sort: current lemma first, then by frequency desc
      lemmas.sort(function (a, b) {
        if (currentLemma && a.arabic === currentLemma) return -1;
        if (currentLemma && b.arabic === currentLemma) return 1;
        return b.occurrences - a.occurrences;
      });
      for (var i = 0; i < lemmas.length; i++) {
        var lm = lemmas[i];
        var cls = (currentLemma && lm.arabic === currentLemma)
          ? ' ssp-freq-item is-current' : 'ssp-freq-item';
        html += '<div class="' + cls + '">';
        html += '<span class="ssp-freq-arabic" lang="ar" dir="rtl">' +
          esc(lm.arabic) + '</span>';
        html += '<span class="ssp-freq-count">' + lm.occurrences + '×</span>';
        html += '</div>';
      }
      html += '</div>';

      // Sub-col 2: Concordance (occurrences in this surah)
      // Dedupe by surah:ayah — a verse can have the root at multiple
      // word positions (e.g., 2:87 has Aty twice), so the raw list
      // contains the same verse multiple times. Also truncate the
      // verse text — full verses are too long for a sidebar column.
      html += '<div class="ssp-col">';
      html += '<div class="ssp-col-head">Concordance</div>';
      var occs = rootData.occ_list || [];
      var surahNum = wordData.surah;
      var seen = {};
      var shown = 0;
      for (var j = 0; j < occs.length && shown < 10; j++) {
        var o = occs[j];
        if (o.surah !== surahNum) continue;
        var vKey = o.surah + ':' + o.ayah;
        if (seen[vKey]) continue;
        seen[vKey] = true;
        html += '<div class="ssp-conc-item">';
        html += '<a class="ssp-conc-link" href="' + esc(o.link) + '">' +
          esc(o.title) + '</a>';
        if (o.verse_text) {
          html += '<div class="ssp-conc-verse" lang="ar" dir="rtl">' +
            esc(trunc(o.verse_text, 120)) + '</div>';
        }
        html += '</div>';
        shown++;
      }
      if (shown === 0) {
        html += '<p class="ssp-empty">No occurrences in this surah.</p>';
      }
      html += '</div>';

      // Sub-col 3: Shades of meaning
      html += '<div class="ssp-col">';
      html += '<div class="ssp-col-head">Shades</div>';
      if (rootData.meaning_en) {
        html += '<div class="ssp-shade-item">';
        html += '<p class="ssp-shade-body">' +
          esc(trunc(rootData.meaning_en, 300)) + '</p>';
        html += '</div>';
      }
      if (rootData.meaning_ar) {
        html += '<div class="ssp-shade-item">';
        html += '<p class="ssp-shade-body" lang="ar" dir="rtl">' +
          esc(trunc(rootData.meaning_ar, 200)) + '</p>';
        html += '</div>';
      }
      if (rootData.core_semantic) {
        html += '<div class="ssp-shade-item">';
        html += '<span class="ssp-shade-head">Core:</span> ';
        html += '<span class="ssp-shade-body">' +
          esc(trunc(rootData.core_semantic, 200)) + '</span>';
        html += '</div>';
      }
      html += '<a class="ssp-link" href="/root/detailed/' +
        esc(rootData.root) + '">Full root detail →</a>';
      html += '</div>';

      html += '</div>'; // .ssp-root-grid
    } else if (!rootData && !(wordData.root && wordData.root.buckwalter)) {
      html += '<p class="ssp-empty">No trilateral root.</p>';
    }

    content.innerHTML = html;
  }

  function trunc(s, n) { return s && s.length > n ? s.slice(0, n) + '…' : s || ''; }
  function esc(s) {
    if (s == null) return "";
    return String(s).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }
})();
