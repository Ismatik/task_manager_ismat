package domain

import (
	"fmt"
	"time"
)

// StatusChange is one row of a cascade plan: the status a node must be written
// with, and the value its completed_at must be written with at the same time.
//
// CompletedAt is ALWAYS written, and nil means NULL. That is not an oversight:
// every node in a plan either becomes done, in which case it needs a completion
// time, or leaves done for a working column, in which case the completion time
// it is carrying is now a lie. Making the field optional would let the two
// columns disagree, which is the whole class of bug D2 exists to remove.
type StatusChange struct {
	NodeID      string
	Status      Status
	CompletedAt *time.Time
}

// PlanCascade is what dragging a card to a column produces (D2).
//
// # There is no rejection any more
//
// The old rule refused to move a parent to Done. It was REPLACED. Dragging a
// node to column target sets status = target on every descendant that is not
// already done and is not a note, and the parent then DERIVES its status from
// what is underneath it.
//
// # What the plan contains, and what it deliberately does not
//
//   - A leaf — no children, or children that are all notes — is the one node
//     the plan writes when the drag lands on a leaf. Exactly one change.
//   - A parent that is not a leaf gets NO ROW OF ITS OWN, at any depth of the
//     drag. Its status is derived by DeriveStatus after the plan is applied.
//     Writing a status onto a parent is precisely the drift D2 exists to
//     prevent, so the plan is built so that it cannot happen.
//   - done descendants are never touched. A finished task does not get
//     un-finished because somebody dragged its parent to Today, and its
//     completed_at survives.
//   - note descendants are never touched. Notes have no column.
//
// # doing is different, because doing means a timer
//
// Only leaves start a timer (D2), so a drag to Doing cascades doing onto the
// non-done, non-note LEAF descendants only. Intermediate parents get no stored
// status; they derive Doing from the leaves underneath them, which is what puts
// them in the Doing column on screen without anything having started a timer on
// them.
//
// For the other targets the cascade writes to every non-done, non-note
// descendant, intermediate parents included, which is what D2 says literally.
// Their stored status stays inert — derivation never reads it — but their
// completed_at is real, and a project that finished needs one.
//
// # Moving out of done
//
// The only node that can leave done is the drag target itself, when it is a
// leaf: done descendants are skipped, so nothing else can. Such a leaf gets the
// new status and a nil CompletedAt, clearing the completion time. Something that
// is back in Today has not been completed, and leaving the old timestamp behind
// would put a completion date on an unfinished task for every report that ever
// reads the column.
//
// Re-dropping an already-done leaf onto Done keeps its original completed_at
// rather than re-stamping it: the work finished when it finished.
//
// The result is a deterministic, ordered slice — pre-order over the subtree —
// that the service applies in one transaction. now is read once, so every node
// completed by one drag shares one timestamp.
func PlanCascade(nodes []Node, rootID string, target Status, now func() time.Time) ([]StatusChange, error) {
	if !target.Valid() {
		return nil, invalid("node", "status", "%q is not a column a node can be dragged to", target)
	}

	subtree, err := Subtree(nodes, rootID)
	if err != nil {
		return nil, err
	}
	root := subtree[0]

	// Notes have no column (PLAN.md §4, behaviour by type), so there is nothing
	// to cascade and nothing to write.
	if root.Type == NodeTypeNote {
		return nil, nil
	}

	kids := groupByParent(nodes)
	at := now()

	// A leaf is dragged on its own: exactly one change, itself, whatever its
	// current status.
	if root.IsLeaf(kids[root.ID]) {
		return []StatusChange{statusChange(root, target, at)}, nil
	}

	out := make([]StatusChange, 0, len(subtree)-1)
	for _, n := range subtree[1:] {
		if n.Type == NodeTypeNote || n.Status == StatusDone {
			continue
		}
		if target == StatusDoing && !n.IsLeaf(kids[n.ID]) {
			continue
		}
		out = append(out, statusChange(n, target, at))
	}
	return out, nil
}

// statusChange builds the change for one node, deciding its completed_at.
func statusChange(n Node, target Status, at time.Time) StatusChange {
	c := StatusChange{NodeID: n.ID, Status: target}
	switch {
	case target != StatusDone:
		// Leaving done, or never having been in it: there is no completion.
		c.CompletedAt = nil
	case n.Status == StatusDone && n.CompletedAt != nil:
		// Already finished; the work finished when it finished.
		kept := *n.CompletedAt
		c.CompletedAt = &kept
	default:
		stamp := at
		c.CompletedAt = &stamp
	}
	return c
}

// ApplyStatusChanges returns a copy of nodes with the plan applied. It exists so
// that a caller — and the tests — can ask what the tree looks like afterwards,
// in particular what the parent then DERIVES, without going near a database.
//
// A change naming a node that is not in the set is an error rather than a
// silent no-op: a plan that has drifted from the set it was built for is a bug,
// and applying three quarters of a cascade is worse than applying none.
func ApplyStatusChanges(nodes []Node, changes []StatusChange) ([]Node, error) {
	out := make([]Node, len(nodes))
	copy(out, nodes)

	index := make(map[string]int, len(out))
	for i, n := range out {
		index[n.ID] = i
	}

	for _, c := range changes {
		i, ok := index[c.NodeID]
		if !ok {
			return nil, fmt.Errorf("domain: apply status change to %q: %w", c.NodeID, ErrNodeNotFound)
		}
		out[i].Status = c.Status
		out[i].CompletedAt = copyTime(c.CompletedAt)
	}
	return out, nil
}

func copyTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}
