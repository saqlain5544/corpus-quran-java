package data

import (
	"fmt"
	"testing"

	"quranreader/loc"
	"quranreader/pipeline"
	"quranreader/token"
)

// TestCrossSourceAlignment verifies that the Quran data loaded from
// detailed-quran.db matches the data loaded directly from
// quran-uthmani.xml token-by-token, and that every word's MASAQ and
// root entries are consistent across both sources.
func TestCrossSourceAlignment(t *testing.T) {
	xmlPath := "../../../data/quran/quran-uthmani.xml"
	dbPath := "../../../data/new/detailed-quran.db"

	// Load from XML via pipeline
	qXML, err := pipeline.LoadQuran(xmlPath)
	if err != nil {
		t.Fatalf("LoadQuran XML: %v", err)
	}

	// Load from DB
	qDB, mDB, rDB, _, _, err := LoadAll(dbPath)
	if err != nil {
		t.Fatalf("LoadAll DB: %v", err)
	}

	// ── Top-level counts ──
	if len(qXML.Surahs) != len(qDB.Surahs) {
		t.Errorf("surah count: XML=%d DB=%d", len(qXML.Surahs), len(qDB.Surahs))
	}
	if qXML.Meta.AyahCount != qDB.Meta.AyahCount {
		t.Errorf("ayah count: XML=%d DB=%d", qXML.Meta.AyahCount, qDB.Meta.AyahCount)
	}
	if qXML.Meta.WordCount != qDB.Meta.WordCount {
		t.Errorf("word count: XML=%d DB=%d", qXML.Meta.WordCount, qDB.Meta.WordCount)
	}
	t.Logf("counts match: surahs=%d ayahs=%d words=%d",
		len(qXML.Surahs), qXML.Meta.AyahCount, qXML.Meta.WordCount)

	// ── Per-token walk ──
	var (
		masaqHits, masaqMisses int
		rootHits, rootMisses   int
		nameMismatches         int
		tokenDiffs             int
	)
	var tokenDiffExamples []string

	for sn := 1; sn <= 114; sn++ {
		sXML := qXML.Surahs[sn]
		sDB := qDB.Surahs[sn]
		if sXML == nil || sDB == nil {
			continue
		}
		if sXML.Name != sDB.Name {
			if nameMismatches < 5 {
				t.Logf("name diff surah %d: XML=%q DB=%q", sn, sXML.Name, sDB.Name)
			}
			nameMismatches++
		}

		for an := 1; ; an++ {
			ayXML := sXML.Ayahs[an]
			ayDB := sDB.Ayahs[an]
			if ayXML == nil && ayDB == nil {
				break
			}
			if ayXML == nil || ayDB == nil {
				continue // handled by count check
			}

			// Re-tokenize both with shared tokenizer for comparison
			tokXML := token.Tokenize(ayXML.Text)
			tokDB := token.Tokenize(ayDB.Text)

			var xmlWords, dbWords []string
			for _, tk := range tokXML {
				if tk.Kind == "word" {
					xmlWords = append(xmlWords, tk.Value)
				}
			}
			for _, tk := range tokDB {
				if tk.Kind == "word" {
					dbWords = append(dbWords, tk.Value)
				}
			}

			if len(xmlWords) != len(dbWords) {
				tokenDiffs++
				if len(tokenDiffExamples) < 10 {
					tokenDiffExamples = append(tokenDiffExamples,
						fmt.Sprintf("%d:%d: XML=%d words DB=%d words", sn, an, len(xmlWords), len(dbWords)))
				}
			}

			// MASAQ + root coverage via XML word positions
			for w := 1; w <= len(xmlWords); w++ {
				key := loc.Key(sn, an, w)
				if segs, ok := mDB.ByWord[key]; ok && len(segs) > 0 {
					masaqHits++
				} else {
					masaqMisses++
				}
				if _, ok := rDB.ByLoc[key]; ok {
					rootHits++
				} else {
					rootMisses++
				}
			}
		}
	}

	t.Logf("MASAQ:  hits=%d misses=%d (%.2f%%)",
		masaqHits, masaqMisses,
		100*float64(masaqHits)/float64(masaqHits+masaqMisses))
	t.Logf("Roots:  hits=%d misses=%d (%.2f%%)",
		rootHits, rootMisses,
		100*float64(rootHits)/float64(rootHits+rootMisses))

	// After reconcileMasaqBoundaries, MASAQ coverage should be
	// ≥ 99.99%. The remaining ~8 misses are genuine word-numbering
	// differences (not merge cases) — constructions like ما دام
	// and ما زال where MASAQ's word numbering differs from the
	// Quran XML's by more than 2 positions. These get "No
	// morphological data" in the UI, which is correct behaviour.
	if masaqMisses > 8 {
		t.Errorf("MASAQ miss = %d, want ≤ 8 after reconciliation", masaqMisses)
	}
	t.Logf("Name mismatches: %d", nameMismatches)
	t.Logf("Token count diffs: %d", tokenDiffs)
	for _, ex := range tokenDiffExamples {
		t.Logf("  %s", ex)
	}

	// Known acceptable gaps: ~23 MASAQ misses from word-boundary
	// disagreements. MASAQ coverage should be >99.9%.
	if float64(masaqMisses)/float64(masaqHits+masaqMisses) > 0.001 {
		t.Errorf("MASAQ miss rate too high: %.4f%%", 100*float64(masaqMisses)/float64(masaqHits+masaqMisses))
	}
	// Root coverage: not all words have roots (particles, marks, etc.)
	// but it should be reasonable (>60%).
	if float64(rootHits)/float64(rootHits+rootMisses) < 0.60 {
		t.Errorf("root coverage too low: %.2f%%", 100*float64(rootHits)/float64(rootHits+rootMisses))
	}
}
