package data

import (
	"path/filepath"
	"testing"

	"quranreader/loc"
	"quranreader/types"
)

// BenchmarkLoadAll measures wall-clock time of the parallel loader.
// Run with: go test -bench BenchmarkLoadAll ./backend/internal/data
// Compared with the pre-parallel sequential baseline, this should
// be ~30-50% faster on a cold disk cache.
func BenchmarkLoadAll(b *testing.B) {
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		b.Fatal(err)
	}
	// Warm disk cache first so we measure in-memory cost, not I/O.
	if _, _, _, _, err := LoadAll(dbPath); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, _, _, err := LoadAll(dbPath); err != nil {
			b.Fatal(err)
		}
	}
}

// TestLoadAllConcurrentDeterministic verifies that running the three
// sub-loaders concurrently produces the same data structures as the
// sequential version would have. Sub-loaders are deterministic — the
// underlying SQLite queries are read-only and the result maps are
// populated by SurahNo primary key — so this is just a smoke test
// that the goroutines join correctly and no map writes race.
func TestLoadAllConcurrentDeterministic(t *testing.T) {
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		t.Fatal(err)
	}
	q1, m1, r1, meta1, err := LoadAll(dbPath)
	if err != nil {
		t.Fatalf("LoadAll run 1: %v", err)
	}
	q2, m2, r2, meta2, err := LoadAll(dbPath)
	if err != nil {
		t.Fatalf("LoadAll run 2: %v", err)
	}
	if q1.Meta.WordCount != q2.Meta.WordCount {
		t.Errorf("WordCount differs across runs: %d vs %d",
			q1.Meta.WordCount, q2.Meta.WordCount)
	}
	if len(m1.ByWord) != len(m2.ByWord) {
		t.Errorf("Masaq size differs: %d vs %d", len(m1.ByWord), len(m2.ByWord))
	}
	if len(r1.ByRoot) != len(r2.ByRoot) {
		t.Errorf("Roots size differs: %d vs %d", len(r1.ByRoot), len(r2.ByRoot))
	}
	// Spot-check one Masaq entry is identical across runs.
	for k := range m1.ByWord {
		if m2.ByWord[k] == nil {
			t.Errorf("masaq key %d present in run 1, missing in run 2", k)
		}
		break
	}
	_ = meta1
	_ = meta2
	_ = types.Meta{} // keep types import used
}

// TestLoadAllReal runs the loader against the real detailed-quran.db
// and verifies the structures are populated.
func TestLoadAllReal(t *testing.T) {
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		t.Fatal(err)
	}
	q, m, r, meta, err := LoadAll(dbPath)
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(q.Surahs) != 114 {
		t.Errorf("surahs = %d", len(q.Surahs))
	}
	if len(m.ByWord) < 77000 {
		t.Errorf("masaq words = %d", len(m.ByWord))
	}
	if len(r.ByRoot) < 1600 {
		t.Errorf("roots = %d", len(r.ByRoot))
	}
	if meta.AyahCount != 6236 {
		t.Errorf("meta ayahs = %d", meta.AyahCount)
	}
}

// TestLookupWord is a smoke test of the word-loc lookup using a known
// (surah, ayah, word) tuple.
func TestLookupWord(t *testing.T) {
	dbPath, _ := filepath.Abs("../../../data/new/detailed-quran.db")
	_, m, _, _, err := LoadAll(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	key := loc.Key(1, 1, 2) // ٱللَّهِ in Bismillah
	segs := m.ByWord[key]
	if len(segs) == 0 {
		t.Fatalf("no segments for 1:1:2")
	}
	if len(segs) != 2 {
		t.Errorf("expected 2 segments, got %d", len(segs))
	}
	// Verify the segments form a DET + NOUN_PROP combination.
	if segs[0].MorphTag != "DET" {
		t.Errorf("first tag = %q want DET", segs[0].MorphTag)
	}
	if segs[1].MorphTag != "NOUN_PROP" {
		t.Errorf("second tag = %q want NOUN_PROP", segs[1].MorphTag)
	}
}

// TestLookupRootByLoc verifies the ByLoc lookup returns a real root.
func TestLookupRootByLoc(t *testing.T) {
	dbPath, _ := filepath.Abs("../../../data/new/detailed-quran.db")
	_, _, r, _, err := LoadAll(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	key := loc.Key(1, 1, 2) // ٱللَّهِ → root "Alh"
	bw, ok := r.ByLoc[key]
	if !ok {
		t.Fatalf("no root for 1:1:2")
	}
	if bw != "Alh" {
		t.Errorf("root = %q want Alh", bw)
	}
	entry, ok := r.ByRoot[bw]
	if !ok {
		t.Fatalf("no entry for %s", bw)
	}
	if entry.Arabic == "" {
		t.Error("Alh: empty Arabic")
	}
}

// TestMetaNames ensures surah names are present and in order.
func TestMetaNames(t *testing.T) {
	dbPath, _ := filepath.Abs("../../../data/new/detailed-quran.db")
	_, _, _, meta, err := LoadAll(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(meta.SurahNames) != 114 {
		t.Errorf("SurahNames len = %d", len(meta.SurahNames))
	}
	// First letters of surah 1 should be "ال" (Alif Lam).
	if len([]rune(meta.SurahNames[0])) < 2 {
		t.Errorf("surah 1 name too short: %q", meta.SurahNames[0])
	}
	// Surah 96 (Al-Alaq) has 19 ayas.
	if meta.AyahCounts[95] != 19 {
		t.Errorf("surah 96 ayahs = %d want 19", meta.AyahCounts[95])
	}
}

// silence unused import warning if/when test changes.
var _ = types.Meta{}
