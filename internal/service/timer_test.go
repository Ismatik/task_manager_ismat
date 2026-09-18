package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

// timers returns a timer service over the fixture's database, sharing its
// clock — so a test moves time by assigning to f.now and never sleeps.
func (f *fixture) timers() *service.TimerService {
	f.t.Helper()

	return service.NewTimerService(f.begin, f.nodes, f.entries, f.clock(), f.nextID)
}

// openEntries counts the rows with ended_at IS NULL, in SQL, across the whole
// table. The single-active invariant is a statement about that number and
// nothing else, so the test asserts exactly it.
func (f *fixture) openEntries() int {
	f.t.Helper()

	var n int
	if err := f.db.QueryRow("SELECT count(*) FROM time_entries WHERE ended_at IS NULL").Scan(&n); err != nil {
		f.t.Fatalf("counting open entries: %v", err)
	}
	return n
}

// entriesOf returns every entry on one node, newest first.
func (f *fixture) entriesOf(nodeID string) []domain.TimeEntry {
	f.t.Helper()

	got, err := f.entries.ListByNode(context.Background(), nodeID)
	if err != nil {
		f.t.Fatalf("ListByNode(%q): %v", nodeID, err)
	}
	return got
}

// THE overlapping-timer test: starting on B while A is running leaves exactly
// one open entry, on B, and A's is closed at the instant B started.
func TestTimerStartClosesTheRunningEntry(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	timers := f.timers()

	a := f.create(draft("a", domain.NodeTypeTask, nil))
	b := f.create(draft("b", domain.NodeTypeTask, nil))

	first, err := timers.Start(ctx, a.ID)
	if err != nil {
		t.Fatalf("Start(a) = %v", err)
	}

	switchedAt := testNow.Add(25 * time.Minute)
	f.now = switchedAt

	second, err := timers.Start(ctx, b.ID)
	if err != nil {
		t.Fatalf("Start(b) = %v", err)
	}

	t.Run("exactly one entry is open, across the whole table", func(t *testing.T) {
		if got := f.openEntries(); got != 1 {
			t.Fatalf("%d open entries, want exactly 1", got)
		}
	})

	t.Run("the open one is B's", func(t *testing.T) {
		open, err := f.entries.OpenEntry(ctx)
		if err != nil {
			t.Fatalf("OpenEntry: %v", err)
		}
		if open.ID != second.ID || open.NodeID != b.ID {
			t.Errorf("the open entry is %+v, want %q on %q", open, second.ID, b.ID)
		}
		if !open.StartedAt.Equal(switchedAt) {
			t.Errorf("started_at = %v, want %v", open.StartedAt, switchedAt)
		}
	})

	t.Run("A's entry is closed at the instant B started", func(t *testing.T) {
		closed, err := f.entries.Get(ctx, first.ID)
		if err != nil {
			t.Fatalf("Get(%q): %v", first.ID, err)
		}
		if closed.EndedAt == nil {
			t.Fatal("A's entry is still open")
		}
		if !closed.EndedAt.Equal(switchedAt) {
			t.Errorf("ended_at = %v, want %v", closed.EndedAt, switchedAt)
		}
		if got, want := closed.Duration(f.clock()), 25*time.Minute; got != want {
			t.Errorf("duration = %v, want %v", got, want)
		}
	})
}

// Stop then start on the same node produces two rows, not one long one.
func TestTimerStartStopStartProducesTwoEntries(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	timers := f.timers()
	n := f.create(draft("a", domain.NodeTypeTask, nil))

	if _, err := timers.Start(ctx, n.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	f.now = testNow.Add(10 * time.Minute)
	if _, err := timers.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	f.now = testNow.Add(30 * time.Minute)
	if _, err := timers.Start(ctx, n.ID); err != nil {
		t.Fatalf("Start again: %v", err)
	}

	f.now = testNow.Add(45 * time.Minute)
	if _, err := timers.Stop(ctx); err != nil {
		t.Fatalf("Stop again: %v", err)
	}

	got := f.entriesOf(n.ID)
	if len(got) != 2 {
		t.Fatalf("%d entries, want 2 — a stop and a start must not be merged", len(got))
	}
	total := time.Duration(0)
	for _, e := range got {
		total += e.Duration(f.clock())
	}
	if want := 25 * time.Minute; total != want {
		t.Errorf("total tracked = %v, want %v", total, want)
	}
}

// Starting on the node that is already running changes nothing: the same entry,
// the same started_at, no second row.
func TestTimerStartOnTheRunningNodeIsANoOp(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	timers := f.timers()
	n := f.create(draft("a", domain.NodeTypeTask, nil))

	first, err := timers.Start(ctx, n.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	f.now = testNow.Add(20 * time.Minute)
	again, err := timers.Start(ctx, n.ID)
	if err != nil {
		t.Fatalf("Start again: %v", err)
	}

	if again.ID != first.ID {
		t.Errorf("entry id = %q, want the running %q", again.ID, first.ID)
	}
	if !again.StartedAt.Equal(first.StartedAt) {
		t.Errorf("started_at = %v, want the original %v", again.StartedAt, first.StartedAt)
	}
	if got := len(f.entriesOf(n.ID)); got != 1 {
		t.Errorf("%d entries, want 1 — restarting must not fragment the log", got)
	}
	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want 1", got)
	}
}

// Who may be timed: D2, D7 and D9, decided by the domain and surfaced here.
func TestTimerStartRefusesWhatCannotBeTimed(t *testing.T) {
	ctx := context.Background()

	t.Run("a node with non-note children", func(t *testing.T) {
		f := newFixture(t)
		parent := f.create(draft("parent", domain.NodeTypeTask, nil))
		f.create(draft("child", domain.NodeTypeTask, &parent.ID))

		_, err := f.timers().Start(ctx, parent.ID)
		if !errors.Is(err, service.ErrTimerNotAllowed) {
			t.Fatalf("Start = %v, want service.ErrTimerNotAllowed", err)
		}
		if got := f.openEntries(); got != 0 {
			t.Errorf("%d entries were opened, want 0", got)
		}
	})

	t.Run("D9: an EMPTY project, which has no children at all", func(t *testing.T) {
		f := newFixture(t)
		p := f.create(draft("empty project", domain.NodeTypeProject, nil))

		_, err := f.timers().Start(ctx, p.ID)
		if !errors.Is(err, service.ErrTimerNotAllowed) {
			t.Fatalf("Start = %v, want service.ErrTimerNotAllowed", err)
		}
		if !errors.Is(err, domain.ErrProjectNeverDoing) {
			t.Errorf("Start = %v, want it to name domain.ErrProjectNeverDoing", err)
		}
		if got := f.openEntries(); got != 0 {
			t.Errorf("%d entries were opened, want 0", got)
		}
	})

	t.Run("D9: a project whose children are all notes", func(t *testing.T) {
		f := newFixture(t)
		p := f.create(draft("project", domain.NodeTypeProject, nil))
		f.create(draft("memo", domain.NodeTypeNote, &p.ID))

		if _, err := f.timers().Start(ctx, p.ID); !errors.Is(err, domain.ErrProjectNeverDoing) {
			t.Fatalf("Start = %v, want domain.ErrProjectNeverDoing", err)
		}
	})

	t.Run("a habit is checked off, never timed", func(t *testing.T) {
		f := newFixture(t)
		d := draft("habit", domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		h := f.create(d)

		if _, err := f.timers().Start(ctx, h.ID); !errors.Is(err, service.ErrTimerNotAllowed) {
			t.Fatalf("Start = %v, want service.ErrTimerNotAllowed", err)
		}
	})

	t.Run("a note has no column and no timer", func(t *testing.T) {
		f := newFixture(t)
		memo := f.create(draft("memo", domain.NodeTypeNote, nil))

		if _, err := f.timers().Start(ctx, memo.ID); !errors.Is(err, service.ErrTimerNotAllowed) {
			t.Fatalf("Start = %v, want service.ErrTimerNotAllowed", err)
		}
	})

	t.Run("an archived node", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("a", domain.NodeTypeTask, nil))
		if _, err := f.tasks.ArchiveNode(ctx, n.ID); err != nil {
			t.Fatalf("ArchiveNode: %v", err)
		}

		_, err := f.timers().Start(ctx, n.ID)
		if !errors.Is(err, service.ErrNodeArchived) {
			t.Fatalf("Start = %v, want service.ErrNodeArchived", err)
		}
		if !errors.Is(err, service.ErrTimerNotAllowed) {
			t.Errorf("Start = %v, want it to match ErrTimerNotAllowed too", err)
		}
	})

	t.Run("a done node: finish the card or reopen it, do not time it", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("a", domain.NodeTypeTask, nil))
		if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDone); err != nil {
			t.Fatalf("MoveToColumn(done): %v", err)
		}

		_, err := f.timers().Start(ctx, n.ID)
		if !errors.Is(err, service.ErrNodeDone) {
			t.Fatalf("Start = %v, want service.ErrNodeDone", err)
		}

		// Moving it back out of Done makes it timeable again, in the same
		// breath that clears completed_at.
		if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday); err != nil {
			t.Fatalf("MoveToColumn(today): %v", err)
		}
		if _, err := f.timers().Start(ctx, n.ID); err != nil {
			t.Errorf("Start after reopening = %v, want it to start", err)
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		f := newFixture(t)

		if _, err := f.timers().Start(ctx, "ghost"); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("Start(ghost) = %v, want store.ErrNotFound", err)
		}
	})
}

// A node whose children are ALL notes is a leaf (D2), so it starts a timer.
func TestTimerStartOnAnAllNotesParentSucceeds(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	parent := f.create(draft("parent", domain.NodeTypeTask, nil))
	f.create(draft("memo", domain.NodeTypeNote, &parent.ID))
	f.create(draft("another memo", domain.NodeTypeNote, &parent.ID))

	if _, err := f.timers().Start(ctx, parent.ID); err != nil {
		t.Fatalf("Start = %v, want it to start: a node with only note children is a leaf", err)
	}
	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want 1", got)
	}
}

func TestTimerStop(t *testing.T) {
	ctx := context.Background()

	t.Run("with nothing running: no error, no rows changed", func(t *testing.T) {
		f := newFixture(t)

		got, err := f.timers().Stop(ctx)
		if err != nil {
			t.Fatalf("Stop = %v, want no error", err)
		}
		if got != nil {
			t.Errorf("Stop returned %+v, want nil", got)
		}

		var rows int
		if err := f.db.QueryRow("SELECT count(*) FROM time_entries").Scan(&rows); err != nil {
			t.Fatalf("counting entries: %v", err)
		}
		if rows != 0 {
			t.Errorf("%d rows, want 0", rows)
		}
	})

	t.Run("closes the running entry at the injected clock", func(t *testing.T) {
		f := newFixture(t)
		timers := f.timers()
		n := f.create(draft("a", domain.NodeTypeTask, nil))

		started, err := timers.Start(ctx, n.ID)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}

		stoppedAt := testNow.Add(90 * time.Minute)
		f.now = stoppedAt

		got, err := timers.Stop(ctx)
		if err != nil {
			t.Fatalf("Stop: %v", err)
		}
		if got == nil {
			t.Fatal("Stop returned nil, want the closed entry")
		}
		if got.ID != started.ID {
			t.Errorf("entry id = %q, want %q", got.ID, started.ID)
		}
		if got.EndedAt == nil || !got.EndedAt.Equal(stoppedAt) {
			t.Errorf("ended_at = %v, want %v", got.EndedAt, stoppedAt)
		}
		if n := f.openEntries(); n != 0 {
			t.Errorf("%d open entries after Stop, want 0", n)
		}
	})

	t.Run("stopping twice is a no-op the second time", func(t *testing.T) {
		f := newFixture(t)
		timers := f.timers()
		n := f.create(draft("a", domain.NodeTypeTask, nil))

		if _, err := timers.Start(ctx, n.ID); err != nil {
			t.Fatalf("Start: %v", err)
		}
		f.now = testNow.Add(time.Minute)
		if _, err := timers.Stop(ctx); err != nil {
			t.Fatalf("Stop: %v", err)
		}

		f.now = testNow.Add(2 * time.Minute)
		got, err := timers.Stop(ctx)
		if err != nil {
			t.Fatalf("second Stop = %v, want no error", err)
		}
		if got != nil {
			t.Errorf("second Stop returned %+v, want nil", got)
		}
		if e := f.entriesOf(n.ID); len(e) != 1 || e[0].EndedAt == nil ||
			!e[0].EndedAt.Equal(testNow.Add(time.Minute)) {
			t.Errorf("entries = %+v, want one closed at the first stop", e)
		}
	})
}

func TestTimerCurrent(t *testing.T) {
	ctx := context.Background()

	t.Run("nothing running", func(t *testing.T) {
		f := newFixture(t)

		got, err := f.timers().Current(ctx)
		if err != nil {
			t.Fatalf("Current = %v", err)
		}
		if got != nil {
			t.Errorf("Current = %+v, want nil", got)
		}
	})

	t.Run("the open entry and its elapsed time, from the injected clock", func(t *testing.T) {
		f := newFixture(t)
		timers := f.timers()
		n := f.create(draft("a", domain.NodeTypeTask, nil))

		started, err := timers.Start(ctx, n.ID)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}

		f.now = testNow.Add(42 * time.Minute)
		got, err := timers.Current(ctx)
		if err != nil {
			t.Fatalf("Current = %v", err)
		}
		if got == nil {
			t.Fatal("Current = nil, want the running entry")
		}
		if got.Entry.ID != started.ID || got.Entry.NodeID != n.ID {
			t.Errorf("entry = %+v, want %q on %q", got.Entry, started.ID, n.ID)
		}
		if want := 42 * time.Minute; got.Elapsed != want {
			t.Errorf("elapsed = %v, want %v", got.Elapsed, want)
		}
	})
}

// A failure between the close and the open rolls back BOTH: the timer that was
// running is still running, and no second entry exists.
func TestTimerStartRollsBackTheCloseWhenTheOpenFails(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	a := f.create(draft("a", domain.NodeTypeTask, nil))
	b := f.create(draft("b", domain.NodeTypeTask, nil))

	if _, err := f.timers().Start(ctx, a.ID); err != nil {
		t.Fatalf("Start(a): %v", err)
	}

	f.now = testNow.Add(5 * time.Minute)

	// One Exec succeeds — closing A — and the INSERT that opens B fails.
	brittle := service.NewTimerService(
		failingBeginner{inner: f.begin, after: 1}, f.nodes, f.entries, f.clock(), f.nextID)

	if _, err := brittle.Start(ctx, b.ID); !errors.Is(err, errBoom) {
		t.Fatalf("Start(b) = %v, want the injected errBoom", err)
	}

	open, err := f.entries.OpenEntry(ctx)
	if err != nil {
		t.Fatalf("OpenEntry: %v", err)
	}
	if open.NodeID != a.ID {
		t.Errorf("the open entry is on %q, want A (%q) — the close must have rolled back", open.NodeID, a.ID)
	}
	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want 1", got)
	}
	if got := len(f.entriesOf(b.ID)); got != 0 {
		t.Errorf("%d entries on B, want 0", got)
	}
}

// The service is the policy; the schema is the backstop. Bypassing the service
// and opening a second entry straight through the repository is refused by the
// `one_open_timer` index, as store.ErrTimerRunning.
func TestTheSchemaRefusesASecondOpenEntry(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	a := f.create(draft("a", domain.NodeTypeTask, nil))
	b := f.create(draft("b", domain.NodeTypeTask, nil))

	if _, err := f.timers().Start(ctx, a.ID); err != nil {
		t.Fatalf("Start(a): %v", err)
	}

	_, err := f.entries.Open(ctx, "smuggled", b.ID, f.now.Add(time.Minute))
	if !errors.Is(err, store.ErrTimerRunning) {
		t.Fatalf("Open() = %v, want store.ErrTimerRunning from the partial unique index", err)
	}
	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want 1", got)
	}
}

// Every read and write path reports a transaction that will not start, rather
// than reporting success.
func TestTimerSurfacesTransactionFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("a", domain.NodeTypeTask, nil))

	blocked := service.NewTimerService(blockedBeginner{}, f.nodes, f.entries, f.clock(), f.nextID)
	if _, err := blocked.Start(ctx, n.ID); !errors.Is(err, errBoom) {
		t.Fatalf("Start = %v, want the injected errBoom", err)
	}

	broken := service.NewTimerService(
		wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return queryFailingTx{Tx: tx}
		}}, f.nodes, f.entries, f.clock(), f.nextID)
	if _, err := broken.Start(ctx, n.ID); err == nil {
		t.Fatal("Start succeeded with a failing read; want an error")
	}
}
