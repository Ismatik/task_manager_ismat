package domain

import (
	"strings"
	"time"
)

// TimeEntry is one stretch of tracked time on a node. An entry with no EndedAt
// is still running; at most one such entry exists across the whole database, an
// invariant the 0002 migration enforces with a partial unique index and the
// timer service enforces again above it.
// The JSON names are the wire contract (S2-02): explicit lowerCamelCase tags,
// chosen here rather than inherited from the Go identifiers. The timestamps are
// RFC 3339 and a nil EndedAt is null, which is how the frontend tells a running
// entry from a finished one without asking a second question.
type TimeEntry struct {
	ID        string     `json:"id"`
	NodeID    string     `json:"nodeId"`
	StartedAt time.Time  `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"` // nil means the timer is still running
}

// IsOpen reports whether the entry is still running.
func (e TimeEntry) IsOpen() bool { return e.EndedAt == nil }

// Duration returns how long the entry has lasted.
//
// For a closed entry that is EndedAt - StartedAt and now is never called. For an
// open entry it is now() - StartedAt: the current time arrives as the injected
// clock, never from time.Now, so "the timer says 25 minutes" is a value a test
// can assert on rather than a value that changes while the assertion runs.
//
// A negative span — a clock that went backwards, or an entry whose EndedAt
// precedes its StartedAt — reports zero rather than a negative duration. There
// is no such thing as negative tracked time, and a negative value would
// propagate into totals and the PMP timelog as a silent subtraction.
func (e TimeEntry) Duration(now func() time.Time) time.Duration {
	end := e.EndedAt
	if end == nil {
		t := now()
		end = &t
	}
	if d := end.Sub(e.StartedAt); d > 0 {
		return d
	}
	return 0
}

// Validate reports the first field of e that holds a forbidden value.
func (e TimeEntry) Validate() error {
	const entity = "time_entry"

	if strings.TrimSpace(e.ID) == "" {
		return invalid(entity, "id", "must not be empty")
	}
	if strings.TrimSpace(e.NodeID) == "" {
		return invalid(entity, "node_id", "must not be empty")
	}
	if e.StartedAt.IsZero() {
		return invalid(entity, "started_at", "must be set")
	}
	if e.EndedAt != nil && e.EndedAt.Before(e.StartedAt) {
		return invalid(entity, "ended_at", "%s is before started_at %s",
			e.EndedAt.Format(time.RFC3339), e.StartedAt.Format(time.RFC3339))
	}
	return nil
}
