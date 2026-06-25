package search

import (
	"cmp"
	"encoding/gob"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"quranreader/loc"
	"quranreader/types"
)

// Arabic searches MASAQ for Arabic-word matches, optionally filtered
// to a single surah (filterSurah > 0).
//
// Algorithm (post-inverted-index, see /tmp/algo-research/01_data_structures.md):
//
//  1. NormalizeArabic the query.
//  2. Look up `m.ByArabicFormPostings[normalized query]` for exact
//     matches (class 0).
//  3. Scan the vocabulary to find words where the query is a substring
//     of the vocab word OR vice versa (class 1). Skip vocab entries
//     shorter than 2 chars to avoid noise from single-letter morphemes
//     (like 'ل' the definite article prefix, which would match every
//     query containing 'ل' anywhere).
//  4. Union the candidate postings and rank by (class asc, surah asc,
//     ayah asc, word asc).
//
// Falls back to the O(N) linear scan via arabicOrLemmaLinear when the
// MasaqIndex has no inverted index (e.g. constructed manually in tests).
func Arabic(q string, m *types.MasaqIndex, quran *types.Quran, limit int, filterSurah int) []SearchResult {
	return arabicOrLemma(q, m, quran, limit, filterSurah, false)
}

// Lemma is the shared implementation used by Arabic() and the Lemma()
// search. The `isLemma` flag selects whether to search against the
// computed lemma form instead of the raw word.
func arabicOrLemma(q string, m *types.MasaqIndex, quran *types.Quran, limit, filterSurah int, isLemma bool) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = NormalizeArabic(q)
	// Reject very short queries — single-character Arabic queries
	// match too many things (the morpheme 'ل' alone appears in
	// thousands of words as the definite article prefix). Mirror
	// the English-search threshold (len < 2 chars returns nil).
	// We use utf8.RuneCountInString because Arabic chars are
	// multi-byte in UTF-8 — len("ل") is 2 bytes, not 1 rune.
	if q == "" || utf8.RuneCountInString(q) < 2 {
		return nil
	}
	kind := "arabic"
	postings := m.ByArabicFormPostings
	if isLemma {
		kind = "lemma"
		// The Lemma index is built separately in load.go — we use the
		// Arabic form postings as a fallback for now (Lemma is
		// computed at search time).
		postings = m.ByArabicFormPostings
	}
	if postings == nil {
		return arabicOrLemmaLinear(q, m, quran, limit, filterSurah, isLemma)
	}

	// Build candidate set.
	cands := make(map[uint64]int, 4096) // doc_id → class
	// 1. Exact match
	for _, key := range postings[q] {
		cands[key] = 0
	}
	// 2. Bidirectional substring match. Both directions of containment
	//    are treated as class 1. We skip single-character vocab entries
	//    to avoid spurious matches (e.g., a 1-char query matches every
	//    word containing that character).
	if len(q) >= 2 {
		for v, posts := range postings {
			if v == q || len(v) < 2 {
				continue
			}
			matched := false
			if len(v) > len(q) && strings.Contains(v, q) {
				matched = true
			} else if len(q) > len(v) && strings.Contains(q, v) {
				matched = true
			}
			if !matched {
				continue
			}
			for _, key := range posts {
				if _, ok := cands[key]; !ok || cands[key] > 1 {
					cands[key] = 1
				}
			}
		}
	}
	out := make([]SearchResult, 0, len(cands))
	for key, cls := range cands {
		surah, ayah, word := loc.Decode(key)
		if filterSurah > 0 && surah != filterSurah {
			continue
		}
		// For display: use the word's stored imla text (first non-empty
		// gloss segment). For Lemma search, use the computed lemma.
		segs := m.ByWord[key]
		var title string
		if isLemma {
			title = computeLemmaFromSegs(segs)
		} else {
			for _, s := range segs {
				if s.Word != "" {
					title = s.Word
					break
				}
			}
		}
		out = append(out, SearchResult{
			Kind:  kind,
			Surah: surah,
			Ayah:  ayah,
			Word:  word,
			Title: title,
			Link:  "/surah/" + strconv.Itoa(surah) + "#verse-" + strconv.Itoa(ayah),
			Score: cls * 1000, // rank-encoded for API stability
		})
	}
	// Sort: class asc, surah asc, ayah asc, word asc
	slices.SortFunc(out, func(a, b SearchResult) int {
		return cmp.Or(
			cmp.Compare(a.Score, b.Score),
			cmp.Compare(a.Surah, b.Surah),
			cmp.Compare(a.Ayah, b.Ayah),
			cmp.Compare(a.Word, b.Word),
		)
	})
	if len(out) > limit {
		out = out[:limit]
	}
	for i := range out {
		out[i].Snippet = HighlightSnippet(out[i].Title, q)
		out[i].VerseText = verseText(quran, out[i].Surah, out[i].Ayah)
	}
	return out
}

// arabicOrLemmaLinear is the legacy O(N) fallback for tests / data
// paths that build a MasaqIndex without the inverted index.
func arabicOrLemmaLinear(q string, m *types.MasaqIndex, quran *types.Quran, limit, filterSurah int, isLemma bool) []SearchResult {
	if m == nil || limit <= 0 {
		return nil
	}
	q = NormalizeArabic(q)
	// Reject very short queries — see arabicOrLemma for rationale.
	if q == "" || len(q) < 2 {
		return nil
	}
	kind := "arabic"
	if isLemma {
		kind = "lemma"
	}
	bestByKey := make(map[uint64]SearchResult, 1000)
	scoreByKey := make(map[uint64]int, 1000)

	for key, segs := range m.ByWord {
		var best int = 1 << 30
		for _, s := range segs {
			cand := scoreArabic(q, s.Word)
			if alt := scoreArabic(q, s.WithoutDiacritics); alt < cand {
				cand = alt
			}
			if cand < best {
				best = cand
			}
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
		var title string
		if isLemma {
			title = computeLemmaFromSegs(segs)
		} else {
			for _, s := range segs {
				if s.Word != "" {
					title = s.Word
					break
				}
			}
		}
		bestByKey[key] = SearchResult{
			Kind:  kind,
			Surah: surah,
			Ayah:  ayah,
			Word:  word,
			Title: title,
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
//
// Algorithm (post-inverted-index, see /tmp/algo-research/01_data_structures.md):
//
//  1. Tokenize the query.
//  2. Use the pre-built inverted index (m.ByEnTokenPostings or
//     m.ByTranslationPostings) to expand candidate docs:
//     - Literal postings (union over query tokens)
//     - Prefix postings (vocabulary words starting with a query token)
//     - Fuzzy postings (vocabulary words within edit distance ≤ auto-fuzz)
//  3. Classify each candidate against the stored gloss text. Keep the
//     BEST match class per doc (a word may have multiple segments with
//     different glosses — the best one wins).
//  4. Rank by (class asc, BM25 desc, surah asc, ayah asc, word asc).
//
// The old implementation scanned all 77K segments per query and
// recomputed IDF each time. This one expands to (literal ∪ prefix ∪
// fuzzy) postings, which is bounded by the vocabulary size (~5K)
// not the corpus size (~77K). BM25 stats are pre-computed.
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
	postings := m.ByEnTokenPostings
	if isTranslation {
		kind = "translation"
		postings = m.ByTranslationPostings
	}
	// If no inverted index (e.g. legacy data path or tests that
	// build a MasaqIndex directly), fall back to the linear scan.
	if postings == nil {
		return englishOrTranslationLinear(q, m, quran, limit, filterSurah, isTranslation)
	}

	// Candidate expansion via the inverted index.
	var bkTree types.BKTreeIface
	if isTranslation {
		bkTree = m.TranslationBKTree
	} else {
		bkTree = m.GlossBKTree
	}
	candidates := expandPostings(postings, bkTree, qTokens)

	// Pre-computed BM25 stats (cached on MasaqIndex at LoadAll time).
	// Old code rebuilt these per request (~10ms saved per query).
	idxTotalDocs := len(m.ByWord)
	idxAvgDocLen := m.GlossAvgDocLen
	if isTranslation {
		idxAvgDocLen = m.TranslationAvgDocLen
	}

	// For each candidate doc, find the best match class.
	type scored struct {
		key   uint64
		class int     // 0..4
		bm25  float64 // higher is better — we'll negate for sort
		title string
		surah int
		ayah  int
		word  int
	}
	cands := make([]scored, 0, len(candidates))
	for _, key := range candidates {
		surah, ayah, word := loc.Decode(key)
		if filterSurah > 0 && surah != filterSurah {
			continue
		}
		// Pick the gloss for classification. We use the first non-empty
		// gloss from any segment for this LocKey.
		segs := m.ByWord[key]
		var gloss string
		for _, s := range segs {
			var t string
			if isTranslation {
				t = s.Translation
			} else {
				t = s.Gloss
			}
			if t != "" {
				gloss = t
				break
			}
		}
		if gloss == "" {
			continue
		}
		detail := ClassifyMatch(q, gloss)
		if detail.Class == MatchNone {
			continue
		}
		// Compute BM25 over the doc's stored tokens.
		docTokens := EnTokenize(gloss)
		bm25 := bm25ScoreForDocWithText(postings, qTokens, docTokens, idxAvgDocLen, idxTotalDocs)
		cands = append(cands, scored{
			key:   key,
			class: int(detail.Class),
			bm25:  bm25,
			title: gloss,
			surah: surah,
			ayah:  ayah,
			word:  word,
		})
	}
	if len(cands) == 0 {
		return nil
	}
	// Primary sort: class asc (lower = better match).
	// Secondary sort: BM25 desc (higher = more relevant).
	// Tertiary: deterministic location order.
	slices.SortFunc(cands, func(a, b scored) int {
		if a.class != b.class {
			return cmp.Compare(a.class, b.class)
		}
		// BM25 higher is better; we want it descending, so negate.
		if math.Abs(a.bm25-b.bm25) > 1e-9 {
			if a.bm25 > b.bm25 {
				return -1
			}
			return 1
		}
		return cmp.Or(
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
			Score: c.class*1000 - int(c.bm25), // rank-encoded score for API
		}
		r.Snippet = highlightWord(c.title, qTokens)
		r.VerseText = verseText(quran, c.surah, c.ayah)
		out[i] = r
	}
	return out
}

// englishOrTranslationLinear is the legacy O(N) fallback used when
// no inverted index is available (e.g. unit tests that build a
// MasaqIndex directly via a constructor). It produces the same
// results as the inverted-index path but scans all 77K segments per
// query.
func englishOrTranslationLinear(q string, m *types.MasaqIndex, quran *types.Quran, limit, filterSurah int, isTranslation bool) []SearchResult {
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
	type scored struct {
		key   uint64
		score int
		class int
		title string
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
			idfSum := 0.0
			for _, qt := range qTokens {
				if _, ok := tSet[qt]; ok {
					idfSum += math.Log(float64(len(m.ByWord)+1) / float64(seenCount(m, qt, field)+1))
				}
			}
			score := int(detail.Class)*1000000 - int(idfSum*10) + len(tTokens)
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
		r.Snippet = highlightWord(c.title, qTokens)
		r.VerseText = verseText(quran, c.surah, c.ayah)
		out[i] = r
	}
	return out
}

// seenCount counts the number of docs containing token `tok` in
// the given field. O(N) — only used in the legacy linear path.
func seenCount(m *types.MasaqIndex, tok string, field func(types.MasaqSegment) string) int {
	n := 0
	seen := make(map[uint64]struct{})
	for key, segs := range m.ByWord {
		if _, dup := seen[key]; dup {
			continue
		}
		for _, s := range segs {
			for _, t := range EnTokenize(field(s)) {
				if t == tok {
					seen[key] = struct{}{}
					n++
					break
				}
			}
		}
	}
	return n
}

// expandPostings returns the union of postings for all query tokens,
// extended with vocabulary words that share a prefix or are within
// AUTO-fuzziness edit distance of any query token. Returns the
// candidate set as a deduplicated, sorted []uint64 slice.
//
// Complexity:
//   - Literal: O(Σ|posting_lit|)
//   - Prefix: O(V) scan of vocabulary (acceptable for V ≈ 5K)
//   - Fuzzy: O(c^k · log V) via the BK-tree if `bkTree != nil`,
//            else falls back to O(V) linear scan
//
// When the corpus grows past ~50K unique tokens, swap in a proper
// radix tree for prefix queries too — see
// /tmp/algo-research/01_data_structures.md section 2.
func expandPostings(postings map[string][]uint64, bkTree types.BKTreeIface, qTokens []string) []uint64 {
	if len(postings) == 0 {
		return nil
	}
	// Map keyed by LocKey to dedupe candidates from multiple postings.
	candSet := make(map[uint64]struct{}, 4096)
	for _, qt := range qTokens {
		for _, key := range postings[qt] {
			candSet[key] = struct{}{}
		}
	}
	// Prefix expansion: any vocab word starting with a query token
	// (and longer) is a candidate via the PREFIX class. Linear scan
	// over V — acceptable for V ≈ 5K; would benefit from a radix
	// tree at V > ~50K.
	for _, qt := range qTokens {
		if len(qt) < 2 {
			continue
		}
		for v, posts := range postings {
			if v == qt {
				continue
			}
			if len(v) > len(qt) && hasPrefix(v, qt) {
				for _, key := range posts {
					candSet[key] = struct{}{}
				}
			}
		}
	}
	// Fuzzy expansion: vocab words within AUTO-fuzz edit distance
	// of the query token, with a 2-char prefix lock, are candidates
	// via the FUZZY class.
	for _, qt := range qTokens {
		if len(qt) < 3 {
			continue
		}
		af := autoFuzziness(len(qt))
		pre2 := qt[:2]
		if bkTree != nil {
			// Sub-linear via BK-tree. Query returns words within
			// distance ≤ af; we then filter by the 2-char prefix
			// lock (the BK-tree prune doesn't know about it).
			matches := bkTree.Query(qt, af)
			for _, m := range matches {
				if m.Word == qt {
					continue
				}
				if len(m.Word) < 3 || m.Word[:2] != pre2 {
					continue
				}
				for _, key := range postings[m.Word] {
					candSet[key] = struct{}{}
				}
			}
		} else {
			// Linear fallback — same semantics, O(V) per query.
			for v, posts := range postings {
				if v == qt {
					continue
				}
				if len(v) < 3 || v[:2] != pre2 {
					continue
				}
				if abs(len(v)-len(qt)) > af {
					continue
				}
				if Levenshtein(qt, v) <= af {
					for _, key := range posts {
						candSet[key] = struct{}{}
					}
				}
			}
		}
	}
	out := make([]uint64, 0, len(candSet))
	for k := range candSet {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// hasPrefix is a thin alias for strings.HasPrefix — inlined to keep
// this file's hot path free of `strings` lookups for the trim cases.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// bm25ScoreForDocWithText computes the BM25 score of a doc against
// query tokens, given the doc's pre-tokenized text and the corpus
// `postings` map (for IDF lookup).
//
// BM25 (Robertson–Sparck Jones, with the +1 IDF form so it's never
// negative — see /tmp/algo-research/01_data_structures.md section 4):
//
//   score(D, Q) = Σ_{qi ∈ Q} IDF(qi) · (f(qi,D)·(k1+1)) /
//                                 (f(qi,D) + k1·(1 - b + b·|D|/avgdl))
//
//   IDF(qi)    = ln((N - df + 0.5) / (df + 0.5) + 1)
//
// We use k1=1.2, b=0.75 (Lucene/Elasticsearch defaults — empirically
// tuned on TREC short-text corpora).
func bm25ScoreForDocWithText(postings map[string][]uint64, qTokens []string, docTokens []string, avgDocLen float64, totalDocs int) float64 {
	if len(docTokens) == 0 || avgDocLen == 0 || totalDocs == 0 {
		return 0
	}
	dl := len(docTokens)
	qSet := make(map[string]struct{}, len(qTokens))
	for _, qt := range qTokens {
		qSet[qt] = struct{}{}
	}
	tf := make(map[string]int, len(qTokens))
	for _, t := range docTokens {
		if _, ok := qSet[t]; ok {
			tf[t]++
		}
	}
	if len(tf) == 0 {
		return 0
	}
	const k1 = 1.2
	const b = 0.75
	score := 0.0
	for qt := range qSet {
		df := len(postings[qt])
		if df == 0 {
			continue
		}
		f := tf[qt]
		if f == 0 {
			continue
		}
		idf := math.Log((float64(totalDocs)-float64(df)+0.5)/(float64(df)+0.5) + 1.0)
		score += idf * (float64(f)*(k1+1)) /
			(float64(f) + k1*(1-b+b*float64(dl)/avgDocLen))
	}
	return score
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
