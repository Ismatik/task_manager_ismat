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
//   - A leaf — no children, or children none of which has a column — reports its
//     own stored status. Derivation is for parents only.
//   - A parent reports the least-advanced status among its non-done children,
//     and done only when every one of those children is done.
//   - A child whose TYPE has no Kanban column is excluded entirely: from the
//     "least advanced" scan and from the "all done" test. A project whose only
//     unfinished child is a note or a habit is done.
//   - Derivation is recursive. A grandparent derives from its children's
//     DERIVED statuses, not from whatever their status column happens to hold —
//     a parent's stored status is meaningless and is never written.
//
// # Why the exclusion is NodeType.HasColumn and not "is it a note" (D10)
//
// §4's rule was written for the note alone, and derivation obeyed it literally,
// so a habit child — which has no column, sits inert at backlog and can never
// become done — dragged its parent down for ever: a project holding one done
// task and one habit rendered in the Backlog column while its progress bar read
// 100%. The column and the bar disagreed, and the bar was right, because
// ComputeProgress had already generalised the rule.
//
// D10 settles it the way D9 settled the cascade: a type with no column cannot
// contribute to a column-derived status. NodeType.HasColumn is the one place
// that rule lives — the same predicate PlanCascade, CanEnterDoing, CheckStatus
// and the board's onTheBoard start from — so derivation cannot drift from them
// about a type the way it just did about a habit.
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

	// The SELF cut, at the same predicate and now in the same position as
	// walkProgress's (D10). A type with no column has no column to derive: the
	// answer is backlog, the inert value its NOT NULL column carries, which means
	// "no column" and not "in the Backlog column".
	//
	// Until S1-09 the cut was applied to the CHILDREN only, a few lines below.
	// That made a habit WITH children a non-leaf, which fell through to the
	// parent branch and derived a real Kanban column for something PLAN.md §4
	// says never appears in one. ValidateMove now refuses to create that shape at
	// all; this line makes the derivation right even if one is already stored.
	if !n.Type.HasColumn() {
		return StatusBacklog, nil
	}

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
		// D10: a child with no column contributes nothing to a column-derived
		// status, and neither does anything under it — the subtree is not
		// descended into at all. The test is NodeType.HasColumn rather than a
		// list of type constants; it named the note alone until D10, which is
		// how a habit child came to hold its parent at backlog for ever.
		//
		// The recursive call would cut such a child at the top of the function
		// anyway now. Skipping it here as well is deliberate: a child that is not
		// on the board must not enter the "all done" test either, and returning
		// backlog for it would hold the parent at backlog for ever — the very
		// defect D10 removed.
		if !c.Type.HasColumn() {
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
// # Zero countable leaves is not zero per cent
//
// PLAN.md does not say what a subtree with no work in it should report, and this
// is where it is decided: Total is 0, Done is 0, and Defined reports false.
//
// It is deliberately neither 0% ("nothing done") nor 100% ("all done"), because
// both are lies that a caller would render as a progress bar. A project that
// contains nothing but notes has no work in it; the honest answer is that the
// question does not apply, and the caller draws no bar at all. Every consumer
// must check Defined before reading Fraction or Percent.
//
// This is the answer to "what is inside this project?", and D11 does not touch
// it. Whether that same project is itself a unit in its PARENT's denominator is
// a different question with a different answer; see ComputeProgress.
type Progress struct {
	Done  int
	Total int
}

// Defined reports whether the subtree contains any work to measure. It is false
// exactly when there are no leaves that count as work — countsAsWork's answer,
// which excludes every type with no Kanban column (a note AND a habit, D10) and
// the measured node when it is a project. It said "no non-note leaves" until
// S1-09, which was the note-only spelling of a rule D10 had already generalised.
// See the type's documentation.
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
// half finished the moment that task was done. Which leaf types are units of
// work is countsAsWork's answer: not a note, not a habit — neither has a column
// to be finished in. The subtree of a type with no column is not descended into
// at all, which is the same cut DeriveStatus makes (D10): work parked under a
// habit is off the board, so counting it here would put a bar on screen that the
// derived column contradicts — the very disagreement D10 exists to remove.
//
// # The two questions about a project, and why they differ (D11)
//
// A project is asked about twice, and the answers are not the same:
//
//   - "What is inside this project?" — asked when the project itself is the node
//     being measured. A project with no work beneath it reports an UNDEFINED
//     progress: an empty project, a project holding only notes and a project
//     holding only habits all draw no bar, because both 0% and 100% would be
//     lies. This is the Progress type's promise above and D7's, and D11 leaves
//     it exactly as it was.
//
//   - "Is this project finished?" — asked when the project is found BELOW the
//     node being measured. A project that is a leaf (nothing under it has a
//     column) is one unit of work in its parent's denominator, done iff its own
//     stored status is done. D11: an empty project is real work that has not
//     been broken down yet, so it is counted rather than wished away.
//
// Both hold at once for the same node and they do not conflict: the first looks
// down and finds nothing, the second looks at the project itself and finds it
// unfinished. So project{empty sub-project, task:done} derives backlog and
// reads 1 of 2 — the column and the bar agree that it is not done — while the
// empty sub-project's own bar is still not drawn at all.
//
// A project WITH children that have columns is not a unit either way: its
// children are counted, exactly as before. D11 is about the leaf case only, and
// it is the only part of D9's "a project is never a unit of work" it reverses.
//
// A leaf's own stored status decides whether it is done; nothing is derived
// here, because a leaf is where the stored status is the truth.
func ComputeProgress(nodes []Node, id string) (Progress, error) {
	byID := indexByID(nodes)
	kids := groupByParent(nodes)

	var p Progress
	// measured = true: the walk starts AT the node being asked about, which is
	// the one position where a leaf project counts for nothing (D11, above).
	if err := walkProgress(byID, kids, id, make(map[string]bool, len(nodes)), &p, true); err != nil {
		return Progress{}, err
	}
	return p, nil
}

// countsAsWork reports whether a leaf of type t is one of the units the progress
// bar measures.
//
// A type with NO COLUMN — a note or a habit — is not. Progress counts work
// through the Kanban columns, and something that never enters one can never
// become done: counting a habit would leave a project whose every task is
// finished reporting 1 of 2 for ever. The rule is NodeType.HasColumn's, the same
// one the cascade and CheckStatus use, rather than a second list of types.
// Since D10 walkProgress stops at such a node before it ever gets here, so that
// half of the condition no longer fires in practice. It stays anyway: this
// predicate has to be right on its own terms, not only at the one call site that
// happens to filter for it first.
//
// A PROJECT is not counted either, even though it has a column (D7, D9): a
// project is the thing a progress bar is drawn FOR, not one of the units the bar
// measures, so a project asked about ITSELF and holding no work reports an
// UNDEFINED progress rather than 0% of 1 — a bar about nothing.
//
// That last paragraph is a statement about the node being MEASURED, and it is
// all this predicate answers. D11 added the other half — a leaf project counts
// as one unit inside its PARENT — and that half is deliberately not here: it is
// not a property of the type, it depends on where the node was found, so it
// lives at walkProgress's single leaf site instead of becoming a second,
// position-blind spelling of the project rule.
func countsAsWork(t NodeType) bool {
	return t.HasColumn() && t != NodeTypeProject
}

// walkProgress accumulates the subtree rooted at id into p. measured is true
// only for the node ComputeProgress was asked about, and false everywhere below
// it; it is what tells the two D11 questions apart.
func walkProgress(byID map[string]Node, kids map[string][]Node, id string, visiting map[string]bool, p *Progress, measured bool) error {
	n, ok := byID[id]
	if !ok {
		return fmt.Errorf("domain: progress of %q: %w", id, ErrNodeNotFound)
	}
	if visiting[id] {
		return fmt.Errorf("domain: progress of %q: %w", id, ErrCycle)
	}
	visiting[id] = true
	defer delete(visiting, id)

	// A type with no column is not work and hides none: the whole subtree is
	// skipped, matching the cut deriveStatus makes over the same predicate. This
	// said `== NodeTypeNote` before D10, so a task under a habit was counted in
	// the bar while the derived column ignored it.
	if !n.Type.HasColumn() {
		return nil
	}

	children := kids[id]
	if n.IsLeaf(children) {
		// A leaf of a type that IS work counts wherever it is found. A leaf
		// project counts too, but only below the node being measured (D11): the
		// project a bar is being drawn for is not one of the units on its own
		// bar. Everything with no column has already returned above, so the
		// !measured half admits exactly the leaf project and nothing else.
		if !countsAsWork(n.Type) && measured {
			return nil
		}
		p.Total++
		if n.Status == StatusDone {
			p.Done++
		}
		return nil
	}

	for _, c := range children {
		if err := walkProgress(byID, kids, c.ID, visiting, p, false); err != nil {
			return err
		}
	}
	return nil
}
