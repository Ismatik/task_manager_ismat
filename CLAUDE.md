# CLAUDE.md — working rules for Nexus

Read this first. It is the short list of things that are easy to get wrong here.
It does not restate the plan: [`PLAN.md`](./PLAN.md) is the brief and the decisions
(D1–D15, E1–E3), [`ARCHITECTURE.md`](./ARCHITECTURE.md) is the layout,
[`TASKS.md`](./TASKS.md) is the ticket breakdown, and
[`design/README.md`](./design/README.md) is the token handoff.

---

## RESUME HERE — temporary handoff, delete when Stage 2 starts

> This section is scaffolding for picking the work back up on another machine.
> **Delete it once Stage 2 is underway.** Everything below it is permanent.

### Where the work stopped

| | |
|---|---|
| Stage 0 | **CLOSED** — Reviewer PASS |
| Stage 1 | **CLOSED** — Reviewer PASS on the fourth round, at `a1f09b7` |
| Stage 2 | **PLANNED** — twenty-two tickets, `S2-01` … `S2-22` in `TASKS.md`. **No code written yet.** |

Coverage: `internal/domain` **100.0%**, `internal/service` **92.9%** (bar is ≥90%),
`internal/store` 86.4% (not gated). All five `make check` gates green, including
`wails build -tags webkit2_41`.

### Do this first

**Implement `S2-01`** — one ticket at a time, one commit each. Stage 2 is planned: the
tickets are in `TASKS.md`, and **C1–C5 are absorbed** into named ones (C1 → S2-03,
C2 → S2-01, C3 → S2-08, C4 → S2-06, C5 → S2-14) rather than left floating.

The Go engine is complete and tested but **nothing is wired to a UI yet**. `main.go`
still does not open the store or construct a service, so Stage 2's first job is that
wiring. `frontend/src` is still the untouched Wails scaffold (an `App.tsx` demo with a
`Greet` box) and there is no `locales/` directory yet, even though i18n is required from
the first component.

If `frontend/wailsjs/*` ever shows as modified with files truncated to zero bytes, an
interrupted `wails build` did it. `make build` regenerates them.

### Stage 2 must not forget these

- **Wire `MoveToColumn(doing)` to `TimerService.Start`** — `PLAN.md` §4 couples them and
  no Stage 1 ticket did it. **Now S2-03**, with **D13** settling the three things §4
  leaves open (same transaction; a cascade opens no timer; leaving `doing` closes it).
- **`Board()` is O(n²·log n)** — `snapshot.view` rebuilds its index per node. **Now
  S2-01**, scheduled first, because the instruction was *fix it before the board is on
  screen*.
- **No type rule may be spelled twice.** Three of Stage 1's four review failures were
  the same defect: a rule written down in two places and edited in one. The inventory
  is currently clean — `HasColumn`, `HasDue`, `DoingRefusal`/`CanBeDoing`,
  `countsAsWork`, habit-requires-recurrence and `DefaultActivity` are each the single
  definition of their rule. **This applies to TypeScript too**: the frontend renders
  what Go returns and re-implements none of it.
- **Screenshots and README.** `README.md` is deliberately not written yet — every
  screen it needs belongs to Stage 2+. Write it once the board exists. Automated
  capture will need `sudo apt install xvfb imagemagick` (none of `xvfb-run`, `scrot`,
  `import`, `grim` are installed).

### Known issues — all now DECIDED and scheduled (rulings in `PLAN.md` §7)

- **K1** — `BackgroundColour` never reaches GTK on this machine. `LC_NUMERIC=ru_RU.UTF-8`
  makes Wails' `window.c:205` emit `rgba(27, 38, 54, 0,0)` with a comma, so GTK fails to
  parse it. **DECIDED — D12** (the user's): force `LC_NUMERIC=C` **and** seed the
  background from the persisted palette/theme. **Ticket S2-08.** The containment rule
  matters: `main.go` must gain **no hex literal**.
- **K2** — since D11 a leaf project's stale *stored* status has teeth: archiving the last
  real child of a project an earlier cascade wrote `done` leaves it counted as a done unit.
  **DECIDED — D14** (PM ruling): archiving rewrites the stored status of a node it turned
  into a leaf, to the status that node derived immediately before. **Ticket S2-06.**
- **K3** — an empty project stored `done` renders in the Done column with no bar at all.
  **DECIDED — D15** (PM ruling): `progress.defined === false` renders a localised
  "empty project" marker instead of a bar, in every column. **Ticket S2-14.**

**D13, D14 and D15 are PM rulings, not the user's** — recorded so each has one spelling,
and overturnable by the user. **D1–D12 are the user's and are authoritative.**

### Environment, already done

- `pkg-config`, `libgtk-3-dev`, `libwebkit2gtk-4.1-dev` are installed. Ubuntu 24.04 has
  **no** `libwebkit2gtk-4.0-dev`; that is why every wails command needs `-tags webkit2_41`.
- Wails CLI v2.16.0 at `$(go env GOPATH)/bin/wails`, not necessarily on `PATH`.
- `origin` → `github.com/Ismatik/task_manager_ismat`. `main` was force-pushed to
  `a1f09b7` to drop a stray `7e50583 "Tests over other parts"` WIP commit that
  `9664506` already superseded. Nothing was lost.

### How the work is run

Orchestrator plus three sub-agents: **PM** owns `PLAN.md`/`TASKS.md` and decides when a
stage is done (never writes code); **Dev** implements exactly one ticket at a time and
commits per ticket; **Reviewer** reads the diff, runs the five gates, checks acceptance
criteria, and returns PASS or blocking issues. **A stage cannot close without PASS.**
One stage at a time, then stop and wait.

Two habits that earned their keep in Stage 1 and should continue: **negative-control
testing** (break the fix, confirm the specific test fails, restore) and **property
sweeps with an independent reference implementation** — `internal/domain/sweep_test.go`
is the committed example, and it caught bugs four rounds of example-based tests missed.

---

## What Nexus is

A personal task manager for the Linux desktop: a single Wails v2 binary (Go backend,
React + TypeScript frontend) over one SQLite file. Tasks, projects, habits, notes and
bugs are rows in one tree, with Kanban, a tree view, a calendar, time tracking and
streaks.

**Local-only, and that is a hard rule.** No server, no account, no sync, no telemetry,
no analytics, no crash reporting, no font or asset CDN, no update check. If a change
introduces a network call, it is wrong. Fonts, icons and every other asset are vendored
and loaded from disk.

---

## The five gate commands

`make check` is the single entry point. It runs exactly these, in this order, and stops
at the first failure:

| # | Command | Run from |
|---|---|---|
| 1 | `go test` over `./...` **excluding `node_modules`** | repo root |
| 2 | `go vet` over `./...` **excluding `node_modules`** | repo root |
| 3 | `npm run lint` | `frontend/` |
| 4 | `npm run typecheck` | `frontend/` |
| 5 | `wails build -tags webkit2_41` | repo root |

Each is also a target on its own: `make test`, `make vet`, `make lint`,
`make typecheck`, `make build`. "Green" means `make check` exited 0 — not that a
selected subset looked fine.

Gates 1 and 2 are specified over this project's own packages, not a literal `./...`:
a JS dependency ships Go source (`flatted` ships `golang/pkg/flatted`), so `./...`
would drag third-party code into the gates. The Makefile expands
`GOPKGS = $(shell go list ./... | grep -v /node_modules/)` and fails loudly if that
list is ever empty, so the gate cannot pass vacuously.

---

## Environment facts that are not negotiable

### E1 — `-tags webkit2_41` on every wails command

**`libwebkit2gtk-4.0-dev` does not exist on Ubuntu 24.04.** Only `libwebkit2gtk-4.1-dev`
is packaged. Wails v2 probes for the 4.0 pkg-config module by default and only looks
for 4.1 when the `webkit2_41` build tag is set. So:

```sh
wails build -tags webkit2_41
wails dev   -tags webkit2_41
```

The Makefile bakes it in as `$(TAGS)`. A wails command run without the tag failing to
link is the **expected** outcome, not a bug to debug.

System prerequisites (`sudo`, installed once by the user):

```sh
sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

Until those are installed, gate 5 fails with
`exec: "pkg-config": executable file not found in $PATH`. That is the environment, not
the code — do not work around it and do not disable the gate.

### E2 — the wails CLI is not necessarily on `PATH`

It lives at `$(go env GOPATH)/bin/wails` (here `/home/ismat/go/bin/wails`) and is not
guaranteed to be on `PATH` in non-login shells. Resolve it explicitly
(`WAILS ?= $(shell go env GOPATH)/bin/wails`); never assume a bare `wails` works.

### No cgo

The SQLite driver is **`modernc.org/sqlite`**, pure Go. `github.com/mattn/go-sqlite3`
or any other cgo driver must never enter the dependency graph.

```sh
CGO_ENABLED=0 go build ./...        # must always succeed
go list -deps ./... | grep -i mattn # must always return nothing
```

If a feature turns out to be unavailable in the pure-Go driver (FTS5 is the candidate),
**the feature degrades** — FTS5 falls back to `LIKE`-based search. The no-cgo guarantee
does not bend.

---

## Code rules

### All domain logic lives in Go

The frontend renders what Go returns and **computes nothing**: not a status, not a
streak, not progress, not an overdue flag, not a due date. Every derived value arrives
as a plain field on the DTO. New derivations go in `internal/domain` and are surfaced
through `internal/service` — never reimplemented in TypeScript. Two implementations of
a rule is two rules.

`internal/domain` is pure: no `database/sql`, no `os`, no `net`, and no `time.Now()` —
anything time-dependent takes an injected `now func() time.Time`.

Dependency direction is one-way: `main.go` → `internal/service` → `internal/store` →
`internal/domain`. Every Wails-bound method returns `(T, error)`.

### Colours only via Tailwind token names

`bg`, `surface`, `elevated`, `line`, `ink`, `muted`, `accent`, `accent-2`, `on-accent`,
`danger`, `warning`, `success`. Radii only from `rounded-sm/md/lg`, motion only from
`duration-fast/base/slow`, and JetBrains Mono for every number, date, timer and shortcut
hint.

**Zero hex literals in components** — this is grepped:

```sh
git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src   # must return nothing
```

**`design/` is read-only.** `design/tokens.css` and `design/tailwind.config.js` are the
source of truth; the frontend config consumes them and adds only its `content` globs.
Do not edit, copy or restate their values.

### i18n from day one

No hard-coded user-visible strings, anywhere, from the first component. Every string is
a key in `frontend/src/locales/en.json` and `frontend/src/locales/ru.json`. Russian runs
roughly **30% wider** than English — buttons, labels and columns must have the room, and
a layout that only works in English is a broken layout. Retrofitting i18n is miserable,
which is why it is not deferred.

---

## Commits

Conventional commits, **one commit per ticket**, subject referencing the ticket ID:

```
build(makefile): add make check with the five gates (S0-11)
```

Commits are authored as the configured git user, `Ismat <mukhamejanov.ismat@gmail.com>`.
**No AI author, no AI co-author trailer, no "Generated with" line, ever** (D7).

Stay inside the ticket's declared scope. If a ticket cannot be finished without touching
a file outside it, stop and say so rather than widening it silently. Every commit leaves
the tree in a working state: the gates that exist at that point still pass.

---

## See also

- [`PLAN.md`](./PLAN.md) — the brief, the data model, decisions D1–D11 and E1–E3.
- [`ARCHITECTURE.md`](./ARCHITECTURE.md) — package layout, dependency direction, build
  constraints, data and runtime paths.
- [`TASKS.md`](./TASKS.md) — tickets, acceptance criteria, stage DONE criteria.
- [`design/README.md`](./design/README.md) — tokens, palettes, fonts, motion (read-only).
- [`design/pmp-timelog-format.md`](./design/pmp-timelog-format.md) — the deterministic
  PMP KIT timelog spec.
