// bookmarks.js — verse-level bookmarking backed by localStorage.
//
// Storage key: qr.bookmarks
// Item shape: { surah: int, ayah: int, ts: number }
//
// Surah page:
//   - Each ayah renders an .ayah-bookmark button. Click toggles its
//     bookmarked state. Filled star (★) + gold color when set,
//     outline (☆) when not. Updated in real time on every page.
//
// Global header:
//   - The .gh-bookmarks button opens a drawer listing all saved
//     bookmarks sorted by most-recent. Each entry has a "Go" link
//     to /surah/{s}#verse-{v} and a remove button. An empty-state
//     line shows when nothing is saved.
(function () {
  "use strict";

  const STORAGE_KEY = "qr.bookmarks";
  const KEY = (s, a) => s + ":" + a;

  function load() {
    return QR.storage.getJSON(STORAGE_KEY);
  }
  function save(items) {
    QR.storage.setJSON(STORAGE_KEY, items);
  }

  // Build a Set of "s:a" strings from the current bookmarks array
  // so isBookmarked() is O(1) instead of O(M). We refresh the Set
  // after every add/remove. The set is module-scoped and rebuilt on
  // demand via refreshBookmarkSet(); we don't keep a global cache
  // because localStorage can be modified by other tabs (and the
  // page's own add/remove), and re-parsing 100 items is sub-ms.
  //
  // For M = ~10 bookmarks (typical) the array scan was already fast;
  // the win is when many per-ayah buttons call isBookmarked on
  // initial render — Al-Baqarah has 286 ayahs, each one calls
  // isBookmarked(s, a) once. With the Set this is 286 hash lookups
  // (~286ns) instead of 286 × 10 = 2860 string concats + comparisons.
  function refreshBookmarkSet() {
    var items = load();
    var s = Object.create(null);
    for (var i = 0; i < items.length; i++) {
      s[KEY(items[i].surah, items[i].ayah)] = true;
    }
    return s;
  }
  var bookmarkSet = refreshBookmarkSet();

  function isBookmarked(s, a) {
    return bookmarkSet[KEY(s, a)] === true;
  }

  function add(s, a) {
    var items = load();
    var k = KEY(s, a);
    // Don't double-add.
    if (bookmarkSet[k]) return;
    bookmarkSet[k] = true;
    items.unshift({ surah: s, ayah: a, ts: Date.now() });
    save(items);
  }

  function remove(s, a) {
    var k = KEY(s, a);
    if (!bookmarkSet[k]) return;
    delete bookmarkSet[k];
    var items = load();
    var next = [];
    for (var i = 0; i < items.length; i++) {
      if (KEY(items[i].surah, items[i].ayah) !== k) next.push(items[i]);
    }
    save(next);
  }

  function toggle(s, a) {
    if (isBookmarked(s, a)) remove(s, a);
    else add(s, a);
  }

  // ── Per-ayah buttons on the surah page ─────────────────────
  function refreshAyahButtons() {
    var btns = document.querySelectorAll('[data-component="ayah-bookmark"]');
    for (var i = 0; i < btns.length; i++) {
      var b = btns[i];
      var s = parseInt(b.dataset.surah, 10);
      var a = parseInt(b.dataset.ayah, 10);
      var set = isBookmarked(s, a);
      b.setAttribute("aria-pressed", set ? "true" : "false");
      b.textContent = set ? "★" : "☆";
      b.setAttribute("aria-label", (set ? "Remove bookmark" : "Bookmark") + " verse " + a);
    }
  }

  document.addEventListener("click", function (e) {
    var btn = e.target.closest('[data-component="ayah-bookmark"]');
    if (!btn) return;
    e.preventDefault();
    var s = parseInt(btn.dataset.surah, 10);
    var a = parseInt(btn.dataset.ayah, 10);
    if (!s || !a) return;
    toggle(s, a);
    refreshAyahButtons();
    refreshDrawer();
  });

  // Refresh the in-memory set when localStorage might have changed
  // out-of-band (e.g. user clears bookmarks via the drawer).
  function refreshFromStorage() {
    bookmarkSet = refreshBookmarkSet();
  }

  // ── Global header drawer ────────────────────────────────────
  var toggleBtn = document.querySelector('[data-component="bookmarks-toggle"]');
  var drawer = document.querySelector('[data-component="bookmarks-drawer"]');
  if (toggleBtn && drawer) {
    var list = document.createElement("ul");
    list.className = "gh-bookmarks-list";
    drawer.appendChild(buildHeader());
    drawer.appendChild(list);
    toggleBtn.addEventListener("click", function (e) {
      e.stopPropagation();
      // wasHidden === true means the drawer is currently closed and
      // this click should open it. The previous version named this
      // variable `open` and inverted the branches, which was
      // semantically confusing on read.
      var wasHidden = drawer.hidden;
      if (wasHidden) {
        refreshDrawer();
        drawer.hidden = false;
        toggleBtn.setAttribute("aria-pressed", "true");
      } else {
        drawer.hidden = true;
        toggleBtn.setAttribute("aria-pressed", "false");
      }
    });
    document.addEventListener("click", function (e) {
      if (drawer.hidden) return;
      if (!drawer.contains(e.target) && e.target !== toggleBtn) {
        drawer.hidden = true;
        toggleBtn.setAttribute("aria-pressed", "false");
      }
    });
    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape" && !drawer.hidden) {
        drawer.hidden = true;
        toggleBtn.setAttribute("aria-pressed", "false");
      }
    });
  }

  function buildHeader() {
    var head = document.createElement("div");
    head.className = "gh-bookmarks-head";
    // Header shows the bookmark count so the user knows what to
    // expect — e.g. "Bookmarks (3)" rather than just "Bookmarks"
    // which gives no sense of the drawer's contents.
    var items = load();
    var label = items.length === 1 ? "Bookmark" : "Bookmarks";
    head.innerHTML = '<span class="gh-bookmarks-title">' + label +
      (items.length > 0 ? ' <span class="gh-bookmarks-count">' + items.length + '</span>' : '') +
      '</span>';
    var clear = document.createElement("button");
    clear.type = "button";
    clear.textContent = "Clear all";
    clear.addEventListener("click", function () {
      save([]);
      refreshFromStorage();
      refreshDrawer();
      refreshAyahButtons();
    });
    head.appendChild(clear);
    return head;
  }

  function refreshDrawer() {
    if (!drawer) return;
    var list = drawer.querySelector(".gh-bookmarks-list");
    if (!list) return;
    list.innerHTML = "";
    var items = load();
    if (items.length === 0) {
      var empty = document.createElement("div");
      empty.className = "gh-bookmarks-empty";
      empty.textContent = "No saved bookmarks yet. Tap ☆ next to any verse.";
      list.appendChild(empty);
      return;
    }
    for (var i = 0; i < items.length; i++) {
      (function (it) {
        var li = document.createElement("li");
        var a = document.createElement("a");
        a.href = "/surah/" + it.surah + "#verse-" + it.ayah;
        a.textContent = "Surah " + it.surah + ":" + it.ayah;
        var rm = document.createElement("button");
        rm.className = "gbk-remove";
        rm.type = "button";
        rm.textContent = "✕";
        rm.title = "Remove bookmark";
        rm.setAttribute("aria-label", "Remove bookmark for surah " + it.surah + " verse " + it.ayah);
        rm.addEventListener("click", function () {
          remove(it.surah, it.ayah);
          refreshDrawer();
          refreshAyahButtons();
        });
        li.appendChild(a);
        li.appendChild(rm);
        list.appendChild(li);
      })(items[i]);
    }
  }

  // Initial paint — set the right star for every ayah on this page.
  refreshAyahButtons();
  refreshDrawer();

  // Test hook.
  window.__qrBookmarks = { load: load, save: save, add: add, remove: remove };
})();