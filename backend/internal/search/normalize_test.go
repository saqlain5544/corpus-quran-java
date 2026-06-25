package search

import "testing"

func TestNormalizeArabic(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// strip tashkeel
		{"بِسْمِ", "بسم"},
		{"ٱللَّهِ", "الله"},
		// unify alif variants
		{"إِنَّ", "ان"},
		{"أَ", "ا"},
		{"ٱلْكِتَاب", "الكتاب"},
		// unify ya
		{"عَلَىٰ", "علي"},
		// unify ta marbuta
		{"نِعْمَة", "نعمه"},
		// collapse whitespace
		{"بسم   الله", "بسم الله"},
		// idempotent
		{"ٱلرَّحْمَـٰنِ", "الرحمن"},
	}
	for _, c := range cases {
		got := NormalizeArabic(c.in)
		if got != c.want {
			t.Errorf("NormalizeArabic(%q) = %q want %q", c.in, got, c.want)
		}
		// idempotency
		if again := NormalizeArabic(got); again != got {
			t.Errorf("not idempotent: %q → %q → %q", c.in, got, again)
		}
	}
}

func TestBuckwalter(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"قَوْل", "qwl"},
		{"كَوْن", "kwn"},
		{"رَبّ", "rbb"},
	}
	for _, c := range cases {
		got := Buckwalter(c.in)
		if got != c.want {
			t.Errorf("Buckwalter(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"same", "same", 0},
		{"abc", "abd", 1},
	}
	for _, c := range cases {
		got := Levenshtein(c.a, c.b)
		if got != c.want {
			t.Errorf("Levenshtein(%q,%q) = %d want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestHighlightSnippet(t *testing.T) {
	cases := []struct {
		text, match, want string
	}{
		{"hello world", "world", "hello <mark>world</mark>"},
		{"no match here", "missing", "no match here"},
		{"only first", "first", "only <mark>first</mark>"},
	}
	for _, c := range cases {
		got := HighlightSnippet(c.text, c.match)
		if got != c.want {
			t.Errorf("HighlightSnippet(%q,%q) = %q want %q", c.text, c.match, got, c.want)
		}
	}
}
