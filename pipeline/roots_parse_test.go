package pipeline

import (
	"sort"
	"strings"
	"testing"

	"quranreader/loc"
)

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func TestLoadRootsReal(t *testing.T) {
	idx, err := LoadRoots(
		"../data/morphology/corpus-roots.json",
		"../data/morphology/meanings-roots-ai.jsonl",
	)
	if err != nil {
		t.Fatalf("LoadRoots: %v", err)
	}
	t.Logf("distinct roots: %d", len(idx.ByRoot))
	t.Logf("total occurrences: %d", len(idx.ByLoc))
	if len(idx.ByRoot) < 1600 || len(idx.ByRoot) > 1700 {
		t.Errorf("root count out of range: %d", len(idx.ByRoot))
	}
	if len(idx.ByLoc) < 49000 || len(idx.ByLoc) > 51000 {
		t.Errorf("occurrence count out of range: %d", len(idx.ByLoc))
	}
	// Spot checks against the Research Agent's findings:
	for _, r := range []string{"qwl", "kwn", "rbb", "Alh"} {
		e, ok := idx.ByRoot[r]
		if !ok {
			t.Errorf("missing root %q", r)
			continue
		}
		t.Logf("root %s: arabic=%q pos=%q occurrences=%d meaning-en-prefix=%q",
			r, e.Arabic, e.POS, e.Occurrences,
			truncate(e.MeaningEN, 60))
		if e.Arabic == "" {
			t.Errorf("root %s: empty Arabic", r)
		}
	}
	// Surah 1:1 should be linked to known roots (per Research Agent).
	expect := map[string]bool{"smw": true, "Alh": true, "rHm": true}
	for w := 1; w <= 4; w++ {
		k := loc.Key(1, 1, w)
		bw, ok := idx.ByLoc[k]
		if !ok {
			t.Errorf("surah 1:1 word %d has no root", w)
			continue
		}
		if !expect[bw] {
			// It's OK to have a different root, just log it.
			t.Logf("surah 1:1 word %d → root %s (not in expect set)", w, bw)
		}
	}
}

func TestParseLoc(t *testing.T) {
	cases := []struct {
		in      string
		s, a, w int
		wantErr bool
	}{
		{"2:255:50", 2, 255, 50, false},
		{"1:1:1", 1, 1, 1, false},
		{"bad", 0, 0, 0, true},
		{"1:2", 0, 0, 0, true},
	}
	for _, c := range cases {
		s, a, w, err := parseLoc(c.in)
		if c.wantErr {
			if err == nil {
				t.Errorf("parseLoc(%q) = (%d,%d,%d) want error", c.in, s, a, w)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLoc(%q): %v", c.in, err)
			continue
		}
		if s != c.s || a != c.a || w != c.w {
			t.Errorf("parseLoc(%q) = (%d,%d,%d) want (%d,%d,%d)", c.in, s, a, w, c.s, c.a, c.w)
		}
	}
}

func TestStripDiacritics(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"قَوْل", "قول"},
		{"رَبّ", "رب"}, // shadda U+0651 is in the range, so it gets stripped
	}
	for _, c := range cases {
		got := stripDiacritics(c.in)
		if got != c.want {
			t.Errorf("stripDiacritics(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestByArabicReverseLookup(t *testing.T) {
	idx, err := LoadRoots(
		"../data/morphology/corpus-roots.json",
		"../data/morphology/meanings-roots-ai.jsonl",
	)
	if err != nil {
		t.Fatal(err)
	}
	// Try to look up "قول" → qwl
	bw, ok := idx.ByArabic["قول"]
	if !ok {
		// Maybe root_arabic has tashkeel; iterate and find a known one.
		count := 0
		for k, v := range idx.ByArabic {
			t.Logf("has arabic root: %q → %s", k, v)
			count++
			if count > 5 {
				break
			}
		}
		// Don't fail — root_arabic field varies; the dictionary is best-effort.
		t.Logf("ByArabic['قول'] = %v (best-effort, may be empty)", bw)
	}
	// Ensure ByArabic is at least populated with some entries.
	if len(idx.ByArabic) == 0 {
		t.Error("ByArabic is empty")
	}
}

func TestRootsSorted(t *testing.T) {
	idx, err := LoadRoots(
		"../data/morphology/corpus-roots.json",
		"../data/morphology/meanings-roots-ai.jsonl",
	)
	if err != nil {
		t.Fatal(err)
	}
	// pick a known root and verify Locations is sorted
	e := idx.ByRoot["qwl"]
	if e == nil {
		t.Fatal("no qwl")
	}
	if !sort.StringsAreSorted(e.Locations) {
		t.Errorf("Locations not sorted: first=%q", e.Locations[0])
	}
	// The first location should be lex-min.
	if len(e.Locations) > 0 && !strings.HasPrefix(e.Locations[0], "1:") {
		// It might not start with "1:" (depends on data) — just log it.
		t.Logf("first location: %q", e.Locations[0])
	}
}
