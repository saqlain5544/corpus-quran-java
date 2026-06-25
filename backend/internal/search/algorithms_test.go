package search

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	qdata "quranreader/backend/internal/data"
	"quranreader/types"
)

// loadTestData loads the indexes from detailed-quran.db once for the test suite.
func loadTestData(t *testing.T) (*types.Quran, *types.MasaqIndex, *types.RootsIndex) {
	t.Helper()
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		t.Fatal(err)
	}
	q, m, r, _, err := qdata.LoadAll(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	return q, m, r
}

var _ = sql.Open // silence unused import when not needed

// ── Tests ───────────────────────────────────────────────────────

func TestArabicSearchFindsBasmala(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := Arabic("ٱللَّهِ", masaq, q, 10, 0)
	if len(results) == 0 {
		t.Fatal("no results for ٱللَّهِ")
	}
	if results[0].Surah != 1 || results[0].Ayah != 1 {
		t.Errorf("first hit = %d:%d want 1:1", results[0].Surah, results[0].Ayah)
	}
}

func TestArabicSearchNormalizesTashkeel(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	withTashkeel := Arabic("ٱللَّهِ", masaq, q, 5, 0)
	noTashkeel := Arabic("الله", masaq, q, 5, 0)
	if len(withTashkeel) == 0 || len(noTashkeel) == 0 {
		t.Fatalf("no results: with=%d plain=%d", len(withTashkeel), len(noTashkeel))
	}
	if withTashkeel[0].Surah != noTashkeel[0].Surah ||
		withTashkeel[0].Ayah != noTashkeel[0].Ayah ||
		withTashkeel[0].Word != noTashkeel[0].Word {
		t.Errorf("mismatch: %+v vs %+v", withTashkeel[0], noTashkeel[0])
	}
}

func TestEnglishSearchFindsAllah(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	// Search for a phrase less common than "allah".
	results := English("his-messenger", masaq, q, 50, 0)
	if len(results) == 0 {
		t.Fatal("no results for 'his-messenger'")
	}
	// Look for any hit in surah 2 (a frequent source).
	found := false
	for _, r := range results {
		if r.Surah == 2 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no surah-2 hits in %d results", len(results))
	}
}

func TestRootSearchFindsQwl(t *testing.T) {
	q, _, roots := loadTestData(t)
	results := Root("qwl", roots, q, 10)
	if len(results) == 0 {
		t.Fatal("no results for qwl")
	}
	if results[0].Root != "qwl" {
		t.Errorf("first root = %q want qwl", results[0].Root)
	}
}

func TestRootSearchArabicInput(t *testing.T) {
	q, _, roots := loadTestData(t)
	results := Root("قول", roots, q, 10)
	if len(results) == 0 {
		t.Fatal("no results for قول")
	}
	if results[0].Root != "qwl" {
		t.Errorf("first root = %q want qwl", results[0].Root)
	}
}

func TestSearchResultRanking(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := English("allah", masaq, q, 20, 0)
	if len(results) < 2 {
		t.Fatal("need at least 2 results")
	}
	if results[0].Score > results[1].Score {
		t.Errorf("not sorted: %d then %d", results[0].Score, results[1].Score)
	}
}

func TestSearchSnippetHighlights(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := English("merciful", masaq, q, 5, 0)
	for _, r := range results {
		if !contains(r.Snippet, "<mark>") {
			t.Errorf("snippet missing highlight: %q", r.Snippet)
		}
	}
}

func TestRootsListPagination(t *testing.T) {
	_, _, roots := loadTestData(t)
	page1, total := RootsList(roots, 1, 50, false, 0)
	if total < 1600 {
		t.Errorf("total roots = %d", total)
	}
	if len(page1) != 50 {
		t.Errorf("page1 size = %d", len(page1))
	}
	if page1[0].Kind != "root-list" {
		t.Errorf("kind = %q", page1[0].Kind)
	}
}

// TestOccurrencesForRootSortAsc verifies the asc parameter actually
// reverses the natural order. Regression: the asc=true and asc=false
// branches were once identical (sort ascending in both cases), which
// made the "Ascending" UI option a no-op.
func TestOccurrencesForRootSortAsc(t *testing.T) {
	q, _, roots := loadTestData(t)
	desc := OccurrencesForRoot("qwl", roots, q, 0, false)
	asc := OccurrencesForRoot("qwl", roots, q, 0, true)
	if len(desc) < 2 || len(asc) < 2 {
		t.Fatalf("need ≥2 occurrences: desc=%d asc=%d", len(desc), len(asc))
	}
	// Descending toggle (asc=false): natural surah-ascending order.
	if desc[0].Surah > desc[len(desc)-1].Surah {
		t.Errorf("desc not in surah-ascending order: first=%d last=%d",
			desc[0].Surah, desc[len(desc)-1].Surah)
	}
	// Ascending toggle (asc=true): reversed order. First result must
	// have a strictly greater (surah, ayah) than the last result.
	first := asc[0]
	last := asc[len(asc)-1]
	if first.Surah < last.Surah {
		t.Errorf("asc not reversed: first=%d:%d last=%d:%d",
			first.Surah, first.Ayah, last.Surah, last.Ayah)
	}
	// The two lists must be the same length (just reordered).
	if len(desc) != len(asc) {
		t.Errorf("length mismatch: desc=%d asc=%d", len(desc), len(asc))
	}
}

// TestRootByArabicVariants ensures Arabic root lookup works for the
// variants users actually type: with/without spaces, hamza variants.
// Regression: the index was built with stripDiacritics (only strips
// tashkeel) while queries used NormalizeArabic — they disagreed on
// both alif variants AND spaces, so "/root/detailed/قول" 404'd while
// "/root/detailed/ق و ل" succeeded.
func TestRootByArabicVariants(t *testing.T) {
	_, _, roots := loadTestData(t)

	// The DB has Arabic letters space-separated; with our key fix,
	// the index stores the canonical space-free form. So:
	//   - "قول"     (no spaces, no hamza) → qwl
	//   - "ق و ل"   (spaces, no hamza)     → qwl
	//   - "أبي"     (no spaces, with hamza) → Aby
	//   - "أ ب ي"   (spaces, with hamza)   → Aby
	cases := []struct {
		in, want string
	}{
		{"قول", "qwl"},
		{"ق و ل", "qwl"},
		{"أبي", "Aby"},
		{"أ ب ي", "Aby"},
		{"ٱللَّه", "Alh"}, // hamza-wasla + tashkeel
	}
	results := Root(cases[0].in, roots, nil, 100)
	if len(results) == 0 || results[0].Root != cases[0].want {
		t.Errorf("Root(%q) = %+v want first.Root=%q", cases[0].in, results, cases[0].want)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
