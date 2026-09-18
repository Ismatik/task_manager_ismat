package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalid is the sentinel every validation failure in this package wraps, so
// that a caller can ask "is this a domain validation problem?" with
// errors.Is(err, domain.ErrInvalid) without having to know which field tripped.
// The field itself is on *ValidationError, which errors.As recovers.
var ErrInvalid = errors.New("domain: invalid value")

// ValidationError names the entity and the exact field that failed validation.
//
// Field is spelled as the database column (parent_id, due_source, estimate_min)
// rather than as the Go field, because that is the name the SQL, the migration's
// CHECK constraints and the error a user eventually sees all share. A message
// that says "priority" when the column is `priority` is one fewer translation
// step for whoever is reading the log at the time.
type ValidationError struct {
	Entity string // "node", "tag", "time_entry", "habit_check"
	Field  string // the column name
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("domain: %s.%s: %s", e.Entity, e.Field, e.Reason)
}

// Unwrap makes every ValidationError match errors.Is(err, ErrInvalid).
func (e *ValidationError) Unwrap() error { return ErrInvalid }

// invalid builds a *ValidationError with a formatted reason.
func invalid(entity, field, reason string, a ...any) error {
	return &ValidationError{Entity: entity, Field: field, Reason: fmt.Sprintf(reason, a...)}
}

// dateLayout is the one representation of a date in this project: the same
// 'YYYY-MM-DD' the 0002 migration stores in nodes.due and habit_checks.date.
const dateLayout = "2006-01-02"

// Date is a calendar date — a year, a month and a day — with no time of day and
// no zone.
//
// # Why not time.Time
//
// `due` and `habit_checks.date` are dates, not instants. Storing them as a
// time.Time invites a zone onto a value that has none, and "overdue" (D1)
// compares a due date with today: one value built in UTC and one in the local
// zone differ by up to a day, which the user sees as a task that goes red a day
// early, every morning, forever. A Date cannot carry that bug because it has
// nowhere to put a zone.
//
// Date is comparable, so it works as a map key and with ==; Compare, Before and
// After give the ordering. The zero Date is not a real date — use IsZero.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// NewDate returns the date year-month-day. It does not normalise: use Valid to
// find out whether the result names a real calendar date.
func NewDate(year int, month time.Month, day int) Date {
	return Date{Year: year, Month: month, Day: day}
}

// DateOf returns the calendar date t falls on in t's own location.
//
// The location is deliberately t's and not UTC: callers hand this function the
// result of the injected now func() time.Time, and "today" means the user's
// today. Converting to UTC first would make the small hours of the evening
// belong to tomorrow for anyone east of Greenwich.
func DateOf(t time.Time) Date {
	y, m, d := t.Date()
	return Date{Year: y, Month: m, Day: d}
}

// ParseDate parses a 'YYYY-MM-DD' string. It rejects anything else, including a
// date that is well-formed but impossible such as 2026-02-30.
func ParseDate(s string) (Date, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return Date{}, fmt.Errorf("domain: %q is not a YYYY-MM-DD date: %w", s, ErrInvalid)
	}
	return DateOf(t), nil
}

// Time returns midnight UTC on d. UTC is an implementation detail of the
// arithmetic helpers (AddDays, Weekday): a day in UTC is always exactly 24
// hours, so date arithmetic done through it can never lose or gain an hour to a
// daylight-saving transition.
func (d Date) Time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.UTC)
}

// String renders d as 'YYYY-MM-DD', the representation the database stores.
func (d Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", d.Year, int(d.Month), d.Day)
}

// IsZero reports whether d is the zero value rather than a real date.
func (d Date) IsZero() bool { return d == Date{} }

// Valid reports whether d names a real calendar date — 2026-02-29 does not.
func (d Date) Valid() bool {
	t := d.Time()
	return t.Year() == d.Year && t.Month() == d.Month && t.Day() == d.Day
}

// Compare returns -1 if d is earlier than o, +1 if later and 0 if they are the
// same day.
func (d Date) Compare(o Date) int {
	switch {
	case d.Year != o.Year:
		return sign(d.Year - o.Year)
	case d.Month != o.Month:
		return sign(int(d.Month) - int(o.Month))
	case d.Day != o.Day:
		return sign(d.Day - o.Day)
	default:
		return 0
	}
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}

// Before reports whether d is strictly earlier than o.
func (d Date) Before(o Date) bool { return d.Compare(o) < 0 }

// After reports whether d is strictly later than o.
func (d Date) After(o Date) bool { return d.Compare(o) > 0 }

// Equal reports whether d and o are the same day.
func (d Date) Equal(o Date) bool { return d == o }

// AddDays returns the date n days after d (n may be negative). Month and year
// boundaries are normalised, so 2026-01-31 plus one day is 2026-02-01.
func (d Date) AddDays(n int) Date { return DateOf(d.Time().AddDate(0, 0, n)) }

// Weekday returns the day of the week d falls on.
func (d Date) Weekday() time.Weekday { return d.Time().Weekday() }

// DaysUntil returns the number of whole days from d to o: positive when o is
// later, negative when earlier.
func (d Date) DaysUntil(o Date) int {
	return int(o.Time().Sub(d.Time()) / (24 * time.Hour))
}

// Node is one row of the single `nodes` tree: task, project, habit, note and bug
// are all Nodes, told apart by Type. The fields mirror the columns created by
// migration 0002, with Go types chosen to make the illegal states awkward —
// nullable columns are pointers, enumerated columns are their enum type rather
// than a string.
//
// Status is meaningful on leaves only. A parent's status is derived from its
// children every time it is needed and is never written (D2); see DeriveStatus.
type Node struct {
	ID            string
	ParentID      *string // nil at the root of the tree
	Type          NodeType
	Title         string
	DescriptionMD string
	Status        Status
	Due           *Date // a date, not an instant
	DueSource     DueSource
	Priority      Priority
	EstimateMin   *int    // minutes
	Recurrence    *string // RRULE; required for habits
	Activity      *Activity
	SortOrder     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time // set when the node becomes done
	ArchivedAt    *time.Time // archived rows are hidden, never deleted
}

// Validate reports the first field of n that holds a value the schema or the
// rules forbid, as a *ValidationError naming that field.
func (n Node) Validate() error {
	const entity = "node"

	if strings.TrimSpace(n.ID) == "" {
		return invalid(entity, "id", "must not be empty")
	}
	if n.ParentID != nil {
		if strings.TrimSpace(*n.ParentID) == "" {
			return invalid(entity, "parent_id", "must not be empty when set; use nil for a root node")
		}
		if *n.ParentID == n.ID {
			return invalid(entity, "parent_id", "a node cannot be its own parent")
		}
	}
	if !n.Type.Valid() {
		return invalid(entity, "type", "%q is not one of task, project, habit, note, bug", n.Type)
	}
	if strings.TrimSpace(n.Title) == "" {
		return invalid(entity, "title", "must not be empty")
	}
	if !n.Status.Valid() {
		return invalid(entity, "status", "%q is not one of backlog, week, today, doing, done", n.Status)
	}
	if n.Due != nil && !n.Due.Valid() {
		return invalid(entity, "due", "%s is not a real calendar date", n.Due)
	}
	if !n.DueSource.Valid() {
		return invalid(entity, "due_source", "%q is not manual or auto", n.DueSource)
	}
	if !n.Priority.Valid() {
		return invalid(entity, "priority", "%d is outside 1..4", int(n.Priority))
	}
	if n.EstimateMin != nil && *n.EstimateMin < 0 {
		return invalid(entity, "estimate_min", "%d minutes is negative", *n.EstimateMin)
	}
	if n.Recurrence != nil && strings.TrimSpace(*n.Recurrence) == "" {
		return invalid(entity, "recurrence", "must not be blank when set; use nil for no recurrence")
	}
	if n.Type == NodeTypeHabit && n.Recurrence == nil {
		return invalid(entity, "recurrence", "a habit requires a recurrence rule")
	}
	if n.Activity != nil && !n.Activity.Valid() {
		return invalid(entity, "activity", "%q is not one of the seven D4 activities", *n.Activity)
	}
	if n.CreatedAt.IsZero() {
		return invalid(entity, "created_at", "must be set")
	}
	if n.UpdatedAt.IsZero() {
		return invalid(entity, "updated_at", "must be set")
	}
	return nil
}

// IsLeaf reports whether n behaves as a leaf, given its direct children.
//
// A node is a leaf when it has no children **or every child is a note** (D2).
// This is the predicate the rest of the package is built on: status derivation,
// progress, the cascade and "only leaves start a timer" all ask it, so the
// all-notes case is decided here once instead of being re-derived — and
// re-fumbled — in four places.
//
// children must be n's direct children; nothing else in the slice is filtered
// out, and archived children are counted like any other. Callers load the node
// set they mean.
func (n Node) IsLeaf(children []Node) bool {
	for _, c := range children {
		if c.Type != NodeTypeNote {
			return false
		}
	}
	return true
}

// CanEnterDoing reports whether n may be given the STORED status doing, given
// its direct children (D2, D7, D9).
//
// # A type with no column cannot be in one
//
// The first question is not about leaf-ness at all: doing is a Kanban column,
// and a note or a habit has none (NodeType.HasColumn, PLAN.md §4), so neither
// can be doing whatever its children look like. Starting from HasColumn is what
// keeps this predicate agreeing with CheckStatus and with PlanCascade — it
// previously said yes to a childless habit, and the cascade believed it.
//
// # Type beats the leaf rule (D9)
//
// D7 says a project never enters doing and never runs a timer; D2 says a node
// with no children, or with only note children, behaves as a leaf and can be
// dragged and timed. An empty project satisfies both descriptions, and the user
// has ruled that the TYPE wins: the per-type rule is the more specific one, so
// a project is never doing no matter how few children it has. A freshly created
// project is therefore inert until it gains children; adding a child is how
// work becomes timeable.
//
// A parent that is not a leaf is excluded because doing means a timer and only
// leaves run one. Such a parent still RENDERS in the Doing column when a leaf
// underneath it is running — that is DeriveStatus's answer, not a stored status.
func (n Node) CanEnterDoing(children []Node) bool {
	if !n.Type.HasColumn() || n.Type == NodeTypeProject {
		return false
	}
	return n.IsLeaf(children)
}

// CanStartTimer reports whether a time entry may be opened on n, given its
// direct children (D2, D7, D9).
//
// A timer is what doing MEANS, so the answer is CanEnterDoing's: a node that may
// not be in the Doing column may not be timed either, and there is no third
// thing to check. A habit used to be named here a second time, because
// CanEnterDoing let one through; it is refused one level down now, by the
// no-column rule that also says a habit is checked off rather than timed and has
// no PMP activity (D4). A project at any size, a note, and a node with non-note
// children are refused for the reasons CanEnterDoing gives.
//
// It stays a separate method: "may this be dragged to Doing?" and "may a timer
// be opened on this?" are asked by different callers, and the day one of them
// gains a rule the other has not got, this is where it goes.
func (n Node) CanStartTimer(children []Node) bool {
	return n.CanEnterDoing(children)
}

// ErrTypeHasNoColumn is returned when a node whose TYPE has no Kanban column —
// a note or a habit — is given a column's status, by any door: a drag, the
// create path, or whatever is written next (PLAN.md §4, Type.HasColumn).
var ErrTypeHasNoColumn = errors.New("domain: this node type has no kanban column")

// ErrTypeHasNoDue is returned when a node whose TYPE may not carry a due date —
// a note — is given one, by any door: the create path, a user edit, or whatever
// is written next (PLAN.md §4, NodeType.HasDue).
//
// It is a separate sentinel from ErrTypeHasNoColumn because it is a separate
// rule with a different membership: a habit has no column but may be due on a
// date. A caller that matched one sentinel for both would refuse the habit too.
var ErrTypeHasNoDue = errors.New("domain: this node type has no due date")

// CheckDue reports whether n may be STORED carrying n.Due.
//
// It is the due-date twin of CheckStatus and exists for the same reason: the
// rule was spelled in prose ("a note has no status, no due") and enforced on the
// status half only, so a note could be created with a due date and given one
// afterwards by a hand edit. Both doors ask this now.
//
// Clearing a due date is always allowed — a nil Due is nothing to refuse, and a
// note that somehow has one must be able to lose it.
func (n Node) CheckDue() error {
	if n.Due != nil && !n.Type.HasDue() {
		return fmt.Errorf("domain: a %s cannot carry the due date %s: %w", n.Type, n.Due, ErrTypeHasNoDue)
	}
	return nil
}

// CheckStatus reports whether n may be STORED carrying n.Status, given its
// direct children, and names the rule that refuses it when it may not.
//
// # Why this is not part of Validate
//
// Validate answers "is this a well-formed row?", one field at a time and with
// reference to nothing else. This asks a question about the TYPE and the
// children — D2, D7 and D9 — which is not a question about a field. Keeping it
// separate also keeps Validate usable on a node read back out of the database,
// where the children are not to hand.
//
// Every door into a stored status is meant to come through here. The defect
// that put it in the package was exactly the absence of that: CreateNode stored
// a project as doing while MoveToColumn refused the very same state, so an
// illegal row was one call away from a rule that was working.
//
// The three refusals:
//
//   - A note or a habit has no column (Type.HasColumn), so the only status it
//     may carry is backlog — the value its NOT NULL column needs, not a claim
//     that it sits in the Backlog column. This is the same predicate
//     CanEnterDoing and PlanCascade start from, so the three cannot disagree
//     about a type the way they once did about a habit.
//   - A project never enters doing, empty or not (D9) — the same sentinel the
//     drag returns, because it is the same rule.
//   - Anything else that is not a leaf may not be stored as doing either, since
//     doing means a timer and only leaves run one (D2). Such a parent RENDERS
//     in Doing through DeriveStatus instead; nothing is written.
func (n Node) CheckStatus(children []Node) error {
	if !n.Type.HasColumn() {
		if n.Status != StatusBacklog {
			return fmt.Errorf("domain: a %s cannot carry the status %q: %w", n.Type, n.Status, ErrTypeHasNoColumn)
		}
		return nil
	}
	if n.Status == StatusDoing && !n.CanEnterDoing(children) {
		if n.Type == NodeTypeProject {
			return fmt.Errorf("domain: node %q: %w", n.ID, ErrProjectNeverDoing)
		}
		return invalid("node", "status", "only a leaf is stored as doing, and %q has children that are not notes", n.ID)
	}
	return nil
}

// DefaultActivity returns the PMP activity a new node of type t starts with
// (D4), or nil when the type has no sensible default. The user may override it
// at any time, which is why this is a starting value in the domain rather than
// a column DEFAULT in the schema.
//
// A habit gets nil: habits are not billable work and never reach a timelog.
func DefaultActivity(t NodeType) *Activity {
	switch t {
	case NodeTypeTask:
		return activityPtr(ActivityDevelopment)
	case NodeTypeBug:
		return activityPtr(ActivityTesting)
	case NodeTypeProject:
		return activityPtr(ActivityManagement)
	case NodeTypeNote:
		return activityPtr(ActivityDocumentation)
	case NodeTypeHabit:
		return nil
	default:
		return nil
	}
}

func activityPtr(a Activity) *Activity { return &a }

// Tag is a colour-coded label. A node carries any number of them through the
// node_tags join table.
type Tag struct {
	ID    string
	Name  string
	Color string // '' means "use the palette's default chip colour"
}

// Validate reports the first field of t that holds a forbidden value.
func (t Tag) Validate() error {
	const entity = "tag"

	if strings.TrimSpace(t.ID) == "" {
		return invalid(entity, "id", "must not be empty")
	}
	if strings.TrimSpace(t.Name) == "" {
		return invalid(entity, "name", "must not be empty")
	}
	if t.Color != "" && !isHexColour(t.Color) {
		return invalid(entity, "color", "%q is not empty or a #RRGGBB colour", t.Color)
	}
	return nil
}

// isHexColour reports whether s is exactly '#' followed by six hex digits.
func isHexColour(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
