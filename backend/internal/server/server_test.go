package server

import (
	"context"
	"embed"
	"encoding/json"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"quranreader/backend/internal/data"
	"quranreader/loc"
)

// ── Test harness ──────────────────────────────────────────────
//
// The Server package embeds templates, static and fonts at compile
// time. Tests reuse those same embeddings (go test compiles the
// package the same way). Data is loaded from detailed-quran.db.

//go:embed testdata
var testdataFS embed.FS

// testServer builds a Server backed by the detailed-quran.db and
// the package's embedded templates/static/fonts.
//
// The embed.FS in this file is rooted at "testdata/", so we wrap it
// in fs.Sub to expose the inner directories at the expected paths
// (templates/, static/, fonts/).
func testServer(t *testing.T) *Server {
	t.Helper()
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		t.Fatal(err)
	}
	q, m, r, meta, err := data.LoadAll(dbPath)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	// Wrap testdata so its root is testdata/, exposing templates/,
	// static/, fonts/ as sub-directories that New() can fs.Sub.
	rooted, err := fs.Sub(testdataFS, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	// fs.Sub returns fs.FS which doesn't have ReadFile/ReadDir; use
	// our subFS wrapper to re-expose those methods.
	wrapped := subRoot{fsys: rooted}
	srv, err := New(q, m, r, meta, Options{
		TemplatesFS:  wrapped,
		StaticFS:     wrapped,
		FontsFS:      wrapped,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		Translations: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	return srv
}

// wrapSub produces a sub-filesystem rooted at dir, returning a fresh
// embed.FS-shaped value. Used so handlers can read paths like
// "static/css/tokens.css" after the sub is mounted.
func wrapSub(fsys fs.FS) fs.FS { return fsys }

// ── Tests ─────────────────────────────────────────────────────

func TestHomepage(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	srv.Handler().ServeHTTP(w, r)

	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("content-type = %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "الفاتحة") {
		t.Error("homepage should mention surah 1 (الفاتحة)")
	}
	if !strings.Contains(body, "الناس") {
		t.Error("homepage should mention surah 114 (الناس)")
	}
}

func TestSurahPageValid(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d body = %.500s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "البقرة") {
		t.Error("expected البقرة in body")
	}
	if !strings.Contains(body, `data-component="word"`) {
		t.Error("expected word buttons")
	}
}

func TestSurahPageInvalid(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/115", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestSurahVersePage(t *testing.T) {
	// Verse-anchored navigation is client-side via #verse-N hash
	// fragments. The server renders all verses with id anchors so
	// that verse-scroll.js can find and scroll to them.
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `id="verse-255"`) {
		t.Error("expected anchor verse-255")
	}
	// Rootless words have data-root="" (empty string); words with
	// roots have a buckwalter code. Verify we see at least one.
	if !strings.Contains(body, `data-root="Alh"`) {
		t.Error("expected data-root='Alh' for Allah occurrences")
	}
}

func TestSearchEnglish(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=merciful&type=english", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "merciful") {
		t.Error("expected 'merciful' in result")
	}
}

func TestSearchRoot(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=qwl&type=root", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestSearchEmpty(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

func TestRootsList(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/roots?page=1&sort=desc", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestRootsListAsc(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/roots?page=1&sort=asc", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestRootDetail(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/root/detailed/qwl", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d body = %.500s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "qwl") {
		t.Error("expected 'qwl' in body")
	}
}

func TestRootDetailNotFound(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/root/detailed/zzzNotARealRoot", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 404 {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// TestRootDetailArabicVariants verifies the Arabic ByArabic reverse
// lookup works for the variants users actually type: with/without
// spaces, hamza variants, tashkeel. Regression: the index was built
// with stripDiacritics (only strips tashkeel) while queries used
// NormalizeArabic — they disagreed on both alif variants AND spaces,
// so these URLs all returned 404.
func TestRootDetailArabicVariants(t *testing.T) {
	srv := testServer(t)
	cases := []struct {
		name  string
		input string // URL-encoded Arabic letters
	}{
		{"no_spaces_no_hamza", "%D9%82%D9%88%D9%84"},           // قول
		{"with_spaces_no_hamza", "%D9%82%20%D9%88%20%D9%84"},   // ق و ل
		{"no_spaces_with_hamza", "%D8%A3%D8%A8%D9%8A"},         // أبي
		{"with_spaces_with_hamza", "%D8%A3%20%D8%A8%20%D9%8A"}, // أ ب ي
		{"with_tashkeel", "%D9%82%D9%8E%D9%88%D9%92%D9%84"},    // قَوْل
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/root/detailed/"+c.input, nil)
			srv.Handler().ServeHTTP(w, r)
			if w.Code != 200 {
				t.Errorf("status = %d, want 200; body=%.200s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAbout(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/about", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

// ── API tests ─────────────────────────────────────────────────

func TestAPIWordValid(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	q := url.Values{}
	q.Set("s", "1")
	q.Set("a", "1")
	q.Set("w", "2")
	r := httptest.NewRequest("GET", "/api/word?"+q.Encode(), nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Errorf("content-type = %q", w.Header().Get("Content-Type"))
	}
	var resp apiWordResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Surah != 1 || resp.Ayah != 1 || resp.Word != 2 {
		t.Errorf("resp = %+v", resp)
	}
	if len(resp.Segments) < 1 {
		t.Error("expected segments")
	}
	if resp.Root == nil {
		t.Error("expected root for 1:1:2 (ٱللَّهِ)")
	}
	// New tooltip fields.
	if resp.Lemma == "" {
		t.Error("Lemma should be populated")
	}
	if resp.Gloss == "" {
		t.Error("Gloss should be populated")
	}
	if resp.POS == "" {
		t.Error("POS should be populated")
	}
	if resp.Function == "" {
		t.Error("Function should be populated")
	}
	if resp.Root != nil && resp.Root.Letters == "" {
		t.Error("Root.Letters should be populated")
	}
}

func TestAPIWordInvalid(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/word?s=200&a=1&w=1", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 400 {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAPISearch(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/search?q=allah&type=english", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		Query   string `json:"query"`
		Type    string `json:"type"`
		Results []struct {
			Surah int `json:"surah"`
			Ayah  int `json:"ayah"`
		} `json:"results"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if len(resp.Results) == 0 {
		t.Error("expected at least 1 result")
	}
}

func TestAPIRoots(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/roots?page=1&sort=desc", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAPIRootOccurrences(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/root/qwl/occurrences", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		Root        string `json:"root"`
		Arabic      string `json:"arabic"`
		Occurrences []struct {
			Surah int `json:"surah"`
			Ayah  int `json:"ayah"`
		} `json:"occurrences"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("json: %v", err)
	}
	if resp.Root != "qwl" {
		t.Errorf("root = %q", resp.Root)
	}
	if len(resp.Occurrences) == 0 {
		t.Error("expected occurrences")
	}
}

func TestAPIHealth(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/health", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

// ── Static asset tests ────────────────────────────────────────

func TestStaticCSS(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/css/tokens.css", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "--bg") {
		t.Error("expected tokens.css to define --bg variable")
	}
}

func TestStaticJS(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/js/app.js", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
}

// ── Direct handler unit tests ─────────────────────────────────

func TestRespondJSON(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	srv.respondJSON(w, http.StatusTeapot, map[string]string{"hello": "world"})
	if w.Code != http.StatusTeapot {
		t.Errorf("status = %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"hello":"world"`) {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestLocKeyInHandler(t *testing.T) {
	k := loc.Key(1, 1, 1)
	if loc.String(k) != "1:1:1" {
		t.Errorf("got %s", loc.String(k))
	}
}

func TestSearchURLEncoding(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	encoded := url.QueryEscape("ٱللَّهِ")
	r := httptest.NewRequest("GET", "/api/search?q="+encoded+"&type=arabic", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

func TestPaginationBoundary(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/roots?page=9999", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestContextServer(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil).WithContext(context.Background())
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("status = %d", w.Code)
	}
}

// silence unused warnings if strconv isn't otherwise referenced.
var _ = strconv.Itoa
