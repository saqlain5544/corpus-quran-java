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
      QR.storage.set("qr.theme", next);
    });
  }

  // ── Restore font size / line height on surah pages ────────────
  // surah-header.js owns the canonical restore logic (it also
  // re-syncs the slider output and the --fill gradient). We only
  // mirror the CSS custom properties here for the rare case where
  // the user lands on a surah page before that script runs (e.g.,
  // a flash of unstyled font before defer scripts execute).
  const surah = document.querySelector('.page-surah');
  if (surah) {
    var fs = QR.storage.get("qr.fs");
    var lh = QR.storage.get("qr.lh");
      if (fs) surah.style.setProperty('--fs-quran', fs + 'px');
      if (lh) surah.style.setProperty('--lh-quran', lh);


    // ── Fixed left-sidebar positioning ─────────────────────────
    // The surah-side-panel is position:fixed on desktop, so it is
    // taken out of the grid flow. We measure the grid columns of
    // .page-surah and write the sidebar's left/width as CSS
    // variables so the panel visually aligns with col 2 of the
    // 2fr 30fr 68fr grid (gap / sidebar / main).
    const panel = document.getElementById('surah-side-panel');
    if (panel) {
      const positionPanel = () => {
        if (window.innerWidth < 960) {
          // Mobile: let CSS handle the fixed overlay.
          surah.style.removeProperty('--ssp-left');
          surah.style.removeProperty('--ssp-width');
          return;
        }
        const rect = surah.getBoundingClientRect();
        const styles = getComputedStyle(surah);
        const cols = styles.gridTemplateColumns.split(' ');
        const gap = parseFloat(styles.columnGap) || 0;
        if (cols.length < 3) return;
        const col1 = parseFloat(cols[0]);
        const col2 = parseFloat(cols[1]);
        const left = rect.left + col1 + gap;
        surah.style.setProperty('--ssp-left', left + 'px');
        surah.style.setProperty('--ssp-width', col2 + 'px');
      };
      positionPanel();
      // Debounce resize — fires 60+ times during a window drag.
      window.addEventListener('resize', QR.dom.debounce(positionPanel, 100));
      // Reflow after the hafs.woff2 font swap (it can change line
      // heights and therefore column widths).
      if (document.fonts && document.fonts.ready) {
        document.fonts.ready.then(positionPanel).catch(() => {});
      }
    }
  }

  // ── Global search form already submits as GET /search — no JS needed.
  // Hook here only for future enhancements (autocomplete dropdown).

  // ── Auto-submit filter forms (replaces inline onchange handlers) ──
  // Any <form data-component="auto-submit-filters"> submits when a
  // <select> inside it changes. The form ALSO has a visible "Apply"
  // button for keyboard / screen-reader users who don't trigger
  // change events.
  document.addEventListener("change", (e) => {
    const t = e.target;
    if (!t || t.tagName !== "SELECT") return;
    const form = t.closest('form[data-component="auto-submit-filters"]');
    if (form) form.submit();
  });

  // Note: verse-scroll.js handles the scroll-to-verse behavior on
  // /surah/{id} URLs with a #verse-N hash. It adds the .ayah-focus
  // class client-side before smooth-scrolling, then strips it after
  // a 2.5s delay so the highlight fades out.
})();
