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
    const pos = data.pos || (data.root && data.root.pos) || "";
    const fn = data.function || "";
    const rootLetters = (data.root && data.root.letters) || "";

    // ── 1. Word (top, large, Arabic with tashkeel) ─────────────
    let html = '<p class="wt-word" lang="ar" dir="rtl">' +
      escape(wordText) + '</p>';

    // ── 2. Root (Arabic letters with spaces between) ───────────
    if (rootLetters) {
      html += '<dl class="wt-fields">' +
        '<dt>Root</dt>' +
        '<dd class="wt-root" lang="ar" dir="rtl">' + escape(rootLetters) + '</dd>';
    }

    // ── 3. Lemma (bare stem) ───────────────────────────────────
    if (lemma) {
      html += (rootLetters ? '' : '<dl class="wt-fields">') +
        '<dt>Lemma</dt>' +
        '<dd class="wt-lemma" lang="ar" dir="rtl">' + escape(lemma) + '</dd>';
    }

    // ── 4. Gloss (English) ─────────────────────────────────────
    if (gloss) {
      html += '<dt>Gloss</dt>' +
        '<dd class="wt-gloss">' + escape(gloss) + '</dd>';
    }

    if (rootLetters || lemma || gloss) {
      html += '</dl>';
    }

    // ── 5. POS · Function (small caps, bottom) ─────────────────
    const metaBits = [];
    if (pos) metaBits.push('<span class="wt-pos">' + escape(pos) + '</span>');
    if (fn) metaBits.push('<span class="wt-fn">' + escape(fn) + '</span>');
    if (metaBits.length) {
      html += '<div class="wt-meta">' +
        metaBits.join('<span class="wt-sep" aria-hidden="true">·</span>') +
      '</div>';
    }

    // Empty-data fallback.
    if (!rootLetters && !lemma && !gloss && !pos && !fn) {
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
