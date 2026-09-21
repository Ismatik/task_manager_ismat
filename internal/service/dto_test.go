package service_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"nexus/internal/domain"
	"nexus/internal/service"
)

// THE wire contract (S2-02).
//
// This string is the contract. Every key the frontend binds to is visible in
// this file, so renaming one is a diff hunk a reviewer can read rather than a
// change that compiles in Go and fails at runtime in TypeScript.
//
// What it pins, deliberately:
//
//   - lowerCamelCase everywhere, no exceptions: descriptionMd, dueSource,
//     estimateMin, sortOrder, completedAt, archivedAt, entryId, elapsedSeconds;
//   - a due date is the string "2026-10-02", NOT {"Year":...,"Month":...};
//   - instants are RFC 3339;
//   - an absent optional is null, and a list is never null.
const goldenColumn = `{
  "status": "today",
  "nodes": [
    {
      "node": {
        "id": "n1",
        "parentId": null,
        "type": "project",
        "title": "Ship the board",
        "descriptionMd": "# Ship it",
        "status": "backlog",
        "due": "2026-10-02",
        "dueSource": "auto",
        "priority": 2,
        "estimateMin": 90,
        "recurrence": null,
        "activity": "Управление проектом",
        "sortOrder": 3,
        "createdAt": "2026-09-01T08:00:00Z",
        "updatedAt": "2026-09-21T09:00:00Z",
        "completedAt": null,
        "archivedAt": null
      },
      "status": "today",
      "progress": {
        "done": 1,
        "total": 2,
        "percent": 50,
        "defined": true
      },
      "overdue": false,
      "isLeaf": false,
      "tags": [
        {
          "id": "t1",
          "name": "work",
          "color": "#38bdf8"
        }
      ],
      "timer": {
        "running": true,
        "entryId": "e1",
        "startedAt": "2026-09-21T09:15:00Z",
        "elapsedSeconds": 900
      },
      "children": [
        {
          "node": {
            "id": "n2",
            "parentId": "n1",
            "type": "task",
            "title": "The card",
            "descriptionMd": "",
            "status": "today",
            "due": null,
            "dueSource": "manual",
            "priority": 4,
            "estimateMin": null,
            "recurrence": null,
            "activity": null,
            "sortOrder": 0,
            "createdAt": "2026-09-01T08:00:00Z",
            "updatedAt": "2026-09-21T09:00:00Z",
            "completedAt": null,
            "archivedAt": null
          },
          "status": "today",
          "progress": {
            "done": 0,
            "total": 1,
            "percent": 0,
            "defined": true
          },
          "overdue": true,
          "isLeaf": true,
          "tags": [],
          "timer": {
            "running": false,
            "entryId": "",
            "startedAt": null,
            "elapsedSeconds": 0
          },
          "children": []
        }
      ]
    }
  ]
}`

func TestColumnViewJSONContract(t *testing.T) {
	created := time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC)
	updated := time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC)
	started := time.Date(2026, time.September, 21, 9, 15, 0, 0, time.UTC)
	due := domain.NewDate(2026, time.October, 2)
	activity := domain.ActivityManagement
	estimate := 90

	child := service.NodeView{
		Node: domain.Node{
			ID:        "n2",
			ParentID:  ptr("n1"),
			Type:      domain.NodeTypeTask,
			Title:     "The card",
			Status:    domain.StatusToday,
			DueSource: domain.DueSourceManual,
			Priority:  domain.Priority4,
			CreatedAt: created,
			UpdatedAt: updated,
		},
		Status:   domain.StatusToday,
		Progress: service.ProgressView{Done: 0, Total: 1, Percent: 0, Defined: true},
		Overdue:  true,
		IsLeaf:   true,
	}

	parent := service.NodeView{
		Node: domain.Node{
			ID:            "n1",
			Type:          domain.NodeTypeProject,
			Title:         "Ship the board",
			DescriptionMD: "# Ship it",
			Status:        domain.StatusBacklog,
			Due:           &due,
			DueSource:     domain.DueSourceAuto,
			Priority:      domain.Priority2,
			EstimateMin:   &estimate,
			Activity:      &activity,
			SortOrder:     3,
			CreatedAt:     created,
			UpdatedAt:     updated,
		},
		Status:   domain.StatusToday,
		Progress: service.ProgressView{Done: 1, Total: 2, Percent: 50, Defined: true},
		Tags:     []domain.Tag{{ID: "t1", Name: "work", Color: "#38bdf8"}},
		Timer: service.TimerView{
			Running:        true,
			EntryID:        "e1",
			StartedAt:      &started,
			ElapsedSeconds: 900,
		},
		Children: []service.NodeView{child},
	}

	got, err := json.MarshalIndent(service.ColumnView{
		Status: domain.StatusToday,
		Nodes:  []service.NodeView{parent},
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshalling the column: %v", err)
	}
	if string(got) != goldenColumn {
		t.Errorf("the wire contract changed.\n--- got ---\n%s\n--- want ---\n%s", got, goldenColumn)
	}
}

// A list is never null. The zero NodeView is the worst case — nothing was
// populated at all — and it still has to carry [] for both of its lists.
func TestEmptyViewsMarshalEmptyListsNotNull(t *testing.T) {
	t.Run("an empty NodeView", func(t *testing.T) {
		got, err := json.Marshal(service.NodeView{})
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		for _, want := range []string{`"tags":[]`, `"children":[]`} {
			if !strings.Contains(string(got), want) {
				t.Errorf("marshalled as %s, want it to contain %s", got, want)
			}
		}
		if strings.Contains(string(got), "null,\"children\"") || strings.Contains(string(got), `"tags":null`) {
			t.Errorf("marshalled as %s, want no null list", got)
		}
	})

	t.Run("an empty ColumnView", func(t *testing.T) {
		got, err := json.Marshal(service.ColumnView{Status: domain.StatusBacklog})
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		if want := `"nodes":[]`; !strings.Contains(string(got), want) {
			t.Errorf("marshalled as %s, want it to contain %s", got, want)
		}
	})

	t.Run("a nested nil child list is filled in too", func(t *testing.T) {
		got, err := json.Marshal(service.ColumnView{
			Status: domain.StatusBacklog,
			Nodes:  []service.NodeView{{Children: []service.NodeView{{}}}},
		})
		if err != nil {
			t.Fatalf("marshalling: %v", err)
		}
		// The optional scalars are legitimately null; the three LISTS never are.
		for _, list := range []string{"tags", "children", "nodes"} {
			if bad := `"` + list + `":null`; strings.Contains(string(got), bad) {
				t.Errorf("marshalled as %s, want no %s", got, bad)
			}
		}
	})
}

// wireTypes is every Go struct that crosses the Wails boundary: the six views,
// the three domain rows they carry, and — the half S2-02 first forgot — the two
// INPUTS, NewNode and SearchOptions.
//
// An input is on the contract exactly as much as an output is. App.CreateNode
// and App.Search take these structs, so Wails generates a TypeScript class for
// each and json.Unmarshals whatever the frontend sends back into them. Leaving
// them off this list is what let the create and search inputs ship in
// PascalCase while every read was lowerCamelCase.
func wireTypes() []reflect.Type {
	return []reflect.Type{
		reflect.TypeOf(service.NodeView{}),
		reflect.TypeOf(service.ProgressView{}),
		reflect.TypeOf(service.TimerView{}),
		reflect.TypeOf(service.ColumnView{}),
		reflect.TypeOf(service.HabitView{}),
		reflect.TypeOf(service.SettingsView{}),
		reflect.TypeOf(service.NewNode{}),
		reflect.TypeOf(service.SearchOptions{}),
		reflect.TypeOf(domain.Node{}),
		reflect.TypeOf(domain.Tag{}),
		reflect.TypeOf(domain.TimeEntry{}),
	}
}

// Every exported field of every wire type carries a tag, asserted by walking
// the struct rather than by eye. A field added later without one would otherwise
// arrive on the wire under its Go name and nobody would notice until the
// frontend read undefined.
func TestEveryWireFieldIsTagged(t *testing.T) {
	for _, typ := range wireTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			for i := range typ.NumField() {
				f := typ.Field(i)
				if !f.IsExported() {
					continue
				}
				tag, ok := f.Tag.Lookup("json")
				if !ok || tag == "" {
					t.Errorf("%s.%s has no json tag: it would cross the wire as %q",
						typ.Name(), f.Name, f.Name)
					continue
				}
				name := strings.Split(tag, ",")[0]
				switch {
				case name == "":
					t.Errorf("%s.%s has the empty json name %q", typ.Name(), f.Name, tag)
				case name[0] < 'a' || name[0] > 'z':
					t.Errorf("%s.%s is tagged %q: the convention is lowerCamelCase, no exceptions",
						typ.Name(), f.Name, name)
				case strings.ContainsAny(name, "_-"):
					t.Errorf("%s.%s is tagged %q: the convention is lowerCamelCase, not snake or kebab",
						typ.Name(), f.Name, name)
				}
			}
		})
	}
}

// The habit strip's entry is on the same contract as the board's card: explicit
// lowerCamelCase keys, visible here (S2-02, S2-04).
const goldenHabit = `{
  "node": {
    "id": "h1",
    "parentId": null,
    "type": "habit",
    "title": "Stretch every morning",
    "descriptionMd": "",
    "status": "backlog",
    "due": null,
    "dueSource": "manual",
    "priority": 4,
    "estimateMin": null,
    "recurrence": "FREQ=DAILY",
    "activity": null,
    "sortOrder": 0,
    "createdAt": "2026-09-01T08:00:00Z",
    "updatedAt": "2026-09-21T09:00:00Z",
    "completedAt": null,
    "archivedAt": null
  },
  "scheduledToday": true,
  "checkedToday": false,
  "streak": 12
}`

func TestHabitViewJSONContract(t *testing.T) {
	got, err := json.MarshalIndent(service.HabitView{
		Node: domain.Node{
			ID:         "h1",
			Type:       domain.NodeTypeHabit,
			Title:      "Stretch every morning",
			Status:     domain.StatusBacklog,
			DueSource:  domain.DueSourceManual,
			Priority:   domain.Priority4,
			Recurrence: ptr("FREQ=DAILY"),
			CreatedAt:  time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2026, time.September, 21, 9, 0, 0, 0, time.UTC),
		},
		ScheduledToday: true,
		Streak:         12,
	}, "", "  ")
	if err != nil {
		t.Fatalf("marshalling the habit view: %v", err)
	}
	if string(got) != goldenHabit {
		t.Errorf("the wire contract changed.\n--- got ---\n%s\n--- want ---\n%s", got, goldenHabit)
	}
}

// Exactly the fields whose Go shape differs from their JSON shape carry
// ts_type:"string", and no others (S2-02).
//
// # What this test is for
//
// Wails builds frontend/wailsjs/go/models.ts by REFLECTING over these structs,
// and reflection cannot see a MarshalJSON method. So the generated TypeScript
// described `due` as a {Year,Month,Day} class and every instant as `any`, while
// the JSON that actually crossed was "2026-10-02" and an RFC 3339 string. Go was
// right and the generated types were wrong, which is the worse way round: the
// frontend's entire job is to trust them.
//
// There are exactly two such Go types — domain.Date and time.Time — and the
// tag is asserted BY REFLECTION rather than by eye so that the date field
// somebody adds in Stage 3 cannot quietly reintroduce the defect.
//
// The converse half matters as much: a field that does NOT need the tag must not
// have one. ts_type overrides the generator outright, so a stray one would be a
// second, unchecked declaration of a type the generator already gets right —
// the same "two spellings of one rule" this project keeps paying for.
func TestWireDates(t *testing.T) {
	dateType := reflect.TypeOf(domain.Date{})
	timeType := reflect.TypeOf(time.Time{})

	for _, typ := range wireTypes() {
		t.Run(typ.String(), func(t *testing.T) {
			for i := range typ.NumField() {
				f := typ.Field(i)
				if !f.IsExported() {
					continue
				}

				ft := f.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				needsTag := ft == dateType || ft == timeType
				tag := f.Tag.Get("ts_type")

				switch {
				case needsTag && tag != "string":
					t.Errorf(`%s.%s is a %s, which crosses as a string but generates as %s without a tag; want ts_type:"string", got %q`,
						typ.Name(), f.Name, ft, generatedShapeOf(ft), tag)
				case !needsTag && tag != "":
					t.Errorf(`%s.%s is a %s and needs no ts_type tag, but carries ts_type:%q; the generator already describes it correctly`,
						typ.Name(), f.Name, f.Type, tag)
				}
			}
		})
	}
}

// generatedShapeOf names what Wails would have written for an untagged field of
// this type, so the failure above says what the frontend would have been told.
func generatedShapeOf(ft reflect.Type) string {
	if ft == reflect.TypeOf(time.Time{}) {
		return "`any`"
	}
	return "an object class"
}

// A date survives frontend → Go → frontend unchanged (S2-02).
//
// The string in the middle is the whole point: it is the one Go sends on a read,
// it is the one the generated `due?: string` tells the frontend to send back,
// and NewNode takes it through domain.Date.UnmarshalJSON — which is
// domain.ParseDate, the only date parser in the project.
func TestNewNodeDateRoundTrip(t *testing.T) {
	due := domain.NewDate(2026, time.October, 2)

	// Out: what a read hands the frontend.
	out, err := json.Marshal(domain.Node{Due: &due})
	if err != nil {
		t.Fatalf("marshalling the node: %v", err)
	}
	if want := `"due":"2026-10-02"`; !strings.Contains(string(out), want) {
		t.Fatalf("a read sent %s, want it to contain %s", out, want)
	}

	// Back in: the same string, in the create input.
	var draft service.NewNode
	if err := json.Unmarshal([]byte(`{"title":"Ship it","type":"task","due":"2026-10-02"}`), &draft); err != nil {
		t.Fatalf("unmarshalling the draft: %v", err)
	}
	if draft.Due == nil {
		t.Fatalf("the draft's due is nil, want %s", due)
	}
	if *draft.Due != due {
		t.Errorf("the date came back as %s, want %s", draft.Due, due)
	}

	// And out again, byte for byte what went in.
	again, err := json.Marshal(domain.Node{Due: draft.Due})
	if err != nil {
		t.Fatalf("marshalling the round-tripped node: %v", err)
	}
	if string(again) != string(out) {
		t.Errorf("the round trip changed the wire value:\n\tout  %s\n\tback %s", out, again)
	}
}

// An invalid date in the create input is REFUSED, with an error naming it.
//
// Silently zeroing it would be the dangerous outcome: a node created with
// 0001-01-01 is permanently overdue, on a board, with nothing on screen to
// explain it. json.Unmarshal calls domain.Date.UnmarshalJSON, which calls
// domain.ParseDate, so there is one answer to "is this a date?" and this is the
// door it guards.
func TestNewNodeRefusesAnInvalidDueDate(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"a date that does not exist", `{"due":"2026-02-30"}`},
		{"the wrong layout", `{"due":"02/10/2026"}`},
		{"an empty string", `{"due":""}`},
		{"a number", `{"due":20261002}`},
		// The shape the OLD generated TypeScript described, and would have had a
		// frontend send. That it is refused is the proof the old type was wrong:
		// a create built from `Due?: domain.Date` could never have succeeded.
		{"the object the generated Date class would have produced", `{"due":{"Year":2026,"Month":10,"Day":2}}`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var draft service.NewNode
			err := json.Unmarshal([]byte(tc.body), &draft)
			if err == nil {
				t.Fatalf("unmarshalling %s succeeded with due = %v, want an error", tc.body, draft.Due)
			}
			if !errors.Is(err, domain.ErrInvalid) {
				t.Errorf("unmarshalling %s gave %v, want it to wrap domain.ErrInvalid", tc.body, err)
			}
		})
	}
}

// An absent or null due date is not an error, and arrives as nil.
//
// Both spellings have to work: `due?: string` lets the frontend leave the key
// out, and a form that clears the field will send null. Neither is a date and
// neither is a failure.
func TestNewNodeTakesAnAbsentDueDate(t *testing.T) {
	for _, body := range []string{`{"title":"x"}`, `{"title":"x","due":null}`} {
		t.Run(body, func(t *testing.T) {
			var draft service.NewNode
			if err := json.Unmarshal([]byte(body), &draft); err != nil {
				t.Fatalf("unmarshalling %s: %v", body, err)
			}
			if draft.Due != nil {
				t.Errorf("due came back as %v, want nil", draft.Due)
			}
		})
	}
}

// THE wire contract for the two INPUTS (S2-02).
//
// Read it the way the golden column above is read: every key the frontend sends
// is visible here, in the casing it must use. `due` is the same "YYYY-MM-DD"
// string a read returns — one date format in both directions, not two.
const goldenNewNode = `{
  "parentId": "n1",
  "type": "task",
  "title": "Ship the board",
  "descriptionMd": "# Ship it",
  "status": "today",
  "due": "2026-10-02",
  "priority": 2,
  "estimateMin": 90,
  "recurrence": null,
  "activity": "Разработка"
}`

const goldenSearchOptions = `{
  "includeArchived": true,
  "limit": 25
}`

func TestInputsAreOnTheSameWireContract(t *testing.T) {
	t.Run("NewNode", func(t *testing.T) {
		due := domain.NewDate(2026, time.October, 2)
		activity := domain.ActivityDevelopment
		estimate := 90

		draft := service.NewNode{
			ParentID:      ptr("n1"),
			Type:          domain.NodeTypeTask,
			Title:         "Ship the board",
			DescriptionMD: "# Ship it",
			Status:        domain.StatusToday,
			Due:           &due,
			Priority:      domain.Priority2,
			EstimateMin:   &estimate,
			Activity:      &activity,
		}

		got, err := json.MarshalIndent(draft, "", "  ")
		if err != nil {
			t.Fatalf("marshalling the draft: %v", err)
		}
		if string(got) != goldenNewNode {
			t.Errorf("the create input's contract changed.\n--- got ---\n%s\n--- want ---\n%s", got, goldenNewNode)
		}

		// And it survives the trip back, which is the direction that matters:
		// this struct is an input.
		var back service.NewNode
		if err := json.Unmarshal(got, &back); err != nil {
			t.Fatalf("unmarshalling the draft: %v", err)
		}
		if !reflect.DeepEqual(back, draft) {
			t.Errorf("the create input did not round-trip:\n\tgot  %+v\n\twant %+v", back, draft)
		}
	})

	t.Run("SearchOptions", func(t *testing.T) {
		opts := service.SearchOptions{IncludeArchived: true, Limit: 25}

		got, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			t.Fatalf("marshalling the options: %v", err)
		}
		if string(got) != goldenSearchOptions {
			t.Errorf("the search input's contract changed.\n--- got ---\n%s\n--- want ---\n%s", got, goldenSearchOptions)
		}

		var back service.SearchOptions
		if err := json.Unmarshal(got, &back); err != nil {
			t.Fatalf("unmarshalling the options: %v", err)
		}
		if back != opts {
			t.Errorf("the search input did not round-trip: got %+v, want %+v", back, opts)
		}
	})
}
