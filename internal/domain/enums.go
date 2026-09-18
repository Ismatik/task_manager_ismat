package domain

import "strconv"

// NodeType distinguishes the kinds of row stored in the single `nodes` tree.
// A subtask and a project differ by type and position, not by table.
type NodeType string

// The five node types. These strings are the values persisted in the database.
const (
	NodeTypeTask    NodeType = "task"
	NodeTypeProject NodeType = "project"
	NodeTypeHabit   NodeType = "habit"
	NodeTypeNote    NodeType = "note"
	NodeTypeBug     NodeType = "bug"
)

// NodeTypes returns every valid NodeType, in declaration order.
func NodeTypes() []NodeType {
	return []NodeType{
		NodeTypeTask,
		NodeTypeProject,
		NodeTypeHabit,
		NodeTypeNote,
		NodeTypeBug,
	}
}

// String returns the persisted representation of the node type.
func (t NodeType) String() string { return string(t) }

// Valid reports whether t is one of the five known node types.
func (t NodeType) Valid() bool {
	switch t {
	case NodeTypeTask, NodeTypeProject, NodeTypeHabit, NodeTypeNote, NodeTypeBug:
		return true
	default:
		return false
	}
}

// HasColumn reports whether a node of type t belongs in a Kanban column at all
// (PLAN.md §4, "Behaviour by type").
//
// Two types do not. A note has "no status, no due" and is excluded from parent
// derivation; a habit lives in the habit strip and "never appears in Kanban
// columns". Neither can be dragged to a column and neither may be STORED
// carrying one — which is the rule PlanCascade already applies to a note when
// it plans no change for one.
//
// Such a node still holds a status, because the column is NOT NULL: it is
// backlog, the schema's inert default, and it means "no column", not "in the
// Backlog column".
//
// task, project and bug do have columns. Whether a node that has a column may
// enter doing is a different and narrower question — that one is about running
// a timer and is answered by Node.CanEnterDoing (D2, D9).
//
// An unknown type has no column either: a type nobody recognises is not one
// this rule can vouch for. Rejecting it as a type is Node.Validate's job.
func (t NodeType) HasColumn() bool {
	switch t {
	case NodeTypeTask, NodeTypeProject, NodeTypeBug:
		return true
	default:
		return false
	}
}

// Status is a node's Kanban column. It is stored on leaves only: a parent's
// status is derived from its children and is never written (Stage 1, D2).
type Status string

// The five statuses, ordered from least to most advanced.
const (
	StatusBacklog Status = "backlog"
	StatusWeek    Status = "week"
	StatusToday   Status = "today"
	StatusDoing   Status = "doing"
	StatusDone    Status = "done"
)

// Statuses returns every valid Status, ordered from least to most advanced.
func Statuses() []Status {
	return []Status{
		StatusBacklog,
		StatusWeek,
		StatusToday,
		StatusDoing,
		StatusDone,
	}
}

// String returns the persisted representation of the status.
func (s Status) String() string { return string(s) }

// Valid reports whether s is one of the five known statuses.
func (s Status) Valid() bool {
	switch s {
	case StatusBacklog, StatusWeek, StatusToday, StatusDoing, StatusDone:
		return true
	default:
		return false
	}
}

// DueSource records the provenance of a node's due date (D1). Moving a card
// between columns writes an auto due date; any user edit of the due date makes
// it manual. Only an auto due date is cleared by a move back to Backlog.
type DueSource string

// The two due-date provenances. There are exactly two (D1).
const (
	DueSourceManual DueSource = "manual"
	DueSourceAuto   DueSource = "auto"
)

// DueSources returns both valid DueSource values, in declaration order.
func DueSources() []DueSource {
	return []DueSource{DueSourceManual, DueSourceAuto}
}

// String returns the persisted representation of the due source.
func (d DueSource) String() string { return string(d) }

// Valid reports whether d is manual or auto.
func (d DueSource) Valid() bool {
	switch d {
	case DueSourceManual, DueSourceAuto:
		return true
	default:
		return false
	}
}

// Activity is the PMP KIT "Деятельность" field of a node (D4). The seven values
// are fixed by design/SKILL.md and are Russian by design: they are copied
// verbatim into the PMP timelog form, so they are data, not UI strings, and are
// never translated.
type Activity string

// The seven activities. Exactly these, per D4 and design/SKILL.md.
const (
	ActivityDevelopment   Activity = "Разработка"
	ActivityAnalysis      Activity = "Анализ"
	ActivityTesting       Activity = "Тестирование"
	ActivityDocumentation Activity = "Документация"
	ActivityMeeting       Activity = "Совещание"
	ActivityApproval      Activity = "Согласование"
	ActivityManagement    Activity = "Управление проектом"
)

// Activities returns all seven valid activities, in declaration order.
func Activities() []Activity {
	return []Activity{
		ActivityDevelopment,
		ActivityAnalysis,
		ActivityTesting,
		ActivityDocumentation,
		ActivityMeeting,
		ActivityApproval,
		ActivityManagement,
	}
}

// String returns the persisted representation of the activity.
func (a Activity) String() string { return string(a) }

// Valid reports whether a is one of the seven known activities.
func (a Activity) Valid() bool {
	switch a {
	case ActivityDevelopment,
		ActivityAnalysis,
		ActivityTesting,
		ActivityDocumentation,
		ActivityMeeting,
		ActivityApproval,
		ActivityManagement:
		return true
	default:
		return false
	}
}

// Priority is a node's priority, 1..4. Lower is more urgent.
type Priority int

// The four priorities.
const (
	Priority1 Priority = 1
	Priority2 Priority = 2
	Priority3 Priority = 3
	Priority4 Priority = 4
)

// Priorities returns every valid Priority, from most to least urgent.
func Priorities() []Priority {
	return []Priority{Priority1, Priority2, Priority3, Priority4}
}

// String returns the decimal representation of the priority, which is also how
// it is persisted.
func (p Priority) String() string { return strconv.Itoa(int(p)) }

// Valid reports whether p is within 1..4.
func (p Priority) Valid() bool { return p >= Priority1 && p <= Priority4 }
