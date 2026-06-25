// quranreader is the server binary entry point.
//
// Usage:
//
//	quranreader [-addr :8080] [-db ./data/new/detailed-quran.db]
//
// Data is loaded from the detailed-quran.db SQLite database. The
// server embeds templates, static assets and the Hafs font at
// compile time.
package main

import (
	"embed"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"quranreader/backend/internal/data"
	"quranreader/backend/internal/server"
)

//go:embed templates/*.tmpl
var tmplFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed fonts
var fontsFS embed.FS

func main() {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in main", "err", r, "stack", string(debug.Stack()))
		}
	}()

	// Tighten the Go GC: aim for a 96 MB soft memory limit and a
	// moderate GC trigger so we stay snappy without thrashing. The
	// default GOMEMLIMIT is math.MaxInt64 (no limit), which lets
	// the heap grow during spike load (e.g., the first large
	// surah render). Capping at 96 MB keeps RSS predictable.
	if os.Getenv("GOMEMLIMIT") == "" {
		debug.SetMemoryLimit(96 * 1024 * 1024)
	}
	// Reduce GC frequency — we have a stable working set once data
	// is loaded; aggressive GC just adds CPU.
	if os.Getenv("GOGC") == "" {
		_ = debug.SetGCPercent(200)
	}

	// Reasonable defaults for concurrency.
	if runtime.NumCPU() > 4 {
		runtime.GOMAXPROCS(4)
	}

	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "./data/new/detailed-quran.db", "Path to detailed-quran.db")
	transDir := flag.String("translations", "./data/quran/translations", "Translations directory")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("loading data", "path", *dbPath)
	startLoad := time.Now()
	q, m, r, meta, err := data.LoadAll(*dbPath)
	if err != nil {
		logger.Error("load data", "err", err)
		os.Exit(1)
	}
	logger.Info("data loaded",
		"surahs", len(q.Surahs),
		"words", q.Meta.WordCount,
		"ayahs", q.Meta.AyahCount,
		"masaq_entries", len(m.ByWord),
		"roots", len(r.ByRoot),
		"duration_ms", time.Since(startLoad).Milliseconds(),
	)

	// Load verse-level translations (optional) and concordance
	// (optional) in parallel. They're independent I/O-bound paths so
	// running them concurrently shaves noticeable time off startup
	// when the disk cache is cold.
	var (
		translations     *data.Translations
		concordance      *data.Concordance
		loadWg           sync.WaitGroup
		transErr, conErr error
	)
	loadWg.Add(2)
	go func() {
		defer loadWg.Done()
		if *transDir == "" {
			return
		}
		t, err := data.LoadTranslations(*transDir)
		if err != nil {
			transErr = err
			return
		}
		translations = t
	}()
	go func() {
		defer loadWg.Done()
		conPath := "./data/morphology/concordance.jsonl"
		c, err := data.LoadConcordance(conPath)
		if err != nil {
			conErr = err
			return
		}
		concordance = c
	}()
	loadWg.Wait()
	if transErr != nil {
		logger.Warn("load translations", "err", transErr)
		translations = nil
	} else if translations != nil {
		logger.Info("translations loaded", "sets", len(translations.Sets))
	}
	if conErr != nil {
		logger.Warn("load concordance", "err", conErr)
		concordance = nil
	} else if concordance != nil {
		logger.Info("concordance loaded", "roots", len(concordance.ByRoot))
	}

	srv, err := server.New(q, m, r, meta, server.Options{
		TemplatesFS:  tmplFS,
		StaticFS:     staticFS,
		FontsFS:      fontsFS,
		Logger:       logger,
		Translations: translations,
		Concordance:  concordance,
	})
	if err != nil {
		logger.Error("build server", "err", err)
		os.Exit(1)
	}

	logger.Info("listening", "addr", *addr)
	h := srv.Handler()
	if err := http.ListenAndServe(*addr, h); err != nil {
		logger.Error("listen", "err", err)
		os.Exit(1)
	}
}
