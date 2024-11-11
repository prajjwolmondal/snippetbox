package main

import (
	"path/filepath"
	"text/template"
	"time"

	"snippetbox.prajjmon.net/internal/models"
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
	return t.Format("02 Jan 2006 at 15:04")
}

// Initialize a template.FuncMap object and store it in a global variable. This is essentially
// a string-keyed map which acts as a lookup between the names of our custom template
// functions and the functions themselves.
var functions = template.FuncMap{
	"humanDate": humanDate,
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		// Extract the file name (like 'home.tmpl') from the full filepath
		name := filepath.Base(page)

		// The template.FuncMap must be registered with the template set before we call the
		// ParseFiles() method. This means we have to use template.New() to create an empty
		// template set, use the Funcs() method to register the template.FuncMap, and then
		// parse the file as normal.
		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Call ParseGlob() *on this template set* to add any partials
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		// Call ParseFiles() *on this template set* to add the page template.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
