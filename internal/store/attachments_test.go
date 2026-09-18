package store

import (
	"context"
	"database/sql"
	"errors"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"
)

// attachmentRepo returns an attachment repository over a freshly migrated temp
// database that already has nodes "n1" and "n2".
func attachmentRepo(t *testing.T) (*AttachmentRepo, *NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	nodes := NewNodeRepo(db)

	ctx := context.Background()
	for _, id := range []string{"n1", "n2"} {
		mustCreate(t, ctx, nodes, newTestNode(id))
	}
	return NewAttachmentRepo(db), nodes, db
}

func attachmentIDs(as []Attachment) []string {
	out := make([]string, len(as))
	for i, a := range as {
		out[i] = a.ID
	}
	return out
}

func TestAttachmentRepoRoundTrip(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		att  Attachment
	}{
		{
			name: "a file with a mime type",
			att:  Attachment{ID: "a1", NodeID: "n1", Path: "attachments/n1/spec.pdf", MIME: "application/pdf"},
		},
		{
			// 0002 defaults mime to '', meaning "not recognised". It is a
			// value, not an absence, and it has to survive as one.
			name: "an unrecognised file keeps an empty mime rather than NULL",
			att:  Attachment{ID: "a2", NodeID: "n1", Path: "attachments/n1/notes", MIME: ""},
		},
		{
			name: "a path with non-ASCII characters",
			att:  Attachment{ID: "a3", NodeID: "n1", Path: "attachments/n1/Отчёт за сентябрь.docx", MIME: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, _ := attachmentRepo(t)

			if err := r.AddAttachment(ctx, tc.att); err != nil {
				t.Fatalf("AddAttachment: %v", err)
			}

			got, err := r.GetAttachment(ctx, tc.att.ID)
			if err != nil {
				t.Fatalf("GetAttachment: %v", err)
			}
			if got != tc.att {
				t.Errorf("GetAttachment = %+v, want %+v", got, tc.att)
			}

			list, err := r.ListAttachments(ctx, tc.att.NodeID)
			if err != nil {
				t.Fatalf("ListAttachments: %v", err)
			}
			if len(list) != 1 || list[0] != tc.att {
				t.Errorf("ListAttachments = %+v, want exactly [%+v]", list, tc.att)
			}
		})
	}
}

func TestAttachmentRepoListAndDelete(t *testing.T) {
	ctx := context.Background()

	seed := func(t *testing.T) (*AttachmentRepo, *sql.DB) {
		t.Helper()
		r, _, db := attachmentRepo(t)
		for _, a := range []Attachment{
			{ID: "a2", NodeID: "n1", Path: "b.png", MIME: "image/png"},
			{ID: "a1", NodeID: "n1", Path: "a.png", MIME: "image/png"},
			{ID: "a3", NodeID: "n2", Path: "c.png", MIME: "image/png"},
		} {
			if err := r.AddAttachment(ctx, a); err != nil {
				t.Fatalf("AddAttachment(%q): %v", a.ID, err)
			}
		}
		return r, db
	}

	t.Run("lists one node's attachments, ordered by path", func(t *testing.T) {
		r, _ := seed(t)

		got, err := r.ListAttachments(ctx, "n1")
		if err != nil {
			t.Fatalf("ListAttachments: %v", err)
		}
		if want := []string{"a1", "a2"}; !slices.Equal(attachmentIDs(got), want) {
			t.Errorf("ListAttachments(n1) = %v, want %v", attachmentIDs(got), want)
		}
	})

	t.Run("a node with no attachments lists nothing rather than nil", func(t *testing.T) {
		r, _, _ := attachmentRepo(t)

		got, err := r.ListAttachments(ctx, "n1")
		if err != nil {
			t.Fatalf("ListAttachments: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListAttachments = %v, want an empty non-nil slice", got)
		}
	})

	t.Run("delete removes one row and leaves the rest", func(t *testing.T) {
		r, db := seed(t)

		if err := r.DeleteAttachment(ctx, "a1"); err != nil {
			t.Fatalf("DeleteAttachment: %v", err)
		}
		if n := countRows(t, db, "attachments"); n != 2 {
			t.Errorf("attachments has %d rows, want 2", n)
		}
		if _, err := r.GetAttachment(ctx, "a1"); !errors.Is(err, ErrNotFound) {
			t.Errorf("GetAttachment(deleted) error = %v, want one matching ErrNotFound", err)
		}
	})

	t.Run("deleting an attachment that is not there is ErrNotFound", func(t *testing.T) {
		r, _ := seed(t)

		if err := r.DeleteAttachment(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("DeleteAttachment(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})

	t.Run("an unknown id is ErrNotFound", func(t *testing.T) {
		r, _ := seed(t)

		if _, err := r.GetAttachment(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("GetAttachment(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})

	t.Run("a node that does not exist is refused by the foreign key", func(t *testing.T) {
		r, _, _ := attachmentRepo(t)

		err := r.AddAttachment(ctx, Attachment{ID: "a1", NodeID: "ghost", Path: "x.png"})
		if !errors.Is(err, ErrConstraint) {
			t.Errorf("AddAttachment(unknown node) error = %v, want one matching ErrConstraint", err)
		}
	})

	t.Run("a duplicate id is refused by the primary key", func(t *testing.T) {
		r, _ := seed(t)

		err := r.AddAttachment(ctx, Attachment{ID: "a1", NodeID: "n1", Path: "other.png"})
		if !errors.Is(err, ErrConstraint) {
			t.Errorf("AddAttachment(duplicate id) error = %v, want one matching ErrConstraint", err)
		}
	})
}

func TestAttachmentsCascadeWithTheNode(t *testing.T) {
	ctx := context.Background()
	r, nodes, db := attachmentRepo(t)

	for _, a := range []Attachment{
		{ID: "a1", NodeID: "n1", Path: "a.png"},
		{ID: "a2", NodeID: "n2", Path: "b.png"},
	} {
		if err := r.AddAttachment(ctx, a); err != nil {
			t.Fatalf("AddAttachment(%q): %v", a.ID, err)
		}
	}

	if err := nodes.Delete(ctx, "n1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if n := countRows(t, db, "attachments"); n != 1 {
		t.Errorf("attachments has %d rows, want 1 — only n1's should be gone", n)
	}
	if _, err := r.GetAttachment(ctx, "a2"); err != nil {
		t.Errorf("n2's attachment was removed with n1: %v", err)
	}
}

func TestAttachmentRepoRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	r, _, db := attachmentRepo(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}
	if err := r.WithExecutor(tx).AddAttachment(ctx, Attachment{ID: "a1", NodeID: "n1", Path: "a.png"}); err != nil {
		t.Fatalf("AddAttachment in the transaction: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if n := countRows(t, db, "attachments"); n != 0 {
		t.Errorf("attachments has %d rows after a rollback, want 0", n)
	}
}

// The repository stores a path; it does not touch the filesystem. Copying files
// into the app data directory is Stage 3's, above this layer, and a repository
// that wrote files could not be tested with a temp database alone — and a
// DELETE that cascaded from `nodes` would silently destroy the user's
// documents from SQL, with nothing in Go having decided to.
//
// Asserting on the import list rather than on behaviour is the point: this is a
// constraint on what the file may ever do, not on what it happens to do today.
func TestAttachmentsTouchesNoFilesystem(t *testing.T) {
	const path = "attachments.go"

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	forbidden := map[string]bool{
		"os":            true,
		"io/fs":         true,
		"io/ioutil":     true,
		"path/filepath": true,
		"os/exec":       true,
	}

	var imported []string
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			t.Fatalf("unquoting import %s: %v", spec.Path.Value, err)
		}
		imported = append(imported, importPath)
		if forbidden[importPath] {
			t.Errorf("%s imports %q; this repository must not touch the filesystem", path, importPath)
		}
	}
	if len(imported) == 0 {
		t.Fatalf("%s reported no imports at all; the parse cannot be right", path)
	}
}
