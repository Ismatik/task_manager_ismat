# Nexus — TASKS

**Stage 0 — Scaffold.** Tickets only. Nothing below is implemented yet.

Decisions referenced as **D1–D7 / E1–E3** live in [`PLAN.md` §7](./PLAN.md).
Do not plan or implement beyond Stage 0.

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

## ⚠️ Blocker the user must clear (sudo required)

Stage 0 **cannot fully pass** until these are installed. They are not present on this
machine right now — `pkg-config` itself is missing, so no GTK/WebKit probe can even run:

```sh
sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev
```

Consequences until that is done:

- Gate 5 (`wails build -tags webkit2_41`) **will fail**. That is expected, not a bug.
- The Stage 0 ACCEPT criteria "app opens an empty window" and "second launch focuses
  the first" **cannot be demonstrated**.
- Gates 1–4 (`go test`, `go vet`, `npm run lint`, `npm run typecheck`) are unaffected
  and must be kept green throughout.

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
| 1 | `go test ./...` | repo root |
| 2 | `go vet ./...` | repo root |
| 3 | `npm run lint` | `frontend/` |
| 4 | `npm run typecheck` | `frontend/` |
| 5 | `wails build -tags webkit2_41` | repo root |

**`make check` is introduced by ticket [S0-11](#s0-11--build-make-check-the-five-gates).**
Earlier tickets add the individual pieces; S0-11 is the ticket that wires all five
into one gate and makes "green" a single, checkable claim.

---

## Ticket index

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
- [ ] `git log` shows exactly one commit on `main`.
- [ ] `git status --porcelain` is empty.
- [ ] `PLAN.md`, `TASKS.md` and all four `design/` files are tracked.
- [ ] No `.go`, `.ts`, `.tsx`, `.css`, `.js` or config file other than `.gitignore`
      was added.
- [ ] Commit author is `Ismat <mukhamejanov.ismat@gmail.com>`; no co-author trailer.

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
- [ ] `ARCHITECTURE.md` exists at repo root and lists all eight paths above.
- [ ] States the dependency direction and the purity rule for `internal/domain`.
- [ ] States the no-cgo rule and the `-tags webkit2_41` requirement.
- [ ] Notes that `main.go` lives at the **repo root** (Wails convention), not
      `cmd/nexus/`.
- [ ] No code files added.

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
- [ ] `go build ./...` succeeds (this does **not** need GTK/WebKit).
- [ ] `go vet ./...` is clean.
- [ ] `cd frontend && npm install && npm run build` succeeds.
- [ ] `go.mod` declares `go 1.26` and module `nexus`.
- [ ] `frontend/wailsjs/` is tracked; `frontend/node_modules/` is **not** tracked.
- [ ] `main.go` and `app.go` sit at the repo root; `ARCHITECTURE.md`, `PLAN.md`,
      `TASKS.md`, `design/` are untouched.
- [ ] `git grep -n "webkit2gtk-4.0"` returns nothing.

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
- [ ] `make` with no argument prints help and exits 0.
- [ ] `make test` runs `go test ./...` and passes.
- [ ] `make clean` removes `build/bin` and `frontend/dist` and is idempotent.
- [ ] `grep -c 'webkit2_41' Makefile` ≥ 1, and **no** wails invocation in the file
      lacks `$(TAGS)`.
- [ ] `make build` is *expected to fail* on this machine until the apt packages land;
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
- [ ] `frontend/tailwind.config.js` consumes `design/tailwind.config.js` rather than
      restating the token map.
- [ ] `design/tokens.css` is imported by the entry CSS and is **byte-identical** to
      its committed state (`git diff --exit-code design/`).
- [ ] `npm run build` succeeds and the emitted CSS contains the token custom
      properties.
- [ ] A throwaway element using `bg-surface text-ink rounded-md font-mono
      duration-fast` resolves to the token values (verify in the built CSS).
- [ ] `index.html` has `data-palette="aurora"` and `class="dark"` on `<html>`.
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing.
- [ ] Tailwind resolves to a `3.x` version in `package-lock.json`.

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
- [ ] `cd frontend && npm run lint` exits 0 with zero warnings.
- [ ] `cd frontend && npm run typecheck` exits 0.
- [ ] `npx prettier --check .` exits 0.
- [ ] `tsconfig.json` has `"strict": true`.
- [ ] Deliberately introducing an unused variable makes `npm run lint` exit non-zero
      (demonstrate, then revert).
- [ ] `npm run build` still succeeds.

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
- [ ] All five packages exist and `go build ./...` succeeds.
- [ ] `go test ./...` passes and reports **no** `[no test files]` for
      `internal/domain`.
- [ ] `go vet ./...` is clean.
- [ ] `internal/domain` imports nothing from `store`, `service`, `parse`, `platform`,
      `database/sql`, `os` or `net` — assert this with a test or a `go list` check
      documented in the commit body.
- [ ] `DueSource` has exactly the values `manual` and `auto`; `Activity` has exactly
      the seven values from D4.
- [ ] No status derivation, no streak logic, no column↔due logic (Stage 1).

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
- [ ] `CGO_ENABLED=0 go build ./...` succeeds.
- [ ] `go list -deps ./... | grep -i mattn` returns nothing (no cgo sqlite driver
      crept in).
- [ ] Test: migrating a fresh temp DB succeeds; migrating **again** is a no-op and
      applies zero migrations.
- [ ] Test: `foreign_keys` reads back as `1` and `journal_mode` as `wal`.
- [ ] Test: a deliberately failing migration rolls back and leaves
      `schema_migrations` unchanged.
- [ ] Tests use `t.TempDir()`; no `*.db` file is left in the working tree
      (`git status --porcelain` empty after `go test ./...`).
- [ ] `go test ./...` and `go vet ./...` pass.

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
- [ ] Migration applies cleanly on a fresh DB via the S0-08 runner.
- [ ] Test: round-trip `Set` → `Get`; `Set` twice on the same key upserts rather than
      erroring or duplicating.
- [ ] Test: `Get` on a missing key returns `found == false` and **no** error.
- [ ] Test: seeding defaults twice leaves exactly four rows; a user-modified
      `palette=studio` survives a second seed.
- [ ] `go test ./...` and `go vet ./...` pass; no stray `*.db` in the tree.

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
- [ ] Test (no GUI needed): first `Acquire` succeeds; a second `Acquire` in the same
      test reports "already running" and the message is received by the first.
- [ ] Test: a stale socket file with no listener is reclaimed, not fatal.
- [ ] Test: `--quick` and bare launch produce distinguishable messages on the wire.
- [ ] Test: the socket file is removed after the primary shuts down.
- [ ] Tests use a temp runtime dir and leave nothing behind.
- [ ] `go test ./...`, `go vet ./...` pass.
- [ ] **Manual (blocked on apt, see the blocker note):** `make build && ./build/bin/nexus`
      opens an empty window; a second `./build/bin/nexus` exits immediately and the
      first window is focused.
- [ ] No quick-add UI was built (Stage 4 scope).

**Commit:** `feat(platform): add single-instance lock with --quick ipc message (S0-10)`

---

## S0-11 — build: `make check` — the five gates

**This is the ticket that introduces `make check`.**

**Scope (may touch):** `Makefile`, `frontend/package.json` (only if a script name
needs aligning).

`make check` runs exactly these five, in order, **failing fast on the first failure**:

| # | Command | Run from |
|---|---|---|
| 1 | `go test ./...` | repo root |
| 2 | `go vet ./...` | repo root |
| 3 | `npm run lint` | `frontend/` |
| 4 | `npm run typecheck` | `frontend/` |
| 5 | `wails build -tags webkit2_41` | repo root |

Requirements:

- Individual targets `test`, `vet`, `lint`, `typecheck`, `build` exist and are
  runnable alone; `check` depends on them in that order.
- Gate 5 goes through `$(WAILS)` and `$(TAGS)` from S0-04 — **never** a bare `wails`.
- If `frontend/node_modules` is missing, gates 3–4 must `npm ci` first rather than
  fail confusingly.
- On gate-5 failure from a missing system library, print the
  `sudo apt install pkg-config libgtk-3-dev libwebkit2gtk-4.1-dev` hint.
- Ordering matters: the cheap, currently-passable gates run **before** the one that is
  environmentally blocked, so a Dev gets useful signal even before the apt install.

**Acceptance criteria**
- [ ] `make check` exists and runs all five commands in the stated order.
- [ ] Gates 1–4 pass on this machine **today**.
- [ ] Gate 5 is the only failing gate, and only for the missing-apt-package reason.
- [ ] Making gate 1 fail on purpose stops the run before gate 2 (demonstrate, then
      revert).
- [ ] `make check` re-run twice in a row gives the same result (no hidden state).
- [ ] `git grep -n 'wails ' Makefile` shows every invocation carrying `$(TAGS)`.

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
- [ ] `CLAUDE.md` exists at repo root and covers every bullet above.
- [ ] Lists the five gate commands verbatim.
- [ ] Mentions `webkit2_41` and the no-cgo rule explicitly.
- [ ] Does not duplicate `PLAN.md` wholesale — it links instead.
- [ ] No code files touched.

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
- [ ] File exists at `design/pmp-timelog-format.md`.
- [ ] Lists all seven activities **and** all four type-defaults.
- [ ] States the 15-minute rounding rule with a worked example including a boundary
      case.
- [ ] States that Комментарий is generated then user-edited.
- [ ] States that hours come only from measured `time_entries`.
- [ ] Shows the exact output block.
- [ ] `git diff --exit-code design/SKILL.md design/tokens.css
      design/tailwind.config.js design/README.md` is clean.

**Commit:** `docs: add deterministic pmp timelog format spec (S0-13)`

---

## Stage 0 — DONE criteria

Stage 0 closes only when **all** of these hold:

1. All thirteen tickets are committed, one commit each, in order.
2. `make check` is green — **all five gates**, including
   `wails build -tags webkit2_41`.
3. `./build/bin/nexus` opens an empty window.
4. A second `./build/bin/nexus` exits immediately and focuses the first window.
5. `git status --porcelain` is empty after a full `make check`.
6. `CLAUDE.md`, `ARCHITECTURE.md` and `design/pmp-timelog-format.md` all exist.
7. `git log` shows no AI author or co-author on any commit.

Criteria 2–4 are **blocked on the apt install**. Until then Stage 0 is
**PASS-pending-environment**, not PASS. Nothing gets marked done that was not
actually verified.

**Out of Stage 0 scope** (do not start): the `nodes` schema, any domain rule (status
derivation, column↔due, streaks, timer), FTS5, Kanban, quick-add UI, tray, i18n string
extraction beyond the locale file stubs.
