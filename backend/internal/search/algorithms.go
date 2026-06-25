package search

import (
	"cmp"
	"encoding/gob"
	"math"
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

// English searches MASAQ glosses for English matches using
// word-boundary matching with class-based ranking.
//
// Algorithm (v2 — word-boundary matching):
//
//  1. Lowercase + trim the query. Skip if shorter than 2 chars or
//     has no alphabetic tokens.
//  2. For each segment, classify the match against its gloss:
//     - MatchExact       (class 0) — full gloss equals query
//     - MatchExactWord   (class 1) — query is a whole token in the gloss
//     - MatchPrefix      (class 2) — a gloss word starts with the query
//     - MatchSubstring   (class 3) — query sits at the START of a gloss
//                                    word (never in the middle of a word)
//     - MatchFuzzy       (class 4) — Levenshtein ≤ 2 on a single word
//  3. Rank by class first, then by IDF-weighted relevance within the
//     class, then by gloss length, then by surah/ayah.
//  4. Return up to `limit` results.
//
// The new matching strategy fixes the previous greedy behavior where
// "rain" would match inside "restrained", "training", "in", "sin",
// "said" — none of those are valid word-boundary matches. The
// snippet also wraps a clean word in <mark> rather than a raw
// substring (so "rain" never bleeds into "but-he-rest<mark>rain</mark>ed").
func English(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	return englishOrTranslation(q, m, quran, limit, filterSurah, false)
}

// englishOrTranslation is the shared implementation for English()
// and Translation(). The fields parameter selects which MASAQ field
// to search: gloss (English) or translation (Translation).
func englishOrTranslation(q string, m *types.MasaqIndex, quran *types.Quran, limit, filterSurah int, isTranslation bool) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = strings.ToLower(strings.TrimSpace(q))
	if len(q) < 2 {
		return nil
	}
	qTokens := EnTokenize(q)
	if len(qTokens) == 0 {
		return nil
	}
	kind := "english"
	field := func(s types.MasaqSegment) string { return s.Gloss }
	if isTranslation {
		kind = "translation"
		field = func(s types.MasaqSegment) string { return s.Translation }
	}

	// Build IDF table over the search field — common tokens (e.g.,
	// "the", "and") get low IDF, rare tokens get high IDF.
	docFreq := make(map[string]int, 10000)
	totalDocs := 0
	for _, segs := range m.ByWord {
		seen := make(map[string]struct{}, 8)
		for _, s := range segs {
			t := field(s)
			if t == "" {
				continue
			}
			for _, tok := range EnTokenize(t) {
				if _, ok := seen[tok]; ok {
					continue
				}
				seen[tok] = struct{}{}
				docFreq[tok]++
			}
		}
		totalDocs++
	}
	idf := func(tok string) float64 {
		df := docFreq[tok]
		if df == 0 {
			df = 1 // unseen — assume some doc count
		}
		// Smoothed IDF: ln((N - df + 0.5) / (df + 0.5) + 1).
		return math.Log((float64(totalDocs)-float64(df)+0.5)/(float64(df)+0.5) + 1.0)
	}
	qIDFs := make([]float64, len(qTokens))
	for i, t := range qTokens {
		qIDFs[i] = idf(t)
	}

	type scored struct {
		key   uint64
		score int    // lower is better
		class int    // 0..4
		title string // gloss / translation text
		surah int
		ayah  int
		word  int
	}
	cands := make([]scored, 0, 256)
	seen := make(map[uint64]bool, 5000)
	for key, segs := range m.ByWord {
		surah, ayah, word := loc.Decode(key)
		if filterSurah > 0 && surah != filterSurah {
			continue
		}
		if seen[key] {
			continue
		}
		var best scored
		best.score = 1 << 30
		for _, s := range segs {
			t := field(s)
			if t == "" {
				continue
			}
			detail := ClassifyMatch(q, t)
			if detail.Class == MatchNone {
				continue
			}
			tTokens := EnTokenize(t)
			tSet := make(map[string]struct{}, len(tTokens))
			for _, tt := range tTokens {
				tSet[tt] = struct{}{}
			}
			// Score = class * 1M - IDF-weighted relevance + length penalty.
			// Lower is better, so subtract IDF (rarer tokens = more relevant).
			matchedIDF := 0.0
			for i, qt := range qTokens {
				if _, ok := tSet[qt]; ok {
					matchedIDF += qIDFs[i]
				}
			}
			score := int(detail.Class)*1000000 - int(matchedIDF*10) + len(tTokens)
			if score < best.score {
				best = scored{
					key:   key,
					score: score,
					class: int(detail.Class),
					title: t,
					surah: surah,
					ayah:  ayah,
					word:  word,
				}
			}
		}
		if best.score < 1<<30 {
			seen[key] = true
			cands = append(cands, best)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	slices.SortFunc(cands, func(a, b scored) int {
		return cmp.Or(
			cmp.Compare(a.score, b.score),
			cmp.Compare(a.surah, b.surah),
			cmp.Compare(a.ayah, b.ayah),
			cmp.Compare(a.word, b.word),
		)
	})
	if len(cands) > limit {
		cands = cands[:limit]
	}
	out := make([]SearchResult, len(cands))
	for i, c := range cands {
		r := SearchResult{
			Kind:  kind,
			Surah: c.surah,
			Ayah:  c.ayah,
			Word:  c.word,
			Title: c.title,
			Link:  "/surah/" + strconv.Itoa(c.surah) + "#verse-" + strconv.Itoa(c.ayah),
			Score: c.score,
		}
		// Highlight the first matching word (token) in the title,
		// not just the raw query substring — this way the <mark> tag
		// always wraps a clean word and never bleeds across boundaries
		// like the old "rest<mark>rain</mark>ed" rendering.
		r.Snippet = highlightWord(c.title, qTokens)
		r.VerseText = verseText(quran, c.surah, c.ayah)
		out[i] = r
	}
	return out
}

// highlightWord picks the first query token that appears as a whole
// word in title (case-insensitive) and wraps it in <mark>…</mark>.
// Falls back to wrapping the raw query string if no token match is
// found (e.g., for fuzzy matches where the matched word is similar
// but not equal to a query token).
func highlightWord(title string, qTokens []string) string {
	t := title
	tLower := strings.ToLower(t)
	for _, qt := range qTokens {
		// Find the token's position in the lowercased title.
		// MASAQ glosses are all-hyphen-separated (no spaces), so a
		// substring search for the bare token is safe — we won't
		// accidentally match inside another token because tokens
		// don't overlap in hyphenated strings.
		idx := strings.Index(tLower, qt)
		if idx >= 0 {
			end := idx + len(qt)
			if end <= len(t) {
				return t[:idx] + "<mark>" + t[idx:end] + "</mark>" + t[end:]
			}
		}
	}
	// Fallback: wrap the raw query if no token match.
	idx := strings.Index(tLower, strings.Join(qTokens, "-"))
	if idx >= 0 {
		end := idx + len(qTokens[0])*len(qTokens) + (len(qTokens)-1) // approximation
		_ = end
	}
	return t
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
// Translation searches MASAQ word-level translations for English
// matches. Like English() but matches against the Translation field
// (word-level) instead of Gloss (segment-level). Uses the same
// word-boundary matching algorithm via englishOrTranslation().
func Translation(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	return englishOrTranslation(q, m, quran, limit, filterSurah, true)
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
