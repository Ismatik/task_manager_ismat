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
			// D9 in the denominator: a project is what the bar is drawn for,
			// never a unit it measures, so a project shaped like a leaf is not
			// counted and the subtree has no work in it at all.
			name: "a PROJECT whose children are all notes is not a unit of work",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("p1", "p", domain.StatusDone),
				note("n", "p1"),
			},
			id: "p", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
		{
			name: "an empty project is undefined, not zero per cent",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
			},
			id: "p", wantDone: 0, wantTotal: 0, wantDefined: false,
		},
		{
			name: "an empty project among real tasks counts for nothing",
			nodes: []domain.Node{
				project("p", "", domain.StatusBacklog),
				project("empty", "p", domain.StatusBacklog),
				task("t", "p", domain.StatusDone),
			},
			id: "p", wantDone: 1, wantTotal: 1, wantDefined: true,
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
			// p is a leaf here (all children are notes) and p is a project, so
			// it is NOT counted: there is no work in this subtree to measure.
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

// The headline of D7's undefined case: a subtree with no work in it reports
// neither 0% nor 100%, and the caller renders no bar.
func TestProgressWithZeroNonNoteLeavesIsUndefined(t *testing.T) {
	nodes := []domain.Node{
		project("p", "", domain.StatusBacklog),
		project("sub", "p", domain.StatusBacklog),
		note("n1", "sub"),
	}

	// "sub" is a leaf (all-notes children) and is itself a project, so it is not
	// counted either; this case asserts the note-rooted subtree.
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
