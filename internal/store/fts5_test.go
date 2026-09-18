package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

// objectCount returns how many objects of that name exist in the given schema
// ("main" or "temp") as seen from conn. Temp objects are private to a
// connection, so the connection matters and the query has to run on the same
// one the probe used.
func objectCount(t *testing.T, ctx context.Context, conn *sql.Conn, schema, name string) int {
	t.Helper()

	var n int
	query := "SELECT count(*) FROM " + schema + ".sqlite_master WHERE name = ?"
	if err := conn.QueryRowContext(ctx, query, name).Scan(&n); err != nil {
		t.Fatalf("counting %s.sqlite_master rows for %q: %v", schema, name, err)
	}
	return n
}

// schemaSnapshot is every object in the main schema, as one sorted string.
func schemaSnapshot(t *testing.T, db *sql.DB) string {
	t.Helper()

	rows, err := db.Query("SELECT type, name, coalesce(sql, '') FROM sqlite_master ORDER BY type, name")
	if err != nil {
		t.Fatalf("reading sqlite_master: %v", err)
	}
	defer rows.Close()

	var b strings.Builder
	for rows.Next() {
		var typ, name, ddl string
		if err := rows.Scan(&typ, &name, &ddl); err != nil {
			t.Fatalf("scanning sqlite_master: %v", err)
		}
		b.WriteString(typ + "\t" + name + "\t" + ddl + "\n")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("reading sqlite_master: %v", err)
	}
	return b.String()
}

// The headline question, and the one the decision block in fts5.go answers.
// This machine's driver has the module, so the probe must say so — and if a
// driver upgrade ever takes FTS5 away, this test is what notices.
func TestFTS5AvailableOnThisDriver(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	available, err := FTS5Available(ctx, db)
	if err != nil {
		t.Fatalf("FTS5Available: %v", err)
	}
	if !available {
		t.Fatalf("FTS5Available = false, want true: the decision block on FTS5Available records " +
			"modernc.org/sqlite v1.53.0 as having the module, and S1-17 was planned on it")
	}
}

// The observed behaviour of the AVAILABLE branch, asserted directly rather than
// through the probe: a real fts5 table, a real MATCH, the expected row back.
// The probe could in principle return true for the wrong reason; this cannot.
func TestFTS5MatchReturnsTheIndexedRow(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup

	if _, err := conn.ExecContext(ctx,
		`CREATE VIRTUAL TABLE temp.search_probe USING fts5(title, description_md)`); err != nil {
		t.Fatalf("CREATE VIRTUAL TABLE ... USING fts5: %v", err)
	}

	rows := []struct{ title, body string }{
		{"deploy the release", "roll the build out to production"},
		{"write the changelog", "summarise what shipped this week"},
	}
	for _, r := range rows {
		if _, err := conn.ExecContext(ctx,
			`INSERT INTO temp.search_probe (title, description_md) VALUES (?, ?)`, r.title, r.body); err != nil {
			t.Fatalf("inserting %q: %v", r.title, err)
		}
	}

	for _, tc := range []struct {
		name  string
		query string
		want  string
	}{
		{name: "term in the title", query: "changelog", want: "write the changelog"},
		{name: "term in the description", query: "production", want: "deploy the release"},
		{name: "column-qualified term", query: "title:deploy", want: "deploy the release"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			err := conn.QueryRowContext(ctx,
				`SELECT title FROM temp.search_probe WHERE search_probe MATCH ?`, tc.query).Scan(&got)
			if err != nil {
				t.Fatalf("MATCH %q: %v", tc.query, err)
			}
			if got != tc.want {
				t.Errorf("MATCH %q returned %q, want %q", tc.query, got, tc.want)
			}
		})
	}

	t.Run("a term that is not indexed matches nothing", func(t *testing.T) {
		var got string
		err := conn.QueryRowContext(ctx,
			`SELECT title FROM temp.search_probe WHERE search_probe MATCH ?`, "kangaroo").Scan(&got)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("MATCH on an absent term returned (%q, %v), want sql.ErrNoRows", got, err)
		}
	})

	if _, err := conn.ExecContext(ctx, `DROP TABLE temp.search_probe`); err != nil {
		t.Fatalf("DROP TABLE: %v", err)
	}
}

// Two calls on the same database must agree. A probe whose answer depends on
// whether it has already run is a probe that left something behind.
func TestFTS5AvailableIsConsistentAcrossCalls(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	first, err := FTS5Available(ctx, db)
	if err != nil {
		t.Fatalf("FTS5Available (first call): %v", err)
	}

	second, err := FTS5Available(ctx, db)
	if err != nil {
		t.Fatalf("FTS5Available (second call): %v", err)
	}

	if first != second {
		t.Errorf("FTS5Available returned %v then %v; the probe is not repeatable", first, second)
	}
}

// The scratch table must be gone from the connection it was created on, not
// merely invisible from another one. Hence the pinned connection.
func TestFTS5ProbeDropsItsScratchTable(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup

	if _, err := fts5AvailableOnConn(ctx, conn); err != nil {
		t.Fatalf("fts5AvailableOnConn: %v", err)
	}

	if n := objectCount(t, ctx, conn, "temp", fts5ProbeTable); n != 0 {
		t.Errorf("temp.sqlite_master still has %d row(s) for %q; the scratch table was not dropped", n, fts5ProbeTable)
	}
	if n := objectCount(t, ctx, conn, "main", fts5ProbeTable); n != 0 {
		t.Errorf("main.sqlite_master has %d row(s) for %q; the probe wrote to the database file", n, fts5ProbeTable)
	}
}

// Safe to call on the live database: an already-migrated schema and its
// bookkeeping must come out the other side byte-identical.
func TestFTS5ProbeLeavesAnExistingSchemaUndisturbed(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	if _, err := Migrate(ctx, db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	if _, err := NewSettingsRepo(db).SeedDefaults(ctx); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}

	beforeSchema := schemaSnapshot(t, db)
	beforeMigrations := countRows(t, db, "schema_migrations")
	beforeSettings := countRows(t, db, "settings")

	if _, err := FTS5Available(ctx, db); err != nil {
		t.Fatalf("FTS5Available: %v", err)
	}

	if after := schemaSnapshot(t, db); after != beforeSchema {
		t.Errorf("the probe changed the main schema:\nbefore:\n%s\nafter:\n%s", beforeSchema, after)
	}
	if after := countRows(t, db, "schema_migrations"); after != beforeMigrations {
		t.Errorf("schema_migrations has %d rows after the probe, want %d", after, beforeMigrations)
	}
	if after := countRows(t, db, "settings"); after != beforeSettings {
		t.Errorf("settings has %d rows after the probe, want %d", after, beforeSettings)
	}

	// And the probe must not have broken the migration bookkeeping either.
	applied, err := Migrate(ctx, db)
	if err != nil {
		t.Fatalf("Migrate after the probe: %v", err)
	}
	if applied != 0 {
		t.Errorf("Migrate after the probe applied %d migrations, want 0", applied)
	}
}

// A failure that is not "no such module" must surface as an error. This is the
// case that protects search from being switched off forever by an unrelated
// database problem, so it is tested with a real one: a name collision in the
// temp schema, which fails the CREATE without FTS5 being involved at all.
func TestFTS5ProbeReportsNonModuleFailuresAsErrors(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer conn.Close() //nolint:errcheck // test cleanup

	if _, err := conn.ExecContext(ctx,
		`CREATE TABLE temp.`+fts5ProbeTable+` (squatter TEXT)`); err != nil {
		t.Fatalf("creating the colliding table: %v", err)
	}

	available, err := fts5AvailableOnConn(ctx, conn)
	if err == nil {
		t.Fatalf("fts5AvailableOnConn = (%v, nil), want an error: a failing CREATE that is not "+
			"a missing module must never be answered with a quiet false", available)
	}
	if available {
		t.Errorf("fts5AvailableOnConn returned available = true alongside an error")
	}

	// The probe must not have dropped a table it did not create.
	if n := objectCount(t, ctx, conn, "temp", fts5ProbeTable); n != 1 {
		t.Errorf("the colliding table is gone (%d rows in temp.sqlite_master); the probe dropped "+
			"an object it does not own", n)
	}
}

// The connection cannot even be acquired. Also an error, never a false.
func TestFTS5AvailableOnAClosedDatabaseIsAnError(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	available, err := FTS5Available(ctx, db)
	if err == nil {
		t.Fatalf("FTS5Available on a closed database = (%v, nil), want an error", available)
	}
	if available {
		t.Errorf("FTS5Available returned available = true alongside an error")
	}
}

// The classifier is the single place where "unavailable" is decided, so every
// message that must and must not mean it gets a case of its own.
func TestIsNoSuchFTS5Module(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "the exact sqlite message",
			err:  errors.New("no such module: fts5"),
			want: true,
		},
		{
			name: "wrapped and prefixed by the driver",
			err:  errors.New("SQL logic error: no such module: fts5 (1)"),
			want: true,
		},
		{
			name: "upper case",
			err:  errors.New("SQL logic error: No such module: FTS5 (1)"),
			want: true,
		},
		{
			name: "a different missing module is not our answer",
			err:  errors.New("no such module: rtree"),
			want: false,
		},
		{
			name: "read-only database",
			err:  errors.New("attempt to write a readonly database (8)"),
			want: false,
		},
		{
			name: "disk full",
			err:  errors.New("database or disk is full (13)"),
			want: false,
		},
		{
			name: "permission denied",
			err:  errors.New("unable to open database file: permission denied (14)"),
			want: false,
		},
		{
			name: "an unrelated failure that merely mentions fts5",
			err:  errors.New("disk I/O error while writing the fts5 index (10)"),
			want: false,
		},
		{
			name: "table already exists",
			err:  errors.New("table nexus_fts5_probe already exists (1)"),
			want: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNoSuchFTS5Module(tc.err); got != tc.want {
				t.Errorf("isNoSuchFTS5Module(%q) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
