package domain

import "time"

// DueUpdate is the (due, due_source) pair a column move or a user edit
// produces. It is a value rather than a mutation so that the two ways of
// changing a due date cannot be confused for one another at a call site: you
// ask DueForColumnMove or DueForUserEdit for the pair, then write it.
type DueUpdate struct {
	Due       *Date
	DueSource DueSource
}

// ApplyTo returns a copy of n carrying the update.
func (u DueUpdate) ApplyTo(n Node) Node {
	n.Due = copyDate(u.Due)
	n.DueSource = u.DueSource
	return n
}

// copyDate returns an independent copy of d, so that no two nodes ever end up
// sharing a *Date that one of them could edit under the other.
func copyDate(d *Date) *Date {
	if d == nil {
		return nil
	}
	c := *d
	return &c
}

// Today returns the calendar date the injected clock is currently on.
//
// # Which zone
//
// The user's. now is the injected now func() time.Time; the service hands down
// a clock reading the system wall clock, which carries the local zone, and
// DateOf takes the calendar day in the instant's own location without
// converting. That is deliberate: a task due "today" means the user's today, and folding
// through UTC first would move every evening east of Greenwich into tomorrow —
// a card that goes overdue while the user is still looking at it.
//
// The one place UTC is used is the arithmetic inside Date (AddDays, Weekday),
// where a day is always exactly 24 hours and cannot be bent by a daylight-saving
// transition.
func Today(now func() time.Time) Date { return DateOf(now()) }

// UpcomingFriday returns the Friday of the current week-ahead window (D1).
//
// If today IS a Friday the answer is today, not next week: "this week" means the
// end of the week the user is standing in, and pushing a Friday card seven days
// out would quietly turn "this week" into "next week" every Friday. Today,
// 2026-09-18, is a Friday, so this branch is live on day one.
func UpcomingFriday(today Date) Date {
	delta := (int(time.Friday) - int(today.Weekday()) + 7) % 7
	return today.AddDays(delta)
}

// DueForColumnMove returns the due date and provenance a node must have after
// being moved to the target column (D1).
//
//   - today   → due = today, due_source = auto.
//   - week    → due = the upcoming Friday (today itself when today is Friday),
//     due_source = auto.
//   - backlog → due is cleared ONLY when due_source is auto. A date the user
//     typed survives, and its due_source stays manual.
//   - doing / done → due and due_source are left exactly as they are. These two
//     columns say something about work, not about when it is due.
//
// # A column move to Today or This-week overwrites a manual date
//
// This is the reading of D1 confirmed by the user, and it is unconditional: the
// move rules in D1 are stated without an exception for manual dates, so a drag
// to Today or This-week always writes the date AND flips due_source to auto,
// destroying whatever the user had typed there.
//
// The alternative — letting a manual date survive the move — was rejected
// because it lets the column and the date disagree: a card sitting in Today
// showing a due date of next month is a board that has stopped telling the
// truth, and there is nothing on screen to explain why that one card is
// different. The column move always wins, so the two can never contradict each
// other.
//
// The consequence is accepted rather than hidden: a hand-typed date is
// destroyed by the drag, and because the date is now auto, a later move to
// Backlog clears it entirely.
//
// # Clearing resets the provenance
//
// When a move to Backlog clears an auto date, due_source goes back to manual,
// the column default. due_source describes a date that exists; claiming "we set
// this automatically" about a date that is no longer there would be a statement
// about nothing.
//
// now is only read for the today and week targets.
func DueForColumnMove(n Node, target Status, now func() time.Time) DueUpdate {
	switch target {
	case StatusToday:
		today := Today(now)
		return DueUpdate{Due: &today, DueSource: DueSourceAuto}

	case StatusWeek:
		friday := UpcomingFriday(Today(now))
		return DueUpdate{Due: &friday, DueSource: DueSourceAuto}

	case StatusBacklog:
		if n.DueSource == DueSourceAuto {
			return DueUpdate{Due: nil, DueSource: DueSourceManual}
		}
		return DueUpdate{Due: copyDate(n.Due), DueSource: n.DueSource}

	default:
		// doing, done, and any status this function does not know about: the
		// safe answer to "what should the due date be now?" is "whatever it
		// already was".
		return DueUpdate{Due: copyDate(n.Due), DueSource: n.DueSource}
	}
}

// DueForUserEdit returns the pair for a due date the user set by hand (D1).
//
// It is always manual — including when the new date equals the old one, and
// including when the user clears it. Typing the date that a column move happened
// to pick is still the user claiming the date, and from then on a move back to
// Backlog must not take it away.
//
// This is a separate exported function from DueForColumnMove on purpose. The two
// have opposite provenance, and a single function with a boolean would make it
// possible to apply one while meaning the other by getting an argument backwards
// at one call site.
func DueForUserEdit(due *Date) DueUpdate {
	return DueUpdate{Due: copyDate(due), DueSource: DueSourceManual}
}

// IsOverdue reports whether n is past its due date (D1):
//
//	due != nil && due < today && status != done
//
// Due exactly today is NOT overdue — the day is not over. A done node is never
// overdue however old its date, because the work is finished and a red card
// would be asking for something that has already happened.
//
// n.Status is taken as given. For a parent, the caller passes the node carrying
// the status it renders in — DeriveStatus's result — because a parent whose
// children are all finished is done and must not glow red.
func IsOverdue(n Node, today Date) bool {
	return n.Due != nil && n.Due.Before(today) && n.Status != StatusDone
}
