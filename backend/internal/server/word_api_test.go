package server

import (
	"testing"

	"quranreader/types"
)

func TestComputeLemmaStem(t *testing.T) {
	// Noun with PREP prefix + Stem: the lemma is just the stem.
	// The preposition is NOT part of the lemma. (Was incorrectly
	// returning "باسم" before the fix — the article-detector was
	// concatenating ALL prefixes regardless of word class.)
	segs := []types.MasaqSegment{
		{Word: "بِسْمِ", WithoutDiacritics: "بسم", SegmentedWord: "ب", MorphType: "Prefix", MorphTag: "PREP"},
		{Word: "بِسْمِ", WithoutDiacritics: "بسم", SegmentedWord: "اسم", MorphType: "Stem", MorphTag: "NOUN_ABSTRACT"},
	}
	if got := computeLemma(segs); got != "اسم" {
		t.Errorf("computeLemma = %q want اسم (stem only — preposition is not part of noun lemma)", got)
	}
}

func TestComputeLemmaCompoundVerb(t *testing.T) {
	// Form IV verb أَنذِرِ (warn!) — MASAQ tags it as Prefix أ (CV_PREF)
	// + Stem نذر (CV). Lemma combines both → أنذر. Verbal prefixes
	// carry semantic content (imperative mood here) so they ARE
	// part of the lemma.
	segs := []types.MasaqSegment{
		{Word: "أَنذِرِ", WithoutDiacritics: "أنذر", SegmentedWord: "أ", MorphType: "Prefix", MorphTag: "CV_PREF"},
		{Word: "أَنذِرِ", WithoutDiacritics: "أنذر", SegmentedWord: "نذر", MorphType: "Stem", MorphTag: "CV"},
	}
	if got := computeLemma(segs); got != "أنذر" {
		t.Errorf("computeLemma = %q want أنذر", got)
	}
}

func TestComputeLemmaDETPrefixExcluded(t *testing.T) {
	// لِّلْمُتَّقِينَ: PREP لِ + DET لْ + NOUN_ACTIVE_PART مُتَّقِ +
	// plural suffix. The DET (article's bare lām after a vowel-ending
	// preposition) must NEVER be part of the lemma. The lemma is
	// just the noun stem.
	segs := []types.MasaqSegment{
		{Word: "لِّلْمُتَّقِينَ", WithoutDiacritics: "ل", SegmentedWord: "لِ", MorphType: "Prefix", MorphTag: "PREP"},
		{Word: "لِّلْمُتَّقِينَ", WithoutDiacritics: "ال", SegmentedWord: "لْ", MorphType: "Prefix", MorphTag: "DET"},
		{Word: "لِّلْمُتَّقِينَ", WithoutDiacritics: "متق", SegmentedWord: "مُتَّقِ", MorphType: "Stem", MorphTag: "NOUN_ACTIVE_PART"},
		{Word: "لِّلْمُتَّقِينَ", WithoutDiacritics: "ين", SegmentedWord: "ينَ", MorphType: "Suffix", MorphTag: "NSUFF_MASC_PL_GEN"},
	}
	if got := computeLemma(segs); got != "مُتَّقِ" {
		t.Errorf("computeLemma = %q want مُتَّقِ (DET prefix excluded, only stem)", got)
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
