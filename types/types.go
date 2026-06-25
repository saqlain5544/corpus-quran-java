// Package types defines the shared data structures used by both the
// pipeline (data/preprocess) and the runtime server. Keeping these in
// one place guarantees the gob files we emit can be loaded back without
// ad-hoc converters.
package types

import "strings"

// Quran is the top-level in-memory structure holding all surahs.
type Quran struct {
	Surahs map[int]*Surah
	Meta   Meta
}

// Surah holds the metadata and verses for one chapter.
type Surah struct {
	Number             int
	Name               string // Arabic
	NameLatin          string // optional English transliteration (empty if unknown)
	RevelationType     string // "Meccan" or "Medinan"
	EnglishTranslation string // short English descriptor (e.g. "The Opening")
	Bismillah          string // raw text, may be empty for surah 9
	Ayahs              map[int]*Ayah
}

// Ayah holds the raw text (with marks inline) and a pre-parsed list of
// tokens (words + marks) for rendering.
type Ayah struct {
	Number int
	Text   string
	Tokens []Token
}

// Token is one element of an ayah's tokenized form.
//
// Kind == "word"  → Value is the Arabic word (with tashkeel), Marks may
//
//	contain trailing marks peeled off the source token,
//	WordNo is the 1-based position within the ayah.
//
// Kind == "mark"  → Value is a single mark character (e.g., "۞" or "۩"),
//
//	WordNo is zero.
type Token struct {
	Kind   string
	Value  string
	Marks  []string
	WordNo int
}

// Meta holds aggregate counts and lookup tables that are useful at
// request time without re-scanning all surahs.
type Meta struct {
	SurahNames []string // index 0 == surah 1
	AyahCounts []int    // ayahs per surah, index 0 == surah 1
	WordCount  int      // total word tokens across all surahs
	AyahCount  int      // total ayas
}

// MasaqSegment is one row of MASAQ.csv. A word can have 1..N segments.
type MasaqSegment struct {
	ID                   int
	SuraNo               int
	VerseNo              int
	WordNo               int
	SegmentNo            int
	Word                 string
	WithoutDiacritics    string
	SegmentedWord        string
	MorphTag             string
	MorphType            string
	PunctuationMark      string
	InvariableDeclinable string
	SyntacticRole        string
	PossessiveConstruct  string
	CaseMood             string
	CaseMoodMarker       string
	Phrase               string
	PhrasalFunction      string
	Gloss                string
	Translation          string // word-level English translation (plan.md §Translation)
}

// MasaqIndex is the MASAQ lookup structure. ByWord is keyed by the
// 64-bit LocKey(surah, ayah, word) so each fetch is a single hashmap
// lookup with no allocations.
//
// Note: there are ~23 cases where MASAQ and the XML disagree on word
// boundaries (MASAQ merges "waw + ma" into "wama" while the XML
// keeps them as two tokens). For those, the strict ByWord lookup
// misses and the server reports "no morphological data" rather than
// try a fuzzy match. The known divergences are documented in
// docs/research/data-validation.md.
type MasaqIndex struct {
	ByWord map[uint64][]MasaqSegment
}

// RootEntry is one root with its meanings and occurrences.
type RootEntry struct {
	Buckwalter     string          // ASCII key, e.g. "qwl"
	Arabic         string          // Arabic letters with optional tashkeel
	Letters        string          // space-separated Arabic letters
	POS            string          // V | N | P | ...
	Occurrences    int             // total occurrences in Quran
	MeaningEN      string          // English meaning summary
	MeaningAR      string          // Arabic meaning summary
	CoreSemantic   string          // core semantic field (AI-generated)
	IbnFaris       string          // Ibn Faris dictionary reference
	AlRaghib       string          // Al-Raghib dictionary reference
	QuranExamples  []QuranExample  // illustrative verses
	HadithExamples []HadithExample // relevant hadith
	Locations      []string        // "s:v:w" strings sorted
}

// QuranExample is one illustrative Quranic verse from the AI meanings.
type QuranExample struct {
	Ref     string `json:"ref"`     // "s:v"
	Arabic  string `json:"ar"`      // arabic text
	English string `json:"en"`      // english translation
	Context string `json:"context"` // why this verse illustrates the root
}

// HadithExample is a relevant hadith from the AI meanings.
type HadithExample struct {
	Arabic  string `json:"ar"`     // arabic text
	English string `json:"en"`     // english translation
	Source  string `json:"source"` // collection reference
}

// RootsIndex groups all root data structures together.
type RootsIndex struct {
	ByRoot       map[string]*RootEntry   // buckwalter key
	ByLoc        map[uint64]string       // LocKey → buckwalter key
	ByArabic     map[string]string       // normalized Arabic letters → buckwalter key
	RootsBySurah map[int]map[string]bool // surah number → set of buckwalter roots
}

// LoadMetaFromFile reads meta.json from disk. The Meta struct is the
// small JSON sidecar that travels alongside the gob files.
func LoadMetaFromFile(path string) (*Meta, error) {
	// Implementation lives in the same package so we can use json
	// here without re-importing. We pull the body into a local
	// function to avoid a circular dep with the JSON emitter.
	return loadMetaJSON(path)
}

// NormalizeArabicRoot canonicalizes a root-letters string for the
// ByArabic reverse lookup. The DB stores letters space-separated
// (e.g. "ق و ل") for readability, but the lookup key must be
// invariant under:
//
//   - whitespace (the user types "قول", not "ق و ل")
//   - alif/ya/ta-marbuta variants (the user types "اب ي", not "أب ي")
//   - tashkeel (stripped by NormalizeArabic)
//
// Both the load path (data.LoadAll → ByArabic) and the query path
// (search.Root, server.handleRootDetail) call this function, so they
// agree on the canonical key by construction. Implemented in the
// types package — which has no other dependencies — to avoid an
// import cycle between data and search.
func NormalizeArabicRoot(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r >= 0x064B && r <= 0x065F: // tashkeel
			continue
		case r == 0x0670, r == 0x0640: // alef khanjariya, tatweel
			continue
		case r == 'أ', r == 'إ', r == 'آ', r == 'ٱ':
			b.WriteRune('ا')
		case r == 'ى', r == 'ٰ':
			b.WriteRune('ي')
		case r == 'ة':
			b.WriteRune('ه')
		case r == 'ؤ', r == 'ئ', r == 'ء':
			b.WriteRune('ا')
		case r == ' ':
			// Skip — the lookup key has no spaces regardless of
			// how the user typed the query.
			continue
		default:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
