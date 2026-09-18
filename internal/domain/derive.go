package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

// ErrNodeNotFound is returned when an operation names a node that is not in the
// loaded set. Callers match it with errors.Is; it is wrapped with the id.
var ErrNodeNotFound = errors.New("domain: node not found")

// ErrCycle is returned when the loaded node set is not a tree: some node is its
// own ancestor. That is corrupt data rather than a user action — the move
// validation in tree.go refuses to create a cycle — but every walk in this
// package is bounded so that reading corrupt data produces an error instead of
// a hung process.
var ErrCycle = errors.New("domain: cycle in the node tree")

// statusRank orders the statuses from least to most advanced, using the
// declaration order of Statuses(). "Least advanced" in D2 means lowest rank.
var statusRank = func() map[Status]int {
	m := make(map[Status]int, len(Statuses()))
	for i, s := range Statuses() {
		m[s] = i
	}
	return m
}()

// indexByID keys the loaded set by node id.
func indexByID(nodes []Node) map[string]Node {
	byID := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		byID[n.ID] = n
	}
	return byID
}

// groupByParent buckets the loaded set by parent id, with "" for the roots, and
// sorts each bucket by (sort_order, id).
//
// The id is the tie-break on purpose: sort_order is not unique — it defaults to
// 0 and nothing stops two siblings sharing a value — and every plan this package
// produces is asserted to be deterministic. Leaving the order of two nodes with
// the same sort_order down to the order the caller happened to load them in
// would make "same input, same output" true only by luck.
func groupByParent(nodes []Node) map[string][]Node {
	kids := make(map[string][]Node)
	for _, n := range nodes {
		key := ""
		if n.ParentID != nil {
			key = *n.ParentID
		}
		kids[key] = append(kids[key], n)
	}
	for key := range kids {
		bucket := kids[key]
		sort.SliceStable(bucket, func(i, j int) bool {
			if bucket[i].SortOrder != bucket[j].SortOrder {
				return bucket[i].SortOrder < bucket[j].SortOrder
			}
			return bucket[i].ID < bucket[j].ID
		})
	}
	return kids
}

// DeriveStatus returns the status the node with the given id renders in (D2).
//
// # The rule
//
//   - A leaf — no children, or children that are all notes — reports its own
//     stored status. Derivation is for parents only.
//   - A parent reports the least-advanced status among its non-done children,
//     and done only when every one of those children is done.
//   - note children are excluded entirely: from the "least advanced" scan and
//     from the "all done" test. A project whose only unfinished child is a note
//     is done.
//   - Derivation is recursive. A grandparent derives from its children's
//     DERIVED statuses, not from whatever their status column happens to hold —
//     a parent's stored status is meaningless and is never written.
//
// # This function can return doing, and that is not a contradiction
//
// "Parents never enter doing" (D2) is a rule about starting a timer: only a leaf
// opens a time entry, and the cascade in cascade.go never writes doing onto a
// parent. Rendering is a different question. A project whose least-advanced
// unfinished child is being worked on right now shows up in the Doing column,
// because that is where the work is. So DeriveStatus does return StatusDoing for
// such a parent, and the caller must not read that as permission to store it or
// to start a timer on it.
//
// The returned value is a render value. It is never written back to the status
// column, which is what makes it impossible for it to drift from the children.
func DeriveStatus(nodes []Node, id string) (Status, error) {
	byID := indexByID(nodes)
	kids := groupByParent(nodes)
	return deriveStatus(byID, kids, id, make(map[string]bool, len(nodes)))
}

func deriveStatus(byID map[string]Node, kids map[string][]Node, id string, visiting map[string]bool) (Status, error) {
	n, ok := byID[id]
	if !ok {
		return "", fmt.Errorf("domain: derive status of %q: %w", id, ErrNodeNotFound)
	}
	if visiting[id] {
		return "", fmt.Errorf("domain: derive status of %q: %w", id, ErrCycle)
	}
	visiting[id] = true
	defer delete(visiting, id)

	children := kids[id]
	if n.IsLeaf(children) {
		if !n.Status.Valid() {
			return "", invalid("node", "status", "%q on node %q is not a known status", n.Status, n.ID)
		}
		return n.Status, nil
	}

	least := Status("")
	allDone := true
	for _, c := range children {
		if c.Type == NodeTypeNote {
			continue
		}
		s, err := deriveStatus(byID, kids, c.ID, visiting)
		if err != nil {
			return "", err
		}
		if s == StatusDone {
			continue
		}
		allDone = false
		if least == "" || statusRank[s] < statusRank[least] {
			least = s
		}
	}
	if allDone {
		return StatusDone, nil
	}
	return least, nil
}

// Progress is the done-leaves-over-total-leaves fraction of a subtree (D7).
//
// # Zero non-note leaves is not zero per cent
//
// PLAN.md does not say what a subtree with no work in it should report, and this
// is where it is decided: Total is 0, Done is 0, and Defined reports false.
//
// It is deliberately neither 0% ("nothing done") nor 100% ("all done"), because
// both are lies that a caller would render as a progress bar. A project that
// contains nothing but notes has no work in it; the honest answer is that the
// question does not apply, and the caller draws no bar at all. Every consumer
// must check Defined before reading Fraction or Percent.
type Progress struct {
	Done  int
	Total int
}

// Defined reports whether the subtree contains any work to measure. It is false
// exactly when there are no non-note leaves; see the type's documentation.
func (p Progress) Defined() bool { return p.Total > 0 }

// Fraction returns the completed share of the subtree in 0..1, and 0 when the
// progress is not Defined. Check Defined first: an undefined progress is not
// "nothing done".
func (p Progress) Fraction() float64 {
	if !p.Defined() {
		return 0
	}
	return float64(p.Done) / float64(p.Total)
}

// Percent returns Fraction as a whole percentage, rounded half away from zero,
// and 0 when the progress is not Defined. Check Defined first.
func (p Progress) Percent() int {
	return int(math.Round(p.Fraction() * 100))
}

// ComputeProgress counts the done and total leaves of the subtree rooted at id
// (D7).
//
// Only leaves count. An intermediate parent is not a unit of work — its leaves
// are — so counting it would make a project with one deeply nested task look
// half finished the moment that task was done. note leaves are excluded from
// both the numerator and the denominator, consistent with status derivation, and
// a note's subtree is not descended into at all.
//
// A PROJECT is excluded as well, even when it has no children or only note
// children and is therefore a leaf by shape (D7, D9). A project is the thing a
// progress bar is drawn FOR; it is not one of the units the bar measures. An
// empty project and a project holding only notes both report an UNDEFINED
// progress — no work in it to measure — which is what the type's documentation
// above promises and what the board and the tree render as no bar at all.
//
// A leaf's own stored status decides whether it is done; nothing is derived
// here, because a leaf is where the stored status is the truth.
func ComputeProgress(nodes []Node, id string) (Progress, error) {
	byID := indexByID(nodes)
	kids := groupByParent(nodes)

	var p Progress
	if err := walkProgress(byID, kids, id, make(map[string]bool, len(nodes)), &p); err != nil {
		return Progress{}, err
	}
	return p, nil
}

func walkProgress(byID map[string]Node, kids map[string][]Node, id string, visiting map[string]bool, p *Progress) error {
	n, ok := byID[id]
	if !ok {
		return fmt.Errorf("domain: progress of %q: %w", id, ErrNodeNotFound)
	}
	if visiting[id] {
		return fmt.Errorf("domain: progress of %q: %w", id, ErrCycle)
	}
	visiting[id] = true
	defer delete(visiting, id)

	if n.Type == NodeTypeNote {
		return nil
	}

	children := kids[id]
	if n.IsLeaf(children) {
		// D9, applied to the denominator: a PROJECT is a container with a
		// progress bar, never a unit of work itself, so it is not counted even
		// when it is shaped like a leaf — empty, or holding only notes. Type
		// wins over the leaf rule, exactly as it does for doing and for the
		// timer (see Node.CanEnterDoing). Counting it would give an empty
		// project a progress bar reading 0% of 1, which is a bar about nothing.
		if n.Type == NodeTypeProject {
			return nil
		}
		p.Total++
		if n.Status == StatusDone {
			p.Done++
		}
		return nil
	}

	for _, c := range children {
		if err := walkProgress(byID, kids, c.ID, visiting, p); err != nil {
			return err
		}
	}
	return nil
}
