// global-header.js — minimal helpers for the global header.
// Three independent features are wired up here:
//   1. Global search form — submit / type persistence / "/" focus
//   2. Hamburger menu toggle — on narrow viewports, taps on the
//      hamburger reveal/hide the slide-in nav drawer
//   3. Recent searches dropdown — populated by recent-searches.js
//      (this file just leaves a hook in case other code wants to
//       close the drawer on outside click)

(function () {
  "use strict";

  /* ── Search form ─────────────────────────────────────────── */
  const form = document.querySelector('[data-component="global-search"]');
  if (form) {
    const input = form.querySelector('input[name="q"]');
    const select = form.querySelector('select[name="type"]');

    // Restore last-used search type (e.g. "root" or "arabic") so the
    // user doesn't need to re-select it on every search.
    var last = QR.storage.get("qr.searchType");
    if (last && select) {
      var opt = Array.from(select.options).find(function(o) { return o.value === last; });
      if (opt) select.value = last;
    }

    // Persist on submit.
    form.addEventListener("submit", function() {
      if (select) QR.storage.set("qr.searchType", select.value);
    });
  }

  /* ── Hamburger menu toggle ────────────────────────────────── */
  // The hamburger button is only visible on narrow viewports (see
  // .gh-hamburger { display: none } + the @media (max-width: 900px)
  // override). Tapping it toggles the .gh-nav-drawer.
  const hamburger = document.querySelector('[data-component="gh-hamburger"]');
  const drawer = document.querySelector('[data-component="gh-nav-drawer"]');
  if (hamburger && drawer) {
    function setOpen(open) {
      hamburger.setAttribute("aria-expanded", open ? "true" : "false");
      hamburger.setAttribute("aria-label", open ? "Close navigation menu" : "Open navigation menu");
      drawer.classList.toggle("is-open", !!open);
    }

    hamburger.addEventListener("click", function() {
      var isOpen = hamburger.getAttribute("aria-expanded") === "true";
      setOpen(!isOpen);
    });

    // Close the drawer when a link inside is clicked — the page is
    // about to navigate, leaving the drawer open looks broken.
    drawer.querySelectorAll('[data-component="gh-nav-link"]').forEach(function(link) {
      link.addEventListener("click", function() {
        setOpen(false);
      });
    });

    // Close the drawer with Escape — standard disclosure pattern.
    document.addEventListener("keydown", function(e) {
      if (e.key === "Escape" && hamburger.getAttribute("aria-expanded") === "true") {
        setOpen(false);
        hamburger.focus();
      }
    });

    // Close the drawer on outside click. Cheap to test; doesn't
    // matter if the drawer is hidden because the listener only fires
    // when the drawer is in the DOM (which it always is).
    document.addEventListener("click", function(e) {
      if (hamburger.getAttribute("aria-expanded") !== "true") return;
      if (drawer.contains(e.target) || hamburger.contains(e.target)) return;
      setOpen(false);
    });

    // Close the drawer if the viewport widens past 900px (the
    // hamburger hides and the drawer has no trigger). Without this,
    // re-opening it would show stale state.
    var mq = window.matchMedia("(min-width: 901px)");
    function onMq(e) { if (e.matches) setOpen(false); }
    if (mq.addEventListener) mq.addEventListener("change", onMq);
    else if (mq.addListener) mq.addListener(onMq); // old WebKit fallback
  }

  /* ── Keyboard: "/" focuses search, "Esc" clears it ────────── */
  document.addEventListener("keydown", function(e) {
    if (!form) return;
    const input = form.querySelector('input[name="q"]');
    if (e.key === "/" && document.activeElement !== input &&
        !document.activeElement.matches("input, textarea, select")) {
      e.preventDefault();
      if (input) input.focus();
    } else if (e.key === "Escape" && document.activeElement === input) {
      input.value = "";
    }
  });
})();
