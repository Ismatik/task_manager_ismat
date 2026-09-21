package service_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"nexus/internal/domain"
	"nexus/internal/service"
)

// column returns the cards in one column of a board.
func column(t *testing.T, board []service.ColumnView, status domain.Status) []service.NodeView {
	t.Helper()

	for _, c := range board {
		if c.Status == status {
			return c.Nodes
		}
	}
	t.Fatalf("the board has no %q column", status)
	return nil
}

// titles lists the titles of a column, which is what a failure message should
// say rather than six struct dumps.
func titles(views []service.NodeView) []string {
	out := make([]string, len(views))
	for i, v := range views {
		out[i] = v.Node.Title
	}
	return out
}

// find returns the view of the node with this id from a board, or nil.
func find(board []service.ColumnView, id string) *service.NodeView {
	for _, c := range board {
		for i := range c.Nodes {
			if c.Nodes[i].Node.ID == id {
				return &c.Nodes[i]
			}
		}
	}
	return nil
}

// THE derived-placement test: a parent whose STORED status disagrees with its
// children is placed by what its children say, not by what the column says.
func TestBoardPlacesAParentInItsDerivedColumn(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	parent := f.create(draft("parent", domain.NodeTypeProject, nil))
	child := f.create(draft("child", domain.NodeTypeTask, &parent.ID))
	if _, err := f.tasks.MoveToColumn(ctx, child.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn(child, today): %v", err)
	}

	// A stored status that says something else entirely — the state a bug, or
	// an older version, could leave behind. It must not reach the board.
	stored := f.get(parent.ID)
	stored.Status = domain.StatusDone
	if err := f.nodes.Update(ctx, stored); err != nil {
		t.Fatalf("Update: %v", err)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	if got := titles(column(t, board, domain.StatusToday)); len(got) != 2 {
		t.Errorf("the today column holds %v, want both the parent and the child", got)
	}
	if got := titles(column(t, board, domain.StatusDone)); len(got) != 0 {
		t.Errorf("the done column holds %v, want nothing", got)
	}

	v := find(board, parent.ID)
	if v == nil {
		t.Fatal("the parent is not on the board at all")
	}
	if v.Status != domain.StatusToday {
		t.Errorf("the parent's derived status = %q, want today", v.Status)
	}
	if v.Node.Status != domain.StatusDone {
		t.Errorf("the stored status was rewritten to %q; the board must only READ it", v.Node.Status)
	}
	if v.IsLeaf {
		t.Error("the parent reports IsLeaf = true")
	}
}

// D2: habits are the habit strip's, never a column's. This is the board half of
// the cross-check whose other half is in habit_test.go (S1-20).
func TestBoardExcludesHabits(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	d := draft("stretch every morning", domain.NodeTypeHabit, nil)
	d.Recurrence = ptr("FREQ=DAILY")
	h := f.create(d)
	f.create(draft("a real task", domain.NodeTypeTask, nil))

	// Even with a Kanban status stored on it, a habit is not a card.
	stored := f.get(h.ID)
	stored.Status = domain.StatusToday
	if err := f.nodes.Update(ctx, stored); err != nil {
		t.Fatalf("Update: %v", err)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if v := find(board, h.ID); v != nil {
		t.Errorf("the habit is on the board, in the %q column", v.Status)
	}

	total := 0
	for _, c := range board {
		total += len(c.Nodes)
	}
	if total != 1 {
		t.Errorf("%d cards on the board, want only the task", total)
	}
}

// D10, end to end: a project holding one done task and one habit is placed in
// the DONE column, and its bar reads 100%.
//
// This is the motivating case of the decision, asserted where the user actually
// sees it. The project rendered in Backlog before D10 — DeriveStatus scanned the
// habit, which is parked at the inert backlog its NOT NULL column needs and can
// never leave — while ComputeProgress, which had already generalised the rule to
// NodeType.HasColumn, drew a full bar on the same card. The column and the bar
// are asserted in one test so that neither can go green alone.
func TestBoardPlacesAProjectWithADoneTaskAndAHabitInDone(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("project", domain.NodeTypeProject, nil))
	task := f.create(draft("the only work in it", domain.NodeTypeTask, &p.ID))

	d := draft("stretch every morning", domain.NodeTypeHabit, &p.ID)
	d.Recurrence = ptr("FREQ=DAILY")
	h := f.create(d)

	if _, err := f.tasks.MoveToColumn(ctx, task.ID, domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(task, done): %v", err)
	}
	// The habit really is sitting at backlog, which is the whole point: it is
	// "no column", not "in the Backlog column".
	if got := f.get(h.ID).Status; got != domain.StatusBacklog {
		t.Fatalf("the habit's stored status = %q, want backlog", got)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	v := find(board, p.ID)
	if v == nil {
		t.Fatal("the project is not on the board at all")
	}
	if v.Status != domain.StatusDone {
		t.Errorf("the project derives %q and sits in that column; want done — the habit has no "+
			"column and cannot hold it back", v.Status)
	}
	if got := titles(column(t, board, domain.StatusBacklog)); len(got) != 0 {
		t.Errorf("the backlog column holds %v, want nothing", got)
	}

	if !v.Progress.Defined || v.Progress.Done != 1 || v.Progress.Total != 1 || v.Progress.Percent != 100 {
		t.Errorf("progress = %+v, want 1/1 at 100%% and defined", v.Progress)
	}
	if (v.Status == domain.StatusDone) != (v.Progress.Percent == 100) {
		t.Errorf("the column (%q) and the bar (%d%%) disagree on the same card",
			v.Status, v.Progress.Percent)
	}
	// The task child DOES have a column, so the project is still a parent. Only
	// the habit is invisible to the leaf rule.
	if v.IsLeaf {
		t.Error("IsLeaf = true, but this project has a task child, which has a column")
	}

	t.Run("archive the task and the habit alone leaves a leaf", func(t *testing.T) {
		if _, err := f.tasks.ArchiveNode(ctx, task.ID); err != nil {
			t.Fatalf("ArchiveNode: %v", err)
		}
		board, err := f.tasks.Board(ctx)
		if err != nil {
			t.Fatalf("Board: %v", err)
		}
		v := find(board, p.ID)
		if v == nil {
			t.Fatal("the project left the board")
		}
		if !v.IsLeaf {
			t.Error("IsLeaf = false: the project's only remaining child is a habit, which has " +
				"no column")
		}
		// A leaf reports its own stored status, and nothing has ever written a
		// column onto this project, so it falls back to backlog.
		if v.Status != domain.StatusBacklog {
			t.Errorf("the project derives %q, want its own stored backlog", v.Status)
		}
	})
}

// S1-09, the reviewer's end-to-end reproduction, against a real database and
// through the public service API only.
//
// What it produced before the fix, with a task created under a habit:
//
//	project renders status="done" progress={Done:1 Total:1 Percent:100 Defined:true}
//	habit   renders status="today" isLeaf=false progress={Defined:false}
//	the buried task IS on the board, in column "today"
//
// Two rules broken at once: §4 says a habit never appears in a Kanban column,
// and a project read done at 100% while an unfinished task sat on the board
// underneath it. Both came from the same door — ValidateMove refusing a note
// parent but not a habit one — and both are asserted here where the user sees
// them.
func TestNothingCanBeParkedUnderAHabit(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("project", domain.NodeTypeProject, nil))
	work := f.create(draft("the only work in it", domain.NodeTypeTask, &p.ID))

	d := draft("stretch every morning", domain.NodeTypeHabit, &p.ID)
	d.Recurrence = ptr("FREQ=DAILY")
	h := f.create(d)

	if _, err := f.tasks.MoveToColumn(ctx, work.ID, domain.StatusDone); err != nil {
		t.Fatalf("MoveToColumn(work, done): %v", err)
	}

	// The door the reviewer came through: create.
	before := f.all()
	buried, err := f.tasks.CreateNode(ctx, draft("buried", domain.NodeTypeTask, &h.ID))
	if err == nil {
		t.Fatalf("CreateNode(task under a habit) created %q; want it refused", buried.ID)
	}
	if !errors.Is(err, domain.ErrTypeHasNoChildren) {
		t.Fatalf("CreateNode(task under a habit) = %v, want domain.ErrTypeHasNoChildren", err)
	}
	if after := f.all(); !reflect.DeepEqual(before, after) {
		t.Error("a refused create wrote to the database")
	}

	// And the other door: drag an existing task onto the habit.
	loose := f.create(draft("loose", domain.NodeTypeTask, nil))
	if _, err := f.tasks.MoveToColumn(ctx, loose.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn(loose, today): %v", err)
	}
	before = f.all()
	if _, err := f.tasks.MoveNode(ctx, loose.ID, &h.ID, 0); !errors.Is(err, domain.ErrTypeHasNoChildren) {
		t.Fatalf("MoveNode(task under a habit) = %v, want domain.ErrTypeHasNoChildren", err)
	}
	if after := f.all(); !reflect.DeepEqual(before, after) {
		t.Error("a refused move wrote to the database")
	}

	// What the user sees now.
	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if v := find(board, h.ID); v != nil {
		t.Errorf("the habit is on the board, in the %q column: §4 says it never appears in one",
			v.Status)
	}

	tree, err := f.tasks.Tree(ctx, nil)
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	hv := findInTree(tree, h.ID)
	if hv == nil {
		t.Fatal("the habit is not in the tree at all")
	}
	if hv.Status != domain.StatusBacklog {
		t.Errorf("the habit renders status %q, want backlog — which means \"no column\"", hv.Status)
	}
	if !hv.IsLeaf {
		t.Error("the habit renders IsLeaf = false, want true: nothing may be under it")
	}
	assertNoDoneAncestorOverUnfinishedWork(t, tree, f.all())

	// The derivation half holds even for data the service can no longer create:
	// a row written straight through the repository, as an older build could
	// have left behind.
	t.Run("a task already parked under a habit does not give it a column", func(t *testing.T) {
		legacy := domain.Node{
			ID: "legacy", ParentID: &h.ID, Type: domain.NodeTypeTask,
			Title: "written by an older build", Status: domain.StatusToday,
			DueSource: domain.DueSourceManual, Priority: domain.Priority4,
			CreatedAt: testNow, UpdatedAt: testNow,
		}
		if err := f.nodes.Create(ctx, legacy); err != nil {
			t.Fatalf("Create(legacy): %v", err)
		}

		board, err := f.tasks.Board(ctx)
		if err != nil {
			t.Fatalf("Board: %v", err)
		}
		if v := find(board, h.ID); v != nil {
			t.Errorf("the habit is on the board, in the %q column", v.Status)
		}

		tree, err := f.tasks.Tree(ctx, nil)
		if err != nil {
			t.Fatalf("Tree: %v", err)
		}
		hv := findInTree(tree, h.ID)
		if hv == nil {
			t.Fatal("the habit is not in the tree at all")
		}
		if hv.Status != domain.StatusBacklog {
			t.Errorf("the habit renders status %q, want backlog even with a task under it",
				hv.Status)
		}
		if !hv.IsLeaf {
			t.Error("the habit renders IsLeaf = false, want true even with a task under it")
		}

		// The column and the bar still tell the same story about the project.
		pv := findInTree(tree, p.ID)
		if pv == nil {
			t.Fatal("the project is not in the tree at all")
		}
		if (pv.Status == domain.StatusDone) != (pv.Progress.Percent == 100) {
			t.Errorf("the column (%q) and the bar (%d%%) disagree on the project",
				pv.Status, pv.Progress.Percent)
		}
	})
}

// findInTree returns the view of a node anywhere in a nested Tree result.
func findInTree(views []service.NodeView, id string) *service.NodeView {
	for i := range views {
		if views[i].Node.ID == id {
			return &views[i]
		}
		if v := findInTree(views[i].Children, id); v != nil {
			return v
		}
	}
	return nil
}

// assertNoDoneAncestorOverUnfinishedWork is the reviewer's second invariant: no
// node may render done at 100% while unfinished work sits anywhere below it in
// the STORED tree.
//
// "Below" is deliberately the raw parent_id chain, with no cut of its own: the
// board shows every node whose type has a column wherever it sits, so a task
// hidden under a habit is still on screen and still unfinished. The walk here
// shares no code with the derivation it checks.
func assertNoDoneAncestorOverUnfinishedWork(t *testing.T, views []service.NodeView, set []domain.Node) {
	t.Helper()

	kids := map[string][]domain.Node{}
	for _, n := range set {
		key := ""
		if n.ParentID != nil {
			key = *n.ParentID
		}
		kids[key] = append(kids[key], n)
	}

	// leaf reports whether n is where a stored status is the truth: nothing under
	// it has a column, so nothing derives over it.
	leaf := func(n domain.Node) bool {
		for _, c := range kids[n.ID] {
			if c.Type.HasColumn() {
				return false
			}
		}
		return true
	}

	// unfinished reports the first LEAF below id that has a column and is not
	// done. Only leaves are asked: a parent's stored status is never written and
	// means nothing, so reading it here would invent violations.
	var unfinished func(id string) string
	unfinished = func(id string) string {
		for _, c := range kids[id] {
			if c.Type.HasColumn() && leaf(c) && c.Status != domain.StatusDone {
				return c.ID
			}
			if found := unfinished(c.ID); found != "" {
				return found
			}
		}
		return ""
	}

	var walk func(vs []service.NodeView)
	walk = func(vs []service.NodeView) {
		for _, v := range vs {
			if v.Status == domain.StatusDone && v.Progress.Defined && v.Progress.Percent == 100 {
				if found := unfinished(v.Node.ID); found != "" {
					t.Errorf("%q renders done at 100%% while %q is unfinished below it",
						v.Node.ID, found)
				}
			}
			walk(v.Children)
		}
	}
	walk(views)
}

// Notes have no column, and they are not work: they count in no denominator.
func TestBoardAndProgressExcludeNotes(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("project", domain.NodeTypeProject, nil))
	memo := f.create(draft("memo", domain.NodeTypeNote, &p.ID))
	task := f.create(draft("task", domain.NodeTypeTask, &p.ID))

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if v := find(board, memo.ID); v != nil {
		t.Errorf("the note is on the board, in the %q column", v.Status)
	}

	t.Run("the note is not in the denominator", func(t *testing.T) {
		got, err := f.tasks.Progress(ctx, p.ID)
		if err != nil {
			t.Fatalf("Progress: %v", err)
		}
		if got.Total != 1 || got.Done != 0 {
			t.Errorf("progress = %d/%d, want 0/1 — only the task counts", got.Done, got.Total)
		}
		if !got.Defined {
			t.Error("Defined = false, want true")
		}
	})

	t.Run("finishing the only task is 100%", func(t *testing.T) {
		if _, err := f.tasks.MoveToColumn(ctx, task.ID, domain.StatusDone); err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
		got, err := f.tasks.Progress(ctx, p.ID)
		if err != nil {
			t.Fatalf("Progress: %v", err)
		}
		if got.Done != 1 || got.Total != 1 || got.Percent != 100 {
			t.Errorf("progress = %+v, want 1/1 at 100%%", got)
		}
	})
}

// A project whose only leaves are notes has no work in it: the answer is
// "undefined", not 0% and not 100% (D7).
func TestProgressIsUndefinedForAProjectOfNotes(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("reading list", domain.NodeTypeProject, nil))
	f.create(draft("a memo", domain.NodeTypeNote, &p.ID))
	f.create(draft("another memo", domain.NodeTypeNote, &p.ID))

	got, err := f.tasks.Progress(ctx, p.ID)
	if err != nil {
		t.Fatalf("Progress: %v", err)
	}
	if got.Defined {
		t.Errorf("Defined = true for %+v, want false", got)
	}
	if got.Total != 0 || got.Done != 0 {
		t.Errorf("progress = %d/%d, want 0/0", got.Done, got.Total)
	}

	t.Run("and the same on its view", func(t *testing.T) {
		views, err := f.tasks.Tree(ctx, &p.ID)
		if err != nil {
			t.Fatalf("Tree: %v", err)
		}
		if views[0].Progress.Defined {
			t.Errorf("the view reports Defined = true: %+v", views[0].Progress)
		}
	})

	t.Run("a node that does not exist", func(t *testing.T) {
		if _, err := f.tasks.Progress(ctx, "ghost"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("Progress(ghost) = %v, want domain.ErrNodeNotFound", err)
		}
	})
}

// Archived nodes are hidden from every read on this path.
func TestBoardAndTreeExcludeArchivedNodes(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	live := f.create(draft("live", domain.NodeTypeTask, nil))
	gone := f.create(draft("archived", domain.NodeTypeTask, nil))
	if _, err := f.tasks.ArchiveNode(ctx, gone.ID); err != nil {
		t.Fatalf("ArchiveNode: %v", err)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if v := find(board, gone.ID); v != nil {
		t.Error("an archived node is on the board")
	}
	if v := find(board, live.ID); v == nil {
		t.Error("the live node is missing from the board")
	}

	views, err := f.tasks.Tree(ctx, nil)
	if err != nil {
		t.Fatalf("Tree: %v", err)
	}
	if len(views) != 1 || views[0].Node.ID != live.ID {
		t.Errorf("the tree holds %v, want only the live node", titles(views))
	}
}

// D1's overdue, computed in Go and delivered as a field.
func TestOverdueIsComputedOnTheView(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name string
		due  domain.Date
		done bool
		want bool
	}{
		{"due yesterday and not done", domain.NewDate(2026, time.September, 17), false, true},
		{"due today", domain.NewDate(2026, time.September, 18), false, false},
		{"due tomorrow", domain.NewDate(2026, time.September, 19), false, false},
		{"due yesterday but done", domain.NewDate(2026, time.September, 17), true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			n := f.create(draft("x", domain.NodeTypeTask, nil))
			if _, err := f.tasks.SetDue(ctx, n.ID, &tc.due); err != nil {
				t.Fatalf("SetDue: %v", err)
			}
			if tc.done {
				if _, err := f.tasks.MoveToColumn(ctx, n.ID, domain.StatusDone); err != nil {
					t.Fatalf("MoveToColumn(done): %v", err)
				}
			}

			board, err := f.tasks.Board(ctx)
			if err != nil {
				t.Fatalf("Board: %v", err)
			}
			v := find(board, n.ID)
			if v == nil {
				t.Fatal("the node is not on the board")
			}
			if v.Overdue != tc.want {
				t.Errorf("Overdue = %v, want %v", v.Overdue, tc.want)
			}
		})
	}

	t.Run("a parent whose children are all done is not overdue", func(t *testing.T) {
		f := newFixture(t)
		p := f.create(draft("p", domain.NodeTypeProject, nil))
		child := f.create(draft("child", domain.NodeTypeTask, &p.ID))

		yesterday := domain.NewDate(2026, time.September, 17)
		if _, err := f.tasks.SetDue(ctx, p.ID, &yesterday); err != nil {
			t.Fatalf("SetDue: %v", err)
		}
		if _, err := f.tasks.MoveToColumn(ctx, child.ID, domain.StatusDone); err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}

		board, err := f.tasks.Board(ctx)
		if err != nil {
			t.Fatalf("Board: %v", err)
		}
		v := find(board, p.ID)
		if v == nil {
			t.Fatal("the parent is not on the board")
		}
		if v.Status != domain.StatusDone {
			t.Fatalf("the parent derives %q, want done", v.Status)
		}
		if v.Overdue {
			t.Error("the parent is reported overdue although it derives done")
		}
	})
}

// The board is ordered: the five columns in their own order, and the cards
// inside one column by sort_order.
func TestBoardOrder(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	first := f.create(draft("first", domain.NodeTypeTask, nil))
	second := f.create(draft("second", domain.NodeTypeTask, nil))
	third := f.create(draft("third", domain.NodeTypeTask, nil))
	for _, id := range []string{first.ID, second.ID, third.ID} {
		if _, err := f.tasks.MoveToColumn(ctx, id, domain.StatusToday); err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}
	}
	// Put the third one at the top of the range.
	if _, err := f.tasks.MoveNode(ctx, third.ID, nil, 0); err != nil {
		t.Fatalf("MoveNode: %v", err)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	t.Run("the columns are in D2's order", func(t *testing.T) {
		got := make([]domain.Status, len(board))
		for i, c := range board {
			got[i] = c.Status
		}
		want := domain.Statuses()
		if len(got) != len(want) {
			t.Fatalf("%d columns, want %d", len(got), len(want))
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("column %d = %q, want %q", i, got[i], want[i])
			}
		}
	})

	t.Run("cards follow sort_order", func(t *testing.T) {
		got := titles(column(t, board, domain.StatusToday))
		want := []string{"third", "first", "second"}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("today = %v, want %v", got, want)
		}
	})
}

// Tree nests the children, carries the derivations at every level and respects
// sort_order.
func TestTree(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("p", domain.NodeTypeProject, nil))
	mid := f.create(draft("mid", domain.NodeTypeProject, &p.ID))
	deep := f.create(draft("deep", domain.NodeTypeTask, &mid.ID))
	f.create(draft("memo", domain.NodeTypeNote, &p.ID))
	other := f.create(draft("another root", domain.NodeTypeTask, nil))

	if _, err := f.tasks.MoveToColumn(ctx, deep.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn: %v", err)
	}

	t.Run("the whole forest", func(t *testing.T) {
		views, err := f.tasks.Tree(ctx, nil)
		if err != nil {
			t.Fatalf("Tree: %v", err)
		}
		if got := titles(views); len(got) != 2 || got[0] != "p" || got[1] != "another root" {
			t.Fatalf("roots = %v, want [p another root]", got)
		}
		if got := titles(views[0].Children); len(got) != 2 || got[0] != "mid" || got[1] != "memo" {
			t.Errorf("p's children = %v, want [mid memo]", got)
		}
		if got := views[0].Children[0].Children; len(got) != 1 || got[0].Node.ID != deep.ID {
			t.Errorf("mid's children = %v, want [deep]", titles(got))
		}
		if views[0].Status != domain.StatusToday {
			t.Errorf("p derives %q, want today", views[0].Status)
		}
		if views[1].Node.ID != other.ID {
			t.Errorf("the second root is %q, want %q", views[1].Node.ID, other.ID)
		}
	})

	t.Run("one named subtree", func(t *testing.T) {
		views, err := f.tasks.Tree(ctx, &mid.ID)
		if err != nil {
			t.Fatalf("Tree: %v", err)
		}
		if len(views) != 1 || views[0].Node.ID != mid.ID {
			t.Fatalf("Tree(mid) = %v, want [mid]", titles(views))
		}
		if len(views[0].Children) != 1 || views[0].Children[0].Node.ID != deep.ID {
			t.Errorf("mid's children = %v, want [deep]", titles(views[0].Children))
		}
		if views[0].Status != domain.StatusToday {
			t.Errorf("mid derives %q, want today", views[0].Status)
		}
	})

	t.Run("a root that does not exist", func(t *testing.T) {
		if _, err := f.tasks.Tree(ctx, ptr("ghost")); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Fatalf("Tree(ghost) = %v, want domain.ErrNodeNotFound", err)
		}
	})
}

// The tags are resolved by Go and arrive on the card; the UI never joins.
func TestViewsCarryTheirTags(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	n := f.create(draft("x", domain.NodeTypeTask, nil))
	other := f.create(draft("y", domain.NodeTypeTask, nil))
	work := domain.Tag{ID: "t1", Name: "work", Color: "#8b5cf6"}
	urgent := domain.Tag{ID: "t2", Name: "urgent"}
	for _, tag := range []domain.Tag{work, urgent} {
		if err := f.tags.CreateTag(ctx, tag); err != nil {
			t.Fatalf("CreateTag: %v", err)
		}
	}
	if err := f.tags.AttachTag(ctx, n.ID, work.ID); err != nil {
		t.Fatalf("AttachTag: %v", err)
	}
	if err := f.tags.AttachTag(ctx, n.ID, urgent.ID); err != nil {
		t.Fatalf("AttachTag: %v", err)
	}

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	v := find(board, n.ID)
	if v == nil {
		t.Fatal("the node is not on the board")
	}
	if len(v.Tags) != 2 {
		t.Fatalf("tags = %+v, want both", v.Tags)
	}
	names := map[string]bool{}
	for _, tag := range v.Tags {
		names[tag.Name] = true
	}
	if !names["work"] || !names["urgent"] {
		t.Errorf("tags = %+v, want work and urgent", v.Tags)
	}
	if other := find(board, other.ID); other == nil || len(other.Tags) != 0 {
		t.Errorf("the untagged node carries %+v", other)
	}
}

// The running timer is an indication on the card: which node, and for how long,
// computed in Go against the injected clock.
func TestViewsCarryTheRunningTimer(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	a := f.create(draft("a", domain.NodeTypeTask, nil))
	b := f.create(draft("b", domain.NodeTypeTask, nil))

	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if find(board, a.ID).Timer.Running {
		t.Error("a card reports a running timer before anything started")
	}

	started, err := f.timers().Start(ctx, a.ID)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	f.now = testNow.Add(90 * time.Second)

	board, err = f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}

	got := find(board, a.ID).Timer
	if !got.Running {
		t.Fatal("the running card reports Running = false")
	}
	if got.EntryID != started.ID {
		t.Errorf("EntryID = %q, want %q", got.EntryID, started.ID)
	}
	if got.ElapsedSeconds != 90 {
		t.Errorf("ElapsedSeconds = %d, want 90", got.ElapsedSeconds)
	}
	if got.StartedAt == nil || !got.StartedAt.Equal(testNow) {
		t.Errorf("StartedAt = %v, want %v", got.StartedAt, testNow)
	}
	if find(board, b.ID).Timer.Running {
		t.Error("the other card reports a running timer too; exactly one may")
	}
}

// ---------------------------------------------------------------------------
// No N+1

// countingTx counts the statements a read issues, so that "no per-node queries"
// is a number in a failure message rather than a claim.
type countingTx struct {
	service.Tx
	queries *int
}

func (tx countingTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	*tx.queries++
	return tx.Tx.QueryContext(ctx, query, args...)
}

func (tx countingTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	*tx.queries++
	return tx.Tx.QueryRowContext(ctx, query, args...)
}

// countedTree reads the whole forest through a counting transaction and returns
// the number of statements it took.
func (f *fixture) countedTree(ctx context.Context, t *testing.T) int {
	t.Helper()

	count := 0
	svc := service.NewTaskService(
		wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return countingTx{Tx: tx, queries: &count}
		}},
		f.nodes, f.tags, f.clock(), f.nextID)

	if _, err := svc.Tree(ctx, nil); err != nil {
		t.Fatalf("Tree: %v", err)
	}
	return count
}

func TestTreeIsAssembledWithoutPerNodeQueries(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	parent := f.create(draft("parent", domain.NodeTypeProject, nil))
	for i := range 199 {
		f.create(draft(fmt.Sprintf("child %d", i), domain.NodeTypeTask, &parent.ID))
	}

	small := newFixture(t)
	smallParent := small.create(draft("parent", domain.NodeTypeProject, nil))
	small.create(draft("child", domain.NodeTypeTask, &smallParent.ID))

	big := f.countedTree(ctx, t)
	tiny := small.countedTree(ctx, t)

	if big != tiny {
		t.Errorf("a 200-node tree issued %d queries and a 2-node tree %d; the count must not depend on the node count",
			big, tiny)
	}
	if big > 6 {
		t.Errorf("a tree read issued %d queries, want a small constant", big)
	}
}

// A read that cannot reach the database says so, rather than returning an empty
// board that looks like an empty database.
func TestReadPathsSurfaceStoreFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	p := f.create(draft("p", domain.NodeTypeProject, nil))
	f.create(draft("child", domain.NodeTypeTask, &p.ID))

	beginners := map[string]service.Beginner{
		"the transaction cannot be started": blockedBeginner{},
		"the reads fail": wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return queryFailingTx{Tx: tx}
		}},
	}

	for name, begin := range beginners {
		t.Run(name, func(t *testing.T) {
			svc := service.NewTaskService(begin, f.nodes, f.tags, f.clock(), f.nextID)

			if _, err := svc.Tree(ctx, nil); err == nil {
				t.Error("Tree succeeded")
			}
			if _, err := svc.Tree(ctx, &p.ID); err == nil {
				t.Error("Tree(root) succeeded")
			}
			if _, err := svc.Board(ctx); err == nil {
				t.Error("Board succeeded")
			}
			if _, err := svc.Progress(ctx, p.ID); err == nil {
				t.Error("Progress succeeded")
			}
		})
	}
}

// The tag index is read once for the whole board, not once per card: adding
// cards must not add queries, and adding tags adds one each.
func TestTagIndexCostsQueriesPerTagNotPerNode(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	base := f.countedTree(ctx, t)

	for i := range 20 {
		f.create(draft(fmt.Sprintf("task %d", i), domain.NodeTypeTask, nil))
	}
	if got := f.countedTree(ctx, t); got != base {
		t.Errorf("twenty more cards cost %d queries, was %d", got, base)
	}

	if err := f.tags.CreateTag(ctx, domain.Tag{ID: "t1", Name: "work"}); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if got, want := f.countedTree(ctx, t), base+1; got != want {
		t.Errorf("one tag cost %d queries, want %d", got, want)
	}
}

// C2, the acceptance criterion of S2-01: assembling the board builds ONE
// derivation index, whatever the board holds.
//
// # Why the counter is in the domain and not around this call
//
// The regression this guards against is a view that goes back to
// domain.DeriveStatus(set, id) once per card — the O(n²·log n) shape Board() had
// until S2-01. A counter wrapped around the service's own call to
// domain.NewIndex would still read exactly 1 while five hundred indexes were
// built underneath it by the whole-slice entry points, so the count is taken
// from domain.IndexBuilds, which counts EVERY construction in the process. Put
// the per-node call back and this test reports 501.
//
// The card count is asserted as well: a board that derived nothing would also
// build one index, and "1" would then be a number about nothing.
func TestBoardBuildsTheDerivationIndexOnce(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	// 500 cards in 100 subtrees, so that the derivation really recurses rather
	// than reading 500 stored statuses off 500 roots.
	const (
		projects       = 100
		tasksPerParent = 4
		cards          = projects * (1 + tasksPerParent)
	)
	for p := range projects {
		parent := bareNode(fmt.Sprintf("p%03d", p), nil, domain.NodeTypeProject)
		if err := f.nodes.Create(ctx, parent); err != nil {
			t.Fatalf("Create(%q): %v", parent.ID, err)
		}
		for c := range tasksPerParent {
			child := bareNode(fmt.Sprintf("p%03d-t%d", p, c), &parent.ID, domain.NodeTypeTask)
			if err := f.nodes.Create(ctx, child); err != nil {
				t.Fatalf("Create(%q): %v", child.ID, err)
			}
		}
	}

	before := domain.IndexBuilds()
	board, err := f.tasks.Board(ctx)
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	built := domain.IndexBuilds() - before

	total := 0
	for _, c := range board {
		total += len(c.Nodes)
	}
	if total != cards {
		t.Fatalf("the board holds %d cards, want %d — the index count below would be about nothing",
			total, cards)
	}
	if built != 1 {
		t.Errorf("a %d-card board built %d derivation indexes, want 1 (C2: the index is built once "+
			"per Board() call, not once per card)", cards, built)
	}
}

// bareNode is a valid row written straight through the repository, for the
// tests that need hundreds of them: CreateNode re-reads the whole set per call,
// which is fine for a handful and quadratic for five hundred.
func bareNode(id string, parent *string, typ domain.NodeType) domain.Node {
	return domain.Node{
		ID:        id,
		ParentID:  parent,
		Type:      typ,
		Title:     id,
		Status:    domain.StatusBacklog,
		DueSource: domain.DueSourceAuto,
		Priority:  domain.Priority4,
		CreatedAt: testNow,
		UpdatedAt: testNow,
	}
}
