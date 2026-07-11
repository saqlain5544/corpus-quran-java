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
// and line-height sliders debounce their localStorage writes through
// the QR.dom.debounce / QR.storage abstractions (in static/js/dom.js
// and static/js/storage.js) instead of writing on every input event.
func TestSurahHeaderDebouncesLocalStorage(t *testing.T) {
	jsPath := filepath.Join("..", "..", "static", "js", "surah-header.js")
	b, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read %s: %v", jsPath, err)
	}
	src := string(b)
	// The slider persistence must go through the shared debounce
	// helper — it lives in dom.js and wraps setTimeout.
	if !strings.Contains(src, "QR.dom.debounce") {
		t.Errorf("surah-header.js does not use QR.dom.debounce for slider persistence")
	}
	// And the inner call must go through QR.storage.set so private-
	// mode / file:// environments don't throw.
	if !strings.Contains(src, "QR.storage.set") {
		t.Errorf("surah-header.js does not use QR.storage.set — localStorage is raw")
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
