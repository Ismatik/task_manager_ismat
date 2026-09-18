package domain_test

import (
	"testing"

	"nexus/internal/domain"
)

func TestNodeTypeValid(t *testing.T) {
	tests := []struct {
		name string
		in   domain.NodeType
		want bool
	}{
		{"task", domain.NodeTypeTask, true},
		{"project", domain.NodeTypeProject, true},
		{"habit", domain.NodeTypeHabit, true},
		{"note", domain.NodeTypeNote, true},
		{"bug", domain.NodeTypeBug, true},
		{"empty", domain.NodeType(""), false},
		{"unknown", domain.NodeType("epic"), false},
		{"wrong case", domain.NodeType("Task"), false},
		{"padded", domain.NodeType(" task"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.want {
				t.Errorf("NodeType(%q).Valid() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNodeTypeString(t *testing.T) {
	tests := []struct {
		in   domain.NodeType
		want string
	}{
		{domain.NodeTypeTask, "task"},
		{domain.NodeTypeProject, "project"},
		{domain.NodeTypeHabit, "habit"},
		{domain.NodeTypeNote, "note"},
		{domain.NodeTypeBug, "bug"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNodeTypesIsTheCompleteSet(t *testing.T) {
	want := []string{"task", "project", "habit", "note", "bug"}

	got := domain.NodeTypes()
	if len(got) != len(want) {
		t.Fatalf("NodeTypes() has %d values, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].String() != w {
			t.Errorf("NodeTypes()[%d] = %q, want %q", i, got[i], w)
		}
		if !got[i].Valid() {
			t.Errorf("NodeTypes()[%d] = %q is not Valid()", i, got[i])
		}
	}
}

// PLAN.md §4: a note has no status and no due, and a habit never appears in a
// Kanban column. Everything else does have one.
func TestNodeTypeHasColumn(t *testing.T) {
	tests := []struct {
		name string
		in   domain.NodeType
		want bool
	}{
		{"a task has a column", domain.NodeTypeTask, true},
		{"a project has a column — D9 is about doing, not about having one", domain.NodeTypeProject, true},
		{"a bug has a column", domain.NodeTypeBug, true},
		{"a note has none: no status, no due", domain.NodeTypeNote, false},
		{"a habit has none: the habit strip, never a column", domain.NodeTypeHabit, false},
		{"an unknown type has none", domain.NodeType("epic"), false},
		{"the empty type has none", domain.NodeType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.HasColumn(); got != tt.want {
				t.Errorf("NodeType(%q).HasColumn() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// PLAN.md §4: a note is "no status, no due". Every other known type may carry a
// date — a habit included, because §4 denies it a COLUMN, not a date.
func TestNodeTypeHasDue(t *testing.T) {
	tests := []struct {
		name string
		in   domain.NodeType
		want bool
	}{
		{"a task may be due", domain.NodeTypeTask, true},
		{"a project may be due", domain.NodeTypeProject, true},
		{"a bug may be due", domain.NodeTypeBug, true},
		{"a habit may be due: it has no column, which is a different rule", domain.NodeTypeHabit, true},
		{"a note may not: no status, no due", domain.NodeTypeNote, false},
		{"an unknown type may not", domain.NodeType("epic"), false},
		{"the empty type may not", domain.NodeType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.HasDue(); got != tt.want {
				t.Errorf("NodeType(%q).HasDue() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name string
		in   domain.Status
		want bool
	}{
		{"backlog", domain.StatusBacklog, true},
		{"week", domain.StatusWeek, true},
		{"today", domain.StatusToday, true},
		{"doing", domain.StatusDoing, true},
		{"done", domain.StatusDone, true},
		{"empty", domain.Status(""), false},
		{"unknown", domain.Status("archived"), false},
		{"wrong case", domain.Status("Done"), false},
		{"this week", domain.Status("this_week"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.want {
				t.Errorf("Status(%q).Valid() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestStatusesAreOrderedLeastToMostAdvanced(t *testing.T) {
	want := []string{"backlog", "week", "today", "doing", "done"}

	got := domain.Statuses()
	if len(got) != len(want) {
		t.Fatalf("Statuses() has %d values, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].String() != w {
			t.Errorf("Statuses()[%d] = %q, want %q", i, got[i], w)
		}
		if !got[i].Valid() {
			t.Errorf("Statuses()[%d] = %q is not Valid()", i, got[i])
		}
	}
}

// D1: due_source has exactly two values.
func TestDueSourceValid(t *testing.T) {
	tests := []struct {
		name string
		in   domain.DueSource
		want bool
	}{
		{"manual", domain.DueSourceManual, true},
		{"auto", domain.DueSourceAuto, true},
		{"empty", domain.DueSource(""), false},
		{"legacy boolean", domain.DueSource("true"), false},
		{"wrong case", domain.DueSource("Auto"), false},
		{"unknown", domain.DueSource("derived"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.want {
				t.Errorf("DueSource(%q).Valid() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestDueSourcesHasExactlyManualAndAuto(t *testing.T) {
	want := []string{"manual", "auto"}

	got := domain.DueSources()
	if len(got) != len(want) {
		t.Fatalf("DueSources() has %d values, want exactly %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].String() != w {
			t.Errorf("DueSources()[%d] = %q, want %q", i, got[i], w)
		}
	}
}

// D4: the seven PMP activities, Russian by design.
func TestActivityValid(t *testing.T) {
	tests := []struct {
		name string
		in   domain.Activity
		want bool
	}{
		{"development", domain.ActivityDevelopment, true},
		{"analysis", domain.ActivityAnalysis, true},
		{"testing", domain.ActivityTesting, true},
		{"documentation", domain.ActivityDocumentation, true},
		{"meeting", domain.ActivityMeeting, true},
		{"approval", domain.ActivityApproval, true},
		{"management", domain.ActivityManagement, true},
		{"empty", domain.Activity(""), false},
		{"english translation", domain.Activity("Development"), false},
		{"lower case", domain.Activity("разработка"), false},
		{"unknown", domain.Activity("Обучение"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.want {
				t.Errorf("Activity(%q).Valid() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestActivitiesHasExactlyTheSevenD4Values(t *testing.T) {
	want := []string{
		"Разработка",
		"Анализ",
		"Тестирование",
		"Документация",
		"Совещание",
		"Согласование",
		"Управление проектом",
	}

	got := domain.Activities()
	if len(got) != 7 {
		t.Fatalf("Activities() has %d values, want exactly 7: %v", len(got), got)
	}
	for i, w := range want {
		if got[i].String() != w {
			t.Errorf("Activities()[%d] = %q, want %q", i, got[i], w)
		}
		if !got[i].Valid() {
			t.Errorf("Activities()[%d] = %q is not Valid()", i, got[i])
		}
	}
}

func TestPriorityValid(t *testing.T) {
	tests := []struct {
		name string
		in   domain.Priority
		want bool
	}{
		{"1", domain.Priority1, true},
		{"2", domain.Priority2, true},
		{"3", domain.Priority3, true},
		{"4", domain.Priority4, true},
		{"zero", domain.Priority(0), false},
		{"negative", domain.Priority(-1), false},
		{"five", domain.Priority(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.want {
				t.Errorf("Priority(%d).Valid() = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		in   domain.Priority
		want string
	}{
		{domain.Priority1, "1"},
		{domain.Priority2, "2"},
		{domain.Priority3, "3"},
		{domain.Priority4, "4"},
		{domain.Priority(9), "9"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("Priority(%d).String() = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPrioritiesRunFromOneToFour(t *testing.T) {
	got := domain.Priorities()
	if len(got) != 4 {
		t.Fatalf("Priorities() has %d values, want 4: %v", len(got), got)
	}
	for i, p := range got {
		if int(p) != i+1 {
			t.Errorf("Priorities()[%d] = %d, want %d", i, p, i+1)
		}
		if !p.Valid() {
			t.Errorf("Priorities()[%d] = %d is not Valid()", i, p)
		}
	}
}
