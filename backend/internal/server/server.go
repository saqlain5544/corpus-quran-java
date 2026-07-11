// Package server wires HTTP handlers to the loaded Quran data and
// renders html/template pages. It is the runtime side of the
// pipeline ↔ data boundary.
package server

import (
	"bytes"
	"cmp"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"quranreader/backend/internal/data"
	"quranreader/backend/internal/search"
	"quranreader/loc"
	"quranreader/types"
)

// Server is the assembled HTTP handler.
type Server struct {
	quran        *types.Quran
	masaq        *types.MasaqIndex
	roots        *types.RootsIndex
	meta         *types.Meta
	translations *data.Translations
	concordance  *data.Concordance
	fontMeta     FontMetaPageData

	templates map[string]*template.Template
	staticFS  fs.FS
	fontsFS   fs.FS

	startTime time.Time
	log       *slog.Logger
}

// TemplatesFS is the minimal interface loadTemplates needs:
// ReadDir to enumerate *.tmpl files, and ReadFile to load them.
// embed.FS and fs.ReadFileFS both satisfy this.
type TemplatesFS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

// ReadFS combines fs.ReadFileFS (for serving) with ReadDir (for
// any optional enumeration). embed.FS satisfies both directly.
type ReadFS = fs.ReadFileFS

// Options controls optional behavior.
type Options struct {
	// Translations holds verse-level translations (loaded from
	// data/quran/translations/). nil means no translations available.
	Translations *data.Translations
	// Concordance holds the concordance index (loaded from
	// data/morphology/concordance.jsonl). nil means no concordance.
	Concordance *data.Concordance
	// TemplatesFS is the embedded templates directory.
	TemplatesFS TemplatesFS
	// StaticFS is the embedded static/ directory (css, js).
	StaticFS ReadFS
	// FontsFS is the embedded fonts/ directory (hafs.woff2).
	FontsFS ReadFS
	// Logger is the structured logger to use; nil falls back to slog.Default().
	Logger *slog.Logger
}

// New builds a Server with the given data and option overrides.
func New(q *types.Quran, m *types.MasaqIndex, r *types.RootsIndex, meta *types.Meta, opts Options) (*Server, error) {
	s := &Server{
		quran:        q,
		masaq:        m,
		roots:        r,
		meta:         meta,
		translations: opts.Translations,
		concordance:  opts.Concordance,
		startTime:    time.Now(),
		log:          opts.Logger,
	}
	if s.log == nil {
		s.log = slog.Default()
	}

	// Templates — pass the FS root; loadTemplates prepends "templates/".
	t, err := loadTemplates(opts.TemplatesFS)
	if err != nil {
		return nil, fmt.Errorf("load templates: %w", err)
	}
	s.templates = t

	// Static + fonts: each Options.FS is expected to be rooted at
	// the directory that contains "static" / "fonts". In production
	// that's the backend dir; in tests it's the testdata dir.
	if opts.StaticFS != nil {
		fsys := opts.StaticFS
		sub, err := fs.Sub(fsys, "static")
		if err != nil {
			return nil, fmt.Errorf("static sub: %w", err)
		}
		s.staticFS = sub
	}
	if opts.FontsFS != nil {
		fsys := opts.FontsFS
		sub, err := fs.Sub(fsys, "fonts")
		if err != nil {
			return nil, fmt.Errorf("fonts sub: %w", err)
		}
		s.fontsFS = sub
	}

	// Font-meta debug page. Loads the embedded cmap JSON (built by
	// scripts/extract-font-cmaps.py) once at startup and assembles
	// the page data. If the cmap JSON can't be parsed the page is
	// still served but the "what's in this font" sections are empty.
	hafsCmap, amiriCmap, err := loadFontCMaps()
	if err != nil {
		s.log.Warn("fontcmaps unavailable; /meta/fonts will be partial", "err", err)
	}
	s.fontMeta = buildFontMeta(q, hafsCmap, amiriCmap)

	return s, nil
}

// Handler returns the http.Handler ready to serve.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Pages. The pattern "GET /{$}" matches ONLY the literal root
	// "/" — NOT any path with "/" as a prefix. (In Go 1.22's
	// ServeMux, the bare pattern "GET /" is a catch-all that matches
	// every path not handled by another pattern; the "/{$}" form
	// scopes it to the exact root.)
	mux.HandleFunc("GET /{$}", s.handleHomepage)
	mux.HandleFunc("GET /surah/{id}", s.handleSurah)
	mux.HandleFunc("GET /search", s.handleSearch)
	mux.HandleFunc("GET /roots", s.handleRootsList)
	mux.HandleFunc("GET /root/detailed/{root}", s.handleRootDetail)
	mux.HandleFunc("GET /about", s.handleAbout)
	mux.HandleFunc("GET /concordance", s.handleConcordance)
	mux.HandleFunc("GET /meta/fonts", s.handleFontMeta)

	// APIs.
	mux.HandleFunc("GET /api/word", s.handleAPIWord)
	mux.HandleFunc("GET /api/search", s.handleAPISearch)
	mux.HandleFunc("GET /api/roots", s.handleAPIRoots)
	mux.HandleFunc("GET /api/root/{root}/occurrences", s.handleAPIRootOccurrences)
	mux.HandleFunc("GET /api/root/{root}/summary", s.handleAPIRootSummary)
	mux.HandleFunc("GET /api/health", s.handleHealth)

	// Static + fonts.
	mux.Handle("GET /static/", immutableCache(http.StripPrefix("/static/", http.FileServerFS(s.staticFS))))
	mux.Handle("GET /fonts/", immutableCache(http.StripPrefix("/fonts/", http.FileServerFS(s.fontsFS))))

	// Catch-all 404 — registered LAST so the more-specific patterns
	// above win under Go 1.22's longest-match-wins mux. Routes that
	// don't match anything fall through here and render the styled
	// error page (HTML) or JSON 404 (for /api/ paths).
	mux.HandleFunc("GET /{rest...}", s.handleCatchAll)

	return s.recovery(s.logging(mux))
}

// handleCatchAll is the catch-all 404 handler. Browser requests
// get the styled HTML error page; API requests get a JSON 404.
func (s *Server) handleCatchAll(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.respondJSON(w, http.StatusNotFound, map[string]string{
			"error": "not found",
			"path":  r.URL.Path,
		})
		return
	}
	s.handleNotFound(w, r)
}

// ─── Middleware ────────────────────────────────────────────────

// immutableCache wraps a handler so every response carries headers
// that let the browser cache the asset forever. Static assets
// (CSS, JS, fonts) are embedded in the binary at compile time via
// go:embed — they NEVER change at runtime. Setting an immutable
// 1-year max-age avoids ~50 KB of font re-fetch and ~50 KB of
// CSS/JS revalidation per page load.
//
// If we ever want to ship asset updates without a binary release,
// add a content hash to the URL (/static/main.{hash}.css) and bump
// the URL — the old URL stays cached.
func immutableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		s.log.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"ua", r.Header.Get("User-Agent"),
		)
	})
}

func (s *Server) recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				s.log.Error("panic", "err", rec, "path", r.URL.Path)
				s.respondError(w, r, http.StatusInternalServerError, "internal error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// ─── Page handlers ─────────────────────────────────────────────

// handleNotFound is the 404 fallback for any GET path that didn't
// match a more-specific route above. Go 1.22's ServeMux uses
// longest-match-wins, so the more-specific patterns (e.g.
// "/surah/{id}") win over the "{rest...}" catch-all here.
//
// We render the error page with a helpful message and links back
// to the homepage and search, so the user has a clear path forward
// instead of seeing a bare 404 or (worse) the homepage content
// masquerading as a different page.
func (s *Server) handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	data := struct {
		Title  string
		Status int
		Msg    string
		Path   string
	}{
		Title:  "Page not found",
		Status: http.StatusNotFound,
		Msg: fmt.Sprintf("No page matches the path %q. It may have been moved "+
			"or never existed.", r.URL.Path),
		Path: r.URL.Path,
	}
	t, ok := s.templates["error.tmpl"]
	if !ok {
		fmt.Fprintf(w, "404 Not Found: %s", data.Msg)
		return
	}
	if err := t.ExecuteTemplate(w, "error.tmpl", data); err != nil {
		s.log.Error("404 template", "err", err, "path", r.URL.Path)
		fmt.Fprintf(w, "404 Not Found: %s", data.Msg)
	}
}

func (s *Server) handleHomepage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title  string
		Surahs []*types.Surah
		Meta   *types.Meta
		Stats  *homepageStats
	}{Title: "Quran Reader", Surahs: nil, Meta: s.meta}
	// Build an ordered slice of surahs.
	data.Surahs = make([]*types.Surah, 114)
	for i := 1; i <= 114; i++ {
		data.Surahs[i-1] = s.quran.Surahs[i]
	}

	// Aggregate stats for the header strip. Counting roots and
	// translation sets directly off the loaded indexes keeps the
	// numbers honest (any future data changes propagate). Both
	// fields are optional (tests may construct a Server with
	// neither), so each is nil-guarded.
	stats := &homepageStats{
		SurahCount: 114,
		AyahCount:  s.meta.AyahCount,
		WordCount:  s.meta.WordCount,
	}
	if s.roots != nil {
		stats.RootCount = len(s.roots.ByRoot)
	}
	if s.translations != nil {
		stats.TranslationCount = len(s.translations.Sets)
	}
	data.Stats = stats
	s.render(w, "homepage.tmpl", data)
}

// homepageStats holds the aggregate counts shown above the surah
// grid. Each field is sourced from the live data so the displayed
// numbers are always accurate.
type homepageStats struct {
	SurahCount       int
	AyahCount        int
	WordCount        int
	RootCount        int
	TranslationCount int
}

// surahPageAyah wraps an Ayah with per-word root lookups for the
// surah page template. The WordRoots map is keyed by word number
// (1-based) and holds the buckwalter code (empty if no root).
type surahPageAyah struct {
	*types.Ayah
	WordRoots map[int]string // wordNo → buckwalter
}

func (s *Server) handleSurah(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 || id > 114 {
		s.respondError(w, r, http.StatusNotFound, "Surah not found")
		return
	}
	surah := s.quran.Surahs[id]
	if surah == nil {
		s.respondError(w, r, http.StatusNotFound, "Surah not found")
		return
	}

	// Build per-ayah page data with root lookups pre-computed.
	// Each ayah is independent (read-only ByLoc + own per-ayah map),
	// so we process ayahs in parallel chunks. Chunk size balances
	// goroutine count against work-per-task: too small and we
	// pay goroutine setup overhead; too large and the longest
	// ayah (long tail of Al-Baqarah) dominates one worker.
	type ayahEntry struct {
		Ayah      *types.Ayah
		WordRoots map[int]string
	}
	const chunkSize = 16
	totalAyahs := len(surah.Ayahs)
	ayahEntries := make([]ayahEntry, totalAyahs)

	buildEntry := func(an int) ayahEntry {
		ay := surah.Ayahs[an]
		if ay == nil {
			return ayahEntry{}
		}
		wr := map[int]string{}
		for _, t := range ay.Tokens {
			if t.Kind == "word" {
				key := loc.Key(id, an, t.WordNo)
				if s.roots != nil {
					if bw, ok := s.roots.ByLoc[key]; ok {
						wr[t.WordNo] = bw
					}
				}
			}
		}
		return ayahEntry{Ayah: ay, WordRoots: wr}
	}

	var (
		wg  sync.WaitGroup
		sem = make(chan struct{}, runtime.GOMAXPROCS(0))
	)
	for start := 1; start <= totalAyahs; start += chunkSize {
		end := start + chunkSize - 1
		if end > totalAyahs {
			end = totalAyahs
		}
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() { <-sem }()
			defer wg.Done()
			for an := start; an <= end; an++ {
				ayahEntries[an-1] = buildEntry(an)
			}
		}()
	}
	wg.Wait()

	type transOption struct {
		Idx   int
		Label string
		Dir   string
	}
	var transOpts []transOption
	allAyahTrans := make([][]string, 0) // [transIdx][ayahIdx]
	transIdx := 0
	if s.translations != nil && len(s.translations.Sets) > 0 {
		for i, ts := range s.translations.Sets {
			transOpts = append(transOpts, transOption{Idx: i, Label: ts.Label, Dir: ts.Dir})
		}
		transIdx = s.translations.Default
		// Build per-ayah translation array for ALL sets.
		for _, set := range s.translations.Sets {
			oneTrans := make([]string, len(ayahEntries))
			if set.Data != nil {
				if ay, ok := set.Data[id]; ok {
					for i, ae := range ayahEntries {
						oneTrans[i] = ay[ae.Ayah.Number]
					}
				}
			}
			allAyahTrans = append(allAyahTrans, oneTrans)
		}
	}

	data := struct {
		Title        string
		Surah        *types.Surah
		AyahEntries  []ayahEntry
		AllAyahTrans [][]string // [transIdx][ayahIdx]
		TransOptions []transOption
		TransIdx     int
		Prev         *types.Surah
		Next         *types.Surah
	}{
		Title:        surah.Name,
		Surah:        surah,
		AyahEntries:  ayahEntries,
		AllAyahTrans: allAyahTrans,
		TransOptions: transOpts,
		TransIdx:     transIdx,
	}
	if id > 1 {
		data.Prev = s.quran.Surahs[id-1]
	}
	if id < 114 {
		data.Next = s.quran.Surahs[id+1]
	}
	s.render(w, "surah.tmpl", data)
}

func (s *Server) handleRootsList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 50
	asc := r.URL.Query().Get("sort") == "asc"
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	results, total := search.RootsList(s.roots, page, pageSize, asc, filterSurah)

	type surahOption struct {
		Number int
		Name   string
	}
	surahOpts := make([]surahOption, 114)
	for i := 0; i < 114; i++ {
		surahOpts[i] = surahOption{Number: i + 1, Name: s.meta.SurahNames[i]}
	}

	data := struct {
		Title       string
		Page        int
		PageSize    int
		Total       int
		HasPrev     bool
		HasNext     bool
		Asc         bool
		FilterSurah int
		SurahOpts   []surahOption
		Results     []search.SearchResult
	}{
		Title:       "Roots",
		Page:        page,
		PageSize:    pageSize,
		Total:       total,
		HasPrev:     page > 1,
		HasNext:     page*pageSize < total,
		Asc:         asc,
		FilterSurah: filterSurah,
		SurahOpts:   surahOpts,
		Results:     results,
	}
	s.render(w, "roots-list.tmpl", data)
}

func (s *Server) handleRootDetail(w http.ResponseWriter, r *http.Request) {
	root := r.PathValue("root")
	entry, ok := s.roots.ByRoot[root]
	if !ok {
		// Try the Arabic reverse lookup using the same canonical key
		// (no spaces, hamza/ya/ta-marbuta unified) the data layer
		// uses when building the index.
		if hit, ok := s.roots.ByArabic[types.NormalizeArabicRoot(root)]; ok {
			root = hit
			entry = s.roots.ByRoot[root]
			if entry == nil {
				s.respondError(w, r, http.StatusNotFound, "Root not found")
				return
			}
		} else {
			// Last-chance: transliterate Arabic input to Buckwalter
			// (mirrors search.Root so /root/detailed/قول resolves
			// the same way /api/search?q=قول does).
			if bw := search.Buckwalter(search.NormalizeArabic(root)); bw != "" {
				if e, ok := s.roots.ByRoot[bw]; ok {
					root = bw
					entry = e
					http.Redirect(w, r, "/root/detailed/"+bw, http.StatusMovedPermanently)
					return
				}
			}
			s.respondError(w, r, http.StatusNotFound, "Root not found")
			return
		}
	}
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	asc := r.URL.Query().Get("sort") == "asc"
	occurrences := search.OccurrencesForRoot(root, s.roots, s.quran, s.masaq, filterSurah, asc)

	// Build concordance lemmas with verse text + translations.
	type lemmaOccurrence struct {
		Ref         string // "s:v"
		Link        string // "/surah/S#verse-V"
		VerseText   string // Uthmani verse text
		Translation string // verse-level English translation
	}
	type concordLemma struct {
		Arabic      string
		Occurrences int
		Verses      []lemmaOccurrence
	}
	var conLemmas []concordLemma
	if s.concordance != nil {
		if ce, ok := s.concordance.ByRoot[root]; ok {
			// Pre-resolve the active translation set once. We use
			// it inside the parallel worker below; reading it per
			// lemma would require either locking or making a copy.
			var trans *data.TranslationSet
			if s.translations != nil && s.translations.Default < len(s.translations.Sets) {
				trans = &s.translations.Sets[s.translations.Default]
			}

			// Snapshot the concordance lemmas: each lemma maps an
			// Arabic letter to (total occurrences, []verse refs).
			// Build a parallel-safe input slice — anonymous structs
			// can't be named, so we use a local alias type.
			type lemmaInner struct {
				TotalOccurrences int
				Occurrences      []string
			}
			type lemmaInput struct {
				arabic string
				inner  lemmaInner
			}
			ins := make([]lemmaInput, 0, len(ce.Lemmas))
			for ar, lm := range ce.Lemmas {
				ins = append(ins, lemmaInput{arabic: ar, inner: lemmaInner{
					TotalOccurrences: lm.TotalOccurrences,
					Occurrences:      lm.Occurrences,
				}})
			}

			// Build each lemma in parallel. Lemmas are independent —
			// each reads from the read-only ByRoot map and writes
			// into a uniquely-indexed slot of `outs`, so there's no
			// shared mutable state across goroutines. The chunk size
			// keeps the goroutine count bounded on roots with many
			// lemmas (most roots have 5-50).
			outs := make([]concordLemma, len(ins))
			const lemmaChunk = 4
			var wg sync.WaitGroup
			for start := 0; start < len(ins); start += lemmaChunk {
				end := start + lemmaChunk
				if end > len(ins) {
					end = len(ins)
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := start; i < end; i++ {
						in := ins[i]
						occList := make([]lemmaOccurrence, 0, len(in.inner.Occurrences))
						for _, sv := range in.inner.Occurrences {
							parts := strings.SplitN(sv, ":", 2)
							if len(parts) != 2 {
								continue
							}
							sNum, sErr := strconv.Atoi(parts[0])
							vNum, vErr := strconv.Atoi(parts[1])
							if sErr != nil || vErr != nil || sNum < 1 || sNum > 114 || vNum < 1 {
								continue
							}
							lo := lemmaOccurrence{
								Ref:       sv,
								Link:      fmt.Sprintf("/surah/%d#verse-%d", sNum, vNum),
								VerseText: verseTextSafe(s.quran, sNum, vNum),
							}
							if trans != nil && trans.Data != nil {
								if ay, ok := trans.Data[sNum]; ok {
									lo.Translation = ay[vNum]
								}
							}
							occList = append(occList, lo)
						}
						outs[i] = concordLemma{
							Arabic:      in.arabic,
							Occurrences: in.inner.TotalOccurrences,
							Verses:      occList,
						}
					}
				}()
			}
			wg.Wait()

			conLemmas = outs
			// Sort lemmas by occurrence count descending
			sort.Slice(conLemmas, func(i, j int) bool {
				return conLemmas[i].Occurrences > conLemmas[j].Occurrences
			})
		}
	}

	type surahOption struct {
		Number int
		Name   string
	}
	surahOpts := make([]surahOption, 114)
	for i := 0; i < 114; i++ {
		surahOpts[i] = surahOption{Number: i + 1, Name: s.meta.SurahNames[i]}
	}

	// Distribution of this root's occurrences across surahs (count
	// per surah). Computed from entry.Locations so it stays in sync
	// with the underlying data and includes every surah the root
	// appears in.
	type surahDist struct {
		Number int
		Name   string
		Count  int
	}
	var surahDistribution []surahDist
	if len(entry.Locations) > 0 {
		counts := make(map[int]int)
		for _, loc := range entry.Locations {
			parts := strings.SplitN(loc, ":", 2)
			if len(parts) < 1 {
				continue
			}
			n, err := strconv.Atoi(parts[0])
			if err != nil {
				continue
			}
			counts[n]++
		}
		surahDistribution = make([]surahDist, 0, len(counts))
		for n, c := range counts {
			name := ""
			if n >= 1 && n <= 114 && s.meta != nil && n-1 < len(s.meta.SurahNames) {
				name = s.meta.SurahNames[n-1]
			}
			surahDistribution = append(surahDistribution, surahDist{Number: n, Name: name, Count: c})
		}
		sort.Slice(surahDistribution, func(i, j int) bool {
			if surahDistribution[i].Count != surahDistribution[j].Count {
				return surahDistribution[i].Count > surahDistribution[j].Count
			}
			return surahDistribution[i].Number < surahDistribution[j].Number
		})
	}

	// MorphSummary aggregates morphological pattern counts across all
	// occurrences so the template can render a TAG → count → % table.
	type morphTagCount struct {
		Tag   string
		Type  string
		Count int
		Pct   float64
	}
	var ms []morphTagCount
	if len(occurrences) > 0 {
		raw := make(map[string]int) // key = "TAG|TYPE"
		for _, occ := range occurrences {
			for _, seg := range occ.MorphSegments {
				if seg.MorphTag == "" && seg.MorphType == "" {
					continue
				}
				key := seg.MorphTag + "|" + seg.MorphType
				raw[key]++
			}
		}
		total := 0
		for _, c := range raw {
			total += c
		}
		for k, c := range raw {
			parts := strings.SplitN(k, "|", 2)
			tag, typ := k, ""
			if len(parts) == 2 {
				tag, typ = parts[0], parts[1]
			}
			var pct float64
			if total > 0 {
				pct = float64(c) * 100 / float64(total)
			}
			ms = append(ms, morphTagCount{Tag: tag, Type: typ, Count: c, Pct: pct})
		}
		sort.Slice(ms, func(i, j int) bool {
			return ms[i].Count > ms[j].Count
		})
	}

	data := struct {
		Title             string
		Root              string
		Entry             *types.RootEntry
		Occurrences       []search.SearchResult
		ConLemmas         []concordLemma
		SurahDistribution []surahDist
		MorphSummary      []morphTagCount
		FilterSurah       int
		Asc               bool
		SurahOpts         []surahOption
	}{Title: entry.Buckwalter, Root: root, Entry: entry, Occurrences: occurrences, ConLemmas: conLemmas, SurahDistribution: surahDistribution, MorphSummary: ms, FilterSurah: filterSurah, Asc: asc, SurahOpts: surahOpts}
	s.render(w, "root-detail.tmpl", data)
}

// verseTextSafe returns Uthmani verse text for (surah, ayah), or "" if not found.
func verseTextSafe(quran *types.Quran, surah, ayah int) string {
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

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	s.render(w, "about.tmpl", struct {
		Title string
	}{Title: "About"})
}

// handleFontMeta serves /meta/fonts — the font-debug page that
// visualises every codepoint in the Quran text and every glyph in
// each font's cmap. See font_meta.go for the data shape and
// font_meta_load.go for how the cmaps are sourced.
func (s *Server) handleFontMeta(w http.ResponseWriter, r *http.Request) {
	s.render(w, "font_meta.tmpl", s.fontMeta)
}

func (s *Server) handleConcordance(w http.ResponseWriter, r *http.Request) {
	if s.concordance == nil {
		s.respondError(w, r, http.StatusNotFound, "Concordance data not loaded")
		return
	}
	// Pagination.
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 100
	total := len(s.concordance.Keys)
	start := (page - 1) * pageSize
	if start >= total {
		// Out-of-range page — return 404 rather than silently
		// bouncing the user back to page 1 (which would also leave
		// the rendered "Page N" indicator inconsistent with the
		// actual data shown).
		s.respondError(w, r, http.StatusNotFound, fmt.Sprintf("No page %d (have %d pages)", page, (total+pageSize-1)/pageSize))
		return
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	type lemmaView struct {
		Arabic      string
		Occurrences int
		Verses      []string
	}
	type rootView struct {
		Buckwalter  string
		Occurrences int
		Lemmas      []lemmaView
	}
	views := make([]rootView, 0, end-start)
	for _, bw := range s.concordance.Keys[start:end] {
		e := s.concordance.ByRoot[bw]
		lemmas := make([]lemmaView, 0, len(e.Lemmas))
		for ar, lm := range e.Lemmas {
			lemmas = append(lemmas, lemmaView{Arabic: ar, Occurrences: lm.TotalOccurrences, Verses: lm.Occurrences})
		}
		sort.Slice(lemmas, func(i, j int) bool {
			return lemmas[i].Occurrences > lemmas[j].Occurrences
		})
		views = append(views, rootView{Buckwalter: bw, Occurrences: e.TotalOccurrences, Lemmas: lemmas})
	}

	data := struct {
		Title    string
		Page     int
		PageSize int
		Total    int
		HasPrev  bool
		HasNext  bool
		Roots    []rootView
	}{
		Title:    "Concordance",
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		HasPrev:  page > 1,
		HasNext:  end < total,
		Roots:    views,
	}
	s.render(w, "concordance.tmpl", data)
}

// ─── API handlers ──────────────────────────────────────────────

func (s *Server) handleAPIRoots(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	asc := r.URL.Query().Get("sort") == "asc"
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	results, total := search.RootsList(s.roots, page, 50, asc, filterSurah)
	sortLabel := "desc"
	if asc {
		sortLabel = "asc"
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"page":      page,
		"page_size": 50,
		"total":     total,
		"sort":      sortLabel,
		"roots":     results,
	})
}

func (s *Server) handleAPIRootOccurrences(w http.ResponseWriter, r *http.Request) {
	root := r.PathValue("root")
	entry, ok := s.roots.ByRoot[root]
	if !ok {
		s.respondError(w, r, http.StatusNotFound, "root not found")
		return
	}
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	asc := r.URL.Query().Get("sort") == "asc"
	occ := search.OccurrencesForRoot(root, s.roots, s.quran, s.masaq, filterSurah, asc)
	s.respondJSON(w, http.StatusOK, map[string]any{
		"root":        root,
		"arabic":      entry.Arabic,
		"occurrences": occ,
	})
}

// handleAPIRootSummary returns everything needed for the surah page
// side panel: root metadata + AI meaning data + lemma frequency.
//
// The panel does NOT include the concordance occurrence list — the
// plan explicitly says the side panel focuses on morphology and root
// meanings, not the per-verse concordance (that lives on /concordance
// and /root/detailed/[root]).
func (s *Server) handleAPIRootSummary(w http.ResponseWriter, r *http.Request) {
	root := r.PathValue("root")
	entry, ok := s.roots.ByRoot[root]
	if !ok {
		s.respondError(w, r, http.StatusNotFound, "root not found")
		return
	}

	// Lemma frequency list — comes from the concordance index but we
	// only surface the lemma + occurrence count (not the verse refs,
	// which is what the panel should be hiding).
	type lemmaView struct {
		Arabic      string `json:"arabic"`
		Occurrences int    `json:"occurrences"`
	}
	lemmas := []lemmaView{}
	if s.concordance != nil {
		if ce, ok := s.concordance.ByRoot[root]; ok {
			for ar, lm := range ce.Lemmas {
				lemmas = append(lemmas, lemmaView{
					Arabic:      ar,
					Occurrences: lm.TotalOccurrences,
				})
			}
			slices.SortFunc(lemmas, func(a, b lemmaView) int {
				return cmp.Compare(b.Occurrences, a.Occurrences)
			})
		}
	}

	s.respondJSON(w, http.StatusOK, map[string]any{
		"root":           root,
		"arabic":         entry.Arabic,
		"letters":        entry.Letters,
		"pos":            entry.POS,
		"occurrences":    entry.Occurrences,
		"meaning_en":     entry.MeaningEN,
		"meaning_ar":     entry.MeaningAR,
		"core_semantic":  entry.CoreSemantic,
		"ibn_faris":      entry.IbnFaris,
		"al_raghib":      entry.AlRaghib,
		"quran_examples": entry.QuranExamples,
		"hadith":         entry.HadithExamples,
		"lemmas":         lemmas,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]any{
		"status":         "ok",
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
	})
}

// ─── Render helpers ────────────────────────────────────────────

func (s *Server) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, ok := s.templates[name]
	if !ok {
		s.log.Error("template not found", "name", name)
		http.Error(w, "template missing", http.StatusInternalServerError)
		return
	}
	// Execute into a buffer first so we can detect template errors
	// before any bytes reach the wire. Otherwise a runtime error in
	// a deeply-nested template leaves the client with a 200 status
	// and a half-rendered page.
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		s.log.Error("template execute", "name", name, "err", err)
		// Template errors are always 500 with a plain-text fallback
		// (the HTML error page is itself a template, so recursive
		// failures are possible). We skip the respondError path
		// because render() has no access to the incoming *Request.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "template %q failed: %s", name, err.Error())
		return
	}
	w.Write(buf.Bytes())
}

func (s *Server) respondError(w http.ResponseWriter, r *http.Request, status int, msg string) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		s.respondJSON(w, status, map[string]any{
			"error": msg,
			"path":  r.URL.Path,
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	data := struct {
		Title  string
		Status int
		Msg    string
	}{Title: "Error", Status: status, Msg: msg}
	t, ok := s.templates["error.tmpl"]
	if !ok {
		fmt.Fprintf(w, "Error %d: %s", status, msg)
		return
	}
	if err := t.ExecuteTemplate(w, "error.tmpl", data); err != nil {
		s.log.Error("error template", "err", err)
		fmt.Fprintf(w, "Error %d: %s", status, msg)
	}
}

func (s *Server) respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := jsonEncode(w, v); err != nil {
		s.log.Error("json encode", "err", err)
	}
}
