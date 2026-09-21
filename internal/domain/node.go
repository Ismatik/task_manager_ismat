package domain

import (
	"encoding/json"
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

// MarshalJSON renders d as the JSON string "YYYY-MM-DD" — the same
// representation the database column holds, and the same one String gives.
//
// Without this a Date crosses the wire as {"Year":2026,"Month":9,"Day":21},
// which is three numbers a frontend can do arithmetic on. Date exists precisely
// so that a calendar date cannot acquire a zone or an hour (see the type's
// documentation); handing TypeScript the components back invites it to build a
// JavaScript Date out of them, in the browser's zone, and to re-derive "overdue"
// — a rule that lives in Go exactly once (IsOverdue). A string it can only
// display is the point.
//
// A nil *Date needs nothing here: encoding/json writes null for a nil pointer.
func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.String()) }

// UnmarshalJSON reads the "YYYY-MM-DD" string MarshalJSON writes, and rejects
// anything else — including a well-formed but impossible date such as
// 2026-02-30, because ParseDate does.
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("domain: a date must be a %q string: %w", dateLayout, ErrInvalid)
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// Node is one row of the single `nodes` tree: task, project, habit, note and bug
// are all Nodes, told apart by Type. The fields mirror the columns created by
// migration 0002, with Go types chosen to make the illegal states awkward —
// nullable columns are pointers, enumerated columns are their enum type rather
// than a string.
//
// Status is meaningful on leaves only. A parent's status is derived from its
// children every time it is needed and is never written (D2); see DeriveStatus.
//
// # The JSON names are the wire contract (S2-02)
//
// Every field carries an explicit lowerCamelCase tag, so the name the frontend
// binds to is chosen here and reviewed as one diff hunk rather than inherited
// from whatever the Go identifier happens to be. A rename in Go is then a
// deliberate act: dropping the `Md` from DescriptionMD would otherwise silently
// break every TypeScript call site. Nullable columns are pointers and marshal as
// null; Date marshals as "YYYY-MM-DD" and time.Time as RFC 3339.
type Node struct {
	ID            string     `json:"id"`
	ParentID      *string    `json:"parentId"` // nil at the root of the tree
	Type          NodeType   `json:"type"`
	Title         string     `json:"title"`
	DescriptionMD string     `json:"descriptionMd"`
	Status        Status     `json:"status"`
	Due           *Date      `json:"due"` // a date, not an instant
	DueSource     DueSource  `json:"dueSource"`
	Priority      Priority   `json:"priority"`
	EstimateMin   *int       `json:"estimateMin"` // minutes
	Recurrence    *string    `json:"recurrence"`  // RRULE; required for habits
	Activity      *Activity  `json:"activity"`
	SortOrder     int        `json:"sortOrder"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	CompletedAt   *time.Time `json:"completedAt"` // set when the node becomes done
	ArchivedAt    *time.Time `json:"archivedAt"`  // archived rows are hidden, never deleted
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
// A node is a leaf when its OWN type has no Kanban column, or when it has no
// children, or when no child of it has a Kanban column (D2, D10). This is the
// predicate the rest of the package is built on:
// status derivation, progress, the cascade and "only leaves start a timer" all
// ask it, so the case is decided here once instead of being re-derived — and
// re-fumbled — in four places.
//
// # Why the test is NodeType.HasColumn and not "is it a note"
//
// It named the note type literally until D10, and a habit child therefore made
// its parent a non-leaf. That is the same divergence PlanCascade, CanEnterDoing
// and onTheBoard each had: a habit has no column (PLAN.md §4), so it can neither
// be finished nor be dragged, and a parent holding nothing but habits has
// nothing underneath it that a column-derived answer could come from. Such a
// parent is a leaf and reports its own stored status, exactly as an all-notes
// parent already did — which is what stopped a task whose only children are
// habits from producing an empty cascade plan.
//
// # A node with no column of its own is a leaf whatever is under it (S1-09)
//
// Nothing may be parented under such a node any more (ValidateMove,
// ErrTypeHasNoChildren), and if corrupt data holds one anyway, every walk in
// this package already cuts the subtree off at it: deriveStatus, walkProgress
// and PlanCascade all stop there. Answering "not a leaf" for it would be the one
// remaining disagreement — and it was exactly the one that put a habit holding a
// task into a Kanban column.
//
// children must be n's direct children; nothing else in the slice is filtered
// out, and archived children are counted like any other. Callers load the node
// set they mean.
func (n Node) IsLeaf(children []Node) bool {
	if !n.Type.HasColumn() {
		return true
	}
	for _, c := range children {
		if c.Type.HasColumn() {
			return false
		}
	}
	return true
}

// DoingRefusal names the TYPE rule that forbids a node of type t the doing
// status, or nil when the type itself permits it (D2, D7, D9).
//
// # The one place the "who may be doing" type rules live
//
// There are two of them and they have different reasons and different sentinels:
//
//   - a type with NO KANBAN COLUMN — a note, a habit (NodeType.HasColumn,
//     PLAN.md §4) — cannot be in a column at all, so it cannot be in the Doing
//     one either, whatever its children look like;
//   - a PROJECT has a column and still never enters doing, empty or not (D9):
//     it carries a progress bar and never runs a timer, and the user has ruled
//     that this per-type rule beats D2's "a node with no children behaves as a
//     leaf".
//
// It returns the reason rather than a bool so that a caller can keep the
// specific message it had — "a project never enters doing" reads very
// differently from "a habit has no column" — without holding its own copy of
// either rule. That copy is what this function exists to remove: the project
// half was spelled out in four places (CanEnterDoing, CheckStatus, PlanCascade
// and the timer service) with no predicate behind it, which is the same setup
// that let the no-column rule diverge one copy at a time.
//
// NodeType.CanBeDoing is the bool form for callers that only need yes or no.
// This is a question about the TYPE alone; whether a node that passes it is
// also a LEAF is Node.CanEnterDoing's.
func DoingRefusal(t NodeType) error {
	switch {
	case !t.HasColumn():
		return ErrTypeHasNoColumn
	case t == NodeTypeProject:
		return ErrProjectNeverDoing
	default:
		return nil
	}
}

// CanEnterDoing reports whether n may be given the STORED status doing, given
// its direct children (D2, D7, D9).
//
// The type question comes first and is DoingRefusal's: a note or a habit has no
// column to be doing in, and a project never enters doing however few children
// it has. A freshly created project is therefore inert until it gains children;
// adding a child is how work becomes timeable.
//
// A parent that is not a leaf is excluded because doing means a timer and only
// leaves run one. Such a parent still RENDERS in the Doing column when a leaf
// underneath it is running — that is DeriveStatus's answer, not a stored status.
func (n Node) CanEnterDoing(children []Node) bool {
	if !n.Type.CanBeDoing() {
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
// no PMP activity (D4). A project at any size, a note, and a node with children
// that have a Kanban column are refused for the reasons CanEnterDoing gives.
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
		// Which rule refused is DoingRefusal's answer, not a second copy of the
		// type list here: above this line the type is known to have a column, so
		// the only type rule left to fire is the project one (D9), and it keeps
		// the sentinel the drag returns because it is the same rule.
		if reason := DoingRefusal(n.Type); reason != nil {
			return fmt.Errorf("domain: node %q: %w", n.ID, reason)
		}
		return invalid("node", "status",
			"only a leaf is stored as doing, and %q has children that have a kanban column", n.ID)
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
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"` // '' means "use the palette's default chip colour"
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
