package data

import (
	"bufio"
	"encoding/json"
	"os"

	"quranreader/types"
)

type richRoot struct {
	Root             string `json:"root"`
	RootArabic       string `json:"root_arabic"`
	RootLetters      string `json:"root_letters"`
	POS              string `json:"pos"`
	OccurrencesQuran int    `json:"occurrences_quran"`
	Meaning          struct {
		EN string `json:"en"`
		AR string `json:"ar"`
	} `json:"meaning"`
	LexicalAnalysis struct {
		CoreSemanticField string `json:"core_semantic_field"`
		IbnFaris          string `json:"ibn_faris"`
		AlRaghib          string `json:"al_raghib"`
	} `json:"lexical_analysis"`
	QuranExamples []struct {
		Ref     string `json:"ref"`
		Arabic  string `json:"ar"`
		English string `json:"en"`
		Context string `json:"context"`
	} `json:"quran_examples"`
	Hadith []struct {
		Arabic  string `json:"ar"`
		English string `json:"en"`
		Source  string `json:"source"`
	} `json:"hadith"`
}

// loadRichMeanings reads meanings-roots-ai.jsonl and returns a map
// keyed by buckwalter code with the enriched types.
func loadRichMeanings() map[string]richRoot {
	// Try multiple paths: from server cwd or test cwd.
	paths := []string{
		"./data/morphology/meanings-roots-ai.jsonl",
		"../../../data/morphology/meanings-roots-ai.jsonl",
	}
	var rich map[string]richRoot
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		rich = make(map[string]richRoot, 1700)
		sc := bufio.NewScanner(f)
		sc.Buffer(make([]byte, 128*1024), 4*1024*1024)
		for sc.Scan() {
			var rr richRoot
			if err := json.Unmarshal(sc.Bytes(), &rr); err != nil {
				continue
			}
			if rr.Root == "" {
				continue
			}
			rich[rr.Root] = rr
		}
		f.Close()
		break
	}
	return rich
}

// toRootEntry converts a richRoot into the types used by RootEntry.
// Used as a helper when we need the full enriched types.
func toRootEntry(r richRoot) *types.RootEntry {
	e := &types.RootEntry{
		Buckwalter:   r.Root,
		Arabic:       r.RootArabic,
		Letters:      r.RootLetters,
		POS:          r.POS,
		Occurrences:  r.OccurrencesQuran,
		MeaningEN:    r.Meaning.EN,
		MeaningAR:    r.Meaning.AR,
		CoreSemantic: r.LexicalAnalysis.CoreSemanticField,
		IbnFaris:     r.LexicalAnalysis.IbnFaris,
		AlRaghib:     r.LexicalAnalysis.AlRaghib,
	}
	for _, ex := range r.QuranExamples {
		e.QuranExamples = append(e.QuranExamples, types.QuranExample{
			Ref: ex.Ref, Arabic: ex.Arabic, English: ex.English, Context: ex.Context,
		})
	}
	for _, h := range r.Hadith {
		e.HadithExamples = append(e.HadithExamples, types.HadithExample{
			Arabic: h.Arabic, English: h.English, Source: h.Source,
		})
	}
	return e
}
