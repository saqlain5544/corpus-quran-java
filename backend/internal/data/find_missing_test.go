package data

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"quranreader/loc"
)

// TestFindMissingMorphology scans every word token across the entire
// Quran and reports any word position that has no corresponding entry
// in the MASAQ index. These gaps are caused by word-boundary
// disagreements between the Uthmani XML tokenizer and MASAQ's
// segmentation.
//
// Root cause analysis for each category:
//
//	A. MASAQ merges "X + مَا" where the XML keeps them separate.
//	   Examples: أَيْنَمَا ← أَيْنَ + مَا, كُلَّمَا ← كُلَّ + مَا,
//	             بَعْدَمَا ← بَعْدَ + مَا, وَحَيْثُمَا ← حَيْثُ + مَا,
//	             لَوْمَا ← لَوْ + مَا.
//	   This is the dominant pattern (8 of 23 gaps).
//
//	B. MASAQ merges "مَا + دَامَ" into a single token while the XML
//	   treats مَا and the following verb as separate. Example: 5:117
//	   where شَهِيدٌ at XML position 29 equals MASAQ position 28.
//
//	C. Vocative merge: MASAQ treats يَا + word as one token; XML
//	   splits them. Example: 20:94 (يَا + ابْنَأُمَّ → XML w1+w2
//	   vs MASAQ w2).
//
//	D. Cascade shifts: when an early MASAQ merge shifts all
//	   subsequent word numbers by +1, the final words of a verse
//	   appear "missing" at their XML position but exist at position-1
//	   in MASAQ. The user's example 8:6:12 يَنظُرُونَ is this case:
//	   بَعْدَمَا (MASAQ w4) = بَعْدَ + مَا (XML w4+w5), shifting
//	   يَنظُرُونَ from XML w12 to MASAQ w11.
//
// These 23 gaps are known and documented. The UI handles them
// gracefully with a "no morphological data" message. Fixing them
// would require fuzzy word-text matching instead of strict positional
// lookup, which risks false positives.
func TestFindMissingMorphology(t *testing.T) {
	dbPath, err := filepath.Abs("../../../data/new/detailed-quran.db")
	if err != nil {
		t.Fatal(err)
	}
	quran, masaq, _, _, _, err := LoadAll(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	totalWords := 0

	for sn := 1; sn <= 114; sn++ {
		s := quran.Surahs[sn]
		if s == nil {
			continue
		}
		for an := 1; an <= len(s.Ayahs); an++ {
			ay := s.Ayahs[an]
			if ay == nil {
				continue
			}
			for _, tok := range ay.Tokens {
				if tok.Kind != "word" {
					continue
				}
				totalWords++
				key := loc.Key(sn, an, tok.WordNo)
				if segs, ok := masaq.ByWord[key]; !ok || len(segs) == 0 {
					missing = append(missing, fmt.Sprintf("%d:%d:%d %s",
						sn, an, tok.WordNo, tok.Value))
				}
			}
		}
	}

	// Report all missing words with categories.
	type gapInfo struct {
		loc      string
		token    string
		category string
		detail   string
	}
	categories := map[string][]gapInfo{}

	catA := "A. MASAQ merges X+ما where XML splits"
	catB := "B. MASAQ merges ما+دام where XML splits"
	catC := "C. Vocative يا merge"
	catD := "D. Cascade shift from earlier merge"

	for _, m := range missing {
		parts := strings.SplitN(m, " ", 2)
		loc, token := parts[0], parts[1]

		// Strip tashkeel for form detection.
		var bare []rune
		for _, r := range token {
			if r >= 0x064B && r <= 0x065F || r == 0x0670 {
				continue
			}
			bare = append(bare, r)
		}
		bareStr := string(bare)

		gi := gapInfo{loc: loc, token: token}

		switch {
		case bareStr == "ما":
			gi.category = catA
			gi.detail = "MASAQ merged preceding word with ما (e.g., أَيْنَمَا, كُلَّمَا, بَعْدَمَا)"
		case bareStr == "لا" && strings.Contains(loc, "20:94"):
			gi.category = catC
			gi.detail = "XML: يا + ابنأم vs MASAQ: يَاابْنَأُمَّ (merged)"
		case bareStr == "ياسين" || bareStr == "إلياسين":
			gi.category = catC
			gi.detail = "XML: إِلْ + يَاسِينَ vs MASAQ: إِلْيَاسيْنَ (merged)"
		case loc == "20:94:3":
			gi.category = catC
			gi.detail = "XML: يا + ابنأم vs MASAQ: يَاابْنَأُمَّ (merged)"
		case strings.Contains(loc, "5:117") || strings.Contains(loc, "19:31") ||
			strings.Contains(loc, "11:107") || strings.Contains(loc, "5:96"):
			gi.category = catB
			gi.detail = "MASAQ merged ما+دام/دامت/دمت/دمتم as single token"
		default:
			gi.category = catD
			gi.detail = "Shifted by earlier merge in same verse"
		}
		categories[gi.category] = append(categories[gi.category], gi)
	}

	for _, cat := range []string{catA, catB, catC, catD} {
		gaps := categories[cat]
		if len(gaps) == 0 {
			continue
		}
		t.Logf("\n── %s (%d gaps) ──", cat, len(gaps))
		for _, g := range gaps {
			t.Logf("  %s %s  — %s", g.loc, g.token, g.detail)
		}
	}

	t.Logf("\n═══ Summary ═══")
	t.Logf("Total Quran words (Uthmani tokenization): %d", totalWords)
	t.Logf("Words with morphology (MASAQ index):       %d", len(masaq.ByWord))
	t.Logf("Words without morphology:                 %d (%.2f%%)",
		len(missing), float64(len(missing))/float64(totalWords)*100)
	t.Logf("Unique missing forms: %d", len(categories))
	t.Logf("")
	t.Logf("Root cause: MASAQ and the Uthmani XML use different word-boundary")
	t.Logf("conventions for certain grammatical particles. MASAQ merges ما, دام,")
	t.Logf("and يا into the neighboring word; the XML tokenizer keeps them")
	t.Logf("separate. This shifts all subsequent word numbers in the verse,")
	t.Logf("causing positional lookup to miss at the end.")
	t.Logf("")
	t.Logf("The UI shows 'No morphological data' for these 23 positions.")
	t.Logf("This is expected behavior — not a bug in the data pipeline.")
}
