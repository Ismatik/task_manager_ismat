package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/options"

	"nexus/internal/domain"
	"nexus/internal/service"
)

// tokensFS holds design/tokens.css, the source of truth for every colour in
// this application (D12, S2-08).
//
// # Why the CSS file is read by Go at all
//
// GTK paints the window background before the WebView has evaluated anything,
// so it is the first colour the user sees, and Wails wants it as an RGBA at
// startup. That makes main a second place that knows a background colour — and
// a second spelling of a value is the defect class Stage 1 spent three review
// rounds on. So the value is not retyped here: the one file that owns it is
// embedded and parsed, and if a token changes, this changes with it.
//
// design/ stays read-only. This reads it and never writes it.
//
//go:embed design/tokens.css
var tokensFS embed.FS

// tokensPath is the embedded file's name.
const tokensPath = "design/tokens.css"

// appearance is one (palette, theme) pair — the two settings that together pick
// a --bg (D6).
type appearance struct {
	Palette domain.Palette
	Theme   domain.Theme
}

// forceNumericLocale sets LC_NUMERIC=C for this process (D12, closing K1).
//
// Wails formats the window background in C at window.c:205 using the PROCESS
// locale, so under a comma-decimal locale — this machine's LC_NUMERIC is
// ru_RU.UTF-8 — it emits "rgba(27, 38, 54, 0,0)". GTK's CSS parser needs a
// decimal point there, rejects the declaration, and discards the whole thing
// SILENTLY: nothing is logged, and the background is simply never applied.
//
// Only LC_NUMERIC is forced, never LC_ALL. Go's own formatting is
// locale-independent — strconv and fmt do not consult the C locale at all — so
// the blast radius of this is exactly the cgo/GTK layer, which is the thing
// being fixed. Forcing LC_ALL would additionally change collation and time
// formatting for anything that ever asks the C library, for no benefit.
//
// It must run BEFORE wails.Run, because GTK reads the environment when it
// initialises and never again.
func forceNumericLocale() error {
	if err := os.Setenv("LC_NUMERIC", "C"); err != nil {
		return fmt.Errorf("nexus: forcing LC_NUMERIC=C: %w", err)
	}
	return nil
}

// startupAppearance reads the persisted palette and theme, falling back to the
// seeded defaults when the settings cannot be read at all.
//
// A settings row is never worth refusing to start over (S2-05 says the same
// thing about a corrupt value): the failure is logged and the window opens
// looking like a fresh installation, which is the state the user can act on.
func startupAppearance(ctx context.Context, settings *service.SettingsService) appearance {
	view, err := settings.Settings(ctx)
	if err != nil {
		log.Printf("nexus: reading the appearance settings: %v; falling back to the defaults", err)
		return appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme}
	}
	return appearance{Palette: view.Palette, Theme: view.Theme}
}

// windowBackground returns the --bg design/tokens.css declares for this
// appearance, as the RGBA Wails wants.
//
// Alpha is 255 because options.RGBA is four uint8s in 0..255, alpha included —
// S1-02 fixed an A: 1 that meant roughly 0.4% opacity rather than "opaque". That
// fix was unobservable until now, because K1 meant the colour never reached GTK
// at all; forceNumericLocale is what makes it visible, and what makes this
// function worth having.
//
// It NEVER fails the startup. An appearance the tokens do not describe falls
// back to the seeded default (D6), and a tokens file this cannot parse at all
// returns nil — which leaves Wails its own default and still opens a window. A
// colour is not worth a blank screen, and nothing here may become a place where
// a colour is typed by hand.
func windowBackground(want appearance) *options.RGBA {
	backgrounds, err := tokenBackgrounds()
	if err != nil {
		log.Printf("nexus: %v; the window background is left to Wails", err)
		return nil
	}

	hex, ok := backgrounds[want]
	if !ok {
		fallback := appearance{Palette: domain.DefaultPalette, Theme: domain.DefaultTheme}
		if hex, ok = backgrounds[fallback]; !ok {
			log.Printf("nexus: %s declares no --bg for %s/%s or for the default %s/%s",
				tokensPath, want.Palette, want.Theme, fallback.Palette, fallback.Theme)
			return nil
		}
		log.Printf("nexus: %s declares no --bg for %s/%s; using %s/%s",
			tokensPath, want.Palette, want.Theme, fallback.Palette, fallback.Theme)
	}

	rgba, err := parseHexRGBA(hex)
	if err != nil {
		log.Printf("nexus: %v; the window background is left to Wails", err)
		return nil
	}
	return &rgba
}

// tokenBackgrounds returns the --bg declared for every (palette, theme) in
// design/tokens.css.
//
// The file's shape, which is the contract this reads:
//
//	[data-palette='aurora']           { … --bg: <light>; … }
//	[data-palette='aurora'].dark,
//	[data-palette='aurora'] .dark     { … --bg: <dark>;  … }
//
// so the palette comes from the selector's data-palette attribute and the theme
// from whether that selector mentions .dark. A block with no data-palette — the
// :root block, the reduced-motion media query — has no appearance to belong to
// and is skipped, media query and all.
func tokenBackgrounds() (map[appearance]string, error) {
	raw, err := tokensFS.ReadFile(tokensPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", tokensPath, err)
	}

	out := map[appearance]string{}
	for _, block := range cssBlocks(stripCSSComments(string(raw))) {
		palette, ok := selectorPalette(block.selector)
		if !ok {
			continue
		}
		value, ok := declaration(block.body, "--bg")
		if !ok {
			continue
		}
		if !domain.IsHexColour(value) {
			// The one colour syntax in Go is domain.IsHexColour's, so this
			// refuses rather than teaching a second parser a second shape.
			return nil, fmt.Errorf("%s: --bg for %s is %q, which is not a #RRGGBB colour",
				tokensPath, block.selector, value)
		}

		theme := domain.ThemeLight
		if strings.Contains(block.selector, ".dark") {
			theme = domain.ThemeDark
		}
		out[appearance{Palette: palette, Theme: theme}] = value
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("%s declares no palette background at all", tokensPath)
	}
	return out, nil
}

// cssBlock is one top-level rule: everything before the brace, and everything
// inside it.
type cssBlock struct {
	selector string
	body     string
}

// cssBlocks splits css into its TOP-LEVEL rules, counting braces so that a
// nested rule — the `*` inside the reduced-motion media query — does not end
// the block that contains it.
func cssBlocks(css string) []cssBlock {
	var (
		out   []cssBlock
		depth int
		start int
		sel   string
	)

	for i, r := range css {
		switch r {
		case '{':
			if depth == 0 {
				sel = strings.TrimSpace(css[start:i])
				start = i + 1
			}
			depth++
		case '}':
			depth--
			if depth == 0 {
				out = append(out, cssBlock{selector: sel, body: css[start:i]})
				start = i + 1
			}
		}
	}
	return out
}

// stripCSSComments removes /* … */ so that a commented-out declaration cannot
// be read as a live one.
func stripCSSComments(css string) string {
	var b strings.Builder
	for {
		open := strings.Index(css, "/*")
		if open < 0 {
			b.WriteString(css)
			return b.String()
		}
		b.WriteString(css[:open])

		close := strings.Index(css[open:], "*/")
		if close < 0 {
			return b.String()
		}
		css = css[open+close+2:]
	}
}

// selectorPalette returns the palette a selector selects, from its
// data-palette='…' attribute.
func selectorPalette(selector string) (domain.Palette, bool) {
	const marker = "data-palette='"

	at := strings.Index(selector, marker)
	if at < 0 {
		return "", false
	}
	rest := selector[at+len(marker):]

	end := strings.IndexByte(rest, '\'')
	if end < 0 {
		return "", false
	}
	palette := domain.Palette(rest[:end])

	// A palette the app does not know is not one it can be set to, so it is not
	// one this map needs an entry for. domain.Palettes() is the single
	// definition of that set (S2-05).
	return palette, palette.Valid()
}

// declaration returns the value of a custom property in a rule body, trimmed
// and without its semicolon.
func declaration(body, property string) (string, bool) {
	for _, line := range strings.Split(body, ";") {
		name, value, found := strings.Cut(line, ":")
		if !found || strings.TrimSpace(name) != property {
			continue
		}
		return strings.TrimSpace(value), true
	}
	return "", false
}

// parseHexRGBA turns a #RRGGBB colour into an opaque options.RGBA.
//
// The syntax is domain.IsHexColour's — the ONE place a colour's shape is known
// in Go (S2-05) — so this validates through it and then reads the three bytes
// it has already vouched for.
func parseHexRGBA(hex string) (options.RGBA, error) {
	if !domain.IsHexColour(hex) {
		return options.RGBA{}, fmt.Errorf("%q is not a #RRGGBB colour", hex)
	}

	value, err := strconv.ParseUint(hex[1:], 16, 32)
	if err != nil {
		return options.RGBA{}, fmt.Errorf("%q is not a #RRGGBB colour: %w", hex, err)
	}

	return options.RGBA{
		R: uint8(value >> 16),
		G: uint8(value >> 8),
		B: uint8(value),
		// Opaque. Every field of options.RGBA is a uint8 in 0..255, alpha
		// included; it is not a 0..1 fraction (S1-02).
		A: 255,
	}, nil
}
