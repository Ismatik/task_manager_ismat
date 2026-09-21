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
- **[Stage 2 — Kanban + Habits strip](#stage-2--kanban--habits-strip): IMPLEMENTED,
  NOT CLOSED.** Twenty-two tickets, **S2-01 … S2-22, all committed**, plus the
  gap-closing `7af4d1d`. `make check` green (all five gates), `make cover` **100.0% /
  93.8%**, `make front-test` **22 files / 251 tests**, `make guard` clean over all **six**
  checks with `GUARD_ALLOW_RE` **empty**. **The Reviewer has not ruled — a stage closes
  only on PASS**, and four checks are recorded as
  [owed to a hand pass](#owed-to-a-hand-pass--not-verified), not as verified. All five
  items under [Carried into Stage 2](#carried-into-stage-2) are absorbed into named
  tickets — **C1 → S2-03, C2 → S2-01, C3 → S2-08, C4 → S2-06, C5 → S2-14** — and the
  fifth, *one rule one spelling*, is enforced by `make guard` (**S2-10**) and re-checked
  on every frontend ticket. Two planning defects were **corrected in place** mid-stage,
  both reported by the Dev rather than silently widened: nothing owned `App.tsx` or
  `main.tsx` (see [Composition — who mounts what](#composition--who-mounts-what)), and
  the out-of-scope line contradicted the brief on *set priority* (see
  [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2)).

Decisions referenced as **D1–D16 / E1–E3** and known issues **K1–K5** live in
[`PLAN.md` §7](./PLAN.md). Four decisions were confirmed by the user *during* Stage 1
and are recorded there — **D8** (a column move always overwrites the due date),
**D9** (type beats leaf-ness — a project is never timeable), **D10** (a node with no
Kanban column is excluded from parent derivation, generalising §4) and **D11** (an
empty project is one unfinished work leaf in its parent). Four more were added when
Stage 2 was planned: **D12** (force `LC_NUMERIC=C` and seed the window background from
`settings` — the user's, closing **K1**), and the PM rulings **D13** (what the
Doing↔timer coupling actually does), **D14** (archiving re-inspects a node that becomes
a leaf — closing **K2**) and **D15** (a card whose progress is undefined says so —
closing **K3**). **D16** was added *during* Stage 2, when the Dev asked what draws
Aurora's background drift: the gate is built, the visual is deferred, and **K5** records
the consequence. **Every known issue now has a decision; none is open.**

**Do not implement anything that is not on a ticket**, and do not add tickets here —
the PM writes them.

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

**ABSORBED — this is the record, the tickets are the work.** Stage 2 planning turned
all five into tickets. Nothing below is a floating obligation any more:

| | Obligation | Ticket | Decision it needed |
|---|---|---|---|
| **C1** | `MoveToColumn(doing)` → `TimerService.Start` | [S2-03](#s2-03--feat-the-doingtimer-coupling-c1-d13) | **D13** (PM ruling) |
| **C2** | `Board()` is O(n²·log n) | [S2-01](#s2-01--perf-build-the-derivation-index-once-per-board-c2) | — |
| **C3** | K1, the `LC_NUMERIC` background bug | [S2-08](#s2-08--fix-force-lc_numericc-and-seed-the-window-background-from-settings-c3-k1-d12) | **D12** (user) |
| **C4** | K2, a leaf project's stale stored status | [S2-06](#s2-06--fix-archiving-re-inspects-a-node-that-becomes-a-leaf-c4-k2-d14) | **D14** (PM ruling) |
| **C5** | K3, a done card with no bar | [S2-14](#s2-14--feat-the-card-c5-k3-d15) | **D15** (PM ruling) |
| — | one rule, one spelling — in TypeScript too | [S2-10](#s2-10--build-vitest-and-make-guard-the-mechanical-rules-check) `make guard`, re-checked on every frontend ticket | — |

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

---

## Stage 2 — Kanban + Habits strip

**Status: IMPLEMENTED, NOT CLOSED — S2-01 … S2-22 all committed, plus the gap-closing
`7af4d1d`. The Reviewer runs next.**
Twenty-two tickets, **S2-01 … S2-22**, one conventional commit each. This is the
**launch screen**: the first stage whose output a user can look at.

**It is not closed, and this document must not be read as saying it is.** Per
`PLAN.md` §5 a stage closes only on a Reviewer **PASS**, and none has been returned. The
[DONE criteria](#stage-2--done-criteria) below are therefore still **unticked**: they are
the Reviewer's to verify, not the PM's to assert. What the PM *can* state is what was
measured on the tree as committed:

| | |
|---|---|
| `make check` | green — all five gates, unchanged in number and definition |
| `make cover` | `internal/domain` **100.0%**, `internal/service` **93.8%** (bar ≥90%) |
| `make front-test` | **22 files, 251 tests** |
| `make guard` | all **six** checks pass, `GUARD_ALLOW_RE` **empty** |

Four things nobody on this machine can check are listed under
[Owed to a hand pass](#owed-to-a-hand-pass--not-verified). They are **owed**, not done.

**Corrected in place after S2-13.** The Dev reported, in `97873d9`'s body rather than by
silently widening scope, that **no remaining ticket had `frontend/src/App.tsx` or
`frontend/src/main.tsx` in its Scope** — so every ticket from S2-14 on could pass its own
criteria while the running app stayed an empty screen. That was a planning defect, not a
Dev defect, and it is fixed below: see
[Composition — who mounts what](#composition--who-mounts-what). Two smaller corrections
came with it — `frontend/src/lib/format.ts` is now in S2-11's Scope where its
Requirements had already named it, and S2-13's six-line `main.tsx` widening is
[ratified](#s2-13--feat-the-go-client-the-zustand-store-and-the-error-toast). **No
committed ticket's meaning changed**; S2-01 … S2-13 stand as reviewed.

### The plan correction — "set priority" from the palette is Stage 2

**This document contradicted the brief, and the brief wins.**

[What Stage 2 must NOT do](#what-stage-2-must-not-do) and the out-of-scope paragraph
under the DONE criteria both listed *"the tag/due/priority/estimate **editors**"*. The
user's brief specifies the Stage 2 command palette's action set as *"new, move to column,
**set priority**, start/stop timer, switch view, toggle theme, switch language"* — and
[S2-20](#s2-20--feat-the-command-palette-ctrlk)'s own Requirements table has carried the
`set priority` row since the ticket was written.

**The contradiction had teeth.** No binding set a priority, so S2-20 (`8aeb6e1`)
registered its four `set priority` rows as **unavailable with a reason**, and said so in
the commit body: the criterion *"every action in the table is reachable and executable by
keyboard alone"* could not be met for that row without a Go change the ticket's Scope did
not permit, so it **needs a PM ruling**. The Dev **refused to widen into Go and reported
it** — the second time in this stage that rule caught a planning defect rather than a
coding one.

**The ruling.** Setting a priority **from the palette** is **Stage 2**. The **priority
editor in a detail panel** is **Stage 3**, together with the tag, due and estimate
editors. The out-of-scope line said *editors* and meant *editors*; it was too coarse to
say so, and it is corrected in both places it appears. This is a **plan correction, not a
scope change by the Dev**: the Dev did the right thing by stopping.

**What landed, in `7af4d1d`** — `feat(service): add a priority setter and wire the palette
rows (S2-20)`:

- `TaskService.SetPriority`, built to `SetDue`'s shape — read, apply one field, let
  `domain.Node.Validate` decide, write, read back. A refused value writes nothing.
- **The 1..4 range is stated exactly once**, in `domain.Priority.Valid` — not in the
  service, not in `app.go`, not in `lib/client.ts`, not in the palette. One rule, one
  spelling; the stage's prime directive is intact.
- `App.SetPriority` takes a plain `int` for the reason `MoveToColumn` does: a
  `domain.Priority` parameter generates a type name the binding generator never emits into
  `models.ts`. The regenerated client says `arg2:number`, which is what `models.ts` has
  always called `Node.priority`. `frontend/wailsjs` is committed in the same commit
  (rule 12).
- The palette's rows take their set from the `palette.priority` label table as before and
  call the **store**, never the client. Nothing is written locally, so a Go rejection
  raises its one toast and leaves no applied change on screen.

**S2-20's criterion *"every action in the table is reachable and executable by keyboard
alone"* is therefore now met.**

### Three disclosed scope widenings — all ratified

Each was minimal, each was **stated in its commit body** naming the file and the reason,
and each falls under the general rule set in
[S2-13's ratification](#s2-13--feat-the-go-client-the-zustand-store-and-the-error-toast):
a widening that is minimal, disclosed, and needed to avoid leaving a rule with two
spellings is **acceptable**; an undisclosed one, or one that adds behaviour rather than
removing a duplicate, is **not**. **The Reviewer does not need to re-litigate any of the
three.**

| Ticket | Commit | Widening | Ruling |
|---|---|---|---|
| **S2-13** | `97873d9` | `frontend/src/main.tsx`, six lines | **RATIFIED** — recorded in full at the foot of [S2-13](#s2-13--feat-the-go-client-the-zustand-store-and-the-error-toast). The zustand store subsumed two hand-rolled stores, so enforcing Scope would have *preserved* a duplicate rule. Kept on the record here so it stays ratified. |
| **S2-21** | `642e677` | six existing test files, plus a new shared `frontend/src/test/render.tsx` | **RATIFIED** — see below. |
| **S2-22** | `58552bf` | the `Makefile`, to add `make guard` check 6 | **RATIFIED** — see below. |

**S2-21 — the six test files and `test/render.tsx`.** Mounting the appearance controls
into the header put focusable elements **ahead of the board in DOM order**, so six test
files that asserted *"the first `Tab` lands on the board"* (or *"the only button on the
page"*) became mechanically wrong the moment region 1 stopped being empty — which is
precisely what that ticket was for. The Dev introduced **`tabUntil(user, arrived)`** in a
new shared helper: **one spelling of "walk past the header" instead of six hard-coded
tab-stop counts.** So the widening **removed duplication rather than adding it**, which is
the same ground S2-13 was ratified on. **No assertion was weakened** — every updated test
still makes the same claim about the same element; the files are named one by one in the
commit body with what changed in each. Nothing outside `frontend/src` changed.

**S2-22 — the `Makefile`.** The ticket's Scope named the accept test and its helpers but
not the `Makefile`; its own acceptance criterion, however, is *"`make guard` fails if a
mouse event is introduced into the accept test"*. **The rule was required to live in
`make guard`**, so the Scope list was simply **incomplete** — this is a Scope-list
omission, not a widening in substance. Check 6 is scoped to that one file, greps the
three words a pointing device is spelt with (case-insensitively, code or comment), and
**fails if the file is missing or empty**, because otherwise the cheapest way to pass
would be to delete the thing being checked. The rest of the suite stays free to use a
pointer where a pointer is what is under test (S2-17's drag). **The five gates are still
five**: `guard` is not one of them and `make check` is untouched.

### Two known limits, recorded so nobody over-trusts them

- **The orphan check has a blind spot.** `App.mount.test.tsx` walks the **static import**
  closure of `main.tsx`, so **deleting an element while keeping its import leaves the
  check green.** The Dev demonstrated this deliberately, twice — removing
  `<CommandPalette />` from `App.tsx` under S2-20 and `<AppearanceControls />` under
  S2-21, each time watching the per-ticket reachability tests go red while the orphan walk
  stayed green. That is the two halves of the composition check doing **different jobs**,
  and it is exactly why the `render(<App />)` reachability criterion exists on every
  mounting ticket alongside the walk. **Neither half is sufficient alone**, and a future
  reader should not treat an empty orphan difference as proof the screen is assembled.
- **The palette's remaining unavailable rows are honest and transient.** After `7af4d1d`
  the only rows still registered unavailable are `view:tree`, `view:calendar` and
  `view:kanban`, and each carries a reason tied to a **later stage** — *"you are looking
  at it"* for Kanban, *"coming in stage N"* for the others — not a permanent limitation.
  `TASKS.md` offered *"disabled with a localised coming-in-stage-N, or omitted — pick one
  and be consistent"*; registered-with-a-reason is the choice, applied to every view row
  including Kanban's own. **No row silently does nothing.**

### ACCEPT

> **Create a task, move it across all five columns, and complete it — with no mouse.**

Plus the coupling `PLAN.md` §4 states and Stage 1 never wired: **moving a card to Doing
opens a `time_entry`** (**C1**, **D13**, ticket **S2-03**).

How it is demonstrated is not left to interpretation — see
[ACCEPT — how the no-mouse flow is demonstrated](#accept--how-the-no-mouse-flow-is-demonstrated)
below. It is demonstrated **twice**: once mechanically, by an automated test that can
only press keys, and once by hand against the real binary, following a numbered script.

### What Stage 2 must deliver

- The **wiring that does not exist yet**. `main.go` opens no database and constructs no
  service; `app.go` binds nothing but the scaffold's `Greet`. Everything else in this
  stage is blocked on that, which is why it is near the front.
- The two **carried couplings and costs**: C1 (Doing → timer) and C2 (`Board()`'s
  complexity), both before the board is on screen.
- The three **known issues**, each now with a decision: K1/**D12**, K2/**D14**,
  K3/**D15**.
- The **launch screen** itself: five columns, full card chrome, a habits strip,
  drag-and-drop, a complete keyboard model, quick-add, a command palette, the
  palette/theme/accent controls, and EN/RU.

### What Stage 2 must NOT do

Out of scope, and it must stay out: the detail slide-over, the Markdown editor, the
**tag / due / priority / estimate editors** (see the qualification below), the tree
view, the RRULE editor, attachments, the archive view, the search UI (the *service*
exists; the screen is Stage 3), the standalone quick-add **window** and the NL parser
(Stage 4 — Stage 2's quick-add is **in-app** and takes a plain title), the tray,
D-Bus sleep/lock, backup/export, the PMP timelog screen, the calendar, stats and Gantt.

> **Qualification on "the priority editor" — corrected, see
> [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2).**
> What is out of Stage 2 is the **editor**: a field in a detail panel where a user sets a
> tag, a due date, a priority or an estimate. The **command palette's `set priority`
> action is in Stage 2**, because the brief names it in the palette's action set and
> S2-20's Requirements table has always carried the row. The earlier, uncorrected wording
> of this line said only *"the priority editor"* and was read as forbidding the palette
> action too; that reading blocked S2-20's own acceptance criterion and is now settled the
> other way. The detail-panel editors remain **Stage 3**.

`README.md` is still deliberately unwritten. Every screenshot it needs belongs to this
stage or later; it is written once the board exists, not before.

### Rules for the Dev agent, Stage 2 additions

The six rules at the top of this file still apply in full. These are added:

7. **The frontend computes nothing.** Not a status, not a column, not progress, not a
   streak, not an overdue flag, not a due date, not "can this be dragged to Doing". If a
   component needs a value that is not on the DTO, the answer is a Go change, not a
   TypeScript one. `make guard` (**S2-10**) greps for the specific violations and is an
   acceptance criterion on **every** frontend ticket from S2-10 onwards.
8. **No hard-coded user-visible string, from the very first component.** Keys in
   `frontend/src/locales/en.json` and `ru.json`, both complete, on the same commit.
   **Russian runs ~30% wider**: a layout that only works in English is a broken layout,
   and "it fits in English" is not a passing answer.
9. **No hex literal in `frontend/src`.** `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src`
   returns nothing. Colours only through the token names (`bg`, `surface`, `elevated`,
   `line`, `ink`, `muted`, `accent`, `accent-2`, `on-accent`, `danger`, `warning`,
   `success`), radii only `rounded-sm/md/lg`, motion only `duration-fast/base/slow`,
   **JetBrains Mono (`font-mono`) for every number, date, timer and shortcut hint**.
10. **`design/` is read-only.** Not edited, not copied, not restated. Any diff touching
    `design/` is rejected on sight.
11. **Local-only.** No network call, no CDN, no font fetch, no update check, no
    telemetry. Fonts are already vendored through `@fontsource`. A new dependency that
    phones home at runtime is rejected.
12. **`frontend/wailsjs/` is generated and tracked.** `wails build` regenerates it, so a
    ticket that changes a bound signature **must commit the regenerated files in the same
    commit** — otherwise `git status --porcelain` is dirty the moment the next person runs
    `make check`, and criterion 8 of the DONE list fails for a reason unrelated to their
    work. Never hand-edit it.
13. **Every bound method returns `(T, error)`** (`ARCHITECTURE.md` §4), and the frontend
    surfaces every rejection in a toast. No silent failure, anywhere.
14. **`make cover` must stay green** — `internal/domain` and `internal/service` both
    ≥90%. New service code lands **with its tests in the same commit**, not in a
    follow-up; the bar is currently 100.0% / 92.9% and there is no headroom budget to
    spend.
15. **Negative-control testing, as in Stage 1.** For every ticket whose acceptance
    criterion says "a test fails if X is removed": actually remove X, watch the *named*
    test fail, restore, and say so in the commit body. A test nobody has seen fail is a
    test nobody has checked.

### New dependencies, and why each one is justified

Every one is a normal npm package, installed into `node_modules` and bundled by Vite.
**None makes a network call at runtime** — that is checkable and is an acceptance
criterion on the ticket that adds it.

| Package | Ticket | Why it, and not hand-rolled |
|---|---|---|
| `zustand` | S2-13 | Named in the brief. A store hydrated from Go, subscribed to by the board, the strip and the palette; hand-rolling a subscription store is a week of bugs nobody asked for. |
| `@dnd-kit/core`, `@dnd-kit/sortable`, `@dnd-kit/utilities` | S2-17 | Named in the brief. Pointer/keyboard sensors, collision detection, drop-target highlighting and accessibility announcements — all of which we would otherwise reimplement badly. |
| `i18next`, `react-i18next` | S2-11 | Interpolation, pluralisation (Russian has three plural forms — `one/few/many`, which a hand-rolled lookup gets wrong on day two) and lazy namespace loading. Bundled locally; no backend plugin, no HTTP loader. |
| `lucide-react` | S2-14 | Type icons, as inline SVG React components. Vendored in `node_modules`, no icon-font CDN, no sprite fetch, tree-shaken to the handful actually used. Colour comes from `currentColor`, so icons inherit the tokens and add no hex. |
| `vitest`, `jsdom`, `@testing-library/react`, `@testing-library/user-event`, `@testing-library/jest-dom` | S2-10 | **devDependencies.** The justification is the ACCEPT criterion itself: "keyboard-only" is an interaction property that neither `tsc` nor ESLint can see, and a claim only verifiable by a human at a keyboard is a claim the Reviewer cannot reproduce. `user-event` drives real key events through the real event pipeline. |

**Rejected:** a browser-driving e2e runner (Playwright/WebdriverIO). It would need a real
WebKit window, a display, and `xvfb`+`imagemagick` which are **not installed** on this
machine; it downloads browser binaries, which is a network dependency in a project whose
first rule is local-only; and the thing under test is a React keyboard model that jsdom
reproduces faithfully. The parts jsdom genuinely cannot judge — that the window actually
appears, that the background colour is right, that Russian does not clip — are covered by
the **hand** half of the ACCEPT script, honestly labelled as manual.

### Two non-gate make targets — the five gates stay five

`make check` remains **exactly** the five gates in the table at the top of this file.
Stage 2 adds two targets alongside it, on the same precedent as Stage 1's `make cover`:

| Target | Ticket | What it does |
|---|---|---|
| `make front-test` | S2-10 | `npm run test -- --run` in `frontend/` — the vitest suite. |
| `make guard` | S2-10 | The mechanical rules greps: no hex literal, no status-string literal, no recomputed rule, no bare user-visible string. Exits non-zero on any hit. **As shipped it runs six checks** — S2-10's five, plus **check 6**, the no-mouse rule over `App.accept.test.tsx`, added by S2-22 because that ticket's own criterion required the rule to live here ([ratified](#s2-22--test-the-no-mouse-accept-flow-and-the-ru--a11y-audit)). |

Both are **acceptance criteria on every frontend ticket from S2-10 onwards**, and both
are Stage 2 DONE criteria. Neither is a sixth gate. "Green" still means `make check`
exited 0.

---

## Stage 2 ticket index

| ID | Title | Prefix |
|---|---|---|
| [S2-01](#s2-01--perf-build-the-derivation-index-once-per-board-c2) | Build the derivation index once per `Board()` (**C2**) | `perf:` |
| [S2-02](#s2-02--feat-the-json-wire-contract-for-the-dtos) | The JSON wire contract for the DTOs | `feat:` |
| [S2-03](#s2-03--feat-the-doingtimer-coupling-c1-d13) | The Doing↔timer coupling (**C1**, **D13**) | `feat:` |
| [S2-04](#s2-04--feat-the-habit-strip-read) | The habit strip read | `feat:` |
| [S2-05](#s2-05--feat-settingsservice) | `SettingsService` | `feat:` |
| [S2-06](#s2-06--fix-archiving-re-inspects-a-node-that-becomes-a-leaf-c4-k2-d14) | Archiving re-inspects a node that becomes a leaf (**C4**, **K2**, **D14**) | `fix:` |
| [S2-07](#s2-07--feat-open-the-store-construct-the-services-bind-them) | Open the store, construct the services, bind them | `feat:` |
| [S2-08](#s2-08--fix-force-lc_numericc-and-seed-the-window-background-from-settings-c3-k1-d12) | Force `LC_NUMERIC=C`, seed the background from settings (**C3**, **K1**, **D12**) | `fix:` |
| [S2-09](#s2-09--chore-delete-the-scaffold-demo-and-lay-out-frontendsrc) | Delete the scaffold demo, lay out `frontend/src` | `chore:` |
| [S2-10](#s2-10--build-vitest-and-make-guard-the-mechanical-rules-check) | vitest and `make guard`, the mechanical rules check | `build:` |
| [S2-11](#s2-11--feat-the-i18n-runtime-enjson-and-rujson) | The i18n runtime, `en.json` and `ru.json` | `feat:` |
| [S2-12](#s2-12--feat-the-appearance-runtime--palette-theme-accent) | The appearance runtime — palette, theme, accent | `feat:` |
| [S2-13](#s2-13--feat-the-go-client-the-zustand-store-and-the-error-toast) | The Go client, the Zustand store and the error toast | `feat:` |
| [S2-14](#s2-14--feat-the-card-c5-k3-d15) | The card (**C5**, **K3**, **D15**) | `feat:` |
| [S2-15](#s2-15--feat-the-app-shell-and-the-five-columns) | **The app shell** and the five columns | `feat:` |
| [S2-16](#s2-16--feat-the-keyboard-model--focus-navigation-and-keyboard-move) | The keyboard model — focus, navigation, keyboard move | `feat:` |
| [S2-17](#s2-17--feat-dnd-kit-drag-and-drop-with-optimistic-move-and-rollback) | dnd-kit drag and drop, optimistic move with rollback | `feat:` |
| [S2-18](#s2-18--feat-the-habits-strip) | The habits strip | `feat:` |
| [S2-19](#s2-19--feat-quick-add-ctrln) | Quick add (Ctrl+N) | `feat:` |
| [S2-20](#s2-20--feat-the-command-palette-ctrlk) | The command palette (Ctrl+K) | `feat:` |
| [S2-21](#s2-21--feat-the-appearance-and-language-controls) | The appearance and language controls | `feat:` |
| [S2-22](#s2-22--test-the-no-mouse-accept-flow-and-the-ru--a11y-audit) | The no-mouse ACCEPT flow, and the RU / a11y audit | `test:` |

**Sequencing.** **S2-01 … S2-08 are Go** and touch no `frontend/src` file: they pay the
carried debt, give the DTOs a stable wire shape, add the two reads the screen needs, and
wire the application together. **S2-09 … S2-13 are the frontend's foundation** — nothing
renders a card until S2-14. **S2-14 … S2-22 are the launch screen**, with the keyboard
model (S2-16) deliberately **before** drag-and-drop (S2-17), because the ACCEPT criterion
is the keyboard path and it must not end up as an afterthought bolted onto a pointer API.

Nothing later in the list is needed by anything earlier. Three ordering constraints are
real and must not be reshuffled:

- **S2-02 before S2-07.** Binding a struct whose field names then change is a
  regeneration of `frontend/wailsjs` plus a rename across every consumer.
- **S2-07 before every frontend ticket.** Until the services are constructed and bound
  there is nothing for the frontend to call.
- **S2-05 before S2-08.** D12 seeds the window background *from settings*; there must be
  a settings read to seed it from.

### Composition — who mounts what

**The defect this section exists to fix.** As first written, no ticket after S2-12 had
`frontend/src/App.tsx` or `frontend/src/main.tsx` in its Scope. `ToastList` (S2-13), the
Kanban view (S2-15) and the overlays (S2-19, S2-20) would each have been built, tested in
isolation, and mounted by nobody. Every ticket would have passed, `make check`,
`make front-test` and `make guard` would all have been green, and **the app would have
opened on an empty page** — with the ACCEPT criterion, which is a property of the whole
screen, impossible to demonstrate. It was reported by the Dev under S2-13.

**The rule, from here to the end of the stage:**

> **A component that is built and not mounted is an unfinished ticket.** The ticket that
> builds a top-level piece is the ticket that mounts it, in the same commit. "Mounted"
> means reachable in the running app from `main.tsx` — not merely exported, and not
> merely rendered by its own test.

**Why the pieces own their own mounting, rather than one assembly ticket at the end.**
A single "assemble the screen" ticket reproduces the failure it is meant to prevent: nine
tickets would pass in a row with nothing on screen, and the one ticket that finds out
whether any of it composes is the last one, next to the deadline. An assembly ticket at
the *front* cannot mount components that do not exist yet, so it would leave empty slots
and every later ticket would have to touch `App.tsx` anyway — the distribution, without
the honesty. So: **S2-15 creates the shell**, because it is the first ticket that renders
a top-level region and because S2-13's `ToastList` is orphaned until it does; **every
later ticket that adds a top-level piece has `App.tsx` in its Scope** and mounts its own.

**The regions, and the ticket that fills each:**

| Region in `App.tsx` | Filled by | Mounted in |
|---|---|---|
| shell, store provider, board hydration | the shell itself | **S2-15** |
| `ToastList` layer (built in S2-13, orphaned until here) | `components/Toast.tsx` | **S2-15** |
| board region | `views/Kanban.tsx` | **S2-15** |
| global key handling and region order (`Tab` between regions) | `lib/keyboard.ts` | **S2-16** |
| habits strip region | `components/HabitStrip.tsx` | **S2-18** |
| overlay layer — quick add | `components/QuickAdd.tsx` | **S2-19** |
| overlay layer — command palette | `components/CommandPalette.tsx` | **S2-20** |
| header region — appearance and language | `components/AppearanceControls.tsx` | **S2-21** |

Everything else is mounted by its parent: the card of S2-14 by S2-15's `Column`, the
`AccentPicker` of S2-21 by its `AppearanceControls`, and so on. S2-14 is the one launch-
screen ticket with no mounting obligation, because a card is not a region — it becomes
reachable when S2-15 renders the column that holds it, which is one ticket later and is
S2-15's criterion, not S2-14's.

**How it is checked, rather than assumed.** Two halves, both mechanical:

1. **The orphan check** — `frontend/src/App.mount.test.tsx`, created in S2-15 and green
   from then on. It enumerates every non-test `.tsx` under `frontend/src/components/` and
   `frontend/src/views/` with `import.meta.glob`, computes the import closure of
   `frontend/src/main.tsx` by walking relative import specifiers, and **fails naming any
   module that is in the first set and not in the second**. A component built and left
   unmounted turns this test red on the commit that builds it. Walking from `main.tsx`
   and not from the test file is the whole point: a component imported only by its own
   test is still an orphan.
2. **A reachability criterion per mounting ticket** — because "imported" is not
   "reachable": a component can be mounted behind a condition that is never true. Each
   mounting ticket therefore asserts its piece **through `render(<App />)`**, driven by
   keyboard, with the test importing `App` and **not** importing the component it is
   checking for. That import restriction is the assertion; without it the test passes by
   rendering the component itself and proves nothing.

The orphan check is a **vitest** test and runs under `make front-test`. It is deliberately
not a `make guard` grep — reachability is a graph walk, not a pattern, and `make guard` is
specified to stay a set of greps that finish in under two seconds. **The five gates stay
five**, and `make front-test` and `make guard` stay non-gate targets.

---

## S2-01 — perf: build the derivation index once per `Board()` (C2)

**This is C2**, and it is scheduled first because the instruction attached to it was
*fix it before the board is on screen* — once the Kanban renders, this stops being a
number in a profile and becomes a stutter competing with feature work.

`snapshot.view` (`internal/service/read.go`) calls `domain.DeriveStatus(snap.all, n.ID)`
and `domain.ComputeProgress(snap.all, n.ID)`, handing each the **whole** node slice.
Both begin with `indexByID(nodes)` plus a children map — so the index is rebuilt **once
per node**, and `Board()` over n nodes is O(n²·log n). At the ~200 nodes a personal
board holds this is invisible. At 10k it is not.

**Scope (may touch):** `internal/domain/derive.go`, `internal/domain/derive_test.go`,
`internal/service/read.go`, `internal/service/read_test.go`, and a new
`internal/domain/derive_bench_test.go`.

Requirements:

- A **prebuilt index** type in `domain` — `NewIndex(nodes []Node) *Index` or equivalent —
  exposing the derivations as methods, e.g. `(*Index).Status(id)` and
  `(*Index).Progress(id)`.
- **`DeriveStatus` and `ComputeProgress` survive as thin wrappers over it**
  (`NewIndex(nodes).Status(id)`). They are called from elsewhere, they are the documented
  entry points, and — this is the whole point — **the derivation logic must exist exactly
  once**. A second traversal written "for the indexed path" is the defect class that cost
  Stage 1 three review rounds. This is an **indexing** change and nothing else.
- `snapshot` builds the index **once**, in `loadSnapshot`, and `view` uses it.
- `internal/domain` stays pure: no clock, no I/O, `TestDomainIsPure` and
  `TestDomainReadsNoClock` pass **unmodified**.

**Acceptance criteria**
- [ ] The index is constructed **once per `Board()` call**. Proven, not asserted in
      prose: a counter (a test-only hook or a package-level `var` guarded by the test)
      shows exactly one construction for a 500-node board.
- [ ] A **benchmark** — `BenchmarkBoardDerivation` or similar — over sized inputs
      (100 / 1 000 / 10 000 nodes) is committed, and the commit body records the
      before/after numbers. The growth must be visibly sub-quadratic.
- [ ] **No behaviour change.** Every existing `derive_test.go`, `read_test.go` and
      `sweep_test.go` case passes **unmodified** — this is a refactor and the sweep is
      the independent reference implementation that proves it.
- [ ] `git grep -n 'func deriveStatus\|func walkProgress' internal/domain` shows each
      recursion **exactly once**.
- [ ] `make cover` green (`internal/domain` and `internal/service` both ≥90%).
- [ ] `make check` green.

**Commit:** `perf(service): build the derivation index once per board (S2-01)`

---

## S2-02 — feat: the JSON wire contract for the DTOs

Every DTO in `internal/service/dto.go` and every type in `internal/domain` that they
carry has **no JSON tag anywhere in the repository** — checked:
`git grep -n 'json:"' internal/` returns nothing. Bound as they stand, Wails would
generate a TypeScript model with Go's exported field names (`Node`, `DescriptionMD`,
`DueSource`) and, worse, would render `domain.Date` as `{Year, Month, Day}` and
`time.Time` as whatever the marshaller happens to do. The frontend would then be built
against an **accidental** contract, and the first time a field is renamed in Go every
call site in TypeScript breaks silently.

So the contract is designed once, deliberately, **before** anything is bound — and before
S2-07 regenerates `frontend/wailsjs`.

**Scope (may touch):** `internal/service/dto.go`, a new
`internal/service/dto_test.go`, `internal/domain/node.go` (a `MarshalJSON` /
`UnmarshalJSON` on `Date` only), `internal/domain/node_test.go`.

Requirements:

- **`json` tags on every exported field** of `NodeView`, `ProgressView`, `TimerView`,
  `ColumnView`, `domain.Node`, `domain.Tag`, `domain.TimeEntry`, in **lowerCamelCase**
  (`descriptionMd`, `dueSource`, `estimateMin`, `sortOrder`, `completedAt`,
  `archivedAt`). One convention, no exceptions.
- **`domain.Date` marshals as `"YYYY-MM-DD"`** and unmarshals back. It is a calendar
  date, not an instant, and `{"Year":2026,"Month":9,"Day":21}` is a shape that invites
  TypeScript to do date arithmetic on it. A nil `*Date` is `null`.
- `time.Time` fields marshal as RFC 3339, which is the default — nothing to do but pin
  it with a test.
- **Nil slices marshal as `[]`, not `null`** — `Tags`, `Children`, `ColumnView.Nodes`.
  A frontend that must guard every list against `null` will forget once.
- `encoding/json` is the only import added to `domain`. It is **not** on the forbidden
  list and it is pure; `TestDomainIsPure` must pass unmodified, unedited.

**Acceptance criteria**
- [ ] A golden test marshals a fully-populated `ColumnView` containing a `NodeView` with
      a due date, tags, a running timer and children, and compares against an inline
      expected JSON string. The **exact key names are visible in the test source** — that
      string is the contract, and reviewing a rename means reviewing one diff hunk.
- [ ] `Date` round-trips: `marshal → unmarshal → equal`, including a nil `*Date` ⇄ `null`.
- [ ] An empty `NodeView` marshals `"tags":[]` and `"children":[]`, never `null`.
- [ ] `git grep -cn 'json:"' internal/service/dto.go internal/domain/node.go` is
      non-zero, and **no exported field on those types lacks a tag** — asserted by a
      reflection test walking the struct fields, not by eye.
- [ ] `TestDomainIsPure` and `TestDomainReadsNoClock` pass **unmodified**.
- [ ] `make cover` green. `make check` green.

**Commit:** `feat(service): give the view DTOs a stable JSON wire contract (S2-02)`

---

## S2-03 — feat: the Doing↔timer coupling (C1, D13)

**This is C1**, the rule `PLAN.md` §4 states and no Stage 1 ticket implemented:
*"Moving a card to Doing opens a `time_entry`."* `TaskService.MoveToColumn` and
`TimerService.Start` are both written, both tested, and merely **composable** — nothing
calls one from the other. The Reviewer's standing warning: unless this ships as an
explicit acceptance criterion, it silently never ships.

**Read `PLAN.md` §7 D13 before starting.** It settles the three things §4 leaves open,
and they are requirements, not suggestions: the timer opens **in the same transaction**
as the move; a **cascade opens no timer**; and moving **out** of `doing` — to any column,
`done` included — **closes** the node's open entry.

**Scope (may touch):** `internal/service/task.go`, `internal/service/timer.go`,
`internal/service/task_test.go`, `internal/service/timer_test.go`.

Requirements:

- One transaction. `TimerService.Start` currently opens its own via `runInTx`; factor out
  an **executor-bound** internal `start(ctx, exec, nodeID)` (and the matching `stop`) so
  that the public method and the coupled move both drive the *same* code, bound to the
  caller's transaction. **The single-active invariant must not be reimplemented** — it
  keeps its one spelling, in the schema's `one_open_timer` index and in the existing
  service path.
- **Exactly one move-to-column entry point is reachable from `app.go`, and it is the
  coupled one.** Two methods — one coupled, one not — is a second spelling with a
  call-site-shaped fuse (D13). If the shape chosen requires a new composing type, it is
  the *only* one bound; the uncoupled method must not be bindable.
- `domain.DoingRefusal` still decides who may be `doing` (**D9**). This ticket adds **no
  new predicate** and asks **no new question about a type**.

**Acceptance criteria**
- [ ] Moving a timeable leaf to `doing` opens **exactly one** open `time_entry` for it,
      committed with the move.
- [ ] **Atomicity, both ways:** if the timer insert fails, the move is rolled back too —
      no card sits in Doing with no entry. Test it by forcing the failure (start a timer
      on another node through a path that leaves it open, or inject a failing executor).
- [ ] Moving a **second** card to `doing` closes the first card's entry and opens one for
      the second. Never two open entries. (**Existing invariant, re-asserted through the
      new path.**)
- [ ] A **cascade** — dragging a parent to `doing`, which sets `doing` on every
      unfinished descendant with a column — opens **no** entry at all (**D13** §2).
- [ ] Moving a node **out** of `doing` closes its open entry, `ended_at` set from the
      injected clock. Asserted for **each** of the four destinations: `backlog`, `week`,
      `today`, `done` (**D13** §3).
- [ ] Moving a **project** to `doing` is refused by `domain.ErrProjectNeverDoing` and
      opens **no** entry and closes none (**D9**).
- [ ] Moving a node with **no Kanban column** is refused as before and touches no timer.
- [ ] **Negative control, and say so in the commit body:** delete the coupling, run the
      suite, watch the *named* test fail, restore. Name the test in the body.
- [ ] `git grep -n 'func.*MoveToColumn' internal/service` shows a single bindable
      entry point.
- [ ] `make cover` green. `make check` green.

**Commit:** `feat(service): open a time entry when a card moves to doing (S2-03)`

---

## S2-04 — feat: the habit strip read

The habits strip needs, for each habit, in **one** call: the habit, whether it is
scheduled today, whether it is checked today, and its streak. `HabitService` has
`Check`, `Uncheck`, `IsChecked`, `Streak`, `DueToday` and `DueOn` — **all per node** —
and **no way to list habits at all**: `Board()` excludes them by design (**D2**), and
`Tree()` returns the whole forest. Without this read the frontend would have to list the
tree, filter it by `type === 'habit'` and call three methods per habit — which is a
filter rule in TypeScript, an N+1, and exactly the shape rule 7 forbids.

**Scope (may touch):** `internal/service/habit.go`, `internal/service/dto.go`,
`internal/service/habit_test.go`.

Requirements:

- `Strip(ctx) ([]HabitView, error)` — every non-archived habit, in `sort_order` then
  `id`, each carrying: the node, **`ScheduledToday`**, **`CheckedToday`** and
  **`Streak`**, all computed in Go.
- The strip's **membership rule is `NodeType`**, asked of the domain — the same single
  spelling every other caller uses. Not a string comparison in the service.
- A habit **nested under a project** appears in the strip exactly once, like any other
  (**D10** — it contributes nothing to its parent, which is a different question).
- Archived habits are excluded.
- **Fixed number of queries**, independent of the habit count, in the style of
  `loadSnapshot`. A test pins it.
- The clock is the injected one; no `time.Now()` anywhere on the path.
- Whether the strip shows habits **not** scheduled today is the **frontend's** filter to
  apply — the service reports `ScheduledToday` and does not pre-filter, so the UI can
  decide without a second call.

**Acceptance criteria**
- [ ] A weekly habit checked four weeks running reports `Streak == 4` (**D5**) through
      this read, not only through `HabitService.Streak`.
- [ ] Today's still-pending occurrence does not break a live streak, seen through
      `Strip` (**D5**).
- [ ] `CheckedToday` flips with `Check`/`Uncheck` on the same fixed clock.
- [ ] A habit under a project appears exactly once; a task and a note never appear.
- [ ] Archived habits never appear.
- [ ] Query count is constant across 1 and 50 habits.
- [ ] `HabitView` carries JSON tags in the S2-02 convention and is in the golden test.
- [ ] `make cover` green. `make check` green.

**Commit:** `feat(service): add the habit strip read (S2-04)`

---

## S2-05 — feat: `SettingsService`

`store.SettingsRepo` exists — `Get`, `Set`, `All`, `SeedDefaults`, with the four keys
`palette`, `theme`, `accent`, `language` and the defaults `aurora` / `dark` / `""` /
`en` (**D6**). There is **no service over it**, so there is nothing for `app.go` to bind
and nothing for the appearance runtime (S2-12) or **D12** (S2-08) to read.

It also has **no validation**: `Set("theme", "purple")` succeeds today, and the value is
read back at the next startup by the code that paints the window.

**Scope (may touch):** a new `internal/service/settings.go` and
`internal/service/settings_test.go`; `internal/service/dto.go` if the view type lives
there.

Requirements:

- `Settings(ctx) (SettingsView, error)` returning the four values as a typed struct with
  S2-02's JSON tags — **not** a bare `map[string]string`, which gives TypeScript no
  contract at all.
- `SetPalette`, `SetTheme`, `SetAccent`, `SetLanguage`, each `(T, error)` and each
  returning the resulting `SettingsView`, so the frontend re-renders from the answer
  rather than from what it hoped it wrote.
- **Validation, with one spelling per enumeration.** `palette ∈ {aurora, studio}`,
  `theme ∈ {dark, light}`, `language ∈ {en, ru}`, `accent` either empty (meaning "use the
  palette's own") or a valid CSS colour. Each allowed set is defined **exactly once** in
  Go; a second list in TypeScript is a rule with two spellings, so the frontend builds
  its pickers from what the service reports, not from a literal array.
- An unreadable or unknown stored value **falls back to the default and does not
  error** — a settings row is not worth refusing to start over. A corrupt row read at
  startup must never be able to prevent the window opening.
- The accent validator is the **only** place a colour syntax is known in Go. It
  validates *shape*, it does not know any palette's values.

**Acceptance criteria**
- [ ] The four getters and setters round-trip through a real SQLite file.
- [ ] Every invalid value is refused with a distinguishable error: `SetTheme("purple")`,
      `SetPalette("")`, `SetLanguage("de")`, `SetAccent("not a colour")`.
- [ ] `SetAccent("")` is **accepted** — empty means "use the palette's accent" and is the
      seeded default.
- [ ] A row hand-corrupted to an unknown value reads back as the **default**, with no
      error.
- [ ] Each allowed set appears **once** in the Go source:
      `git grep -n 'aurora' internal/ | grep -v _test` shows exactly one definition site.
- [ ] `make cover` green. `make check` green.

**Commit:** `feat(service): add the settings service over the settings repo (S2-05)`

---

## S2-06 — fix: archiving re-inspects a node that becomes a leaf (C4, K2, D14)

**This is C4 / K2**, and **`PLAN.md` §7 D14 is the ruling — read it before starting.**

Since **D11** a leaf project's *stored* status decides whether it counts as done in its
parent's denominator. So: `P{C1:done, C2:backlog}` derives `backlog` and renders in
Backlog, while an old cascade left `done` sitting in `P`'s stored status. Archive `C1`
and `C2` and `P` becomes a leaf — and silently starts counting as a **done** unit in its
parent's bar, on a status nobody set. The state is self-consistent, column and bar agree,
and that is exactly why nothing will ever flag it.

**D14: when archiving leaves a node with no remaining child that has a Kanban column,
that node's stored status is rewritten to the status it derived *immediately before* the
archive**, with `completed_at` set or cleared to match. The principle is **continuity** —
what the board showed before the archive is what it shows after.

**Scope (may touch):** `internal/domain/tree.go` (the archive plan),
`internal/domain/tree_test.go`, `internal/service/task.go`,
`internal/service/task_test.go`.

Requirements:

- The rule is computed in `domain`, inside the existing archive planning, and applied by
  the service **in the archive's own transaction**.
- **No new predicate.** "Has no remaining child with a column" is `NodeType.HasColumn` +
  `Node.IsLeaf` (**D10**) asked of the set **as it will be after** the archive; the value
  written is `domain.DeriveStatus` of the set **as it was before**. Both already exist.
  Adding a third way to ask either question is the defect this project has paid for three
  times.
- The rewrite is **bounded and checkable**: it may touch only an ancestor that *this*
  archive turned into a leaf, and only to the status that ancestor was already
  displaying.
- It applies to **any** node type, not only `project` — the situation is not special to
  projects; **D11** is merely where it acquired teeth.
- **Restore needs no rule.** Once a node has column-bearing children again, derivation
  takes over and the stored status stops being consulted. Do not add one.

**Acceptance criteria**
- [ ] **The K2 scenario, named as such in the test:** `P` stored `done`,
      children `C1:done` + `C2:backlog`, `P` derives `backlog`; archive both children;
      `P`'s stored status is now `backlog` and `completed_at` is `NULL`; `P` counts as
      **one unfinished** work leaf in its parent (**D11**).
- [ ] The honest-completion case still works: `P` with children all `done` derives
      `done`; archive them; `P` stays `done` with `completed_at` preserved, and counts
      as a done unit.
- [ ] Archiving a child that leaves **other column-bearing children** behind rewrites
      **nothing** — the node is not a leaf and derivation still applies.
- [ ] Archiving a `note` or `habit` child rewrites nothing, because the node was
      **already** a leaf by **D10** and nothing changed.
- [ ] The rewrite happens in the **same transaction**: a failed archive leaves the
      stored status untouched.
- [ ] `RestoreNode` adds no symmetric rule, and a test shows that restoring a child
      brings derivation back without any stored-status write.
- [ ] The whole `sweep_test.go` property sweep passes unmodified.
- [ ] **Negative control, named in the commit body.**
- [ ] `make cover` green. `make check` green.

**Commit:** `fix(domain): re-inspect a node's stored status when archiving makes it a leaf (S2-06)`

---

## S2-07 — feat: open the store, construct the services, bind them

**Nothing is wired.** `main.go` acquires the single-instance lock and calls `wails.Run`;
it opens no database, runs no migration and constructs no service. `app.go` holds the
scaffold's `Greet` and the IPC handler. Every ticket after this one is blocked on it,
and it is the ticket that turns two closed Go stages into an application.

**Scope (may touch):** `main.go`, `app.go`, a new `app_test.go`,
`frontend/wailsjs/**` (**regenerated by `wails build`, committed with this commit** —
rule 12).

Requirements:

- On startup, in `main.go`: resolve the DB path via `store.DefaultPath()`, `store.Open`,
  run the migrations, `SeedDefaults()`, construct `TaskService`, `TimerService`,
  `HabitService`, `SearchService` and `SettingsService` with the **real wall clock** and
  a uuid generator, and bind **one** `App` carrying them.
- A **failure to open or migrate the database is fatal and visible**: log it and exit
  non-zero. Opening a window onto a database that is not there is worse than not opening
  one — a silent empty board looks exactly like "you have no tasks".
- `app.go` gains **thin delegations only**. Every one returns `(T, error)`
  (`ARCHITECTURE.md` §4), every one takes and returns the S2-02 DTOs, and **not one of
  them contains a rule**. A conditional in `app.go` that decides something is a rule in
  the wrong layer.
- The bound surface for Stage 2, and nothing beyond it: `Board`, `Tree`, `Progress`,
  `CreateNode`, `MoveToColumn` (the **coupled** one from S2-03 — the only one bindable),
  `MoveNode`, `SetDue`, `ArchiveNode`, `RestoreNode`, `Search`, `HabitStrip`,
  `CheckHabit`, `UncheckHabit`, `TimerStart`, `TimerStop`, `TimerCurrent`, `Settings`,
  `SetPalette`, `SetTheme`, `SetAccent`, `SetLanguage`.
- **`Greet` is deleted**, along with its frontend caller (S2-09 removes the caller; if
  the ordering makes that awkward, delete the demo box here and say so).
- `main.go` stays thin (`ARCHITECTURE.md` §1): flags, lock, open, construct, bind. **No
  business rule.**
- `internal/domain` and `internal/store` are **not** touched.

**Acceptance criteria**
- [ ] A fresh run against an empty `XDG_DATA_HOME` creates the database, applies every
      migration and seeds the four settings — asserted by a Go test that points
      `store.DefaultPath()` at a temp dir, not by launching the app.
- [ ] Every bound method returns `(T, error)` — asserted by a **reflection test** over
      `App`'s exported methods, so a future method cannot forget.
- [ ] No bound method contains a branch that decides a domain question:
      `git grep -nE 'domain\.(Status|NodeType)[A-Za-z]* ==' app.go` returns nothing.
- [ ] `frontend/wailsjs/go/main/App.d.ts` and the generated `models.ts` list the bound
      surface above, and are **committed in this commit**.
- [ ] `git status --porcelain` is **empty** after `make check` — the regenerated bindings
      are in the commit, not left dirty.
- [ ] `Greet` appears nowhere: `git grep -n Greet` returns nothing.
- [ ] The app opens a window and the frontend can call `Board()` and receive five
      columns. Verified by hand and reported.
- [ ] `make cover` green. `make check` green.

**Commit:** `feat(app): open the store, construct the services and bind them (S2-07)`

---

## S2-08 — fix: force `LC_NUMERIC=C` and seed the window background from settings (C3, K1, D12)

**This is C3 / K1, and `PLAN.md` §7 D12 is the decision — read it before starting.**
The user chose option 1 **plus a second half** the recorded options did not contain.

Wails formats the window background in C at `window.c:205` with the **process** locale,
so under this machine's `LC_NUMERIC=ru_RU.UTF-8` it emits `rgba(27, 38, 54, 0,0)` — a
comma where GTK's CSS parser needs a decimal point — and GTK **silently discards the
whole declaration**. Nothing is logged.

Two halves:

1. **Force `LC_NUMERIC=C` before `wails.Run`**, so the colour actually reaches GTK.
2. **Seed `BackgroundColour` from the persisted `palette` + `theme`** (S2-05), so the
   first frame — painted by GTK before the WebView has evaluated anything — matches the
   theme the user last chose, in **both** themes, with no flash of the wrong colour.

**Scope (may touch):** `main.go`, a new `background.go` + `background_test.go` at the
repo root (or `internal/platform`, if the Dev prefers and says why).

Requirements:

- **`main.go` must not contain a hex literal or an RGB triple.** This is the containment
  rule D12 spells out: the fix makes `main.go` a second place that knows a background
  colour, and a second spelling of a value is the defect class Stage 1 paid for three
  times. The colour is **derived** from `palette` + `theme` against the token values
  owned by `design/tokens.css` — parsed or generated from that file, never retyped.
- **`design/` is read-only** and is not edited by this ticket, only read.
- A value that cannot be resolved falls back to the **seeded default** (`aurora` +
  `dark`) rather than to a constant typed here, and never prevents the window opening.
- Only `LC_NUMERIC` is forced, not `LC_ALL`. Go's own formatting is locale-independent,
  so the blast radius is the cgo/GTK layer — which is the thing being fixed. Say so in a
  comment.
- The comment S1-02 left in `main.go`, which says the alpha fix is unobservable on this
  machine because of K1, is now **wrong** and must be updated, not left to rot.

**Acceptance criteria**
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}|RGBA\{[^}]*[0-9]' main.go` returns nothing.
- [ ] A unit test maps each of the four (palette, theme) combinations to the `--bg`
      value that `design/tokens.css` actually declares — read from the file, so the
      test fails if the token changes and the mapping does not.
- [ ] `LC_NUMERIC` is set to `C` **before** `wails.Run`, proven by a test on the
      extracted function rather than by reading `main`.
- [ ] `LC_ALL` is not touched: `git grep -n 'LC_ALL' .` returns nothing.
- [ ] **By hand, on this machine, and reported:** launch with
      `LC_NUMERIC=ru_RU.UTF-8`; the window background is the dark Aurora `--bg` from the
      first frame, with no flash. Then `SetTheme("light")`, restart, and the first frame
      is the **light** background. This is the half a unit test cannot see, and it is the
      half K1 is about.
- [ ] The stale S1-02 comment is updated.
- [ ] `make check` green.

**Commit:** `fix(app): force LC_NUMERIC=C and seed the window background from settings (S2-08)`

---

## S2-09 — chore: delete the scaffold demo and lay out `frontend/src`

`frontend/src` is still the untouched Wails scaffold: an `App.tsx` with a `Greet` demo
box, `App.css`, `style.css`, `main.tsx`, a Nunito woff2 under `src/assets/fonts/` and a
`logo-universal.png`. `App.css` carries a hard-coded `border-radius: 3px`, which is a
radius not from the tokens, in a project whose rule is `rounded-sm/md/lg` only.

This ticket empties the room before anything is built in it, and establishes the
**baseline** the mechanical checks are run against from S2-10 onwards.

**Scope (may touch):** `frontend/src/App.tsx`, `frontend/src/App.css`,
`frontend/src/style.css`, `frontend/src/main.tsx`, `frontend/src/assets/**`,
`frontend/index.html`, and the new empty directories
`frontend/src/{views,components,store,lib,locales}` (`ARCHITECTURE.md` §6).

Requirements:

- Delete `App.css`, the Nunito font files and `OFL.txt`, and `logo-universal.png`. The
  fonts that stay are the vendored `@fontsource` packages (Space Grotesk, Figtree,
  JetBrains Mono) — **loaded from `node_modules`, never from a CDN**.
- `App.tsx` becomes an empty shell: the token-styled page background and nothing else.
  No `Greet`, no demo.
- `style.css` keeps only the Tailwind directives and the `@fontsource` imports.
- `index.html` keeps `<html data-palette="aurora" class="dark">` as the **static
  default** — S2-12 makes it dynamic.
- Create the five directories of `ARCHITECTURE.md` §6 with a one-line `README` or a
  `.gitkeep` each, so the layout is visible before it is populated.

**Acceptance criteria**
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns **nothing**.
- [ ] `git grep -n 'border-radius\|borderRadius' frontend/src` returns nothing — radii
      come from Tailwind's `rounded-*`.
- [ ] `git grep -rn 'nunito\|Nunito\|logo-universal' frontend/` returns nothing.
- [ ] No remote URL anywhere: `git grep -nE 'https?://' frontend/src frontend/index.html`
      returns nothing.
- [ ] `frontend/src/{views,components,store,lib,locales}` all exist and are tracked.
- [ ] The app still builds and opens a window showing an empty, correctly-coloured page.
- [ ] `make check` green — gate 3 and gate 4 included.

**Commit:** `chore(frontend): remove the scaffold demo and lay out src (S2-09)`

---

## S2-10 — build: vitest and `make guard`, the mechanical rules check

Two things land here, both **before** the first component, because a rule introduced
after the code it governs is a rule that arrives as a rewrite.

**1. A test runner.** The frontend has none. The ACCEPT criterion is *"keyboard only, no
mouse"* — an **interaction** property that `tsc` cannot see, ESLint cannot see, and a
human at a keyboard cannot reproduce for the next reviewer. `@testing-library/user-event`
drives real key events through the real event pipeline, which makes "no mouse" a thing a
machine can assert. See the dependency table above for why **not** Playwright.

**2. `make guard`.** Stage 1 failed review three times on *a rule written twice and
edited once*. Stage 2's version of that defect is a rule re-derived in TypeScript. The
one defence that does not depend on a reviewer's attention is a grep that runs every
time.

**Scope (may touch):** `frontend/package.json`, `frontend/vitest.config.ts` (or the
`test` block in `vite.config.ts`), `frontend/src/test/setup.ts`,
`frontend/tsconfig.json`, `frontend/eslint.config.js`, `Makefile`, `.gitignore`.

Requirements:

- devDependencies: `vitest`, `jsdom`, `@testing-library/react`,
  `@testing-library/user-event`, `@testing-library/jest-dom`. `npm run test` runs vitest;
  `make front-test` runs it non-interactively from the repo root.
- Test files live beside their subject as `*.test.ts`/`*.test.tsx` and are covered by
  gate 3 (lint) and gate 4 (typecheck) like any other source.
- `make guard` runs these greps and **exits non-zero on any hit**:

  | # | Check | Grep |
  |---|---|---|
  | 1 | no hex literal | `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` |
  | 2 | no status string literal in code | the five status names, quoted, in `frontend/src` **excluding `frontend/src/locales/`** |
  | 3 | no re-derived rule | `overdue`/`isOverdue`/`derive`/`streak`/`progress` appearing as a **computation** — an assignment or a function body — rather than a field read off a DTO |
  | 4 | no bare user-visible string | JSX text nodes and `title=`/`aria-label=`/`placeholder=` attributes that are string literals rather than `t(...)` calls |

- Checks 3 and 4 are necessarily heuristic. **Say so in the target's own comment**, and
  keep them **tight enough to be actionable and loose enough not to cry wolf** — a guard
  that produces false positives gets disabled, which is worse than no guard. Where a
  genuine exception is needed it is a **named, commented allow-list entry inside the
  target**, never an inline suppression comment scattered through the source.
- **The one authorised presentation mapping**: priority → chip (`1→P0`, `2→P1`, `3→P2`,
  `4→no chip`, `ARCHITECTURE.md` §6) lives in **exactly one** module,
  `frontend/src/lib/priority.ts`, and the guard asserts `P0|P1|P2` appears nowhere else
  in `frontend/src` outside `locales/`.
- Both targets are **non-gate**, exactly like `make cover`. The five gates stay five.

**Acceptance criteria**
- [ ] `make front-test` runs and passes (a single trivial smoke test is enough at this
      point — the runner is the deliverable).
- [ ] `make guard` passes on the S2-09 baseline.
- [ ] **Each of the four checks is proven to fire.** Introduce a violation of each, one
      at a time, see `make guard` exit non-zero naming the file and line, revert. Record
      all four in the commit body — a guard nobody has watched fail is a guard nobody has
      checked.
- [ ] `make guard` completes in under two seconds; it will be run on every ticket.
- [ ] A test file is type-checked by gate 4 and linted by gate 3.
- [ ] No new **runtime** dependency: every package added here is a `devDependency`.
- [ ] `git status --porcelain` is empty after `make check && make front-test && make guard`
      — no coverage or cache artefact left in the tree.
- [ ] `make check` green.

**Commit:** `build(frontend): add vitest and the make guard rules check (S2-10)`

---

## S2-11 — feat: the i18n runtime, `en.json` and `ru.json`

**i18n from day one** (`PLAN.md` §2) — before the first component, which is the entire
point. There is no `frontend/src/locales/` directory yet.

**Scope (may touch):** `frontend/src/locales/en.json`, `frontend/src/locales/ru.json`,
`frontend/src/lib/i18n.ts`, `frontend/src/lib/format.ts`, `frontend/src/main.tsx`,
`frontend/package.json`, and the tests beside them.

> **Scope correction, after the fact.** `frontend/src/lib/format.ts` is named in this
> ticket's Requirements below — *"numbers, dates and durations are formatted in exactly
> one module"* — and was missing from the Scope list above, which is a PM error. The Dev
> created the file under this ticket and flagged the omission rather than inventing
> permission. **The Scope list is corrected to match the Requirements it already
> contained; the ticket's meaning is unchanged and `853077f` is not re-opened.**

Requirements:

- `i18next` + `react-i18next`, configured **entirely from bundled resources** — no HTTP
  backend plugin, no language-detector that fetches anything. The language comes from
  `SettingsService` (S2-05), not from the browser.
- Both locale files are **complete on this commit and stay complete on every later
  one**. A key added to `en.json` without `ru.json` is an incomplete ticket.
- Russian plurals use i18next's `one/few/many/other` forms — this is the concrete reason
  the library is here rather than a lookup object.
- The initial language is read from settings; `changeLanguage` writes it back through
  `SetLanguage` and the whole UI re-renders. Persisted across restart.
- **Numbers, dates and durations are formatted in exactly one module**
  (`frontend/src/lib/format.ts`), using `Intl` with the active locale, and always
  rendered in `font-mono` (`PLAN.md` §3). Formatting is presentation; the **values**
  still come from Go.

**Acceptance criteria**
- [ ] A test asserts `en.json` and `ru.json` have **identical key sets**, recursively,
      and fails naming the missing keys. This test outlives the stage.
- [ ] No empty-string values in either file.
- [ ] A test renders a component in `ru` and asserts the Russian string appears — the
      wiring is proven, not assumed.
- [ ] A Russian plural test covers 1 / 3 / 5 / 21 items and gets three distinct forms.
- [ ] Switching language persists: change it, reload the store, it is still Russian.
- [ ] `make guard` green — check 4 (no bare user-visible string) now has something to
      check.
- [ ] No network: `git grep -nE 'https?://' frontend/src` returns nothing, and the
      i18next config contains no backend plugin.
- [ ] `make front-test` green. `make check` green.

**Commit:** `feat(frontend): add the i18n runtime with complete en and ru locales (S2-11)`

---

## S2-12 — feat: the appearance runtime — palette, theme, accent

`index.html` hard-codes `<html data-palette="aurora" class="dark">`. The three values
already live in `settings` (**D6**) and now have a service (S2-05). This makes them live.

**Scope (may touch):** `frontend/src/lib/appearance.ts`, `frontend/src/store/**`,
`frontend/src/main.tsx`, `frontend/index.html`, and tests beside them.

Requirements:

- On startup, read `Settings()` and apply, exactly as `design/README.md` prescribes:
  `document.documentElement.dataset.palette = palette`,
  `classList.toggle('dark', isDark)`,
  `style.setProperty('--accent', accent)` when the accent is non-empty.
- **Apply before first paint** — a `<script>` in `index.html` or a blocking effect —
  so the WebView does not flash the default theme. This is the frontend's half of the
  same flash **D12** fixes on the GTK side.
- An empty accent **removes** the inline `--accent` override so the palette's own value
  takes over. Clearing must be clearing, not setting the empty string.
- Writes go through `SetPalette` / `SetTheme` / `SetAccent`; on a rejection the UI
  **reverts to the value the service reports** and raises a toast. The service's answer
  is the truth (S2-05 returns the resulting view for exactly this reason).
- `prefers-reduced-motion` is respected: `tokens.css` already kills transitions, and
  **Aurora's ~60s background drift must pause too** (`design/README.md`) — an animation
  driven by JS or by a CSS animation the media query does not already cover is this
  ticket's to gate.
- **No hex anywhere.** The accent value is whatever the user picked, held as a string and
  handed to `setProperty`; it is never a literal in the source.

**Acceptance criteria**
- [ ] All four (palette, theme) combinations apply the right `data-palette` and `dark`
      class — a test asserts the DOM attributes, not the colours.
- [ ] A non-empty accent sets the inline `--accent`; setting it back to `""` **removes**
      the property rather than setting it empty.
- [ ] Restarting the app restores the last chosen palette, theme and accent.
- [ ] A rejected `SetTheme` leaves the DOM on the service's value and raises one toast.
- [ ] Under `prefers-reduced-motion: reduce`, the Aurora drift is **not** running —
      asserted in a test, and confirmed by hand. **Read the note below before judging
      this one.**
- [ ] No flash of the default theme on startup when the stored theme is `light`.
      Confirmed by hand, together with S2-08's GTK-side half.
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing.
- [ ] `make guard`, `make front-test`, `make check` all green.

**Commit:** `feat(frontend): apply palette, theme and accent from settings (S2-12)`

> **PM ruling on the Aurora drift — D16, made during Stage 2. Read this with the
> criterion above.**
>
> This ticket asked for the drift to be **gated**, and `auroraDriftEnabled()` gates it:
> the decision is taken in exactly one place, respects `prefers-reduced-motion`, and is
> published on `<html>` as `data-drift="on"|"off"`. What no ticket in this stage ever
> asked for is the drift itself — **nothing draws it**, so today `data-drift` is a
> contract with no consumer. The Dev raised this rather than inventing a visual, which
> was the correct call: `design/` names **no component and no drift geometry**, and
> `PLAN.md` §3 and **D6 as amended** forbid inventing a spec and attributing it to the
> export.
>
> **Ruling: the gate ships in Stage 2, the visual does not.** `design/README.md`'s one
> sentence — *"Aurora's background drift (~60s) must also pause"* under reduced motion —
> is a **constraint on a drift**, not a specification of one: it fixes the period and the
> pause condition and says nothing about geometry, layer count, opacity or motion path.
> Building from that means inventing four things and attributing them to a document that
> contains none of them. Against that, the drift is pure decoration: it moves no card,
> blocks no key, and has **zero bearing on the ACCEPT criterion**. Inventing it is
> forbidden; guessing it and shipping it under the design export's name is worse than
> not shipping it. So it is **deliberately out of Stage 2**, recorded as **K5** in
> `PLAN.md`, and it becomes a Stage 3 ticket the moment a drift spec exists — at which
> point the gate is already built, already tested, and the ticket is only the drawing.
>
> **What this criterion therefore means, stated honestly so the Reviewer does not hunt
> for a visual that is not there:** the assertion is that `data-drift` is `off` under
> `prefers-reduced-motion: reduce` and `on` otherwise on Aurora — which
> `appearance.test.ts` asserts. Nobody may claim to have watched a drift pause. S2-22's
> a11y audit and the ACCEPT verification table are corrected to say the same thing.

---

## S2-13 — feat: the Go client, the Zustand store and the error toast

The single place TypeScript talks to Go, and the single place domain state lives on the
client.

**Scope (may touch):** `frontend/src/lib/client.ts`, `frontend/src/store/**`,
`frontend/src/components/Toast.tsx`, `frontend/package.json`, tests beside them.

Requirements:

- `lib/client.ts` wraps every generated `frontend/wailsjs/go/main/App` binding. It is the
  **only** file in `frontend/src` that imports from `wailsjs`, which makes the call
  surface greppable and mockable in one place.
- **Every rejection is surfaced in a toast** (`PLAN.md` §1 — no silent failures). The
  toast text is an i18n key; the raw Go error goes to the console, not to the user, and
  is never swallowed.
- A `zustand` store holding the **board, the habit strip and the settings as Go returned
  them**, plus genuinely-client-only UI state: current selection, focused column, which
  overlay is open, which card is being dragged (`ARCHITECTURE.md` §6 — *"`store` holds UI
  state; domain state is owned by Go and fetched"*).
- **The store never mutates a domain field to a value Go did not produce**, with one
  bounded exception, written down here so nobody has to guess: the **optimistic move**
  of S2-17, which moves a card between columns in the local copy and **reverts from Go's
  answer on error**. It is optimism about a value Go is about to compute, not a
  recomputation of it. Everything else re-reads.
- No derived value is ever computed here: no status, no progress, no overdue, no streak,
  no due date.
- Timer display ticks locally between reads — `dto.go` authorises exactly this
  (*"the UI ticks its own display between reads rather than polling Go every second"*).
  The **elapsed value it starts from** is Go's `elapsedSeconds`.

**Acceptance criteria**
- [ ] `git grep -rn "wailsjs" frontend/src --files-with-matches` lists **only**
      `frontend/src/lib/client.ts`.
- [ ] A rejected call raises exactly one toast carrying a translated message, and the
      store is unchanged.
- [ ] Hydration from a mocked client populates board, strip and settings; a test asserts
      the store's shape matches the DTOs field-for-field.
- [ ] `make guard` check 3 passes: the store computes no derived value.
- [ ] Toast text is a `t(...)` key in both locales.
- [ ] `make front-test`, `make guard`, `make check` all green.

**Commit:** `feat(frontend): add the Go client, the store and the error toast (S2-13)`

> **PM ruling — the `main.tsx` widening in `97873d9` is RATIFIED. The Reviewer does not
> need to re-litigate it.**
>
> `frontend/src/main.tsx` is not in this ticket's Scope and six lines of it were changed:
> S2-12 had created two hand-rolled stores there, and this ticket's zustand store
> subsumes both. The widening is accepted on four counts, and the fourth is the one that
> matters.
>
> 1. **It removed a second spelling, which is this stage's prime directive.** Leaving the
>    two stores in the tree would have left client state with two owners — the exact
>    defect that cost Stage 1 three review rounds, installed deliberately, in the stage
>    written to prevent it. The Scope rule exists to stop silent sprawl; enforcing it here
>    would have used it to *preserve* a duplicate rule.
> 2. **It was disclosed, not smuggled.** The commit body names the file, the line count
>    and the reason, under a heading. That is precisely the behaviour the Scope rule is
>    for: stop and say so. The Dev said so.
> 3. **It was minimal and mechanical** — construct one store instead of two, no behaviour
>    invented, no rule moved into TypeScript. All gates and both non-gate targets green.
> 4. **The cause was a PM defect, not a Dev defect.** Under the plan as written, S2-12 was
>    the last ticket that owned `main.tsx`, so there was no later ticket to hand the fix
>    to — the Dev was choosing between a permanent duplicate and a disclosed widening.
>    That gap is now fixed: see
>    [Composition — who mounts what](#composition--who-mounts-what).
>
> **The general rule this sets, so it is not re-argued per ticket:** a widening that is
> (a) minimal, (b) required to avoid leaving a rule with two spellings, and (c) stated
> plainly in the commit body naming the file and the reason, is **acceptable**. A
> widening that is undisclosed, or that adds behaviour rather than removing a duplicate,
> is **not** — it is still "stop and report". Nothing here loosens rule 15 or the Scope
> discipline; it names the one exception that was already being applied correctly.

---

## S2-14 — feat: the card (C5, K3, D15)

**This ticket also closes C5 / K3, and `PLAN.md` §7 D15 is the ruling — read it before
starting.**

The card renders, and computes nothing. Everything on it is a field Go already filled in:
`status`, `progress.percent`, `progress.defined`, `overdue`, `timer.running`, `tags`.

**Scope (may touch):** `frontend/src/components/Card.tsx`,
`frontend/src/components/{PriorityChip,DueBadge,ProgressBar,TagList,TypeIcon,TimerDot}.tsx`,
`frontend/src/lib/priority.ts`, the locale files, tests beside them.

Requirements — the card shows:

- the **title**;
- the **type icon** — `lucide-react`, coloured through tokens (`success` for task,
  `danger` for bug, per `design/README.md`);
- the **priority chip** through the one authorised mapping (`ARCHITECTURE.md` §6):
  `1→P0`, `2→P1` in `danger`; `3→P2` in `warning`; **`4` renders no chip at all** — not a
  grey "P3". This mapping lives **only** in `frontend/src/lib/priority.ts`;
- the **due badge**, `font-mono`, in `danger` **when `overdue` is true** — the flag from
  Go, never a date comparison in TypeScript;
- the **tags**;
- the **estimate**, `font-mono`;
- the **project progress bar** — see D15 below;
- the **active-timer indicator** when `timer.running`.

**D15 — the progress slot has exactly three states, and the branch is on `defined`:**

| Condition | Renders |
|---|---|
| not a project / nothing to measure and not a project | nothing |
| `progress.defined === true` | the bar, at `progress.percent`, with `done/total` in `font-mono` |
| `progress.defined === false` | **the localised "empty project" marker**, `muted`, one line, in the bar's place |

Never a percentage, never `0/0`, never an empty track: `defined === false` exists
precisely because **neither 0% nor 100% is true** (**D7**, **D9**). The marker fires in
**every** column, not only in Done — the condition is `defined`, not `status === 'done'`.

**Acceptance criteria**
- [ ] **The K3 case, named in the test:** an empty project whose derived status is `done`
      renders in Done **with the empty-project marker and no bar**. The card is not blank.
- [ ] The same empty project in **Backlog** renders the same marker — the rule has no
      column in it.
- [ ] A project with `defined === true` renders a bar at exactly `progress.percent` and
      no marker.
- [ ] Priority 4 renders **no chip**; 1 and 2 render `danger`; 3 renders `warning`.
- [ ] The overdue badge is driven **only** by the `overdue` field: a test passes
      `overdue: true` with a **future** due date and the badge is still red. That is the
      assertion that proves TypeScript is not deciding.
- [ ] `P0|P1|P2` appears in `frontend/src` only inside `lib/priority.ts` and `locales/`.
- [ ] Every number and date on the card is `font-mono`.
- [ ] The card renders correctly in **RU** with a long title and long tags; nothing
      clips, nothing overflows. A test at a fixed narrow width in both locales.
- [ ] `make guard` green — no status literal, no recomputed value.
- [ ] `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the card with full chrome and the empty-project marker (S2-14)`

---

## S2-15 — feat: the app shell and the five columns

**Two deliverables, and the first one is the reason this ticket is bigger than its
title used to be.** After S2-13 the app mounts `App.tsx`, which is still S2-09's empty
`<div>`: the store is built in `main.tsx` and handed to nobody, and `ToastList` is built
and rendered by nothing. This ticket makes `App.tsx` the **composition root** — the shell
with named regions that every later ticket mounts into — and then fills its board region
with the five columns. Read
[Composition — who mounts what](#composition--who-mounts-what) first; the rule and the
orphan check it describes start here.

**Scope (may touch):** `frontend/src/App.tsx`, `frontend/src/App.test.tsx`,
`frontend/src/App.mount.test.tsx`, `frontend/src/main.tsx`,
`frontend/src/views/Kanban.tsx`, `frontend/src/components/Column.tsx`,
`frontend/src/store/**`, the locale files, tests beside them.

`App.test.tsx` is listed explicitly because it exists: S2-10's smoke test asserts the
shell is an empty `bg-bg` div, which stops being true here. Updating it is part of this
ticket, not a scope widening.

### Part 1 — the shell

- **`App.tsx` is the composition root.** It renders the page background it already
  renders — token classes only, no hex — plus the **named regions** of the launch screen,
  in this DOM order, so that the `Tab` order S2-16 specifies falls out of the markup
  rather than out of a `tabIndex` ladder:

  | Order | Region | Filled by |
  |---|---|---|
  | 1 | header | **S2-21** — empty here, and an empty region renders nothing, not a blank bar |
  | 2 | habits strip | **S2-18** — same |
  | 3 | board | **this ticket** — `views/Kanban.tsx` |
  | 4 | overlay layer | **S2-19**, **S2-20** — same |
  | 5 | toast layer | **this ticket** — S2-13's `ToastList`, which nothing has mounted until now |

  A region with nothing in it yet is an empty slot in the source, named in a comment with
  its ticket. It is **not** a placeholder component, not localised filler text and not a
  "coming soon" box — a placeholder is a thing somebody has to remember to delete.
- **The store reaches the tree.** `main.tsx` builds the store today and drops it:
  `createAppStore(wailsClient)` is called and `<App />` is rendered without it. Give it
  one route down — a context provider, or the store passed in as a prop — chosen once,
  here, and used by every consumer. **A second module-level store singleton is a second
  spelling and is refused**; the factory exists so tests get a fresh store, and that
  property must survive this ticket.
- **The board is hydrated.** Something must call `Board()` before the columns can render.
  `main.tsx` already reads settings before first paint (S2-12) — the board read follows
  the same shape, and a failed read is a **toast and an empty board, never a blank
  window** (S2-13's rule, unchanged).
- **The orphan check lands here**, as `frontend/src/App.mount.test.tsx`, exactly as
  specified in [Composition — who mounts what](#composition--who-mounts-what): enumerate
  every non-test `.tsx` under `components/` and `views/`, walk the relative-import closure
  of `main.tsx`, fail naming anything in the first set and not the second. It is an
  acceptance criterion on **every ticket from here to S2-22**.

### Part 2 — the five columns

Requirements:

- The board renders **from `Board()`'s answer**: five `ColumnView`s, **in the order Go
  returned them**, each with its cards in the order Go returned them. The frontend does
  **not** hold a list of the five statuses, does not sort, and does not filter.
- Column headings are i18n keys **looked up by the status Go supplied** — the lookup
  table is in `locales/`, which is the one authorised place a status name may be written
  in the frontend (a label table is presentation; a status list in code is a rule).
- Per-column card count, `font-mono`.
- An empty column renders a localised empty state, not a blank rectangle.
- Aurora surfaces translucent (`bg-surface` + `backdrop-blur-glass`), Studio solid
  (`shadow-sm`) — both via the tokens, no palette conditional in the component.
- The board scrolls; the columns do not collapse below a legible width in **Russian**,
  which is ~30% wider.

**Acceptance criteria — the shell**
- [ ] **`render(<App />)` with a mocked client shows the board.** The test imports `App`
      and **does not import `Kanban`, `Column` or `Card`** — importing the component you
      are looking for turns this into a test of that component and proves nothing about
      mounting. Assert on what the user would see: the column headings from the mocked
      `Board()`.
- [ ] **`render(<App />)` shows a toast when a call is rejected.** S2-13 built `ToastList`
      and `Toast.test.tsx` renders it directly; this asserts it through `App`, which is
      the thing that was missing. Exactly one toast, translated, dismissable by keyboard.
- [ ] **The orphan check exists and is green**: `App.mount.test.tsx` enumerates
      `components/**/*.tsx` and `views/**/*.tsx`, walks the import closure of `main.tsx`,
      and reports the difference. **Proven to fire** — add a component that nothing
      imports, watch the test fail **naming that file**, delete it. Record it in the
      commit body; a check nobody has watched fail is a check nobody has checked.
- [ ] The store has exactly one construction route into the tree, and
      `createAppStore(...)` is still a factory: two tests in the same file get
      independent stores. `git grep -n "createAppStore" frontend/src` shows no
      module-level singleton.
- [ ] The four regions this ticket does not fill render **nothing at all** — no empty
      bar, no placeholder text, no reserved blank box. Asserted on the DOM.
- [ ] The built binary opens on the board, not on an empty page. Confirmed by hand with
      `make build && ./build/bin/nexus`, and reported.

**Acceptance criteria — the columns**
- [ ] With a mocked client returning five columns in a deliberately unusual order, the UI
      renders **that** order — proving the order comes from Go.
- [ ] No status string literal in `frontend/src` outside `locales/` (`make guard` #2).
- [ ] No `.sort(`, no `.filter(` on the board data in any component.
- [ ] Column headings and empty states are translated; the RU board at 1024px wide shows
      all five columns with no clipping and no horizontal scroll inside a column.
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): mount the app shell and render the five kanban columns (S2-15)`

---

## S2-16 — feat: the keyboard model — focus, navigation and keyboard move

**This is the ACCEPT criterion's backbone**, and it lands **before** drag-and-drop on
purpose. Built after a pointer API, the keyboard path becomes a bolted-on translation of
mouse gestures; built first, the pointer is the alternative route.

**Scope (may touch):** `frontend/src/lib/keyboard.ts`, `frontend/src/App.tsx`,
`frontend/src/views/Kanban.tsx`, `frontend/src/components/{Column,Card}.tsx`,
`frontend/src/store/**`, the locale files, tests beside them.

**`App.tsx` is in scope because two rows of the map below are not the board's.**
`Tab`/`Shift+Tab` moves between **regions** — habits strip ⇄ board ⇄ overlays — and
`Ctrl+N` and `Ctrl+K` must fire wherever focus happens to be, including in a region the
board does not own. Both are properties of S2-15's shell. Put the region order and the
global shortcut listener there, once, and leave the within-board arrows to the board:
**the map has one spelling either way**, in `lib/keyboard.ts`, and S2-20 reads its hints
from that module rather than restating them.

**The keyboard map — normative, and the command palette (S2-20) must match it:**

| Keys | Action |
|---|---|
| `Tab` / `Shift+Tab` | move between regions: habits strip ⇄ board ⇄ overlays |
| `ArrowLeft` / `ArrowRight` | focus the adjacent column, landing on the nearest card |
| `ArrowUp` / `ArrowDown` | focus the previous / next card in the column |
| `Home` / `End` | first / last card in the column |
| `Ctrl+Shift+ArrowRight` | **move the focused card one column right** |
| `Ctrl+Shift+ArrowLeft` | **move the focused card one column left** |
| `Ctrl+N` | quick add (S2-19) |
| `Ctrl+K` | command palette (S2-20) |
| `Escape` | close the topmost overlay |
| `Enter` on a card | **reserved for Stage 3's detail slide-over. A documented no-op.** It must not be silently swallowed — no handler at all is correct. |
| `Space` on a habit | toggle today's check (S2-18) |

Requirements:

- **Roving tabindex**: the board is one tab stop, and the arrows move focus inside it.
  Five columns × n cards as n+5 tab stops is unusable with a keyboard.
- **Focus follows the card it is on.** After a move the focus is on the **same card in
  its new column** — not reset to the top of the board, and not left on the vacated
  slot. This is the property that makes four consecutive `Ctrl+Shift+ArrowRight` presses
  a flow rather than four separate hunts.
- Moving **right from Done** and **left from Backlog** are no-ops, not errors, not
  wrap-arounds.
- A move refused by Go (a project to Doing, **D9**) raises a toast and leaves focus on
  the card.
- The target column is `column.status` from Go's board — **never** a computed "next
  status". The frontend knows the columns' **order** because Go returned them in order;
  it does not know their **names**.
- **`:focus-visible` ring in `accent`, always visible, on every focusable element**
  (`PLAN.md` §2). A focus ring that only appears on some elements is the bug that makes
  keyboard use guesswork.
- Shortcut hints rendered anywhere are `font-mono` (`PLAN.md` §3).

**Acceptance criteria**
- [ ] Arrow navigation reaches every card in every column, driven **only** by
      `user-event.keyboard`.
- [ ] `Ctrl+Shift+ArrowRight` from Backlog to Done, four presses, calls `MoveToColumn`
      four times with the statuses **Go supplied**, in order, and focus is on the same
      card at every step.
- [ ] Right from Done and left from Backlog do nothing at all — no call, no toast.
- [ ] A rejected move raises exactly one toast; focus does not move.
- [ ] Roving tabindex: exactly **one** element inside the board has `tabIndex=0` at any
      moment.
- [ ] The focus ring is visible on every focusable element — asserted by a test on the
      `:focus-visible` class/attribute, and confirmed by hand in both palettes.
- [ ] `Enter` on a card does nothing and is not intercepted.
- [ ] **Driven through `render(<App />)`, not through `<Kanban />`.** The navigation and
      move tests mount the real shell — that is what makes them the ACCEPT criterion's
      rehearsal rather than a component demo, and it is what catches a key handler
      attached to a node the shell never renders.
- [ ] `Tab` from the board reaches the next region in DOM order and `Shift+Tab` comes
      back, asserted on the shell even though the strip and the overlays are still empty
      — an empty region is skipped, not focused.
- [ ] The orphan check (`App.mount.test.tsx`) is still green.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the board keyboard model and keyboard column moves (S2-16)`

---

## S2-17 — feat: dnd-kit drag and drop, with optimistic move and rollback

**Scope (may touch):** `frontend/src/views/Kanban.tsx`,
`frontend/src/components/{Column,Card}.tsx`, `frontend/src/store/**`,
`frontend/package.json`, tests beside them.

Requirements:

- `@dnd-kit/core` + `@dnd-kit/sortable`: drag a card between columns and reorder inside
  one. Dragging a card **takes its whole subtree** — which is what Go already does
  (`MoveNode` moves `parent_id` + `sort_order`; `MoveToColumn` cascades, **D2**). The
  frontend issues the call and re-reads; it does **not** walk the subtree itself.
- **Drop-target highlight** on the hovered column, via tokens.
- **Optimistic UI with rollback**: the card moves locally on drop, the call goes out, and
  **on rejection the store is restored from Go's answer** — not from a remembered
  snapshot, which can be stale if anything else changed — and a toast is raised.
- **Reordering within a column** calls `MoveNode`; **crossing columns** calls the coupled
  `MoveToColumn` (S2-03). Two gestures, two calls, and the frontend does not decide what
  either one means beyond which gesture happened.
- **dnd-kit's `KeyboardSensor` must not fight S2-16.** Either disable it, or bind it to
  keys that do not collide with the map above. **S2-16's map is normative**; if dnd-kit's
  default `Space`-to-lift conflicts with the habit-strip `Space`, dnd-kit yields. State
  in a comment which way it was resolved.
- `prefers-reduced-motion`: drag transitions respect it.

**Acceptance criteria**
- [ ] Dropping a card on another column calls `MoveToColumn` once with that column's
      status from Go, and the card is in the new column **before** the call resolves.
- [ ] A rejected drop returns the card to its original column **and** re-reads from Go; a
      toast is raised. Test with a client that rejects.
- [ ] Reordering inside a column calls `MoveNode`, not `MoveToColumn`.
- [ ] Dragging a parent moves its subtree — asserted on the **call**, with no subtree
      walking in TypeScript (`make guard` #3).
- [ ] The keyboard map of S2-16 still works **unchanged** with dnd-kit mounted. The S2-16
      test suite passes untouched — that is the regression this criterion exists for.
- [ ] Drop-target highlight uses token colours only.
- [ ] The orphan check (`App.mount.test.tsx`) is still green. **`App.tsx` is deliberately
      not in this ticket's scope**: dnd-kit adds no region, so its `DndContext` belongs
      inside the board it wraps. If it turns out to need the shell, that is a stop-and-say
      rather than a quiet edit.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add dnd-kit drag and drop with optimistic rollback (S2-17)`

---

## S2-18 — feat: the habits strip

**Scope (may touch):** `frontend/src/components/HabitStrip.tsx`,
`frontend/src/components/HabitChip.tsx`, `frontend/src/App.tsx`,
`frontend/src/store/**`, the locale files, tests beside them.

**This ticket fills S2-15's habits-strip region** and hydrates the strip from
`HabitStrip()` the same way S2-15 hydrates the board. A strip that exists and is not in
the shell is an unfinished ticket — see
[Composition — who mounts what](#composition--who-mounts-what).

Requirements:

- Renders `HabitStrip()` from S2-04: each habit with a **checkbox** and its **streak**.
- The streak number is Go's (**D5** — consecutive scheduled RRULE occurrences, *not*
  calendar days). `font-mono`, `warning` for the flame (`design/README.md`), `success`
  for the check.
- Checking calls `CheckHabit`, unchecking `UncheckHabit`, and both re-read. **Optimistic
  is allowed here on the same terms as a move** — revert from Go's answer on error, with
  a toast.
- **Habits never appear in a Kanban column** (`PLAN.md` §4) — and the frontend does not
  need to enforce that, because `Board()` already excludes them. A filter here would be a
  second spelling of the rule. **Do not add one.**
- Whether to show habits **not** scheduled today is a UI choice over S2-04's
  `scheduledToday` flag, not a recomputation of the schedule.
- Keyboard: the strip is one tab stop, arrows move between habits, `Space` toggles
  today's check, and it is reachable from the board with `Tab` (S2-16).

**Acceptance criteria**
- [ ] A habit with streak 4 shows `4`, in `font-mono`, taken from the DTO. A test that
      supplies a streak inconsistent with the check history still renders the DTO's
      number — proving TypeScript is not counting.
- [ ] `Space` on a focused habit toggles the check and calls the right method; a
      rejection reverts and raises one toast.
- [ ] `git grep -n "habit" frontend/src` shows **no** filtering of board data by type.
- [ ] The strip is keyboard-reachable from the board and back, mouse untouched.
- [ ] **Mounted and reachable**: `render(<App />)` with a mocked client shows the habits,
      and `Tab` from the board reaches them. The test imports `App` and **does not import
      `HabitStrip`** — that restriction is the assertion.
- [ ] RU labels fit; nothing clips at 1024px.
- [ ] The orphan check (`App.mount.test.tsx`) is still green.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the habits strip with checks and streaks (S2-18)`

---

## S2-19 — feat: quick add (Ctrl+N)

**In-app only.** The frameless standalone quick-add **window** and the natural-language
parser are **Stage 4** and must not be started here. Stage 2's quick-add takes a title
and creates a node.

**Scope (may touch):** `frontend/src/components/QuickAdd.tsx`, `frontend/src/App.tsx`,
`frontend/src/store/**`, the locale files, tests beside them.

**This ticket mounts its own overlay** into S2-15's overlay layer, on S2-16's `Ctrl+N`.
An overlay that only opens in its own test is an unfinished ticket, and this one is
**step 1 of the ACCEPT script** — if it is not reachable from the running app, the
criterion cannot be demonstrated at all. See
[Composition — who mounts what](#composition--who-mounts-what).

Requirements:

- `Ctrl+N` opens an overlay with focus **in the title field**; `Escape` closes it and
  **returns focus to where it came from**; `Enter` creates and closes.
- Creates a `task` in `backlog` by default — the defaults are Go's
  (`NewNode`'s zero values: empty status means `backlog`, zero priority means 4). The
  frontend sends a title and a type; it does not fill in defaults it invented.
- A type selector (task / project / habit / note / bug), **keyboard-operable**, built
  from what Go exposes.
- A habit **requires a recurrence** — that rule is Go's, and Go will refuse. The frontend
  **surfaces the refusal in a toast**; it does not pre-validate, because pre-validating
  is writing the rule a second time.
- After creation, the board re-reads and **focus lands on the new card** — this is what
  makes step 2 of the ACCEPT script flow into step 3.
- Focus is **trapped** in the overlay while it is open.
- An empty or whitespace-only title: Go refuses (`Node.Validate`), the toast says so.

**Acceptance criteria**
- [ ] **Mounted and reachable**: in `render(<App />)`, `Ctrl+N` opens the overlay. The
      test imports `App` and **does not import `QuickAdd`**. Every criterion below is
      driven through `App` for the same reason — this is the first half of the ACCEPT
      flow and it has to work where the user is, not where the test is.
- [ ] `Ctrl+N` → type → `Enter` creates a node with the typed title and calls
      `CreateNode` **once**.
- [ ] Focus starts in the title field and, after creation, is on the **new card** in
      Backlog. Asserted with `user-event.keyboard` only.
- [ ] `Escape` closes and restores focus to the previously focused element.
- [ ] Focus is trapped while open: `Tab` cycles inside the overlay.
- [ ] Creating a habit with no recurrence raises **Go's** refusal in a toast; there is no
      recurrence validation in `frontend/src` (`make guard`).
- [ ] Every string is translated; the RU overlay does not clip.
- [ ] The orphan check (`App.mount.test.tsx`) is still green.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the in-app quick add overlay (S2-19)`

---

## S2-20 — feat: the command palette (Ctrl+K)

**Scope (may touch):** `frontend/src/components/CommandPalette.tsx`,
`frontend/src/lib/commands.ts`, `frontend/src/App.tsx`, `frontend/src/store/**`, the
locale files, tests beside them.

**This ticket mounts its own overlay** into S2-15's overlay layer, on S2-16's `Ctrl+K`,
alongside S2-19's. Step 8 of the hand ACCEPT script drives the completion leg through
this palette, so "built but not mounted" would take the second independent keyboard route
with it. See [Composition — who mounts what](#composition--who-mounts-what).

Requirements — the action set from the brief, and **only** it:

| Action | Notes |
|---|---|
| new task | opens quick-add (S2-19) |
| move to column | one entry **per column Go returned**, labelled from `locales/` |
| set priority | 1–4; the chip mapping stays in `lib/priority.ts`. **In Stage 2 scope** — see the ruling below; the *editor* in a detail panel is Stage 3 |
| start / stop timer | `TimerStart` / `TimerStop`; a node Go refuses (**D9**) produces a toast, and the palette does **not** hide the entry — hiding it would be re-deriving `DoingRefusal` in TypeScript |
| switch view | Kanban is the only view in Stage 2; the others are registered as **disabled with a localised "coming in stage N"**, or omitted. **Pick one and be consistent** — a dead entry that silently does nothing is the worse option. |
| toggle theme | through `SetTheme` (S2-05) |
| switch language | through `SetLanguage` |

- `Ctrl+K` opens with focus in the filter field; typing filters; `ArrowUp`/`ArrowDown`
  selects; `Enter` runs; `Escape` closes and restores focus.
- Filtering matches on the **translated label**, so it works in Russian. A palette that
  only answers to English words is broken in RU.
- Each entry shows its shortcut hint where it has one, in `font-mono`, and the hints
  **match S2-16's normative map** — one keyboard map, one spelling.
- Actions operate on the **currently focused card** where they need a target, and are
  disabled with a reason when there is none.
- Focus trapped while open; full a11y roles (`combobox`/`listbox`/`option`).

**Acceptance criteria**
- [ ] **Mounted and reachable**: in `render(<App />)`, `Ctrl+K` opens the palette. The
      test imports `App` and **does not import `CommandPalette`**. Every criterion below
      is driven through `App`.
- [ ] Two overlays coexist: `Ctrl+N` and `Ctrl+K` each open their own, `Escape` closes
      **the topmost** (S2-16's map), and neither is mounted over the other by accident.
- [ ] Every action in the table is reachable and executable by keyboard alone.
- [ ] "Move to column" lists exactly the columns Go returned, in Go's order, with no
      status literal in the component (`make guard` #2).
- [ ] Filtering in RU finds the Russian labels.
- [ ] Start-timer on a project produces **Go's** refusal in a toast; the entry is not
      hidden and no `DoingRefusal` logic exists in `frontend/src`.
- [ ] Shortcut hints match S2-16 exactly — a test compares the palette's hint strings
      against the keyboard map's single source.
- [ ] `Escape` restores focus to the element that had it.
- [ ] The orphan check (`App.mount.test.tsx`) is still green.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the ctrl+k command palette (S2-20)`

> **PM ruling — `set priority` is a Stage 2 palette action, and the criterion is now
> met. This is a plan correction, not a scope change by the Dev.**
>
> S2-20 shipped in `8aeb6e1` with its four `set priority` rows registered **unavailable
> with a reason**, because no binding set a priority and adding one was a Go change
> outside the ticket's Scope — while
> [What Stage 2 must NOT do](#what-stage-2-must-not-do) separately listed *"the
> tag/due/priority/estimate editors"* as out of scope. The Dev said so in the commit body
> and **asked for a ruling instead of widening**. The criterion *"every action in the
> table is reachable and executable by keyboard alone"* was therefore unmet on this row,
> through no fault of the implementation.
>
> **The brief wins over the plan.** The user's brief names the Stage 2 palette's actions
> as *"new, move to column, **set priority**, start/stop timer, switch view, toggle
> theme, switch language"*, and this ticket's own Requirements table has carried the row
> since it was written. The out-of-scope line meant the **editors** and was too coarse to
> say it; it is corrected in both places it appears.
>
> **`7af4d1d`** closes the gap: `TaskService.SetPriority` (`SetDue`'s shape),
> `App.SetPriority` taking a plain `int`, the regenerated `frontend/wailsjs` committed
> alongside, and the palette rows wired through the **store**. The 1..4 range lives
> **only** in `domain.Priority.Valid` — no second spelling in the service, in `app.go`, in
> the client or in the palette — and a refused value writes nothing and raises one toast.
> Full reasoning in
> [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2).
>
> **Status: this ticket's action-set criterion is MET.** The only rows still unavailable
> are the `switch view` rows, whose reasons are transient and point at later stages.

---

## S2-21 — feat: the appearance and language controls

S2-12 made palette, theme and accent **live**; this gives the user something to set them
with beyond the command palette.

**Scope (may touch):** `frontend/src/components/AppearanceControls.tsx`,
`frontend/src/components/AccentPicker.tsx`, `frontend/src/App.tsx`, the locale files,
tests beside them.

**This ticket fills S2-15's header region** — the last empty one. Its own requirement
below, *"the controls live somewhere unobtrusive on the launch screen … and are reachable
with `Tab`, not only through the palette"*, is unsatisfiable without the shell, which is
why `App.tsx` is in scope. `AccentPicker` is mounted by `AppearanceControls`, not by the
shell. See [Composition — who mounts what](#composition--who-mounts-what).

Requirements:

- Palette (aurora / studio), theme (dark / light), language (EN / RU) and an **accent
  picker**, all persisted through `SettingsService` and all keyboard-operable.
- The accent picker offers a **small preset set plus a free value**, and **"use the
  palette's accent"** as an explicit choice that writes `""` — the seeded default must be
  reachable, not a state you can only leave.
- **The presets must not be hex literals in `frontend/src`.** They come from tokens or
  from Go. If that forces the preset list into `design/`, it does **not** go there —
  `design/` is read-only — so it comes from Go, or the picker offers only the free value
  plus the palette default. **Decide it in the ticket and say which in the commit body.**
- Rejected values revert to the service's answer, with a toast (S2-12's rule).
- The controls live somewhere unobtrusive on the launch screen — a header or a corner —
  and are reachable with `Tab`, not only through the palette.

**Acceptance criteria**
- [ ] **Mounted and reachable**: in `render(<App />)`, `Tab` from the board reaches the
      controls without opening the command palette. The test imports `App` and **does not
      import `AppearanceControls`**.
- [ ] **All five regions of the shell are now filled**, and the orphan check has nothing
      to report: this is the last mounting ticket, so after it every component built in
      Stage 2 is reachable from `main.tsx`. State it in the commit body.
- [ ] All four settings are changeable by keyboard alone and survive a restart.
- [ ] "Use the palette's accent" writes `""` and **removes** the inline `--accent`.
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing — including the
      presets.
- [ ] An invalid free accent value is refused by Go and produces a toast; the UI reverts.
- [ ] Labels translated; the RU control strip does not clip at 1024px.
- [ ] The orphan check (`App.mount.test.tsx`) is green with an **empty** difference.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the appearance and language controls (S2-21)`

> **PM ruling — the six test files and `frontend/src/test/render.tsx` in `642e677` are
> RATIFIED. The Reviewer does not need to re-litigate it.**
>
> This ticket's Scope is `AppearanceControls.tsx`, `AccentPicker.tsx`, `App.tsx`, the
> locale files and tests beside them. Filling region 1 put **focusable elements ahead of
> the board in the shell's DOM order** — which is the entire point of the ticket — and six
> existing test files asserted *"one `Tab` lands on the board / the strip"* or *"the only
> button / textbox on the page"*. Those assertions were about a header that was empty and
> is no longer.
>
> Ratified on the same three counts as S2-13, plus one specific to it:
>
> 1. **It removed duplication rather than adding it.** The fix is a new shared
>    `tabUntil(user, arrived)` helper — **one spelling of "walk past the header" instead
>    of six hard-coded tab-stop counts**. Updating six copies of a magic number would have
>    been the worse outcome and would have broken again on the next mounting ticket.
> 2. **No assertion was weakened.** Every updated test still makes the **same claim about
>    the same element**; the queries that changed were narrowed (the toast button by name,
>    the quick-add title field scoped to its dialog, because the accent field is a textbox
>    too), not loosened.
> 3. **It was disclosed, not smuggled** — the commit body names all seven files under a
>    heading with what changed in each and why, exactly as rule 15 and the Scope
>    discipline ask.
> 4. **It was mechanical**, confined to `frontend/src`, with `design/` untouched and all
>    gates plus both non-gate targets green.
>
> Nothing here loosens the Scope rule: a widening that adds **behaviour**, or that is not
> disclosed, is still "stop and report".

---

## S2-22 — test: the no-mouse ACCEPT flow, and the RU / a11y audit

The last ticket, and the one that makes the ACCEPT criterion a thing a machine can check
rather than a thing somebody says.

**Scope (may touch):** `frontend/src/App.accept.test.tsx`, any test helper it
needs, the locale files if the audit finds a missing key, and **no production behaviour**
beyond fixing what the audit finds.

> The file sits beside `App.tsx` rather than under `views/`, because since S2-15 the
> thing under test is the **assembled screen**, not the Kanban view. `accept` is still in
> the filename, so it is still findable by name.

Requirements:

1. **The automated flow test.** One test, driving the whole ACCEPT criterion against a
   mocked client:
   `Ctrl+N` → type a title → `Enter` → four × `Ctrl+Shift+ArrowRight` → assert the card
   is in Done. It asserts the **exact sequence of client calls**: one `CreateNode`, then
   `MoveToColumn` with each of the four statuses Go supplied, in order.
   **It renders `<App />`** — the real composition root of S2-15, with the store, the
   overlays, the strip and the toast layer all mounted — and imports **no** component
   below it. The ACCEPT criterion is a property of the assembled screen; a flow test that
   mounts `<Kanban />` and passes a quick-add in by hand is testing an arrangement no user
   ever gets. This is the criterion the whole composition rule exists to protect.
2. **It can only press keys.** The test uses `user-event.keyboard` and `user-event.tab`
   and **nothing else** — no `click`, no `pointer`, no `fireEvent.mouse*`. This is
   asserted mechanically, as a `make guard` check over the test file, so the proof cannot
   quietly rot.
3. **The timer assertion**: after the move into Doing the card shows the running-timer
   indicator, and after the move into Done it does not (**D13** §3). That is C1 visible
   from the UI.
4. **The RU audit.** Every screen at 1024×768 in Russian: board, strip, quick-add,
   palette, appearance controls, toast. Nothing clipped, nothing overflowing, no
   horizontal scroll. Snapshot-tested at a fixed width where it can be, inspected by hand
   where it cannot, and **reported honestly either way**.
5. **The a11y audit.** Every interactive element keyboard-reachable; `:focus-visible` in
   `accent` visible on every one of them; `prefers-reduced-motion` respected; sensible
   roles and labels on the overlays.

   **On the Aurora drift, and say it this way in the report (D16, K5):** what is audited
   is that `tokens.css` kills transitions under `prefers-reduced-motion` and that
   `auroraDriftEnabled()` sets `data-drift="off"`. **Nothing draws the drift** — the gate
   shipped in S2-12, the visual is deliberately out of Stage 2 because `design/` specifies
   no drift to build and inventing one is forbidden. Do **not** report having watched a
   drift pause; report that the gate is correct and that there is nothing behind it yet.
6. **The mounting audit.** `App.mount.test.tsx` reports an empty difference, and the
   report names, one by one, every component Stage 2 built and where in the running app
   it is reached. This is the last chance to catch a piece that exists and is not on
   screen, and it is the defect the composition rule was written for.

**Acceptance criteria**
- [ ] The flow test passes and is named so it is findable (`accept`, in the filename).
- [ ] **The flow test renders `<App />`** and imports no component beneath it —
      `git grep -n "^import" frontend/src/App.accept.test.tsx` shows `App`, the
      test helpers and nothing from `components/` or `views/`.
- [ ] `make guard` fails if a mouse event is introduced into the accept test — proven by
      introducing one, watching it fail, and reverting. **Record it in the commit body.**
- [ ] The timer indicator appears on Doing and is gone on Done.
- [ ] The RU audit is done and reported, screen by screen, with any fix committed here.
- [ ] Every interactive element is keyboard-reachable and shows the accent focus ring.
- [ ] **`App.mount.test.tsx` reports an empty difference**, and the mounting audit is
      reported component by component.
- [ ] The Aurora drift is reported as **gated and not drawn** (D16 / K5), not as
      "verified pausing".
- [ ] The **hand** script below is executed on the real binary and reported step by step.
- [ ] `make guard`, `make front-test`, `make cover`, `make check` all green.

**Commit:** `test(frontend): add the no-mouse accept flow and the ru and a11y audit (S2-22)`

> **PM ruling — the `Makefile` change in `58552bf` is RATIFIED, and it is a Scope-list
> omission rather than a widening. The Reviewer does not need to re-litigate it.**
>
> This ticket's Scope names `App.accept.test.tsx`, its helpers and the locale files. It
> does **not** name the `Makefile` — but requirement 2 and acceptance criterion 3 say the
> no-mouse proof must be *"asserted mechanically, as a `make guard` check over the test
> file"*. **The rule was required by the ticket to live in `make guard`**, so the Scope
> list was incomplete; the Dev had no way to satisfy the criterion inside it, disclosed
> the change under a heading, and kept it minimal.
>
> Check 6 is **scoped to that one file**, greps the three words a pointing device is spelt
> with, case-insensitively, in code or in a comment, and **fails if the file is missing or
> empty** — because otherwise the cheapest way to pass the check would be to delete the
> thing it checks. The rest of the suite stays free to use a pointer where a pointer is
> what is under test (S2-17's drag), and check 6 says so.
>
> **The five gates are still five.** `make guard` is a non-gate target by construction
> (see [Two non-gate make targets](#two-non-gate-make-targets--the-five-gates-stay-five))
> and `make check` is untouched. The negative control the criterion asks for was run both
> ways: a `user.click(...)` added to the accept test made check 6 fail naming the file and
> line, and the file removed entirely made it fail rather than pass quietly.

---

## ACCEPT — how the no-mouse flow is demonstrated

> **Create a task, move it across all five columns, and complete it — with no mouse.**

**Demonstrated twice, deliberately.** The automated half proves the interaction is
keyboard-only in a way that survives the next change; the hand half proves the thing
actually works in a real WebKit window, which jsdom cannot tell anyone.

### Half 1 — mechanical, and re-run on every commit

`frontend/src/App.accept.test.tsx` (S2-22), run by `make front-test`.

It renders **`<App />`** — the whole assembled screen of S2-15 onwards, not a hand-built
arrangement of components — and then presses keys and nothing else. `make guard` asserts
the file contains no `click`, no `pointer`, no `fireEvent.mouse*` — so "no mouse" is not a
claim in a commit message, it is a check that fails.

It asserts the exact call sequence: `CreateNode` once, then `MoveToColumn` with each of
the four statuses **as Go supplied them**, in order, then the card in Done.

### Half 2 — by hand, on the real binary, mouse untouched

The Reviewer runs this and **reports each step**. Build and launch:

```sh
make build && ./build/bin/nexus
```

Then push the mouse out of reach and do not touch it. Steps 1–10 are keyboard only:

| # | Keys | Expected |
|---|---|---|
| 1 | `Ctrl+N` | Quick-add opens; the caret is in the title field; the accent focus ring is visible |
| 2 | type `Keyboard accept`, `Enter` | Overlay closes; a new card appears in **Backlog**; **focus is on that card** |
| 3 | `Ctrl+Shift+→` | Card is in **This week**; the due badge reads the **upcoming Friday** (D1, D8); focus followed the card |
| 4 | `Ctrl+Shift+→` | Card is in **Today**; the due badge reads **today** |
| 5 | `Ctrl+Shift+→` | Card is in **Doing**; the **running-timer indicator** appears on it — and on no other card (**C1**, **D13**) |
| 6 | `Ctrl+Shift+→` | Card is in **Done**; the timer indicator is **gone** (**D13** §3) |
| 7 | `Ctrl+Shift+←` ×4 | Back to **Backlog**, one column per press; the due date is **cleared**, because the column moves set `due_source = auto` (D1, D8) |
| 8 | `Ctrl+K`, type `done`, `Enter` on "Move to Done" | Card is in **Done** again — the completion leg driven through the palette, a second independent keyboard route |
| 9 | `Ctrl+K`, `theme`, `Enter` · then `Ctrl+K`, `language`, `Enter` | Theme flips with no flash; UI switches to Russian; **nothing clips** and all five columns are still readable |
| 10 | quit, relaunch | The card is still in **Done**; theme and language persisted; **no flash of the wrong background** on the first frame (**D12**) |

**The mouse was not touched.** The Reviewer states this explicitly. The mechanical proof
is half 1; this half is the real window, which is exactly what half 1 cannot see.

### What is verified by automation vs by hand — stated honestly

| Verified by automation | Verified by hand |
|---|---|
| the whole keyboard flow, and that it uses no mouse event | that a real WebKit window opens and is usable |
| the exact sequence and arguments of the Go calls | the **window background** and the absence of a startup flash (**D12** / K1) — GTK-side, invisible to jsdom |
| the timer indicator appearing on Doing and gone on Done | Russian text not clipping at a real 1024×768 |
| that every Stage 2 component is mounted and reached from `main.tsx` (`App.mount.test.tsx`) | that the focus ring is genuinely visible against both palettes |
| that `data-drift` is `off` under `prefers-reduced-motion` and `on` otherwise | that the launch screen is actually assembled — header, strip, board, overlays, toasts — and not five tests that pass in isolation |
| Go-side atomicity, the single-active timer, the cascade rule (**S2-03**) | |
| `en.json`/`ru.json` key parity | |
| no hex literal, no status literal, no recomputed rule (`make guard`) | |

Nothing in the left column is claimed as hand-verified, and nothing in the right column
is claimed as automated.

**Verified by neither, and deliberately so:** that Aurora's background drift pauses. The
*gate* is automated, above; the **drift itself is not drawn in Stage 2** — see **D16** and
**K5**. It was on the hand list in the first draft of this plan, which would have had the
Reviewer confirm the behaviour of something nothing renders. Removed rather than left to
be discovered at review.

### Owed to a hand pass — NOT VERIFIED

**Four checks are outstanding, and no document in this repository may describe them as
done.** They are not outstanding through negligence: **there is no display on this
machine**, and `xvfb-run`, `scrot`, `import` and `grim` are **all absent**, so the
Reviewer cannot run them either. They are owed to a human at a real keyboard in front of
a real window, and they are the precise contents of the right-hand column above.

1. **Steps 1–10 of [half 2](#half-2--by-hand-on-the-real-binary-mouse-untouched)** on
   `./build/bin/nexus`, mouse untouched, reported step by step. **Including the due-badge
   assertions at steps 3, 4 and 7** — the upcoming Friday, today, and cleared on the way
   back. The automated flow's fake client **does not model D8's due rewrite**, so those
   three claims have **no mechanical backing at all**; nothing in the suite stands behind
   them. This is DONE criterion 5's hand half.
2. **No flash of the wrong background on the first frame, in *both* themes**
   (**D12** / **K1**). GTK-side and pre-paint; jsdom cannot see it. This is DONE
   criterion 8's last clause, and it is the one still open on it.
3. **Russian not clipping at a real 1024×768.** jsdom has **no layout engine** — every
   `offsetWidth` is 0 and nothing ever overflows — so the suite asserts the **mechanisms**
   that prevent clipping (`flex-wrap` on the header, the strip, the type row and the
   palette rows; `break-words`; no `truncate`; no `nowrap`; `min-w`/`basis-0` on the
   columns; `overflow-x` on the board and never inside a column) and **never the absence
   of clipping**. The mechanisms being right is a different claim from the text fitting.
   This is DONE criterion 13's last clause.
4. **The `:focus-visible` accent ring actually painted, in both palettes.** jsdom
   evaluates `:focus-visible` as **false** for programmatic focus, and the test host runs
   with `css: false`, so what is asserted mechanically is **reachability and
   focusability**, element by element — not a painted ring. This is the right-hand
   column's fourth row.

**And, held to the same standard, the Aurora drift.** The gate is implemented and was
audited **both ways** in S2-22 — `data-drift` is `on` without reduced motion and `off`
with it — but **nothing draws a drift, so nobody has watched one pause**, and this is not
owed to a human either, because there is nothing to look at. Do not let any document imply
the visual exists (**D16**, **K5**, DONE criterion 21).

---

## Stage 2 — DONE criteria

Stage 2 closes only when **all** of these hold, verified by the Reviewer. Per
`PLAN.md` §5 a stage cannot close without a **PASS**.

**The boxes below are deliberately still unticked.** Every ticket is implemented and
committed, but these are the **Reviewer's** to verify, not the PM's to assert — ticking
them here would be exactly the "marked done, not verified" failure the last line of this
section forbids. What the PM has recorded is the **measurement** (`make check` green,
`make cover` 100.0% / 93.8%, `make front-test` 22 files / 251 tests, `make guard` clean
over six checks with an empty `GUARD_ALLOW_RE`) and, separately, the four things
**nobody on this machine can check** — see
[Owed to a hand pass](#owed-to-a-hand-pass--not-verified), which carries the hand half of
criterion 5, the last clause of criterion 8 and the last clause of criterion 13.

1. [ ] All twenty-two tickets are committed, one conventional commit each, in order,
   `S2-01` … `S2-22`, authored solely by `Ismat <mukhamejanov.ismat@gmail.com>` with **no
   AI author and no co-author trailer**. **One further commit, `7af4d1d`, closes the
   `set priority` gap under S2-20** and is part of the stage — see
   [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2).
2. [ ] `make check` is green — **all five gates, unchanged in number and definition**.
3. [ ] `make cover` is green: `internal/domain` ≥ 90.0% **and** `internal/service` ≥ 90.0%,
   measured per package after `go clean -testcache`.
4. [ ] `make front-test` is green and `make guard` exits 0.
5. [ ] **The ACCEPT criterion is met, both halves** — the automated flow test passes, and
   the ten-step hand script has been executed on the real binary and reported step by
   step.
6. [ ] **C1 is wired and cannot be unwired silently**: moving a card to Doing opens
   exactly one `time_entry`; a cascade opens none; moving out of Doing closes it; a
   project opens none. The negative control is recorded in S2-03's commit body.
7. [ ] **C2 is fixed**: the derivation index is built once per `Board()`, with a committed
   benchmark and before/after numbers in the commit body.
8. [ ] **C3/K1 is fixed (D12)**: `LC_NUMERIC=C` is forced, the background is seeded from
   settings, `main.go` contains **no hex literal**, and the no-flash behaviour is
   confirmed by hand in **both** themes.
9. [ ] **C4/K2 is fixed (D14)** and **C5/K3 is fixed (D15)**, each with the named test.
10. [ ] **No rule has a second spelling.** The Stage 1 inventory — `HasColumn`, `HasDue`,
    `DoingRefusal`/`CanBeDoing`, `countsAsWork`, habit-requires-recurrence,
    `DefaultActivity` — is still one definition each, and **no `frontend/src` file
    computes a status, a column eligibility, progress, a streak, an overdue flag or a due
    date**. `make guard` passes and the Reviewer has read it rather than trusting it.
11. [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing.
12. [ ] `design/` is byte-identical to its Stage 1 state:
    `git diff <stage-1-tag-or-sha> -- design/` is empty.
13. [ ] Both locale files are **complete and key-identical**; every user-visible string is
    a key; the Russian UI does not clip at 1024×768.
14. [ ] Every `App` method returns `(T, error)` (reflection test), every rejection reaches
    a toast, and `frontend/src/lib/client.ts` is the only importer of `wailsjs`.
15. [ ] **Local-only holds**: no network call, no CDN, no font fetch.
    `git grep -nE 'https?://' frontend/src` returns nothing, and every added dependency is
    bundled from `node_modules`.
16. [ ] `CGO_ENABLED=0 go build ./...` succeeds and `go list -deps ./... | grep -i mattn`
    prints nothing.
17. [ ] `internal/domain` is still pure — `TestDomainIsPure` and `TestDomainReadsNoClock`
    pass **unmodified**.
18. [ ] `git status --porcelain` is empty after
    `make check && make cover && make front-test && make guard` — including the generated
    `frontend/wailsjs`, which is committed and not left dirty.
19. [ ] `0001_settings.sql` and every Stage 1 migration are byte-identical to their
    closed state.
20. [ ] **Every component Stage 2 built is reachable in the running app.**
    `frontend/src/App.mount.test.tsx` reports an **empty** difference between the
    components under `frontend/src/{components,views}` and the import closure of
    `frontend/src/main.tsx`; the header, habits strip, board, overlay and toast regions of
    the shell are all filled; and the ACCEPT flow test drives `<App />`. A stage whose
    every ticket passed and whose app opens blank has not delivered a launch screen —
    this is the criterion that says so.
21. [ ] **Nothing is claimed that was not verified, specifically on the Aurora drift**:
    the report says the drift is **gated and not drawn** (**D16**, **K5**) and nobody
    claims to have watched it pause.
22. [ ] **The Reviewer returns PASS.**

Nothing may be marked done that was not actually verified. Stage 1 closed on its fourth
review; a first-round FAIL here is a normal outcome, not a failure of the process.

**Out of Stage 2 scope, and it must stay out**: the detail slide-over, the Markdown
editor, inline subtasks, the tag/due/priority/estimate **editors** — the detail-panel
fields, **not** the command palette's `set priority` action, which is **in** Stage 2 by
[the plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2) and
shipped in `7af4d1d` — the RRULE editor,
attachments, the type switcher, the tree view, the archive view, the search **screen**,
the standalone quick-add **window** and the NL parser, focus mode, the tray, autostart,
D-Bus sleep/lock, backup, export, the PMP timelog screen, the calendar, stats and Gantt —
and **Aurora's background drift**, whose reduced-motion gate ships and whose visual does
not, by **D16**. That last one is the only item on this list that is out of scope because
the design export does not specify it rather than because it belongs to a later stage;
it returns when a spec exists, and **it is not to be invented** (**K5**).
