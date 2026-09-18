// Package store is the persistence layer: the SQLite connection, the embedded
// schema migrations, the repositories that read and write nodes, tags, time
// entries and settings, and the search index.
//
// What is here today: Open, which connects to the SQLite file and verifies its
// pragmas; Migrate, which applies the embedded schema migrations; and
// SettingsRepo, the key/value store behind the palette, theme, accent and
// language preferences.
//
// Still to come in this stage: the node, tag, time-entry, attachment and
// habit-check repositories, and the search backend over titles and
// descriptions.
//
// # Imports
//
// store sits between service and domain: it may import nexus/internal/domain
// and must never import nexus/internal/service. Repositories expose data, they
// do not orchestrate — a repository reaching back up into a service is a cycle
// waiting to happen.
//
// # No cgo
//
// The driver is modernc.org/sqlite, which is pure Go.
// github.com/mattn/go-sqlite3, or any other cgo driver, must never enter the
// dependency graph: CGO_ENABLED=0 go build ./... has to keep succeeding. If a
// feature turns out to be unavailable in the pure-Go driver the feature
// degrades; the no-cgo guarantee does not.
package store
