// dom.js — shared DOM utilities used by multiple components.
// Import this BEFORE any file that uses these helpers.
//
// Usage:
//   QR.dom.escape("<mark>")           → "&lt;mark&gt;"
//   QR.dom.focusHighlight(el, 2500)   → adds .ayah-focus, removes after 2.5s
//   QR.dom.debounce(fn, 150)          → debounced fn
//   QR.dom.closestWord(el)            → el.closest('[data-component="word"]')

(function () {
  "use strict";
  window.QR = window.QR || {};
  window.QR.dom = {
    // HTML-escape a string so user input / Arabic text can be
    // safely interpolated into innerHTML.
    escape: function (s) {
      if (s == null) return "";
      return String(s).replace(/[&<>"']/g, function (c) {
        return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
      });
    },

    // Add .ayah-focus class to el, then remove after ms (default 2500).
    // Used by verse-scroll.js and surah-header.js for the gradient
    // underline highlight that fades out.
    focusHighlight: function (el, ms) {
      if (!el) return;
      ms = ms || 2500;
      el.classList.add("ayah-focus");
      setTimeout(function () { el.classList.remove("ayah-focus"); }, ms);
    },

    // Debounce a function — common pattern for scroll/resize/slider
    // handlers that trigger expensive operations.
    debounce: function (fn, ms) {
      ms = ms || 150;
      var timer = null;
      return function () {
        var ctx = this, args = arguments;
        if (timer) clearTimeout(timer);
        timer = setTimeout(function () { fn.apply(ctx, args); }, ms);
      };
    },

    // Find the closest word component element (event delegation).
    closestWord: function (el) {
      return el && el.closest ? el.closest('[data-component="word"]') : null;
    },

    // Find the closest ayah bookmark button.
    closestBookmark: function (el) {
      return el && el.closest ? el.closest('[data-component="ayah-bookmark"]') : null;
    }
  };
})();
