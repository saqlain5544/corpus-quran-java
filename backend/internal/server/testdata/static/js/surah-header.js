// surah-header.js — local search, font-size, line-height, prev/next.
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

  // ── Update one slider end-to-end ──────────────────────────────
  // - sets the input value
  // - writes the live numeric display to its <output>
  // - recalculates the CSS --fill so the track fill follows the thumb
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
      // Default value — still need to seed the fill on first paint.
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

  // ── Local search ─────────────────────────────────────────────
  if (searchForm) {
    const input = searchForm.querySelector("input");
    searchForm.addEventListener("submit", (e) => {
      e.preventDefault();
      const q = (input.value || "").trim();
      input.classList.remove("is-error");
      if (!q) return;
      // Find the first ayah whose text contains q (case-insensitive).
      const ayahs = surahEl.querySelectorAll(".ayah");
      let firstHit = null;
      for (const ay of ayahs) {
        const text = (ay.textContent || "").replace(/\s+/g, " ");
        if (text.toLowerCase().indexOf(q.toLowerCase()) !== -1) {
          firstHit = ay;
          break;
        }
      }
      if (firstHit) {
        const id = firstHit.id || ("verse-" + firstHit.dataset.ayah);
        firstHit.id = id;
        firstHit.scrollIntoView({ behavior: "smooth", block: "start" });
        firstHit.classList.add("ayah-focus");
        setTimeout(() => firstHit.classList.remove("ayah-focus"), 2500);
        // Reflect in URL for shareability.
        history.replaceState(null, "", "#" + id);
      } else {
        input.classList.add("is-error");
        input.focus();
      }
    });
    // Clear error state as user types.
    if (input) {
      input.addEventListener("input", () => input.classList.remove("is-error"));
    }
  }
})();
