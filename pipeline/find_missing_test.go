package pipeline

import (
	"strings"
	"testing"

	"quranreader/loc"
)

// TestFindMissingWords enumerates Quran words that have no MASAQ
// entry and reports them. The output is informational — the test
// always passes; it just shows where the pipeline and MASAQ diverge.
func TestFindMissingWords(t *testing.T) {
	quran, err := LoadQuran("../data/quran/quran-uthmani.xml")
	if err != nil {
		t.Fatal(err)
	}
	masaq, err := LoadMasaq("../data/morphology/MASAQ.csv")
	if err != nil {
		t.Fatal(err)
	}
	var missing []string
	for _, s := range quran.Surahs {
		for _, ay := range s.Ayahs {
			for _, tok := range ay.Tokens {
				if tok.Kind != "word" {
					continue
				}
				k := loc.Key(s.Number, ay.Number, tok.WordNo)
				if _, ok := masaq.ByWord[k]; !ok {
					missing = append(missing, loc.String(k)+" "+tok.Value)
				}
			}
		}
	}
	t.Logf("missing words: %d", len(missing))
	for _, m := range missing {
		t.Logf("  %s", m)
	}
	if len(missing) == 0 {
		t.Log("all Quran words have MASAQ entries")
	}
	// Find common prefix to see if they cluster.
	if len(missing) > 0 {
		prefix := missing[0]
		for _, m := range missing[1:] {
			i := 0
			for i < len(prefix) && i < len(m) && prefix[i] == m[i] {
				i++
			}
			prefix = prefix[:i]
		}
		t.Logf("common prefix: %q", prefix)
		// Count by first word.
		wordCounts := map[string]int{}
		for _, m := range missing {
			parts := strings.SplitN(m, " ", 2)
			if len(parts) == 2 {
				wordCounts[parts[1]]++
			}
		}
		t.Logf("distinct missing words: %d", len(wordCounts))
		for w, c := range wordCounts {
			if c > 1 {
				t.Logf("  %dx %s", c, w)
			}
		}
	}
}
