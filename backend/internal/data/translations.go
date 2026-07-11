package data

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"sync"
)

// TranslationSet holds verse-level translation text for a single
// language, keyed by surah number (1-based) and ayah number (1-based).
type TranslationSet struct {
	Label string                 // display name, e.g. "Saheeh International"
	Lang  string                 // "en", "ur"
	Dir   string                 // "ltr" or "rtl"
	Data  map[int]map[int]string // surah → ayah → text
}

// Translations holds one or more loaded translation sets.
type Translations struct {
	Sets    []TranslationSet
	Default int // index into Sets, 0 = English
}

// LoadTranslations reads translation XML files from dir and returns
// the assembled Translations. It expects files matching
// *.sahih.xml (English) and *.junagarhi.xml (Urdu).
// Transliteration files are deliberately skipped per plan.md §Translation.
//
// English and Urdu files are loaded concurrently — they're independent
// I/O + XML parse, so they run in parallel and roughly halve total
// wall time when the disk is cold.
func LoadTranslations(dir string) (*Translations, error) {
	type result struct {
		set TranslationSet
		err error
	}
	var wg sync.WaitGroup
	var enRes, urRes result
	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := loadTransSet(dir+"/en.sahih.xml", "Saheeh International", "en", "ltr")
		enRes = result{s, err}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := loadTransSet(dir+"/ur.junagarhi.xml", "محمد جوناگڑھی", "ur", "rtl")
		urRes = result{s, err}
	}()
	wg.Wait()

	if enRes.err != nil {
		return nil, fmt.Errorf("load en.sahih.xml: %w", enRes.err)
	}
	if urRes.err != nil {
		return nil, fmt.Errorf("load ur.junagarhi.xml: %w", urRes.err)
	}

	t := &Translations{
		Sets:    []TranslationSet{enRes.set, urRes.set},
		Default: 0, // English
	}
	return t, nil
}

func loadTransSet(path, label, lang, dir string) (TranslationSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return TranslationSet{}, err
	}
	// The Tanzil translation XML files contain comments with runs of
	// dashes ("# ----...") that violate the XML spec (-- not allowed
	// inside comments). Strip the comment block before parsing.
	raw = stripXMLComment(raw)

	type ayaXML struct {
		Index int    `xml:"index,attr"`
		Text  string `xml:"text,attr"`
	}
	type suraXML struct {
		Index int      `xml:"index,attr"`
		Ayas  []ayaXML `xml:"aya"`
	}
	type quranXML struct {
		Surahs []suraXML `xml:"sura"`
	}

	var doc quranXML
	if err := xml.NewDecoder(bytes.NewReader(raw)).Decode(&doc); err != nil {
		return TranslationSet{}, fmt.Errorf("decode %s: %w", path, err)
	}

	data := make(map[int]map[int]string, 114)
	for _, s := range doc.Surahs {
		if s.Index < 1 || s.Index > 114 {
			continue
		}
		ayas := make(map[int]string, len(s.Ayas))
		hasText := false
		for _, a := range s.Ayas {
			if a.Index > 0 && a.Text != "" {
				ayas[a.Index] = a.Text
				hasText = true
			}
		}
		if hasText {
			data[s.Index] = ayas
		}
	}

	return TranslationSet{
		Label: label,
		Lang:  lang,
		Dir:   dir,
		Data:  data,
	}, nil
}

// stripXMLComment removes every XML comment block (<!-- ... -->) from
// the input. Tanzil translation files contain one leading comment
// block with dashes illegal under strict XML (-- inside comments),
// but defensive: strip all occurrences.
func stripXMLComment(raw []byte) []byte {
	for {
		start := bytes.Index(raw, []byte("<!--"))
		if start < 0 {
			return raw
		}
		end := bytes.Index(raw[start:], []byte("-->"))
		if end < 0 {
			return raw
		}
		end += start + 3
		// Recursively trim the comment and continue for any remaining.
		raw = append(raw[:start], raw[end:]...)
	}
}
