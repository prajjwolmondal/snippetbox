package models

import "errors"

// We're returning the ErrNoRecord error from our SnippetModel.Get() method, instead of
// sql.ErrNoRows directly, to help encapsulate the model completely. This way our handlers
// aren’t concerned with the underlying datastore or reliant on datastore-specific
// errors (like sql.ErrNoRows) for its behavior.
var ErrNoRecord = errors.New("models: no matching record found")
