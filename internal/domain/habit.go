package domain

import "strings"

// HabitCheck records that a habit was ticked off on a given day.
//
// It is keyed by date and not by instant on purpose: a day is either done or it
// is not, and checking twice is the same fact rather than two facts — which is
// what the composite primary key on habit_checks says in SQL. Streaks are
// therefore counted in scheduled occurrences (D5), never in hours.
type HabitCheck struct {
	NodeID string
	Date   Date
}

// Validate reports the first field of c that holds a forbidden value.
func (c HabitCheck) Validate() error {
	const entity = "habit_check"

	if strings.TrimSpace(c.NodeID) == "" {
		return invalid(entity, "node_id", "must not be empty")
	}
	if c.Date.IsZero() {
		return invalid(entity, "date", "must be set")
	}
	if !c.Date.Valid() {
		return invalid(entity, "date", "%s is not a real calendar date", c.Date)
	}
	return nil
}
