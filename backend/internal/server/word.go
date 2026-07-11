package server

import (
	"net/http"
	"strconv"

	"quranreader/backend/internal/morph"
	"quranreader/loc"
	"quranreader/types"
)

// ─── Word API types ─────────────────────────────────────────────

type apiWordResponse struct {
	Surah       int                  `json:"surah"`
	Ayah        int                  `json:"ayah"`
	Word        int                  `json:"word"`
	WordText    string               `json:"word_text"`
	Lemma       string               `json:"lemma"`
	Gloss       string               `json:"gloss"`
	Translation string               `json:"translation"`
	POS         string               `json:"pos"`
	Function    string               `json:"function"`
	Segments    []types.MasaqSegment `json:"segments"`
	Root        *rootSummary         `json:"root,omitempty"`
}

// rootSummary enriches the basic root pointer from the /api/word
// response with all fields the tooltip and side panel both need,
// so neither client has to fetch /api/root/{root}/summary separately.
type rootSummary struct {
	Buckwalter   string `json:"buckwalter"`
	Arabic       string `json:"arabic"`
	Letters      string `json:"letters"`
	Occurrences  int    `json:"occurrences"`
	POS          string `json:"pos"`
	MeaningEN    string `json:"meaning_en"`
	MeaningAR    string `json:"meaning_ar"`
	CoreSemantic string `json:"core_semantic,omitempty"`
	Link         string `json:"link"`
}

// ─── Word API handler ──────────────────────────────────────────

func (s *Server) handleAPIWord(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	surah, _ := strconv.Atoi(q.Get("s"))
	ayah, _ := strconv.Atoi(q.Get("a"))
	wordNo, _ := strconv.Atoi(q.Get("w"))
	if !loc.Valid(surah, ayah, wordNo) {
		s.respondError(w, r, http.StatusBadRequest, "invalid s/a/w")
		return
	}
	key := loc.Key(surah, ayah, wordNo)
	segs := s.masaq.ByWord[key]

	var wordText string
	if surahPtr := s.quran.Surahs[surah]; surahPtr != nil {
		if ay := surahPtr.Ayahs[ayah]; ay != nil {
			idx := 0
			for _, t := range ay.Tokens {
				if t.Kind != "word" {
					continue
				}
				idx++
				if idx == wordNo {
					wordText = t.Value
					break
				}
			}
		}
	}

	resp := apiWordResponse{
		Surah:       surah,
		Ayah:        ayah,
		Word:        wordNo,
		WordText:    wordText,
		Lemma:       morph.Lemma(segs),
		Gloss:       morph.Gloss(segs),
		Translation: morph.Translation(segs),
		POS:         "",
		Function:    morph.Function(segs),
		Segments:    segs,
	}
	if s.roots != nil {
		if bw, ok := s.roots.ByLoc[key]; ok {
			if e, ok := s.roots.ByRoot[bw]; ok {
				resp.Root = &rootSummary{
					Buckwalter:   e.Buckwalter,
					Arabic:       e.Arabic,
					Letters:      e.Letters,
					Occurrences:  e.Occurrences,
					POS:          e.POS,
					MeaningEN:    e.MeaningEN,
					MeaningAR:    e.MeaningAR,
					CoreSemantic: e.CoreSemantic,
					Link:         "/root/detailed/" + e.Buckwalter,
				}
				if e.POS != "" {
					resp.POS = e.POS
				}
			}
		}
	}
	if resp.POS == "" {
		for _, sg := range segs {
			switch sg.MorphTag {
			case "VERB", "IV", "IV1P", "IV1S", "IV2MP", "IV3FS", "IV3MP", "IV3MS", "IV_PASS", "PV", "CV":
				resp.POS = "V"
			case "NOUN", "NOUN_ABSTRACT", "NOUN_CONCRETE", "NOUN_PROP",
				"NOUN_ACTIVE_PART", "NOUN_PASSIVE_PART", "NOUN_DIMINUTIVE",
				"NOUN_FIVE", "NOUN_INSTRUMENT", "NOUN_NUM", "NOUN_TIME_PLACE",
				"NOUN_VERB_LIKE", "NOUN_RELATIVE", "ADJ_QUALIT", "ADJ_COMP",
				"ADJ_INTENS":
				resp.POS = "N"
			case "PREP", "CONJ", "NEG_PART", "INTERROG", "INTERROG_PART",
				"REL_PRON", "DEM_PRON", "DEM_PRON_F", "DEM_PRON_FS",
				"DEM_PRON_MP", "DEM_PRON_MS", "CERT_PART", "CONDITION_PART",
				"FUTURE_PART", "JUSSIVE_PART", "EMPHATIC_NUN", "ANNUL_PART",
				"EXCEPT_PART", "PRON", "PART", "DET", "POSS_PRON", "OBJ_PRON":
				resp.POS = "P"
			}
			if resp.POS != "" {
				break
			}
		}
	}
	s.respondJSON(w, http.StatusOK, resp)
}
