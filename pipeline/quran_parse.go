// Package pipeline contains the preprocessing tool that turns the
// raw data files (quran-uthmani.xml, MASAQ.csv, corpus-roots.json,
// meanings-roots-ai.jsonl) into the optimized gob-encoded form loaded
// at server startup.
//
// The parser is intentionally separate from the runtime data loader:
// it does not need to be goroutine-safe, and it allocates
// aggressively because it runs once at build time.
package pipeline

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"

	"quranreader/loc"
	"quranreader/token"
	"quranreader/types"
)

// LoadQuran reads quran-uthmani.xml and returns the in-memory Quran
// struct. The parser uses an event-based xml.Decoder so peak memory
// stays well below the file size — only the output structs are kept.
//
// Bismillah rules implemented (per docs/pen-and-paper/architecture.md
// and the existing render-to-html reference):
//
//   - Surah 1: Bismillah IS verse 1 (no separate bismillah attr).
//   - Surah 9: No Bismillah at all.
//   - Other surahs: if aya 1's `bismillah` attribute is present OR
//     aya 1's text equals the canonical Bismillah, store it on the
//     surah and treat aya 1 as a regular verse.
func LoadQuran(path string) (*types.Quran, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseQuran(f)
}

func parseQuran(r io.Reader) (*types.Quran, error) {
	dec := xml.NewDecoder(r)

	q := &types.Quran{
		Surahs: make(map[int]*Surah, 114),
	}
	// We will set Meta after parsing completes.

	const (
		stateDoc     = 0
		stateInSurah = 1
		stateInAyah  = 2
	)
	state := stateDoc
	var curSurah *Surah

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "sura":
				if state != stateDoc {
					return nil, fmt.Errorf("nested <sura> at line %d", dec.InputOffset())
				}
				sn, err := attrInt(t.Attr, "index")
				if err != nil {
					return nil, fmt.Errorf("sura index: %w", err)
				}
				if !loc.Valid(sn, 1, 1) {
					return nil, fmt.Errorf("invalid surah number: %d", sn)
				}
				curSurah = &Surah{
					Number: sn,
					Name:   attrStr(t.Attr, "name"),
					Ayahs:  make(map[int]*Ayah),
				}
				state = stateInSurah
			case "aya":
				if state != stateInSurah {
					return nil, fmt.Errorf("<aya> outside <sura>")
				}
				an, err := attrInt(t.Attr, "index")
				if err != nil {
					return nil, err
				}
				text := attrStr(t.Attr, "text")
				bismillah := attrStr(t.Attr, "bismillah")
				ayah := &Ayah{Number: an, Text: text}

				// Bismillah handling.
				if curSurah.Number == 1 {
					if an == 1 {
						curSurah.Bismillah = text
					}
				} else if curSurah.Number != 9 && an == 1 {
					if bismillah != "" {
						curSurah.Bismillah = bismillah
				} else if token.IsBismillah(text) {
					curSurah.Bismillah = text
				}
			}

			// Skip the standalone-Bismillah-as-aya-1 case (not
			// Surah 1, since Surah 1's aya 1 IS the Bismillah).
			if curSurah.Number != 1 && an == 1 && token.IsBismillah(text) {
					// aya 1 is Bismillah itself: don't store as verse.
				} else {
					curSurah.Ayahs[an] = ayah
				}
				state = stateInAyah
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "sura":
				q.Surahs[curSurah.Number] = curSurah
				curSurah = nil
				state = stateDoc
			case "aya":
				state = stateInSurah
			}
		}
	}

	if len(q.Surahs) != 114 {
		return nil, fmt.Errorf("expected 114 surahs, parsed %d", len(q.Surahs))
	}

	// Build meta and tokenize ayas now that the structure is complete.
	meta := types.Meta{
		SurahNames: make([]string, 114),
		AyahCounts: make([]int, 114),
	}
	for n := 1; n <= 114; n++ {
		s, ok := q.Surahs[n]
		if !ok {
			return nil, fmt.Errorf("missing surah %d", n)
		}
		meta.SurahNames[n-1] = s.Name
		meta.AyahCounts[n-1] = len(s.Ayahs)
		meta.AyahCount += len(s.Ayahs)
		for _, ay := range s.Ayahs {
			ay.Tokens = token.Tokenize(ay.Text)
			for _, t := range ay.Tokens {
				if t.Kind == "word" {
					meta.WordCount++
				}
			}
		}
	}
	q.Meta = meta
	return q, nil
}

// attrInt parses an XML attribute as int.
func attrInt(attrs []xml.Attr, name string) (int, error) {
	for _, a := range attrs {
		if a.Name.Local == name {
			return atoi(a.Value)
		}
	}
	return 0, fmt.Errorf("missing attribute %q", name)
}

// attrStr reads an XML attribute as a string ("" if missing).
func attrStr(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// atoi is a tiny strconv.Atoi wrapper that returns an error with the
// value baked in.
func atoi(s string) (int, error) {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number: %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// Surah and Ayah are local aliases to keep the pipeline source
// readable; the output written to gob is the public types.Surah /
// types.Ayah.
type Surah = types.Surah
type Ayah = types.Ayah

// SortSurahsByNumber returns the surah numbers in 1..114 order. Used
// by the build report.
func SortSurahsByNumber(q *types.Quran) []int {
	out := make([]int, 0, len(q.Surahs))
	for n := range q.Surahs {
		out = append(out, n)
	}
	sort.Ints(out)
	return out
}
