package pipeline

import (
	"strings"
	"testing"

	"quranreader/loc"
)

// TestRowToSegment checks the conversion of one CSV row into a
// MasaqSegment with sample values drawn from the actual corpus.
func TestRowToSegment(t *testing.T) {
	row := []string{
		"1", "1", "1", "1", "1", "بِسْمِ", "بسم", "ب",
		"PREP", "Prefix", "", "INVAR", "PREP",
		"NOT_CONSTRUCT", "INVARIABLE", "KASRA", "PHRASE", "PRED",
		"in-(the)-name",
	}
	seg, err := rowToSegment(row)
	if err != nil {
		t.Fatal(err)
	}
	if seg.ID != 1 || seg.SuraNo != 1 || seg.VerseNo != 1 ||
		seg.WordNo != 1 || seg.SegmentNo != 1 {
		t.Errorf("ids wrong: %+v", seg)
	}
	if seg.Word != "بِسْمِ" {
		t.Errorf("Word = %q", seg.Word)
	}
	if seg.Gloss != "in-(the)-name" {
		t.Errorf("Gloss = %q", seg.Gloss)
	}
}

// TestLoadMasaqReal runs the parser against the real MASAQ.csv and
// reports the totals.
func TestLoadMasaqReal(t *testing.T) {
	const path = "../data/morphology/MASAQ.csv"
	idx, err := LoadMasaq(path)
	if err != nil {
		t.Fatalf("LoadMasaq: %v", err)
	}
	t.Logf("distinct words in MasaqIndex: %d", len(idx.ByWord))
	if len(idx.ByWord) < 77000 || len(idx.ByWord) > 79000 {
		t.Errorf("word count out of expected range: %d", len(idx.ByWord))
	}
	// Spot check: 1:1:1 should have 2 segments (the prefix and the stem).
	got := idx.ByWord[loc.Key(1, 1, 1)]
	if len(got) != 2 {
		t.Errorf("1:1:1 segments = %d want 2", len(got))
	}
	if got[0].MorphType != "Prefix" || got[1].MorphType != "Stem" {
		t.Errorf("1:1:1 types = (%q, %q) want (Prefix, Stem)",
			got[0].MorphType, got[1].MorphType)
	}
}

// TestMasaqSpotCheck verifies a handful of well-known words.
func TestMasaqSpotCheck(t *testing.T) {
	idx, err := LoadMasaq("../data/morphology/MASAQ.csv")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		key             uint64
		wantAnyGlossSub string // substr of any segment's gloss
		wantFirstTag    string // morph tag of the first segment
	}{
		{loc.Key(2, 255, 1), "allah", "DET"},  // ٱللَّهُ (DET + NOUN_PROP)
		{loc.Key(1, 1, 2), "allah", "DET"},    // ٱللَّهِ in Bismillah
		{loc.Key(1, 1, 3), "gracious", "DET"}, // ٱلرَّحْمَـٰنِ
	}
	for _, c := range cases {
		segs, ok := idx.ByWord[c.key]
		if !ok || len(segs) == 0 {
			t.Errorf("no entry for key %s", loc.String(c.key))
			continue
		}
		if segs[0].MorphTag != c.wantFirstTag {
			t.Errorf("key %s: first tag = %q want %q", loc.String(c.key), segs[0].MorphTag, c.wantFirstTag)
		}
		found := false
		for _, s := range segs {
			if strings.Contains(strings.ToLower(s.Gloss), c.wantAnyGlossSub) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("key %s: no gloss contains %q in %+v", loc.String(c.key), c.wantAnyGlossSub, segs)
		}
	}
}
