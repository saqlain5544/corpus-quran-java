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
  function topmostAyah() {
    var ayahs = document.querySelectorAll(".ayah");
    if (!ayahs.length) return null;
    var top = window.innerHeight * 0.25; // focus on upper quarter
    for (var i = 0; i < ayahs.length; i++) {
      var r = ayahs[i].getBoundingClientRect();
      if (r.bottom > top) return parseInt(ayahs[i].dataset.ayah, 10);
    }
    return null;
  }

  var saveTimer = null;
  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer);
    saveTimer = setTimeout(function () {
      var s = currentSurah();
      if (!s || !s.surah) return;
      var ayah = topmostAyah();
      if (!ayah) return;
      save({ surah: s.surah, ayah: ayah, ts: Date.now() });
    }, 800);
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

  // Scroll → debounced save of the topmost ayah.
  var pageSurah = document.querySelector(".page-surah");
  if (pageSurah) {
    window.addEventListener("scroll", scheduleSave, { passive: true });
    // Also save on initial page-load (so a brand-new visit is recorded).
    scheduleSave();
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