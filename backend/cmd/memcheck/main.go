// memcheck is a small helper that loads all data and prints
// runtime.MemStats. Used during integration to verify we stay
// under the 80 MB target.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"quranreader/backend/internal/data"
)

func main() {
	dbPath := "../data/new/detailed-quran.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	q, m, r, meta, err := data.LoadAll(dbPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "load:", err)
		os.Exit(1)
	}
	// Force GC and return memory to OS.
	runtime.GC()
	debug.FreeOSMemory()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	fmt.Printf("HeapAlloc:   %6.1f MB\n", float64(ms.HeapAlloc)/1024/1024)
	fmt.Printf("HeapSys:     %6.1f MB\n", float64(ms.HeapSys)/1024/1024)
	fmt.Printf("HeapInuse:   %6.1f MB\n", float64(ms.HeapInuse)/1024/1024)
	fmt.Printf("HeapObjects: %d\n", ms.HeapObjects)
	fmt.Printf("Sys:         %6.1f MB\n", float64(ms.Sys)/1024/1024)
	fmt.Printf("NumGC:       %d\n", ms.NumGC)
	fmt.Printf("Surahs=%d Words=%d MasaqEntries=%d Roots=%d\n",
		len(q.Surahs), q.Meta.WordCount, len(m.ByWord), len(r.ByRoot))
	_ = meta
}
