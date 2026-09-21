package service

import (
	"encoding/json"
	"time"

	"nexus/internal/domain"
)

// # The JSON wire contract (S2-02)
//
// Every exported field of every type in this file carries an explicit
// lowerCamelCase json tag, and so do the domain types they carry (domain.Node,
// domain.Tag, domain.TimeEntry). The names are chosen here, once, and reviewed
// as one diff hunk — not inherited from the Go identifiers, where a rename in Go
// would silently rename the wire field and break every TypeScript call site at
// runtime rather than at compile time. dto_test.go pins the whole shape against
// an inline golden string and asserts by reflection that no exported field is
// missing a tag.
//
// Two conventions the frontend may rely on:
//
//   - a calendar date is the string "YYYY-MM-DD" (domain.Date), an instant is
//     RFC 3339 (time.Time), and an absent one of either is null;
//   - a LIST IS NEVER null. Tags, Children and ColumnView.Nodes marshal as []
//     when they are empty, which is what the MarshalJSON methods below are for.
//     A frontend that has to guard every list against null will forget once, and
//     the crash will be in whichever view was written last.

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
	Node domain.Node `json:"node"`

	// Status is the DERIVED status (D2): the node's own for a leaf, and for a
	// parent the least-advanced status among its non-done children that HAVE a
	// Kanban column, done only when all of them are done. A child with no column
	// — a note or a habit (D10) — is excluded along with its whole subtree, and
	// a node with no column of its own always reads backlog, which means "no
	// column" rather than "in the Backlog column". This is the column the card
	// belongs in, and it is never written back.
	Status domain.Status `json:"status"`

	// Progress is done leaves over total leaves (D7). Leaves with no Kanban
	// column — notes and habits — are excluded, and so is everything beneath
	// them. Check Defined before drawing a bar: a subtree with no work in it is
	// neither 0% nor 100%.
	Progress ProgressView `json:"progress"`

	// Overdue is due < today && derived status != done (D1), computed against
	// the injected clock.
	Overdue bool `json:"overdue"`

	// IsLeaf reports whether the node behaves as a leaf — no children, or no
	// child with a Kanban column, which is notes and habits (D2, D10). A node
	// with no column of its own is always a leaf: nothing may be parented under
	// one, and every derivation cuts the subtree off there anyway (S1-09). It is
	// what the UI needs to decide between a progress bar and a timer button.
	IsLeaf bool `json:"isLeaf"`

	// Tags are the node's labels, already resolved; the UI never joins.
	Tags []domain.Tag `json:"tags"`

	// Timer says whether the single global timer is running on THIS node, and
	// for how long.
	Timer TimerView `json:"timer"`

	// Children is the subtree, in sort_order. It is populated by Tree and left
	// empty by Board and Search, where a card stands on its own.
	Children []NodeView `json:"children"`
}

// MarshalJSON writes v with its two lists never null: a card with no tags
// carries "tags":[] and a card with no children carries "children":[].
//
// The local `wire` type is the whole trick: it has NodeView's fields and none of
// its methods, so marshalling it does not call this method again. The children
// keep theirs, so a nil list three levels down is filled in too.
func (v NodeView) MarshalJSON() ([]byte, error) {
	type wire NodeView

	w := wire(v)
	if w.Tags == nil {
		w.Tags = []domain.Tag{}
	}
	if w.Children == nil {
		w.Children = []NodeView{}
	}
	return json.Marshal(w)
}

// ProgressView is domain.Progress with the percentage precomputed.
//
// Percent is rounded in Go rather than in the view layer for the same reason
// everything else here is: two implementations of a rounding rule are two
// rounding rules, and the one on screen would be the untested one.
type ProgressView struct {
	Done    int `json:"done"`
	Total   int `json:"total"`
	Percent int `json:"percent"`

	// Defined is false exactly when there are no leaves that count as work to
	// measure — a leaf with no Kanban column, which is a note and a habit alike
	// (D10), is not one. Neither 0% nor 100% is an honest answer then, and the
	// caller draws no bar at all (D7).
	Defined bool `json:"defined"`
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
	Running        bool       `json:"running"`
	EntryID        string     `json:"entryId"`
	StartedAt      *time.Time `json:"startedAt"`
	ElapsedSeconds int        `json:"elapsedSeconds"`
}

// HabitView is one entry of the habits strip, with every derived value already
// computed by Go (D5).
//
// Habits are not cards: they never appear in a Kanban column (D2), so the board
// excludes them and this is their read. It carries everything the strip draws,
// so that the frontend neither filters the tree by type nor asks three further
// questions per habit — both of which would be rules in TypeScript.
type HabitView struct {
	// Node is the stored habit row.
	Node domain.Node `json:"node"`

	// ScheduledToday reports whether today is an occurrence of the habit's
	// recurrence rule. The strip decides what to do with a habit that is not
	// scheduled today — dim it, hide it — which is presentation; the service
	// reports the fact and pre-filters nothing, so that decision costs no
	// second call.
	ScheduledToday bool `json:"scheduledToday"`

	// CheckedToday reports whether the habit is ticked for today.
	CheckedToday bool `json:"checkedToday"`

	// Streak is consecutive SCHEDULED occurrences that were checked, ending at
	// the present (D5) — occurrences, not days, and today's still-pending one
	// does not break it.
	Streak int `json:"streak"`
}

// ColumnView is one Kanban column: the status it stands for and the cards in
// it, in sort_order.
//
// The board is a slice of these rather than a map, because the five columns have
// an order — backlog, this week, today, doing, done — and a map has none, in Go
// or in JSON.
type ColumnView struct {
	Status domain.Status `json:"status"`
	Nodes  []NodeView    `json:"nodes"`
}

// MarshalJSON writes c with an empty column as "nodes":[] rather than null. An
// empty column is the normal state of at least one column on most boards, so
// this is the list a frontend would forget to guard first.
func (c ColumnView) MarshalJSON() ([]byte, error) {
	type wire ColumnView

	w := wire(c)
	if w.Nodes == nil {
		w.Nodes = []NodeView{}
	}
	return json.Marshal(w)
}
