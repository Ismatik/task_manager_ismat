package domain

import (
	"errors"
	"fmt"
	"time"
)

// ErrNoRecurrence is returned when a streak is asked for on a node that has no
// recurrence rule. A streak counts scheduled occurrences (D5), so without a
// schedule there is nothing to count — and answering 0 would be a number the
// user would read as "you broke it" rather than as "this habit is misconfigured".
var ErrNoRecurrence = errors.New("domain: a habit requires a recurrence rule")

// Streak returns how many consecutive scheduled occurrences of n's recurrence
// rule were checked, ending at the present (D5).
//
// # It counts occurrences, not days
//
// This is the part that is not the obvious implementation. A FREQ=WEEKLY habit
// checked four weeks running has a streak of 4 — not 4 days, not 28, and
// emphatically not 0 because six days were "missed" every week. The six days in
// between were never scheduled, so there was nothing to miss.
//
// # What breaks it
//
// Walking backwards from today, the streak ends at the first SCHEDULED
// occurrence that passed unchecked. Unscheduled days are invisible to the walk.
//
// # Today does not break it
//
// A habit scheduled for today and not yet ticked keeps yesterday's streak: the
// day is not over. Ticking it increments. Without this rule every streak in the
// app would read as broken every morning until the user got round to it, which
// is precisely when the number is supposed to be encouraging.
//
// # A check on an unscheduled date does not count
//
// The user can tick a habit on a whim, and habit_checks will happily store it —
// it is keyed by (node_id, date) and knows nothing about schedules. It does not
// contribute to the streak and it does not repair a break, because D5 counts
// consecutive SCHEDULED occurrences and that date was not one. Nor does it do
// any harm: the walk only ever looks at scheduled dates.
//
// Checks belonging to other nodes are ignored, so the caller may pass an
// unfiltered set. DTSTART is the day n was created, taken in the stored
// timestamp's own location.
//
// Any node with a recurrence rule may be asked; the type is not checked, so a
// recurring task would work the same way.
func Streak(n Node, checks []HabitCheck, now func() time.Time) (int, error) {
	if n.Recurrence == nil {
		return 0, fmt.Errorf("domain: streak of %q: %w", n.ID, ErrNoRecurrence)
	}
	if n.CreatedAt.IsZero() {
		return 0, invalid("node", "created_at", "must be set to anchor a habit's schedule")
	}

	dates := make([]Date, 0, len(checks))
	for _, c := range checks {
		if c.NodeID == n.ID {
			dates = append(dates, c.Date)
		}
	}
	return StreakOf(*n.Recurrence, DateOf(n.CreatedAt), dates, now)
}

// StreakOf is Streak with the rule, the anchor and the checked dates given
// explicitly. See Streak for the rules it implements.
func StreakOf(rule string, dtstart Date, checked []Date, now func() time.Time) (int, error) {
	r, err := ParseRecurrence(rule)
	if err != nil {
		return 0, err
	}

	ticked := make(map[Date]bool, len(checked))
	for _, d := range checked {
		ticked[d] = true
	}

	today := Today(now)
	count := 0

	// The upper bound of the backwards walk. Previous is exclusive, so starting
	// one day after today makes today itself the first candidate.
	cursor := today.AddDays(1)

	if r.Matches(dtstart, today) {
		if ticked[today] {
			count++
		}
		// Checked or not, today is settled: an unchecked occurrence that has
		// not finished yet is pending, not missed, so the walk simply steps
		// over it instead of stopping.
		cursor = today
	}

	// The loop is bounded by the number of checks: it only continues while it
	// keeps finding ticked occurrences, and there are finitely many of those.
	for {
		prev, ok := r.Previous(dtstart, cursor)
		if !ok || !ticked[prev] {
			return count, nil
		}
		count++
		cursor = prev
	}
}
