package domain

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

// ErrCircularParent is returned when a move would put a node inside its own
// subtree — directly (a node as its own parent) or at any depth (a node under
// one of its own descendants). Either one detaches the whole subtree from the
// root: it still has parents all the way up, but the walk never terminates, so
// the nodes exist and nothing can ever display them.
var ErrCircularParent = errors.New("domain: a node cannot be moved into its own subtree")

// ErrTypeHasNoChildren is returned when a move — or a create, which validates
// through the same function — would park something under a node whose TYPE has
// no Kanban column: a note or a habit (NodeType.HasColumn, PLAN.md §4).
//
// # Why the predicate and not the note type
//
// Such a node holds no children, and that is a decision rather than an
// oversight. A node with no column is excluded from status derivation and from
// the progress denominator, and its whole subtree is cut with it (D2, D7, D10),
// so work parked under one would be invisible to both: the parent above would
// derive a status that ignores it and a progress bar that does not count it, and
// the user would see a project reported as done with unfinished tasks inside it.
// Refusing the move is the only way to keep that unreachable.
//
// Every word of that is true of a HABIT and not only of a note — a habit is
// excluded by exactly the same cut — which is why this sentinel was renamed from
// ErrNoteParent and the test generalised to NodeType.HasColumn. It was the last
// place the no-column rule was still spelled as "is it a note", and it produced
// precisely the described outcome: a task created under a habit put the habit
// itself in a Kanban column and left its project reading done at 100% with that
// task unfinished on the board underneath it.
//
// It is NOT the rule about being a child. A habit may still be grouped UNDER a
// project; only the other direction is refused.
var ErrTypeHasNoChildren = errors.New("domain: this node type cannot have children")

// rootKey is the parent key of a node with no parent.
const rootKey = ""

// parentKey returns the grouping key of n's parent: its parent id, or rootKey
// when n is a root.
func parentKey(n Node) string {
	if n.ParentID == nil {
		return rootKey
	}
	return *n.ParentID
}

// Children returns the direct children of parentID, ordered by (sort_order, id).
// Pass rootKey — the empty string — for the roots of the loaded set.
func Children(nodes []Node, parentID string) []Node {
	return groupByParent(nodes)[parentID]
}

// Descendants returns every node strictly below id, in deterministic pre-order:
// each child in (sort_order, id) order, followed by that child's own subtree.
func Descendants(nodes []Node, id string) ([]Node, error) {
	sub, err := Subtree(nodes, id)
	if err != nil {
		return nil, err
	}
	return sub[1:], nil
}

// Subtree returns id followed by every node below it, in deterministic
// pre-order. The first element is always the node itself.
func Subtree(nodes []Node, id string) ([]Node, error) {
	byID := indexByID(nodes)
	kids := groupByParent(nodes)

	root, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("domain: subtree of %q: %w", id, ErrNodeNotFound)
	}

	out := []Node{root}
	visiting := map[string]bool{id: true}
	if err := collectSubtree(kids, id, visiting, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func collectSubtree(kids map[string][]Node, id string, visiting map[string]bool, out *[]Node) error {
	for _, c := range kids[id] {
		if visiting[c.ID] {
			return fmt.Errorf("domain: subtree of %q: %w", c.ID, ErrCycle)
		}
		visiting[c.ID] = true
		*out = append(*out, c)
		if err := collectSubtree(kids, c.ID, visiting, out); err != nil {
			return err
		}
		delete(visiting, c.ID)
	}
	return nil
}

// Ancestors returns the chain from id's parent up to its root, nearest first.
//
// A parent id that is not in the loaded set ends the chain rather than failing:
// a partially loaded subtree is a normal thing to hand this package, and the
// top of it legitimately points at a row that was not loaded.
//
// The walk is bounded by the size of the loaded set, so a corrupt graph in which
// a node is its own ancestor returns ErrCycle instead of looping forever.
func Ancestors(nodes []Node, id string) ([]Node, error) {
	byID := indexByID(nodes)

	n, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("domain: ancestors of %q: %w", id, ErrNodeNotFound)
	}

	var out []Node
	seen := map[string]bool{id: true}
	for n.ParentID != nil {
		parent, ok := byID[*n.ParentID]
		if !ok {
			break
		}
		if seen[parent.ID] {
			return nil, fmt.Errorf("domain: ancestors of %q: %w", id, ErrCycle)
		}
		seen[parent.ID] = true
		out = append(out, parent)
		n = parent
	}
	return out, nil
}

// ValidateMove reports whether nodeID may be re-parented to newParentID. A nil
// newParentID means "make it a root", which is always allowed for a node that
// exists.
//
// It rejects, in this order:
//
//   - a nodeID that is not in the loaded set (ErrNodeNotFound),
//   - a newParentID that is not in the loaded set (ErrNodeNotFound),
//   - a node moved into itself or into its own subtree, at any depth
//     (ErrCircularParent),
//   - a new parent whose TYPE has no Kanban column — a note or a habit
//     (NodeType.HasColumn, ErrTypeHasNoChildren),
//   - a loaded set that is already cyclic (ErrCycle) — the walk is bounded, so
//     corrupt data produces an error rather than a hang.
//
// The circular check comes before the no-column check so that a note dragged
// onto itself reports the more specific reason.
//
// Only the PARENT's type is examined. What is being moved does not matter: a
// habit grouped under a project is legal and stays legal.
func ValidateMove(nodes []Node, nodeID string, newParentID *string) error {
	byID := indexByID(nodes)

	if _, ok := byID[nodeID]; !ok {
		return fmt.Errorf("domain: move %q: %w", nodeID, ErrNodeNotFound)
	}
	if newParentID == nil {
		return nil
	}

	parent, ok := byID[*newParentID]
	if !ok {
		return fmt.Errorf("domain: move %q under %q: %w", nodeID, *newParentID, ErrNodeNotFound)
	}
	if parent.ID == nodeID {
		return fmt.Errorf("domain: move %q under itself: %w", nodeID, ErrCircularParent)
	}

	ancestors, err := Ancestors(nodes, parent.ID)
	if err != nil {
		return err
	}
	for _, a := range ancestors {
		if a.ID == nodeID {
			return fmt.Errorf("domain: move %q under its own descendant %q: %w",
				nodeID, parent.ID, ErrCircularParent)
		}
	}

	// The rule is NodeType.HasColumn's, the same predicate deriveStatus,
	// walkProgress, PlanCascade, Node.IsLeaf and the board's onTheBoard start
	// from. It named the note type alone until S1-09, which is how a task came to
	// be parkable under a habit — invisible to every one of them.
	if !parent.Type.HasColumn() {
		return fmt.Errorf("domain: move %q under the %s %q: %w",
			nodeID, parent.Type, parent.ID, ErrTypeHasNoChildren)
	}
	return nil
}

// OrderChange is one row of a reordering plan: the sort_order a node must be
// written with.
type OrderChange struct {
	NodeID    string
	SortOrder int
}

// Reorder returns the complete new ordering of a sibling set after nodeID is
// moved to toIndex — a dense, gapless 0..n-1 sequence, one entry per sibling,
// in index order.
//
// The whole sequence is returned rather than only the rows that moved, because
// sort_order arrives from the database with gaps, duplicates and defaults of 0
// in it. Rewriting the range densely is what makes the next reorder
// predictable; emitting only the difference would preserve whatever mess was
// there. Applying the plan twice is a no-op.
//
// The starting order is (sort_order, id), so two siblings that share a
// sort_order still have one defined order rather than whatever order they came
// back from the database in. Everything not moved keeps its relative position.
//
// toIndex is clamped into 0..n-1: a drag to "past the end" is a drag to the end,
// which is what the gesture meant.
func Reorder(siblings []Node, nodeID string, toIndex int) ([]OrderChange, error) {
	ordered := sortSiblings(siblings)

	from := -1
	for i, n := range ordered {
		if n.ID == nodeID {
			from = i
			break
		}
	}
	if from < 0 {
		return nil, fmt.Errorf("domain: reorder %q: %w", nodeID, ErrNodeNotFound)
	}

	moved := ordered[from]
	return denseOrder(insertAt(without(ordered, nodeID), moved, toIndex)), nil
}

// sortSiblings returns a copy of nodes ordered by (sort_order, id). The id is
// the tie-break because sort_order is neither unique nor gapless as it comes
// back from the database.
func sortSiblings(nodes []Node) []Node {
	out := make([]Node, len(nodes))
	copy(out, nodes)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].SortOrder != out[j].SortOrder {
			return out[i].SortOrder < out[j].SortOrder
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// without returns nodes with nodeID removed, order preserved.
func without(nodes []Node, nodeID string) []Node {
	out := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		if n.ID != nodeID {
			out = append(out, n)
		}
	}
	return out
}

// insertAt returns dest with n spliced in at index, clamped into 0..len(dest).
// A drag to "past the end" is a drag to the end, which is what the gesture
// meant.
func insertAt(dest []Node, n Node, index int) []Node {
	if index < 0 {
		index = 0
	}
	if index > len(dest) {
		index = len(dest)
	}
	out := make([]Node, 0, len(dest)+1)
	out = append(out, dest[:index]...)
	out = append(out, n)
	out = append(out, dest[index:]...)
	return out
}

// MovePlan is what a drag produces: exactly one parent_id change and the dense
// re-ordering of the sibling ranges it disturbed.
//
// # Moving a node moves its whole subtree
//
// The subtree comes along by construction. Descendants are attached to the moved
// node by their own parent_id, which the move does not touch, so a plan that
// changes one parent_id relocates the entire subtree and leaves every
// descendant's parent_id AND status exactly as it was. That is why there is no
// list of descendants here: a plan that mentioned them could get them wrong.
type MovePlan struct {
	// NodeID is the only node whose parent_id changes.
	NodeID string
	// NewParentID is its new parent, nil for a root.
	NewParentID *string
	// OldSiblings is the dense re-ordering of the range the node left, empty
	// when the node did not change parent.
	OldSiblings []OrderChange
	// NewSiblings is the dense re-ordering of the range the node joined,
	// including the node itself.
	NewSiblings []OrderChange
}

// PlanMove produces the plan for re-parenting nodeID under newParentID at
// position toIndex among its new siblings. It validates the move first, so a
// circular parent never reaches a plan.
//
// Nothing is written here; S1-13 and S1-18 apply the plan in one transaction.
func PlanMove(nodes []Node, nodeID string, newParentID *string, toIndex int) (MovePlan, error) {
	if err := ValidateMove(nodes, nodeID, newParentID); err != nil {
		return MovePlan{}, err
	}

	byID := indexByID(nodes)
	node := byID[nodeID]

	oldKey := parentKey(node)
	newKey := rootKey
	if newParentID != nil {
		newKey = *newParentID
	}

	kids := groupByParent(nodes)
	plan := MovePlan{NodeID: nodeID, NewParentID: copyString(newParentID)}

	reparented := node
	reparented.ParentID = copyString(newParentID)

	// The range the node leaves, closed up.
	dest := without(kids[oldKey], nodeID)
	if oldKey != newKey {
		// It really left, so that range is rewritten densely on its own and
		// the node is spliced into a different one.
		plan.OldSiblings = denseOrder(dest)
		dest = kids[newKey]
	}
	plan.NewSiblings = denseOrder(insertAt(dest, reparented, toIndex))
	return plan, nil
}

// denseOrder numbers an already-ordered slice 0..n-1.
func denseOrder(nodes []Node) []OrderChange {
	out := make([]OrderChange, len(nodes))
	for i, n := range nodes {
		out[i] = OrderChange{NodeID: n.ID, SortOrder: i}
	}
	return out
}

func copyString(s *string) *string {
	if s == nil {
		return nil
	}
	c := *s
	return &c
}

// ArchiveChange is one row of an archive or restore plan: the value to write to
// archived_at. A non-nil ArchivedAt archives, nil restores.
type ArchiveChange struct {
	NodeID     string
	ArchivedAt *time.Time
}

// ArchivePlan is everything an archive or a restore writes.
//
// Archived is the archived_at column, one row per node. Statuses is D14's
// rewrite — see PlanArchive — and is empty for a restore and for most archives.
// They are one plan because they are one transaction: half an archive is a
// state nobody designed.
type ArchivePlan struct {
	Archived []ArchiveChange
	Statuses []StatusChange
}

// PlanArchive archives id and its whole subtree at the injected now, and
// rewrites the stored status of any node this archive turns into a leaf (D14).
//
// A node that is already archived is left out of the plan, keeping its original
// archived_at. Archiving is a fact about when something was put away, and
// re-stamping a child that was archived last month because its parent was
// archived today would quietly rewrite that history for no gain.
//
// # Why archiving writes a status at all (D14, closing K2)
//
// Since D11 a LEAF's stored status decides whether it counts as done in its
// parent's progress, while a parent's stored status is dead weight — derivation
// ignores it. Archiving can turn a parent into a leaf, and at that moment the
// dead weight comes alive: P{C1:done, C2:backlog} derives backlog and renders in
// the Backlog column while an old cascade left `done` sitting in its status
// column; archive both children and P silently becomes a DONE unit in its
// parent's bar, on a status nobody set. The state is self-consistent — the
// column and the bar agree — which is exactly why nothing would ever flag it.
//
// D14's principle is continuity: what the board showed before the archive is
// what it shows after. So such a node is rewritten to the status it DERIVED
// immediately before the archive, with completed_at set or cleared to match by
// the same statusChange the cascade uses — a project whose children were all
// done stays done, and keeps the completion time it already had.
//
// It applies to any node type; a project is merely where D11 gave it teeth.
//
// Restore needs no symmetric rule: once a node has column-bearing children
// again, derivation takes over and the stored status stops being consulted. See
// PlanRestore.
func PlanArchive(nodes []Node, id string, now func() time.Time) (ArchivePlan, error) {
	subtree, err := Subtree(nodes, id)
	if err != nil {
		return ArchivePlan{}, err
	}

	at := now()
	plan := ArchivePlan{Archived: make([]ArchiveChange, 0, len(subtree))}

	// hidden is the whole subtree, already-archived rows included: what the
	// archive removes from view is the subtree, not merely the rows it writes.
	hidden := make(map[string]bool, len(subtree))
	for _, n := range subtree {
		hidden[n.ID] = true
		if n.ArchivedAt != nil {
			continue
		}
		stamp := at
		plan.Archived = append(plan.Archived, ArchiveChange{NodeID: n.ID, ArchivedAt: &stamp})
	}

	if plan.Statuses, err = planLeafRewrites(nodes, hidden, at); err != nil {
		return ArchivePlan{}, err
	}
	return plan, nil
}

// planLeafRewrites is D14's rule: every node this archive turns into a leaf,
// rewritten to the status it was already displaying.
//
// # No new predicate
//
// "Has no remaining child with a column" is Node.IsLeaf, which is
// NodeType.HasColumn's rule (D10), asked twice — of the set as it WAS and of the
// set as it WILL BE — and the difference between the two answers is the whole
// condition. The value written is the derived status of the set as it was. Both
// already existed; a third way to ask either question is the defect this project
// has paid for three times.
//
// The bound is checkable: a node is rewritten only if IT flipped. A node that
// keeps a column-bearing child never flips, and a node that was ALREADY a leaf —
// one whose only children are notes or habits (D10), or one with no children at
// all — never flips either, so neither is touched.
//
// Archived rows are in neither set: they are not on the board, so they are not
// what derivation or the leaf test are about.
func planLeafRewrites(nodes []Node, hidden map[string]bool, at time.Time) ([]StatusChange, error) {
	before := make([]Node, 0, len(nodes))
	after := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		if n.ArchivedAt != nil {
			continue
		}
		before = append(before, n)
		if !hidden[n.ID] {
			after = append(after, n)
		}
	}

	beforeKids := groupByParent(before)
	afterKids := groupByParent(after)
	index := NewIndex(before)

	var out []StatusChange
	for _, n := range after {
		if n.IsLeaf(beforeKids[n.ID]) || !n.IsLeaf(afterKids[n.ID]) {
			continue
		}
		derived, err := index.Status(n.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, statusChange(n, derived, at))
	}

	// The result holds at most one change, so it is deterministic whatever
	// order the caller loaded the rows in — and no sort is needed to make it
	// so. An archived subtree is connected: every node in it except its root
	// has its parent in it too, so exactly ONE node outside the subtree loses a
	// child, and that node is the root's parent. Nothing else can flip.
	return out, nil
}

// PlanRestore clears archived_at on id and its whole subtree.
//
// # The nested case
//
// Restoring is unconditional: a child that was archived on its own, before its
// parent was archived, is restored along with everything else when the parent is
// restored. archived_at is not a stack, and the alternative — remembering that
// this one child was archived separately and leaving it hidden — produces a
// subtree whose top is visible and whose bottom is not, with nothing in the UI
// to explain the gap or to find the missing child again. Restore means "show
// this subtree", and it means all of it.
//
// Nodes that are not archived are left out of the plan; there is nothing to
// write.
//
// # No status rewrite, deliberately (D14)
//
// PlanArchive rewrites the stored status of a node it turns into a leaf.
// Restoring is NOT the mirror of that and must not become one: a node that gets
// column-bearing children back is a parent again, and a parent's stored status
// is never read — DeriveStatus takes over the moment the children are visible.
// Writing one here would put a value nobody reads back into circulation, which
// is how K2 started. Statuses is therefore always empty.
func PlanRestore(nodes []Node, id string) (ArchivePlan, error) {
	subtree, err := Subtree(nodes, id)
	if err != nil {
		return ArchivePlan{}, err
	}

	plan := ArchivePlan{Archived: make([]ArchiveChange, 0, len(subtree))}
	for _, n := range subtree {
		if n.ArchivedAt == nil {
			continue
		}
		plan.Archived = append(plan.Archived, ArchiveChange{NodeID: n.ID, ArchivedAt: nil})
	}
	return plan, nil
}
