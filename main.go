package main

import (
	"embed"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"nexus/internal/platform"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	socket := platform.SocketPath()

	// Create an instance of the app structure
	app := NewApp()

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

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "nexus",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
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
