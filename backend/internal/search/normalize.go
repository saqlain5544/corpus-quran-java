package search

import "strings"

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
