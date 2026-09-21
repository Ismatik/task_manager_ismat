package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"nexus/internal/domain"
)

// The setting keys Nexus reads at startup (D6, design/README.md).
const (
	// KeyPalette selects the design palette — one of domain.Palettes(), which
	// is where the allowed values are defined. It ends up on
	// <html data-palette="...">.
	KeyPalette = "palette"

	// KeyTheme selects the theme — one of domain.Themes(). It drives the
	// "dark" class on <html>.
	KeyTheme = "theme"

	// KeyAccent is the user's accent colour override. Empty means "use the
	// palette's own --accent", which is why the column is NOT NULL rather than
	// nullable: absence is spelled "", not NULL.
	KeyAccent = "accent"

	// KeyLanguage selects the UI language — one of domain.Languages().
	KeyLanguage = "language"
)

// defaultSettings are the values seeded on first run (D6).
//
// The values come from internal/domain rather than being spelled out here: the
// settings service validates against the same constants, and a palette name
// written in two packages is one edit away from a default its own validator
// rejects. This file names no palette, theme or language literally.
var defaultSettings = []struct{ key, value string }{
	{KeyPalette, domain.DefaultPalette.String()},
	{KeyTheme, domain.DefaultTheme.String()},
	{KeyAccent, domain.DefaultAccent},
	{KeyLanguage, domain.DefaultLanguage.String()},
}

// ErrEmptyKey is returned when a setting key is the empty string. SQLite would
// happily store a row under "", so the guard is ours to make.
var ErrEmptyKey = errors.New("store: setting key is empty")

// SettingsRepo reads and writes the settings table.
type SettingsRepo struct {
	db *sql.DB
}

// NewSettingsRepo returns a repository over db. The settings table must already
// exist — run Migrate first.
func NewSettingsRepo(db *sql.DB) *SettingsRepo {
	return &SettingsRepo{db: db}
}

// Get returns the value stored under key. A key that has never been set is not
// an error: it returns found == false and a nil error, and the caller decides
// what the absence means.
func (r *SettingsRepo) Get(ctx context.Context, key string) (string, bool, error) {
	if key == "" {
		return "", false, ErrEmptyKey
	}

	var value string
	err := r.db.QueryRowContext(ctx, "SELECT value FROM settings WHERE key = ?", key).Scan(&value)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", false, nil
	case err != nil:
		return "", false, fmt.Errorf("store: reading setting %q: %w", key, err)
	}
	return value, true, nil
}

// Set writes value under key, replacing whatever was there. Writing a key for
// the second time is an update, not a duplicate and not an error.
func (r *SettingsRepo) Set(ctx context.Context, key, value string) error {
	if key == "" {
		return ErrEmptyKey
	}

	const upsert = `INSERT INTO settings (key, value) VALUES (?, ?)
	                ON CONFLICT (key) DO UPDATE SET value = excluded.value`
	if _, err := r.db.ExecContext(ctx, upsert, key, value); err != nil {
		return fmt.Errorf("store: writing setting %q: %w", key, err)
	}
	return nil
}

// All returns every setting. The map is never nil, so an empty settings table
// reads back as an empty map rather than something the caller has to guard.
func (r *SettingsRepo) All(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT key, value FROM settings")
	if err != nil {
		return nil, fmt.Errorf("store: reading settings: %w", err)
	}
	defer rows.Close()

	settings := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("store: reading settings: %w", err)
		}
		settings[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: reading settings: %w", err)
	}
	return settings, nil
}

// SeedDefaults inserts any default setting that is not present yet and reports
// how many it inserted. It is safe to call on every start: a key the user has
// already changed is left exactly as they left it, which is the difference
// between seeding and resetting.
func (r *SettingsRepo) SeedDefaults(ctx context.Context) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("store: seeding settings: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op once the transaction is committed

	// DO NOTHING rather than DO UPDATE: this is the whole "never overwrite a
	// user-set value" rule, expressed in one clause.
	const seed = `INSERT INTO settings (key, value) VALUES (?, ?)
	              ON CONFLICT (key) DO NOTHING`

	seeded := 0
	for _, d := range defaultSettings {
		result, err := tx.ExecContext(ctx, seed, d.key, d.value)
		if err != nil {
			return 0, fmt.Errorf("store: seeding setting %q: %w", d.key, err)
		}
		inserted, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("store: seeding setting %q: %w", d.key, err)
		}
		seeded += int(inserted)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("store: seeding settings: %w", err)
	}
	return seeded, nil
}
