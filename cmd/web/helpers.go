package main

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/justinas/nosurf"
)

// Writes a log entry at Error level (including the request method and URI as attributes),
// then sends a generic 500 Internal Server Error response to the user.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)

	app.logger.Error(err.Error(), slog.String("method", method), slog.String("uri", uri), slog.String("trace", trace))
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// Sends a specific status code and corresponding description to the user. Should be used to send
// responses like 400 "Bad Request" when there's a problem with the request that the user sent.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, r, err)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)

}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
		CsrfToken:       nosurf.Token(r),
	}
}

func (app *application) decodePostForm(r *http.Request, targetDst any) error {

	// r.ParseForm() adds any data in POST request bodies to the r.PostForm map.
	// This also works in the same way for PUT and PATCH requests.
	err := r.ParseForm()
	if err != nil {
		return err
	}

	err = app.formDecoder.Decode(targetDst, r.PostForm)
	if err != nil {

		// If we try to use an invalid target destination, the Decode() method will return an
		// error with the type *form.InvalidDecoderError. We raise a panic for this particular
		// error rather than returning the error.
		var invalidDecoderError *form.InvalidDecoderError
		if errors.As(err, &invalidDecoderError) {
			panic(err)
		}

		return err
	}

	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)

	// If there isn’t a value in the request context with the isAuthenticatedContextKey key,
	// or the underlying value isn’t a bool, then this type assertion will fail, and we take
	// a "safe" fall back and return false
	if !ok {
		return false
	}

	return isAuthenticated
}
