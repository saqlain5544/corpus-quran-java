package data

import (
	"bufio"
	"database/sql"
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

// LoadConcordance reads concordance.jsonl from file and returns the
// assembled concordance. Prefer LoadConcordanceFromDB when a DB is
// available; this file-based loader is kept for standalone use.
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

// LoadConcordanceFromDB reads concordance data from the SQLite database
// and returns the assembled Concordance, ordered by total occurrences
// descending. Uses root_lemmas + lemma_positions tables.
func LoadConcordanceFromDB(db *sql.DB) (*Concordance, error) {
	c := &Concordance{
		ByRoot: make(map[string]*ConcordanceEntry, 1700),
	}

	// Load lemma data with positions. JOIN through verses table to
	// get surah_id and verse_number — lemma_positions only stores
	// lemma_id + verse_id (normalized).
	rows, err := db.Query(`
		SELECT rl.root_buckwalter, rl.lemma_arabic, rl.occurrences,
		       v.surah_id, v.verse_number
		FROM root_lemmas rl
		LEFT JOIN lemma_positions lp ON rl.id = lp.lemma_id
		LEFT JOIN verses v ON lp.verse_id = v.id
		ORDER BY rl.root_buckwalter, rl.lemma_arabic, v.surah_id, v.verse_number
	`)
	if err != nil {
		return nil, fmt.Errorf("query root_lemmas: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bw, lemmaAr string
		var occ int
		var surahID, verseNum sql.NullInt64
		if err := rows.Scan(&bw, &lemmaAr, &occ, &surahID, &verseNum); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}

		entry, ok := c.ByRoot[bw]
		if !ok {
			entry = &ConcordanceEntry{
				Root: bw,
				Lemmas: make(map[string]struct {
					TotalOccurrences int      `json:"total_occurrences"`
					Occurrences      []string `json:"occurrences"`
				}),
			}
			c.ByRoot[bw] = entry
		}

		lemma, ok := entry.Lemmas[lemmaAr]
		if !ok {
			lemma = struct {
				TotalOccurrences int      `json:"total_occurrences"`
				Occurrences      []string `json:"occurrences"`
			}{
				TotalOccurrences: occ,
				Occurrences:      nil,
			}
		}

		if surahID.Valid && verseNum.Valid {
			pos := fmt.Sprintf("%d:%d", surahID.Int64, verseNum.Int64)
			lemma.Occurrences = append(lemma.Occurrences, pos)
		}

		entry.Lemmas[lemmaAr] = lemma
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows: %w", err)
	}

	// TotalOccurrences comes from the roots table — one value per root,
	// independent of how many lemma_positions rows joined.
	freqRows, qErr := db.Query(`SELECT root_buckwalter, occurrences_quran FROM roots`)
	if qErr != nil {
		return nil, fmt.Errorf("query roots frequencies: %w", qErr)
	}
	defer freqRows.Close()
	for freqRows.Next() {
		var bw string
		var f int
		if err := freqRows.Scan(&bw, &f); err != nil {
			continue
		}
		if entry, ok := c.ByRoot[bw]; ok {
			entry.TotalOccurrences = f
		}
	}
	if err := freqRows.Err(); err != nil {
		return nil, fmt.Errorf("scan roots frequencies: %w", err)
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
