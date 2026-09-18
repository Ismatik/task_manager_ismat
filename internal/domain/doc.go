// Package domain holds the pure core of Nexus: the node tree, time entries and
// every rule derived from them (status derivation, the column<->due coupling,
// progress and habit streaks).
//
// The package holds the enumerations and their validators, the Node, Tag,
// TimeEntry and HabitCheck types, and the Date type used by the columns that are
// dates rather than instants. The rules that operate on them — status
// derivation, progress, the column<->due coupling, the tree and cascade plans
// and habit streaks — live in this same package, alongside their tests.
//
// # Purity
//
// This package performs no I/O of any kind and must stay that way:
//
//   - no database access — it does not import database/sql or any driver,
//   - no filesystem access — it does not import os,
//   - no network — it does not import anything under net,
//   - no clock reads — time.Now is never called here. Anything that needs the
//     current time takes an injected now func() time.Time, which keeps every
//     time-dependent rule deterministically testable.
//
// # Imports
//
// domain is the innermost layer and knows about nobody: it imports nothing from
// nexus/internal/store, nexus/internal/service, nexus/internal/parse or
// nexus/internal/platform. The dependency direction is
// main.go -> service -> store -> domain, strictly one-way.
//
// The constraint is enforced mechanically by TestDomainIsPure in this package's
// external test package; the equivalent manual check is
//
//	go list -f '{{join .Imports "\n"}}' ./internal/domain
package domain
