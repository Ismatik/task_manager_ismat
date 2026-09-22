package main

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"nexus/internal/domain"
	"nexus/internal/platform"
	"nexus/internal/service"
)

// Services is everything the bound surface delegates to, constructed in main.go
// and handed over whole.
//
// It is a struct rather than six parameters so that adding a service later is
// not a change at every construction site, and so that main.go reads as a list
// of what the application is made of.
type Services struct {
	Tasks    *service.TaskService
	Timers   *service.TimerService
	Habits   *service.HabitService
	Search   *service.SearchService
	Settings *service.SettingsService
}

// App is the struct bound to Wails. Its exported methods are the frontend's
// entire surface onto Go, and every one of them returns (T, error) — see
// ARCHITECTURE.md §4: Wails turns the second return value into a rejected JS
// promise, and without it the frontend cannot tell "empty result" from "it blew
// up". TestEveryBoundMethodReturnsAnError asserts that mechanically.
//
// # Every method here is a delegation and nothing else
//
// No branch in this file decides anything. Which column a card belongs in, who
// may be `doing`, what a streak is, whether a palette name is allowed — every
// one of those is a rule, and a rule in the binding layer is a rule with a
// second spelling that no test in internal/ can see. The two conversions that do
// appear — an empty rootID meaning "the whole forest", an empty due date meaning
// "clear it" — are wire conventions, not decisions: the frontend cannot send a
// Go nil, and both are documented on the method that performs them.
type App struct {
	// mu guards ctx, which is written by startup on the main goroutine and read
	// by onIPCMessage on the single-instance listener's goroutine.
	mu  sync.RWMutex
	ctx context.Context

	svc Services
}

// NewApp returns the bound application over the constructed services.
func NewApp(svc Services) *App {
	return &App{svc: svc}
}

// refuse is the one line every bound method's error passes through.
//
// service.Refuse (D25) is what produces the wire code; this exists only so that
// a method body stays a single delegation — `return refuse(a.svc.X.Y(ctx))` —
// instead of growing a two-line branch twenty-two times over. It decides
// nothing: which errors are refusals is internal/service/refusal.go's table and
// nothing here reads it.
func refuse[T any](value T, err error) (T, error) {
	return value, service.Refuse(err)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

// context returns the context Wails handed startup, or a background context
// when a call somehow arrives before it.
//
// A bound method cannot take a context.Context — Wails does not supply one — so
// this is where every call below gets theirs. Falling back rather than failing
// is deliberate: the fallback is unreachable from the frontend, which cannot
// call anything before the runtime has started, and an error there would be one
// no user could act on.
func (a *App) context() context.Context {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.ctx == nil {
		return context.Background()
	}
	return a.ctx
}

// ---------------------------------------------------------------------------
// The board and the tree.

// Board returns the five Kanban columns, in order, each with its cards.
func (a *App) Board() ([]service.ColumnView, error) {
	return refuse(a.svc.Tasks.Board(a.context()))
}

// Tree returns the subtree rooted at rootID, or the whole forest when rootID is
// empty.
//
// The empty string stands for "no root" because JavaScript cannot send a Go
// nil. It is a wire convention and not a decision: what a nil root means is
// TaskService.Tree's, unchanged.
func (a *App) Tree(rootID string) ([]service.NodeView, error) {
	if rootID == "" {
		return refuse(a.svc.Tasks.Tree(a.context(), nil))
	}
	return refuse(a.svc.Tasks.Tree(a.context(), &rootID))
}

// Progress returns the done-over-total progress of a subtree.
func (a *App) Progress(nodeID string) (service.ProgressView, error) {
	return refuse(a.svc.Tasks.Progress(a.context(), nodeID))
}

// ---------------------------------------------------------------------------
// Writes.

// CreateNode inserts a node and returns it as it was stored.
func (a *App) CreateNode(draft service.NewNode) (domain.Node, error) {
	return refuse(a.svc.Tasks.CreateNode(a.context(), draft))
}

// MoveToColumn drags a card to a Kanban column, named by its wire value —
// "backlog", "week", "today", "doing" or "done".
//
// This is the COUPLED move (C1, D13): moving a card to Doing opens a time entry
// in the same transaction, and moving it out closes one. There is no uncoupled
// variant to bind — see TaskService.MoveToColumn.
//
// # Why the parameter is a string (S2-02)
//
// It took a domain.Status, and Wails' client generator wrote that into
// frontend/wailsjs/go/main/App.d.ts as `arg2: domain.Status` — a type it never
// emitted into models.ts, because it only generates a TypeScript enum for a type
// registered with EnumBind. The generated client therefore did not compile:
// "Namespace 'domain' has no exported member 'Status'". Nothing caught it
// because tsconfig only includes src/, and nothing in src/ imports the client
// yet; S2-13 is the ticket that would have found it the hard way.
//
// A string is what a status IS on the wire — models.ts has always described
// Node.status as one — and it is the convention the rest of this file already
// follows: SetPalette, SetTheme and SetLanguage all take the value as a string
// and let the service validate it. The conversion below is a cast and not a
// check: which strings are columns is domain.Status.Valid's answer, asked by
// domain.PlanCascade, which refuses an unknown one with the message naming it.
// No rule moves up here.
func (a *App) MoveToColumn(nodeID string, target string) (domain.Node, error) {
	return refuse(a.svc.Tasks.MoveToColumn(a.context(), nodeID, domain.Status(target)))
}

// MoveNode re-parents a node, with its whole subtree, at a position among its
// new siblings. An empty newParentID makes it a root — the same wire convention
// Tree uses.
func (a *App) MoveNode(nodeID string, newParentID string, toIndex int) (domain.Node, error) {
	if newParentID == "" {
		return refuse(a.svc.Tasks.MoveNode(a.context(), nodeID, nil, toIndex))
	}
	return refuse(a.svc.Tasks.MoveNode(a.context(), nodeID, &newParentID, toIndex))
}

// SetDue is the user editing a due date by hand (D1). An empty due CLEARS the
// date, which is the same wire convention again: there is no Go nil to send.
//
// The date arrives as the "YYYY-MM-DD" string domain.Date marshals to (S2-02),
// and is parsed by domain.ParseDate — the one parser, which refuses a date that
// does not exist.
func (a *App) SetDue(nodeID string, due string) (domain.Node, error) {
	if due == "" {
		return refuse(a.svc.Tasks.SetDue(a.context(), nodeID, nil))
	}

	parsed, err := domain.ParseDate(due)
	if err != nil {
		return refuse(domain.Node{}, err)
	}
	return refuse(a.svc.Tasks.SetDue(a.context(), nodeID, &parsed))
}

// SetPriority is the user changing a card's priority — 1..4, lower is more
// urgent. It is what the command palette's four priority rows call.
//
// # Why the parameter is a plain int
//
// For MoveToColumn's reason, one type along. A domain.Priority parameter would
// be written into frontend/wailsjs/go/main/App.d.ts as `arg2: domain.Priority`,
// a name the generator never emits into models.ts — exactly the failure S2-02
// records for domain.Status, where the generated client did not compile. A
// number is also what a priority IS on the wire: models.ts has always described
// Node.priority as one.
//
// The conversion below is a cast and not a check. Which numbers are priorities
// is domain.Priority.Valid's answer, asked by domain.Node.Validate inside
// TaskService.SetPriority, which refuses an out-of-range one with the message
// naming it. No rule moves up here.
func (a *App) SetPriority(nodeID string, priority int) (domain.Node, error) {
	return refuse(a.svc.Tasks.SetPriority(a.context(), nodeID, domain.Priority(priority)))
}

// ArchiveNode hides a node and its subtree, and reports how many rows it
// archived.
func (a *App) ArchiveNode(nodeID string) (int, error) {
	return refuse(a.svc.Tasks.ArchiveNode(a.context(), nodeID))
}

// RestoreNode brings a node and its subtree back, and reports how many rows it
// restored.
func (a *App) RestoreNode(nodeID string) (int, error) {
	return refuse(a.svc.Tasks.RestoreNode(a.context(), nodeID))
}

// ---------------------------------------------------------------------------
// Search, habits and the timer.

// Search returns the cards whose title or description match, most relevant
// first.
func (a *App) Search(query string, opts service.SearchOptions) ([]service.NodeView, error) {
	return refuse(a.svc.Search.Search(a.context(), query, opts))
}

// HabitStrip returns every non-archived habit with its schedule, today's check
// and its streak (D5).
func (a *App) HabitStrip() ([]service.HabitView, error) {
	return refuse(a.svc.Habits.Strip(a.context()))
}

// CheckHabitToday ticks a habit for TODAY and returns the strip as it now is —
// so the caller re-renders from the answer.
//
// # It takes no date, and that is the point (S2-18)
//
// Which day "today" is is domain.Today(clock)'s answer, asked here through
// HabitService.CheckToday — the SAME clock and the same call HabitStrip derives
// CheckedToday and ScheduledToday against. A date parameter would invite the
// caller to name the day, and the only caller that can is the frontend, which
// would have to compute it: two clocks for one rule, disagreeing for one minute
// either side of local midnight. CLAUDE.md forbids the frontend deriving a date
// in so many words, so the parameter is gone rather than merely unused.
//
// The dated door is not lost: service.HabitService.Check still takes a
// domain.Date, which is what Stage 7's calendar will bind when it needs to tick
// a day that is not today — a day the user PICKED rather than one the frontend
// worked out.
func (a *App) CheckHabitToday(nodeID string) ([]service.HabitView, error) {
	if err := a.svc.Habits.CheckToday(a.context(), nodeID); err != nil {
		return refuse[[]service.HabitView](nil, err)
	}
	return refuse(a.svc.Habits.Strip(a.context()))
}

// UncheckHabitToday removes today's tick and returns the strip as it now is.
// Today is the clock's, for CheckHabitToday's reason.
func (a *App) UncheckHabitToday(nodeID string) ([]service.HabitView, error) {
	if err := a.svc.Habits.UncheckToday(a.context(), nodeID); err != nil {
		return refuse[[]service.HabitView](nil, err)
	}
	return refuse(a.svc.Habits.Strip(a.context()))
}

// TimerStart opens a timer on a node and reports it as the card renders it.
func (a *App) TimerStart(nodeID string) (service.TimerView, error) {
	if _, err := a.svc.Timers.Start(a.context(), nodeID); err != nil {
		return refuse(service.TimerView{}, err)
	}
	return a.TimerCurrent()
}

// TimerStop closes the running timer, if any, and reports the timer as it now
// is — which is "not running". Stopping when nothing runs is not an error.
func (a *App) TimerStop() (service.TimerView, error) {
	if _, err := a.svc.Timers.Stop(a.context()); err != nil {
		return refuse(service.TimerView{}, err)
	}
	return a.TimerCurrent()
}

// TimerCurrent reports the single global timer, running or not.
//
// It returns the same TimerView a card carries (S2-02) rather than the service's
// own Running struct, so the indicator on a card and the one in the toolbar
// cannot disagree about a shape.
func (a *App) TimerCurrent() (service.TimerView, error) {
	running, err := a.svc.Timers.Current(a.context())
	if err != nil {
		return refuse(service.TimerView{}, err)
	}
	if running == nil {
		return service.TimerView{}, nil
	}
	return service.TimerView{
		Running:        true,
		EntryID:        running.Entry.ID,
		StartedAt:      &running.Entry.StartedAt,
		ElapsedSeconds: int(running.Elapsed.Seconds()),
	}, nil
}

// ---------------------------------------------------------------------------
// Settings.

// Settings returns the four persisted preferences (D6).
func (a *App) Settings() (service.SettingsView, error) {
	return refuse(a.svc.Settings.Settings(a.context()))
}

// SetPalette stores the palette and returns the settings as they now are.
func (a *App) SetPalette(value string) (service.SettingsView, error) {
	return refuse(a.svc.Settings.SetPalette(a.context(), value))
}

// SetTheme stores the theme and returns the settings as they now are.
func (a *App) SetTheme(value string) (service.SettingsView, error) {
	return refuse(a.svc.Settings.SetTheme(a.context(), value))
}

// SetAccent stores the accent override — "" means "use the palette's own" — and
// returns the settings as they now are.
func (a *App) SetAccent(value string) (service.SettingsView, error) {
	return refuse(a.svc.Settings.SetAccent(a.context(), value))
}

// SetLanguage stores the UI language and returns the settings as they now are.
func (a *App) SetLanguage(value string) (service.SettingsView, error) {
	return refuse(a.svc.Settings.SetLanguage(a.context(), value))
}

// ---------------------------------------------------------------------------

// onIPCMessage handles one line from a secondary instance. It is unexported, so
// Wails does not bind it; it is wired to the single-instance lock in main.go.
func (a *App) onIPCMessage(msg platform.Message) {
	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()

	// A message can arrive between Acquire and Wails' startup callback. There
	// is no window to focus yet, and the one being built will come up focused.
	if ctx == nil {
		return
	}

	if msg == platform.MsgQuick {
		// Stage 0 goes no further than this on purpose: the quick-add window is
		// Stage 4. Logging it proves the flag travelled end to end.
		runtime.LogInfo(ctx, "ipc: --quick requested; quick-add arrives in Stage 4, focusing the main window")
	}

	// Unminimise before Show: a window that is merely iconified is still
	// "shown", so Show alone would be a no-op for it. Raising above other
	// applications is up to the window manager and is best-effort on Wayland.
	runtime.WindowUnminimise(ctx)
	runtime.WindowShow(ctx)
}
