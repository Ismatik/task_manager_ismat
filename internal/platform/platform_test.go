package platform_test

import (
	"errors"
	"fmt"
	"testing"

	"nexus/internal/platform"
)

func TestAppName(t *testing.T) {
	if got, want := platform.AppName, "nexus"; got != want {
		t.Errorf("AppName = %q, want %q; the XDG paths and the .desktop entry depend on it", got, want)
	}
}

func TestErrUnsupportedSurvivesWrapping(t *testing.T) {
	wrapped := fmt.Errorf("start tray: %w", platform.ErrUnsupported)

	if !errors.Is(wrapped, platform.ErrUnsupported) {
		t.Errorf("errors.Is(%v, ErrUnsupported) = false, want true", wrapped)
	}
	if got, want := platform.ErrUnsupported.Error(), "platform: not supported on this system"; got != want {
		t.Errorf("ErrUnsupported.Error() = %q, want %q", got, want)
	}
}
