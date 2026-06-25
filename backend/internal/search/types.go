// Package search implements server-side fuzzy search for the three
// modes documented in docs/pen-and-paper/routes-api.md:
//
//   - root      — match against RootEntry.Buckwalter or RootEntry.Arabic
//   - english   — match against MasaqSegment.Gloss (with Levenshtein)
//   - arabic    — match against MasaqSegment.Word (with normalization)
//
// All functions return up to N SearchResult values, ranked by mode-
// specific relevance. They are read-only over the loaded indexes.
package search

// SearchResult is the union shape returned by all three search modes.
// Fields used depend on the mode:
//   - root mode: Root, Surah, Ayah, Word all set; Snippet is empty
//   - english mode: Surah/Ayah/Word set; Snippet is the matched gloss
//   - arabic mode:  Surah/Ayah/Word set; Snippet is the matched word
type SearchResult struct {
	Kind      string `json:"kind"`       // "root" | "english" | "arabic" | "root-list"
	Surah     int    `json:"surah"`      // 0 for root-list
	Ayah      int    `json:"ayah"`       // 0 for root-list
	Word      int    `json:"word"`       // 0 for root-list
	Snippet   string `json:"snippet"`    // text excerpt with <mark>…</mark>
	VerseText string `json:"verse_text"` // full verse text (plan.md line 55)
	Title     string `json:"title"`      // short label
	Link      string `json:"link"`       // relative URL to follow
	Score     int    `json:"score"`      // lower = better match; for sorting
	Root      string `json:"root"`       // buckwalter, when present
}
