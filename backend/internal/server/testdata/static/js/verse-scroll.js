// verse-scroll.js — when a /surah/{id}/{verse} URL is loaded, scroll
// to the target verse and apply a brief focus highlight.
//
// The server already marks the target <section> with the class
// `.ayah-focus` (see surah.tmpl). Our job is to find it on load and
// smooth-scroll it into view, also updating the URL hash so the
// address bar reflects the verse for shareability.
//
// We also handle the legacy case where someone navigates to a URL
// with an explicit `#verse-N` hash.

(function () {
  "use strict";

  function findTarget() {
    // 1. Prefer an explicit #verse-N hash (most shareable).
    const hash = window.location.hash;
    if (hash && hash.startsWith("#verse-")) {
      const el = document.querySelector(hash);
      if (el) return el;
    }
    // 2. Fall back to the ayah-focus class that the server template
    //    applies when the URL contains a verse number.
    const focused = document.querySelector(".ayah.ayah-focus");
    if (focused) return focused;
    // 3. Last resort: look at any [data-ayah] on the page that has
    //    the hash id (defensive).
    return null;
  }

  function go() {
    const target = findTarget();
    if (!target) return;

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

    // Strip the focus class after a delay so the gradient underline
    // fades out (mirrors the local-search highlight in
    // surah-header.js).
    setTimeout(() => target.classList.remove("ayah-focus"), 2500);
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
