package platform

import "errors"

// AppName is the directory and identifier Nexus uses under the XDG base
// directories: $XDG_DATA_HOME/nexus, $XDG_RUNTIME_DIR/nexus and the
// nexus.desktop entry.
const AppName = "nexus"

// ErrUnsupported is returned when a desktop facility is not available on the
// running system — no AppIndicator for the tray, no session bus for the sleep
// and lock listeners. It is never fatal: the caller degrades and warns.
var ErrUnsupported = errors.New("platform: not supported on this system")
