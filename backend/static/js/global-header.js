// global-header.js — minimal helpers for the global search form.
// The form already submits as GET /search so we only add:
//   - type persistence (last-used type sticks across submissions)
//   - keyboard: "/" focuses the search box, "Esc" clears it

(function () {
  "use strict";

  const form = document.querySelector('[data-component="global-search"]');
  if (!form) return;
  const input = form.querySelector('input[name="q"]');
  const select = form.querySelector('select[name="type"]');

  // Restore last-used type.
  try {
    const last = localStorage.getItem("qr.searchType");
    if (last && select) {
      const opt = Array.from(select.options).find(o => o.value === last);
      if (opt) select.value = last;
    }
  } catch (e) {}

  // Persist on submit.
  form.addEventListener("submit", () => {
    try {
      if (select) localStorage.setItem("qr.searchType", select.value);
    } catch (e) {}
  });

  // "/" focuses the search box.
  document.addEventListener("keydown", (e) => {
    if (e.key === "/" && document.activeElement !== input &&
        !document.activeElement.matches("input, textarea")) {
      e.preventDefault();
      if (input) input.focus();
    } else if (e.key === "Escape" && document.activeElement === input) {
      input.value = "";
    }
  });
})();
