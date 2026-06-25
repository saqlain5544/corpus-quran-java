// surah-header.js — local search, font-size, line-height, prev/next,
// and bottom search-result navigation with prev/next scrolling.
//
// All behavior is event-delegated and idempotent. State is persisted
// to localStorage as `qr.fs` (font-size in px) and `qr.lh` (line-height
// multiplier).

(function () {
  "use strict";

  const surahEl = document.querySelector(".page-surah");
  if (!surahEl) return;

  const searchForm = surahEl.querySelector('[data-component="surah-local-search"]');
  const fontInput = surahEl.querySelector('[data-component="font-size"]');
  const lineInput = surahEl.querySelector('[data-component="line-height"]');
  const fontOutput = document.getElementById("font-size-value");
  const lineOutput = document.getElementById("line-height-value");

  // ── Search result navigation state ─────────────────────────────
  let srIdx = -1;
  let srMatches = [];

  // Build / remove the bottom navigation footer.
  function ensureFooter() {
    let ft = document.getElementById("surah-search-footer");
    if (!ft) {
      ft = document.createElement("footer");
      ft.id = "surah-search-footer";
      ft.className = "surah-search-footer";
      ft.setAttribute("role", "status");
      ft.innerHTML =
        '<span class="ssf-info"></span>' +
        '<button type="button" class="ssf-prev" aria-label="Previous result">←</button>' +
        '<button type="button" class="ssf-next" aria-label="Next result">→</button>' +
        '<button type="button" class="ssf-close" aria-label="Close results">✕</button>';
      surahEl.appendChild(ft);
      ft.querySelector(".ssf-prev").addEventListener("click", goPrev);
      ft.querySelector(".ssf-next").addEventListener("click", goNext);
      ft.querySelector(".ssf-close").addEventListener("click", clearSearch);
    }
    return ft;
  }

  function updateFooter() {
    const ft = ensureFooter();
    if (srMatches.length === 0) {
      ft.style.display = "none";
      return;
    }
    ft.style.display = "";
    ft.querySelector(".ssf-info").textContent =
      (srIdx + 1) + "/" + srMatches.length;
  }

  function goPrev() {
    if (srMatches.length === 0) return;
    srIdx = srIdx > 0 ? srIdx - 1 : srMatches.length - 1;
    scrollToMatch();
  }

  function goNext() {
    if (srMatches.length === 0) return;
    srIdx = srIdx < srMatches.length - 1 ? srIdx + 1 : 0;
    scrollToMatch();
  }

  function scrollToMatch() {
    const ay = srMatches[srIdx];
    if (!ay) return;
    ay.scrollIntoView({ behavior: "smooth", block: "start" });
    ay.classList.add("ayah-focus");
    setTimeout(() => ay.classList.remove("ayah-focus"), 2500);
    updateFooter();
  }

  function clearSearch() {
    srMatches = [];
    srIdx = -1;
    updateFooter();
    const inp = searchForm && searchForm.querySelector("input");
    if (inp) {
      inp.value = "";
      inp.classList.remove("is-error");
    }
  }

  // ── Update one slider end-to-end ──────────────────────────────
  function syncSlider(input, output, value, unit) {
    input.value = value;
    const display = unit ? value + " " + unit : value;
    output.textContent = display;
    const min = parseFloat(input.min);
    const max = parseFloat(input.max);
    const pct = max > min
      ? ((parseFloat(value) - min) / (max - min)) * 100
      : 0;
    input.style.setProperty("--fill", pct.toFixed(1) + "%");
  }

  // ── Restore persisted values ─────────────────────────────────
  try {
    const fs = localStorage.getItem("qr.fs");
    const lh = localStorage.getItem("qr.lh");
    if (fs && fontInput && fontOutput) {
      surahEl.style.setProperty("--fs-quran", fs + "px");
      syncSlider(fontInput, fontOutput, fs, "px");
    } else if (fontInput && fontOutput) {
      syncSlider(fontInput, fontOutput, fontInput.value, "px");
    }
    if (lh && lineInput && lineOutput) {
      surahEl.style.setProperty("--lh-quran", lh);
      syncSlider(lineInput, lineOutput, lh, "");
    } else if (lineInput && lineOutput) {
      syncSlider(lineInput, lineOutput, lineInput.value, "");
    }
  } catch (e) {}

  // ── Font-size ────────────────────────────────────────────────
  if (fontInput && fontOutput) {
    fontInput.addEventListener("input", () => {
      const v = fontInput.value;
      surahEl.style.setProperty("--fs-quran", v + "px");
      syncSlider(fontInput, fontOutput, v, "px");
      try { localStorage.setItem("qr.fs", v); } catch (e) {}
    });
  }

  // ── Line-height ──────────────────────────────────────────────
  if (lineInput && lineOutput) {
    lineInput.addEventListener("input", () => {
      const v = lineInput.value;
      surahEl.style.setProperty("--lh-quran", v);
      syncSlider(lineInput, lineOutput, v, "");
      try { localStorage.setItem("qr.lh", v); } catch (e) {}
    });
  }

  // ── Translation selector ────────────────────────────────────
  const transSelect = surahEl.querySelector('[data-component="translation-select"]');
  if (transSelect) {
    transSelect.addEventListener("change", () => {
      const idx = transSelect.value;
      const allTrans = surahEl.querySelectorAll(".ayah-translation");
      if (idx === "-1") {
        // Hide all
        allTrans.forEach(el => el.hidden = true);
      } else {
        allTrans.forEach(el => {
          el.hidden = el.dataset.transIdx !== idx;
        });
      }
      try { localStorage.setItem("qr.transIdx", idx); } catch (e) {}
    });
    // Restore persisted value
    try {
      const saved = localStorage.getItem("qr.transIdx");
      if (saved !== null) {
        transSelect.value = saved;
        transSelect.dispatchEvent(new Event("change"));
      }
    } catch (e) {}
  }

  // ── Local search with result navigation ──────────────────────
  if (searchForm) {
    const input = searchForm.querySelector("input");
    searchForm.addEventListener("submit", (e) => {
      e.preventDefault();
      const q = (input.value || "").trim();
      input.classList.remove("is-error");
      if (!q) {
        clearSearch();
        return;
      }
      // Find all matching ayahs.
      const ayahs = surahEl.querySelectorAll(".ayah");
      srMatches = [];
      for (const ay of ayahs) {
        const text = (ay.textContent || "").replace(/\s+/g, " ");
        if (text.toLowerCase().indexOf(q.toLowerCase()) !== -1) {
          srMatches.push(ay);
        }
      }
      if (srMatches.length === 0) {
        // Flag the input as having an error so the user can see
        // their query, but DO NOT clear it — wiping the box leaves
        // the user with no way to see what didn't match or fix it.
        input.classList.add("is-error");
        input.focus();
        return;
      }
      srIdx = 0;
      scrollToMatch();
    });
    if (input) {
      input.addEventListener("input", () => input.classList.remove("is-error"));
    }
  }
})();
