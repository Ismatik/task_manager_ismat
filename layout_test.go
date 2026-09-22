package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"nexus/internal/domain"
)

// Nexus — the DRIFT test for the window's minimum size (S3-03, K9, D21).
//
// # The criterion is drift, not correctness
//
// Nobody can check "624 is the right floor" from inside Go; there is no layout
// engine here and no display on this machine. What CAN be checked, and what D21
// actually asks for, is that the constants in layout.go still say the same thing
// as the files that own those numbers. Change `min-w-36` to `min-w-32` in
// Column.tsx, or `html { font-size: 13px }` in style.css, and this test goes red
// until layout.go changes with it.
//
// That is the containment for the hardest rule in this repository: a minimum
// width in Go IS the CSS floor written down a second time, and the only thing
// that makes a second copy survivable is a test that fails when the two disagree.
//
// # It reads frontend sources; it does not touch them
//
// Reading a file is not modifying it. S3-03's scope says no file under frontend/
// is modified, and none is — the paths below are opened read-only. Doing this at
// TEST time rather than at run time is deliberate: see layout.go's header for
// why the binary must not read frontend/src.

// classLine finds the ONE line of a source file containing `marker` and returns
// its Tailwind class tokens.
//
// Line-based rather than a TSX parser, and that is a considered trade: every
// className in this codebase is written on a single line, and a parser for JSX
// template literals would be a large thing to own for a lookup a regex does
// honestly. If a className is ever split across lines, this fails loudly with
// "no line contains" rather than reading the wrong element — which is the
// failure mode that matters.
func classLine(t *testing.T, path, marker string) []string {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	var found []string
	for _, line := range strings.Split(string(raw), "\n") {
		// `className` as well as the marker. S3-01 and S3-02 deliberately left
		// paragraphs behind that NAME these classes while explaining them, and a
		// test that read one of those would be asserting against prose. Every
		// class attribute in this codebase carries the word `className`; no
		// comment paragraph does.
		if strings.Contains(line, "className") && strings.Contains(line, marker) {
			found = append(found, line)
		}
	}

	if len(found) != 1 {
		t.Fatalf("%s: %d className attributes contain %q, want exactly 1", path, len(found), marker)
	}
	return regexp.MustCompile(`[a-zA-Z0-9_.\[\]()%,-]+`).FindAllString(found[0], -1)
}

// utilityUnits reads the numeric suffix of the one token starting with `prefix`.
func utilityUnits(t *testing.T, tokens []string, prefix string) float64 {
	t.Helper()

	var matched []string
	for _, token := range tokens {
		if strings.HasPrefix(token, prefix) {
			matched = append(matched, token)
		}
	}
	if len(matched) != 1 {
		t.Fatalf("%d tokens start with %q, want exactly 1: %v", len(matched), prefix, matched)
	}

	units, err := strconv.ParseFloat(strings.TrimPrefix(matched[0], prefix), 64)
	if err != nil {
		t.Fatalf("%q does not end in a number: %v", matched[0], err)
	}
	return units
}

// TestLayoutInputsStillMatchTheFrontend is THE criterion of S3-03.
//
// Each row names an input, the one file that owns it, and how to find it there.
// A failure names the input and both numbers, so the fix is obvious: change
// layout.go to agree, or change the frontend back.
func TestLayoutInputsStillMatchTheFrontend(t *testing.T) {
	const (
		app    = "frontend/src/App.tsx"
		column = "frontend/src/components/Column.tsx"
		card   = "frontend/src/components/Card.tsx"
		kanban = "frontend/src/views/Kanban.tsx"
	)

	// The markers are chosen so that they are NOT the value under test: looking
	// the column's class list up by `min-w-36` and then reading `min-w-` back out
	// of it would be a test that can never fail.
	columnClasses := classLine(t, column, "basis-0")
	cardClasses := classLine(t, card, "rounded-md")
	shellClasses := classLine(t, app, "bg-bg")
	boardClasses := classLine(t, kanban, "overflow-x-auto")

	for _, tc := range []struct {
		input string
		want  float64
		got   float64
	}{
		{"columnMinWidthUnits", columnMinWidthUnits, utilityUnits(t, columnClasses, "min-w-")},
		{"columnPadUnits", columnPadUnits, utilityUnits(t, columnClasses, "p-")},
		{"columnGapUnits", columnGapUnits, utilityUnits(t, columnClasses, "gap-")},
		{"cardPadUnits", cardPadUnits, utilityUnits(t, cardClasses, "p-")},
		{"cardGapUnits", cardGapUnits, utilityUnits(t, cardClasses, "gap-")},
		{"shellPadUnits", shellPadUnits, utilityUnits(t, shellClasses, "p-")},
		{"shellGapUnits", shellGapUnits, utilityUnits(t, shellClasses, "gap-")},
		{"boardGapUnits", boardGapUnits, utilityUnits(t, boardClasses, "gap-")},
	} {
		if tc.got != tc.want {
			t.Errorf("layout.go's %s is %v, but the frontend now says %v", tc.input, tc.want, tc.got)
		}
	}
}

// TestRootFontSizeStillMatchesTheStylesheet covers the one input that is not a
// Tailwind class, and it is the input every other number is scaled by.
func TestRootFontSizeStillMatchesTheStylesheet(t *testing.T) {
	const path = "frontend/src/style.css"

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}

	// Comments are stripped with background.go's own helper rather than a second
	// one: a commented-out declaration must not be read as a live one, and that
	// rule already has a home in this package.
	css := stripCSSComments(string(raw))

	match := regexp.MustCompile(`html\s*\{[^}]*font-size:\s*([0-9.]+)px`).FindStringSubmatch(css)
	if match == nil {
		t.Fatalf("%s: found no `html { font-size: …px }` rule at all", path)
	}

	size, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		t.Fatalf("%s: font-size %q is not a number: %v", path, match[1], err)
	}
	if size != rootFontSizePx {
		t.Errorf("layout.go's rootFontSizePx is %v, but %s now says %v", rootFontSizePx, path, size)
	}
}

// TestMinWidthFollowsTheStatusSet is the other half of "never typed": the column
// count is domain.Statuses()'s answer and nothing else's.
func TestMinWidthFollowsTheStatusSet(t *testing.T) {
	statuses := len(domain.Statuses())

	if got, want := minWindowWidth(), floorPx(minWidthTerms(statuses)); got != want {
		t.Errorf("minWindowWidth() = %d, want the floor for %d columns, %d", got, statuses, want)
	}

	// A sixth status must move the floor, with no edit to layout.go and none to
	// main.go. Asserted by asking for one more column rather than by mutating a
	// package-level set, which is not a thing a test may do to domain.
	//
	// Compared on the UNROUNDED sums: at a 13px root a column is 117px and a gap
	// is 6.5px, so the two ceilings differ by 124 while the real growth is
	// 123.5, and an assertion over the rounded numbers would be asserting the
	// rounding.
	grew := sumPx(minWidthTerms(statuses+1)) - sumPx(minWidthTerms(statuses))
	wantGrowth := unitsPx(columnMinWidthUnits) + unitsPx(boardGapUnits)

	if grew != wantGrowth {
		t.Errorf("a sixth column moves the floor by %vpx, want %vpx (one column plus one gap)", grew, wantGrowth)
	}
	if floorPx(minWidthTerms(statuses+1)) <= minWindowWidth() {
		t.Error("the window's own minimum did not move with the column count")
	}
}

// TestMinHeightDisclosesItsEstimates. D21: a term that cannot be derived is
// disclosed, not rounded up quietly. This is what stops a later edit quietly
// promoting an estimate to a fact, or padding the total to hide one.
func TestMinHeightDisclosesItsEstimates(t *testing.T) {
	terms := minHeightTerms()

	var estimated int
	total := 0.0
	for _, term := range terms {
		if term.Name == "" {
			t.Errorf("an unnamed term of %vpx: the composition has to be readable", term.Px)
		}
		if term.Px <= 0 {
			t.Errorf("term %q contributes %vpx, which is not a floor", term.Name, term.Px)
		}
		if term.Estimated {
			estimated++
		}
		total += term.Px
	}

	if estimated == 0 {
		t.Error("no term of the vertical floor is marked Estimated, which cannot be true: " +
			"a card's height depends on chip count and title wrapping, and neither is computable here")
	}
	if got := minWindowHeight(); float64(got) < total {
		t.Errorf("minWindowHeight() = %d, below the sum of its own terms %v", got, total)
	}
}

// TestLayoutArithmeticIsReportable prints the two floors term by term.
//
// Not an assertion — a record. The commit body has to show the arithmetic, and
// re-deriving it by hand is how a commit body comes to disagree with the code.
func TestLayoutArithmeticIsReportable(t *testing.T) {
	report := func(what string, terms []layoutTerm, total int) {
		var b strings.Builder
		fmt.Fprintf(&b, "\n%s = %dpx\n", what, total)
		for _, term := range terms {
			mark := "derived "
			if term.Estimated {
				mark = "ESTIMATE"
			}
			fmt.Fprintf(&b, "  %s  %8.2fpx  %s\n", mark, term.Px, term.Name)
		}
		t.Log(b.String())
	}

	report("minWindowWidth", minWidthTerms(len(domain.Statuses())), minWindowWidth())
	report("minWindowHeight", minHeightTerms(), minWindowHeight())
}
