package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"nexus/internal/domain"

	sqlite3 "modernc.org/sqlite/lib"
)

// ErrTimerRunning reports that a timer is already running: the attempt to open
// a second time entry was refused by the `one_open_timer` partial unique index
// that migration 0002 puts on time_entries.
//
// The rejection is surfaced as a sentinel rather than as the driver's "UNIQUE
// constraint failed: time_entries" for two reasons. The message names an index
// expression, which tells a user nothing. And the service layer's
// single-active policy (S1-19 — close the running entry, then open the new
// one) has to be able to recognise this exact condition and act on it, which it
// cannot do by matching prose.
//
// The repository reports; it does not decide. Closing the previous entry so
// that a new one may start is a policy, it belongs to TimerService, and it is
// deliberately not implemented here.
var ErrTimerRunning = errors.New("store: a timer is already running")

// ErrEntryClosed reports that a time entry has already been stopped. Closing an
// entry twice is not idempotent the way attaching a tag twice is: the second
// call carries a different end time, and silently ignoring it would discard a
// correction or, worse, hide the fact that two code paths both believe they own
// the timer.
var ErrEntryClosed = errors.New("store: the time entry is already closed")

// timeEntryColumns is the SELECT list of `time_entries`, qualified for joins.
const timeEntryColumns = "time_entries.id, time_entries.node_id, time_entries.started_at, time_entries.ended_at"

// TimeEntryRepo reads and writes `time_entries`.
//
// It is thin in the same way NodeRepo is: it opens, closes and lists entries,
// and it computes no durations and no totals. domain.TimeEntry.Duration is the
// one implementation of "how long did this take", and it takes an injected
// clock precisely so that nothing below the service layer has to read one.
type TimeEntryRepo struct {
	exec Executor
}

// NewTimeEntryRepo returns a repository over exec. Run Migrate first.
func NewTimeEntryRepo(exec Executor) *TimeEntryRepo { return &TimeEntryRepo{exec: exec} }

// WithExecutor returns a copy bound to exec, which is what lets S1-19 close the
// running entry and open the next one inside a single transaction — the two
// halves of "move a card to Doing" must not be able to land separately.
func (r *TimeEntryRepo) WithExecutor(exec Executor) *TimeEntryRepo {
	return &TimeEntryRepo{exec: exec}
}

func scanTimeEntry(s rowScanner) (domain.TimeEntry, error) {
	var (
		e         domain.TimeEntry
		startedAt string
		endedAt   sql.NullString
	)
	if err := s.Scan(&e.ID, &e.NodeID, &startedAt, &endedAt); err != nil {
		return domain.TimeEntry{}, err
	}

	var err error
	if e.StartedAt, err = parseTimestamp(startedAt); err != nil {
		return domain.TimeEntry{}, fmt.Errorf("store: time entry %q: started_at: %w", e.ID, err)
	}
	if e.EndedAt, err = parseNullTimestamp(endedAt); err != nil {
		return domain.TimeEntry{}, fmt.Errorf("store: time entry %q: ended_at: %w", e.ID, err)
	}
	return e, nil
}

// Open starts a timer on nodeID at startedAt and returns the entry it wrote.
//
// At most one entry may be open across the whole database. That is the schema's
// rule — the `one_open_timer` partial unique index — and this method's only job
// where the invariant is concerned is to report the refusal as ErrTimerRunning
// instead of as a driver string. It does not close the other entry first;
// whether a running timer should be stopped so that this one may start is the
// timer service's policy, not the repository's.
func (r *TimeEntryRepo) Open(ctx context.Context, id, nodeID string, startedAt time.Time) (domain.TimeEntry, error) {
	const stmt = `INSERT INTO time_entries (id, node_id, started_at, ended_at) VALUES (?, ?, ?, NULL)`

	if _, err := r.exec.ExecContext(ctx, stmt, id, nodeID, formatTimestamp(startedAt)); err != nil {
		if isOpenTimerConflict(err) {
			return domain.TimeEntry{}, fmt.Errorf("store: opening a timer on node %q: %w (%w)",
				nodeID, ErrTimerRunning, wrapExec("opening a timer", err))
		}
		return domain.TimeEntry{}, wrapExec(fmt.Sprintf("opening a timer on node %q", nodeID), err)
	}

	return domain.TimeEntry{ID: id, NodeID: nodeID, StartedAt: startedAt.UTC()}, nil
}

// isOpenTimerConflict reports whether err is the one_open_timer index refusing
// a second running entry.
//
// The index is the only UNIQUE constraint on time_entries other than the
// primary key, so a UNIQUE violation that is not a duplicate id is this one.
// The primary key reports its own extended code — SQLITE_CONSTRAINT_PRIMARYKEY
// — which is what keeps the two apart without reading the message.
func isOpenTimerConflict(err error) bool {
	return isConstraintCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE)
}

// Close stops the entry at endedAt.
//
// An entry that is already closed is ErrEntryClosed and an id that does not
// exist is ErrNotFound; the two are told apart with a second query rather than
// conflated, because "you stopped it twice" and "there is no such entry" call
// for different handling upstream.
//
// The schema does not check that ended_at is after started_at —
// domain.TimeEntry.Validate does — so the repository writes what it is given.
func (r *TimeEntryRepo) Close(ctx context.Context, id string, endedAt time.Time) error {
	const stmt = `UPDATE time_entries SET ended_at = ? WHERE id = ? AND ended_at IS NULL`

	result, err := r.exec.ExecContext(ctx, stmt, formatTimestamp(endedAt), id)
	if err != nil {
		return wrapExec(fmt.Sprintf("closing time entry %q", id), err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: closing time entry %q: %w", id, err)
	}
	if affected == 1 {
		return nil
	}

	// Nothing changed: either the row is not there, or it is there and already
	// closed. Ask which.
	if _, getErr := r.Get(ctx, id); getErr != nil {
		return getErr
	}
	return fmt.Errorf("store: closing time entry %q: %w", id, ErrEntryClosed)
}

// CloseAll stops every open entry at endedAt and reports how many it closed.
//
// There is at most one, by construction, so the count is 0 or 1. The method
// exists anyway because its callers — the timer service, and later the
// sleep/lock handler, which has to stop the clock without knowing whether one
// was running — want "make sure nothing is running" rather than "stop this id",
// and a caller that had to look the entry up first would have a race between
// the lookup and the close.
func (r *TimeEntryRepo) CloseAll(ctx context.Context, endedAt time.Time) (int, error) {
	const stmt = `UPDATE time_entries SET ended_at = ? WHERE ended_at IS NULL`

	result, err := r.exec.ExecContext(ctx, stmt, formatTimestamp(endedAt))
	if err != nil {
		return 0, wrapExec("closing all open timers", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("store: closing all open timers: %w", err)
	}
	return int(affected), nil
}

// Get returns one entry by id, or ErrNotFound.
func (r *TimeEntryRepo) Get(ctx context.Context, id string) (domain.TimeEntry, error) {
	row := r.exec.QueryRowContext(ctx,
		"SELECT "+timeEntryColumns+" FROM time_entries WHERE time_entries.id = ?", id)

	e, err := scanTimeEntry(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.TimeEntry{}, fmt.Errorf("store: time entry %q: %w", id, ErrNotFound)
	case err != nil:
		return domain.TimeEntry{}, fmt.Errorf("store: reading time entry %q: %w", id, err)
	}
	return e, nil
}

// OpenEntry returns the single globally open entry, or ErrNotFound when no
// timer is running. There cannot be two: the schema does not allow it.
func (r *TimeEntryRepo) OpenEntry(ctx context.Context) (domain.TimeEntry, error) {
	row := r.exec.QueryRowContext(ctx,
		"SELECT "+timeEntryColumns+" FROM time_entries WHERE time_entries.ended_at IS NULL")

	e, err := scanTimeEntry(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.TimeEntry{}, fmt.Errorf("store: no timer is running: %w", ErrNotFound)
	case err != nil:
		return domain.TimeEntry{}, fmt.Errorf("store: reading the open time entry: %w", err)
	}
	return e, nil
}

// ListByNode returns every entry on one node, newest first, open entry included.
func (r *TimeEntryRepo) ListByNode(ctx context.Context, nodeID string) ([]domain.TimeEntry, error) {
	const query = "SELECT " + timeEntryColumns + ` FROM time_entries
	               WHERE time_entries.node_id = ?
	               ORDER BY time_entries.started_at DESC, time_entries.id`
	return r.queryEntries(ctx, query, nodeID)
}

// ListByDay returns every entry that STARTED on date, oldest first.
//
// # The bucketing rule, stated because Stage 6 depends on it
//
// An entry belongs to the day its started_at falls on, and to that day only. An
// entry that runs from 23:30 to 00:30 is wholly a Tuesday entry; it is not
// split, and none of it is reported on Wednesday. That is what the PMP timelog
// wants — a line of work is logged on the day it was begun — and splitting
// would turn one 60-minute line into two half-hour ones that no form has a
// place for.
//
// # The zone, which is the part that is easy to get wrong
//
// started_at is stored in UTC and date is a calendar date in the user's own
// zone, so the day's boundaries are converted, not assumed. The bounds are
// computed in Go from an explicit *time.Location rather than in SQL, because
// SQLite's date functions only know 'utc' and 'localtime' — the process's local
// zone — and a query that silently means "the machine's timezone" is a query
// that reports a different day to a user who travels.
//
// The range is half-open, [start of date, start of the next date), so the last
// second of the day belongs to the day and midnight belongs to the next one.
func (r *TimeEntryRepo) ListByDay(ctx context.Context, date domain.Date, loc *time.Location) ([]domain.TimeEntry, error) {
	if loc == nil {
		loc = time.UTC
	}

	start := time.Date(date.Year, date.Month, date.Day, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 1)

	const query = "SELECT " + timeEntryColumns + ` FROM time_entries
	               WHERE time_entries.started_at >= ? AND time_entries.started_at < ?
	               ORDER BY time_entries.started_at, time_entries.id`
	return r.queryEntries(ctx, query, formatTimestamp(start), formatTimestamp(end))
}

// queryEntries runs a SELECT of timeEntryColumns and scans every row. The
// result is never nil.
func (r *TimeEntryRepo) queryEntries(ctx context.Context, query string, args ...any) ([]domain.TimeEntry, error) {
	rows, err := r.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("store: listing time entries: %w", err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []domain.TimeEntry{}
	for rows.Next() {
		e, err := scanTimeEntry(rows)
		if err != nil {
			return nil, fmt.Errorf("store: listing time entries: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing time entries: %w", err)
	}
	return out, nil
}
