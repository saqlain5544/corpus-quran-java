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
	"regexp"
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

// ── Structural CSS vocabulary ────────────────────────────────────
//
// These tests guard against accidental removal of the surfaces.css
// vocabulary from the page templates. They are intentionally
// "shape" tests (looking for class names) rather than full snapshot
// comparisons — that way cosmetic markup changes don't break the
// suite, but a refactor that drops the layout/surface primitives
// will fail loudly.

func TestStructuralClasses_Homepage(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	must := []string{
		`class="layout-header"`,         // page-level header
		`class="surface surface--bare"`, // card list lives on a bare surface
		`<ol class="surah-grid">`,
	}
	for _, m := range must {
		if !strings.Contains(body, m) {
			t.Errorf("homepage missing %q", m)
		}
	}
}

func TestStructuralClasses_RootDetail(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/root/detailed/qwl", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	must := []string{
		`class="layout-shell layout-shell--two-col"`,
		`class="layout-aside layout-aside--sticky stack"`,
		// Both columns must use surfaces; the aside stacks distribution +
		// concordance lemmas.
		`class="surface stack"`,
	}
	for _, m := range must {
		if !strings.Contains(body, m) {
			t.Errorf("root detail missing %q", m)
		}
	}
}

func TestStructuralClasses_Search(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/search?q=allah&type=english", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `class="layout-header"`) {
		t.Error("search page missing layout-header")
	}
	if !strings.Contains(body, `measure-wide`) {
		t.Error("search page missing measure-wide")
	}
	// Results, if present, must be wrapped in a surface.
	if strings.Contains(body, `<ol class="search-results">`) &&
		!strings.Contains(body, `<ol class="search-results surface">`) {
		t.Error("search-results <ol> should carry class=\"surface\"")
	}
}

func TestStructuralClasses_About(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/about", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	// measure-base appears on the page wrapper, not standalone.
	if !strings.Contains(body, `measure-base`) {
		t.Error("about page missing measure-base constraint")
	}
	// Each section on the about page should be a surface.
	count := strings.Count(body, `<section class="surface`)
	if count < 3 {
		t.Errorf("about page: expected ≥3 surface sections, got %d", count)
	}
}

func TestStructuralClasses_RootsList(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/roots", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `layout-header-actions`) {
		t.Error("roots list missing layout-header-actions")
	}
	if !strings.Contains(body, `class="cluster"`) {
		t.Error("roots list controls not in a cluster")
	}
}

func TestStructuralClasses_Surah(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/1", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	// Bismillah must be on a surface--bare.
	if !strings.Contains(body, `class="bismillah surface surface--bare"`) {
		t.Error("surah page: bismillah should carry surface surface--bare")
	}
	// Each ayah section must be a bare surface (no chrome, just spacing).
	if !strings.Contains(body, `class="ayah surface surface--bare"`) {
		t.Error("surah page: ayah should carry surface surface--bare")
	}
}

// ── Verse-mark / word-cluster tests ─────────────────────────────
//
// The Quran verse mark (ayah number) used to wrap to its own line
// when text justification pushed the last word onto the previous line.
// The fix: wrap each word + its trailing marks in a
// <span class="word-cluster"> (display:inline-block + nowrap), and
// include the ayah number SVG inside the LAST word's cluster so the
// mark is physically anchored to the last word.

func TestWordClusterEveryWord(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	// Every word button must be wrapped in a word-cluster span.
	buttons := strings.Count(body, `data-component="word"`)
	clusters := strings.Count(body, `class="word-cluster`)
	if buttons != clusters {
		t.Errorf("word buttons (%d) and word-cluster spans (%d) differ — every word should be wrapped", buttons, clusters)
	}
	if clusters == 0 {
		t.Error("no word-cluster spans rendered")
	}
}

func TestAyahNumberSVGInsideLastWord(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()

	// Each verse must have exactly one ayah-number SVG.
	verses := strings.Count(body, `id="verse-`)
	svgs := strings.Count(body, `<svg class="ayah-number"`)
	if verses != svgs {
		t.Errorf("verses (%d) and ayah-number SVGs (%d) differ", verses, svgs)
	}

	// The ayah-number SVG must be inside a word-cluster--end span
	// (i.e., physically nested inside the last word's cluster).
	// Count word-cluster--end spans: should match verse count.
	endClusters := strings.Count(body, `word-cluster--end`)
	if endClusters != verses {
		t.Errorf("word-cluster--end (%d) and verse count (%d) differ — every verse should have one last-word cluster", endClusters, verses)
	}

	// Each verse's word-cluster--end span must contain both the word
	// button AND the ayah-number SVG. We spot-check a handful.
	for _, vid := range []string{"verse-1", "verse-100", "verse-286"} {
		idx := strings.Index(body, `id="`+vid+`"`)
		if idx < 0 {
			t.Errorf("missing %s", vid)
			continue
		}
		// Find the next </section> after the verse anchor.
		end := strings.Index(body[idx:], "</section>")
		if end < 0 {
			t.Errorf("verse %s: no </section>", vid)
			continue
		}
		verse := body[idx : idx+end]
		if !strings.Contains(verse, "word-cluster--end") {
			t.Errorf("verse %s: no word-cluster--end", vid)
			continue
		}
		// The word-cluster--end span and the SVG must coexist in
		// the same verse — the SVG must come after the word button
		// inside the cluster span.
		openIdx := strings.Index(verse, `class="word-cluster word-cluster--end"`)
		if openIdx < 0 {
			continue
		}
		// Find the matching </span> by counting opens.
		depth := 1
		pos := openIdx + len(`class="word-cluster word-cluster--end"`)
		for pos < len(verse) {
			next := strings.IndexAny(verse[pos:], "<>")
			if next < 0 {
				break
			}
			pos += next
			if pos < len(verse) && verse[pos] == '<' {
				if pos+5 < len(verse) && verse[pos:pos+5] == "<span" {
					depth++
				} else if pos+6 < len(verse) && verse[pos:pos+6] == "</span" {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			pos++
		}
		cluster := verse[openIdx:pos]
		if !strings.Contains(cluster, "data-component=\"word\"") {
			t.Errorf("verse %s: word-cluster--end has no word button", vid)
		}
		if !strings.Contains(cluster, "<svg class=\"ayah-number\"") {
			t.Errorf("verse %s: word-cluster--end has no ayah-number SVG — the verse mark is not anchored to the last word", vid)
		}
	}
}

func TestStructuralClasses_Error(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/115", nil) // invalid -> 404
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 404 {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `class="error-page surface surface--elev measure-narrow"`) {
		t.Error("error page must be a narrow elevated surface")
	}
}

func TestSurfacesCSSAsset(t *testing.T) {
	// The structural vocabulary lives in backend/static/css/surfaces.css
	// and is referenced from the layout's <link>. Confirm both that
	// the asset is served and that it contains the vocabulary.
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/css/surfaces.css", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("surfaces.css status = %d", w.Code)
	}
	body := w.Body.String()
	for _, sel := range []string{
		".surface",
		".surface--sunken",
		".surface--elev",
		".surface--bare",
		".stack",
		".cluster",
		".layout-shell--two-col",
		".layout-header",
		".layout-aside--sticky",
		".measure-narrow",
	} {
		if !strings.Contains(body, sel) {
			t.Errorf("surfaces.css missing selector %q", sel)
		}
	}
}

// ── oklch accent + panel highlight ─────────────────────────────
//
// Gold accent in CSS Color 4 oklch() — the perceptually-uniform
// color space — gives us a manuscript-ink palette that holds its
// hue across lightness/chroma tweaks. The surah side panel is a
// "live" surface (it appears on word click) and uses gold-tinted
// background + gold border to read as distinct from the main
// surah column.

func TestTokensCSSUsesOklchAccent(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/css/tokens.css", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("tokens.css status = %d", w.Code)
	}
	body := w.Body.String()

	// Accent must be gold-ish: oklch hue in the yellow range (60–95).
	// Match the literal "--accent:" assignment line and verify its
	// oklch() hue falls in [60, 95] for the light theme.
	m := regexp.MustCompile(`--accent:\s*oklch\(\s*([\d.]+)%\s+([\d.]+)\s+([\d.]+)\s*\)`)
	matches := m.FindAllStringSubmatch(body, -1)
	if len(matches) < 2 {
		t.Fatalf("tokens.css: expected ≥2 --accent: oklch() declarations (light + dark), got %d", len(matches))
	}
	for _, mm := range matches {
		hue, _ := strconv.ParseFloat(mm[3], 64)
		if hue < 60 || hue > 95 {
			t.Errorf("--accent oklch hue = %.1f° — expected gold (60-95°), got %s", hue, mm[0])
		}
	}

	// The panel-bg token must also be oklch (panel tinting).
	if !strings.Contains(body, "--panel-bg:") {
		t.Error("tokens.css missing --panel-bg token")
	}
	if !strings.Contains(body, "--panel-border:") {
		t.Error("tokens.css missing --panel-border token")
	}
}

func TestSidePanelHeader(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/surah/2", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status = %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `<aside class="surah-side-panel"`) {
		t.Error("surah page missing side-panel aside")
	}
	if !strings.Contains(body, `class="ssp-header"`) {
		t.Error("side panel missing ssp-header — the gold accent label that anchors the panel to its purpose")
	}
	if !strings.Contains(body, `class="ssp-header-label"`) {
		t.Error("side panel header missing ssp-header-label")
	}
	if !strings.Contains(body, `>Word analysis<`) {
		t.Error("side panel header label text is not 'Word analysis'")
	}
}

func TestSidePanelHighlightStyling(t *testing.T) {
	srv := testServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/css/quran.css", nil)
	srv.Handler().ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("quran.css status = %d", w.Code)
	}
	body := w.Body.String()
	// Side panel must declare its distinguishing visual treatment
	// via the panel-bg / panel-border tokens. If someone removes
	// the gold-tinted background or the gold border, the panel
	// will blend back into the main column and this test fires.
	for _, sel := range []string{
		".surah-side-panel",
		"var(--panel-bg)",
		"var(--panel-border)",
		"border-inline-start",
		"box-shadow",
	} {
		if !strings.Contains(body, sel) {
			t.Errorf("quran.css side-panel missing %q — the panel won't read as highlighted", sel)
		}
	}
}
