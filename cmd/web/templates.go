package main

import (
	"io/fs"
	"path/filepath"
	"text/template"
	"time"

	"snippetbox.prajjmon.net/internal/models"
	"snippetbox.prajjmon.net/ui"
)

type templateData struct {
	CurrentYear     int
	Snippet         models.Snippet
	Snippets        []models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CsrfToken       string
}

// Create a humanDate function which returns a nicely formatted string representation of
// a time.Time object. Note: Using the exact layout "02 Jan 2006 15:04" allows Go to interpret
// each component uniquely, so always refer to these exact values for reliable time formatting in Go.
func humanDate(t time.Time) string {

	if t.IsZero() {
		return ""
	}

	return t.UTC().Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is essentially
// a string-keyed map which acts as a lookup between the names of our custom template
// functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// This essentially gives us a slice of all the 'page' templates for the application
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		// Extract the file name (like 'home.tmpl') from the full filepath
		name := filepath.Base(page)

		// Create slice containing the filepath patterns for the templates we want to parse.
		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		// Parse the template files from the ui.Files embedded filesystem.
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
