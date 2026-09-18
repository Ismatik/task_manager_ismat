package service_test

import (
	"context"
	"testing"
	"time"

	"nexus/internal/domain"
	"nexus/internal/service"
)

// searches returns a search service over the fixture's database and clock.
func (f *fixture) searches() *service.SearchService {
	f.t.Helper()

	return service.NewSearchService(f.begin, f.nodes, f.tags, f.search, f.clock())
}

// ids lists the node ids of a result, which is what a failure message should
// say.
func ids(views []service.NodeView) []string {
	out := make([]string, len(views))
	for i, v := range views {
		out[i] = v.Node.ID
	}
	return out
}

// contains reports whether a result holds this node.
func contains(views []service.NodeView, id string) bool {
	for _, v := range views {
		if v.Node.ID == id {
			return true
		}
	}
	return false
}

func TestSearchFindsByTitleAndDescription(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	byTitle := draft("quarterly report", domain.NodeTypeTask, nil)
	titled := f.create(byTitle)

	byBody := draft("something else entirely", domain.NodeTypeTask, nil)
	byBody.DescriptionMD = "remember to attach the quarterly numbers"
	described := f.create(byBody)

	f.create(draft("unrelated", domain.NodeTypeTask, nil))

	t.Run("by title", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "report", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 1 || got[0].Node.ID != titled.ID {
			t.Fatalf("results = %v, want [%s]", ids(got), titled.ID)
		}
	})

	t.Run("by description", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "attach", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 1 || got[0].Node.ID != described.ID {
			t.Fatalf("results = %v, want [%s]", ids(got), described.ID)
		}
	})

	t.Run("a word in both fields finds both", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "quarterly", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("results = %v, want both nodes", ids(got))
		}
	})

	t.Run("a word nobody typed finds nothing", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "wombat", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("results = %v, want none", ids(got))
		}
	})

	t.Run("punctuation is not a syntax error", func(t *testing.T) {
		for _, q := range []string{`report*`, `"report"`, `50% report`, `report AND`} {
			if _, err := f.searches().Search(ctx, q, service.SearchOptions{}); err != nil {
				t.Errorf("Search(%q) = %v, want the user's typing to be treated as text", q, err)
			}
		}
	})
}

// A result is a card: the derived status, the progress, the overdue flag and the
// tags arrive with it, exactly as they do on the board.
func TestSearchResultsAreFullCards(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	p := f.create(draft("migration project", domain.NodeTypeProject, nil))
	child := f.create(draft("migration step one", domain.NodeTypeTask, &p.ID))

	if _, err := f.tasks.MoveToColumn(ctx, child.ID, domain.StatusToday); err != nil {
		t.Fatalf("MoveToColumn: %v", err)
	}
	yesterday := domain.NewDate(2026, time.September, 17)
	if _, err := f.tasks.SetDue(ctx, child.ID, &yesterday); err != nil {
		t.Fatalf("SetDue: %v", err)
	}
	if err := f.tags.CreateTag(ctx, domain.Tag{ID: "t1", Name: "work"}); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if err := f.tags.AttachTag(ctx, child.ID, "t1"); err != nil {
		t.Fatalf("AttachTag: %v", err)
	}

	got, err := f.searches().Search(ctx, "migration", service.SearchOptions{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("results = %v, want both nodes", ids(got))
	}

	var parentView, childView *service.NodeView
	for i := range got {
		switch got[i].Node.ID {
		case p.ID:
			parentView = &got[i]
		case child.ID:
			childView = &got[i]
		}
	}
	if parentView == nil || childView == nil {
		t.Fatalf("results = %v, want the project and its task", ids(got))
	}

	t.Run("the parent carries its derived status and progress", func(t *testing.T) {
		if parentView.Status != domain.StatusToday {
			t.Errorf("derived status = %q, want today", parentView.Status)
		}
		if !parentView.Progress.Defined || parentView.Progress.Total != 1 {
			t.Errorf("progress = %+v, want 0/1 defined", parentView.Progress)
		}
		if parentView.IsLeaf {
			t.Error("IsLeaf = true for a node with a task under it")
		}
	})

	t.Run("the task carries overdue and its tags", func(t *testing.T) {
		if !childView.Overdue {
			t.Error("Overdue = false for a task due yesterday and not done")
		}
		if len(childView.Tags) != 1 || childView.Tags[0].Name != "work" {
			t.Errorf("tags = %+v, want work", childView.Tags)
		}
		if !childView.IsLeaf {
			t.Error("IsLeaf = false for a childless task")
		}
	})

	t.Run("the running timer shows on a result too", func(t *testing.T) {
		if _, err := f.timers().Start(ctx, child.ID); err != nil {
			t.Fatalf("Start: %v", err)
		}
		f.now = testNow.Add(time.Minute)

		got, err := f.searches().Search(ctx, "migration", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		for _, v := range got {
			if v.Node.ID == child.ID && !v.Timer.Running {
				t.Error("the running task's result does not show the timer")
			}
			if v.Node.ID == p.ID && v.Timer.Running {
				t.Error("the project's result shows a timer of its own")
			}
		}
	})
}

// countingBeginner counts the transactions a call opens.
type countingBeginner struct {
	inner  service.Beginner
	begins *int
}

func (b countingBeginner) Begin(ctx context.Context) (service.Tx, error) {
	*b.begins++
	return b.inner.Begin(ctx)
}

// An empty search box returns nothing — not everything — and does not even go to
// the database.
func TestSearchWithNothingToSearchFor(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	for i := range 3 {
		f.create(draft(string(rune('a'+i)), domain.NodeTypeTask, nil))
	}

	for _, q := range []string{"", " ", "\t\n  ", "   "} {
		t.Run("query "+q, func(t *testing.T) {
			begins := 0
			svc := service.NewSearchService(
				countingBeginner{inner: f.begin, begins: &begins},
				f.nodes, f.tags, f.search, f.clock())

			got, err := svc.Search(ctx, q, service.SearchOptions{})
			if err != nil {
				t.Fatalf("Search(%q) = %v", q, err)
			}
			if len(got) != 0 {
				t.Errorf("results = %v, want none", ids(got))
			}
			// The service answers a blank box itself: it does not open a
			// transaction, hand the whitespace to the backend and hope the
			// backend has the same opinion about it.
			if begins != 0 {
				t.Errorf("%d transactions were started for a blank search box, want 0", begins)
			}
		})
	}

	t.Run("punctuation with no words in it", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "!!! ...", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("results = %v, want none", ids(got))
		}
	})
}

// Archived nodes stay hidden unless the caller asks for them.
func TestSearchExcludesArchivedByDefault(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)

	live := f.create(draft("archive me later", domain.NodeTypeTask, nil))
	gone := f.create(draft("archive me now", domain.NodeTypeTask, nil))
	if _, err := f.tasks.ArchiveNode(ctx, gone.ID); err != nil {
		t.Fatalf("ArchiveNode: %v", err)
	}

	t.Run("by default", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "archive", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 1 || got[0].Node.ID != live.ID {
			t.Fatalf("results = %v, want only the live node", ids(got))
		}
	})

	t.Run("when asked for", func(t *testing.T) {
		got, err := f.searches().Search(ctx, "archive", service.SearchOptions{IncludeArchived: true})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 2 || !contains(got, gone.ID) {
			t.Fatalf("results = %v, want both nodes", ids(got))
		}
		for _, v := range got {
			if v.Node.ID == gone.ID && v.Node.ArchivedAt == nil {
				t.Error("the archived node comes back without its archived_at")
			}
		}
	})
}

func TestSearchLimit(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	for i := range 5 {
		f.create(draft("report "+string(rune('a'+i)), domain.NodeTypeTask, nil))
	}

	got, err := f.searches().Search(ctx, "report", service.SearchOptions{Limit: 2})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("results = %v, want 2", ids(got))
	}

	all, err := f.searches().Search(ctx, "report", service.SearchOptions{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(all) != 5 {
		t.Errorf("results = %v, want all five", ids(all))
	}
}

// A node that is renamed, or that stops existing, is searchable — or not —
// immediately: the index follows the rows rather than a rebuild.
func TestSearchFollowsTheRows(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	n := f.create(draft("original title", domain.NodeTypeTask, nil))

	renamed := f.get(n.ID)
	renamed.Title = "renamed entirely"
	if err := f.nodes.Update(ctx, renamed); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got, err := f.searches().Search(ctx, "original", service.SearchOptions{}); err != nil || len(got) != 0 {
		t.Errorf("searching for the old title = %v, %v; want no results", ids(got), err)
	}
	if got, err := f.searches().Search(ctx, "renamed", service.SearchOptions{}); err != nil || len(got) != 1 {
		t.Errorf("searching for the new title = %v, %v; want one result", ids(got), err)
	}

	if err := f.nodes.Delete(ctx, n.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got, err := f.searches().Search(ctx, "renamed", service.SearchOptions{}); err != nil || len(got) != 0 {
		t.Errorf("a deleted node is still searchable: %v, %v", ids(got), err)
	}

	t.Run("and a rebuild changes nothing that is already correct", func(t *testing.T) {
		kept := f.create(draft("still here", domain.NodeTypeTask, nil))
		if err := f.searches().Rebuild(ctx); err != nil {
			t.Fatalf("Rebuild: %v", err)
		}
		got, err := f.searches().Search(ctx, "still", service.SearchOptions{})
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(got) != 1 || got[0].Node.ID != kept.ID {
			t.Errorf("results = %v, want [%s]", ids(got), kept.ID)
		}
	})
}

// A search that cannot reach the database reports it rather than returning no
// results, which would read as "nothing matched".
func TestSearchSurfacesStoreFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.create(draft("report", domain.NodeTypeTask, nil))

	blocked := service.NewSearchService(blockedBeginner{}, f.nodes, f.tags, f.search, f.clock())
	if _, err := blocked.Search(ctx, "report", service.SearchOptions{}); err == nil {
		t.Error("Search succeeded with no transaction")
	}
	if err := blocked.Rebuild(ctx); err == nil {
		t.Error("Rebuild succeeded with no transaction")
	}

	broken := service.NewSearchService(
		wrappingBeginner{inner: f.begin, wrap: func(tx service.Tx) service.Tx {
			return queryFailingTx{Tx: tx}
		}}, f.nodes, f.tags, f.search, f.clock())
	if _, err := broken.Search(ctx, "report", service.SearchOptions{}); err == nil {
		t.Error("Search succeeded with failing reads")
	}
}
