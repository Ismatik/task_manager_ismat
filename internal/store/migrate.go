package store

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
)

// migrationsFS carries the schema migrations inside the binary. They are never
// read from disk at runtime: a Nexus binary copied to a machine with no source
// tree still knows how to bring a database up to date.
//
// The directory rather than the migrations/*.sql glob is embedded. The glob
// would work today — the set is not empty — but Go rejects an embed pattern
// that matches no files, so the glob form would turn an empty migration
// directory from a runtime no-op into a compile error. Embedding the directory
// costs nothing in exchange: loadMigrations applies the migrations/*.sql glob
// itself, so anything in the directory that is not a .sql file (the README) is
// carried in the binary but is never mistaken for a migration.
//
//go:embed migrations
var migrationsFS embed.FS

// migrationsDir is the path of the migration directory inside migrationsFS.
const migrationsDir = "migrations"

// migrationNamePattern enforces NNNN_snake_case.sql, zero-padded to four
// digits. The zero padding is what makes filename order, byte order and
// numeric order the same thing.
var migrationNamePattern = regexp.MustCompile(`^([0-9]{4})_[a-z0-9]+(_[a-z0-9]+)*\.sql$`)

// createSchemaMigrations bootstraps the bookkeeping table. It is deliberately
// not a migration itself: something has to exist before anything can be
// recorded as applied.
const createSchemaMigrations = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at TEXT NOT NULL
)`

// migration is one embedded .sql file.
type migration struct {
	version string // the four-digit prefix, e.g. "0001"
	name    string // the file name, e.g. "0001_settings.sql"
	sql     string // the file contents
}

// Migrate applies every embedded migration that this database has not seen yet,
// in filename order, and reports how many it applied. Running it on an
// up-to-date database applies nothing and returns 0.
//
// Each migration runs inside its own transaction together with the row that
// records it, so a migration either applies completely and is remembered, or
// fails and leaves no trace. Migration files must therefore not contain BEGIN,
// COMMIT or ROLLBACK of their own.
func Migrate(ctx context.Context, db *sql.DB) (int, error) {
	return migrateFS(ctx, db, migrationsFS)
}

// migrateFS is Migrate against an arbitrary filesystem, so that the runner can
// be tested against migration sets that the production binary must never ship
// — a deliberately broken one, most usefully.
func migrateFS(ctx context.Context, db *sql.DB, fsys fs.FS) (int, error) {
	migrations, err := loadMigrations(fsys)
	if err != nil {
		return 0, err
	}

	if _, err := db.ExecContext(ctx, createSchemaMigrations); err != nil {
		return 0, fmt.Errorf("store: creating schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, db)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		if err := applyMigration(ctx, db, m); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

// loadMigrations reads and validates the migration set, sorted by filename.
func loadMigrations(fsys fs.FS) ([]migration, error) {
	names, err := fs.Glob(fsys, migrationsDir+"/*.sql")
	if err != nil {
		return nil, fmt.Errorf("store: listing migrations: %w", err)
	}
	sort.Strings(names)

	migrations := make([]migration, 0, len(names))
	seen := make(map[string]string, len(names))

	for _, path := range names {
		base := path[len(migrationsDir)+1:]

		match := migrationNamePattern.FindStringSubmatch(base)
		if match == nil {
			return nil, fmt.Errorf("store: migration %q is not named NNNN_snake_case.sql", base)
		}
		version := match[1]

		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("store: migrations %q and %q share version %s", other, base, version)
		}
		seen[version] = base

		content, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil, fmt.Errorf("store: reading migration %q: %w", base, err)
		}

		migrations = append(migrations, migration{version: version, name: base, sql: string(content)})
	}

	return migrations, nil
}

// appliedVersions returns the set of migration versions already recorded.
func appliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("store: reading schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("store: reading schema_migrations: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reading schema_migrations: %w", err)
	}
	return applied, nil
}

// applyMigration runs one migration and records it, both inside a single
// transaction. SQLite makes DDL transactional, so a statement failing halfway
// through a file rolls the whole file back — including the schema_migrations
// row, which is why a failed migration is retried rather than skipped.
func applyMigration(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("store: migration %s: beginning transaction: %w", m.name, err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once the transaction is committed

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return fmt.Errorf("store: migration %s: %w", m.name, err)
	}

	const record = `INSERT INTO schema_migrations (version, name, applied_at)
	                VALUES (?, ?, strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))`
	if _, err := tx.ExecContext(ctx, record, m.version, m.name); err != nil {
		return fmt.Errorf("store: migration %s: recording: %w", m.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("store: migration %s: committing: %w", m.name, err)
	}
	return nil
}
