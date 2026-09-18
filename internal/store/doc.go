// Package store is the persistence layer: the SQLite connection, the embedded
// schema migrations, the repositories that read and write nodes, tags, time
// entries and settings, and the search index.
//
// Stage 0 delivers only a placeholder; the connection and the migration runner
// arrive in S0-08 and the settings repository in S0-09.
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
