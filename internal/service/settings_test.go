package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

// settingsFixture returns a settings service over a freshly migrated database
// with the four defaults seeded, plus the repository underneath it — the tests
// need the repository to write a value the service would never write.
func settingsFixture(t *testing.T) (*service.SettingsService, *store.SettingsRepo) {
	t.Helper()

	f := newFixture(t)
	repo := store.NewSettingsRepo(f.db)
	if _, err := repo.SeedDefaults(context.Background()); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	return service.NewSettingsService(repo), repo
}

// The seeded state is D6's: aurora, dark, no accent override, English.
func TestSettingsReturnsTheSeededDefaults(t *testing.T) {
	ctx := context.Background()
	svc, _ := settingsFixture(t)

	got, err := svc.Settings(ctx)
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	want := service.SettingsView{
		Palette:  domain.PaletteAurora,
		Theme:    domain.ThemeDark,
		Accent:   "",
		Language: domain.LanguageEN,
	}
	if got != want {
		t.Errorf("Settings = %+v, want %+v", got, want)
	}
}

// Every setter round-trips through a real SQLite file and hands back the whole
// view, so the caller renders from the answer rather than from what it wrote.
func TestSettingsSettersRoundTrip(t *testing.T) {
	ctx := context.Background()

	for name, tc := range map[string]struct {
		set  func(*service.SettingsService, context.Context, string) (service.SettingsView, error)
		to   string
		want service.SettingsView
	}{
		"palette": {
			set: (*service.SettingsService).SetPalette,
			to:  domain.PaletteStudio.String(),
			want: service.SettingsView{
				Palette: domain.PaletteStudio, Theme: domain.ThemeDark, Language: domain.LanguageEN,
			},
		},
		"theme": {
			set: (*service.SettingsService).SetTheme,
			to:  domain.ThemeLight.String(),
			want: service.SettingsView{
				Palette: domain.PaletteAurora, Theme: domain.ThemeLight, Language: domain.LanguageEN,
			},
		},
		"accent": {
			set: (*service.SettingsService).SetAccent,
			to:  "#38bdf8",
			want: service.SettingsView{
				Palette: domain.PaletteAurora, Theme: domain.ThemeDark,
				Accent: "#38bdf8", Language: domain.LanguageEN,
			},
		},
		"language": {
			set: (*service.SettingsService).SetLanguage,
			to:  domain.LanguageRU.String(),
			want: service.SettingsView{
				Palette: domain.PaletteAurora, Theme: domain.ThemeDark, Language: domain.LanguageRU,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			svc, _ := settingsFixture(t)

			got, err := tc.set(svc, ctx, tc.to)
			if err != nil {
				t.Fatalf("set %s = %v", name, err)
			}
			if got != tc.want {
				t.Errorf("the setter returned %+v, want %+v", got, tc.want)
			}

			// And it is really persisted, not merely returned.
			reread, err := svc.Settings(ctx)
			if err != nil {
				t.Fatalf("Settings: %v", err)
			}
			if reread != tc.want {
				t.Errorf("re-read %+v, want %+v", reread, tc.want)
			}
		})
	}
}

// Every invalid value is refused, and the refusal names the setting so a toast
// can say which control was wrong.
func TestSettingsRefusesInvalidValues(t *testing.T) {
	refusals := map[string]struct {
		set   func(*service.SettingsService, context.Context, string) (service.SettingsView, error)
		value string
		key   string
	}{
		"a theme nobody defined":      {(*service.SettingsService).SetTheme, "purple", store.KeyTheme},
		"an empty palette":            {(*service.SettingsService).SetPalette, "", store.KeyPalette},
		"a language we do not ship":   {(*service.SettingsService).SetLanguage, "de", store.KeyLanguage},
		"an accent that is not a hex": {(*service.SettingsService).SetAccent, "not a colour", store.KeyAccent},
		"an accent with no hash":      {(*service.SettingsService).SetAccent, "38bdf8", store.KeyAccent},
		"a three-digit accent":        {(*service.SettingsService).SetAccent, "#fff", store.KeyAccent},
		"a palette in the wrong case": {(*service.SettingsService).SetPalette, "Aurora", store.KeyPalette},
	}

	for name, tc := range refusals {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			svc, _ := settingsFixture(t)
			before, err := svc.Settings(ctx)
			if err != nil {
				t.Fatalf("Settings: %v", err)
			}

			view, err := tc.set(svc, ctx, tc.value)
			if !errors.Is(err, service.ErrInvalidSetting) {
				t.Fatalf("set(%q) = %v, want service.ErrInvalidSetting", tc.value, err)
			}
			if !strings.Contains(err.Error(), tc.key) {
				t.Errorf("the refusal %q does not name the setting %q", err, tc.key)
			}
			if view != (service.SettingsView{}) {
				t.Errorf("a refused setter returned %+v, want the zero view", view)
			}

			after, err := svc.Settings(ctx)
			if err != nil {
				t.Fatalf("Settings: %v", err)
			}
			if after != before {
				t.Errorf("a refused setter changed the settings: %+v, was %+v", after, before)
			}
		})
	}
}

// The empty accent is ACCEPTED: it is the seeded default and it means "use the
// palette's own --accent" (D6). It is also how the user clears an override.
func TestSetAccentAcceptsTheEmptyOverride(t *testing.T) {
	ctx := context.Background()
	svc, _ := settingsFixture(t)

	if _, err := svc.SetAccent(ctx, "#c96f3b"); err != nil {
		t.Fatalf("SetAccent: %v", err)
	}
	got, err := svc.SetAccent(ctx, "")
	if err != nil {
		t.Fatalf("SetAccent(\"\") = %v, want it accepted", err)
	}
	if got.Accent != "" {
		t.Errorf("Accent = %q, want it cleared", got.Accent)
	}
}

// A row hand-corrupted to something nobody recognises reads back as the
// default, with NO error. The window painted from these values (D12) has to
// open even when the database has been edited by hand.
func TestSettingsFallsBackToTheDefaultForACorruptRow(t *testing.T) {
	ctx := context.Background()

	corrupt := map[string]string{
		store.KeyPalette:  "neon",
		store.KeyTheme:    "sepia",
		store.KeyAccent:   "rebeccapurple",
		store.KeyLanguage: "kk",
	}

	for key, bad := range corrupt {
		t.Run(key, func(t *testing.T) {
			svc, repo := settingsFixture(t)
			if err := repo.Set(ctx, key, bad); err != nil {
				t.Fatalf("Set: %v", err)
			}

			got, err := svc.Settings(ctx)
			if err != nil {
				t.Fatalf("Settings = %v, want no error: a corrupt row must not stop the app", err)
			}
			want := service.SettingsView{
				Palette:  domain.DefaultPalette,
				Theme:    domain.DefaultTheme,
				Accent:   domain.DefaultAccent,
				Language: domain.DefaultLanguage,
			}
			if got != want {
				t.Errorf("Settings = %+v, want every value at its default %+v", got, want)
			}
		})
	}
}

// A table with no rows at all — a database that has not been seeded — reads as
// the defaults too, for the same reason.
func TestSettingsOnAnUnseededTable(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	svc := service.NewSettingsService(store.NewSettingsRepo(f.db))

	got, err := svc.Settings(ctx)
	if err != nil {
		t.Fatalf("Settings = %v, want no error", err)
	}
	if got.Palette != domain.DefaultPalette || got.Theme != domain.DefaultTheme ||
		got.Language != domain.DefaultLanguage || got.Accent != domain.DefaultAccent {
		t.Errorf("Settings = %+v, want the defaults", got)
	}
}

// A read that cannot reach the database says so, rather than reporting the
// defaults as if they were the user's choices.
func TestSettingsSurfacesStoreFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	svc := service.NewSettingsService(store.NewSettingsRepo(f.db))

	if err := f.db.Close(); err != nil {
		t.Fatalf("closing the database: %v", err)
	}

	if _, err := svc.Settings(ctx); err == nil {
		t.Error("Settings succeeded against a closed database")
	}
	if _, err := svc.SetTheme(ctx, domain.ThemeLight.String()); err == nil {
		t.Error("SetTheme succeeded against a closed database")
	}
}

// The allowed sets have one definition each, and it is in the domain. Asserted
// as a property rather than by eye: the service must agree with domain.Palettes,
// domain.Themes and domain.Languages about exactly which values it accepts, so
// a value added to one and not the other cannot pass.
func TestTheAllowedSetsComeFromTheDomain(t *testing.T) {
	ctx := context.Background()

	t.Run("every palette the domain declares is accepted", func(t *testing.T) {
		svc, _ := settingsFixture(t)
		for _, p := range domain.Palettes() {
			if _, err := svc.SetPalette(ctx, p.String()); err != nil {
				t.Errorf("SetPalette(%q) = %v, want it accepted", p, err)
			}
		}
	})

	t.Run("every theme the domain declares is accepted", func(t *testing.T) {
		svc, _ := settingsFixture(t)
		for _, th := range domain.Themes() {
			if _, err := svc.SetTheme(ctx, th.String()); err != nil {
				t.Errorf("SetTheme(%q) = %v, want it accepted", th, err)
			}
		}
	})

	t.Run("every language the domain declares is accepted", func(t *testing.T) {
		svc, _ := settingsFixture(t)
		for _, l := range domain.Languages() {
			if _, err := svc.SetLanguage(ctx, l.String()); err != nil {
				t.Errorf("SetLanguage(%q) = %v, want it accepted", l, err)
			}
		}
	})

	t.Run("the seeded defaults are values the service accepts", func(t *testing.T) {
		svc, _ := settingsFixture(t)
		seeded, err := svc.Settings(ctx)
		if err != nil {
			t.Fatalf("Settings: %v", err)
		}
		if _, err := svc.SetPalette(ctx, seeded.Palette.String()); err != nil {
			t.Errorf("the seeded palette %q is refused by the validator: %v", seeded.Palette, err)
		}
		if _, err := svc.SetTheme(ctx, seeded.Theme.String()); err != nil {
			t.Errorf("the seeded theme %q is refused by the validator: %v", seeded.Theme, err)
		}
		if _, err := svc.SetLanguage(ctx, seeded.Language.String()); err != nil {
			t.Errorf("the seeded language %q is refused by the validator: %v", seeded.Language, err)
		}
		if _, err := svc.SetAccent(ctx, seeded.Accent); err != nil {
			t.Errorf("the seeded accent %q is refused by the validator: %v", seeded.Accent, err)
		}
	})
}
