# CLAUDE.md — working rules for Nexus

Read this first. It is the short list of things that are easy to get wrong here.
It does not restate the plan: [`PLAN.md`](./PLAN.md) is the brief and the decisions
(D1–D15, E1–E3), [`ARCHITECTURE.md`](./ARCHITECTURE.md) is the layout,
[`TASKS.md`](./TASKS.md) is the ticket breakdown, and
[`design/README.md`](./design/README.md) is the token handoff.

---

## RESUME HERE — temporary handoff, delete when Stage 3 starts

> This section is scaffolding for picking the work back up on another machine.
> **Delete it once Stage 3 is underway.** Everything below it is permanent.

### Where the work stopped

| | |
|---|---|
| Stage 0 | **CLOSED** — Reviewer PASS |
| Stage 1 | **CLOSED** — Reviewer PASS on the fourth round, at `a1f09b7` |
| Stage 2 | **CLOSED** — Reviewer PASS on the second round, at `8818db4` |
| Stage 3 | Not started, not yet planned |

At `138b040` (85 commits): all five `make check` gates green including
`wails build -tags webkit2_41`; `make cover` `internal/domain` **100.0%**,
`internal/service` **93.8%** (bar ≥90%), `internal/store` 86.4% (not gated);
`make front-test` **247 tests across 22 files**; `make guard` six of six; tree clean.

### Do this first

**Plan Stage 3** — task detail panel, tree view, search and archive. Nothing else is
outstanding. `TASKS.md` has a `## Carried into Stage 3` section that Stage 3's tickets
must absorb rather than rediscover.

**The app runs now.** The Kanban board, habits strip, quick add (Ctrl+N) and command
palette (Ctrl+K) are all live and keyboard-reachable, in EN and RU.

### ⚠️ A green `make guard` is NOT a proof

This is the most important thing Stage 2 learned. `make guard`'s checks 3a/3b are
**name-based heuristics** — they fire on identifiers matching
`overdue|derive|streak|progress|percent`. A frontend function called `wireDate` computed
the local calendar day and sent it to Go for an entire stage, behind a permanently green
guard, until the Reviewer read the diff. The Reviewer then reproduced it: a planted
regression recomputing today, recomputing the overdue flag, and inlining a percentage in
JSX **passed all six checks**.

**No status line may offer a green `make guard` as evidence that the frontend computes
nothing. Reading the diff is still the check.** See `PLAN.md` §7 **D17**.

`GUARD_ALLOW_RE` is **empty and must stay empty** — every hit so far has been a real bug
fixed at the source, never allow-listed.

### Stage 3 must not forget these

- **Five things are owed to a hand pass** and cannot be verified on this machine (no
  display; `xvfb-run`, `scrot`, `import`, `grim` all absent): the ten-step keyboard-only
  run on `./build/bin/nexus` — including three D8 due-badge assertions that have **no**
  mechanical backing at all; no flash of the wrong background in either theme
  (D12/K1); Russian not clipping at a real 1024×768 (jsdom has no layout engine, so the
  suite asserts the anti-clipping *mechanisms*, never the absence of clipping); the
  `:focus-visible` ring actually painted in both palettes; and "a window opens and the
  first `Board()` returns five columns".
- **C6** — five enum sets (statuses, types, themes, palettes, accents) are spelled a
  second time as keys in `en.json`/`ru.json`, with nothing tying them to Go. Ruled a
  judgement call rather than a defect, since locale files must name what they translate.
  Stage 3 either binds a set-publishing method or adds a parity test, so that adding a
  value to a Go set turns something red.
- **No type rule may be spelled twice.** Three of Stage 1's four review failures, and
  one of Stage 2's two, were the same defect: a rule written down in two places and
  edited in one. The Go inventory is clean — `HasColumn`, `HasDue`,
  `DoingRefusal`/`CanBeDoing`, `countsAsWork`, habit-requires-recurrence and
  `DefaultActivity` are each the single definition of their rule. **This applies to
  TypeScript**: the frontend renders what Go returns and re-implements none of it.
- **Screenshots and README.** `README.md` is still not written. The board exists now, so
  it is finally writable — but automated capture needs
  `sudo apt install xvfb imagemagick` first.
- If `frontend/wailsjs/*` ever shows as modified with files truncated to zero bytes, an
  interrupted `wails build` did it. `make build` regenerates them.

### Known issues — decided; K1 shipped in Stage 2 (rulings in `PLAN.md` §7)

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
