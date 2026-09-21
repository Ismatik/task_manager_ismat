package main

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

// D12's first half: LC_NUMERIC is forced to C, and LC_ALL is not touched.
//
// It is tested on the extracted function rather than by reading main, so that
// "it is set before wails.Run" is a claim about code that runs in a test rather
// than about a line somebody read.
func TestForceNumericLocale(t *testing.T) {
	t.Setenv("LC_NUMERIC", "ru_RU.UTF-8")
	t.Setenv("LC_ALL", "")
	if err := os.Unsetenv("LC_ALL"); err != nil {
		t.Fatalf("unsetting LC_ALL: %v", err)
	}

	if err := forceNumericLocale(); err != nil {
		t.Fatalf("forceNumericLocale: %v", err)
	}

	if got := os.Getenv("LC_NUMERIC"); got != "C" {
		t.Errorf("LC_NUMERIC = %q, want C — GTK parses the background colour with it", got)
	}
	if got, set := os.LookupEnv("LC_ALL"); set {
		t.Errorf("LC_ALL = %q; only LC_NUMERIC may be forced (D12)", got)
	}
}

// bgPattern finds each palette block and its --bg, INDEPENDENTLY of the parser
// under test: one regular expression over the raw file, rather than the
// comment-stripping, brace-counting scan tokenBackgrounds performs.
//
// The two agreeing is the point. If design/tokens.css changes a background and
// the mapping does not follow, this test is what fails.
var bgPattern = regexp.MustCompile(`(?s)\[data-palette='([a-z]+)'\](\.dark)?[^{]*\{(.*?)\}`)

var bgValue = regexp.MustCompile(`--bg:\s*([^;]+);`)

// tokensFromFile reads the four (palette, theme) -> --bg pairs straight out of
// design/tokens.css.
func tokensFromFile(t *testing.T) map[appearance]string {
	t.Helper()

	raw, err := os.ReadFile("design/tokens.css")
	if err != nil {
		t.Fatalf("reading design/tokens.css: %v", err)
	}

	out := map[appearance]string{}
	for _, block := range bgPattern.FindAllStringSubmatch(string(raw), -1) {
		value := bgValue.FindStringSubmatch(block[3])
		if value == nil {
			continue
		}
		theme := domain.ThemeLight
		if block[2] != "" {
			theme = domain.ThemeDark
		}
		out[appearance{Palette: domain.Palette(block[1]), Theme: theme}] = value[1]
	}
	return out
}

// Each of the four (palette, theme) combinations maps to the --bg that
// design/tokens.css actually declares — read from the file, never retyped here.
func TestTokenBackgroundsMatchTheTokensFile(t *testing.T) {
	want := tokensFromFile(t)
	if len(want) != 4 {
		t.Fatalf("design/tokens.css yielded %d backgrounds, want four (two palettes x two themes): %v",
			len(want), want)
	}

	got, err := tokenBackgrounds()
	if err != nil {
		t.Fatalf("tokenBackgrounds: %v", err)
	}

	for _, palette := range domain.Palettes() {
		for _, theme := range domain.Themes() {
			key := appearance{Palette: palette, Theme: theme}
			t.Run(palette.String()+"/"+theme.String(), func(t *testing.T) {
				if got[key] != want[key] {
					t.Errorf("--bg = %q, but design/tokens.css declares %q", got[key], want[key])
				}
				if got[key] == "" {
					t.Error("no --bg at all for this combination")
				}
			})
		}
	}

	if len(got) != len(want) {
		t.Errorf("tokenBackgrounds returned %d entries, the file declares %d: %v vs %v",
			len(got), len(want), got, want)
	}
}

// The parsed value really becomes the RGBA Wails is handed: the right three
// bytes, and opaque.
func TestWindowBackgroundIsTheTokensValueOpaque(t *testing.T) {
	tokens := tokensFromFile(t)

	for _, palette := range domain.Palettes() {
		for _, theme := range domain.Themes() {
			key := appearance{Palette: palette, Theme: theme}
			t.Run(palette.String()+"/"+theme.String(), func(t *testing.T) {
				want, err := parseHexRGBA(tokens[key])
				if err != nil {
					t.Fatalf("the token %q does not parse: %v", tokens[key], err)
				}

				got := windowBackground(key)
				if got == nil {
					t.Fatal("windowBackground returned nil for a combination the tokens describe")
				}
				if *got != want {
					t.Errorf("background = %+v, want %+v (%s)", *got, want, tokens[key])
				}
				if got.A != 255 {
					t.Errorf("alpha = %d, want 255: options.RGBA is 0..255, not a fraction (S1-02)", got.A)
				}
			})
		}
	}
}

// An appearance the tokens do not describe falls back to the SEEDED default —
// aurora and dark (D6) — rather than to a constant typed here.
func TestWindowBackgroundFallsBackToTheSeededDefault(t *testing.T) {
	tokens := tokensFromFile(t)
	want, err := parseHexRGBA(tokens[appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme}])
	if err != nil {
		t.Fatalf("the default token does not parse: %v", err)
	}

	for _, unknown := range []appearance{
		{Palette: "neon", Theme: domain.ThemeDark},
		{Palette: domain.PaletteStudio, Theme: "sepia"},
		{},
	} {
		got := windowBackground(unknown)
		if got == nil {
			t.Fatalf("windowBackground(%+v) = nil, want the default background", unknown)
		}
		if *got != want {
			t.Errorf("windowBackground(%+v) = %+v, want the aurora/dark default %+v", unknown, *got, want)
		}
	}
}

// The startup read is the persisted palette and theme, and a settings read that
// fails does not stop the window: it reports the defaults.
func TestStartupAppearance(t *testing.T) {
	ctx := context.Background()
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	db, err := openStore(ctx)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	settings := service.NewSettingsService(store.NewSettingsRepo(db))

	t.Run("a fresh installation is the seeded default", func(t *testing.T) {
		want := appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme}
		if got := startupAppearance(ctx, settings); got != want {
			t.Errorf("startupAppearance = %+v, want %+v", got, want)
		}
	})

	t.Run("what the user last chose", func(t *testing.T) {
		if _, err := settings.SetPalette(ctx, domain.PaletteStudio.String()); err != nil {
			t.Fatalf("SetPalette: %v", err)
		}
		if _, err := settings.SetTheme(ctx, domain.ThemeLight.String()); err != nil {
			t.Fatalf("SetTheme: %v", err)
		}

		want := appearance{Palette: domain.PaletteStudio, Theme: domain.ThemeLight}
		if got := startupAppearance(ctx, settings); got != want {
			t.Errorf("startupAppearance = %+v, want %+v", got, want)
		}

		// And it really is a different first frame from the default one.
		light := windowBackground(want)
		dark := windowBackground(appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme})
		if light == nil || dark == nil || *light == *dark {
			t.Errorf("the light and dark first frames are the same colour: %+v", light)
		}
	})

	t.Run("a settings read that fails still opens a window", func(t *testing.T) {
		if err := db.Close(); err != nil {
			t.Fatalf("closing the database: %v", err)
		}

		want := appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme}
		if got := startupAppearance(ctx, settings); got != want {
			t.Errorf("startupAppearance = %+v, want the defaults %+v", got, want)
		}
	})
}

// The CSS scan is only as good as its handling of the file's awkward parts: a
// commented-out declaration, a nested rule inside a media query, and a block
// that belongs to no palette.
func TestCSSScanning(t *testing.T) {
	t.Run("comments are stripped", func(t *testing.T) {
		if got := stripCSSComments("a /* b */ c"); got != "a  c" {
			t.Errorf("stripCSSComments = %q", got)
		}
		if got := stripCSSComments("a /* unterminated"); got != "a " {
			t.Errorf("an unterminated comment leaves %q", got)
		}
		if got := stripCSSComments("plain"); got != "plain" {
			t.Errorf("stripCSSComments = %q", got)
		}
	})

	t.Run("a nested rule does not end the block that contains it", func(t *testing.T) {
		blocks := cssBlocks("@media (x) { * { a: 1; } }\n[data-palette='aurora'] { --bg: #ffffff; }")
		if len(blocks) != 2 {
			t.Fatalf("cssBlocks found %d top-level blocks, want 2: %+v", len(blocks), blocks)
		}
		if blocks[1].selector != "[data-palette='aurora']" {
			t.Errorf("the second selector is %q", blocks[1].selector)
		}
	})

	t.Run("a selector with no palette belongs to nothing", func(t *testing.T) {
		for _, selector := range []string{":root", "@media (prefers-reduced-motion: reduce)", "[data-palette='neon']", "[data-palette='"} {
			if _, ok := selectorPalette(selector); ok {
				t.Errorf("selectorPalette(%q) claimed a palette", selector)
			}
		}
		if got, ok := selectorPalette("[data-palette='studio'] .dark"); !ok || got != domain.PaletteStudio {
			t.Errorf("selectorPalette = %q, %v; want studio, true", got, ok)
		}
	})

	t.Run("a declaration is found by name, not by prefix", func(t *testing.T) {
		body := "--bg-soft: #111111;\n  --bg: #222222;\n  --border: #333333;"
		got, ok := declaration(body, "--bg")
		if !ok || got != "#222222" {
			t.Errorf("declaration(--bg) = %q, %v; want #222222, true", got, ok)
		}
		if _, ok := declaration(body, "--missing"); ok {
			t.Error("declaration found a property that is not there")
		}
	})
}

// The colour parser accepts what domain.IsHexColour accepts and nothing else —
// one colour syntax in Go, not two.
func TestParseHexRGBA(t *testing.T) {
	got, err := parseHexRGBA("#38bdf8")
	if err != nil {
		t.Fatalf("parseHexRGBA: %v", err)
	}
	if want := (options.RGBA{R: 0x38, G: 0xbd, B: 0xf8, A: 255}); got != want {
		t.Errorf("parseHexRGBA = %+v, want %+v", got, want)
	}

	for _, bad := range []string{"", "#fff", "38bdf8", "#gggggg", "rgba(1,2,3,1)"} {
		if _, err := parseHexRGBA(bad); err == nil {
			t.Errorf("parseHexRGBA(%q) succeeded", bad)
		}
	}
}
