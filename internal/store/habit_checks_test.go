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

// habitCheckRepo returns a habit-check repository over a freshly migrated temp
// database that already has a habit "h1" (and a second habit "h2").
func habitCheckRepo(t *testing.T) (*HabitCheckRepo, *NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	nodes := NewNodeRepo(db)

	ctx := context.Background()
	for _, id := range []string{"h1", "h2"} {
		h := newTestNode(id)
		h.Type = domain.NodeTypeHabit
		h.Recurrence = strPtr("FREQ=DAILY")
		mustCreate(t, ctx, nodes, h)
	}
	return NewHabitCheckRepo(db), nodes, db
}

func checkDates(checks []domain.HabitCheck) []string {
	out := make([]string, len(checks))
	for i, c := range checks {
		out[i] = c.Date.String()
	}
	return out
}

func mustCheck(t *testing.T, ctx context.Context, r *HabitCheckRepo, nodeID string, date domain.Date) {
	t.Helper()

	if err := r.Check(ctx, nodeID, date); err != nil {
		t.Fatalf("Check(%q, %s): %v", nodeID, date, err)
	}
}

func TestHabitCheckRepoCheckAndUncheck(t *testing.T) {
	ctx := context.Background()
	day := domain.NewDate(2026, time.September, 18)

	t.Run("check makes the day checked, uncheck makes it unchecked again", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		checked, err := r.IsChecked(ctx, "h1", day)
		if err != nil {
			t.Fatalf("IsChecked: %v", err)
		}
		if checked {
			t.Error("IsChecked on a fresh habit = true, want false")
		}

		mustCheck(t, ctx, r, "h1", day)

		if checked, err = r.IsChecked(ctx, "h1", day); err != nil {
			t.Fatalf("IsChecked: %v", err)
		}
		if !checked {
			t.Error("IsChecked after a check = false, want true")
		}

		if err := r.Uncheck(ctx, "h1", day); err != nil {
			t.Fatalf("Uncheck: %v", err)
		}
		if checked, err = r.IsChecked(ctx, "h1", day); err != nil {
			t.Fatalf("IsChecked: %v", err)
		}
		if checked {
			t.Error("IsChecked after an uncheck = true, want false")
		}
	})

	// A day is either done or not. The second check is the same fact.
	t.Run("checking twice leaves exactly one row and returns no error", func(t *testing.T) {
		r, _, db := habitCheckRepo(t)

		mustCheck(t, ctx, r, "h1", day)
		if err := r.Check(ctx, "h1", day); err != nil {
			t.Fatalf("second Check = %v, want nil: checking twice is a no-op", err)
		}
		if n := countRows(t, db, "habit_checks"); n != 1 {
			t.Errorf("habit_checks has %d rows, want exactly 1", n)
		}
	})

	t.Run("unchecking a day that was never checked is a no-op", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		if err := r.Uncheck(ctx, "h1", day); err != nil {
			t.Errorf("Uncheck on an unchecked day = %v, want nil", err)
		}
	})

	t.Run("two habits keep their own checks on the same day", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		mustCheck(t, ctx, r, "h1", day)

		checked, err := r.IsChecked(ctx, "h2", day)
		if err != nil {
			t.Fatalf("IsChecked: %v", err)
		}
		if checked {
			t.Error("checking h1 also checked h2")
		}
	})

	t.Run("a habit that does not exist is refused by the foreign key", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		if err := r.Check(ctx, "ghost", day); !errors.Is(err, ErrConstraint) {
			t.Errorf("Check(unknown node) error = %v, want one matching ErrConstraint", err)
		}
	})

	t.Run("IsChecked on a habit that does not exist is false, not an error", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		checked, err := r.IsChecked(ctx, "ghost", day)
		if err != nil {
			t.Fatalf("IsChecked(ghost): %v", err)
		}
		if checked {
			t.Error("IsChecked(ghost) = true, want false")
		}
	})
}

// The from/to window is inclusive on both ends, and the final day is the one a
// streak turns on, so both ends get their own subtest.
func TestHabitCheckRepoChecksForNode(t *testing.T) {
	ctx := context.Background()

	// Checks on the 14th through the 20th of September 2026.
	seed := func(t *testing.T) *HabitCheckRepo {
		t.Helper()
		r, _, _ := habitCheckRepo(t)
		for day := 14; day <= 20; day++ {
			mustCheck(t, ctx, r, "h1", domain.NewDate(2026, time.September, day))
		}
		return r
	}

	for _, tc := range []struct {
		name     string
		from, to domain.Date
		want     []string
	}{
		{
			name: "the from day is included",
			from: domain.NewDate(2026, time.September, 16),
			to:   domain.NewDate(2026, time.September, 18),
			want: []string{"2026-09-16", "2026-09-17", "2026-09-18"},
		},
		{
			name: "the to day is included",
			from: domain.NewDate(2026, time.September, 18),
			to:   domain.NewDate(2026, time.September, 20),
			want: []string{"2026-09-18", "2026-09-19", "2026-09-20"},
		},
		{
			name: "the day before from is excluded",
			from: domain.NewDate(2026, time.September, 15),
			to:   domain.NewDate(2026, time.September, 15),
			want: []string{"2026-09-15"},
		},
		{
			name: "a window wider than the data returns all of it",
			from: domain.NewDate(2026, time.January, 1),
			to:   domain.NewDate(2026, time.December, 31),
			want: []string{
				"2026-09-14", "2026-09-15", "2026-09-16", "2026-09-17",
				"2026-09-18", "2026-09-19", "2026-09-20",
			},
		},
		{
			name: "a window that misses everything returns nothing",
			from: domain.NewDate(2026, time.October, 1),
			to:   domain.NewDate(2026, time.October, 31),
			want: []string{},
		},
		{
			name: "from after to returns nothing rather than erroring",
			from: domain.NewDate(2026, time.September, 20),
			to:   domain.NewDate(2026, time.September, 14),
			want: []string{},
		},
		{
			// The column is 'YYYY-MM-DD' text, so a month boundary is only
			// ordered correctly because of the zero padding 0002 insists on.
			name: "a single day is a from == to window",
			from: domain.NewDate(2026, time.September, 17),
			to:   domain.NewDate(2026, time.September, 17),
			want: []string{"2026-09-17"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := seed(t)

			got, err := r.ChecksForNode(ctx, "h1", tc.from, tc.to)
			if err != nil {
				t.Fatalf("ChecksForNode: %v", err)
			}
			if !slices.Equal(checkDates(got), tc.want) {
				t.Errorf("ChecksForNode(%s..%s) = %v, want %v", tc.from, tc.to, checkDates(got), tc.want)
			}
			for _, c := range got {
				if c.NodeID != "h1" {
					t.Errorf("a check came back for node %q, want h1", c.NodeID)
				}
			}
		})
	}

	t.Run("another habit's checks are not included", func(t *testing.T) {
		r := seed(t)
		mustCheck(t, ctx, r, "h2", domain.NewDate(2026, time.September, 18))

		got, err := r.ChecksForNode(ctx, "h2", domain.NewDate(2026, time.September, 1), domain.NewDate(2026, time.September, 30))
		if err != nil {
			t.Fatalf("ChecksForNode: %v", err)
		}
		if want := []string{"2026-09-18"}; !slices.Equal(checkDates(got), want) {
			t.Errorf("ChecksForNode(h2) = %v, want %v", checkDates(got), want)
		}
	})

	t.Run("a habit with no checks lists nothing rather than nil", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)

		got, err := r.ChecksForNode(ctx, "h1", domain.NewDate(2026, time.September, 1), domain.NewDate(2026, time.September, 30))
		if err != nil {
			t.Fatalf("ChecksForNode: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ChecksForNode = %v, want an empty non-nil slice", got)
		}
	})

	t.Run("checks cross a month boundary in calendar order", func(t *testing.T) {
		r, _, _ := habitCheckRepo(t)
		for _, d := range []domain.Date{
			domain.NewDate(2026, time.October, 1),
			domain.NewDate(2026, time.September, 30),
			domain.NewDate(2026, time.September, 9),
		} {
			mustCheck(t, ctx, r, "h1", d)
		}

		got, err := r.ChecksForNode(ctx, "h1", domain.NewDate(2026, time.September, 1), domain.NewDate(2026, time.October, 31))
		if err != nil {
			t.Fatalf("ChecksForNode: %v", err)
		}
		want := []string{"2026-09-09", "2026-09-30", "2026-10-01"}
		if !slices.Equal(checkDates(got), want) {
			t.Errorf("ChecksForNode = %v, want %v", checkDates(got), want)
		}
	})

	t.Run("a date the repository did not write is an error, not a zero date", func(t *testing.T) {
		r, _, db := habitCheckRepo(t)
		mustCheck(t, ctx, r, "h1", domain.NewDate(2026, time.September, 18))

		// Well formed, inside the window, and not a day that exists.
		if _, err := db.ExecContext(ctx, "UPDATE habit_checks SET date = '2026-02-30' WHERE node_id = 'h1'"); err != nil {
			t.Fatalf("corrupting date: %v", err)
		}
		_, err := r.ChecksForNode(ctx, "h1", domain.NewDate(2026, time.January, 1), domain.NewDate(2026, time.December, 31))
		if err == nil {
			t.Error("ChecksForNode over an unparseable date returned no error")
		}
	})
}

func TestHabitChecksCascadeWithTheNode(t *testing.T) {
	ctx := context.Background()
	r, nodes, db := habitCheckRepo(t)

	mustCheck(t, ctx, r, "h1", domain.NewDate(2026, time.September, 18))
	mustCheck(t, ctx, r, "h2", domain.NewDate(2026, time.September, 18))

	if err := nodes.Delete(ctx, "h1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if n := countRows(t, db, "habit_checks"); n != 1 {
		t.Errorf("habit_checks has %d rows, want 1 — only h1's check should be gone", n)
	}
	checked, err := r.IsChecked(ctx, "h2", domain.NewDate(2026, time.September, 18))
	if err != nil {
		t.Fatalf("IsChecked: %v", err)
	}
	if !checked {
		t.Error("deleting h1 removed h2's check as well")
	}
}

func TestHabitCheckRepoRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	r, _, db := habitCheckRepo(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	day := domain.NewDate(2026, time.September, 18)
	if err := r.WithExecutor(tx).Check(ctx, "h1", day); err != nil {
		t.Fatalf("Check in the transaction: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	checked, err := r.IsChecked(ctx, "h1", day)
	if err != nil {
		t.Fatalf("IsChecked: %v", err)
	}
	if checked {
		t.Error("the check survived a rollback")
	}
}
