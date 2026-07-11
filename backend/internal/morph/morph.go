// Package morph provides pure functions for deriving display-ready
// morphological information from MASAQ segments — lemma, gloss,
// translation, grammatical function, POS classification, friendly
// role/mood names. These operate solely on []types.MasaqSegment;
// they have no HTTP or database dependencies.
package morph

import (
	"strings"

	"quranreader/types"
)

// Lemma returns the dictionary-form lemma for a word given its MASAQ
// segments. Strategy:
//
//  1. Proper noun → Stem segment's WithoutDiacritics (restore elided
//     alif if the article was absorbed by a vowel-ending preposition).
//  2. Verb → concatenate non-DET prefixes + stem.
//  3. Noun/particle → bare stem.
//  4. Fallback → first segment's WithoutDiacritics.
func Lemma(segs []types.MasaqSegment) string {
	// 1. Proper noun.
	for _, s := range segs {
		if s.MorphTag == "NOUN_PROP" {
			for _, ss := range segs {
				if ss.MorphType == "Stem" {
					lemma := ss.WithoutDiacritics
					if !strings.HasPrefix(lemma, "ال") {
						lemma = "ال" + lemma
					}
					return lemma
				}
			}
			// Stem not found; fall back to the first segment's bare form.
			if len(segs) > 0 {
				return segs[0].WithoutDiacritics
			}
			return ""
		}
	}

	// 2. Locate the stem.
	var stem string
	for _, s := range segs {
		if s.MorphType == "Stem" && s.SegmentedWord != "" {
			stem = s.SegmentedWord
			break
		}
	}

	// 3. Verb → prefix + stem.
	if IsVerb(segs) && stem != "" {
		var b strings.Builder
		for _, s := range segs {
			if s.MorphType != "Prefix" {
				continue
			}
			if s.MorphTag == "DET" {
				continue
			}
			if s.SegmentedWord != "" {
				b.WriteString(s.SegmentedWord)
			}
		}
		b.WriteString(stem)
		return b.String()
	}

	// 4. Noun/particle → stem.
	if stem != "" {
		return stem
	}

	// 5. Fallback.
	if len(segs) > 0 {
		return segs[0].WithoutDiacritics
	}
	return ""
}

// IsVerb reports whether the segments describe a verb (MASAQ stem
// tags starting with "V" or containing "VERB").
func IsVerb(segs []types.MasaqSegment) bool {
	for _, s := range segs {
		if s.MorphType != "Stem" {
			continue
		}
		t := s.MorphTag
		if t == "V" || t == "IV" || t == "PV" || t == "CV" || t == "VERB" {
			return true
		}
		if strings.HasPrefix(t, "V_") || strings.HasSuffix(t, "_VERB") {
			return true
		}
	}
	return false
}

// Translation returns the word-level English translation (first
// non-empty segment translation).
func Translation(segs []types.MasaqSegment) string {
	for _, s := range segs {
		t := strings.TrimSpace(s.Translation)
		if t != "" {
			return t
		}
	}
	return ""
}

// Gloss returns the canonical English gloss, preferring the Stem
// segment's gloss as the most informative.
func Gloss(segs []types.MasaqSegment) string {
	for _, s := range segs {
		if s.MorphType == "Stem" {
			g := strings.TrimSpace(s.Gloss)
			if g != "" {
				return g
			}
		}
	}
	for _, s := range segs {
		g := strings.TrimSpace(s.Gloss)
		if g != "" {
			return g
		}
	}
	return ""
}

// Function inspects the stem segment's SyntacticRole and CaseMood
// and produces a short English phrase like "Genitive construct" or
// "Subject". Falls back to a friendly morph-tag description.
func Function(segs []types.MasaqSegment) string {
	role, mood := pickRole(segs)
	if friendly := RoleName(role, mood); friendly != "" {
		return friendly
	}
	for _, s := range segs {
		if s.MorphType == "Stem" {
			if m := TagName(s.MorphTag, s.MorphType); m != "" {
				return m
			}
		}
	}
	for _, s := range segs {
		if m := TagName(s.MorphTag, s.MorphType); m != "" {
			return m
		}
	}
	return ""
}

// pickRole selects the most representative SyntacticRole + CaseMood
// from a segment list (preferring the Stem segment).
func pickRole(segs []types.MasaqSegment) (role, mood string) {
	for _, s := range segs {
		if s.MorphType == "Stem" && s.SyntacticRole != "" {
			return s.SyntacticRole, s.CaseMood
		}
	}
	for i := len(segs) - 1; i >= 0; i-- {
		if segs[i].SyntacticRole != "" {
			return segs[i].SyntacticRole, segs[i].CaseMood
		}
	}
	if len(segs) > 0 {
		return segs[0].SyntacticRole, segs[0].CaseMood
	}
	return "", ""
}

// TagName produces a friendly description for a MASAQ morph tag.
func TagName(tag, morphType string) string {
	if morphType != "Stem" && morphType != "Prefix" && morphType != "Suffix" {
		return ""
	}
	switch tag {
	case "CV":
		return "Imperfect verb"
	case "IV":
		return "Perfect verb"
	case "PV":
		return "Verbal noun"
	case "CV_PREF":
		return "Imperfect verb prefix"
	case "IV_PREF":
		return "Perfect verb prefix"
	case "IVSUFF_SUBJ:MP_MOOD:I":
		return "Imperfect verb subjunctive"
	case "IVSUFF_SUBJ:MP_MOOD:SJ":
		return "Imperfect verb jussive"
	case "IVSUFF_DO:3MS":
		return "Imperfect verb (do: him)"
	}
	return ""
}

// RoleName maps a MASAQ syntactic role code to a human-readable
// English label, optionally appending the case mood.
func RoleName(role, mood string) string {
	roleMap := map[string]string{
		"PREP":           "Preposition",
		"PREP_OBJ":       "Object of preposition",
		"GEN_CONS":       "Genitive construct",
		"NOUN_CONS":      "Construct noun",
		"ADJ":            "Adjective",
		"SUBJ":           "Subject",
		"OBJ":            "Object",
		"PRED":           "Predicate",
		"VERB":           "Verb",
		"ACC_SPECIF":     "Accusative specifier",
		"CIRCUM":         "Circumstantial",
		"ADV_TIME":       "Adverb of time",
		"ADV_PLCE":       "Adverb of place",
		"COMIT":          "Comitative",
		"ANNUL_PART":     "Annulation particle",
		"CERT_PART":      "Certainty particle",
		"CONDITION_PART": "Conditional particle",
		"EXCEPT_NOUN":    "Excepted noun",
		"FUTURE_PART":    "Future particle",
		"JUSSIVE_PART":   "Jussive particle",
		"NEG":            "Negation",
		"NEG_CAT":        "Categorical negation",
		"NEG_MAA":        "Exceptive negation",
		"NEG_PROH":       "Prohibitive negation",
		"PART_COP_PRED":  "Predicate of copula",
		"PART_COP_V":     "Copula verb",
		"PART_CONDITION": "Conditional particle",
		"PART_EXCEPT":    "Exceptive particle",
		"PART_INHIB":     "Inhibitor particle",
		"PART_INTERROG":  "Interrogative particle",
		"PART_JUSSIVE":   "Jussive particle",
		"PART_PREV":      "Preventive particle",
		"PASS_SUBJ":      "Passive subject",
		"PURP":           "Purpose clause",
		"SUBJ_COP_PART":  "Subject of copular sentence",
		"SUBJ_COP_V":     "Copular verb subject",
		"SUBJ_DELA":      "Delayed subject",
		"SUBJ_NEG_CAT":   "Subject of categorical negation",
		"SUBJUNC_PART":   "Subjunctive particle",
		"SUBOR_ANN_CONJ": "Annulling subordinating conjunction",
		"SUBS_COG_ACC":   "Cognate accusative",
		"V_COP_PRED":     "Copular predicate",
		"VOC":            "Vocative",
		"VOC_PART":       "Vocative particle",
		"INTENCIF":       "Intentifier",
		"INTERJ_CV":      "Interjection (imperfect)",
		"INTERJ_IV":      "Interjection (perfect)",
		"INTERJ_PV":      "Interjection (verbal)",
		"NON_INFLECT":    "Non-inflecting",
		"ACRON":          "Acronym",
		"APPOS":          "Apposition",
		"AGNT":           "Agent",
		"COGN":           "Cognate",
		"COMPL":          "Complement",
		"CONJ":           "Conjunction",
		"CONJ_N":         "Conjunction (negative)",
		"CV":             "Imperfect verb",
		"CV_COP":         "Copula (imperfect)",
		"EXCP":           "Exception",
		"EXPLET":         "Expletive",
		"IV":             "Perfect verb",
		"IV_COP":         "Copula (perfect)",
		"IV_PASS":        "Passive verb",
		"NUM_COMP":       "Compound number",
		"PV":             "Verbal noun (masdar)",
		"SUBOR_CONJ":     "Subordinating conjunction",
	}
	if r, ok := roleMap[role]; ok {
		if mood != "" && mood != "INVARIABLE" {
			return r + " (" + MoodName(mood) + ")"
		}
		return r
	}
	if role != "" {
		return role
	}
	return ""
}

// MoodName maps a MASAQ case mood code to a human-readable label.
func MoodName(m string) string {
	moodMap := map[string]string{
		"NOMINATIVE":       "nominative",
		"ACCUSATIVE":       "accusative",
		"GENITIVE":         "genitive",
		"INVARIABLE":       "indeclinable",
		"INVARIABLE_KASRA": "indeclinable (kasra)",
	}
	if r, ok := moodMap[m]; ok {
		return r
	}
	return strings.ToLower(strings.ReplaceAll(m, "_", " "))
}
