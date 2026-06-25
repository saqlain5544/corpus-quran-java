package pipeline

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"quranreader/loc"
	"quranreader/types"
)

// LoadRoots reads corpus-roots.json and meanings-roots-ai.jsonl and
// returns a unified RootsIndex.
//
// corpus-roots.json format:
//
//	{ "qwl": ["1:2:6", "1:3:9", ...], "kwn": [...] }
//
// meanings-roots-ai.jsonl format (one JSON object per line):
//
//	{"id":1,"root":"qwl","root_arabic":"قَوْل","root_letters":"ق و ل",
//	 "pos":"V","occurrences_quran":1722,"meaning":{"en":"...","ar":"..."},
//	 ...}
//
// The two files are joined on the buckwalter key. corpus-roots.json
// supplies the locations; meanings-ai supplies the meaning text.
func LoadRoots(corpusJSON, aiJSONL string) (*types.RootsIndex, error) {
	b, err := os.ReadFile(corpusJSON)
	if err != nil {
		return nil, fmt.Errorf("read corpus-roots.json: %w", err)
	}
	b2, err := os.ReadFile(aiJSONL)
	if err != nil {
		return nil, fmt.Errorf("read meanings-roots-ai.jsonl: %w", err)
	}
	return parseRoots(b, b2)
}

// parseRoots is the in-memory counterpart to LoadRoots. The two file
// reads can be parallelized by the caller (see pipeline/build.go);
// this function operates on the already-loaded byte slices.
func parseRoots(corpusJSON, aiJSONL []byte) (*types.RootsIndex, error) {
	type corpusFile map[string][]string
	var raw corpusFile
	if err := json.Unmarshal(corpusJSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal corpus-roots.json: %w", err)
	}

	meanings := make(map[string]aiEntry)
	sc := bufio.NewScanner(bytes.NewReader(aiJSONL))
	sc.Buffer(make([]byte, 1024*1024), 8*1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var e aiEntry
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("parse line: %w", err)
		}
		if e.Root == "" {
			continue
		}
		meanings[e.Root] = e
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan meanings-ai: %w", err)
	}

	idx := &types.RootsIndex{
		ByRoot:   make(map[string]*types.RootEntry, len(raw)),
		ByLoc:    make(map[uint64]string, 50000),
		ByArabic: make(map[string]string, len(raw)),
	}
	for bw, locs := range raw {
		entry := &types.RootEntry{
			Buckwalter:  bw,
			Locations:   append([]string(nil), locs...),
			Occurrences: len(locs),
		}
		if m, ok := meanings[bw]; ok {
			entry.Arabic = m.RootArabic
			entry.Letters = m.RootLetters
			entry.POS = m.Pos
			// Note: meanings-ai's occurrences_quran sometimes
			// counts unique verses rather than word occurrences,
			// so we only override when it's clearly the larger
			// value. Otherwise we keep len(locs) which is the
			// word-occurrence count from corpus-roots.json.
			if m.OccurrencesQuran > entry.Occurrences {
				entry.Occurrences = m.OccurrencesQuran
			}
			entry.MeaningEN = m.Meaning.En
			entry.MeaningAR = m.Meaning.Ar
		}
		sort.Strings(entry.Locations)
		idx.ByRoot[bw] = entry
		for _, l := range locs {
			s, a, w, err := parseLoc(l)
			if err != nil {
				return nil, fmt.Errorf("root %s: bad loc %q: %w", bw, l, err)
			}
			idx.ByLoc[loc.Key(s, a, w)] = bw
		}
		if entry.Arabic != "" {
			letters := stripDiacritics(entry.Arabic)
			idx.ByArabic[letters] = bw
		}
	}
	return idx, nil
}

type aiEntry struct {
	Root             string `json:"root"`
	RootArabic       string `json:"root_arabic"`
	RootLetters      string `json:"root_letters"`
	Pos              string `json:"pos"`
	OccurrencesQuran int    `json:"occurrences_quran"`
	Meaning          struct {
		En string `json:"en"`
		Ar string `json:"ar"`
	} `json:"meaning"`
}

func parseLoc(s string) (surah, ayah, word int, err error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("expected s:v:w, got %q", s)
	}
	surah, err = strconv.Atoi(parts[0])
	if err != nil {
		return
	}
	ayah, err = strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	word, err = strconv.Atoi(parts[2])
	return
}

// stripDiacritics removes tashkeel (Unicode combining marks) from an
// Arabic string so that lookups by raw Arabic letters are robust.
func stripDiacritics(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r >= 0x064B && r <= 0x065F { // tashkeel range
			continue
		}
		if r == 0x0670 { // alef khanjariya
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
