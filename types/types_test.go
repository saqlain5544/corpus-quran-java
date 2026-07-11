package types

import "testing"

// TestNormalizeArabicRoot verifies the canonical form used by the
// ByArabic reverse lookup. This is the single source of truth shared
// by the load path (data.LoadAll) and the query path
// (search.Root, server.handleRootDetail); changes here must keep the
// data tests in lock-step.
func TestNormalizeArabicRoot(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		// Strip spaces.
		{"ق و ل", "قول"},
		{"أ ب ي", "ابي"},
		// Unify alif variants.
		{"إِنَّ", "ان"},
		{"أَ", "ا"},
		{"ٱلْكِتَاب", "الكتاب"},
		// Strip tashkeel.
		{"بِسْمِ", "بسم"},
		// Unify ya variants.
		{"عَلَىٰ", "علي"},
		// Unify ta marbuta.
		{"نِعْمَة", "نعمه"},
		// Idempotent.
		{"ٱلرَّحْمَـٰنِ", "الرحمن"},
		// Empty in → empty out.
		{"", ""},
		// Only spaces → empty out.
		{"   ", ""},
	}
	for _, c := range cases {
		got := NormalizeArabicRoot(c.in)
		if got != c.want {
			t.Errorf("NormalizeArabicRoot(%q) = %q want %q", c.in, got, c.want)
		}
		// Idempotency.
		if again := NormalizeArabicRoot(got); again != got {
			t.Errorf("not idempotent: %q → %q → %q", c.in, got, again)
		}
	}
}
