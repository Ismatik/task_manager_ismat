package store

import "errors"

// ErrNotFound is returned by repositories when the requested row does not
// exist. Callers match it with errors.Is rather than comparing strings, so
// repositories are free to wrap it with context.
var ErrNotFound = errors.New("store: not found")
