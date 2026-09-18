package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	// modernc.org/sqlite is the pure-Go SQLite driver. It is registered under
	// the name "sqlite" (not "sqlite3"). No cgo driver may ever replace it.
	_ "modernc.org/sqlite"
)

const (
	// driverName is the name modernc.org/sqlite registers itself under.
	driverName = "sqlite"

	// appDirName is the per-application directory under the XDG data home.
	appDirName = "nexus"

	// dbFileName is the single SQLite file holding all of the user's data.
	dbFileName = "nexus.db"

	// busyTimeoutMS is how long a connection waits for a lock held by another
	// connection before giving up with SQLITE_BUSY. WAL keeps readers out of
	// the writer's way, so this only has to cover writer-versus-writer.
	busyTimeoutMS = 5000

	// dirPerm is the mode of the created data directory. The database is
	// personal data on a single-user desktop: owner only.
	dirPerm = 0o700
)

// DefaultPath returns the location of the database file:
// $XDG_DATA_HOME/nexus/nexus.db, falling back to ~/.local/share/nexus/nexus.db
// when XDG_DATA_HOME is unset or, per the XDG base directory specification,
// is not an absolute path.
func DefaultPath() (string, error) {
	if dataHome := os.Getenv("XDG_DATA_HOME"); filepath.IsAbs(dataHome) {
		return filepath.Join(dataHome, appDirName, dbFileName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("store: locating home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", appDirName, dbFileName), nil
}

// dsn builds the driver connection string for path. The pragmas travel in the
// DSN rather than being executed after connecting because journal_mode is the
// only one of the three that is persisted in the file: foreign_keys and
// busy_timeout are per-connection settings, and database/sql owns a pool of
// connections it opens whenever it likes. Putting them in the DSN is what makes
// them true of every connection in that pool.
func dsn(path string) string {
	params := url.Values{}
	params.Add("_pragma", "journal_mode(WAL)")
	params.Add("_pragma", "foreign_keys(1)")
	params.Add("_pragma", fmt.Sprintf("busy_timeout(%d)", busyTimeoutMS))

	u := url.URL{Scheme: "file", Path: path, RawQuery: params.Encode()}
	return u.String()
}

// Open opens the database at path, creating its parent directory (and the file
// itself) if they do not exist. An empty path means DefaultPath.
//
// The returned handle has journal_mode=WAL, foreign_keys=ON and a busy timeout
// on every connection; Open verifies that they read back as expected rather
// than assuming. It does not run migrations — call Migrate for that.
//
// The caller owns the handle and must Close it.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	if path == "" {
		var err error
		if path, err = DefaultPath(); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil {
		return nil, fmt.Errorf("store: creating data directory: %w", err)
	}

	db, err := sql.Open(driverName, dsn(path))
	if err != nil {
		return nil, fmt.Errorf("store: opening %s: %w", path, err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: connecting to %s: %w", path, err)
	}

	if err := verifyPragmas(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// verifyPragmas reads the connection settings back out of SQLite. A pragma that
// silently failed to apply would cost us referential integrity or leave writers
// erroring out on contention instead of waiting, so it is worth one round trip
// at startup.
func verifyPragmas(ctx context.Context, db *sql.DB) error {
	var foreignKeys int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		return fmt.Errorf("store: reading foreign_keys pragma: %w", err)
	}
	if foreignKeys != 1 {
		return fmt.Errorf("store: foreign_keys pragma is %d, want 1", foreignKeys)
	}

	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		return fmt.Errorf("store: reading journal_mode pragma: %w", err)
	}
	if journalMode != "wal" {
		return fmt.Errorf("store: journal_mode pragma is %q, want %q", journalMode, "wal")
	}

	var busyTimeout int
	if err := db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		return fmt.Errorf("store: reading busy_timeout pragma: %w", err)
	}
	if busyTimeout != busyTimeoutMS {
		return fmt.Errorf("store: busy_timeout pragma is %d, want %d", busyTimeout, busyTimeoutMS)
	}

	return nil
}
