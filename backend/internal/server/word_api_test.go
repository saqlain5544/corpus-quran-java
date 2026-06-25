package server

import (
	"testing"

	"quranreader/types"
)

func TestComputeLemmaStem(t *testing.T) {
	// Word with Prefix + Stem: lemma should be prefix + stem.
	segs := []types.MasaqSegment{
		{Word: "بِسْمِ", WithoutDiacritics: "بسم", SegmentedWord: "ب", MorphType: "Prefix"},
		{Word: "بِسْمِ", WithoutDiacritics: "بسم", SegmentedWord: "اسم", MorphType: "Stem"},
	}
	if got := computeLemma(segs); got != "باسم" {
		t.Errorf("computeLemma = %q want باسم", got)
	}
}

func TestComputeLemmaCompoundVerb(t *testing.T) {
	// Form IV verb أَنذِرِ (warn!) — MASAQ tags it as Prefix أ (CV_PREF)
	// + Stem نذر (CV). Lemma should combine both → أنذر.
	segs := []types.MasaqSegment{
		{Word: "أَنذِرِ", WithoutDiacritics: "أنذر", SegmentedWord: "أ", MorphType: "Prefix"},
		{Word: "أَنذِرِ", WithoutDiacritics: "أنذر", SegmentedWord: "نذر", MorphType: "Stem"},
	}
	if got := computeLemma(segs); got != "أنذر" {
		t.Errorf("computeLemma = %q want أنذر", got)
	}
}

func TestComputeLemmaProperNoun(t *testing.T) {
	// Proper noun (NOUN_PROP): lemma should be the full word
	// without diacritics, NOT just the stem segment.
	segs := []types.MasaqSegment{
		{Word: "ٱللَّهِ", WithoutDiacritics: "الله", SegmentedWord: "ال", MorphType: "Prefix"},
		{Word: "ٱللَّهِ", WithoutDiacritics: "الله", SegmentedWord: "له", MorphType: "Stem", MorphTag: "NOUN_PROP"},
	}
	if got := computeLemma(segs); got != "الله" {
		t.Errorf("computeLemma = %q want الله (proper noun full form)", got)
	}
}

func TestComputeLemmaNoStem(t *testing.T) {
	// Word with no Stem segment: fall back to first segment's
	// WithoutDiacritics.
	segs := []types.MasaqSegment{
		{Word: "كَ", WithoutDiacritics: "ك", SegmentedWord: "ك", MorphType: "Prefix"},
		{Word: "كَ", WithoutDiacritics: "ك", SegmentedWord: "", MorphType: "Other"},
	}
	if got := computeLemma(segs); got != "ك" {
		t.Errorf("computeLemma = %q want ك", got)
	}
}

func TestComputeLemmaEmpty(t *testing.T) {
	if got := computeLemma(nil); got != "" {
		t.Errorf("computeLemma(nil) = %q", got)
	}
}

func TestCombineGlosses(t *testing.T) {
	segs := []types.MasaqSegment{
		{Gloss: "in-(the)-name", MorphType: "Stem"},
		{Gloss: "in-(the)-name", MorphType: "Prefix"},
	}
	if got := combineGlosses(segs); got != "in-(the)-name" {
		t.Errorf("combineGlosses = %q want in-(the)-name (stem segment wins)", got)
	}

	segs2 := []types.MasaqSegment{
		{Gloss: ""},
		{Gloss: "wrote", MorphType: "Stem"},
		{Gloss: "it", MorphType: "Suffix"},
	}
	if got := combineGlosses(segs2); got != "wrote" {
		t.Errorf("combineGlosses = %q want wrote", got)
	}

	segs3 := []types.MasaqSegment{
		{Gloss: "(of)-allah", MorphType: "Prefix"},
		{Gloss: "(of)-allah", MorphType: "Stem"},
	}
	if got := combineGlosses(segs3); got != "(of)-allah" {
		t.Errorf("combineGlosses = %q want (of)-allah (no duplication)", got)
	}
}

func TestGrammaticalFunction(t *testing.T) {
	segs := []types.MasaqSegment{
		{SyntacticRole: "PREP", MorphType: "Prefix"},
		{SyntacticRole: "PREP_OBJ", MorphType: "Stem"},
	}
	if got := grammaticalFunction(segs); got == "" {
		t.Error("expected non-empty function")
	}

	segs2 := []types.MasaqSegment{
		{SyntacticRole: "GEN_CONS", CaseMood: "GENITIVE", MorphType: "Stem"},
	}
	if got := grammaticalFunction(segs2); got == "" {
		t.Error("expected non-empty function for GEN_CONS")
	}

	// Empty segments should yield empty function.
	if got := grammaticalFunction(nil); got != "" {
		t.Errorf("grammaticalFunction(nil) = %q want empty", got)
	}
}

func TestFriendlyRoleKnown(t *testing.T) {
	cases := []struct{ in, want string }{
		{"PREP", "Preposition"},
		{"PREP_OBJ", "Object of preposition"},
		{"GEN_CONS", "Genitive construct"},
		{"ADJ", "Adjective"},
		{"SUBJ", "Subject"},
		{"OBJ", "Object"},
		{"PRED", "Predicate"},
	}
	for _, c := range cases {
		if got := friendlyRole(c.in, ""); got != c.want {
			t.Errorf("friendlyRole(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestFriendlyRoleWithCaseMood(t *testing.T) {
	got := friendlyRole("GEN_CONS", "GENITIVE")
	if got == "" {
		t.Error("expected non-empty with case mood")
	}
	// INVARIABLE should not be appended.
	gotInv := friendlyRole("PREP", "INVARIABLE")
	if gotInv != "Preposition" {
		t.Errorf("friendlyRole INVARIABLE not stripped: %q", gotInv)
	}
}

func TestFriendlyMood(t *testing.T) {
	cases := []struct{ in, want string }{
		{"NOMINATIVE", "nominative"},
		{"ACCUSATIVE", "accusative"},
		{"GENITIVE", "genitive"},
		{"UNKNOWN_VALUE", "unknown value"},
	}
	for _, c := range cases {
		if got := friendlyMood(c.in); got != c.want {
			t.Errorf("friendlyMood(%q) = %q want %q", c.in, got, c.want)
		}
	}
}

func TestPickSyntacticRolePrefersStem(t *testing.T) {
	segs := []types.MasaqSegment{
		{SyntacticRole: "PREP", MorphType: "Prefix"},
		{SyntacticRole: "OBJ", MorphType: "Stem"},
		{SyntacticRole: "PRON", MorphType: "Suffix"},
	}
	role, mood := pickSyntacticRole(segs)
	if role != "OBJ" {
		t.Errorf("pickSyntacticRole = %q want OBJ", role)
	}
	if mood != "" {
		t.Errorf("mood = %q want empty", mood)
	}
}

func TestGrammaticalFunctionVerbFallback(t *testing.T) {
	// Verb tagged only as CV with no SyntacticRole — function
	// should fall back to friendlyMorphTag and show "Imperfect verb".
	segs := []types.MasaqSegment{
		{MorphTag: "CV_PREF", MorphType: "Prefix"},
		{MorphTag: "CV", MorphType: "Stem"},
	}
	got := grammaticalFunction(segs)
	if got != "Imperfect verb" {
		t.Errorf("grammaticalFunction = %q want Imperfect verb", got)
	}
}

func TestFriendlyMorphTag(t *testing.T) {
	cases := []struct{ tag, morphType, want string }{
		{"CV", "Stem", "Imperfect verb"},
		{"IV", "Stem", "Perfect verb"},
		{"PV", "Stem", "Verbal noun"},
		{"CV_PREF", "Prefix", "Imperfect verb prefix"},
		{"UNKNOWN", "Stem", ""},
		{"CV", "Other", ""}, // only stem/prefix/suffix get mapped
	}
	for _, c := range cases {
		if got := friendlyMorphTag(c.tag, c.morphType); got != c.want {
			t.Errorf("friendlyMorphTag(%q,%q) = %q want %q", c.tag, c.morphType, got, c.want)
		}
	}
}
