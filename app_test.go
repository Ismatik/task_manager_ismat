package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

// A first run against an empty XDG_DATA_HOME creates the database, applies
// every migration and seeds the four settings — asserted here rather than by
// launching the app, so that the startup path is covered by `go test`.
func TestOpenStoreCreatesMigratesAndSeeds(t *testing.T) {
	ctx := context.Background()
	dataHome := t.TempDir()
	t.Setenv("XDG_DATA_HOME", dataHome)

	path := filepath.Join(dataHome, "nexus", "nexus.db")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("the database already exists at %s", path)
	}

	db, err := openStore(ctx)
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer db.Close() //nolint:errcheck // the test is over either way

	if _, err := os.Stat(path); err != nil {
		t.Errorf("the database was not created at %s: %v", path, err)
	}

	t.Run("every migration is applied", func(t *testing.T) {
		// Migrating again applies nothing: the first run did all of it.
		applied, err := store.Migrate(ctx, db)
		if err != nil {
			t.Fatalf("Migrate: %v", err)
		}
		if applied != 0 {
			t.Errorf("%d migrations were still pending after openStore", applied)
		}
		// And the schema is really there.
		for _, table := range []string{"nodes", "tags", "node_tags", "time_entries", "habit_checks", "settings"} {
			if !tableExists(t, db, table) {
				t.Errorf("the %q table does not exist", table)
			}
		}
	})

	t.Run("the four settings are seeded", func(t *testing.T) {
		settings, err := store.NewSettingsRepo(db).All(ctx)
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		if len(settings) != 4 {
			t.Errorf("%d settings rows, want the four defaults: %v", len(settings), settings)
		}
		for _, key := range []string{store.KeyPalette, store.KeyTheme, store.KeyAccent, store.KeyLanguage} {
			if _, ok := settings[key]; !ok {
				t.Errorf("the %q setting was not seeded", key)
			}
		}
	})

	t.Run("a second run adopts the existing database", func(t *testing.T) {
		again, err := openStore(ctx)
		if err != nil {
			t.Fatalf("openStore, second run: %v", err)
		}
		defer again.Close() //nolint:errcheck // the test is over either way

		settings, err := store.NewSettingsRepo(again).All(ctx)
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		if len(settings) != 4 {
			t.Errorf("%d settings rows after a second run, want 4", len(settings))
		}
	})
}

// tableExists reports whether the schema holds a table of this name.
func tableExists(t *testing.T, db *sql.DB, name string) bool {
	t.Helper()

	var found string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&found)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		t.Fatalf("reading sqlite_master: %v", err)
	}
	return true
}

// A database that cannot be opened is an error the caller can act on, not a
// window onto nothing. main exits non-zero on it.
func TestOpenStoreReportsAnUnusableDataDirectory(t *testing.T) {
	// A regular file where the data directory has to be: MkdirAll fails.
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("writing the blocking file: %v", err)
	}
	t.Setenv("XDG_DATA_HOME", blocked)

	db, err := openStore(context.Background())
	if err == nil {
		db.Close() //nolint:errcheck // unreachable in a passing run
		t.Fatal("openStore succeeded with a file where the data directory should be")
	}
}

// The services are all constructed, with nothing left nil for a bound method to
// dereference at the first click.
func TestNewServicesConstructsEverything(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	db, err := openStore(context.Background())
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer db.Close() //nolint:errcheck // the test is over either way

	svc := reflect.ValueOf(newServices(db))
	for i := range svc.NumField() {
		if svc.Field(i).IsNil() {
			t.Errorf("Services.%s is nil", svc.Type().Field(i).Name)
		}
	}
}

// boundSurface is the Stage 2 binding, exactly (S2-07).
//
// It is an assertion and not documentation: a method added to App is bound the
// moment it is exported, and the frontend's generated client changes with it.
// Widening the surface should be a deliberate edit to this list, reviewed with
// the method.
var boundSurface = []string{
	"ArchiveNode",
	"Board",
	"CheckHabit",
	"CreateNode",
	"HabitStrip",
	"MoveNode",
	"MoveToColumn",
	"Progress",
	"RestoreNode",
	"Search",
	"SetAccent",
	"SetDue",
	"SetLanguage",
	"SetPalette",
	"SetTheme",
	"Settings",
	"TimerCurrent",
	"TimerStart",
	"TimerStop",
	"Tree",
	"UncheckHabit",
}

func TestTheBoundSurfaceIsExactlyTheStageTwoList(t *testing.T) {
	var got []string
	typ := reflect.TypeOf(&App{})
	for i := range typ.NumMethod() {
		got = append(got, typ.Method(i).Name)
	}
	slices.Sort(got)

	if !slices.Equal(got, boundSurface) {
		t.Errorf("the bound surface is\n\t%v\nwant\n\t%v", got, boundSurface)
	}
}

// ARCHITECTURE.md §4: every bound method returns (T, error).
//
// Wails turns the second return value into a rejected JS promise; without it
// the frontend cannot tell an empty result from a failure, and every rejection
// is supposed to reach a toast. Asserted by reflection so that a method added
// later cannot forget.
func TestEveryBoundMethodReturnsAnError(t *testing.T) {
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	typ := reflect.TypeOf(&App{})
	for i := range typ.NumMethod() {
		m := typ.Method(i)
		t.Run(m.Name, func(t *testing.T) {
			out := m.Type.NumOut()
			if out != 2 {
				t.Fatalf("%s returns %d values, want exactly (T, error)", m.Name, out)
			}
			if last := m.Type.Out(1); last != errorType {
				t.Errorf("%s's second return is %s, want error", m.Name, last)
			}
		})
	}
}

// The bound surface really answers: a freshly wired App returns the five
// columns, in order, with a card in the one it was created in.
//
// This is the Go half of "the app opens a window and the frontend can call
// Board()". The WebKit half cannot be asserted here and is the ticket's manual
// step; everything below the window is covered by this.
func TestBoardThroughTheBoundSurface(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())

	db, err := openStore(context.Background())
	if err != nil {
		t.Fatalf("openStore: %v", err)
	}
	defer db.Close() //nolint:errcheck // the test is over either way

	app := NewApp(newServices(db))

	board, err := app.Board()
	if err != nil {
		t.Fatalf("Board: %v", err)
	}
	if len(board) != 5 {
		t.Fatalf("Board returned %d columns, want five", len(board))
	}
	want := []string{"backlog", "week", "today", "doing", "done"}
	for i, c := range board {
		if string(c.Status) != want[i] {
			t.Errorf("column %d is %q, want %q", i, c.Status, want[i])
		}
		if c.Nodes == nil {
			t.Errorf("column %q has a nil card list; it must marshal as []", c.Status)
		}
	}

	t.Run("a created card lands on the board and moves across it", func(t *testing.T) {
		created, err := app.CreateNode(service.NewNode{Type: domain.NodeTypeTask, Title: "keyboard accept"})
		if err != nil {
			t.Fatalf("CreateNode: %v", err)
		}

		board, err := app.Board()
		if err != nil {
			t.Fatalf("Board: %v", err)
		}
		if len(board[0].Nodes) != 1 || board[0].Nodes[0].Node.ID != created.ID {
			t.Fatalf("the backlog column holds %+v, want the new card", board[0].Nodes)
		}

		if _, err := app.MoveToColumn(created.ID, domain.StatusDoing); err != nil {
			t.Fatalf("MoveToColumn: %v", err)
		}

		// C1/D13 through the binding: the move opened the timer.
		timer, err := app.TimerCurrent()
		if err != nil {
			t.Fatalf("TimerCurrent: %v", err)
		}
		if !timer.Running {
			t.Error("no timer is running after the card moved to Doing (C1, D13)")
		}

		strip, err := app.HabitStrip()
		if err != nil {
			t.Fatalf("HabitStrip: %v", err)
		}
		if len(strip) != 0 {
			t.Errorf("the habit strip holds %+v, want nothing — the card is a task", strip)
		}
	})

	t.Run("the settings read back as the seeded defaults", func(t *testing.T) {
		got, err := app.Settings()
		if err != nil {
			t.Fatalf("Settings: %v", err)
		}
		if got.Palette != domain.DefaultPalette || got.Theme != domain.DefaultTheme ||
			got.Language != domain.DefaultLanguage || got.Accent != domain.DefaultAccent {
			t.Errorf("Settings = %+v, want the seeded defaults", got)
		}
	})
}
