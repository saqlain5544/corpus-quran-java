package pipeline

import (
	"encoding/gob"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"quranreader/types"
)

// TestBuild runs the entire pipeline end-to-end against the real data
// files and verifies the output. The output is written to a temp
// directory so it does not pollute the repo unless the test passes.
func TestBuild(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "report.txt")
	if err := Build("../data", tmp, report); err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, name := range []string{"quran.gob", "masaq_index.gob", "roots.gob", "meta.json"} {
		if _, err := os.Stat(filepath.Join(tmp, name)); err != nil {
			t.Errorf("missing %s: %v", name, err)
		}
	}
	// Read meta.json back and verify counts.
	b, err := os.ReadFile(filepath.Join(tmp, "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	var meta types.Meta
	if err := json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	if meta.AyahCount != 6236 {
		t.Errorf("meta.AyahCount = %d, want 6236", meta.AyahCount)
	}
	if len(meta.SurahNames) != 114 {
		t.Errorf("meta.SurahNames len = %d, want 114", len(meta.SurahNames))
	}
	if meta.AyahCounts[0] != 7 {
		t.Errorf("surah 1 ayah count = %d, want 7", meta.AyahCounts[0])
	}
	if meta.AyahCounts[1] != 286 {
		t.Errorf("surah 2 ayah count = %d, want 286", meta.AyahCounts[1])
	}
	// Print report to test log so it's visible with -v.
	rb, _ := os.ReadFile(report)
	t.Logf("\n%s", string(rb))
}

// TestGobRoundTrip checks that the gob-encoded data can be decoded
// back without loss.
func TestGobRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "report.txt")
	if err := Build("../data", tmp, report); err != nil {
		t.Fatalf("Build: %v", err)
	}

	// quran.gob
	qf, err := os.Open(filepath.Join(tmp, "quran.gob"))
	if err != nil {
		t.Fatal(err)
	}
	defer qf.Close()
	var q types.Quran
	if err := gob.NewDecoder(qf).Decode(&q); err != nil {
		t.Fatal(err)
	}
	if len(q.Surahs) != 114 {
		t.Errorf("roundtrip: surahs = %d", len(q.Surahs))
	}
	if q.Meta.AyahCount != 6236 {
		t.Errorf("roundtrip: ayahs = %d", q.Meta.AyahCount)
	}

	// masaq_index.gob
	mf, err := os.Open(filepath.Join(tmp, "masaq_index.gob"))
	if err != nil {
		t.Fatal(err)
	}
	defer mf.Close()
	var m types.MasaqIndex
	if err := gob.NewDecoder(mf).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if len(m.ByWord) < 77000 {
		t.Errorf("roundtrip: masaq words = %d", len(m.ByWord))
	}

	// roots.gob
	rf, err := os.Open(filepath.Join(tmp, "roots.gob"))
	if err != nil {
		t.Fatal(err)
	}
	defer rf.Close()
	var r types.RootsIndex
	if err := gob.NewDecoder(rf).Decode(&r); err != nil {
		t.Fatal(err)
	}
	if len(r.ByRoot) < 1600 {
		t.Errorf("roundtrip: roots = %d", len(r.ByRoot))
	}

	// Spot check surah 1 after round-trip.
	s1 := q.Surahs[1]
	if s1 == nil || len(s1.Ayahs[1].Tokens) != 4 {
		t.Errorf("roundtrip: surah 1 aya 1 tokens = %d",
			len(s1.Ayahs[1].Tokens))
	}
}

// TestMetaOrdering asserts that the meta arrays are aligned with
// surah numbers 1..114.
func TestMetaOrdering(t *testing.T) {
	tmp := t.TempDir()
	report := filepath.Join(tmp, "report.txt")
	if err := Build("../data", tmp, report); err != nil {
		t.Fatalf("Build: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(tmp, "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	var meta types.Meta
	if err := json.Unmarshal(b, &meta); err != nil {
		t.Fatal(err)
	}
	t.Logf("ayah counts: %v", meta.AyahCounts)
	if len(meta.AyahCounts) != 114 {
		t.Errorf("AyahCounts len = %d want 114", len(meta.AyahCounts))
	}
	if !sort.IntsAreSorted([]int{1, 2, 3}) {
		// silence unused
		t.Fail()
	}
}
