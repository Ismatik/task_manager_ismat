package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"nexus/internal/domain"

	sqlite3 "modernc.org/sqlite/lib"
)

// testNow is the fixed instant every node test is written against. Nothing in
// the store reads a clock, so every timestamp in these tests is one the test
// chose and can therefore assert on exactly.
var testNow = time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC)

// newTestNode returns a node that is valid in every column, so that a test only
// has to say what it is actually about.
func newTestNode(id string) domain.Node {
	return domain.Node{
		ID:            id,
		Type:          domain.NodeTypeTask,
		Title:         "node " + id,
		DescriptionMD: "",
		Status:        domain.StatusBacklog,
		DueSource:     domain.DueSourceManual,
		Priority:      domain.Priority4,
		CreatedAt:     testNow,
		UpdatedAt:     testNow,
	}
}

func strPtr(s string) *string            { return &s }
func intPtr(i int) *int                  { return &i }
func timePtr(t time.Time) *time.Time     { return &t }
func datePtr(d domain.Date) *domain.Date { return &d }
func activityPtr(a domain.Activity) *domain.Activity {
	return &a
}

// nodeRepo returns a repository over a freshly migrated temp database.
func nodeRepo(t *testing.T) (*NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	return NewNodeRepo(db), db
}

// mustCreate inserts n through the repository and fails the test if it cannot.
func mustCreate(t *testing.T, ctx context.Context, r *NodeRepo, n domain.Node) {
	t.Helper()

	if err := r.Create(ctx, n); err != nil {
		t.Fatalf("Create(%q): %v", n.ID, err)
	}
}

// nodeFields is every field of domain.Node, named, so that a round-trip test
// can report exactly which column did not survive rather than dumping two
// structs and leaving the reader to spot the difference. A field added to
// domain.Node and forgotten here is caught by TestNodeFieldsAreAllCompared.
func nodeFields() []struct {
	name string
	get  func(domain.Node) any
} {
	return []struct {
		name string
		get  func(domain.Node) any
	}{
		{"ID", func(n domain.Node) any { return n.ID }},
		{"ParentID", func(n domain.Node) any { return derefString(n.ParentID) }},
		{"Type", func(n domain.Node) any { return n.Type }},
		{"Title", func(n domain.Node) any { return n.Title }},
		{"DescriptionMD", func(n domain.Node) any { return n.DescriptionMD }},
		{"Status", func(n domain.Node) any { return n.Status }},
		{"Due", func(n domain.Node) any { return derefDate(n.Due) }},
		{"DueSource", func(n domain.Node) any { return n.DueSource }},
		{"Priority", func(n domain.Node) any { return n.Priority }},
		{"EstimateMin", func(n domain.Node) any { return derefInt(n.EstimateMin) }},
		{"Recurrence", func(n domain.Node) any { return derefString(n.Recurrence) }},
		{"Activity", func(n domain.Node) any { return derefActivity(n.Activity) }},
		{"SortOrder", func(n domain.Node) any { return n.SortOrder }},
		{"CreatedAt", func(n domain.Node) any { return stamp(&n.CreatedAt) }},
		{"UpdatedAt", func(n domain.Node) any { return stamp(&n.UpdatedAt) }},
		{"CompletedAt", func(n domain.Node) any { return stamp(n.CompletedAt) }},
		{"ArchivedAt", func(n domain.Node) any { return stamp(n.ArchivedAt) }},
	}
}

func derefString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

func derefInt(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func derefDate(p *domain.Date) any {
	if p == nil {
		return nil
	}
	return p.String()
}

func derefActivity(p *domain.Activity) any {
	if p == nil {
		return nil
	}
	return string(*p)
}

// stamp renders a time for comparison. Comparing the rendered instant rather
// than the time.Time is what makes "UTC normalisation" an assertion instead of
// an assumption: a value written in +03:00 and read back in UTC is the same
// instant and must compare equal.
func stamp(p *time.Time) any {
	if p == nil {
		return nil
	}
	return p.UTC().Format(time.RFC3339Nano)
}

// assertNodeEqual compares every field of domain.Node by name.
func assertNodeEqual(t *testing.T, got, want domain.Node) {
	t.Helper()

	for _, f := range nodeFields() {
		if g, w := f.get(got), f.get(want); !reflect.DeepEqual(g, w) {
			t.Errorf("%s = %v (%T), want %v (%T)", f.name, g, g, w, w)
		}
	}
}

// A field on domain.Node that nodeFields does not mention would be a field the
// round-trip test silently skips — which is precisely the "written but never
// read back" bug this stage exists to prevent.
func TestNodeFieldsAreAllCompared(t *testing.T) {
	compared := map[string]bool{}
	for _, f := range nodeFields() {
		compared[f.name] = true
	}

	typ := reflect.TypeOf(domain.Node{})
	for i := range typ.NumField() {
		if name := typ.Field(i).Name; !compared[name] {
			t.Errorf("domain.Node.%s is not compared by the round-trip test; add it to nodeFields", name)
		}
	}
	if len(compared) != typ.NumField() {
		t.Errorf("nodeFields compares %d fields, domain.Node has %d", len(compared), typ.NumField())
	}
}

// The column list the repository builds its statements from has to be the
// schema's, in the schema's order, or nodeValues and scanNode are reading and
// writing different things.
func TestNodeColumnsMatchTheSchema(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	rows, err := db.QueryContext(ctx, "SELECT name FROM pragma_table_info('nodes') ORDER BY cid")
	if err != nil {
		t.Fatalf("reading pragma_table_info: %v", err)
	}
	defer rows.Close()

	var schema []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scanning pragma_table_info: %v", err)
		}
		schema = append(schema, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading pragma_table_info: %v", err)
	}

	if !slices.Equal(schema, nodeColumnNames) {
		t.Errorf("nodeColumnNames = %v,\nschema columns   = %v", nodeColumnNames, schema)
	}
}

// Create then Get, with every nullable column NULL in one case and set in the
// other. A column that is written and never read back is the bug this test
// exists for.
func TestNodeRepoRoundTripsEveryColumn(t *testing.T) {
	ctx := context.Background()

	due := domain.NewDate(2026, time.September, 18)
	completed := time.Date(2026, 9, 18, 17, 30, 0, 0, time.UTC)
	archived := time.Date(2026, 9, 19, 8, 15, 0, 0, time.UTC)

	for _, tc := range []struct {
		name string
		node domain.Node
	}{
		{
			name: "every nullable column NULL",
			node: newTestNode("plain"),
		},
		{
			name: "every nullable column set",
			node: domain.Node{
				ID:            "full",
				ParentID:      strPtr("root"),
				Type:          domain.NodeTypeHabit,
				Title:         "полный узел",
				DescriptionMD: "# heading\n\nbody with *markdown*",
				Status:        domain.StatusDoing,
				Due:           datePtr(due),
				DueSource:     domain.DueSourceAuto,
				Priority:      domain.Priority1,
				EstimateMin:   intPtr(90),
				Recurrence:    strPtr("FREQ=WEEKLY;BYDAY=MO,WE,FR"),
				Activity:      activityPtr(domain.ActivityAnalysis),
				SortOrder:     7,
				CreatedAt:     testNow,
				UpdatedAt:     testNow.Add(time.Hour),
				CompletedAt:   timePtr(completed),
				ArchivedAt:    timePtr(archived),
			},
		},
		{
			// due_source and activity are the two columns D1 and D4 added and
			// the two most likely to be dropped on the floor by a mapper that
			// was written against PLAN.md's first sketch of the table.
			name: "due_source auto and a D4 activity survive on their own",
			node: func() domain.Node {
				n := newTestNode("d1d4")
				n.Due = datePtr(due)
				n.DueSource = domain.DueSourceAuto
				n.Activity = activityPtr(domain.ActivityManagement)
				return n
			}(),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _ := nodeRepo(t)

			if tc.node.ParentID != nil {
				mustCreate(t, ctx, r, newTestNode(*tc.node.ParentID))
			}
			mustCreate(t, ctx, r, tc.node)

			got, err := r.Get(ctx, tc.node.ID)
			if err != nil {
				t.Fatalf("Get(%q): %v", tc.node.ID, err)
			}
			assertNodeEqual(t, got, tc.node)
		})
	}
}

// A timestamp handed in outside UTC comes back as the same instant in UTC, and
// the column really does hold the 0002 format rather than something the driver
// invented.
func TestNodeRepoNormalisesTimestampsToUTC(t *testing.T) {
	ctx := context.Background()
	r, db := nodeRepo(t)

	east := time.FixedZone("UTC+3", 3*60*60)
	created := time.Date(2026, 9, 18, 12, 0, 0, 0, east) // 09:00:00Z

	n := newTestNode("zoned")
	n.CreatedAt = created
	n.UpdatedAt = created
	mustCreate(t, ctx, r, n)

	var raw string
	if err := db.QueryRowContext(ctx, "SELECT created_at FROM nodes WHERE id = 'zoned'").Scan(&raw); err != nil {
		t.Fatalf("reading created_at: %v", err)
	}
	if want := "2026-09-18T09:00:00Z"; raw != want {
		t.Errorf("stored created_at = %q, want %q", raw, want)
	}

	got, err := r.Get(ctx, "zoned")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("CreatedAt = %s, want the same instant as %s", got.CreatedAt, created)
	}
	if got.CreatedAt.Location() != time.UTC {
		t.Errorf("CreatedAt location = %s, want UTC", got.CreatedAt.Location())
	}
}

func TestNodeRepoGetUnknownIDIsErrNotFound(t *testing.T) {
	ctx := context.Background()
	r, _ := nodeRepo(t)

	_, err := r.Get(ctx, "nobody")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(unknown) error = %v, want one matching ErrNotFound", err)
	}
}

func TestNodeRepoUpdateRewritesEveryColumn(t *testing.T) {
	ctx := context.Background()

	t.Run("an existing row takes every new value", func(t *testing.T) {
		r, _ := nodeRepo(t)
		mustCreate(t, ctx, r, newTestNode("n"))

		updated := domain.Node{
			ID:            "n",
			Type:          domain.NodeTypeBug,
			Title:         "renamed",
			DescriptionMD: "severity: high",
			Status:        domain.StatusDone,
			Due:           datePtr(domain.NewDate(2026, time.October, 1)),
			DueSource:     domain.DueSourceAuto,
			Priority:      domain.Priority2,
			EstimateMin:   intPtr(15),
			Activity:      activityPtr(domain.ActivityTesting),
			SortOrder:     3,
			CreatedAt:     testNow,
			UpdatedAt:     testNow.Add(2 * time.Hour),
			CompletedAt:   timePtr(testNow.Add(time.Hour)),
		}
		if err := r.Update(ctx, updated); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := r.Get(ctx, "n")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		assertNodeEqual(t, got, updated)
	})

	t.Run("a nullable column can be cleared back to NULL", func(t *testing.T) {
		r, _ := nodeRepo(t)

		n := newTestNode("n")
		n.Due = datePtr(domain.NewDate(2026, time.October, 1))
		n.EstimateMin = intPtr(30)
		n.CompletedAt = timePtr(testNow)
		mustCreate(t, ctx, r, n)

		cleared := newTestNode("n")
		if err := r.Update(ctx, cleared); err != nil {
			t.Fatalf("Update: %v", err)
		}

		got, err := r.Get(ctx, "n")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		assertNodeEqual(t, got, cleared)
	})

	t.Run("a row that is not there is ErrNotFound, not a silent no-op", func(t *testing.T) {
		r, _ := nodeRepo(t)

		if err := r.Update(ctx, newTestNode("ghost")); !errors.Is(err, ErrNotFound) {
			t.Errorf("Update(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})
}

func TestNodeRepoDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("deleting a node takes its subtree with it", func(t *testing.T) {
		r, db := nodeRepo(t)
		mustCreate(t, ctx, r, newTestNode("root"))
		child := newTestNode("child")
		child.ParentID = strPtr("root")
		mustCreate(t, ctx, r, child)
		grandchild := newTestNode("grandchild")
		grandchild.ParentID = strPtr("child")
		mustCreate(t, ctx, r, grandchild)

		if err := r.Delete(ctx, "root"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if n := countRows(t, db, "nodes"); n != 0 {
			t.Errorf("after deleting the root, nodes has %d rows, want 0 (ON DELETE CASCADE)", n)
		}
	})

	t.Run("a row that is not there is ErrNotFound", func(t *testing.T) {
		r, _ := nodeRepo(t)

		if err := r.Delete(ctx, "ghost"); !errors.Is(err, ErrNotFound) {
			t.Errorf("Delete(unknown) error = %v, want one matching ErrNotFound", err)
		}
	})
}

// seedTree builds
//
//	root ── a ── a1
//	     └─ b
//	other ── o1
//
// so that a subtree query has both depth and a sibling subtree to leave alone.
func seedTree(t *testing.T, ctx context.Context, r *NodeRepo) {
	t.Helper()

	for _, n := range []struct {
		id     string
		parent *string
		order  int
	}{
		{"root", nil, 0},
		{"a", strPtr("root"), 0},
		{"b", strPtr("root"), 1},
		{"a1", strPtr("a"), 0},
		{"other", nil, 1},
		{"o1", strPtr("other"), 0},
	} {
		node := newTestNode(n.id)
		node.ParentID = n.parent
		node.SortOrder = n.order
		mustCreate(t, ctx, r, node)
	}
}

func ids(nodes []domain.Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.ID
	}
	return out
}

func TestNodeRepoListSubtree(t *testing.T) {
	ctx := context.Background()

	t.Run("a three-level tree returns the root and every descendant", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListSubtree(ctx, "root", false)
		if err != nil {
			t.Fatalf("ListSubtree: %v", err)
		}
		// sort_order first, id as the tie-break: a, a1 and root all sit at 0.
		want := []string{"a", "a1", "root", "b"}
		if !slices.Equal(ids(got), want) {
			t.Errorf("ListSubtree(root) = %v, want %v", ids(got), want)
		}
	})

	t.Run("nothing from a sibling subtree leaks in", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListSubtree(ctx, "root", false)
		if err != nil {
			t.Fatalf("ListSubtree: %v", err)
		}
		for _, n := range got {
			if n.ID == "other" || n.ID == "o1" {
				t.Errorf("ListSubtree(root) returned %q from the sibling subtree", n.ID)
			}
		}
	})

	t.Run("an inner node returns only what is below it", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListSubtree(ctx, "a", false)
		if err != nil {
			t.Fatalf("ListSubtree: %v", err)
		}
		if want := []string{"a", "a1"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListSubtree(a) = %v, want %v", ids(got), want)
		}
	})

	t.Run("a leaf returns just itself", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListSubtree(ctx, "a1", false)
		if err != nil {
			t.Fatalf("ListSubtree: %v", err)
		}
		if want := []string{"a1"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListSubtree(a1) = %v, want %v", ids(got), want)
		}
	})

	t.Run("an id that does not exist is an empty subtree, not an error", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListSubtree(ctx, "ghost", false)
		if err != nil {
			t.Fatalf("ListSubtree(ghost): %v", err)
		}
		if len(got) != 0 {
			t.Errorf("ListSubtree(ghost) = %v, want empty", ids(got))
		}
	})
}

func TestNodeRepoListChildrenAndRoots(t *testing.T) {
	ctx := context.Background()

	t.Run("direct children only, in sort order", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListChildren(ctx, strPtr("root"), false)
		if err != nil {
			t.Fatalf("ListChildren: %v", err)
		}
		if want := []string{"a", "b"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListChildren(root) = %v, want %v (a1 is a grandchild)", ids(got), want)
		}
	})

	t.Run("a nil parent means the roots", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListChildren(ctx, nil, false)
		if err != nil {
			t.Fatalf("ListChildren(nil): %v", err)
		}
		if want := []string{"root", "other"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListChildren(nil) = %v, want %v", ids(got), want)
		}
	})

	t.Run("ListRoots agrees with ListChildren(nil)", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		roots, err := r.ListRoots(ctx, false)
		if err != nil {
			t.Fatalf("ListRoots: %v", err)
		}
		children, err := r.ListChildren(ctx, nil, false)
		if err != nil {
			t.Fatalf("ListChildren(nil): %v", err)
		}
		if !slices.Equal(ids(roots), ids(children)) {
			t.Errorf("ListRoots = %v, ListChildren(nil) = %v", ids(roots), ids(children))
		}
	})

	t.Run("a childless parent lists nothing rather than nil", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		got, err := r.ListChildren(ctx, strPtr("a1"), false)
		if err != nil {
			t.Fatalf("ListChildren: %v", err)
		}
		if got == nil || len(got) != 0 {
			t.Errorf("ListChildren(a1) = %v, want an empty non-nil slice", got)
		}
	})
}

// Archiving is how Nexus hides a node without destroying it, so every list has
// to leave archived rows out unless it is asked for them — one subtest each,
// because a filter that is right in three of four queries is still a bug.
func TestNodeRepoArchivedExclusion(t *testing.T) {
	ctx := context.Background()

	archivedAt := testNow.Add(24 * time.Hour)

	setup := func(t *testing.T) *NodeRepo {
		t.Helper()
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)
		if err := r.SetArchivedAt(ctx, []string{"b"}, &archivedAt, testNow); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}
		return r
	}

	for _, tc := range []struct {
		name            string
		list            func(*NodeRepo, bool) ([]domain.Node, error)
		wantExcluded    []string
		wantIncluded    []string
		archivedMemberB bool
	}{
		{
			name:         "ListChildren",
			list:         func(r *NodeRepo, inc bool) ([]domain.Node, error) { return r.ListChildren(ctx, strPtr("root"), inc) },
			wantExcluded: []string{"a"},
			wantIncluded: []string{"a", "b"},
		},
		{
			name:         "ListSubtree",
			list:         func(r *NodeRepo, inc bool) ([]domain.Node, error) { return r.ListSubtree(ctx, "root", inc) },
			wantExcluded: []string{"a", "a1", "root"},
			wantIncluded: []string{"a", "a1", "b", "root"},
		},
		{
			name:         "ListAll",
			list:         func(r *NodeRepo, inc bool) ([]domain.Node, error) { return r.ListAll(ctx, inc) },
			wantExcluded: []string{"a", "a1", "o1", "other", "root"},
			wantIncluded: []string{"a", "a1", "b", "o1", "other", "root"},
		},
	} {
		t.Run(tc.name+": archived rows are excluded by default", func(t *testing.T) {
			r := setup(t)

			got, err := tc.list(r, false)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			sorted := ids(got)
			slices.Sort(sorted)
			if !slices.Equal(sorted, tc.wantExcluded) {
				t.Errorf("%s(includeArchived=false) = %v, want %v", tc.name, sorted, tc.wantExcluded)
			}
		})

		t.Run(tc.name+": archived rows are included on request", func(t *testing.T) {
			r := setup(t)

			got, err := tc.list(r, true)
			if err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			sorted := ids(got)
			slices.Sort(sorted)
			if !slices.Equal(sorted, tc.wantIncluded) {
				t.Errorf("%s(includeArchived=true) = %v, want %v", tc.name, sorted, tc.wantIncluded)
			}
		})
	}

	t.Run("ListRoots: an archived root is excluded by default and included on request", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)
		if err := r.SetArchivedAt(ctx, []string{"other"}, &archivedAt, testNow); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}

		got, err := r.ListRoots(ctx, false)
		if err != nil {
			t.Fatalf("ListRoots: %v", err)
		}
		if want := []string{"root"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListRoots(false) = %v, want %v", ids(got), want)
		}

		got, err = r.ListRoots(ctx, true)
		if err != nil {
			t.Fatalf("ListRoots: %v", err)
		}
		if want := []string{"root", "other"}; !slices.Equal(ids(got), want) {
			t.Errorf("ListRoots(true) = %v, want %v", ids(got), want)
		}
	})

	// Hiding an archived parent must not hide the live children underneath it:
	// the archived filter belongs on the result, not on the walk.
	t.Run("an archived node does not hide its unarchived descendants", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)
		if err := r.SetArchivedAt(ctx, []string{"a"}, &archivedAt, testNow); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}

		got, err := r.ListSubtree(ctx, "root", false)
		if err != nil {
			t.Fatalf("ListSubtree: %v", err)
		}
		if !slices.Contains(ids(got), "a1") {
			t.Errorf("ListSubtree(root) = %v, want it to still contain a1, whose parent a is archived", ids(got))
		}
	})
}

func TestNodeRepoSetArchivedAt(t *testing.T) {
	ctx := context.Background()
	archivedAt := testNow.Add(24 * time.Hour)

	t.Run("stamps every id in the batch", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		if err := r.SetArchivedAt(ctx, []string{"a", "a1"}, &archivedAt, testNow.Add(time.Hour)); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}
		for _, id := range []string{"a", "a1"} {
			n, err := r.Get(ctx, id)
			if err != nil {
				t.Fatalf("Get(%q): %v", id, err)
			}
			if n.ArchivedAt == nil || !n.ArchivedAt.Equal(archivedAt) {
				t.Errorf("%s.ArchivedAt = %v, want %s", id, n.ArchivedAt, archivedAt)
			}
			if !n.UpdatedAt.Equal(testNow.Add(time.Hour)) {
				t.Errorf("%s.UpdatedAt = %s, want it bumped to %s", id, n.UpdatedAt, testNow.Add(time.Hour))
			}
		}
	})

	t.Run("a nil time un-archives", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		if err := r.SetArchivedAt(ctx, []string{"a"}, &archivedAt, testNow); err != nil {
			t.Fatalf("SetArchivedAt: %v", err)
		}
		if err := r.SetArchivedAt(ctx, []string{"a"}, nil, testNow); err != nil {
			t.Fatalf("SetArchivedAt(nil): %v", err)
		}

		n, err := r.Get(ctx, "a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.ArchivedAt != nil {
			t.Errorf("ArchivedAt = %v, want nil", n.ArchivedAt)
		}
	})

	t.Run("an unknown id is ErrNotFound and nothing is archived", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		err := r.SetArchivedAt(ctx, []string{"a", "ghost"}, &archivedAt, testNow)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("SetArchivedAt error = %v, want one matching ErrNotFound", err)
		}
		n, err := r.Get(ctx, "a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.ArchivedAt != nil {
			t.Errorf("a.ArchivedAt = %v after a failed batch, want nil — the statement must be all or nothing", n.ArchivedAt)
		}
	})

	t.Run("an empty id list is a no-op", func(t *testing.T) {
		r, _ := nodeRepo(t)

		if err := r.SetArchivedAt(ctx, nil, &archivedAt, testNow); err != nil {
			t.Errorf("SetArchivedAt(nil ids) = %v, want nil", err)
		}
	})
}

func TestNodeRepoUpdateParentAndOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("re-parents and repositions in one write", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		at := testNow.Add(3 * time.Hour)
		if err := r.UpdateParentAndOrder(ctx, "a1", strPtr("b"), 5, at); err != nil {
			t.Fatalf("UpdateParentAndOrder: %v", err)
		}

		n, err := r.Get(ctx, "a1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.ParentID == nil || *n.ParentID != "b" {
			t.Errorf("ParentID = %v, want b", derefString(n.ParentID))
		}
		if n.SortOrder != 5 {
			t.Errorf("SortOrder = %d, want 5", n.SortOrder)
		}
		if !n.UpdatedAt.Equal(at) {
			t.Errorf("UpdatedAt = %s, want %s", n.UpdatedAt, at)
		}
	})

	t.Run("a nil parent makes the node a root", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		if err := r.UpdateParentAndOrder(ctx, "a1", nil, 0, testNow); err != nil {
			t.Fatalf("UpdateParentAndOrder: %v", err)
		}
		n, err := r.Get(ctx, "a1")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.ParentID != nil {
			t.Errorf("ParentID = %v, want nil", *n.ParentID)
		}
	})

	t.Run("an unknown node is ErrNotFound", func(t *testing.T) {
		r, _ := nodeRepo(t)

		if err := r.UpdateParentAndOrder(ctx, "ghost", nil, 0, testNow); !errors.Is(err, ErrNotFound) {
			t.Errorf("error = %v, want one matching ErrNotFound", err)
		}
	})

	t.Run("a parent that does not exist is refused by the foreign key", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		err := r.UpdateParentAndOrder(ctx, "a1", strPtr("ghost"), 0, testNow)
		if !errors.Is(err, ErrConstraint) {
			t.Errorf("error = %v, want one matching ErrConstraint", err)
		}
	})
}

// UpdateStatuses is how a cascade plan reaches the database. The plan is
// domain.PlanCascade's; this checks only that the plan is written whole.
func TestNodeRepoUpdateStatuses(t *testing.T) {
	ctx := context.Background()
	at := testNow.Add(4 * time.Hour)

	t.Run("a multi-row change is applied in one call", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		completed := testNow.Add(90 * time.Minute)
		changes := []domain.StatusChange{
			{NodeID: "a", Status: domain.StatusToday},
			{NodeID: "a1", Status: domain.StatusDone, CompletedAt: &completed},
			{NodeID: "b", Status: domain.StatusDoing},
		}
		if err := r.UpdateStatuses(ctx, changes, at); err != nil {
			t.Fatalf("UpdateStatuses: %v", err)
		}

		for _, want := range changes {
			n, err := r.Get(ctx, want.NodeID)
			if err != nil {
				t.Fatalf("Get(%q): %v", want.NodeID, err)
			}
			if n.Status != want.Status {
				t.Errorf("%s.Status = %q, want %q", want.NodeID, n.Status, want.Status)
			}
			if !reflect.DeepEqual(stamp(n.CompletedAt), stamp(want.CompletedAt)) {
				t.Errorf("%s.CompletedAt = %v, want %v", want.NodeID, stamp(n.CompletedAt), stamp(want.CompletedAt))
			}
			if !n.UpdatedAt.Equal(at) {
				t.Errorf("%s.UpdatedAt = %s, want %s", want.NodeID, n.UpdatedAt, at)
			}
		}
	})

	// A StatusChange always carries completed_at, nil included: a node leaving
	// done must lose the completion time in the same write.
	t.Run("a nil CompletedAt clears the column rather than leaving it", func(t *testing.T) {
		r, _ := nodeRepo(t)

		n := newTestNode("done")
		n.Status = domain.StatusDone
		n.CompletedAt = timePtr(testNow)
		mustCreate(t, ctx, r, n)

		changes := []domain.StatusChange{{NodeID: "done", Status: domain.StatusToday}}
		if err := r.UpdateStatuses(ctx, changes, at); err != nil {
			t.Fatalf("UpdateStatuses: %v", err)
		}

		got, err := r.Get(ctx, "done")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.CompletedAt != nil {
			t.Errorf("CompletedAt = %v, want nil", got.CompletedAt)
		}
	})

	t.Run("a bad row in the batch leaves none of them applied", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		changes := []domain.StatusChange{
			{NodeID: "a", Status: domain.StatusToday},
			{NodeID: "b", Status: domain.Status("Done")}, // wrong case; the CHECK refuses it
			{NodeID: "a1", Status: domain.StatusDoing},
		}
		err := r.UpdateStatuses(ctx, changes, at)
		if !errors.Is(err, ErrConstraint) {
			t.Fatalf("UpdateStatuses error = %v, want one matching ErrConstraint", err)
		}

		for _, id := range []string{"a", "a1", "b"} {
			n, getErr := r.Get(ctx, id)
			if getErr != nil {
				t.Fatalf("Get(%q): %v", id, getErr)
			}
			if n.Status != domain.StatusBacklog {
				t.Errorf("%s.Status = %q after a rejected batch, want the untouched %q",
					id, n.Status, domain.StatusBacklog)
			}
		}
	})

	t.Run("an id that matches no row is ErrNotFound and nothing is applied", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		changes := []domain.StatusChange{
			{NodeID: "a", Status: domain.StatusToday},
			{NodeID: "ghost", Status: domain.StatusToday},
		}
		if err := r.UpdateStatuses(ctx, changes, at); !errors.Is(err, ErrNotFound) {
			t.Fatalf("UpdateStatuses error = %v, want one matching ErrNotFound", err)
		}

		n, err := r.Get(ctx, "a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.Status != domain.StatusBacklog {
			t.Errorf("a.Status = %q, want the untouched %q", n.Status, domain.StatusBacklog)
		}
	})

	t.Run("an empty plan is a no-op", func(t *testing.T) {
		r, _ := nodeRepo(t)

		if err := r.UpdateStatuses(ctx, nil, at); err != nil {
			t.Errorf("UpdateStatuses(nil) = %v, want nil", err)
		}
	})

	t.Run("a repeated id is counted once", func(t *testing.T) {
		r, _ := nodeRepo(t)
		seedTree(t, ctx, r)

		changes := []domain.StatusChange{
			{NodeID: "a", Status: domain.StatusToday},
			{NodeID: "a", Status: domain.StatusDoing},
		}
		if err := r.UpdateStatuses(ctx, changes, at); err != nil {
			t.Fatalf("UpdateStatuses: %v", err)
		}
		n, err := r.Get(ctx, "a")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if n.Status != domain.StatusToday {
			t.Errorf("a.Status = %q, want the first change's %q", n.Status, domain.StatusToday)
		}
	})
}

// The schema's CHECK constraints are the backstop underneath the domain's
// validators (0002 says so). These subtests prove the repository is wired to
// them and reports them as ErrConstraint rather than as a driver string.
func TestNodeRepoSurfacesSchemaConstraints(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		node  func() domain.Node
		check func(domain.Node) domain.Node
	}{
		{
			name: "type: an unknown kind",
			node: func() domain.Node { n := newTestNode("x"); n.Type = "epic"; return n },
		},
		{
			name: "status: right value, wrong case",
			node: func() domain.Node { n := newTestNode("x"); n.Status = "Done"; return n },
		},
		{
			name: "due_source: neither manual nor auto",
			node: func() domain.Node { n := newTestNode("x"); n.DueSource = "inherited"; return n },
		},
		{
			name: "priority: 5 is above the range",
			node: func() domain.Node { n := newTestNode("x"); n.Priority = 5; return n },
		},
		{
			name: "activity: not one of the seven D4 values",
			node: func() domain.Node {
				n := newTestNode("x")
				n.Activity = activityPtr(domain.Activity("Отдых"))
				return n
			},
		},
		{
			name: "parent_id: a parent that does not exist",
			node: func() domain.Node { n := newTestNode("x"); n.ParentID = strPtr("ghost"); return n },
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, db := nodeRepo(t)

			err := r.Create(ctx, tc.node())
			if !errors.Is(err, ErrConstraint) {
				t.Fatalf("Create error = %v, want one matching ErrConstraint", err)
			}
			if n := countRows(t, db, "nodes"); n != 0 {
				t.Errorf("nodes has %d rows after a refused insert, want 0", n)
			}
		})
	}

	t.Run("a duplicate id is refused by the primary key", func(t *testing.T) {
		r, _ := nodeRepo(t)
		mustCreate(t, ctx, r, newTestNode("dup"))

		if err := r.Create(ctx, newTestNode("dup")); !errors.Is(err, ErrConstraint) {
			t.Errorf("Create(duplicate) error = %v, want one matching ErrConstraint", err)
		}
	})

	t.Run("an unrelated failure is not reported as a constraint", func(t *testing.T) {
		r, _ := nodeRepo(t)

		// A cancelled context is the cheapest non-constraint failure there is.
		cancelled, cancel := context.WithCancel(ctx)
		cancel()

		err := r.Create(cancelled, newTestNode("cancelled"))
		if err == nil {
			t.Fatal("Create on a cancelled context returned nil, want an error")
		}
		if errors.Is(err, ErrConstraint) {
			t.Errorf("error = %v, want one that does NOT match ErrConstraint", err)
		}
	})
}

// The whole reason the repository is built over Executor: a service composes
// several writes into one transaction, and a rollback has to take all of them.
func TestNodeRepoRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	r, db := nodeRepo(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	inTx := r.WithExecutor(tx)
	mustCreate(t, ctx, inTx, newTestNode("pending"))

	if _, err := inTx.Get(ctx, "pending"); err != nil {
		t.Fatalf("Get inside the transaction: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	if _, err := r.Get(ctx, "pending"); !errors.Is(err, ErrNotFound) {
		t.Errorf("after a rollback, Get = %v, want one matching ErrNotFound", err)
	}
	if n := countRows(t, db, "nodes"); n != 0 {
		t.Errorf("nodes has %d rows after a rollback, want 0", n)
	}
}

// A row whose text columns the schema accepts but the mapper cannot read must
// be an error, not a zero value quietly handed to the UI.
func TestNodeRepoRejectsUnreadableRows(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name   string
		column string
		value  string
	}{
		{"due is not a date", "due", "next tuesday"},
		{"due is well formed but impossible", "due", "2026-02-30"},
		{"created_at is not a timestamp", "created_at", "2026-09-18 09:00:00"},
		{"updated_at carries an offset", "updated_at", "2026-09-18T09:00:00+03:00"},
		{"completed_at is empty", "completed_at", ""},
		{"archived_at is a date, not a timestamp", "archived_at", "2026-09-18"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, db := nodeRepo(t)
			mustCreate(t, ctx, r, newTestNode("n"))

			// Straight SQL: the point is a row the repository did not write.
			stmt := fmt.Sprintf("UPDATE nodes SET %s = ? WHERE id = 'n'", tc.column)
			if _, err := db.ExecContext(ctx, stmt, tc.value); err != nil {
				t.Fatalf("corrupting %s: %v", tc.column, err)
			}

			if _, err := r.Get(ctx, "n"); err == nil {
				t.Errorf("Get on a row with %s = %q returned no error", tc.column, tc.value)
			}
		})
	}
}

// The classifier has to tell a constraint failure from everything else by
// result code, not by message: a CHECK and a UNIQUE are both constraints, a
// syntax error is not, and none of them can be trusted to keep their wording.
func TestConstraintClassification(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	mustInsertNode(t, ctx, db, "n", nil)

	for _, tc := range []struct {
		name           string
		stmt           string
		wantConstraint bool
		wantCode       int
	}{
		{
			name:           "CHECK",
			stmt:           "UPDATE nodes SET status = 'Done' WHERE id = 'n'",
			wantConstraint: true,
			wantCode:       sqlite3.SQLITE_CONSTRAINT_CHECK,
		},
		{
			name:           "PRIMARY KEY",
			stmt:           "INSERT INTO nodes (id, type, title, status, created_at, updated_at) VALUES ('n', 'task', 't', 'backlog', '2026-09-18T09:00:00Z', '2026-09-18T09:00:00Z')",
			wantConstraint: true,
			wantCode:       sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY,
		},
		{
			name:           "FOREIGN KEY",
			stmt:           "UPDATE nodes SET parent_id = 'ghost' WHERE id = 'n'",
			wantConstraint: true,
			wantCode:       sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY,
		},
		{
			name:           "NOT NULL",
			stmt:           "UPDATE nodes SET title = NULL WHERE id = 'n'",
			wantConstraint: true,
			wantCode:       sqlite3.SQLITE_CONSTRAINT_NOTNULL,
		},
		{
			name:           "a syntax error is not a constraint",
			stmt:           "UPDATE nodes SET nonexistent_column = 1",
			wantConstraint: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := db.ExecContext(ctx, tc.stmt)
			if err == nil {
				t.Fatal("the statement succeeded; it was supposed to fail")
			}
			if got := isConstraintViolation(err); got != tc.wantConstraint {
				t.Errorf("isConstraintViolation(%v) = %v, want %v", err, got, tc.wantConstraint)
			}
			if tc.wantConstraint && !isConstraintCode(err, tc.wantCode) {
				code, _ := sqliteErrorCode(err)
				t.Errorf("extended code = %d, want %d", code, tc.wantCode)
			}
			if wrapped := wrapExec("op", err); errors.Is(wrapped, ErrConstraint) != tc.wantConstraint {
				t.Errorf("errors.Is(wrapExec(err), ErrConstraint) = %v, want %v",
					errors.Is(wrapped, ErrConstraint), tc.wantConstraint)
			}
		})
	}
}

// The timestamp format is a contract with migration 0002 and with every later
// query that compares two timestamps as strings.
func TestTimestampFormat(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   time.Time
		want string
	}{
		{"UTC", time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC), "2026-09-18T09:00:00Z"},
		{"east of UTC", time.Date(2026, 9, 18, 12, 0, 0, 0, time.FixedZone("+3", 3*3600)), "2026-09-18T09:00:00Z"},
		{"west of UTC", time.Date(2026, 9, 18, 4, 0, 0, 0, time.FixedZone("-5", -5*3600)), "2026-09-18T09:00:00Z"},
		{"sub-second precision is dropped, not rounded up", time.Date(2026, 9, 18, 9, 0, 0, 999_000_000, time.UTC), "2026-09-18T09:00:00Z"},
		{"midnight", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), "2026-01-01T00:00:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatTimestamp(tc.in); got != tc.want {
				t.Errorf("formatTimestamp(%s) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}

	t.Run("parseTimestamp is formatTimestamp's inverse", func(t *testing.T) {
		got, err := parseTimestamp("2026-09-18T09:00:00Z")
		if err != nil {
			t.Fatalf("parseTimestamp: %v", err)
		}
		if want := time.Date(2026, 9, 18, 9, 0, 0, 0, time.UTC); !got.Equal(want) {
			t.Errorf("parseTimestamp = %s, want %s", got, want)
		}
		if got.Location() != time.UTC {
			t.Errorf("location = %s, want UTC", got.Location())
		}
	})

	t.Run("parseTimestamp refuses anything else", func(t *testing.T) {
		for _, bad := range []string{"", "2026-09-18", "2026-09-18 09:00:00", "2026-09-18T09:00:00+03:00", "not a time"} {
			if _, err := parseTimestamp(bad); err == nil {
				t.Errorf("parseTimestamp(%q) = nil error, want one", bad)
			}
		}
	})
}
