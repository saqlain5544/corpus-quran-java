// word-modal.js — event delegation for word modal (click), focus trap, ESC/backdrop close.

(function () {
  "use strict";

  const modal = document.getElementById("word-modal");
  if (!modal) return;

  const body = modal.querySelector('[data-component="modal-body"]');
  let lastFocused = null;

  function openFor(word) {
    lastFocused = word;
    const s = word.dataset.surah, a = word.dataset.ayah, w = word.dataset.word;

    body.innerHTML = '<p class="modal-loading">Loading…</p>';
    lockBodyScroll();
    if (typeof modal.showModal === "function") {
      modal.showModal();
    } else {
      modal.setAttribute("open", "");
    }
    trapFocus();

    fetch("/api/word?s=" + encodeURIComponent(s) +
          "&a=" + encodeURIComponent(a) +
          "&w=" + encodeURIComponent(w))
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => renderBody(data))
      .catch(err => {
        body.innerHTML = '<p class="modal-no-root">Could not load data (' +
          escape(String(err)) + ').</p>';
      });
  }

  function renderBody(data) {
    const wordText = data.word_text || "";
    const segs = data.segments || [];
    const lemma = data.lemma || "";
    const gloss = data.gloss || "";
    const pos = data.pos || (data.root && data.root.pos) || "";
    const fn = data.function || "";
    const rootLetters = (data.root && data.root.letters) || "";

    // ── 1. Word (large, centered) ─────────────────────────────
    let html = '<p class="modal-word-text" lang="ar" dir="rtl">' +
      escape(wordText) + '</p>';

    // ── 2–4. Root / Lemma / Gloss list ─────────────────────────
    const fields = [];
    if (rootLetters) fields.push(["Root", '<span class="m-root" lang="ar" dir="rtl">' + escape(rootLetters) + '</span>']);
    if (lemma) fields.push(["Lemma", '<span class="m-lemma" lang="ar" dir="rtl">' + escape(lemma) + '</span>']);
    if (gloss) fields.push(["Gloss", '<span class="m-gloss">' + escape(gloss) + '</span>']);

    if (fields.length) {
      html += '<dl class="m-fields">';
      for (const [label, val] of fields) {
        html += '<dt>' + escape(label) + '</dt><dd>' + val + '</dd>';
      }
      html += '</dl>';
    }

    // ── 5. POS · Function (small caps, bottom) ─────────────────
    const metaBits = [];
    if (pos) metaBits.push('<span class="m-pos">' + escape(pos) + '</span>');
    if (fn) metaBits.push('<span class="m-fn">' + escape(fn) + '</span>');
    if (metaBits.length) {
      html += '<div class="m-meta">' +
        metaBits.join('<span class="m-sep" aria-hidden="true">·</span>') +
      '</div>';
    }

    // ── Segments table removed — the gloss column duplicates the
    // word-level gloss above. Morphological detail belongs to the
    // root page (/root/{root}) which lists all occurrences. ──────

    // ── Root deep-dive (meanings, all occurrences) ─────────────
    if (data.root) {
      const r = data.root;
      html += '<div class="modal-root">' +
        '<div class="m-root-line">' +
          '<span class="mr-arabic" lang="ar" dir="rtl">' + escape(r.arabic || "") + '</span>' +
          '<span class="mr-buckwalter">' + escape(r.buckwalter || "") + '</span>' +
          (r.pos ? '<span class="mr-pos">' + escape(r.pos) + '</span>' : '') +
        '</div>' +
        (r.meaning_en ? '<p class="mr-meaning">' + escape(truncate(r.meaning_en, 320)) + '</p>' : '') +
        '<a class="mr-link" href="' + escape(r.link) + '">See all occurrences →</a>' +
      '</div>';
    } else {
      html += '<p class="modal-no-root">This word has no trilateral root (particle, pronoun, etc.).</p>';
    }

    body.innerHTML = html;
    // Move focus to the close button for accessibility.
    const closeBtn = modal.querySelector('.modal-close');
    if (closeBtn) closeBtn.focus();
  }

  function close() {
    if (typeof modal.close === "function") {
      modal.close();
    } else {
      modal.removeAttribute("open");
    }
    unlockBodyScroll();
    if (lastFocused && typeof lastFocused.focus === "function") {
      lastFocused.focus();
    }
    body.innerHTML = "";
  }

  // Simple focus trap: cycle Tab within the modal.
  function trapFocus() {
    const sel = 'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"]), input:not([disabled])';
    const handler = (e) => {
      if (e.key !== "Tab") return;
      const focusables = Array.from(modal.querySelectorAll(sel)).filter(el => el.offsetParent !== null);
      if (focusables.length === 0) return;
      const first = focusables[0];
      const last = focusables[focusables.length - 1];
      if (e.shiftKey && document.activeElement === first) {
        e.preventDefault();
        last.focus();
      } else if (!e.shiftKey && document.activeElement === last) {
        e.preventDefault();
        first.focus();
      }
    };
    modal.addEventListener("keydown", handler);
    modal.addEventListener("close", () => {
      modal.removeEventListener("keydown", handler);
    }, { once: true });
  }

  // Click on a word button → open modal (after tooltip logic).
  document.addEventListener("click", e => {
    const w = e.target.closest('[data-component="word"]');
    if (!w) return;
    // If user is selecting text or modifier-click, let browser handle it.
    if (e.ctrlKey || e.metaKey || e.shiftKey) return;
    e.preventDefault();
    openFor(w);
  });

  // Close handlers.
  modal.addEventListener("click", e => {
    if (e.target === modal) close(); // backdrop click
    const act = e.target.closest('[data-action="close"]');
    if (act) close();
  });
  // Native <dialog> dispatches "cancel" on Escape.
  modal.addEventListener("cancel", e => {
    e.preventDefault();
    close();
  });

  function truncate(s, n) {
    if (!s) return "";
    if (s.length <= n) return s;
    return s.slice(0, n) + "…";
  }
  function escape(s) {
    if (s == null) return "";
    return String(s).replace(/[&<>"']/g, c => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }

  // ── Body scroll lock ───────────────────────────────────────
  // While the modal is open we freeze background scrolling. The
  // <html> element has `scrollbar-gutter: stable` so toggling
  // overflow doesn't cause horizontal layout shift when the
  // scrollbar disappears.
  let scrollLockCount = 0;
  function lockBodyScroll() {
    if (scrollLockCount++ === 0) {
      document.documentElement.classList.add("modal-open");
    }
  }
  function unlockBodyScroll() {
    if (scrollLockCount > 0 && --scrollLockCount === 0) {
      document.documentElement.classList.remove("modal-open");
    }
  }
})();
