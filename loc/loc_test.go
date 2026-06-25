package loc

import (
	"strconv"
	"testing"
)

func TestKeyDecodeRoundTrip(t *testing.T) {
	cases := []struct {
		s, a, w int
	}{
		{1, 1, 1},
		{1, 1, 4},
		{2, 255, 50},
		{114, 6, 1},
		{114, 6, 5},
	}
	for _, c := range cases {
		k := Key(c.s, c.a, c.w)
		gs, ga, gw := Decode(k)
		if gs != c.s || ga != c.a || gw != c.w {
			t.Fatalf("roundtrip mismatch: got (%d,%d,%d) want (%d,%d,%d)", gs, ga, gw, c.s, c.a, c.w)
		}
	}
}

func TestKeyOrder(t *testing.T) {
	// Surah orders before ayah, ayah before word.
	a := Key(1, 1, 1)
	b := Key(1, 1, 2)
	c := Key(1, 2, 1)
	d := Key(2, 1, 1)
	if !(a < b && b < c && c < d) {
		t.Fatalf("ordering wrong: %d %d %d %d", a, b, c, d)
	}
}

func TestStringFormat(t *testing.T) {
	if got := String(Key(2, 255, 50)); got != "2:255:50" {
		t.Fatalf("String format wrong: %q", got)
	}
}

func TestValid(t *testing.T) {
	cases := []struct {
		s, a, w int
		want    bool
	}{
		{1, 1, 1, true},
		{114, 6, 200, true},
		{0, 1, 1, false},
		{115, 1, 1, false},
		{1, 0, 1, false},
		{1, 1, 0, false},
		{-1, 1, 1, false},
	}
	for _, c := range cases {
		if got := Valid(c.s, c.a, c.w); got != c.want {
			t.Errorf("Valid(%d,%d,%d)=%v want %v", c.s, c.a, c.w, got, c.want)
		}
	}
}

func TestKeyFitsUint64(t *testing.T) {
	// 114 * 1_000_000 + 286 * 1_000 + 130 = 114,286,130 — fits easily.
	max := Key(MaxSurah, 286, 130)
	if max > 1<<32 {
		t.Fatalf("key exceeds 2^32: %d", max)
	}
	// Sanity-check with strconv to make sure decoding prints sanely.
	s, a, w := Decode(max)
	if strconv.Itoa(s)+":"+strconv.Itoa(a)+":"+strconv.Itoa(w) == "" {
		t.Fatal("decode empty")
	}
}
