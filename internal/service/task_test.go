package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

// testNow is the instant every service test is written against: 2026-09-18,
// 10:30 UTC — a FRIDAY, which is the live edge case of the "this week" due rule
// (D1). Nothing here reads a real clock, so every timestamp in these tests is a
// value the test chose and can assert on exactly.
var testNow = time.Date(2026, time.September, 18, 10, 30, 0, 0, time.UTC)

// fixture is a migrated temp database, the repositories over it, a clock the
// test can advance by hand, and the services built on all of that.
type fixture struct {
	t       *testing.T
	db      *sql.DB
	begin   service.Beginner
	nodes   *store.NodeRepo
	tags    *store.TagRepo
	entries *store.TimeEntryRepo
	checks  *store.HabitCheckRepo
	search  *store.SearchRepo

	// now is what the injected Clock reports. A test that needs time to pass
	// assigns to it; nothing ever sleeps.
	now time.Time

	ids   int
	tasks *service.TaskService
}

// newFixture opens a fresh database in a temp directory, applies every
// migration and returns the services over it.
func newFixture(t *testing.T) *fixture {
	t.Helper()

	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "nexus.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("closing the database: %v", err)
		}
	})
	if _, err := store.Migrate(ctx, db); err != nil {
		t.Fatalf("store.Migrate: %v", err)
	}

	f := &fixture{
		t:       t,
		db:      db,
		begin:   service.NewBeginner(db),
		nodes:   store.NewNodeRepo(db),
		tags:    store.NewTagRepo(db),
		entries: store.NewTimeEntryRepo(db),
		checks:  store.NewHabitCheckRepo(db),
		search:  store.NewSearchRepo(db),
		now:     testNow,
	}
	f.tasks = service.NewTaskService(f.begin, f.nodes, f.tags, f.clock(), f.nextID)
	return f
}

// clock is the injected Clock: it reads f.now every time, so a test can move
// time forward by assigning to the field.
func (f *fixture) clock() service.Clock {
	return func() time.Time { return f.now }
}

// nextID is the injected id generator: n1, n2, n3… Deterministic ids make a
// failure message name the node the reader can find in the test above.
func (f *fixture) nextID() string {
	f.ids++
	return fmt.Sprintf("n%d", f.ids)
}

// create inserts a draft through the service and fails the test if it cannot.
func (f *fixture) create(draft service.NewNode) domain.Node {
	f.t.Helper()

	n, err := f.tasks.CreateNode(context.Background(), draft)
	if err != nil {
		f.t.Fatalf("CreateNode(%q): %v", draft.Title, err)
	}
	return n
}

// all returns every row of `nodes`, archived included, in a deterministic order.
func (f *fixture) all() []domain.Node {
	f.t.Helper()

	set, err := f.nodes.ListAll(context.Background(), true)
	if err != nil {
		f.t.Fatalf("ListAll: %v", err)
	}
	return set
}

// get returns one node, or fails the test.
func (f *fixture) get(id string) domain.Node {
	f.t.Helper()

	n, err := f.nodes.Get(context.Background(), id)
	if err != nil {
		f.t.Fatalf("Get(%q): %v", id, err)
	}
	return n
}

// draft builds a NewNode with the two fields every test has to supply.
func draft(title string, typ domain.NodeType, parent *string) service.NewNode {
	return service.NewNode{ParentID: parent, Type: typ, Title: title}
}

func ptr[T any](v T) *T { return &v }

// headline builds D2's headline tree, through the service, and returns its ids
// by name:
//
//	p (project)
//	├── mid (project)
//	│   ├── deep (task, week)
//	│   └── finished (task, done a month ago)
//	├── shallow (task, backlog)
//	└── memo (note)
func (f *fixture) headline() map[string]string {
	f.t.Helper()

	ctx := context.Background()
	p := f.create(draft("p", domain.NodeTypeProject, nil))
	mid := f.create(draft("mid", domain.NodeTypeProject, &p.ID))

	deepDraft := draft("deep", domain.NodeTypeTask, &mid.ID)
	deepDraft.Status = domain.StatusWeek
	deep := f.create(deepDraft)

	finished := f.create(draft("finished", domain.NodeTypeTask, &mid.ID))
	shallow := f.create(draft("shallow", domain.NodeTypeTask, &p.ID))
	memo := f.create(draft("memo", domain.NodeTypeNote, &p.ID))

	// finished was completed a month ago, through the same cascade path the
	// rest of the app uses, then the clock is put back.
	f.now = testNow.AddDate(0, -1, 0)
	if _, err := f.tasks.MoveToColumn(ctx, finished.ID, domain.StatusDone); err != nil {
		f.t.Fatalf("finishing %q: %v", finished.ID, err)
	}
	f.now = testNow

	return map[string]string{
		"p": p.ID, "mid": mid.ID, "deep": deep.ID,
		"finished": finished.ID, "shallow": shallow.ID, "memo": memo.ID,
	}
}

// ---------------------------------------------------------------------------
// CreateNode

func TestCreateNode(t *testing.T) {
	ctx := context.Background()

	t.Run("stamps the id, the clock and D1's manual provenance", func(t *testing.T) {
		f := newFixture(t)

		n := f.create(draft("write the report", domain.NodeTypeTask, nil))

		if n.ID != "n1" {
			t.Errorf("ID = %q, want the injected generator's n1", n.ID)
		}
		if !n.CreatedAt.Equal(testNow) || !n.UpdatedAt.Equal(testNow) {
			t.Errorf("CreatedAt/UpdatedAt = %v/%v, want %v", n.CreatedAt, n.UpdatedAt, testNow)
		}
		if n.Status != domain.StatusBacklog {
			t.Errorf("Status = %q, want backlog by default", n.Status)
		}
		if n.Priority != domain.Priority4 {
			t.Errorf("Priority = %d, want 4 by default", int(n.Priority))
		}
		if n.DueSource != domain.DueSourceManual {
			t.Errorf("DueSource = %q, want manual", n.DueSource)
		}
	})

	t.Run("applies D4's default activity per type", func(t *testing.T) {
		for _, typ := range domain.NodeTypes() {
			t.Run(typ.String(), func(t *testing.T) {
				f := newFixture(t)

				d := draft("x", typ, nil)
				if typ == domain.NodeTypeHabit {
					d.Recurrence = ptr("FREQ=DAILY")
				}
				n := f.create(d)

				want := domain.DefaultActivity(typ)
				switch {
				case want == nil && n.Activity != nil:
					t.Fatalf("Activity = %q, want none", *n.Activity)
				case want == nil:
				case n.Activity == nil:
					t.Fatalf("Activity = nil, want %q", *want)
				case *n.Activity != *want:
					t.Errorf("Activity = %q, want %q", *n.Activity, *want)
				}
			})
		}
	})

	t.Run("an activity chosen by the caller survives", func(t *testing.T) {
		f := newFixture(t)

		d := draft("x", domain.NodeTypeTask, nil)
		d.Activity = ptr(domain.ActivityMeeting)
		n := f.create(d)

		if n.Activity == nil || *n.Activity != domain.ActivityMeeting {
			t.Errorf("Activity = %v, want the caller's meeting", n.Activity)
		}
	})

	t.Run("sort_order is appended after the existing siblings", func(t *testing.T) {
		f := newFixture(t)

		parent := f.create(draft("parent", domain.NodeTypeProject, nil))
		var got []int
		for range 3 {
			got = append(got, f.create(draft("child", domain.NodeTypeTask, &parent.ID)).SortOrder)
		}
		if want := []int{0, 1, 2}; !reflect.DeepEqual(got, want) {
			t.Errorf("sort_order = %v, want %v", got, want)
		}

		// Roots are numbered in their own range, not in the children's.
		second := f.create(draft("another root", domain.NodeTypeTask, nil))
		if second.SortOrder != 1 {
			t.Errorf("the second root's sort_order = %d, want 1", second.SortOrder)
		}
	})

	t.Run("an invalid node is refused and nothing is written", func(t *testing.T) {
		cases := []struct {
			name  string
			draft service.NewNode
		}{
			{"an empty title", draft("", domain.NodeTypeTask, nil)},
			{"an unknown type", draft("x", domain.NodeType("epic"), nil)},
			{"a habit without a recurrence", draft("x", domain.NodeTypeHabit, nil)},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := newFixture(t)

				if _, err := f.tasks.CreateNode(ctx, tc.draft); !errors.Is(err, domain.ErrInvalid) {
					t.Fatalf("CreateNode() = %v, want a domain.ErrInvalid", err)
				}
				if got := len(f.all()); got != 0 {
					t.Errorf("%d rows were written by a refused create, want 0", got)
				}
			})
		}
	})

	t.Run("a parent that does not exist is refused", func(t *testing.T) {
		f := newFixture(t)

		_, err := f.tasks.CreateNode(ctx, draft("x", domain.NodeTypeTask, ptr("ghost")))
		if !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("CreateNode() = %v, want domain.ErrNodeNotFound", err)
		}
		if got := len(f.all()); got != 0 {
			t.Errorf("%d rows written, want 0", got)
		}
	})

	t.Run("a note cannot be a parent", func(t *testing.T) {
		f := newFixture(t)

		memo := f.create(draft("memo", domain.NodeTypeNote, nil))

		_, err := f.tasks.CreateNode(ctx, draft("x", domain.NodeTypeTask, &memo.ID))
		if !errors.Is(err, domain.ErrNoteParent) {
			t.Fatalf("CreateNode() = %v, want domain.ErrNoteParent", err)
		}
		if got := len(f.all()); got != 1 {
			t.Errorf("%d rows, want only the note", got)
		}
	})
}

// ---------------------------------------------------------------------------
// MoveToColumn — the cascade

// THE headline case of D2, end to end against a real database: drag a parent to
// Done and every unfinished non-note descendant finishes, with completed_at,
// while the note and the already-finished task are untouched — and reading the
// tree back, the parent DERIVES done.
func TestMoveToColumnParentToDone(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	ids := f.headline()

	completedBefore := f.get(ids["finished"]).CompletedAt
	memoBefore := f.get(ids["memo"])

	if _, err := f.tasks.MoveToColumn(ctx, ids["p"], domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(p, done) = %v", err)
	}

	t.Run("every unfinished non-note descendant is done in the database", func(t *testing.T) {
		for _, name := range []string{"mid", "deep", "shallow"} {
			n := f.get(ids[name])
			if n.Status != domain.StatusDone {
				t.Errorf("%s: status = %q, want done", name, n.Status)
			}
			if n.CompletedAt == nil {
				t.Fatalf("%s: completed_at is NULL, want the injected clock", name)
			}
			if !n.CompletedAt.Equal(testNow) {
				t.Errorf("%s: completed_at = %v, want %v", name, n.CompletedAt, testNow)
			}
			if !n.UpdatedAt.Equal(testNow) {
				t.Errorf("%s: updated_at = %v, want %v", name, n.UpdatedAt, testNow)
			}
		}
	})

	t.Run("the note is untouched", func(t *testing.T) {
		memo := f.get(ids["memo"])
		if !reflect.DeepEqual(memo, memoBefore) {
			t.Errorf("the note changed:\n got %+v\nwant %+v", memo, memoBefore)
		}
	})

	t.Run("an already finished task keeps its own completed_at", func(t *testing.T) {
		got := f.get(ids["finished"]).CompletedAt
		if got == nil || completedBefore == nil || !got.Equal(*completedBefore) {
			t.Errorf("completed_at = %v, want the original %v", got, completedBefore)
		}
	})

	t.Run("the parent derives done when the tree is read back", func(t *testing.T) {
		got, err := domain.DeriveStatus(f.all(), ids["p"])
		if err != nil {
			t.Fatalf("DeriveStatus: %v", err)
		}
		if got != domain.StatusDone {
			t.Errorf("DeriveStatus(p) = %q, want done", got)
		}
	})
}

// D9 — confirmed by the user: a project NEVER enters Doing, with children or
// without. The rejection is the domain's, surfaced as an error by the service,
// and nothing is written.
func TestMoveToColumnProjectNeverEntersDoing(t *testing.T) {
	ctx := context.Background()

	t.Run("an empty project is refused and the database is unchanged", func(t *testing.T) {
		f := newFixture(t)
		p := f.create(draft("empty project", domain.NodeTypeProject, nil))
		before := f.all()

		_, err := f.tasks.MoveToColumn(ctx, p.ID, domain.StatusDoing)
		if !errors.Is(err, domain.ErrProjectNeverDoing) {
			t.Fatalf("MoveToColumn(project, doing) = %v, want domain.ErrProjectNeverDoing", err)
		}
		if after := f.all(); !reflect.DeepEqual(before, after) {
			t.Errorf("the database changed:\n got %+v\nwant %+v", after, before)
		}
	})

	t.Run("a project with children is refused", func(t *testing.T) {
		f := newFixture(t)
		ids := f.headline()
		before := f.all()

		_, err := f.tasks.MoveToColumn(ctx, ids["p"], domain.StatusDoing)
		if !errors.Is(err, domain.ErrProjectNeverDoing) {
			t.Fatalf("MoveToColumn(project, doing) = %v, want domain.ErrProjectNeverDoing", err)
		}
		if after := f.all(); !reflect.DeepEqual(before, after) {
			t.Error("the database changed on a refused move")
		}
	})

	t.Run("an empty project below the dragged node is not stored as doing", func(t *testing.T) {
		f := newFixture(t)
		root := f.create(draft("root", domain.NodeTypeTask, nil))
		empty := f.create(draft("empty project", domain.NodeTypeProject, &root.ID))
		leaf := f.create(draft("leaf", domain.NodeTypeTask, &root.ID))

		if _, err := f.tasks.MoveToColumn(ctx, root.ID, domain.StatusDoing); err != nil {
			t.Fatalf("MoveToColumn(root, doing) = %v", err)
		}
		if got := f.get(empty.ID).Status; got == domain.StatusDoing {
			t.Errorf("the empty project's status = %q; D9 says a project is never doing", got)
		}
		if got := f.get(leaf.ID).Status; got != domain.StatusDoing {
			t.Errorf("the leaf's status = %q, want doing", got)
		}
	})

	t.Run("the same project moves to every other column", func(t *testing.T) {
		for _, target := range domain.Statuses() {
			if target == domain.StatusDoing {
				continue
			}
			t.Run(target.String(), func(t *testing.T) {
				f := newFixture(t)
				p := f.create(draft("empty project", domain.NodeTypeProject, nil))

				if _, err := f.tasks.MoveToColumn(ctx, p.ID, target); err != nil {
					t.Fatalf("MoveToColumn(project, %s) = %v", target, err)
				}
				if got := f.get(p.ID).Status; got != target {
					t.Errorf("status = %q, want %q", got, target)
				}
			})
		}
	})
}

// ---------------------------------------------------------------------------
// The column <-> due coupling (D1, D8) and all four due_source transitions.

func TestMoveToColumnWritesTheDueDate(t *testing.T) {
	ctx := context.Background()
	friday := domain.NewDate(2026, time.September, 18) // testNow is a Friday

	t.Run("to today: due = today, due_source = auto", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		moved, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday)
		if err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		if moved.Due == nil || !moved.Due.Equal(friday) {
			t.Errorf("due = %v, want %v", moved.Due, friday)
		}
		if moved.DueSource != domain.DueSourceAuto {
			t.Errorf("due_source = %q, want auto", moved.DueSource)
		}
	})

	t.Run("to this week on a Friday: due = today", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		moved, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusWeek)
		if err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		if moved.Due == nil || !moved.Due.Equal(friday) {
			t.Errorf("due = %v, want the same Friday %v", moved.Due, friday)
		}
		if moved.DueSource != domain.DueSourceAuto {
			t.Errorf("due_source = %q, want auto", moved.DueSource)
		}
	})

	t.Run("D8: a column move overwrites a manual date and takes it over", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))
		typed := domain.NewDate(2026, time.December, 24)

		if _, err := f.tasks.SetDue(ctx, n.ID, &typed); err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		moved, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday)
		if err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		if moved.Due == nil || !moved.Due.Equal(friday) {
			t.Errorf("due = %v, want the column's %v — the move always wins", moved.Due, friday)
		}
		if moved.DueSource != domain.DueSourceAuto {
			t.Errorf("due_source = %q, want auto", moved.DueSource)
		}
	})

	t.Run("to backlog: an auto date is cleared", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday); err != nil {
			t.Fatalf("MoveToColumn(today): %v", err)
		}
		moved, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusBacklog)
		if err != nil {
			t.Fatalf("MoveToColumn(backlog): %v", err)
		}
		if moved.Due != nil {
			t.Errorf("due = %v, want it cleared", moved.Due)
		}
		if moved.DueSource != domain.DueSourceManual {
			t.Errorf("due_source = %q, want manual once the date is gone", moved.DueSource)
		}
	})

	t.Run("to backlog: a manual date survives", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))
		typed := domain.NewDate(2026, time.December, 24)

		if _, err := f.tasks.SetDue(ctx, n.ID, &typed); err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		moved, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusBacklog)
		if err != nil {
			t.Fatalf("MoveToColumn(backlog): %v", err)
		}
		if moved.Due == nil || !moved.Due.Equal(typed) {
			t.Errorf("due = %v, want the user's %v", moved.Due, typed)
		}
		if moved.DueSource != domain.DueSourceManual {
			t.Errorf("due_source = %q, want manual", moved.DueSource)
		}
	})

	t.Run("to doing and done: the due date is left alone", func(t *testing.T) {
		for _, target := range []domain.Status{domain.StatusDoing, domain.StatusDone} {
			t.Run(target.String(), func(t *testing.T) {
				f := newFixture(t)
				n := f.create(draft("x", domain.NodeTypeTask, nil))
				typed := domain.NewDate(2026, time.December, 24)

				if _, err := f.tasks.SetDue(ctx, n.ID, &typed); err != nil {
					t.Fatalf("SetDue: %v", err)
				}
				moved, err := f.tasks.MoveToColumn(ctx, n.ID, target)
				if err != nil {
					t.Fatalf("MoveToColumn(%s): %v", target, err)
				}
				if moved.Due == nil || !moved.Due.Equal(typed) {
					t.Errorf("due = %v, want %v untouched", moved.Due, typed)
				}
				if moved.DueSource != domain.DueSourceManual {
					t.Errorf("due_source = %q, want manual", moved.DueSource)
				}
			})
		}
	})
}

func TestSetDueIsAlwaysManual(t *testing.T) {
	ctx := context.Background()

	t.Run("setting a date", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))
		typed := domain.NewDate(2026, time.October, 1)

		got, err := f.tasks.SetDue(ctx, n.ID, &typed)
		if err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		if got.Due == nil || !got.Due.Equal(typed) {
			t.Errorf("due = %v, want %v", got.Due, typed)
		}
		if got.DueSource != domain.DueSourceManual {
			t.Errorf("due_source = %q, want manual", got.DueSource)
		}
	})

	t.Run("typing the date a column move picked is still manual", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		auto, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday)
		if err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		got, err := f.tasks.SetDue(ctx, n.ID, auto.Due)
		if err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		if got.DueSource != domain.DueSourceManual {
			t.Errorf("due_source = %q, want manual", got.DueSource)
		}
	})

	t.Run("clearing a date is still manual", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday); err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		got, err := f.tasks.SetDue(ctx, n.ID, nil)
		if err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		if got.Due != nil {
			t.Errorf("due = %v, want nil", got.Due)
		}
		if got.DueSource != domain.DueSourceManual {
			t.Errorf("due_source = %q, want manual", got.DueSource)
		}
	})

	t.Run("an impossible date is refused", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))
		impossible := domain.NewDate(2026, time.February, 30)

		if _, err := f.tasks.SetDue(ctx, n.ID, &impossible); !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("SetDue(2026-02-30) = %v, want domain.ErrInvalid", err)
		}
		if got := f.get(n.ID).Due; got != nil {
			t.Errorf("due = %v, want it unwritten", got)
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		f := newFixture(t)

		if _, err := f.tasks.SetDue(ctx, "ghost", nil); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("SetDue(ghost) = %v, want store.ErrNotFound", err)
		}
	})
}

// ---------------------------------------------------------------------------
// MoveNode

func TestMoveNode(t *testing.T) {
	ctx := context.Background()

	t.Run("re-parents the node and takes its subtree with it", func(t *testing.T) {
		f := newFixture(t)
		a := f.create(draft("a", domain.NodeTypeProject, nil))
		b := f.create(draft("b", domain.NodeTypeProject, nil))
		child := f.create(draft("child", domain.NodeTypeTask, &a.ID))
		grand := f.create(draft("grand", domain.NodeTypeTask, &child.ID))

		moved, err := f.tasks.MoveNode(ctx, child.ID, &b.ID, 0)
		if err != nil {
			t.Fatalf("MoveNode: %v", err)
		}
		if moved.ParentID == nil || *moved.ParentID != b.ID {
			t.Errorf("parent_id = %v, want %q", moved.ParentID, b.ID)
		}
		if got := f.get(grand.ID); got.ParentID == nil || *got.ParentID != child.ID {
			t.Errorf("the grandchild was re-parented: %v", got.ParentID)
		}
		if got := f.get(grand.ID).Status; got != domain.StatusBacklog {
			t.Errorf("the grandchild's status changed to %q; a move must not touch it", got)
		}
	})

	t.Run("re-orders the range it joins, densely", func(t *testing.T) {
		f := newFixture(t)
		parent := f.create(draft("parent", domain.NodeTypeProject, nil))
		first := f.create(draft("first", domain.NodeTypeTask, &parent.ID))
		second := f.create(draft("second", domain.NodeTypeTask, &parent.ID))
		third := f.create(draft("third", domain.NodeTypeTask, &parent.ID))

		if _, err := f.tasks.MoveNode(ctx, third.ID, &parent.ID, 0); err != nil {
			t.Fatalf("MoveNode: %v", err)
		}
		want := map[string]int{third.ID: 0, first.ID: 1, second.ID: 2}
		for id, order := range want {
			if got := f.get(id).SortOrder; got != order {
				t.Errorf("%q: sort_order = %d, want %d", id, got, order)
			}
		}
	})

	t.Run("a node dragged past the end lands at the end", func(t *testing.T) {
		f := newFixture(t)
		parent := f.create(draft("parent", domain.NodeTypeProject, nil))
		first := f.create(draft("first", domain.NodeTypeTask, &parent.ID))
		f.create(draft("second", domain.NodeTypeTask, &parent.ID))

		if _, err := f.tasks.MoveNode(ctx, first.ID, &parent.ID, 99); err != nil {
			t.Fatalf("MoveNode: %v", err)
		}
		if got := f.get(first.ID).SortOrder; got != 1 {
			t.Errorf("sort_order = %d, want 1 (the end)", got)
		}
	})

	t.Run("a node can be moved to the root", func(t *testing.T) {
		f := newFixture(t)
		parent := f.create(draft("parent", domain.NodeTypeProject, nil))
		child := f.create(draft("child", domain.NodeTypeTask, &parent.ID))

		moved, err := f.tasks.MoveNode(ctx, child.ID, nil, 0)
		if err != nil {
			t.Fatalf("MoveNode: %v", err)
		}
		if moved.ParentID != nil {
			t.Errorf("parent_id = %v, want nil", moved.ParentID)
		}
		if moved.SortOrder != 0 {
			t.Errorf("sort_order = %d, want 0", moved.SortOrder)
		}
	})

	t.Run("a sibling that did not move keeps its updated_at", func(t *testing.T) {
		f := newFixture(t)
		parent := f.create(draft("parent", domain.NodeTypeProject, nil))
		first := f.create(draft("first", domain.NodeTypeTask, &parent.ID))
		second := f.create(draft("second", domain.NodeTypeTask, &parent.ID))

		f.now = testNow.Add(time.Hour)
		if _, err := f.tasks.MoveNode(ctx, second.ID, &parent.ID, 1); err != nil {
			t.Fatalf("MoveNode: %v", err)
		}
		if got := f.get(first.ID).UpdatedAt; !got.Equal(testNow) {
			t.Errorf("first.updated_at = %v, want the original %v", got, testNow)
		}
	})
}

// The circular-parent rejection, end to end: an error the caller can match, and
// a database that did not change — the rollback, not just the error.
func TestMoveNodeRejectsACircularParent(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	root := f.create(draft("root", domain.NodeTypeProject, nil))
	child := f.create(draft("child", domain.NodeTypeProject, &root.ID))
	grand := f.create(draft("grand", domain.NodeTypeProject, &child.ID))
	before := f.all()

	t.Run("under its own grandchild", func(t *testing.T) {
		_, err := f.tasks.MoveNode(ctx, root.ID, &grand.ID, 0)
		if !errors.Is(err, domain.ErrCircularParent) {
			t.Fatalf("MoveNode = %v, want domain.ErrCircularParent", err)
		}
		if after := f.all(); !reflect.DeepEqual(before, after) {
			t.Errorf("the database changed:\n got %+v\nwant %+v", after, before)
		}
	})

	t.Run("under itself", func(t *testing.T) {
		_, err := f.tasks.MoveNode(ctx, root.ID, &root.ID, 0)
		if !errors.Is(err, domain.ErrCircularParent) {
			t.Fatalf("MoveNode = %v, want domain.ErrCircularParent", err)
		}
		if after := f.all(); !reflect.DeepEqual(before, after) {
			t.Error("the database changed on a refused move")
		}
	})

	t.Run("under a note", func(t *testing.T) {
		memo := f.create(draft("memo", domain.NodeTypeNote, nil))
		snapshot := f.all()

		_, err := f.tasks.MoveNode(ctx, root.ID, &memo.ID, 0)
		if !errors.Is(err, domain.ErrNoteParent) {
			t.Fatalf("MoveNode = %v, want domain.ErrNoteParent", err)
		}
		if after := f.all(); !reflect.DeepEqual(snapshot, after) {
			t.Error("the database changed on a refused move")
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		if _, err := f.tasks.MoveNode(ctx, "ghost", nil, 0); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("MoveNode = %v, want domain.ErrNodeNotFound", err)
		}
	})
}

// ---------------------------------------------------------------------------
// Rollback

// errBoom is the injected failure: a statement that fails for a reason that has
// nothing to do with the data, part way through a multi-statement write.
var errBoom = errors.New("boom: the injected executor refused a statement")

// failingTx is a real transaction whose ExecContext starts failing after a
// given number of successful calls. Everything else — including Rollback — is
// the real thing, so a test asserts a real rollback against a real database.
type failingTx struct {
	service.Tx
	remaining *int
}

func (tx failingTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if *tx.remaining <= 0 {
		return nil, errBoom
	}
	*tx.remaining--
	return tx.Tx.ExecContext(ctx, query, args...)
}

// failingBeginner hands out failingTx instead of the transaction it wraps.
type failingBeginner struct {
	inner service.Beginner
	after int
}

func (b failingBeginner) Begin(ctx context.Context) (service.Tx, error) {
	tx, err := b.inner.Begin(ctx)
	if err != nil {
		return nil, err
	}
	budget := b.after
	return failingTx{Tx: tx, remaining: &budget}, nil
}

// A failure part way through a cascade takes the whole move with it: the due
// date written by step 2 is rolled back along with the statuses of step 3.
func TestMoveToColumnRollsBackAPartialCascade(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	ids := f.headline()
	before := f.all()

	// The clock moves on, so that a due date AND an updated_at that survived
	// the failure would both be visible in the comparison below.
	f.now = testNow.Add(time.Hour)

	// One successful Exec — the due-date write — then the cascade's own
	// statement fails.
	brittle := service.NewTaskService(
		failingBeginner{inner: f.begin, after: 1}, f.nodes, f.tags, f.clock(), f.nextID)

	_, err := brittle.MoveToColumn(ctx, ids["p"], domain.StatusToday)
	if !errors.Is(err, errBoom) {
		t.Fatalf("MoveToColumn = %v, want the injected errBoom", err)
	}

	t.Run("not one row changed", func(t *testing.T) {
		if after := f.all(); !reflect.DeepEqual(before, after) {
			t.Errorf("the database changed:\n got %+v\nwant %+v", after, before)
		}
	})

	t.Run("the due date written before the failure is gone too", func(t *testing.T) {
		if got := f.get(ids["p"]).Due; got != nil {
			t.Errorf("due = %v, want nil — the first statement must roll back as well", got)
		}
	})
}

func TestCreateNodeRollsBackAFailedInsert(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	brittle := service.NewTaskService(
		failingBeginner{inner: f.begin, after: 0}, f.nodes, f.tags, f.clock(), f.nextID)

	if _, err := brittle.CreateNode(ctx, draft("x", domain.NodeTypeTask, nil)); !errors.Is(err, errBoom) {
		t.Fatalf("CreateNode = %v, want the injected errBoom", err)
	}
	if got := len(f.all()); got != 0 {
		t.Errorf("%d rows, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Archive and restore

func TestArchiveAndRestoreCoverTheSubtree(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	ids := f.headline()

	archivedAt := testNow.Add(2 * time.Hour)
	f.now = archivedAt

	n, err := f.tasks.ArchiveNode(ctx, ids["p"])
	if err != nil {
		t.Fatalf("ArchiveNode: %v", err)
	}

	t.Run("the whole subtree is archived at the injected clock", func(t *testing.T) {
		if n != len(ids) {
			t.Errorf("archived %d nodes, want %d", n, len(ids))
		}
		for name, id := range ids {
			got := f.get(id)
			if got.ArchivedAt == nil {
				t.Fatalf("%s: archived_at is NULL", name)
			}
			if !got.ArchivedAt.Equal(archivedAt) {
				t.Errorf("%s: archived_at = %v, want %v", name, got.ArchivedAt, archivedAt)
			}
		}
	})

	t.Run("archived rows are hidden from the default listing", func(t *testing.T) {
		live, err := f.nodes.ListAll(ctx, false)
		if err != nil {
			t.Fatalf("ListAll: %v", err)
		}
		if len(live) != 0 {
			t.Errorf("%d live rows, want 0", len(live))
		}
	})

	t.Run("archiving again is a no-op that keeps the original stamp", func(t *testing.T) {
		f.now = archivedAt.Add(time.Hour)
		count, err := f.tasks.ArchiveNode(ctx, ids["p"])
		if err != nil {
			t.Fatalf("ArchiveNode: %v", err)
		}
		if count != 0 {
			t.Errorf("archived %d rows the second time, want 0", count)
		}
		if got := f.get(ids["deep"]).ArchivedAt; got == nil || !got.Equal(archivedAt) {
			t.Errorf("archived_at = %v, want the original %v", got, archivedAt)
		}
		f.now = archivedAt
	})

	t.Run("restore brings the whole subtree back", func(t *testing.T) {
		count, err := f.tasks.RestoreNode(ctx, ids["p"])
		if err != nil {
			t.Fatalf("RestoreNode: %v", err)
		}
		if count != len(ids) {
			t.Errorf("restored %d nodes, want %d", count, len(ids))
		}
		for name, id := range ids {
			if got := f.get(id).ArchivedAt; got != nil {
				t.Errorf("%s: archived_at = %v, want NULL", name, got)
			}
		}
	})

	t.Run("restoring a live subtree is a no-op", func(t *testing.T) {
		count, err := f.tasks.RestoreNode(ctx, ids["p"])
		if err != nil {
			t.Fatalf("RestoreNode: %v", err)
		}
		if count != 0 {
			t.Errorf("restored %d rows, want 0", count)
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		if _, err := f.tasks.ArchiveNode(ctx, "ghost"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("ArchiveNode(ghost) = %v, want domain.ErrNodeNotFound", err)
		}
		if _, err := f.tasks.RestoreNode(ctx, "ghost"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("RestoreNode(ghost) = %v, want domain.ErrNodeNotFound", err)
		}
	})
}

// A subtree archived on its own keeps its own archived_at when its parent is
// archived later, and comes back with the parent all the same.
func TestArchiveKeepsAnEarlierStamp(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	parent := f.create(draft("parent", domain.NodeTypeProject, nil))
	child := f.create(draft("child", domain.NodeTypeTask, &parent.ID))

	lastMonth := testNow.AddDate(0, -1, 0)
	f.now = lastMonth
	if _, err := f.tasks.ArchiveNode(ctx, child.ID); err != nil {
		t.Fatalf("ArchiveNode(child): %v", err)
	}

	f.now = testNow
	if _, err := f.tasks.ArchiveNode(ctx, parent.ID); err != nil {
		t.Fatalf("ArchiveNode(parent): %v", err)
	}

	if got := f.get(child.ID).ArchivedAt; got == nil || !got.Equal(lastMonth) {
		t.Errorf("the child's archived_at = %v, want its own %v", got, lastMonth)
	}
	if got := f.get(parent.ID).ArchivedAt; got == nil || !got.Equal(testNow) {
		t.Errorf("the parent's archived_at = %v, want %v", got, testNow)
	}

	if _, err := f.tasks.RestoreNode(ctx, parent.ID); err != nil {
		t.Fatalf("RestoreNode: %v", err)
	}
	if got := f.get(child.ID).ArchivedAt; got != nil {
		t.Errorf("the child is still archived (%v); restore means the whole subtree", got)
	}
}

// ---------------------------------------------------------------------------
// The failure paths of the unit of work itself.

// errRollback is the second failure: the undo failing after the work failed.
var errRollback = errors.New("boom: the injected transaction refused to roll back")

// blockedBeginner cannot start a transaction at all — the disk is gone, the
// file is locked, the process is out of handles.
type blockedBeginner struct{}

func (blockedBeginner) Begin(context.Context) (service.Tx, error) { return nil, errBoom }

// queryFailingTx is a transaction whose reads fail: every write path starts by
// loading the node set, and that load can fail like anything else.
type queryFailingTx struct{ service.Tx }

func (tx queryFailingTx) QueryContext(context.Context, string, ...any) (*sql.Rows, error) {
	return nil, errBoom
}

func (tx queryFailingTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	// A row that always reports errBoom, built by asking the real transaction
	// for a statement that cannot parse.
	return tx.Tx.QueryRowContext(ctx, "SELECT this is not sql")
}

// rollbackFailingTx fails its statements and then fails to undo them.
type rollbackFailingTx struct{ service.Tx }

func (tx rollbackFailingTx) ExecContext(context.Context, string, ...any) (sql.Result, error) {
	return nil, errBoom
}

func (tx rollbackFailingTx) Rollback() error { return errRollback }

// wrappingBeginner applies wrap to every transaction it hands out.
type wrappingBeginner struct {
	inner service.Beginner
	wrap  func(service.Tx) service.Tx
}

func (b wrappingBeginner) Begin(ctx context.Context) (service.Tx, error) {
	tx, err := b.inner.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return b.wrap(tx), nil
}

// Every write path reports the failure it met rather than pretending to have
// worked, whichever half of the unit of work broke.
func TestWritePathsSurfaceTransactionFailures(t *testing.T) {
	ctx := context.Background()

	beginners := []struct {
		name string
		of   func(f *fixture) service.Beginner
	}{
		{"the transaction cannot be started", func(*fixture) service.Beginner {
			return blockedBeginner{}
		}},
		{"the load of the node set fails", func(f *fixture) service.Beginner {
			return wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
				return queryFailingTx{Tx: tx}
			}}
		}},
		{"the write fails and so does the rollback", func(f *fixture) service.Beginner {
			return wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
				return rollbackFailingTx{Tx: tx}
			}}
		}},
	}

	// Each call is given the fixture so that it can name the node it needs; the
	// second root exists so that MoveNode has a re-ordering to write rather
	// than a no-op to skip.
	calls := []struct {
		name string
		call func(f *fixture, svc *service.TaskService) error
	}{
		{"CreateNode", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.CreateNode(ctx, draft("x", domain.NodeTypeTask, nil))
			return err
		}},
		{"MoveNode", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.MoveNode(ctx, "n2", nil, 0)
			return err
		}},
		{"MoveToColumn", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.MoveToColumn(ctx, "n1", domain.StatusToday)
			return err
		}},
		{"SetDue", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.SetDue(ctx, "n1", nil)
			return err
		}},
		{"ArchiveNode", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.ArchiveNode(ctx, "n1")
			return err
		}},
		{"RestoreNode", func(f *fixture, svc *service.TaskService) error {
			_, err := svc.RestoreNode(ctx, "n1")
			return err
		}},
	}

	for _, b := range beginners {
		t.Run(b.name, func(t *testing.T) {
			for _, c := range calls {
				t.Run(c.name, func(t *testing.T) {
					f := newFixture(t)
					first := f.create(draft("seed", domain.NodeTypeTask, nil))
					f.create(draft("sibling", domain.NodeTypeTask, nil))
					// The node has to be archived for RestoreNode to have
					// anything to write.
					if c.name == "RestoreNode" {
						if _, err := f.tasks.ArchiveNode(ctx, first.ID); err != nil {
							t.Fatalf("ArchiveNode: %v", err)
						}
					}

					brittle := service.NewTaskService(b.of(f), f.nodes, f.tags, f.clock(), f.nextID)
					if err := c.call(f, brittle); err == nil {
						t.Fatal("the call succeeded; want the injected failure")
					}
				})
			}
		})
	}
}

// A failed rollback is reported alongside the failure that caused it: the user
// needs the first, the log needs the second.
func TestAFailedRollbackIsJoinedOntoTheOriginalError(t *testing.T) {
	f := newFixture(t)
	n := f.create(draft("seed", domain.NodeTypeTask, nil))

	brittle := service.NewTaskService(
		wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return rollbackFailingTx{Tx: tx}
		}}, f.nodes, f.tags, f.clock(), f.nextID)

	_, err := brittle.MoveToColumn(context.Background(), n.ID, domain.StatusToday)
	if !errors.Is(err, errBoom) {
		t.Errorf("err = %v, does not carry the original failure", err)
	}
	if !errors.Is(err, errRollback) {
		t.Errorf("err = %v, does not carry the rollback failure", err)
	}
}

// The commit itself can fail, and the caller must hear about it rather than
// being told the write landed.
func TestAFailedCommitIsReported(t *testing.T) {
	f := newFixture(t)

	brittle := service.NewTaskService(
		wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return commitFailingTx{Tx: tx}
		}}, f.nodes, f.tags, f.clock(), f.nextID)

	if _, err := brittle.CreateNode(context.Background(), draft("x", domain.NodeTypeTask, nil)); !errors.Is(err, errCommit) {
		t.Fatalf("CreateNode = %v, want the injected commit failure", err)
	}
}

var errCommit = errors.New("boom: the injected transaction refused to commit")

type commitFailingTx struct{ service.Tx }

func (tx commitFailingTx) Commit() error {
	// Undo the real work so the database is left as the failed commit claims.
	if err := tx.Tx.Rollback(); err != nil {
		return err
	}
	return errCommit
}
