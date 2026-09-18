package platform

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The single-instance lock is a unix domain socket, not D-Bus. D-Bus would drag
// in a session-bus dependency and a well-known service name for what is a
// one-line message; the D-Bus work this package owes (sleep and screen lock,
// Stage 4) stays separate from it deliberately.
//
// The socket doubles as the lock and as the channel: whoever manages to Listen
// is the primary instance, and everybody else connects, writes one line and
// exits.

// SocketName is the file name of the single-instance socket inside the
// runtime directory.
const SocketName = "ipc.sock"

// Message is the one line a secondary instance writes to the primary. It is a
// whole protocol: one line, no arguments, no reply.
type Message string

const (
	// MsgFocus asks the primary instance to show, unminimise and raise its
	// main window. It is what a bare second launch sends.
	MsgFocus Message = "focus"

	// MsgQuick asks the primary instance for the quick-add entry point, and is
	// what `nexus --quick` sends. In Stage 0 the primary logs it and focuses
	// the main window; the quick-add window itself is Stage 4.
	MsgQuick Message = "quick"
)

// ErrAlreadyRunning is returned by Acquire when another Nexus instance holds
// the lock. The caller is then a secondary instance: it should Send its message
// and exit 0.
var ErrAlreadyRunning = errors.New("platform: another nexus instance is already running")

const (
	// dialTimeout bounds both the liveness probe in Acquire and Send. A local
	// unix socket either answers at once or is not there at all.
	dialTimeout = 2 * time.Second
	// ioTimeout bounds reading the one line from a peer, so a connection that
	// opens and then says nothing cannot pin a goroutine forever.
	ioTimeout = 5 * time.Second
)

// RuntimeDir returns the directory holding the single-instance socket:
// $XDG_RUNTIME_DIR/nexus, falling back to /tmp/nexus-$UID when the variable is
// unset (a bare `su`, a cron job, a container without a logind session).
func RuntimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return filepath.Join(dir, AppName)
	}
	return filepath.Join(os.TempDir(), AppName+"-"+strconv.Itoa(os.Getuid()))
}

// SocketPath returns the full path of the single-instance socket.
func SocketPath() string {
	return filepath.Join(RuntimeDir(), SocketName)
}

// MessageFromArgs maps process arguments onto the message a secondary instance
// sends: --quick (or -quick) means MsgQuick, anything else means MsgFocus.
//
// It deliberately does not use the flag package. flag.Parse terminates the
// process on any argument it does not recognise, and this binary is handed
// arguments it does not own — `wails dev` forwards its own, and GTK/WebKit add
// theirs. Unknown arguments are ignored here and os.Args is left untouched so
// that Wails' argument handling still sees exactly what it was given.
func MessageFromArgs(args []string) Message {
	for _, arg := range args {
		if arg == "--quick" || arg == "-quick" {
			return MsgQuick
		}
	}
	return MsgFocus
}

// Instance is a held single-instance lock: a listening socket plus the
// goroutine accepting messages from secondary instances. Close releases it.
type Instance struct {
	listener net.Listener
	path     string
	handle   func(Message)

	closeOnce sync.Once
	closing   chan struct{}
	wg        sync.WaitGroup
}

// Acquire tries to become the primary instance by listening on path, creating
// the parent directory if needed. On success it returns an Instance and starts
// serving; handle is then called, on its own goroutine, once per message
// received. handle may be nil.
//
// If the socket exists and something is accepting on it, Acquire returns
// ErrAlreadyRunning and the caller is a secondary instance. If the socket
// exists but nothing answers — a corpse left by a crash — it is unlinked and
// the listen retried exactly once before giving up with a descriptive error.
func Acquire(path string, handle func(Message)) (*Instance, error) {
	dir := filepath.Dir(path)
	// 0700: the socket is per-user and nobody else has any business writing to it.
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("platform: create runtime dir %s: %w", dir, err)
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		if isAccepting(path) {
			return nil, ErrAlreadyRunning
		}
		// Nothing is listening, so whatever is on disk is stale. Unlink and
		// retry once; a second failure is real and gets reported as such.
		if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			return nil, fmt.Errorf("platform: remove stale socket %s: %w", path, rmErr)
		}
		listener, err = net.Listen("unix", path)
		if err != nil {
			return nil, fmt.Errorf("platform: listen on %s after reclaiming a stale socket: %w", path, err)
		}
	}

	inst := &Instance{
		listener: listener,
		path:     path,
		handle:   handle,
		closing:  make(chan struct{}),
	}
	inst.wg.Add(1)
	go inst.serve()
	return inst, nil
}

// Path is the socket file this instance holds.
func (i *Instance) Path() string { return i.path }

// Close stops accepting, waits for in-flight messages and removes the socket
// file. It is safe to call more than once; only the first call does anything.
func (i *Instance) Close() error {
	var err error
	i.closeOnce.Do(func() {
		close(i.closing)
		err = i.listener.Close()
		i.wg.Wait()
		// Go's unix listener unlinks the socket it created, so this is normally
		// a no-op; it is kept for the case where it did not.
		if rmErr := os.Remove(i.path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) && err == nil {
			err = fmt.Errorf("platform: remove socket %s: %w", i.path, rmErr)
		}
	})
	return err
}

func (i *Instance) serve() {
	defer i.wg.Done()
	for {
		conn, err := i.listener.Accept()
		if err != nil {
			// Close is the only expected way out; any other error means the
			// listener is unusable, and spinning on it would burn a core.
			return
		}
		i.wg.Add(1)
		go func() {
			defer i.wg.Done()
			i.receive(conn)
		}()
	}
}

// receive reads exactly one line and dispatches it. Anything after the first
// line is ignored: the protocol is one message per connection.
func (i *Instance) receive(conn net.Conn) {
	defer conn.Close() //nolint:errcheck // read-only side of a local socket

	_ = conn.SetReadDeadline(time.Now().Add(ioTimeout))
	line, err := bufio.NewReader(conn).ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if i.handle != nil {
		i.handle(Message(line))
	}
}

// Send delivers exactly one message to the primary instance listening at path.
// It is what a secondary instance calls immediately before exiting 0.
func Send(path string, msg Message) error {
	conn, err := net.DialTimeout("unix", path, dialTimeout)
	if err != nil {
		return fmt.Errorf("platform: connect to the running instance at %s: %w", path, err)
	}
	defer conn.Close() //nolint:errcheck // the write below is what can fail

	_ = conn.SetWriteDeadline(time.Now().Add(ioTimeout))
	if _, err := io.WriteString(conn, string(msg)+"\n"); err != nil {
		return fmt.Errorf("platform: send %q to the running instance: %w", msg, err)
	}
	return nil
}

// isAccepting reports whether something is actually accepting on path. A socket
// file that nobody answers is stale and may be reclaimed.
func isAccepting(path string) bool {
	conn, err := net.DialTimeout("unix", path, dialTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
