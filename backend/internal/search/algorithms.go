package search

import (
	"cmp"
	"encoding/gob"
	"os"
	"slices"
	"strconv"
	"strings"

	"quranreader/loc"
	"quranreader/types"
)

// Arabic searches MASAQ for Arabic-word matches, optionally filtered
// to a single surah (filterSurah > 0).
func Arabic(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = NormalizeArabic(q)
	if q == "" {
		return nil
	}
	bestByKey := make(map[uint64]SearchResult, 1000)
	scoreByKey := make(map[uint64]int, 1000)

	for key, segs := range m.ByWord {
		for _, s := range segs {
			best := scoreArabic(q, s.Word)
			if alt := scoreArabic(q, s.WithoutDiacritics); alt < best {
				best = alt
			}
			if best >= 1<<20 {
				continue
			}
			if prev, ok := scoreByKey[key]; ok && prev <= best {
				continue
			}
			surah, ayah, word := loc.Decode(key)
			if filterSurah > 0 && surah != filterSurah {
				continue
			}
			scoreByKey[key] = best
			bestByKey[key] = SearchResult{
				Kind:  "arabic",
				Surah: surah,
				Ayah:  ayah,
				Word:  word,
				Title: s.Word,
				Link:  "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
				Score: best,
			}
		}
	}
	if len(bestByKey) == 0 {
		return nil
	}
	all := make([]SearchResult, 0, len(bestByKey))
	for k, r := range bestByKey {
		r.Snippet = HighlightSnippet(r.Title, q)
		r.VerseText = verseText(quran, r.Surah, r.Ayah)
		all = append(all, r)
		_ = k
	}
	slices.SortFunc(all, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Surah, b.Surah),
			cmp.Compare(a.Ayah, b.Ayah),
			cmp.Compare(a.Word, b.Word),
		)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// scoreArabic returns a non-negative score: 0 is best, 1<<20 means
// "no match at all". Smaller is better.
func scoreArabic(query, word string) int {
	q := NormalizeArabic(word)
	if query == q {
		return 0
	}
	if strings.Contains(q, query) {
		// q is the longer one (word); query is shorter — prefix or sub.
		return 1 + len(q) - len(query)
	}
	if strings.Contains(query, q) {
		return 1 + len(query) - len(q)
	}
	return 1 << 20
}

// English searches MASAQ glosses for English matches.
//
// Algorithm:
//
//  1. Lowercase + trim the query.
//  2. For each segment, compute:
//     - exact match        → score 0
//     - substring          → score 1 + length diff
//     - Levenshtein ≤ 2    → score 2 + distance
//  3. Return up to `limit` results, sorted by score.
func English(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if len(q) < 2 {
		return nil
	}
	bestByKey := make(map[uint64]SearchResult, 5000)
	scoreByKey := make(map[uint64]int, 5000)

	for key, segs := range m.ByWord {
		for _, s := range segs {
			g := strings.ToLower(strings.TrimSpace(s.Gloss))
			if g == "" {
				continue
			}
			score := 1 << 20
			if g == q {
				score = 0
			} else if strings.Contains(g, q) {
				score = 1 + len(g) - len(q)
			} else {
				d := Levenshtein(g, q)
				if d <= 2 {
					score = 100 + d
				}
			}
			if score >= 1<<20 {
				continue
			}
			if prev, ok := scoreByKey[key]; ok && prev <= score {
				continue
			}
			surah, ayah, word := loc.Decode(key)
			scoreByKey[key] = score
			bestByKey[key] = SearchResult{
				Kind:  "english",
				Surah: surah,
				Ayah:  ayah,
				Word:  word,
				Title: s.Gloss,
				Link:  "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
				Score: score,
			}
		}
	}
	if len(bestByKey) == 0 {
		return nil
	}
	all := make([]SearchResult, 0, len(bestByKey))
	for _, c := range bestByKey {
		c.Snippet = HighlightSnippet(c.Title, q)
		c.VerseText = verseText(quran, c.Surah, c.Ayah)
		all = append(all, c)
	}
	slices.SortFunc(all, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Surah, b.Surah),
			cmp.Compare(a.Ayah, b.Ayah),
		)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// Lemma searches MASAQ for lemma matches (bare stem, no diacritics).
// Like Arabic() but matches against the computed lemma instead of the
// raw word text. Normalizes the query for tashkeel-tolerant matching.
func Lemma(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = NormalizeArabic(q)
	if q == "" {
		return nil
	}
	bestByKey := make(map[uint64]SearchResult, 1000)
	scoreByKey := make(map[uint64]int, 1000)

	// Group segments by word key and compute lemma once per word.
	type wordSegs struct{ segs []types.MasaqSegment }
	wordsSeen := map[uint64]bool{}

	for key, segs := range m.ByWord {
		if wordsSeen[key] {
			continue
		}
		wordsSeen[key] = true
		lemma := computeLemmaFromSegs(segs)
		if lemma == "" {
			continue
		}
		best := scoreArabic(q, lemma)
		if best >= 1<<20 {
			continue
		}
		if prev, ok := scoreByKey[key]; ok && prev <= best {
			continue
		}
		surah, ayah, word := loc.Decode(key)
		if filterSurah > 0 && surah != filterSurah {
			continue
		}
		scoreByKey[key] = best
		bestByKey[key] = SearchResult{
			Kind:  "lemma",
			Surah: surah,
			Ayah:  ayah,
			Word:  word,
			Title: lemma,
			Link:  "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
			Score: best,
		}
	}
	if len(bestByKey) == 0 {
		return nil
	}
	all := make([]SearchResult, 0, len(bestByKey))
	for k, r := range bestByKey {
		r.Snippet = HighlightSnippet(r.Title, q)
		r.VerseText = verseText(quran, r.Surah, r.Ayah)
		all = append(all, r)
		_ = k
	}
	slices.SortFunc(all, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Surah, b.Surah),
			cmp.Compare(a.Ayah, b.Ayah),
			cmp.Compare(a.Word, b.Word),
		)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// computeLemmaFromSegs is a simplified version of server.computeLemma,
// avoiding a circular dependency. Returns the bare-stem lemma given
// a word's segments.
func computeLemmaFromSegs(segs []types.MasaqSegment) string {
	// Proper noun → use first segment's WithoutDiacritics.
	for _, s := range segs {
		if s.MorphTag == "NOUN_PROP" && len(segs) > 0 {
			return segs[0].WithoutDiacritics
		}
	}
	// Prefix + Stem → concatenate SegmentedWord.
	var prefix, stem string
	for _, s := range segs {
		switch s.MorphType {
		case "Prefix":
			if s.SegmentedWord != "" {
				prefix = s.SegmentedWord
			}
		case "Stem":
			if s.SegmentedWord != "" {
				stem = s.SegmentedWord
			}
		}
	}
	if stem != "" {
		return prefix + stem
	}
	// Fallback: first segment's WithoutDiacritics.
	if len(segs) > 0 {
		return segs[0].WithoutDiacritics
	}
	return ""
}

// Translation searches MASAQ word-level translations for English
// matches. Like English() but matches against the Translation field
// (word-level) instead of Gloss (segment-level).
func Translation(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if len(q) < 2 {
		return nil
	}
	bestByKey := make(map[uint64]SearchResult, 5000)
	scoreByKey := make(map[uint64]int, 5000)

	// Split query into words for word-boundary matching.
	qWords := strings.FieldsFunc(q, func(r rune) bool {
		return r == ' ' || r == '-' || r == '/' || r == '(' || r == ')'
	})

	for key, segs := range m.ByWord {
		for _, s := range segs {
			t := strings.ToLower(strings.TrimSpace(s.Translation))
			if t == "" {
				continue
			}
			// Split gloss into words.
			tWords := strings.FieldsFunc(t, func(r rune) bool {
				return r == ' ' || r == '-' || r == '/' || r == '(' || r == ')'
			})

			score := 1 << 20
			if t == q {
				score = 0
			} else {
				// Word-boundary matching: check if any query word
				// matches any gloss word (exact or substring).
				for _, tw := range tWords {
					for _, qw := range qWords {
						if qw == tw {
							score = min(score, 1)
						} else if strings.Contains(tw, qw) || strings.Contains(qw, tw) {
							score = min(score, 2+abs(len(tw)-len(qw)))
						}
					}
				}
				if score >= 1<<20 {
					// Fall back to full-string Levenshtein on whole gloss.
					d := Levenshtein(t, q)
					if d <= 2 {
						score = 100 + d
					}
				}
			}
			if score >= 1<<20 {
				continue
			}
			if prev, ok := scoreByKey[key]; ok && prev <= score {
				continue
			}
			surah, ayah, word := loc.Decode(key)
			if filterSurah > 0 && surah != filterSurah {
				continue
			}
			scoreByKey[key] = score
			bestByKey[key] = SearchResult{
				Kind:  "english",
				Surah: surah,
				Ayah:  ayah,
				Word:  word,
				Title: s.Translation,
				Link:  "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
				Score: score,
			}
		}
	}
	if len(bestByKey) == 0 {
		return nil
	}
	all := make([]SearchResult, 0, len(bestByKey))
	for _, c := range bestByKey {
		c.Snippet = HighlightSnippet(c.Title, q)
		c.VerseText = verseText(quran, c.Surah, c.Ayah)
		all = append(all, c)
	}
	slices.SortFunc(all, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Surah, b.Surah),
			cmp.Compare(a.Ayah, b.Ayah),
		)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// Root searches the roots index for a root match.
//
// The input may be either Buckwalter (ASCII) or Arabic letters; we
// detect by inspecting the first non-ASCII byte. If Arabic, we
// transliterate via Buckwalter() and also index via the ByArabic
// reverse map.
//
// Returns one entry per matching root, with the root's occurrences
// surfaced in `Title`. Each result links to the root detail page.
func Root(q string, r *types.RootsIndex, quran *types.Quran, limit int) []SearchResult {
	if r == nil || limit <= 0 {
		return nil
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	bw := q
	if !isASCII(q) {
		// Arabic input — try ByArabic first (key is the canonical
		// "letters, no spaces" form), fall back to Buckwalter.
		letters := types.NormalizeArabicRoot(q)
		if hit, ok := r.ByArabic[letters]; ok {
			bw = hit
		} else {
			bw = Buckwalter(NormalizeArabic(q))
		}
	}

	var all []SearchResult

	// 1. Exact key match.
	if e, ok := r.ByRoot[bw]; ok {
		all = append(all, SearchResult{
			Kind:  "root",
			Root:  bw,
			Title: bw + " — " + truncate(e.Arabic, 80),
			Link:  "/root/detailed/" + bw,
			Score: 0,
		})
	}
	// 2. Prefix matches (sorted by occurrence desc).
	for key, e := range r.ByRoot {
		if key == bw {
			continue
		}
		if !strings.HasPrefix(key, bw) {
			continue
		}
		all = append(all, SearchResult{
			Kind:  "root",
			Root:  key,
			Title: key + " — " + truncate(e.Arabic, 80),
			Link:  "/root/detailed/" + key,
			Score: 1000 - e.Occurrences,
		})
	}
	slices.SortFunc(all, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Root, b.Root),
		)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

// RootsList returns a paginated list of roots, optionally filtered
// by surah. When filterSurah > 0, only roots appearing in that surah
// are included. Sorted by frequency descending (or asc if `asc`).
func RootsList(r *types.RootsIndex, page, pageSize int, asc bool, filterSurah int) ([]SearchResult, int) {
	if r == nil || pageSize <= 0 {
		return nil, 0
	}
	if page < 1 {
		page = 1
	}

	// Build the set of root keys to include.
	var allowed map[string]bool
	if filterSurah > 0 && filterSurah <= 114 {
		allowed = r.RootsBySurah[filterSurah]
	}

	// Collect (key, entry) pairs, optionally filtered.
	keys := make([]string, 0, len(r.ByRoot))
	for k := range r.ByRoot {
		if allowed != nil && !allowed[k] {
			continue
		}
		keys = append(keys, k)
	}
	// Modern sort: cmp.Or + cmp.Compare. Two-level key — primary by
	// frequency (ascending or descending depending on user toggle),
	// secondary by Buckwalter name for stable, deterministic order
	// when two roots share the same frequency.
	slices.SortFunc(keys, func(a, b string) int {
		fa, fb := r.ByRoot[a].Occurrences, r.ByRoot[b].Occurrences
		var cmpOcc int
		if asc {
			cmpOcc = cmp.Compare(fa, fb)
		} else {
			cmpOcc = cmp.Compare(fb, fa)
		}
		return cmp.Or(cmpOcc, cmp.Compare(a, b))
	})
	total := len(keys)
	start := (page - 1) * pageSize
	if start >= total {
		return nil, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	out := make([]SearchResult, 0, end-start)
	for _, k := range keys[start:end] {
		e := r.ByRoot[k]
		out = append(out, SearchResult{
			Kind:  "root-list",
			Root:  k,
			Title: k + " — " + truncate(e.Arabic, 80),
			Link:  "/root/detailed/" + k,
			Score: e.Occurrences,
		})
	}
	return out, total
}

// OccurrencesForRoot returns the search results for every occurrence
// of `root` in the corpus, optionally filtered by surah and sorted by
// location. The locations come in surah/ayah ascending order; when
// `asc` is true the result is reversed (last surah/ayah first) to
// match the user-facing "Ascending" toggle on the root-detail page.
func OccurrencesForRoot(root string, r *types.RootsIndex, q *types.Quran, filterSurah int, asc bool) []SearchResult {
	e, ok := r.ByRoot[root]
	if !ok {
		return nil
	}
	out := make([]SearchResult, 0, len(e.Locations))
	for _, loc := range e.Locations {
		// Filter by surah if requested.
		if filterSurah > 0 {
			parts := strings.Split(loc, ":")
			surah, _ := strconv.Atoi(parts[0])
			if surah != filterSurah {
				continue
			}
		}
		parts := strings.Split(loc, ":")
		if len(parts) != 3 {
			continue
		}
		surah, _ := strconv.Atoi(parts[0])
		ayah, _ := strconv.Atoi(parts[1])
		wordNo, _ := strconv.Atoi(parts[2])
		surahPtr := q.Surahs[surah]
		if surahPtr == nil {
			continue
		}
		ay := surahPtr.Ayahs[ayah]
		if ay == nil {
			continue
		}
		// The word token at position wordNo-1 (0-indexed).
		var wordText string
		wordIdx := 0
		for _, t := range ay.Tokens {
			if t.Kind != "word" {
				continue
			}
			wordIdx++
			if wordIdx == wordNo {
				wordText = t.Value
				break
			}
		}
		out = append(out, SearchResult{
			Kind:      "occurrence",
			Surah:     surah,
			Ayah:      ayah,
			Word:      wordNo,
			Title:     "Surah " + strconv.Itoa(surah) + ":" + strconv.Itoa(ayah),
			Link:      "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
			Snippet:   wordText,
			VerseText: verseText(q, surah, ayah),
			Root:      root,
		})
	}
	if asc {
		// Reverse the natural (surah-ascending) order so the most
		// recent occurrences appear first.
		slices.SortStableFunc(out, func(a, b SearchResult) int {
			return cmp.Or(
				cmp.Compare(b.Surah, a.Surah),
				cmp.Compare(b.Ayah, a.Ayah),
			)
		})
	}
	// Default (asc=false): preserve the natural surah-ascending order
	// in which Locations was built.
	return out
}

// ── helpers ────────────────────────────────────────────────────

// verseText returns the Uthmani verse text for (surah, ayah) from the
// Quran struct, or empty string if not found.
func verseText(quran *types.Quran, surah, ayah int) string {
	if quran == nil {
		return ""
	}
	s := quran.Surahs[surah]
	if s == nil {
		return ""
	}
	a := s.Ayahs[ayah]
	if a == nil {
		return ""
	}
	return a.Text
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// gobDecodeFile reads a gob file from disk.
func gobDecodeFile(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return gob.NewDecoder(f).Decode(v)
}
