package store

import (
	"context"
	"database/sql"
	"slices"
	"strings"
	"testing"
)

// testStamp is a valid ISO-8601 UTC timestamp in the format 0002 documents.
const testStamp = "2026-09-18T09:00:00Z"

// openMigratedDB opens a fresh temp database with the whole embedded migration
// set applied.
func openMigratedDB(t *testing.T) *sql.DB {
	t.Helper()

	db := openTestDB(t)
	if _, err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

// insertNode inserts one row into `nodes`, applying overrides on top of a row
// that is valid in every column, and returns the driver's error rather than
// failing the test — the CHECK-constraint tests exist to assert on that error.
//
// Overriding by column name keeps each test case down to the single value it is
// about, so a case that says {"status": "Done"} is visibly a test of the status
// constraint and of nothing else.
func insertNode(ctx context.Context, db *sql.DB, id string, overrides map[string]any) error {
	row := map[string]any{
		"id":             id,
		"parent_id":      nil,
		"type":           "task",
		"title":          "a node",
		"description_md": "",
		"status":         "backlog",
		"due":            nil,
		"due_source":     "manual",
		"priority":       4,
		"estimate_min":   nil,
		"recurrence":     nil,
		"activity":       nil,
		"sort_order":     0,
		"created_at":     testStamp,
		"updated_at":     testStamp,
		"completed_at":   nil,
		"archived_at":    nil,
	}

	for column, value := range overrides {
		if _, known := row[column]; !known {
			panic("insertNode: unknown column " + column)
		}
		row[column] = value
	}

	columns := make([]string, 0, len(row))
	for column := range row {
		columns = append(columns, column)
	}
	slices.Sort(columns)

	values := make([]any, 0, len(columns))
	placeholders := make([]string, 0, len(columns))
	for _, column := range columns {
		values = append(values, row[column])
		placeholders = append(placeholders, "?")
	}

	query := "INSERT INTO nodes (" + strings.Join(columns, ", ") + ") VALUES (" +
		strings.Join(placeholders, ", ") + ")"
	_, err := db.ExecContext(ctx, query, values...)
	return err
}

// mustInsertNode is insertNode for the rows a test needs to exist rather than
// to fail.
func mustInsertNode(t *testing.T, ctx context.Context, db *sql.DB, id string, overrides map[string]any) {
	t.Helper()

	if err := insertNode(ctx, db, id, overrides); err != nil {
		t.Fatalf("inserting node %q: %v", id, err)
	}
}

// Every enumerated column is defended by a CHECK, and every CHECK gets a
// subtest naming the value it must refuse. The database is the last line of
// defence: a bug that writes status='Done' has to fail at the point it happens
// rather than become a row no view will ever show.
func TestNodesCheckConstraintsRejectBadValues(t *testing.T) {
	ctx := context.Background()

	for _, tc := range []struct {
		name      string
		overrides map[string]any
	}{
		{name: "type: an unknown kind", overrides: map[string]any{"type": "epic"}},
		{name: "type: right value, wrong case", overrides: map[string]any{"type": "Task"}},
		{name: "type: empty", overrides: map[string]any{"type": ""}},

		{name: "status: an unknown column", overrides: map[string]any{"status": "in-progress"}},
		{name: "status: right value, wrong case", overrides: map[string]any{"status": "Done"}},
		{name: "status: empty", overrides: map[string]any{"status": ""}},

		{name: "due_source: neither manual nor auto", overrides: map[string]any{"due_source": "inherited"}},
		{name: "due_source: right value, wrong case", overrides: map[string]any{"due_source": "Auto"}},

		{name: "priority: 0 is below the range", overrides: map[string]any{"priority": 0}},
		{name: "priority: 5 is above the range", overrides: map[string]any{"priority": 5}},
		{name: "priority: negative", overrides: map[string]any{"priority": -1}},

		{name: "activity: not one of the seven", overrides: map[string]any{"activity": "Отдых"}},
		{name: "activity: the English translation", overrides: map[string]any{"activity": "Development"}},
		{name: "activity: empty string is not the same as NULL", overrides: map[string]any{"activity": ""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := openMigratedDB(t)

			err := insertNode(ctx, db, "bad", tc.overrides)
			if err == nil {
				t.Fatalf("insert with %v succeeded; the CHECK constraint did not fire", tc.overrides)
			}
			if !strings.Contains(strings.ToUpper(err.Error()), "CONSTRAINT") {
				t.Errorf("insert failed with %v, which does not look like a constraint violation", err)
			}
			if n := countRows(t, db, "nodes"); n != 0 {
				t.Errorf("nodes has %d rows after a rejected insert, want 0", n)
			}
		})
	}
}

// The mirror image: every value the enums declare has to be accepted, or the
// CHECK is not a guard, it is a bug. A constraint that rejects everything would
// pass every test above.
func TestNodesCheckConstraintsAcceptEveryValidValue(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		column string
		values []any
	}{
		{column: "type", values: []any{"task", "project", "habit", "note", "bug"}},
		{column: "status", values: []any{"backlog", "week", "today", "doing", "done"}},
		{column: "due_source", values: []any{"manual", "auto"}},
		{column: "priority", values: []any{1, 2, 3, 4}},
		{column: "activity", values: []any{
			nil,
			"Разработка",
			"Анализ",
			"Тестирование",
			"Документация",
			"Совещание",
			"Согласование",
			"Управление проектом",
		}},
	}

	for _, tc := range cases {
		t.Run(tc.column, func(t *testing.T) {
			db := openMigratedDB(t)

			for i, value := range tc.values {
				id := tc.column + "-" + strings.Repeat("x", i+1)
				if err := insertNode(ctx, db, id, map[string]any{tc.column: value}); err != nil {
					t.Errorf("%s = %v was rejected: %v", tc.column, value, err)
				}
			}
			if n := countRows(t, db, "nodes"); n != len(tc.values) {
				t.Errorf("nodes has %d rows, want %d", n, len(tc.values))
			}
		})
	}
}

// The column defaults are part of the contract: D1 says due_source defaults to
// 'manual', and a node arriving without a priority is the least urgent one.
func TestNodesColumnDefaults(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	if _, err := db.ExecContext(ctx,
		`INSERT INTO nodes (id, type, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		"defaults", "task", "a node", "backlog", testStamp, testStamp); err != nil {
		t.Fatalf("inserting with only the required columns: %v", err)
	}

	var (
		dueSource     string
		priority      int
		sortOrder     int
		descriptionMD string
	)
	if err := db.QueryRowContext(ctx,
		`SELECT due_source, priority, sort_order, description_md FROM nodes WHERE id = ?`, "defaults").
		Scan(&dueSource, &priority, &sortOrder, &descriptionMD); err != nil {
		t.Fatalf("reading the row back: %v", err)
	}

	if dueSource != "manual" {
		t.Errorf("due_source defaulted to %q, want %q (D1)", dueSource, "manual")
	}
	if priority != 4 {
		t.Errorf("priority defaulted to %d, want 4", priority)
	}
	if sortOrder != 0 {
		t.Errorf("sort_order defaulted to %d, want 0", sortOrder)
	}
	if descriptionMD != "" {
		t.Errorf("description_md defaulted to %q, want the empty string", descriptionMD)
	}
}

// Deleting a node takes its whole subtree and everything attached to it. The
// foreign_keys pragma is on for every connection, so ON DELETE CASCADE is live
// rather than decorative.
func TestDeletingANodeCascadesThroughTheWholeSubtree(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	mustInsertNode(t, ctx, db, "root", map[string]any{"type": "project"})
	mustInsertNode(t, ctx, db, "child", map[string]any{"parent_id": "root"})
	mustInsertNode(t, ctx, db, "grandchild", map[string]any{"parent_id": "child"})
	mustInsertNode(t, ctx, db, "survivor", nil)

	if _, err := db.ExecContext(ctx, `INSERT INTO tags (id, name, color) VALUES ('tag', 'work', '')`); err != nil {
		t.Fatalf("inserting a tag: %v", err)
	}

	for _, stmt := range []struct {
		name string
		sql  string
		args []any
	}{
		{name: "node_tags", sql: `INSERT INTO node_tags (node_id, tag_id) VALUES (?, 'tag')`, args: []any{"grandchild"}},
		{name: "time_entries", sql: `INSERT INTO time_entries (id, node_id, started_at, ended_at) VALUES ('te', ?, ?, ?)`, args: []any{"grandchild", testStamp, testStamp}},
		{name: "attachments", sql: `INSERT INTO attachments (id, node_id, path, mime) VALUES ('att', ?, 'a.png', 'image/png')`, args: []any{"child"}},
		{name: "habit_checks", sql: `INSERT INTO habit_checks (node_id, date) VALUES (?, '2026-09-18')`, args: []any{"child"}},
	} {
		if _, err := db.ExecContext(ctx, stmt.sql, stmt.args...); err != nil {
			t.Fatalf("seeding %s: %v", stmt.name, err)
		}
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM nodes WHERE id = 'root'`); err != nil {
		t.Fatalf("deleting the root: %v", err)
	}

	for _, table := range []string{"node_tags", "time_entries", "attachments", "habit_checks"} {
		t.Run(table+" is emptied", func(t *testing.T) {
			if n := countRows(t, db, table); n != 0 {
				t.Errorf("%s has %d rows after the cascade, want 0", table, n)
			}
		})
	}

	t.Run("the subtree is gone and nothing else is", func(t *testing.T) {
		rows, err := db.QueryContext(ctx, `SELECT id FROM nodes ORDER BY id`)
		if err != nil {
			t.Fatalf("listing nodes: %v", err)
		}
		defer rows.Close()

		var ids []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				t.Fatalf("scanning: %v", err)
			}
			ids = append(ids, id)
		}
		if err := rows.Err(); err != nil {
			t.Fatalf("listing nodes: %v", err)
		}

		if len(ids) != 1 || ids[0] != "survivor" {
			t.Errorf("nodes left after the cascade = %v, want [survivor]", ids)
		}
	})

	t.Run("the tag itself survives its node", func(t *testing.T) {
		if n := countRows(t, db, "tags"); n != 1 {
			t.Errorf("tags has %d rows, want 1: deleting a node must not delete the tag", n)
		}
	})
}

// The single-active timer invariant, enforced by the database rather than only
// by the service, because a race the service loses must not corrupt the data.
func TestOnlyOneOpenTimeEntryIsAllowedGlobally(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	mustInsertNode(t, ctx, db, "a", nil)
	mustInsertNode(t, ctx, db, "b", nil)

	openEntry := func(id, nodeID string) error {
		_, err := db.ExecContext(ctx,
			`INSERT INTO time_entries (id, node_id, started_at, ended_at) VALUES (?, ?, ?, NULL)`,
			id, nodeID, testStamp)
		return err
	}
	closedEntry := func(id, nodeID string) error {
		_, err := db.ExecContext(ctx,
			`INSERT INTO time_entries (id, node_id, started_at, ended_at) VALUES (?, ?, ?, ?)`,
			id, nodeID, testStamp, "2026-09-18T10:00:00Z")
		return err
	}

	t.Run("the first open entry is accepted", func(t *testing.T) {
		if err := openEntry("open-1", "a"); err != nil {
			t.Fatalf("the first open entry was rejected: %v", err)
		}
	})

	t.Run("a second open entry on another node is rejected", func(t *testing.T) {
		err := openEntry("open-2", "b")
		if err == nil {
			t.Fatal("a second open time entry was accepted; the single-active invariant is not enforced")
		}
		if !strings.Contains(strings.ToUpper(err.Error()), "UNIQUE CONSTRAINT") {
			t.Errorf("rejected with %v, which does not look like the unique index firing", err)
		}
	})

	t.Run("a second open entry on the same node is rejected too", func(t *testing.T) {
		if err := openEntry("open-3", "a"); err == nil {
			t.Error("a second open entry on the same node was accepted")
		}
	})

	t.Run("any number of closed entries is fine", func(t *testing.T) {
		for _, e := range []struct{ id, node string }{
			{"closed-1", "a"}, {"closed-2", "a"}, {"closed-3", "b"},
		} {
			if err := closedEntry(e.id, e.node); err != nil {
				t.Errorf("closed entry %s was rejected: %v", e.id, err)
			}
		}
	})

	t.Run("closing the open entry frees the slot", func(t *testing.T) {
		if _, err := db.ExecContext(ctx,
			`UPDATE time_entries SET ended_at = ? WHERE id = 'open-1'`, "2026-09-18T11:00:00Z"); err != nil {
			t.Fatalf("closing the open entry: %v", err)
		}
		if err := openEntry("open-4", "b"); err != nil {
			t.Errorf("a new timer could not be started after the previous one was closed: %v", err)
		}
	})

	t.Run("re-opening a closed entry by UPDATE is rejected as well", func(t *testing.T) {
		_, err := db.ExecContext(ctx, `UPDATE time_entries SET ended_at = NULL WHERE id = 'closed-1'`)
		if err == nil {
			t.Error("an UPDATE that produced a second open entry was accepted")
		}
	})
}

func TestHabitChecksRejectADuplicateDay(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	mustInsertNode(t, ctx, db, "habit", map[string]any{"type": "habit", "recurrence": "FREQ=DAILY"})
	mustInsertNode(t, ctx, db, "other", map[string]any{"type": "habit", "recurrence": "FREQ=DAILY"})

	check := func(nodeID, date string) error {
		_, err := db.ExecContext(ctx, `INSERT INTO habit_checks (node_id, date) VALUES (?, ?)`, nodeID, date)
		return err
	}

	t.Run("the first check for a day is accepted", func(t *testing.T) {
		if err := check("habit", "2026-09-18"); err != nil {
			t.Fatalf("first check: %v", err)
		}
	})

	t.Run("checking the same habit on the same day again is rejected", func(t *testing.T) {
		err := check("habit", "2026-09-18")
		if err == nil {
			t.Fatal("a duplicate (node_id, date) was accepted")
		}
		if !strings.Contains(strings.ToUpper(err.Error()), "UNIQUE CONSTRAINT") {
			t.Errorf("rejected with %v, which does not look like the primary key firing", err)
		}
	})

	t.Run("another day for the same habit is fine", func(t *testing.T) {
		if err := check("habit", "2026-09-19"); err != nil {
			t.Errorf("a second day was rejected: %v", err)
		}
	})

	t.Run("the same day for another habit is fine", func(t *testing.T) {
		if err := check("other", "2026-09-18"); err != nil {
			t.Errorf("the same day on a different habit was rejected: %v", err)
		}
	})
}

func TestTagNamesAreUnique(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	if _, err := db.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES ('t1', 'work')`); err != nil {
		t.Fatalf("inserting the first tag: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES ('t2', 'work')`); err == nil {
		t.Error("a duplicate tag name was accepted")
	}
}

func TestNodeTagsRejectADuplicatePairing(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	mustInsertNode(t, ctx, db, "n", nil)
	if _, err := db.ExecContext(ctx, `INSERT INTO tags (id, name) VALUES ('t', 'work')`); err != nil {
		t.Fatalf("inserting a tag: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO node_tags (node_id, tag_id) VALUES ('n', 't')`); err != nil {
		t.Fatalf("first pairing: %v", err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO node_tags (node_id, tag_id) VALUES ('n', 't')`); err == nil {
		t.Error("a duplicate (node_id, tag_id) pairing was accepted")
	}
}

// A row pointing at a node that does not exist must be refused outright, or the
// cascade has nothing to cascade along.
func TestForeignKeysRejectDanglingReferences(t *testing.T) {
	ctx := context.Background()
	db := openMigratedDB(t)

	for _, tc := range []struct {
		name string
		sql  string
	}{
		{name: "nodes.parent_id", sql: `INSERT INTO nodes (id, parent_id, type, title, status, created_at, updated_at) VALUES ('orphan', 'nobody', 'task', 't', 'backlog', '` + testStamp + `', '` + testStamp + `')`},
		{name: "time_entries.node_id", sql: `INSERT INTO time_entries (id, node_id, started_at) VALUES ('te', 'nobody', '` + testStamp + `')`},
		{name: "attachments.node_id", sql: `INSERT INTO attachments (id, node_id, path) VALUES ('att', 'nobody', 'a.png')`},
		{name: "habit_checks.node_id", sql: `INSERT INTO habit_checks (node_id, date) VALUES ('nobody', '2026-09-18')`},
		{name: "node_tags.tag_id", sql: `INSERT INTO node_tags (node_id, tag_id) VALUES ('nobody', 'nothing')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := db.ExecContext(ctx, tc.sql); err == nil {
				t.Error("a dangling foreign key was accepted")
			}
		})
	}
}

// Every table and index the schema promises has to be in sqlite_master. A
// missing index is invisible until the board is slow with a thousand nodes in
// it; a missing unique index is invisible until two timers run at once.
func TestCoreSchemaObjectsExist(t *testing.T) {
	db := openMigratedDB(t)

	t.Run("tables", func(t *testing.T) {
		for _, table := range []string{
			"nodes", "tags", "node_tags", "time_entries", "attachments", "habit_checks", "settings",
		} {
			if !tableExists(t, db, table) {
				t.Errorf("table %q is missing", table)
			}
		}
	})

	t.Run("indexes", func(t *testing.T) {
		for _, index := range []string{
			"idx_nodes_parent_id",
			"idx_nodes_status",
			"idx_nodes_due",
			"idx_nodes_archived_at",
			"idx_time_entries_node_id",
			"one_open_timer",
		} {
			var name string
			err := db.QueryRow(
				`SELECT name FROM sqlite_master WHERE type = 'index' AND name = ?`, index).Scan(&name)
			if err != nil {
				t.Errorf("index %q is missing from sqlite_master: %v", index, err)
			}
		}
	})
}
