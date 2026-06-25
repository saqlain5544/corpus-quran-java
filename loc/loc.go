// Package loc provides the canonical 64-bit word-location key used by
// every index in the system. Using a single integer instead of a
// "{surah}:{ayah}:{word}" string cuts allocations and hashmap cost
// dramatically. The encoding is:
//
//	LocKey = surah * 1_000_000 + ayah * 1_000 + word
//
// Maximum observed values: surah ≤ 114, ayah ≤ 286, word ≤ ~130, so
// the result is well under 2^32 even before we widen to uint64.
package loc

// MaxSurah is the largest surah number (inclusive).
const MaxSurah = 114

// MaxAyah is the largest ayah number observed in the corpus (Surah 2,
// Ayat al-Kursi). Generous headroom used for validation.
const MaxAyah = 300

// MaxWord is the largest word-position-in-ayah observed. Generous
// headroom used for validation.
const MaxWord = 200

// Key encodes (surah, ayah, word) into a single uint64.
//
// Panics only if the caller provides an out-of-range value; otherwise
// returns the packed key directly. The encoding is monotonic in the
// lexicographic order of (surah, ayah, word).
func Key(surah, ayah, word int) uint64 {
	if surah < 1 || surah > MaxSurah {
		panic("loc: surah out of range")
	}
	if ayah < 1 || ayah > MaxAyah {
		panic("loc: ayah out of range")
	}
	if word < 1 || word > MaxWord {
		panic("loc: word out of range")
	}
	return uint64(surah)*1_000_000 + uint64(ayah)*1_000 + uint64(word)
}

// Decode is the inverse of Key. Useful for debug/log output.
func Decode(k uint64) (surah, ayah, word int) {
	word = int(k % 1000)
	ayah = int(k/1000) % 1000
	surah = int(k / 1_000_000)
	return
}

// String renders the key in the canonical "s:v:w" form used by
// corpus-roots.json, log messages, and the search snippet output.
func String(k uint64) string {
	s, a, w := Decode(k)
	// Avoid fmt.Sprintf allocation in the hot path: this is fine
	// because Sprintf is not on any per-request critical path, but
	// we keep it readable.
	return itoa(s) + ":" + itoa(a) + ":" + itoa(w)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Valid reports whether the key encodes a plausible location. It is a
// cheap pre-flight check before hashmap lookups.
func Valid(surah, ayah, word int) bool {
	return surah >= 1 && surah <= MaxSurah &&
		ayah >= 1 && ayah <= MaxAyah &&
		word >= 1 && word <= MaxWord
}
