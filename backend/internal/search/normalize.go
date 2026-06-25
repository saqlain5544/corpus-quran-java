package search

import (
	"regexp"
	"strings"
)

// NormalizeArabic canonicalizes an Arabic string for fuzzy matching:
//
//   - strip tashkeel (U+064B..U+065F, U+0670 alef khanjariya,
//     U+0640 tatweel)
//   - unify hamza variants (أ إ آ ٱ → ا)
//   - unify ya/alef-maqsura (ى ٰ → ي)
//   - unify ta marbuta / ha at word boundary (ة → ه)
//   - unify hamza-on-line/under/standalone (ؤ ئ ء → ا)
//   - collapse whitespace
//
// The function is idempotent: NormalizeArabic(NormalizeArabic(s)) ==
// NormalizeArabic(s).
//
// Buckwalter ASCII transliteration is not produced here; callers that
// need it should call Buckwalter(s) afterwards.
func NormalizeArabic(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		// strip tashkeel
		case r >= 0x064B && r <= 0x065F:
			continue
		case r == 0x0670: // alef khanjariya
			continue
		case r == 0x0640: // tatweel
			continue
		// unify alif variants
		case r == 'أ' || r == 'إ' || r == 'آ' || r == 'ٱ':
			b.WriteRune('ا')
		// unify ya variants
		case r == 'ى' || r == 'ٰ':
			b.WriteRune('ي')
		// unify ta marbuta
		case r == 'ة':
			b.WriteRune('ه')
		// unify hamza variants
		case r == 'ؤ' || r == 'ئ' || r == 'ء':
			b.WriteRune('ا')
		case r == ' ' || r == '\t' || r == '\n' || r == '\r':
			b.WriteRune(' ')
		default:
			b.WriteRune(r)
		}
	}
	// collapse whitespace runs.
	out := b.String()
	var c strings.Builder
	c.Grow(len(out))
	prevSpace := false
	for _, r := range out {
		if r == ' ' {
			if !prevSpace {
				c.WriteRune(' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		c.WriteRune(r)
	}
	return strings.TrimSpace(c.String())
}

// Buckwalter transliterates an Arabic string to the ASCII
// Buckwalter encoding used in corpus-roots.json.
//
// Tashkeel is stripped except for shadda (U+0651), which doubles the
// preceding consonant (this is how root gemination is encoded in
// Buckwalter: رَبّ → rbb).
//
// This is best-effort: characters we don't recognize are passed
// through unchanged.
func Buckwalter(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		// shadda: repeat last emitted consonant
		if r == 0x0651 {
			out := b.String()
			if n := len(out); n > 0 {
				b.WriteByte(out[n-1])
			}
			continue
		}
		// strip other tashkeel
		if r >= 0x064B && r <= 0x065F {
			continue
		}
		if r == 0x0670 || r == 0x0640 {
			continue
		}
		switch r {
		case 'ا':
			b.WriteByte('A')
		case 'ب':
			b.WriteByte('b')
		case 'ت':
			b.WriteByte('t')
		case 'ث':
			b.WriteByte('v')
		case 'ج':
			b.WriteByte('j')
		case 'ح':
			b.WriteByte('H')
		case 'خ':
			b.WriteByte('x')
		case 'د':
			b.WriteByte('d')
		case 'ذ':
			b.WriteByte('*')
		case 'ر':
			b.WriteByte('r')
		case 'ز':
			b.WriteByte('z')
		case 'س':
			b.WriteByte('s')
		case 'ش':
			b.WriteByte('$')
		case 'ص':
			b.WriteByte('S')
		case 'ض':
			b.WriteByte('D')
		case 'ط':
			b.WriteByte('T')
		case 'ظ':
			b.WriteByte('Z')
		case 'ع':
			b.WriteByte('E')
		case 'غ':
			b.WriteByte('g')
		case 'ف':
			b.WriteByte('f')
		case 'ق':
			b.WriteByte('q')
		case 'ك':
			b.WriteByte('k')
		case 'ل':
			b.WriteByte('l')
		case 'م':
			b.WriteByte('m')
		case 'ن':
			b.WriteByte('n')
		case 'ه':
			b.WriteByte('h')
		case 'و':
			b.WriteByte('w')
		case 'ي':
			b.WriteByte('y')
		case 'ء':
			b.WriteByte('\'') // hamza-on-line
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Levenshtein computes the edit distance between two strings using
// the classic O(len(a)*len(b)) DP. Memory is bounded by O(min(len(a),
// len(b))) by rolling the previous row only.
//
// Used by English search to rank near-matches.
func Levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	// Ensure la <= lb so the inner row is the shorter one.
	if la > lb {
		a, b = b, a
		la, lb = lb, la
	}
	prev := make([]int, la+1)
	curr := make([]int, la+1)
	for i := 0; i <= la; i++ {
		prev[i] = i
	}
	for j := 1; j <= lb; j++ {
		curr[0] = j
		for i := 1; i <= la; i++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			del := prev[i] + 1
			ins := curr[i-1] + 1
			sub := prev[i-1] + cost
			best := del
			if ins < best {
				best = ins
			}
			if sub < best {
				best = sub
			}
			curr[i] = best
		}
		prev, curr = curr, prev
	}
	return prev[la]
}

// HighlightSnippet wraps the first occurrence of `match` inside `text`
// with `<mark>…</mark>`. Returns text unchanged if match is empty or
// not found. The matching is case-sensitive for English and uses
// NormalizeArabic for Arabic text.
//
// The returned string is HTML-safe at the caller — server templates
// render snippets via html/template which auto-escapes; the <mark>
// tag is intentionally not escaped (it is added by us, not by the
// user input).
func HighlightSnippet(text, match string) string {
	if match == "" {
		return text
	}
	i := strings.Index(text, match)
	if i < 0 {
		return text
	}
	return text[:i] + "<mark>" + match + "</mark>" + text[i+len(match):]
}

// ── English / Translation word-tokenization ──────────────────
//
// MASAQ glosses look like "the-prayer", "and-restrained",
// "(from)-the-sky-(rain)". Hyphens mark morpheme boundaries within
// a compound word (NOT word boundaries — the whole compound is one
// token in the underlying linguistic unit). Apostrophes mark
// clitics. We tokenize on hyphen AND on whitespace, treating each
// hyphen-separated chunk as its own word.
//
// Examples:
//   "heavy-rain"        → ["heavy", "rain"]         (2 words)
//   "(rain-from)-sky"    → ["rain", "from", "sky"]   (3 words)
//   "and-the-prayer;"    → ["and", "the", "prayer"]  (3 words)
//
// This is the right granularity for word-boundary matching:
// queries like "rain" should match "(rain-from)-sky" as exact-word
// (because "rain" is a token after hyphen-split), but should NOT
// match "(from)-the-sky-(rain)" if the hyphen split doesn't put
// "rain" at a token boundary — i.e., we'd never see "rain" embedded
// in the middle of "from-the-sky-rain" without it being its own
// token.
var enTokenRe = regexp.MustCompile(`[A-Za-z]+(?:'[A-Za-z]+)?`)

// EnTokenize returns the lowercased token list of an English/Translation
// gloss. Used by the English() and Translation() search functions to
// do word-boundary matching.
func EnTokenize(s string) []string {
	return enTokenRe.FindAllString(strings.ToLower(s), -1)
}

// MatchClass is a tiered classification of how a query matches a
// candidate gloss. Lower numbers are better matches and sort first.
//
//   0 — exact full-text match (the whole gloss equals the query, or
//        the query is a single token and the gloss is that single
//        token)
//   1 — exact word-boundary match (every query token is a full
//        word in the gloss, after tokenization on hyphens)
//   2 — prefix match (a gloss word starts with the query and the
//        query is shorter than that word — e.g., "rain" → "raining")
//   3 — substring inside a token (e.g., "rain" → "rainfall") — only
//        when the substring is at the START of a token, never in
//        the middle (so "rain" doesn't match "restrain").
//   4 — fuzzy match (Levenshtein ≤ 2 on a single-word query vs a
//        single gloss word)
//   9 — no useful match
type MatchClass int

const (
	MatchExact        MatchClass = 0
	MatchExactWord    MatchClass = 1
	MatchPrefix       MatchClass = 2
	MatchSubstring    MatchClass = 3
	MatchFuzzy        MatchClass = 4
	MatchNone         MatchClass = 9
)

// String renders a MatchClass as a stable string label (used in
// logs and tests).
func (m MatchClass) String() string {
	switch m {
	case MatchExact:
		return "exact"
	case MatchExactWord:
		return "exact_word"
	case MatchPrefix:
		return "prefix"
	case MatchSubstring:
		return "substr_start"
	case MatchFuzzy:
		return "fuzzy"
	default:
		return "none"
	}
}

// MatchDetail records how a query matched a candidate text.
type MatchDetail struct {
	Class   MatchClass
	Kind    string // same as Class.String()
	Matched string // the gloss word that matched (for debugging)
	Dist    int    // Levenshtein distance (for fuzzy matches)
}

// ClassifyMatch inspects how `query` matches `text` (both already
// lowercased by the caller) and returns the best match class +
// details. The matching strategy is word-boundary: we tokenize the
// text on hyphens + whitespace, then look for the query as a whole
// token, a prefix, or a fuzzy match within a token.
//
// Critical: the function NEVER matches a query as a substring that
// crosses a word boundary in the candidate text. For example,
// "rain" must not match "restrained" (rain is embedded inside
// "restrained" at non-token-start position), nor must it match
// "training" (rain at end), nor "in", "sin", "said" (Levenshtein
// ≤ 2 but the result is misleading for short queries — we use
// AUTO fuzziness plus a 2-char prefix lock so "rain" → "in" is
// rejected).
//
// AUTO fuzziness (Lucene convention): 0 for query length 0..2, 1
// for length 3..5, 2 for length ≥ 6. This keeps the threshold tight
// for short queries where a 2-edit distance can collapse unrelated
// words ("rain" → "in" is a tempting 2-edit match but unrelated).
func ClassifyMatch(query, text string) MatchDetail {
	q := strings.ToLower(strings.TrimSpace(query))
	t := strings.ToLower(strings.TrimSpace(text))
	if q == "" || t == "" {
		return MatchDetail{Class: MatchNone}
	}
	if q == t {
		return MatchDetail{Class: MatchExact, Kind: "exact", Matched: q}
	}
	qTokens := EnTokenize(q)
	tTokens := EnTokenize(t)
	if len(qTokens) == 0 {
		return MatchDetail{Class: MatchNone}
	}
	if len(qTokens) == 1 {
		qt := qTokens[0]
		// Minimum query length: a 1-char query is too short for any
		// meaningful prefix or fuzzy match (every gloss word would
		// be a prefix of "a" or "i", etc.). The caller already
		// filters len(q) < 2, but defence-in-depth: skip prefix
		// matches for very short queries too.
		if len(qt) < 2 {
			return MatchDetail{Class: MatchNone}
		}
		// Class 1: exact word match (after hyphen-split the gloss
		// tokens are independent words).
		for _, tt := range tTokens {
			if qt == tt {
				return MatchDetail{Class: MatchExactWord, Kind: "exact_word", Matched: qt}
			}
		}
		// Class 2: prefix match (text token starts with the query).
		// Catches "rain" → "raining", "rain" → "rained",
		// "rain" → "rainfall", "rain" → "raintime", etc.
		for _, tt := range tTokens {
			if len(tt) > len(qt) && strings.HasPrefix(tt, qt) {
				return MatchDetail{Class: MatchPrefix, Kind: "prefix", Matched: tt}
			}
		}
		// Class 4: fuzzy match with AUTO fuzziness + 2-char prefix
		// lock. The prefix lock rejects "rain" → "in" / "sin" /
		// "said" (all 2 edits away but start with completely
		// different letters); AUTO fuzziness handles the rare
		// cross-typo case like "prayer" → "pryers" (1 edit).
		af := autoFuzziness(len(qt))
		var bestDist int = 99
		var bestMatch string
		for _, tt := range tTokens {
			// Prefix lock: query length ≥ 3 AND gloss word length ≥ 3
			// AND first 2 chars must match. This keeps short fuzzy
			// matches from collapsing unrelated words. For len < 3
			// queries the prefix lock doesn't apply (we have very
			// few 1-2 char queries anyway since len(q) < 2 is
			// filtered out at the caller).
			if len(qt) >= 3 && len(tt) >= 3 && qt[:2] != tt[:2] {
				continue
			}
			d := Levenshtein(qt, tt)
			if d <= af && d < bestDist {
				bestDist = d
				bestMatch = tt
			}
		}
		if bestDist <= af {
			return MatchDetail{Class: MatchFuzzy, Kind: "fuzzy", Matched: bestMatch, Dist: bestDist}
		}
		return MatchDetail{Class: MatchNone}
	}
	// Multi-token query — require all tokens to be present in the
	// gloss as whole words (class 1).
	qset := make(map[string]struct{}, len(qTokens))
	for _, t := range qTokens {
		qset[t] = struct{}{}
	}
	tset := make(map[string]struct{}, len(tTokens))
	for _, t := range tTokens {
		tset[t] = struct{}{}
	}
	for qt := range qset {
		if _, ok := tset[qt]; !ok {
			return MatchDetail{Class: MatchNone}
		}
	}
	return MatchDetail{Class: MatchExactWord, Kind: "exact_words", Matched: joinTokens(qTokens)}
}

// autoFuzziness returns the Lucene "AUTO" fuzziness threshold for a
// query token of the given length. Short queries get tight
// thresholds (so "rain" → "in" doesn't match — that would be
// misleading), longer queries can absorb more edits.
func autoFuzziness(queryLen int) int {
	switch {
	case queryLen <= 2:
		return 0
	case queryLen <= 5:
		return 1
	default:
		return 2
	}
}

// joinTokens renders a token slice as a single string for logging /
// debug display.
func joinTokens(toks []string) string {
	out := ""
	for i, t := range toks {
		if i > 0 {
			out += " "
		}
		out += t
	}
	return out
}
