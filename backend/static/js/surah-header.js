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
    QR.dom.focusHighlight(ay, 2500);
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

  // ── ArrowLeft / ArrowRight — prev/next surah navigation ────────
  // The .sh-prev / .sh-next anchors exist in the markup; we just
  // bind keyboard shortcuts so a reader deep in the Quran can
  // navigate between surahs without scrolling back to the header.
  // Only fires when no input/select/contenteditable is focused.
  const prevLink = surahEl.querySelector(".sh-prev[href]:not(.disabled)");
  const nextLink = surahEl.querySelector(".sh-next[href]:not(.disabled)");
  if (prevLink || nextLink) {
    document.addEventListener("keydown", function (e) {
      if (e.altKey || e.ctrlKey || e.metaKey) return;
      const a = document.activeElement;
      if (a && (a.tagName === "INPUT" || a.tagName === "TEXTAREA"
                || a.tagName === "SELECT" || a.isContentEditable)) return;
      if (e.key === "ArrowLeft" && prevLink) {
        e.preventDefault();
        prevLink.click();
      } else if (e.key === "ArrowRight" && nextLink) {
        e.preventDefault();
        nextLink.click();
      }
    });
  }

  // ── Restore persisted values ─────────────────────────────────
  var fs = QR.storage.get("qr.fs");
  var lh = QR.storage.get("qr.lh");
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

  // ── Slider persistence ──────────────────────────────────────
  var debouncedSliderPersist = QR.dom.debounce(function(key, value) {
    QR.storage.set(key, value);
  }, 50);

  // ── Font-size ────────────────────────────────────────────────
  if (fontInput && fontOutput) {
    fontInput.addEventListener("input", () => {
      const v = fontInput.value;
      surahEl.style.setProperty("--fs-quran", v + "px");
      syncSlider(fontInput, fontOutput, v, "px");
      debouncedSliderPersist("qr.fs", v);
    });
  }

  // ── Line-height ──────────────────────────────────────────────
  if (lineInput && lineOutput) {
    lineInput.addEventListener("input", () => {
      const v = lineInput.value;
      surahEl.style.setProperty("--lh-quran", v);
      syncSlider(lineInput, lineOutput, v, "");
      debouncedSliderPersist("qr.lh", v);
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
      QR.storage.set("qr.transIdx", String(idx));
    });
    // Restore persisted value
    var saved = QR.storage.get("qr.transIdx");
    if (saved !== "") {
      transSelect.value = saved;
      transSelect.dispatchEvent(new Event("change"));
    }
  }

  // ── Local search with result navigation ──────────────────────
  //
  // Hot-path optimization: each ayah's normalized search text is
  // pre-computed once and stashed in `data-search-text` on the
  // element. This avoids re-reading textContent (which traverses
  // every child node including verse numbers, sajda markers, etc.)
  // on every submit. For Al-Baqarah that's 286 reads replaced by
  // 286 attribute reads — about 3-5× faster.
  //
  // The script that builds the surah page (templates/surah.tmpl)
  // is responsible for emitting `data-search-text` on each .ayah.
  // We back-fill it here for any ayah that's missing the
  // attribute (defensive — older cached pages or template bugs).
  if (searchForm) {
    const input = searchForm.querySelector("input");
    const ayahs = surahEl.querySelectorAll(".ayah");
    // One-time cache fill. Uses textContent which is the live
    // rendered text (includes any Arabic transliteration already
    // resolved).
    for (const ay of ayahs) {
      if (!ay.dataset.searchText) {
        ay.dataset.searchText = (ay.textContent || "")
          .replace(/\s+/g, " ").toLowerCase();
      }
    }
    searchForm.addEventListener("submit", (e) => {
      e.preventDefault();
      const q = (input.value || "").trim();
      input.classList.remove("is-error");
      if (!q) {
        clearSearch();
        return;
      }
      // Find all matching ayahs.
      const needle = q.toLowerCase();
      srMatches = [];
      for (const ay of ayahs) {
        if (ay.dataset.searchText.indexOf(needle) !== -1) {
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
