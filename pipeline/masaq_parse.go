package pipeline

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	"quranreader/loc"
	"quranreader/types"
)

// LoadMasaq reads MASAQ.csv and returns a MasaqIndex keyed by loc.Key(surah, ayah, word).
//
// The CSV is simple comma-separated with no quoted fields, so we use
// encoding/csv for safety (handles CRLF, embedded quotes, etc.). Each
// row's ID is the same as the existing corpus ID (sequential 1..N);
// we keep it because the join with corpus-roots.json will eventually
// want it.
//
// Expected column order (verified by Research Agent):
//
//	ID, Sura_No, Verse_No, Word_No, Segment_No, Word,
//	Without_Diacritics, Segmented_Word, Morph_Tag, Morph_Type,
//	Punctuation_Mark, Invariable_Declinable, Syntactic_Role,
//	Possessive_Construct, Case_Mood, Case_Mood_Marker, Phrase,
//	Phrasal_Function, Gloss
func LoadMasaq(path string) (*types.MasaqIndex, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseMasaq(f)
}

func parseMasaq(r io.Reader) (*types.MasaqIndex, error) {
	rdr := csv.NewReader(r)
	rdr.FieldsPerRecord = 19
	idx := &types.MasaqIndex{
		ByWord: make(map[uint64][]types.MasaqSegment, 80000),
	}

	header, err := rdr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if len(header) != 19 {
		return nil, fmt.Errorf("unexpected header width: %d", len(header))
	}

	line := 1
	for {
		row, err := rdr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		line++

		seg, err := rowToSegment(row)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		key := loc.Key(seg.SuraNo, seg.VerseNo, seg.WordNo)
		idx.ByWord[key] = append(idx.ByWord[key], seg)
	}
	return idx, nil
}

// rowToSegment converts one CSV row into a MasaqSegment. Centralized
// so it is easy to unit-test independently of the file reader.
func rowToSegment(row []string) (types.MasaqSegment, error) {
	if len(row) != 19 {
		return types.MasaqSegment{}, fmt.Errorf("expected 19 cols, got %d", len(row))
	}
	intCols := []int{0, 1, 2, 3, 4} // ID, Sura, Verse, Word, Segment
	ints := make([]int, len(intCols))
	for i, c := range intCols {
		n, err := strconv.Atoi(row[c])
		if err != nil {
			return types.MasaqSegment{}, fmt.Errorf("col %d: %w", c, err)
		}
		ints[i] = n
	}
	return types.MasaqSegment{
		ID:                   ints[0],
		SuraNo:               ints[1],
		VerseNo:              ints[2],
		WordNo:               ints[3],
		SegmentNo:            ints[4],
		Word:                 row[5],
		WithoutDiacritics:    row[6],
		SegmentedWord:        row[7],
		MorphTag:             row[8],
		MorphType:            row[9],
		PunctuationMark:      row[10],
		InvariableDeclinable: row[11],
		SyntacticRole:        row[12],
		PossessiveConstruct:  row[13],
		CaseMood:             row[14],
		CaseMoodMarker:       row[15],
		Phrase:               row[16],
		PhrasalFunction:      row[17],
		Gloss:                row[18],
	}, nil
}
