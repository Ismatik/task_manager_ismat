package domain_test

import (
	"errors"
	"testing"
	"time"

	"nexus/internal/domain"
)

// fixedNow is 2026-09-18 — a Friday, the day the project started, which is the
// live edge case for the "upcoming Friday" rule (D1).
var fixedNow = time.Date(2026, time.September, 18, 10, 30, 0, 0, time.UTC)

func ptr[T any](v T) *T { return &v }

// validNode returns a node that passes Validate, so that each subtest can break
// exactly one field and be sure that field is the reason it failed.
func validNode() domain.Node {
	return domain.Node{
		ID:            "n1",
		ParentID:      nil,
		Type:          domain.NodeTypeTask,
		Title:         "Write the derivation tests",
		DescriptionMD: "",
		Status:        domain.StatusBacklog,
		Due:           nil,
		DueSource:     domain.DueSourceManual,
		Priority:      domain.Priority4,
		CreatedAt:     fixedNow,
		UpdatedAt:     fixedNow,
	}
}

// assertValidationError checks that err is a *domain.ValidationError naming
// field, and that it matches errors.Is(err, domain.ErrInvalid).
func assertValidationError(t *testing.T, err error, entity, field string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Validate() = nil, want an error naming %s.%s", entity, field)
	}
	if !errors.Is(err, domain.ErrInvalid) {
		t.Errorf("errors.Is(err, ErrInvalid) = false for %v", err)
	}

	var ve *domain.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("errors.As(%v, *ValidationError) = false", err)
	}
	if ve.Entity != entity {
		t.Errorf("Entity = %q, want %q (err: %v)", ve.Entity, entity, err)
	}
	if ve.Field != field {
		t.Errorf("Field = %q, want %q (err: %v)", ve.Field, field, err)
	}
}

func TestNodeValidateRejectsEachField(t *testing.T) {
	tests := []struct {
		name  string
		field string
		mut   func(*domain.Node)
	}{
		{"empty id", "id", func(n *domain.Node) { n.ID = "" }},
		{"blank id", "id", func(n *domain.Node) { n.ID = "   " }},
		{"empty parent id", "parent_id", func(n *domain.Node) { n.ParentID = ptr("") }},
		{"self parent", "parent_id", func(n *domain.Node) { n.ParentID = ptr("n1") }},
		{"unknown type", "type", func(n *domain.Node) { n.Type = domain.NodeType("epic") }},
		{"empty title", "title", func(n *domain.Node) { n.Title = "" }},
		{"blank title", "title", func(n *domain.Node) { n.Title = "\t\n " }},
		{"unknown status", "status", func(n *domain.Node) { n.Status = domain.Status("Done") }},
		{"impossible due date", "due", func(n *domain.Node) {
			n.Due = ptr(domain.NewDate(2026, time.February, 30))
		}},
		{"unknown due source", "due_source", func(n *domain.Node) {
			n.DueSource = domain.DueSource("derived")
		}},
		{"priority zero", "priority", func(n *domain.Node) { n.Priority = domain.Priority(0) }},
		{"priority five", "priority", func(n *domain.Node) { n.Priority = domain.Priority(5) }},
		{"negative estimate", "estimate_min", func(n *domain.Node) { n.EstimateMin = ptr(-1) }},
		{"blank recurrence", "recurrence", func(n *domain.Node) { n.Recurrence = ptr("  ") }},
		{"habit without recurrence", "recurrence", func(n *domain.Node) {
			n.Type = domain.NodeTypeHabit
			n.Recurrence = nil
		}},
		{"unknown activity", "activity", func(n *domain.Node) {
			n.Activity = ptr(domain.Activity("Обучение"))
		}},
		{"missing created_at", "created_at", func(n *domain.Node) { n.CreatedAt = time.Time{} }},
		{"missing updated_at", "updated_at", func(n *domain.Node) { n.UpdatedAt = time.Time{} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := validNode()
			tt.mut(&n)
			assertValidationError(t, n.Validate(), "node", tt.field)
		})
	}
}

// The message must carry the entity, the column name and the reason, because
// that string is what ends up in a log line with no struct around it.
func TestValidationErrorMessageNamesEntityAndField(t *testing.T) {
	n := validNode()
	n.Priority = domain.Priority(9)

	err := n.Validate()
	if err == nil {
		t.Fatal("Validate() = nil, want an error")
	}
	if got, want := err.Error(), "domain: node.priority: 9 is outside 1..4"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestNodeValidateAcceptsValidNodes(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*domain.Node)
	}{
		{"the baseline task", func(*domain.Node) {}},
		{"a child node", func(n *domain.Node) { n.ParentID = ptr("root") }},
		{"a real due date", func(n *domain.Node) {
			n.Due = ptr(domain.NewDate(2026, time.September, 18))
		}},
		{"a leap day due date", func(n *domain.Node) {
			n.Due = ptr(domain.NewDate(2028, time.February, 29))
		}},
		{"zero estimate", func(n *domain.Node) { n.EstimateMin = ptr(0) }},
		{"an activity", func(n *domain.Node) { n.Activity = ptr(domain.ActivityAnalysis) }},
		{"a habit with a recurrence", func(n *domain.Node) {
			n.Type = domain.NodeTypeHabit
			n.Recurrence = ptr("FREQ=DAILY")
		}},
		{"a completed, archived node", func(n *domain.Node) {
			n.Status = domain.StatusDone
			n.CompletedAt = ptr(fixedNow)
			n.ArchivedAt = ptr(fixedNow)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := validNode()
			tt.mut(&n)
			if err := n.Validate(); err != nil {
				t.Errorf("Validate() = %v, want nil", err)
			}
		})
	}
}

// D2: a node with no children, or whose children are all notes, behaves as a
// leaf. The all-notes case is the one the decision calls out by name.
func TestNodeIsLeaf(t *testing.T) {
	child := func(tp domain.NodeType) domain.Node {
		n := validNode()
		n.Type = tp
		return n
	}

	tests := []struct {
		name     string
		children []domain.Node
		want     bool
	}{
		{"no children", nil, true},
		{"empty slice of children", []domain.Node{}, true},
		{"one task child", []domain.Node{child(domain.NodeTypeTask)}, false},
		{"one note child", []domain.Node{child(domain.NodeTypeNote)}, true},
		{"several notes only", []domain.Node{
			child(domain.NodeTypeNote), child(domain.NodeTypeNote), child(domain.NodeTypeNote),
		}, true},
		{"mixed note and task", []domain.Node{
			child(domain.NodeTypeNote), child(domain.NodeTypeTask),
		}, false},
		{"mixed task and note", []domain.Node{
			child(domain.NodeTypeTask), child(domain.NodeTypeNote),
		}, false},
		{"one project child", []domain.Node{child(domain.NodeTypeProject)}, false},
		{"one habit child", []domain.Node{child(domain.NodeTypeHabit)}, false},
		{"one bug child", []domain.Node{child(domain.NodeTypeBug)}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := validNode()
			parent.Type = domain.NodeTypeProject
			if got := parent.IsLeaf(tt.children); got != tt.want {
				t.Errorf("IsLeaf(%d children) = %v, want %v", len(tt.children), got, tt.want)
			}
		})
	}
}

// D4: the defaults by node type, habit having none.
func TestDefaultActivity(t *testing.T) {
	tests := []struct {
		name string
		in   domain.NodeType
		want *domain.Activity
	}{
		{"task", domain.NodeTypeTask, ptr(domain.ActivityDevelopment)},
		{"bug", domain.NodeTypeBug, ptr(domain.ActivityTesting)},
		{"project", domain.NodeTypeProject, ptr(domain.ActivityManagement)},
		{"note", domain.NodeTypeNote, ptr(domain.ActivityDocumentation)},
		{"habit has no default", domain.NodeTypeHabit, nil},
		{"unknown type has no default", domain.NodeType("epic"), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DefaultActivity(tt.in)
			switch {
			case tt.want == nil && got != nil:
				t.Fatalf("DefaultActivity(%q) = %q, want nil", tt.in, *got)
			case tt.want == nil:
				return
			case got == nil:
				t.Fatalf("DefaultActivity(%q) = nil, want %q", tt.in, *tt.want)
			case *got != *tt.want:
				t.Errorf("DefaultActivity(%q) = %q, want %q", tt.in, *got, *tt.want)
			}
			if !got.Valid() {
				t.Errorf("DefaultActivity(%q) = %q, which is not a valid Activity", tt.in, *got)
			}
		})
	}
}

// DefaultActivity must hand back a fresh pointer each time: a shared one would
// let a caller that edits one node's activity edit every node's.
func TestDefaultActivityReturnsAnIndependentPointer(t *testing.T) {
	a := domain.DefaultActivity(domain.NodeTypeTask)
	b := domain.DefaultActivity(domain.NodeTypeTask)
	if a == b {
		t.Fatal("DefaultActivity returned the same pointer twice; callers could alias each other's activity")
	}
	*a = domain.ActivityMeeting
	if *b != domain.ActivityDevelopment {
		t.Errorf("mutating one result changed the other: %q", *b)
	}
}

func TestTagValidate(t *testing.T) {
	valid := domain.Tag{ID: "t1", Name: "deep work", Color: "#7F5AF0"}

	t.Run("a valid tag", func(t *testing.T) {
		if err := valid.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})
	t.Run("an empty colour means the palette default", func(t *testing.T) {
		tag := valid
		tag.Color = ""
		if err := tag.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})
	t.Run("lower case hex is accepted", func(t *testing.T) {
		tag := valid
		tag.Color = "#7f5af0"
		if err := tag.Validate(); err != nil {
			t.Errorf("Validate() = %v, want nil", err)
		}
	})

	bad := []struct {
		name  string
		field string
		mut   func(*domain.Tag)
	}{
		{"empty id", "id", func(tg *domain.Tag) { tg.ID = "" }},
		{"blank name", "name", func(tg *domain.Tag) { tg.Name = " " }},
		{"colour without a hash", "color", func(tg *domain.Tag) { tg.Color = "7F5AF0" }},
		{"short colour", "color", func(tg *domain.Tag) { tg.Color = "#fff" }},
		{"non hex digits", "color", func(tg *domain.Tag) { tg.Color = "#GGGGGG" }},
		{"named colour", "color", func(tg *domain.Tag) { tg.Color = "purple" }},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			tag := valid
			tt.mut(&tag)
			assertValidationError(t, tag.Validate(), "tag", tt.field)
		})
	}
}

func TestDateStringAndParse(t *testing.T) {
	tests := []struct {
		name string
		in   domain.Date
		want string
	}{
		{"the project start", domain.NewDate(2026, time.September, 18), "2026-09-18"},
		{"single digit month and day", domain.NewDate(2026, time.January, 2), "2026-01-02"},
		{"a leap day", domain.NewDate(2028, time.February, 29), "2028-02-29"},
		{"the last day of a year", domain.NewDate(2026, time.December, 31), "2026-12-31"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
			back, err := domain.ParseDate(tt.want)
			if err != nil {
				t.Fatalf("ParseDate(%q) = %v", tt.want, err)
			}
			if back != tt.in {
				t.Errorf("ParseDate(%q) = %v, want %v", tt.want, back, tt.in)
			}
		})
	}
}

func TestParseDateRejects(t *testing.T) {
	tests := []string{
		"",
		"2026-9-18",
		"18/09/2026",
		"2026-02-30",
		"2026-13-01",
		"2026-09-18T10:00:00Z",
		"yesterday",
	}

	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			got, err := domain.ParseDate(in)
			if err == nil {
				t.Fatalf("ParseDate(%q) = %v, want an error", in, got)
			}
			if !errors.Is(err, domain.ErrInvalid) {
				t.Errorf("errors.Is(err, ErrInvalid) = false for %v", err)
			}
		})
	}
}

func TestDateOfKeepsTheLocationsCalendarDay(t *testing.T) {
	// 23:30 on the 18th in a zone three hours ahead of UTC is still the 18th
	// there, even though it is the 20:30 of the 18th in UTC. Converting to UTC
	// first is the classic off-by-one-day bug; DateOf must not do it.
	east := time.FixedZone("UTC+3", 3*60*60)
	got := domain.DateOf(time.Date(2026, time.September, 18, 23, 30, 0, 0, east))
	if want := domain.NewDate(2026, time.September, 18); got != want {
		t.Errorf("DateOf(23:30 UTC+3) = %v, want %v", got, want)
	}

	// And the same instant one hour later really is the 19th there.
	got = domain.DateOf(time.Date(2026, time.September, 19, 0, 30, 0, 0, east))
	if want := domain.NewDate(2026, time.September, 19); got != want {
		t.Errorf("DateOf(00:30 UTC+3) = %v, want %v", got, want)
	}
}

func TestDateValidAndIsZero(t *testing.T) {
	tests := []struct {
		name      string
		in        domain.Date
		wantValid bool
		wantZero  bool
	}{
		{"a real date", domain.NewDate(2026, time.September, 18), true, false},
		{"a leap day in a leap year", domain.NewDate(2028, time.February, 29), true, false},
		{"a leap day in a common year", domain.NewDate(2026, time.February, 29), false, false},
		{"the 31st of a 30-day month", domain.NewDate(2026, time.April, 31), false, false},
		{"month zero", domain.NewDate(2026, time.Month(0), 1), false, false},
		{"month thirteen", domain.NewDate(2026, time.Month(13), 1), false, false},
		{"day zero", domain.NewDate(2026, time.September, 0), false, false},
		{"the zero date", domain.Date{}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.Valid(); got != tt.wantValid {
				t.Errorf("Valid() = %v, want %v", got, tt.wantValid)
			}
			if got := tt.in.IsZero(); got != tt.wantZero {
				t.Errorf("IsZero() = %v, want %v", got, tt.wantZero)
			}
		})
	}
}

func TestDateOrdering(t *testing.T) {
	tests := []struct {
		name        string
		a, b        domain.Date
		wantCompare int
	}{
		{"same day", domain.NewDate(2026, time.September, 18), domain.NewDate(2026, time.September, 18), 0},
		{"earlier day", domain.NewDate(2026, time.September, 17), domain.NewDate(2026, time.September, 18), -1},
		{"later day", domain.NewDate(2026, time.September, 19), domain.NewDate(2026, time.September, 18), 1},
		{"earlier month", domain.NewDate(2026, time.August, 31), domain.NewDate(2026, time.September, 1), -1},
		{"later month", domain.NewDate(2026, time.October, 1), domain.NewDate(2026, time.September, 30), 1},
		{"earlier year", domain.NewDate(2025, time.December, 31), domain.NewDate(2026, time.January, 1), -1},
		{"later year", domain.NewDate(2027, time.January, 1), domain.NewDate(2026, time.December, 31), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Compare(tt.b); got != tt.wantCompare {
				t.Errorf("Compare() = %d, want %d", got, tt.wantCompare)
			}
			if got, want := tt.a.Before(tt.b), tt.wantCompare < 0; got != want {
				t.Errorf("Before() = %v, want %v", got, want)
			}
			if got, want := tt.a.After(tt.b), tt.wantCompare > 0; got != want {
				t.Errorf("After() = %v, want %v", got, want)
			}
			if got, want := tt.a.Equal(tt.b), tt.wantCompare == 0; got != want {
				t.Errorf("Equal() = %v, want %v", got, want)
			}
		})
	}
}

func TestDateAddDaysAndDaysUntil(t *testing.T) {
	tests := []struct {
		name string
		from domain.Date
		n    int
		want domain.Date
	}{
		{"no move", domain.NewDate(2026, time.September, 18), 0, domain.NewDate(2026, time.September, 18)},
		{"next day", domain.NewDate(2026, time.September, 18), 1, domain.NewDate(2026, time.September, 19)},
		{"previous day", domain.NewDate(2026, time.September, 18), -1, domain.NewDate(2026, time.September, 17)},
		{"across a month end", domain.NewDate(2026, time.January, 31), 1, domain.NewDate(2026, time.February, 1)},
		{"across a year end", domain.NewDate(2026, time.December, 31), 1, domain.NewDate(2027, time.January, 1)},
		{"backwards across a year end", domain.NewDate(2027, time.January, 1), -1, domain.NewDate(2026, time.December, 31)},
		{"across a leap day", domain.NewDate(2028, time.February, 28), 1, domain.NewDate(2028, time.February, 29)},
		{"a whole common year", domain.NewDate(2026, time.September, 18), 365, domain.NewDate(2027, time.September, 18)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.AddDays(tt.n)
			if got != tt.want {
				t.Fatalf("AddDays(%d) = %v, want %v", tt.n, got, tt.want)
			}
			if d := tt.from.DaysUntil(got); d != tt.n {
				t.Errorf("DaysUntil() = %d, want %d", d, tt.n)
			}
		})
	}
}

func TestDateWeekdayAndTime(t *testing.T) {
	tests := []struct {
		in   domain.Date
		want time.Weekday
	}{
		{domain.NewDate(2026, time.September, 14), time.Monday},
		{domain.NewDate(2026, time.September, 15), time.Tuesday},
		{domain.NewDate(2026, time.September, 16), time.Wednesday},
		{domain.NewDate(2026, time.September, 17), time.Thursday},
		{domain.NewDate(2026, time.September, 18), time.Friday},
		{domain.NewDate(2026, time.September, 19), time.Saturday},
		{domain.NewDate(2026, time.September, 20), time.Sunday},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			if got := tt.in.Weekday(); got != tt.want {
				t.Errorf("%v.Weekday() = %v, want %v", tt.in, got, tt.want)
			}
			ts := tt.in.Time()
			if ts.Location() != time.UTC {
				t.Errorf("Time() location = %v, want UTC", ts.Location())
			}
			if h, m, s := ts.Clock(); h|m|s != 0 {
				t.Errorf("Time() = %v, want midnight", ts)
			}
		})
	}
}

// D9: type beats the leaf rule. A project never enters doing and never runs a
// timer, and neither does a note, a habit or a node with non-note children.
func TestNodeCanEnterDoingAndCanStartTimer(t *testing.T) {
	child := func(tp domain.NodeType) domain.Node {
		n := validNode()
		n.Type = tp
		return n
	}
	node := func(tp domain.NodeType) domain.Node {
		n := validNode()
		n.Type = tp
		return n
	}

	tests := []struct {
		name      string
		node      domain.Node
		children  []domain.Node
		wantDoing bool
		wantTimer bool
	}{
		{"a childless task is both", node(domain.NodeTypeTask), nil, true, true},
		{"a task with only note children is both", node(domain.NodeTypeTask),
			[]domain.Node{child(domain.NodeTypeNote)}, true, true},
		{"a task with a task child is neither", node(domain.NodeTypeTask),
			[]domain.Node{child(domain.NodeTypeTask)}, false, false},
		{"a childless bug is both", node(domain.NodeTypeBug), nil, true, true},
		{"D9: an EMPTY project is neither", node(domain.NodeTypeProject), nil, false, false},
		{"D9: a project with only note children is neither", node(domain.NodeTypeProject),
			[]domain.Node{child(domain.NodeTypeNote)}, false, false},
		{"D9: a project with children is neither", node(domain.NodeTypeProject),
			[]domain.Node{child(domain.NodeTypeTask)}, false, false},
		{"a note is neither", node(domain.NodeTypeNote), nil, false, false},
		{"a habit may hold doing but never a timer", node(domain.NodeTypeHabit), nil, true, false},
		{"a habit with children holds neither", node(domain.NodeTypeHabit),
			[]domain.Node{child(domain.NodeTypeTask)}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.CanEnterDoing(tt.children); got != tt.wantDoing {
				t.Errorf("CanEnterDoing() = %v, want %v", got, tt.wantDoing)
			}
			if got := tt.node.CanStartTimer(tt.children); got != tt.wantTimer {
				t.Errorf("CanStartTimer() = %v, want %v", got, tt.wantTimer)
			}
		})
	}
}

// CheckStatus is the one gate every door into a stored status goes through, so
// the table is the whole type x status matrix plus the child shapes that change
// the answer.
func TestNodeCheckStatus(t *testing.T) {
	child := func(tp domain.NodeType) domain.Node {
		n := validNode()
		n.ID = "c1"
		n.Type = tp
		return n
	}
	node := func(tp domain.NodeType, s domain.Status) domain.Node {
		n := validNode()
		n.Type = tp
		n.Status = s
		return n
	}

	tests := []struct {
		name     string
		node     domain.Node
		children []domain.Node
		wantErr  error // nil means the combination is legal
	}{
		// The types that do have a column may hold any of the five, doing
		// included, as long as they are leaves.
		{"a leaf task may hold backlog", node(domain.NodeTypeTask, domain.StatusBacklog), nil, nil},
		{"a leaf task may hold week", node(domain.NodeTypeTask, domain.StatusWeek), nil, nil},
		{"a leaf task may hold today", node(domain.NodeTypeTask, domain.StatusToday), nil, nil},
		{"a leaf task may hold doing", node(domain.NodeTypeTask, domain.StatusDoing), nil, nil},
		{"a leaf task may hold done", node(domain.NodeTypeTask, domain.StatusDone), nil, nil},
		{"a bug is a task for this purpose", node(domain.NodeTypeBug, domain.StatusDoing), nil, nil},
		{"a task whose children are all notes is still a leaf", node(domain.NodeTypeTask, domain.StatusDoing),
			[]domain.Node{child(domain.NodeTypeNote)}, nil},

		// D2: doing means a timer, and only a leaf runs one.
		{"a task with a task child may not hold doing", node(domain.NodeTypeTask, domain.StatusDoing),
			[]domain.Node{child(domain.NodeTypeTask)}, domain.ErrInvalid},
		{"a task with a task child may still hold today", node(domain.NodeTypeTask, domain.StatusToday),
			[]domain.Node{child(domain.NodeTypeTask)}, nil},

		// D9: a project never enters doing, empty or not — the same sentinel
		// the drag returns.
		{"D9: an EMPTY project may not hold doing", node(domain.NodeTypeProject, domain.StatusDoing),
			nil, domain.ErrProjectNeverDoing},
		{"D9: a project with note children may not hold doing", node(domain.NodeTypeProject, domain.StatusDoing),
			[]domain.Node{child(domain.NodeTypeNote)}, domain.ErrProjectNeverDoing},
		{"D9: a project with children may not hold doing", node(domain.NodeTypeProject, domain.StatusDoing),
			[]domain.Node{child(domain.NodeTypeTask)}, domain.ErrProjectNeverDoing},
		{"a project may hold today", node(domain.NodeTypeProject, domain.StatusToday), nil, nil},
		{"a project may hold done", node(domain.NodeTypeProject, domain.StatusDone), nil, nil},

		// PLAN.md §4: a note has no status and a habit is never in a column.
		// Backlog is the inert NOT NULL default and is the only value either
		// may carry.
		{"a note may hold the default backlog", node(domain.NodeTypeNote, domain.StatusBacklog), nil, nil},
		{"a note may not hold week", node(domain.NodeTypeNote, domain.StatusWeek), nil, domain.ErrTypeHasNoColumn},
		{"a note may not hold today", node(domain.NodeTypeNote, domain.StatusToday), nil, domain.ErrTypeHasNoColumn},
		{"a note may not hold doing", node(domain.NodeTypeNote, domain.StatusDoing), nil, domain.ErrTypeHasNoColumn},
		{"a note may not hold done", node(domain.NodeTypeNote, domain.StatusDone), nil, domain.ErrTypeHasNoColumn},
		{"a habit may hold the default backlog", node(domain.NodeTypeHabit, domain.StatusBacklog), nil, nil},
		{"a habit may not hold week", node(domain.NodeTypeHabit, domain.StatusWeek), nil, domain.ErrTypeHasNoColumn},
		{"a habit may not hold today", node(domain.NodeTypeHabit, domain.StatusToday), nil, domain.ErrTypeHasNoColumn},
		{"a habit may not hold doing, which CanEnterDoing alone would allow",
			node(domain.NodeTypeHabit, domain.StatusDoing), nil, domain.ErrTypeHasNoColumn},
		{"a habit may not hold done", node(domain.NodeTypeHabit, domain.StatusDone), nil, domain.ErrTypeHasNoColumn},

		// An unrecognised type is refused the same way; Validate is what names
		// the type itself as the problem.
		{"an unknown type has no column either", node(domain.NodeType("epic"), domain.StatusToday),
			nil, domain.ErrTypeHasNoColumn},
		{"an unknown type carrying backlog passes this gate", node(domain.NodeType("epic"), domain.StatusBacklog),
			nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.node.CheckStatus(tt.children)
			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("CheckStatus() = %v, want nil", err)
			case tt.wantErr == nil:
			case !errors.Is(err, tt.wantErr):
				t.Fatalf("CheckStatus() = %v, want an error matching %v", err, tt.wantErr)
			}
		})
	}
}

// The non-leaf refusal names the field, so that a caller can report which
// column of which row the rule objected to.
func TestNodeCheckStatusNamesTheStatusField(t *testing.T) {
	parent := validNode()
	parent.Status = domain.StatusDoing
	child := validNode()
	child.ID = "c1"

	assertValidationError(t, parent.CheckStatus([]domain.Node{child}), "node", "status")
}
