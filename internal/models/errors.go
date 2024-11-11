package models

import "errors"

var (

	// We're returning the ErrNoRecord error from our SnippetModel.Get() method, instead of
	// sql.ErrNoRows directly, to help encapsulate the model completely. This way our handlers
	// aren’t concerned with the underlying datastore or reliant on datastore-specific
	// errors (like sql.ErrNoRows) for its behavior.
	ErrNoRecord = errors.New("models: no matching record found")

	// Thrown if user tries to login with incorrect email or password
	ErrInvalidCredentials = errors.New("models: invalid credentials")

	ErrDuplicateEmail = errors.New("models: duplicate email")
)
