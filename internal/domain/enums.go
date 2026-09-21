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
// a timer and is answered by NodeType.CanBeDoing and Node.CanEnterDoing
// (D2, D9).
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

// CanBeDoing reports whether the TYPE t permits the doing status at all — which
// is to say, permits a timer, since doing is what a timer means (D2, D7, D9).
//
// It is the bool form of DoingRefusal, defined in terms of it so that the two
// can never answer differently: task and bug yes, project no (D9), note and
// habit no (they have no column to be doing in). Whether a particular NODE of a
// permitted type may be doing depends on its children as well and is
// Node.CanEnterDoing's question.
//
// # Not the same rule as countsAsWork, which happens to admit the same types
//
// countsAsWork in derive.go excludes a project too, and for an unrelated reason:
// a project is the thing a progress bar is drawn FOR rather than a unit the bar
// measures. Two questions with the same answer today are still two questions —
// the precedent is HasColumn and HasDue, which a habit answers differently —
// and folding them together would mean a future change to one silently moving
// the other.
func (t NodeType) CanBeDoing() bool { return DoingRefusal(t) == nil }

// HasDue reports whether a node of type t may carry a due date at all (PLAN.md
// §4, "Behaviour by type").
//
// Exactly one type may not: a note is "no status, **no due**". A note is a piece
// of writing attached to something else — it is not work, it is not on the
// board, and there is nothing about it that can be late.
//
// A HABIT may. §4 says only that a habit "never appears in Kanban columns", and
// a column is not a date: a habit that has to be done by the end of the month is
// a sentence the specification nowhere forbids. This is why HasDue is a separate
// predicate from HasColumn rather than the same one under two names — the two
// questions genuinely have different answers, and D9's prose was corrected to
// stop claiming otherwise.
//
// An unknown type may not either, on HasColumn's reasoning: a type nobody
// recognises is not one this rule can vouch for. Rejecting it as a type is
// Node.Validate's job.
func (t NodeType) HasDue() bool {
	if t == NodeTypeNote {
		return false
	}
	return t.Valid()
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

// ---------------------------------------------------------------------------
// Appearance and language (D6).
//
// These three are the settings the user picks and the app persists — the
// `palette`, `theme` and `language` rows of the settings table. They live here,
// with every other enumerated value in this project, for the reason S2-05 makes
// an acceptance criterion: each allowed set must be defined EXACTLY ONCE in Go.
// The store seeds its defaults from these constants and the settings service
// validates against them, so "which palettes exist" cannot come to mean two
// different things in two packages — and the frontend builds its pickers from
// what the service reports rather than from a literal array of its own, which
// would be a third.

// Palette selects the design palette. design/tokens.css defines exactly these
// two, as [data-palette='...'] blocks.
type Palette string

// The two palettes. Aurora is the default (D6).
const (
	PaletteAurora Palette = "aurora"
	PaletteStudio Palette = "studio"
)

// DefaultPalette is the palette a fresh installation starts with (D6).
const DefaultPalette = PaletteAurora

// Palettes returns every valid Palette, in declaration order.
func Palettes() []Palette { return []Palette{PaletteAurora, PaletteStudio} }

// String returns the persisted representation of the palette.
func (p Palette) String() string { return string(p) }

// Valid reports whether p is one of the two known palettes.
func (p Palette) Valid() bool {
	switch p {
	case PaletteAurora, PaletteStudio:
		return true
	default:
		return false
	}
}

// Theme selects light or dark. It drives the `dark` class on <html>, and — so
// that the first frame is not the wrong colour — the GTK window background
// (D12).
type Theme string

// The two themes. Dark is the default (D6).
const (
	ThemeDark  Theme = "dark"
	ThemeLight Theme = "light"
)

// DefaultTheme is the theme a fresh installation starts with (D6).
const DefaultTheme = ThemeDark

// Themes returns every valid Theme, in declaration order.
func Themes() []Theme { return []Theme{ThemeDark, ThemeLight} }

// String returns the persisted representation of the theme.
func (t Theme) String() string { return string(t) }

// Valid reports whether t is dark or light.
func (t Theme) Valid() bool {
	switch t {
	case ThemeDark, ThemeLight:
		return true
	default:
		return false
	}
}

// Language selects the UI language. Both locale files are complete and shipped;
// there is no runtime download and no third language (PLAN.md §1).
type Language string

// The two languages. English is the default (D6).
const (
	LanguageEN Language = "en"
	LanguageRU Language = "ru"
)

// DefaultLanguage is the language a fresh installation starts with (D6).
const DefaultLanguage = LanguageEN

// Languages returns every valid Language, in declaration order.
func Languages() []Language { return []Language{LanguageEN, LanguageRU} }

// String returns the persisted representation of the language.
func (l Language) String() string { return string(l) }

// Valid reports whether l is en or ru.
func (l Language) Valid() bool {
	switch l {
	case LanguageEN, LanguageRU:
		return true
	default:
		return false
	}
}

// DefaultAccent is the seeded accent override: empty, meaning "use the
// palette's own --accent" (D6). It is a value the user may set to a colour and
// clear again, not one of an enumerated set, which is why it has a default and
// no type of its own.
const DefaultAccent = ""
