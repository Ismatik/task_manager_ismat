# Nexus — TASKS

- **[Stage 0 — Scaffold](#stage-0--scaffold): DONE. Reviewer returned PASS.** All
  thirteen tickets committed, all five gates green, all seven DONE criteria met.
  Kept below as the record and as the template for ticket quality.
- **[Stage 1 — Domain + store](#stage-1--domain--store): DONE. Reviewer returned PASS
  on the fourth review, at `a1f09b7`.** Go only, no UI, no Wails bindings. All
  twenty-two tickets **S1-01 … S1-22** committed; ACCEPT met at **100.0% / 92.9%**
  against a ≥90% bar; all eleven DONE criteria met. **Review returned FAIL three times
  first**, on one family of related defects — that history is kept in full below and is
  the most useful thing in this file. See
  [Stage 1 — DONE criteria](#stage-1--done-criteria) and
  [Carried into Stage 2](#carried-into-stage-2).
- **Stage 2 — Kanban + Habits strip: not started, not yet planned.** What Stage 1 owes
  it is listed under [Carried into Stage 2](#carried-into-stage-2); those items are
  acceptance criteria, not suggestions.

Decisions referenced as **D1–D11 / E1–E3** and known issues **K1–K3** (plus **K4**,
now **RESOLVED** in `9664506`) live in
[`PLAN.md` §7](./PLAN.md). Four decisions were confirmed by the user *during* Stage 1
and are now recorded there — **D8** (a column move always overwrites the due date),
**D9** (type beats leaf-ness — a project is never timeable), **D10** (a node with no
Kanban column is excluded from parent derivation, generalising §4) and **D11** (an
empty project is one unfinished work leaf in its parent). The tickets below encode
their behaviour. **Stage 1 is closed; Stage 2 is not planned yet.** Do not implement
anything below that is not on a Stage 2 ticket, and do not invent Stage 2 tickets here
— the PM writes them in a separate pass.

---

## Rules for the Dev agent

1. **One ticket at a time.** Implement it, test it, commit it, stop. Do not start the
   next ticket in the same run.
2. **One commit per ticket**, conventional-commit format, using the prefix printed on
   the ticket. Subject line references the ticket ID, e.g.
   `build(makefile): add make check with the five gates (S0-11)`.
3. **Commits are authored as the configured git user**, `Ismat
   <mukhamejanov.ismat@gmail.com>`. **No AI author, no AI co-author trailer, no
   `Generated with` line.** (D7)
4. **Stay inside the ticket's Scope.** If a ticket cannot be completed without
   touching a file outside its scope, stop and report — do not widen it silently.
5. **Every ticket must leave the tree in a working state**: whatever gates exist at
   that point in the sequence must still pass.
6. **No cgo, ever.** `CGO_ENABLED=0` must remain a true statement about this build.

---

## Blocker the user had to clear (sudo required) — CLEARED

```sh
sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

These were **missing** while Stage 0 was being written, which is why gate 5 and the two
manual Stage 0 criteria were blocked. The user has since installed them:
`pkg-config --exists gtk+-3.0 webkit2gtk-4.1` now succeeds, gate 5 passes, and Stage 0
closed on a real build and a real window rather than on a promise. Nothing in Stage 1
is blocked on the environment.

**E1 — `libwebkit2gtk-4.0-dev` does not exist on Ubuntu 24.04.** Only the `-4.1`
package is available. Therefore **every `wails build` and `wails dev` invocation in
this project must pass `-tags webkit2_41`**, and the Makefile bakes it in. A build
command without that tag failing to link is the expected outcome, not something to
debug.

**E2 —** the Wails CLI (v2.16.0) is at `$(go env GOPATH)/bin/wails`
(`/home/ismat/go/bin/wails`) and is **not guaranteed to be on `PATH`** in non-login
shells. The Makefile must resolve it explicitly.

---

## The five gate commands

`make check` runs exactly these, in this order, failing fast:

| # | Command | Run from |
|---|---|---|
| 1 | `go test` over `./...` **excluding `node_modules`** | repo root |
| 2 | `go vet` over `./...` **excluding `node_modules`** | repo root |
| 3 | `npm run lint` | `frontend/` |
| 4 | `npm run typecheck` | `frontend/` |
| 5 | `wails build -tags webkit2_41` | repo root |

Gates 1 and 2 exclude `node_modules` because a JS dependency ships Go source
(`flatted` ships `golang/pkg/flatted`), and the gates must cover only this project's
packages. The Makefile expands the list as
`GOPKGS = $(shell go list ./... | grep -v /node_modules/)` and aborts loudly if that
list ever comes back empty, so an empty expansion can never pass vacuously.

**`make check` is introduced by ticket [S0-11](#s0-11--build-make-check-the-five-gates).**
Earlier tickets add the individual pieces; S0-11 is the ticket that wires all five
into one gate and makes "green" a single, checkable claim.

**"Green" always means these five and only these five.** The Stage 1 coverage
threshold is measured by a **separate** target, `make cover`
([S1-03](#s1-03--build-make-cover--the-90-coverage-measurement)); it is deliberately
**not** a sixth gate, so that the gate definition above stays exactly as blessed.

**Wherever an acceptance criterion below says `go test` or `go vet`, it means over
this project's packages — `./...` excluding `node_modules`** — spelled out as
`go test $(go list ./... | grep -v /node_modules/)` in tickets that predate the
Makefile's `$(GOPKGS)`, and as `make test` / `make vet` afterwards. A literal
`go test ./...` is never the gate.

---

## Stage 0 — Scaffold

**Status: DONE. Reviewer returned PASS.** Thirteen tickets, one commit each,
`S0-01` … `S0-13`. Kept in full below as the record of what was agreed and as the
template for Stage 1 ticket quality. **Do not re-open these tickets.**

## Stage 0 ticket index

| ID | Title | Prefix |
|---|---|---|
| [S0-01](#s0-01--chore-repository-baseline-and-gitignore) | Repository baseline and `.gitignore` | `chore:` |
| [S0-02](#s0-02--docs-architecturemd) | `ARCHITECTURE.md` | `docs:` |
| [S0-03](#s0-03--chore-wails-v2-react-ts-scaffold) | Wails v2 react-ts scaffold | `chore:` |
| [S0-04](#s0-04--build-makefile-dev--build--test) | Makefile — `dev` / `build` / `test` | `build:` |
| [S0-05](#s0-05--feat-tailwind-wired-to-the-design-tokens) | Tailwind wired to the design tokens | `feat:` |
| [S0-06](#s0-06--chore-eslint-prettier-and-typecheck) | ESLint, Prettier and `typecheck` | `chore:` |
| [S0-07](#s0-07--feat-go-package-layout-skeleton) | Go package layout skeleton | `feat:` |
| [S0-08](#s0-08--feat-embedded-sql-migrations-and-sqlite-open) | Embedded SQL migrations and SQLite open | `feat:` |
| [S0-09](#s0-09--feat-settings-table-and-repository) | `settings` table and repository | `feat:` |
| [S0-10](#s0-10--feat-single-instance-lock-and---quick-ipc) | Single-instance lock and `--quick` IPC | `feat:` |
| [S0-11](#s0-11--build-make-check-the-five-gates) | `make check` — the five gates | `build:` |
| [S0-12](#s0-12--docs-claudemd) | `CLAUDE.md` | `docs:` |
| [S0-13](#s0-13--docs-designpmp-timelog-formatmd) | `design/pmp-timelog-format.md` | `docs:` |

---

## S0-01 — chore: repository baseline and `.gitignore`

The repo has **zero commits**. This ticket creates the first one, so that every later
ticket has a clean diff to be reviewed against.

**Scope (may touch):** `.gitignore` (new). Commits the already-present `PLAN.md`,
`TASKS.md`, `design/README.md`, `design/tokens.css`, `design/tailwind.config.js`,
`design/SKILL.md`. **Creates no other file.**

`.gitignore` must cover, at minimum:

```
# build artifacts
/build/bin/
/frontend/dist/
/nexus

# deps
/frontend/node_modules/

# local data
*.db
*.db-shm
*.db-wal

# editor / os
.DS_Store
*.swp
```

**Acceptance criteria**
- [x] `git log` shows exactly one commit on `main`.
- [x] `git status --porcelain` is empty.
- [x] `PLAN.md`, `TASKS.md` and all four `design/` files are tracked.
- [x] No `.go`, `.ts`, `.tsx`, `.css`, `.js` or config file other than `.gitignore`
      was added.
- [x] Commit author is `Ismat <mukhamejanov.ismat@gmail.com>`; no co-author trailer.

**Commit:** `chore: initialise repository with plan, tasks, design handoff and gitignore (S0-01)`

---

## S0-02 — docs: `ARCHITECTURE.md`

Write down the layout **before** any code creates it, so S0-03 and S0-07 have an
authoritative target. Content comes from **D3**.

**Scope (may touch):** `ARCHITECTURE.md` (new). Nothing else.

Must contain, verbatim in substance:

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

Plus these rules, each stated explicitly:

- **All domain logic lives in Go**; the frontend renders what Go returns and computes
  no status, streak, progress, overdue flag or due date.
- `internal/domain` is **pure** — no database, no filesystem, no clock reads except
  through an injected `now func() time.Time`.
- **Dependency direction:** `main.go` → `service` → `store` → `domain`.
  `domain` imports nothing from the other three. `store` never imports `service`.
- Every Wails-bound method returns `(T, error)`.
- **No cgo**, ever — `modernc.org/sqlite` only.
- The `webkit2_41` build-tag requirement (**E1**) and why.

**Acceptance criteria**
- [x] `ARCHITECTURE.md` exists at repo root and lists all eight paths above.
- [x] States the dependency direction and the purity rule for `internal/domain`.
- [x] States the no-cgo rule and the `-tags webkit2_41` requirement.
- [x] Notes that `main.go` lives at the **repo root** (Wails convention), not
      `cmd/nexus/`.
- [x] No code files added.

**Commit:** `docs: add ARCHITECTURE.md describing the Go and frontend layout (S0-02)`

---

## S0-03 — chore: Wails v2 react-ts scaffold

**Scope (may touch):** `wails.json`, `main.go`, `app.go`, `go.mod`, `go.sum`,
`frontend/**` (as generated), `build/**` (as generated), `.gitignore` (append only if
the generated tree needs it).

**Notes the Dev must heed**

- The repo root is **not empty**, and `wails init` refuses to scaffold into a
  non-empty directory. Generate into a temporary directory
  (`wails init -n nexus -t react-ts -d /tmp/nexus-scaffold`) and move the result in.
  Do **not** delete or overwrite `PLAN.md`, `TASKS.md`, `ARCHITECTURE.md` or `design/`.
- Invoke the CLI as `$(go env GOPATH)/bin/wails` (**E2**).
- Module name: `nexus`. `go.mod` must say **`go 1.26`** (**D7**) — edit it if the
  template writes something older.
- **Commit `frontend/wailsjs/`.** It is generated by `wails dev`/`wails build`, but
  `npm run typecheck` (gate 4) imports from it, and gate 4 must be able to pass on a
  machine where gate 5 cannot yet run. Generate it once and track it.
- Do not add Tailwind, ESLint or Prettier here — those are S0-05 and S0-06.

**Acceptance criteria**
- [x] `go build ./...` succeeds (this does **not** need GTK/WebKit).
- [x] `go vet $(go list ./... | grep -v /node_modules/)` is clean.
- [x] `cd frontend && npm install && npm run build` succeeds.
- [x] `go.mod` declares `go 1.26` and module `nexus`.
- [x] `frontend/wailsjs/` is tracked; `frontend/node_modules/` is **not** tracked.
- [x] `main.go` and `app.go` sit at the repo root; `ARCHITECTURE.md`, `PLAN.md`,
      `TASKS.md`, `design/` are untouched.
- [x] `git grep -n "webkit2gtk-4.0"` returns nothing.

**Commit:** `chore: scaffold wails v2 react-ts application (S0-03)`

---

## S0-04 — build: Makefile — `dev` / `build` / `test`

Get the correct invocations recorded in one place **early**, so nobody hand-types a
`wails build` without the tag. `make check` is deliberately *not* in this ticket.

**Scope (may touch):** `Makefile` (new).

Requirements:

- `WAILS ?= $(shell go env GOPATH)/bin/wails` — resolve the CLI explicitly (**E2**),
  with a friendly error if that path does not exist.
- `TAGS ?= webkit2_41` used by **every** wails invocation (**E1**).
- `CGO_ENABLED=0` exported for Go builds and tests.
- Targets: `dev`, `build`, `test`, `clean`, and `help` as the default goal.
- `.PHONY` on every target.
- A comment at the top of the file pointing at the apt prerequisites and stating
  plainly that `-tags webkit2_41` is mandatory and why.

**Acceptance criteria**
- [x] `make` with no argument prints help and exits 0.
- [x] `make test` runs `go test` over this project's packages — `./...` **excluding
      `node_modules`** — and passes. (S0-11 formalises the exclusion as `$(GOPKGS)`;
      `make test` must already not be a bare `go test ./...`.)
- [x] `make clean` removes `build/bin` and `frontend/dist` and is idempotent.
- [x] `grep -c 'webkit2_41' Makefile` ≥ 1, and **no** wails invocation in the file
      lacks `$(TAGS)`.
- [x] `make build` is *expected to fail* on this machine until the apt packages land;
      the failure must be the linker/pkg-config error, not "wails: command not found".

**Commit:** `build: add Makefile with dev, build and test targets (S0-04)`

---

## S0-05 — feat: Tailwind wired to the design tokens

**Scope (may touch):** `frontend/package.json`, `frontend/package-lock.json`,
`frontend/tailwind.config.js` (new), `frontend/postcss.config.js` (new),
`frontend/src/style.css` (or the template's equivalent entry CSS),
`frontend/index.html`, `frontend/src/main.tsx`. **`design/` is read-only.**

**Notes the Dev must heed**

- **Pin Tailwind v3.** `design/tailwind.config.js` is a CommonJS v3 config with
  `darkMode: 'class'` and `theme.extend`. Tailwind v4 changes the configuration model
  and would silently drop it. Install `tailwindcss@^3`.
- `design/tokens.css` is **the source of truth and must not be edited or duplicated**.
  Import it; do not copy its values. Same for `design/tailwind.config.js` — the
  frontend config should `require`/re-export it and add only `content` globs.
- Set `<html data-palette="aurora" class="dark">` as the static default (**D6**,
  `design/README.md`). Runtime switching is Stage 2 — do not build it here.
- Load Space Grotesk, Figtree and JetBrains Mono **locally** (this is an offline,
  local-only app — no CDN `<link>` to Google Fonts).

**Acceptance criteria**
- [x] `frontend/tailwind.config.js` consumes `design/tailwind.config.js` rather than
      restating the token map.
- [x] `design/tokens.css` is imported by the entry CSS and is **byte-identical** to
      its committed state (`git diff --exit-code design/`).
- [x] `npm run build` succeeds and the emitted CSS contains the token custom
      properties.
- [x] A throwaway element using `bg-surface text-ink rounded-md font-mono
      duration-fast` resolves to the token values (verify in the built CSS).
- [x] `index.html` has `data-palette="aurora"` and `class="dark"` on `<html>`.
- [x] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing.
- [x] Tailwind resolves to a `3.x` version in `package-lock.json`.

**Commit:** `feat(frontend): wire tailwind to the design tokens and palettes (S0-05)`

---

## S0-06 — chore: ESLint, Prettier and `typecheck`

**Scope (may touch):** `frontend/package.json`, `frontend/package-lock.json`,
`frontend/eslint.config.js` (or `.eslintrc.cjs`), `frontend/.prettierrc`,
`frontend/.prettierignore`, `frontend/tsconfig.json`, and mechanical formatting-only
changes to existing `frontend/src/**`.

Requirements:

- Scripts: `"lint"`, `"lint:fix"`, `"format"`, `"typecheck"` (`tsc --noEmit`).
- `npm run lint` must exit **non-zero on warnings** (`--max-warnings=0`) — a gate that
  tolerates warnings is not a gate.
- TypeScript `strict: true`.
- ESLint must **ignore** `frontend/wailsjs/` (generated) and `frontend/dist/`.
- Pick **one** ESLint major and stick to it; if using v9, use flat config
  (`eslint.config.js`) and do not leave an `.eslintrc` behind.
- Prettier and ESLint must not fight — no formatting rules in ESLint.

**Acceptance criteria**
- [x] `cd frontend && npm run lint` exits 0 with zero warnings.
- [x] `cd frontend && npm run typecheck` exits 0.
- [x] `npx prettier --check .` exits 0.
- [x] `tsconfig.json` has `"strict": true`.
- [x] Deliberately introducing an unused variable makes `npm run lint` exit non-zero
      (demonstrate, then revert).
- [x] `npm run build` still succeeds.

**Commit:** `chore(frontend): add eslint, prettier and typecheck scripts (S0-06)`

---

## S0-07 — feat: Go package layout skeleton

Create the packages from **D3 / `ARCHITECTURE.md`** as compiling, tested skeletons.
**No domain rules are implemented here** — those are Stage 1.

**Scope (may touch):** `internal/domain/**`, `internal/store/**`,
`internal/service/**`, `internal/parse/**`, `internal/platform/**`.

Each package gets a `doc.go` stating its responsibility and its import constraints,
plus at least one real (non-empty, non-skipped) test so the package is exercised by
gate 1.

Minimum real content — types only, no behaviour:

- `internal/domain` — `NodeType`, `Status`, `DueSource` (`manual` / `auto`, **D1**),
  `Activity` (the seven values from **D4**), `Priority`, with `String()` and a
  `Valid()` validator, and table-driven tests for those. **No I/O, no `time.Now()`.**
- `internal/store`, `internal/service`, `internal/parse`, `internal/platform` —
  `doc.go` and a compile-level placeholder only.

**Acceptance criteria**
- [x] All five packages exist and `go build ./...` succeeds.
- [x] `go test $(go list ./... | grep -v /node_modules/)` passes and reports **no**
      `[no test files]` for `internal/domain`.
- [x] `go vet $(go list ./... | grep -v /node_modules/)` is clean.
- [x] `internal/domain` imports nothing from `store`, `service`, `parse`, `platform`,
      `database/sql`, `os` or `net` — assert this with a test or a `go list` check
      documented in the commit body.
- [x] `DueSource` has exactly the values `manual` and `auto`; `Activity` has exactly
      the seven values from D4.
- [x] No status derivation, no streak logic, no column↔due logic (Stage 1).

**Commit:** `feat: add internal package layout skeleton with core enums (S0-07)`

---

## S0-08 — feat: embedded SQL migrations and SQLite open

**Scope (may touch):** `internal/store/migrations/**` (new `.sql` files),
`internal/store/migrate.go`, `internal/store/db.go`, `internal/store/*_test.go`,
`go.mod`, `go.sum`.

**Scope decision — what migration 0001 contains.** Stage 0 delivers the **migration
machinery plus the `settings` table only**. The full `nodes` / `tags` / `time_entries`
/ `attachments` / `habit_checks` schema lands in Stage 1 alongside the repositories
and domain rules that use it — writing it now would guarantee churn before anything
reads it. The `settings` table itself is created by **S0-09**; this ticket delivers
the runner and an empty-but-valid migration set.

Requirements:

- `//go:embed migrations/*.sql` — migrations compiled into the binary, never read
  from disk at runtime.
- Driver `modernc.org/sqlite` (pure Go). **`CGO_ENABLED=0 go build ./...` must
  succeed** — that is the no-cgo proof.
- A `schema_migrations` table recording applied versions; each migration applied
  **once**, in filename order, inside a transaction.
- Filenames `NNNN_snake_case.sql`, zero-padded to 4.
- DB path from an injectable parameter, defaulting to
  `$XDG_DATA_HOME/nexus/nexus.db` (fallback `~/.local/share/nexus/nexus.db`), with
  the parent directory created on open.
- Pragmas on open: `journal_mode=WAL`, `foreign_keys=ON`, `busy_timeout`.

**Acceptance criteria**
- [x] `CGO_ENABLED=0 go build ./...` succeeds.
- [x] `go list -deps ./... | grep -i mattn` returns nothing (no cgo sqlite driver
      crept in).
- [x] Test: migrating a fresh temp DB succeeds; migrating **again** is a no-op and
      applies zero migrations.
- [x] Test: `foreign_keys` reads back as `1` and `journal_mode` as `wal`.
- [x] Test: a deliberately failing migration rolls back and leaves
      `schema_migrations` unchanged.
- [x] Tests use `t.TempDir()`; no `*.db` file is left in the working tree
      (`git status --porcelain` empty after a full test run).
- [x] `go test` and `go vet` over `./...` **excluding `node_modules`** both pass.

**Commit:** `feat(store): add embedded sql migrations and sqlite connection (S0-08)`

---

## S0-09 — feat: `settings` table and repository

**Scope (may touch):** `internal/store/migrations/0001_settings.sql` (new),
`internal/store/settings.go`, `internal/store/settings_test.go`.

Requirements:

- `settings(key TEXT PRIMARY KEY, value TEXT NOT NULL)`.
- Repository with `Get(key) (string, bool, error)`, `Set(key, value) error` (upsert),
  `All() (map[string]string, error)`.
- Defaults seeded idempotently on first run, per `design/README.md` and **D6**:

  | key | default |
  |---|---|
  | `palette` | `aurora` |
  | `theme` | `dark` |
  | `accent` | *(empty — means "use the palette's `--accent`")* |
  | `language` | `en` |

- Seeding must **never overwrite** a value the user has already set.

**Acceptance criteria**
- [x] Migration applies cleanly on a fresh DB via the S0-08 runner.
- [x] Test: round-trip `Set` → `Get`; `Set` twice on the same key upserts rather than
      erroring or duplicating.
- [x] Test: `Get` on a missing key returns `found == false` and **no** error.
- [x] Test: seeding defaults twice leaves exactly four rows; a user-modified
      `palette=studio` survives a second seed.
- [x] `go test` and `go vet` over `./...` **excluding `node_modules`** both pass; no
      stray `*.db` in the tree.

**Commit:** `feat(store): add settings table with seeded defaults (S0-09)`

---

## S0-10 — feat: single-instance lock and `--quick` IPC

The last functional piece of the Stage 0 ACCEPT criteria.

**Scope (may touch):** `main.go`, `app.go`, `internal/platform/singleinstance.go`,
`internal/platform/singleinstance_test.go`.

**Design (decided — do not re-litigate):** a **unix domain socket**, not D-Bus. D-Bus
drags in a session-bus dependency and a service name for what is a one-line message,
and `internal/platform` already owes D-Bus work in Stage 4 for sleep/lock — keeping
these separate keeps Stage 0 small.

- Socket at `$XDG_RUNTIME_DIR/nexus/ipc.sock`, falling back to `/tmp/nexus-$UID/`.
- On start, try to `Listen`. Success ⇒ **this is the primary instance**.
- If the socket exists but nothing accepts (stale socket from a crash), **unlink and
  retry once**, then give up with a clear error.
- Failure to listen because someone *is* accepting ⇒ **this is a secondary instance**:
  connect, send one line, exit **0** immediately.
  - `--quick` ⇒ send `quick`.
  - no flag ⇒ send `focus`.
- Primary handles messages: `focus` ⇒ show/unminimise/raise the main window.
  `quick` ⇒ for Stage 0, **log the request and focus the main window**; the real
  quick-add window is **Stage 4** and must not be built here.
- Clean up the socket on shutdown and on `SIGINT`/`SIGTERM`.
- Flag parsing must not break Wails' own argument handling.

**Acceptance criteria**
- [x] Test (no GUI needed): first `Acquire` succeeds; a second `Acquire` in the same
      test reports "already running" and the message is received by the first.
- [x] Test: a stale socket file with no listener is reclaimed, not fatal.
- [x] Test: `--quick` and bare launch produce distinguishable messages on the wire.
- [x] Test: the socket file is removed after the primary shuts down.
- [x] Tests use a temp runtime dir and leave nothing behind.
- [x] `go test` and `go vet` over `./...` **excluding `node_modules`** both pass.
- [x] **Manual (blocked on apt, see the blocker note):** `make build && ./build/bin/nexus`
      opens an empty window; a second `./build/bin/nexus` exits immediately and the
      first window is focused.
- [x] No quick-add UI was built (Stage 4 scope).

**Commit:** `feat(platform): add single-instance lock with --quick ipc message (S0-10)`

---

## S0-11 — build: `make check` — the five gates

**This is the ticket that introduces `make check`.**

**Scope (may touch):** `Makefile`, `frontend/package.json` (only if a script name
needs aligning).

`make check` runs exactly these five, in order, **failing fast on the first failure**:

| # | Command | Run from |
|---|---|---|
| 1 | `go test` over `./...` **excluding `node_modules`** | repo root |
| 2 | `go vet` over `./...` **excluding `node_modules`** | repo root |
| 3 | `npm run lint` | `frontend/` |
| 4 | `npm run typecheck` | `frontend/` |
| 5 | `wails build -tags webkit2_41` | repo root |

Requirements:

- Individual targets `test`, `vet`, `lint`, `typecheck`, `build` exist and are
  runnable alone; `check` depends on them in that order.
- Gates 1 and 2 run over `$(GOPKGS) = $(shell go list ./... | grep -v /node_modules/)`,
  not a literal `./...` — `frontend/node_modules` contains real Go source. The
  expansion must abort with a hard error when the list is empty, so that a broken
  `go list` cannot degrade the gate into a bare `go test` of the root directory.
- Gate 5 goes through `$(WAILS)` and `$(TAGS)` from S0-04 — **never** a bare `wails`.
- If `frontend/node_modules` is missing, gates 3–4 must `npm ci` first rather than
  fail confusingly.
- On gate-5 failure from a missing system library, print the
  `sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev` hint.
- Ordering matters: the cheap, currently-passable gates run **before** the one that is
  environmentally blocked, so a Dev gets useful signal even before the apt install.

**Acceptance criteria**
- [x] `make check` exists and runs all five commands in the stated order.
- [x] Gates 1–4 pass on this machine **today**.
- [x] Gate 5 is the only failing gate, and only for the missing-apt-package reason.
- [x] Making gate 1 fail on purpose stops the run before gate 2 (demonstrate, then
      revert).
- [x] `make check` re-run twice in a row gives the same result (no hidden state).
- [x] `git grep -n 'wails ' Makefile` shows every invocation carrying `$(TAGS)`.

**Commit:** `build: add make check running the five gate commands (S0-11)`

---

## S0-12 — docs: `CLAUDE.md`

**Scope (may touch):** `CLAUDE.md` (new).

Must state, concisely and without restating `PLAN.md`:

- What Nexus is, and **local-only / no network** as a hard rule.
- The **five gate commands** and that `make check` is the single entry point.
- **E1**: `-tags webkit2_41` on every wails command, and that `libwebkit2gtk-4.0-dev`
  does not exist on Ubuntu 24.04.
- **E2**: the wails CLI lives at `$(go env GOPATH)/bin/wails`, possibly off `PATH`.
- **No cgo.** `modernc.org/sqlite` only.
- **All domain logic in Go**; the frontend computes nothing.
- **Colours only via Tailwind token names**; zero hex literals in components;
  `design/` is read-only.
- **i18n from day one** — no hard-coded UI strings; every string a key in
  `frontend/src/locales/{en,ru}.json`; Russian needs ~30% more width.
- Commit convention: conventional commits, one per ticket, **no AI author or
  co-author** (**D7**).
- Pointers to `PLAN.md`, `ARCHITECTURE.md`, `TASKS.md`, `design/README.md`.

**Acceptance criteria**
- [x] `CLAUDE.md` exists at repo root and covers every bullet above.
- [x] Lists the five gate commands verbatim.
- [x] Mentions `webkit2_41` and the no-cgo rule explicitly.
- [x] Does not duplicate `PLAN.md` wholesale — it links instead.
- [x] No code files touched.

**Commit:** `docs: add CLAUDE.md with project rules and gate commands (S0-12)`

---

## S0-13 — docs: `design/pmp-timelog-format.md`

The deterministic counterpart to `design/SKILL.md`, per **D4**. `SKILL.md` is an
*interactive* skill; this file specifies a *generator* driven by tracked
`time_entries`. Stage 6 implements it. **This ticket writes the spec only.**

**Scope (may touch):** `design/pmp-timelog-format.md` (new).
**`design/SKILL.md`, `tokens.css`, `tailwind.config.js`, `README.md` are read-only.**

Must specify:

- The four PMP KIT form fields: **Дата**, **Часы**, **Деятельность**, **Комментарий**,
  and the output block format from `SKILL.md` §3.
- Node field **`activity TEXT`**, values restricted to exactly:
  `Разработка, Анализ, Тестирование, Документация, Совещание, Согласование,
  Управление проектом`.
- **Defaults by node type**, user-overridable: `task` → Разработка,
  `bug` → Тестирование, `project` → Управление проектом, `note` → Документация.
  (`habit` is out of scope for time logging — state that.)
- **Комментарий is GENERATED then user-edited**: generated from node `title` (short
  line, ≤80 chars, no trailing period) and `description_md` (detailed block, 2–5
  sentences), then freely editable before copying.
- **Hours rounded to the nearest 15 minutes**, rendered in platform format
  (`2h`, `1.5h`, `30m`). Worked example table of raw minutes → rounded output,
  including a `:07` / `:08` boundary case.
- Hours come **only** from measured `time_entries` — never invented, never estimated
  (`SKILL.md` "Никогда не выдумывать").
- Multiple entries in a period ⇒ listed, with **итого часов** at the end.
- Text rules inherited from `SKILL.md`: impersonal, result-oriented, no filler,
  technical terms untranslated.
- A short section on **what differs from `SKILL.md`**: no clarifying questions are
  asked — every input comes from the database.

**Acceptance criteria**
- [x] File exists at `design/pmp-timelog-format.md`.
- [x] Lists all seven activities **and** all four type-defaults.
- [x] States the 15-minute rounding rule with a worked example including a boundary
      case.
- [x] States that Комментарий is generated then user-edited.
- [x] States that hours come only from measured `time_entries`.
- [x] Shows the exact output block.
- [x] `git diff --exit-code design/SKILL.md design/tokens.css
      design/tailwind.config.js design/README.md` is clean.

**Commit:** `docs: add deterministic pmp timelog format spec (S0-13)`

---

## Stage 0 — DONE criteria — ALL MET, PASS

Stage 0 closed only once **all** of these held. They do:

1. [x] All thirteen tickets are committed, one commit each, in order.
2. [x] `make check` is green — **all five gates**, including
   `wails build -tags webkit2_41`.
3. [x] `./build/bin/nexus` opens an empty window.
4. [x] A second `./build/bin/nexus` exits immediately and focuses the first window.
5. [x] `git status --porcelain` is empty after a full `make check`.
6. [x] `CLAUDE.md`, `ARCHITECTURE.md` and `design/pmp-timelog-format.md` all exist.
7. [x] `git log` shows no AI author or co-author on any commit.

Criteria 2–4 were blocked on the apt install; the user installed the packages, they
were then **verified for real**, and only then did the Reviewer return **PASS**.
Nothing was marked done that was not actually verified.

**Out of Stage 0 scope** (deferred, and now largely Stage 1's job): the `nodes`
schema, every domain rule (status derivation, column↔due, streaks, timer), FTS5.
Still out of scope at the end of Stage 1: Kanban, quick-add UI, tray, i18n string
extraction beyond the locale file stubs.

---

## Stage 1 — Domain + store

**Status: DONE — CLOSED, PASS.** The Reviewer returned **PASS** on the **fourth**
review, verified at commit `a1f09b7`. Tickets `S1-01` … `S1-22`, all twenty-two
committed, one conventional commit each, no AI author or co-author on any of them.
All five gates green including `wails build -tags webkit2_41`. Kept in full below as
the record of what was agreed. **Do not re-open these tickets.**

**ACCEPT met** — the bar is ≥90% statement coverage per package on `internal/domain`
and `internal/service`; re-measured after `go clean -testcache` the actuals are
**100.0%** and **92.9%** (`internal/store` is 86.4% and deliberately not gated).

**Review returned FAIL three times before that**, every time on the same family of
defects: a type rule spelled in more than one place, with one copy diverging. Each
divergence was a door illegal rows walked through. The history is kept in full — it is
the most useful record in this file.

- **First review** — two code defects, fixed by the Dev in `fd5e31d` (illegal
  type/status combinations accepted on create) and `f1802d7` (due dates written onto
  types with no column), plus the missing record of decisions **D8** and **D9**,
  written into `PLAN.md` §7 by the PM in `e7d74cc`.
- **Second review** — the remaining doors of the same rule: `1cbe582` (the rule routed
  through `domain.NodeType.HasColumn()` everywhere, one spelling), `f266bf5` (a due
  date refused on a `note`; D9's wording corrected — a `habit` may have a date),
  `81fb6e4` (**every** no-column type excluded from derivation — **D10**), and
  `d5a170b` (an empty project counted as one unfinished work leaf — **D11**). **D10**
  and **D11** were recorded in `PLAN.md` §7 by the PM in `0854ed5`, which also
  generalised §4's derivation rule to match.
- **Third review** — the last surviving divergence, the one `PLAN.md` had written off
  as harmless in **K4**: `ValidateMove` refused only a `note` as a parent, so a habit
  could take children. That made the habit itself render in a Kanban column and made an
  ancestor project read `done` behind a 100% bar with the unfinished task still on the
  board — ~32,300 + ~800 invariant violations in a 200,000-forest sweep. Fixed by the
  Dev in `9664506`: every no-column parent refused via `NodeType.HasColumn`,
  `deriveStatus` cut at the node itself, and `ErrNoteParent` renamed to
  **`ErrTypeHasNoChildren`** with **no alias** (an alias would be a second spelling).
  The same commit folded `CanEnterDoing`, `CheckStatus`'s sentinel selection,
  `PlanCascade` and `canBeTimed` onto one predicate, **`domain.DoingRefusal`**, with
  `NodeType.CanBeDoing()` defined in terms of it. **K4 is resolved**, not deferred —
  rewritten from "harmless" to **RESOLVED** in `PLAN.md` by the PM in `e3d626f`, and
  kept on the record there rather than deleted so the failure mode stays visible.
- **Fourth review — PASS**, at `a1f09b7`. Every fix from the three failed rounds
  re-checked: five gates green including `wails build -tags webkit2_41`; coverage
  re-measured after `go clean -testcache` at **100.0% / 92.9%**; `PLAN.md` §7 **D2**
  and the **D9** sub-point correctly generalised, with an independent grep finding no
  bullet left that states a *current* rule in `note`-only terms — that final doc fix is
  `a1f09b7` itself; no `.go` file changed since `9664506`; tree clean; sole author
  `Ismat <mukhamejanov.ismat@gmail.com>` throughout. **Stage 1 closes here.**

The eleven commits of the review cycle, in order: `fd5e31d`, `f1802d7`, `e7d74cc`
(round 1) · `1cbe582`, `f266bf5`, `81fb6e4`, `d5a170b`, `0854ed5` (round 2) ·
`9664506`, `e3d626f` (round 3) · `a1f09b7` (round 4, PASS).

**No type rule is spelled twice any more** — `HasColumn`, `HasDue`,
`DoingRefusal`/`CanBeDoing`, `countsAsWork`, the habit-requires-recurrence check and
`DefaultActivity` are each the single definition of their rule. `countsAsWork` and
`DoingRefusal` admit the same types today but answer different questions (D7/D11 vs
D9) and are kept separate on purpose. **Stage 2 must not reintroduce a second
spelling of any of them.**

**All fixes landed and were re-checked.** The two user decisions the D10/D11 commits
required are recorded in `PLAN.md` §7, and §4's derivation rule has been generalised to
match. The Reviewer verified all of it on the fourth round, so **the stage is closed**.
What it owes Stage 2 is in [Carried into Stage 2](#carried-into-stage-2).

**Go only. No UI, no Wails bindings, no TypeScript.** Not one line of
`frontend/src` changes in this stage, and no method is added to `app.go`. Stage 2 owns
the bindings; a binding added now would be a binding designed before the thing it
binds to is finished. The only non-Go files any Stage 1 ticket may touch are the
`Makefile` and `.gitignore` (S1-03), `go.mod` / `go.sum` (S1-11), `.sql` migrations
under `internal/store/migrations/`, and `internal/store/migrations/README.md` (S1-01).

Everything here implements **`PLAN.md` §4** under decisions **D1, D2, D5, D7** and,
as of the Stage 1 reviews, **D8**, **D9**, **D10** and **D11**. Where this file and
`PLAN.md` disagree, `PLAN.md` wins and the disagreement is a bug in this file — report
it rather than guessing.

**D8, D9, D10 and D11 were confirmed by the user mid-stage**, after implementation
exposed questions D1–D7 did not answer. They are recorded in full in `PLAN.md` §7; in
short:

- **D8** — a column move to **Today** or **This week** **always** overwrites an
  existing due date, *including a manual one*, and flips `due_source` to `auto`. The
  column move always wins, so the column and the date can never disagree. Affects
  **S1-08** and **S1-18**.
- **D9** — **type beats leaf-ness**: a `project` never enters `doing` and never starts
  a timer, **even with no children at all**, resolving the contradiction between D7
  (projects never enter `doing`) and D2 (a childless node "behaves as a leaf").
  Consequences already encoded below: `ErrProjectNeverDoing` from `MoveToColumn` and
  from create; `PlanCascade` skips project descendants when cascading `doing`; a
  project is **not a unit of work in the progress denominator**, so an empty project —
  or one with nothing under it that counts as work, all notes, all habits or any mix of
  types with no Kanban column (**D10**) — reports `Defined() == false`, neither 0% nor
  100%; and types with no column (`note`, `habit`) are refused by both create and
  `MoveToColumn` and never receive a due date. Affects **S1-07**, **S1-10**, **S1-18**,
  **S1-19**, **S1-21**.
- **D10** — **a node with no Kanban column is excluded from parent status derivation**,
  exactly as a `note` already was. This **generalises `PLAN.md` §4's derivation rule**,
  which is now worded as "nodes with no Kanban column are excluded" rather than "note
  nodes are excluded". D9's principle: a type with no column cannot contribute to a
  column-derived status. Encoded as: `DeriveStatus` skips any child where
  `!child.Type.HasColumn()`; `Node.IsLeaf` treats a node whose children all lack a
  column as a leaf; `walkProgress` cuts at the same predicate (forced — cutting
  derivation without cutting progress rebuilds the same bug one level down); everything
  routes through `domain.NodeType.HasColumn()`. Affects **S1-06**, **S1-07**,
  **S1-10**, **S1-21**.
- **D11** — **an empty project counts as one unfinished work leaf in its parent's
  progress denominator**, reversing part of D9's "a project is never a unit of work"
  for the empty case only. Two questions, two answers, both true and both binding on
  Stage 2: a project's **own** progress is still `Defined() == false` when nothing
  under it is work (**unchanged**, S1-07's reasoning still binds), while a **leaf**
  project (no children with a column, per D10) counts as **one work leaf** in its
  **parent's** denominator, done iff its own stored status is `done` (**the change**).
  A project with real children is still not a unit itself; its children are counted.
  Derivation needed no change. Affects **S1-07**, **S1-21**.

### What Stage 1 must deliver

- Repositories for `nodes`, `tags`/`node_tags`, `time_entries`, `attachments` and
  `habit_checks`.
- Tree operations: create, move a subtree, reorder, archive, restore.
- Derived status and progress (**D2**, **D7**, **D9**, **D10**, **D11**) — computed,
  never stored.
- The column↔due coupling and `due_source` transitions (**D1**, **D8**).
- A timer service holding the **single-active invariant**, and refusing to time a
  project (**D9**).
- Habit streaks over **scheduled RRULE occurrences** (**D5**).
- Search over `title` + `description_md`, FTS5 if it exists, `LIKE` if it does not.

### ACCEPT — how coverage is measured

> **≥90% statement coverage on `internal/domain` AND `internal/service`,
> per package, separately.**

Not combined, not averaged, not "the repo overall". Each of the two packages must
reach 90.0% on its own. The exact commands — these, verbatim, are what the Reviewer
runs:

```sh
go test -covermode=atomic -coverprofile=coverage.domain.out  ./internal/domain/...
go tool cover -func=coverage.domain.out  | tail -1    # -> total: (statements) NN.N%

go test -covermode=atomic -coverprofile=coverage.service.out ./internal/service/...
go tool cover -func=coverage.service.out | tail -1    # -> total: (statements) NN.N%
```

The number that counts is the `total:` line's percentage from `go tool cover -func`.
**S1-03 wraps exactly these in `make cover`**, which parses both totals and exits
non-zero if either is below `90.0`, so "green" is one command and not a judgement
call. The Reviewer may run either form; they must agree.

Today's baseline, measured on the Stage 0 tree: `internal/domain` **100.0%**,
`internal/service` **100.0%**. The threshold therefore starts satisfied and every
ticket keeps it satisfied — no ticket is allowed to land code and defer its tests to a
later one. `make cover` is an acceptance criterion on **every** ticket from S1-03
onwards that touches `internal/domain` or `internal/service`.

`internal/store` is **not** coverage-gated (it sits at 77.5% today and its error paths
are driver-dependent). That is a deliberate consequence of the design rule below, not
an excuse:

> **Rules live in `domain`; orchestration lives in `service`; `store` stays thin.**
> A repository method that does anything more interesting than map a row to a struct
> and back is a rule in the wrong package. If `store` is where the logic went, the
> coverage number will look fine and the stage will still have failed.

Store repositories are still tested — every ticket below says so — they are just not
where the threshold is enforced.

### Purity, still

`internal/domain` stays pure: no `database/sql`, no `os`, no `net`, **no
`time.Now()`**. The clock arrives as an injected `now func() time.Time`, and
`service.Clock` (already in the tree) is where it comes from. `TestDomainIsPure` and
`TestDomainReadsNoClock` in `internal/domain/purity_test.go` enforce this mechanically
and **must not be weakened** — if a Stage 1 ticket makes one of them fail, the ticket
is wrong, not the test.

Dependency direction is unchanged and one-way:
`main.go → service → store → domain`.

### Migration ordering

The Stage 0 runner applies `NNNN_snake_case.sql` in **filename order**, each inside
its own transaction, recording it in `schema_migrations`. Consequences for this stage:

- **`0001_settings.sql` is shipped and is never edited.** Not to add a column, not to
  fix a typo. Corrections are new files.
- New migrations **append**: `0002_core_schema.sql` (S1-05), `0003_search.sql` (S1-17).
- Migration files contain **no** `BEGIN` / `COMMIT` / `ROLLBACK` — the runner owns the
  transaction.
- A migration that has been committed and run is immutable from the next ticket
  onwards. Within the ticket that introduces it, it may of course still be edited.

### Rules for the Dev agent, Stage 1 additions

In addition to the six rules at the top of this file:

7. **Tests are table-driven** for every rule. One `t.Run` per case with a name that
   reads as the rule it checks. A rule with one test case is an untested rule.
8. **Every edge case named in the ticket gets its own named subtest**, so the Reviewer
   can find it by `go test -run`.
9. **No new dependency without justification in the commit body**, and none that pulls
   in cgo. `CGO_ENABLED=0 go build ./...` and an empty
   `go list -deps ./... | grep -i mattn` remain true after every ticket.
10. **UUIDs**: `github.com/google/uuid` is already in the module graph (indirect, via
    Wails). Promoting it to a direct dependency is fine and expected; generation is a
    **service** concern, not a domain one — `domain` must not generate IDs any more
    than it reads the clock.

---

## Stage 1 ticket index

| ID | Title | Prefix |
|---|---|---|
| [S1-01](#s1-01--chore-refresh-the-stale-forward-looking-doc-comments-in-internalstore) | Refresh stale doc comments in `internal/store` | `chore:` |
| [S1-02](#s1-02--fix-window-background-alpha-is-0255-not-01) | Window background alpha is 0–255, not 0–1 | `fix:` |
| [S1-03](#s1-03--build-make-cover--the-90-coverage-measurement) | `make cover` — the ≥90% coverage measurement | `build:` |
| [S1-04](#s1-04--feat-fts5-availability-spike--the-search-backend-decision-point) | FTS5 availability spike — decision point | `feat:` |
| [S1-05](#s1-05--feat-migration-0002--the-core-schema) | Migration `0002` — the core schema | `feat:` |
| [S1-06](#s1-06--feat-node-tag-timeentry-and-habitcheck-types) | `Node`, `Tag`, `TimeEntry`, `HabitCheck` types | `feat:` |
| [S1-07](#s1-07--feat-derived-status-and-progress-d2-d7-d9) | Derived status and progress (D2, D7, D9; amended by D10, D11) | `feat:` |
| [S1-08](#s1-08--feat-columndue-rules-and-overdue-d1-d8) | Column↔due rules and overdue (D1, D8) | `feat:` |
| [S1-09](#s1-09--feat-tree-operations--move-cycle-rejection-reorder) | Tree ops — move, cycle rejection, reorder | `feat:` |
| [S1-10](#s1-10--feat-the-status-cascade-plan-including-parent--done) | The status cascade plan, incl. parent → Done | `feat:` |
| [S1-11](#s1-11--feat-rrule-occurrence-expansion--library-decision) | RRULE occurrence expansion — library decision | `feat:` |
| [S1-12](#s1-12--feat-habit-streaks-d5) | Habit streaks (D5) | `feat:` |
| [S1-13](#s1-13--feat-node-repository) | Node repository | `feat:` |
| [S1-14](#s1-14--feat-tag-repository-and-node_tags) | Tag repository and `node_tags` | `feat:` |
| [S1-15](#s1-15--feat-time-entry-repository) | Time entry repository | `feat:` |
| [S1-16](#s1-16--feat-habit-check-and-attachment-repositories) | Habit check and attachment repositories | `feat:` |
| [S1-17](#s1-17--feat-migration-0003-and-the-search-backend) | Migration `0003` and the search backend | `feat:` |
| [S1-18](#s1-18--feat-taskservice-writes--create-move-reorder-archive-restore) | TaskService writes | `feat:` |
| [S1-19](#s1-19--feat-timerservice-and-the-single-active-invariant) | TimerService and the single-active invariant | `feat:` |
| [S1-20](#s1-20--feat-habitservice--checks-and-streaks) | HabitService — checks and streaks | `feat:` |
| [S1-21](#s1-21--feat-taskservice-reads--board-and-tree-assembly) | TaskService reads — board and tree assembly | `feat:` |
| [S1-22](#s1-22--feat-searchservice) | SearchService | `feat:` |

**Sequencing:** S1-01 … S1-04 are housekeeping and de-risking and come first. S1-05
lays the schema. S1-06 … S1-12 are pure `domain` and touch no database at all.
S1-13 … S1-17 are `store`. S1-18 … S1-22 are `service` and compose the two. Nothing
later in the list is needed by anything earlier.

---

## S1-01 — chore: refresh the stale forward-looking doc comments in `internal/store`

The Stage 0 doc comments describe a future that has arrived. `internal/store/doc.go`
still says the connection and migration runner "arrive in S0-08 and the settings
repository in S0-09" — all three landed. `migrate.go`'s `//go:embed` rationale
explains that the directory rather than the `migrations/*.sql` glob is embedded
because "the migration set is legitimately empty at this point in Stage 0", which
stopped being true the moment `0001_settings.sql` landed in S0-09.

A comment that describes a plan instead of the code is worse than no comment: the next
reader trusts it. This is the first ticket because S1-05 is about to add more
migrations and would otherwise inherit and entrench the wrong explanation.

**Scope (may touch):** `internal/store/doc.go`, `internal/store/migrate.go`
(**comments only**), `internal/store/migrations/README.md`. **No behaviour change of
any kind** — no statement added, removed or reordered.

Requirements:

- `doc.go` describes what the package **contains now** (connection, migration runner,
  settings repository) and what is coming in this stage without ticket numbers that
  will go stale again — say "the node, tag, time-entry and habit repositories" rather
  than "S1-13".
- `migrate.go`: keep the `//go:embed migrations` decision, **replace the reason**. The
  honest reason now is that embedding the directory keeps non-`.sql` files (the
  `README.md`) out of the migration set via the glob in `loadMigrations`, and that
  `//go:embed migrations/*.sql` would also work today but would break the moment the
  set were emptied. State what is true, not what was true.
- The `# Imports` and `# No cgo` sections of `doc.go` are still accurate — leave them.

**Acceptance criteria**
- [ ] `git show --stat` lists only the files in Scope above.
- [ ] **No behaviour changed.** Every added line in the two `.go` files is a comment:
      `git show -U0 -- internal/store/doc.go internal/store/migrate.go | grep -E '^\+'
      | grep -vE '^\+\+\+|^\+\s*(//|$)'` prints nothing.
- [ ] Coverage of `internal/store` is unchanged from before the commit (a
      comments-only diff cannot move it).
- [ ] No occurrence of `S0-08`, `S0-09` or "Stage 0 delivers only" remains in
      `internal/store/`.
- [ ] `go test` and `go vet` over `./...` **excluding `node_modules`** both pass.
- [ ] `make check` green.

**Commit:** `chore(store): refresh the stale package and migration doc comments (S1-01)`

---

## S1-02 — fix: window background alpha is 0–255, not 0–1

`main.go` passes `BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1}`. Every
field of `options.RGBA` is a `uint8` in `0..255`, so `A: 1` asks for **~0.4% opacity**,
not "opaque". The intent was plainly the design's dark background at full alpha.

**Read `PLAN.md` K1 before starting.** On this machine the fix **cannot be verified by
eye**: `LC_NUMERIC=ru_RU.UTF-8` makes Wails emit `rgba(27, 38, 54, 0,0)` from
`window.c:205` and GTK discards the whole declaration, so the background colour never
reaches the window whatever alpha we pass. K1 is an upstream bug, it is **recorded and
deferred to Stage 2**, and **this ticket must not attempt to work around it** — no
`setlocale`, no `os.Setenv("LC_NUMERIC", ...)`, no wrapper. Fix the wrong constant and
stop.

**Scope (may touch):** `main.go` — the one literal, plus a short comment. Nothing else.

**Acceptance criteria**
- [ ] `main.go` reads `A: 255`.
- [ ] A comment states that alpha is 0–255 and points at K1 in `PLAN.md` as the reason
      the value is currently unobservable on this machine.
- [ ] `git show --stat` lists `main.go` only.
- [ ] `make check` green, including gate 5.
- [ ] No locale manipulation anywhere in the diff:
      `git show | grep -iE 'setlocale|LC_NUMERIC|LC_ALL'` prints nothing.

**Commit:** `fix(app): set the window background alpha to 255 (S1-02)`

---

## S1-03 — build: `make cover` — the ≥90% coverage measurement

Stage 1's ACCEPT is a number. A number that nobody can reproduce on demand is not an
acceptance criterion, so it gets a target — **before** there is any Stage 1 code to
measure, so that it can never be tuned to fit what was written.

**Scope (may touch):** `Makefile`, `.gitignore`.

Requirements:

- A `cover` target that runs, per package and separately:

  ```sh
  go test -covermode=atomic -coverprofile=coverage.domain.out  ./internal/domain/...
  go tool cover -func=coverage.domain.out  | tail -1
  go test -covermode=atomic -coverprofile=coverage.service.out ./internal/service/...
  go tool cover -func=coverage.service.out | tail -1
  ```

- It **prints both totals** and **exits non-zero if either is below `90.0`**, naming
  which package failed and by how much. A target that prints a number and exits 0 is a
  report, not a gate.
- The threshold is one variable (`COVER_MIN ?= 90.0`) so the Reviewer can see it at a
  glance and nobody has to grep for `90` in a shell one-liner.
- The comparison must not depend on the locale's decimal separator — this machine runs
  `LC_NUMERIC=ru_RU.UTF-8` (see **K1**) and `awk` will happily read `77.5` as `77`
  under it. Force `LC_ALL=C` for the comparison. **This is a `make cover`-internal
  measurement concern only and has nothing to do with K1's deferral**, which is about
  the GTK window.
- `cover` is **not** wired into `check`. The five gates stay five. Say so in a comment.
- `.gitignore` gains `coverage*.out`; a coverage run must leave
  `git status --porcelain` empty.
- `help` lists the new target.

**Acceptance criteria**
- [ ] `make cover` prints a total for `internal/domain` and one for `internal/service`
      and exits 0 on today's tree (both are at 100.0%).
- [ ] Temporarily lowering `COVER_MIN` proves nothing; instead demonstrate the failure
      path by temporarily raising it to `100.1`, showing a **non-zero exit** naming
      both packages, then revert.
- [ ] The failure message names the package, its total and the threshold.
- [ ] `make check` still runs exactly five gates and does **not** run `cover`:
      `make -n check` mentions no `-coverprofile`.
- [ ] `git status --porcelain` is empty after `make cover`.
- [ ] `LC_NUMERIC=ru_RU.UTF-8 make cover` and `LC_ALL=C make cover` agree.
- [ ] `make check` green.

**Commit:** `build: add make cover enforcing 90% on domain and service (S1-03)`

---

## S1-04 — feat: FTS5 availability spike — the search-backend decision point

**This is a de-risking ticket and it comes before any search code.** `PLAN.md` §7 FTS5
says: spike it, do not guess. `modernc.org/sqlite v1.53.0` may or may not have been
built with `SQLITE_ENABLE_FTS5`. Guessing wrong costs a schema migration and a rewrite
of the search repository; finding out costs one ticket.

**Scope (may touch):** `internal/store/fts5.go` (new), `internal/store/fts5_test.go`
(new). **No migration, no search API, no schema change** — those are S1-17.

The spike does not leave a throwaway behind. It leaves a **tested capability probe**:

- `func FTS5Available(ctx context.Context, db *sql.DB) (bool, error)` which
  **actually exercises FTS5**, in a scratch/temporary table it drops afterwards:
  1. `CREATE VIRTUAL TABLE ... USING fts5(title, description_md)`,
  2. insert a row,
  3. run a **real `MATCH` query** and check it returns that row,
  4. drop the table.
- Only a genuine "no such module: fts5" style failure means "unavailable" → `false,
  nil`. Any other error is an error, returned as one. A probe that reports `false` for
  a permissions problem would silently downgrade search forever.
- It must be safe to call on the live database.

### The decision point

The Dev **reports which path was taken** in the commit body, in this exact shape:

```
FTS5: AVAILABLE      (or: UNAVAILABLE)
Driver: modernc.org/sqlite v1.53.0
Probe: CREATE VIRTUAL TABLE ... USING fts5 + MATCH query -> <what happened>
Consequence for S1-17: <FTS5 virtual table + triggers | LIKE fallback>
```

and repeats it as a doc comment on `FTS5Available`. **S1-17 reads that decision and
does not re-litigate it.**

- **If AVAILABLE:** S1-17 builds an `fts5` external-content virtual table over
  `nodes(title, description_md)` with triggers keeping it in sync.
- **If UNAVAILABLE:** S1-17 implements `LIKE`-based search over `title` and
  `description_md`, case-insensitively, and that is a **pre-approved** outcome — not a
  failure, not a thing to escalate, and **not** a reason to reach for a cgo driver.

**cgo is not on the table in either branch.** Adding `github.com/mattn/go-sqlite3`, or
any build that needs a C toolchain, is an automatic reject. The no-cgo guarantee does
not bend; the feature degrades.

**Acceptance criteria**
- [ ] `FTS5Available` exists, is documented, and returns `(bool, error)` with the
      false-vs-error distinction described above.
- [ ] A test opens a temp DB (`t.TempDir()`), calls the probe, and asserts the result
      is **consistent across two calls** and that the scratch table is gone afterwards
      (`sqlite_master` has no row for it).
- [ ] A test asserts the probe leaves no trace in `schema_migrations` and does not
      disturb an existing schema.
- [ ] Whichever branch is real, a test **asserts the actual observed behaviour** — if
      FTS5 works, a `MATCH` query returns the expected row; if it does not, the error
      is the module-missing one and the probe returns `false, nil`.
- [ ] The commit body contains the four-line decision block verbatim.
- [ ] `CGO_ENABLED=0 go build ./...` succeeds and
      `go list -deps ./... | grep -i mattn` prints nothing.
- [ ] No new dependency was added.
- [ ] `make check` green.

**Commit:** `feat(store): probe fts5 availability on the pure-go driver (S1-04)`

---

## S1-05 — feat: migration `0002` — the core schema

The whole of `PLAN.md` §4 as SQL, in one append-only migration. Everything after this
ticket reads and writes these tables.

**Scope (may touch):** `internal/store/migrations/0002_core_schema.sql` (new),
`internal/store/migrate_test.go` (extend), `internal/store/schema_test.go` (new).
**`0001_settings.sql` is immutable — do not touch it.** No repository code here.

The schema, exactly per **`PLAN.md` §4** and **D1** / **D4**:

```
nodes(id TEXT PRIMARY KEY,               -- uuid
      parent_id TEXT NULL REFERENCES nodes(id) ON DELETE CASCADE,
      type TEXT NOT NULL,                -- task|project|habit|note|bug
      title TEXT NOT NULL,
      description_md TEXT NOT NULL DEFAULT '',
      status TEXT NOT NULL,              -- backlog|week|today|doing|done
      due TEXT NULL,                     -- date, YYYY-MM-DD
      due_source TEXT NOT NULL DEFAULT 'manual',   -- manual|auto   (D1)
      priority INTEGER NOT NULL DEFAULT 4,          -- 1..4
      estimate_min INTEGER NULL,
      recurrence TEXT NULL,              -- RRULE
      activity TEXT NULL,                -- the seven values of D4
      sort_order INTEGER NOT NULL DEFAULT 0,
      created_at TEXT NOT NULL,
      updated_at TEXT NOT NULL,
      completed_at TEXT NULL,
      archived_at TEXT NULL)

tags(id TEXT PRIMARY KEY, name TEXT NOT NULL UNIQUE, color TEXT NOT NULL DEFAULT '')
node_tags(node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
          tag_id  TEXT NOT NULL REFERENCES tags(id)  ON DELETE CASCADE,
          PRIMARY KEY (node_id, tag_id))
time_entries(id TEXT PRIMARY KEY,
             node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
             started_at TEXT NOT NULL,
             ended_at TEXT NULL)
attachments(id TEXT PRIMARY KEY,
            node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
            path TEXT NOT NULL, mime TEXT NOT NULL DEFAULT '')
habit_checks(node_id TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
             date TEXT NOT NULL,
             PRIMARY KEY (node_id, date))
```

Requirements and the reasoning that must show up as SQL comments:

- **`CHECK` constraints on every enumerated column**: `type`, `status`, `due_source`,
  `priority BETWEEN 1 AND 4`, and `activity` restricted to the seven **D4** values or
  NULL. The database is the last line of defence; a bug that writes `status='Done'`
  should fail loudly rather than become a row nobody can explain later.
- **Timestamps are TEXT in ISO-8601 UTC** (`YYYY-MM-DDTHH:MM:SSZ`), dates are TEXT
  `YYYY-MM-DD`. Pick this, comment it, and be consistent — SQLite has no date type and
  mixing representations is how comparisons start lying. `habit_checks.date` and
  `nodes.due` are **dates**; everything else is a timestamp.
- **Indexes** on `nodes(parent_id)`, `nodes(status)`, `nodes(due)`,
  `nodes(archived_at)`, `time_entries(node_id)` and a **partial unique index
  enforcing at most one open entry globally**:
  `CREATE UNIQUE INDEX one_open_timer ON time_entries((ended_at IS NULL)) WHERE ended_at IS NULL;`
  or an equivalent the Dev proves works. **The single-active invariant is a database
  constraint first and a service rule second** (S1-19) — two layers, because a race
  that the service loses must not be able to corrupt the data.
- `parent_id` is self-referential with `ON DELETE CASCADE`; the `foreign_keys` pragma
  is already on, so this is live.
- No triggers, no FTS table here — S1-17.

**Acceptance criteria**
- [ ] `0001_settings.sql` is byte-identical: `git diff --exit-code` names it never, and
      `git show --stat` does not list it.
- [ ] Migrating a fresh temp DB applies **two** migrations; migrating again applies
      **zero**.
- [ ] A DB already at `0001` (seed it, then migrate) applies exactly **one**.
- [ ] Table-driven test: every `CHECK` constraint rejects a bad value — one subtest per
      column (`type`, `status`, `due_source`, `priority` 0 and 5, `activity`), each
      asserting the insert **fails**.
- [ ] Test: deleting a parent cascades to children, `node_tags`, `time_entries`,
      `attachments` and `habit_checks`.
- [ ] Test: a second open `time_entry` (with `ended_at IS NULL`) is **rejected by the
      database**, while two *closed* entries are fine.
- [ ] Test: `habit_checks` rejects a duplicate `(node_id, date)`.
- [ ] Every index named in the ticket exists in `sqlite_master`.
- [ ] Tests use `t.TempDir()`; `git status --porcelain` empty afterwards.
- [ ] `make check` green.

**Commit:** `feat(store): add the core schema migration for nodes, tags and entries (S1-05)`

---

## S1-06 — feat: `Node`, `Tag`, `TimeEntry` and `HabitCheck` types

The pure types every later ticket speaks in. **`internal/domain` only — no SQL, no
`os`, no `time.Now()`.**

**Scope (may touch):** `internal/domain/node.go`, `internal/domain/time_entry.go`,
`internal/domain/habit.go`, `internal/domain/doc.go` (refresh the "Stage 0 delivers
only the enumerations" line), and their `_test.go` files.

Requirements:

- `Node` mirrors the `0002` columns using Go types that make illegal states hard:
  `ParentID *string`, `Due *Date`, `CompletedAt`/`ArchivedAt` `*time.Time`,
  `Type NodeType`, `Status Status`, `DueSource DueSource`, `Priority Priority`,
  `Activity *Activity`.
- A `Date` type (or a documented `time.Time`-at-midnight-UTC convention) for the
  **date-not-timestamp** columns. Whatever is chosen, `due`'s comparison semantics must
  be unambiguous, because "overdue" (**D1**) compares a date to today and an
  off-by-one-timezone here is a bug the user sees every morning.
- `Node.IsLeaf(children []Node) bool` — true when there are no children **or no child's
  type has a Kanban column** (`!child.Type.HasColumn()` — `note` **and** `habit`;
  **D2** as generalised by **D10**, see the amendment in the criteria below). This
  predicate is load-bearing for status derivation,
  progress, cascades and "only leaves start a timer"; it lives here, once.
- `DefaultActivity(NodeType) *Activity` per **D4**: task→Разработка, bug→Тестирование,
  project→Управление проектом, note→Документация, habit→none.
- `Validate() error` on each type, returning a domain error naming the offending field.
- `TimeEntry` with `EndedAt *time.Time` and `IsOpen() bool`; `Duration(now)` taking the
  injected clock for the open case — **never** `time.Now()`.
- `HabitCheck{NodeID string; Date Date}`.

**Acceptance criteria**
- [ ] Table-driven `Validate` tests: at least one failing case per validated field, and
      the error names the field.
- [ ] `IsLeaf` subtests: no children; one `task` child; one `note` child; mixed
      `note` + `task`; several notes only. **The all-notes case returns true** — it is
      the case D2 calls out and the one that will otherwise be got wrong.
      **Amended by D10:** the predicate is `!child.Type.HasColumn()`, not "is it a
      note", so the **all-habits** and **mixed note+habit** cases return true as well.
- [ ] `DefaultActivity` covers all five node types, `habit` returning nil.
- [ ] `TimeEntry.Duration` is tested with a fixed injected clock for an open entry and
      with `EndedAt` for a closed one.
- [ ] `TestDomainIsPure` and `TestDomainReadsNoClock` still pass, unmodified.
- [ ] `make cover` green (`internal/domain` ≥90%).
- [ ] `make check` green.

**Commit:** `feat(domain): add node, tag, time entry and habit check types (S1-06)`

---

## S1-07 — feat: derived status and progress (D2, D7, D9)

**The central rule of the product.** Get this wrong and every column, every progress
bar and every drag is wrong with it.

**Scope (may touch):** `internal/domain/derive.go`, `internal/domain/derive_test.go`.
Pure — the functions take an already-loaded set of nodes, not a database.

Requirements — from **D2**, **D7** and **D9**:

- `DeriveStatus`: a parent's status is the **least-advanced status among its non-done
  children**, and `done` **only when every child is done**. `note` children are
  **excluded entirely** — from the "least advanced" scan and from the "all done" test.
- Status ordering, least to most advanced, is the existing declaration order:
  `backlog < week < today < doing < done`.
- A node that **is a leaf** (no children, or all children are notes — `Node.IsLeaf`
  from S1-06) reports its **own stored status**. Derivation is for parents only.
- Derivation is **recursive**: a grandparent derives from its children's *derived*
  statuses, not their stored ones.
- **Parents never derive `doing`** as a stored fact — but a parent whose least-advanced
  non-done child is `doing` does *render* in Doing. Spell out which of these the
  function returns and make a test say so, because "parents never enter Doing" (D2)
  is about **starting a timer**, not about rendering.
- `Progress`: **done leaves / total leaves**, `note` leaves **excluded from the
  denominator** (**D7**). Recursive over the subtree, counting leaves only — an
  intermediate parent is not a unit of work, its leaves are.
- **A `project` is not a unit of work** and is therefore **never counted in the
  progress denominator**, even when it is empty and would otherwise qualify as a leaf
  by shape (**D9**). A project is the thing the bar is drawn *for*, not a thing the bar
  counts.
- **Zero non-note leaves.** This was undefined in `PLAN.md` when the ticket was written
  and the ticket decided it; **D9 now records the same answer**:
  `Progress` returns `done=0, total=0` and a **`Defined() bool` that is false**. It
  does **not** return 0% and it does **not** return 100%. A project containing only
  notes has no work in it, and both 0% ("nothing done") and 100% ("all done") are
  lies the UI would render as a bar. The caller renders no bar. Document this on the
  function and test it.

**Acceptance criteria**
- [ ] Table-driven, each case a small literal tree with the expected derived status:
      - [ ] all children `backlog` → `backlog`
      - [ ] children `backlog` + `today` → `backlog` (least advanced wins)
      - [ ] children `done` + `week` → `week` (done children are skipped)
      - [ ] **all** children `done` → `done`
      - [ ] children `done` + `note` → `done` (the note does not hold it back)
      - [ ] only `note` children → **leaf**, reports its own stored status
      - [ ] no children → leaf, reports its own stored status
      - [ ] three levels deep, the grandparent deriving from derived values
- [ ] `Progress` subtests: 0 of 3; 2 of 3; 3 of 3; notes present and excluded from both
      numerator and denominator; nested subtree; **zero non-note leaves →
      `Defined() == false`**; **an empty `project` → `Defined() == false`** (D9).
- [ ] No `time.Now()`, no I/O; purity tests still pass.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): derive parent status and subtree progress (S1-07)`

### Amended after the second review — D10 and D11

The ticket as written above shipped, and the second review found the rule it encodes
was too narrow in two places. Both amendments were confirmed by the user as decisions
and are recorded in full in [`PLAN.md` §7](./PLAN.md) — **D10** and **D11**. They
supersede the corresponding lines above; the follow-up commits are `81fb6e4` and
`d5a170b`.

- **D10 supersedes "`note` children are excluded"** in both the `DeriveStatus` and the
  `Progress` requirements. The exclusion is **every type with no Kanban column**, i.e.
  `!Type.HasColumn()` — `note` **and** `habit`. `Node.IsLeaf` follows the same
  predicate, so a node whose children all lack a column is a leaf reporting its own
  stored status, and `walkProgress` cuts at the same predicate. `project{task:done,
  habit}` used to derive `backlog` behind a 100% bar; it derives `done` now. The same
  fix closed a pre-existing bug where a task whose only children were habits derived
  `done` while stored at `backlog`. `PLAN.md` §4's derivation rule has been reworded to
  match.
- **D11 supersedes "a `project` is … never counted in the progress denominator"** for
  the leaf case only. A **leaf** project — nothing under it has a column — is **one
  work leaf in its parent's denominator**, done iff its own stored status is `done`.
  `project{empty sub-project, task:done}` read 100% behind a `backlog` column; it reads
  **1 of 2** now. **The "zero non-note leaves" requirement above is unchanged** — read
  it as *zero work leaves*, the D10 predicate: asked
  about *itself*, a project with no work beneath it still reports
  `Defined() == false`, neither 0% nor 100%, and the caller still draws no bar. A
  project **with** children that have columns is still not a unit itself. Derivation
  was already correct and did not change.

**Stage 2 must not re-derive either of these.** See also known issues **K2** and **K3**
in `PLAN.md`, which this pair of decisions created or exposed and which are **carried
into Stage 2** as C4 and C5. **K4** — the fourth one this pair exposed — is **not**
deferred: it was a real defect and is **resolved** in `9664506`, which refuses every
no-column type as a parent.

---

## S1-08 — feat: column↔due rules and overdue (D1, D8)

**Scope (may touch):** `internal/domain/due.go`, `internal/domain/due_test.go`.

Requirements — **D1** and **D8**, exactly:

- Moving to **`today`**: `due = today`, `due_source = 'auto'`. Per **D8** this
  **overwrites any existing due date, including a manual one** — the column move
  always wins, unconditionally.
- Moving to **`week`**: `due = the upcoming Friday`, `due_source = 'auto'`; **if today
  is Friday, `due = today`**. Today is 2026-09-18, which *is* a Friday — the edge case
  is live on day one and gets its own subtest.
- Moving to **`backlog`**: clears `due` **only if `due_source == 'auto'`**. A manual
  date survives, and `due_source` stays `manual`.
- Moving to `doing` or `done` does **not** touch `due` or `due_source`.
- **Any user edit of `due` sets `due_source = 'manual'`** — including setting it to the
  same value it already had, and including clearing it. This is a separate exported
  function from the column move; it must be impossible to apply one while meaning the
  other.
- `IsOverdue(node, today) bool` = `due != nil && due < today && status != done`.
  Equal to today is **not** overdue.
- `today` and "the upcoming Friday" are computed from the **injected**
  `now func() time.Time`, never from `time.Now()`, and in a documented, explicit
  location (the local zone — a task due "today" means the user's today).

**Acceptance criteria**
- [ ] Table-driven over **every** weekday: → `week` from Mon…Sun with the expected
      Friday, and **Friday → today itself**.
- [ ] → `today` sets both the date and `due_source='auto'`.
- [ ] `due_source` transition subtests, all four:
      - [ ] `auto` → backlog clears `due`
      - [ ] `manual` → backlog **keeps** `due`
      - [ ] column move (today/week) **overwrites** a `manual` due and flips it to
            `auto` — this is **D8**, now recorded in `PLAN.md` §7 and no longer a
            reading the Dev has to justify in a comment.
      - [ ] user edit of `due` sets `manual`, including the clear-to-nil case
- [ ] `doing` / `done` moves leave `due` and `due_source` untouched.
- [ ] `IsOverdue`: yesterday+`today` status → true; **today** → false; tomorrow →
      false; yesterday + `done` → false; nil due → false.
- [ ] A month-end and a year-end case for "upcoming Friday" (e.g. 2026-12-31).
- [ ] Purity tests still pass.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): add the column-to-due rules and overdue (S1-08)`

---

## S1-09 — feat: tree operations — move, cycle rejection, reorder

**Scope (may touch):** `internal/domain/tree.go`, `internal/domain/tree_test.go`.
Pure: these functions take a loaded node set and return a **plan** — which rows change
and how. They do not write anything; S1-13 and S1-18 do the writing.

Requirements:

- `Children`, `Descendants`, `Ancestors`, `Subtree` over a loaded node set.
- `ValidateMove(nodes, nodeID, newParentID) error` rejecting:
  - **moving a node into its own subtree** — the **circular parent** case named in the
    brief. Both the direct case (`parent = itself`) and the indirect one (`parent = a
    grandchild`), at depth ≥ 3.
  - a `newParentID` that does not exist,
  - a `nodeID` that does not exist,
  - **any parent whose type has no Kanban column** — `note` **and** `habit` — with
    `ErrTypeHasNoChildren`, through `NodeType.HasColumn`. Refusing only a `note` here
    was the third review's blocking defect (see the Stage 1 history above and **K4**,
    now resolved): a habit with children rendered in a Kanban column and corrupted its
    ancestors' derived status and progress. The sentinel has **no alias** — the former
    `ErrNoteParent` name is gone, because an alias is a second spelling of the rule.
- **Moving a node moves its whole subtree**: children keep their own `parent_id`,
  their own status, and come along. The plan changes exactly one `parent_id` — the
  moved node's — and the `sort_order` of the affected sibling ranges. Assert that
  descendants' statuses are untouched.
- `Reorder(siblings, nodeID, toIndex)` producing a **dense, gapless `sort_order`
  sequence** for the affected sibling set. Reordering is deterministic and stable for
  everything not moved.
- `Archive` / `Restore` plans: archiving sets `archived_at` on the node **and its whole
  subtree**; restoring clears it on the same set. A child archived on its own, then
  restored via its parent — decide the rule, document it, test it.
- Detecting a cycle must not hang on a graph that is **already** cyclic (corrupt data):
  bound the walk by the node count and return an error rather than looping.

**Acceptance criteria**
- [ ] **`ValidateMove` rejects a node moved into its own subtree** — subtests for:
      into itself; into its direct child; into a grandchild (depth 3); into a
      great-grandchild. Each asserts a **specific, matchable** error
      (`errors.Is(err, domain.ErrCircularParent)`).
- [ ] `ValidateMove` **accepts** a move to a sibling, to the root (`nil` parent), and
      to an unrelated subtree.
- [ ] `ValidateMove` **rejects a `note` parent and a `habit` parent alike**, each
      asserting `errors.Is(err, domain.ErrTypeHasNoChildren)`, and `deriveStatus`
      returns `backlog` for a no-column node even when handed children (the self-cut,
      so the invariant does not rest on the validator alone).
- [ ] Moving a subtree: descendants' `parent_id` and `status` are unchanged; only the
      moved node's `parent_id` changes.
- [ ] `Reorder` yields a gapless `0..n-1` sequence; moving to the first, last and
      middle positions each get a subtest; reordering a single-element set is a no-op.
- [ ] Archive/restore cover the whole subtree, and the documented nested rule has a
      test.
- [ ] A deliberately cyclic input terminates with an error rather than hanging — give
      the subtest a short `t.Deadline`-friendly shape, no `time.Sleep`.
- [ ] Purity tests still pass.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): add tree move, reorder and archive planning (S1-09)`

---

## S1-10 — feat: the status cascade plan, including parent → Done

**D2's headline rule, and the one the brief calls out by name.** The old "reject a
parent's move to Done" behaviour was **REPLACED**. There is no rejection. There is a
cascade.

**Scope (may touch):** `internal/domain/cascade.go`,
`internal/domain/cascade_test.go`.

Requirements:

- `PlanCascade(nodes, rootID, target Status, now) []StatusChange` — dragging a node to
  column `target` sets `status = target` on **every descendant that is not already
  `done` and is not a `note`**, and on the node itself when it is a leaf.
- Moving to **`done`** additionally sets **`completed_at`** on every node it changes,
  from the injected `now`. Moving *out* of `done` clears `completed_at` — decide,
  document, test.
- `note` descendants are **never** touched. `done` descendants are **never** touched —
  a finished task does not get un-finished by dragging its parent to Today.
- After the cascade, the parent **derives** its status via S1-07's `DeriveStatus`; the
  plan therefore contains **no row for the parent itself** when the parent is not a
  leaf. Assert this explicitly — writing a status onto a parent is exactly the drift
  D2 exists to prevent.
- The cascade never puts a **parent** into `doing`; only leaves. Dragging a parent to
  Doing cascades `doing` onto its non-done, non-note **leaf** descendants, and the
  parent derives Doing from them. Intermediate parents get no stored status.
  **Amended by D9/D10:** "non-note" is really `Type.HasColumn()` — every type with no
  Kanban column is skipped as a descendant, and `project` descendants are skipped when
  the cascaded status is `doing`. Read every "note" and "non-note" in this ticket —
  **above and below this line** — the same way: the predicate is
  `Type.HasColumn()`, so `habit` descendants and all-`habit` children behave exactly
  as notes do here. Implemented as `note`-only, this ticket reproduces the pre-D10
  defect.
- The plan is a **plan**: a deterministic, ordered slice the service applies in one
  transaction (S1-18). Pure, no I/O, injected clock.

**Acceptance criteria**
- [ ] **The headline test:** parent with three unfinished descendants at two depths,
      one already `done`, one `note`, dragged to `done` →
      - [ ] every unfinished non-note descendant becomes `done`,
      - [ ] each gets `completed_at` set to the injected `now`,
      - [ ] the already-`done` one is untouched (its original `completed_at` survives),
      - [ ] the `note` is untouched,
      - [ ] feeding the result back into `DeriveStatus` makes the parent derive
            **`done`** — assert the derivation, not just the writes.
- [ ] Dragging a parent to `today` cascades `today` and sets **no** `completed_at`.
- [ ] Dragging to `doing` touches only leaves; no intermediate parent gets a stored
      status.
- [ ] Moving a subtree out of `done` clears `completed_at` per the documented rule.
- [ ] A leaf dragged anywhere produces exactly one change: itself.
- [ ] A node whose children are all notes is treated as a **leaf**.
- [ ] The plan is deterministic — same input, same order, twice.
- [ ] Purity tests still pass.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): cascade a parent's column move onto its descendants (S1-10)`

---

## S1-11 — feat: RRULE occurrence expansion — library decision

**Streaks need real RRULE expansion** (**D5**). Hand-rolling "every RRULE" is a
multi-week rabbit hole; pretending `FREQ=WEEKLY` is "+7 days" breaks on
`BYDAY=MO,WE,FR`, which is exactly the habit people actually create.

**Scope (may touch):** `internal/domain/recurrence.go`,
`internal/domain/recurrence_test.go`, `go.mod`, `go.sum`.

### The decision the Dev must make and justify

Pick **one**, and justify it in the commit body:

1. **A maintained pure-Go RRULE library.** It must be: actively maintained, pure Go
   (**no cgo, no C, no subprocess**), permissively licensed, and small enough to read.
   The commit body records the module path, version, licence, last release date, and
   the RFC 5545 subset it covers.
2. **A hand-rolled, explicitly restricted subset.** Allowed only with a written
   justification and an explicit list of what is supported. The **minimum** supported
   subset is `FREQ=DAILY`, `FREQ=WEEKLY` (with `BYDAY`), `FREQ=MONTHLY` (with
   `BYMONTHDAY`) and `INTERVAL`. Anything outside the supported set must be **rejected
   at parse time with a clear error**, never silently mis-expanded — a habit that
   silently expands wrong produces a plausible, wrong streak, and nobody will catch it.

Either way:

- `CGO_ENABLED=0 go build ./...` still succeeds; `go list -deps ./... | grep -i mattn`
  still prints nothing; no new transitive cgo.
- **`internal/domain` stays pure.** If the chosen library reads the clock or the
  filesystem, it is wrapped so that domain's entry point still takes
  `now func() time.Time` and nothing else. If it cannot be wrapped that way, it is the
  wrong library.
- The API this ticket exposes is small and is what S1-12 consumes:
  `Occurrences(rrule string, dtstart Date, from, to Date) ([]Date, error)` and
  `PreviousOccurrence` / `NextOccurrence` helpers.
- Timezone handling is stated explicitly. Habits are **date**-granular (`habit_checks`
  is keyed by date), so expansion works in dates, not instants.

**Acceptance criteria**
- [ ] The commit body names the decision, the library (path, version, licence,
      maintenance status) **or** the hand-rolled subset, and why.
- [ ] Table-driven expansion tests: `FREQ=DAILY`; `FREQ=WEEKLY` (plain);
      `FREQ=WEEKLY;BYDAY=MO,WE,FR`; `FREQ=WEEKLY;INTERVAL=2`;
      `FREQ=MONTHLY;BYMONTHDAY=1`; and a **month-end / leap-year** case.
- [ ] An unparseable or unsupported RRULE returns a **matchable error**, never an empty
      slice and never a silently wrong expansion.
- [ ] Expansion is deterministic and bounded — an open-ended rule over a bounded
      window returns a bounded result, and a test proves a large window terminates.
- [ ] `go.mod` gains at most one direct dependency; `go mod tidy` leaves the tree clean.
- [ ] Purity tests still pass, `TestDomainReadsNoClock` included.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): expand rrule occurrences for habit scheduling (S1-11)`

---

## S1-12 — feat: habit streaks (D5)

**Scope (may touch):** `internal/domain/streak.go`, `internal/domain/streak_test.go`.

Requirements — **D5**, which is *not* the obvious implementation:

- A streak is the number of **consecutive scheduled occurrences of the node's RRULE
  that were checked** — **not** consecutive calendar days.
- A `FREQ=WEEKLY` habit checked four weeks running has **streak 4**, not 4 and not 28
  and certainly not 0 because six days were "missed" each week.
- The streak **breaks at the first scheduled occurrence that passed unchecked**,
  walking backwards from now.
- **Today's still-pending occurrence does not break the streak.** A habit due today and
  not yet checked keeps yesterday's streak; checking it increments.
- A check on a **non-scheduled** date: decide the rule, document it, test it. (It can
  exist — the user can check a habit on a whim — the question is whether it counts.)
- Signature takes the loaded checks, the RRULE, and the injected `now`. Pure.

**Acceptance criteria**
- [ ] `FREQ=WEEKLY` checked on four consecutive scheduled dates → **4**.
- [ ] The same habit with the third occurrence missed → streak counts only from the
      break forward, i.e. **2**, not 4 and not 3.
- [ ] `FREQ=DAILY` checked 5 days running → 5.
- [ ] **Today scheduled and unchecked** → the streak is unchanged (previous
      occurrences' count), explicitly **not** 0.
- [ ] Today scheduled and checked → previous + 1.
- [ ] No checks at all → 0.
- [ ] A habit with no RRULE → a matchable error (habits require `recurrence`).
- [ ] `FREQ=WEEKLY;BYDAY=MO,WE,FR` with a mid-week miss.
- [ ] A check on an unscheduled date behaves as documented, with a subtest naming the
      rule.
- [ ] Purity tests still pass.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(domain): count habit streaks over scheduled occurrences (S1-12)`

---

## S1-13 — feat: node repository

**Scope (may touch):** `internal/store/nodes.go`, `internal/store/nodes_test.go`.

**Thin.** This repository maps rows to `domain.Node` and back, and runs queries. It
computes **no** derived status, **no** progress, **no** due dates, and makes **no**
decision about cascades or cycles — all of that is `domain`, applied by `service`. A
`CASE WHEN` in here that reimplements a rule is a Stage 1 failure regardless of what
the coverage number says.

Requirements:

- `Create`, `Get`, `Update`, `Delete`.
- `ListChildren(parentID *string)`, `ListSubtree(rootID)` (recursive CTE),
  `ListRoots()`, `ListAll()` — each with an `includeArchived bool` or an explicit
  archived-excluding default that is **documented and tested**.
- `UpdateParentAndOrder`, `UpdateStatuses(changes)` — bulk, so a cascade plan is one
  round trip — `SetArchivedAt(ids, *time.Time)`.
- Everything that can be asked to run inside a caller's transaction **must** be: take a
  `*sql.Tx`-or-`*sql.DB` executor interface, because S1-18 applies a cascade plus a
  due-date write plus a reorder atomically.
- `ErrNotFound` (already in `store/errors.go`) for a missing row.
- Time and date columns round-trip through the `0002` text formats exactly.

**Acceptance criteria**
- [ ] Round-trip test: create → get → every field equal, including the nullables
      (`parent_id`, `due`, `estimate_min`, `recurrence`, `activity`, `completed_at`,
      `archived_at`) as both NULL and set.
- [ ] `ListSubtree` on a 3-level tree returns the root and all descendants, and
      **nothing** from a sibling subtree.
- [ ] Archived nodes are excluded by default and included on request — one subtest each.
- [ ] `UpdateStatuses` applies a multi-row change in one call and is atomic: a bad row
      in the batch leaves **none** of them applied.
- [ ] `Get` on an unknown id returns `ErrNotFound` via `errors.Is`.
- [ ] Writing an invalid enum value fails against the `0002` `CHECK` constraint (the
      DB is the backstop, and this test proves it is wired).
- [ ] Tests use `t.TempDir()`; `git status --porcelain` empty afterwards.
- [ ] No derived status, progress, due or cascade logic in this file — the Reviewer
      will grep for it.
- [ ] `make check` green.

**Commit:** `feat(store): add the node repository over the core schema (S1-13)`

---

## S1-14 — feat: tag repository and `node_tags`

**Scope (may touch):** `internal/store/tags.go`, `internal/store/tags_test.go`.

Requirements:

- `CreateTag`, `GetTagByName`, `ListTags`, `DeleteTag`.
- `AttachTag(nodeID, tagID)` / `DetachTag` / `TagsForNode(nodeID)` /
  `NodesForTag(tagID)`.
- Attaching a tag twice is a **no-op, not an error** (`ON CONFLICT DO NOTHING`) — the
  UI will do it, and an error there is noise.
- Tag names are unique; the uniqueness is the DB's, and the repository surfaces a
  matchable `ErrDuplicate` rather than a driver string.

**Acceptance criteria**
- [ ] Round-trip create → get by name → list.
- [ ] Duplicate name returns a matchable error, not a raw driver error.
- [ ] Double attach is a no-op and leaves exactly one row.
- [ ] Deleting a tag removes its `node_tags` rows and leaves the nodes alone.
- [ ] Deleting a node removes its `node_tags` rows (FK cascade) and leaves the tags.
- [ ] Tests use `t.TempDir()`; tree clean afterwards.
- [ ] `make check` green.

**Commit:** `feat(store): add the tag repository and node tag links (S1-14)`

---

## S1-15 — feat: time entry repository

**Scope (may touch):** `internal/store/time_entries.go`,
`internal/store/time_entries_test.go`.

Requirements:

- `Open(nodeID, startedAt)`, `Close(id, endedAt)`, `OpenEntry()` — the **single**
  globally open entry, or `ErrNotFound`.
- `ListByNode(nodeID)`, `ListByDay(date)` — the latter is what Stage 6's PMP timelog
  will read, so get the date-boundary semantics right and comment them (entries are
  bucketed by `started_at`'s local date; state it).
- `CloseAll(endedAt)` — closes whatever is open, used by the timer service and later by
  the sleep/lock handler.
- Executor interface, as S1-13, so the service can run open+close in one transaction.
- **The repository does not decide** when a timer may start — it enforces what the
  schema enforces and reports. The invariant's policy is S1-19.

**Acceptance criteria**
- [ ] Open → `OpenEntry` returns it → `Close` → `OpenEntry` returns `ErrNotFound`.
- [ ] **Opening a second entry while one is open fails** — the `0002` partial unique
      index rejects it. This is the database half of the single-active invariant and
      this test is what proves it exists.
- [ ] Two **closed** entries on the same node are fine.
- [ ] `Close` on an already-closed entry is a matchable error.
- [ ] `ListByDay` boundary subtests: an entry started at 00:00 and one at 23:59 of the
      same day both land in it; one from the previous day does not.
- [ ] Timestamps round-trip exactly, including UTC normalisation.
- [ ] Tests use `t.TempDir()`; tree clean afterwards.
- [ ] `make check` green.

**Commit:** `feat(store): add the time entry repository (S1-15)`

---

## S1-16 — feat: habit check and attachment repositories

Two small, unrelated tables, together because neither justifies a ticket alone and
neither depends on the other.

**Scope (may touch):** `internal/store/habit_checks.go`,
`internal/store/habit_checks_test.go`, `internal/store/attachments.go`,
`internal/store/attachments_test.go`.

Requirements:

- `Check(nodeID, date)` / `Uncheck(nodeID, date)` / `ChecksForNode(nodeID, from, to)` /
  `IsChecked(nodeID, date)`.
- Checking the same `(node, date)` twice is a **no-op, not an error**.
- `AddAttachment`, `ListAttachments(nodeID)`, `DeleteAttachment`. **The repository
  stores a path and a mime type and does not touch the filesystem** — copying files
  into the app data dir is Stage 3, and a repository that writes files is a repository
  that cannot be tested with a temp DB alone.

**Acceptance criteria**
- [ ] Check → `IsChecked` true → uncheck → false.
- [ ] Double check leaves exactly one row and returns no error.
- [ ] `ChecksForNode` respects an inclusive `from`/`to` range — boundary subtests on
      both ends.
- [ ] Deleting a node cascades away its checks and attachments.
- [ ] `internal/store/attachments.go` imports neither `os` nor `io/fs`.
- [ ] Tests use `t.TempDir()`; tree clean afterwards.
- [ ] `make check` green.

**Commit:** `feat(store): add the habit check and attachment repositories (S1-16)`

---

## S1-17 — feat: migration `0003` and the search backend

**Reads S1-04's decision and implements that branch. It does not re-run the spike and
does not re-open the question.**

**Scope (may touch):** `internal/store/migrations/0003_search.sql` (new — **only if the
FTS5 branch was taken**), `internal/store/search.go`, `internal/store/search_test.go`.
**`0001` and `0002` are immutable.**

Search covers **`title` + `description_md`**, and nothing else, in this stage.

### If S1-04 reported AVAILABLE

- `0003_search.sql` creates an `fts5` virtual table over `nodes(title,
  description_md)` — external-content, `content='nodes'`, `content_rowid` mapped to a
  stable integer — plus `AFTER INSERT` / `AFTER UPDATE` / `AFTER DELETE` triggers that
  keep it in sync. An index that can drift from its table is a bug generator; the
  triggers are the point.
- A **rebuild** path (`INSERT INTO nodes_fts(nodes_fts) VALUES('rebuild')`) for
  recovery, callable from code.
- Query with `MATCH`, ordered by `rank` (or `bm25()`), with user input **escaped** so
  that a query containing `"` or `*` or `AND` cannot become a syntax error thrown at
  the user or a query that means something else.

### If S1-04 reported UNAVAILABLE

- **No `0003` migration at all** — there is nothing to create, and an empty migration
  is a lie in the ledger.
- `LIKE`-based search over `title` and `description_md`, case-insensitive, with `%` and
  `_` **escaped** in the user's input via an explicit `ESCAPE` clause.
- Documented and commented as the **pre-approved fallback**, with a note that the
  `LIKE` path has no ranking and results are ordered deterministically
  (`updated_at DESC, id`).

### Either way

- **One exported API, identical in both branches**, so that S1-22 and Stage 3 do not
  care which one is underneath:
  `Search(ctx, query string, opts SearchOptions) ([]domain.Node, error)`.
- Archived nodes are excluded unless `opts` asks for them.
- The empty query returns an empty result, not every row.
- No cgo. Still.

**Acceptance criteria**
- [ ] The branch taken matches S1-04's recorded decision; the commit body says which
      and quotes the decision line.
- [ ] Finds a node by a word in `title`; finds one by a word in `description_md`.
- [ ] Updating a node's title makes the **old** word stop matching and the new one
      start (the sync test — the one that catches a missing trigger).
- [ ] Deleting a node removes it from results.
- [ ] Archived nodes excluded by default, included on request.
- [ ] Empty query → empty result.
- [ ] **Injection/escape subtests**: a query containing `%`, `_`, `"`, `*` and `'` is
      treated as literal text and neither errors nor matches everything.
- [ ] Unicode: a Russian-language title is found by a Russian query (the app is EN/RU).
- [ ] If FTS5: a fresh DB applies **three** migrations; re-migrating applies zero;
      `rebuild` restores the index after the virtual table is emptied.
- [ ] If LIKE: `internal/store/migrations/` contains exactly two `.sql` files.
- [ ] Tests use `t.TempDir()`; tree clean afterwards.
- [ ] `make check` green.

**Commit:** `feat(store): add search over node title and description (S1-17)`

---

## S1-18 — feat: TaskService writes — create, move, reorder, archive, restore

Where `domain`'s plans become database rows. **The service owns the transaction; the
domain owns the decision.** If this file contains an `if status ==` that is not simply
dispatching to a domain function, the rule has leaked out of `domain`.

**Scope (may touch):** `internal/service/task.go`, `internal/service/task_test.go`.

Requirements:

- `NewTaskService(nodes, tags *store.…Repo, clock Clock, newID func() string)` — the
  clock and the ID generator are **injected**, so every test is deterministic.
- `CreateNode` — generates the UUID, stamps `created_at`/`updated_at`, applies
  `DefaultActivity` by type (**D4**), appends at the end of its siblings'
  `sort_order`, and rejects an invalid node via `domain.Validate`.
- `MoveNode(nodeID, newParentID, toIndex)` — `domain.ValidateMove` first (**the
  circular-parent rejection surfaces here as an error, not a panic and not a silent
  no-op**), then the reorder plan, all in **one transaction**.
- `MoveToColumn(nodeID, status)` — the money method. In one transaction:
  1. `domain.PlanCascade` (**S1-10**) over the subtree,
  2. `domain.ApplyColumnDue` (**S1-08**) for the due/`due_source` write,
  3. bulk `UpdateStatuses`, `completed_at` where the cascade says so,
  4. `updated_at` on everything touched.
  Either all of it lands or none of it does.
- `SetDue(nodeID, *Date)` — the **user edit** path, therefore `due_source = 'manual'`
  (**D1**), and it must be a different method from the column move so the two can
  never be confused at the call site.
- `ArchiveNode` / `RestoreNode` over the whole subtree.
- Every method returns `(T, error)` — the Wails contract holds even though nothing is
  bound yet (**that is Stage 2**).

**Acceptance criteria**
- [ ] **Parent → Done, end to end against a real temp DB:** every unfinished descendant
      **whose type has a Kanban column** (`!Type.HasColumn()` is skipped — `note`
      **and** `habit`, per **D10**) is `done` in the database, each with `completed_at`
      set to the injected clock; the no-column descendants are untouched; reading the
      tree back derives `done` on the parent.
- [ ] **Circular parent is rejected**: moving a node under its own grandchild returns
      `domain.ErrCircularParent` via `errors.Is` **and the database is unchanged** —
      assert the rollback, not just the error.
- [ ] `MoveToColumn(today)` writes `due = today` and `due_source='auto'`;
      `MoveToColumn(backlog)` clears an `auto` due and keeps a `manual` one.
- [ ] `SetDue` always leaves `due_source='manual'`.
- [ ] A failure partway through a cascade rolls the **whole** move back (inject a
      failing executor or a constraint violation).
- [ ] `CreateNode` assigns `sort_order` after existing siblings and the D4 default
      activity per type.
- [ ] Archive/restore cover the subtree.
- [ ] All time-dependent assertions use a fixed `service.FixedClock`; no `time.Sleep`
      anywhere in the tests.
- [ ] `make cover` green (`internal/service` ≥90%).
- [ ] `make check` green.

**Commit:** `feat(service): add the task service write path (S1-18)`

---

## S1-19 — feat: TimerService and the single-active invariant

> **Exactly one open `time_entry`, globally.**

**Scope (may touch):** `internal/service/timer.go`, `internal/service/timer_test.go`.

Requirements:

- `Start(nodeID)`: in **one transaction**, close any open entry **on any node** at
  `now`, then open a new one on `nodeID`. The overlap window must not exist even
  momentarily.
- `Stop()`: close the open entry; stopping when nothing is running is a **no-op with no
  error** (the tray and a keyboard shortcut will both call it).
- `Current()`: the open entry and its elapsed time, or nil.
- **Only leaves start a timer** (**D2**, **D7**): starting on a node that is not a leaf
  — including any `project` — returns a matchable error. `Node.IsLeaf` from S1-06
  decides this; the service does not re-derive it.
- Starting on a node that is **already** running is a no-op, **not** a stop/start —
  restarting would silently fragment the log and lose the elapsed time the user is
  watching.
- Starting on an `archived` node, or a `done` one: decide, document, test.
- The **database** also enforces the invariant (the S1-05 partial unique index). The
  service is the policy; the index is the backstop. Both are tested, here and in S1-15.
- Sleep/lock handling is **Stage 4** — `PrepareForSleep`, `ActiveChanged` and
  "resume timer?" are out of scope. `CloseAll` exists (S1-15) and Stage 4 will call it.

**Acceptance criteria**
- [ ] **The overlapping-timer test:** start on A, then start on B → A's entry is closed
      with `ended_at == now`, B's is open, and **exactly one** row has
      `ended_at IS NULL` across the whole table. Assert that count with SQL.
- [ ] Start → stop → start on the same node produces **two** closed entries, not one
      long one.
- [ ] Stop with nothing running: no error, no rows changed.
- [ ] Start on a node with children **whose type has a Kanban column** → matchable
      error, **no** entry created (the predicate is `Node.IsLeaf`/`HasColumn`, **D10**).
- [ ] Start on a `project` → matchable error even when it has no children (**D7**:
      projects never enter Doing).
- [ ] Start on a node whose children **all have no Kanban column** — all notes, all
      habits, or a mix — **succeeds** (it is a leaf, **D10**).
- [ ] Start on an already-running node is a no-op: same entry id, same `started_at`.
- [ ] Elapsed time uses the injected clock; no `time.Sleep` in any test.
- [ ] A forced failure between close and open rolls back both.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(service): add the timer service with a single active entry (S1-19)`

---

## S1-20 — feat: HabitService — checks and streaks

**Scope (may touch):** `internal/service/habit.go`, `internal/service/habit_test.go`.

Requirements:

- `Check(nodeID, date)` / `Uncheck(nodeID, date)`, rejecting a node whose `type` is not
  `habit` with a matchable error.
- `Streak(nodeID)` — loads the checks and the `recurrence`, calls `domain.Streak`
  (**S1-12**) with the injected clock, returns the number. **No streak arithmetic in
  this file.**
- `DueToday(nodeID)` — whether today is a scheduled occurrence, for the habit strip.
- A habit with no `recurrence` is invalid; creating one must already have been rejected
  (S1-18) and `Streak` returns a matchable error rather than 0 — **0 and "cannot say"
  are different answers** and the UI will render them differently.
- **Habits never appear in Kanban columns** (**D2**). This service is their only read
  path, and S1-21's board query must exclude them — cross-check that here with a test.

**Acceptance criteria**
- [ ] Check → `Streak` reflects it; uncheck → it does not.
- [ ] A weekly habit checked on four consecutive scheduled dates reports **4** through
      the service, against a real temp DB.
- [ ] Today scheduled but unchecked → the streak is **unchanged**, not 0.
- [ ] `Check` on a `task` → matchable error.
- [ ] `Streak` on a habit with no `recurrence` → matchable error.
- [ ] `DueToday` true on a scheduled date, false otherwise, with a fixed clock.
- [ ] Fixed clock throughout.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(service): add habit checks and streak reads (S1-20)`

---

## S1-21 — feat: TaskService reads — board and tree assembly

The read path that makes every derivation in this stage **reachable**. Without it the
rules are written and never exercised end to end, and Stage 2 would be tempted to
recompute them in TypeScript — which is the one thing `PLAN.md` §1 forbids.

**Scope (may touch):** `internal/service/read.go`, `internal/service/read_test.go`,
`internal/service/dto.go`.

Requirements:

- A `NodeView` DTO carrying, **already computed by Go**: the node's fields, its
  **derived** status, its **progress** (with the `Defined` flag from S1-07), its
  **overdue** flag, its tags, and whether it is a leaf. The frontend computes nothing
  (`PLAN.md` §1) — this DTO is the contract that makes that possible.
- `Tree(rootID *string)` — the full tree with every derivation applied, one query for
  the nodes plus one for the tags. **No N+1**: loading a 200-node tree must not issue
  200 queries, and a test asserts the query count or, at minimum, that the subtree is
  fetched with a single recursive CTE.
- `Board()` — the five columns, each a list of `NodeView`, with:
  - **habits excluded entirely** (**D2**),
  - archived nodes excluded,
  - notes excluded from the columns (they have no status),
  - parents placed in their **derived** column,
  - `sort_order` respected within a column.
- `Progress(nodeID)` for a project.
- All of it read-only; no writes on this path.

**Acceptance criteria**
- [ ] `Board` places a parent in its derived column, not in a stored one — build a tree
      whose parent's stored status disagrees with its children and assert the derived
      placement.
- [ ] Habits never appear in any column.
- [ ] Notes never appear in any column.
- [ ] **No type without a Kanban column** — `note` **and** `habit`, `!Type.HasColumn()`
      (**D10**) — ever appears in a column or in a progress denominator.
- [ ] Archived nodes never appear.
- [ ] `overdue` is true for a node due yesterday and not done, false for one due today
      — computed in Go, present on the DTO.
- [ ] A project with nothing under it that counts as work — all notes, all habits, or
      any mix of types with no Kanban column (**D10**) — reports progress with
      `Defined == false`
      **when that project is the node being asked about** — this is D11 part 1 and is
      *not* superseded. Asked about its **parent**, the same project is one work leaf
      in the denominator (**D11** part 2).
- [ ] A 200-node tree is assembled without per-node queries (assert with a counting
      driver wrapper or by the shape of the calls).
- [ ] Fixed clock throughout.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(service): assemble the board and tree with derived state (S1-21)`

---

## S1-22 — feat: SearchService

**Scope (may touch):** `internal/service/search.go`,
`internal/service/search_test.go`.

Requirements:

- `Search(query string, opts)` over `store.Search` (**S1-17**), returning the same
  `NodeView` DTO as S1-21 so that a search result renders exactly like a card.
- **Backend-agnostic**: this file must not know whether FTS5 or `LIKE` is underneath.
  If it needs to know, S1-17's abstraction is wrong and that is the thing to fix.
- Query trimming: a whitespace-only query returns an empty result, not everything.
- Archived nodes excluded unless asked for.
- Tag / type / status / date **filters are Stage 3** — not here. Leave the `opts`
  struct extensible and empty of them.

**Acceptance criteria**
- [ ] Finds by title and by description, returning full `NodeView`s with derived status
      and overdue populated.
- [ ] Whitespace-only and empty queries return empty results.
- [ ] Archived excluded by default.
- [ ] `internal/service/search.go` contains neither `MATCH` nor `LIKE`
      (`git grep -nE 'MATCH|LIKE' internal/service` returns nothing).
- [ ] The same test suite passes whichever S1-04 branch is in force.
- [ ] `make cover` green.
- [ ] `make check` green.

**Commit:** `feat(service): add the search service over the store backend (S1-22)`

---

## Stage 1 — DONE criteria — ALL MET, PASS

Stage 1 closed only once **all** of these held. They do, verified by the Reviewer at
`a1f09b7` on the fourth review:

1. [x] All twenty-two tickets are committed, one conventional commit each, in order,
   `S1-01` … `S1-22`.
2. [x] `make check` is green — all five gates, unchanged in number and definition.
3. [x] **`make cover` is green: `internal/domain` ≥ 90.0% and `internal/service` ≥ 90.0%,
   measured per package** by
   `go test -covermode=atomic -coverprofile=… ./internal/{domain,service}/...` and read
   off the `total:` line of `go tool cover -func=…`. **Actual, re-measured after
   `go clean -testcache`: `internal/domain` 100.0%, `internal/service` 92.9%.**
4. [x] Every rule in `PLAN.md` §4 has table-driven tests, and specifically these four named
   edge cases each have a findable, named subtest:
   - **parent → Done cascades** to every unfinished descendant **that has a Kanban
     column** (D10 — `note` and `habit` descendants are skipped), sets `completed_at`,
     and the parent then **derives** `done` (S1-10, S1-18),
   - **circular parent** is rejected and the transaction rolls back (S1-09, S1-18),
   - **overlapping timers** are impossible — exactly one open entry, enforced by the
     service *and* by the schema (S1-05, S1-15, S1-19),
   - **`due_source` transitions**, all four (S1-08, S1-18).
5. [x] The FTS5 decision is recorded in S1-04's commit body and S1-17 implements that
   branch and only that branch.
6. [x] `CGO_ENABLED=0 go build ./...` succeeds and `go list -deps ./... | grep -i mattn`
   prints nothing.
7. [x] `internal/domain` is still pure — `TestDomainIsPure` and `TestDomainReadsNoClock`
   pass **unmodified**.
8. [x] `git status --porcelain` is empty after a full `make check && make cover`; no
   `*.db` and no `coverage*.out` in the tree.
9. [x] `git log` shows no AI author and no co-author trailer on any commit.
10. [x] `0001_settings.sql` is byte-identical to its Stage 0 state.
11. [x] **The Reviewer re-checks the three failed reviews' fixes and returns PASS.**
    This happened on the **fourth** review, at `a1f09b7`: the fixes for all three
    earlier FAILs were re-checked, coverage was re-measured from a cleared test cache,
    and an independent grep confirmed the generalised wording of D2 and D9. Per
    `PLAN.md` §5 a stage cannot close without a PASS — this is that PASS.

Nothing was marked done that was not actually verified.

**Out of Stage 1 scope, and it stayed out**: any Wails binding or `app.go` method, any
`frontend/` change, Kanban, the habit strip, quick-add, the command palette, the tray,
D-Bus sleep/lock handling, attachment file copying, backup/export, the PMP timelog
generator, the calendar, stats and Gantt. Also out of scope: the **K1** locale
workaround, **K2** (a leaf project's stale stored status now decides whether it counts
as done) and **K3** (an empty project stored `done` renders in Done with no bar). Those
three are open and now belong to Stage 2 — see below. **K4 is not among them**: a
no-column type taking children was a real defect, **fixed** in `9664506`, and it is
kept on the record in `PLAN.md` §7 as **RESOLVED** rather than deleted.

---

## Carried into Stage 2

**These are Stage 2 acceptance criteria, not suggestions.** Stage 1 closed clean, but it
closed owing five things. Each one is here because it will otherwise be forgotten: two
are couplings and costs no ticket ever picked up, three are the open known issues.
Whoever plans Stage 2 must turn every item below into a ticket with a checkable
criterion, or state explicitly why not.

### C1 — Wire `MoveToColumn(doing)` to `TimerService.Start`

`PLAN.md` §4 couples them: *"Moving a card to Doing opens a `time_entry`."* **No Stage 1
ticket did the wiring.** `TaskService.MoveToColumn` and `TimerService.Start` are both
implemented, both tested, and merely composable — nothing calls one from the other. The
Reviewer's warning: unless this is an **explicit Stage 2 acceptance criterion**, the
coupling silently never ships and the product quietly loses a §4 rule.

It is therefore already written into the §5 Stage 2 acceptance cell in `PLAN.md`:
*moving a card to Doing must itself open a `time_entry`*. Stage 2 must have a test that
fails if the wiring is removed.

- [ ] Moving a node to `doing` opens exactly one `time_entry` for it, in the same
      transaction semantics as the move.
- [ ] The single-active-timer invariant still holds: moving a second card to `doing`
      closes the first card's entry rather than opening a concurrent one.
- [ ] A move refused by `domain.DoingRefusal` (a `project`) opens **no** entry.

### C2 — `Board()` is O(n²·log n); fix it before the board is on screen

`snapshot.view` (`internal/service/read.go`) calls `domain.DeriveStatus` and
`domain.ComputeProgress` with the **whole node slice**, and each of those rebuilds
`indexByID` from scratch — so the index is rebuilt **once per node** and assembling the
board is O(n²·log n). At the ~200 nodes a real personal board holds this is invisible;
at 10k it is not. Fix it **before** the Kanban renders, not after — once the UI is live
the regression is a user-visible stutter and the fix competes with feature work.

- [ ] The index is built once per `Board()` call, not once per node.
- [ ] `internal/domain` stays pure and keeps its single spelling of each rule — the fix
      is an indexing change, not a second derivation path.
- [ ] A benchmark or a sized test documents the complexity change.

### C3 — K1: `BackgroundColour` never reaches GTK under `LC_NUMERIC=ru_RU.UTF-8`

Upstream Wails formats the window background's alpha with the **process** locale at
`window.c:205`, emitting `rgba(27, 38, 54, 0,0)`; GTK's CSS parser rejects the comma and
the background is **silently** never applied. Three options are recorded in `PLAN.md`
§7 — force `LC_NUMERIC=C` before `wails.Run`, leave the window transparent and let the
frontend paint it, or patch and pin a fork — and **none is chosen**. Choose it **with
the palette work**, where the window background first has to match a token.

- [ ] One of the three options is chosen, recorded as a decision, and implemented.
- [ ] The dark theme shows no flash of a wrong background on this machine.

### C4 — K2: a leaf project's stale *stored* status has teeth

Since **D11** a leaf project's stored status decides whether it counts as done in its
parent's denominator. **Archiving the last real child of a project that an earlier
cascade wrote `done`** leaves that project counted as a **done** unit on a status nobody
set deliberately. The state is self-consistent — column and bar agree — so it is not a
contradiction, which is exactly why it will not announce itself.

- [ ] Stage 2 rules on whether a project's stored status is re-inspected when its last
      child is archived, and the ruling is recorded as a decision in `PLAN.md` §7.

### C5 — K3: an empty project stored `done` renders in Done with no bar

Its own progress is undefined (**D11**, part 1) so no bar is drawn, while its status
puts the card in the Done column. **It is the one place a finished card shows nothing.**
This is a card-chrome decision.

- [ ] The card-chrome work decides what a done-but-unmeasurable card renders, and the
      decision is recorded.

### Engineering note — one rule, one spelling. In TypeScript too.

**Three of Stage 1's four review rounds failed on the same defect**: a type rule written
down in two places and edited in one. Every one of those duplicates was a door illegal
rows walked through, and the worst of them — K4 — had been written off in the plan as
harmless before a sweep attributed ~33,100 invariant violations to it.

The inventory is **clean** as of `9664506`. Each of these is **the single definition of
its rule**:

| Rule | Single spelling |
|---|---|
| does this type get a Kanban column? | `domain.NodeType.HasColumn` |
| may this type carry a due date? | `domain.NodeType.HasDue` |
| may this node be `doing` / be timed? | `domain.DoingRefusal`, with `NodeType.CanBeDoing` defined in terms of it |
| is this a unit of work for progress? | `countsAsWork` |
| a habit must have a recurrence | the habit-requires-recurrence check |
| default activity for a type | `DefaultActivity` |

`countsAsWork` and `DoingRefusal` admit the same set of types **today** but answer
different questions (D7/D11 versus D9); they are kept separate on purpose and must not
be merged.

**Stage 2 must not introduce a second spelling of any of them — in Go or in
TypeScript.** The frontend renders what Go returns and **re-implements none of these
rules**: not a status, not a column eligibility, not "can this be dragged to Doing", not
progress, not a streak, not an overdue flag, not a due date. Every derived value arrives
as a plain field on the DTO. A rule re-derived in a React component is a second
spelling, and a second spelling is the defect class that cost this project three review
rounds.

- [ ] No Stage 2 ticket adds a second implementation of any row in the table above.
- [ ] No `frontend/src` file computes a status, progress, streak, overdue flag or column
      eligibility.
