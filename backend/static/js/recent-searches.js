// recent-searches.js — remembers the last 10 searches in localStorage
// and shows them in a dropdown under the global search box. Click a
// recent search to re-run it; click "Clear history" to wipe.
//
// Storage key: qr.recentSearches
// Item shape: { q: string, type: string, ts: number }
(function () {
  "use strict";

  const STORAGE_KEY = "qr.recentSearches";
  const MAX_ITEMS = 10;

  var input = document.querySelector('[data-component="gh-search-input"]');
  var form = document.querySelector('[data-component="global-search"]');
  var dropdown = document.querySelector('[data-component="gh-recent"]');
  var list = dropdown ? dropdown.querySelector(".gh-recent-list") : null;
  var clearBtn = dropdown ? dropdown.querySelector(".gh-recent-clear") : null;
  if (!input || !form || !dropdown || !list) return;

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
    } catch (e) {
      // localStorage may be disabled (e.g., file://, private mode).
      // Silently skip — the dropdown just won't persist across
      // sessions.
    }
  }

  // Render the dropdown items.
  function render() {
    var items = load();
    // De-duplicate by q+type, keep most recent.
    var seen = Object.create(null);
    var unique = [];
    for (var i = 0; i < items.length; i++) {
      var k = items[i].type + ":" + items[i].q;
      if (seen[k]) continue;
      seen[k] = true;
      unique.push(items[i]);
    }
    list.innerHTML = "";
    if (unique.length === 0) {
      var empty = document.createElement("li");
      empty.className = "gh-recent-empty-li";
      empty.textContent = "No recent searches";
      list.appendChild(empty);
      return;
    }
    for (var j = 0; j < unique.length; j++) {
      var item = unique[j];
      var li = document.createElement("li");
      var btn = document.createElement("button");
      btn.type = "button";
      var qSpan = document.createElement("span");
      qSpan.className = "gh-recent-q";
      // Heuristic: Arabic if any Arabic-range codepoint present.
      if (/[\u0600-\u06FF]/.test(item.q || "")) qSpan.classList.add("ar");
      qSpan.textContent = item.q || "(empty)";
      var typeSpan = document.createElement("span");
      typeSpan.className = "gh-recent-type";
      typeSpan.textContent = item.type || "";
      var arrow = document.createElement("span");
      arrow.className = "gh-recent-arrow";
      arrow.textContent = "→";
      btn.appendChild(qSpan);
      btn.appendChild(typeSpan);
      btn.appendChild(arrow);
      btn.addEventListener("click", function (it) {
        return function () {
          // Re-submit the form with the saved query + type.
          input.value = it.q;
          var sel = form.querySelector("select[name=type]");
          if (sel && it.type) sel.value = it.type;
          form.submit();
        };
      }(item));
      li.appendChild(btn);
      list.appendChild(li);
    }
  }

  // Capture submitted searches into the history.
  form.addEventListener("submit", function () {
    var q = input.value.trim();
    var sel = form.querySelector("select[name=type]");
    var type = sel ? sel.value : "english";
    if (!q) return; // don't store empty queries
    var items = load();
    // Remove any existing entry with the same q+type so the
    // dedupe keeps the most recent submission on top.
    var seen = Object.create(null);
    var filtered = [];
    for (var i = 0; i < items.length; i++) {
      var k = items[i].type + ":" + items[i].q;
      if (k === type + ":" + q || seen[k]) continue;
      seen[k] = true;
      filtered.push(items[i]);
    }
    filtered.unshift({ q: q, type: type, ts: Date.now() });
    if (filtered.length > MAX_ITEMS) filtered.length = MAX_ITEMS;
    save(filtered);
  });

  // Show the dropdown on focus (only when there are items).
  function showDropdown() {
    var items = load().filter(function (i) { return i.q; });
    if (items.length === 0) {
      dropdown.hidden = true;
      return;
    }
    render();
    dropdown.hidden = false;
  }
  function hideDropdown() { dropdown.hidden = true; }

  input.addEventListener("focus", showDropdown);
  input.addEventListener("input", showDropdown);
  document.addEventListener("click", function (e) {
    if (!dropdown.contains(e.target) && e.target !== input) hideDropdown();
  });
  input.addEventListener("keydown", function (e) {
    if (e.key === "Escape") hideDropdown();
  });

  if (clearBtn) {
    clearBtn.addEventListener("click", function () {
      save([]);
      hideDropdown();
    });
  }

  // Expose for tests (read-only snapshot).
  window.__qrRecentSearches = { load: load, save: save, render: render };
})();