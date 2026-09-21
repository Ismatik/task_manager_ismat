package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
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

// S2-18 — the midnight case, which is why CheckToday takes no date.
//
// The day is read from the injected clock AT THE MOMENT OF THE CALL, and it is
// the same domain.Today(clock) Strip derives CheckedToday against. The test
// stands one minute either side of local midnight: before it the tick belongs
// to Friday and the strip agrees, after it the very same call means Saturday.
//
// A caller that computed the day instead — the frontend did, and sent it — can
// be one minute stale and name yesterday, at which point the check lands on a
// day the user never chose and the strip comes back still unchecked.
func TestHabitCheckTodayIsTheClocksDayAtTheMomentOfTheCall(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	h := f.weeklyHabit()

	saturday := thisFriday.AddDays(1)

	// 23:59 on the scheduled Friday.
	f.now = thisFriday.Time().Add(23*time.Hour + 59*time.Minute)

	if err := habits.CheckToday(ctx, h.ID); err != nil {
		t.Fatalf("CheckToday = %v", err)
	}
	if checked, err := habits.IsChecked(ctx, h.ID, thisFriday); err != nil || !checked {
		t.Fatalf("IsChecked(%s) = %v, %v; want true, nil", thisFriday, checked, err)
	}
	// The read path agrees with the write path, which is the whole claim: one
	// clock, one day, so the tick the user pressed is the tick that comes back.
	if view := viewOf(f.stripOf(ctx), h.ID); view == nil || !view.CheckedToday {
		t.Fatalf("the strip reports CheckedToday = %v just before midnight; want true", view)
	}

	// Two minutes later it is Saturday, and the very same call means Saturday.
	f.now = f.now.Add(2 * time.Minute)

	if err := habits.CheckToday(ctx, h.ID); err != nil {
		t.Fatalf("CheckToday after midnight = %v", err)
	}
	if checked, err := habits.IsChecked(ctx, h.ID, saturday); err != nil || !checked {
		t.Errorf("IsChecked(%s) = %v, %v; want true, nil", saturday, checked, err)
	}
	if checked, err := habits.IsChecked(ctx, h.ID, thisFriday); err != nil || !checked {
		t.Errorf("Friday's check was disturbed: IsChecked(%s) = %v, %v", thisFriday, checked, err)
	}
}

// UncheckToday removes the same day CheckToday wrote, and only that day.
func TestHabitUncheckTodayIsTheClocksDay(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	h := f.weeklyHabit()

	if err := habits.Check(ctx, h.ID, lastFriday); err != nil {
		t.Fatalf("Check = %v", err)
	}
	if err := habits.CheckToday(ctx, h.ID); err != nil {
		t.Fatalf("CheckToday = %v", err)
	}

	if err := habits.UncheckToday(ctx, h.ID); err != nil {
		t.Fatalf("UncheckToday = %v", err)
	}
	if checked, err := habits.IsChecked(ctx, h.ID, domain.Today(f.clock())); err != nil || checked {
		t.Errorf("today is still checked: %v, %v", checked, err)
	}
	if checked, err := habits.IsChecked(ctx, h.ID, lastFriday); err != nil || !checked {
		t.Errorf("an earlier day was unchecked too: IsChecked(%s) = %v, %v", lastFriday, checked, err)
	}
}

// Both dateless doors refuse a node that is not a habit, like every other
// method here — the convenience wrapper does not open a side entrance.
func TestHabitTodayDoorsRefuseANonHabit(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	habits := f.habits()
	task := f.create(draft("not a habit", domain.NodeTypeTask, nil))

	if err := habits.CheckToday(ctx, task.ID); !errors.Is(err, service.ErrNotAHabit) {
		t.Errorf("CheckToday = %v, want ErrNotAHabit", err)
	}
	if err := habits.UncheckToday(ctx, task.ID); !errors.Is(err, service.ErrNotAHabit) {
		t.Errorf("UncheckToday = %v, want ErrNotAHabit", err)
	}
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

// ---------------------------------------------------------------------------
// S2-04 — the habit strip read.

// stripOf returns the strip, failing the test if it cannot be read.
func (f *fixture) stripOf(ctx context.Context) []service.HabitView {
	f.t.Helper()

	strip, err := f.habits().Strip(ctx)
	if err != nil {
		f.t.Fatalf("Strip: %v", err)
	}
	return strip
}

// viewOf returns the strip entry for one habit, or nil.
func viewOf(strip []service.HabitView, id string) *service.HabitView {
	for i := range strip {
		if strip[i].Node.ID == id {
			return &strip[i]
		}
	}
	return nil
}

// D5 through the strip, not only through HabitService.Streak: a weekly habit
// checked four Fridays running reads 4, and today's pending occurrence does not
// break it.
func TestStripReportsTheStreak(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	h := f.weeklyHabit()

	for _, d := range []domain.Date{fourFridaysAgo, threeFridaysAgo, twoFridaysAgo, lastFriday} {
		if err := f.habits().Check(ctx, h.ID, d); err != nil {
			t.Fatalf("Check(%s) = %v", d, err)
		}
	}

	t.Run("four consecutive Fridays, today still pending", func(t *testing.T) {
		v := viewOf(f.stripOf(ctx), h.ID)
		if v == nil {
			t.Fatal("the habit is not on the strip")
		}
		if v.Streak != 4 {
			t.Errorf("Streak = %d, want 4 — today is scheduled and unchecked, which is pending, not missed", v.Streak)
		}
		if !v.ScheduledToday {
			t.Error("ScheduledToday = false, but today is a Friday and the rule is FREQ=WEEKLY;BYDAY=FR")
		}
		if v.CheckedToday {
			t.Error("CheckedToday = true, but today was never checked")
		}
	})

	t.Run("checking today makes it five", func(t *testing.T) {
		if err := f.habits().Check(ctx, h.ID, thisFriday); err != nil {
			t.Fatalf("Check: %v", err)
		}
		v := viewOf(f.stripOf(ctx), h.ID)
		if v.Streak != 5 {
			t.Errorf("Streak = %d, want 5", v.Streak)
		}
		if !v.CheckedToday {
			t.Error("CheckedToday = false right after Check")
		}
	})

	t.Run("unchecking today takes it back to four", func(t *testing.T) {
		if err := f.habits().Uncheck(ctx, h.ID, thisFriday); err != nil {
			t.Fatalf("Uncheck: %v", err)
		}
		v := viewOf(f.stripOf(ctx), h.ID)
		if v.CheckedToday {
			t.Error("CheckedToday = true right after Uncheck")
		}
		if v.Streak != 4 {
			t.Errorf("Streak = %d, want 4", v.Streak)
		}
	})
}

// A habit that is not scheduled today is still ON the strip, carrying
// ScheduledToday = false. Whether the strip draws it is the frontend's
// presentation decision and must not cost a second call.
func TestStripReportsScheduledTodayWithoutFiltering(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	friday := f.weeklyHabit()

	// A Monday habit, on a Friday: on the strip, not scheduled.
	f.now = fourFridaysAgo.Time().Add(9 * time.Hour)
	d := draft("monday review", domain.NodeTypeHabit, nil)
	d.Recurrence = ptr("FREQ=WEEKLY;BYDAY=MO")
	monday := f.create(d)
	f.now = testNow

	strip := f.stripOf(ctx)
	if len(strip) != 2 {
		t.Fatalf("the strip holds %d habits, want both", len(strip))
	}
	if v := viewOf(strip, friday.ID); v == nil || !v.ScheduledToday {
		t.Errorf("the Friday habit reports %+v, want ScheduledToday = true", v)
	}
	if v := viewOf(strip, monday.ID); v == nil || v.ScheduledToday {
		t.Errorf("the Monday habit reports %+v, want it present with ScheduledToday = false", v)
	}
}

// Membership is the TYPE, asked once: a habit under a project is on the strip
// exactly once, and a task, a note, a project and a bug never are.
func TestStripMembershipIsTheType(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("a project", domain.NodeTypeProject, nil))
	f.create(draft("a task", domain.NodeTypeTask, &p.ID))
	f.create(draft("a note", domain.NodeTypeNote, nil))
	f.create(draft("a bug", domain.NodeTypeBug, nil))

	d := draft("a nested habit", domain.NodeTypeHabit, &p.ID)
	d.Recurrence = ptr("FREQ=DAILY")
	nested := f.create(d)

	strip := f.stripOf(ctx)
	if len(strip) != 1 {
		t.Fatalf("the strip holds %d entries, want exactly the one habit: %+v", len(strip), strip)
	}
	if strip[0].Node.ID != nested.ID {
		t.Errorf("the strip holds %q, want the nested habit %q", strip[0].Node.ID, nested.ID)
	}

	// D10's other half, asserted alongside: the habit contributes nothing to
	// the project it is nested under, which is a different question.
	if v := find(boardOf(ctx, t, f), nested.ID); v != nil {
		t.Errorf("the habit is on the board too, in the %q column", v.Status)
	}
}

// boardOf reads the board, failing the test if it cannot.
func boardOf(ctx context.Context, t *testing.T, f *fixture) []service.ColumnView {
	t.Helper()

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	return board
}

// Archived habits are excluded, as everywhere else.
func TestStripExcludesArchivedHabits(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	h := f.weeklyHabit()

	if got := len(f.stripOf(ctx)); got != 1 {
		t.Fatalf("the strip holds %d habits before the archive, want 1", got)
	}
	if _, err := f.tasks.ArchiveNode(ctx, h.ID); err != nil {
		t.Fatalf("ArchiveNode: %v", err)
	}
	if got := f.stripOf(ctx); len(got) != 0 {
		t.Errorf("the strip holds %+v after archiving the only habit, want nothing", got)
	}
}

// The strip is ordered by (sort_order, id), the same deterministic order every
// other list uses.
func TestStripIsOrderedBySortOrderThenID(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	var want []string
	for i := range 4 {
		d := draft(fmt.Sprintf("habit %d", i), domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		want = append(want, f.create(d).ID)
	}

	var got []string
	for _, v := range f.stripOf(ctx) {
		got = append(got, v.Node.ID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the strip is ordered %v, want %v", got, want)
	}
}

// The cost is fixed: fifty habits must not cost fifty queries. This is the
// property that makes the strip a read rather than an N+1, and the reason
// HabitCheckRepo gained a bulk ChecksInRange.
func TestStripCostsAConstantNumberOfQueries(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	one := f.countedStrip(ctx, t, 1)
	fifty := f.countedStrip(ctx, t, 50)

	if one != fifty {
		t.Errorf("1 habit cost %d queries and 50 cost %d; the count must not depend on the habit count",
			one, fifty)
	}
	if fifty > 3 {
		t.Errorf("the strip issued %d queries, want a small constant", fifty)
	}
}

// countedStrip builds n habits in a fresh database and returns how many
// statements one Strip() call issues.
func (f *fixture) countedStrip(ctx context.Context, t *testing.T, n int) int {
	t.Helper()

	fresh := newFixture(t)
	for i := range n {
		d := draft(fmt.Sprintf("habit %d", i), domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		h := fresh.create(d)
		if err := fresh.habits().Check(ctx, h.ID, domain.Today(fresh.clock())); err != nil {
			t.Fatalf("Check: %v", err)
		}
	}

	count := 0
	counting := countingExec{inner: fresh.db, queries: &count}
	svc := service.NewHabitService(
		store.NewNodeRepo(counting), store.NewHabitCheckRepo(counting), fresh.clock())

	strip, err := svc.Strip(ctx)
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if len(strip) != n {
		t.Fatalf("the strip holds %d habits, want %d", len(strip), n)
	}
	return count
}

// countingExec counts the statements a repository issues against the database.
type countingExec struct {
	inner   *sql.DB
	queries *int
}

func (e countingExec) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	*e.queries++
	return e.inner.ExecContext(ctx, query, args...)
}

func (e countingExec) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	*e.queries++
	return e.inner.QueryContext(ctx, query, args...)
}

func (e countingExec) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	*e.queries++
	return e.inner.QueryRowContext(ctx, query, args...)
}

// An empty database has an empty strip, and reads no checks at all.
func TestStripOnAnEmptyDatabase(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	strip, err := f.habits().Strip(ctx)
	if err != nil {
		t.Fatalf("Strip: %v", err)
	}
	if len(strip) != 0 {
		t.Errorf("the strip holds %+v, want nothing", strip)
	}
}

// failingExec fails the (after+1)-th statement it is given, so that each of the
// strip's two reads can be broken in turn.
type failingExec struct {
	inner     *sql.DB
	remaining *int
}

func (e failingExec) spend() error {
	if *e.remaining <= 0 {
		return errBoom
	}
	*e.remaining--
	return nil
}

func (e failingExec) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := e.spend(); err != nil {
		return nil, err
	}
	return e.inner.ExecContext(ctx, query, args...)
}

func (e failingExec) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := e.spend(); err != nil {
		return nil, err
	}
	return e.inner.QueryContext(ctx, query, args...)
}

func (e failingExec) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if err := e.spend(); err != nil {
		// A *sql.Row carrying an error is not constructible from here, so the
		// query is run and the caller sees the row; Strip does not use
		// QueryRowContext, which is what this stub is for.
		return e.inner.QueryRowContext(ctx, query, args...)
	}
	return e.inner.QueryRowContext(ctx, query, args...)
}

// A strip that cannot reach the database says so, rather than returning an
// empty strip that looks exactly like "you have no habits".
func TestStripSurfacesStoreFailures(t *testing.T) {
	ctx := context.Background()

	for name, budget := range map[string]int{
		"the habits cannot be read": 0,
		"the checks cannot be read": 1,
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			f.weeklyHabit()

			remaining := budget
			exec := failingExec{inner: f.db, remaining: &remaining}
			svc := service.NewHabitService(
				store.NewNodeRepo(exec), store.NewHabitCheckRepo(exec), f.clock())

			if _, err := svc.Strip(ctx); !errors.Is(err, errBoom) {
				t.Fatalf("Strip = %v, want the injected errBoom", err)
			}
		})
	}
}

// A habit row that is corrupt — no recurrence, or one that will not parse, or
// no created_at to anchor the schedule — fails the strip loudly. None of these
// can be produced through CreateNode; they are written straight through the
// repository, which is how a bug or an older version would leave them.
func TestStripRefusesACorruptHabitRow(t *testing.T) {
	ctx := context.Background()

	corrupt := map[string]func(n *domain.Node){
		"no recurrence at all": func(n *domain.Node) { n.Recurrence = nil },
		"a recurrence that will not parse": func(n *domain.Node) {
			n.Recurrence = ptr("FREQ=FORTNIGHTLY")
		},
		"no created_at to anchor the schedule": func(n *domain.Node) { n.CreatedAt = time.Time{} },
	}

	for name, break_ := range corrupt {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)

			n := bareNode("h1", nil, domain.NodeTypeHabit)
			n.Recurrence = ptr("FREQ=DAILY")
			break_(&n)
			if err := f.nodes.Create(ctx, n); err != nil {
				t.Fatalf("Create: %v", err)
			}

			if _, err := f.habits().Strip(ctx); err == nil {
				t.Fatal("Strip succeeded on a corrupt habit row; want an error")
			}
		})
	}
}

// The check window starts at the EARLIEST habit's creation day, whatever order
// the habits come back in: a habit created long ago must not lose the checks
// that its streak walks back through because a newer one was seen first.
func TestStripReadsChecksFromTheEarliestHabitsCreation(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	// The NEWER habit sorts first, so the window's lower bound has to be
	// lowered by the second one the loop sees.
	f.now = testNow.Add(-24 * time.Hour)
	recent := draft("recent", domain.NodeTypeHabit, nil)
	recent.Recurrence = ptr("FREQ=DAILY")
	f.create(recent)

	f.now = fourFridaysAgo.Time().Add(9 * time.Hour)
	old := draft("old", domain.NodeTypeHabit, nil)
	old.Recurrence = ptr("FREQ=WEEKLY;BYDAY=FR")
	h := f.create(old)
	f.now = testNow

	for _, d := range []domain.Date{threeFridaysAgo, twoFridaysAgo, lastFriday} {
		if err := f.habits().Check(ctx, h.ID, d); err != nil {
			t.Fatalf("Check(%s): %v", d, err)
		}
	}

	v := viewOf(f.stripOf(ctx), h.ID)
	if v == nil {
		t.Fatal("the older habit is not on the strip")
	}
	if v.Streak != 3 {
		t.Errorf("Streak = %d, want 3 — the check window must reach back to the oldest habit's creation", v.Streak)
	}
}
