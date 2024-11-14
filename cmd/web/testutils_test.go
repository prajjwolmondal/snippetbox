package main

import (
	"bytes"
	"html"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"snippetbox.prajjmon.net/internal/models/mocks"
)

// Returns an instance of our application struct containing mocked dependencies
func newTestApplication(t *testing.T) *application {

	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()

	// Create a session manager instance that uses the same settings as production, except
	// that we *don't* set a Store for the session manager. If no store is set, the SCS
	// package will default to using a transient in-memory store, which is ideal for testing purposes.
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	// Create a new instance of our application struct.
	return &application{
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

type testServer struct {
	*httptest.Server
}

// Initalizes and returns a new instance of our custom testServer type.
func newTestServer(t *testing.T, h http.Handler) *testServer {

	// Create a new test server, passing in the value returned by our app.routes() method as
	// the handler for the server. This starts up a HTTPS server which listens on a
	// randomly-chosen port of your local machine for the duration of the test.
	// If we're testing a HTTP (not HTTPS) server then we should use the httptest.NewServer function
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client. Any response cookies will
	// now be stored and sent with subsequent requests when using this client.
	ts.Client().Jar = jar

	// Disable redirect-following for the test server client by setting a custom
	// CheckRedirect function. This function will be called whenever a 3xx response is
	// received by the client, and by always returning a http.ErrUseLastResponse error it
	// forces the client to immediately return the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

// Makes a GET request to a given url path using the test server client, and returns the
// response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {

	// The network address that the test server is listening on is contained in the ts.URL
	// field. We can use this along with the ts.Client().Get() method to make a GET /ping
	// request against the test server. This returns a http.Response struct containing the response.
	result, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)
	return result.StatusCode, result.Header, string(body)
}

var csrfTokenRegex = regexp.MustCompile(`<input type="hidden" name="csrf_token" value="(.+)">`)

func extractCsrfToken(t *testing.T, htmlBody string) string {

	// FindStringSubmatch returns an array with the entire matched pattern in the first
	// position, and the values of any captured data in the subsequent positions.
	matches := csrfTokenRegex.FindStringSubmatch(htmlBody)
	if len(matches) < 2 {
		t.Fatal("No CSRF token found in given HTML body")
	}

	// Go’s html/template package automatically escapes all dynamically rendered data,
	// including our CSRF token. Because the CSRF token is a base64 encoded string it will
	// potentially include the + character, and this will be escaped to &#43;. So after
	// extracting the token from the HTML we need to run it through html.UnescapeString() to
	// get the original token value.
	return html.UnescapeString(matches[1])
}

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	body = bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}
