package store

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"

	"nexus/internal/domain"
)

// timeEntryRepo returns a time-entry repository over a freshly migrated temp
// database that already has a node "n1" (and "n2") to track time against.
func timeEntryRepo(t *testing.T) (*TimeEntryRepo, *NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	nodes := NewNodeRepo(db)

	ctx := context.Background()
	for _, id := range []string{"n1", "n2"} {
		mustCreate(t, ctx, nodes, newTestNode(id))
	}
	return NewTimeEntryRepo(db), nodes, db
}

func entryIDs(entries []domain.TimeEntry) []string {
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.ID
	}
	return out
}

// The timer's whole life cycle, in the order the user lives it.
func TestTimeEntryRepoOpenCloseCycle(t *testing.T) {
	ctx := context.Background()
	r, _, _ := timeEntryRepo(t)

	started := testNow
	opened, err := r.Open(ctx, "e1", "n1", started)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !opened.IsOpen() {
		t.Error("the entry Open returned is not open")
	}

	running, err := r.OpenEntry(ctx)
	if err != nil {
		t.Fatalf("OpenEntry: %v", err)
	}
	if running.ID != "e1" || running.NodeID != "n1" {
		t.Errorf("OpenEntry = %+v, want e1 on n1", running)
	}
	if !running.StartedAt.Equal(started) {
		t.Errorf("StartedAt = %s, want %s", running.StartedAt, started)
	}
	if running.EndedAt != nil {
		t.Errorf("EndedAt = %v, want nil", running.EndedAt)
	}

	ended := started.Add(25 * time.Minute)
	if err := r.Close(ctx, "e1", ended); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := r.OpenEntry(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("OpenEntry after a close = %v, want one matching ErrNotFound", err)
	}

	closed, err := r.Get(ctx, "e1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if closed.EndedAt == nil || !closed.EndedAt.Equal(ended) {
		t.Errorf("EndedAt = %v, want %s", closed.EndedAt, ended)
	}
	if closed.IsOpen() {
		t.Error("the closed entry still reports itself as open")
	}
}

// The database half of the single-active invariant (PLAN.md §4, D2). The
// service's half — closing the running entry so the new one may start — is
// S1-19's and is deliberately absent here.
func TestTimeEntryRepoRefusesASecondOpenEntry(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name         string
		secondNodeID string
	}{
		{name: "on another node", secondNodeID: "n2"},
		{name: "on the same node", secondNodeID: "n1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, db := timeEntryRepo(t)

			if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
				t.Fatalf("first Open: %v", err)
			}

			_, err := r.Open(ctx, "e2", tc.secondNodeID, testNow.Add(time.Minute))
			if !errors.Is(err, ErrTimerRunning) {
				t.Fatalf("second Open error = %v, want one matching ErrTimerRunning", err)
			}
			if !errors.Is(err, ErrConstraint) {
				t.Errorf("the rejection does not also match ErrConstraint: %v", err)
			}
			if n := countRows(t, db, "time_entries"); n != 1 {
				t.Errorf("time_entries has %d rows, want 1 — the second entry must not exist", n)
			}

			running, err := r.OpenEntry(ctx)
			if err != nil {
				t.Fatalf("OpenEntry: %v", err)
			}
			if running.ID != "e1" {
				t.Errorf("OpenEntry = %q, want the first entry e1", running.ID)
			}
		})
	}

	t.Run("a duplicate id is a constraint failure but not a running timer", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
			t.Fatalf("Open: %v", err)
		}
		if err := r.Close(ctx, "e1", testNow.Add(time.Minute)); err != nil {
			t.Fatalf("Close: %v", err)
		}

		// Nothing is running now, so a refusal here can only be the primary key.
		_, err := r.Open(ctx, "e1", "n1", testNow.Add(2*time.Minute))
		if !errors.Is(err, ErrConstraint) {
			t.Fatalf("error = %v, want one matching ErrConstraint", err)
		}
		if errors.Is(err, ErrTimerRunning) {
			t.Errorf("a duplicate id was reported as ErrTimerRunning: %v", err)
		}
	})

	t.Run("a node that does not exist is refused by the foreign key", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		_, err := r.Open(ctx, "e1", "ghost", testNow)
		if !errors.Is(err, ErrConstraint) {
			t.Errorf("error = %v, want one matching ErrConstraint", err)
		}
	})
}

// The partial index covers open entries only: a node may accumulate any number
// of finished ones, which is the normal shape of a day's work.
func TestTimeEntryRepoAllowsManyClosedEntries(t *testing.T) {
	ctx := context.Background()
	r, _, _ := timeEntryRepo(t)

	for i := range 3 {
		id := string(rune('a' + i))
		start := testNow.Add(time.Duration(i) * time.Hour)
		if _, err := r.Open(ctx, id, "n1", start); err != nil {
			t.Fatalf("Open(%q): %v", id, err)
		}
		if err := r.Close(ctx, id, start.Add(30*time.Minute)); err != nil {
			t.Fatalf("Close(%q): %v", id, err)
		}
	}

	got, err := r.ListByNode(ctx, "n1")
	if err != nil {
		t.Fatalf("ListByNode: %v", err)
	}
	if want := []string{"c", "b", "a"}; !slices.Equal(entryIDs(got), want) {
		t.Errorf("ListByNode = %v, want %v (newest first)", entryIDs(got), want)
	}
}

func TestTimeEntryRepoClose(t *testing.T) {
	ctx := context.Background()

	t.Run("closing an entry that is already closed is a matchable error", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
			t.Fatalf("Open: %v", err)
		}
		first := testNow.Add(10 * time.Minute)
		if err := r.Close(ctx, "e1", first); err != nil {
			t.Fatalf("Close: %v", err)
		}

		err := r.Close(ctx, "e1", testNow.Add(20*time.Minute))
		if !errors.Is(err, ErrEntryClosed) {
			t.Fatalf("second Close error = %v, want one matching ErrEntryClosed", err)
		}
		if errors.Is(err, ErrNotFound) {
			t.Errorf("an already-closed entry was reported as missing: %v", err)
		}

		got, err := r.Get(ctx, "e1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.EndedAt == nil || !got.EndedAt.Equal(first) {
			t.Errorf("EndedAt = %v, want the first close's %s — a refused close must not overwrite it", got.EndedAt, first)
		}
	})

	t.Run("closing an entry that does not exist is ErrNotFound", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		err := r.Close(ctx, "ghost", testNow)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("error = %v, want one matching ErrNotFound", err)
		}
		if errors.Is(err, ErrEntryClosed) {
			t.Errorf("a missing entry was reported as already closed: %v", err)
		}
	})

	t.Run("Get on an entry that does not exist is ErrNotFound", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Get(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("error = %v, want one matching ErrNotFound", err)
		}
	})
}

func TestTimeEntryRepoCloseAll(t *testing.T) {
	ctx := context.Background()

	t.Run("closes the running entry and reports one", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
			t.Fatalf("Open: %v", err)
		}

		at := testNow.Add(time.Hour)
		closed, err := r.CloseAll(ctx, at)
		if err != nil {
			t.Fatalf("CloseAll: %v", err)
		}
		if closed != 1 {
			t.Errorf("CloseAll = %d, want 1", closed)
		}
		if _, err := r.OpenEntry(ctx); !errors.Is(err, ErrNotFound) {
			t.Errorf("OpenEntry after CloseAll = %v, want one matching ErrNotFound", err)
		}

		got, err := r.Get(ctx, "e1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.EndedAt == nil || !got.EndedAt.Equal(at) {
			t.Errorf("EndedAt = %v, want %s", got.EndedAt, at)
		}
	})

	t.Run("with nothing running it reports zero and is not an error", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		closed, err := r.CloseAll(ctx, testNow)
		if err != nil {
			t.Fatalf("CloseAll: %v", err)
		}
		if closed != 0 {
			t.Errorf("CloseAll = %d, want 0", closed)
		}
	})

	t.Run("an already-closed entry is left exactly as it was", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
			t.Fatalf("Open: %v", err)
		}
		ended := testNow.Add(15 * time.Minute)
		if err := r.Close(ctx, "e1", ended); err != nil {
			t.Fatalf("Close: %v", err)
		}

		if _, err := r.CloseAll(ctx, testNow.Add(time.Hour)); err != nil {
			t.Fatalf("CloseAll: %v", err)
		}
		got, err := r.Get(ctx, "e1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.EndedAt == nil || !got.EndedAt.Equal(ended) {
			t.Errorf("EndedAt = %v, want the untouched %s", got.EndedAt, ended)
		}
	})
}

// ListByDay is what Stage 6's PMP timelog reads, so the boundaries are the
// point of the test rather than an afterthought.
func TestTimeEntryRepoListByDay(t *testing.T) {
	ctx := context.Background()

	// Every entry is closed immediately so that the one_open_timer index does
	// not get in the way of seeding a whole day.
	seed := func(t *testing.T, r *TimeEntryRepo, entries map[string]time.Time) {
		t.Helper()
		for id, start := range entries {
			if _, err := r.Open(ctx, id, "n1", start); err != nil {
				t.Fatalf("Open(%q): %v", id, err)
			}
			if err := r.Close(ctx, id, start.Add(time.Minute)); err != nil {
				t.Fatalf("Close(%q): %v", id, err)
			}
		}
	}

	day := domain.NewDate(2026, time.September, 18)

	t.Run("midnight and the last minute of the day are both in it", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)
		seed(t, r, map[string]time.Time{
			"midnight": time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC),
			"lastmin":  time.Date(2026, 9, 18, 23, 59, 0, 0, time.UTC),
		})

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay: %v", err)
		}
		if want := []string{"midnight", "lastmin"}; !slices.Equal(entryIDs(got), want) {
			t.Errorf("ListByDay = %v, want %v", entryIDs(got), want)
		}
	})

	t.Run("the previous day's last minute is not in it", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)
		seed(t, r, map[string]time.Time{
			"yesterday": time.Date(2026, 9, 17, 23, 59, 59, 0, time.UTC),
			"today":     time.Date(2026, 9, 18, 0, 0, 0, 0, time.UTC),
		})

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay: %v", err)
		}
		if want := []string{"today"}; !slices.Equal(entryIDs(got), want) {
			t.Errorf("ListByDay = %v, want %v", entryIDs(got), want)
		}
	})

	t.Run("the next day's midnight is not in it", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)
		seed(t, r, map[string]time.Time{
			"today":    time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
			"tomorrow": time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC),
		})

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay: %v", err)
		}
		if want := []string{"today"}; !slices.Equal(entryIDs(got), want) {
			t.Errorf("ListByDay = %v, want %v", entryIDs(got), want)
		}
	})

	// The bucket is the STARTED day. An entry that runs past midnight belongs
	// wholly to the day it began on and is never split.
	t.Run("an entry that runs past midnight belongs to the day it started", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		start := time.Date(2026, 9, 18, 23, 30, 0, 0, time.UTC)
		if _, err := r.Open(ctx, "overnight", "n1", start); err != nil {
			t.Fatalf("Open: %v", err)
		}
		if err := r.Close(ctx, "overnight", start.Add(time.Hour)); err != nil {
			t.Fatalf("Close: %v", err)
		}

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay(the 18th): %v", err)
		}
		if want := []string{"overnight"}; !slices.Equal(entryIDs(got), want) {
			t.Errorf("ListByDay(18th) = %v, want %v", entryIDs(got), want)
		}

		next, err := r.ListByDay(ctx, domain.NewDate(2026, time.September, 19), time.UTC)
		if err != nil {
			t.Fatalf("ListByDay(the 19th): %v", err)
		}
		if len(next) != 0 {
			t.Errorf("ListByDay(19th) = %v, want empty — the entry is not split", entryIDs(next))
		}
	})

	// started_at is UTC and the day is the user's. The boundaries move with the
	// zone, which is the whole reason the location is a parameter.
	t.Run("the day boundaries are the caller's zone, not UTC", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)
		east := time.FixedZone("UTC+3", 3*60*60)
		seed(t, r, map[string]time.Time{
			// 2026-09-18T22:00Z is already the 19th at UTC+3.
			"late": time.Date(2026, 9, 18, 22, 0, 0, 0, time.UTC),
			// 2026-09-17T22:00Z is the 18th at UTC+3.
			"early": time.Date(2026, 9, 17, 22, 0, 0, 0, time.UTC),
		})

		inZone, err := r.ListByDay(ctx, day, east)
		if err != nil {
			t.Fatalf("ListByDay(+03:00): %v", err)
		}
		if want := []string{"early"}; !slices.Equal(entryIDs(inZone), want) {
			t.Errorf("ListByDay(18th, +03:00) = %v, want %v", entryIDs(inZone), want)
		}

		inUTC, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay(UTC): %v", err)
		}
		if want := []string{"late"}; !slices.Equal(entryIDs(inUTC), want) {
			t.Errorf("ListByDay(18th, UTC) = %v, want %v", entryIDs(inUTC), want)
		}
	})

	t.Run("a nil location means UTC rather than the machine's zone", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)
		seed(t, r, map[string]time.Time{
			"noon": time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC),
		})

		withNil, err := r.ListByDay(ctx, day, nil)
		if err != nil {
			t.Fatalf("ListByDay(nil): %v", err)
		}
		withUTC, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay(UTC): %v", err)
		}
		if !slices.Equal(entryIDs(withNil), entryIDs(withUTC)) {
			t.Errorf("ListByDay(nil) = %v, ListByDay(UTC) = %v", entryIDs(withNil), entryIDs(withUTC))
		}
	})

	t.Run("a day with nothing on it lists nothing rather than nil", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListByDay on an empty day = %v, want an empty non-nil slice", got)
		}
	})

	t.Run("the open entry is listed too", func(t *testing.T) {
		r, _, _ := timeEntryRepo(t)

		if _, err := r.Open(ctx, "running", "n1", time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)); err != nil {
			t.Fatalf("Open: %v", err)
		}

		got, err := r.ListByDay(ctx, day, time.UTC)
		if err != nil {
			t.Fatalf("ListByDay: %v", err)
		}
		if want := []string{"running"}; !slices.Equal(entryIDs(got), want) {
			t.Errorf("ListByDay = %v, want %v", entryIDs(got), want)
		}
		if !got[0].IsOpen() {
			t.Error("the running entry came back closed")
		}
	})
}

func TestTimeEntryRepoListByNode(t *testing.T) {
	ctx := context.Background()
	r, _, _ := timeEntryRepo(t)

	if _, err := r.Open(ctx, "on-n1", "n1", testNow); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := r.Close(ctx, "on-n1", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if _, err := r.Open(ctx, "on-n2", "n2", testNow.Add(time.Hour)); err != nil {
		t.Fatalf("Open: %v", err)
	}

	got, err := r.ListByNode(ctx, "n1")
	if err != nil {
		t.Fatalf("ListByNode: %v", err)
	}
	if want := []string{"on-n1"}; !slices.Equal(entryIDs(got), want) {
		t.Errorf("ListByNode(n1) = %v, want %v", entryIDs(got), want)
	}

	none, err := r.ListByNode(ctx, "ghost")
	if err != nil {
		t.Fatalf("ListByNode(ghost): %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Errorf("ListByNode(ghost) = %v, want an empty non-nil slice", none)
	}
}

// Timestamps have to survive the round trip exactly, including the UTC
// normalisation: a timer that reads 25 minutes must not read 26 after a
// restart.
func TestTimeEntryRepoRoundTripsTimestamps(t *testing.T) {
	ctx := context.Background()

	east := time.FixedZone("UTC+3", 3*60*60)

	for _, tc := range []struct {
		name            string
		started, ended  time.Time
		wantStartedText string
	}{
		{
			name:            "UTC in, UTC out",
			started:         time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC),
			ended:           time.Date(2026, 9, 18, 9, 25, 0, 0, time.UTC),
			wantStartedText: "2026-09-18T09:00:00Z",
		},
		{
			name:            "a zoned instant is normalised to UTC",
			started:         time.Date(2026, 9, 18, 12, 0, 0, 0, east),
			ended:           time.Date(2026, 9, 18, 12, 25, 0, 0, east),
			wantStartedText: "2026-09-18T09:00:00Z",
		},
		{
			name:            "midnight, the value most likely to be off by a day",
			started:         time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			ended:           time.Date(2026, 1, 1, 0, 0, 1, 0, time.UTC),
			wantStartedText: "2026-01-01T00:00:00Z",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, db := timeEntryRepo(t)

			if _, err := r.Open(ctx, "e1", "n1", tc.started); err != nil {
				t.Fatalf("Open: %v", err)
			}
			if err := r.Close(ctx, "e1", tc.ended); err != nil {
				t.Fatalf("Close: %v", err)
			}

			var raw string
			if err := db.QueryRowContext(ctx, "SELECT started_at FROM time_entries WHERE id = 'e1'").Scan(&raw); err != nil {
				t.Fatalf("reading started_at: %v", err)
			}
			if raw != tc.wantStartedText {
				t.Errorf("stored started_at = %q, want %q", raw, tc.wantStartedText)
			}

			got, err := r.Get(ctx, "e1")
			if err != nil {
				t.Fatalf("Get: %v", err)
			}
			if !got.StartedAt.Equal(tc.started) {
				t.Errorf("StartedAt = %s, want the same instant as %s", got.StartedAt, tc.started)
			}
			if got.EndedAt == nil || !got.EndedAt.Equal(tc.ended) {
				t.Errorf("EndedAt = %v, want the same instant as %s", got.EndedAt, tc.ended)
			}
			if got.StartedAt.Location() != time.UTC {
				t.Errorf("StartedAt location = %s, want UTC", got.StartedAt.Location())
			}
			if want := tc.ended.Sub(tc.started); got.Duration(nil) != want {
				t.Errorf("Duration = %s, want %s", got.Duration(nil), want)
			}
		})
	}

	t.Run("a row the repository did not write is an error, not a zero time", func(t *testing.T) {
		r, _, db := timeEntryRepo(t)

		if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
			t.Fatalf("Open: %v", err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE time_entries SET started_at = 'yesterday' WHERE id = 'e1'"); err != nil {
			t.Fatalf("corrupting started_at: %v", err)
		}
		if _, err := r.Get(ctx, "e1"); err == nil {
			t.Error("Get on an unparseable started_at returned no error")
		}
	})
}

// Deleting a node takes its time entries with it (ON DELETE CASCADE), which is
// also what frees the one_open_timer slot when the node being deleted is the
// one being tracked.
func TestTimeEntriesCascadeWithTheNode(t *testing.T) {
	ctx := context.Background()
	r, nodes, db := timeEntryRepo(t)

	if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := nodes.Delete(ctx, "n1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if n := countRows(t, db, "time_entries"); n != 0 {
		t.Errorf("time_entries has %d rows after the node was deleted, want 0", n)
	}
	if _, err := r.OpenEntry(ctx); !errors.Is(err, ErrNotFound) {
		t.Errorf("OpenEntry = %v, want one matching ErrNotFound", err)
	}
}

// S1-19 has to close the running entry and open the next one atomically; that
// is only possible because the repository runs on the caller's transaction.
func TestTimeEntryRepoRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	r, _, db := timeEntryRepo(t)

	if _, err := r.Open(ctx, "e1", "n1", testNow); err != nil {
		t.Fatalf("Open: %v", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	inTx := r.WithExecutor(tx)

	// close-then-open, the shape of "move a card to Doing".
	if err := inTx.Close(ctx, "e1", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("Close in the transaction: %v", err)
	}
	if _, err := inTx.Open(ctx, "e2", "n2", testNow.Add(time.Minute)); err != nil {
		t.Fatalf("Open in the transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	running, err := r.OpenEntry(ctx)
	if err != nil {
		t.Fatalf("OpenEntry after a rollback: %v", err)
	}
	if running.ID != "e1" {
		t.Errorf("OpenEntry = %q, want the original e1 back", running.ID)
	}
	if n := countRows(t, db, "time_entries"); n != 1 {
		t.Errorf("time_entries has %d rows after a rollback, want 1", n)
	}
}
