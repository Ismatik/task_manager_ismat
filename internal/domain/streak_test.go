package domain_test

import (
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
)

// habitNode builds a habit anchored on createdOn with the given rule.
func habitNode(rule string, createdOn domain.Date) domain.Node {
	n := nd("h1", "", domain.NodeTypeHabit, domain.StatusBacklog)
	n.Recurrence = ptr(rule)
	n.CreatedAt = time.Date(createdOn.Year, createdOn.Month, createdOn.Day, 8, 0, 0, 0, time.UTC)
	n.UpdatedAt = n.CreatedAt
	return n
}

func checksOn(nodeID string, dates ...domain.Date) []domain.HabitCheck {
	out := make([]domain.HabitCheck, len(dates))
	for i, x := range dates {
		out[i] = domain.HabitCheck{NodeID: nodeID, Date: x}
	}
	return out
}

// D5: a streak counts consecutive SCHEDULED occurrences, not calendar days.
func TestStreak(t *testing.T) {
	// A weekly habit every Friday, anchored on Friday 2026-08-28.
	// Its occurrences are 08-28, 09-04, 09-11, 09-18, 09-25, ...
	weeklyStart := d(2026, time.August, 28)
	const weekly = "FREQ=WEEKLY"

	// A daily habit anchored on 2026-09-01.
	dailyStart := d(2026, time.September, 1)
	const daily = "FREQ=DAILY"

	// Mon/Wed/Fri, anchored on Monday 2026-09-07.
	mwfStart := d(2026, time.September, 7)
	const mwf = "FREQ=WEEKLY;BYDAY=MO,WE,FR"

	tests := []struct {
		name    string
		rule    string
		dtstart domain.Date
		checked []domain.Date
		today   domain.Date
		want    int
	}{
		{
			// The headline of D5: four weeks running is 4, not 28 and not 0.
			name: "a weekly habit checked on four consecutive scheduled dates is 4",
			rule: weekly, dtstart: weeklyStart,
			checked: []domain.Date{
				d(2026, time.August, 28), d(2026, time.September, 4),
				d(2026, time.September, 11), d(2026, time.September, 18),
			},
			today: d(2026, time.September, 22), // a Tuesday: nothing scheduled
			want:  4,
		},
		{
			// Of the four occurrences, the third one counting BACK from the
			// most recent (2026-09-04) was missed. The streak is therefore the
			// two occurrences after the break: 09-11 and 09-18. Not 4 (which
			// would ignore the break) and not 3 (the number of checks).
			name: "the third occurrence counting back is missed, so the streak is 2",
			rule: weekly, dtstart: weeklyStart,
			checked: []domain.Date{
				d(2026, time.August, 28), // checked
				// 2026-09-04 missed
				d(2026, time.September, 11), d(2026, time.September, 18),
			},
			today: d(2026, time.September, 22),
			want:  2,
		},
		{
			// The same rule read the other way round: the third occurrence
			// CHRONOLOGICALLY (2026-09-11) is missed, leaving only 09-18 after
			// the break.
			name: "the third occurrence chronologically is missed, so the streak is 1",
			rule: weekly, dtstart: weeklyStart,
			checked: []domain.Date{
				d(2026, time.August, 28), d(2026, time.September, 4),
				// 2026-09-11 missed
				d(2026, time.September, 18),
			},
			today: d(2026, time.September, 22),
			want:  1,
		},
		{
			name: "the most recent occurrence is missed, so the streak is 0",
			rule: weekly, dtstart: weeklyStart,
			checked: []domain.Date{
				d(2026, time.August, 28), d(2026, time.September, 4), d(2026, time.September, 11),
				// 2026-09-18 missed
			},
			today: d(2026, time.September, 22),
			want:  0,
		},
		{
			name: "a daily habit checked five days running is 5",
			rule: daily, dtstart: dailyStart,
			checked: []domain.Date{
				d(2026, time.September, 1), d(2026, time.September, 2), d(2026, time.September, 3),
				d(2026, time.September, 4), d(2026, time.September, 5),
			},
			today: d(2026, time.September, 5),
			want:  5,
		},
		{
			name: "a daily habit with a gap counts only since the gap",
			rule: daily, dtstart: dailyStart,
			checked: []domain.Date{
				d(2026, time.September, 1), d(2026, time.September, 2),
				// 3 September missed
				d(2026, time.September, 4), d(2026, time.September, 5),
			},
			today: d(2026, time.September, 5),
			want:  2,
		},
		{
			name:    "no checks at all is 0",
			rule:    weekly,
			dtstart: weeklyStart,
			checked: nil,
			today:   d(2026, time.September, 22),
			want:    0,
		},
		{
			name: "mon/wed/fri with a mid-week miss",
			rule: mwf, dtstart: mwfStart,
			checked: []domain.Date{
				d(2026, time.September, 7),  // Mon
				d(2026, time.September, 9),  // Wed
				d(2026, time.September, 11), // Fri
				d(2026, time.September, 14), // Mon
				// Wednesday 16 September missed
				d(2026, time.September, 18), // Fri
			},
			today: d(2026, time.September, 19), // Saturday, nothing scheduled
			want:  1,
		},
		{
			name: "mon/wed/fri with no miss counts every scheduled day",
			rule: mwf, dtstart: mwfStart,
			checked: []domain.Date{
				d(2026, time.September, 7), d(2026, time.September, 9), d(2026, time.September, 11),
				d(2026, time.September, 14), d(2026, time.September, 16), d(2026, time.September, 18),
			},
			today: d(2026, time.September, 19),
			want:  6,
		},
		{
			name: "the streak stops at DTSTART rather than running off the calendar",
			rule: daily, dtstart: dailyStart,
			checked: []domain.Date{
				d(2026, time.September, 1), d(2026, time.September, 2), d(2026, time.September, 3),
			},
			today: d(2026, time.September, 3),
			want:  3,
		},
		{
			name: "a monthly habit checked three months running is 3",
			rule: "FREQ=MONTHLY;BYMONTHDAY=1", dtstart: d(2026, time.July, 1),
			checked: []domain.Date{
				d(2026, time.July, 1), d(2026, time.August, 1), d(2026, time.September, 1),
			},
			today: d(2026, time.September, 18),
			want:  3,
		},
		{
			name: "an every-other-week habit counts occurrences, not weeks",
			rule: "FREQ=WEEKLY;INTERVAL=2", dtstart: d(2026, time.August, 28),
			checked: []domain.Date{
				d(2026, time.August, 28), d(2026, time.September, 11),
			},
			today: d(2026, time.September, 18), // an off week
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.StreakOf(tt.rule, tt.dtstart, tt.checked, utcClockAt(tt.today))
			if err != nil {
				t.Fatalf("StreakOf() = %v", err)
			}
			if got != tt.want {
				t.Errorf("StreakOf() = %d, want %d", got, tt.want)
			}
		})
	}
}

// D5's second half: today's still-pending occurrence does not break a live
// streak, and checking it increments.
func TestStreakTodaysPendingOccurrenceDoesNotBreakIt(t *testing.T) {
	const daily = "FREQ=DAILY"
	dtstart := d(2026, time.September, 14)
	today := d(2026, time.September, 18)

	before := []domain.Date{
		d(2026, time.September, 14), d(2026, time.September, 15),
		d(2026, time.September, 16), d(2026, time.September, 17),
	}

	t.Run("scheduled today and not yet checked keeps yesterday's streak", func(t *testing.T) {
		got, err := domain.StreakOf(daily, dtstart, before, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 4 {
			t.Errorf("StreakOf() = %d, want 4 — a pending occurrence is not a missed one", got)
		}
	})

	t.Run("scheduled today and checked is the previous count plus one", func(t *testing.T) {
		got, err := domain.StreakOf(daily, dtstart, append(before, today), utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 5 {
			t.Errorf("StreakOf() = %d, want 5", got)
		}
	})

	t.Run("a pending occurrence today with nothing behind it is still 0", func(t *testing.T) {
		got, err := domain.StreakOf(daily, dtstart, nil, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 0 {
			t.Errorf("StreakOf() = %d, want 0", got)
		}
	})

	t.Run("a pending occurrence today does not rescue a broken streak", func(t *testing.T) {
		// Yesterday was scheduled and missed, so the streak is 0 whatever
		// today's pending occurrence is.
		checked := []domain.Date{d(2026, time.September, 14), d(2026, time.September, 15)}
		got, err := domain.StreakOf(daily, dtstart, checked, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 0 {
			t.Errorf("StreakOf() = %d, want 0", got)
		}
	})

	t.Run("today unscheduled and unchecked leaves the streak alone", func(t *testing.T) {
		// A Mon/Wed/Fri habit asked about on a Saturday.
		checked := []domain.Date{
			d(2026, time.September, 14), d(2026, time.September, 16), d(2026, time.September, 18),
		}
		got, err := domain.StreakOf("FREQ=WEEKLY;BYDAY=MO,WE,FR", d(2026, time.September, 14),
			checked, utcClockAt(d(2026, time.September, 19)))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 3 {
			t.Errorf("StreakOf() = %d, want 3", got)
		}
	})
}

// THE DOCUMENTED RULE for an off-schedule check: it does not count towards the
// streak, and it does not repair a break either. D5 counts consecutive
// SCHEDULED occurrences, and a date that was never scheduled is not one.
func TestStreakIgnoresChecksOnUnscheduledDates(t *testing.T) {
	// Mon/Wed/Fri anchored on Monday 2026-09-14.
	const mwf = "FREQ=WEEKLY;BYDAY=MO,WE,FR"
	dtstart := d(2026, time.September, 14)
	today := d(2026, time.September, 19) // Saturday, nothing scheduled

	t.Run("an extra check on an unscheduled day does not count", func(t *testing.T) {
		withExtra := []domain.Date{
			d(2026, time.September, 14), // Mon, scheduled
			d(2026, time.September, 15), // Tue, NOT scheduled
			d(2026, time.September, 16), // Wed, scheduled
			d(2026, time.September, 18), // Fri, scheduled
		}
		got, err := domain.StreakOf(mwf, dtstart, withExtra, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 3 {
			t.Errorf("StreakOf() = %d, want 3 — the Tuesday tick is not an occurrence", got)
		}
	})

	t.Run("an unscheduled check does not repair a break", func(t *testing.T) {
		// Wednesday, a scheduled day, was missed; Thursday was ticked instead.
		checked := []domain.Date{
			d(2026, time.September, 14), // Mon, scheduled
			d(2026, time.September, 17), // Thu, NOT scheduled — the make-up tick
			d(2026, time.September, 18), // Fri, scheduled
		}
		got, err := domain.StreakOf(mwf, dtstart, checked, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 1 {
			t.Errorf("StreakOf() = %d, want 1 — Wednesday was missed and Thursday does not stand in", got)
		}
	})

	t.Run("checks entirely before DTSTART do not count", func(t *testing.T) {
		checked := []domain.Date{
			d(2026, time.September, 7), d(2026, time.September, 9), d(2026, time.September, 11),
		}
		got, err := domain.StreakOf(mwf, dtstart, checked, utcClockAt(today))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if got != 0 {
			t.Errorf("StreakOf() = %d, want 0", got)
		}
	})
}

func TestStreakOnANode(t *testing.T) {
	habit := habitNode("FREQ=WEEKLY", d(2026, time.August, 28))
	today := utcClockAt(d(2026, time.September, 22))

	t.Run("it reads the node's recurrence and creation date", func(t *testing.T) {
		checks := checksOn("h1",
			d(2026, time.August, 28), d(2026, time.September, 4),
			d(2026, time.September, 11), d(2026, time.September, 18))
		got, err := domain.Streak(habit, checks, today)
		if err != nil {
			t.Fatalf("Streak() = %v", err)
		}
		if got != 4 {
			t.Errorf("Streak() = %d, want 4", got)
		}
	})

	t.Run("checks belonging to other nodes are ignored", func(t *testing.T) {
		checks := append(
			checksOn("h1", d(2026, time.September, 11), d(2026, time.September, 18)),
			checksOn("someone-else",
				d(2026, time.August, 28), d(2026, time.September, 4))...)
		got, err := domain.Streak(habit, checks, today)
		if err != nil {
			t.Fatalf("Streak() = %v", err)
		}
		if got != 2 {
			t.Errorf("Streak() = %d, want 2 — another habit's checks must not count", got)
		}
	})

	t.Run("no checks at all", func(t *testing.T) {
		got, err := domain.Streak(habit, nil, today)
		if err != nil {
			t.Fatalf("Streak() = %v", err)
		}
		if got != 0 {
			t.Errorf("Streak() = %d, want 0", got)
		}
	})
}

func TestStreakErrors(t *testing.T) {
	today := utcClockAt(d(2026, time.September, 18))

	t.Run("a habit with no recurrence is a matchable error", func(t *testing.T) {
		n := nd("h1", "", domain.NodeTypeHabit, domain.StatusBacklog)
		n.Recurrence = nil

		got, err := domain.Streak(n, nil, today)
		if !errors.Is(err, domain.ErrNoRecurrence) {
			t.Fatalf("Streak() = %d, %v; want ErrNoRecurrence", got, err)
		}
		if got != 0 {
			t.Errorf("Streak() = %d alongside the error, want 0", got)
		}
	})

	t.Run("an unsupported recurrence is a matchable error, not a streak of 0", func(t *testing.T) {
		n := habitNode("FREQ=YEARLY", d(2026, time.August, 28))
		if _, err := domain.Streak(n, nil, today); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
			t.Errorf("err = %v, want ErrUnsupportedRecurrence", err)
		}
		if _, err := domain.StreakOf("nonsense", d(2026, time.August, 28), nil, today); !errors.Is(err, domain.ErrUnsupportedRecurrence) {
			t.Errorf("err = %v, want ErrUnsupportedRecurrence", err)
		}
	})

	t.Run("a node with no created_at cannot be anchored", func(t *testing.T) {
		n := habitNode("FREQ=DAILY", d(2026, time.August, 28))
		n.CreatedAt = time.Time{}

		if _, err := domain.Streak(n, nil, today); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("err = %v, want ErrInvalid", err)
		}
	})
}

// The streak is a pure function of the checks, the rule and the injected clock.
func TestStreakIsDeterministicAndReadsNoWallClock(t *testing.T) {
	checked := []domain.Date{
		d(2026, time.September, 14), d(2026, time.September, 15), d(2026, time.September, 16),
	}

	first, err := domain.StreakOf("FREQ=DAILY", d(2026, time.September, 14), checked,
		utcClockAt(d(2026, time.September, 16)))
	if err != nil {
		t.Fatalf("StreakOf() = %v", err)
	}
	for range 5 {
		again, err := domain.StreakOf("FREQ=DAILY", d(2026, time.September, 14), checked,
			utcClockAt(d(2026, time.September, 16)))
		if err != nil {
			t.Fatalf("StreakOf() = %v", err)
		}
		if again != first {
			t.Fatalf("StreakOf() = %d then %d on the same input", first, again)
		}
	}

	// Moving the injected clock forward past an unchecked occurrence breaks it,
	// which proves the clock is the only thing that moved.
	later, err := domain.StreakOf("FREQ=DAILY", d(2026, time.September, 14), checked,
		utcClockAt(d(2026, time.September, 18)))
	if err != nil {
		t.Fatalf("StreakOf() = %v", err)
	}
	if later != 0 {
		t.Errorf("StreakOf(two days later) = %d, want 0 — 17 September passed unchecked", later)
	}
}

// A long streak terminates: the walk is bounded by the number of checks.
func TestStreakOverALongHistoryTerminates(t *testing.T) {
	dtstart := d(2020, time.January, 1)
	today := d(2026, time.September, 18)

	all, err := domain.Occurrences("FREQ=DAILY", dtstart, dtstart, today)
	if err != nil {
		t.Fatalf("Occurrences() = %v", err)
	}

	got, err := domain.StreakOf("FREQ=DAILY", dtstart, all, utcClockAt(today))
	if err != nil {
		t.Fatalf("StreakOf() = %v", err)
	}
	if got != len(all) {
		t.Errorf("StreakOf() = %d, want %d", got, len(all))
	}
}
