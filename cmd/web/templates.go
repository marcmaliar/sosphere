package main

import (
	"html/template" // for the template cache, template function
	"path/filepath" // glob to get all pages
	"time"          // readable time

	"person.mmaliar.com/internal/models" // template holds model to show in webpage
)

// templateData defines what data will be fed into the HTML template
type templateData struct {
	CurrentYear int               // Current year
	Snippet     *models.Snippet   // One snippet
	Snippets    []*models.Snippet // A list of snippest
	Form        any               // Form results
	Flash       string            // Flash from Session
}

// humanDate converts a time into a more human readable time
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

// functions maps template function names to template functions
// for when templates are parsed
var functions = template.FuncMap{
	"humanDate": humanDate,
}

// newTemplateCache parses and caches all page templates into a map
// using nav and partials. If it fails it returns an error.
func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page) // just "home"

		ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partials/*.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts

	}
	return cache, nil
}
