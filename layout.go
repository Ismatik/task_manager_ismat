package main

import (
	"math"

	"nexus/internal/domain"
)

// Nexus — the window's minimum size, DERIVED from the layout rather than typed
// (K9, closed by D21).
//
// # The failure this closes
//
// main.go asked for 1024x768 and no minimum, and Wails calls SetMinSize(0, 0)
// unconditionally (window.go:134 -> window.c:266), so GTK was hinted with
// min_width = min_height = 0 and the window could be dragged roughly 380px below
// the layout's own floor. The user changes display scaling often and it varies,
// so the startup window is not reliably 1024 CSS px and may begin near the floor
// already.
//
// # The rule this collides with is this project's hardest one
//
// A minimum width in Go is the CSS floor written down a second time, and "a rule
// spelled twice is two rules" has cost this project four review rounds. The
// containment is D12's, applied again: D12 forbade a hex literal in main.go and
// made the colour be DERIVED from design/tokens.css, which Go already parses.
// The floor gets the same treatment — every input below names the one file that
// owns it, and layout_test.go reads that file and fails if the two have drifted.
//
// # Why the frontend is parsed at TEST time and not at run time
//
// Two reasons, and both are about what ends up in the binary. Reading
// frontend/src at run time would mean shipping the frontend SOURCES inside the
// embedded assets — today only frontend/dist is embedded — and it would put a
// file read on the startup path for a number that cannot change while the
// process is alive. The constants below are therefore plain constants, and the
// test is what stops them going stale. The criterion is DRIFT, not correctness:
// changing `min-w-36` to `min-w-32` must turn a named Go test red.
//
// # What is derived and what is estimated
//
// The horizontal floor is derived end to end: it is five column minimums, four
// gaps and the shell's padding, and every one of those is a number written in a
// file this package reads.
//
// The vertical floor is SOFTER, and D21 requires that to be said rather than
// rounded up quietly. The board's minimum height is one column heading plus one
// card, and a card's height depends on how many chips it carries and how its
// title wraps — which is layout, and layout is exactly what cannot be computed
// here. Those terms are marked Estimated and are reported as estimates by
// layoutArithmetic(); they are stated in units of TEXT LINES, so the assumption
// is visible instead of being buried in a pixel count.

const (
	// rootFontSizePx is `html { font-size: … }` in frontend/src/style.css.
	//
	// It is the reason every arithmetic here starts somewhere unexpected: at
	// 13px, every Tailwind rem is 13/16 of nominal, so `min-w-36` is 117px and
	// not 144px and `gap-2` is 6.5px and not 8px. K8 is what that costs when it
	// is forgotten.
	rootFontSizePx = 13.0

	// tailwindUnitRem is Tailwind's spacing scale: one unit is 0.25rem.
	//
	// Not ours, and not in any file of this repository — it is the framework's
	// own default, which design/tailwind.config.js does not override. Named here
	// so the conversion from "the number in the class name" to pixels is written
	// once.
	tailwindUnitRem = 0.25

	// tailwindBorderPx is the width of Tailwind's `border` utility.
	//
	// Framework default, like tailwindUnitRem. The column and the card each draw
	// one, top and bottom, and at this scale two of them are a visible term.
	tailwindBorderPx = 1.0

	// textLineRatio is how tall one line of text is, as a multiple of the font
	// size. ESTIMATE, and the largest single assumption in the vertical floor:
	// the real value is the font's own metrics resolved by the WebView, which
	// nothing in Go can ask for. 1.5 is Tailwind's normal leading.
	textLineRatio = 1.5
)

// The unit counts, each read from the ONE file that owns it. Every one of these
// is checked against that file by layout_test.go; none may be edited alone.
const (
	// frontend/src/components/Column.tsx — `min-w-36` on the column surface.
	columnMinWidthUnits = 36
	// frontend/src/components/Column.tsx — `p-2` and `gap-2` on the same surface.
	columnPadUnits = 2
	columnGapUnits = 2

	// frontend/src/views/Kanban.tsx — `gap-2` between the columns.
	boardGapUnits = 2

	// frontend/src/App.tsx — `p-2` and `gap-2` on the shell.
	shellPadUnits = 2
	shellGapUnits = 2

	// frontend/src/components/Card.tsx — `p-2` and `gap-2` on the card.
	cardPadUnits = 2
	cardGapUnits = 2
)

// The counts that are facts about the SHAPE of the markup rather than numbers
// written in it. Each is a claim about App.tsx's DOM order or a component's
// rows, and each is an ESTIMATE in the sense that it assumes nothing wraps.
const (
	// App.tsx puts three regions in the flow — header, habits strip, board — so
	// there are two gaps between them. The overlay and toast layers are `fixed`
	// and take no space in it.
	shellRegionGaps = 2

	// The header and the habits strip are each one row of controls with `py-1`.
	headerRows     = 1
	headerPadUnits = 1
	stripRows      = 1
	stripPadUnits  = 1

	// A column at its floor shows its heading and one card.
	columnHeadingLines = 1

	// A card at its floor is a title row, a chip row and a progress row, so
	// three lines and two gaps.
	cardRows = 3
	cardGaps = 2
)

// unitsPx converts a Tailwind spacing-scale count to pixels at the root font
// size. This is the one place `units -> rem -> px` is written down.
func unitsPx(units float64) float64 {
	return units * tailwindUnitRem * rootFontSizePx
}

// linesPx converts a count of text lines to pixels. ESTIMATE by construction —
// see textLineRatio.
func linesPx(lines float64) float64 {
	return lines * textLineRatio * rootFontSizePx
}

// layoutTerm is one addend of a floor, named so the arithmetic can be read back
// rather than taken on trust.
type layoutTerm struct {
	Name string
	Px   float64
	// Estimated marks a term that could not be derived from a file. D21: a term
	// that cannot be derived is disclosed, not rounded up quietly.
	Estimated bool
}

// minWidthTerms is the horizontal floor, term by term, for a board of `columns`
// columns.
//
// `columns` is a parameter and not a constant on purpose: the column count is
// domain.Statuses()'s answer, the same set internal/service/read.go iterates to
// build the board, and a `5` typed into this file — or worse into main.go —
// would be that set spelled a second time. Taking it as an argument is also what
// lets a test watch the floor move when the set grows.
func minWidthTerms(columns int) []layoutTerm {
	return []layoutTerm{
		{
			Name: "columns at their floor",
			Px:   float64(columns) * unitsPx(columnMinWidthUnits),
		},
		{
			Name: "gaps between the columns",
			Px:   float64(columns-1) * unitsPx(boardGapUnits),
		},
		{
			Name: "the shell's padding, left and right",
			Px:   2 * unitsPx(shellPadUnits),
		},
	}
}

// minHeightTerms is the vertical floor, term by term.
//
// Four of its seven terms are estimates, and they are the four that depend on
// how text lays out. Nothing here rounds up to hide that.
func minHeightTerms() []layoutTerm {
	return []layoutTerm{
		{
			Name: "the shell's padding, top and bottom",
			Px:   2 * unitsPx(shellPadUnits),
		},
		{
			Name: "the gaps between the shell's regions",
			Px:   shellRegionGaps * unitsPx(shellGapUnits),
		},
		{
			Name:      "the header, one row of controls",
			Px:        linesPx(headerRows) + 2*unitsPx(headerPadUnits),
			Estimated: true,
		},
		{
			Name:      "the habits strip, one row of chips",
			Px:        linesPx(stripRows) + 2*unitsPx(stripPadUnits) + 2*tailwindBorderPx,
			Estimated: true,
		},
		{
			Name: "the column's border, padding and one inner gap",
			Px:   2*tailwindBorderPx + 2*unitsPx(columnPadUnits) + unitsPx(columnGapUnits),
		},
		{
			Name:      "the column's heading",
			Px:        linesPx(columnHeadingLines),
			Estimated: true,
		},
		{
			Name:      "one card",
			Px:        2*tailwindBorderPx + 2*unitsPx(cardPadUnits) + cardGaps*unitsPx(cardGapUnits) + linesPx(cardRows),
			Estimated: true,
		},
	}
}

// sumPx adds the terms up, in pixels and without rounding.
func sumPx(terms []layoutTerm) float64 {
	total := 0.0
	for _, term := range terms {
		total += term.Px
	}
	return total
}

// floorPx sums the terms and rounds UP, so the computed minimum is never below
// the layout's own. At a 13px root the terms are rarely whole pixels — `gap-2`
// is 6.5 — so rounding down would hand GTK a hint one pixel tighter than the
// layout actually needs.
func floorPx(terms []layoutTerm) int {
	return int(math.Ceil(sumPx(terms)))
}

// minWindowWidth is the narrowest window in which the board still fits at its
// floor, in CSS pixels.
//
// The column count comes from domain.Statuses() — the set internal/service
// already iterates to build the board — so a sixth status moves this number with
// no edit here and none in main.go.
func minWindowWidth() int {
	return floorPx(minWidthTerms(len(domain.Statuses())))
}

// minWindowHeight is the shortest window in which the header, the habits strip
// and one card in one column still fit, in CSS pixels.
//
// Softer than minWindowWidth, and deliberately not padded to hide it: see
// minHeightTerms for which of its terms are estimates.
func minWindowHeight() int {
	return floorPx(minHeightTerms())
}
