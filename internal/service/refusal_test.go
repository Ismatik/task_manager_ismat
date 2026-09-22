package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"nexus/internal/domain"
)

// Nexus — the refusal table is exhaustive, or something here is red (D25).
//
// The table in refusal.go is only worth having if adding a sentinel to it
// FORCES the two locale files to grow a label. Otherwise the next refusal ships
// as a blank toast and nobody finds out until a user reports a mystery.
//
// So these tests deliberately reach across the repository and read
// frontend/src/locales/*.json. That is the layering exception and it is on
// purpose: Go owns the codes, the locale files own the words, and this is the
// only place both halves are visible at once. internal/service does not import
// anything from the frontend and does not know how it renders them.

// localeDir is where the two translation files live, relative to this package.
const localeDir = "../../frontend/src/locales"

// refusalLabelPrefix is the dotted key each code is labelled under. The
// TypeScript side builds exactly this string; see the drift test at the bottom.
const refusalLabelPrefix = "toast.refusal.body."

func TestRefuseTagsEverySentinelWithItsCode(t *testing.T) {
	for _, r := range refusals {
		t.Run(r.code, func(t *testing.T) {
			tagged := Refuse(r.err)

			want := "[" + RefusalMarker + r.code + "]"
			if !strings.Contains(tagged.Error(), want) {
				t.Errorf("Refuse(%v) = %q, want it to carry %q", r.err, tagged, want)
			}
			// The original message survives in front of the token: it is what
			// goes to the console, and the token is an addition, not a
			// replacement.
			if !strings.HasPrefix(tagged.Error(), r.err.Error()) {
				t.Errorf("Refuse(%v) = %q, want it to start with the original message", r.err, tagged)
			}
			// And the sentinel is still findable, so tagging an error on the
			// way out cannot break a caller that inspects it.
			if !errors.Is(tagged, r.err) {
				t.Errorf("errors.Is lost the sentinel through Refuse(%v)", r.err)
			}
		})
	}
}

func TestRefuseRecognisesAWrappedSentinel(t *testing.T) {
	// How these errors really arrive: a domain sentinel with two layers of
	// context on it. Matching on the message would miss this; errors.Is does
	// not.
	wrapped := fmt.Errorf("service: move node %q: %w", "abc", fmt.Errorf("domain: %w", domain.ErrProjectNeverDoing))

	code, ok := RefusalCode(wrapped)
	if !ok || code != "project-never-doing" {
		t.Fatalf("RefusalCode(wrapped) = %q, %v; want \"project-never-doing\", true", code, ok)
	}
}

func TestRefuseLeavesAFailureExactlyAsItWas(t *testing.T) {
	// An error in no table is a FAILURE, not a refusal, and keeps today's key.
	// Returned by identity, not merely equal: nothing is wrapped, so nothing
	// downstream sees a changed message.
	boom := errors.New("store: disk I/O error")

	if got := Refuse(boom); got != boom { //nolint:errorlint // identity is the assertion
		t.Errorf("Refuse(%v) = %v, want the same error value back", boom, got)
	}
	if _, ok := RefusalCode(boom); ok {
		t.Errorf("RefusalCode(%v) claimed a code", boom)
	}
}

func TestRefuseIsSafeOnNil(t *testing.T) {
	if got := Refuse(nil); got != nil {
		t.Errorf("Refuse(nil) = %v, want nil — a bound method pipes (T, error) through without a branch", got)
	}
	if _, ok := RefusalCode(nil); ok {
		t.Error("RefusalCode(nil) claimed a code")
	}
}

func TestEveryRefusalCodeIsUniqueAndAKeySegment(t *testing.T) {
	segment := regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	seen := map[string]error{}

	for _, r := range refusals {
		if !segment.MatchString(r.code) {
			t.Errorf("code %q is not a kebab-case i18n key segment", r.code)
		}
		if first, ok := seen[r.code]; ok {
			t.Errorf("code %q is used for both %v and %v", r.code, first, r.err)
		}
		seen[r.code] = r.err
		if r.err == nil {
			t.Errorf("code %q has a nil sentinel", r.code)
		}
	}
}

// THE exhaustive-or-red test (D25).
//
// Add a sentinel to `refusals` and this fails until en.json and ru.json both
// name it. Delete a label and it fails too. Negative control is recorded in the
// S3-08 commit body.
func TestEveryRefusalCodeHasALabelInBothLanguages(t *testing.T) {
	for _, file := range []string{"en.json", "ru.json"} {
		t.Run(file, func(t *testing.T) {
			labels := refusalLabels(t, file)

			for _, r := range refusals {
				text, ok := labels[r.code]
				if !ok {
					t.Errorf("%s has no %s%s — a code Go can emit would render as a blank toast", file, refusalLabelPrefix, r.code)
					continue
				}
				if strings.TrimSpace(text) == "" {
					t.Errorf("%s%s is empty in %s", refusalLabelPrefix, r.code, file)
				}
			}

			// The other direction: a label for a code Go cannot emit is a
			// sentence nobody will ever see, and usually the leftover of a
			// rename.
			for code := range labels {
				if _, ok := codeOf(code); !ok {
					t.Errorf("%s has %s%s, but no sentinel produces that code", file, refusalLabelPrefix, code)
				}
			}
		})
	}
}

// The wire token has one spelling on each side of the boundary, and this is
// what keeps them the same one.
//
// It is a literal check on purpose: the TypeScript parser cannot import a Go
// constant, so the only alternative to this test is two constants nobody
// compares until a refusal silently renders as a generic failure.
func TestTheWireMarkerIsSpelledTheSameInTypeScript(t *testing.T) {
	const parser = "../../frontend/src/store/call.ts"

	source, err := os.ReadFile(filepath.Clean(parser))
	if err != nil {
		t.Fatalf("reading %s: %v", parser, err)
	}
	if !strings.Contains(string(source), RefusalMarker) {
		t.Errorf("%s does not contain the marker %q that service.Refuse writes", parser, RefusalMarker)
	}
	if !strings.Contains(string(source), refusalLabelPrefix) {
		t.Errorf("%s does not build keys under %q", parser, refusalLabelPrefix)
	}
}

// codeOf reports whether a code is one the table can produce.
func codeOf(code string) (error, bool) {
	for _, r := range refusals {
		if r.code == code {
			return r.err, true
		}
	}
	return nil, false
}

// refusalLabels reads toast.refusal.body out of one locale file.
func refusalLabels(t *testing.T, file string) map[string]string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(localeDir, file))
	if err != nil {
		t.Fatalf("reading %s: %v", file, err)
	}

	var tree map[string]json.RawMessage
	if err := json.Unmarshal(raw, &tree); err != nil {
		t.Fatalf("parsing %s: %v", file, err)
	}

	node := tree
	for _, key := range strings.Split(strings.TrimSuffix(refusalLabelPrefix, "."), ".") {
		child, ok := node[key]
		if !ok {
			t.Fatalf("%s has no %q section", file, key)
		}
		node = nil
		if err := json.Unmarshal(child, &node); err != nil {
			t.Fatalf("%s: %q is not an object: %v", file, key, err)
		}
	}

	labels := map[string]string{}
	for code, raw := range node {
		var text string
		if err := json.Unmarshal(raw, &text); err != nil {
			t.Fatalf("%s: %s%s is not a string: %v", file, refusalLabelPrefix, code, err)
		}
		labels[code] = text
	}
	return labels
}
