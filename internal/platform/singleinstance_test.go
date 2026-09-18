package platform_test

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"nexus/internal/platform"
)

// recorder collects the messages the primary instance receives. The handler
// runs on the serving goroutine, so a buffered channel is the synchronisation
// point rather than a slice.
type recorder chan platform.Message

func newRecorder() recorder { return make(recorder, 8) }

func (r recorder) handle(msg platform.Message) { r <- msg }

func (r recorder) await(t *testing.T) platform.Message {
	t.Helper()
	select {
	case msg := <-r:
		return msg
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the primary instance to receive a message")
		return ""
	}
}

// socketIn returns a socket path inside a temp directory. Every test that
// touches the filesystem goes through it, so nothing is left behind.
func socketIn(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), platform.SocketName)
}

func TestSocketPathUsesXDGRuntimeDir(t *testing.T) {
	runtimeDir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", runtimeDir)

	want := filepath.Join(runtimeDir, "nexus", platform.SocketName)
	if got := platform.SocketPath(); got != want {
		t.Errorf("SocketPath() = %q, want %q", got, want)
	}
}

func TestSocketPathFallsBackWhenRuntimeDirUnset(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")

	want := filepath.Join(os.TempDir(), "nexus-"+strconv.Itoa(os.Getuid()), platform.SocketName)
	if got := platform.SocketPath(); got != want {
		t.Errorf("SocketPath() with XDG_RUNTIME_DIR unset = %q, want the /tmp/nexus-$UID fallback %q", got, want)
	}
}

func TestAcquireCreatesTheRuntimeDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nexus", platform.SocketName)

	inst, err := platform.Acquire(path, nil)
	if err != nil {
		t.Fatalf("Acquire(%q) = %v, want it to create the parent directory", path, err)
	}
	defer inst.Close() //nolint:errcheck // asserted in its own test

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("socket %q not created: %v", path, err)
	}
}

// The headline behaviour: the second launch does not open a window, it talks to
// the first one.
func TestSecondAcquireReportsAlreadyRunningAndItsMessageArrives(t *testing.T) {
	path := socketIn(t)
	got := newRecorder()

	primary, err := platform.Acquire(path, got.handle)
	if err != nil {
		t.Fatalf("first Acquire = %v, want success", err)
	}
	defer primary.Close() //nolint:errcheck // asserted in its own test

	secondary, err := platform.Acquire(path, nil)
	if !errors.Is(err, platform.ErrAlreadyRunning) {
		t.Fatalf("second Acquire = %v, want ErrAlreadyRunning", err)
	}
	if secondary != nil {
		t.Fatal("second Acquire returned an Instance; a secondary must not hold the lock")
	}

	if err := platform.Send(path, platform.MsgFocus); err != nil {
		t.Fatalf("Send from the secondary = %v, want success", err)
	}
	if msg := got.await(t); msg != platform.MsgFocus {
		t.Errorf("primary received %q, want %q", msg, platform.MsgFocus)
	}
}

// A crash leaves the socket file on disk with nobody behind it. The next launch
// must reclaim it rather than refuse to start forever.
func TestAcquireReclaimsAStaleSocket(t *testing.T) {
	path := socketIn(t)

	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		t.Fatalf("ResolveUnixAddr(%q) = %v", path, err)
	}
	corpse, err := net.ListenUnix("unix", addr)
	if err != nil {
		t.Fatalf("ListenUnix(%q) = %v", path, err)
	}
	// Leave the file behind on close — exactly what a SIGKILL would do.
	corpse.SetUnlinkOnClose(false)
	if err := corpse.Close(); err != nil {
		t.Fatalf("closing the stale listener = %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stale socket not left on disk: %v", err)
	}

	got := newRecorder()
	inst, err := platform.Acquire(path, got.handle)
	if err != nil {
		t.Fatalf("Acquire over a stale socket = %v, want it reclaimed, not fatal", err)
	}
	defer inst.Close() //nolint:errcheck // asserted in its own test

	if err := platform.Send(path, platform.MsgFocus); err != nil {
		t.Fatalf("Send after reclaiming = %v, want the reclaimed socket to work", err)
	}
	if msg := got.await(t); msg != platform.MsgFocus {
		t.Errorf("received %q, want %q", msg, platform.MsgFocus)
	}
}

// --quick and a bare launch must be told apart on the wire; Stage 4 hangs the
// real quick-add window off exactly this distinction.
func TestQuickAndBareLaunchAreDistinguishableOnTheWire(t *testing.T) {
	path := socketIn(t)
	got := newRecorder()

	primary, err := platform.Acquire(path, got.handle)
	if err != nil {
		t.Fatalf("Acquire = %v, want success", err)
	}
	defer primary.Close() //nolint:errcheck // asserted in its own test

	for _, want := range []platform.Message{
		platform.MessageFromArgs([]string{"--quick"}),
		platform.MessageFromArgs(nil),
	} {
		if err := platform.Send(path, want); err != nil {
			t.Fatalf("Send(%q) = %v, want success", want, err)
		}
		if msg := got.await(t); msg != want {
			t.Errorf("primary received %q, want %q", msg, want)
		}
	}

	if platform.MsgQuick == platform.MsgFocus {
		t.Fatal("MsgQuick and MsgFocus are the same string; the two launches are indistinguishable")
	}
}

func TestMessageFromArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want platform.Message
	}{
		{"no arguments", nil, platform.MsgFocus},
		{"empty slice", []string{}, platform.MsgFocus},
		{"double dash quick", []string{"--quick"}, platform.MsgQuick},
		{"single dash quick", []string{"-quick"}, platform.MsgQuick},
		{"quick among others", []string{"-loglevel", "debug", "--quick"}, platform.MsgQuick},
		{"unknown arguments are ignored, not fatal", []string{"--assetdir", "/tmp/x"}, platform.MsgFocus},
		{"a lookalike is not the flag", []string{"--quickly"}, platform.MsgFocus},
		{"positional argument", []string{"quick"}, platform.MsgFocus},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := platform.MessageFromArgs(tc.args); got != tc.want {
				t.Errorf("MessageFromArgs(%q) = %q, want %q", tc.args, got, tc.want)
			}
		})
	}
}

func TestCloseRemovesTheSocketAndReleasesTheLock(t *testing.T) {
	path := socketIn(t)

	primary, err := platform.Acquire(path, nil)
	if err != nil {
		t.Fatalf("Acquire = %v, want success", err)
	}
	if err := primary.Close(); err != nil {
		t.Fatalf("Close = %v, want nil", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("os.Stat(%q) after Close = %v, want the socket to be gone", path, err)
	}

	// Closing twice must not panic or report a spurious failure.
	if err := primary.Close(); err != nil {
		t.Errorf("second Close = %v, want nil", err)
	}

	// The lock is genuinely released: a fresh Acquire becomes primary again.
	next, err := platform.Acquire(path, nil)
	if err != nil {
		t.Fatalf("Acquire after Close = %v, want the lock to have been released", err)
	}
	defer next.Close() //nolint:errcheck // already asserted above
}

func TestSendWithoutAPrimaryFails(t *testing.T) {
	path := socketIn(t)

	if err := platform.Send(path, platform.MsgFocus); err == nil {
		t.Fatal("Send with nobody listening = nil, want an error naming the socket")
	}
}

func TestAcquireWithNilHandlerIgnoresMessages(t *testing.T) {
	path := socketIn(t)

	primary, err := platform.Acquire(path, nil)
	if err != nil {
		t.Fatalf("Acquire = %v, want success", err)
	}
	defer primary.Close() //nolint:errcheck // asserted in its own test

	if err := platform.Send(path, platform.MsgQuick); err != nil {
		t.Fatalf("Send = %v, want success", err)
	}
	// Nothing to observe beyond "this did not panic"; give the serving
	// goroutine a moment to run the message through.
	time.Sleep(50 * time.Millisecond)
}
