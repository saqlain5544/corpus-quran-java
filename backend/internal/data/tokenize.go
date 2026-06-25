package data

import (
	"strings"

	"quranreader/types"
)

// BISMILLAH_TEXT is the canonical Bismillah in Uthmani script. Used
// to detect when a verse is the Bismillah so we can store it on the
// Surah struct and skip it as a numbered verse for surahs 2-114
// (except 9).
const BISMILLAH_TEXT = "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ"

// isBismillah reports whether s is the canonical Bismillah text.
func isBismillah(s string) bool {
	return strings.TrimSpace(s) == BISMILLAH_TEXT
}

// tokenTrailingMarks are the Quran annotation characters that may be
// appended to a word in the source. They match the inline
// Arabic-Extended characters found by the Research Agent.
//
//	U+06D6 .. U+06DB → pause marks (ۖ ۗ ۘ ۙ ۚ ۛ)
//	U+06DC           → sakta (ۜ, rare brief pause)
//	U+06DE           → rub' / hizb (۞)
//	U+06E2           → small high meem (ۢ) — iqlab mark above noon
//	                    before ba; rendered above the previous letter
//	U+06ED           → small low meem (ۭ) — iqlab variant rendered
//	                    below the baseline; same purpose as U+06E2
//	U+06E9           → sajdah (۩)
//
// Other annotation characters (۟ U+06DF, ۠ U+06E0, ۡ U+06E1,
// ۥ U+06E5, ۦ U+06E6, ۧ U+06E7, ۨ U+06E8, and the three stop marks
// U+06EA..U+06EC) are diacritics that belong to the letters
// themselves; they are NOT trailing marks.
var tokenTrailingMarks = map[rune]bool{
	'\u06D6': true,
	'\u06D7': true,
	'\u06D8': true,
	'\u06D9': true,
	'\u06DA': true,
	'\u06DB': true,
	'\u06DC': true,
	'\u06DE': true,
	'\u06E2': true, // ۢ iqlab (high meem)
	'\u06ED': true, // ۭ iqlab (low meem)
	'\u06E9': true,
}

// splitMark peels trailing mark characters off the end of a token.
// It returns the remaining word (possibly empty) and the marks in
// their original order (left-to-right as they appeared).
func splitMark(token string) (string, []string) {
	runes := []rune(token)
	cut := len(runes)
	marks := make([]string, 0, 2)
	for cut > 0 && tokenTrailingMarks[runes[cut-1]] {
		cut--
		marks = append(marks, string(runes[cut]))
	}
	// Reverse marks so we get them in original left-to-right order.
	for i, j := 0, len(marks)-1; i < j; i, j = i+1, j-1 {
		marks[i], marks[j] = marks[j], marks[i]
	}
	return string(runes[:cut]), marks
}

// tokenize breaks aya text into (word | mark) tokens.
//
// The algorithm:
//
//	for each whitespace-delimited token:
//	    word, marks = splitMark(token)
//	    if word:
//	        append word token with Marks = marks
//	    else if marks:
//	        append each mark as its own mark token
//
// Trailing marks are stored in the word token's Marks array — NOT
// emitted as separate mark tokens — so the HTML template renders each
// mark exactly once. Standalone mark tokens are only created when a
// whitespace-delimited position contains only marks (e.g., the rub'
// ۞ at the start of an ayah).
//
// Word numbers are 1-based and count only word tokens (marks do not
// increment the counter).
func tokenize(text string) []types.Token {
	out := make([]types.Token, 0, len(text)/4)
	wordNo := 0
	for _, raw := range strings.Fields(text) {
		word, marks := splitMark(raw)
		if word != "" {
			wordNo++
			t := types.Token{Kind: "word", Value: word, WordNo: wordNo}
			if len(marks) > 0 {
				t.Marks = make([]string, len(marks))
				copy(t.Marks, marks)
			}
			out = append(out, t)
		} else {
			for _, m := range marks {
				out = append(out, types.Token{Kind: "mark", Value: m})
			}
		}
	}
	return out
}
