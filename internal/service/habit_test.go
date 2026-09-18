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

// habits returns a habit service over the fixture's database and clock.
func (f *fixture) habits() *service.HabitService {
	f.t.Helper()

	return service.NewHabitService(f.nodes, f.checks, f.clock())
}

// fridays around testNow (2026-09-18 is a Friday).
var (
	fourFridaysAgo  = domain.NewDate(2026, time.August, 21)
	threeFridaysAgo = domain.NewDate(2026, time.August, 28)
	twoFridaysAgo   = domain.NewDate(2026, time.September, 4)
	lastFriday      = domain.NewDate(2026, time.September, 11)
	thisFriday      = domain.NewDate(2026, time.September, 18)
)

// weeklyHabit creates a habit whose schedule starts four Fridays before
// testNow, so that there are five scheduled occurrences up to and including
// today: a streak has somewhere to run.
func (f *fixture) weeklyHabit() domain.Node {
	f.t.Helper()

	f.now = fourFridaysAgo.Time().Add(9 * time.Hour)
	d := draft("ship the weekly report", domain.NodeTypeHabit, nil)
	d.Recurrence = ptr("FREQ=WEEKLY;BYDAY=FR")
	n := f.create(d)
	f.now = testNow
	return n
}

// D5, through the service and a real database: a weekly habit checked on four
// consecutive SCHEDULED dates has a streak of 4 — not 4 days, not 28.
func TestHabitStreakCountsScheduledOccurrences(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	h := f.weeklyHabit()

	t.Run("no checks yet", func(t *testing.T) {
		got, err := habits.Streak(ctx, h.ID)
		if err != nil {
			t.Fatalf("Streak = %v", err)
		}
		if got != 0 {
			t.Errorf("Streak = %d, want 0", got)
		}
	})

	for _, d := range []domain.Date{threeFridaysAgo, twoFridaysAgo, lastFriday, thisFriday} {
		if err := habits.Check(ctx, h.ID, d); err != nil {
			t.Fatalf("Check(%s) = %v", d, err)
		}
	}

	t.Run("four consecutive scheduled Fridays report 4", func(t *testing.T) {
		got, err := habits.Streak(ctx, h.ID)
		if err != nil {
			t.Fatalf("Streak = %v", err)
		}
		if got != 4 {
			t.Errorf("Streak = %d, want 4", got)
		}
	})

	t.Run("today scheduled but unchecked leaves the streak alone", func(t *testing.T) {
		if err := habits.Uncheck(ctx, h.ID, thisFriday); err != nil {
			t.Fatalf("Uncheck = %v", err)
		}
		got, err := habits.Streak(ctx, h.ID)
		if err != nil {
			t.Fatalf("Streak = %v", err)
		}
		if got != 3 {
			t.Errorf("Streak = %d, want 3 — today is pending, not missed", got)
		}
	})

	t.Run("unchecking a day in the middle breaks the streak there", func(t *testing.T) {
		if err := habits.Uncheck(ctx, h.ID, twoFridaysAgo); err != nil {
			t.Fatalf("Uncheck = %v", err)
		}
		got, err := habits.Streak(ctx, h.ID)
		if err != nil {
			t.Fatalf("Streak = %v", err)
		}
		if got != 1 {
			t.Errorf("Streak = %d, want 1 — only the most recent Friday survives", got)
		}
	})

	t.Run("a check on an unscheduled day does not repair it", func(t *testing.T) {
		wednesday := domain.NewDate(2026, time.September, 9)
		if err := habits.Check(ctx, h.ID, wednesday); err != nil {
			t.Fatalf("Check = %v", err)
		}
		got, err := habits.Streak(ctx, h.ID)
		if err != nil {
			t.Fatalf("Streak = %v", err)
		}
		if got != 1 {
			t.Errorf("Streak = %d, want 1 — an unscheduled tick is not an occurrence", got)
		}
	})
}

// Check and uncheck are visible to the streak, and to IsChecked, immediately.
func TestHabitCheckAndUncheckRoundTrip(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	h := f.weeklyHabit()

	checked, err := habits.IsChecked(ctx, h.ID, thisFriday)
	if err != nil {
		t.Fatalf("IsChecked = %v", err)
	}
	if checked {
		t.Fatal("a fresh habit reports today as checked")
	}

	if err := habits.Check(ctx, h.ID, thisFriday); err != nil {
		t.Fatalf("Check = %v", err)
	}
	if got, err := habits.Streak(ctx, h.ID); err != nil || got != 1 {
		t.Errorf("Streak = %d, %v; want 1, nil", got, err)
	}

	t.Run("checking twice is the same fact, not two", func(t *testing.T) {
		if err := habits.Check(ctx, h.ID, thisFriday); err != nil {
			t.Fatalf("Check again = %v", err)
		}
		if got, err := habits.Streak(ctx, h.ID); err != nil || got != 1 {
			t.Errorf("Streak = %d, %v; want 1, nil", got, err)
		}
	})

	if err := habits.Uncheck(ctx, h.ID, thisFriday); err != nil {
		t.Fatalf("Uncheck = %v", err)
	}
	if got, err := habits.Streak(ctx, h.ID); err != nil || got != 0 {
		t.Errorf("Streak after uncheck = %d, %v; want 0, nil", got, err)
	}

	t.Run("unchecking a day that was never checked is a no-op", func(t *testing.T) {
		if err := habits.Uncheck(ctx, h.ID, lastFriday); err != nil {
			t.Errorf("Uncheck = %v, want no error", err)
		}
	})
}

// Every method refuses a node that is not a habit, and says so in a way the
// caller can match.
func TestHabitServiceRefusesNodesThatAreNotHabits(t *testing.T) {
	ctx := context.Background()

	for _, typ := range []domain.NodeType{
		domain.NodeTypeTask, domain.NodeTypeProject, domain.NodeTypeNote, domain.NodeTypeBug,
	} {
		t.Run(typ.String(), func(t *testing.T) {
			f := newFixture(t)
			habits := f.habits()
			n := f.create(draft("x", typ, nil))

			calls := map[string]func() error{
				"Check":     func() error { return habits.Check(ctx, n.ID, thisFriday) },
				"Uncheck":   func() error { return habits.Uncheck(ctx, n.ID, thisFriday) },
				"IsChecked": func() error { _, err := habits.IsChecked(ctx, n.ID, thisFriday); return err },
				"Streak":    func() error { _, err := habits.Streak(ctx, n.ID); return err },
				"DueToday":  func() error { _, err := habits.DueToday(ctx, n.ID); return err },
			}
			for name, call := range calls {
				t.Run(name, func(t *testing.T) {
					if err := call(); !errors.Is(err, service.ErrNotAHabit) {
						t.Fatalf("%s = %v, want service.ErrNotAHabit", name, err)
					}
				})
			}

			var rows int
			if err := f.db.QueryRow("SELECT count(*) FROM habit_checks").Scan(&rows); err != nil {
				t.Fatalf("counting checks: %v", err)
			}
			if rows != 0 {
				t.Errorf("%d habit_checks rows, want 0", rows)
			}
		})
	}

	t.Run("a node that does not exist", func(t *testing.T) {
		f := newFixture(t)

		if err := f.habits().Check(ctx, "ghost", thisFriday); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("Check(ghost) = %v, want store.ErrNotFound", err)
		}
	})
}

// A habit with no recurrence cannot be created through the service at all, and
// if one is in the file anyway, the streak says "cannot say" rather than 0.
func TestHabitWithoutARecurrence(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	t.Run("CreateNode refuses to make one", func(t *testing.T) {
		if _, err := f.tasks.CreateNode(ctx, draft("x", domain.NodeTypeHabit, nil)); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("CreateNode = %v, want domain.ErrInvalid", err)
		}
	})

	// Written straight through the repository, which is the only way such a row
	// could exist: an older version, or a hand-edited file.
	broken := domain.Node{
		ID:        "legacy-habit",
		Type:      domain.NodeTypeHabit,
		Title:     "a habit from before the rule",
		Status:    domain.StatusBacklog,
		DueSource: domain.DueSourceManual,
		Priority:  domain.Priority4,
		CreatedAt: testNow,
		UpdatedAt: testNow,
	}
	if err := f.nodes.Create(ctx, broken); err != nil {
		t.Fatalf("Create: %v", err)
	}

	t.Run("Streak is an error, not 0", func(t *testing.T) {
		got, err := f.habits().Streak(ctx, broken.ID)
		if !errors.Is(err, domain.ErrNoRecurrence) {
			t.Fatalf("Streak = %d, %v; want domain.ErrNoRecurrence", got, err)
		}
	})

	t.Run("DueToday is an error, not false", func(t *testing.T) {
		if _, err := f.habits().DueToday(ctx, broken.ID); !errors.Is(err, domain.ErrNoRecurrence) {
			t.Fatalf("DueToday = %v, want domain.ErrNoRecurrence", err)
		}
	})

	t.Run("an unparseable rule is an error too", func(t *testing.T) {
		broken.ID = "nonsense-rule"
		broken.Recurrence = ptr("EVERY OTHER TUESDAY")
		if err := f.nodes.Create(ctx, broken); err != nil {
			t.Fatalf("Create: %v", err)
		}

		if _, err := f.habits().Streak(ctx, broken.ID); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
			t.Errorf("Streak = %v, want domain.ErrUnsupportedRecurrence", err)
		}
		if _, err := f.habits().DueToday(ctx, broken.ID); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
			t.Errorf("DueToday = %v, want domain.ErrUnsupportedRecurrence", err)
		}
	})
}

// DueToday is the habit strip's question: is today one of this habit's days?
func TestHabitDueToday(t *testing.T) {
	ctx := context.Background()

	t.Run("a Friday habit on a Friday", func(t *testing.T) {
		f := newFixture(t)
		h := f.weeklyHabit()

		got, err := f.habits().DueToday(ctx, h.ID)
		if err != nil {
			t.Fatalf("DueToday = %v", err)
		}
		if !got {
			t.Error("DueToday = false on the scheduled Friday")
		}
	})

	t.Run("the same habit on the Saturday after", func(t *testing.T) {
		f := newFixture(t)
		h := f.weeklyHabit()
		f.now = testNow.AddDate(0, 0, 1)

		got, err := f.habits().DueToday(ctx, h.ID)
		if err != nil {
			t.Fatalf("DueToday = %v", err)
		}
		if got {
			t.Error("DueToday = true on a Saturday for a Friday habit")
		}
	})

	t.Run("a daily habit is due every day", func(t *testing.T) {
		f := newFixture(t)
		d := draft("stretch", domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		h := f.create(d)

		for offset := range 4 {
			f.now = testNow.AddDate(0, 0, offset)
			got, err := f.habits().DueToday(ctx, h.ID)
			if err != nil {
				t.Fatalf("DueToday = %v", err)
			}
			if !got {
				t.Errorf("day +%d: DueToday = false, want true", offset)
			}
		}
	})

	t.Run("DueOn answers for any day", func(t *testing.T) {
		f := newFixture(t)
		h := f.weeklyHabit()

		for _, tc := range []struct {
			date domain.Date
			want bool
		}{
			{thisFriday, true},
			{lastFriday, true},
			{domain.NewDate(2026, time.September, 17), false},
			// Before the habit existed: not scheduled, whatever the weekday.
			{domain.NewDate(2026, time.August, 14), false},
		} {
			got, err := f.habits().DueOn(ctx, h.ID, tc.date)
			if err != nil {
				t.Fatalf("DueOn(%s) = %v", tc.date, err)
			}
			if got != tc.want {
				t.Errorf("DueOn(%s) = %v, want %v", tc.date, got, tc.want)
			}
		}
	})
}

// D2: habits never appear in Kanban columns, so the habit service is their read
// path — a habit's stored status is not a column placement and nothing here
// consults it.
//
// The other half of this cross-check lives in read_test.go (S1-21), where the
// board exists and TestBoardExcludesHabits asserts that no column ever contains
// one.
func TestHabitReadPathIgnoresTheKanbanStatus(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	h := f.weeklyHabit()

	// Whatever a stray write puts in the status column, the habit answers the
	// same questions the same way.
	before, err := habits.Streak(ctx, h.ID)
	if err != nil {
		t.Fatalf("Streak = %v", err)
	}

	stored := f.get(h.ID)
	stored.Status = domain.StatusDone
	if err := f.nodes.Update(ctx, stored); err != nil {
		t.Fatalf("Update: %v", err)
	}

	after, err := habits.Streak(ctx, h.ID)
	if err != nil {
		t.Fatalf("Streak = %v", err)
	}
	if before != after {
		t.Errorf("the streak changed with the status column: %d -> %d", before, after)
	}
	due, err := habits.DueToday(ctx, h.ID)
	if err != nil {
		t.Fatalf("DueToday = %v", err)
	}
	if !due {
		t.Error("DueToday changed with the status column")
	}
}

// The read paths report a database failure rather than a zero.
func TestHabitServiceSurfacesStoreFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	h := f.weeklyHabit()

	if err := f.db.Close(); err != nil {
		t.Fatalf("closing the database: %v", err)
	}

	habits := f.habits()
	if _, err := habits.Streak(ctx, h.ID); err == nil {
		t.Error("Streak succeeded against a closed database")
	}
	if _, err := habits.DueToday(ctx, h.ID); err == nil {
		t.Error("DueToday succeeded against a closed database")
	}
	if err := habits.Check(ctx, h.ID, thisFriday); err == nil {
		t.Error("Check succeeded against a closed database")
	}
}
