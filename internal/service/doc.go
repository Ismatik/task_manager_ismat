// Package service orchestrates the application: TaskService, TimerService,
// BackupService and ExportService. It composes repositories from
// nexus/internal/store, applies the rules from nexus/internal/domain, and is
// the layer bound to Wails.
//
// Stage 0 delivers only a placeholder; the services arrive in Stage 1.
//
// # Imports
//
// service is the only layer main.go talks to for behaviour. It may import
// nexus/internal/store and nexus/internal/domain; nothing imports service in
// return — the dependency direction is main.go -> service -> store -> domain,
// strictly one-way.
//
// # Contract
//
// Every method that ends up bound to Wails returns (T, error), including ones
// that cannot currently fail, so that adding a failure mode later is not a
// breaking change at every call site.
//
// # Time
//
// Reading the wall clock is a service-layer concern. Services hold a Clock and
// pass it down as the injected now func() time.Time that domain rules require,
// which keeps every time-dependent rule deterministically testable.
package service
