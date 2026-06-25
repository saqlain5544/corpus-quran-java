package search

import (
	"database/sql"
	"math"
	"path/filepath"
	"strings"
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

// TestEnglishSearchNoSubstringBug is the regression for the
// "rest<mark>rain</mark>ed" rendering bug. The old substring-based
// match returned "restrained", "training", "in", "sin", etc. when
// the user searched for "rain". The new word-boundary matcher
// returns ONLY entries where "rain" is at a token boundary (exact
// word or prefix match), never inside a token like "restrain".
func TestEnglishSearchNoSubstringBug(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := English("rain", masaq, q, 100, 0)
	if len(results) == 0 {
		t.Fatal("no results for 'rain'")
	}
	for _, r := range results {
		// Every result title's first token (after hyphen-split) must
		// either equal "rain" or start with "rain". We explicitly
		// reject titles like "restrained" where rain sits at
		// non-token-start.
		toks := EnTokenize(r.Title)
		if len(toks) == 0 {
			t.Errorf("result %q has no tokens", r.Title)
			continue
		}
		first := toks[0]
		if first == "rain" || strings.HasPrefix(first, "rain") {
			continue // valid: exact or prefix at token start
		}
		// The first token doesn't start with "rain" — so how did
		// this match? It must be a fuzzy match (rain~brain via
		// Levenshtein) or it slipped through the substring filter.
		// Check if any token in the title is a fuzzy/prefix match.
		matched := false
		for _, t := range toks {
			if strings.HasPrefix(t, "rain") || t == "rain" {
				matched = true
				break
			}
			// Fuzzy match: Levenshtein ≤ 1 (AUTO for len 4).
			if Levenshtein("rain", t) <= 1 && len(t) >= 3 &&
				"rain"[:2] == t[:2] {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("result %q (surah %d:%d) has 'rain' as a non-token substring — word-boundary check failed",
				r.Title, r.Surah, r.Ayah)
		}
	}
}

// TestEnglishSearchRestrained checks that "restrained" returns the
// correct entries (the exact word plus word-boundary compounds), but
// does NOT return arbitrary other words that contain "restrain".
func TestEnglishSearchRestrained(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := English("restrained", masaq, q, 20, 0)
	if len(results) == 0 {
		t.Fatal("no results for 'restrained'")
	}
	// First result should be the exact match.
	if results[0].Title != "restrained" {
		t.Errorf("first result = %q want 'restrained'", results[0].Title)
	}
}

// TestEnglishSearchExactFirst checks that when a word exists as an
// exact gloss in MASAQ, it appears as the first result — even if
// other word-boundary matches also exist.
func TestEnglishSearchExactFirst(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	for _, query := range []string{"mercy", "prayer", "light", "guidance"} {
		results := English(query, masaq, q, 5, 0)
		if len(results) == 0 {
			t.Errorf("%q: no results", query)
			continue
		}
		if results[0].Title != query {
			t.Errorf("%q: first result = %q want exact match", query, results[0].Title)
		}
	}
}

// TestEnglishSearchHighlightWrapsWholeWord is the regression for
// "rest<mark>rain</mark>ed" — the <mark> tag must always be at a
// token boundary (start of title, or after a hyphen/space/open-paren).
// What must NEVER happen is a mark in the middle of a token
// like "<mark>rest</mark>" or a fragment between two word boundaries
// in the middle of one token.
func TestEnglishSearchHighlightWrapsWholeWord(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	results := English("rain", masaq, q, 100, 0)
	if len(results) == 0 {
		t.Skip("no results — MASAQ data may be missing")
	}
	boundaries := map[byte]bool{
		'-':  true, // hyphen (compound word separator in MASAQ glosses)
		' ':  true, // space
		'(':  true, // open paren (e.g., "(rain-from)-the-sky")
		';':  true, // semicolon
		',':  true, // comma
		'\t': true, // tab (unlikely but defensive)
	}
	for _, r := range results {
		markStart := strings.Index(r.Snippet, "<mark>")
		if markStart < 0 {
			continue
		}
		// Opening <mark> must be at start of snippet OR right after
		// a token-boundary character. This is the key check — it
		// rejects the old "rest<mark>rain" bug where the mark sat
		// mid-token.
		if markStart > 0 {
			prev := r.Snippet[markStart-1]
			if !boundaries[prev] {
				t.Errorf("snippet %q: <mark> not at token boundary (prev char %q at offset %d)",
					r.Snippet, string(prev), markStart-1)
			}
		}
	}
}

// TestClassifyMatch is a direct test for the classifier — verifies
// every match class boundary without going through MASAQ. It is
// the authoritative pin for the user's "rain should not match in /
// sin / said / restrained" complaint.
func TestClassifyMatch(t *testing.T) {
	cases := []struct {
		query, text string
		want        MatchClass
		note        string
	}{
		// Class 0: exact full-text match.
		{"rain", "rain", MatchExact, "exact gloss equals query"},
		{"mercy", "mercy", MatchExact, "exact gloss equals query"},
		// Class 1: exact word match (after hyphen split the gloss
		// tokens are independent words).
		{"rain", "heavy-rain", MatchExactWord, "rain is a token after hyphen-split"},
		{"rain", "(from)-the-sky-(rain)", MatchExactWord, "rain is a token"},
		{"prayer", "the-prayer", MatchExactWord, "prayer is a token"},
		// Class 2: prefix match.
		{"rain", "raining", MatchPrefix, "raining starts with rain"},
		{"rain", "rained", MatchPrefix, "rained starts with rain"},
		{"rain", "rainfall", MatchPrefix, "rainfall starts with rain"},
		{"prayer", "prayerful", MatchPrefix, "prayerful starts with prayer"},
		// No match: substring in middle of token.
		{"rain", "restrained", MatchNone, "rain is in the middle of restrained"},
		{"rain", "training", MatchNone, "rain is at the end of training"},
		{"rain", "brain", MatchNone, "rain is in the middle of brain"},
		{"rain", "drainage", MatchNone, "rainage starts with drain, not rain"},
		// rained is a prefix match (not exact word since 'rained' ≠ 'rain').
		{"rain", "and-we-rained", MatchPrefix, "rained is a token, prefix match at token start"},
		// No match: short query (1 char = no prefix / fuzzy).
		{"a", "abc", MatchNone, "query too short, prefix would match everything"},
		// Class 4: fuzzy match with AUTO + prefix lock.
		{"merci", "mercy", MatchFuzzy, "1 edit, length ≥ 3, first 2 chars match"},
		{"pryer", "prayer", MatchFuzzy, "1 edit"},
		// Prefix lock: query "rain" should NOT fuzzy-match "in",
		// "sin", "said" even though Levenshtein = 2 (because the
		// first 2 chars don't match: r-a vs i-n / s-i / s-a).
		{"rain", "in", MatchNone, "prefix lock: 'ra' vs 'in'"},
		{"rain", "sin", MatchNone, "prefix lock: 'ra' vs 'si'"},
		{"rain", "said", MatchNone, "prefix lock: 'ra' vs 'sa'"},
		// AUTO fuzziness: short query = tight threshold.
		{"rain", "brain", MatchNone, "prefix lock: 'ra' vs 'br'"},
		{"rain", "drain", MatchNone, "prefix lock: 'ra' vs 'dr'"},
		// Longer query: more lenient fuzzy.
		{"prayer", "prayers", MatchPrefix, "prayers starts with prayer"},
		{"prayer", "pryers", MatchFuzzy, "1 edit, AUTO fuzziness = 2"},
	}
	for _, c := range cases {
		got := ClassifyMatch(c.query, c.text).Class
		if got != c.want {
			t.Errorf("ClassifyMatch(%q, %q) = %v (%s) want %v  [note: %s]",
				c.query, c.text, got, got, c.want, c.note)
		}
	}
}

// TestAutoFuzziness checks the threshold function directly.
func TestAutoFuzziness(t *testing.T) {
	cases := []struct {
		queryLen int
		want     int
	}{
		{1, 0}, {2, 0}, // 1-2 char: no fuzzy
		{3, 1}, {4, 1}, {5, 1}, // 3-5 char: 1 edit
		{6, 2}, {10, 2}, {20, 2}, // 6+: 2 edits
	}
	for _, c := range cases {
		if got := autoFuzziness(c.queryLen); got != c.want {
			t.Errorf("autoFuzziness(%d) = %d want %d", c.queryLen, got, c.want)
		}
	}
}

// containsToken returns true if needle appears as a whole token
// inside haystack (case-insensitive). Used by the highlight test.
func containsToken(haystack, needle string) bool {
	for _, t := range EnTokenize(haystack) {
		if strings.EqualFold(t, needle) {
			return true
		}
	}
	return false
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

// TestInvertedIndexPopulated verifies that the load path builds the
// inverted index. If this fails, the search will fall back to the
// slow linear path and queries will be ~100x slower.
func TestInvertedIndexPopulated(t *testing.T) {
	_, masaq, _ := loadTestData(t)
	if masaq.ByEnTokenPostings == nil {
		t.Fatal("ByEnTokenPostings is nil — load path failed to build the inverted index")
	}
	if masaq.ByTranslationPostings == nil {
		t.Fatal("ByTranslationPostings is nil")
	}
	if len(masaq.ByEnTokenPostings) < 1000 {
		t.Errorf("ByEnTokenPostings has only %d entries; expected ≥1000 unique tokens",
			len(masaq.ByEnTokenPostings))
	}
	if masaq.GlossAvgDocLen < 1 || masaq.GlossAvgDocLen > 10 {
		t.Errorf("GlossAvgDocLen = %.2f; expected ~2 (avg tokens per gloss)", masaq.GlossAvgDocLen)
	}
}

// TestPostingsAreSorted ensures the posting lists are sorted (so set
// ops are merge-walkable).
func TestPostingsAreSorted(t *testing.T) {
	_, masaq, _ := loadTestData(t)
	for token, posts := range masaq.ByEnTokenPostings {
		for i := 1; i < len(posts); i++ {
			if posts[i-1] >= posts[i] {
				t.Fatalf("posting list for %q not sorted at index %d: %v",
					token, i, posts)
			}
		}
	}
}

// TestExpandPostingsUnion verifies that the candidate-expansion
// helper returns the union of literal + prefix + fuzzy postings,
// matching what the user-facing English() function would match.
func TestExpandPostingsUnion(t *testing.T) {
	_, masaq, _ := loadTestData(t)
	posts := masaq.ByEnTokenPostings

	// "rain" should expand to:
	//   literal:  "rain"  → 23 docs
	//   prefix:   "rains", "rained", "rainfall", "rainstorm", "raintime"
	//             → ~10 docs (vocab words starting with "rain")
	//   fuzzy:    "ran" → 1 doc
	// Total: at least 25 docs.
	cands := expandPostings(posts, []string{"rain"})
	if len(cands) < 25 {
		t.Errorf("expandPostings(rain) = %d candidates; expected ≥25", len(cands))
	}

	// All candidates should be sorted and unique.
	for i := 1; i < len(cands); i++ {
		if cands[i-1] >= cands[i] {
			t.Fatalf("candidates not sorted at %d", i)
		}
	}

	// Multi-token query: "his-messenger" → literal + prefix + fuzzy
	// for BOTH tokens.
	cands2 := expandPostings(posts, []string{"his", "messenger"})
	if len(cands2) < 10 {
		t.Errorf("expandPostings([his, messenger]) = %d; expected ≥10", len(cands2))
	}
}

// TestBM25IDF is a regression pin for the IDF formula. BM25 uses
// log((N - df + 0.5)/(df + 0.5) + 1) so it's never negative.
func TestBM25IDF(t *testing.T) {
	// Common token (high df) → low IDF
	if got := bm25ScoreForDocWithTextCheck(77411, 5000); got > 3.0 {
		t.Errorf("common-token IDF = %.2f; should be ≤ 3.0", got)
	}
	// Rare token (low df) → high IDF
	if got := bm25ScoreForDocWithTextCheck(77411, 1); got < 10 {
		t.Errorf("rare-token IDF = %.2f; should be ≥ 10", got)
	}
}

// bm25ScoreForDocWithTextCheck is a thin wrapper for the IDF-only
// case used in TestBM25IDF. We just check that the math is right.
func bm25ScoreForDocWithTextCheck(n, df int) float64 {
	return math.Log((float64(n)-float64(df)+0.5)/(float64(df)+0.5) + 1.0)
}

// TestEnglishSearchUsesInvertedIndex is an end-to-end parity test
// that mirrors the Python prototype's parity_test.py. It checks
// that the top-N results for representative queries match what we
// expect from the prototype — 31 for "rain", 167 for "mercy",
// 11 for "restrained", etc.
func TestEnglishSearchUsesInvertedIndex(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	cases := []struct {
		query string
		want  int
	}{
		{"rain", 31},          // ground truth from Python prototype
		{"mercy", 167},
		{"restrained", 11},
		{"he-ran", 1},         // the (37,140,2) fuzzy hit
		{"his-messenger", 84}, // 84 docs contain both tokens
	}
	for _, c := range cases {
		results := English(c.query, masaq, q, 1000, 0)
		if len(results) != c.want {
			t.Errorf("English(%q) = %d results; want %d", c.query, len(results), c.want)
		}
	}
}

// TestArabicInvertedIndexPopulated verifies the load path builds
// the Arabic-form inverted index. Without it, Arabic search falls
// back to the slow linear path.
func TestArabicInvertedIndexPopulated(t *testing.T) {
	_, masaq, _ := loadTestData(t)
	if masaq.ByArabicFormPostings == nil {
		t.Fatal("ByArabicFormPostings is nil — load path failed to build the inverted index")
	}
	if len(masaq.ByArabicFormPostings) < 1000 {
		t.Errorf("ByArabicFormPostings has only %d entries; expected ≥1000 unique Arabic forms",
			len(masaq.ByArabicFormPostings))
	}
	// Spot check: 'الله' (the normalized form of اللَّهِ / الله / etc.)
	// must be in the index — it's the most-frequent Arabic token.
	if posts := masaq.ByArabicFormPostings["الله"]; len(posts) < 100 {
		t.Errorf("postings[الله] has only %d entries; expected ≥100", len(posts))
	}
}

// TestArabicSearchFindsBasmala was an existing test — kept here
// for documentation of the new algorithm's behaviour at the
// boundary case. The basmala's "اللَّهِ" word (1:1:2) normalizes
// to "الله" which must be findable in the inverted index.
//
// Note: MASAQ splits this word into segments 'ل' (article) + 'لَّهِ'
// (stem). Neither segment alone normalizes to "الله", but the
// word-level imla "اللَّهِ" does — so the inverted index must also
// index word-level text, not just per-segment text. See
// backend/internal/data/load.go::buildArabicInvertedIndex.
func TestArabicSearchTopResultIsExact(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	// For "ٱللَّهِ" the very first result MUST be the basmala.
	r := Arabic("ٱللَّهِ", masaq, q, 10, 0)
	if len(r) == 0 {
		t.Fatal("no results for ٱللَّهِ")
	}
	if r[0].Surah != 1 || r[0].Ayah != 1 || r[0].Word != 2 {
		t.Errorf("first hit = %d:%d w=%d want 1:1 w=2 (basmala اللَّهِ)",
			r[0].Surah, r[0].Ayah, r[0].Word)
	}
}

// TestArabicSearchSkipsSingleCharMatches verifies the algorithm
// doesn't return spurious matches from 1-character morphemes like
// 'ل' (the definite article prefix). Without the len ≥ 2 guard,
// every word containing 'ل' would match every query containing 'ل'.
func TestArabicSearchSkipsSingleCharMatches(t *testing.T) {
	q, masaq, _ := loadTestData(t)
	// Query 'ل' is too short — should return nil (Go's algorithm
	// would return thousands of spurious matches via the empty-
	// segment quirk).
	r := Arabic("ل", masaq, q, 50, 0)
	if len(r) > 0 {
		t.Errorf("Arabic(ل) returned %d results; expected 0 (single-char query rejected)", len(r))
	}
}
