package store

import (
	"context"
	"database/sql"
	"io/fs"
	"slices"
	"strings"
	"testing"
	"time"

	"nexus/internal/domain"
)

// searchRepo returns a search repository, a node repository and the database
// behind both, over a freshly migrated temp file.
func searchRepo(t *testing.T) (*SearchRepo, *NodeRepo, *sql.DB) {
	t.Helper()

	db := openMigratedDB(t)
	return NewSearchRepo(db), NewNodeRepo(db), db
}

// searchNode is a node with a title and a description worth indexing.
func searchNode(id, title, description string) domain.Node {
	n := newTestNode(id)
	n.Title = title
	n.DescriptionMD = description
	return n
}

// assertIndexInSync asks FTS5 whether the index still agrees with `nodes`.
//
// This is the assertion the joined Search result cannot make. Search joins
// nodes_fts to nodes on rowid, so a stale index entry for a row that no longer
// exists is silently dropped by the join — the result looks right while the
// index is wrong, and it stays wrong until a rowid is reused and the wrong node
// starts turning up in somebody's search.
//
// The `rank` argument is what makes the check deep: integrity-check on its own
// only verifies that the index is internally consistent, and returns happily
// over an index that has drifted from an external content table. With rank = 1
// FTS5 re-reads `nodes` and compares, and reports SQLITE_CORRUPT_VTAB when the
// two disagree.
func assertIndexInSync(t *testing.T, db *sql.DB) {
	t.Helper()

	const check = "INSERT INTO nodes_fts(nodes_fts, rank) VALUES ('integrity-check', 1)"
	if _, err := db.ExecContext(context.Background(), check); err != nil {
		t.Errorf("the search index has drifted from `nodes`: %v", err)
	}
}

// indexHits counts the rows the INDEX matches, without joining `nodes`, so that
// a stale entry cannot hide behind the join.
func indexHits(t *testing.T, db *sql.DB, query string) int {
	t.Helper()

	match, ok := ftsQuery(query)
	if !ok {
		t.Fatalf("ftsQuery(%q) produced nothing searchable", query)
	}

	var n int
	err := db.QueryRowContext(context.Background(),
		"SELECT count(*) FROM nodes_fts WHERE nodes_fts MATCH ?", match).Scan(&n)
	if err != nil {
		t.Fatalf("counting index hits for %q: %v", query, err)
	}
	return n
}

func mustSearch(t *testing.T, r *SearchRepo, query string, opts SearchOptions) []string {
	t.Helper()

	got, err := r.Search(context.Background(), query, opts)
	if err != nil {
		t.Fatalf("Search(%q): %v", query, err)
	}
	return ids(got)
}

// S1-04 probed this driver and recorded FTS5: AVAILABLE, so this is the FTS5
// branch: three migrations ship, and the third is the search index.
func TestSearchIsBackedByTheFTS5Branch(t *testing.T) {
	ctx := context.Background()

	t.Run("a fresh database applies three migrations and re-migrating applies none", func(t *testing.T) {
		db := openTestDB(t)

		applied, err := Migrate(ctx, db)
		if err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		if applied != 3 {
			t.Errorf("a fresh database applied %d migrations, want 3", applied)
		}

		again, err := Migrate(ctx, db)
		if err != nil {
			t.Fatalf("second Migrate: %v", err)
		}
		if again != 0 {
			t.Errorf("re-migrating applied %d migrations, want 0", again)
		}
	})

	t.Run("the migration set on disk is the three files, 0003 included", func(t *testing.T) {
		names, err := fs.Glob(migrationsFS, migrationsDir+"/*.sql")
		if err != nil {
			t.Fatalf("listing migrations: %v", err)
		}
		slices.Sort(names)

		want := []string{
			migrationsDir + "/0001_settings.sql",
			migrationsDir + "/0002_core_schema.sql",
			migrationsDir + "/0003_search.sql",
		}
		if !slices.Equal(names, want) {
			t.Errorf("migrations = %v, want %v", names, want)
		}
	})

	t.Run("the index and all three sync triggers exist", func(t *testing.T) {
		db := openMigratedDB(t)

		for _, object := range []struct{ kind, name string }{
			{"table", "nodes_fts"},
			{"trigger", "nodes_fts_after_insert"},
			{"trigger", "nodes_fts_after_update"},
			{"trigger", "nodes_fts_after_delete"},
		} {
			var n int
			err := db.QueryRowContext(ctx,
				"SELECT count(*) FROM sqlite_master WHERE type = ? AND name = ?",
				object.kind, object.name).Scan(&n)
			if err != nil {
				t.Fatalf("reading sqlite_master: %v", err)
			}
			if n != 1 {
				t.Errorf("sqlite_master has %d %s(s) named %q, want 1", n, object.kind, object.name)
			}
		}
	})

	// The migration is applied by the shared runner, inside its transaction,
	// with no BEGIN/COMMIT of its own.
	t.Run("0003 contains no transaction control of its own", func(t *testing.T) {
		body, err := fs.ReadFile(migrationsFS, migrationsDir+"/0003_search.sql")
		if err != nil {
			t.Fatalf("reading 0003_search.sql: %v", err)
		}
		for _, keyword := range []string{"BEGIN TRANSACTION", "COMMIT", "ROLLBACK"} {
			if strings.Contains(strings.ToUpper(string(body)), keyword) {
				t.Errorf("0003_search.sql contains %q; the runner owns the transaction", keyword)
			}
		}
	})
}

func TestSearchFindsNodes(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
		want  []string
	}{
		{name: "a word in the title", query: "kanban", want: []string{"n1"}},
		{name: "a word in description_md", query: "zebra", want: []string{"n2"}},
		{name: "a word that is in neither", query: "elephant", want: []string{}},
		{name: "the match is case-insensitive", query: "KANBAN", want: []string{"n1"}},
		{name: "two words are ANDed, not ORed", query: "kanban board", want: []string{"n1"}},
		{name: "two words that are not in one node match nothing", query: "kanban zebra", want: []string{}},
		{name: "a word split across title and description of one node", query: "notes zebra", want: []string{"n2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, nodes, _ := searchRepo(t)
			mustCreate(t, ctx, nodes, searchNode("n1", "kanban board", "columns and cards"))
			mustCreate(t, ctx, nodes, searchNode("n2", "release notes", "the body mentions a zebra"))

			if got := mustSearch(t, r, tc.query, SearchOptions{}); !slices.Equal(got, tc.want) {
				t.Errorf("Search(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

// The triggers are the whole point of an external-content index: it does not
// update itself. These are the subtests that catch a missing one.
func TestSearchIndexStaysInSyncWithNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("an inserted node becomes searchable", func(t *testing.T) {
		r, nodes, db := searchRepo(t)

		if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 0 {
			t.Fatalf("Search before the insert = %v, want empty", got)
		}
		mustCreate(t, ctx, nodes, searchNode("n1", "aardvark", ""))

		if got := mustSearch(t, r, "aardvark", SearchOptions{}); !slices.Equal(got, []string{"n1"}) {
			t.Errorf("Search after the insert = %v, want [n1]", got)
		}
		assertIndexInSync(t, db)
	})

	// The AFTER UPDATE trigger has to retract the OLD values before inserting
	// the new ones. Without the retraction the old word matches forever, which
	// is invisible until somebody searches for a name they renamed away.
	t.Run("renaming a node stops the old word matching and starts the new one", func(t *testing.T) {
		r, nodes, db := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("n1", "aardvark", ""))

		renamed := searchNode("n1", "buffalo", "")
		if err := nodes.Update(ctx, renamed); err != nil {
			t.Fatalf("Update: %v", err)
		}

		if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 0 {
			t.Errorf("Search(old title) = %v, want empty — the update trigger did not retract it", got)
		}
		if got := mustSearch(t, r, "buffalo", SearchOptions{}); !slices.Equal(got, []string{"n1"}) {
			t.Errorf("Search(new title) = %v, want [n1]", got)
		}
		if n := indexHits(t, db, "aardvark"); n != 0 {
			t.Errorf("the index still holds %d entries for the old title", n)
		}
		assertIndexInSync(t, db)
	})

	t.Run("rewriting a description swaps its words too", func(t *testing.T) {
		r, nodes, db := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("n1", "task", "mentions aardvark"))

		rewritten := searchNode("n1", "task", "mentions buffalo")
		if err := nodes.Update(ctx, rewritten); err != nil {
			t.Fatalf("Update: %v", err)
		}

		if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 0 {
			t.Errorf("Search(old description word) = %v, want empty", got)
		}
		if got := mustSearch(t, r, "buffalo", SearchOptions{}); !slices.Equal(got, []string{"n1"}) {
			t.Errorf("Search(new description word) = %v, want [n1]", got)
		}
		assertIndexInSync(t, db)
	})

	// The join to `nodes` would hide a missing delete trigger all on its own,
	// so this asserts on the index directly as well as on the search result.
	t.Run("a deleted node disappears from the results and from the index", func(t *testing.T) {
		r, nodes, db := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("n1", "aardvark", ""))

		if err := nodes.Delete(ctx, "n1"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 0 {
			t.Errorf("Search after the delete = %v, want empty", got)
		}
		if n := indexHits(t, db, "aardvark"); n != 0 {
			t.Errorf("the index still holds %d entries for the deleted node", n)
		}
		assertIndexInSync(t, db)
	})

	// Deleting a parent removes its children through ON DELETE CASCADE, which
	// fires the row triggers for each of them — the case a hand-written
	// "delete from the index when the service deletes a node" would miss.
	t.Run("a cascaded delete takes the descendants out of the index too", func(t *testing.T) {
		r, nodes, db := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("root", "root aardvark", ""))
		child := searchNode("child", "child buffalo", "")
		child.ParentID = strPtr("root")
		mustCreate(t, ctx, nodes, child)

		if err := nodes.Delete(ctx, "root"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if got := mustSearch(t, r, "buffalo", SearchOptions{}); len(got) != 0 {
			t.Errorf("Search for the cascaded child = %v, want empty", got)
		}
		if n := indexHits(t, db, "buffalo"); n != 0 {
			t.Errorf("the index still holds %d entries for the cascaded child", n)
		}
		assertIndexInSync(t, db)
	})

	// Updating a column that is not indexed must not disturb the index.
	t.Run("an unrelated column change leaves the node findable", func(t *testing.T) {
		r, nodes, db := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("n1", "aardvark", ""))

		changes := []domain.StatusChange{{NodeID: "n1", Status: domain.StatusDoing}}
		if err := nodes.UpdateStatuses(ctx, changes, testNow.Add(time.Hour)); err != nil {
			t.Fatalf("UpdateStatuses: %v", err)
		}
		if got := mustSearch(t, r, "aardvark", SearchOptions{}); !slices.Equal(got, []string{"n1"}) {
			t.Errorf("Search after a status change = %v, want [n1]", got)
		}
		assertIndexInSync(t, db)
	})
}

func TestSearchExcludesArchivedByDefault(t *testing.T) {
	ctx := context.Background()
	r, nodes, _ := searchRepo(t)

	mustCreate(t, ctx, nodes, searchNode("live", "aardvark one", ""))
	mustCreate(t, ctx, nodes, searchNode("hidden", "aardvark two", ""))

	at := testNow.Add(time.Hour)
	if err := nodes.SetArchivedAt(ctx, []string{"hidden"}, &at, at); err != nil {
		t.Fatalf("SetArchivedAt: %v", err)
	}

	t.Run("excluded by default", func(t *testing.T) {
		if got := mustSearch(t, r, "aardvark", SearchOptions{}); !slices.Equal(got, []string{"live"}) {
			t.Errorf("Search = %v, want [live]", got)
		}
	})

	t.Run("included on request", func(t *testing.T) {
		got := mustSearch(t, r, "aardvark", SearchOptions{IncludeArchived: true})
		slices.Sort(got)
		if want := []string{"hidden", "live"}; !slices.Equal(got, want) {
			t.Errorf("Search(IncludeArchived) = %v, want %v", got, want)
		}
	})
}

// An empty search box must not list the database.
func TestSearchEmptyQueryReturnsNothing(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
	}{
		{name: "the empty string", query: ""},
		{name: "spaces", query: "   "},
		{name: "a tab and a newline", query: "\t\n"},
		{name: "punctuation only", query: "  ***  "},
		{name: "a single quote", query: "'"},
		{name: "a double quote", query: `"`},
		{name: "a percent sign", query: "%"},
		{name: "an underscore", query: "_"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, nodes, _ := searchRepo(t)
			for _, id := range []string{"n1", "n2", "n3"} {
				mustCreate(t, ctx, nodes, searchNode(id, "node "+id, "some body text"))
			}

			got, err := r.Search(ctx, tc.query, SearchOptions{})
			if err != nil {
				t.Fatalf("Search(%q): %v", tc.query, err)
			}
			if len(got) != 0 {
				t.Errorf("Search(%q) = %v, want empty — an empty query must not list the table", tc.query, ids(got))
			}
			if got == nil {
				t.Error("Search returned a nil slice, want an empty non-nil one")
			}
		})
	}
}

// Every character that means something to FTS5, or to LIKE, or to SQL, typed
// into a search box by somebody who meant it literally. None of them may error,
// and none of them may match everything.
func TestSearchTreatsUserInputAsLiteralText(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
		want  []string
	}{
		{name: "a percent sign is not a wildcard", query: "100%", want: []string{"pct"}},
		{name: "an underscore is not a single-character wildcard", query: "snake_case", want: []string{"snake"}},
		{name: "an asterisk is not a prefix operator", query: "star*", want: []string{"star"}},
		{name: "a double quote does not open a phrase", query: `say "hello"`, want: []string{"quote"}},
		{name: "an apostrophe does not close a string literal", query: "o'brien", want: []string{"irish"}},
		{name: "AND is a word, not an operator", query: "and", want: []string{"conj"}},
		{name: "OR is a word, not an operator", query: "or", want: []string{"conj"}},
		{name: "NOT is a word, not an operator", query: "not", want: []string{"conj"}},
		{name: "NEAR is a word, not a function", query: "near", want: []string{"conj"}},
		{name: "a caret is not a column filter", query: "^title", want: []string{"caret"}},
		{name: "a colon is not a column filter", query: "title:something", want: []string{}},
		{name: "a semicolon cannot end the statement", query: "drop; table", want: []string{}},
		{name: "parentheses are not grouping", query: "(star)", want: []string{"star"}},
		{name: "a lone minus is not a negation", query: "-star", want: []string{"star"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, nodes, _ := searchRepo(t)
			for _, n := range []domain.Node{
				searchNode("pct", "100% complete", ""),
				searchNode("snake", "snake_case naming", ""),
				searchNode("star", "star of the show", ""),
				searchNode("quote", `say "hello" to them`, ""),
				searchNode("irish", "o'brien the builder", ""),
				searchNode("conj", "and or not near", ""),
				searchNode("caret", "title of the piece", ""),
				searchNode("plain", "a perfectly ordinary node", ""),
			} {
				mustCreate(t, ctx, nodes, n)
			}

			got, err := r.Search(ctx, tc.query, SearchOptions{})
			if err != nil {
				t.Fatalf("Search(%q) returned an error, want literal-text handling: %v", tc.query, err)
			}

			gotIDs := ids(got)
			slices.Sort(gotIDs)
			want := slices.Clone(tc.want)
			slices.Sort(want)
			if !slices.Equal(gotIDs, want) {
				t.Errorf("Search(%q) = %v, want %v", tc.query, gotIDs, want)
			}
			if len(got) == 8 {
				t.Errorf("Search(%q) matched every row; the query was interpreted, not quoted", tc.query)
			}
		})
	}
}

// The app is EN/RU. A Russian title has to be findable with a Russian query,
// including with different case, or half the product is unsearchable.
func TestSearchHandlesUnicode(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name  string
		query string
		want  []string
	}{
		{name: "a Russian word in the title", query: "отчёт", want: []string{"ru"}},
		{name: "the same word in upper case", query: "ОТЧЁТ", want: []string{"ru"}},
		{name: "a Russian word in the description", query: "квартал", want: []string{"ru"}},
		{name: "two Russian words are ANDed", query: "отчёт сентябрь", want: []string{"ru"}},
		{name: "a Russian word that is not there", query: "совещание", want: []string{}},
		{name: "an English query does not match the Russian node", query: "report", want: []string{"en"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, nodes, _ := searchRepo(t)
			mustCreate(t, ctx, nodes, searchNode("ru", "Отчёт за сентябрь", "Итоги за квартал"))
			mustCreate(t, ctx, nodes, searchNode("en", "September report", "quarter summary"))

			if got := mustSearch(t, r, tc.query, SearchOptions{}); !slices.Equal(got, tc.want) {
				t.Errorf("Search(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

// The index can only drift if something bypasses the triggers, but a user with
// a sqlite3 shell and a VACUUM is exactly that. Rebuild is the way back.
func TestSearchRebuild(t *testing.T) {
	ctx := context.Background()
	r, nodes, db := searchRepo(t)

	for _, id := range []string{"n1", "n2"} {
		mustCreate(t, ctx, nodes, searchNode(id, "aardvark "+id, "body of "+id))
	}
	if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 2 {
		t.Fatalf("Search before emptying the index = %v, want two results", got)
	}

	// Empty the virtual table behind the triggers' back.
	if _, err := db.ExecContext(ctx, "INSERT INTO nodes_fts(nodes_fts) VALUES ('delete-all')"); err != nil {
		t.Fatalf("emptying the index: %v", err)
	}
	if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 0 {
		t.Fatalf("Search over the emptied index = %v, want empty", got)
	}

	if err := r.Rebuild(ctx); err != nil {
		t.Fatalf("Rebuild: %v", err)
	}

	got := mustSearch(t, r, "aardvark", SearchOptions{})
	slices.Sort(got)
	if want := []string{"n1", "n2"}; !slices.Equal(got, want) {
		t.Errorf("Search after Rebuild = %v, want %v", got, want)
	}
	if body := mustSearch(t, r, "body", SearchOptions{}); len(body) != 2 {
		t.Errorf("Search of the descriptions after Rebuild = %v, want two results", body)
	}
}

func TestSearchOptionsLimit(t *testing.T) {
	ctx := context.Background()
	r, nodes, _ := searchRepo(t)

	for _, id := range []string{"n1", "n2", "n3"} {
		mustCreate(t, ctx, nodes, searchNode(id, "aardvark "+id, ""))
	}

	t.Run("no limit returns everything", func(t *testing.T) {
		if got := mustSearch(t, r, "aardvark", SearchOptions{}); len(got) != 3 {
			t.Errorf("Search = %v, want three results", got)
		}
	})

	t.Run("a limit caps the results", func(t *testing.T) {
		if got := mustSearch(t, r, "aardvark", SearchOptions{Limit: 2}); len(got) != 2 {
			t.Errorf("Search(Limit 2) = %v, want two results", got)
		}
	})

	t.Run("a negative limit means no limit", func(t *testing.T) {
		if got := mustSearch(t, r, "aardvark", SearchOptions{Limit: -1}); len(got) != 3 {
			t.Errorf("Search(Limit -1) = %v, want three results", got)
		}
	})
}

// The result carries whole nodes, not ids: the caller gets every column it
// would get from NodeRepo.Get, so a search result can be rendered without a
// second round trip per row.
func TestSearchReturnsWholeNodes(t *testing.T) {
	ctx := context.Background()
	r, nodes, _ := searchRepo(t)

	want := searchNode("n1", "aardvark", "the body")
	want.Due = datePtr(domain.NewDate(2026, time.September, 18))
	want.DueSource = domain.DueSourceAuto
	want.Activity = activityPtr(domain.ActivityAnalysis)
	want.EstimateMin = intPtr(45)
	want.Priority = domain.Priority2
	mustCreate(t, ctx, nodes, want)

	got, err := r.Search(ctx, "aardvark", SearchOptions{})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("Search = %v, want one result", ids(got))
	}
	assertNodeEqual(t, got[0], want)
}

// Two searches of the same data must produce the same order. bm25 ranks first;
// the tie-break exists because bm25 alone does not order equally relevant rows.
func TestSearchOrderIsDeterministic(t *testing.T) {
	ctx := context.Background()
	r, nodes, _ := searchRepo(t)

	for _, id := range []string{"n3", "n1", "n2"} {
		n := searchNode(id, "aardvark", "")
		n.UpdatedAt = testNow
		mustCreate(t, ctx, nodes, n)
	}

	first := mustSearch(t, r, "aardvark", SearchOptions{})
	for range 5 {
		if got := mustSearch(t, r, "aardvark", SearchOptions{}); !slices.Equal(got, first) {
			t.Fatalf("Search returned %v then %v; the order is not deterministic", first, got)
		}
	}

	t.Run("a shorter, more relevant title outranks a long description", func(t *testing.T) {
		r, nodes, _ := searchRepo(t)
		mustCreate(t, ctx, nodes, searchNode("title", "aardvark", strings.Repeat("filler words here ", 40)))
		mustCreate(t, ctx, nodes, searchNode("body", "something else entirely",
			strings.Repeat("filler words here ", 40)+" aardvark "+strings.Repeat("more filler ", 40)))

		got := mustSearch(t, r, "aardvark", SearchOptions{})
		if len(got) != 2 {
			t.Fatalf("Search = %v, want two results", got)
		}
		if got[0] != "title" {
			t.Errorf("Search = %v, want the short title first (bm25 ranking)", got)
		}
	})
}

// ftsQuery is the whole escaping story, so it is tested on its own as well as
// through the database.
func TestFTSQuery(t *testing.T) {
	for _, tc := range []struct {
		name     string
		in       string
		want     string
		wantAble bool
	}{
		{name: "one word", in: "kanban", want: `"kanban"`, wantAble: true},
		{name: "two words are ANDed", in: "kanban board", want: `"kanban" "board"`, wantAble: true},
		{name: "punctuation separates", in: "snake_case", want: `"snake" "case"`, wantAble: true},
		{name: "a percent sign is dropped", in: "100%", want: `"100"`, wantAble: true},
		{name: "an asterisk is dropped", in: "star*", want: `"star"`, wantAble: true},
		{name: "a double quote is dropped", in: `say "hi"`, want: `"say" "hi"`, wantAble: true},
		{name: "an apostrophe separates", in: "o'brien", want: `"o" "brien"`, wantAble: true},
		{name: "operators become quoted words", in: "a AND b", want: `"a" "AND" "b"`, wantAble: true},
		{name: "Cyrillic is one token", in: "Отчёт", want: `"Отчёт"`, wantAble: true},
		{name: "digits count", in: "2026", want: `"2026"`, wantAble: true},
		{name: "mixed scripts", in: "отчёт report", want: `"отчёт" "report"`, wantAble: true},
		{name: "the empty string is unsearchable", in: "", want: "", wantAble: false},
		{name: "whitespace is unsearchable", in: " \t\n ", want: "", wantAble: false},
		{name: "punctuation alone is unsearchable", in: `*"%_'();`, want: "", wantAble: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ftsQuery(tc.in)
			if ok != tc.wantAble {
				t.Errorf("ftsQuery(%q) searchable = %v, want %v", tc.in, ok, tc.wantAble)
			}
			if got != tc.want {
				t.Errorf("ftsQuery(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Everything the repository produces has to be balanced FTS5 syntax: a query
// expression that does not parse reaches the user as an error message about a
// search language they never agreed to learn.
func TestSearchNeverProducesASyntaxError(t *testing.T) {
	ctx := context.Background()
	r, nodes, _ := searchRepo(t)
	mustCreate(t, ctx, nodes, searchNode("n1", "a perfectly ordinary node", "with a body"))

	for _, query := range []string{
		`"`, `""`, `"""`, `*`, `**`, `^`, `-`, `NOT`, `AND OR NOT`, `a AND`, `AND a`,
		`(`, `)`, `()`, `a:b`, `"unterminated`, `foo"bar`, `NEAR(a b)`, `{col}`,
		`\`, `%`, `_`, `;--`, `' OR 1=1 --`, `a*b`, `ordinary*`, `^ordinary`,
		strings.Repeat("a", 1000), "\x00nul", "emoji 🙂 here",
	} {
		t.Run(query, func(t *testing.T) {
			if _, err := r.Search(ctx, query, SearchOptions{}); err != nil {
				t.Errorf("Search(%q) = %v, want no error", query, err)
			}
		})
	}
}

func TestSearchRunsInsideACallerTransaction(t *testing.T) {
	ctx := context.Background()
	r, nodes, db := searchRepo(t)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("BeginTx: %v", err)
	}

	// A node created inside the transaction is searchable inside it.
	mustCreate(t, ctx, nodes.WithExecutor(tx), searchNode("n1", "aardvark", ""))

	got, err := r.WithExecutor(tx).Search(ctx, "aardvark", SearchOptions{})
	if err != nil {
		t.Fatalf("Search in the transaction: %v", err)
	}
	if !slices.Equal(ids(got), []string{"n1"}) {
		t.Errorf("Search in the transaction = %v, want [n1]", ids(got))
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}

	after, err := r.Search(ctx, "aardvark", SearchOptions{})
	if err != nil {
		t.Fatalf("Search after the rollback: %v", err)
	}
	if len(after) != 0 {
		t.Errorf("Search after the rollback = %v, want empty — the index rolled back with the row", ids(after))
	}
}
