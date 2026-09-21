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

// isHabit reports whether n is one of the rows this service answers for.
//
// One spelling, asked of the TYPE: the per-node methods below use it to refuse
// everything else, and Strip uses it to decide who is on the strip. A membership
// test written out a second time in Strip — or, worse, in TypeScript as
// `type === 'habit'` over the tree — is the defect class this project has paid
// for three times.
func isHabit(n domain.Node) bool { return n.Type == domain.NodeTypeHabit }

// habit loads nodeID and refuses anything that is not a habit.
func (s *HabitService) habit(ctx context.Context, nodeID string) (domain.Node, error) {
	n, err := s.nodes.Get(ctx, nodeID)
	if err != nil {
		return domain.Node{}, err
	}
	if !isHabit(n) {
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

// CheckToday records that the habit was done TODAY — where today is
// domain.Today(s.clock), the injected clock's local calendar day.
//
// # Why this exists rather than the caller naming the day
//
// It is the same day HabitView.CheckedToday and HabitView.ScheduledToday are
// derived against in Strip: the same clock, read in the same process. A caller
// that worked out its own "today" and passed it to Check would be a SECOND
// implementation of the rule, and across a local midnight boundary the two
// disagree — the check lands on a day the user never chose while the strip
// comes back still unchecked, so the tick appears to do nothing at all.
//
// The dated Check above stays exactly as it is. The calendar (Stage 7) ticks
// days that are genuinely not today, and that is a day the caller CHOSE rather
// than one it computed. DueToday/DueOn is the same pair, one question along.
func (s *HabitService) CheckToday(ctx context.Context, nodeID string) error {
	return s.Check(ctx, nodeID, domain.Today(s.clock))
}

// UncheckToday removes today's check, today being the injected clock's, for
// CheckToday's reason.
func (s *HabitService) UncheckToday(ctx context.Context, nodeID string) error {
	return s.Uncheck(ctx, nodeID, domain.Today(s.clock))
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
	return scheduledOn(n, date)
}

// scheduledOn reports whether date is a scheduled occurrence of n's rule.
//
// It takes the node rather than an id because the strip has already loaded
// fifty of them and must not go back to the database per habit. DueOn is this
// function with the load in front of it, so "is this day scheduled?" has one
// answer whichever door asks it.
func scheduledOn(n domain.Node, date domain.Date) (bool, error) {
	if n.Recurrence == nil {
		return false, fmt.Errorf("service: schedule of %q: %w", n.ID, domain.ErrNoRecurrence)
	}

	rule, err := domain.ParseRecurrence(*n.Recurrence)
	if err != nil {
		return false, err
	}
	return rule.Matches(domain.DateOf(n.CreatedAt), date), nil
}

// Strip is the habits strip's whole read: every non-archived habit, in
// sort_order then id, each with everything the strip draws (S2-04).
//
// # Why this exists at all
//
// Board() excludes habits by design (D2) and every other method here is per
// node, so without this the frontend would list the tree, filter it by
// `type === 'habit'` and then ask three more questions per habit. That is a
// membership rule in TypeScript, an N+1, and a streak computed away from D5's
// single implementation. Everything the strip needs arrives here, computed in
// Go, in a fixed number of queries.
//
// # Fixed cost
//
// Two queries, whatever the habit count: one for the nodes and one for the
// checks, then an index by node id — the same shape loadSnapshot uses for tags.
// A habit's whole check history is read (from the earliest habit's creation day
// to today) because that is what a streak walks backwards through.
//
// # What is on the strip
//
//   - membership is isHabit, the same type question the per-node methods ask;
//   - a habit NESTED under a project appears exactly once, like any other. It
//     contributes nothing to that project (D10), which is a different question
//     with a different answer;
//   - archived habits are excluded, as everywhere else;
//   - habits that are not scheduled today are INCLUDED, carrying
//     ScheduledToday = false. Whether the strip draws them is presentation.
func (s *HabitService) Strip(ctx context.Context) ([]HabitView, error) {
	// ListAll is already ordered by (sort_order, id) and already excludes
	// archived rows; neither ordering nor the archive filter is re-implemented
	// here.
	nodes, err := s.nodes.ListAll(ctx, false)
	if err != nil {
		return nil, err
	}

	habits := make([]domain.Node, 0, len(nodes))
	for _, n := range nodes {
		if isHabit(n) {
			habits = append(habits, n)
		}
	}

	today := domain.Today(s.clock)
	checks, err := s.checkIndex(ctx, habits, today)
	if err != nil {
		return nil, err
	}

	out := make([]HabitView, 0, len(habits))
	for _, n := range habits {
		// A misconfigured habit is an error, not a zero: 0 reads as "you broke
		// your streak" and an empty strip reads as "you have no habits", and
		// both are worse than being told the row is wrong.
		scheduled, err := scheduledOn(n, today)
		if err != nil {
			return nil, fmt.Errorf("service: habit strip: %w", err)
		}
		streak, err := domain.Streak(n, checks[n.ID], s.clock)
		if err != nil {
			return nil, fmt.Errorf("service: habit strip: %w", err)
		}

		out = append(out, HabitView{
			Node:           n,
			ScheduledToday: scheduled,
			CheckedToday:   checkedOn(checks[n.ID], today),
			Streak:         streak,
		})
	}
	return out, nil
}

// checkIndex reads the checks of every habit in one query and groups them by
// node id.
//
// The window starts at the earliest habit's creation day, which is where the
// earliest schedule can begin, and ends today. With no habits there is nothing
// to read and no query is issued.
func (s *HabitService) checkIndex(ctx context.Context, habits []domain.Node, today domain.Date) (map[string][]domain.HabitCheck, error) {
	index := map[string][]domain.HabitCheck{}
	if len(habits) == 0 {
		return index, nil
	}

	from := domain.DateOf(habits[0].CreatedAt)
	for _, n := range habits[1:] {
		if created := domain.DateOf(n.CreatedAt); created.Before(from) {
			from = created
		}
	}

	all, err := s.checks.ChecksInRange(ctx, from, today)
	if err != nil {
		return nil, err
	}
	for _, c := range all {
		index[c.NodeID] = append(index[c.NodeID], c)
	}
	return index, nil
}

// checkedOn reports whether today's check is among the ones already loaded.
//
// It is a lookup in rows the strip is holding anyway, not a second spelling of
// IsChecked: the question is the same and so is the table, but asking the
// database again per habit is the N+1 Strip exists to avoid.
func checkedOn(checks []domain.HabitCheck, date domain.Date) bool {
	for _, c := range checks {
		if c.Date == date {
			return true
		}
	}
	return false
}
