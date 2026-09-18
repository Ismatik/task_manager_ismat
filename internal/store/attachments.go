package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Attachment is one file attached to a node: a path inside the application's
// data directory and the MIME type it was recognised as.
//
// It lives here rather than in internal/domain because there is no rule
// attached to it. A node's attachments carry no derived state, take part in no
// status derivation and no progress denominator; they are a list of rows. The
// moment a rule appears — a size budget, a thumbnail policy — the type moves
// down into domain with it.
type Attachment struct {
	ID     string
	NodeID string
	Path   string // relative to the app data directory; files are copied in, never linked
	MIME   string
}

// AttachmentRepo reads and writes `attachments`.
//
// # It does not touch the filesystem, and that is a design decision
//
// The repository stores a path and a MIME type. It does not copy the file into
// the app data directory, does not stat it, does not read it and does not
// delete it when the row goes — that is Stage 3's job, above this layer.
//
// Two reasons. A repository that writes files cannot be tested with a temp
// database alone: every test would need a scratch directory, a permissions
// story and a cleanup path, and the interesting failures would be the
// filesystem's rather than the schema's. And a delete that removed the file
// would make DELETE FROM nodes — which cascades to attachments — silently
// destroy the user's documents, from SQL, with nothing in Go having decided to.
//
// This is why neither os nor io/fs appears in this file, and why a test asserts
// that they never will.
type AttachmentRepo struct {
	exec Executor
}

// NewAttachmentRepo returns a repository over exec. Run Migrate first.
func NewAttachmentRepo(exec Executor) *AttachmentRepo { return &AttachmentRepo{exec: exec} }

// WithExecutor returns a copy bound to exec.
func (r *AttachmentRepo) WithExecutor(exec Executor) *AttachmentRepo {
	return &AttachmentRepo{exec: exec}
}

// attachmentColumns is the SELECT list of `attachments`, qualified for joins.
const attachmentColumns = "attachments.id, attachments.node_id, attachments.path, attachments.mime"

func scanAttachment(s rowScanner) (Attachment, error) {
	var a Attachment
	if err := s.Scan(&a.ID, &a.NodeID, &a.Path, &a.MIME); err != nil {
		return Attachment{}, err
	}
	return a, nil
}

// AddAttachment records a file against a node. A node_id that does not exist is
// refused by the foreign key and comes back as ErrConstraint.
func (r *AttachmentRepo) AddAttachment(ctx context.Context, a Attachment) error {
	const stmt = `INSERT INTO attachments (id, node_id, path, mime) VALUES (?, ?, ?, ?)`

	if _, err := r.exec.ExecContext(ctx, stmt, a.ID, a.NodeID, a.Path, a.MIME); err != nil {
		return wrapExec(fmt.Sprintf("adding attachment %q to node %q", a.ID, a.NodeID), err)
	}
	return nil
}

// GetAttachment returns one attachment by id, or ErrNotFound.
func (r *AttachmentRepo) GetAttachment(ctx context.Context, id string) (Attachment, error) {
	row := r.exec.QueryRowContext(ctx,
		"SELECT "+attachmentColumns+" FROM attachments WHERE attachments.id = ?", id)

	a, err := scanAttachment(row)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Attachment{}, fmt.Errorf("store: attachment %q: %w", id, ErrNotFound)
	case err != nil:
		return Attachment{}, fmt.Errorf("store: reading attachment %q: %w", id, err)
	}
	return a, nil
}

// ListAttachments returns a node's attachments, ordered by path so that the
// list reads the same on every open.
func (r *AttachmentRepo) ListAttachments(ctx context.Context, nodeID string) ([]Attachment, error) {
	const query = "SELECT " + attachmentColumns + ` FROM attachments
	               WHERE attachments.node_id = ?
	               ORDER BY attachments.path, attachments.id`

	rows, err := r.exec.QueryContext(ctx, query, nodeID)
	if err != nil {
		return nil, fmt.Errorf("store: listing attachments of %q: %w", nodeID, err)
	}
	defer rows.Close() //nolint:errcheck // the error surfaces from rows.Err below

	out := []Attachment{}
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("store: listing attachments of %q: %w", nodeID, err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: listing attachments of %q: %w", nodeID, err)
	}
	return out, nil
}

// DeleteAttachment removes the row. The file on disk is not touched — see the
// type's documentation — so a caller that means to delete the file has to say
// so, above this layer, deliberately.
func (r *AttachmentRepo) DeleteAttachment(ctx context.Context, id string) error {
	result, err := r.exec.ExecContext(ctx, "DELETE FROM attachments WHERE id = ?", id)
	if err != nil {
		return wrapExec(fmt.Sprintf("deleting attachment %q", id), err)
	}
	return requireRows(result, 1, fmt.Sprintf("attachment %q", id))
}
