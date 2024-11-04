package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DbPool *pgxpool.Pool
}

// Insert a new snippet into the database.
func (m *SnippetModel) Insert(title, content string) (int, error) {
	stmt := "INSERT INTO snippets(title, content, created, expires) VALUES($1, $2, NOW(), NOW() + INTERVAL '7 days') returning id"

	var id int
	err := m.DbPool.QueryRow(context.Background(), stmt, title, content).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Return a specific snippet based on its id
func (m *SnippetModel) Get(id int) (Snippet, error) {
	stmt := "SELECT id, title, content, created, expires from snippets WHERE expires > NOW() AND id = $1"

	var s Snippet

	row := m.DbPool.QueryRow(context.Background(), stmt, id)

	// Use row.Scan() to copy the values from each field in sql.Row to the corresponding field
	// in the Snippet struct. The arguments to row.Scan are *pointers* to the place you want to
	// copy the data into, and the number of arguments must be exactly the same as the number
	// of columns returned by your statement.
	// Behind the scenes of rows.Scan() your driver will automatically convert the raw output
	// from the SQL database to the required native Go types
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}

	return s, nil
}

// This will return the 10 most recently created snippets.
func (m *SnippetModel) Latest() ([]Snippet, error) {
	stmt := `SELECT id, title, content, created, expires FROM snippets 
	WHERE expires > NOW() 
	ORDER BY id DESC LIMIT 10`

	rows, err := m.DbPool.Query(context.Background(), stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var snippets []Snippet

	// Use rows.Next to iterate through the rows in the resultset. If iteration over all the rows
	// completes then the resultset automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}

		snippets = append(snippets, s)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any error that was
	// encountered during the iteration. It's important to call this! We shouldn't assume that a
	// successful iteration was completed over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return snippets, nil
}
