// verse-scroll.js — when a /surah/{id} URL is loaded with a #verse-N
// hash, smooth-scroll to that verse and apply a brief focus highlight.
//
// Verse-anchored navigation is purely client-side via the URL hash.
// We add the .ayah-focus class so the CSS gradient underline is
// visible during the scroll, then strip it after a short delay so the
// highlight fades out. We also update the URL hash (via replaceState)
// so the address bar stays accurate across back/forward navigation.

(function () {
  "use strict";

  function findTarget() {
    // Look for an explicit #verse-N hash in the URL — the only
    // mechanism for verse-anchored navigation.
    const hash = window.location.hash;
    if (hash && hash.startsWith("#verse-")) {
      const el = document.querySelector(hash);
      if (el) return el;
    }
    return null;
  }

  function go() {
    const target = findTarget();
    if (!target) return;

    // Apply the highlight class so the gradient underline is visible
    // while the smooth-scroll runs. Stripped after a delay so it fades
    // out (mirrors the local-search highlight in surah-header.js).
    QR.dom.focusHighlight(target, 2500);

    // scrollIntoView respects the CSS scroll-margin-top on .ayah
    // (8rem), which already accounts for the sticky global header
    // and the surah header.
    try {
      target.scrollIntoView({ behavior: "smooth", block: "start" });
    } catch (e) {
      target.scrollIntoView();
    }

    // Make the address bar shareable. Use replaceState so we don't
    // bloat the back-history with every verse scroll.
    const id = target.id || ("verse-" + (target.dataset.ayah || ""));
    if (id && window.location.hash !== "#" + id) {
      try {
        history.replaceState(null, "", "#" + id);
      } catch (e) { /* file:// or restrictive env */ }
    }
  }

  // Defer until layout has settled — the server-rendered ayah
  // heights depend on the rendered font which may not be loaded
  // synchronously (hafs.woff2 swap).
  if (document.readyState === "complete") {
    requestAnimationFrame(go);
  } else {
    window.addEventListener("load", () => requestAnimationFrame(go), { once: true });
  }
})();
