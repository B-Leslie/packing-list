// Package web implements the HTTP layer: routing, middleware, handlers, templates.
package web

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"strings"
)

//go:embed templates/*.html templates/pages/*.html templates/partials/*.html
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

type Renderer struct {
	pages    map[string]*template.Template
	partials *template.Template
}

func NewRenderer() (*Renderer, error) {
	funcs := template.FuncMap{
		"add":    func(a, b int) int { return a + b },
		"isZero": func(v any) bool { return v == nil || v == "" || v == 0 },
		"plus":   func(a, b int) int { return a + b },
	}

	base := template.New("").Funcs(funcs)
	base, err := base.ParseFS(templatesFS, "templates/layout.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse base: %w", err)
	}

	pages := map[string]*template.Template{}
	entries, err := templatesFS.ReadDir("templates/pages")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".html")
		cloned, err := base.Clone()
		if err != nil {
			return nil, err
		}
		t, err := cloned.ParseFS(templatesFS, "templates/pages/"+e.Name())
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		pages[name] = t
	}
	return &Renderer{pages: pages, partials: base}, nil
}

// Render writes a page template by name.
func (r *Renderer) Render(w io.Writer, page string, data any) error {
	t, ok := r.pages[page]
	if !ok {
		return errors.New("unknown page: " + page)
	}
	return t.ExecuteTemplate(w, "layout", data)
}

// Partial renders a partial template by name (for HTMX fragment responses).
func (r *Renderer) Partial(w io.Writer, name string, data any) error {
	return r.partials.ExecuteTemplate(w, name, data)
}

// StaticFS returns the embedded /static FS for use with http.FileServer.
func StaticFS() embed.FS { return staticFS }
