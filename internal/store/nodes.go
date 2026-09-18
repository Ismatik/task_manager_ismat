package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"nexus/internal/domain"

	sqlitedriver "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// Executor is the subset of *sql.DB and *sql.Tx that a repository needs.
//
// Every repository is built over this rather than over *sql.DB so that a
// service can run several repositories' writes inside one transaction: S1-18
// applies a cascade, a due-date write and a reorder as a single atomic change,
// and that is only possible if each of those calls can be handed the same
// *sql.Tx. A repository that closed over a *sql.DB would commit each step on
// its own and leave a half-applied drag behind on the first failure.
//
// Both *sql.DB and *sql.Tx satisfy it as they are; nothing has to wrap them.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// rowScanner is what *sql.Row and *sql.Rows have in common, so that one
// scanning function serves both the single-row and the many-row queries.
type rowScanner interface {
	Scan(dest ...any) error
}

// timestampLayout is the one timestamp representation in this project: the
// 'YYYY-MM-DDTHH:MM:SSZ' that migration 0002 documents for created_at,
// updated_at, completed_at, archived_at, started_at and ended_at.
//
// The trailing Z is a literal, not a zone directive — Go only reads "Z" as a
// zone when it is followed by 0700 or 07:00 — which is exactly right here:
// every timestamp is normalised to UTC before it is formatted, so the column
// can never hold two instants that sort in the wrong order because one of them
// carried an offset.
const timestampLayout = "2006-01-02T15:04:05Z"

// formatTimestamp renders t in the 0002 timestamp format, in UTC.
func formatTimestamp(t time.Time) string { return t.UTC().Format(timestampLayout) }

// parseTimestamp reads a 0002 timestamp back. The result is in UTC, which is
// the zone it was written in.
func parseTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(timestampLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("store: %q is not a %s timestamp: %w", s, timestampLayout, err)
	}
	return t, nil
}

// ErrConstraint reports that the database refused a row: a CHECK, a UNIQUE, a
// NOT NULL or a foreign key.
//
// The schema's constraints are the last line of defence (0002 says so in as
// many words), and a caller has to be able to tell "this value is not allowed"
// from "the disk is full" without matching on a driver's message. The driver's
// own error stays in the chain, so a log still shows which constraint fired.
var ErrConstraint = errors.New("store: the database refused the row")

// sqliteErrorCode returns the SQLite extended result code carried by err.
func sqliteErrorCode(err error) (int, bool) {
	var serr *sqlitedriver.Error
	if errors.As(err, &serr) {
		return serr.Code(), true
	}
	return 0, false
}

// isConstraintViolation reports whether err is any SQLITE_CONSTRAINT_* failure.
//
// SQLite's extended result codes are the primary code in the low byte plus a
// discriminator above it, so every constraint failure — CHECK (275), UNIQUE
// (2067), PRIMARY KEY (1555), FOREIGN KEY (787) — masks down to
// SQLITE_CONSTRAINT. Masking rather than listing them means a constraint kind
// this code has not thought of is still classified correctly.
func isConstraintViolation(err error) bool {
	code, ok := sqliteErrorCode(err)
	return ok && code&0xff == sqlite3.SQLITE_CONSTRAINT
}

// isConstraintCode reports whether err is one specific extended constraint
// code, e.g. SQLITE_CONSTRAINT_UNIQUE.
func isConstraintCode(err error, want int) bool {
	code, ok := sqliteErrorCode(err)
	return ok && code == want
}

// wrapExec turns a driver error into this package's vocabulary: a constraint
// failure additionally matches ErrConstraint, everything else is context plus
// the original.
func wrapExec(op string, err error) error {
	if isConstraintViolation(err) {
		return fmt.Errorf("store: %s: %w (%w)", op, ErrConstraint, err)
	}
	return fmt.Errorf("store: %s: %w", op, err)
}

// nodeColumnNames is every column of `nodes`, in the order 0002 declares them.
// It is the single source for the INSERT column list, the SELECT list and the
// argument order of nodeValues and scanNode, so those four cannot drift apart.
// TestNodeColumnsMatchTheSchema pins it against PRAGMA table_info.
var nodeColumnNames = []string{
	"id",
	"parent_id",
	"type",
	"title",
	"description_md",
	"status",
	"due",
	"due_source",
	"priority",
	"estimate_min",
	"recurrence",
	"activity",
	"sort_order",
	"created_at",
	"updated_at",
	"completed_at",
	"archived_at",
}

var (
	// nodeColumns is the SELECT list, qualified with the table name. The
	// qualification is not decoration: the search index joins a virtual table
	// that also has `title` and `description_md` columns, and an unqualified
	// list would become an ambiguous-column error the day that join appears.
	nodeColumns = strings.Join(prefixAll(nodeColumnNames, "nodes."), ", ")

	// insertNodeStmt inserts one full row. Column names are unqualified because
	// INSERT names columns of the table it is already targeting.
	insertNodeStmt = "INSERT INTO nodes (" + strings.Join(nodeColumnNames, ", ") + ") VALUES (" +
		strings.Join(placeholders(len(nodeColumnNames)), ", ") + ")"

	// updateNodeStmt rewrites every column except the primary key.
	updateNodeStmt = "UPDATE nodes SET " +
		strings.Join(assignments(nodeColumnNames[1:]), ", ") + " WHERE id = ?"
)

func prefixAll(names []string, prefix string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = prefix + n
	}
	return out
}

func placeholders(n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = "?"
	}
	return out
}

// idGuard builds the WHERE clause of a bulk update over ids: it restricts the
// statement to those rows AND refuses to touch any of them unless every single
// id is present.
//
// The second half is what makes "all or nothing" true rather than merely
// intended. Without it, `WHERE id IN (a, ghost)` updates a, reports one row
// affected, and leaves the caller holding a half-applied cascade plan that it
// can only undo by hand. With it, a missing id makes the count disagree, the
// WHERE matches nothing, zero rows change, and the caller gets ErrNotFound over
// a database it never modified. One statement, so it holds inside and outside a
// transaction alike.
func idGuard(ids []string) (string, []any) {
	list := strings.Join(placeholders(len(ids)), ", ")

	args := make([]any, 0, 2*len(ids)+1)
	for range 2 {
		for _, id := range ids {
			args = append(args, id)
		}
	}
	args = append(args, len(ids))

	return "id IN (" + list + ") AND (SELECT count(*) FROM nodes WHERE id IN (" + list + ")) = ?", args
}

func assignments(names []string) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = n + " = ?"
	}
	return out
}

// NodeRepo reads and writes the `nodes` table.
//
// It is deliberately thin. It maps rows to domain.Node and back and it runs
// queries; it derives no status, computes no progress, decides no due date and
// plans no cascade. Those rules live in internal/domain and are applied by
// internal/service, which is why there is not a single CASE WHEN in here that
// encodes a rule — the one CASE that does exist, in UpdateStatuses, is a value
// lookup table, not a decision.
type NodeRepo struct {
	exec Executor
}

// NewNodeRepo returns a repository over exec, which is normally the *sql.DB.
// The schema must already exist — run Migrate first.
func NewNodeRepo(exec Executor) *NodeRepo { return &NodeRepo{exec: exec} }

// WithExecutor returns a copy of the repository bound to exec, which is how a
// service runs it inside its own transaction: repo.WithExecutor(tx).
func (r *NodeRepo) WithExecutor(exec Executor) *NodeRepo { return &NodeRepo{exec: exec} }

// nodeValues flattens n into the argument order of nodeColumnNames.
//
// Nullable columns are written as SQL NULL when the Go pointer is nil, never as
// a zero value: an empty-string due date and a 1970 completed_at are both rows
// nobody can interpret afterwards.
func nodeValues(n domain.Node) []any {
	return []any{
		n.ID,
		nullString(n.ParentID),
		string(n.Type),
		n.Title,
		n.DescriptionMD,
		string(n.Status),
		nullDate(n.Due),
		string(n.DueSource),
		int(n.Priority),
		nullInt(n.EstimateMin),
		nullString(n.Recurrence),
		nullActivity(n.Activity),
		n.SortOrder,
		formatTimestamp(n.CreatedAt),
		formatTimestamp(n.UpdatedAt),
		nullTimestamp(n.CompletedAt),
		nullTimestamp(n.ArchivedAt),
	}
}

func nullString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func nullDate(d *domain.Date) any {
	if d == nil {
		return nil
	}
	return d.String()
}

func nullActivity(a *domain.Activity) any {
	if a == nil {
		return nil
	}
	return string(*a)
}

func nullTimestamp(t *time.Time) any {
	if t == nil {
		return nil
	}
	return formatTimestamp(*t)
}

// scanNode reads one row in nodeColumnNames order.
//
// A column that fails to parse is an error rather than a zero value. The
// database is the user's only copy of their data, and a due date that silently
// scans as the zero date is a due date that silently disappears from every
// view that reads it.
func scanNode(s rowScanner) (domain.Node, error) {
	var (
		n                                   domain.Node
		nodeType, status, dueSource         string
		priority                            int
		createdAt, updatedAt                string
		parentID, due, recurrence, activity sql.NullString
		completedAt, archivedAt             sql.NullString
		estimateMin                         sql.NullInt64
	)

	err := s.Scan(
		&n.ID,
		&parentID,
		&nodeType,
		&n.Title,
		&n.DescriptionMD,
		&status,
		&due,
		&dueSource,
		&priority,
		&estimateMin,
		&recurrence,
		&activity,
		&n.SortOrder,
		&createdAt,
		&updatedAt,
		&completedAt,
		&archivedAt,
	)
	if err != nil {
		return domain.Node{}, err
	}

	n.Type = domain.NodeType(nodeType)
	n.Status = domain.Status(status)
	n.DueSource = domain.DueSource(dueSource)
	n.Priority = domain.Priority(priority)

	if parentID.Valid {
		v := parentID.String
		n.ParentID = &v
	}
	if due.Valid {
		d, err := domain.ParseDate(due.String)
		if err != nil {
			return domain.Node{}, fmt.Errorf("store: node %q: due: %w", n.ID, err)
		}
		n.Due = &d
	}
	if estimateMin.Valid {
		v := int(estimateMin.Int64)
		n.EstimateMin = &v
	}
	if recurrence.Valid {
		v := recurrence.String
		n.Recurrence = &v
	}
	if activity.Valid {
		v := domain.Activity(activity.String)
		n.Activity = &v
	}

	if n.CreatedAt, err = parseTimestamp(createdAt); err != nil {
		return domain.Node{}, fmt.Errorf("store: node %q: created_at: %w", n.ID, err)
	}
	if n.UpdatedAt, err = parseTimestamp(updatedAt); err != nil {
		return domain.Node{}, fmt.Errorf("store: node %q: updated_at: %w", n.ID, err)
	}
	if n.CompletedAt, err = parseNullTimestamp(completedAt); err != nil {
		return domain.Node{}, fmt.Errorf("store: node %q: completed_at: %w", n.ID, err)
	}
	if n.ArchivedAt, err = parseNullTimestamp(archivedAt); err != nil {
		return domain.Node{}, fmt.Errorf("store: node %q: archived_at: %w", n.ID, err)
	}

	return n, nil
}

func parseNullTimestamp(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := parseTimestamp(s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Create inserts n.
//
// It does not validate: domain.Node.Validate is the caller's to run, and the
// schema's CHECK constraints are underneath it either way. A value the schema
// refuses comes back as ErrConstraint.
func (r *NodeRepo) Create(ctx context.Context, n domain.Node) error {
	if _, err := r.exec.ExecContext(ctx, insertNodeStmt, nodeValues(n)...); err != nil {
		return wrapExec(fmt.Sprintf("creating node %q", n.ID), err)
	}
	return nil
}

// Get returns the node with this id, archived or not — a caller asking for one
// row by id has already named the row it wants. A missing row is ErrNotFound.
func (r *NodeRepo) Get(ctx context.Context, id string) (domain.Node, error) {
	row := r.exec.QueryRowContext(ctx, "SELECT "+nodeColumns+" FROM nodes WHERE nodes.id = ?", id)

	n, err := scanNode(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Node{}, fmt.Errorf("store: node %q: %w", id, ErrNotFound)
	case err != nil:
		return domain.Node{}, fmt.Errorf("store: reading node %q: %w", id, err)
	}
	return n, nil
}

// Update rewrites every column of the row with n.ID from n. A row that is not
// there is ErrNotFound rather than a silent no-op.
func (r *NodeRepo) Update(ctx context.Context, n domain.Node) error {
	args := append(nodeValues(n)[1:], n.ID)

	result, err := r.exec.ExecContext(ctx, updateNodeStmt, args...)
	if err != nil {
		return wrapExec(fmt.Sprintf("updating node %q", n.ID), err)
	}
	return requireRows(result, 1, fmt.Sprintf("node %q", n.ID))
}

// Delete removes the node. Its children, tags, time entries, attachments and
// habit checks go with it: every one of those foreign keys is ON DELETE
// CASCADE and the foreign_keys pragma is on for every connection.
//
// Nexus archives rather than deletes in the UI (archived_at, never DROP), so
// this is the deliberate, permanent path.
func (r *NodeRepo) Delete(ctx context.Context, id string) error {
	result, err := r.exec.ExecContext(ctx, "DELETE FROM nodes WHERE id = ?", id)
	if err != nil {
		return wrapExec(fmt.Sprintf("deleting node %q", id), err)
	}
	return requireRows(result, 1, fmt.Sprintf("node %q", id))
}

// requireRows turns "the statement matched nothing" into ErrNotFound.
func requireRows(result sql.Result, want int64, subject string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("store: counting affected rows for %s: %w", subject, err)
	}
	if affected != want {
		return fmt.Errorf("store: %s: %d of %d rows matched: %w", subject, affected, want, ErrNotFound)
	}
	return nil
}

// archivedFilter is the WHERE fragment that hides archived rows.
//
// Archived nodes are excluded by default everywhere — every list method takes
// an explicit includeArchived, and passing false is what the board, the tree
// and search all do. Archiving is how Nexus hides things without destroying
// them, so a list that returned archived rows unasked would make the feature
// pointless.
func archivedFilter(includeArchived bool) string {
	if includeArchived {
		return ""
	}
	return " AND nodes.archived_at IS NULL"
}

// nodeOrder is the deterministic order of every list: position among siblings
// first, id as the tie-break so that two nodes sharing a sort_order never swap
// places between two runs of the same query.
const nodeOrder = " ORDER BY nodes.sort_order, nodes.id"

// ListChildren returns the direct children of parentID, or the roots when
// parentID is nil.
func (r *NodeRepo) ListChildren(ctx context.Context, parentID *string, includeArchived bool) ([]domain.Node, error) {
	if parentID == nil {
		return r.ListRoots(ctx, includeArchived)
	}

	query := "SELECT " + nodeColumns + " FROM nodes WHERE nodes.parent_id = ?" +
		archivedFilter(includeArchived) + nodeOrder
	return r.query(ctx, query, *parentID)
}

// ListRoots returns the nodes with no parent.
func (r *NodeRepo) ListRoots(ctx context.Context, includeArchived bool) ([]domain.Node, error) {
	query := "SELECT " + nodeColumns + " FROM nodes WHERE nodes.parent_id IS NULL" +
		archivedFilter(includeArchived) + nodeOrder
	return r.query(ctx, query)
}

// ListAll returns every node.
func (r *NodeRepo) ListAll(ctx context.Context, includeArchived bool) ([]domain.Node, error) {
	query := "SELECT " + nodeColumns + " FROM nodes WHERE 1 = 1" +
		archivedFilter(includeArchived) + nodeOrder
	return r.query(ctx, query)
}

// listSubtreeQuery walks down from one root with a recursive CTE.
//
// The recursion is retrieval shape, not a rule: it decides which rows to fetch
// and nothing about them. The ordering of the result into a tree, the leaf
// test, the derived status and the progress are all domain.Subtree,
// domain.DeriveStatus and friends, working on the slice this returns.
//
// `UNION` rather than `UNION ALL` is the cycle guard. Data cannot legally hold
// a cycle — domain.ValidateMove refuses to create one — but a corrupt file must
// not be able to spin this query forever; UNION drops a row the walk has
// already produced, so the recursion terminates.
//
// The archived filter is applied to the SELECT and not inside the recursion:
// hiding an archived node must not also hide its unarchived children, which is
// what pruning the walk would do.
var listSubtreeQuery = `
WITH RECURSIVE subtree(id) AS (
    SELECT id FROM nodes WHERE id = ?
    UNION
    SELECT nodes.id FROM nodes JOIN subtree ON nodes.parent_id = subtree.id
)
SELECT ` + nodeColumns + ` FROM nodes JOIN subtree ON subtree.id = nodes.id WHERE 1 = 1`

// ListSubtree returns rootID and every node below it, at any depth, and nothing
// from any other subtree. A rootID that does not exist returns an empty slice,
// not an error: "the subtree of a node that is not there is empty" is the
// honest answer, and Get is how a caller asks whether the node exists.
func (r *NodeRepo) ListSubtree(ctx context.Context, rootID string, includeArchived bool) ([]domain.Node, error) {
	return r.query(ctx, listSubtreeQuery+archivedFilter(includeArchived)+nodeOrder, rootID)
}

// query runs a SELECT of nodeColumns and scans every row.
func (r *NodeRepo) query(ctx context.Context, query string, args ...any) ([]domain.Node, error) {
	rows, err := r.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("store: listing nodes: %w", err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	// Never nil: an empty result is an empty slice, so callers do not have to
	// tell "no rows" from "no result".
	out := []domain.Node{}
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("store: listing nodes: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing nodes: %w", err)
	}
	return out, nil
}

// UpdateParentAndOrder re-parents a node and puts it at sortOrder among its new
// siblings. A nil parentID makes it a root.
//
// Whether the move is legal — a cycle, a note as the parent — is
// domain.ValidateMove's decision, taken before this is called. The repository
// writes what it is told; the schema's own foreign key is the backstop for a
// parent that does not exist.
func (r *NodeRepo) UpdateParentAndOrder(ctx context.Context, id string, parentID *string, sortOrder int, updatedAt time.Time) error {
	const stmt = `UPDATE nodes SET parent_id = ?, sort_order = ?, updated_at = ? WHERE id = ?`

	result, err := r.exec.ExecContext(ctx, stmt, nullString(parentID), sortOrder, formatTimestamp(updatedAt), id)
	if err != nil {
		return wrapExec(fmt.Sprintf("moving node %q", id), err)
	}
	return requireRows(result, 1, fmt.Sprintf("node %q", id))
}

// UpdateStatuses applies a whole cascade plan — the slice domain.PlanCascade
// produced — in one statement.
//
// # One statement, on purpose
//
// A plan is applied all or not at all. One statement gives that for free: a
// single UPDATE in SQLite is atomic whether or not it runs inside a
// transaction, so a row the CHECK constraints refuse takes the whole batch
// down with it and nothing is left half-applied. A loop would need a
// transaction of its own, which it cannot open when the repository is already
// bound to the caller's *sql.Tx.
//
// # The CASE is a lookup table, not a rule
//
// `CASE id WHEN 'a' THEN 'today' WHEN 'b' THEN 'done'` is the plan's own
// id -> value mapping spelled in SQL. It computes nothing: which node gets
// which status, and which completed_at goes with it, was decided by
// domain.PlanCascade. There is no condition in here that could disagree with
// the domain.
//
// completed_at is always written, never left alone, because that is what a
// StatusChange means: a node leaving done must lose its completion time in the
// same statement that moves it, or the two columns start telling different
// stories.
//
// An id in the plan that matches no row is ErrNotFound and the whole statement
// has changed nothing — applying three quarters of a cascade is worse than
// applying none.
func (r *NodeRepo) UpdateStatuses(ctx context.Context, changes []domain.StatusChange, updatedAt time.Time) error {
	if len(changes) == 0 {
		return nil
	}

	var (
		statusCase    strings.Builder
		completedCase strings.Builder
		statusArgs    = make([]any, 0, 2*len(changes))
		completedArgs = make([]any, 0, 2*len(changes))
		ids           = make([]string, 0, len(changes))
		seen          = make(map[string]bool, len(changes))
	)

	for _, c := range changes {
		// A repeated id would be listed twice in the IN clause and counted
		// once by RowsAffected. CASE already takes the first matching WHEN, so
		// dropping the repeat here keeps the two in step.
		if seen[c.NodeID] {
			continue
		}
		seen[c.NodeID] = true
		ids = append(ids, c.NodeID)

		statusCase.WriteString(" WHEN ? THEN ?")
		statusArgs = append(statusArgs, c.NodeID, string(c.Status))

		completedCase.WriteString(" WHEN ? THEN ?")
		completedArgs = append(completedArgs, c.NodeID, nullTimestamp(c.CompletedAt))
	}

	guard, guardArgs := idGuard(ids)
	stmt := "UPDATE nodes SET status = CASE id" + statusCase.String() + " END" +
		", completed_at = CASE id" + completedCase.String() + " END" +
		", updated_at = ? WHERE " + guard

	args := make([]any, 0, len(statusArgs)+len(completedArgs)+1+len(guardArgs))
	args = append(args, statusArgs...)
	args = append(args, completedArgs...)
	args = append(args, formatTimestamp(updatedAt))
	args = append(args, guardArgs...)

	result, err := r.exec.ExecContext(ctx, stmt, args...)
	if err != nil {
		return wrapExec("applying status changes", err)
	}
	return requireRows(result, int64(len(ids)), fmt.Sprintf("status changes for %d node(s)", len(ids)))
}

// SetArchivedAt stamps or clears archived_at on every id in one statement —
// archiving a subtree is one call, and un-archiving it is the same call with a
// nil at.
//
// Which ids make up the subtree is domain.PlanArchive's decision; this writes
// the list it is given. An id that matches no row is ErrNotFound and, the
// statement being atomic, nothing has been archived.
func (r *NodeRepo) SetArchivedAt(ctx context.Context, ids []string, at *time.Time, updatedAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}

	unique := make([]string, 0, len(ids))
	seen := make(map[string]bool, len(ids))
	for _, id := range ids {
		if seen[id] {
			continue
		}
		seen[id] = true
		unique = append(unique, id)
	}

	guard, guardArgs := idGuard(unique)
	stmt := "UPDATE nodes SET archived_at = ?, updated_at = ? WHERE " + guard

	args := make([]any, 0, len(guardArgs)+2)
	args = append(args, nullTimestamp(at), formatTimestamp(updatedAt))
	args = append(args, guardArgs...)

	result, err := r.exec.ExecContext(ctx, stmt, args...)
	if err != nil {
		return wrapExec("setting archived_at", err)
	}
	return requireRows(result, int64(len(unique)), fmt.Sprintf("archived_at for %d node(s)", len(unique)))
}
