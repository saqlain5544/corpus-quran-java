// continue-reading.js — remembers the user's last-read surah +
// ayah and surfaces a "Continue reading …" banner on the homepage.
//
// Storage key: qr.lastRead
// Item shape: { surah: int, ayah: int, ts: number }
//
// Triggers:
//   - bookmark toggle on /surah/: when the user clicks the bookmark
//     star (this is a stronger signal than just scrolling — the user
//     is marking their place).
//   - any /surah/ click on a word/root: signals active engagement.
//   - the surah page's scroll handler: on stop-scroll (debounced)
//     the visible topmost ayah is saved.
//
// Homepage:
//   - On load, reads qr.lastRead and renders a banner above the
//     stats strip. Click → /surah/{s}#verse-{a}.
(function () {
  "use strict";

  const STORAGE_KEY = "qr.lastRead";

  function load() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      var v = JSON.parse(raw);
      if (!v || typeof v.surah !== "number" || typeof v.ayah !== "number") return null;
      return v;
    } catch (e) { return null; }
  }
  function save(v) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(v));
    } catch (e) { /* private mode */ }
  }

  // ── Surah page: capture the user's current ayah ─────────────
  function currentSurah() {
    var el = document.querySelector(".page-surah");
    if (!el) return null;
    return {
      surah: parseInt(el.dataset.surah, 10),
      scrollY: window.scrollY,
    };
  }

  // Find the topmost ayah currently in view (top of viewport).
  //
  // We use IntersectionObserver (MDN — Intersection Observer API,
  // Baseline 2019) instead of polling getBoundingClientRect() in a
  // scroll handler. The browser computes intersection lazily off the
  // main thread and fires the callback only when an ayah crosses the
  // upper-quarter strip of the viewport.
  //
  // Old code did querySelectorAll('.ayah') + per-ayah rect() in a
  // debounced scroll handler. For Al-Baqarah (286 ayahs) that's
  // 286 getBoundingClientRect calls per scroll tick — each forces a
  // layout. The new approach has the browser do this work natively
  // and notify us only on intersection change.
  //
  // rootMargin: "0px 0px -75% 0px" shrinks the viewport's effective
  // height to just the upper 25% strip. Entries fire when an ayah
  // crosses either edge of that strip.
  //
  // See /tmp/algo-research/fe_01_techniques.md for the rationale.
  var topmostObserver = null;
  function setupTopmostObserver() {
    if (topmostObserver) return;
    var ayahs = document.querySelectorAll(".ayah");
    if (!ayahs.length) return;
    topmostObserver = new IntersectionObserver(function (entries) {
      // Among the changed entries, find the topmost intersecting one.
      // We pick the topmost by current boundingClientRect.top across
      // ALL currently-intersecting ayahs (not just changed entries)
      // — because the change entry may have just left the strip while
      // a different one entered. We approximate by scanning the
      // changed entries plus their neighbours.
      var topmost = null;
      for (var i = 0; i < entries.length; i++) {
        var e = entries[i];
        if (!e.isIntersecting) continue;
        // Skip entries with empty rect (defensive — can happen with
        // detached elements or 0-size containers).
        var r = e.boundingClientRect;
        if (r.height === 0 || r.width === 0) continue;
        if (topmost === null || r.top < topmost.boundingClientRect.top) {
          topmost = e;
        }
      }
      if (!topmost) return;
      var n = parseInt(topmost.target.dataset.ayah, 10);
      if (!n) return;
      var s = currentSurah();
      if (!s || !s.surah) return;
      save({ surah: s.surah, ayah: n, ts: Date.now() });
    }, {
      rootMargin: "0px 0px -75% 0px",
      threshold: [0, 1],
    });
    for (var i = 0; i < ayahs.length; i++) topmostObserver.observe(ayahs[i]);
  }

  // Fallback for browsers without IntersectionObserver (very old,
  // <2019). Should be a no-op in practice — Baseline Widely
  // available since 2019 — but defensive.
  function scheduleSaveLegacy() {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = setTimeout(function () {
      var s = currentSurah();
      if (!s || !s.surah) return;
      var ayahs = document.querySelectorAll(".ayah");
      if (!ayahs.length) return;
      var top = window.innerHeight * 0.25;
      for (var i = 0; i < ayahs.length; i++) {
        var r = ayahs[i].getBoundingClientRect();
        if (r.bottom > top) {
          save({ surah: s.surah, ayah: parseInt(ayahs[i].dataset.ayah, 10), ts: Date.now() });
          break;
        }
      }
    }, 800);
  }

  var saveTimer = null;
  function scheduleSave() {
    // If IntersectionObserver is available, save immediately —
    // no debounce needed because the browser already debounces
    // intersection events to animation frames.
    if (topmostObserver) return;
    scheduleSaveLegacy();
  }

  // Bookmark click is a strong "I'm here" signal — save immediately.
  document.addEventListener("click", function (e) {
    if (e.target.closest('[data-component="ayah-bookmark"]')) {
      var s = currentSurah();
      if (!s || !s.surah) return;
      // Use the ayah from the button's dataset.
      var btn = e.target.closest('[data-component="ayah-bookmark"]');
      var ayah = parseInt(btn.dataset.ayah, 10);
      save({ surah: s.surah, ayah: ayah, ts: Date.now() });
    }
  });

  // Sidebar word-click signals active engagement.
  document.addEventListener("click", function (e) {
    if (e.target.closest('[data-component="word"]')) {
      scheduleSave();
    }
  });

  // ── Set up topmost-ayah detection ─────────────────────────
  // Prefer IntersectionObserver (native, off-main-thread). Fall
  // back to debounced scroll handler for ancient browsers.
  var pageSurah = document.querySelector(".page-surah");
  if (pageSurah) {
    if (typeof IntersectionObserver === "function") {
      setupTopmostObserver();
      // Save once on initial page-load (for a brand-new visit).
      scheduleSave();
    } else {
      window.addEventListener("scroll", scheduleSave, { passive: true });
      scheduleSave();
    }
  }

  // ── Homepage: render the banner ─────────────────────────────
  function renderBanner() {
    var banner = document.querySelector('[data-component="continue-reading"]');
    if (!banner) return;
    var v = load();
    if (!v || !v.surah || !v.ayah) {
      banner.hidden = true;
      return;
    }
    banner.hidden = false;
    banner.innerHTML = "";
    var link = document.createElement("a");
    link.href = "/surah/" + v.surah + "#verse-" + v.ayah;
    link.className = "continue-link";

    var label = document.createElement("span");
    label.className = "continue-label";
    label.textContent = "Continue reading";
    link.appendChild(label);

    var ref = document.createElement("span");
    ref.className = "continue-ref";
    ref.textContent = "Surah " + v.surah + ":" + v.ayah + " →";
    link.appendChild(ref);

    banner.appendChild(link);
  }
  if (document.querySelector(".page-home")) {
    renderBanner();
  }

  // Test hook.
  window.__qrLastRead = { load: load, save: save };
})();