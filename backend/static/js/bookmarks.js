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
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return [];
      var parsed = JSON.parse(raw);
      return Array.isArray(parsed) ? parsed : [];
    } catch (e) {
      return [];
    }
  }
  function save(items) {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
    } catch (e) { /* private mode — silently skip */ }
  }

  function isBookmarked(s, a) {
    var k = KEY(s, a);
    var items = load();
    for (var i = 0; i < items.length; i++) {
      if (items[i].surah + ":" + items[i].ayah === k) return true;
    }
    return false;
  }

  function add(s, a) {
    var items = load();
    var k = KEY(s, a);
    // Don't double-add.
    for (var i = 0; i < items.length; i++) {
      if (items[i].surah + ":" + items[i].ayah === k) return;
    }
    items.unshift({ surah: s, ayah: a, ts: Date.now() });
    save(items);
  }

  function remove(s, a) {
    var items = load();
    var k = KEY(s, a);
    var next = [];
    for (var i = 0; i < items.length; i++) {
      if (items[i].surah + ":" + items[i].ayah !== k) next.push(items[i]);
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
      var open = !drawer.hidden;
      if (open) {
        drawer.hidden = true;
        toggleBtn.setAttribute("aria-pressed", "false");
      } else {
        refreshDrawer();
        drawer.hidden = false;
        toggleBtn.setAttribute("aria-pressed", "true");
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
    head.innerHTML = '<span>Bookmarks</span>';
    var clear = document.createElement("button");
    clear.type = "button";
    clear.textContent = "Clear all";
    clear.addEventListener("click", function () {
      save([]);
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