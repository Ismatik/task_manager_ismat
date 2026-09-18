package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"nexus/internal/domain"

	sqlite3 "modernc.org/sqlite/lib"
)

// ErrDuplicate reports that a row already exists under a value the schema
// declares unique — today, a tag name.
//
// It is a sentinel and not a driver string on purpose: "Работа already exists"
// is a message the UI has to be able to produce from the error, and matching
// on "UNIQUE constraint failed: tags.name" would tie the frontend's behaviour
// to SQLite's wording. ErrConstraint stays in the chain, so a caller that only
// cares that the database refused the row still matches that.
var ErrDuplicate = errors.New("store: already exists")

// tagColumns is the SELECT list of `tags`, qualified so that the node_tags
// joins below stay unambiguous.
const tagColumns = "tags.id, tags.name, tags.color"

// TagRepo reads and writes `tags` and the `node_tags` link table.
//
// Like NodeRepo it is thin: it stores the name and colour it is given and
// decides nothing about either. A tag's colour is validated by
// domain.Tag.Validate before it gets here.
type TagRepo struct {
	exec Executor
}

// NewTagRepo returns a repository over exec. Run Migrate first.
func NewTagRepo(exec Executor) *TagRepo { return &TagRepo{exec: exec} }

// WithExecutor returns a copy bound to exec, so that tagging can join a
// caller's transaction.
func (r *TagRepo) WithExecutor(exec Executor) *TagRepo { return &TagRepo{exec: exec} }

func scanTag(s rowScanner) (domain.Tag, error) {
	var t domain.Tag
	if err := s.Scan(&t.ID, &t.Name, &t.Color); err != nil {
		return domain.Tag{}, err
	}
	return t, nil
}

// CreateTag inserts t. A name that is already taken comes back as ErrDuplicate.
func (r *TagRepo) CreateTag(ctx context.Context, t domain.Tag) error {
	const stmt = `INSERT INTO tags (id, name, color) VALUES (?, ?, ?)`

	if _, err := r.exec.ExecContext(ctx, stmt, t.ID, t.Name, t.Color); err != nil {
		if isConstraintCode(err, sqlite3.SQLITE_CONSTRAINT_UNIQUE) {
			return fmt.Errorf("store: creating tag %q: %w (%w)", t.Name, ErrDuplicate, wrapExec("creating tag", err))
		}
		return wrapExec(fmt.Sprintf("creating tag %q", t.Name), err)
	}
	return nil
}

// GetTag returns the tag with this id, or ErrNotFound.
func (r *TagRepo) GetTag(ctx context.Context, id string) (domain.Tag, error) {
	row := r.exec.QueryRowContext(ctx, "SELECT "+tagColumns+" FROM tags WHERE tags.id = ?", id)

	t, err := scanTag(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Tag{}, fmt.Errorf("store: tag %q: %w", id, ErrNotFound)
	case err != nil:
		return domain.Tag{}, fmt.Errorf("store: reading tag %q: %w", id, err)
	}
	return t, nil
}

// GetTagByName returns the tag with this exact name, or ErrNotFound.
//
// The match is the column's own: `name` has no COLLATE clause in 0002, so it is
// case-sensitive and "Работа" and "работа" are two tags. That is stated here
// because it is the kind of thing a caller assumes the other way round.
func (r *TagRepo) GetTagByName(ctx context.Context, name string) (domain.Tag, error) {
	row := r.exec.QueryRowContext(ctx, "SELECT "+tagColumns+" FROM tags WHERE tags.name = ?", name)

	t, err := scanTag(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.Tag{}, fmt.Errorf("store: tag %q: %w", name, ErrNotFound)
	case err != nil:
		return domain.Tag{}, fmt.Errorf("store: reading tag %q: %w", name, err)
	}
	return t, nil
}

// ListTags returns every tag, ordered by name so that the tag filter reads the
// same on every open.
func (r *TagRepo) ListTags(ctx context.Context) ([]domain.Tag, error) {
	return r.queryTags(ctx, "SELECT "+tagColumns+" FROM tags ORDER BY tags.name")
}

// DeleteTag removes the tag. Its node_tags rows go with it — that foreign key
// is ON DELETE CASCADE — and the nodes themselves are untouched: removing a
// label is not removing what it was stuck to.
func (r *TagRepo) DeleteTag(ctx context.Context, id string) error {
	result, err := r.exec.ExecContext(ctx, "DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return wrapExec(fmt.Sprintf("deleting tag %q", id), err)
	}
	return requireRows(result, 1, fmt.Sprintf("tag %q", id))
}

// AttachTag links a node and a tag.
//
// Attaching a tag the node already has is a no-op and not an error. The UI will
// do it — a drop onto a node that already carries the tag, a double click, a
// re-run of an import — and an error there would be noise the user has to
// dismiss for something that is already true. ON CONFLICT DO NOTHING is that
// rule, in one clause, on the composite primary key.
//
// A node_id or tag_id that does not exist is a different matter: that is a
// foreign key violation and comes back as ErrConstraint.
func (r *TagRepo) AttachTag(ctx context.Context, nodeID, tagID string) error {
	const stmt = `INSERT INTO node_tags (node_id, tag_id) VALUES (?, ?)
	              ON CONFLICT (node_id, tag_id) DO NOTHING`

	if _, err := r.exec.ExecContext(ctx, stmt, nodeID, tagID); err != nil {
		return wrapExec(fmt.Sprintf("attaching tag %q to node %q", tagID, nodeID), err)
	}
	return nil
}

// DetachTag unlinks a node and a tag. Detaching a tag the node does not have is
// a no-op, for the same reason attaching twice is: the caller's intent —
// "this node must not carry this tag" — is satisfied either way.
func (r *TagRepo) DetachTag(ctx context.Context, nodeID, tagID string) error {
	const stmt = `DELETE FROM node_tags WHERE node_id = ? AND tag_id = ?`

	if _, err := r.exec.ExecContext(ctx, stmt, nodeID, tagID); err != nil {
		return wrapExec(fmt.Sprintf("detaching tag %q from node %q", tagID, nodeID), err)
	}
	return nil
}

// TagsForNode returns the tags attached to one node, ordered by name.
func (r *TagRepo) TagsForNode(ctx context.Context, nodeID string) ([]domain.Tag, error) {
	const query = "SELECT " + tagColumns + ` FROM tags
	               JOIN node_tags ON node_tags.tag_id = tags.id
	               WHERE node_tags.node_id = ?
	               ORDER BY tags.name`
	return r.queryTags(ctx, query, nodeID)
}

// NodesForTag returns the nodes carrying one tag.
//
// Archived nodes are excluded unless asked for, exactly as in NodeRepo: a tag
// filter that quietly resurrected archived cards would undo archiving.
func (r *TagRepo) NodesForTag(ctx context.Context, tagID string, includeArchived bool) ([]domain.Node, error) {
	query := "SELECT " + nodeColumns + ` FROM nodes
	          JOIN node_tags ON node_tags.node_id = nodes.id
	          WHERE node_tags.tag_id = ?` + archivedFilter(includeArchived) + nodeOrder

	rows, err := r.exec.QueryContext(ctx, query, tagID)
	if err != nil {
		return nil, fmt.Errorf("store: listing nodes for tag %q: %w", tagID, err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []domain.Node{}
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, fmt.Errorf("store: listing nodes for tag %q: %w", tagID, err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing nodes for tag %q: %w", tagID, err)
	}
	return out, nil
}

// queryTags runs a SELECT of tagColumns and scans every row. The result is
// never nil.
func (r *TagRepo) queryTags(ctx context.Context, query string, args ...any) ([]domain.Tag, error) {
	rows, err := r.exec.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("store: listing tags: %w", err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []domain.Tag{}
	for rows.Next() {
		t, err := scanTag(rows)
		if err != nil {
			return nil, fmt.Errorf("store: listing tags: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing tags: %w", err)
	}
	return out, nil
}
