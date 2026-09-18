package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"nexus/internal/domain"
)

// HabitCheckRepo reads and writes `habit_checks`: one row per habit per day.
//
// It stores facts and counts nothing. A streak is consecutive scheduled
// occurrences of the habit's RRULE that were checked (D5), which needs the
// recurrence rule and a calendar and is therefore domain.Streak's job, working
// on the checks this returns. There is no SQL in here that counts anything in a
// row, and there must not be: a streak computed in SQL would be a second
// implementation of D5 that nobody would notice diverging.
type HabitCheckRepo struct {
	exec Executor
}

// NewHabitCheckRepo returns a repository over exec. Run Migrate first.
func NewHabitCheckRepo(exec Executor) *HabitCheckRepo { return &HabitCheckRepo{exec: exec} }

// WithExecutor returns a copy bound to exec, so a check can be written in the
// same transaction as whatever else a service is doing.
func (r *HabitCheckRepo) WithExecutor(exec Executor) *HabitCheckRepo {
	return &HabitCheckRepo{exec: exec}
}

// Check records that the habit was done on date.
//
// Checking the same (node, date) twice is a no-op and not an error. A day is
// either done or it is not: the second check is the same fact, not a second
// one, which is exactly what the composite primary key in 0002 says. The UI
// will produce a double check — a double tap, an undo that is retried — and
// refusing it would make the app argue with the user about something that is
// already true.
func (r *HabitCheckRepo) Check(ctx context.Context, nodeID string, date domain.Date) error {
	const stmt = `INSERT INTO habit_checks (node_id, date) VALUES (?, ?)
	              ON CONFLICT (node_id, date) DO NOTHING`

	if _, err := r.exec.ExecContext(ctx, stmt, nodeID, date.String()); err != nil {
		return wrapExec(fmt.Sprintf("checking habit %q on %s", nodeID, date), err)
	}
	return nil
}

// Uncheck removes the check. Unchecking a day that was never checked is a
// no-op, for the same reason checking twice is: the caller's intent — "this day
// is not done" — is satisfied either way.
func (r *HabitCheckRepo) Uncheck(ctx context.Context, nodeID string, date domain.Date) error {
	const stmt = `DELETE FROM habit_checks WHERE node_id = ? AND date = ?`

	if _, err := r.exec.ExecContext(ctx, stmt, nodeID, date.String()); err != nil {
		return wrapExec(fmt.Sprintf("unchecking habit %q on %s", nodeID, date), err)
	}
	return nil
}

// IsChecked reports whether the habit was checked on date.
func (r *HabitCheckRepo) IsChecked(ctx context.Context, nodeID string, date domain.Date) (bool, error) {
	const query = `SELECT 1 FROM habit_checks WHERE node_id = ? AND date = ?`

	var one int
	err := r.exec.QueryRowContext(ctx, query, nodeID, date.String()).Scan(&one)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Not an ErrNotFound: "that day is not checked" is an answer, not a
		// failure, and the caller has a bool to put it in.
		return false, nil
	case err != nil:
		return false, fmt.Errorf("store: reading habit check %q on %s: %w", nodeID, date, err)
	}
	return true, nil
}

// ChecksForNode returns the habit's checks between from and to, both ends
// INCLUSIVE, oldest first.
//
// Inclusive on both ends because the range a caller has is a calendar window —
// "this month", "the last 30 days" — and a half-open window would silently drop
// the final day, which for a streak is the day that decides whether it is still
// alive. The comparison is a plain string comparison on the 'YYYY-MM-DD' column
// 0002 defines, which sorts identically to the calendar.
func (r *HabitCheckRepo) ChecksForNode(ctx context.Context, nodeID string, from, to domain.Date) ([]domain.HabitCheck, error) {
	const query = `SELECT node_id, date FROM habit_checks
	               WHERE node_id = ? AND date >= ? AND date <= ?
	               ORDER BY date`

	rows, err := r.exec.QueryContext(ctx, query, nodeID, from.String(), to.String())
	if err != nil {
		return nil, fmt.Errorf("store: listing habit checks for %q: %w", nodeID, err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []domain.HabitCheck{}
	for rows.Next() {
		var (
			c    domain.HabitCheck
			date string
		)
		if err := rows.Scan(&c.NodeID, &date); err != nil {
			return nil, fmt.Errorf("store: listing habit checks for %q: %w", nodeID, err)
		}
		if c.Date, err = domain.ParseDate(date); err != nil {
			return nil, fmt.Errorf("store: habit check %q: date: %w", nodeID, err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing habit checks for %q: %w", nodeID, err)
	}
	return out, nil
}
