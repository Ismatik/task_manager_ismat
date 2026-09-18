package store_test

import (
	"errors"
	"fmt"
	"testing"

	"nexus/internal/store"
)

func TestErrNotFoundSurvivesWrapping(t *testing.T) {
	wrapped := fmt.Errorf("load node %q: %w", "abc", store.ErrNotFound)

	if !errors.Is(wrapped, store.ErrNotFound) {
		t.Errorf("errors.Is(%v, ErrNotFound) = false, want true", wrapped)
	}
	if errors.Is(errors.New("store: not found"), store.ErrNotFound) {
		t.Error("a distinct error with the same message matched ErrNotFound; it must be compared by identity")
	}
	if got, want := store.ErrNotFound.Error(), "store: not found"; got != want {
		t.Errorf("ErrNotFound.Error() = %q, want %q", got, want)
	}
}
