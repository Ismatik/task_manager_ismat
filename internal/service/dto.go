package service

import (
	"time"

	"nexus/internal/domain"
)

// NodeView is one card, with every derived value ALREADY COMPUTED by Go.
//
// This DTO is the contract that makes PLAN.md §1 enforceable: the frontend
// renders what it is given and computes nothing — not a status, not a progress
// percentage, not an overdue flag. Anything a card displays that is not a stored
// column is a field here, so there is never a reason for TypeScript to
// reimplement a rule, and the rule has exactly one implementation.
type NodeView struct {
	// Node is the stored row. Its Status field is the STORED status, which is
	// meaningful on leaves only and must not be rendered for a parent — use
	// Status below.
	Node domain.Node

	// Status is the DERIVED status (D2): the node's own for a leaf, and for a
	// parent the least-advanced status among its non-done children that HAVE a
	// Kanban column, done only when all of them are done. A child with no column
	// — a note or a habit (D10) — is excluded along with its whole subtree, and
	// a node with no column of its own always reads backlog, which means "no
	// column" rather than "in the Backlog column". This is the column the card
	// belongs in, and it is never written back.
	Status domain.Status

	// Progress is done leaves over total leaves (D7). Leaves with no Kanban
	// column — notes and habits — are excluded, and so is everything beneath
	// them. Check Defined before drawing a bar: a subtree with no work in it is
	// neither 0% nor 100%.
	Progress ProgressView

	// Overdue is due < today && derived status != done (D1), computed against
	// the injected clock.
	Overdue bool

	// IsLeaf reports whether the node behaves as a leaf — no children, or no
	// child with a Kanban column, which is notes and habits (D2, D10). A node
	// with no column of its own is always a leaf: nothing may be parented under
	// one, and every derivation cuts the subtree off there anyway (S1-09). It is
	// what the UI needs to decide between a progress bar and a timer button.
	IsLeaf bool

	// Tags are the node's labels, already resolved; the UI never joins.
	Tags []domain.Tag

	// Timer says whether the single global timer is running on THIS node, and
	// for how long.
	Timer TimerView

	// Children is the subtree, in sort_order. It is populated by Tree and left
	// empty by Board and Search, where a card stands on its own.
	Children []NodeView
}

// ProgressView is domain.Progress with the percentage precomputed.
//
// Percent is rounded in Go rather than in the view layer for the same reason
// everything else here is: two implementations of a rounding rule are two
// rounding rules, and the one on screen would be the untested one.
type ProgressView struct {
	Done    int
	Total   int
	Percent int

	// Defined is false exactly when there are no leaves that count as work to
	// measure — a leaf with no Kanban column, which is a note and a habit alike
	// (D10), is not one. Neither 0% nor 100% is an honest answer then, and the
	// caller draws no bar at all (D7).
	Defined bool
}

// progressView converts a domain.Progress.
func progressView(p domain.Progress) ProgressView {
	return ProgressView{Done: p.Done, Total: p.Total, Percent: p.Percent(), Defined: p.Defined()}
}

// TimerView is the running-timer indication on a card.
//
// Exactly one node in the whole application can have Running = true, which is
// the single-active invariant (S1-19) seen from the read side. ElapsedSeconds is
// measured against the injected clock, so it is a value a test can assert on;
// the UI ticks its own display between reads rather than polling Go every
// second.
type TimerView struct {
	Running        bool
	EntryID        string
	StartedAt      *time.Time
	ElapsedSeconds int
}

// ColumnView is one Kanban column: the status it stands for and the cards in
// it, in sort_order.
//
// The board is a slice of these rather than a map, because the five columns have
// an order — backlog, this week, today, doing, done — and a map has none, in Go
// or in JSON.
type ColumnView struct {
	Status domain.Status
	Nodes  []NodeView
}
