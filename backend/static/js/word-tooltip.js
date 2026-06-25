// word-tooltip.js — event delegation on .ayah-text container, edge-aware
// positioning, renders the morphology tooltip for a hovered word.
//
// Tooltip layout (top → bottom):
//   1. Word            — large Arabic, with tashkeel
//   2. Root            — Arabic letters with a space between each
//   3. Lemma           — bare stem, no diacritics
//   4. Gloss           — combined English gloss
//   5. POS · Function  — small caps at the bottom

(function () {
  "use strict";

  const tooltip = document.getElementById("word-tooltip");
  if (!tooltip) return;

  let hideTimer = null;
  let activeWord = null;
  const cache = new Map();

  function showOn(word) {
    if (hideTimer) { clearTimeout(hideTimer); hideTimer = null; }
    activeWord = word;
    const s = word.dataset.surah, a = word.dataset.ayah, w = word.dataset.word;
    const key = s + ":" + a + ":" + w;

    tooltip.setAttribute("data-state", "loading");
    tooltip.innerHTML = '<p class="wt-loading">Loading…</p>';
    positionTooltip(word);

    if (cache.has(key)) {
      renderTooltip(cache.get(key));
      return;
    }
    word.dataset.loading = "true";
    fetch("/api/word?s=" + encodeURIComponent(s) +
          "&a=" + encodeURIComponent(a) +
          "&w=" + encodeURIComponent(w))
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => {
        cache.set(key, data);
        if (activeWord === word) renderTooltip(data);
      })
      .catch(err => {
        if (activeWord === word) {
          tooltip.innerHTML = '<p class="wt-empty">No data (' +
            escape(String(err)) + ').</p>';
          tooltip.setAttribute("data-state", "visible");
        }
      })
      .finally(() => { word.dataset.loading = "false"; });
  }

  function renderTooltip(data) {
    const wordText = data.word_text || "";
    const lemma = data.lemma || "";
    const gloss = data.gloss || "";
    const fn = data.function || "";
    const rootLetters = (data.root && data.root.letters) || "";
    const segs = data.segments || [];

    // ── 1. Word (Uthmani, large) ───────────────────────────────
    let html = '<p class="wt-word" lang="ar" dir="rtl">' +
      escape(wordText) + '</p>';

    // ── 2. Root (Arabic / Buckwalter) ──────────────────────────
    if (data.root) {
      const r = data.root;
      html += '<p class="wt-root-line">' +
        '<span class="wt-root-ar" lang="ar" dir="rtl">' + escape(r.arabic || "") + '</span>' +
        '<span class="wt-root-bw">' + escape(r.buckwalter || "") + '</span>' +
      '</p>';
    }

    // ── 3. Lemma (bare stem, from MASAQ) ───────────────────────
    if (lemma) {
      html += '<p class="wt-lemma" lang="ar" dir="rtl">' + escape(lemma) + '</p>';
    }

    // ── 4. Gloss ───────────────────────────────────────────────
    if (gloss) {
      html += '<hr class="wt-sep">';
      html += '<p class="wt-gloss">' + escape(gloss) + '</p>';
    }

    // ── 5. Morphology Tags ─────────────────────────────────────
    // Only MorphType + MorphTag (the core POS / segment role).
    // Declension state (InvariableDeclinable, PossessiveConstruct)
    // and grammatical tags (SyntacticRole, CaseMood, etc.) are
    // rendered in their own sections below — including them here
    // would just repeat the same info twice.
    if (segs.length > 0) {
      const hasMorphTags = segs.some(s => s.MorphType || s.MorphTag);
      if (hasMorphTags) {
        html += '<hr class="wt-sep">';
        html += '<div class="wt-tags">';
        for (const s of segs) {
          const bits = [];
          if (s.MorphType) bits.push(s.MorphType);
          if (s.MorphTag) bits.push(s.MorphTag);
          if (bits.length) {
            html += '<span class="wt-tag">' + escape(bits.join(' · ')) + '</span>';
          }
        }
        html += '</div>';
      }
    }

    // ── 6. Phrase + Function ────────────────────────────────────
    // PHRASE / PRED / SUBJ / etc. — only render when the segment
    // is part of a phrase. Empty phrase + null phrasal_function
    // usually means the segment stands alone.
    const phraseBits = [];
    for (const s of segs) {
      if (s.Phrase && s.Phrase.trim()) phraseBits.push(s.Phrase);
      if (s.PhrasalFunction && s.PhrasalFunction.trim()) {
        phraseBits.push(s.PhrasalFunction);
      }
    }
    if (phraseBits.length) {
      html += '<hr class="wt-sep">';
      html += '<dl class="wt-fields">';
      html += '<dt>Phrase</dt><dd>' + escape(phraseBits.join(' / ')) + '</dd>';
      html += '</dl>';
    }

    // ── 7. Declension state ─────────────────────────────────────
    // INVAR / DECLN / DEF_ART + construct state. These are
    // technically grammar tags but they affect how the segment
    // is parsed downstream (e.g. a construct noun has genitive
    // case forced).
    const declBits = [];
    for (const s of segs) {
      if (s.InvariableDeclinable) declBits.push(s.InvariableDeclinable);
      if (s.PossessiveConstruct && s.PossessiveConstruct !== "NOT_CONSTRUCT") {
        declBits.push(s.PossessiveConstruct);
      }
    }
    if (declBits.length) {
      html += '<dl class="wt-fields">';
      html += '<dt>Decl.</dt><dd>' + escape(declBits.join(' / ')) + '</dd>';
      html += '</dl>';
    }

    // ── 8. Grammatical Tags ────────────────────────────────────
    // Dedupe by tracking the last-emitted value, since the same
    // tag can appear across multiple segments of the same word
    // (e.g. two SUFFIX segments both tagged with NSUFF_*).
    const gramBits = [];
    let lastEmitted = "";
    function emit(tag) {
      if (!tag) return;
      if (tag === lastEmitted) return;
      gramBits.push(tag);
      lastEmitted = tag;
    }
    for (const s of segs) {
      emit(s.SyntacticRole);
      emit(s.CaseMood);
      // The case mood marker is the actual vowel (KASRA/DHAMMA/FATHA)
      // attached to the word's last letter. Surface it next to the
      // abstract CaseMood tag so the user can see what sound the
      // case produces in recitation.
      emit(s.CaseMoodMarker);
    }
    if (gramBits.length || fn) {
      html += '<hr class="wt-sep">';
      html += '<p class="wt-gram">';
      if (fn) html += escape(fn);
      if (gramBits.length) {
        html += ' <span class="wt-gram-tags">' + escape(gramBits.join(', ')) + '</span>';
      }
      html += '</p>';
    }

    if (!rootLetters && !lemma && !gloss && segs.length === 0) {
      html += '<p class="wt-empty">No morphological data.</p>';
    }

    tooltip.innerHTML = html;
    tooltip.setAttribute("data-state", "visible");
    positionTooltip(activeWord);
  }

  function positionTooltip(word) {
    if (!word) return;
    tooltip.removeAttribute("hidden");
    tooltip.setAttribute("data-state", "visible");
    // Force layout so we can measure.
    tooltip.style.left = "0px";
    tooltip.style.top = "0px";
    const wRect = word.getBoundingClientRect();
    const tRect = tooltip.getBoundingClientRect();
    const vw = window.innerWidth;
    const vh = window.innerHeight;
    const margin = 8;

    // Vertical: prefer above the word; flip below if no room.
    let top = wRect.top - tRect.height - margin;
    if (top < margin) top = wRect.bottom + margin;
    if (top + tRect.height > vh - margin) top = vh - tRect.height - margin;

    // Horizontal: center on word; clamp to viewport.
    let left = wRect.left + wRect.width / 2 - tRect.width / 2;
    if (left < margin) left = margin;
    if (left + tRect.width > vw - margin) left = vw - tRect.width - margin;

    tooltip.style.left = left + "px";
    tooltip.style.top = top + "px";
  }

  function hideTooltip() {
    activeWord = null;
    hideTimer = setTimeout(() => {
      tooltip.setAttribute("hidden", "");
      tooltip.removeAttribute("data-state");
    }, 120);
  }

  // Event delegation on the document (covers all surah pages).
  document.addEventListener("mouseover", e => {
    const w = e.target.closest('[data-component="word"]');
    if (w) showOn(w);
  });
  document.addEventListener("mouseout", e => {
    const w = e.target.closest('[data-component="word"]');
    if (w) hideTooltip();
  });
  document.addEventListener("focusin", e => {
    const w = e.target.closest('[data-component="word"]');
    if (w) showOn(w);
  });
  document.addEventListener("focusout", e => {
    const w = e.target.closest('[data-component="word"]');
    if (w) hideTooltip();
  });

  // Hide on scroll (tooltips should not lag behind).
  window.addEventListener("scroll", hideTooltip, { passive: true });

  // Escape hides the tooltip.
  document.addEventListener("keydown", e => {
    if (e.key === "Escape") hideTooltip();
  });

  function escape(s) {
    if (s == null) return "";
    return String(s).replace(/[&<>"']/g, c =>
      ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }
})();
