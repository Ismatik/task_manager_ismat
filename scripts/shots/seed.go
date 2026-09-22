//go:build shots

// Command seed prepares a throwaway Nexus database for `make shots`.
//
// # Why it exists at all
//
// `make shots` photographs the running binary in every palette, theme and
// language. Input cannot be driven on this machine — there is no xdotool, no
// xte, no wmctrl and no window manager (E4) — so a state cannot be reached by
// clicking the header. The palette, the theme and the language are rows in the
// `settings` table (D6), so the state is reached by WRITING THOSE ROWS before
// the window opens. That is all this program does.
//
// # It adds no production code and no second spelling of the schema
//
// Everything here goes through internal/store and internal/service, which are
// the only writers to this schema in the project. There is no `sqlite3` binary
// on this machine, and there must not be a second writer in the repository: a
// program that spelled the INSERTs itself would be the schema written twice.
//
// # It is not part of the application, and not part of the gates
//
// The `shots` build tag keeps it out of `go list ./...` entirely, so $(GOPKGS)
// — and therefore gates 1 and 2, and the `make cover` bars — never see it. It
// is compiled only by `go run -tags shots ./scripts/shots`.
//
// # It never touches the user's database
//
// It writes to whatever store.DefaultPath resolves, and `make shots` exports
// XDG_DATA_HOME to a throwaway directory under build/. Where the data lives has
// one spelling — store.DefaultPath — and this program does not restate it.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"

	"nexus/internal/domain"
	"nexus/internal/service"
	"nexus/internal/store"
)

func main() {
	var (
		palette  = flag.String("palette", domain.DefaultPalette.String(), "palette to persist")
		theme    = flag.String("theme", domain.DefaultTheme.String(), "theme to persist")
		language = flag.String("language", domain.DefaultLanguage.String(), "language to persist")
		board    = flag.Bool("board", false, "also seed a card in every column, a project and a habit")
		matrix   = flag.Bool("matrix", false, "print every palette/theme/language triple and exit")
	)
	flag.Parse()

	if *matrix {
		printMatrix(os.Stdout)
		return
	}

	if err := run(context.Background(), *palette, *theme, *language, *board); err != nil {
		log.Printf("seed: %v", err)
		os.Exit(1)
	}
}

// printMatrix writes one "palette theme language" line per state `make shots`
// photographs.
//
// The capture script iterates over THIS, rather than over three lists of its
// own. Which palettes, themes and languages exist is internal/domain's answer —
// it is already the single spelling for the settings service, the store's
// defaults and the frontend's pickers — and a shell array repeating them would
// be a fourth. A palette added to domain.Palettes() is photographed by itself,
// with no edit here and none in the Makefile.
func printMatrix(out io.Writer) {
	for _, palette := range domain.Palettes() {
		for _, theme := range domain.Themes() {
			for _, language := range domain.Languages() {
				fmt.Fprintf(out, "%s %s %s\n", palette, theme, language)
			}
		}
	}
}

func run(ctx context.Context, palette, theme, language string, board bool) error {
	path, err := store.DefaultPath()
	if err != nil {
		return err
	}

	db, err := store.Open(ctx, path)
	if err != nil {
		return err
	}
	defer db.Close() //nolint:errcheck // nothing useful to do at exit

	if _, err := store.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrating %s: %w", path, err)
	}

	settings := store.NewSettingsRepo(db)
	if _, err := settings.SeedDefaults(ctx); err != nil {
		return fmt.Errorf("seeding defaults: %w", err)
	}

	// Each value is checked against the domain's own set before it is stored.
	// Seeding is not a way around validation: an unknown palette would
	// photograph a state the running application refuses to enter, and the
	// picture would be evidence of nothing.
	for _, pair := range []struct{ key, value string }{
		{store.KeyPalette, palette},
		{store.KeyTheme, theme},
		{store.KeyLanguage, language},
	} {
		if err := validate(pair.key, pair.value); err != nil {
			return err
		}
		if err := settings.Set(ctx, pair.key, pair.value); err != nil {
			return err
		}
	}

	if !board {
		return nil
	}
	return seedBoard(ctx, db)
}

// validate refuses a value internal/domain does not know. The three sets are
// the domain's, so this function names no palette, theme or language itself.
func validate(key, value string) error {
	var ok bool
	switch key {
	case store.KeyPalette:
		ok = domain.Palette(value).Valid()
	case store.KeyTheme:
		ok = domain.Theme(value).Valid()
	case store.KeyLanguage:
		ok = domain.Language(value).Valid()
	default:
		return fmt.Errorf("unknown setting %q", key)
	}
	if !ok {
		return fmt.Errorf("%s: %q is not a value internal/domain knows", key, value)
	}
	return nil
}

// seedBoard fills an empty database with one card per Kanban column, a project
// with children so a progress bar has something to draw, and a habit so the
// habits strip is not empty.
//
// It is a no-op on a database that already holds nodes, so re-running
// `make shots` over a kept fixture directory does not multiply the cards.
func seedBoard(ctx context.Context, db *sql.DB) error {
	nodes := store.NewNodeRepo(db)
	existing, err := nodes.ListAll(ctx, true)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	var (
		clock  = service.Clock(service.SystemClock)
		newID  = func() string { return uuid.NewString() }
		begin  = service.NewBeginner(db)
		timers = service.NewTimerService(begin, nodes, store.NewTimeEntryRepo(db), clock, newID)
		tasks  = service.NewTaskService(begin, nodes, store.NewTagRepo(db), timers, clock, newID)
	)

	// One card per column, in domain.Statuses() order. The set is Go's and is
	// not restated here, so a sixth column would be photographed by itself.
	statuses := domain.Statuses()
	for i, status := range statuses {
		if _, err := tasks.CreateNode(ctx, service.NewNode{
			Type:     domain.NodeTypeTask,
			Title:    fmt.Sprintf("Проверка вёрстки — карточка %d", i+1),
			Status:   status,
			Priority: domain.Priority(1 + i%len(domain.Priorities())),
		}); err != nil {
			return fmt.Errorf("creating a %s card: %w", status, err)
		}
	}

	project, err := tasks.CreateNode(ctx, service.NewNode{
		Type:     domain.NodeTypeProject,
		Title:    "Проект с достаточно длинным названием",
		Status:   statuses[0],
		Priority: domain.Priority2,
	})
	if err != nil {
		return fmt.Errorf("creating the project: %w", err)
	}
	for i, status := range []domain.Status{statuses[0], statuses[len(statuses)-1]} {
		if _, err := tasks.CreateNode(ctx, service.NewNode{
			ParentID: &project.ID,
			Type:     domain.NodeTypeTask,
			Title:    fmt.Sprintf("Подзадача %d", i+1),
			Status:   status,
			Priority: domain.Priority4,
		}); err != nil {
			return fmt.Errorf("creating a child: %w", err)
		}
	}

	daily := "FREQ=DAILY"
	if _, err := tasks.CreateNode(ctx, service.NewNode{
		Type:       domain.NodeTypeHabit,
		Title:      "Ежедневная привычка",
		Status:     statuses[0],
		Priority:   domain.Priority3,
		Recurrence: &daily,
	}); err != nil {
		return fmt.Errorf("creating the habit: %w", err)
	}
	return nil
}
