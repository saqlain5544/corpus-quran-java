package data

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// ConcordanceEntry is one root's lemma map from concordance.jsonl.
type ConcordanceEntry struct {
	Root             string              `json:"root"`
	TotalOccurrences int                 `json:"total_occurrences"`
	Lemmas           map[string]struct { // map arabic lemma → ...
		TotalOccurrences int      `json:"total_occurrences"`
		Occurrences      []string `json:"occurrences"` // "s:v" pairs
	} `json:"lemmas"`
}

// Concordance holds the loaded concordance data keyed by buckwalter.
type Concordance struct {
	ByRoot map[string]*ConcordanceEntry // buckwalter → entry
	Keys   []string                     // ordered by freq desc
}

// LoadConcordance reads concordance.jsonl and returns the assembled
// concordance, ordered by total occurrences descending.
func LoadConcordance(path string) (*Concordance, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := &Concordance{
		ByRoot: make(map[string]*ConcordanceEntry, 1700),
	}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var e ConcordanceEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, fmt.Errorf("parse line: %w", err)
		}
		if e.Root == "" {
			continue
		}
		c.ByRoot[e.Root] = &e
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Build ordered key list by frequency desc.
	c.Keys = make([]string, 0, len(c.ByRoot))
	for k := range c.ByRoot {
		c.Keys = append(c.Keys, k)
	}
	sort.Slice(c.Keys, func(i, j int) bool {
		ai := c.ByRoot[c.Keys[i]].TotalOccurrences
		aj := c.ByRoot[c.Keys[j]].TotalOccurrences
		if ai != aj {
			return ai > aj
		}
		return c.Keys[i] < c.Keys[j]
	})
	return c, nil
}
