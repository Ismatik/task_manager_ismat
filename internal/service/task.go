package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// Tx is one unit of work: an Executor that can be committed or rolled back.
//
// *sql.Tx satisfies it as it stands. It is an interface rather than a concrete
// *sql.Tx so that a test can wrap a real transaction and make one statement in
// the middle of a cascade fail, which is the only honest way to assert that the
// whole drag rolls back rather than half-applying.
type Tx interface {
	store.Executor
	Commit() error
	Rollback() error
}

// Beginner starts a unit of work. NewBeginner adapts a *sql.DB.
type Beginner interface {
	Begin(ctx context.Context) (Tx, error)
}

// sqlBeginner is the production Beginner: the database itself.
type sqlBeginner struct{ db *sql.DB }

// NewBeginner returns the Beginner backed by db.
func NewBeginner(db *sql.DB) Beginner { return sqlBeginner{db: db} }

func (b sqlBeginner) Begin(ctx context.Context) (Tx, error) {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("service: beginning a transaction: %w", err)
	}
	return tx, nil
}

// TaskService is the write path of the node tree: create, move, re-column,
// re-date, archive and restore.
//
// # It orchestrates, it does not decide
//
// Every rule in here comes from nexus/internal/domain — PlanCascade for a column
// move, PlanMove and ValidateMove for a drag, DueForColumnMove and DueForUserEdit
// for the due date, PlanArchive and PlanRestore for archiving — and every row it
// writes goes through nexus/internal/store. There is no SQL in this package and
// no `if status ==` that decides anything: a condition in here that is not
// simply dispatching to a domain function would be a second copy of a rule.
//
// # It owns the transaction
//
// A column move is a cascade plus a due-date write plus an updated_at on
// everything it touched. Either all of that lands or none of it does, which is
// why the service — not the repository — holds the unit of work: repositories
// are built over store.Executor precisely so that several of them can be bound
// to one transaction.
//
// # The clock and the ids are injected
//
// Neither a timestamp nor a uuid is ever read from the ambient environment here.
// The Clock is the one the domain rules are handed, and newID is the uuid source
// (generating ids is a service concern, never a domain one), so every test is
// deterministic and every assertion is on an exact value.
//
// # It owns the timer coupling too (D13)
//
// PLAN.md §4 says moving a card to Doing opens a time entry, and D13 settles the
// rest of it. The timer service is a collaborator rather than something the
// caller composes afterwards, because "afterwards" is a second transaction: the
// move would commit and the timer might not. It is a REQUIRED collaborator for
// the same reason there is only one MoveToColumn — an optional coupling is a
// coupling somebody constructs their way around.
type TaskService struct {
	db    Beginner
	nodes *store.NodeRepo
	tags  *store.TagRepo
	timer *TimerService
	clock Clock
	newID func() string
}

// NewTaskService returns a service over the repositories, the timer service, the
// clock and the id generator.
//
// timer is not optional: MoveToColumn drives it inside its own transaction
// (D13), and a nil one would be a card in Doing with no entry behind it.
func NewTaskService(db Beginner, nodes *store.NodeRepo, tags *store.TagRepo, timer *TimerService, clock Clock, newID func() string) *TaskService {
	return &TaskService{db: db, nodes: nodes, tags: tags, timer: timer, clock: clock, newID: newID}
}

// inTx runs fn inside one transaction, committing on success and rolling back
// on any error.
//
// A rollback failure is joined onto the original error rather than replacing it:
// the reason the work failed is what the user needs, and the reason the undo
// also failed is what the log needs.
func (s *TaskService) inTx(ctx context.Context, fn func(exec store.Executor) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return errors.Join(err, fmt.Errorf("service: rolling back: %w", rbErr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("service: committing: %w", err)
	}
	return nil
}

// NewNode is the draft a caller hands CreateNode: everything the user chose,
// and nothing the service is responsible for.
//
// The id, the timestamps and the position among the siblings are deliberately
// absent — those are CreateNode's to produce, and a struct that let a caller
// supply them would be a struct that lets a caller supply two different ids.
//
// The zero value of the two enumerated fields means "the default": an empty
// Status is backlog, where a new card starts, and a zero Priority is 4, the
// schema's own default and the least urgent. A nil Activity takes D4's default
// for the type; a non-nil one is the user overriding it up front.
//
// # It is an INPUT on the same wire contract as every output (S2-02)
//
// This struct is bound: App.CreateNode takes it, so the frontend builds one and
// Wails json.Unmarshals it. It therefore needs the tags for the same reason
// NodeView does, and it did not have them — which made the create input
// PascalCase while every read was lowerCamelCase, and generated `Due?: Date`,
// an object shape domain.Date.UnmarshalJSON would have rejected on arrival.
//
// Due stays a *domain.Date rather than becoming a string field with a parse step
// here: domain.Date.UnmarshalJSON already reads "YYYY-MM-DD" through
// domain.ParseDate, the one parser, and rejects anything else with an error
// naming the offending value. A string field would be a second place that turns
// text into a date, and the second one is always the one without the leap-year
// test. The generated TypeScript says `due?: string` because of the ts_type tag,
// which is what actually crosses.
type NewNode struct {
	ParentID      *string          `json:"parentId"`
	Type          domain.NodeType  `json:"type"`
	Title         string           `json:"title"`
	DescriptionMD string           `json:"descriptionMd"`
	Status        domain.Status    `json:"status"`
	Due           *domain.Date     `json:"due" ts_type:"string"` // "YYYY-MM-DD"
	Priority      domain.Priority  `json:"priority"`
	EstimateMin   *int             `json:"estimateMin"`
	Recurrence    *string          `json:"recurrence"`
	Activity      *domain.Activity `json:"activity"`
}

// CreateNode inserts a new node and returns it as it was stored.
//
// It generates the uuid, stamps created_at and updated_at from the injected
// clock, applies D4's default activity for the type unless the caller chose one,
// and appends the node after its existing siblings. A due date supplied here is
// the user's (D1), so it is stored with due_source = manual.
//
// The node is validated by domain.Validate before anything is written, its
// type/status combination by domain.CheckStatus, its type/due combination by
// domain.CheckDue — a note has no due date (PLAN.md §4) — and the parent — if
// there is one — by domain.ValidateMove, so that a parent that does not exist or
// a note used as a parent is refused by the rules rather than by a foreign key.
func (s *TaskService) CreateNode(ctx context.Context, draft NewNode) (domain.Node, error) {
	now := s.clock()

	n := domain.Node{
		ID:            s.newID(),
		ParentID:      draft.ParentID,
		Type:          draft.Type,
		Title:         draft.Title,
		DescriptionMD: draft.DescriptionMD,
		Status:        draft.Status,
		Due:           draft.Due,
		DueSource:     domain.DueSourceManual,
		Priority:      draft.Priority,
		EstimateMin:   draft.EstimateMin,
		Recurrence:    draft.Recurrence,
		Activity:      draft.Activity,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if n.Status == "" {
		n.Status = domain.StatusBacklog
	}
	if n.Priority == 0 {
		n.Priority = domain.Priority4
	}
	if n.Activity == nil {
		n.Activity = domain.DefaultActivity(n.Type)
	}
	if err := n.Validate(); err != nil {
		return domain.Node{}, fmt.Errorf("service: creating a node: %w", err)
	}
	// Creating is the second door into a stored status, and until this line it
	// was an unlocked one: a project could be created straight into doing, which
	// D9 forbids and which MoveToColumn has always refused. The same domain rule
	// answers both, so the two doors cannot drift apart. A node that is being
	// created has no children yet, hence the empty child set.
	if err := n.CheckStatus(nil); err != nil {
		return domain.Node{}, fmt.Errorf("service: creating a node: %w", err)
	}
	// The due date is the other half of the same sentence in PLAN.md §4 — a note
	// has "no status, no due" — and it was the half nothing enforced: a note
	// created with a due date kept it, and the drag that would have written one
	// had been refused for a year by then. A habit is NOT refused here; §4 gives
	// it no column, not no date.
	if err := n.CheckDue(); err != nil {
		return domain.Node{}, fmt.Errorf("service: creating a node: %w", err)
	}

	var created domain.Node
	err := s.inTx(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		set, err := nodes.ListAll(ctx, true)
		if err != nil {
			return err
		}
		// The new node is put into the set before the move is validated so that
		// "is this parent a note?" and "does this parent exist?" are answered by
		// the same rule that answers them for a drag.
		if err := domain.ValidateMove(append(set, n), n.ID, n.ParentID); err != nil {
			return fmt.Errorf("service: creating a node: %w", err)
		}

		n.SortOrder = nextSortOrder(domain.Children(set, parentKey(n.ParentID)))
		if err := nodes.Create(ctx, n); err != nil {
			return err
		}

		created, err = nodes.Get(ctx, n.ID)
		return err
	})
	if err != nil {
		return domain.Node{}, err
	}
	return created, nil
}

// parentKey is the key domain.Children groups by: the parent id, or "" for the
// roots.
func parentKey(parentID *string) string {
	if parentID == nil {
		return ""
	}
	return *parentID
}

// nextSortOrder returns the position after the last of the siblings. It is
// max+1 rather than len so that a gap left by a deleted row cannot put the new
// node on top of an existing one.
func nextSortOrder(siblings []domain.Node) int {
	next := 0
	for _, sib := range siblings {
		if sib.SortOrder >= next {
			next = sib.SortOrder + 1
		}
	}
	return next
}

// MoveNode re-parents nodeID under newParentID at position toIndex among its
// new siblings, and returns the node as it was stored.
//
// The whole subtree comes with it: descendants are attached by their own
// parent_id, which the move does not touch, so their statuses and their
// positions are untouched too.
//
// domain.PlanMove validates first, which is where a circular parent is refused —
// it surfaces here as an error wrapping domain.ErrCircularParent, and because
// the plan is built before anything is written, the database is not touched at
// all.
func (s *TaskService) MoveNode(ctx context.Context, nodeID string, newParentID *string, toIndex int) (domain.Node, error) {
	var moved domain.Node

	err := s.inTx(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		set, err := nodes.ListAll(ctx, true)
		if err != nil {
			return err
		}

		plan, err := domain.PlanMove(set, nodeID, newParentID, toIndex)
		if err != nil {
			return fmt.Errorf("service: moving node %q: %w", nodeID, err)
		}

		now := s.clock()
		byID := make(map[string]domain.Node, len(set))
		for _, n := range set {
			byID[n.ID] = n
		}

		write := func(changes []domain.OrderChange) error {
			for _, c := range changes {
				current := byID[c.NodeID]
				parent := current.ParentID
				if c.NodeID == plan.NodeID {
					parent = plan.NewParentID
				}
				// A sibling that is already where the plan puts it is left
				// alone: rewriting it would bump updated_at on a row that did
				// not move.
				if current.SortOrder == c.SortOrder && samePointer(current.ParentID, parent) {
					continue
				}
				if err := nodes.UpdateParentAndOrder(ctx, c.NodeID, parent, c.SortOrder, now); err != nil {
					return err
				}
			}
			return nil
		}

		if err := write(plan.OldSiblings); err != nil {
			return err
		}
		if err := write(plan.NewSiblings); err != nil {
			return err
		}

		moved, err = nodes.Get(ctx, nodeID)
		return err
	})
	if err != nil {
		return domain.Node{}, err
	}
	return moved, nil
}

// samePointer reports whether two optional parent ids name the same parent.
func samePointer(a, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

// MoveToColumn is the drag between Kanban columns, and it is one transaction:
//
//  1. domain.PlanCascade decides the status and completed_at of every node the
//     drag touches (D2), including refusing a project dragged to Doing (D9),
//  2. a node whose TYPE has no column is refused outright (PLAN.md §4),
//  3. domain.DueForColumnMove decides the due date and its provenance (D1),
//  4. the due date is written, then the cascade is applied in one statement,
//  5. updated_at is stamped on everything either of them touched,
//  6. the timer is coupled to the move (D13) — see coupleTimer.
//
// Either all of it lands or none of it does, the time entry included: if the
// timer insert fails, the card does not stay in Doing.
//
// # This is the only move-to-column entry point, on purpose (D13)
//
// There is no uncoupled variant to bind by mistake. Two methods, one that opens
// a time entry and one that does not, is a second spelling of PLAN.md §4's rule
// with a call-site-shaped fuse: the day a new caller picks the wrong one, the
// coupling is gone for that path only and nothing fails.
//
// The due rule is applied to the dragged node only. D1 describes the column move
// as writing "the" due date of the card that moved, and a cascade that also
// re-dated every descendant would silently overwrite dates the user set on
// tasks they never dragged.
func (s *TaskService) MoveToColumn(ctx context.Context, nodeID string, target domain.Status) (domain.Node, error) {
	var moved domain.Node

	err := s.inTx(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		set, err := nodes.ListAll(ctx, true)
		if err != nil {
			return err
		}

		changes, err := domain.PlanCascade(set, nodeID, target, s.clock)
		if err != nil {
			return fmt.Errorf("service: moving node %q to %s: %w", nodeID, target, err)
		}

		node, err := nodes.Get(ctx, nodeID)
		if err != nil {
			return err
		}
		// A type with no column has nothing to be dragged to, and the due write
		// below does not know that: it used to stamp due = today and
		// due_source = auto onto a note that PlanCascade had just deliberately
		// refused to touch, and onto a habit that never appears in a column at
		// all. The cascade and the due write disagreed about the same row.
		//
		// The reading is REFUSE, not silently do nothing. PlanCascade's answer
		// for these types is "no change", and both readings honour that — but a
		// drag that reports success while the card stays where it was leaves the
		// user with nothing to explain it, which is the same argument D9 already
		// makes for refusing a project dragged to Doing outright. So the state
		// matches PlanCascade and the outcome is visible.
		//
		// The type rule itself is the domain's: NodeType.HasColumn.
		if !node.Type.HasColumn() {
			return fmt.Errorf("service: moving %s %q to %s: %w",
				node.Type, nodeID, target, domain.ErrTypeHasNoColumn)
		}

		now := s.clock()
		dated := domain.DueForColumnMove(node, target, s.clock).ApplyTo(node)
		dated.UpdatedAt = now
		if err := nodes.Update(ctx, dated); err != nil {
			return err
		}

		// Last, so that the cascade's status and completed_at win over the row
		// the due write just rewrote.
		if err := nodes.UpdateStatuses(ctx, changes, now); err != nil {
			return err
		}

		// After the statuses are written, so that the timer sees the node as it
		// now is: canBeTimed refuses a done node, and a card dragged out of Done
		// into Doing is not done any more by this point.
		if err := s.coupleTimer(ctx, exec, nodeID, changes); err != nil {
			return err
		}

		moved, err = nodes.Get(ctx, nodeID)
		return err
	})
	if err != nil {
		return domain.Node{}, err
	}
	return moved, nil
}

// coupleTimer is the Doing↔timer coupling (C1, D13), applied inside the move's
// own transaction.
//
// # It reads the plan, and asks no new question about a type
//
// Everything it needs is already in the cascade domain.PlanCascade produced, and
// that is deliberate: this ticket adds no predicate. Who may be doing is
// domain.DoingRefusal's answer, applied by PlanCascade before this runs (D9), so
// a project or a no-column type has failed the move already and there is nothing
// here to re-decide.
//
//   - LEAVING doing closes the entry (D13 §3). Whichever node the single global
//     timer is running on, if the plan writes a status other than doing onto it,
//     its entry is closed — to any column, done included. A timer still running
//     on a card the user has moved on from is wrong in the data and, through
//     TimerView.Running, visibly wrong on screen.
//
//   - A DIRECT move to doing opens one (D13 §1). "Direct" is not a new rule
//     either: PlanCascade names the dragged node itself exactly when the drag
//     landed on a leaf, and names only its descendants when it did not. So a plan
//     that contains the dragged node IS the direct move, and a cascade — the
//     drag of a parent, which could put twelve leaves into Doing — contains only
//     descendants and opens nothing (D13 §2). There is no principled way to pick
//     which of twelve cards the one global timer belongs to, and a cascade is a
//     planning gesture rather than a "start working now" one.
//
// start closes whatever else was running before it opens, which is the existing
// single-active invariant reached through the existing code path, not a second
// copy of it.
func (s *TaskService) coupleTimer(ctx context.Context, exec store.Executor, nodeID string, changes []domain.StatusChange) error {
	running, err := s.timer.running(ctx, exec)
	if err != nil {
		return err
	}

	if running != nil {
		for _, c := range changes {
			if c.NodeID != running.NodeID || c.Status == domain.StatusDoing {
				continue
			}
			if _, err := s.timer.stop(ctx, exec); err != nil {
				return err
			}
			break
		}
	}

	for _, c := range changes {
		if c.NodeID != nodeID || c.Status != domain.StatusDoing {
			continue
		}
		if _, err := s.timer.start(ctx, exec, nodeID); err != nil {
			return err
		}
		break
	}
	return nil
}

// SetDue is the user editing a due date by hand, which is why it is a different
// method from MoveToColumn and not a flag on it: the two have opposite
// provenance (D1), and a boolean argument would make it possible to mean one and
// apply the other by getting a call site backwards.
//
// The result always carries due_source = manual — including when the date is
// cleared, and including when it happens to equal the date a column move picked.
// From then on a move to Backlog leaves it alone.
//
// A NOTE is refused a date here (domain.ErrTypeHasNoDue): PLAN.md §4 gives it
// "no status, no due". A habit is not — it has no column, which is a different
// rule and not one about dates.
func (s *TaskService) SetDue(ctx context.Context, nodeID string, due *domain.Date) (domain.Node, error) {
	var updated domain.Node

	err := s.inTx(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		node, err := nodes.Get(ctx, nodeID)
		if err != nil {
			return err
		}

		edited := domain.DueForUserEdit(due).ApplyTo(node)
		edited.UpdatedAt = s.clock()
		if err := edited.Validate(); err != nil {
			return fmt.Errorf("service: setting the due date of %q: %w", nodeID, err)
		}
		// The third door onto a due date, and the one a user actually clicks. A
		// note has none (PLAN.md §4); clearing one is still allowed, because a
		// nil date is nothing to refuse.
		if err := edited.CheckDue(); err != nil {
			return fmt.Errorf("service: setting the due date of %q: %w", nodeID, err)
		}
		if err := nodes.Update(ctx, edited); err != nil {
			return err
		}

		updated, err = nodes.Get(ctx, nodeID)
		return err
	})
	if err != nil {
		return domain.Node{}, err
	}
	return updated, nil
}

// ArchiveNode hides nodeID and its whole subtree, and reports how many rows it
// archived.
//
// A descendant that was already archived keeps its original archived_at:
// archiving records when something was put away, and re-stamping it because a
// parent was archived today would rewrite that history for no gain.
//
// It can also write a status onto a node it did NOT archive — the one this
// archive turned into a leaf (D14, K2). That is domain.PlanArchive's decision,
// applied here in the archive's own transaction: the whole point is that the
// board shows the same thing after the archive as before, and two transactions
// would leave a window in which it did not.
func (s *TaskService) ArchiveNode(ctx context.Context, nodeID string) (int, error) {
	return s.setArchived(ctx, nodeID, func(set []domain.Node) (domain.ArchivePlan, error) {
		return domain.PlanArchive(set, nodeID, s.clock)
	})
}

// RestoreNode brings nodeID and its whole subtree back, and reports how many
// rows it restored. Restoring is unconditional over the subtree: a child that
// was archived separately comes back too, because a subtree whose top is visible
// and whose bottom is not has nothing on screen to explain the gap.
// RestoreNode adds no symmetric rule to D14's: see domain.PlanRestore.
func (s *TaskService) RestoreNode(ctx context.Context, nodeID string) (int, error) {
	return s.setArchived(ctx, nodeID, func(set []domain.Node) (domain.ArchivePlan, error) {
		return domain.PlanRestore(set, nodeID)
	})
}

// setArchived applies whichever of the two archive plans it is given, in one
// transaction. Both write the same column, so they differ only in the plan —
// and in D14's status rewrite, which only an archive ever produces.
func (s *TaskService) setArchived(ctx context.Context, nodeID string, plan func([]domain.Node) (domain.ArchivePlan, error)) (int, error) {
	count := 0

	err := s.inTx(ctx, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)

		set, err := nodes.ListAll(ctx, true)
		if err != nil {
			return err
		}

		archive, err := plan(set)
		if err != nil {
			return fmt.Errorf("service: archiving node %q: %w", nodeID, err)
		}

		changes := archive.Archived
		if len(changes) > 0 {
			// One plan writes one value: PlanArchive stamps every row with the
			// same instant and PlanRestore clears every row, so the batch is
			// one statement rather than one per node.
			ids := make([]string, len(changes))
			for i, c := range changes {
				ids[i] = c.NodeID
			}
			if err := nodes.SetArchivedAt(ctx, ids, copyTime(changes[0].ArchivedAt), s.clock()); err != nil {
				return err
			}
			count = len(ids)
		}

		// D14: the node this archive turned into a leaf keeps displaying what
		// it displayed before. In the SAME transaction, so a failed archive
		// leaves the stored status untouched.
		return nodes.UpdateStatuses(ctx, archive.Statuses, s.clock())
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// copyTime returns an independent copy of t, so that no row ends up sharing a
// *time.Time with the plan it came from.
func copyTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	c := *t
	return &c
}
