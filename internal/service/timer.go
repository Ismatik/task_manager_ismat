package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nexus/internal/domain"
	"nexus/internal/store"
)

// ErrTimerNotAllowed reports that this node is not something a timer can run
// on. Every refusal below wraps it, so a caller that only wants to know "can I
// start here?" matches one error, while a caller that wants to tell the user
// WHY matches the second one in the chain.
var ErrTimerNotAllowed = errors.New("service: a timer cannot run on this node")

// ErrNodeArchived reports that the node is archived.
//
// Archiving is how the user puts something out of sight. Time booked against a
// row that no list shows would be time nobody can find again — neither in the
// day's timelog nor on the card — so the timer refuses rather than quietly
// recording it. Restore the node first.
var ErrNodeArchived = errors.New("service: the node is archived")

// ErrNodeDone reports that the node is already finished.
//
// The decision (the ticket asks for one): a done card cannot be timed. Starting
// work again means the card is not finished, and the way to say that is to move
// it out of Done — which clears completed_at in the same transaction (D2) and
// keeps the two columns telling the same story. Letting a timer run on a done
// node would produce a card that is finished and in progress at once, and a
// completed_at that predates the work it supposedly completed.
var ErrNodeDone = errors.New("service: the node is done")

// Running is an open timer: the entry and how long it has been going, measured
// against the injected clock.
type Running struct {
	Entry   domain.TimeEntry
	Elapsed time.Duration
}

// TimerService holds the single-active invariant:
//
//	exactly one open time_entry, globally.
//
// # Two layers, and they are not the same layer twice
//
// The schema enforces at most one open entry with the `one_open_timer` partial
// unique index, and the repository surfaces the refusal as store.ErrTimerRunning
// (S1-05, S1-15). That is the backstop. The POLICY — that starting a timer on B
// stops the one running on A, rather than failing — is here, and it is why the
// index should never actually fire in normal operation: Start closes the running
// entry before it opens the new one, inside one transaction, so the overlap
// window does not exist even momentarily.
//
// # Who may be timed
//
// domain.Node.CanStartTimer decides, and this file does not re-derive it: a leaf
// that is not a project, not a habit and not a note (D2, D7, D9). A project is
// refused whether or not it has children — the type wins over the leaf rule.
//
// Sleep and lock handling is Stage 4. store.TimeEntryRepo.CloseAll already
// exists for it; nothing here subscribes to anything.
type TimerService struct {
	db      Beginner
	nodes   *store.NodeRepo
	entries *store.TimeEntryRepo
	clock   Clock
	newID   func() string
}

// NewTimerService returns a timer service over the repositories, the clock and
// the id generator.
func NewTimerService(db Beginner, nodes *store.NodeRepo, entries *store.TimeEntryRepo, clock Clock, newID func() string) *TimerService {
	return &TimerService{db: db, nodes: nodes, entries: entries, clock: clock, newID: newID}
}

// runInTx runs fn inside one transaction over b, committing on success and
// rolling back on any error. A rollback that also fails is joined onto the
// original error rather than replacing it.
func runInTx(ctx context.Context, b Beginner, fn func(exec store.Executor) error) error {
	tx, err := b.Begin(ctx)
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

// Start opens a timer on nodeID and returns the entry.
//
// In one transaction it closes whatever was running, on any node, at the
// injected now, and then opens the new entry. Either both land or neither does,
// which is what makes "two timers were briefly open" unrepresentable rather than
// merely unlikely.
//
// Starting on the node that is ALREADY running is a no-op: the same entry comes
// back, with its original started_at. Restarting it would fragment the log into
// two rows and reset the elapsed time the user is watching, which is not what
// pressing start again means.
func (s *TimerService) Start(ctx context.Context, nodeID string) (domain.TimeEntry, error) {
	var started domain.TimeEntry

	err := runInTx(ctx, s.db, func(exec store.Executor) error {
		nodes := s.nodes.WithExecutor(exec)
		entries := s.entries.WithExecutor(exec)

		node, err := nodes.Get(ctx, nodeID)
		if err != nil {
			return err
		}
		children, err := nodes.ListChildren(ctx, &nodeID, true)
		if err != nil {
			return err
		}
		if err := canBeTimed(node, children); err != nil {
			return err
		}

		// What is running now, if anything.
		open, err := entries.OpenEntry(ctx)
		switch {
		case err == nil && open.NodeID == nodeID:
			// Already running here: hand back the entry that is going, rather
			// than stopping and restarting it.
			started = open
			return nil
		case err == nil:
			if err := entries.Close(ctx, open.ID, s.clock()); err != nil {
				return err
			}
		case !errors.Is(err, store.ErrNotFound):
			return err
		}

		started, err = entries.Open(ctx, s.newID(), nodeID, s.clock())
		return err
	})
	if err != nil {
		return domain.TimeEntry{}, err
	}
	return started, nil
}

// canBeTimed reports why a timer may not run on n, or nil.
//
// Every rule it applies comes from the domain; the two service-level refusals
// (archived, done) are facts about the row rather than rules about the tree.
func canBeTimed(n domain.Node, children []domain.Node) error {
	switch {
	case n.Type == domain.NodeTypeProject:
		return fmt.Errorf("service: starting a timer on %q: %w (%w)",
			n.ID, ErrTimerNotAllowed, domain.ErrProjectNeverDoing)
	case n.ArchivedAt != nil:
		return fmt.Errorf("service: starting a timer on %q: %w (%w)",
			n.ID, ErrTimerNotAllowed, ErrNodeArchived)
	case n.Status == domain.StatusDone:
		return fmt.Errorf("service: starting a timer on %q: %w (%w)",
			n.ID, ErrTimerNotAllowed, ErrNodeDone)
	case !n.CanStartTimer(children):
		return fmt.Errorf("service: starting a timer on %q (%s with %d children): %w",
			n.ID, n.Type, len(children), ErrTimerNotAllowed)
	}
	return nil
}

// Stop closes the running timer and returns the entry it closed, or nil when
// nothing was running.
//
// Stopping when nothing is running is a no-op with NO error: the tray menu, the
// keyboard shortcut and Stage 4's sleep handler will all call it without
// knowing, and an error there would be an error the user cannot act on.
func (s *TimerService) Stop(ctx context.Context) (*domain.TimeEntry, error) {
	open, err := s.entries.OpenEntry(ctx)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	at := s.clock()
	if err := s.entries.Close(ctx, open.ID, at); err != nil {
		return nil, err
	}

	ended := at.UTC()
	open.EndedAt = &ended
	return &open, nil
}

// Current returns the running timer and its elapsed time, or nil when nothing is
// running.
//
// The elapsed time is domain.TimeEntry.Duration against the injected clock: one
// implementation of "how long has this been going", in the domain, where the
// tests can pin it to an exact value.
func (s *TimerService) Current(ctx context.Context) (*Running, error) {
	open, err := s.entries.OpenEntry(ctx)
	if errors.Is(err, store.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &Running{Entry: open, Elapsed: open.Duration(s.clock)}, nil
}
