package domain_test

import (
	"testing"
	"time"

	"nexus/internal/domain"
)

// startedAt is 09:00 on the fixed day; fixedNow is 10:30, so an open entry
// started then has run for exactly 90 minutes.
var startedAt = time.Date(2026, time.September, 18, 9, 0, 0, 0, time.UTC)

func validTimeEntry() domain.TimeEntry {
	return domain.TimeEntry{ID: "e1", NodeID: "n1", StartedAt: startedAt}
}

func TestTimeEntryIsOpen(t *testing.T) {
	tests := []struct {
		name string
		in   domain.TimeEntry
		want bool
	}{
		{"no ended_at means running", validTimeEntry(), true},
		{"an ended_at means closed", domain.TimeEntry{
			ID: "e1", NodeID: "n1", StartedAt: startedAt, EndedAt: ptr(fixedNow),
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.IsOpen(); got != tt.want {
				t.Errorf("IsOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeEntryDuration(t *testing.T) {
	// A clock that records whether it was consulted: a closed entry must not
	// need the current time at all.
	calls := 0
	countingNow := func() time.Time {
		calls++
		return fixedNow
	}

	tests := []struct {
		name      string
		entry     domain.TimeEntry
		wantCalls int
		want      time.Duration
	}{
		{
			name:      "an open entry measures against the injected clock",
			entry:     validTimeEntry(),
			wantCalls: 1,
			want:      90 * time.Minute,
		},
		{
			name: "a closed entry measures ended_at minus started_at",
			entry: domain.TimeEntry{ID: "e1", NodeID: "n1", StartedAt: startedAt,
				EndedAt: ptr(startedAt.Add(25 * time.Minute))},
			wantCalls: 0,
			want:      25 * time.Minute,
		},
		{
			name: "a zero-length closed entry is zero, not negative",
			entry: domain.TimeEntry{ID: "e1", NodeID: "n1", StartedAt: startedAt,
				EndedAt: ptr(startedAt)},
			wantCalls: 0,
			want:      0,
		},
		{
			name: "a closed entry that ends before it starts reports zero",
			entry: domain.TimeEntry{ID: "e1", NodeID: "n1", StartedAt: startedAt,
				EndedAt: ptr(startedAt.Add(-time.Hour))},
			wantCalls: 0,
			want:      0,
		},
		{
			name: "an open entry whose clock went backwards reports zero",
			entry: domain.TimeEntry{ID: "e1", NodeID: "n1",
				StartedAt: fixedNow.Add(time.Hour)},
			wantCalls: 1,
			want:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls = 0
			if got := tt.entry.Duration(countingNow); got != tt.want {
				t.Errorf("Duration() = %v, want %v", got, tt.want)
			}
			if calls != tt.wantCalls {
				t.Errorf("the clock was read %d times, want %d", calls, tt.wantCalls)
			}
		})
	}
}

// A closed entry must not touch the clock at all — passing nil proves it.
func TestTimeEntryDurationOfClosedEntryNeverCallsTheClock(t *testing.T) {
	e := validTimeEntry()
	e.EndedAt = ptr(startedAt.Add(45 * time.Minute))
	if got := e.Duration(nil); got != 45*time.Minute {
		t.Errorf("Duration(nil) = %v, want 45m", got)
	}
}

func TestTimeEntryValidate(t *testing.T) {
	t.Run("an open entry is valid", func(t *testing.T) {
		if err := validTimeEntry().Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})
	t.Run("a closed entry is valid", func(t *testing.T) {
		e := validTimeEntry()
		e.EndedAt = ptr(fixedNow)
		if err := e.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})
	t.Run("ending exactly when it started is valid", func(t *testing.T) {
		e := validTimeEntry()
		e.EndedAt = ptr(startedAt)
		if err := e.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})

	bad := []struct {
		name  string
		field string
		mut   func(*domain.TimeEntry)
	}{
		{"empty id", "id", func(e *domain.TimeEntry) { e.ID = "" }},
		{"blank node id", "node_id", func(e *domain.TimeEntry) { e.NodeID = " " }},
		{"missing started_at", "started_at", func(e *domain.TimeEntry) { e.StartedAt = time.Time{} }},
		{"ended before started", "ended_at", func(e *domain.TimeEntry) {
			e.EndedAt = ptr(startedAt.Add(-time.Second))
		}},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			e := validTimeEntry()
			tt.mut(&e)
			assertValidationError(t, e.Validate(), "time_entry", tt.field)
		})
	}
}
