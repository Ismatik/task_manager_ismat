package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"nexus/internal/platform"
)

// App is the struct bound to Wails. Its exported methods are the frontend's
// entire surface onto Go, and every one of them returns (T, error) — see
// ARCHITECTURE.md §4: Wails turns the second return value into a rejected JS
// promise, and without it the frontend cannot tell "empty result" from "it blew
// up".
type App struct {
	// mu guards ctx, which is written by startup on the main goroutine and read
	// by onIPCMessage on the single-instance listener's goroutine.
	mu  sync.RWMutex
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

// Greet returns a greeting for the given name.
//
// It cannot currently fail, and still returns an error: adding a failure mode
// later must not be a breaking change at every call site (ARCHITECTURE.md §4).
func (a *App) Greet(name string) (string, error) {
	return fmt.Sprintf("Hello %s, It's show time!", name), nil
}

// onIPCMessage handles one line from a secondary instance. It is unexported, so
// Wails does not bind it; it is wired to the single-instance lock in main.go.
func (a *App) onIPCMessage(msg platform.Message) {
	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()

	// A message can arrive between Acquire and Wails' startup callback. There
	// is no window to focus yet, and the one being built will come up focused.
	if ctx == nil {
		return
	}

	if msg == platform.MsgQuick {
		// Stage 0 goes no further than this on purpose: the quick-add window is
		// Stage 4. Logging it proves the flag travelled end to end.
		runtime.LogInfo(ctx, "ipc: --quick requested; quick-add arrives in Stage 4, focusing the main window")
	}

	// Unminimise before Show: a window that is merely iconified is still
	// "shown", so Show alone would be a no-op for it. Raising above other
	// applications is up to the window manager and is best-effort on Wayland.
	runtime.WindowUnminimise(ctx)
	runtime.WindowShow(ctx)
}
