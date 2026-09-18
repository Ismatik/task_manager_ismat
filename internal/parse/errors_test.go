package parse_test

import (
	"errors"
	"fmt"
	"testing"

	"nexus/internal/parse"
)

func TestErrEmptyInputSurvivesWrapping(t *testing.T) {
	wrapped := fmt.Errorf("quick add %q: %w", "   ", parse.ErrEmptyInput)

	if !errors.Is(wrapped, parse.ErrEmptyInput) {
		t.Errorf("errors.Is(%v, ErrEmptyInput) = false, want true", wrapped)
	}
	if got, want := parse.ErrEmptyInput.Error(), "parse: empty input"; got != want {
		t.Errorf("ErrEmptyInput.Error() = %q, want %q", got, want)
	}
}
