package service

import (
	"context"
	"errors"
	"fmt"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// ErrInvalidSetting reports that a setting was given a value that is not
// allowed. Every refusal below wraps it and names the setting, so a caller can
// match one sentinel and still tell the user which control was wrong.
var ErrInvalidSetting = errors.New("service: the setting value is not allowed")

// SettingsService is the four persisted preferences — palette, theme, accent
// and language (D6) — read and written as a typed whole.
//
// # Why there is a service over a two-method repository
//
// store.SettingsRepo stores strings and validates nothing: Set("theme",
// "purple") succeeds today, and the value is read back at the next startup by
// the code that paints the window (D12). Validation is a rule, so it is here,
// and the rules it applies come from internal/domain — Palette, Theme, Language
// and IsHexColour — because "which palettes exist" must have exactly ONE
// definition in this codebase. The frontend builds its pickers from what this
// service reports rather than from a literal array of its own, which would be a
// second.
//
// # A bad row never stops the app
//
// Reading falls back to the seeded default for any value that is missing or
// unrecognised, and returns no error. A settings row is not worth refusing to
// start over: the window painted from these values (D12) must open even when
// somebody has hand-edited the database.
type SettingsService struct {
	repo *store.SettingsRepo
}

// NewSettingsService returns a service over the settings repository.
func NewSettingsService(repo *store.SettingsRepo) *SettingsService {
	return &SettingsService{repo: repo}
}

// Settings returns the four preferences, each either as stored or, if the
// stored value is missing or unrecognised, as its default.
//
// One query: the whole table is four rows.
func (s *SettingsService) Settings(ctx context.Context) (SettingsView, error) {
	all, err := s.repo.All(ctx)
	if err != nil {
		return SettingsView{}, err
	}

	return SettingsView{
		Palette:  paletteOrDefault(all[store.KeyPalette]),
		Theme:    themeOrDefault(all[store.KeyTheme]),
		Accent:   accentOrDefault(all[store.KeyAccent]),
		Language: languageOrDefault(all[store.KeyLanguage]),
	}, nil
}

// SetPalette stores the palette and returns the settings as they now are.
//
// Every setter returns the whole view rather than nothing or a bare error, so
// that the frontend re-renders from the answer instead of from what it hoped it
// wrote — the two differ the moment a value is rejected, normalised or changed
// by something else.
func (s *SettingsService) SetPalette(ctx context.Context, value string) (SettingsView, error) {
	if !domain.Palette(value).Valid() {
		return SettingsView{}, refuse(store.KeyPalette, value)
	}
	return s.set(ctx, store.KeyPalette, value)
}

// SetTheme stores the theme and returns the settings as they now are.
func (s *SettingsService) SetTheme(ctx context.Context, value string) (SettingsView, error) {
	if !domain.Theme(value).Valid() {
		return SettingsView{}, refuse(store.KeyTheme, value)
	}
	return s.set(ctx, store.KeyTheme, value)
}

// SetAccent stores the accent override and returns the settings as they now
// are.
//
// The EMPTY string is accepted and is the seeded default: it means "use the
// palette's own --accent" (D6), and it is how the user clears an override. A
// non-empty value must be a colour, which is domain.IsHexColour's answer — the
// one place a colour syntax is known in Go. It validates shape only; no
// palette's values are known here or anywhere else in Go.
func (s *SettingsService) SetAccent(ctx context.Context, value string) (SettingsView, error) {
	if value != domain.DefaultAccent && !domain.IsHexColour(value) {
		return SettingsView{}, refuse(store.KeyAccent, value)
	}
	return s.set(ctx, store.KeyAccent, value)
}

// SetLanguage stores the UI language and returns the settings as they now are.
func (s *SettingsService) SetLanguage(ctx context.Context, value string) (SettingsView, error) {
	if !domain.Language(value).Valid() {
		return SettingsView{}, refuse(store.KeyLanguage, value)
	}
	return s.set(ctx, store.KeyLanguage, value)
}

// set writes one validated value and reads the whole view back.
func (s *SettingsService) set(ctx context.Context, key, value string) (SettingsView, error) {
	if err := s.repo.Set(ctx, key, value); err != nil {
		return SettingsView{}, err
	}
	return s.Settings(ctx)
}

// refuse builds the refusal for one setting, naming both the setting and the
// value so that a log or a toast can say which control was wrong.
func refuse(key, value string) error {
	return fmt.Errorf("service: %s = %q: %w", key, value, ErrInvalidSetting)
}

// paletteOrDefault, themeOrDefault, languageOrDefault and accentOrDefault are
// the read-side fallbacks: an unreadable or unknown stored value reads back as
// the default, with no error. Each asks its domain type whether the stored
// string is one of the values that exist; none of them holds a list.

func paletteOrDefault(raw string) domain.Palette {
	if p := domain.Palette(raw); p.Valid() {
		return p
	}
	return domain.DefaultPalette
}

func themeOrDefault(raw string) domain.Theme {
	if t := domain.Theme(raw); t.Valid() {
		return t
	}
	return domain.DefaultTheme
}

func languageOrDefault(raw string) domain.Language {
	if l := domain.Language(raw); l.Valid() {
		return l
	}
	return domain.DefaultLanguage
}

func accentOrDefault(raw string) string {
	if raw == domain.DefaultAccent || domain.IsHexColour(raw) {
		return raw
	}
	return domain.DefaultAccent
}
