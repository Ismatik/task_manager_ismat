package domain_test

import (
	"errors"
	"testing"

	"nexus/internal/domain"
)

// nd builds a node for a tree literal. parent is "" for a root.
func nd(id, parent string, typ domain.NodeType, status domain.Status) domain.Node {
	n := domain.Node{
		ID:        id,
		Type:      typ,
		Title:     id,
		Status:    status,
		DueSource: domain.DueSourceManual,
		Priority:  domain.Priority4,
		CreatedAt: fixedNow,
		UpdatedAt: fixedNow,
	}
	if parent != "" {
		n.ParentID = ptr(parent)
	}
	return n
}

// task, note and project are shorthands for the tree literals below.
func task(id, parent string, status domain.Status) domain.Node {
	return nd(id, parent, domain.NodeTypeTask, status)
}

func note(id, parent string) domain.Node {
	return nd(id, parent, domain.NodeTypeNote, domain.StatusBacklog)
}

// habit carries the inert backlog its NOT NULL column needs — PLAN.md §4 gives
// it no column at all — and the recurrence a real habit row always has.
func habit(id, parent string) domain.Node {
	n := nd(id, parent, domain.NodeTypeHabit, domain.StatusBacklog)
	n.Recurrence = ptr("FREQ=DAILY")
	return n
}

func project(id, parent string, status domain.Status) domain.Node {
	return nd(id, parent, domain.NodeTypeProject, status)
}

// D2: a parent's status is the least-advanced status among its non-done
// children, done only when all of them are done, with notes excluded.
func TestDeriveStatus(t *testing.T) {
	tests := []struct {
		name  string
		nodes []domain.Node
		id    string
		want  domain.Status
	}{
		{
			name: "all children backlog",
			nodes: []domain.Node{
				project("p", "", domain.StatusDone),
				task("a", "p", domain.StatusBacklog),
				task("b", "p", domain.StatusBacklog),
			},
			id: "p", want: domain.StatusBacklog,
		},
		{
			name: "backlog and today derive backlog, least advanced wins",
			nodes: []domain.Node{
				project("p", "", domain.StatusDone),
				task("a", "p", domain.StatusToday),
				task("b", "p", domain.StatusBacklog),
			},
			id: "p", want: domain.StatusBacklog,
		},
		{
			name: "done and week derive week, done children are skipped",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusWeek),
			},
			id: "p", want: domain.StatusWeek,
		},
		{
			name: "all children done derives done",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusDone),
			},
			id: "p", want: domain.StatusDone,
		},
		{
			name: "a done child and a note derive done, the note does not hold it back",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				note("n", "p"),
			},
			id: "p", want: domain.StatusDone,
		},
		{
			name: "only note children makes it a leaf reporting its own stored status",
			nodes: []domain.Node{
				project("p", "", domain.StatusToday),
				note("n1", "p"),
				note("n2", "p"),
			},
			id: "p", want: domain.StatusToday,
		},
		{
			name: "no children makes it a leaf reporting its own stored status",
			nodes: []domain.Node{
				task("p", "", domain.StatusDoing),
			},
			id: "p", want: domain.StatusDoing,
		},
		{
			name: "the least advanced of a full spread is backlog",
			nodes: []domain.Node{
				project("p", "", domain.StatusDone),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusDoing),
				task("c", "p", domain.StatusToday),
				task("d", "p", domain.StatusWeek),
				task("e", "p", domain.StatusBacklog),
			},
			id: "p", want: domain.StatusBacklog,
		},
		{
			name: "three levels deep, the grandparent derives from derived values",
			nodes: []domain.Node{
				// g -> p1 -> {done, done}   derives done
				// g -> p2 -> {done, week}   derives week
				// so g derives week, not backlog: p1 and p2's OWN stored
				// statuses are backlog and must be ignored entirely.
				project("g", "", domain.StatusDone),
				project("p1", "g", domain.StatusBacklog),
				task("p1a", "p1", domain.StatusDone),
				task("p1b", "p1", domain.StatusDone),
				project("p2", "g", domain.StatusBacklog),
				task("p2a", "p2", domain.StatusDone),
				task("p2b", "p2", domain.StatusWeek),
			},
			id: "g", want: domain.StatusWeek,
		},
		{
			name: "three levels deep, every leaf done makes the grandparent done",
			nodes: []domain.Node{
				project("g", "", domain.StatusBacklog),
				project("p1", "g", domain.StatusBacklog),
				task("p1a", "p1", domain.StatusDone),
				project("p2", "g", domain.StatusBacklog),
				task("p2a", "p2", domain.StatusDone),
				note("n", "p2"),
			},
			id: "g", want: domain.StatusDone,
		},
		{
			name: "a nested all-notes parent is a leaf and contributes its stored status",
			nodes: []domain.Node{
				project("g", "", domain.StatusDone),
				project("p", "g", domain.StatusWeek),
				note("n", "p"),
				task("t", "g", domain.StatusDone),
			},
			id: "g", want: domain.StatusWeek,
		},
		{
			// D10, the motivating case. The habit sits inert at backlog and can
			// never leave it, so before D10 this project rendered in Backlog
			// while its progress bar read 100%.
			name: "a done child and a habit derive done, the habit does not hold it back",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				habit("h", "p"),
			},
			id: "p", want: domain.StatusDone,
		},
		{
			// The mirror of the all-notes row above: a parent with nothing but
			// habits under it is a leaf and reports its own stored status, not
			// the backlog its habits are parked at.
			name: "only habit children makes it a leaf reporting its own stored status",
			nodes: []domain.Node{
				project("p", "", domain.StatusToday),
				habit("h1", "p"),
				habit("h2", "p"),
			},
			id: "p", want: domain.StatusToday,
		},
		{
			name: "mixed notes and habits with one real task derive from the task alone",
			nodes: []domain.Node{
				project("p", "", domain.StatusDone),
				note("n1", "p"),
				habit("h1", "p"),
				note("n2", "p"),
				habit("h2", "p"),
				task("t", "p", domain.StatusWeek),
			},
			id: "p", want: domain.StatusWeek,
		},
		{
			// Depth changes nothing: a habit is skipped as a child of a child,
			// so the grandparent derives from the real work only.
			name: "a habit grandchild does not touch the grandparent",
			nodes: []domain.Node{
				project("g", "", domain.StatusBacklog),
				project("p", "g", domain.StatusBacklog),
				task("pa", "p", domain.StatusDone),
				habit("ph", "p"),
				task("t", "g", domain.StatusDone),
			},
			id: "g", want: domain.StatusDone,
		},
		{
			// The habit is skipped whole: the task parked underneath it is off
			// the board too, so it cannot hold the project back either.
			name: "a task under a habit is not reached, the habit subtree is cut",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				habit("h", "p"),
				task("buried", "h", domain.StatusBacklog),
			},
			id: "p", want: domain.StatusDone,
		},
		{
			// An all-habits parent nested under a grandparent behaves exactly as
			// the all-notes one does: it is a leaf and contributes its own
			// stored status.
			name: "a nested all-habits parent is a leaf and contributes its stored status",
			nodes: []domain.Node{
				project("g", "", domain.StatusDone),
				project("p", "g", domain.StatusWeek),
				habit("h", "p"),
				task("t", "g", domain.StatusDone),
			},
			id: "g", want: domain.StatusWeek,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.DeriveStatus(tt.nodes, tt.id)
			if err != nil {
				t.Fatalf("DeriveStatus() = %v", err)
			}
			if got != tt.want {
				t.Errorf("DeriveStatus(%q) = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

// D10's headline: the derived column and the progress bar agree about a project
// holding one done task and one habit.
//
// This is the defect the decision was taken for. DeriveStatus excluded the note
// type by name, so the habit — which has no column, sits at the backlog its NOT
// NULL column needs and can never become done — was scanned like real work and
// held the project at backlog for ever. ComputeProgress had already generalised
// the rule to NodeType.HasColumn, so the same project reported 100%. The card
// sat in Backlog with a full bar on it.
//
// Both halves are asserted together on purpose: either one alone would pass
// again if the two rules drifted apart a second time.
func TestDeriveStatusAndProgressAgreeAboutAHabitSibling(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		task("t", "p", domain.StatusDone),
		habit("h", "p"),
	}

	status, err := domain.DeriveStatus(nodes, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if status != domain.StatusDone {
		t.Errorf("DeriveStatus(p) = %q, want %q: every child with a column is done and "+
			"a habit has none", status, domain.StatusDone)
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 1 || got.Total != 1 {
		t.Fatalf("ComputeProgress(p) = %d/%d, want 1/1", got.Done, got.Total)
	}
	if !got.Defined() {
		t.Errorf("Defined() = false, want true: there is one unit of work in this project")
	}
	if got.Percent() != 100 {
		t.Errorf("Percent() = %d, want 100", got.Percent())
	}

	if (status == domain.StatusDone) != (got.Percent() == 100) {
		t.Errorf("the column (%q) and the bar (%d%%) disagree", status, got.Percent())
	}
}

// The progress walk cuts a no-column subtree off entirely, which is the cut
// DeriveStatus makes over the same predicate (D10).
//
// Before D10 the walk stopped at a note only, so a task parked under a habit was
// counted in the denominator while the derived column ignored it — the same
// column-versus-bar disagreement one level down. A habit cannot hold children
// that matter to the board: everything under it is off the board too.
func TestComputeProgressCutsOffAHabitSubtree(t *testing.T) {
	// p ├── t (task, done)
	//   └── h (habit)
	//       └── buried (task, week)  <- off the board, so out of the bar
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		task("t", "p", domain.StatusDone),
		habit("h", "p"),
		task("buried", "h", domain.StatusWeek),
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 1 || got.Total != 1 {
		t.Fatalf("ComputeProgress(p) = %d/%d, want 1/1 — the habit subtree is not descended into",
			got.Done, got.Total)
	}

	status, err := domain.DeriveStatus(nodes, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if (status == domain.StatusDone) != (got.Percent() == 100) {
		t.Errorf("the column (%q) and the bar (%d%%) disagree about a task under a habit",
			status, got.Percent())
	}

	t.Run("a habit asked about itself is undefined, like a note", func(t *testing.T) {
		got, err := domain.ComputeProgress(nodes, "h")
		if err != nil {
			t.Fatalf("ComputeProgress(h) = %v", err)
		}
		if got.Defined() {
			t.Errorf("Defined() = true (%d/%d), want false — a habit has no column to finish in",
				got.Done, got.Total)
		}
	})
}

// "Parents never enter doing" (D2) is about starting a timer, not about
// rendering. DeriveStatus returns doing for a parent whose least-advanced
// unfinished child is being worked on: that is where the work is, so that is
// the column the parent renders in. The value is a render value and is never
// written back.
func TestDeriveStatusReturnsDoingAsARenderValue(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		task("a", "p", domain.StatusDone),
		task("b", "p", domain.StatusDoing),
	}

	got, err := domain.DeriveStatus(nodes, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if got != domain.StatusDoing {
		t.Errorf("DeriveStatus(parent) = %q, want %q: a parent renders in Doing when its "+
			"least-advanced unfinished child is doing", got, domain.StatusDoing)
	}
}

// A parent's own status column is meaningless: changing it must not change what
// the parent derives.
func TestDeriveStatusIgnoresTheParentsOwnStoredStatus(t *testing.T) {
	for _, stored := range domain.Statuses() {
		t.Run("stored "+stored.String(), func(t *testing.T) {
			nodes := []domain.Node{
				project("p", "", stored),
				task("a", "p", domain.StatusWeek),
			}
			got, err := domain.DeriveStatus(nodes, "p")
			if err != nil {
				t.Fatalf("DeriveStatus() = %v", err)
			}
			if got != domain.StatusWeek {
				t.Errorf("DeriveStatus() = %q, want week regardless of the stored %q", got, stored)
			}
		})
	}
}

func TestDeriveStatusErrors(t *testing.T) {
	t.Run("an unknown id", func(t *testing.T) {
		_, err := domain.DeriveStatus([]domain.Node{task("a", "", domain.StatusBacklog)}, "nope")
		if !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	t.Run("a parent that is not in the loaded set", func(t *testing.T) {
		// A partially loaded subtree: the child is present, its parent is not.
		// Asking about the child still works; asking about the absent parent
		// is ErrNodeNotFound rather than a silent empty answer.
		nodes := []domain.Node{task("orphan", "ghost", domain.StatusWeek)}
		if _, err := domain.DeriveStatus(nodes, "ghost"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("DeriveStatus(absent parent) = %v, want ErrNodeNotFound", err)
		}
		got, err := domain.DeriveStatus(nodes, "orphan")
		if err != nil {
			t.Fatalf("DeriveStatus(orphan) = %v", err)
		}
		if got != domain.StatusWeek {
			t.Errorf("DeriveStatus(orphan) = %q, want week", got)
		}
	})

	t.Run("a leaf with a corrupt status", func(t *testing.T) {
		nodes := []domain.Node{
			project("p", "", domain.StatusBacklog),
			task("a", "p", domain.Status("Done")),
		}
		_, err := domain.DeriveStatus(nodes, "p")
		if !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("err = %v, want ErrInvalid", err)
		}
	})

	// Corrupt data, not a user action: a two-node cycle must return rather than
	// recurse forever. No sleeps, no deadline - if this regresses the test hangs
	// and `go test` kills the package.
	t.Run("a cyclic tree terminates with ErrCycle", func(t *testing.T) {
		a := task("a", "b", domain.StatusWeek)
		b := task("b", "a", domain.StatusWeek)
		if _, err := domain.DeriveStatus([]domain.Node{a, b}, "a"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})

	t.Run("a self-parenting node terminates with ErrCycle", func(t *testing.T) {
		a := task("a", "a", domain.StatusWeek)
		if _, err := domain.DeriveStatus([]domain.Node{a}, "a"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})
}

// D7: done leaves over total leaves, notes excluded from both.
func TestComputeProgress(t *testing.T) {
	tests := []struct {
		name        string
		nodes       []domain.Node
		id          string
		wantDone    int
		wantTotal   int
		wantDefined bool
	}{
		{
			name: "none of three",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusBacklog),
				task("b", "p", domain.StatusWeek),
				task("c", "p", domain.StatusDoing),
			},
			id: "p", wantDone: 0, wantTotal: 3, wantDefined: true,
		},
		{
			name: "two of three",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusDone),
				task("c", "p", domain.StatusToday),
			},
			id: "p", wantDone: 2, wantTotal: 3, wantDefined: true,
		},
		{
			name: "three of three",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusDone),
				task("c", "p", domain.StatusDone),
			},
			id: "p", wantDone: 3, wantTotal: 3, wantDefined: true,
		},
		{
			name: "notes are excluded from the numerator and the denominator",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("a", "p", domain.StatusDone),
				task("b", "p", domain.StatusWeek),
				note("n1", "p"),
				note("n2", "p"),
			},
			id: "p", wantDone: 1, wantTotal: 2, wantDefined: true,
		},
		{
			name: "a nested subtree counts only its leaves",
			nodes: []domain.Node{
				// p has two sub-projects; neither of them is a unit of work.
				// D11 does not change this: a project is counted as a unit only
				// when it is a LEAF, and both of these have children with
				// columns, so their children are what the bar measures.
				project("p", "", domain.StatusBacklog),
				project("p1", "p", domain.StatusBacklog),
				task("p1a", "p1", domain.StatusDone),
				task("p1b", "p1", domain.StatusDone),
				project("p2", "p", domain.StatusBacklog),
				task("p2a", "p2", domain.StatusWeek),
				note("p2n", "p2"),
			},
			id: "p", wantDone: 2, wantTotal: 3, wantDefined: true,
		},
		{
			name: "a TASK whose children are all notes is itself the leaf",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				task("t1", "p", domain.StatusDone),
				note("n", "t1"),
			},
			id: "p", wantDone: 1, wantTotal: 1, wantDefined: true,
		},
		{
			// CHANGED BY D11. This case read 0/0 undefined, on the D9 amendment's
			// reasoning that "a project is never a unit of work" — so p1, a leaf
			// project, was skipped and p came out with nothing in it at all.
			//
			// That expectation was wrong about the question being asked. Nobody is
			// drawing a bar for p1 here; we are drawing one for p, and p1 is real
			// work under it that has simply not been broken down. Asserting 0/0
			// made p claim to contain no work while holding a finished sub-project
			// — the bar could not even show that something below it was done.
			// Under D11 p1 is one work leaf, done because its own stored status is
			// done. p1's OWN progress is still undefined; that is the other
			// question and TestEmptyProjectCountsInItsParent pins it.
			name: "a PROJECT whose children are all notes counts as one work leaf in its parent",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("p1", "p", domain.StatusDone),
				note("n", "p1"),
			},
			id: "p", wantDone: 1, wantTotal: 1, wantDefined: true,
		},
		{
			// UNCHANGED BY D11, and the regression guard for the first of D11's two
			// questions: p is the node being MEASURED, and a project with no work
			// beneath it draws no bar rather than an honest-looking 0%.
			name: "an empty project asked about itself is undefined, not zero per cent",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
			},
			id: "p", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
		{
			// CHANGED BY D11 — the motivating case. This read 1/1 at 100%: the
			// empty sub-project was skipped entirely, so a project that plainly
			// still had work in it reported itself finished, while DeriveStatus
			// put the same card in Backlog. The column and the bar contradicted
			// each other on one card, which is the defect D11 was decided on.
			// An empty project is unfinished work, so it is in the denominator:
			// 1 of 2, and both halves now say "not done".
			name: "an empty project among real tasks is one unfinished unit",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("empty", "p", domain.StatusBacklog),
				task("t", "p", domain.StatusDone),
			},
			id: "p", wantDone: 1, wantTotal: 2, wantDefined: true,
		},
		{
			name: "a leaf asked about itself",
			nodes: []domain.Node{
				task("a", "", domain.StatusDone),
			},
			id: "a", wantDone: 1, wantTotal: 1, wantDefined: true,
		},
		{
			name: "zero non-note leaves is undefined, not zero per cent",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				note("n1", "p"),
				note("n2", "p"),
			},
			// UNCHANGED BY D11. p is a leaf here (all children are notes) and p
			// is the node being MEASURED, so it is NOT counted: there is no work
			// inside this subtree to draw a bar from. The same p one level down,
			// under a parent, IS one unit — see the all-notes case above.
			id: "p", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
		{
			name: "a note asked about itself is undefined",
			nodes: []domain.Node{
				note("n", ""),
			},
			id: "n", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
		{
			name: "a project containing only notes below a sub-project is undefined",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				note("n1", "p"),
				note("n2", "p"),
				note("n3", "n1"),
			},
			// Descending into p reaches only notes, and a note's subtree is not
			// descended into at all.
			id: "n1", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ComputeProgress(tt.nodes, tt.id)
			if err != nil {
				t.Fatalf("ComputeProgress() = %v", err)
			}
			if got.Done != tt.wantDone || got.Total != tt.wantTotal {
				t.Errorf("ComputeProgress(%q) = %d/%d, want %d/%d",
					tt.id, got.Done, got.Total, tt.wantDone, tt.wantTotal)
			}
			if got.Defined() != tt.wantDefined {
				t.Errorf("Defined() = %v, want %v", got.Defined(), tt.wantDefined)
			}
		})
	}
}

// D11, the motivating case: project{empty sub-project, task:done} derives
// backlog and its bar reads 1 of 2.
//
// Before D11 the derivation and the bar disagreed on this one card. The empty
// sub-project has a column and is not done, so DeriveStatus held the project at
// backlog — correctly — while ComputeProgress skipped the sub-project entirely
// as "not a unit of work" and reported 1/1 at 100%: a card sitting in Backlog
// with a full bar on it, the same defect shape D10 was decided on.
//
// The user's decision is that an empty project IS unfinished work: it is work
// that has not been broken down yet, not an absence of work. Both halves are
// asserted in one test so that neither can go green alone if the two rules drift
// apart again.
func TestEmptyProjectIsOneUnfinishedUnitInItsParent(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("empty", "p", domain.StatusBacklog),
		task("t", "p", domain.StatusDone),
	}

	status, err := domain.DeriveStatus(nodes, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if status != domain.StatusBacklog {
		t.Errorf("DeriveStatus(p) = %q, want %q: the empty sub-project has a column and is not "+
			"done, so it holds the parent back", status, domain.StatusBacklog)
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 1 || got.Total != 2 || !got.Defined() || got.Percent() != 50 {
		t.Errorf("ComputeProgress(p) = {Done:%d Total:%d Percent:%d Defined:%v}, "+
			"want {Done:1 Total:2 Percent:50 Defined:true} — the empty sub-project is one "+
			"unfinished unit", got.Done, got.Total, got.Percent(), got.Defined())
	}

	if (status == domain.StatusDone) != (got.Percent() == 100) {
		t.Errorf("the column (%q) and the bar (%d%%) disagree", status, got.Percent())
	}
}

// The other half of D11, and the regression guard for S1-07: the very same empty
// project, asked about ITSELF, still has no bar.
//
// "A project containing only notes has no work in it, and both 0% ('nothing
// done') and 100% ('all done') are lies the UI would render as a bar." That is
// still binding. The two answers are not in conflict — the question here is
// "what is inside this project?" (nothing), and the question in the test above
// is "is this project finished?" (no).
func TestEmptyProjectAskedAboutItselfIsStillUndefined(t *testing.T) {
	cases := []struct {
		name  string
		nodes []domain.Node
	}{
		{
			name: "an empty sub-project",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("sub", "p", domain.StatusBacklog),
				task("t", "p", domain.StatusDone),
			},
		},
		{
			name: "a sub-project holding only notes",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("sub", "p", domain.StatusBacklog),
				note("n1", "sub"),
				note("n2", "sub"),
				task("t", "p", domain.StatusDone),
			},
		},
		{
			name: "a sub-project holding only habits",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("sub", "p", domain.StatusBacklog),
				habit("h", "sub"),
				task("t", "p", domain.StatusDone),
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			own, err := domain.ComputeProgress(tt.nodes, "sub")
			if err != nil {
				t.Fatalf("ComputeProgress(sub) = %v", err)
			}
			if own.Defined() {
				t.Errorf("ComputeProgress(sub).Defined() = true (%d/%d), want false — there is no "+
					"work inside it, and 0%% and 100%% would both be lies", own.Done, own.Total)
			}

			// ... and the same node is one undone unit in its parent, at the same
			// time, from the same node set.
			parent, err := domain.ComputeProgress(tt.nodes, "p")
			if err != nil {
				t.Fatalf("ComputeProgress(p) = %v", err)
			}
			if parent.Done != 1 || parent.Total != 2 {
				t.Errorf("ComputeProgress(p) = %d/%d, want 1/2 — sub is one undone work leaf",
					parent.Done, parent.Total)
			}
		})
	}
}

// A leaf project is counted as DONE when its own stored status says so. Its
// stored status is the truth for the same reason any other leaf's is: there is
// nothing underneath it to derive from.
func TestLeafProjectStoredDoneCountsAsDone(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("empty", "p", domain.StatusDone),
		task("t", "p", domain.StatusDone),
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 2 || got.Total != 2 || got.Percent() != 100 {
		t.Errorf("ComputeProgress(p) = {Done:%d Total:%d Percent:%d}, want {Done:2 Total:2 "+
			"Percent:100} — a leaf project stored as done is a done unit",
			got.Done, got.Total, got.Percent())
	}

	status, err := domain.DeriveStatus(nodes, "p")
	if err != nil {
		t.Fatalf("DeriveStatus() = %v", err)
	}
	if (status == domain.StatusDone) != (got.Percent() == 100) {
		t.Errorf("the column (%q) and the bar (%d%%) disagree", status, got.Percent())
	}
}

// D11 is about LEAF projects only. A project with real children under it is
// still not a unit of work itself — its children are what the bar measures — so
// the denominator must not grow by one per level of nesting.
func TestProjectWithRealChildrenIsNotCountedAsAUnit(t *testing.T) {
	// p ├── p1 ├── p1a (done)
	//   │      └── p1b (week)
	//   └── p2 └── p2a (done)
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("p1", "p", domain.StatusBacklog),
		task("p1a", "p1", domain.StatusDone),
		task("p1b", "p1", domain.StatusWeek),
		project("p2", "p", domain.StatusBacklog),
		task("p2a", "p2", domain.StatusDone),
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 2 || got.Total != 3 {
		t.Errorf("ComputeProgress(p) = %d/%d, want 2/3 — p1 and p2 have children with columns, "+
			"so they are measured through those children and never counted themselves",
			got.Done, got.Total)
	}
}

// Nesting adds levels, not units: a chain of empty projects is ONE work leaf at
// the bottom, however many projects are stacked above it.
//
// Counting per level instead of per leaf is the obvious way to get D11 wrong,
// and it would make a project look less finished the more deeply somebody had
// filed an empty folder.
func TestNestedEmptyProjectsCountOncePerLeaf(t *testing.T) {
	// p ├── a ── b ── c        (three projects deep, c is empty -> ONE leaf)
	//   ├── e                  (empty -> one leaf)
	//   └── t (task, done)     (one leaf, done)
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("a", "p", domain.StatusBacklog),
		project("b", "a", domain.StatusBacklog),
		project("c", "b", domain.StatusBacklog),
		project("e", "p", domain.StatusBacklog),
		task("t", "p", domain.StatusDone),
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 1 || got.Total != 3 {
		t.Errorf("ComputeProgress(p) = %d/%d, want 1/3 — the a-b-c chain is one work leaf (c), "+
			"not three", got.Done, got.Total)
	}

	t.Run("and the chain measured from the middle is the same one leaf", func(t *testing.T) {
		got, err := domain.ComputeProgress(nodes, "a")
		if err != nil {
			t.Fatalf("ComputeProgress(a) = %v", err)
		}
		if got.Done != 0 || got.Total != 1 {
			t.Errorf("ComputeProgress(a) = %d/%d, want 0/1 — a is measured, b is not a leaf, "+
				"c is the single work leaf beneath it", got.Done, got.Total)
		}
	})
}

// A habit is not a unit of work, so it is not in the progress denominator.
//
// This is the consequence of the no-column rule the reviewer measured: a habit
// has no column, so it never becomes done, and counting it would leave a
// project whose every task is finished stuck below 100% for ever.
func TestComputeProgressExcludesAHabit(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		task("t", "p", domain.StatusDone),
		habit("h", "p"),
	}

	got, err := domain.ComputeProgress(nodes, "p")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Done != 1 || got.Total != 1 {
		t.Fatalf("ComputeProgress(p) = %d/%d, want 1/1 — the habit is not work", got.Done, got.Total)
	}
	if got.Percent() != 100 {
		t.Errorf("Percent() = %d, want 100", got.Percent())
	}

	t.Run("a subtree of nothing but habits has no progress at all", func(t *testing.T) {
		only := []domain.Node{
			project("p", "", domain.StatusBacklog),
			habit("h", "p"),
		}
		got, err := domain.ComputeProgress(only, "p")
		if err != nil {
			t.Fatalf("ComputeProgress() = %v", err)
		}
		if got.Defined() {
			t.Errorf("Defined() = true (%d/%d), want false — there is no work in it", got.Done, got.Total)
		}
	})
}

// The headline of D7's undefined case: a subtree with no work in it reports
// neither 0% nor 100%, and the caller renders no bar.
func TestProgressWithZeroNonNoteLeavesIsUndefined(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("sub", "p", domain.StatusBacklog),
		note("n1", "sub"),
	}

	// The subtree asserted here is rooted at the NOTE: a note has no column, so
	// the walk stops on it at once and there is nothing to measure. ("sub" is a
	// leaf project, which since D11 is one unit inside p's denominator — but p
	// is not what is being asked about here.)
	got, err := domain.ComputeProgress(nodes, "n1")
	if err != nil {
		t.Fatalf("ComputeProgress() = %v", err)
	}
	if got.Defined() {
		t.Fatalf("Defined() = true for a note-rooted subtree, want false (got %d/%d)", got.Done, got.Total)
	}
	if got.Fraction() != 0 {
		t.Errorf("Fraction() = %v, want 0 for an undefined progress", got.Fraction())
	}
	if got.Percent() != 0 {
		t.Errorf("Percent() = %d, want 0 for an undefined progress", got.Percent())
	}
}

func TestProgressFractionAndPercent(t *testing.T) {
	tests := []struct {
		name        string
		in          domain.Progress
		wantFrac    float64
		wantPercent int
	}{
		{"undefined", domain.Progress{Done: 0, Total: 0}, 0, 0},
		{"none of four", domain.Progress{Done: 0, Total: 4}, 0, 0},
		{"one of four", domain.Progress{Done: 1, Total: 4}, 0.25, 25},
		{"one of three rounds down", domain.Progress{Done: 1, Total: 3}, 1.0 / 3.0, 33},
		{"two of three rounds up", domain.Progress{Done: 2, Total: 3}, 2.0 / 3.0, 67},
		{"all of four", domain.Progress{Done: 4, Total: 4}, 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Fraction(); got != tt.wantFrac {
				t.Errorf("Fraction() = %v, want %v", got, tt.wantFrac)
			}
			if got := tt.in.Percent(); got != tt.wantPercent {
				t.Errorf("Percent() = %d, want %d", got, tt.wantPercent)
			}
		})
	}
}

func TestComputeProgressErrors(t *testing.T) {
	t.Run("an unknown id", func(t *testing.T) {
		_, err := domain.ComputeProgress([]domain.Node{task("a", "", domain.StatusDone)}, "nope")
		if !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	t.Run("a parent that is not in the loaded set", func(t *testing.T) {
		nodes := []domain.Node{task("orphan", "ghost", domain.StatusWeek)}
		if _, err := domain.ComputeProgress(nodes, "ghost"); !errors.Is(err, domain.ErrNodeNotFound) {
			t.Errorf("err = %v, want ErrNodeNotFound", err)
		}
	})

	t.Run("a cyclic tree terminates with ErrCycle", func(t *testing.T) {
		nodes := []domain.Node{
			project("a", "b", domain.StatusWeek),
			project("b", "a", domain.StatusWeek),
		}
		if _, err := domain.ComputeProgress(nodes, "a"); !errors.Is(err, domain.ErrCycle) {
			t.Errorf("err = %v, want ErrCycle", err)
		}
	})
}

// Every function in this package must give the same answer whatever order the
// caller loaded the rows in: sort_order is not unique, so the id is the
// tie-break.
func TestDerivationIsIndependentOfLoadOrder(t *testing.T) {
	forward := []domain.Node{
		project("p", "", domain.StatusDone),
		task("a", "p", domain.StatusDone),
		task("b", "p", domain.StatusWeek),
		note("n", "p"),
	}
	// Explicit, distinct sort_orders that run counter to the load order, so
	// that the bucketing really is by sort_order and not by arrival.
	for i := range forward {
		forward[i].SortOrder = len(forward) - i
	}
	reversed := make([]domain.Node, 0, len(forward))
	for i := len(forward) - 1; i >= 0; i-- {
		reversed = append(reversed, forward[i])
	}

	gotF, err := domain.DeriveStatus(forward, "p")
	if err != nil {
		t.Fatalf("DeriveStatus(forward) = %v", err)
	}
	gotR, err := domain.DeriveStatus(reversed, "p")
	if err != nil {
		t.Fatalf("DeriveStatus(reversed) = %v", err)
	}
	if gotF != gotR {
		t.Errorf("DeriveStatus depends on load order: %q vs %q", gotF, gotR)
	}

	pF, err := domain.ComputeProgress(forward, "p")
	if err != nil {
		t.Fatalf("ComputeProgress(forward) = %v", err)
	}
	pR, err := domain.ComputeProgress(reversed, "p")
	if err != nil {
		t.Fatalf("ComputeProgress(reversed) = %v", err)
	}
	if pF != pR {
		t.Errorf("ComputeProgress depends on load order: %+v vs %+v", pF, pR)
	}
}
