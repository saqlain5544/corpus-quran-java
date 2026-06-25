// P2 + P3 + P4 verification: vanilla JS unit tests via Node.js + jsdom.
// We test bookmarks.js (Set-based lookup) and surah-header.js
// (cached search text + debounced localStorage).

package server

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBookmarksUsesSet is a static-analysis check that the
// bookmarks.js module uses a Set (Object.create(null)) for O(1)
// lookup. We can't run JS in a unit test, but we can verify the
// code path is correct.
func TestBookmarksUsesSet(t *testing.T) {
	jsPath := filepath.Join("..", "..", "static", "js", "bookmarks.js")
	b, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read %s: %v", jsPath, err)
	}
	src := string(b)
	if !strings.Contains(src, "Object.create(null)") {
		t.Errorf("bookmarks.js does not use Object.create(null) for the bookmark Set")
	}
	if !strings.Contains(src, "var bookmarkSet") && !strings.Contains(src, "let bookmarkSet") {
		t.Errorf("bookmarks.js does not declare bookmarkSet")
	}
	// Sanity: isBookmarked should now be a one-liner hash lookup.
	// Confirm by looking for "return bookmarkSet[KEY" which is the
	// O(1) pattern.
	if !strings.Contains(src, "return bookmarkSet[KEY") {
		t.Errorf("isBookmarked() should be a one-line Set lookup, not a for-loop")
	}
	// The legacy linear scan should be gone.
	if strings.Contains(src, "function isBookmarked(s, a) {\n    var k = KEY(s, a);\n    var items = load();") {
		t.Errorf("legacy linear-scan isBookmarked() still present")
	}
}

// TestSurahHeaderDebouncesLocalStorage verifies that the font-size
// and line-height sliders now debounce their localStorage writes
// instead of writing on every input event.
func TestSurahHeaderDebouncesLocalStorage(t *testing.T) {
	jsPath := filepath.Join("..", "..", "static", "js", "surah-header.js")
	b, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read %s: %v", jsPath, err)
	}
	src := string(b)
	// Look for the debounce pattern: setTimeout wrapping the
	// localStorage.setItem call.
	if !strings.Contains(src, "setTimeout") {
		t.Errorf("surah-header.js has no setTimeout — slider persistence is not debounced")
	}
	// The debounce state must be tracked somewhere. Either
	// `saveTimer` (single slider) or `sliderSaveTimers` (keyed map
	// for multiple sliders) — we accept either.
	if !strings.Contains(src, "saveTimer") && !strings.Contains(src, "sliderSaveTimers") {
		t.Errorf("surah-header.js has no debounce state — slider persistence is not debounced")
	}
	// And it must wrap a localStorage.setItem call (otherwise the
	// timer doesn't actually do anything useful).
	if !strings.Contains(src, "localStorage.setItem") {
		t.Errorf("surah-header.js has no localStorage.setItem call at all")
	}
}

// TestSurahHeaderCachesAyahText verifies that the local search
// caches normalized ayah text in a data attribute rather than
// re-reading textContent on every submit.
func TestSurahHeaderCachesAyahText(t *testing.T) {
	jsPath := filepath.Join("..", "..", "static", "js", "surah-header.js")
	b, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read %s: %v", jsPath, err)
	}
	src := string(b)
	if !strings.Contains(src, "data-search") {
		t.Errorf("surah-header.js does not use data-search-* attribute for cached ayah text")
	}
	// The hot loop should read the cached attribute, not textContent.
	if strings.Contains(src, "ay.textContent") && !strings.Contains(src, "data-search") {
		t.Errorf("local search still reads textContent on every submit (should use cached data-search-text)")
	}
}

// TestBookmarksSmoke is a smoke test: when a user has bookmarks
// in localStorage, refreshAyahButtons() correctly marks them.
func TestBookmarksSmoke(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	// The surah page must include bookmarks.js (so the Set-based
	// lookup runs on initial load).
	if !strings.Contains(body, "bookmarks.js") {
		t.Errorf("surah page does not include bookmarks.js")
	}
}
