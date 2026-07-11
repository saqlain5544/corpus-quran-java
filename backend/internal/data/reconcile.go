package data

import (
	"quranreader/loc"
	"quranreader/types"
)

// ─── Word-boundary reconciliation ──────────────────────────────

// reconcileMasaqBoundaries fixes word-boundary disagreements between
// the Quran XML tokenizer and MASAQ's morphological segmentation.
func reconcileMasaqBoundaries(m *types.MasaqIndex, q *types.Quran) {
	for sn := 1; sn <= 114; sn++ {
		surah := q.Surahs[sn]
		if surah == nil {
			continue
		}
		for an := 1; an <= len(surah.Ayahs); an++ {
			ayah := surah.Ayahs[an]
			if ayah == nil {
				continue
			}
			wordTokens := wordTokensFromAyah(ayah)
			verseMasaq := collectVerseMasaq(m, sn, an)

			for wn := 2; wn <= len(wordTokens); wn++ {
				key := loc.Key(sn, an, wn)
				if _, ok := m.ByWord[key]; ok {
					continue
				}
				missingBare := stripTashkeel(wordTokens[wn-1])

				// 1. Position-adjacent: try wn-1, then wn-2.
				if tryMerge(m, sn, an, wn-1, missingBare, key) {
					continue
				}
				if wn >= 3 && tryMerge(m, sn, an, wn-2, missingBare, key) {
					continue
				}

				// 2. Content-based: search verse's MASAQ entries.
				for _, vm := range verseMasaq {
					for _, bf := range vm.bareForms {
						if bf == missingBare {
							m.ByWord[key] = vm.segs
							goto fixed
						}
					}
					if len(vm.fullBare) >= len(missingBare) &&
						vm.fullBare[len(vm.fullBare)-len(missingBare):] == missingBare {
						m.ByWord[key] = vm.segs
						goto fixed
					}
				}
			fixed:
			}
		}
	}
}

type verseMasaqEntry struct {
	wordNo    int
	bareForms []string
	fullBare  string
	segs      []types.MasaqSegment
}

func collectVerseMasaq(m *types.MasaqIndex, surah, ayah int) []verseMasaqEntry {
	var out []verseMasaqEntry
	for wn := 1; wn <= loc.MaxWord; wn++ {
		key := loc.Key(surah, ayah, wn)
		segs, ok := m.ByWord[key]
		if !ok || len(segs) == 0 {
			continue
		}
		bareForms := make([]string, 0, len(segs))
		var fullBareBuilder []rune
		for _, s := range segs {
			wd := s.WithoutDiacritics
			if wd != "" && wd != "None" {
				bareForms = append(bareForms, wd)
				fullBareBuilder = append(fullBareBuilder, []rune(wd)...)
			}
		}
		out = append(out, verseMasaqEntry{wn, bareForms, string(fullBareBuilder), segs})
	}
	return out
}

func tryMerge(m *types.MasaqIndex, surah, ayah, prevPos int, missingBare string, missingKey uint64) bool {
	prevSegs, ok := m.ByWord[loc.Key(surah, ayah, prevPos)]
	if !ok || len(prevSegs) == 0 {
		return false
	}
	lastSeg := prevSegs[len(prevSegs)-1]
	if lastSeg.WithoutDiacritics == missingBare ||
		stripTashkeel(lastSeg.SegmentedWord) == missingBare {
		m.ByWord[missingKey] = prevSegs
		return true
	}
	return false
}

// stripTashkeel removes Arabic diacritical marks, formatting
// characters, and Quranic annotation marks.
func stripTashkeel(s string) string {
	var out []rune
	for _, r := range s {
		switch {
		case r == 0x0640: // tatweel
		case r == 0x0670: // dagger alif → alif
			out = append(out, 0x0627)
		case r == 0x0671: // alif wasla → alif
			out = append(out, 0x0627)
		case r >= 0x064B && r <= 0x065F: // tashkeel
		case r >= 0x06D6 && r <= 0x06ED: // Quranic annotation marks
		default:
			out = append(out, r)
		}
	}
	return string(out)
}

func wordTokensFromAyah(ayah *types.Ayah) []string {
	var words []string
	for _, t := range ayah.Tokens {
		if t.Kind == "word" {
			words = append(words, t.Value)
		}
	}
	return words
}
