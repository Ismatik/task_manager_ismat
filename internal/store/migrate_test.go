package store

import (
	"context"
	"database/sql"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// openTestDB opens a database in a fresh temp directory.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := Open(context.Background(), tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	})
	return db
}

// migrationSet builds an in-memory migration directory laid out exactly like
// the embedded one.
func migrationSet(files map[string]string) fs.FS {
	fsys := fstest.MapFS{}
	for name, body := range files {
		fsys[migrationsDir+"/"+name] = &fstest.MapFile{Data: []byte(body)}
	}
	return fsys
}

// countRows returns the number of rows in table, or fails if it does not exist.
func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()

	var n int
	if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("counting %s: %v", table, err)
	}
	return n
}

// tableExists reports whether a table of that name is in the schema.
func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()

	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		t.Fatalf("looking for table %s: %v", table, err)
	}
	return true
}

// The embedded set is what actually ships, so it has to migrate a real file.
func TestMigrateEmbeddedSetIsApplicableAndIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	first, err := Migrate(ctx, db)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	second, err := Migrate(ctx, db)
	if err != nil {
		t.Fatalf("Migrate (second run): %v", err)
	}
	if second != 0 {
		t.Errorf("second Migrate applied %d migrations, want 0", second)
	}

	if got := countRows(t, db, "schema_migrations"); got != first {
		t.Errorf("schema_migrations has %d rows after two runs, want %d", got, first)
	}
}

func TestMigrateAppliesEachMigrationOnce(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	fsys := migrationSet(map[string]string{
		"0001_first.sql":  "CREATE TABLE first (id INTEGER PRIMARY KEY);",
		"0002_second.sql": "CREATE TABLE second (id INTEGER PRIMARY KEY);",
	})

	applied, err := migrateFS(ctx, db, fsys)
	if err != nil {
		t.Fatalf("migrateFS: %v", err)
	}
	if applied != 2 {
		t.Errorf("first run applied %d migrations, want 2", applied)
	}
	for _, table := range []string{"first", "second"} {
		if !tableExists(t, db, table) {
			t.Errorf("table %q was not created", table)
		}
	}

	applied, err = migrateFS(ctx, db, fsys)
	if err != nil {
		t.Fatalf("migrateFS (second run): %v", err)
	}
	if applied != 0 {
		t.Errorf("second run applied %d migrations, want 0 — migrations must run exactly once", applied)
	}
	if got := countRows(t, db, "schema_migrations"); got != 2 {
		t.Errorf("schema_migrations has %d rows, want 2", got)
	}

	// A new migration added later is the only one that runs.
	fsys = migrationSet(map[string]string{
		"0001_first.sql":  "CREATE TABLE first (id INTEGER PRIMARY KEY);",
		"0002_second.sql": "CREATE TABLE second (id INTEGER PRIMARY KEY);",
		"0003_third.sql":  "CREATE TABLE third (id INTEGER PRIMARY KEY);",
	})
	applied, err = migrateFS(ctx, db, fsys)
	if err != nil {
		t.Fatalf("migrateFS (third run): %v", err)
	}
	if applied != 1 {
		t.Errorf("third run applied %d migrations, want 1", applied)
	}
}

func TestMigrateAppliesInFilenameOrder(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// 0002 depends on the table 0001 creates, and the map is deliberately not
	// in order: only sorted filename order makes this work.
	fsys := migrationSet(map[string]string{
		"0010_last.sql":   "INSERT INTO ordered (step) VALUES (3);",
		"0002_middle.sql": "INSERT INTO ordered (step) VALUES (2);",
		"0001_first.sql":  "CREATE TABLE ordered (step INTEGER NOT NULL);\nINSERT INTO ordered (step) VALUES (1);",
	})

	if _, err := migrateFS(ctx, db, fsys); err != nil {
		t.Fatalf("migrateFS: %v", err)
	}

	rows, err := db.QueryContext(ctx, "SELECT step FROM ordered ORDER BY rowid")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var steps []int
	for rows.Next() {
		var step int
		if err := rows.Scan(&step); err != nil {
			t.Fatalf("scan: %v", err)
		}
		steps = append(steps, step)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}

	want := []int{1, 2, 3}
	if len(steps) != len(want) {
		t.Fatalf("got %v, want %v", steps, want)
	}
	for i := range want {
		if steps[i] != want[i] {
			t.Fatalf("got %v, want %v", steps, want)
		}
	}
}

// The transaction is the whole point: a file that fails halfway through must
// leave neither its own half-applied schema nor a schema_migrations row.
func TestFailingMigrationRollsBack(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// The failure is a runtime constraint violation rather than a syntax
	// error, so SQLite cannot reach it without having genuinely executed the
	// CREATE TABLE and the first INSERT. That is what makes the assertion
	// below a test of the rollback and not of statement preparation.
	fsys := migrationSet(map[string]string{
		"0001_good.sql": "CREATE TABLE good (id INTEGER PRIMARY KEY);",
		"0002_broken.sql": "CREATE TABLE half_applied (id INTEGER PRIMARY KEY);\n" +
			"INSERT INTO half_applied (id) VALUES (1);\n" +
			"INSERT INTO half_applied (id) VALUES (1);",
	})

	applied, err := migrateFS(ctx, db, fsys)
	if err == nil {
		t.Fatal("migrateFS succeeded on a broken migration, want an error")
	}
	if !strings.Contains(err.Error(), "0002_broken.sql") {
		t.Errorf("error %q does not name the failing migration", err)
	}
	if !strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
		t.Fatalf("error %q is not the expected constraint violation; the migration may have failed before executing anything", err)
	}
	if applied != 1 {
		t.Errorf("migrateFS applied %d migrations before failing, want 1", applied)
	}

	if tableExists(t, db, "half_applied") {
		t.Error("table created by the first statement of the failed migration survived; the transaction did not roll back")
	}
	if !tableExists(t, db, "good") {
		t.Error("the migration that succeeded before the failure was rolled back too")
	}

	// schema_migrations records the good migration and nothing else.
	var versions []string
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatalf("scan: %v", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	if len(versions) != 1 || versions[0] != "0001" {
		t.Errorf("schema_migrations = %v, want exactly [0001]", versions)
	}
}

// Failing on the very first migration must leave schema_migrations empty.
func TestFailingFirstMigrationLeavesSchemaMigrationsEmpty(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	fsys := migrationSet(map[string]string{
		"0001_broken.sql": "CREATE TABLE probe (id INTEGER PRIMARY KEY);\nSELECT nonexistent_function();",
	})

	if _, err := migrateFS(ctx, db, fsys); err == nil {
		t.Fatal("migrateFS succeeded on a broken migration, want an error")
	}

	if got := countRows(t, db, "schema_migrations"); got != 0 {
		t.Errorf("schema_migrations has %d rows after a failed migration, want 0", got)
	}
	if tableExists(t, db, "probe") {
		t.Error("table probe survived a rolled-back migration")
	}

	// Once the migration is fixed it applies cleanly — a failure is retried,
	// never skipped.
	fixed := migrationSet(map[string]string{
		"0001_broken.sql": "CREATE TABLE probe (id INTEGER PRIMARY KEY);",
	})
	applied, err := migrateFS(ctx, db, fixed)
	if err != nil {
		t.Fatalf("migrateFS after fixing: %v", err)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}
}

func TestMigrateRejectsBadFilenames(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{"not zero padded", "1_settings.sql"},
		{"five digits", "00001_settings.sql"},
		{"no version", "settings.sql"},
		{"camel case", "0001_Settings.sql"},
		{"kebab case", "0001-settings.sql"},
		{"spaces", "0001_two words.sql"},
		{"no name", "0001_.sql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			db := openTestDB(t)

			fsys := migrationSet(map[string]string{tt.file: "SELECT 1;"})

			if _, err := migrateFS(ctx, db, fsys); err == nil {
				t.Errorf("migrateFS accepted migration named %q, want an error", tt.file)
			}
		})
	}
}

func TestMigrateRejectsDuplicateVersions(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	fsys := migrationSet(map[string]string{
		"0001_settings.sql": "CREATE TABLE a (id INTEGER PRIMARY KEY);",
		"0001_other.sql":    "CREATE TABLE b (id INTEGER PRIMARY KEY);",
	})

	if _, err := migrateFS(ctx, db, fsys); err == nil {
		t.Error("migrateFS accepted two migrations with version 0001, want an error")
	}
}

// The embedded filesystem must contain migrations, not migrations plus a
// surprise: every .sql file in it has to be a well-formed migration.
func TestEmbeddedMigrationsAreWellFormed(t *testing.T) {
	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		t.Fatalf("loadMigrations: %v", err)
	}

	for i, m := range migrations {
		if m.sql == "" {
			t.Errorf("migration %s is empty", m.name)
		}
		if i > 0 && migrations[i-1].version >= m.version {
			t.Errorf("migration %s is not ordered after %s", m.name, migrations[i-1].name)
		}
	}
}
