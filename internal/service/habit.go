package service

import (
	"context"
	"errors"
	"fmt"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// ErrNotAHabit reports that the node is not a habit.
//
// Checking off a task, or asking a project for its streak, is a call that has
// gone to the wrong place: habits are the only rows with a recurrence and a
// habit_checks history, and answering anything at all for the others — 0, false,
// an empty list — would make a caller's bug look like data.
var ErrNotAHabit = errors.New("service: the node is not a habit")

// HabitService is the habit strip's whole read and write path: check a day,
// uncheck it, ask for the streak, ask whether today is scheduled.
//
// # Habits are not on the board
//
// Habits never appear in Kanban columns (D2), so this service — not the board —
// is where they are read. The board's half of that rule is enforced in the read
// path (S1-21), which filters them out of every column.
//
// # No streak arithmetic in this file
//
// A streak is consecutive SCHEDULED occurrences of the node's RRULE that were
// checked (D5) — not calendar days, and today's still-pending occurrence does
// not break it. That is domain.Streak, called with the injected clock. This file
// loads the checks and the rule and hands them over; if a number were computed
// here there would be two implementations of D5 and no way to notice them
// diverging.
type HabitService struct {
	nodes  *store.NodeRepo
	checks *store.HabitCheckRepo
	clock  Clock
}

// NewHabitService returns a habit service over the repositories and the clock.
func NewHabitService(nodes *store.NodeRepo, checks *store.HabitCheckRepo, clock Clock) *HabitService {
	return &HabitService{nodes: nodes, checks: checks, clock: clock}
}

// habit loads nodeID and refuses anything that is not a habit.
func (s *HabitService) habit(ctx context.Context, nodeID string) (domain.Node, error) {
	n, err := s.nodes.Get(ctx, nodeID)
	if err != nil {
		return domain.Node{}, err
	}
	if n.Type != domain.NodeTypeHabit {
		return domain.Node{}, fmt.Errorf("service: node %q is a %s: %w", nodeID, n.Type, ErrNotAHabit)
	}
	return n, nil
}

// Check records that the habit was done on date.
//
// Checking the same day twice is the same fact, not a second one, so it is a
// no-op rather than an error — the UI will produce a double tap sooner or later.
func (s *HabitService) Check(ctx context.Context, nodeID string, date domain.Date) error {
	if _, err := s.habit(ctx, nodeID); err != nil {
		return err
	}
	return s.checks.Check(ctx, nodeID, date)
}

// Uncheck removes the check on date. Unchecking a day that was never checked is
// a no-op, for the same reason.
func (s *HabitService) Uncheck(ctx context.Context, nodeID string, date domain.Date) error {
	if _, err := s.habit(ctx, nodeID); err != nil {
		return err
	}
	return s.checks.Uncheck(ctx, nodeID, date)
}

// IsChecked reports whether the habit was checked on date.
func (s *HabitService) IsChecked(ctx context.Context, nodeID string, date domain.Date) (bool, error) {
	if _, err := s.habit(ctx, nodeID); err != nil {
		return false, err
	}
	return s.checks.IsChecked(ctx, nodeID, date)
}

// Streak returns how many consecutive scheduled occurrences of the habit's rule
// were checked, ending at the present (D5).
//
// A habit with no recurrence is invalid — CreateNode refuses to make one — and
// here it is an error wrapping domain.ErrNoRecurrence rather than a 0. Zero and
// "cannot say" are different answers: the first tells the user they broke a
// streak, the second tells them the habit is misconfigured, and the strip
// renders them differently.
func (s *HabitService) Streak(ctx context.Context, nodeID string) (int, error) {
	n, err := s.habit(ctx, nodeID)
	if err != nil {
		return 0, err
	}
	if n.Recurrence == nil {
		return 0, fmt.Errorf("service: streak of %q: %w", nodeID, domain.ErrNoRecurrence)
	}

	// The schedule is anchored at the day the habit was created, so nothing
	// before that can be an occurrence and nothing after today can be counted.
	from := domain.DateOf(n.CreatedAt)
	to := domain.Today(s.clock)

	checks, err := s.checks.ChecksForNode(ctx, nodeID, from, to)
	if err != nil {
		return 0, err
	}
	return domain.Streak(n, checks, s.clock)
}

// DueToday reports whether today is a scheduled occurrence of the habit — what
// the habit strip needs to know to show the day at all.
func (s *HabitService) DueToday(ctx context.Context, nodeID string) (bool, error) {
	return s.DueOn(ctx, nodeID, domain.Today(s.clock))
}

// DueOn reports whether date is a scheduled occurrence of the habit. DueToday is
// this with the injected clock's today, and the calendar view will want other
// days.
func (s *HabitService) DueOn(ctx context.Context, nodeID string, date domain.Date) (bool, error) {
	n, err := s.habit(ctx, nodeID)
	if err != nil {
		return false, err
	}
	if n.Recurrence == nil {
		return false, fmt.Errorf("service: schedule of %q: %w", nodeID, domain.ErrNoRecurrence)
	}

	rule, err := domain.ParseRecurrence(*n.Recurrence)
	if err != nil {
		return false, err
	}
	return rule.Matches(domain.DateOf(n.CreatedAt), date), nil
}
