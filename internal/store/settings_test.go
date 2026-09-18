package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

// migratedDB opens a fresh temp database and brings it up to date, so every
// test below also exercises 0001_settings.sql going through the S0-08 runner.
func migratedDB(t *testing.T) *sql.DB {
	t.Helper()

	db := openTestDB(t)
	if _, err := Migrate(context.Background(), db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	return db
}

func TestSettingsMigrationCreatesTheTable(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)

	if !tableExists(t, db, "settings") {
		t.Fatal("migration 0001 did not create the settings table")
	}

	// key is the primary key and value is NOT NULL.
	if _, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES ('k', 'a'), ('k', 'b')"); err == nil {
		t.Error("two rows with the same key were accepted; key is not a primary key")
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO settings (key, value) VALUES ('nullable', NULL)"); err == nil {
		t.Error("a NULL value was accepted; value is not NOT NULL")
	}

	// The migration seeds nothing: seeding is the repository's job.
	if got := countRows(t, db, "settings"); got != 0 {
		t.Errorf("settings has %d rows straight after migrating, want 0", got)
	}
}

func TestSetGetRoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepo(migratedDB(t))

	if err := repo.Set(ctx, KeyPalette, "studio"); err != nil {
		t.Fatalf("Set: %v", err)
	}

	value, found, err := repo.Get(ctx, KeyPalette)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("Get found = false after Set, want true")
	}
	if value != "studio" {
		t.Errorf("Get = %q, want %q", value, "studio")
	}
}

func TestSetUpsertsRatherThanDuplicating(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	for _, value := range []string{"aurora", "studio", "aurora"} {
		if err := repo.Set(ctx, KeyPalette, value); err != nil {
			t.Fatalf("Set(%q): %v", value, err)
		}
	}

	if got := countRows(t, db, "settings"); got != 1 {
		t.Errorf("settings has %d rows after three Sets of one key, want 1", got)
	}

	value, _, err := repo.Get(ctx, KeyPalette)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if value != "aurora" {
		t.Errorf("Get = %q, want the last value written, %q", value, "aurora")
	}
}

// A missing key is a normal answer, not a failure: "" and found == false.
func TestGetMissingKeyIsNotAnError(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepo(migratedDB(t))

	value, found, err := repo.Get(ctx, "never-set")
	if err != nil {
		t.Fatalf("Get on a missing key returned an error: %v", err)
	}
	if found {
		t.Error("Get found = true for a key that was never set")
	}
	if value != "" {
		t.Errorf("Get = %q for a missing key, want the empty string", value)
	}
}

// accent defaults to "" and that has to survive the round trip: an empty value
// means "use the palette's own --accent", which is different from being unset.
func TestEmptyValueIsStoredAndDistinctFromMissing(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepo(migratedDB(t))

	if err := repo.Set(ctx, KeyAccent, ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	value, found, err := repo.Get(ctx, KeyAccent)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Error("an explicitly stored empty value reads back as missing")
	}
	if value != "" {
		t.Errorf("Get = %q, want the empty string", value)
	}
}

func TestEmptyKeyIsRejected(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	if err := repo.Set(ctx, "", "value"); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Set(\"\") error = %v, want ErrEmptyKey", err)
	}
	if _, _, err := repo.Get(ctx, ""); !errors.Is(err, ErrEmptyKey) {
		t.Errorf("Get(\"\") error = %v, want ErrEmptyKey", err)
	}
	if got := countRows(t, db, "settings"); got != 0 {
		t.Errorf("settings has %d rows after a rejected Set, want 0", got)
	}
}

func TestAll(t *testing.T) {
	ctx := context.Background()
	repo := NewSettingsRepo(migratedDB(t))

	empty, err := repo.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if empty == nil {
		t.Error("All returned a nil map for an empty table, want an empty one")
	}
	if len(empty) != 0 {
		t.Errorf("All returned %d settings from an empty table, want 0", len(empty))
	}

	want := map[string]string{KeyPalette: "studio", KeyTheme: "light", KeyAccent: ""}
	for key, value := range want {
		if err := repo.Set(ctx, key, value); err != nil {
			t.Fatalf("Set(%q): %v", key, err)
		}
	}

	got, err := repo.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("All returned %d settings, want %d: %v", len(got), len(want), got)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("All()[%q] = %q, want %q", key, got[key], value)
		}
	}
}

func TestSeedDefaults(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	seeded, err := repo.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	if seeded != 4 {
		t.Errorf("SeedDefaults inserted %d settings, want 4", seeded)
	}

	want := map[string]string{
		KeyPalette:  "aurora",
		KeyTheme:    "dark",
		KeyAccent:   "",
		KeyLanguage: "en",
	}
	got, err := repo.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("All returned %d settings, want %d: %v", len(got), len(want), got)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("All()[%q] = %q, want %q", key, got[key], value)
		}
	}
}

// Seeding runs on every start, so running it twice must change nothing.
func TestSeedDefaultsIsIdempotent(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	if _, err := repo.SeedDefaults(ctx); err != nil {
		t.Fatalf("first SeedDefaults: %v", err)
	}

	seeded, err := repo.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("second SeedDefaults: %v", err)
	}
	if seeded != 0 {
		t.Errorf("second SeedDefaults inserted %d settings, want 0", seeded)
	}
	if got := countRows(t, db, "settings"); got != 4 {
		t.Errorf("settings has %d rows after seeding twice, want exactly 4", got)
	}
}

// The rule that matters: seeding tops up what is missing, it never resets what
// the user chose.
func TestSeedDefaultsNeverOverwritesAUserValue(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	if _, err := repo.SeedDefaults(ctx); err != nil {
		t.Fatalf("first SeedDefaults: %v", err)
	}

	// The user switches palette and language, and clears a value entirely.
	if err := repo.Set(ctx, KeyPalette, "studio"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := repo.Set(ctx, KeyLanguage, "ru"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if err := repo.Set(ctx, KeyTheme, ""); err != nil {
		t.Fatalf("Set: %v", err)
	}

	seeded, err := repo.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("second SeedDefaults: %v", err)
	}
	if seeded != 0 {
		t.Errorf("SeedDefaults inserted %d settings over existing ones, want 0", seeded)
	}

	want := map[string]string{
		KeyPalette:  "studio",
		KeyTheme:    "",
		KeyAccent:   "",
		KeyLanguage: "ru",
	}
	got, err := repo.All(ctx)
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("settings has %d rows, want 4: %v", len(got), got)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("All()[%q] = %q, want %q — seeding overwrote a user value", key, got[key], value)
		}
	}
}

// A default that is missing because the user (or a future migration) removed it
// comes back on the next start, without disturbing the others.
func TestSeedDefaultsRestoresOnlyMissingKeys(t *testing.T) {
	ctx := context.Background()
	db := migratedDB(t)
	repo := NewSettingsRepo(db)

	if _, err := repo.SeedDefaults(ctx); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	if err := repo.Set(ctx, KeyPalette, "studio"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM settings WHERE key = ?", KeyLanguage); err != nil {
		t.Fatalf("delete: %v", err)
	}

	seeded, err := repo.SeedDefaults(ctx)
	if err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	if seeded != 1 {
		t.Errorf("SeedDefaults inserted %d settings, want 1", seeded)
	}

	language, found, err := repo.Get(ctx, KeyLanguage)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found || language != "en" {
		t.Errorf("language = %q (found %v), want %q", language, found, "en")
	}
	palette, _, err := repo.Get(ctx, KeyPalette)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if palette != "studio" {
		t.Errorf("palette = %q, want the user's %q", palette, "studio")
	}
}
