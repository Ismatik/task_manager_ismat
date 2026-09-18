package domain_test

import (
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
)

// deepTree is the fixture every move subtest works against:
//
//	root
//	├── a            (project)
//	│   ├── a1       (task, week)
//	│   │   └── a1x  (task, doing)
//	│   │       └── a1xy (task, backlog)   <- great-grandchild of a
//	│   └── a2       (task, done)
//	├── b            (project)
//	│   └── b1       (task, backlog)
//	└── n            (note)
func deepTree() []domain.Node {
	return []domain.Node{
		project("root", "", domain.StatusBacklog),
		project("a", "root", domain.StatusBacklog),
		task("a1", "a", domain.StatusWeek),
		task("a1x", "a1", domain.StatusDoing),
		task("a1xy", "a1x", domain.StatusBacklog),
		task("a2", "a", domain.StatusDone),
		project("b", "root", domain.StatusBacklog),
		task("b1", "b", domain.StatusBacklog),
		note("n", "root"),
	}
}

func ids(nodes []domain.Node) []string {
	out := make([]string, len(nodes))
	for i, n := range nodes {
		out[i] = n.ID
	}
	return out
}

func orderIDs(changes []domain.OrderChange) []string {
	out := make([]string, len(changes))
	for i, c := range changes {
		out[i] = c.NodeID
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestChildren(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name   string
		parent string
		want   []string
	}{
		{"the roots", "", []string{"root"}},
		{"a branch", "root", []string{"a", "b", "n"}},
		{"a nested branch", "a", []string{"a1", "a2"}},
		{"a leaf has none", "a1xy", nil},
		{"a note has none", "n", nil},
		{"an unknown id has none", "nope", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ids(domain.Children(nodes, tt.parent))
			if !equalStrings(got, tt.want) {
				t.Errorf("Children(%q) = %v, want %v", tt.parent, got, tt.want)
			}
		})
	}
}

// Siblings come back in (sort_order, id) order, not in load order.
func TestChildrenAreOrderedBySortOrderThenID(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		task("z", "p", domain.StatusBacklog),
		task("y", "p", domain.StatusBacklog),
		task("late", "p", domain.StatusBacklog),
	}
	nodes[1].SortOrder = 0 // z
	nodes[2].SortOrder = 0 // y  — ties with z, so the id breaks it
	nodes[3].SortOrder = 5 // late

	want := []string{"y", "z", "late"}
	if got := ids(domain.Children(nodes, "p")); !equalStrings(got, want) {
		t.Errorf("Children() = %v, want %v", got, want)
	}
}

func TestSubtreeAndDescendants(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name string
		id   string
		want []string // pre-order, including the root of the subtree
	}{
		{"the whole tree", "root", []string{"root", "a", "a1", "a1x", "a1xy", "a2", "b", "b1", "n"}},
		{"a branch", "a", []string{"a", "a1", "a1x", "a1xy", "a2"}},
		{"a chain", "a1", []string{"a1", "a1x", "a1xy"}},
		{"a leaf", "a1xy", []string{"a1xy"}},
		{"a note", "n", []string{"n"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub, err := domain.Subtree(nodes, tt.id)
			if err != nil {
				t.Fatalf("Subtree(%q) = %v", tt.id, err)
			}
			if got := ids(sub); !equalStrings(got, tt.want) {
				t.Errorf("Subtree(%q) = %v, want %v", tt.id, got, tt.want)
			}

			desc, err := domain.Descendants(nodes, tt.id)
			if err != nil {
				t.Fatalf("Descendants(%q) = %v", tt.id, err)
			}
			if got := ids(desc); !equalStrings(got, tt.want[1:]) {
				t.Errorf("Descendants(%q) = %v, want %v", tt.id, got, tt.want[1:])
			}
		})
	}
}

func TestSubtreeErrors(t *testing.T) {
	t.Run("an unknown id", func(t *testing.T) {
		if _, err := domain.Subtree(deepTree(), "nope"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
		if _, err := domain.Descendants(deepTree(), "nope"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	// Corrupt data: the walk is bounded, so this returns rather than hanging.
	// If it regresses, the test never finishes and `go test` kills the package.
	t.Run("an already cyclic graph terminates with ErrCycle", func(t *testing.T) {
		nodes := []domain.Node{
			project("x", "y", domain.StatusBacklog),
			project("y", "x", domain.StatusBacklog),
		}
		if _, err := domain.Subtree(nodes, "x"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})

	t.Run("a self-parenting node terminates with ErrCycle", func(t *testing.T) {
		nodes := []domain.Node{project("x", "x", domain.StatusBacklog)}
		if _, err := domain.Subtree(nodes, "x"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})

	t.Run("a three-node ring terminates with ErrCycle", func(t *testing.T) {
		nodes := []domain.Node{
			project("x", "z", domain.StatusBacklog),
			project("y", "x", domain.StatusBacklog),
			project("z", "y", domain.StatusBacklog),
		}
		if _, err := domain.Subtree(nodes, "x"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})
}

func TestAncestors(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name string
		id   string
		want []string // nearest first
	}{
		{"the root has none", "root", nil},
		{"one level", "a", []string{"root"}},
		{"three levels", "a1xy", []string{"a1x", "a1", "a", "root"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.Ancestors(nodes, tt.id)
			if err != nil {
				t.Fatalf("Ancestors(%q) = %v", tt.id, err)
			}
			if !equalStrings(ids(got), tt.want) {
				t.Errorf("Ancestors(%q) = %v, want %v", tt.id, ids(got), tt.want)
			}
		})
	}

	t.Run("an unknown id", func(t *testing.T) {
		if _, err := domain.Ancestors(nodes, "nope"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	// A partially loaded subtree is normal: the chain stops at the top of what
	// was loaded rather than failing.
	t.Run("a parent outside the loaded set ends the chain", func(t *testing.T) {
		partial := []domain.Node{task("orphan", "not-loaded", domain.StatusWeek)}
		got, err := domain.Ancestors(partial, "orphan")
		if err != nil {
			t.Fatalf("Ancestors() = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("Ancestors() = %v, want empty", ids(got))
		}
	})

	t.Run("a cyclic chain terminates with ErrCycle", func(t *testing.T) {
		nodes := []domain.Node{
			project("x", "y", domain.StatusBacklog),
			project("y", "x", domain.StatusBacklog),
		}
		if _, err := domain.Ancestors(nodes, "x"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})
}

// The circular-parent rejection, at every depth the brief names.
func TestValidateMoveRejectsACircularParent(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name      string
		nodeID    string
		newParent string
	}{
		{"into itself", "a", "a"},
		{"into itself at the root", "root", "root"},
		{"into its direct child", "a", "a1"},
		{"into a grandchild at depth 3", "a", "a1x"},
		{"into a great-grandchild at depth 4", "a", "a1xy"},
		{"the whole tree into its deepest leaf", "root", "a1xy"},
		{"a mid-chain node into its own child", "a1", "a1x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidateMove(nodes, tt.nodeID, ptr(tt.newParent))
			if !errors.Is(err, domain.ErrCircularParent) {
				t.Errorf("ValidateMove(%q -> %q) = %v, want ErrCircularParent",
					tt.nodeID, tt.newParent, err)
			}
		})
	}
}

func TestValidateMoveAccepts(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name      string
		nodeID    string
		newParent *string
	}{
		{"to a sibling", "a1", ptr("a2")},
		{"to the root", "a1", nil},
		{"a root stays a root", "root", nil},
		{"to an unrelated subtree", "a1", ptr("b")},
		{"to an unrelated leaf", "a1", ptr("b1")},
		{"to its own current parent", "a1", ptr("a")},
		{"upwards to its grandparent", "a1x", ptr("root")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := domain.ValidateMove(nodes, tt.nodeID, tt.newParent); err != nil {
				t.Errorf("ValidateMove() = %v, want nil", err)
			}
		})
	}
}

func TestValidateMoveRejectsMissingNodes(t *testing.T) {
	nodes := deepTree()

	t.Run("an unknown node", func(t *testing.T) {
		if err := domain.ValidateMove(nodes, "nope", ptr("a")); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
	t.Run("an unknown node moved to the root", func(t *testing.T) {
		if err := domain.ValidateMove(nodes, "nope", nil); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
	t.Run("an unknown new parent", func(t *testing.T) {
		if err := domain.ValidateMove(nodes, "a1", ptr("nope")); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
}

// Notes hold no children: work parked under a note would be invisible to both
// status derivation and the progress denominator, so the move is refused.
func TestValidateMoveRejectsANoteAsParent(t *testing.T) {
	nodes := deepTree()

	for _, id := range []string{"a1", "b", "a2"} {
		t.Run("moving "+id+" under a note", func(t *testing.T) {
			err := domain.ValidateMove(nodes, id, ptr("n"))
			if !errors.Is(err, domain.ErrNoteParent) {
				t.Errorf("ValidateMove(%q -> note) = %v, want ErrNoteParent", id, err)
			}
		})
	}

	t.Run("a note under itself is circular, not a note-parent problem", func(t *testing.T) {
		err := domain.ValidateMove(nodes, "n", ptr("n"))
		if !errors.Is(err, domain.ErrCircularParent) {
			t.Errorf("err = %v, want ErrCircularParent", err)
		}
	})
}

func TestValidateMoveOnAlreadyCyclicData(t *testing.T) {
	nodes := []domain.Node{
		project("x", "y", domain.StatusBacklog),
		project("y", "x", domain.StatusBacklog),
		project("free", "", domain.StatusBacklog),
	}
	if err := domain.ValidateMove(nodes, "free", ptr("x")); !errors.Is(err, domain.ErrCycle) {
		t.Errorf("err = %v, want ErrCycle", err)
	}
}

func TestReorder(t *testing.T) {
	siblings := func() []domain.Node {
		out := []domain.Node{
			task("s0", "p", domain.StatusBacklog),
			task("s1", "p", domain.StatusBacklog),
			task("s2", "p", domain.StatusBacklog),
			task("s3", "p", domain.StatusBacklog),
		}
		// Gappy, database-shaped sort_orders.
		for i := range out {
			out[i].SortOrder = i * 10
		}
		return out
	}

	tests := []struct {
		name    string
		nodeID  string
		toIndex int
		want    []string
	}{
		{"to the first position", "s2", 0, []string{"s2", "s0", "s1", "s3"}},
		{"to the last position", "s0", 3, []string{"s1", "s2", "s3", "s0"}},
		{"to a middle position", "s3", 1, []string{"s0", "s3", "s1", "s2"}},
		{"to where it already is", "s1", 1, []string{"s0", "s1", "s2", "s3"}},
		{"past the end is clamped to the end", "s0", 99, []string{"s1", "s2", "s3", "s0"}},
		{"before the start is clamped to the start", "s3", -5, []string{"s3", "s0", "s1", "s2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes, err := domain.Reorder(siblings(), tt.nodeID, tt.toIndex)
			if err != nil {
				t.Fatalf("Reorder() = %v", err)
			}
			if got := orderIDs(changes); !equalStrings(got, tt.want) {
				t.Fatalf("Reorder() = %v, want %v", got, tt.want)
			}
			assertGapless(t, changes)
		})
	}

	t.Run("a single-element set is a no-op", func(t *testing.T) {
		only := []domain.Node{task("only", "p", domain.StatusBacklog)}
		changes, err := domain.Reorder(only, "only", 0)
		if err != nil {
			t.Fatalf("Reorder() = %v", err)
		}
		want := []domain.OrderChange{{NodeID: "only", SortOrder: 0}}
		if len(changes) != 1 || changes[0] != want[0] {
			t.Errorf("Reorder() = %+v, want %+v", changes, want)
		}
	})

	t.Run("an unknown node", func(t *testing.T) {
		if _, err := domain.Reorder(siblings(), "nope", 0); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	t.Run("applying the plan twice is a no-op", func(t *testing.T) {
		first, err := domain.Reorder(siblings(), "s3", 1)
		if err != nil {
			t.Fatalf("Reorder() = %v", err)
		}
		applied := applyOrder(siblings(), first)
		second, err := domain.Reorder(applied, "s3", 1)
		if err != nil {
			t.Fatalf("Reorder() = %v", err)
		}
		if !equalStrings(orderIDs(first), orderIDs(second)) {
			t.Errorf("second pass = %v, want %v", orderIDs(second), orderIDs(first))
		}
	})

	t.Run("ties are broken by id so the result never depends on load order", func(t *testing.T) {
		tied := []domain.Node{
			task("c", "p", domain.StatusBacklog),
			task("a", "p", domain.StatusBacklog),
			task("b", "p", domain.StatusBacklog),
		}
		changes, err := domain.Reorder(tied, "b", 2)
		if err != nil {
			t.Fatalf("Reorder() = %v", err)
		}
		if got, want := orderIDs(changes), []string{"a", "c", "b"}; !equalStrings(got, want) {
			t.Errorf("Reorder() = %v, want %v", got, want)
		}
	})
}

func assertGapless(t *testing.T, changes []domain.OrderChange) {
	t.Helper()
	for i, c := range changes {
		if c.SortOrder != i {
			t.Errorf("changes[%d].SortOrder = %d, want %d — the sequence must be dense 0..n-1",
				i, c.SortOrder, i)
		}
	}
}

func applyOrder(nodes []domain.Node, changes []domain.OrderChange) []domain.Node {
	out := make([]domain.Node, len(nodes))
	copy(out, nodes)
	for _, c := range changes {
		for i := range out {
			if out[i].ID == c.NodeID {
				out[i].SortOrder = c.SortOrder
			}
		}
	}
	return out
}

// Moving a node moves its whole subtree: exactly one parent_id changes, and no
// descendant's parent_id or status is touched.
func TestPlanMoveMovesTheWholeSubtree(t *testing.T) {
	nodes := deepTree()

	plan, err := domain.PlanMove(nodes, "a1", ptr("b"), 0)
	if err != nil {
		t.Fatalf("PlanMove() = %v", err)
	}

	if plan.NodeID != "a1" {
		t.Errorf("NodeID = %q, want a1", plan.NodeID)
	}
	if plan.NewParentID == nil || *plan.NewParentID != "b" {
		t.Fatalf("NewParentID = %v, want b", plan.NewParentID)
	}

	// The plan names exactly one node as re-parented, and it is not a
	// descendant.
	descendants, err := domain.Descendants(nodes, "a1")
	if err != nil {
		t.Fatalf("Descendants() = %v", err)
	}
	for _, d := range descendants {
		if d.ID == plan.NodeID {
			t.Errorf("the plan re-parents the descendant %q", d.ID)
		}
	}

	// The original nodes are untouched: statuses and parent ids of the subtree
	// are exactly what they were.
	after := deepTree()
	for i := range after {
		before := nodes[i]
		if after[i].Status != before.Status {
			t.Errorf("%q status changed from %q to %q", before.ID, before.Status, after[i].Status)
		}
	}
	for _, d := range descendants {
		switch d.ID {
		case "a1x":
			if d.ParentID == nil || *d.ParentID != "a1" || d.Status != domain.StatusDoing {
				t.Errorf("a1x = parent %v status %q, want parent a1 status doing", d.ParentID, d.Status)
			}
		case "a1xy":
			if d.ParentID == nil || *d.ParentID != "a1x" || d.Status != domain.StatusBacklog {
				t.Errorf("a1xy = parent %v status %q, want parent a1x status backlog", d.ParentID, d.Status)
			}
		}
	}

	// The two disturbed sibling ranges are dense.
	if got, want := orderIDs(plan.OldSiblings), []string{"a2"}; !equalStrings(got, want) {
		t.Errorf("OldSiblings = %v, want %v", got, want)
	}
	assertGapless(t, plan.OldSiblings)

	if got, want := orderIDs(plan.NewSiblings), []string{"a1", "b1"}; !equalStrings(got, want) {
		t.Errorf("NewSiblings = %v, want %v", got, want)
	}
	assertGapless(t, plan.NewSiblings)
}

func TestPlanMove(t *testing.T) {
	// twoRoots is deepTree with a second root, so that a root can be moved
	// under another node — the case where the node being moved has no parent
	// at all.
	twoRoots := func() []domain.Node {
		return append(deepTree(), project("root2", "", domain.StatusBacklog))
	}

	tests := []struct {
		name         string
		nodes        func() []domain.Node
		nodeID       string
		newParent    *string
		toIndex      int
		wantOld      []string
		wantNew      []string
		wantParentID *string
	}{
		{
			name: "to the end of another branch", nodes: deepTree,
			nodeID: "a2", newParent: ptr("b"), toIndex: 9,
			wantOld: []string{"a1"}, wantNew: []string{"b1", "a2"}, wantParentID: ptr("b"),
		},
		{
			name: "to the root", nodes: deepTree,
			nodeID: "a1", newParent: nil, toIndex: 0,
			wantOld: []string{"a2"}, wantNew: []string{"a1", "root"}, wantParentID: nil,
		},
		{
			name: "within the same parent leaves the old range empty", nodes: deepTree,
			nodeID: "a", newParent: ptr("root"), toIndex: 2,
			wantOld: nil, wantNew: []string{"b", "n", "a"}, wantParentID: ptr("root"),
		},
		{
			name: "the only child of its parent", nodes: deepTree,
			nodeID: "b1", newParent: ptr("a"), toIndex: 0,
			wantOld: nil, wantNew: []string{"b1", "a1", "a2"}, wantParentID: ptr("a"),
		},
		{
			name: "a root moved under another node leaves the root range", nodes: twoRoots,
			nodeID: "root2", newParent: ptr("b"), toIndex: 0,
			wantOld: []string{"root"}, wantNew: []string{"root2", "b1"}, wantParentID: ptr("b"),
		},
		{
			name: "a root reordered among the roots", nodes: twoRoots,
			nodeID: "root2", newParent: nil, toIndex: 0,
			wantOld: nil, wantNew: []string{"root2", "root"}, wantParentID: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := domain.PlanMove(tt.nodes(), tt.nodeID, tt.newParent, tt.toIndex)
			if err != nil {
				t.Fatalf("PlanMove() = %v", err)
			}
			switch {
			case tt.wantParentID == nil && plan.NewParentID != nil:
				t.Errorf("NewParentID = %q, want nil", *plan.NewParentID)
			case tt.wantParentID != nil && plan.NewParentID == nil:
				t.Errorf("NewParentID = nil, want %q", *tt.wantParentID)
			case tt.wantParentID != nil && *plan.NewParentID != *tt.wantParentID:
				t.Errorf("NewParentID = %q, want %q", *plan.NewParentID, *tt.wantParentID)
			}
			if got := orderIDs(plan.OldSiblings); !equalStrings(got, tt.wantOld) {
				t.Errorf("OldSiblings = %v, want %v", got, tt.wantOld)
			}
			if got := orderIDs(plan.NewSiblings); !equalStrings(got, tt.wantNew) {
				t.Errorf("NewSiblings = %v, want %v", got, tt.wantNew)
			}
			assertGapless(t, plan.OldSiblings)
			assertGapless(t, plan.NewSiblings)
		})
	}
}

func TestPlanMoveRefusesAnInvalidMove(t *testing.T) {
	nodes := deepTree()

	tests := []struct {
		name      string
		nodeID    string
		newParent *string
		want      error
	}{
		{"into its own subtree", "a", ptr("a1x"), domain.ErrCircularParent},
		{"into itself", "a", ptr("a"), domain.ErrCircularParent},
		{"under a note", "a", ptr("n"), domain.ErrNoteParent},
		{"an unknown node", "nope", ptr("a"), domain.ErrNodeNotFound},
		{"an unknown parent", "a", ptr("nope"), domain.ErrNodeNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := domain.PlanMove(nodes, tt.nodeID, tt.newParent, 0)
			if !errors.Is(err, tt.want) {
				t.Fatalf("PlanMove() = %v, want %v", err, tt.want)
			}
			if plan.NodeID != "" || plan.NewSiblings != nil || plan.OldSiblings != nil {
				t.Errorf("PlanMove() returned a plan alongside the error: %+v", plan)
			}
		})
	}
}

// The plan must not alias the caller's parent id.
func TestPlanMoveDoesNotAliasTheCallersParentID(t *testing.T) {
	parent := "b"
	plan, err := domain.PlanMove(deepTree(), "a1", &parent, 0)
	if err != nil {
		t.Fatalf("PlanMove() = %v", err)
	}
	if plan.NewParentID == &parent {
		t.Fatal("PlanMove stored the caller's pointer")
	}
	*plan.NewParentID = "mutated"
	if parent != "b" {
		t.Errorf("mutating the plan changed the caller's value: %q", parent)
	}
}

func TestPlanArchiveCoversTheWholeSubtree(t *testing.T) {
	nodes := deepTree()
	at := fixedNow
	now := func() time.Time { return at }

	changes, err := domain.PlanArchive(nodes, "a", now)
	if err != nil {
		t.Fatalf("PlanArchive() = %v", err)
	}

	want := []string{"a", "a1", "a1x", "a1xy", "a2"}
	got := make([]string, len(changes))
	for i, c := range changes {
		got[i] = c.NodeID
		if c.ArchivedAt == nil {
			t.Fatalf("%q: ArchivedAt = nil, want the injected now", c.NodeID)
		}
		if !c.ArchivedAt.Equal(at) {
			t.Errorf("%q: ArchivedAt = %v, want %v", c.NodeID, c.ArchivedAt, at)
		}
	}
	if !equalStrings(got, want) {
		t.Errorf("PlanArchive() = %v, want %v", got, want)
	}

	t.Run("a leaf archives only itself", func(t *testing.T) {
		changes, err := domain.PlanArchive(nodes, "b1", now)
		if err != nil {
			t.Fatalf("PlanArchive() = %v", err)
		}
		if len(changes) != 1 || changes[0].NodeID != "b1" {
			t.Errorf("PlanArchive(leaf) = %+v, want exactly b1", changes)
		}
	})

	t.Run("an unknown id", func(t *testing.T) {
		if _, err := domain.PlanArchive(nodes, "nope", now); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
}

// Archiving does not re-stamp a node that was already archived: archived_at
// records when something was put away, and that history is not rewritten.
func TestPlanArchiveKeepsAnExistingArchivedAt(t *testing.T) {
	nodes := deepTree()
	lastMonth := fixedNow.AddDate(0, -1, 0)
	for i := range nodes {
		if nodes[i].ID == "a1x" {
			nodes[i].ArchivedAt = &lastMonth
		}
	}

	changes, err := domain.PlanArchive(nodes, "a", func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("PlanArchive() = %v", err)
	}
	for _, c := range changes {
		if c.NodeID == "a1x" {
			t.Fatalf("the plan re-stamps the already-archived a1x with %v", c.ArchivedAt)
		}
	}
	if got, want := len(changes), 4; got != want {
		t.Errorf("len(changes) = %d, want %d", got, want)
	}
}

// Each change carries its own timestamp, so applying one cannot alter another.
func TestPlanArchiveDoesNotShareOneTimestampPointer(t *testing.T) {
	changes, err := domain.PlanArchive(deepTree(), "a", func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("PlanArchive() = %v", err)
	}
	if len(changes) < 2 {
		t.Fatalf("expected several changes, got %d", len(changes))
	}
	if changes[0].ArchivedAt == changes[1].ArchivedAt {
		t.Fatal("two changes share one *time.Time")
	}
	*changes[0].ArchivedAt = time.Time{}
	if changes[1].ArchivedAt.IsZero() {
		t.Error("mutating one change's timestamp changed another's")
	}
}

// The documented nested rule: restoring a parent restores everything under it,
// including a child that was archived separately beforehand. archived_at is not
// a stack.
func TestPlanRestoreClearsTheWholeSubtreeIncludingSeparatelyArchivedChildren(t *testing.T) {
	nodes := deepTree()
	lastMonth := fixedNow.AddDate(0, -1, 0)
	today := fixedNow

	for i := range nodes {
		switch nodes[i].ID {
		case "a1x":
			// archived on its own, a month before its parent
			nodes[i].ArchivedAt = &lastMonth
		case "a", "a1", "a1xy", "a2":
			nodes[i].ArchivedAt = &today
		}
	}

	changes, err := domain.PlanRestore(nodes, "a")
	if err != nil {
		t.Fatalf("PlanRestore() = %v", err)
	}

	want := []string{"a", "a1", "a1x", "a1xy", "a2"}
	got := make([]string, len(changes))
	for i, c := range changes {
		got[i] = c.NodeID
		if c.ArchivedAt != nil {
			t.Errorf("%q: ArchivedAt = %v, want nil", c.NodeID, c.ArchivedAt)
		}
	}
	if !equalStrings(got, want) {
		t.Fatalf("PlanRestore() = %v, want %v — the separately archived a1x must come back too", got, want)
	}
}

func TestPlanRestoreSkipsNodesThatAreNotArchived(t *testing.T) {
	nodes := deepTree() // nothing is archived

	changes, err := domain.PlanRestore(nodes, "a")
	if err != nil {
		t.Fatalf("PlanRestore() = %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("PlanRestore() = %+v, want no changes", changes)
	}

	t.Run("an unknown id", func(t *testing.T) {
		if _, err := domain.PlanRestore(nodes, "nope"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
}

// Archive then restore is a round trip over the same node set.
func TestArchiveThenRestoreIsARoundTrip(t *testing.T) {
	nodes := deepTree()

	archived, err := domain.PlanArchive(nodes, "a", func() time.Time { return fixedNow })
	if err != nil {
		t.Fatalf("PlanArchive() = %v", err)
	}
	applied := make([]domain.Node, len(nodes))
	copy(applied, nodes)
	for _, c := range archived {
		for i := range applied {
			if applied[i].ID == c.NodeID {
				applied[i].ArchivedAt = c.ArchivedAt
			}
		}
	}

	restored, err := domain.PlanRestore(applied, "a")
	if err != nil {
		t.Fatalf("PlanRestore() = %v", err)
	}
	if len(restored) != len(archived) {
		t.Errorf("restore covers %d nodes, archive covered %d", len(restored), len(archived))
	}
	for i := range restored {
		if restored[i].NodeID != archived[i].NodeID {
			t.Errorf("restored[%d] = %q, want %q", i, restored[i].NodeID, archived[i].NodeID)
		}
	}
}

// Every plan in this package must be byte-identical across runs on the same
// input.
func TestTreePlansAreDeterministic(t *testing.T) {
	nodes := deepTree()
	now := func() time.Time { return fixedNow }

	for range 3 {
		m1, err := domain.PlanMove(nodes, "a1", ptr("b"), 1)
		if err != nil {
			t.Fatalf("PlanMove() = %v", err)
		}
		m2, err := domain.PlanMove(nodes, "a1", ptr("b"), 1)
		if err != nil {
			t.Fatalf("PlanMove() = %v", err)
		}
		if !equalStrings(orderIDs(m1.NewSiblings), orderIDs(m2.NewSiblings)) ||
			!equalStrings(orderIDs(m1.OldSiblings), orderIDs(m2.OldSiblings)) {
			t.Fatalf("PlanMove is not deterministic: %+v vs %+v", m1, m2)
		}

		a1, err := domain.PlanArchive(nodes, "root", now)
		if err != nil {
			t.Fatalf("PlanArchive() = %v", err)
		}
		a2, err := domain.PlanArchive(nodes, "root", now)
		if err != nil {
			t.Fatalf("PlanArchive() = %v", err)
		}
		for i := range a1 {
			if a1[i].NodeID != a2[i].NodeID {
				t.Fatalf("PlanArchive is not deterministic at %d: %q vs %q", i, a1[i].NodeID, a2[i].NodeID)
			}
		}
	}
}
