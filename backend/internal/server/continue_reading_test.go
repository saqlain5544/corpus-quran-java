package server

// Vanilla-JS test for continue-reading.js. We verify that the
// static analysis reflects the IO-based topmost-ayah detection
// and that the served page references the script.
//
// This is a smoke test (no JS unit-test framework), but it
// exercises the real code path against the real backend.

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type lastRead struct {
	Surah int `json:"surah"`
	Ayah  int `json:"ayah"`
}

// TestContinueReadingUsesIntersectionObserver verifies that the
// static analysis reflects the IO-based topmost-ayah detection.
func TestContinueReadingUsesIntersectionObserver(t *testing.T) {
	jsPath := filepath.Join("..", "..", "static", "js", "continue-reading.js")
	b, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatalf("read %s: %v", jsPath, err)
	}
	src := string(b)
	if !strings.Contains(src, "IntersectionObserver") {
		t.Errorf("continue-reading.js does not reference IntersectionObserver — IO path missing")
	}
	if !strings.Contains(src, "rootMargin") {
		t.Errorf("continue-reading.js does not configure rootMargin")
	}
}

// TestSurahPageIncludesContinueReading is the smoke test: when
// the page is loaded, continue-reading.js runs and (eventually)
// writes to localStorage. We can't drive a real browser here,
// but we can verify the served page references the script and
// contains .ayah elements for the observer to watch.
func TestSurahPageIncludesContinueReading(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "continue-reading.js") {
		t.Errorf("surah page does not include continue-reading.js")
	}
	if !strings.Contains(body, `class="ayah`) {
		t.Errorf("surah page has no .ayah elements")
	}
}

// Unused but kept to anchor the JSON shape in tests.
var _ = lastRead{}
