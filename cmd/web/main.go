package main

import (
	"context"
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
	snippets       *models.SnippetModel
	templateCache  map[string]*template.Template
	formDecoder    *form.Decoder
	sessionManager *scs.SessionManager
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

	app := &application{
		logger:         logger,
		snippets:       &models.SnippetModel{DbPool: dbpool},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}

	// The value returned from the flag.String() function is a pointer to the flag
	// value, not the value itself.
	logger.Info("starting server", slog.String("addr", ":4000"))

	err = http.ListenAndServe(*addr, app.routes())
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
