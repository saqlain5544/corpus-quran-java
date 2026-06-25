// app.js — top-level: theme persistence, global search submit, layout glue.

(function () {
  "use strict";

  // ── Theme toggle ──────────────────────────────────────────────
  const themeBtn = document.querySelector('[data-component="theme-toggle"]');
  if (themeBtn) {
    themeBtn.addEventListener("click", () => {
      const cur = document.documentElement.getAttribute("data-theme") || "light";
      const next = cur === "dark" ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", next);
      try { localStorage.setItem("qr.theme", next); } catch (e) {}
    });
  }

  // ── Restore font size / line height on surah pages ────────────
  if (document.body.matches('.page-surah, .page-surah *') || document.querySelector('.page-surah')) {
    try {
      const fs = localStorage.getItem("qr.fs");
      const lh = localStorage.getItem("qr.lh");
      const surah = document.querySelector('.page-surah');
      if (surah) {
        if (fs) surah.style.setProperty('--fs-quran', fs + 'px');
        if (lh) surah.style.setProperty('--lh-quran', lh);
      }
    } catch (e) {}
  }

  // ── Global search form already submits as GET /search — no JS needed.
  // Hook here only for future enhancements (autocomplete dropdown).

  // Note: verse-scroll.js handles the scroll-to-verse behavior on
  // /surah/{id}/{verse} URLs (it looks for either a URL hash or the
  // .ayah-focus class the server applies).
})();
