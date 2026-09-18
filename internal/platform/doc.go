// Package platform holds the Linux desktop integrations: the single-instance
// lock and its --quick IPC message, the system tray, the autostart .desktop
// entry and installer, and the D-Bus listeners for sleep and screen lock.
//
// Stage 0 delivers only a placeholder; the single-instance lock arrives in
// S0-10 and the tray, autostart and D-Bus work in Stages 4 and 5.
//
// # Imports
//
// platform is a leaf concern depended on by main.go. It does not import
// nexus/internal/service or nexus/internal/store: it reports events and exposes
// operating-system facilities, and main.go wires those to the services.
package platform
