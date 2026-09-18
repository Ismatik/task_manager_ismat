package domain_test

import (
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
)

// cascadeNow is the injected clock for every cascade below.
var cascadeNow = func() time.Time { return fixedNow }

// changeByID finds the plan row for id, or nil.
func changeByID(plan []domain.StatusChange, id string) *domain.StatusChange {
	for i := range plan {
		if plan[i].NodeID == id {
			return &plan[i]
		}
	}
	return nil
}

func changeIDs(plan []domain.StatusChange) []string {
	out := make([]string, len(plan))
	for i, c := range plan {
		out[i] = c.NodeID
	}
	return out
}

// headlineTree is D2's headline case:
//
//	p (project)
//	├── mid (project)          <- an intermediate parent, unfinished
//	│   ├── deep (task, week)  <- unfinished, at depth 2
//	│   └── finished (task, done, completed a month ago)
//	├── shallow (task, backlog) <- unfinished, at depth 1
//	└── memo (note)            <- must never be touched
func headlineTree() []domain.Node {
	longAgo := fixedNow.AddDate(0, -1, 0)
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("mid", "p", domain.StatusBacklog),
		task("deep", "mid", domain.StatusWeek),
		task("finished", "mid", domain.StatusDone),
		task("shallow", "p", domain.StatusBacklog),
		note("memo", "p"),
	}
	for i := range nodes {
		if nodes[i].ID == "finished" {
			nodes[i].CompletedAt = &longAgo
		}
	}
	return nodes
}

// THE headline test of D2: drag a parent to Done and everything unfinished
// beneath it finishes, with completed_at, while the note and the already-done
// task are left alone — and the parent then DERIVES done.
func TestPlanCascadeParentToDone(t *testing.T) {
	nodes := headlineTree()

	plan, err := domain.PlanCascade(nodes, "p", domain.StatusDone, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}

	t.Run("every unfinished non-note descendant becomes done", func(t *testing.T) {
		for _, id := range []string{"mid", "deep", "shallow"} {
			c := changeByID(plan, id)
			if c == nil {
				t.Fatalf("no change for %q", id)
			}
			if c.Status != domain.StatusDone {
				t.Errorf("%q: Status = %q, want done", id, c.Status)
			}
		}
	})

	t.Run("each gets completed_at set to the injected now", func(t *testing.T) {
		for _, id := range []string{"mid", "deep", "shallow"} {
			c := changeByID(plan, id)
			if c.CompletedAt == nil {
				t.Fatalf("%q: CompletedAt = nil, want the injected now", id)
			}
			if !c.CompletedAt.Equal(fixedNow) {
				t.Errorf("%q: CompletedAt = %v, want %v", id, c.CompletedAt, fixedNow)
			}
		}
	})

	t.Run("the already-done descendant is untouched", func(t *testing.T) {
		if c := changeByID(plan, "finished"); c != nil {
			t.Fatalf("the plan touches the already-done %q: %+v", c.NodeID, c)
		}
		// Its original completed_at survives into the applied tree.
		applied, err := domain.ApplyStatusChanges(nodes, plan)
		if err != nil {
			t.Fatalf("ApplyStatusChanges() = %v", err)
		}
		for _, n := range applied {
			if n.ID != "finished" {
				continue
			}
			want := fixedNow.AddDate(0, -1, 0)
			if n.CompletedAt == nil || !n.CompletedAt.Equal(want) {
				t.Errorf("finished.CompletedAt = %v, want the original %v", n.CompletedAt, want)
			}
		}
	})

	t.Run("the note is untouched", func(t *testing.T) {
		if c := changeByID(plan, "memo"); c != nil {
			t.Errorf("the plan touches the note %q: %+v", c.NodeID, c)
		}
	})

	t.Run("the plan contains no row for the dragged parent itself", func(t *testing.T) {
		if c := changeByID(plan, "p"); c != nil {
			t.Errorf("the plan writes a status onto the parent: %+v — a parent's status is derived", c)
		}
	})

	t.Run("the parent then derives done", func(t *testing.T) {
		applied, err := domain.ApplyStatusChanges(nodes, plan)
		if err != nil {
			t.Fatalf("ApplyStatusChanges() = %v", err)
		}
		got, err := domain.DeriveStatus(applied, "p")
		if err != nil {
			t.Fatalf("DeriveStatus() = %v", err)
		}
		if got != domain.StatusDone {
			t.Errorf("DeriveStatus(p) = %q after the cascade, want done", got)
		}
		// And so does the intermediate parent.
		got, err = domain.DeriveStatus(applied, "mid")
		if err != nil {
			t.Fatalf("DeriveStatus() = %v", err)
		}
		if got != domain.StatusDone {
			t.Errorf("DeriveStatus(mid) = %q after the cascade, want done", got)
		}
	})
}

// Dragging a parent to Today cascades today and sets no completed_at.
func TestPlanCascadeParentToToday(t *testing.T) {
	nodes := headlineTree()

	plan, err := domain.PlanCascade(nodes, "p", domain.StatusToday, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}

	want := []string{"mid", "deep", "shallow"}
	if got := changeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v", got, want)
	}
	for _, c := range plan {
		if c.Status != domain.StatusToday {
			t.Errorf("%q: Status = %q, want today", c.NodeID, c.Status)
		}
		if c.CompletedAt != nil {
			t.Errorf("%q: CompletedAt = %v, want nil — nothing was completed", c.NodeID, c.CompletedAt)
		}
	}

	applied, err := domain.ApplyStatusChanges(nodes, plan)
	if err != nil {
		t.Fatalf("ApplyStatusChanges() = %v", err)
	}
	got, err := domain.DeriveStatus(applied, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if got != domain.StatusToday {
		t.Errorf("DeriveStatus(p) = %q after the cascade, want today", got)
	}
}

// Only leaves start a timer, so a drag to Doing touches only leaves and no
// intermediate parent gets a stored status.
func TestPlanCascadeToDoingTouchesOnlyLeaves(t *testing.T) {
	nodes := headlineTree()

	plan, err := domain.PlanCascade(nodes, "p", domain.StatusDoing, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}

	want := []string{"deep", "shallow"}
	if got := changeIDs(plan); !equalStrings(got, want) {
		t.Fatalf("plan = %v, want %v — only leaves", got, want)
	}
	if c := changeByID(plan, "mid"); c != nil {
		t.Errorf("the intermediate parent %q got a stored status: %+v", c.NodeID, c)
	}
	if c := changeByID(plan, "p"); c != nil {
		t.Errorf("the dragged parent got a stored status: %+v", c)
	}

	// The parents render in Doing all the same, derived from the leaves.
	applied, err := domain.ApplyStatusChanges(nodes, plan)
	if err != nil {
		t.Fatalf("ApplyStatusChanges() = %v", err)
	}
	for _, id := range []string{"p", "mid"} {
		got, err := domain.DeriveStatus(applied, id)
		if err != nil {
			t.Fatalf("DeriveStatus(%q) = %v", id, err)
		}
		if got != domain.StatusDoing {
			t.Errorf("DeriveStatus(%q) = %q, want doing (derived from its leaves)", id, got)
		}
	}
}

// A leaf dragged anywhere produces exactly one change: itself.
func TestPlanCascadeOfALeafIsExactlyOneChange(t *testing.T) {
	nodes := headlineTree()

	for _, target := range domain.Statuses() {
		t.Run("to "+target.String(), func(t *testing.T) {
			plan, err := domain.PlanCascade(nodes, "shallow", target, cascadeNow)
			if err != nil {
				t.Fatalf("PlanCascade() = %v", err)
			}
			if len(plan) != 1 || plan[0].NodeID != "shallow" {
				t.Fatalf("plan = %v, want exactly [shallow]", changeIDs(plan))
			}
			if plan[0].Status != target {
				t.Errorf("Status = %q, want %q", plan[0].Status, target)
			}
			wantCompleted := target == domain.StatusDone
			if got := plan[0].CompletedAt != nil; got != wantCompleted {
				t.Errorf("CompletedAt set = %v, want %v", got, wantCompleted)
			}
		})
	}
}

// A node whose children are all notes is a leaf: it gets a row of its own.
func TestPlanCascadeTreatsAnAllNotesParentAsALeaf(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		note("n1", "p"),
		note("n2", "p"),
	}

	plan, err := domain.PlanCascade(nodes, "p", domain.StatusToday, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}
	if len(plan) != 1 || plan[0].NodeID != "p" {
		t.Fatalf("plan = %v, want exactly [p] — an all-notes parent is a leaf", changeIDs(plan))
	}
	if plan[0].Status != domain.StatusToday {
		t.Errorf("Status = %q, want today", plan[0].Status)
	}

	t.Run("and the notes below it are never written", func(t *testing.T) {
		for _, id := range []string{"n1", "n2"} {
			if c := changeByID(plan, id); c != nil {
				t.Errorf("the plan touches the note %q: %+v", id, c)
			}
		}
	})
}

// Moving out of done clears completed_at. Only the drag target can leave done,
// because done descendants are skipped.
func TestPlanCascadeOutOfDoneClearsCompletedAt(t *testing.T) {
	completed := fixedNow.AddDate(0, 0, -3)
	leaf := task("leaf", "", domain.StatusDone)
	leaf.CompletedAt = &completed
	nodes := []domain.Node{leaf}

	tests := []struct {
		name            string
		target          domain.Status
		wantCompletedAt *time.Time
	}{
		{"back to today", domain.StatusToday, nil},
		{"back to backlog", domain.StatusBacklog, nil},
		{"back to week", domain.StatusWeek, nil},
		{"back to doing", domain.StatusDoing, nil},
		{"re-dropped onto done keeps the original timestamp", domain.StatusDone, &completed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plan, err := domain.PlanCascade(nodes, "leaf", tt.target, cascadeNow)
			if err != nil {
				t.Fatalf("PlanCascade() = %v", err)
			}
			if len(plan) != 1 {
				t.Fatalf("plan = %v, want one change", changeIDs(plan))
			}
			switch {
			case tt.wantCompletedAt == nil && plan[0].CompletedAt != nil:
				t.Errorf("CompletedAt = %v, want nil", plan[0].CompletedAt)
			case tt.wantCompletedAt != nil && plan[0].CompletedAt == nil:
				t.Errorf("CompletedAt = nil, want %v", *tt.wantCompletedAt)
			case tt.wantCompletedAt != nil && !plan[0].CompletedAt.Equal(*tt.wantCompletedAt):
				t.Errorf("CompletedAt = %v, want %v", plan[0].CompletedAt, *tt.wantCompletedAt)
			}
		})
	}
}

// A done leaf with no completed_at at all still gets one when dropped on Done.
func TestPlanCascadeStampsADoneLeafThatHasNoCompletedAt(t *testing.T) {
	nodes := []domain.Node{task("leaf", "", domain.StatusDone)}

	plan, err := domain.PlanCascade(nodes, "leaf", domain.StatusDone, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}
	if plan[0].CompletedAt == nil || !plan[0].CompletedAt.Equal(fixedNow) {
		t.Errorf("CompletedAt = %v, want the injected now", plan[0].CompletedAt)
	}
}

// Dragging a whole subtree out of done: the descendants that are done stay
// done, because a finished task is not un-finished by its parent moving.
func TestPlanCascadeNeverUnfinishesADoneDescendant(t *testing.T) {
	nodes := headlineTree()

	for _, target := range []domain.Status{
		domain.StatusBacklog, domain.StatusWeek, domain.StatusToday, domain.StatusDoing,
	} {
		t.Run("to "+target.String(), func(t *testing.T) {
			plan, err := domain.PlanCascade(nodes, "p", target, cascadeNow)
			if err != nil {
				t.Fatalf("PlanCascade() = %v", err)
			}
			if c := changeByID(plan, "finished"); c != nil {
				t.Errorf("the plan un-finishes %q: %+v", c.NodeID, c)
			}
		})
	}
}

func TestPlanCascadeNotes(t *testing.T) {
	nodes := headlineTree()

	t.Run("a note dragged anywhere produces no changes", func(t *testing.T) {
		plan, err := domain.PlanCascade(nodes, "memo", domain.StatusToday, cascadeNow)
		if err != nil {
			t.Fatalf("PlanCascade() = %v", err)
		}
		if len(plan) != 0 {
			t.Errorf("plan = %v, want empty — notes have no column", changeIDs(plan))
		}
	})
}

func TestPlanCascadeErrors(t *testing.T) {
	nodes := headlineTree()

	t.Run("an unknown node", func(t *testing.T) {
		if _, err := domain.PlanCascade(nodes, "nope", domain.StatusToday, cascadeNow); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})
	t.Run("an unknown target column", func(t *testing.T) {
		_, err := domain.PlanCascade(nodes, "p", domain.Status("Done"), cascadeNow)
		if !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("err = %v, want ErrInvalid", err)
		}
	})
	t.Run("a cyclic tree terminates with ErrCycle", func(t *testing.T) {
		cyclic := []domain.Node{
			project("x", "y", domain.StatusBacklog),
			project("y", "x", domain.StatusBacklog),
		}
		if _, err := domain.PlanCascade(cyclic, "x", domain.StatusToday, cascadeNow); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})
}

// The plan is a plan: same input, same order, every time.
func TestPlanCascadeIsDeterministic(t *testing.T) {
	nodes := headlineTree()

	for _, target := range domain.Statuses() {
		t.Run("to "+target.String(), func(t *testing.T) {
			first, err := domain.PlanCascade(nodes, "p", target, cascadeNow)
			if err != nil {
				t.Fatalf("PlanCascade() = %v", err)
			}
			for range 3 {
				again, err := domain.PlanCascade(nodes, "p", target, cascadeNow)
				if err != nil {
					t.Fatalf("PlanCascade() = %v", err)
				}
				if !equalStrings(changeIDs(first), changeIDs(again)) {
					t.Fatalf("plan order changed: %v vs %v", changeIDs(first), changeIDs(again))
				}
			}
		})
	}
}

// Every node finished by one drag shares one timestamp: the clock is read once.
func TestPlanCascadeReadsTheClockOnce(t *testing.T) {
	calls := 0
	now := func() time.Time {
		calls++
		return fixedNow.Add(time.Duration(calls) * time.Second)
	}

	plan, err := domain.PlanCascade(headlineTree(), "p", domain.StatusDone, now)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}
	if calls != 1 {
		t.Errorf("the clock was read %d times, want 1", calls)
	}
	for _, c := range plan {
		if !c.CompletedAt.Equal(*plan[0].CompletedAt) {
			t.Errorf("%q: CompletedAt = %v, want the same instant as the rest of the plan (%v)",
				c.NodeID, c.CompletedAt, plan[0].CompletedAt)
		}
	}
}

// Each change carries its own timestamp, so applying one cannot alter another.
func TestPlanCascadeDoesNotShareOneTimestampPointer(t *testing.T) {
	plan, err := domain.PlanCascade(headlineTree(), "p", domain.StatusDone, cascadeNow)
	if err != nil {
		t.Fatalf("PlanCascade() = %v", err)
	}
	if len(plan) < 2 {
		t.Fatalf("expected several changes, got %d", len(plan))
	}
	if plan[0].CompletedAt == plan[1].CompletedAt {
		t.Fatal("two changes share one *time.Time")
	}
}

func TestApplyStatusChanges(t *testing.T) {
	nodes := headlineTree()

	t.Run("the argument is not mutated", func(t *testing.T) {
		plan, err := domain.PlanCascade(nodes, "p", domain.StatusDone, cascadeNow)
		if err != nil {
			t.Fatalf("PlanCascade() = %v", err)
		}
		if _, err := domain.ApplyStatusChanges(nodes, plan); err != nil {
			t.Fatalf("ApplyStatusChanges() = %v", err)
		}
		for _, n := range nodes {
			if n.ID == "shallow" && n.Status != domain.StatusBacklog {
				t.Errorf("ApplyStatusChanges mutated its argument: shallow is now %q", n.Status)
			}
		}
	})

	t.Run("a change for a node outside the set is an error", func(t *testing.T) {
		_, err := domain.ApplyStatusChanges(nodes, []domain.StatusChange{
			{NodeID: "ghost", Status: domain.StatusDone},
		})
		if !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	t.Run("the applied completed_at does not alias the plan", func(t *testing.T) {
		plan, err := domain.PlanCascade(nodes, "p", domain.StatusDone, cascadeNow)
		if err != nil {
			t.Fatalf("PlanCascade() = %v", err)
		}
		applied, err := domain.ApplyStatusChanges(nodes, plan)
		if err != nil {
			t.Fatalf("ApplyStatusChanges() = %v", err)
		}
		for i := range applied {
			if applied[i].ID != plan[0].NodeID {
				continue
			}
			if applied[i].CompletedAt == plan[0].CompletedAt {
				t.Error("the applied node shares the plan's *time.Time")
			}
		}
	})

	t.Run("clearing completed_at", func(t *testing.T) {
		applied, err := domain.ApplyStatusChanges(nodes, []domain.StatusChange{
			{NodeID: "finished", Status: domain.StatusToday, CompletedAt: nil},
		})
		if err != nil {
			t.Fatalf("ApplyStatusChanges() = %v", err)
		}
		for _, n := range applied {
			if n.ID == "finished" && n.CompletedAt != nil {
				t.Errorf("CompletedAt = %v, want nil", n.CompletedAt)
			}
		}
	})
}
