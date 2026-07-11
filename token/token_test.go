package token

import "testing"

func TestTokenizeSimple(t *testing.T) {
	in := "بِسْمِ ٱللَّهِ ٱلرَّحْمَـٰنِ ٱلرَّحِيمِ"
	got := Tokenize(in)
	if len(got) != 4 {
		t.Fatalf("want 4 tokens, got %d: %+v", len(got), got)
	}
	for i, g := range got {
		if g.Kind != "word" {
			t.Errorf("token %d: kind=%q want word", i, g.Kind)
		}
		if g.WordNo != i+1 {
			t.Errorf("token %d: WordNo=%d want %d", i, g.WordNo, i+1)
		}
	}
}

func TestTokenizeTrailingSajdah(t *testing.T) {
	in := "كَلَّا لَا تُطِعْهُ وَٱسْجُدْ وَٱقْتَرِب ۩"
	toks := Tokenize(in)
	if len(toks) == 0 {
		t.Fatal("empty tokens")
	}
	last := toks[len(toks)-1]
	if last.Kind != "mark" || last.Value != "۩" {
		t.Errorf("last token = %+v, want mark ۩", last)
	}
	secondLast := toks[len(toks)-2]
	if secondLast.Kind != "word" || secondLast.Value != "وَٱقْتَرِب" {
		t.Errorf("second-last = %+v", secondLast)
	}
	if len(secondLast.Marks) != 0 {
		t.Errorf("second-last should have no marks, got %v", secondLast.Marks)
	}
}

func TestTokenizeInlinePause(t *testing.T) {
	in := "ذَٰلِكَ ٱلْكِتَـٰبُ لَا رَيْبَ ۛ فِيهِ"
	toks := Tokenize(in)
	want := []struct {
		kind, value string
	}{
		{"word", "ذَٰلِكَ"},
		{"word", "ٱلْكِتَـٰبُ"},
		{"word", "لَا"},
		{"word", "رَيْبَ"},
		{"mark", "ۛ"},
		{"word", "فِيهِ"},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens want %d: %+v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Kind != w.kind || toks[i].Value != w.value {
			t.Errorf("token %d = %+v want {%s %s}", i, toks[i], w.kind, w.value)
		}
	}
}

func TestTokenizeIqlabHighMeem(t *testing.T) {
	in := "عَلَيٍّۢ بَصِيرًا"
	toks := Tokenize(in)
	if len(toks) != 2 {
		t.Fatalf("expected exactly 2 tokens (word + word), got %d: %+v", len(toks), toks)
	}
	if toks[0].Kind != "word" {
		t.Errorf("token 0 kind = %q want word", toks[0].Kind)
	}
	if len(toks[0].Marks) != 1 || toks[0].Marks[0] != "\u06E2" {
		t.Errorf("token 0 marks = %v want [ۢ]", toks[0].Marks)
	}
	for i, tk := range toks {
		if tk.Kind == "mark" {
			t.Errorf("token %d is an unexpected separate mark token: %+v", i, tk)
		}
	}
}

func TestTokenizeIqlabLowMeem(t *testing.T) {
	in := "ٱلْكِتَـٰبِۭ بِسْمِ"
	toks := Tokenize(in)
	if len(toks) != 2 {
		t.Fatalf("expected exactly 2 tokens, got %d: %+v", len(toks), toks)
	}
	if toks[0].Kind != "word" {
		t.Errorf("token 0 kind = %q want word", toks[0].Kind)
	}
	if len(toks[0].Marks) != 1 || toks[0].Marks[0] != "\u06ED" {
		t.Errorf("token 0 marks = %v want [ۭ]", toks[0].Marks)
	}
}

func TestTokenizeStandaloneMark(t *testing.T) {
	in := "ٱ ۞ بِسْمِ ۖ ٱللَّهِ"
	toks := Tokenize(in)
	want := []struct {
		kind, value string
	}{
		{"word", "ٱ"},
		{"mark", "۞"},
		{"word", "بِسْمِ"},
		{"mark", "ۖ"},
		{"word", "ٱللَّهِ"},
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens want %d: %+v", len(toks), len(want), toks)
	}
	for i, w := range want {
		if toks[i].Kind != w.kind || toks[i].Value != w.value {
			t.Errorf("token %d = {kind=%q value=%q} want {kind=%q value=%q}",
				i, toks[i].Kind, toks[i].Value, w.kind, w.value)
		}
	}
}

func TestSplitMark(t *testing.T) {
	cases := []struct {
		in       string
		wantWord string
		wantMark string
	}{
		{"ٱقْتَرِب", "ٱقْتَرِب", ""},
		{"عَلَيٍّۢ", "عَلَيٍّ", "\u06E2"},
		{"ٱلْكِتَـٰبِۭ", "ٱلْكِتَـٰبِ", "\u06ED"},
	}
	for _, c := range cases {
		w, m := SplitMark(c.in)
		if w != c.wantWord {
			t.Errorf("SplitMark(%q) word = %q want %q", c.in, w, c.wantWord)
		}
		gotMark := ""
		if len(m) > 0 {
			gotMark = m[len(m)-1]
		}
		if gotMark != c.wantMark {
			t.Errorf("SplitMark(%q) marks = %v want last %q", c.in, m, c.wantMark)
		}
	}
}

func TestIsBismillah(t *testing.T) {
	cases := []string{
		BISMILLAH_TEXT,
		"  " + BISMILLAH_TEXT + "  ",
		"ٱلْحَمْدُ لِلَّهِ",
	}
	for i, c := range cases {
		got := IsBismillah(c)
		want := i < 2
		if got != want {
			t.Errorf("IsBismillah(%q) = %v want %v", c, got, want)
		}
	}
}
