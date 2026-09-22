package service

import (
	"errors"
	"fmt"

	"nexus/internal/domain"
)

// Nexus — refusal codes (D25).
//
// # A refusal is not a failure
//
// Some of the errors that cross the Wails boundary are malfunctions: the disk
// is gone, the database is locked, a row is missing. The rest are RULES doing
// their job — a project may not enter doing (D9), a note has no due date, a
// timer may not run on an archived node. Until this file both arrived at the
// frontend as one key, "Nexus could not finish that", so a rule working exactly
// as specified was rendered as a malfunction and the user learned nothing.
//
// # Go classifies. Nothing else does.
//
// The table below is the ONE place an error is decided to be a refusal, and a
// stable code is the only thing that leaves. The frontend maps that code to an
// i18n key and renders the sentence in whichever language is current; it never
// reads Go's English prose, because a message is not an API. An error in no
// table is a failure and keeps today's key — the classification is a whitelist,
// so a new error is a failure until somebody decides otherwise, which is the
// safe direction.
//
// # Why the code rides in the message
//
// Wails marshals `error` into a rejected JS promise carrying `err.Error()` and
// nothing else — there is no second channel, and every bound method still
// returns (T, error). So Refuse appends a fixed machine token, produced here
// and parsed by exactly one TypeScript function (frontend/src/store/call.ts).
// The token is bracketed and prefixed so that it cannot be confused with prose,
// and the original message is left in front of it untouched: it still goes to
// the console, where the detail has always lived.

// RefusalMarker is the fixed prefix the code travels behind.
//
// It is deliberately not a word anybody would write in a sentence. The
// TypeScript parser looks for this literal; internal/service/refusal_test.go
// asserts that the two spellings have not drifted apart.
const RefusalMarker = "nexus-refusal:"

// refusal binds one sentinel to one stable code.
//
// A slice and not a map: map iteration order is random, and two sentinels could
// in principle both match a wrapped error. First match wins, deterministically.
type refusal struct {
	err  error
	code string
}

// refusals is THE table (D25). Adding a sentinel here without adding
// `toast.refusal.body.<code>` to en.json AND ru.json turns
// TestEveryRefusalCodeHasALabelInBothLanguages red, which is the whole point of
// it being a table and not twelve if-statements.
//
// The codes are kebab-case because they end up as i18n key segments. They are
// stable: renaming one is a breaking change to the frontend's labels, and the
// test above is what says so out loud.
var refusals = []refusal{
	{domain.ErrProjectNeverDoing, "project-never-doing"},
	{domain.ErrTypeHasNoColumn, "type-has-no-column"},
	{domain.ErrTypeHasNoDue, "type-has-no-due"},
	{domain.ErrTypeHasNoChildren, "type-has-no-children"},
	{domain.ErrCircularParent, "circular-parent"},
	{domain.ErrNoRecurrence, "no-recurrence"},
	{domain.ErrUnsupportedRecurrence, "unsupported-recurrence"},
	{ErrTimerNotAllowed, "timer-not-allowed"},
	{ErrNodeArchived, "node-archived"},
	{ErrNodeDone, "node-done"},
	{ErrNotAHabit, "not-a-habit"},
	{ErrInvalidSetting, "invalid-setting"},
}

// RefusalCode reports the stable code of a refusal, and whether err is one.
//
// It matches with errors.Is, so a sentinel wrapped in context all the way up
// from the domain is still recognised — which is how these errors actually
// arrive, since every layer adds its own "%w".
func RefusalCode(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	for _, r := range refusals {
		if errors.Is(err, r.err) {
			return r.code, true
		}
	}
	return "", false
}

// Refuse tags a refusal with its code, and leaves everything else alone.
//
// This is the ONE function that produces the wire token. It is applied at the
// binding layer (app.go), on the way out, so that the services and the domain
// go on returning their own sentinels to their own callers and no rule moves up
// here.
//
// It is idempotent in practice and safe on nil: wrapping nil returns nil, so a
// bound method can pipe its (T, error) straight through without a branch.
func Refuse(err error) error {
	code, ok := RefusalCode(err)
	if !ok {
		return err
	}
	// %w, not %v: errors.Is still finds the sentinel afterwards, so tagging an
	// error on the way out cannot break a caller that inspects it.
	return fmt.Errorf("%w [%s%s]", err, RefusalMarker, code)
}
