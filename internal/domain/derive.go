package domain

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"sync/atomic"
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

// Index is the loaded node set, keyed once: id -> node, and parent id ->
// children in (sort_order, id) order.
//
// # Why it exists (C2)
//
// Every derivation in this file needs both maps, and both are O(n) to build. As
// long as the only entry points were DeriveStatus(nodes, id) and
// ComputeProgress(nodes, id), a caller that wanted the derived status of EVERY
// node — which is exactly what assembling a Kanban board is — rebuilt them once
// per card: O(n²·log n) for the board, invisible at the 200 nodes a personal
// board holds and not invisible at 10 000.
//
// So the maps are lifted into a value the caller can hold. The board builds one
// Index and asks it n questions; nothing else changes.
//
// # It is an INDEX, not a second derivation path
//
// The recursions are still deriveStatus and walkProgress, exactly as they were,
// each written once. Index.Status and Index.Progress are the same two calls with
// the maps already built, and DeriveStatus and ComputeProgress are those methods
// over a freshly built Index. A second traversal "for the indexed path" is the
// defect class that cost Stage 1 three review rounds; there is not one here.
//
// An Index is a snapshot. It holds the nodes it was built from by value, so a
// write to the database afterwards does not reach it — build a new one after a
// write, which is what every service read already does.
type Index struct {
	byID map[string]Node
	kids map[string][]Node
}

// indexBuilds counts how many Index values this process has constructed.
//
// It is instrumentation, and it lives HERE rather than at the call site because
// the property it protects can only be observed where every construction passes.
// C2's regression is a view that goes back to calling DeriveStatus(all, id) once
// per card: a counter wrapped around the service's own call to NewIndex would
// still read exactly 1 while the board built five hundred indexes underneath it,
// which is the one number the test must not be able to miss. DeriveStatus and
// ComputeProgress construct an Index too, so counting constructions here counts
// those as well.
//
// It is monotonic and affects no result; nothing in this package reads it. Tests
// take a difference across the call they are measuring rather than an absolute
// value, so an earlier test's constructions cannot be mistaken for this one's.
var indexBuilds atomic.Int64

// IndexBuilds reports how many Index values this process has constructed since
// it started. It exists for the complexity test described on indexBuilds: take
// the value before and after a Board() call and the difference is the number of
// indexes that board built. Production code has no reason to call it.
func IndexBuilds() int64 { return indexBuilds.Load() }

// NewIndex keys nodes by id and groups them by parent, once.
func NewIndex(nodes []Node) *Index {
	indexBuilds.Add(1)
	return &Index{byID: indexByID(nodes), kids: groupByParent(nodes)}
}

// Children returns the direct children of parentID in (sort_order, id) order —
// the same order Children(nodes, parentID) gives, from the index that is already
// built. Pass "" for the roots of the loaded set.
//
// The returned slice is the index's own: read it, do not sort or append to it.
func (ix *Index) Children(parentID string) []Node { return ix.kids[parentID] }

// Status is DeriveStatus over an index that is already built. See DeriveStatus
// for the rule; this is the same derivation, not a second one.
func (ix *Index) Status(id string) (Status, error) {
	return deriveStatus(ix.byID, ix.kids, id, newVisiting())
}

// Progress is ComputeProgress over an index that is already built. See
// ComputeProgress for the rule; this is the same walk, not a second one.
func (ix *Index) Progress(id string) (Progress, error) {
	var p Progress
	// measured = true: the walk starts AT the node being asked about, which is
	// the one position where a leaf project counts for nothing (D11, below).
	if err := walkProgress(ix.byID, ix.kids, id, newVisiting(), &p, true); err != nil {
		return Progress{}, err
	}
	return p, nil
}

// newVisiting returns the cycle guard the two recursions carry down.
//
// It is deliberately NOT pre-sized to the node set. Both walks delete their
// entry on the way back up, so the map never holds more than the current path —
// its depth, not the set's size. Sizing it len(nodes) was harmless while the
// maps around it were rebuilt per call anyway; once the index is built once and
// asked n questions (C2) it is the last O(n) allocation on a per-question path,
// and it made the indexed board four times slower per card at 1 000 nodes than
// at 100.
func newVisiting() map[string]bool { return map[string]bool{} }

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
//
// It builds an Index and asks it one question. A caller deriving the status of
// more than one node out of the same set — the board — should build the Index
// once and call Index.Status, or it pays for the indexing per card (C2).
func DeriveStatus(nodes []Node, id string) (Status, error) {
	return NewIndex(nodes).Status(id)
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
//
// It builds an Index and asks it one question, exactly as DeriveStatus does, and
// the same note applies: a caller measuring more than one node out of the same
// set builds the Index once and calls Index.Progress (C2).
func ComputeProgress(nodes []Node, id string) (Progress, error) {
	return NewIndex(nodes).Progress(id)
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
