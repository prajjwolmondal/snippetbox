package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"snippetbox.prajjmon.net/internal/models"
	"snippetbox.prajjmon.net/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	snippets, err := app.snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.html", data)
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	snippet, err := app.snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{}

	app.render(w, r, http.StatusOK, "create.html", data)
}

// Represents the form data and validation errors for the form fields.
// Note that most of the struct fields are deliberately exported (i.e. start with a
// capital letter). This is because struct fields must be exported in order to be read
// by the html/template package when rendering the template.
type snippetCreateForm struct {
	Title   string
	Content string
	validator.Validator
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {

	// r.ParseForm() adds any data in POST request bodies to the r.PostForm map.
	// This also works in the same way for PUT and PATCH requests.
	err := r.ParseForm()
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := snippetCreateForm{
		// The r.PostForm.Get() method always returns the form data as a *string*.
		Title:   r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "Title can't be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "Title can't be more than 100 chars long")
	form.CheckField(validator.NotBlank(form.Content), "content", "Content field can't be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.snippets.Insert(form.Title, form.Content)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
