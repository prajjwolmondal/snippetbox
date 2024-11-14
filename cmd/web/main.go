package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"snippetbox.prajjmon.net/internal/models"
)

// This struct will hold application-wide dependencies
type application struct {
	logger         *slog.Logger
	snippets       models.SnippetModelInterface
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
	users          models.UserModelInterface
}

func main() {

	// Go also has a range of other functions including flag.Int(), flag.Bool(),
	// flag.Float64() and flag.Duration() for defining flags
	addr := flag.String("addr", ":4000", "HTTP network addresss")

	dsn := flag.String("dsn", "postgres://snippetuser:localdevpassword@localhost:5432/snippetbox", "pgx data source name")

	// Must be called after all flags are defined and before flags are accessed by the program.
	flag.Parse()

	// Custom loggers created by slog.New() are concurrency-safe. You can share a single logger and
	// use it across multiple goroutines and in your HTTP handlers without needing to worry about race conditions.
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // Include the filename & line number of the calling source code in the log entries
	}))
	// Alternatively, if we wanted the log entries as JSON objects we could do:
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dbpool, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	// Ensure that the connection pool is closed before the main() function exits.
	defer dbpool.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(dbpool)
	sessionManager.Lifetime = 12 * time.Hour

	// This means that the cookie will only be sent by a user's web browser when a HTTPS connection is being used
	sessionManager.Cookie.Secure = true

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DbPool: dbpool},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
		users:          &models.UserModel{DbPool: dbpool},
	}

	// Holds the non-default TLS settings we want the server to use. In this case the only
	// thing that we're changing is the curve preferences value, so that only elliptic curves
	// with assembly implementations are used.
	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	server := &http.Server{
		Addr:         *addr,
		Handler:      app.routes(),
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelWarn), // Logging server errors at Warn level
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,      // all keep-alive connections will be automatically closed after 1 min of inactivity
		ReadTimeout:  5 * time.Second,  // if the request headers or body are still being read 5 seconds after the request is first accepted, then Go will close the underlying connection. This helps to mitigate the risk from slow-client attacks such as Slowloris
		WriteTimeout: 10 * time.Second, // will close the underlying connection if our server attempts to write to the connection after the given period.  If using HTTPS it’s sensible to set WriteTimeout to a value greater than ReadTimeout.
	}

	// The value returned from the flag.String() function is a pointer to the flag
	// value, not the value itself.
	logger.Info("starting server", slog.String("addr", server.Addr))

	err = server.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	logger.Error(err.Error())
	os.Exit(1)
}

func openDB(dsn string) (*pgxpool.Pool, error) {
	dbpool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	err = dbpool.Ping(context.Background())
	if err != nil {
		dbpool.Close()
		return nil, err
	}

	return dbpool, nil
}
