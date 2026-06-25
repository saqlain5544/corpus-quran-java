// pipeline/build.go — Command build converts the raw data files into
// the gob-encoded form loaded by the server at startup.
//
// Usage:
//
//	go run ./pipeline [options]
//
// Options:
//
//	-data <dir>   directory containing the raw data files
//	              (default: ../data)
//
//	-out <dir>    directory to write processed/*.gob and meta.json
//	              (default: ../data/processed)
//
//	-report <f>   path to write the build verification report
//	              (default: ../docs/data-verification/build-report.txt)
//
// Exit status is non-zero if any verification check fails.
package pipeline

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"quranreader/loc"
	"quranreader/types"
)

// Build is the entry point. It is exported so it can be invoked from
// tests or from a wrapper main.
//
// The three source-file reads (Quran XML, MASAQ CSV, roots JSON +
// meanings AI JSONL) are kicked off in parallel. Each parser then
// runs on its already-buffered byte slice so the parallel window
// covers disk I/O — the bulk of the wall-clock time on cold caches.
func Build(dataDir, outDir, reportPath string) error {
	start := time.Now()

	// Make sure the directories exist.
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return err
	}

	quranPath := filepath.Join(dataDir, "quran", "quran-uthmani.xml")
	masaqPath := filepath.Join(dataDir, "morphology", "MASAQ.csv")
	rootsJSON := filepath.Join(dataDir, "morphology", "corpus-roots.json")
	rootsAI := filepath.Join(dataDir, "morphology", "meanings-roots-ai.jsonl")

	// Read all four raw files concurrently. Each parser then runs
	// on its already-loaded byte slice.
	type readResult struct {
		name string
		data []byte
		err  error
	}
	results := make([]readResult, 4)
	var wg sync.WaitGroup
	wg.Add(4)
	reads := []struct {
		i    int
		name string
		path string
	}{
		{0, "quran", quranPath},
		{1, "masaq", masaqPath},
		{2, "roots-json", rootsJSON},
		{3, "roots-ai", rootsAI},
	}
	for _, r := range reads {
		r := r
		go func() {
			defer wg.Done()
			b, err := os.ReadFile(r.path)
			results[r.i] = readResult{name: r.name, data: b, err: err}
		}()
	}
	wg.Wait()
	for _, r := range results {
		if r.err != nil {
			return fmt.Errorf("read %s: %w", r.name, r.err)
		}
	}

	log := &strings.Builder{}
	logLn(log, "Quran Reader build pipeline")
	logLn(log, "===========================")
	logLn(log, "data dir : %s", dataDir)
	logLn(log, "out dir  : %s", outDir)
	logLn(log, "")

	// 1. Quran.
	logLn(log, "[1/3] parsing %s", quranPath)
	t0 := time.Now()
	quran, err := parseQuran(bytes.NewReader(results[0].data))
	if err != nil {
		return fmt.Errorf("quran: %w", err)
	}
	logLn(log, "      surahs=%d ayas=%d words=%d took=%s",
		len(quran.Surahs), quran.Meta.AyahCount, quran.Meta.WordCount, time.Since(t0))

	// 2. MASAQ.
	logLn(log, "[2/3] parsing %s", masaqPath)
	t0 = time.Now()
	masaq, err := parseMasaq(bytes.NewReader(results[1].data))
	if err != nil {
		return fmt.Errorf("masaq: %w", err)
	}
	logLn(log, "      distinct words=%d took=%s", len(masaq.ByWord), time.Since(t0))

	// 3. Roots.
	logLn(log, "[3/3] parsing %s + %s", rootsJSON, rootsAI)
	t0 = time.Now()
	roots, err := parseRoots(results[2].data, results[3].data)
	if err != nil {
		return fmt.Errorf("roots: %w", err)
	}
	logLn(log, "      distinct roots=%d total occurrences=%d took=%s",
		len(roots.ByRoot), len(roots.ByLoc), time.Since(t0))

	// Verification.
	logLn(log, "")
	logLn(log, "verification")
	logLn(log, "------------")
	failures := 0
	failures += check("surah count == 114", len(quran.Surahs) == 114, log)
	failures += check("ayah count == 6236", quran.Meta.AyahCount == 6236, log)
	failures += check("masaq word count >= 77000", len(masaq.ByWord) >= 77000, log)
	failures += check("root count > 1600", len(roots.ByRoot) > 1600, log)

	// MASAQ strict positional coverage.
	missing := 0
	var missingExamples []string
	for _, s := range quran.Surahs {
		for _, ay := range s.Ayahs {
			for _, tok := range ay.Tokens {
				if tok.Kind != "word" {
					continue
				}
				key := loc.Key(s.Number, ay.Number, tok.WordNo)
				if _, ok := masaq.ByWord[key]; !ok {
					missing++
					if len(missingExamples) < 5 {
						missingExamples = append(missingExamples, loc.String(key)+" "+tok.Value)
					}
				}
			}
		}
	}
	// This is informational, not a failure: MASAQ uses slightly
	// different word boundaries for some particles (e.g., merges
	// "waw + ma" into "wama"). For the ~23 known cases the UI
	// shows "no morphological data" instead of trying a fuzzy
	// match.
	logLn(log, "  [INFO] strict positional MASAQ misses: %d / %d words (shown as no-data in UI)",
		missing, quran.Meta.WordCount)
	for _, ex := range missingExamples {
		logLn(log, "         e.g. %s", ex)
	}

	// Root coverage report.
	rootsCovered := 0
	for _, s := range quran.Surahs {
		for _, ay := range s.Ayahs {
			for _, tok := range ay.Tokens {
				if tok.Kind != "word" {
					continue
				}
				key := loc.Key(s.Number, ay.Number, tok.WordNo)
				if _, ok := roots.ByLoc[key]; ok {
					rootsCovered++
				}
			}
		}
	}
	logLn(log, "  Quran words with a root: %d / %d (%.1f%%)",
		rootsCovered, quran.Meta.WordCount,
		100*float64(rootsCovered)/float64(quran.Meta.WordCount))

	// Emit gob files.
	logLn(log, "")
	logLn(log, "emitting")
	logLn(log, "--------")
	if err := emitGob(filepath.Join(outDir, "quran.gob"), quran); err != nil {
		return err
	}
	logLn(log, "  wrote quran.gob")
	if err := emitGob(filepath.Join(outDir, "masaq_index.gob"), masaq); err != nil {
		return err
	}
	logLn(log, "  wrote masaq_index.gob")
	if err := emitGob(filepath.Join(outDir, "roots.gob"), roots); err != nil {
		return err
	}
	logLn(log, "  wrote roots.gob")
	if err := emitMeta(filepath.Join(outDir, "meta.json"), &quran.Meta); err != nil {
		return err
	}
	logLn(log, "  wrote meta.json")

	// File sizes.
	logLn(log, "")
	logLn(log, "sizes")
	logLn(log, "-----")
	for _, name := range []string{"quran.gob", "masaq_index.gob", "roots.gob", "meta.json"} {
		fi, err := os.Stat(filepath.Join(outDir, name))
		if err != nil {
			continue
		}
		logLn(log, "  %-20s %8d KB", name, fi.Size()/1024)
	}

	logLn(log, "")
	logLn(log, "done in %s, failures=%d", time.Since(start), failures)

	// Write report.
	if err := os.WriteFile(reportPath, []byte(log.String()), 0o644); err != nil {
		return err
	}

	if failures > 0 {
		return fmt.Errorf("%d verification failures (see %s)", failures, reportPath)
	}
	return nil
}

// Main is the default `main` for the pipeline command. We export it
// as a function so it is callable from tests with a stubbed stdin/stdout.
func Main() int {
	dataDir := "../data"
	outDir := "../data/processed"
	reportPath := "../docs/data-verification/build-report.txt"

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-data":
			if i+1 < len(args) {
				dataDir = args[i+1]
				i++
			}
		case "-out":
			if i+1 < len(args) {
				outDir = args[i+1]
				i++
			}
		case "-report":
			if i+1 < len(args) {
				reportPath = args[i+1]
				i++
			}
		case "-h", "-help":
			fmt.Fprintln(os.Stderr, "usage: build [-data DIR] [-out DIR] [-report FILE]")
			return 0
		}
	}

	if err := Build(dataDir, outDir, reportPath); err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		return 1
	}
	return 0
}

func logLn(b *strings.Builder, format string, args ...any) {
	b.WriteString(fmt.Sprintf(format+"\n", args...))
}

func check(name string, ok bool, log *strings.Builder) int {
	if ok {
		logLn(log, "  [PASS] %s", name)
		return 0
	}
	logLn(log, "  [FAIL] %s", name)
	return 1
}

func emitGob(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	return enc.Encode(v)
}

func emitMeta(path string, m *types.Meta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ── Tests moved to build_test.go ─────────────────────────────────
