package main

import (
	"net/http"

	"github.com/justinas/alice"
	"snippetbox.prajjmon.net/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Use the mux.Handle() function to register the file server as the handler for
	// all URL paths that start with "/static/".
	// Use the http.FileServerFS() function to create a HTTP handler which serves
	// the embedded files in ui.Files.
	mux.Handle("GET /static/", http.FileServerFS(ui.Files))

	// This middleware chain is specific to the dynamic app routes (non-static) that
	// are unprotected (AKA no-auth required)
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.home)) // Restrict this route to exact matches on "/" only.
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.snippetView))
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))

	// Protected (authenticated-only) application routes, using a new "protected"
	// middleware chain which includes the requireAuthentication middleware.
	protected := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /snippet/create", protected.ThenFunc(app.snippetCreate))
	mux.Handle("POST /snippet/create", protected.ThenFunc(app.snippetCreatePost))
	mux.Handle("POST /user/logout", protected.ThenFunc(app.userLogoutPost))

	// Create a middleware chain containing our 'standard' middleware which will be used for
	// every request our application receives.
	standard := alice.New(app.recoverFromPanic, app.logRequest, commonHeaders)

	// flow of control (reading from left to right) looks like this:
	// 		logRequest ↔ commonHeaders ↔ servemux ↔ application handler
	return standard.Then(mux)
}
