# Nexus — Architecture

The authoritative layout of the repository. Decided in [`PLAN.md` §7, D3](./PLAN.md),
written down **before** the code that creates it, so every later ticket has a target to
build against rather than a shape to invent.

---

## 1. Layout

```
main.go              — Wails bootstrap, single-instance lock, --quick flag
internal/domain      — Node, TimeEntry, status derivation, column<->due rules,
                       streaks (PURE, no I/O)
internal/store       — SQLite repos, embedded migrations, FTS
internal/service     — TaskService, TimerService, BackupService, ExportService
internal/parse       — quick-add NL parser
internal/platform    — tray, autostart .desktop, D-Bus sleep/lock
frontend/src/{views,components,store,lib,locales}
build/linux/         — .desktop, autostart, install.sh
```

### `main.go` lives at the repo root

`main.go` (and its companion `app.go`) sit at the **repository root**, which is the
Wails v2 convention — the `wails.json` at the root, the `main.go` beside it and
`frontend/` as a sibling. They are **not** under `cmd/nexus/`. Wails' code generation
(`frontend/wailsjs/`) and its build pipeline both assume the root-level entry point;
moving it would mean fighting the toolchain for no gain.

`main.go` is deliberately thin: it parses flags, acquires the single-instance lock,
opens the store, constructs the services and hands them to Wails as bound structs. No
business rule ever lives there.

---

## 2. Where logic lives

### All domain logic lives in Go

The frontend **renders what Go returns and computes nothing**. Specifically, the
frontend never computes:

- a node's **status** (derived from its children — see D2),
- a **streak** (derived from scheduled RRULE occurrences — see D5),
- **progress** (completed / total, with leaves whose type has no Kanban column —
  `note`, `habit` — excluded from the denominator; see D7, D10, D11),
- an **overdue** flag,
- a **due date** (including the column↔due coupling — see D1).

Every one of those arrives from Go already computed, as a plain field on the DTO the
frontend receives. If the UI needs a new derived value, the derivation is added in
`internal/domain` and surfaced through `internal/service` — never reimplemented in
TypeScript. Two implementations of the same rule is two rules.

### `internal/domain` is pure

`internal/domain` performs **no I/O of any kind**:

- no database access — it does not import `database/sql` or any driver,
- no filesystem access — it does not import `os`,
- no network — it does not import anything under `net`,
- **no clock reads**. `time.Now()` is never called inside `domain`. Anything that
  needs the current time takes an injected `now func() time.Time`, which makes every
  time-dependent rule (overdue, streaks, "today", "this week") deterministically
  testable.

The package is therefore testable with nothing but `go test` — no temp dirs, no
fixtures, no fake clocks beyond a closure.

---

## 3. Dependency direction

```
main.go  →  internal/service  →  internal/store  →  internal/domain
```

Strictly one-way. The rules that follow from it:

- **`domain` imports nothing** from `store`, `service`, `parse` or `platform`. It is
  the innermost layer and knows about nobody.
- **`store` never imports `service`.** Repositories expose data; they do not
  orchestrate. A repository that reaches back up into a service is a cycle waiting to
  happen.
- **`service` is the only layer `main.go` talks to** for behaviour. It composes
  repositories from `store`, applies rules from `domain`, and is what gets bound to
  Wails.
- `internal/parse` (quick-add NL parsing) may depend on `domain` types; it must not
  depend on `store` or `service`.
- `internal/platform` (tray, autostart, D-Bus, single-instance) is a leaf concern
  depended on by `main.go`; it does not import `service` or `store`.

---

## 4. The Wails binding contract

**Every Wails-bound method returns `(T, error)`.**

That is not stylistic. Wails marshals a second return value into a rejected JS
promise, so a method that returns only `T` gives the frontend no way to distinguish
"empty result" from "it blew up". Even methods that cannot currently fail return an
`error`, so that adding a failure mode later is not a breaking change to every call
site.

Bound methods live on the `App` struct in `app.go` and are thin delegations to
`internal/service`.

---

## 5. Build constraints

### No cgo — ever

The SQLite driver is **`modernc.org/sqlite`**, which is pure Go. `github.com/mattn/go-sqlite3`
(or any other cgo driver) must never enter the dependency graph.

```sh
CGO_ENABLED=0 go build ./...   # must always succeed
go list -deps ./... | grep -i mattn   # must always return nothing
```

This keeps the build reproducible, cross-compilable and free of a C toolchain
requirement. If a feature (e.g. FTS5) turns out to be unavailable in the pure-Go
driver, **the feature degrades** (FTS5 → `LIKE`-based search) — the no-cgo guarantee
does not.

### `-tags webkit2_41` on every wails invocation (E1)

**`libwebkit2gtk-4.0-dev` does not exist on Ubuntu 24.04.** The archive ships only
`libwebkit2gtk-4.1-dev`. Wails v2 defaults to probing for the 4.0 pkg-config module
and only looks for 4.1 when the `webkit2_41` build tag is set. Therefore:

> **Every `wails build` and `wails dev` invocation in this project must pass
> `-tags webkit2_41`.**

```sh
wails build -tags webkit2_41
wails dev   -tags webkit2_41
```

The `Makefile` bakes the tag in (`TAGS ?= webkit2_41`) so that nobody has to remember
it. A wails command run *without* the tag failing to link is the **expected** outcome
on this machine, not a bug to debug.

System prerequisites (installed by the user, `sudo` required):

```sh
sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

### The wails CLI is resolved explicitly (E2)

The CLI (v2.16.0) lives at `$(go env GOPATH)/bin/wails` and is not guaranteed to be on
`PATH` in non-login shells. Build tooling resolves it as
`WAILS ?= $(shell go env GOPATH)/bin/wails` rather than assuming a bare `wails` works.

---

## 6. Frontend layout

```
frontend/src/views       — one file per top-level view (Kanban, Tree, Calendar, ...)
frontend/src/components  — reusable presentational components
frontend/src/store       — client-side UI state only (selection, open panels, filters)
frontend/src/lib         — formatting, Wails call wrappers, small helpers
frontend/src/locales     — en.json / ru.json; every user-visible string is a key
frontend/wailsjs/        — GENERATED by wails; committed so typecheck can run
                           without a full wails build. Never hand-edited.
```

`frontend/src/store` holds **UI** state. Domain state is owned by Go and fetched; it is
not mirrored into a client store and mutated locally.

Colours, radii, shadows and fonts resolve **only** through the Tailwind token names
defined in `design/tailwind.config.js` / `design/tokens.css`. There are no hex
literals in components, and **`design/` is read-only**.

### Priority → chip mapping (authoritative)

`design/README.md` talks about `P0/P1/P2` chips while `domain.Priority` is `1..4`.
The mapping between the two is fixed here so that Stage 2 does not re-decide it:

| `domain.Priority` | Chip | Token |
|---|---|---|
| 1 | `P0` | `danger` |
| 2 | `P1` | `danger` |
| 3 | `P2` | `warning` |
| 4 | *(no chip)* | — |

This is a **rendering** mapping only. The domain stores and reasons about `1..4`; Go
remains the only place that decides anything, and the chip label is presentation. It
lines up with the existing design semantics (`design/README.md` "Semantics", `PLAN.md`
§3): `danger` is used for overdue and **P0/P1**, `warning` for **P2**. Priority 4 is
the unremarkable default and renders **no chip at all** — not a grey "P3".

---

## 7. Data and runtime paths

| What | Where |
|---|---|
| SQLite database | `$XDG_DATA_HOME/nexus/nexus.db` (fallback `~/.local/share/nexus/nexus.db`) |
| Single-instance socket | `$XDG_RUNTIME_DIR/nexus/ipc.sock` (fallback `/tmp/nexus-$UID/`) |
| Migrations | embedded in the binary via `//go:embed`, never read from disk |

Nexus is **local-only**: no network calls, no telemetry, no cloud sync.

---

## 8. See also

- [`PLAN.md`](./PLAN.md) — the brief, the data model and decisions D1–D11 / E1–E3.
- [`TASKS.md`](./TASKS.md) — the ticket breakdown.
- [`design/README.md`](./design/README.md) — token and palette handoff (read-only).
