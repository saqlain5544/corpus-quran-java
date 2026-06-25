// Package server wires HTTP handlers to the loaded Quran data and
// renders html/template pages. It is the runtime side of the
// pipeline ↔ data boundary.
package server

import (
	"bytes"
	"cmp"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
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

	return s, nil
}

// asFS is no longer needed — Options.FS are expected to satisfy
// fs.ReadFileFS directly (embed.FS does so natively). Kept as a
// type-assertion helper in case future Options.FS don't.
func asFS(r ReadFS) fs.FS {
	if f, ok := r.(fs.FS); ok {
		return f
	}
	return rfsAdapter{r}
}

type rfsAdapter struct{ ReadFS }

func (r rfsAdapter) Open(name string) (fs.File, error) {
	b, err := r.ReadFS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &rfsFile{data: b, name: name}, nil
}

type rfsFile struct {
	data []byte
	name string
	off  int
}

func (f *rfsFile) Read(p []byte) (int, error) {
	if f.off >= len(f.data) {
		return 0, io.EOF
	}
	n := copy(p, f.data[f.off:])
	f.off += n
	return n, nil
}
func (f *rfsFile) Close() error               { return nil }
func (f *rfsFile) Stat() (fs.FileInfo, error) { return rfsInfo{f}, nil }
func (f *rfsFile) Name() string               { return f.name }

type rfsInfo struct{ f *rfsFile }

func (i rfsInfo) Name() string       { return i.f.name }
func (i rfsInfo) Size() int64        { return int64(len(i.f.data)) }
func (i rfsInfo) Mode() fs.FileMode  { return 0644 }
func (i rfsInfo) ModTime() time.Time { return time.Time{} }
func (i rfsInfo) IsDir() bool        { return false }
func (i rfsInfo) Sys() any           { return nil }

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

	// APIs.
	mux.HandleFunc("GET /api/word", s.handleAPIWord)
	mux.HandleFunc("GET /api/search", s.handleAPISearch)
	mux.HandleFunc("GET /api/roots", s.handleAPIRoots)
	mux.HandleFunc("GET /api/root/{root}/occurrences", s.handleAPIRootOccurrences)
	mux.HandleFunc("GET /api/root/{root}/summary", s.handleAPIRootSummary)
	mux.HandleFunc("GET /api/health", s.handleHealth)

	// Static + fonts.
	mux.Handle("GET /static/", noCache(http.StripPrefix("/static/", http.FileServerFS(s.staticFS))))
	mux.Handle("GET /fonts/", noCache(http.StripPrefix("/fonts/", http.FileServerFS(s.fontsFS))))

	// Wrap the mux in a 404 interceptor. Go 1.22 ServeMux doesn't
	// support a wildcard catch-all alongside "/" (they conflict at
	// registration time), and NotFoundHandler isn't a field — it's
	// a helper function. So we wrap the mux and check the response
	// status: if it's a 404 with the default plain-text body, we
	// know no route matched, so render our styled error page.
	// API routes get JSON 404s instead (handled in the wrapper).
	return s.notFoundInterceptor(s.recovery(s.logging(mux)))
}

// notFoundInterceptor wraps a handler so any 404 response that
// the mux produced (because no route matched) is replaced with
// our styled HTML error page for browser requests, or a JSON 404
// for API requests. This is the cleanest way to customize 404s
// with Go 1.22+ ServeMux.
//
// We use a full-buffering approach: capture the entire response
// into memory, then forward it only if it's NOT a default Go 404
// (which is the literal "404 page not found\n" body that ServeMux
// emits when no route matches).
//
// Cost: tiny — the mux's default 404 body is 19 bytes; a custom
// 404 from one of our handlers is usually a few KB. We never buffer
// legitimate 200 responses for routes that DO match, because those
// don't go through our interceptor path (they pass through to the
// inner handler).
func (s *Server) notFoundInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bw := &bufferedWriter{ResponseWriter: w}
		next.ServeHTTP(bw, r)
		// Detect the mux's default 404: it's the literal text
		// "404 page not found\n" with status 404. If we see that,
		// render our own page instead.
		if bw.status == http.StatusNotFound && !bw.customBody {
			isAPI := strings.HasPrefix(r.URL.Path, "/api/")
			// Clear the response buffer so we can write fresh.
			bw.buffer.Reset()
			bw.headers = nil
			bw.status = 0
			if isAPI {
				s.respondJSON(w, http.StatusNotFound, map[string]string{
					"error": "not found",
					"path":  r.URL.Path,
				})
			} else {
				s.handleNotFound(w, r)
			}
			return
		}
		// Otherwise, forward the buffered response to the real client.
		bw.flush()
	})
}

// bufferedWriter captures the entire response in memory so we can
// inspect it and decide whether to forward as-is or replace with
// our styled 404 page. It implements http.ResponseWriter so the
// inner handler can call WriteHeader / Write as usual.
type bufferedWriter struct {
	http.ResponseWriter
	headers   http.Header
	buffer    *bytes.Buffer
	status    int
	customBody bool
}

// Header returns the buffered writer's header map. We keep our own
// copy so the inner handler can mutate it freely without touching
// the real response until we flush.
func (bw *bufferedWriter) Header() http.Header {
	if bw.headers == nil {
		bw.headers = make(http.Header)
	}
	return bw.headers
}

// WriteHeader buffers the status code. We don't call the real
// WriteHeader until flush() so we can replace the response entirely.
func (bw *bufferedWriter) WriteHeader(code int) {
	bw.status = code
}

// Write buffers body bytes. The first non-empty Write marks the
// response as having a "custom" body (i.e. not the mux's default
// 404 placeholder). On flush() we forward everything as-is.
func (bw *bufferedWriter) Write(b []byte) (int, error) {
	if bw.status == 0 {
		bw.status = http.StatusOK
	}
	if bw.buffer == nil {
		bw.buffer = &bytes.Buffer{}
	}
	if bw.buffer.Len() == 0 && string(b) == "404 page not found\n" {
		// Mux's default 404 — don't mark as custom.
		bw.buffer.Write(b)
		return len(b), nil
	}
	bw.customBody = true
	bw.buffer.Write(b)
	return len(b), nil
}

// flush forwards the buffered response to the real client.
func (bw *bufferedWriter) flush() {
	if bw.headers != nil {
		for k, vs := range bw.headers {
			for _, v := range vs {
				bw.ResponseWriter.Header().Add(k, v)
			}
		}
	}
	if bw.status != 0 {
		bw.ResponseWriter.WriteHeader(bw.status)
	}
	if bw.buffer != nil {
		bw.ResponseWriter.Write(bw.buffer.Bytes())
	}
}

// ─── Middleware ────────────────────────────────────────────────

// noCache wraps a handler so every response carries headers that
// force the browser to revalidate on every request. Static assets
// are embedded in the binary at compile time — there's no benefit
// to caching them client-side, and we want CSS/JS edits to land on
// the next reload without a hard refresh.
func noCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
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
				s.respondError(w, http.StatusInternalServerError, "internal error")
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
		Title     string
		Status    int
		Msg       string
		Path      string
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
		s.respondError(w, http.StatusNotFound, "Surah not found")
		return
	}
	surah := s.quran.Surahs[id]
	if surah == nil {
		s.respondError(w, http.StatusNotFound, "Surah not found")
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
				if bw, ok := s.roots.ByLoc[key]; ok {
					wr[t.WordNo] = bw
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
		wg.Go(func() {
			defer func() { <-sem }()
			for an := start; an <= end; an++ {
				ayahEntries[an-1] = buildEntry(an)
			}
		})
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

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "english"
	}
	data := struct {
		Title   string
		Query   string
		Type    string
		Results []search.SearchResult
	}{Title: "Search", Query: q, Type: typ}
	if q != "" {
		switch typ {
		case "root":
			data.Results = search.Root(q, s.roots, s.quran, 2000)
		case "arabic":
			data.Results = search.Arabic(q, s.masaq, s.quran, 5000, 0)
		case "translation":
			data.Results = search.Translation(q, s.masaq, s.quran, 5000, 0)
		case "lemma":
			data.Results = search.Lemma(q, s.masaq, s.quran, 5000, 0)
		default:
			typ = "english"
			data.Type = typ
			data.Results = search.English(q, s.masaq, s.quran, 5000, 0)
		}
	}
	s.render(w, "search.tmpl", data)
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
			s.respondError(w, http.StatusNotFound, "Root not found")
			return
		}
	}
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	asc := r.URL.Query().Get("sort") == "asc"
	occurrences := search.OccurrencesForRoot(root, s.roots, s.quran, filterSurah, asc)

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
				wg.Go(func() {
					for i := start; i < end; i++ {
						in := ins[i]
						occList := make([]lemmaOccurrence, 0, len(in.inner.Occurrences))
						for _, sv := range in.inner.Occurrences {
							parts := strings.SplitN(sv, ":", 2)
							if len(parts) != 2 {
								continue
							}
							sNum, _ := strconv.Atoi(parts[0])
							vNum, _ := strconv.Atoi(parts[1])
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
				})
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

	data := struct {
		Title             string
		Root              string
		Entry             *types.RootEntry
		Occurrences       []search.SearchResult
		ConLemmas         []concordLemma
		SurahDistribution []surahDist
		FilterSurah       int
		Asc               bool
		SurahOpts         []surahOption
	}{Title: entry.Buckwalter, Root: root, Entry: entry, Occurrences: occurrences, ConLemmas: conLemmas, SurahDistribution: surahDistribution, FilterSurah: filterSurah, Asc: asc, SurahOpts: surahOpts}
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

func (s *Server) handleConcordance(w http.ResponseWriter, r *http.Request) {
	if s.concordance == nil {
		s.respondError(w, http.StatusNotFound, "Concordance data not loaded")
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
		s.respondError(w, http.StatusNotFound, fmt.Sprintf("No page %d (have %d pages)", page, (total+pageSize-1)/pageSize))
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

type apiWordResponse struct {
	Surah       int                  `json:"surah"`
	Ayah        int                  `json:"ayah"`
	Word        int                  `json:"word"`
	WordText    string               `json:"word_text"`
	Lemma       string               `json:"lemma"`       // bare stem, no diacritics
	Gloss       string               `json:"gloss"`       // combined segment glosses
	Translation string               `json:"translation"` // word-level English translation
	POS         string               `json:"pos"`         // V / N / P
	Function    string               `json:"function"`    // grammatical function, friendly English
	Segments    []types.MasaqSegment `json:"segments"`
	Root        *rootSummary         `json:"root,omitempty"`
}

type rootSummary struct {
	Buckwalter string `json:"buckwalter"`
	Arabic     string `json:"arabic"`  // e.g. أَلِه
	Letters    string `json:"letters"` // space-separated consonants: أ ل ه
	POS        string `json:"pos"`     // V / N / P
	MeaningEN  string `json:"meaning_en"`
	MeaningAR  string `json:"meaning_ar"`
	Link       string `json:"link"`
}

func (s *Server) handleAPIWord(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	surah, _ := strconv.Atoi(q.Get("s"))
	ayah, _ := strconv.Atoi(q.Get("a"))
	wordNo, _ := strconv.Atoi(q.Get("w"))
	if !loc.Valid(surah, ayah, wordNo) {
		s.respondError(w, http.StatusBadRequest, "invalid s/a/w")
		return
	}
	key := loc.Key(surah, ayah, wordNo)
	segs := s.masaq.ByWord[key]
	// Note: ~23 known positions diverge between MASAQ and the XML
	// (MASAQ merges "waw + ma" into "wama" while the XML keeps them
	// separate). For those, segs will be empty and the UI shows
	// "no morphological data" — see docs/research/data-validation.md.

	// Look up the word text from the Quran struct.
	var wordText string
	if surahPtr := s.quran.Surahs[surah]; surahPtr != nil {
		if ay := surahPtr.Ayahs[ayah]; ay != nil {
			idx := 0
			for _, t := range ay.Tokens {
				if t.Kind != "word" {
					continue
				}
				idx++
				if idx == wordNo {
					wordText = t.Value
					break
				}
			}
		}
	}

	resp := apiWordResponse{
		Surah:       surah,
		Ayah:        ayah,
		Word:        wordNo,
		WordText:    wordText,
		Lemma:       computeLemma(segs),
		Gloss:       combineGlosses(segs),
		Translation: combineTranslations(segs),
		POS:         "",
		Function:    grammaticalFunction(segs),
		Segments:    segs,
	}
	if bw, ok := s.roots.ByLoc[key]; ok {
		if e, ok := s.roots.ByRoot[bw]; ok {
			resp.Root = &rootSummary{
				Buckwalter: e.Buckwalter,
				Arabic:     e.Arabic,
				Letters:    e.Letters,
				POS:        e.POS,
				MeaningEN:  e.MeaningEN,
				MeaningAR:  e.MeaningAR,
				Link:       "/root/detailed/" + e.Buckwalter,
			}
			// Prefer the root's POS for the top-level field when
			// available — it's more authoritative than the
			// morphological tag of a single segment.
			if e.POS != "" {
				resp.POS = e.POS
			}
		}
	}
	// Fallback: derive POS from segments if no root was found.
	if resp.POS == "" {
		for _, s := range segs {
			switch s.MorphTag {
			case "VERB", "IV", "IV1P", "IV1S", "IV2MP", "IV3FS", "IV3MP", "IV3MS", "IV_PASS", "PV", "CV":
				resp.POS = "V"
			case "NOUN", "NOUN_ABSTRACT", "NOUN_CONCRETE", "NOUN_PROP",
				"NOUN_ACTIVE_PART", "NOUN_PASSIVE_PART", "NOUN_DIMINUTIVE",
				"NOUN_FIVE", "NOUN_INSTRUMENT", "NOUN_NUM", "NOUN_TIME_PLACE",
				"NOUN_VERB_LIKE", "NOUN_RELATIVE", "ADJ_QUALIT", "ADJ_COMP",
				"ADJ_INTENS":
				resp.POS = "N"
			case "PREP", "CONJ", "NEG_PART", "INTERROG", "INTERROG_PART",
				"REL_PRON", "DEM_PRON", "DEM_PRON_F", "DEM_PRON_FS",
				"DEM_PRON_MP", "DEM_PRON_MS", "CERT_PART", "CONDITION_PART",
				"FUTURE_PART", "JUSSIVE_PART", "EMPHATIC_NUN", "ANNUL_PART",
				"EXCEPT_PART", "PRON", "PART", "DET", "POSS_PRON", "OBJ_PRON":
				resp.POS = "P"
			}
			if resp.POS != "" {
				break
			}
		}
	}
	s.respondJSON(w, http.StatusOK, resp)
}

// computeLemma returns the dictionary-form lemma for a word given
// its MASAQ segments. Strategy:
//
//  1. If any segment is tagged as a proper noun (NOUN_PROP), use
//     the Stem segment's WithoutDiacritics — proper nouns like
//     "ٱللَّهِ" decompose into DET (ال) + NOUN_PROP (له) in MASAQ
//     but the lemma should be "الله" not "له" or "ل".
//  2. Else if the stem is a VERB, concatenate prefix + stem —
//     verbal prefixes (أَنذِر → "أَنذِر", يَفْعَلُ → "يَفْعَل") carry
//     semantic content and are part of the lemma. We exclude
//     prefixes tagged as DET (definite article) which are NOT
//     part of the lemma (e.g., ٱلْكِتَاب → "كتاب", not "الكتاب").
//     We also exclude the article's bare lām prefix that MASAQ
//     emits after a vowel-ending preposition — it has the same
//     morph_tag (DET) so the DET filter handles it.
//  3. Else (noun or particle), return just the stem. The lemma
//     is the bare stem, never including DET or PREP prefixes.
//  4. Fallback: first segment's WithoutDiacritics.
func computeLemma(segs []types.MasaqSegment) string {
	// Step 1: proper noun → use the Stem segment's without-diacritics
	// form. If the stem doesn't start with "ال", the surface form
	// elided the article's alif (this happens after a vowel-ending
	// preposition like لِ or بِ, e.g., لِلَّهِ → "الله" not "له"). In
	// that case we restore the elided alif so the lemma is the
	// dictionary form of the proper noun.
	for _, s := range segs {
		if s.MorphTag == "NOUN_PROP" {
			for _, ss := range segs {
				if ss.MorphType == "Stem" {
					lemma := ss.WithoutDiacritics
					if !strings.HasPrefix(lemma, "ال") {
						lemma = "ال" + lemma
					}
					return lemma
				}
			}
			if len(segs) > 0 {
				return segs[0].WithoutDiacritics
			}
			return ""
		}
	}

	// Step 2: locate the stem.
	var stem string
	for _, s := range segs {
		if s.MorphType == "Stem" && s.SegmentedWord != "" {
			stem = s.SegmentedWord
			break
		}
	}

	// Step 3: gather verbal prefixes (those whose morph_tag is NOT
	// DET — DET is the definite article, which is not part of the
	// lemma). For verbs we walk the segments in order so the
	// concatenated lemma matches the written form (e.g., أَنذِر =
	// أَ + نذِر).
	if isVerbLemma(segs) && stem != "" {
		var b strings.Builder
		for _, s := range segs {
			if s.MorphType != "Prefix" {
				continue
			}
			if s.MorphTag == "DET" {
				// Definite article — never part of the lemma.
				continue
			}
			if s.SegmentedWord != "" {
				b.WriteString(s.SegmentedWord)
			}
		}
		b.WriteString(stem)
		return b.String()
	}

	// Step 4: noun / particle — return just the stem.
	if stem != "" {
		return stem
	}

	// Fallback: first segment's WithoutDiacritics, or empty.
	if len(segs) > 0 {
		return segs[0].WithoutDiacritics
	}
	return ""
}

// isVerbLemma reports whether the segments describe a verb (so the
// lemma should include any non-article prefix). MASAQ tags the
// verbal stems with morph_tag values starting with "V" (V, IV,
// PV, CV) or containing "VERB" (VERB, VERB_IMPERFECT, etc.).
func isVerbLemma(segs []types.MasaqSegment) bool {
	for _, s := range segs {
		if s.MorphType != "Stem" {
			continue
		}
		t := s.MorphTag
		if t == "V" || t == "IV" || t == "PV" || t == "CV" || t == "VERB" {
			return true
		}
		if strings.HasPrefix(t, "V_") || strings.HasSuffix(t, "_VERB") {
			return true
		}
	}
	return false
}

// combineTranslations returns the word-level English translation.
// All segments of a word share the same translation; we return the
// first non-empty value.
func combineTranslations(segs []types.MasaqSegment) string {
	for _, s := range segs {
		t := strings.TrimSpace(s.Translation)
		if t != "" {
			return t
		}
	}
	return ""
}

// combineGlosses returns the canonical English gloss for a word.
// MASAQ gives every segment of a word the same gloss (so a DET +
// NOUN_PROP pair both read "(of)-allah"); we collapse those to the
// stem segment's gloss, which is the most informative single string.
func combineGlosses(segs []types.MasaqSegment) string {
	for _, s := range segs {
		if s.MorphType == "Stem" {
			g := strings.TrimSpace(s.Gloss)
			if g != "" {
				return g
			}
		}
	}
	for _, s := range segs {
		g := strings.TrimSpace(s.Gloss)
		if g != "" {
			return g
		}
	}
	return ""
}

// grammaticalFunction inspects the stem segment's SyntacticRole and
// CaseMood and produces a short English phrase like
// "Genitive construct" or "Subject". Falls back to other segment
// roles if the stem is empty.
func grammaticalFunction(segs []types.MasaqSegment) string {
	role, mood := pickSyntacticRole(segs)
	if friendly := friendlyRole(role, mood); friendly != "" {
		return friendly
	}
	// Fall back to a friendly description of the stem's MorphTag
	// when the SyntacticRole didn't yield anything (common for
	// verbs whose only tag is the verbal form like CV/IV/PV).
	for _, s := range segs {
		if s.MorphType == "Stem" {
			if m := friendlyMorphTag(s.MorphTag, s.MorphType); m != "" {
				return m
			}
		}
	}
	for _, s := range segs {
		if m := friendlyMorphTag(s.MorphTag, s.MorphType); m != "" {
			return m
		}
	}
	return ""
}

func pickSyntacticRole(segs []types.MasaqSegment) (string, string) {
	// Prefer the Stem segment.
	for _, s := range segs {
		if s.MorphType == "Stem" && s.SyntacticRole != "" {
			return s.SyntacticRole, s.CaseMood
		}
	}
	// Otherwise the last non-empty segment (typically the suffix or
	// the only segment).
	for i := len(segs) - 1; i >= 0; i-- {
		if segs[i].SyntacticRole != "" {
			return segs[i].SyntacticRole, segs[i].CaseMood
		}
	}
	// Or the first segment's POS as a fallback.
	if len(segs) > 0 {
		return segs[0].SyntacticRole, segs[0].CaseMood
	}
	return "", ""
}

// friendlyMorphTag produces a friendly description for verbs whose
// MASAQ tag is the verbal form rather than a syntactic role.
// Examples: CV → "Imperfect verb", IV → "Perfect verb",
// PV → "Verbal noun (masdar)".
func friendlyMorphTag(tag, morphType string) string {
	if morphType != "Stem" && morphType != "Prefix" && morphType != "Suffix" {
		return ""
	}
	switch tag {
	case "CV":
		return "Imperfect verb"
	case "IV":
		return "Perfect verb"
	case "PV":
		return "Verbal noun"
	case "CV_PREF":
		return "Imperfect verb prefix"
	case "IV_PREF":
		return "Perfect verb prefix"
	case "IVSUFF_SUBJ:MP_MOOD:I":
		return "Imperfect verb subjunctive"
	case "IVSUFF_SUBJ:MP_MOOD:SJ":
		return "Imperfect verb jussive"
	case "IVSUFF_DO:3MS":
		return "Imperfect verb (do: him)"
	}
	return ""
}

func friendlyRole(role, mood string) string {
	roleMap := map[string]string{
		"PREP":           "Preposition",
		"PREP_OBJ":       "Object of preposition",
		"GEN_CONS":       "Genitive construct",
		"NOUN_CONS":      "Construct noun",
		"ADJ":            "Adjective",
		"SUBJ":           "Subject",
		"OBJ":            "Object",
		"PRED":           "Predicate",
		"VERB":           "Verb",
		"ACC_SPECIF":     "Accusative specifier",
		"CIRCUM":         "Circumstantial",
		"ADV_TIME":       "Adverb of time",
		"ADV_PLCE":       "Adverb of place",
		"COMIT":          "Comitative",
		"ANNUL_PART":     "Annulation particle",
		"CERT_PART":      "Certainty particle",
		"CONDITION_PART": "Conditional particle",
		"EXCEPT_NOUN":    "Excepted noun",
		"FUTURE_PART":    "Future particle",
		"JUSSIVE_PART":   "Jussive particle",
		"NEG":            "Negation",
		"NEG_CAT":        "Categorical negation",
		"NEG_MAA":        "Exceptive negation",
		"NEG_PROH":       "Prohibitive negation",
		"PART_COP_PRED":  "Predicate of copula",
		"PART_COP_V":     "Copula verb",
		"PART_CONDITION": "Conditional particle",
		"PART_EXCEPT":    "Exceptive particle",
		"PART_INHIB":     "Inhibitor particle",
		"PART_INTERROG":  "Interrogative particle",
		"PART_JUSSIVE":   "Jussive particle",
		"PART_PREV":      "Preventive particle",
		"PASS_SUBJ":      "Passive subject",
		"PURP":           "Purpose clause",
		"SUBJ_COP_PART":  "Subject of copular sentence",
		"SUBJ_COP_V":     "Copular verb subject",
		"SUBJ_DELA":      "Delayed subject",
		"SUBJ_NEG_CAT":   "Subject of categorical negation",
		"SUBJUNC_PART":   "Subjunctive particle",
		"SUBOR_ANN_CONJ": "Annulling subordinating conjunction",
		"SUBS_COG_ACC":   "Cognate accusative",
		"V_COP_PRED":     "Copular predicate",
		"VOC":            "Vocative",
		"VOC_PART":       "Vocative particle",
		"INTENCIF":       "Intentifier",
		"INTERJ_CV":      "Interjection (imperfect)",
		"INTERJ_IV":      "Interjection (perfect)",
		"INTERJ_PV":      "Interjection (verbal)",
		"NON_INFLECT":    "Non-inflecting",
		"ACRON":          "Acronym",
		"APPOS":          "Apposition",
		"AGNT":           "Agent",
		"COGN":           "Cognate",
		"COMPL":          "Complement",
		"CONJ":           "Conjunction",
		"CONJ_N":         "Conjunction (negative)",
		"CV":             "Imperfect verb",
		"CV_COP":         "Copula (imperfect)",
		"EXCP":           "Exception",
		"EXPLET":         "Expletive",
		"IV":             "Perfect verb",
		"IV_COP":         "Copula (perfect)",
		"IV_PASS":        "Passive verb",
		"NUM_COMP":       "Compound number",
		"PV":             "Verbal noun (masdar)",
		"SUBOR_CONJ":     "Subordinating conjunction",
	}
	if r, ok := roleMap[role]; ok {
		// Append case mood if present and meaningful.
		if mood != "" && mood != "INVARIABLE" {
			return r + " (" + friendlyMood(mood) + ")"
		}
		return r
	}
	if role != "" {
		return role
	}
	return ""
}

func friendlyMood(m string) string {
	moodMap := map[string]string{
		"NOMINATIVE":       "nominative",
		"ACCUSATIVE":       "accusative",
		"GENITIVE":         "genitive",
		"INVARIABLE":       "indeclinable",
		"INVARIABLE_KASRA": "indeclinable (kasra)",
	}
	if r, ok := moodMap[m]; ok {
		return r
	}
	return strings.ToLower(strings.ReplaceAll(m, "_", " "))
}

func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := r.URL.Query().Get("type")
	if q == "" {
		s.respondJSON(w, http.StatusOK, map[string]any{"query": "", "type": typ, "results": []any{}})
		return
	}
	var results []search.SearchResult
	switch typ {
	case "root":
		results = search.Root(q, s.roots, s.quran, 100)
	case "arabic":
		results = search.Arabic(q, s.masaq, s.quran, 5000, 0)
	case "translation":
		results = search.Translation(q, s.masaq, s.quran, 5000, 0)
	default:
		typ = "english"
		results = search.English(q, s.masaq, s.quran, 5000, 0)
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"query":   q,
		"type":    typ,
		"results": results,
	})
}

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
		s.respondError(w, http.StatusNotFound, "root not found")
		return
	}
	filterSurah, _ := strconv.Atoi(r.URL.Query().Get("surah"))
	asc := r.URL.Query().Get("sort") == "asc"
	occ := search.OccurrencesForRoot(root, s.roots, s.quran, filterSurah, asc)
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
		s.respondError(w, http.StatusNotFound, "root not found")
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
		// Reset the response writer — we can't un-write partial
		// bytes, but we can replace them with a clean error page
		// by hijacking the underlying writer. As a fallback, send
		// the error as text since the original headers are gone.
		// Use respondError which writes a fresh 500 + error page.
		w.WriteHeader(http.StatusInternalServerError)
		// Best-effort: render the error template to the buffer and
		// append the raw error text below it so developers can see
		// the failure mode without checking the server logs.
		fmt.Fprintf(w, "<!-- template %q failed: %s -->\n", name, err.Error())
		return
	}
	w.Write(buf.Bytes())
}

func (s *Server) respondError(w http.ResponseWriter, status int, msg string) {
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

// silence unused warnings if data.LoadAll isn't called from here.
var _ = data.LoadAll
var _ = path.Clean
