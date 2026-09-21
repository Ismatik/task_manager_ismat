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
	f.tasks = service.NewTaskService(f.begin, f.nodes, f.tags, f.timers(), f.clock(), f.nextID)
	return f
}

// tasksOver returns a task service over a substitute Beginner — the brittle and
// blocked transactions the failure-path tests inject — with everything else the
// fixture's. The timer collaborator is required (D13, S2-03), and it is the
// fixture's own: the coupled path runs on the executor the TASK service's
// transaction hands it, so an injected failure reaches the timer too.
func (f *fixture) tasksOver(begin service.Beginner) *service.TaskService {
	f.t.Helper()

	return service.NewTaskService(begin, f.nodes, f.tags, f.timers(), f.clock(), f.nextID)
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
		if !errors.Is(err, domain.ErrTypeHasNoChildren) {
			t.Fatalf("CreateNode() = %v, want domain.ErrTypeHasNoChildren", err)
		}
		if got := len(f.all()); got != 1 {
			t.Errorf("%d rows, want only the note", got)
		}
	})

	// S1-09. The create path is the door the reviewer came through: a habit has
	// no column either, so nothing may be parented under one, and the create
	// path must refuse it for the same reason and with the same sentinel as a
	// note. Every type is refused, because the rule is about the PARENT.
	t.Run("a habit cannot be a parent", func(t *testing.T) {
		for _, typ := range domain.NodeTypes() {
			t.Run(string(typ), func(t *testing.T) {
				f := newFixture(t)

				hd := draft("stretch every morning", domain.NodeTypeHabit, nil)
				hd.Recurrence = ptr("FREQ=DAILY")
				h := f.create(hd)

				d := draft("x", typ, &h.ID)
				if typ == domain.NodeTypeHabit {
					d.Recurrence = ptr("FREQ=DAILY")
				}
				_, err := f.tasks.CreateNode(ctx, d)
				if !errors.Is(err, domain.ErrTypeHasNoChildren) {
					t.Fatalf("CreateNode(%s under a habit) = %v, want domain.ErrTypeHasNoChildren",
						typ, err)
				}
				if got := len(f.all()); got != 1 {
					t.Errorf("%d rows, want only the habit", got)
				}
			})
		}
	})

	// The user's earlier decision is the opposite direction and is untouched:
	// grouping a habit under a project is legal.
	t.Run("a habit under a project is still created", func(t *testing.T) {
		f := newFixture(t)

		p := f.create(draft("project", domain.NodeTypeProject, nil))
		d := draft("stretch every morning", domain.NodeTypeHabit, &p.ID)
		d.Recurrence = ptr("FREQ=DAILY")

		h, err := f.tasks.CreateNode(ctx, d)
		if err != nil {
			t.Fatalf("CreateNode(habit under a project) = %v, want it created", err)
		}
		if h.ParentID == nil || *h.ParentID != p.ID {
			t.Errorf("the habit's parent = %v, want %q", h.ParentID, p.ID)
		}
	})
}

// The whole type x status matrix on the create path, against a real database.
//
// Creating is the second door into a stored status and it used to be unlocked:
// CreateNode{project, doing} stored a project with status = doing and the Board
// then rendered it in the Doing column, while MoveToColumn(project, doing) had
// always refused the identical state (D9). The same holds, less loudly, for a
// note or a habit given a column: PLAN.md §4 gives a note "no status, no due"
// and keeps a habit out of the columns entirely.
//
// Every case asserts the DATABASE, not just the returned error: a create that
// is refused must leave no row behind, and a create that is allowed must store
// the status it was asked for.
func TestCreateNodeTypeAndStatusCombinations(t *testing.T) {
	ctx := context.Background()

	// A habit needs a recurrence to get past Validate, which is a different
	// rule and not the one under test here.
	newDraft := func(typ domain.NodeType, status domain.Status) service.NewNode {
		d := draft("x", typ, nil)
		d.Status = status
		if typ == domain.NodeTypeHabit {
			d.Recurrence = ptr("FREQ=DAILY")
		}
		return d
	}

	// wantErr nil means the combination is legal and must be stored.
	want := func(typ domain.NodeType, status domain.Status) error {
		switch {
		case typ == domain.NodeTypeNote || typ == domain.NodeTypeHabit:
			if status != domain.StatusBacklog {
				return domain.ErrTypeHasNoColumn
			}
		case typ == domain.NodeTypeProject && status == domain.StatusDoing:
			return domain.ErrProjectNeverDoing
		}
		return nil
	}

	for _, typ := range domain.NodeTypes() {
		for _, status := range domain.Statuses() {
			t.Run(typ.String()+" in "+status.String(), func(t *testing.T) {
				f := newFixture(t)

				n, err := f.tasks.CreateNode(ctx, newDraft(typ, status))
				wantErr := want(typ, status)

				if wantErr != nil {
					if !errors.Is(err, wantErr) {
						t.Fatalf("CreateNode(%s, %s) = %v, want %v", typ, status, err, wantErr)
					}
					// The database, not the return value: the bug was a row.
					if got := f.all(); len(got) != 0 {
						t.Fatalf("%d rows written by a refused create, want 0: %+v", len(got), got)
					}
					return
				}

				if err != nil {
					t.Fatalf("CreateNode(%s, %s) = %v, want it stored", typ, status, err)
				}
				if got := f.get(n.ID).Status; got != status {
					t.Errorf("stored status = %q, want %q", got, status)
				}
			})
		}
	}
}

// The reviewer's exact reproduction, kept as its own named case: the project
// must not reach the board through the create path either.
func TestCreateNodeProjectInDoingNeverReachesTheBoard(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	d := draft("ship it", domain.NodeTypeProject, nil)
	d.Status = domain.StatusDoing

	if _, err := f.tasks.CreateNode(ctx, d); !errors.Is(err, domain.ErrProjectNeverDoing) {
		t.Fatalf("CreateNode(project, doing) = %v, want domain.ErrProjectNeverDoing", err)
	}
	if got := f.all(); len(got) != 0 {
		t.Fatalf("%d rows written, want 0: %+v", len(got), got)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	for _, col := range board {
		if len(col.Nodes) != 0 {
			t.Errorf("the %s column has %d cards, want none", col.Status, len(col.Nodes))
		}
	}
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

// Every node type against every target column, on a real database.
//
// MoveToColumn used to run PlanCascade — which correctly plans NO change for a
// note — and then write DueForColumnMove's answer unconditionally, so a type
// with no column came out of a drag in a state nothing had agreed to:
//
//	note   after MoveToColumn(today): status=backlog due=2026-09-18 due_source=auto
//	habit  after MoveToColumn(today): status=today   due=2026-09-18 due_source=auto
//
// The clock is moved on before every move, so a write that survived would show
// up in updated_at even if it wrote the same due date twice.
func TestMoveToColumnEveryTypeAgainstEveryColumn(t *testing.T) {
	ctx := context.Background()

	want := func(typ domain.NodeType, target domain.Status) error {
		switch {
		case !typ.HasColumn():
			return domain.ErrTypeHasNoColumn
		case typ == domain.NodeTypeProject && target == domain.StatusDoing:
			return domain.ErrProjectNeverDoing
		}
		return nil
	}

	for _, typ := range domain.NodeTypes() {
		for _, target := range domain.Statuses() {
			t.Run(typ.String()+" to "+target.String(), func(t *testing.T) {
				f := newFixture(t)

				d := draft("x", typ, nil)
				if typ == domain.NodeTypeHabit {
					d.Recurrence = ptr("FREQ=DAILY")
				}
				n := f.create(d)

				before := f.all()
				f.now = testNow.Add(time.Hour)

				moved, err := f.tasks.MoveToColumn(ctx, n.ID, target)

				if wantErr := want(typ, target); wantErr != nil {
					if !errors.Is(err, wantErr) {
						t.Fatalf("MoveToColumn(%s, %s) = %v, want %v", typ, target, err, wantErr)
					}
					// Every field of every row, not just the ones the reviewer
					// caught: status, due, due_source and updated_at included.
					if after := f.all(); !reflect.DeepEqual(before, after) {
						t.Errorf("the database changed on a refused move:\n got %+v\nwant %+v", after, before)
					}
					return
				}

				if err != nil {
					t.Fatalf("MoveToColumn(%s, %s) = %v, want it to move", typ, target, err)
				}
				if moved.Status != target {
					t.Errorf("status = %q, want %q", moved.Status, target)
				}
			})
		}
	}
}

// The reviewer's exact reproduction, field by field: the two types with no
// column come out of a drag to Today with the due date they went in with.
func TestMoveToColumnDoesNotDateATypeWithNoColumn(t *testing.T) {
	ctx := context.Background()

	for _, typ := range []domain.NodeType{domain.NodeTypeNote, domain.NodeTypeHabit} {
		t.Run(typ.String(), func(t *testing.T) {
			f := newFixture(t)

			d := draft("x", typ, nil)
			if typ == domain.NodeTypeHabit {
				d.Recurrence = ptr("FREQ=DAILY")
			}
			n := f.create(d)
			f.now = testNow.Add(time.Hour)

			if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusToday); !errors.Is(err, domain.ErrTypeHasNoColumn) {
				t.Fatalf("MoveToColumn(%s, today) = %v, want domain.ErrTypeHasNoColumn", typ, err)
			}

			got := f.get(n.ID)
			if got.Due != nil {
				t.Errorf("due = %v, want it still unset — a %s has no column to date it by", got.Due, typ)
			}
			if got.DueSource != domain.DueSourceManual {
				t.Errorf("due_source = %q, want manual", got.DueSource)
			}
			if got.Status != domain.StatusBacklog {
				t.Errorf("status = %q, want the inert backlog it was created with", got.Status)
			}
			if !got.UpdatedAt.Equal(testNow) {
				t.Errorf("updated_at = %v, want the original %v — nothing was written", got.UpdatedAt, testNow)
			}
		})
	}
}

// The reviewer's reproduction of the third door into an illegal status: the
// drag root is a legal one, and the CASCADE underneath it wrote a column status
// onto a habit child — `habit` with status "today" is precisely the row the
// no-column rule exists to make impossible.
//
// Every one of the five columns is dragged, and the habit's whole row is
// compared field by field: status, due, due_source and updated_at included.
func TestMoveToColumnDoesNotCascadeOntoAHabitChild(t *testing.T) {
	ctx := context.Background()

	for _, target := range domain.Statuses() {
		t.Run("to "+target.String(), func(t *testing.T) {
			f := newFixture(t)

			// A task parent, not a project, so that the drag to doing is legal
			// in the first place (D9) and the cascade really runs.
			parent := f.create(draft("parent", domain.NodeTypeTask, nil))
			hd := draft("stretch", domain.NodeTypeHabit, &parent.ID)
			hd.Recurrence = ptr("FREQ=DAILY")
			h := f.create(hd)
			sibling := f.create(draft("real work", domain.NodeTypeTask, &parent.ID))

			before := f.get(h.ID)
			f.now = testNow.Add(time.Hour)

			if _, err := f.tasks.MoveToColumn(ctx, parent.ID, target); err != nil {
				t.Fatalf("MoveToColumn(parent, %s) = %v", target, err)
			}

			if after := f.get(h.ID); !reflect.DeepEqual(after, before) {
				t.Errorf("the habit row changed:\n got %+v\nwant %+v", after, before)
			}

			// The positive control: the drag did do its job on the sibling that
			// does have a column, so the assertion above is not passing because
			// nothing happened at all.
			if got := f.get(sibling.ID).Status; got != target {
				t.Fatalf("the task sibling is %q, want %q — the cascade did not run", got, target)
			}

			// And the illegal row cannot reach the board even by derivation.
			board, err := f.tasks.Board(ctx)
			if err != nil {
				t.Fatalf("Board: %v", err)
			}
			for _, col := range board {
				for _, v := range col.Nodes {
					if v.Node.ID == h.ID {
						t.Errorf("the habit is on the board in the %s column", col.Status)
					}
				}
			}
		})
	}
}

// The other half of what the corrupt status corrupted: a habit is not work, so
// a project holding one finished task and one habit is finished — not 1 of 2.
func TestProgressDoesNotCountAHabit(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("p", domain.NodeTypeProject, nil))
	task := f.create(draft("real work", domain.NodeTypeTask, &p.ID))
	hd := draft("stretch", domain.NodeTypeHabit, &p.ID)
	hd.Recurrence = ptr("FREQ=DAILY")
	f.create(hd)

	if _, err := f.tasks.MoveToColumn(ctx, task.ID, domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(task, done) = %v", err)
	}

	got, err := f.tasks.Progress(ctx, p.ID)
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	want := service.ProgressView{Done: 1, Total: 1, Defined: true, Percent: 100}
	if got != want {
		t.Errorf("Progress(p) = %+v, want %+v — the habit is not a unit of work", got, want)
	}
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
// SetPriority

func TestSetPriority(t *testing.T) {
	ctx := context.Background()

	t.Run("every valid priority round-trips", func(t *testing.T) {
		// domain.Priorities() rather than 1, 2, 3, 4 written out: the set is
		// the domain's and this test is not a second copy of it. If a fifth
		// priority is ever added, this case covers it without being edited.
		for _, want := range domain.Priorities() {
			t.Run(want.String(), func(t *testing.T) {
				f := newFixture(t)
				n := f.create(draft("x", domain.NodeTypeTask, nil))

				got, err := f.tasks.SetPriority(ctx, n.ID, want)
				if err != nil {
					t.Fatalf("SetPriority(%s): %v", want, err)
				}
				if got.Priority != want {
					t.Errorf("returned priority = %d, want %d", got.Priority, want)
				}
				// The returned value is one thing; what is in the database is
				// the claim. Read it back independently.
				if stored := f.get(n.ID).Priority; stored != want {
					t.Errorf("stored priority = %d, want %d", stored, want)
				}
			})
		}
	})

	t.Run("it stamps updated_at and touches nothing else", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))

		f.now = testNow.Add(90 * time.Minute)

		got, err := f.tasks.SetPriority(ctx, n.ID, domain.Priority1)
		if err != nil {
			t.Fatalf("SetPriority: %v", err)
		}
		if !got.UpdatedAt.Equal(f.now) {
			t.Errorf("updated_at = %v, want the injected clock's %v", got.UpdatedAt, f.now)
		}
		if !got.CreatedAt.Equal(testNow) {
			t.Errorf("created_at = %v, want it untouched at %v", got.CreatedAt, testNow)
		}

		// The one field, and only the one field. Everything the edit had no
		// business in is compared against the node as it was created.
		before, after := n, got
		before.Priority, before.UpdatedAt = after.Priority, after.UpdatedAt
		if !reflect.DeepEqual(before, after) {
			t.Errorf("SetPriority changed more than the priority:\n\tbefore %+v\n\tafter  %+v", before, after)
		}
	})

	t.Run("an out-of-range priority is refused by the domain, and nothing is written", func(t *testing.T) {
		// The range is domain.Priority.Valid's, reached through Node.Validate:
		// the service states it nowhere, so this is the domain refusing, not a
		// service-level guard that happens to agree with it.
		for _, bad := range []domain.Priority{-1, 0, 5, 99} {
			t.Run(bad.String(), func(t *testing.T) {
				f := newFixture(t)
				n := f.create(draft("x", domain.NodeTypeTask, nil))

				if _, err := f.tasks.SetPriority(ctx, n.ID, bad); !errors.Is(err, domain.ErrInvalid) {
					t.Fatalf("SetPriority(%d) = %v, want domain.ErrInvalid", bad, err)
				}

				// The DATABASE, not the return value. A method that refused and
				// wrote anyway would pass a test that only read what it handed
				// back.
				stored := f.get(n.ID)
				if stored.Priority != n.Priority {
					t.Errorf("stored priority = %d, want the original %d — nothing should have been written",
						stored.Priority, n.Priority)
				}
				if !stored.UpdatedAt.Equal(n.UpdatedAt) {
					t.Errorf("updated_at = %v, want the original %v — the row was touched by a refused edit",
						stored.UpdatedAt, n.UpdatedAt)
				}
			})
		}
	})

	t.Run("a zero is refused rather than read as the default 4", func(t *testing.T) {
		// NewNode reads a zero priority as 4 because a draft is a form with
		// blanks in it. An edit is not a form, and the two readings must not be
		// confused: created at the default, an edit to 0 leaves it alone AND
		// fails, rather than silently "succeeding" at the value it already had.
		f := newFixture(t)
		n := f.create(draft("x", domain.NodeTypeTask, nil))
		if n.Priority != domain.Priority4 {
			t.Fatalf("the fixture node starts at %d, want the default 4", n.Priority)
		}

		if _, err := f.tasks.SetPriority(ctx, n.ID, 0); err == nil {
			t.Fatal("SetPriority(0) succeeded, want it refused — 0 is not 'the default' on an edit")
		}
	})

	t.Run("every type may be prioritised, including the ones with no column", func(t *testing.T) {
		// Unlike a due date, a priority has no type rule: PLAN.md §4 excludes a
		// note from status and due, and a habit from columns, and says nothing
		// about priority. If this ever starts failing, a predicate was added.
		f := newFixture(t)

		for _, typ := range domain.NodeTypes() {
			d := draft(typ.String(), typ, nil)
			if typ == domain.NodeTypeHabit {
				d.Recurrence = ptr("FREQ=DAILY")
			}
			n := f.create(d)

			got, err := f.tasks.SetPriority(ctx, n.ID, domain.Priority2)
			if err != nil {
				t.Errorf("SetPriority on a %s: %v", typ, err)
				continue
			}
			if got.Priority != domain.Priority2 {
				t.Errorf("%s: priority = %d, want 2", typ, got.Priority)
			}
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		f := newFixture(t)

		if _, err := f.tasks.SetPriority(ctx, "ghost", domain.Priority1); !errors.Is(err, store.ErrNotFound) {
			t.Fatalf("SetPriority(ghost) = %v, want store.ErrNotFound", err)
		}
	})
}

// PLAN.md §4 spells a note out as "no status, no due". The status half was
// locked; the due half was not, and both doors were open — CreateNode stored the
// date it was handed and SetDue wrote one afterwards.
//
// A HABIT is the control, and it is deliberately NOT refused: §4 says only that
// a habit never appears in a column, and nothing anywhere says a habit may not
// be due on a date. D9's prose over-claimed; the prose is what was corrected.
func TestANoteNeverReceivesADueDate(t *testing.T) {
	ctx := context.Background()
	date := domain.NewDate(2026, time.September, 18)

	t.Run("the create path refuses it and writes no row", func(t *testing.T) {
		f := newFixture(t)

		d := draft("memo", domain.NodeTypeNote, nil)
		d.Due = &date

		if _, err := f.tasks.CreateNode(ctx, d); !errors.Is(err, domain.ErrTypeHasNoDue) {
			t.Fatalf("CreateNode(note, due) = %v, want domain.ErrTypeHasNoDue", err)
		}
		if got := f.all(); len(got) != 0 {
			t.Fatalf("%d rows written by a refused create, want 0: %+v", len(got), got)
		}
	})

	t.Run("SetDue refuses it and writes nothing", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("memo", domain.NodeTypeNote, nil))

		before := f.get(n.ID)
		f.now = testNow.Add(time.Hour)

		if _, err := f.tasks.SetDue(ctx, n.ID, &date); !errors.Is(err, domain.ErrTypeHasNoDue) {
			t.Fatalf("SetDue(note, %s) = %v, want domain.ErrTypeHasNoDue", date, err)
		}
		if after := f.get(n.ID); !reflect.DeepEqual(after, before) {
			t.Errorf("the note row changed:\n got %+v\nwant %+v", after, before)
		}
	})

	t.Run("clearing the due date of a note is still allowed", func(t *testing.T) {
		f := newFixture(t)
		n := f.create(draft("memo", domain.NodeTypeNote, nil))

		if _, err := f.tasks.SetDue(ctx, n.ID, nil); err != nil {
			t.Fatalf("SetDue(note, nil) = %v, want it to succeed — there is no date to refuse", err)
		}
	})

	t.Run("a habit may have one, by both doors", func(t *testing.T) {
		f := newFixture(t)

		d := draft("stretch", domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		d.Due = &date
		created := f.create(d)
		if created.Due == nil || !created.Due.Equal(date) {
			t.Fatalf("created habit due = %v, want %v", created.Due, date)
		}

		later := domain.NewDate(2026, time.October, 1)
		edited, err := f.tasks.SetDue(ctx, created.ID, &later)
		if err != nil {
			t.Fatalf("SetDue(habit, %s) = %v — §4 forbids a habit a COLUMN, not a date", later, err)
		}
		if edited.Due == nil || !edited.Due.Equal(later) {
			t.Errorf("habit due = %v, want %v", edited.Due, later)
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

	// The regression guard for the user's earlier decision: S1-09 refuses a
	// parent with no column, which is the OPPOSITE direction. Grouping a habit
	// under a project was ruled legal and stays legal, by drag as well as by
	// create.
	t.Run("a habit may still be dragged under a project", func(t *testing.T) {
		f := newFixture(t)
		p := f.create(draft("project", domain.NodeTypeProject, nil))
		d := draft("stretch every morning", domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		h := f.create(d)

		moved, err := f.tasks.MoveNode(ctx, h.ID, &p.ID, 0)
		if err != nil {
			t.Fatalf("MoveNode(habit under a project) = %v, want it to succeed", err)
		}
		if moved.ParentID == nil || *moved.ParentID != p.ID {
			t.Errorf("parent_id = %v, want %q", moved.ParentID, p.ID)
		}
		if got := f.get(h.ID).ParentID; got == nil || *got != p.ID {
			t.Errorf("the stored parent_id = %v, want %q", got, p.ID)
		}

		// And back out to a root again.
		if _, err := f.tasks.MoveNode(ctx, h.ID, nil, 0); err != nil {
			t.Fatalf("MoveNode(habit to a root) = %v, want it to succeed", err)
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
		if !errors.Is(err, domain.ErrTypeHasNoChildren) {
			t.Fatalf("MoveNode = %v, want domain.ErrTypeHasNoChildren", err)
		}
		if after := f.all(); !reflect.DeepEqual(snapshot, after) {
			t.Error("the database changed on a refused move")
		}
	})

	// S1-09: a habit has no column either, so it holds no children either.
	t.Run("under a habit", func(t *testing.T) {
		d := draft("stretch every morning", domain.NodeTypeHabit, nil)
		d.Recurrence = ptr("FREQ=DAILY")
		h := f.create(d)
		snapshot := f.all()

		_, err := f.tasks.MoveNode(ctx, root.ID, &h.ID, 0)
		if !errors.Is(err, domain.ErrTypeHasNoChildren) {
			t.Fatalf("MoveNode = %v, want domain.ErrTypeHasNoChildren", err)
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
	brittle := f.tasksOver(failingBeginner{inner: f.begin, after: 1})

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

	brittle := f.tasksOver(failingBeginner{inner: f.begin, after: 0})

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

					brittle := f.tasksOver(b.of(f))
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

	brittle := f.tasksOver(wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
		return rollbackFailingTx{Tx: tx}
	}})

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

	brittle := f.tasksOver(wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
		return commitFailingTx{Tx: tx}
	}})

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

// ---------------------------------------------------------------------------
// C1 / D13 — the Doing↔timer coupling.
//
// PLAN.md §4: "Moving a card to Doing opens a time_entry." Stage 1 wrote both
// halves and wired neither; these are the tests that make the wiring impossible
// to remove quietly.

// D13 §1: a direct move of a timeable leaf to doing opens exactly one entry,
// committed with the move.
func TestMoveToDoingOpensExactlyOneTimeEntry(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("the card", domain.NodeTypeTask, nil))

	if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(doing): %v", err)
	}

	if got := f.openEntries(); got != 1 {
		t.Fatalf("%d open entries, want exactly 1", got)
	}
	entries := f.entriesOf(n.ID)
	if len(entries) != 1 {
		t.Fatalf("%d entries on the card, want 1", len(entries))
	}
	if !entries[0].IsOpen() {
		t.Error("the entry is closed; the card is in Doing and the timer must be running")
	}
	if !entries[0].StartedAt.Equal(testNow) {
		t.Errorf("started_at = %s, want the injected clock's %s", entries[0].StartedAt, testNow)
	}
	if got := f.get(n.ID).Status; got != domain.StatusDoing {
		t.Errorf("the card's status = %q, want doing", got)
	}
}

// Atomicity, the half that is easy to get wrong: the move must not survive a
// timer that did not open. If it did, a card would sit in Doing with no entry
// behind it and §4's coupling would be a lie for that card for ever.
func TestMoveToDoingRollsBackWhenTheTimerInsertFails(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("the card", domain.NodeTypeTask, nil))
	before := f.all()

	f.now = testNow.Add(time.Hour)

	// Two Execs succeed — the due-date write and the cascade — and the third,
	// which is the time entry's INSERT, fails.
	brittle := f.tasksOver(failingBeginner{inner: f.begin, after: 2})

	if _, err := brittle.MoveToColumn(ctx, n.ID, domain.StatusDoing); !errors.Is(err, errBoom) {
		t.Fatalf("MoveToColumn = %v, want the injected errBoom", err)
	}

	if got := f.openEntries(); got != 0 {
		t.Errorf("%d open entries, want 0", got)
	}
	if after := f.all(); !reflect.DeepEqual(before, after) {
		t.Errorf("the move survived a timer that did not open:\n got %+v\nwant %+v", after, before)
	}
	if got := f.get(n.ID).Status; got == domain.StatusDoing {
		t.Error("the card is in Doing with no time entry behind it")
	}
}

// The single-active invariant, re-asserted through the coupled path: it is the
// same TimerService.start underneath, not a second copy of the policy.
func TestMovingASecondCardToDoingClosesTheFirstsEntry(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	a := f.create(draft("a", domain.NodeTypeTask, nil))
	b := f.create(draft("b", domain.NodeTypeTask, nil))

	if _, err := f.tasks.MoveToColumn(ctx, a.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(a, doing): %v", err)
	}
	f.now = testNow.Add(10 * time.Minute)
	if _, err := f.tasks.MoveToColumn(ctx, b.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(b, doing): %v", err)
	}

	if got := f.openEntries(); got != 1 {
		t.Fatalf("%d open entries, want exactly 1 — never two", got)
	}

	entriesA := f.entriesOf(a.ID)
	if len(entriesA) != 1 || entriesA[0].IsOpen() {
		t.Errorf("A's entries = %+v, want exactly one, closed", entriesA)
	}
	if entriesA[0].EndedAt == nil || !entriesA[0].EndedAt.Equal(f.now.UTC()) {
		t.Errorf("A's entry ended at %v, want the injected clock's %s", entriesA[0].EndedAt, f.now)
	}
	entriesB := f.entriesOf(b.ID)
	if len(entriesB) != 1 || !entriesB[0].IsOpen() {
		t.Errorf("B's entries = %+v, want exactly one, open", entriesB)
	}
}

// D13 §2: a cascade opens no timer. Dragging a parent to Doing puts every
// unfinished leaf under it into the Doing column; the single active timer is
// global, so at most one of them could have it and there is no principled way to
// choose. A cascade is a planning gesture, not a "start working now" one.
func TestACascadeToDoingOpensNoTimer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	// A TASK with subtasks, not a project: a project is refused the Doing
	// column outright (D9), so it could never reach the cascade at all.
	p := f.create(draft("parent", domain.NodeTypeTask, nil))
	one := f.create(draft("one", domain.NodeTypeTask, &p.ID))
	two := f.create(draft("two", domain.NodeTypeTask, &p.ID))

	if _, err := f.tasks.MoveToColumn(ctx, p.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(parent, doing): %v", err)
	}

	// The parent itself gets no stored status — it DERIVES doing from the
	// leaves (D2) — so the plan named only the two children.
	if got := f.get(p.ID).Status; got == domain.StatusDoing {
		t.Fatalf("the parent was stored as doing; the cascade must leave a parent's status alone")
	}

	// The cascade really did happen — otherwise "no timer" would be a fact
	// about nothing.
	for _, id := range []string{one.ID, two.ID} {
		if got := f.get(id).Status; got != domain.StatusDoing {
			t.Fatalf("child %q is %q, want doing — the cascade did not run", id, got)
		}
	}
	if got := f.openEntries(); got != 0 {
		t.Errorf("%d open entries, want 0 — a cascade opens none", got)
	}
}

// D13 §3: leaving doing closes the entry, for every destination including done.
func TestMovingOutOfDoingClosesTheEntry(t *testing.T) {
	ctx := context.Background()

	for _, target := range []domain.Status{
		domain.StatusBacklog,
		domain.StatusWeek,
		domain.StatusToday,
		domain.StatusDone,
	} {
		t.Run(string(target), func(t *testing.T) {
			f := newFixture(t)
			n := f.create(draft("the card", domain.NodeTypeTask, nil))

			if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDoing); err != nil {
				t.Fatalf("MoveToColumn(doing): %v", err)
			}
			if got := f.openEntries(); got != 1 {
				t.Fatalf("%d open entries after the move to doing, want 1", got)
			}

			f.now = testNow.Add(25 * time.Minute)
			if _, err := f.tasks.MoveToColumn(ctx, n.ID, target); err != nil {
				t.Fatalf("MoveToColumn(%s): %v", target, err)
			}

			if got := f.openEntries(); got != 0 {
				t.Errorf("%d open entries after moving to %s, want 0", got, target)
			}
			entries := f.entriesOf(n.ID)
			if len(entries) != 1 {
				t.Fatalf("%d entries on the card, want 1", len(entries))
			}
			if entries[0].EndedAt == nil {
				t.Fatalf("the entry is still open after the card moved to %s", target)
			}
			if !entries[0].EndedAt.Equal(f.now.UTC()) {
				t.Errorf("ended_at = %s, want the injected clock's %s", entries[0].EndedAt, f.now)
			}
		})
	}
}

// D9, reached through the coupling: a project is refused the Doing column, so
// there is no move for a timer to be coupled to. It must not open one, and it
// must not close somebody else's.
func TestMovingAProjectToDoingTouchesNoTimer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	running := f.create(draft("a real task", domain.NodeTypeTask, nil))
	p := f.create(draft("a project", domain.NodeTypeProject, nil))

	if _, err := f.tasks.MoveToColumn(ctx, running.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(task, doing): %v", err)
	}

	_, err := f.tasks.MoveToColumn(ctx, p.ID, domain.StatusDoing)
	if !errors.Is(err, domain.ErrProjectNeverDoing) {
		t.Fatalf("MoveToColumn(project, doing) = %v, want ErrProjectNeverDoing", err)
	}

	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want the task's 1 — the refused move must neither open nor close", got)
	}
	if got := len(f.entriesOf(p.ID)); got != 0 {
		t.Errorf("%d entries on the project, want 0", got)
	}
	entries := f.entriesOf(running.ID)
	if len(entries) != 1 || !entries[0].IsOpen() {
		t.Errorf("the task's entries = %+v, want exactly one, still open", entries)
	}
}

// A type with no Kanban column is refused the move as before (PLAN.md §4), and
// the refusal reaches the timer no more than the move reaches the board.
func TestMovingANoColumnTypeToDoingTouchesNoTimer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	d := draft("stretch every morning", domain.NodeTypeHabit, nil)
	d.Recurrence = ptr("FREQ=DAILY")
	h := f.create(d)
	memo := f.create(draft("a memo", domain.NodeTypeNote, nil))

	for _, n := range []domain.Node{h, memo} {
		t.Run(string(n.Type), func(t *testing.T) {
			_, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDoing)
			if !errors.Is(err, domain.ErrTypeHasNoColumn) {
				t.Fatalf("MoveToColumn(%s, doing) = %v, want ErrTypeHasNoColumn", n.Type, err)
			}
			if got := f.openEntries(); got != 0 {
				t.Errorf("%d open entries, want 0", got)
			}
			if got := len(f.entriesOf(n.ID)); got != 0 {
				t.Errorf("%d entries on the %s, want 0", got, n.Type)
			}
		})
	}
}

// A timer opened by the timer service directly is closed by the move as well:
// the coupling is about the node the entry is running on, not about how the
// entry came to be open.
func TestMovingOutOfDoingClosesAnEntryTheTimerServiceOpened(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("the card", domain.NodeTypeTask, nil))

	if _, err := f.timers().Start(ctx, n.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	f.now = testNow.Add(5 * time.Minute)
	if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(done): %v", err)
	}

	if got := f.openEntries(); got != 0 {
		t.Errorf("%d open entries, want 0", got)
	}
}

// Re-dropping a card that is already in Doing onto Doing is not a restart: the
// entry that is running keeps its original started_at, exactly as
// TimerService.Start already promised, because it IS TimerService.start.
func TestMovingToDoingTwiceKeepsTheSameEntry(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("the card", domain.NodeTypeTask, nil))

	if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(doing): %v", err)
	}
	f.now = testNow.Add(20 * time.Minute)
	if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDoing); err != nil {
		t.Fatalf("MoveToColumn(doing) again: %v", err)
	}

	entries := f.entriesOf(n.ID)
	if len(entries) != 1 {
		t.Fatalf("%d entries, want 1 — the second drop must not fragment the log", len(entries))
	}
	if !entries[0].StartedAt.Equal(testNow) {
		t.Errorf("started_at = %s, want the original %s", entries[0].StartedAt, testNow)
	}
	if got := f.openEntries(); got != 1 {
		t.Errorf("%d open entries, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// D14 / K2 — archiving re-inspects a node that becomes a leaf (S2-06).

// THE K2 scenario, end to end through the service and a real database.
//
// P is stored done — a state an earlier drag of P to Done really does produce —
// while deriving backlog from {C1:done, C2:backlog}, so the board shows it in
// Backlog. Archive both children and P becomes a leaf, at which point D11 says
// its STORED status decides whether it counts as done in its parent's bar. Left
// alone it would silently become a done unit on a status nobody set, minutes
// after the board showed it as Backlog — self-consistent, and therefore never
// flagged.
func TestArchivingTheLastChildRewritesTheStaleStoredStatus(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	root := f.create(draft("root", domain.NodeTypeProject, nil))
	p := f.create(draft("p", domain.NodeTypeProject, &root.ID))
	c1 := f.create(draft("c1", domain.NodeTypeTask, &p.ID))
	c2 := f.create(draft("c2", domain.NodeTypeTask, &p.ID))

	if _, err := f.tasks.MoveToColumn(ctx, c1.ID, domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(c1, done): %v", err)
	}
	// The stale done: an earlier drag of P itself to Done, back when it was a
	// leaf. It is written the way the bug writes it — straight onto the row.
	stale := f.get(p.ID)
	stale.Status = domain.StatusDone
	completed := testNow.AddDate(0, -1, 0)
	stale.CompletedAt = &completed
	if err := f.nodes.Update(ctx, stale); err != nil {
		t.Fatalf("Update: %v", err)
	}

	// What the board shows before the archive, which is what must survive it.
	v := find(boardOf(ctx, t, f), p.ID)
	if v == nil || v.Status != domain.StatusBacklog {
		t.Fatalf("before the archive p renders %+v, want backlog", v)
	}

	if _, err := f.tasks.ArchiveNode(ctx, c1.ID); err != nil {
		t.Fatalf("ArchiveNode(c1): %v", err)
	}
	t.Run("archiving one of two children rewrites nothing", func(t *testing.T) {
		if got := f.get(p.ID).Status; got != domain.StatusDone {
			t.Errorf("p's stored status is %q, want the stale done still there — c2 has a column", got)
		}
	})

	if _, err := f.tasks.ArchiveNode(ctx, c2.ID); err != nil {
		t.Fatalf("ArchiveNode(c2): %v", err)
	}

	t.Run("p's stored status is now the backlog it was displaying", func(t *testing.T) {
		got := f.get(p.ID)
		if got.Status != domain.StatusBacklog {
			t.Errorf("p's stored status = %q, want backlog (D14)", got.Status)
		}
		if got.CompletedAt != nil {
			t.Errorf("p's completed_at = %v, want NULL: it is not done", got.CompletedAt)
		}
	})

	t.Run("the board shows the same thing after the archive as before", func(t *testing.T) {
		v := find(boardOf(ctx, t, f), p.ID)
		if v == nil {
			t.Fatal("p left the board")
		}
		if v.Status != domain.StatusBacklog {
			t.Errorf("p renders in %q, want the backlog it rendered in before the archive", v.Status)
		}
	})

	t.Run("p counts as one unfinished work leaf in root's progress (D11)", func(t *testing.T) {
		got, err := f.tasks.Progress(ctx, root.ID)
		if err != nil {
			t.Fatalf("Progress: %v", err)
		}
		want := service.ProgressView{Done: 0, Total: 1, Percent: 0, Defined: true}
		if got != want {
			t.Errorf("root's progress = %+v, want %+v — p is one UNFINISHED unit", got, want)
		}
	})
}

// The honest-completion case, which rejected alternative (a) — "reset to
// backlog" — would have broken: archiving the finished children of a finished
// project must not un-finish it.
func TestArchivingFinishedChildrenKeepsTheProjectDone(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	root := f.create(draft("root", domain.NodeTypeProject, nil))
	p := f.create(draft("p", domain.NodeTypeProject, &root.ID))
	c1 := f.create(draft("c1", domain.NodeTypeTask, &p.ID))
	c2 := f.create(draft("c2", domain.NodeTypeTask, &p.ID))

	for _, id := range []string{c1.ID, c2.ID} {
		if _, err := f.tasks.MoveToColumn(ctx, id, domain.StatusDone); err != nil {
			t.Fatalf("MoveToColumn(%q, done): %v", id, err)
		}
	}
	// P genuinely finished, with the completion time that drag recorded.
	finished := f.get(p.ID)
	finished.Status = domain.StatusDone
	completedAt := testNow
	finished.CompletedAt = &completedAt
	if err := f.nodes.Update(ctx, finished); err != nil {
		t.Fatalf("Update: %v", err)
	}

	f.now = testNow.AddDate(0, 0, 7)
	for _, id := range []string{c1.ID, c2.ID} {
		if _, err := f.tasks.ArchiveNode(ctx, id); err != nil {
			t.Fatalf("ArchiveNode(%q): %v", id, err)
		}
	}

	got := f.get(p.ID)
	if got.Status != domain.StatusDone {
		t.Errorf("p's stored status = %q, want done — its children really were finished", got.Status)
	}
	if got.CompletedAt == nil || !got.CompletedAt.Equal(completedAt) {
		t.Errorf("p's completed_at = %v, want the original %v, not the day of the archive",
			got.CompletedAt, completedAt)
	}

	progress, err := f.tasks.Progress(ctx, root.ID)
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if want := (service.ProgressView{Done: 1, Total: 1, Percent: 100, Defined: true}); progress != want {
		t.Errorf("root's progress = %+v, want %+v — p is one DONE unit", progress, want)
	}
}

// The rewrite happens in the archive's own transaction: a failed archive leaves
// the stored status exactly as it was.
func TestTheStatusRewriteRollsBackWithAFailedArchive(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("p", domain.NodeTypeProject, nil))
	c := f.create(draft("c", domain.NodeTypeTask, &p.ID))
	if _, err := f.tasks.MoveToColumn(ctx, c.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn: %v", err)
	}
	stale := f.get(p.ID)
	stale.Status = domain.StatusDone
	if err := f.nodes.Update(ctx, stale); err != nil {
		t.Fatalf("Update: %v", err)
	}
	before := f.all()

	// The archive's own UPDATE succeeds and the status rewrite that follows it
	// fails, which is the ordering that could leave half a plan applied.
	f.now = testNow.Add(time.Hour)
	brittle := f.tasksOver(failingBeginner{inner: f.begin, after: 1})

	if _, err := brittle.ArchiveNode(ctx, c.ID); !errors.Is(err, errBoom) {
		t.Fatalf("ArchiveNode = %v, want the injected errBoom", err)
	}
	if after := f.all(); !reflect.DeepEqual(before, after) {
		t.Errorf("a failed archive changed the database:\n got %+v\nwant %+v", after, before)
	}
}

// Restore adds no symmetric rule: bringing a child back hands the parent to
// derivation again, and nothing writes a stored status on the way.
func TestRestoreWritesNoStatus(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("p", domain.NodeTypeProject, nil))
	c := f.create(draft("c", domain.NodeTypeTask, &p.ID))
	if _, err := f.tasks.MoveToColumn(ctx, c.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn: %v", err)
	}

	if _, err := f.tasks.ArchiveNode(ctx, c.ID); err != nil {
		t.Fatalf("ArchiveNode: %v", err)
	}
	// The archive left p showing what it showed: today.
	if got := f.get(p.ID).Status; got != domain.StatusToday {
		t.Fatalf("p's stored status after the archive = %q, want today (D14)", got)
	}

	f.now = testNow.Add(time.Hour)
	stored := f.get(p.ID)
	if _, err := f.tasks.RestoreNode(ctx, c.ID); err != nil {
		t.Fatalf("RestoreNode: %v", err)
	}

	after := f.get(p.ID)
	if after.Status != stored.Status || !after.UpdatedAt.Equal(stored.UpdatedAt) {
		t.Errorf("restoring rewrote p: %+v, was %+v — restore adds no rule (D14)", after, stored)
	}

	// Derivation has taken over again, which is why no rule is needed.
	v := find(boardOf(ctx, t, f), p.ID)
	if v == nil || v.Status != domain.StatusToday {
		t.Errorf("p renders %+v, want today derived from the restored child", v)
	}
	if v.IsLeaf {
		t.Error("p reports IsLeaf = true although its child is back")
	}
}
