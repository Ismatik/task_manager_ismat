package store

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// tempDBPath returns a path inside t.TempDir() for a database file that does
// not exist yet, in a directory that does not exist yet either — so every test
// also exercises the "create the parent directory on open" requirement.
func tempDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "data", "nexus", "nexus.db")
}

func TestOpenCreatesParentDirectoryAndFile(t *testing.T) {
	ctx := context.Background()
	path := tempDBPath(t)

	if _, err := os.Stat(filepath.Dir(path)); !os.IsNotExist(err) {
		t.Fatalf("precondition: %s already exists", filepath.Dir(path))
	}

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	// SQLite creates the file lazily, so force a write before looking for it.
	if _, err := db.ExecContext(ctx, "CREATE TABLE probe (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("database file %s: %v", path, err)
	}
}

// Pragmas are per-connection and database/sql opens connections whenever it
// likes, so read them back through the pool rather than trusting the DSN.
func TestOpenAppliesPragmas(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	var foreignKeys int
	if err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("PRAGMA foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("foreign_keys = %d, want 1", foreignKeys)
	}

	var journalMode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("journal_mode = %q, want %q", journalMode, "wal")
	}

	var busyTimeout int
	if err := db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("PRAGMA busy_timeout: %v", err)
	}
	if busyTimeout != busyTimeoutMS {
		t.Errorf("busy_timeout = %d, want %d", busyTimeout, busyTimeoutMS)
	}
}

// foreign_keys=ON has to hold on every connection in the pool, not just the
// first one, so force several to be open at once.
func TestForeignKeysAreEnforced(t *testing.T) {
	ctx := context.Background()

	db, err := Open(ctx, tempDBPath(t))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	const schema = `
CREATE TABLE parent (id INTEGER PRIMARY KEY);
CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id));`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	// Hold a connection open so the insert below is forced onto a second one.
	held, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	defer held.Close()

	if _, err := db.ExecContext(ctx, "INSERT INTO child (id, parent_id) VALUES (1, 42)"); err == nil {
		t.Error("insert violating a foreign key succeeded; foreign_keys is not enforced on this connection")
	}
}

func TestOpenRejectsAnUnusablePath(t *testing.T) {
	ctx := context.Background()

	// A regular file where a directory has to be: MkdirAll must fail.
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	db, err := Open(ctx, filepath.Join(blocker, "nexus", "nexus.db"))
	if err == nil {
		db.Close()
		t.Fatal("Open succeeded on a path that cannot exist, want an error")
	}
}

func TestDefaultPathUsesXDGDataHome(t *testing.T) {
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	got, err := DefaultPath()
	if err != nil {
		t.Fatalf("DefaultPath: %v", err)
	}
	if want := filepath.Join(dataHome, "nexus", "nexus.db"); got != want {
		t.Errorf("DefaultPath() = %q, want %q", got, want)
	}
}

func TestDefaultPathFallsBackToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	tests := []struct {
		name      string
		dataHome  string
		setEnvVar bool
	}{
		{name: "unset", setEnvVar: false},
		{name: "empty", dataHome: "", setEnvVar: true},
		{name: "relative is ignored per the XDG spec", dataHome: "relative/share", setEnvVar: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.Setenv registers the restore, including for the unset case.
			t.Setenv("XDG_DATA_HOME", tt.dataHome)
			if !tt.setEnvVar {
				if err := os.Unsetenv("XDG_DATA_HOME"); err != nil {
					t.Fatalf("Unsetenv: %v", err)
				}
			}

			got, err := DefaultPath()
			if err != nil {
				t.Fatalf("DefaultPath: %v", err)
			}
			if want := filepath.Join(home, ".local", "share", "nexus", "nexus.db"); got != want {
				t.Errorf("DefaultPath() = %q, want %q", got, want)
			}
		})
	}
}
