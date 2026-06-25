package pipeline

import (
	"fmt"
	"strings"
	"testing"

	"quranreader/loc"
)

// ─── Tokenizer ──────────────────────────────────────────────────

func TestTokenizeSimple(t *testing.T) {
	in := "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ"
	got := tokenize(in)
	if len(got) != 4 {
		t.Fatalf("want 4 tokens, got %d: %+v", len(got), got)
	}
	for i, g := range got {
		if g.Kind != "word" {
			t.Errorf("token %d: kind=%q want word", i, g.Kind)
		}
		if g.WordNo != i+1 {
			t.Errorf("token %d: WordNo=%d want %d", i, g.WordNo, i+1)
		}
	}
}

func TestTokenizeTrailingSajdah(t *testing.T) {
	// Last token is "ٱقْتَرِب ۩" — word + sajdah mark.
	in := "كَلَّا لَا تُطِعْهُ وَٱسْجُدْ وَٱقْتَرِب ۩"
	toks := tokenize(in)
	if len(toks) == 0 {
		t.Fatal("empty tokens")
	}
	last := toks[len(toks)-1]
	if last.Kind != "mark" || last.Value != "۩" {
		t.Errorf("last token = %+v, want mark ۩", last)
	}
	// Second-to-last must be the word "وَٱقْتَرِب" with no trailing mark
	// attached to itself.
	secondLast := toks[len(toks)-2]
	if secondLast.Kind != "word" || secondLast.Value != "وَٱقْتَرِب" {
		t.Errorf("second-last = %+v", secondLast)
	}
	if len(secondLast.Marks) != 0 {
		t.Errorf("second-last should have no marks, got %v", secondLast.Marks)
	}
}

func TestTokenizeInlinePause(t *testing.T) {
	// Pause marks appear inline as their own tokens.
	in := "ذَٰلِكَ ٱلْكِتَـٰبُ لَا رَيْبَ ۛ فِيهِ"
	toks := tokenize(in)
	// Expected: word word word mark word
	// (pause mark "ۛ" is a token of its own because it's space-delimited)
	want := []struct {
		kind, value string
	}{
		{"word", "ذَٰلِكَ"},
		{"word", "ٱلْكِتَـٰبُ"},
		{"word", "لَا"},
		{"word", "رَيْبَ"},
		{"mark", "ۛ"},
		{"word", "فِيهِ"},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens want %d: %+v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Kind != w.kind || toks[i].Value != w.value {
			t.Errorf("token %d = %+v want {%s %s}", i, toks[i], w.kind, w.value)
		}
	}
}

func TestTokenizeIqlabHighMeem(t *testing.T) {
	// Tanween kasra followed by ba → small high meem ۢ (U+06E2)
	// attached to the previous word as a trailing mark, not part
	// of the word text itself. The ۢ indicates the noon (of
	// tanween) assimilates to meem before ba.
	//
	// The trailing mark is stored on the WORD token's Marks array,
	// NOT emitted as a separate mark token, so the template renders
	// each iqlab exactly once.
	in := "عَلَيٍّۢ بَصِيرًا"
	toks := tokenize(in)
	if len(toks) != 2 {
		t.Fatalf("expected exactly 2 tokens (word + word), got %d: %+v", len(toks), toks)
	}
	if toks[0].Kind != "word" {
		t.Errorf("token 0 kind = %q want word", toks[0].Kind)
	}
	if contains(toks[0].Value, "\u06E2") {
		t.Errorf("token 0 word value should not contain iqlab mark ۢ: %q", toks[0].Value)
	}
	if len(toks[0].Marks) != 1 || toks[0].Marks[0] != "\u06E2" {
		t.Errorf("token 0 marks = %v want [ۢ]", toks[0].Marks)
	}
	// No separate mark token — the mark lives in the word's Marks array.
	for i, tk := range toks {
		if tk.Kind == "mark" {
			t.Errorf("token %d is an unexpected separate mark token: %+v", i, tk)
		}
	}
}

func TestTokenizeIqlabLowMeem(t *testing.T) {
	// Same as above but for the low-meem variant ۭ (U+06ED).
	in := "ٱلْكِتَـٰبِۭ بِسْمِ"
	toks := tokenize(in)
	if len(toks) != 2 {
		t.Fatalf("expected exactly 2 tokens, got %d: %+v", len(toks), toks)
	}
	if toks[0].Kind != "word" {
		t.Errorf("token 0 kind = %q want word", toks[0].Kind)
	}
	if len(toks[0].Marks) != 1 || toks[0].Marks[0] != "\u06ED" {
		t.Errorf("token 0 marks = %v want [ۭ]", toks[0].Marks)
	}
}

func TestTokenizeStandaloneMark(t *testing.T) {
	// Standalone marks (whitespace-delimited, no word body) still
	// produce their own mark tokens. Example: the rub' ۞ at the
	// start of an ayah, or a space-delimited pause mark ۖ between
	// two words.
	in := "ٱ ۞ بِسْمِ ۖ ٱللَّهِ"
	toks := tokenize(in)
	want := []struct {
		kind, value string
	}{
		{"word", "ٱ"},
		{"mark", "۞"},
		{"word", "بِسْمِ"},
		{"mark", "ۖ"},
		{"word", "ٱللَّهِ"},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens want %d: %+v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Kind != w.kind || toks[i].Value != w.value {
			t.Errorf("token %d = {kind=%q value=%q} want {kind=%q value=%q}",
				i, toks[i].Kind, toks[i].Value, w.kind, w.value)
		}
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

func TestSplitMark(t *testing.T) {
	cases := []struct {
		in       string
		wantWord string
		wantMark string
	}{
		{"ٱقْتَرِب", "ٱقْتَرِب", ""},
	}
	for _, c := range cases {
		w, m := splitMark(c.in)
		if w != c.wantWord {
			t.Errorf("splitMark(%q) word = %q want %q", c.in, w, c.wantWord)
		}
		// Verify the mark set rather than the exact slice to handle
		// edge cases with empty strings cleanly.
		gotMark := ""
		if len(m) > 0 {
			gotMark = m[len(m)-1]
		}
		if gotMark != c.wantMark {
			t.Errorf("splitMark(%q) marks = %v want last %q", c.in, m, c.wantMark)
		}
	}
}

// ─── Bismillah detection ────────────────────────────────────────

func TestIsBismillah(t *testing.T) {
	cases := []string{
		BISMILLAH_TEXT,
		"  " + BISMILLAH_TEXT + "  ",
		"ٱلْحَمْدُ لِلَّهِ", // not bismillah
	}
	for i, c := range cases {
		got := isBismillah(c)
		want := i < 2
		if got != want {
			t.Errorf("isBismillah(%q) = %v want %v", c, got, want)
		}
	}
}

// ─── Full XML parser smoke test ─────────────────────────────────

func TestLoadQuranReal(t *testing.T) {
	const path = "../data/quran/quran-uthmani.xml"
	q, err := LoadQuran(path)
	if err != nil {
		t.Fatalf("LoadQuran: %v", err)
	}
	if len(q.Surahs) != 114 {
		t.Fatalf("surahs = %d, want 114", len(q.Surahs))
	}
	if q.Meta.AyahCount != 6236 {
		t.Errorf("ayah count = %d, want 6236", q.Meta.AyahCount)
	}
	// Surah 1 sanity check.
	s1 := q.Surahs[1]
	if s1 == nil || s1.Number != 1 {
		t.Fatalf("surah 1 = %+v", s1)
	}
	if s1.Bismillah == "" {
		t.Error("surah 1: bismillah empty")
	}
	// Surah 9 sanity: no bismillah.
	if q.Surahs[9].Bismillah != "" {
		t.Errorf("surah 9: bismillah = %q want empty", q.Surahs[9].Bismillah)
	}
	// Surah 2 aya 1 should be the standalone Bismillah-as-verse-1.
	// If the XML stores it as text equal to BISMILLAH_TEXT, our
	// pipeline skips it; check that.
	if ay, ok := q.Surahs[2].Ayahs[1]; ok {
		if strings.TrimSpace(ay.Text) == BISMILLAH_TEXT {
			t.Errorf("surah 2 aya 1 = Bismillah itself; pipeline should have skipped")
		}
	}
	// Total word count should be in the right ballpark (~77k).
	if q.Meta.WordCount < 75000 || q.Meta.WordCount > 80000 {
		t.Errorf("word count = %d out of expected range", q.Meta.WordCount)
	}
	// All words should have unique WordNo per ayah.
	for n, s := range q.Surahs {
		for an, ay := range s.Ayahs {
			seen := map[int]bool{}
			for _, tk := range ay.Tokens {
				if tk.Kind == "word" {
					if seen[tk.WordNo] {
						t.Errorf("surah %d aya %d: duplicate word no %d", n, an, tk.WordNo)
					}
					seen[tk.WordNo] = true
				}
			}
		}
	}
}

// ─── loc key helper used by tokens (no equivalent in token; here for symmetry) ───

func TestLocKeyFromSurah(t *testing.T) {
	// Sanity: loc.Key is also exposed; we use it indirectly when
	// building cross-references later.
	if loc.Key(1, 1, 1) == 0 {
		t.Error("loc.Key should never be 0")
	}
}

// ─── helper string builder for tests ──────────────────────────────

func TestTokenSurah1Aya1(t *testing.T) {
	const path = "../data/quran/quran-uthmani.xml"
	q, err := LoadQuran(path)
	if err != nil {
		t.Fatal(err)
	}
	ay := q.Surahs[1].Ayahs[1]
	toks := ay.Tokens
	if len(toks) != 4 {
		t.Fatalf("surah 1 aya 1 tokens = %d want 4", len(toks))
	}
	// The test source may be saved with a different Unicode
	// normalization than the Quran, so we compare by first-rune and
	// length only.
	wantStarts := []rune{'ب', 'ٱ', 'ٱ', 'ٱ'}
	for i, want := range wantStarts {
		if []rune(toks[i].Value)[0] != want {
			t.Errorf("token %d first rune = %q want %q", i, []rune(toks[i].Value)[0], want)
		}
		if len([]rune(toks[i].Value)) < 3 {
			t.Errorf("token %d too short: %q", i, toks[i].Value)
		}
	}
}

// helper that prints counts; not strictly a test but useful when
// running with -v.
func TestReportCounts(t *testing.T) {
	const path = "../data/quran/quran-uthmani.xml"
	q, err := LoadQuran(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("surahs=%d ayas=%d words=%d",
		len(q.Surahs), q.Meta.AyahCount, q.Meta.WordCount)
	for i, name := range q.Meta.SurahNames {
		t.Logf("surah %d: name=%q ayas=%d", i+1, name, q.Meta.AyahCounts[i])
		if q.Meta.AyahCounts[i] == 0 {
			t.Errorf("surah %d has zero ayas", i+1)
		}
	}
}

// ensures unused imports are caught in some build configurations.
var _ = fmt.Sprintf
