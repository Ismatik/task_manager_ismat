package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"nexus/internal/platform"
	"nexus/internal/service"
	"nexus/internal/store"
)

//go:embed all:frontend/dist
var assets embed.FS

// openStore opens the database at the default path, applies every migration and
// seeds the settings defaults.
//
// The caller owns the handle and must Close it. Everything it can fail at is
// fatal to the caller (see main): opening a window onto a database that is not
// there is worse than not opening one, because a silent empty board looks
// exactly like "you have no tasks".
func openStore(ctx context.Context) (*sql.DB, error) {
	path, err := store.DefaultPath()
	if err != nil {
		return nil, err
	}

	db, err := store.Open(ctx, path)
	if err != nil {
		return nil, err
	}

	if _, err := store.Migrate(ctx, db); err != nil {
		db.Close() //nolint:errcheck // the migration error is the one that matters
		return nil, fmt.Errorf("nexus: migrating %s: %w", path, err)
	}
	if _, err := store.NewSettingsRepo(db).SeedDefaults(ctx); err != nil {
		db.Close() //nolint:errcheck // as above
		return nil, fmt.Errorf("nexus: seeding settings in %s: %w", path, err)
	}
	return db, nil
}

// newServices constructs every service over db, with the real wall clock and a
// uuid generator.
//
// The clock and the id source are injected here and nowhere else: internal/
// never reads either from the ambient environment, which is what makes every
// rule in it testable at an exact instant (ARCHITECTURE.md §2).
func newServices(db *sql.DB) Services {
	var (
		clock   = service.Clock(service.SystemClock)
		newID   = func() string { return uuid.NewString() }
		begin   = service.NewBeginner(db)
		nodes   = store.NewNodeRepo(db)
		tags    = store.NewTagRepo(db)
		entries = store.NewTimeEntryRepo(db)
		checks  = store.NewHabitCheckRepo(db)
		search  = store.NewSearchRepo(db)
	)

	// The timer service is built first: the task service drives it inside the
	// move's own transaction, which is what couples Doing to a time entry
	// (C1, D13).
	timers := service.NewTimerService(begin, nodes, entries, clock, newID)

	return Services{
		Tasks:    service.NewTaskService(begin, nodes, tags, timers, clock, newID),
		Timers:   timers,
		Habits:   service.NewHabitService(nodes, checks, clock),
		Search:   service.NewSearchService(begin, nodes, tags, search, clock),
		Settings: service.NewSettingsService(store.NewSettingsRepo(db)),
	}
}

func main() {
	socket := platform.SocketPath()

	// The database is the application: if it cannot be opened or migrated,
	// there is nothing to show and saying so is the only honest outcome.
	db, err := openStore(context.Background())
	if err != nil {
		log.Printf("nexus: %v", err)
		os.Exit(1)
	}
	defer db.Close() //nolint:errcheck // nothing useful to do at exit

	// Create an instance of the app structure
	app := NewApp(newServices(db))

	// Nexus is single-instance: whoever manages to listen on the socket owns
	// the window, and every later launch is a message to it.
	instance, err := platform.Acquire(socket, app.onIPCMessage)
	switch {
	case errors.Is(err, platform.ErrAlreadyRunning):
		// Secondary instance: hand the request over and get out of the way.
		// --quick sends "quick", a bare launch sends "focus".
		if sendErr := platform.Send(socket, platform.MessageFromArgs(os.Args[1:])); sendErr != nil {
			log.Printf("nexus: %v", sendErr)
			os.Exit(1)
		}
		os.Exit(0)
	case err != nil:
		log.Printf("nexus: single-instance lock: %v", err)
		os.Exit(1)
	}
	defer instance.Close() //nolint:errcheck // nothing useful to do at exit

	stopSignals := releaseSocketOnSignal(instance)
	defer stopSignals()

	// K1, closed by D12. Wails formats the window background in C with the
	// PROCESS locale, so under a comma-decimal locale GTK silently discards the
	// colour; this has to happen before wails.Run, because GTK reads the
	// environment once, when it initialises. Only LC_NUMERIC is forced — see
	// forceNumericLocale for why not LC_ALL.
	if err := forceNumericLocale(); err != nil {
		// Not fatal: the window still opens, it is only the first frame's
		// colour that is at risk.
		log.Printf("nexus: %v", err)
	}

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "nexus",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// The background GTK paints before the WebView has evaluated anything,
		// so it is the first colour the user sees: it is read from the palette
		// and theme the user last chose (D12) against design/tokens.css, which
		// owns every colour in this application. No colour is named here, or
		// anywhere else in Go — see background.go.
		BackgroundColour: windowBackground(startupAppearance(context.Background(), app.svc.Settings)),
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Printf("nexus: %v", err)
	}
}

// releaseSocketOnSignal removes the IPC socket when the process is asked to
// terminate. wails.Run does not return on SIGINT/SIGTERM, so without this a
// Ctrl-C would leave the socket on disk for the next launch to reclaim as
// stale. The signal is re-raised with its default disposition afterwards, so
// the exit status stays conventional.
//
// The returned function stops the handler; it is called on the normal exit
// path, where the deferred Close already does the cleanup.
func releaseSocketOnSignal(instance *platform.Instance) func() {
	received := make(chan os.Signal, 1)
	signal.Notify(received, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig, ok := <-received
		if !ok {
			return
		}
		if err := instance.Close(); err != nil {
			log.Printf("nexus: %v", err)
		}
		signal.Stop(received)
		if self, err := os.FindProcess(os.Getpid()); err == nil {
			_ = self.Signal(sig)
		}
	}()

	return func() {
		signal.Stop(received)
		close(received)
	}
}
