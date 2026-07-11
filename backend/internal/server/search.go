package server

import (
	"net/http"
	"strings"

	"quranreader/backend/internal/search"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := r.URL.Query().Get("type")
	if q == "" {
		data := struct {
			Title   string
			Query   string
			Type    string
			Results []search.SearchResult
		}{Title: "Search", Query: "", Type: typ}
		s.render(w, "search.tmpl", data)
		return
	}
	var results []search.SearchResult
	switch typ {
	case "root":
		results = search.Root(q, s.roots, s.quran, 2000)
	case "arabic":
		results = search.Arabic(q, s.masaq, s.quran, 5000, 0)
	case "translation":
		results = search.Translation(q, s.masaq, s.quran, 5000, 0)
	case "lemma":
		results = search.Lemma(q, s.masaq, s.quran, 5000, 0)
	default:
		typ = "english"
		results = search.English(q, s.masaq, s.quran, 5000, 0)
	}
	data := struct {
		Title   string
		Query   string
		Type    string
		Results []search.SearchResult
	}{Title: "Search: " + q, Query: q, Type: typ, Results: results}
	s.render(w, "search.tmpl", data)
}

func (s *Server) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	typ := r.URL.Query().Get("type")
	if q == "" {
		s.respondJSON(w, http.StatusOK, map[string]any{"query": "", "type": typ, "results": []any{}})
		return
	}
	var results []search.SearchResult
	switch typ {
	case "root":
		results = search.Root(q, s.roots, s.quran, 100)
	case "arabic":
		results = search.Arabic(q, s.masaq, s.quran, 5000, 0)
	case "translation":
		results = search.Translation(q, s.masaq, s.quran, 5000, 0)
	case "lemma":
		results = search.Lemma(q, s.masaq, s.quran, 5000, 0)
	default:
		typ = "english"
		results = search.English(q, s.masaq, s.quran, 5000, 0)
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"query":   q,
		"type":    typ,
		"results": results,
	})
}
