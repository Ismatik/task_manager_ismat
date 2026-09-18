package store

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"

	"nexus/internal/domain"
)

// tagRepo returns a tag repository, a node repository and the database behind
// both, over a freshly migrated temp file.
func tagRepo(t *testing.T) (*TagRepo, *NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	return NewTagRepo(db), NewNodeRepo(db), db
}

func newTestTag(id, name string) domain.Tag {
	return domain.Tag{ID: id, Name: name, Color: ""}
}

func mustCreateTag(t *testing.T, ctx context.Context, r *TagRepo, tag domain.Tag) {
	t.Helper()

	if err := r.CreateTag(ctx, tag); err != nil {
		t.Fatalf("CreateTag(%q): %v", tag.Name, err)
	}
}

func tagNames(tags []domain.Tag) []string {
	out := make([]string, len(tags))
	for i, tag := range tags {
		out[i] = tag.Name
	}
	return out
}

func TestTagRepoRoundTrip(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name string
		tag  domain.Tag
	}{
		{name: "a tag with no colour uses the palette default", tag: newTestTag("t1", "work")},
		{name: "a tag with a colour keeps it", tag: domain.Tag{ID: "t2", Name: "urgent", Color: "#ff0055"}},
		{name: "a Russian name survives", tag: domain.Tag{ID: "t3", Name: "Работа", Color: "#0af0c1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, _ := tagRepo(t)
			mustCreateTag(t, ctx, r, tc.tag)

			byName, err := r.GetTagByName(ctx, tc.tag.Name)
			if err != nil {
				t.Fatalf("GetTagByName(%q): %v", tc.tag.Name, err)
			}
			if byName != tc.tag {
				t.Errorf("GetTagByName = %+v, want %+v", byName, tc.tag)
			}

			byID, err := r.GetTag(ctx, tc.tag.ID)
			if err != nil {
				t.Fatalf("GetTag(%q): %v", tc.tag.ID, err)
			}
			if byID != tc.tag {
				t.Errorf("GetTag = %+v, want %+v", byID, tc.tag)
			}

			list, err := r.ListTags(ctx)
			if err != nil {
				t.Fatalf("ListTags: %v", err)
			}
			if len(list) != 1 || list[0] != tc.tag {
				t.Errorf("ListTags = %+v, want exactly [%+v]", list, tc.tag)
			}
		})
	}

	t.Run("ListTags is ordered by name and never nil", func(t *testing.T) {
		r, _, _ := tagRepo(t)

		empty, err := r.ListTags(ctx)
		if err != nil {
			t.Fatalf("ListTags: %v", err)
		}
		if empty == nil || len(empty) != 0 {
			t.Errorf("ListTags on an empty table = %v, want an empty non-nil slice", empty)
		}

		for _, tag := range []domain.Tag{newTestTag("c", "chore"), newTestTag("a", "admin"), newTestTag("b", "bug")} {
			mustCreateTag(t, ctx, r, tag)
		}

		list, err := r.ListTags(ctx)
		if err != nil {
			t.Fatalf("ListTags: %v", err)
		}
		if want := []string{"admin", "bug", "chore"}; !slices.Equal(tagNames(list), want) {
			t.Errorf("ListTags = %v, want %v", tagNames(list), want)
		}
	})

	t.Run("an unknown tag is ErrNotFound", func(t *testing.T) {
		r, _, _ := tagRepo(t)

		if _, err := r.GetTag(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("GetTag(unknown) error = %v, want one matching ErrNotFound", err)
		}
		if _, err := r.GetTagByName(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("GetTagByName(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})
}

// The uniqueness is the database's; the repository's job is to report it in a
// form the UI can act on without reading a driver's prose.
func TestTagRepoDuplicateNameIsMatchable(t *testing.T) {
	ctx := context.Background()
	r, _, _ := tagRepo(t)

	mustCreateTag(t, ctx, r, newTestTag("t1", "work"))

	err := r.CreateTag(ctx, newTestTag("t2", "work"))
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("CreateTag(duplicate name) error = %v, want one matching ErrDuplicate", err)
	}
	if !errors.Is(err, ErrConstraint) {
		t.Errorf("the duplicate error does not also match ErrConstraint: %v", err)
	}

	t.Run("a duplicate id is a constraint failure but not a duplicate name", func(t *testing.T) {
		err := r.CreateTag(ctx, newTestTag("t1", "different"))
		if !errors.Is(err, ErrConstraint) {
			t.Errorf("error = %v, want one matching ErrConstraint", err)
		}
	})

	t.Run("names are case-sensitive, so a different case is a different tag", func(t *testing.T) {
		if err := r.CreateTag(ctx, newTestTag("t3", "Work")); err != nil {
			t.Errorf("CreateTag(%q) = %v, want nil: 0002 declares no COLLATE NOCASE", "Work", err)
		}
	})
}

func TestTagRepoAttachAndDetach(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*TagRepo, *NodeRepo, *sql.DB) {
		t.Helper()
		tags, nodes, db := tagRepo(t)
		mustCreate(t, ctx, nodes, newTestNode("n1"))
		mustCreate(t, ctx, nodes, newTestNode("n2"))
		mustCreateTag(t, ctx, tags, newTestTag("t1", "work"))
		mustCreateTag(t, ctx, tags, newTestTag("t2", "admin"))
		return tags, nodes, db
	}

	t.Run("attach then read back", func(t *testing.T) {
		tags, _, _ := setup(t)

		if err := tags.AttachTag(ctx, "n1", "t1"); err != nil {
			t.Fatalf("AttachTag: %v", err)
		}
		got, err := tags.TagsForNode(ctx, "n1")
		if err != nil {
			t.Fatalf("TagsForNode: %v", err)
		}
		if want := []string{"work"}; !slices.Equal(tagNames(got), want) {
			t.Errorf("TagsForNode(n1) = %v, want %v", tagNames(got), want)
		}
	})

	// The UI will attach a tag a node already has. That is not an error; it is
	// a fact that is already true.
	t.Run("attaching twice is a no-op and leaves exactly one row", func(t *testing.T) {
		tags, _, db := setup(t)

		if err := tags.AttachTag(ctx, "n1", "t1"); err != nil {
			t.Fatalf("first AttachTag: %v", err)
		}
		if err := tags.AttachTag(ctx, "n1", "t1"); err != nil {
			t.Fatalf("second AttachTag = %v, want nil: attaching twice is a no-op", err)
		}
		if n := countRows(t, db, "node_tags"); n != 1 {
			t.Errorf("node_tags has %d rows, want exactly 1", n)
		}

		got, err := tags.TagsForNode(ctx, "n1")
		if err != nil {
			t.Fatalf("TagsForNode: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("TagsForNode(n1) = %v, want one tag", tagNames(got))
		}
	})

	t.Run("detach removes the link and leaves node and tag alone", func(t *testing.T) {
		tags, nodes, db := setup(t)

		if err := tags.AttachTag(ctx, "n1", "t1"); err != nil {
			t.Fatalf("AttachTag: %v", err)
		}
		if err := tags.DetachTag(ctx, "n1", "t1"); err != nil {
			t.Fatalf("DetachTag: %v", err)
		}
		if n := countRows(t, db, "node_tags"); n != 0 {
			t.Errorf("node_tags has %d rows, want 0", n)
		}
		if _, err := nodes.Get(ctx, "n1"); err != nil {
			t.Errorf("the node is gone after a detach: %v", err)
		}
		if _, err := tags.GetTag(ctx, "t1"); err != nil {
			t.Errorf("the tag is gone after a detach: %v", err)
		}
	})

	t.Run("detaching a tag the node does not have is a no-op", func(t *testing.T) {
		tags, _, _ := setup(t)

		if err := tags.DetachTag(ctx, "n1", "t1"); err != nil {
			t.Errorf("DetachTag on an absent link = %v, want nil", err)
		}
	})

	t.Run("TagsForNode is ordered by name and never nil", func(t *testing.T) {
		tags, _, _ := setup(t)

		empty, err := tags.TagsForNode(ctx, "n1")
		if err != nil {
			t.Fatalf("TagsForNode: %v", err)
		}
		if empty == nil || len(empty) != 0 {
			t.Errorf("TagsForNode with no tags = %v, want an empty non-nil slice", empty)
		}

		for _, id := range []string{"t1", "t2"} {
			if err := tags.AttachTag(ctx, "n1", id); err != nil {
				t.Fatalf("AttachTag(%q): %v", id, err)
			}
		}
		got, err := tags.TagsForNode(ctx, "n1")
		if err != nil {
			t.Fatalf("TagsForNode: %v", err)
		}
		if want := []string{"admin", "work"}; !slices.Equal(tagNames(got), want) {
			t.Errorf("TagsForNode(n1) = %v, want %v", tagNames(got), want)
		}
	})

	t.Run("a node or tag that does not exist is refused by the foreign key", func(t *testing.T) {
		tags, _, _ := setup(t)

		if err := tags.AttachTag(ctx, "ghost", "t1"); !errors.Is(err, ErrConstraint) {
			t.Errorf("AttachTag(unknown node) error = %v, want one matching ErrConstraint", err)
		}
		if err := tags.AttachTag(ctx, "n1", "ghost"); !errors.Is(err, ErrConstraint) {
			t.Errorf("AttachTag(unknown tag) error = %v, want one matching ErrConstraint", err)
		}
	})
}

func TestTagRepoNodesForTag(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*TagRepo, *NodeRepo) {
		t.Helper()
		tags, nodes, _ := tagRepo(t)
		for _, id := range []string{"n1", "n2", "n3"} {
			mustCreate(t, ctx, nodes, newTestNode(id))
		}
		mustCreateTag(t, ctx, tags, newTestTag("t1", "work"))
		for _, id := range []string{"n1", "n2"} {
			if err := tags.AttachTag(ctx, id, "t1"); err != nil {
				t.Fatalf("AttachTag(%q): %v", id, err)
			}
		}
		return tags, nodes
	}

	t.Run("returns exactly the tagged nodes", func(t *testing.T) {
		tags, _ := setup(t)

		got, err := tags.NodesForTag(ctx, "t1", false)
		if err != nil {
			t.Fatalf("NodesForTag: %v", err)
		}
		if want := []string{"n1", "n2"}; !slices.Equal(ids(got), want) {
			t.Errorf("NodesForTag(t1) = %v, want %v", ids(got), want)
		}
	})

	t.Run("archived nodes are excluded by default and included on request", func(t *testing.T) {
		tags, nodes := setup(t)

		at := testNow
		if err := nodes.SetArchivedAt(ctx, []string{"n2"}, &at, testNow); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}

		got, err := tags.NodesForTag(ctx, "t1", false)
		if err != nil {
			t.Fatalf("NodesForTag: %v", err)
		}
		if want := []string{"n1"}; !slices.Equal(ids(got), want) {
			t.Errorf("NodesForTag(t1, false) = %v, want %v", ids(got), want)
		}

		got, err = tags.NodesForTag(ctx, "t1", true)
		if err != nil {
			t.Fatalf("NodesForTag: %v", err)
		}
		if want := []string{"n1", "n2"}; !slices.Equal(ids(got), want) {
			t.Errorf("NodesForTag(t1, true) = %v, want %v", ids(got), want)
		}
	})

	t.Run("a tag nobody carries lists nothing rather than nil", func(t *testing.T) {
		tags, _ := setup(t)
		mustCreateTag(t, ctx, tags, newTestTag("t2", "unused"))

		got, err := tags.NodesForTag(ctx, "t2", false)
		if err != nil {
			t.Fatalf("NodesForTag: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("NodesForTag(t2) = %v, want an empty non-nil slice", ids(got))
		}
	})
}

// Both foreign keys in node_tags are ON DELETE CASCADE, in opposite directions.
// Deleting either end must take the link and only the link.
func TestNodeTagsCascade(t *testing.T) {
	ctx := context.Background()

	setup := func(t *testing.T) (*TagRepo, *NodeRepo, *sql.DB) {
		t.Helper()
		tags, nodes, db := tagRepo(t)
		for _, id := range []string{"n1", "n2"} {
			mustCreate(t, ctx, nodes, newTestNode(id))
		}
		for _, tag := range []domain.Tag{newTestTag("t1", "work"), newTestTag("t2", "admin")} {
			mustCreateTag(t, ctx, tags, tag)
		}
		for _, link := range [][2]string{{"n1", "t1"}, {"n1", "t2"}, {"n2", "t1"}} {
			if err := tags.AttachTag(ctx, link[0], link[1]); err != nil {
				t.Fatalf("AttachTag%v: %v", link, err)
			}
		}
		return tags, nodes, db
	}

	t.Run("deleting a tag removes its links and leaves the nodes", func(t *testing.T) {
		tags, nodes, db := setup(t)

		if err := tags.DeleteTag(ctx, "t1"); err != nil {
			t.Fatalf("DeleteTag: %v", err)
		}

		if n := countRows(t, db, "node_tags"); n != 1 {
			t.Errorf("node_tags has %d rows, want 1 (only n1<->t2 survives)", n)
		}
		for _, id := range []string{"n1", "n2"} {
			if _, err := nodes.Get(ctx, id); err != nil {
				t.Errorf("node %q was removed with the tag: %v", id, err)
			}
		}
		got, err := tags.TagsForNode(ctx, "n1")
		if err != nil {
			t.Fatalf("TagsForNode: %v", err)
		}
		if want := []string{"admin"}; !slices.Equal(tagNames(got), want) {
			t.Errorf("TagsForNode(n1) = %v, want %v", tagNames(got), want)
		}
	})

	t.Run("deleting a node removes its links and leaves the tags", func(t *testing.T) {
		tags, nodes, db := setup(t)

		if err := nodes.Delete(ctx, "n1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}

		if n := countRows(t, db, "node_tags"); n != 1 {
			t.Errorf("node_tags has %d rows, want 1 (only n2<->t1 survives)", n)
		}
		if n := countRows(t, db, "tags"); n != 2 {
			t.Errorf("tags has %d rows, want 2 — deleting a node must not delete tags", n)
		}
		list, err := tags.ListTags(ctx)
		if err != nil {
			t.Fatalf("ListTags: %v", err)
		}
		if want := []string{"admin", "work"}; !slices.Equal(tagNames(list), want) {
			t.Errorf("ListTags = %v, want %v", tagNames(list), want)
		}
	})

	t.Run("deleting a tag that is not there is ErrNotFound", func(t *testing.T) {
		tags, _, _ := setup(t)

		if err := tags.DeleteTag(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("DeleteTag(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})
}

func TestTagRepoRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	tags, nodes, db := tagRepo(t)
	mustCreate(t, ctx, nodes, newTestNode("n1"))

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	inTx := tags.WithExecutor(tx)
	mustCreateTag(t, ctx, inTx, newTestTag("t1", "work"))
	if err := inTx.AttachTag(ctx, "n1", "t1"); err != nil {
		t.Fatalf("AttachTag in the transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if n := countRows(t, db, "tags"); n != 0 {
		t.Errorf("tags has %d rows after a rollback, want 0", n)
	}
	if n := countRows(t, db, "node_tags"); n != 0 {
		t.Errorf("node_tags has %d rows after a rollback, want 0", n)
	}
}
