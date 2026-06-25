package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"reflect"
	"strings"
	"time"
)

// loadTemplates parses all *.tmpl files from the given filesystem and
// returns one template set PER PAGE, each containing the layout,
// shared partials (e.g. global-header), and the page's own content.
//
// Go's html/template has a subtle gotcha with {{block}}: when the
// layout's `{{block "content" .}}` is shared across many page
// templates that each define a "content" template, the LAST one
// parsed wins — so e.g. the homepage ends up using surah.tmpl's
// content definition and dies at runtime with a field-not-found
// error. Building per-page sets sidesteps this.
//
// Layout + partials are parsed first so they exist before the page
// (which can call them via {{template "global-header" .}} or have
// its content rendered inside the layout's block).
func loadTemplates(fsys TemplatesFS) (map[string]*template.Template, error) {
	entries, err := fsys.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	var pageFiles []string
	layoutBytes := map[string][]byte{}
	partialsBytes := map[string][]byte{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tmpl") {
			continue
		}
		b, err := fsys.ReadFile("templates/" + e.Name())
		if err != nil {
			return nil, err
		}
		switch e.Name() {
		case "layout.tmpl":
			layoutBytes["layout"] = b
		case "homepage.tmpl", "surah.tmpl", "search.tmpl",
			"roots-list.tmpl", "root-detail.tmpl", "about.tmpl",
			"error.tmpl", "concordance.tmpl":
			pageFiles = append(pageFiles, e.Name())
		default:
			// Anything else is a partial (global-header, word-tooltip,
			// word-modal, …). Stored as the base-name "template
			// name" without the .tmpl suffix so it can be referenced
			// via {{template "global-header" .}}.
			base := strings.TrimSuffix(e.Name(), ".tmpl")
			partialsBytes[base] = b
		}
	}
	if _, ok := layoutBytes["layout"]; !ok {
		return nil, fmt.Errorf("layout.tmpl not found")
	}

	out := make(map[string]*template.Template, len(pageFiles))
	for _, name := range pageFiles {
		pageBytes, err := fsys.ReadFile("templates/" + name)
		if err != nil {
			return nil, err
		}
		t, err := buildPageSet(name, pageBytes, layoutBytes["layout"], partialsBytes)
		if err != nil {
			return nil, fmt.Errorf("build %s: %w", name, err)
		}
		out[name] = t
	}
	return out, nil
}

// buildPageSet creates a fresh template.Template per page, populated
// with: layout, all shared partials, and the page itself. The page
// defines its own "content" template that the layout's block resolves
// to. Each page is isolated, so there is no cross-page symbol
// collision.
func buildPageSet(pageName string, pageBytes, layoutBytes []byte, partialsBytes map[string][]byte) (*template.Template, error) {
	t := template.New("").Funcs(funcMap())
	if _, err := t.New("layout").Parse(string(layoutBytes)); err != nil {
		return nil, fmt.Errorf("parse layout: %w", err)
	}
	for name, b := range partialsBytes {
		if _, err := t.New(name).Parse(string(b)); err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", name, err)
		}
	}
	if _, err := t.New(pageName).Parse(string(pageBytes)); err != nil {
		return nil, fmt.Errorf("parse %s: %w", pageName, err)
	}
	return t, nil
}

// unused but kept for compatibility with the funcMap helper.

func funcMap() template.FuncMap {
	return template.FuncMap{
		"truncate": func(s string, n int) string {
			if len(s) <= n {
				return s
			}
			return s[:n] + "…"
		},
		"fmtAyah": func(n int) string {
			// Format surah:ayah in the canonical "S:V" form.
			return fmt.Sprintf("%d", n)
		},
		"inc":    func(n int) int { return n + 1 },
		"dec":    func(n int) int { return n - 1 },
		"sub":    func(a, b int) int { return a - b },
		"len":    func(v any) int { return reflectLen(v) },
		"splitn": func(s, sep string, n int) []string { return strings.SplitN(s, sep, n) },
		"add":    func(a, b, c int) int { return a + b + c },
		"year":   func() int { return time.Now().Year() },
	}
}

// reflectLen returns the length of v, supporting strings, arrays,
// slices, maps, and channels. Returns 0 for any other type.
func reflectLen(v any) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case string:
		return len(x)
	case []string:
		return len(x)
	case []any:
		return len(x)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		return rv.Len()
	}
	return 0
}

// jsonEncode is a thin wrapper to allow swappable encoding later.
func jsonEncode(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
