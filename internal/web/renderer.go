package web

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"time"

	"github.com/labstack/echo/v4"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

//go:embed static/*
var StaticFS embed.FS

// TemplateRenderer implements echo.Renderer using html/template.
// Each page gets its own template set (layout + page) so that
// multiple {{define "content"}} blocks don't collide.
type TemplateRenderer struct {
	templates map[string]*template.Template
}

var funcMap = template.FuncMap{
	"formatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05 UTC")
	},
	"shortID": func(id string) string {
		if len(id) > 8 {
			return id[:8]
		}
		return id
	},
	"isRevoked": func(t *time.Time) bool {
		return t != nil
	},
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"seq": func(n int) []int {
		s := make([]int, n)
		for i := range s {
			s[i] = i
		}
		return s
	},
	"renderTime": func() string {
		return ""
	},
}

// NewRenderer parses embedded templates and returns a renderer.
// Pages that use the layout are parsed as layout.tmpl + <page>.tmpl pairs.
// The login page is standalone (no layout).
func NewRenderer() *TemplateRenderer {
	pages := map[string][]string{
		"login":      {"templates/login.tmpl"},
		"providers":  {"templates/layout.tmpl", "templates/providers.tmpl"},
		"keys":       {"templates/layout.tmpl", "templates/keys.tmpl"},
		"usage":      {"templates/layout.tmpl", "templates/usage.tmpl"},
		"guardrails": {"templates/layout.tmpl", "templates/guardrails.tmpl"},
		"playground": {"templates/layout.tmpl", "templates/playground.tmpl"},
	}

	templates := make(map[string]*template.Template)
	for name, files := range pages {
		t := template.Must(
			template.New("").Funcs(funcMap).ParseFS(templateFS, files...),
		)
		templates[name] = t
	}

	return &TemplateRenderer{templates: templates}
}

// layoutPages lists pages that use the "layout" wrapper.
var layoutPages = map[string]bool{
	"providers":  true,
	"keys":       true,
	"usage":      true,
	"guardrails": true,
	"playground": true,
}

// Render satisfies the echo.Renderer interface.
// Pass the page name (e.g. "providers", "login"). For layout-wrapped pages
// the renderer executes the "layout" define from that page's template set.
// It injects a RenderTime field into the template data measuring execution time.
func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	t, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %q not found", name)
	}

	start := time.Now()

	// Clone the template to inject a per-request renderTime function.
	clone, err := t.Clone()
	if err != nil {
		return err
	}
	clone.Funcs(template.FuncMap{
		"renderTime": func() string {
			return time.Since(start).String()
		},
	})

	if layoutPages[name] {
		return clone.ExecuteTemplate(w, "layout", data)
	}
	return clone.ExecuteTemplate(w, name, data)
}
