package service

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// snapshot is one consistent reading of the tree: the node set every derivation
// is computed over, the tag index, the running timer and today's date.
//
// It exists so that a board, a tree or a page of search results costs a fixed
// number of queries instead of a query per card. Nothing in here decides
// anything — the indexes are retrieval shape; every rule applied to them comes
// from nexus/internal/domain.
type snapshot struct {
	all     []domain.Node
	kids    map[string][]domain.Node
	tags    map[string][]domain.Tag
	running *domain.TimeEntry
	today   domain.Date
	clock   Clock
}

// loadSnapshot reads the node set, the tag index and the running timer through
// one executor — in practice one transaction, so the three agree with each
// other.
//
// The cost is fixed: one query for the nodes, one listing the tags, one per tag
// for its members, and one for the open entry. It is O(number of tags) and never
// O(number of nodes), which is the property the no-N+1 test pins.
func loadSnapshot(ctx context.Context, exec store.Executor, set []domain.Node, clock Clock) (*snapshot, error) {
	tagIndex, err := loadTagIndex(ctx, store.NewTagRepo(exec))
	if err != nil {
		return nil, err
	}

	running, err := loadRunning(ctx, store.NewTimeEntryRepo(exec))
	if err != nil {
		return nil, err
	}

	kids := make(map[string][]domain.Node, len(set))
	for _, n := range set {
		kids[parentKey(n.ParentID)] = append(kids[parentKey(n.ParentID)], n)
	}

	return &snapshot{
		all:     set,
		kids:    kids,
		tags:    tagIndex,
		running: running,
		today:   domain.Today(clock),
		clock:   clock,
	}, nil
}

// loadTagIndex builds nodeID -> tags with one query per TAG rather than one per
// node.
//
// Tags are a handful of rows and nodes are thousands, so walking the tags is
// what keeps a 200-node board at a constant number of queries. A bulk read of
// node_tags in the store would make this two queries flat; it would be a change
// to the repository layer, and this shape already has the property that matters.
func loadTagIndex(ctx context.Context, tags *store.TagRepo) (map[string][]domain.Tag, error) {
	all, err := tags.ListTags(ctx)
	if err != nil {
		return nil, err
	}

	index := make(map[string][]domain.Tag)
	for _, tag := range all {
		// Archived nodes are included: whether a card is shown is the caller's
		// filter, and an index with holes in it would quietly drop the tags of
		// anything an archive view later asks for.
		members, err := tags.NodesForTag(ctx, tag.ID, true)
		if err != nil {
			return nil, err
		}
		for _, n := range members {
			index[n.ID] = append(index[n.ID], tag)
		}
	}
	return index, nil
}

// loadRunning returns the single open time entry, or nil when no timer runs.
func loadRunning(ctx context.Context, entries *store.TimeEntryRepo) (*domain.TimeEntry, error) {
	open, err := entries.OpenEntry(ctx)
	switch {
	case errors.Is(err, store.ErrNotFound):
		return nil, nil
	case err != nil:
		return nil, err
	}
	return &open, nil
}

// view assembles one NodeView: every derived value, computed here so that
// nothing downstream has to.
func (snap *snapshot) view(n domain.Node) (NodeView, error) {
	status, err := domain.DeriveStatus(snap.all, n.ID)
	if err != nil {
		return NodeView{}, fmt.Errorf("service: reading node %q: %w", n.ID, err)
	}

	progress, err := domain.ComputeProgress(snap.all, n.ID)
	if err != nil {
		return NodeView{}, fmt.Errorf("service: reading node %q: %w", n.ID, err)
	}

	// Overdue is asked of the node as it RENDERS, which is why the derived
	// status goes in first: a parent whose children are all finished is done and
	// must not glow red over a due date nobody has to act on any more.
	rendered := n
	rendered.Status = status

	v := NodeView{
		Node:     n,
		Status:   status,
		Progress: progressView(progress),
		Overdue:  domain.IsOverdue(rendered, snap.today),
		IsLeaf:   n.IsLeaf(snap.kids[n.ID]),
		Tags:     snap.tags[n.ID],
	}
	if snap.running != nil && snap.running.NodeID == n.ID {
		v.Timer = TimerView{
			Running:        true,
			EntryID:        snap.running.ID,
			StartedAt:      &snap.running.StartedAt,
			ElapsedSeconds: int(snap.running.Duration(snap.clock).Seconds()),
		}
	}
	return v, nil
}

// subtreeViews builds the views of parentID's children, each carrying its own
// children, in sort_order.
func (snap *snapshot) subtreeViews(parentID string) ([]NodeView, error) {
	children := snap.kids[parentID]
	out := make([]NodeView, 0, len(children))

	for _, c := range children {
		v, err := snap.view(c)
		if err != nil {
			return nil, err
		}
		if v.Children, err = snap.subtreeViews(c.ID); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// read runs fn against one transaction, so that the node set, the tags and the
// running timer are one consistent reading. Nothing on this path writes.
func (s *TaskService) read(ctx context.Context, fn func(exec store.Executor) error) error {
	return runInTx(ctx, s.db, fn)
}

// Tree returns the subtree rooted at rootID — or the whole forest when rootID is
// nil — with every derivation applied and the children nested in sort_order.
//
// Archived nodes are excluded, as everywhere else; the archive view (Stage 3)
// asks a different question.
//
// # No N+1
//
// The nodes arrive in ONE query: a single recursive CTE for a subtree
// (store.ListSubtree) or one SELECT for the forest. The tags and the running
// timer are read once for the whole call. Loading a 200-node tree therefore
// issues the same number of queries as loading a 2-node one.
func (s *TaskService) Tree(ctx context.Context, rootID *string) ([]NodeView, error) {
	var out []NodeView

	err := s.read(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		var (
			set []domain.Node
			err error
		)
		if rootID == nil {
			set, err = nodes.ListAll(ctx, false)
		} else {
			set, err = nodes.ListSubtree(ctx, *rootID, false)
		}
		if err != nil {
			return err
		}

		snap, err := loadSnapshot(ctx, exec, set, s.clock)
		if err != nil {
			return err
		}

		if rootID == nil {
			out, err = snap.subtreeViews("")
			return err
		}

		// A named root is returned as the single top of the result, with its
		// own derivations, rather than as a bare list of its children.
		root, err := findNode(set, *rootID)
		if err != nil {
			return err
		}
		v, err := snap.view(root)
		if err != nil {
			return err
		}
		if v.Children, err = snap.subtreeViews(root.ID); err != nil {
			return err
		}
		out = []NodeView{v}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// findNode returns the node with this id from an already loaded set.
func findNode(set []domain.Node, id string) (domain.Node, error) {
	for _, n := range set {
		if n.ID == id {
			return n, nil
		}
	}
	return domain.Node{}, fmt.Errorf("service: node %q: %w", id, domain.ErrNodeNotFound)
}

// Board returns the five Kanban columns, in order, each holding the cards that
// belong in it.
//
// What is on the board, and what is not:
//
//   - HABITS never appear, in any column (D2). They are the habit strip's, and
//     HabitService is their read path.
//   - NOTES never appear: a note has no status and no due date, and it is
//     excluded from its parent's derivation and from progress as well.
//   - ARCHIVED nodes never appear.
//   - PARENTS appear in their DERIVED column, which is the point of the whole
//     exercise: the stored status of a parent is meaningless and is never
//     written, so a board that read it would show cards in the wrong columns.
//   - A subtask appears as a card of its own, at any depth. The tree is one
//     tree; a task under a project is still a task that has to be dragged
//     through the columns and finished.
//
// Within a column the order is sort_order, then id — the same deterministic
// order every list in the store uses, so two cards sharing a sort_order never
// swap places between two reads.
func (s *TaskService) Board(ctx context.Context) ([]ColumnView, error) {
	columns := make([]ColumnView, 0, len(domain.Statuses()))
	for _, status := range domain.Statuses() {
		columns = append(columns, ColumnView{Status: status, Nodes: []NodeView{}})
	}

	err := s.read(ctx, func(exec store.Executor) error {
		set, err := s.nodes.WithExecutor(exec).ListAll(ctx, false)
		if err != nil {
			return err
		}

		snap, err := loadSnapshot(ctx, exec, set, s.clock)
		if err != nil {
			return err
		}

		index := make(map[domain.Status]int, len(columns))
		for i, c := range columns {
			index[c.Status] = i
		}

		for _, n := range set {
			if !onTheBoard(n) {
				continue
			}
			v, err := snap.view(n)
			if err != nil {
				return err
			}
			i, ok := index[v.Status]
			if !ok {
				return fmt.Errorf("service: node %q derives the status %q, which is not a column",
					n.ID, v.Status)
			}
			columns[i].Nodes = append(columns[i].Nodes, v)
		}

		for i := range columns {
			nodes := columns[i].Nodes
			sort.SliceStable(nodes, func(a, b int) bool {
				if nodes[a].Node.SortOrder != nodes[b].Node.SortOrder {
					return nodes[a].Node.SortOrder < nodes[b].Node.SortOrder
				}
				return nodes[a].Node.ID < nodes[b].Node.ID
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return columns, nil
}

// onTheBoard reports whether a node type belongs in a Kanban column at all.
//
// The rule is the domain's — NodeType.HasColumn — not a second copy of the type
// list. It read `!= habit && != note` before, which is the same rule spelled a
// fourth time, and a rule spelled four times is a rule that will disagree with
// itself: it did, in PlanCascade, which is how a habit came to hold the status
// today while this function quietly kept the row off the screen.
func onTheBoard(n domain.Node) bool {
	return n.Type.HasColumn()
}

// Progress returns the done-leaves-over-total-leaves progress of a subtree (D7)
// — what a project's progress bar draws.
//
// A subtree with no non-note leaves reports Defined = false. That is not 0% and
// not 100%: there is no work in it to measure, and the caller draws no bar.
func (s *TaskService) Progress(ctx context.Context, nodeID string) (ProgressView, error) {
	var out ProgressView

	err := s.read(ctx, func(exec store.Executor) error {
		set, err := s.nodes.WithExecutor(exec).ListAll(ctx, false)
		if err != nil {
			return err
		}

		p, err := domain.ComputeProgress(set, nodeID)
		if err != nil {
			return fmt.Errorf("service: progress of %q: %w", nodeID, err)
		}
		out = progressView(p)
		return nil
	})
	if err != nil {
		return ProgressView{}, err
	}
	return out, nil
}
