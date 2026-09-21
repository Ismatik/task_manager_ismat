package service_test

import (
	"encoding/json"
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

// Every exported field of every wire type carries a tag, asserted by walking
// the struct rather than by eye. A field added later without one would otherwise
// arrive on the wire under its Go name and nobody would notice until the
// frontend read undefined.
func TestEveryWireFieldIsTagged(t *testing.T) {
	types := []reflect.Type{
		reflect.TypeOf(service.NodeView{}),
		reflect.TypeOf(service.ProgressView{}),
		reflect.TypeOf(service.TimerView{}),
		reflect.TypeOf(service.ColumnView{}),
		reflect.TypeOf(service.HabitView{}),
		reflect.TypeOf(domain.Node{}),
		reflect.TypeOf(domain.Tag{}),
		reflect.TypeOf(domain.TimeEntry{}),
	}

	for _, typ := range types {
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
