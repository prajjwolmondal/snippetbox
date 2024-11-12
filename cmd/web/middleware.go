package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// There are also Set(), Del(), Get() and Values() methods that you can use to manipulate and read from the header map
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		// Any code here will execute on the way down the chain.
		next.ServeHTTP(w, r)
		// Any code here will execute on the way back up the chain.
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}

// Note: our middleware will only recover panics that happen in the same goroutine that
// executed the recoverPanic() middleware. If, for example, you have a handler which spins up
// another goroutine (e.g. to do some background processing), then any panics that happen in
// the second goroutine will not be recovered — not by the recoverPanic() middleware… and not
// by the panic recovery built into Go HTTP server. They will cause your application to exit
// and bring down the server. So, if you are spinning up additional goroutines from within
// your web application and there is any chance of a panic, you must make sure that you
// recover any panics from within those too.
func (app *application) recoverFromPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Create a deferred function (which will always be run in the event of a panic as
		// Go unwinds the stack).
		defer func() {
			// recover() retrieves the error value passed to the call of panic
			if err := recover(); err != nil {
				// This header acts as a trigger to make Go’s HTTP server automatically close
				// the current connection after a response has been sent. It also informs the
				// user that the connection will be closed
				w.Header().Set("Connection", "close")

				// Since the recover() function has the type "any", and its underlying type
				// could be string, error, or something else — we normalize this into an error
				// by using the fmt.Errorf() function to create a new error object containing
				// the default textual representation of the "any" value
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}() // The "()" here can be used to pass arguments to the anon function (func) - if it had params

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !app.isAuthenticated(r) {
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return // return from the middleware chain so that no subsequent handlers in the chain are executed.
		}

		// Set the "Cache-Control: no-store" header so that pages that require auth are
		// not stored in the users browser cache (or other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}

// Creates a NoSurf middleware function which uses a customized CSRF cookie with
// the Secure, Path and HttpOnly attributes set.
func noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
	})

	return csrfHandler
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetInt(r.Context(), "authenticatedUserID")
		if id == 0 { // "authenticatedUserID" not found in context, proceeding to next handler in chain
			next.ServeHTTP(w, r)
			return
		}

		exists, err := app.users.Exists(id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		if exists {
			cntxt := context.WithValue(r.Context(), isAuthenticatedContextKey, true)
			r = r.WithContext(cntxt)
		}

		next.ServeHTTP(w, r)
	})
}
