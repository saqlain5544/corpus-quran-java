// storage.js — safe localStorage helpers. All read/write operations
// are wrapped in try/catch so private-mode / file:// environments
// don't throw. Import this BEFORE any file that touches localStorage.
//
// Usage:
//   QR.storage.get("qr.theme")         → string or ""
//   QR.storage.set("qr.theme", "dark")
//   QR.storage.getJSON("qr.bookmarks") → parsed JSON or []
//   QR.storage.setJSON("qr.bookmarks", [{surah:1,ayah:1}])

(function () {
  "use strict";
  window.QR = window.QR || {};
  window.QR.storage = {
    get: function (key) {
      try { return localStorage.getItem(key) || ""; } catch (e) { return ""; }
    },
    set: function (key, value) {
      try { localStorage.setItem(key, value); } catch (e) {}
    },
    getJSON: function (key) {
      try {
        var raw = localStorage.getItem(key);
        if (!raw) return [];
        var parsed = JSON.parse(raw);
        return Array.isArray(parsed) ? parsed : [];
      } catch (e) { return []; }
    },
    setJSON: function (key, value) {
      try { localStorage.setItem(key, JSON.stringify(value)); } catch (e) {}
    }
  };
})();
