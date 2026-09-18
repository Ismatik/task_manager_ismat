package parse

import "errors"

// ErrEmptyInput is returned when there is nothing to parse: the quick-add line
// is empty or contains only whitespace. Callers match it with errors.Is, so it
// can be wrapped with context.
var ErrEmptyInput = errors.New("parse: empty input")
