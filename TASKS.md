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
- **[Stage 2 — Kanban + Habits strip](#stage-2--kanban--habits-strip): DONE. Reviewer
  returned PASS on the second review, at `8818db4`.** Twenty-two tickets, **S2-01 …
  S2-22, all committed**, plus the gap-closing `7af4d1d` and round 1's two fixes
  `ae6befd` / `4b1af9c`. `make check` green (**all five gates**), `make cover` **100.0% /
  93.8%**, `make front-test` **22 files / 247 tests / 0 failures**, `make guard` **six of
  six** with `GUARD_ALLOW_RE` **empty**; **all twenty-two DONE criteria met**, and the
  **ACCEPT criterion met** by `frontend/src/App.accept.test.tsx` driving the assembled
  `<App />`. **Review round 1 returned FAIL on two blocking issues; both were fixed and
  then verified closed by reverting each fix and watching the specific test fail** — that
  history is kept in full below and is the most useful thing in this section. See
  [Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed),
  [Review round 2](#review-round-2--pass-at-8818db4) and
  [Stage 2 — DONE criteria](#stage-2--done-criteria--all-met-pass). **Five checks remain
  [owed to a hand pass](#owed-to-a-hand-pass--not-verified)** — the PASS did not cover
  them and they are [carried into Stage 3](#carried-into-stage-3).
  All five items under [Carried into Stage 2](#carried-into-stage-2) are absorbed into named
  tickets — **C1 → S2-03, C2 → S2-01, C3 → S2-08, C4 → S2-06, C5 → S2-14** — and the
  fifth, *one rule one spelling*, is enforced by `make guard` (**S2-10**) and re-checked
  on every frontend ticket. Two planning defects were **corrected in place** mid-stage,
  both reported by the Dev rather than silently widened: nothing owned `App.tsx` or
  `main.tsx` (see [Composition — who mounts what](#composition--who-mounts-what)), and
  the out-of-scope line contradicted the brief on *set priority* (see
  [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2)).
- **[Stage 3 — Defect remediation, then Detail + Tree + Search/Archive](#stage-3--defect-remediation-then-detail--tree--searcharchive):
  PLANNED, not started.** Twenty-nine tickets, **S3-01 … S3-29**, in **two blocks**.
  **Block A, [S3-01 … S3-09](#block-a--the-blocking-defect-block-s3-01--s3-09), is
  blocking**: the user ran the app by hand on real hardware after Stage 2 closed and found
  **nine defects** — recorded as **K6 – K14** in [`PLAN.md` §7](./PLAN.md) and ruled as
  **D18 – D26** — and their condition for proceeding was *"if these details are resolved in
  the next steps, we are good to go."* **No feature ticket starts until block A is green.**
  Block B, S3-10 … S3-29, is the feature stage: the detail slide-over, the tree view,
  search and archive, plus **C6** (closed by **D26**) and `README.md`. What Stage 3
  inherits is listed under [Carried into Stage 3](#carried-into-stage-3) — every item there
  is now absorbed into a named ticket or an explicit refusal. **Both open questions have
  been answered by the user and are CLOSED**: **OQ1 → Inter** (unblocking S3-06) and
  **OQ2 → the middle option**, which puts `COUNT` and `UNTIL` into the recurrence language
  (**D27**, **D28**) and adds **S3-19** — both in [`PLAN.md` §7](./PLAN.md).

Decisions referenced as **D1–D28 / E1–E3** and known issues **K1–K14** live in
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
the consequence. **D17** was added *after* Stage 2 was implemented, out of the Reviewer's
first round: **Go decides the habit check day and the frontend never names a date Go will
act on** — and, as part of the same entry, that `make guard`'s checks 3a/3b are
**name-based heuristics**, so a green guard is not proof that the frontend computes
nothing. **Nine more known issues, K6 – K14, came out of the user's hand pass on real
hardware after Stage 2 closed**, and were ruled on while Stage 3 was planned: **D18** (the
dragged card is drawn in a `DragOverlay`), **D19** (blur the few large surfaces, not every
small one), **D20** (no localised string may refuse to shrink), **D21** (the window minimum
is derived from the layout floor, never typed), **D22** (one height chain; the board
scrolls, the document does not; focus never scrolls), **D23** — **the user's** — (vendor a
Cyrillic face under the same two family names via `unicode-range`), **D24** (toasts are
capped, de-duplicated and self-dismissing), **D25** (a refusal is not a failure, and every
toast says what failed) and **D26** (Go publishes every enum set — this is what closes
**C6**). **Every known issue now has a decision; none is open.** **Both *questions* have
now been answered by the user and are CLOSED**, and the answers are **D27** — **the
user's** — (`COUNT` and `UNTIL` enter the recurrence language, everything else stays
rejected at parse time; this is **OQ2**'s middle option and **not** the PM's
recommendation) and **D28** (the PM ruling that spells out what a terminated series means
for **D5** streaks, the habit strip and *"scheduled today"*). **OQ1** was answered
**Inter**, which fixes **D23**'s one open half and unblocks **S3-06**. Both questions are
kept in `PLAN.md` §7, marked closed with their answers rather than deleted.

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

- [x] Moving a node to `doing` opens exactly one `time_entry` for it, in the same
      transaction semantics as the move.
- [x] The single-active-timer invariant still holds: moving a second card to `doing`
      closes the first card's entry rather than opening a concurrent one.
- [x] A move refused by `domain.DoingRefusal` (a `project`) opens **no** entry.

**MET by S2-03 (D13), verified at Stage 2's PASS — DONE criterion 6.**

### C2 — `Board()` is O(n²·log n); fix it before the board is on screen

`snapshot.view` (`internal/service/read.go`) calls `domain.DeriveStatus` and
`domain.ComputeProgress` with the **whole node slice**, and each of those rebuilds
`indexByID` from scratch — so the index is rebuilt **once per node** and assembling the
board is O(n²·log n). At the ~200 nodes a real personal board holds this is invisible;
at 10k it is not. Fix it **before** the Kanban renders, not after — once the UI is live
the regression is a user-visible stutter and the fix competes with feature work.

- [x] The index is built once per `Board()` call, not once per node.
- [x] `internal/domain` stays pure and keeps its single spelling of each rule — the fix
      is an indexing change, not a second derivation path.
- [x] A benchmark or a sized test documents the complexity change.

**MET by S2-01, verified at Stage 2's PASS — DONE criterion 7.**

### C3 — K1: `BackgroundColour` never reaches GTK under `LC_NUMERIC=ru_RU.UTF-8`

Upstream Wails formats the window background's alpha with the **process** locale at
`window.c:205`, emitting `rgba(27, 38, 54, 0,0)`; GTK's CSS parser rejects the comma and
the background is **silently** never applied. Three options are recorded in `PLAN.md`
§7 — force `LC_NUMERIC=C` before `wails.Run`, leave the window transparent and let the
frontend paint it, or patch and pin a fork — and **none is chosen**. Choose it **with
the palette work**, where the window background first has to match a token.

- [x] One of the three options is chosen, recorded as a decision, and implemented.
- [ ] The dark theme shows no flash of a wrong background on this machine.

**Half met.** **D12** chose it and **S2-08** implemented it — DONE criterion 8. The
second box is **still open**: it needs a display, it was not covered by Stage 2's PASS,
and it is [carried into Stage 3](#carried-into-stage-3) as owed item 2.

### C4 — K2: a leaf project's stale *stored* status has teeth

Since **D11** a leaf project's stored status decides whether it counts as done in its
parent's denominator. **Archiving the last real child of a project that an earlier
cascade wrote `done`** leaves that project counted as a **done** unit on a status nobody
set deliberately. The state is self-consistent — column and bar agree — so it is not a
contradiction, which is exactly why it will not announce itself.

- [x] Stage 2 rules on whether a project's stored status is re-inspected when its last
      child is archived, and the ruling is recorded as a decision in `PLAN.md` §7.

**MET — the ruling is D14, implemented in S2-06 with its named test; DONE criterion 9.**

### C5 — K3: an empty project stored `done` renders in Done with no bar

Its own progress is undefined (**D11**, part 1) so no bar is drawn, while its status
puts the card in the Done column. **It is the one place a finished card shows nothing.**
This is a card-chrome decision.

- [x] The card-chrome work decides what a done-but-unmeasurable card renders, and the
      decision is recorded.

**MET — the ruling is D15, implemented in S2-14 with its named test; DONE criterion 9.**

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

- [x] No Stage 2 ticket adds a second implementation of any row in the table above.
- [x] No `frontend/src` file computes a status, progress, streak, overdue flag or column
      eligibility.

**MET — DONE criterion 10.** Round 1 found one violation, `wireDate`, and it was fixed
in `ae6befd` and ruled on as **D17**. It was found **by reading the diff**, not by
`make guard`, whose checks 3a/3b are name-based heuristics: **a green guard is not a
proof** — see [Carried into Stage 3](#carried-into-stage-3).

---

## Stage 2 — Kanban + Habits strip

**Status: DONE — CLOSED on a Reviewer PASS, returned on the SECOND review and verified
at `8818db4`.** S2-01 … S2-22 all committed, plus the gap-closing `7af4d1d` and round
1's two fixes `ae6befd` and `4b1af9c`.
Twenty-two tickets, **S2-01 … S2-22**, one conventional commit each. This is the
**launch screen**: the first stage whose output a user can look at.

Per `PLAN.md` §5 a stage closes only on a Reviewer **PASS**. Round 1 returned **FAIL** on
two blocking issues; both were fixed, and round 2 returned **PASS** — see
[Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed) and
[Review round 2](#review-round-2--pass-at-8818db4). **All twenty-two
[DONE criteria](#stage-2--done-criteria--all-met-pass) are met and ticked**, by the Reviewer's
verification and not by PM assertion. Re-measured at `8818db4`:

| | |
|---|---|
| `make check` | green — **all five gates**, unchanged in number and definition, including `wails build -tags webkit2_41` |
| `make cover` | `internal/domain` **100.0%**, `internal/service` **93.8%** (bar ≥90%) |
| `make front-test` | **22 files, 247 tests, 0 failures** |
| `make guard` | **six of six** checks pass, `GUARD_ALLOW_RE` **empty** |

(247, not the 251 this document carried before round 1: `ae6befd` deleted the orphaned
tests of the dead TypeScript time-derivations along with the code.)

**Five things nobody on this machine can check are still owed**, and the PASS did not
cover them: see [Owed to a hand pass](#owed-to-a-hand-pass--not-verified). They are
**owed, not done**, and they are [carried into Stage 3](#carried-into-stage-3).

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

### The two-round review history — kept in full

**Stage 2 failed its first review and passed its second.** Both rounds are kept below.
Round 1 is not deleted now that the stage is closed, because **what it caught is the most
useful content in this section**: neither blocker was a missing feature, and both were the
*same* defect class that cost Stage 1 three rounds — a rule with a second spelling. Round
1 fixed in `ae6befd` and `4b1af9c`; round 2 returned **PASS**, verified at `8818db4`.

### Review round 1 — FAIL on two blocking issues, both fixed

Neither was a missing feature. **Both were a rule with a second spelling** — the same
family that cost Stage 1 three rounds, arriving this time in TypeScript and in the grep
that was supposed to stop it.

**Blocking issue 1 — the habit check day was computed in TypeScript.**
`frontend/src/store/data.ts` held `wireDate`, which built the **local calendar day** and
passed it to `CheckHabit(nodeID, date)`, while `internal/service/habit.go` independently
computed `domain.Today(s.clock)` for the `checkedToday` / `scheduledToday` flags **the
same strip renders**. Two clocks, one rule. Across a local midnight the frontend writes a
check for *yesterday*, Go answers `checkedToday: false`, the optimistic tick reverts, and
the check lands on a day the user never chose.

The Reviewer's point was not only the bug: *"no ticket or decision ever ruled on it — the
decision was made in a code comment."* So the ruling is now written down as **D17** in
[`PLAN.md` §7](./PLAN.md), with its rejected alternative and its accepted consequence.
Fixed in `ae6befd`: `CheckHabitToday(nodeID)` / `UncheckHabitToday(nodeID)` are the bound
surface, `HabitService.CheckToday`/`UncheckToday` delegate to the existing **dated**
`Check`/`Uncheck` with `domain.Today(s.clock)` so today keeps **one spelling**, and
**`wireDate` was deleted rather than left unused**. The dated pair stays on the service
for **Stage 7's calendar** — where the user *picks* a day — and is deliberately **not
bound**, so nothing can feed a computed date back in. The new store test drives the
toggle at `23:59:30` and again at `00:00:30` and asserts the two calls are identical,
argument for argument.

> **S2-07's bound surface is therefore amended**: `CheckHabit` / `UncheckHabit` are
> `CheckHabitToday` / `UncheckHabitToday`, each taking a node id only. The count of bound
> methods is unchanged; the dated signature is gone from the wire.

**Blocking issue 2 — `make guard` check 2 claimed to be EXACT and was case-sensitive.**
It ran under `-E`, so `'Done'`, `'Doing'` and `'Today'` walked past a grep that stopped
`'done'`, and the Makefile's own paragraph — *"a hit is a defect and there is nothing to
argue about"* — was false. The Reviewer proved it by getting
`view.status === 'Done' || view.status === 'Doing'` through a clean run: `domain.Status`'s
rule, spelled a second time in TypeScript. Fixed in `4b1af9c` (`-E` → `-iE`) and verified
**both ways** — the proof line passes the old grep and fails the new one, and `-iE`
returns **zero hits** across `frontend/src` as it stands, so nothing is grandfathered and
**`GUARD_ALLOW_RE` stays empty**. The same comment said *"Five checks"* while the recipe
runs six; corrected to match.

**A green `make guard` is not proof that the frontend computes nothing.** Checks **3a and
3b are name-based heuristics** over `overdue|derive|streak|progress|percent`, so a
derivation named anything else — `wireDate` being the worked example — is invisible to
them **by construction**. The Reviewer walked a recomputed today, a recomputed overdue
flag and an inline percentage past a clean guard to demonstrate it. The Makefile discloses
this; **D17** records it in the plan as well, because the risk is a future reader citing a
green guard as evidence. It is evidence that five *names* are absent. **Reading the diff
is still the check**, and rule 7 below is still the rule.

**Dead TypeScript time-derivations deleted in the same commit** (`ae6befd`):
`displayElapsedSeconds`, its `timerReadAt` input, `timerStartedAt`, and `format.ts`'s
`formatTime` and `formatDuration`, with their orphaned tests. Every one was wall-clock
arithmetic on top of Go's `elapsedSeconds`, **tested but rendered by no component** —
which is exactly how a second implementation of a rule waits for its first caller, and
exactly the shape that failed Stage 1 three times. Removed before it could be wired up.
**No ticker was added**: drawing the running clock is Stage 3's, and that ticket adds
precisely what it renders.

### Review round 2 — PASS, at `8818db4`

**This is the PASS that closes Stage 2.** Per `PLAN.md` §5 a stage cannot close without
one.

**Both round-1 blockers were verified closed the hard way.** The Reviewer did not take
the fixes on trust: for each, they **reverted it and watched the specific test go red**,
then restored it.

- `ae6befd` reverted → the store test that drives the habit toggle at `23:59:30` and
  again at `00:00:30` fails, because the two calls stop being identical.
- `4b1af9c` reverted (`-iE` back to `-E`) → the planted
  `view.status === 'Done' || view.status === 'Doing'` line walks past `make guard` again.

A revert that leaves the suite green proves the fix is untested. Both reverts went red.

**The Reviewer also reproduced the guard's blind spot deliberately**, planting a
regression named outside `overdue|derive|streak|progress|percent` and walking it past a
clean `make guard` — the same shape as `wireDate`. That is the evidence behind **D17**
and behind the warning at the head of [Carried into Stage 3](#carried-into-stage-3):
**reading the diff is what caught it, not the guard.**

Re-measured and independently confirmed at `8818db4`:

- `make check` **green, all five gates**, including `wails build -tags webkit2_41`.
- `make cover`: `internal/domain` **100.0%**, `internal/service` **93.8%** (bar ≥90%).
- `make front-test`: **22 files, 247 tests, 0 failures**.
- `make guard`: **six of six**, `GUARD_ALLOW_RE` **empty**.
- **Zero hex literals** in `frontend/src`; `design/` **byte-identical** to its Stage 1
  state; **no cgo** and no `mattn` in the dependency graph; Stage 0's and Stage 1's
  migrations **untouched**; `TestDomainIsPure` and `TestDomainReadsNoClock` **unmodified
  and passing**.
- All **84 commits** on `main` authored solely by `Ismat
  <mukhamejanov.ismat@gmail.com>`, **no AI author and no co-author trailer** on any.
- **No test was weakened by the `CheckHabitToday` / `UncheckHabitToday` binding rename**;
  the Reviewer read each changed test and found **two of them strictly stronger** than
  what they replaced.

**What the PASS did not cover**, and says so: the **five checks owed to a hand pass**.
See [Owed to a hand pass](#owed-to-a-hand-pass--not-verified). They are carried into
Stage 3 unchanged.

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
| `make guard` | S2-10 | The mechanical rules greps: no hex literal, no status-string literal, no recomputed rule, no bare user-visible string. Exits non-zero on any hit. **As shipped it runs six checks** — S2-10's five, plus **check 6**, the no-mouse rule over `App.accept.test.tsx`, added by S2-22 because that ticket's own criterion required the rule to live here ([ratified](#s2-22--test-the-no-mouse-accept-flow-and-the-ru--a11y-audit)). **Check 2 is case-insensitive since `4b1af9c`** — it was `-E` and let `'Done'` past. **Checks 3a/3b are name-based heuristics and a green run proves nothing about an unnamed derivation** (**D17**); see [Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed). |

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
  `CheckHabitToday`, `UncheckHabitToday`, `TimerStart`, `TimerStop`, `TimerCurrent`,
  `Settings`, `SetPalette`, `SetTheme`, `SetAccent`, `SetLanguage`.
  **Amended by [review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed)
  (`ae6befd`, D17):** this line originally said `CheckHabit` / `UncheckHabit`, each taking
  a node id **and a date** — and that date was being computed in TypeScript. The bound
  pair is now `CheckHabitToday` / `UncheckHabitToday`, **node id only**; Go decides the
  day from `s.clock`. The dated `HabitService.Check`/`Uncheck` **remain on the service,
  unbound**, for Stage 7's calendar. The count of bound methods is unchanged.
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
      columns. Verified by hand and reported. **This one cannot be checked on this
      machine** — it is item 5 under
      [Owed to a hand pass](#owed-to-a-hand-pass--not-verified).
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

> **Two corrections from
> [review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed), neither
> reopening this ticket.**
> **(a)** Check 2 shipped case-**sensitive** under `-E` while the target's own comment
> claimed it was EXACT, so `'Done'`, `'Doing'` and `'Today'` walked past it. `4b1af9c`
> makes it `-iE`; **zero hits** on the tree as it stands, so nothing is grandfathered and
> `GUARD_ALLOW_RE` **stays empty**.
> **(b)** Checks **3a/3b match on identifier *names*** —
> `overdue|derive|streak|progress|percent` — so **a derivation named anything else is
> invisible to them by construction**, and a green `make guard` is **not** proof that the
> frontend computes nothing. The round-1 `wireDate` defect is the worked example
> (**D17**). The heuristic is not a bug in this ticket; treating its silence as evidence
> would be.

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
- Checking calls `CheckHabitToday`, unchecking `UncheckHabitToday`, **passing a node id
  and nothing else**, and both re-read. **Optimistic is allowed here on the same terms as
  a move** — revert from Go's answer on error, with a toast.
- **The frontend does not name the day** (**D17**). Go derives it from its clock, with
  the same `domain.Today(clock)` that fills the `checkedToday` and `scheduledToday` flags
  this strip renders. A calendar day built in TypeScript and sent over the wire is a
  second clock, and this ticket's first implementation shipped one; see
  [Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed).
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
- [ ] **No date crosses the wire** (**D17**): the toggle driven at `23:59:30` and again
      at `00:00:30` produces **identical calls, argument for argument**, and
      `git grep -n 'new Date\|toISOString' frontend/src/store` finds no date being built
      for a habit call.
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
| no hex literal, no status literal, no `P0/P1/P2` outside its module, no mouse in the ACCEPT test (`make guard` checks 1, 2, 5, 6 — **exact**) | **that no rule is re-derived in TypeScript.** `make guard` checks 3a/3b are **name-based heuristics**; a derivation named outside `overdue\|derive\|streak\|progress\|percent` does not trip them (**D17**). This one is verified by **reading the diff**, not by either column's tooling |

Nothing in the left column is claimed as hand-verified, and nothing in the right column
is claimed as automated.

**Verified by neither, and deliberately so:** that Aurora's background drift pauses. The
*gate* is automated, above; the **drift itself is not drawn in Stage 2** — see **D16** and
**K5**. It was on the hand list in the first draft of this plan, which would have had the
Reviewer confirm the behaviour of something nothing renders. Removed rather than left to
be discovered at review.

### Owed to a hand pass — NOT VERIFIED

**PARTLY DISCHARGED BY THE USER SINCE, AND ONE OF THEM CAME BACK FAILED** — read the box
first. The heading is left as it was so that every link into this section still resolves.

> **Updated when Stage 3 was planned.** The user has since run the app by hand on real
> hardware. **That pass discharged item 5, returned item 3 as a FAILURE, and left items 1,
> 2 and 4 still owed** — and it found nine defects nothing here could have found,
> **K6 – K14** in [`PLAN.md` §7](./PLAN.md). The per-item status is in the list below;
> everything still owed is re-issued as [**S3-09**](#s3-09--test-the-hand-pass-that-closes-the-blocking-block),
> together with the new checks the nine defects created.
>
> | # | Owed by Stage 2 | After the user's pass |
> |---|---|---|
> | 1 | The ten-step keyboard run, incl. the three D8 due-badge assertions | **STILL OWED.** The user drove the app with the mouse; no step-by-step keyboard report exists. The three D8 badge claims **still have no mechanical backing** — S3-09 adds the two mechanical halves that *can* exist and states exactly what the hand run adds on top |
> | 2 | No background flash on the first frame, in **both** themes | **STILL OWED** — not reported either way |
> | 3 | Russian not clipping at a real 1024×768 | **VERIFIED — AND IT FAILED.** It clips, in Russian *and* in English. This is **K8**, fixed by [S3-02](#s3-02--fix-no-localised-string-may-refuse-to-shrink-k8-d20), and **re-verification is owed** |
> | 4 | The `:focus-visible` ring actually painted, in both palettes | **STILL OWED** — not reported either way |
> | 5 | S2-07's *"a window opens and the first `Board()` returns five columns"* | **DISCHARGED.** The user ran the built binary, saw the board with its columns and cards, dragged cards, and switched palette, theme and language. Verified by observation rather than by a step-by-step report, and that is enough for this claim |
>
> **Nothing above is re-ticked in the Stage 2 DONE list.** Stage 2 closed on what Stage 2
> verified; this note records what happened afterwards.

**Five checks were outstanding at the close of Stage 2, and no document in this repository
may describe any of them as done except where the box above says so.** **Stage 2's PASS did
not cover any of them**, and closing the stage did not close them: they were
[carried into Stage 3](#carried-into-stage-3) exactly as they stand
here. They were not outstanding through negligence: **there is no display on this
machine**, and `xvfb-run`, `scrot`, `import` and `grim` are **all absent**, so the
Reviewer cannot run them either. They are owed to a human at a real keyboard in front of
a real window. Items 1–4 are the precise contents of the right-hand column above; item 5
is a ticket criterion that belonged on this list from the start and was **missing until
review round 1 enumerated it**.

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
5. **[S2-07](#s2-07--feat-open-the-store-construct-the-services-bind-them)'s own last
   criterion**: *"the app opens a window and the frontend can call `Board()` and receive
   five columns. Verified by hand and reported."* It is as unverifiable here as the four
   above and was **not enumerated** until review round 1 found it. `make build` proves the
   binary links; **nothing on this machine proves a window opens**, or that the first
   `Board()` across the real IPC bridge returns five columns. It overlaps item 1 in
   practice — step 1 of the hand script cannot start without it — but it is a criterion on
   a named ticket and is owed in its own right.

**And, held to the same standard, the Aurora drift.** The gate is implemented and was
audited **both ways** in S2-22 — `data-drift` is `on` without reduced motion and `off`
with it — but **nothing draws a drift, so nobody has watched one pause**, and this is not
owed to a human either, because there is nothing to look at. Do not let any document imply
the visual exists (**D16**, **K5**, DONE criterion 21).

---

## Stage 2 — DONE criteria — ALL MET, PASS

Stage 2 closed only once **all** of these held. They do, **verified by the Reviewer at
`8818db4` on the second review**. Per `PLAN.md` §5 a stage cannot close without a
**PASS**; criterion 22 is that PASS.

**The boxes are ticked on the Reviewer's verification, not on the PM's assertion.**
Re-measured at `8818db4`: `make check` green over all five gates, `make cover`
**100.0% / 93.8%**, `make front-test` **22 files / 247 tests / 0 failures**, `make guard`
**six of six** with an empty `GUARD_ALLOW_RE`.

**What is ticked is what was verified, and nothing else.** The **five** things **nobody
on this machine can check** are *not* covered by this PASS — see
[Owed to a hand pass](#owed-to-a-hand-pass--not-verified), which carries the **hand half
of criterion 5**, the **last clause of criterion 8** and the **last clause of criterion
13**. Each of those three criteria is ticked on its mechanical half only, and says so
inline. They are [carried into Stage 3](#carried-into-stage-3).

**Review round 1 returned FAIL**, on two blocking issues fixed in `ae6befd` and
`4b1af9c` and then verified closed by **reverting each fix and watching the specific test
fail**; see
[Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed) and
[Review round 2](#review-round-2--pass-at-8818db4). A first-round FAIL is a normal
outcome here — Stage 1 closed on its fourth review.

1. [x] All twenty-two tickets are committed, one conventional commit each, in order,
   `S2-01` … `S2-22`, authored solely by `Ismat <mukhamejanov.ismat@gmail.com>` with **no
   AI author and no co-author trailer**. **One further commit, `7af4d1d`, closes the
   `set priority` gap under S2-20** and is part of the stage — see
   [The plan correction](#the-plan-correction--set-priority-from-the-palette-is-stage-2).
   **Two more, `ae6befd` and `4b1af9c`, fix review round 1's blocking issues** and are
   likewise part of the stage — see
   [Review round 1](#review-round-1--fail-on-two-blocking-issues-both-fixed).
   **All 84 commits on `main` confirmed authored solely by `Ismat`, no AI trailer.**
2. [x] `make check` is green — **all five gates, unchanged in number and definition**.
   Confirmed at `8818db4`, including `wails build -tags webkit2_41`.
3. [x] `make cover` is green: `internal/domain` ≥ 90.0% **and** `internal/service` ≥ 90.0%,
   measured per package after `go clean -testcache`. **Actual: `internal/domain`
   100.0%, `internal/service` 93.8%.**
4. [x] `make front-test` is green and `make guard` exits 0. **Actual: 22 files, 247 tests,
   0 failures; `make guard` six of six with `GUARD_ALLOW_RE` empty.**
5. [x] **The ACCEPT criterion is MET** — `frontend/src/App.accept.test.tsx` drives the
   **assembled `<App />`**, imports no component beneath it, presses keys and only keys,
   and asserts the **exact sequence of client calls**: one `CreateNode`, then
   `MoveToColumn` with each of the four statuses Go supplied, in order —
   create → move across all five columns → complete, no mouse. **`make guard` check 6
   fails if a `click`, a `pointer` or a `fireEvent.mouse*` ever enters that file**, and
   fails too if the file is deleted or emptied, so the proof cannot rot quietly.
   **The hand half is NOT covered by this PASS**: the ten-step script on
   `./build/bin/nexus` is still [owed](#owed-to-a-hand-pass--not-verified), the three
   **D8 due-badge assertions** included, and is
   [carried into Stage 3](#carried-into-stage-3).
6. [x] **C1 is wired and cannot be unwired silently**: moving a card to Doing opens
   exactly one `time_entry`; a cascade opens none; moving out of Doing closes it; a
   project opens none. The negative control is recorded in S2-03's commit body.
7. [x] **C2 is fixed**: the derivation index is built once per `Board()`, with a committed
   benchmark and before/after numbers in the commit body.
8. [x] **C3/K1 is fixed (D12)**: `LC_NUMERIC=C` is forced, the background is seeded from
   settings, `main.go` contains **no hex literal**. **The last clause is NOT covered by
   this PASS**: the no-flash behaviour in **both themes** is GTK-side and pre-paint,
   jsdom cannot see it, and it is still [owed](#owed-to-a-hand-pass--not-verified) and
   [carried into Stage 3](#carried-into-stage-3).
9. [x] **C4/K2 is fixed (D14)** and **C5/K3 is fixed (D15)**, each with the named test.
10. [x] **No rule has a second spelling.** The Stage 1 inventory — `HasColumn`, `HasDue`,
    `DoingRefusal`/`CanBeDoing`, `countsAsWork`, habit-requires-recurrence,
    `DefaultActivity` — is still one definition each, and **no `frontend/src` file
    computes a status, a column eligibility, progress, a streak, an overdue flag or a due
    date**, **and does not compute a *date* Go will act on** (**D17**). `make guard`
    passes **and the Reviewer has read the diff rather than trusting it** — checks 3a/3b
    are name-based heuristics and a derivation named outside
    `overdue|derive|streak|progress|percent` is invisible to them, which is how the
    round-1 `wireDate` defect survived a green guard for a whole stage.
    **The Reviewer read the diff and, at round 2, reproduced the blind spot with a
    planted regression that walked past a clean `make guard`.** The guard is not the
    evidence; the reading is.
11. [x] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing. **Confirmed:
    zero hits.**
12. [x] `design/` is byte-identical to its Stage 1 state:
    `git diff <stage-1-tag-or-sha> -- design/` is empty. **Confirmed.**
13. [x] Both locale files are **complete and key-identical**; every user-visible string is
    a key. **The last clause is NOT covered by this PASS**: jsdom has no layout engine, so
    what is asserted is the **mechanisms** that prevent clipping, never the absence of
    clipping. **Russian not clipping at a real 1024×768** is still
    [owed](#owed-to-a-hand-pass--not-verified) and
    [carried into Stage 3](#carried-into-stage-3).
14. [x] Every `App` method returns `(T, error)` (reflection test), every rejection reaches
    a toast, and `frontend/src/lib/client.ts` is the only importer of `wailsjs`.
15. [x] **Local-only holds**: no network call, no CDN, no font fetch.
    `git grep -nE 'https?://' frontend/src` returns nothing, and every added dependency is
    bundled from `node_modules`.
16. [x] `CGO_ENABLED=0 go build ./...` succeeds and `go list -deps ./... | grep -i mattn`
    prints nothing. **Confirmed.**
17. [x] `internal/domain` is still pure — `TestDomainIsPure` and `TestDomainReadsNoClock`
    pass **unmodified**. **Confirmed: the test files are byte-identical.**
18. [x] `git status --porcelain` is empty after
    `make check && make cover && make front-test && make guard` — including the generated
    `frontend/wailsjs`, which is committed and not left dirty.
19. [x] `0001_settings.sql` and every Stage 1 migration are byte-identical to their
    closed state.
20. [x] **Every component Stage 2 built is reachable in the running app.**
    `frontend/src/App.mount.test.tsx` reports an **empty** difference between the
    components under `frontend/src/{components,views}` and the import closure of
    `frontend/src/main.tsx`; the header, habits strip, board, overlay and toast regions of
    the shell are all filled; and the ACCEPT flow test drives `<App />`. A stage whose
    every ticket passed and whose app opens blank has not delivered a launch screen —
    this is the criterion that says so.
21. [x] **Nothing is claimed that was not verified, specifically on the Aurora drift**:
    the report says the drift is **gated and not drawn** (**D16**, **K5**) and nobody
    claims to have watched it pause. **Confirmed — and it stays true: nothing draws a
    drift, and no document may be edited into implying the visual exists.**
22. [x] **The Reviewer returns PASS.** **Returned on the SECOND review, verified at
    `8818db4`**, with both round-1 blockers checked by **reverting each fix and watching
    the specific test fail**, and with no test weakened by the binding rename — two are
    strictly stronger. See
    [Review round 2](#review-round-2--pass-at-8818db4). **This is the PASS that closes
    Stage 2.**

Nothing was marked done that was not actually verified — which is why criteria **5, 8 and
13** are ticked on their mechanical halves only and name the half that is still owed.
Stage 1 closed on its fourth review; a first-round FAIL here was a normal outcome, not a
failure of the process.

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

---

## Carried into Stage 3

**These are Stage 3 acceptance criteria, not suggestions.** Stage 2 closed clean, but it
closed owing the items below. **Whoever plans Stage 3 must turn each one into a ticket
with a checkable criterion, or state explicitly why not.** Same contract as
[Carried into Stage 2](#carried-into-stage-2). **Nothing here reopens a Stage 2 ticket.**

> **DISCHARGED BY THE STAGE 3 PLAN — every item below is now absorbed.** The contract said
> *a ticket with a checkable criterion, or an explicit refusal*; here is the disposition,
> so that Stage 3 absorbs this list rather than rediscovering it.
>
> | Carried item | Absorbed by |
> |---|---|
> | The warning that a green `make guard` is **not a proof** | Not a ticket — it is [rule 16 for the Dev agent in Stage 3](#rules-for-the-dev-agent-stage-3-additions) and a DONE criterion. It also **came true a second time**: see **K8**/**D20**, where the RU anti-clipping audit was green on a real clipping bug |
> | Hand-pass check 1 — the ten-step keyboard run and the three D8 badge assertions | [**S3-09**](#s3-09--test-the-hand-pass-that-closes-the-blocking-block), which also adds the two mechanical halves the badge claim was missing |
> | Hand-pass check 2 — no background flash, both themes | [**S3-09**](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) |
> | Hand-pass check 3 — Russian not clipping at 1024×768 | **It came back FAILED** — that is **K8**. Fixed by [**S3-02**](#s3-02--fix-no-localised-string-may-refuse-to-shrink-k8-d20); re-verified in [**S3-09**](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) |
> | Hand-pass check 4 — the `:focus-visible` ring painted | [**S3-09**](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) |
> | Hand-pass check 5 — a window opens and `Board()` returns five columns | **DISCHARGED** by the user's pass |
> | **K1/D12**, **K2/D14**, **K3/D15** | Implemented in Stage 2 and unchanged. K1's no-flash half is hand-pass check 2, above |
> | **D16 / K5** — the drift gate with nothing behind it | **Explicitly refused again.** No drift specification exists, so no ticket. Not to be invented |
> | **C6** — enum membership spelled twice | [**S3-18**](#s3-18--featservice-publish-every-enum-set-c6-d26), taking the first of the two recorded options (**D26**). With an honest correction: current locale parity is **clean**, so C6 is a **drift risk**, not a present defect |
> | Draw the running timer clock | [**S3-24**](#s3-24--featfrontend-attachments-the-editable-time-log-and-the-running-clock), which adds **exactly what it renders** |
> | Aurora's background drift | Refused, as above |

### ⚠️ Read this first — a green `make guard` is NOT a proof

**No status line, report, commit body or ticket in this repository may offer a green
`make guard` as evidence that the frontend computes nothing.** **D17** records why:
`make guard` checks **3a and 3b are name-based heuristics** over
`overdue|derive|streak|progress|percent`. A derivation named anything else is invisible
to them **by construction**.

This is not hypothetical. **`wireDate` was named after none of those trigger words and
sat behind a green guard for an entire stage**, computing the habit check day in
TypeScript on a decision that had only ever been written in a code comment. It was caught
by a human reading the diff, and at round 2 the **Reviewer reproduced it with a planted
regression** that walked straight past a clean guard.

A green guard is evidence that **five names are absent**. **Reading the diff is still the
check**, and [rule 7 for the Dev agent](#rules-for-the-dev-agent) is still the rule.

### The five checks owed to a hand pass — still owed, none verifiable here

**Stage 2's PASS did not cover these.** There is no display on this machine and
`xvfb-run`, `scrot`, `import` and `grim` are all absent, so neither the Dev nor the
Reviewer could run them. They are owed to a human at a real keyboard in front of a real
window, and they are reproduced in full under
[Owed to a hand pass](#owed-to-a-hand-pass--not-verified).

| # | Owed | Ties back to |
|---|---|---|
| 1 | The **ten-step keyboard run** on `./build/bin/nexus`, mouse untouched, reported step by step — **including the three D8 due-badge assertions at steps 3, 4 and 7, which have *no* mechanical backing anywhere in the suite** | DONE criterion 5, hand half |
| 2 | **No flash of the wrong background on the first frame, in *both* themes** (**D12**, **K1**) — GTK-side and pre-paint | DONE criterion 8, last clause |
| 3 | **Russian not clipping at a real 1024×768** — jsdom has no layout engine, so only the *mechanisms* are asserted | DONE criterion 13, last clause |
| 4 | **The `:focus-visible` accent ring actually painted, in both palettes** — jsdom evaluates `:focus-visible` as false for programmatic focus | S2-22's a11y audit |
| 5 | **S2-07's** *"the app opens a window and the frontend can call `Board()` and receive five columns"* | [S2-07](#s2-07--feat-open-the-store-construct-the-services-bind-them) |

**`make build` proves the binary links. Nothing on this machine proves a window opens.**

### Known issues carried through from Stage 1, unaffected by Stage 2

Stage 2 implemented the decisions; the issues stay on the record in [`PLAN.md`
§7](./PLAN.md) because they explain why the code looks the way it does.

| | Issue | Decision | Where it stands |
|---|---|---|---|
| **K1** | `BackgroundColour` never reaches GTK under `LC_NUMERIC=ru_RU.UTF-8` | **D12** (user) | Implemented in **S2-08**; the **no-flash behaviour is item 2 above and is still unwatched** |
| **K2** | A leaf project's stale *stored* status has teeth since D11 | **D14** (PM ruling, overturnable) | Implemented in **S2-06** with its named test |
| **K3** | An empty project stored `done` renders in Done with no bar | **D15** (PM ruling, overturnable) | Implemented in **S2-14** with its named test |

### D16 / K5 — the drift gate is built and NOTHING DRAWS A DRIFT

`data-drift` is set correctly — `on` without reduced motion, `off` with it, on every
palette — it was **audited both ways** in S2-22, and **no code reads it**. Nobody has
watched a drift pause, because **there is no drift to watch**. This is not on the
hand-pass list above: it is not owed to a human either, because there is nothing to look
at.

**No document in this repository may be edited into implying the visual exists**, and the
drift is **not to be invented**. It returns as a ticket when a drift **specification**
does (§3, **D6** as amended).

### The rest

| | Item | Source | Status |
|---|---|---|---|
| **C6** | Five enum sets are spelled a second time in the locale files, with nothing tying them to Go | Review round 1, non-blocking finding 4 | **Judgement call, not a defect.** Stage 3 picks one of two fixes — see below |
| — | Draw the running timer clock | `ae6befd` deleted the untested-in-anger wall-clock helpers no component rendered | Stage 3 adds **exactly what it renders**, on Go's `elapsedSeconds` |
| — | Aurora's background drift | **D16**, **K5** | Only when a drift **specification** exists. **Not to be invented** |

### C6 — enum membership is spelled twice, and nothing checks it

`frontend/src/lib/commands.ts` and `frontend/src/components/AppearanceControls.tsx` take
`Object.keys` of **`settings.palette`, `settings.theme`, `settings.language`,
`palette.priority` and `card.type`** out of `en.json` and treat them as the
**authoritative sets**. Go owns all five — `domain.Palettes()`, `domain.Themes()`,
`domain.Priorities()` and the node types — and **publishes none of them over the wire**.
`locales.test.ts` checks **en/ru key parity** and nothing ties *either* file to Go, so if
Go gains a palette or a priority, **the UI silently will not offer it and no test goes
red**.

**The Reviewer ruled this an acceptable judgement call for Stage 2, not a defect**, and
the contrast that makes it instructive is inside the same stage: **where Go *does*
publish a set — the five Kanban columns — the frontend takes the set from Go and the
locale file only supplies the *labels*.** That is the shape the other five should end up
in. Nothing today is wrong on screen; what is missing is the thing that would go red if
it became wrong.

Stage 3 picks **one** of:

- [ ] **Bind a set-publishing method** for each set, so the frontend enumerates what Go
      enumerates and the locale file is reduced to labels — the columns' shape,
      generalised. Preferred, because it removes the second spelling rather than
      detecting it.
- [ ] **Add a parity test** that fails when a Go set and its locale table disagree, in
      either direction. Cheaper, and it leaves the duplicate in place with an alarm on
      it.

Either way the criterion is the same: **adding a value to a Go set must turn something
red** until the frontend offers it.

**Stage 3 takes the first option**, as **D26**, and it is ticket
[**S3-18**](#s3-18--featservice-publish-every-enum-set-c6-d26).

---

## Stage 3 — Defect remediation, then Detail + Tree + Search/Archive

**Status: PLANNED, not started.** Twenty-nine tickets, **S3-01 … S3-29**, one
conventional commit each, in **two blocks**.

**Block A — [S3-01 … S3-09](#block-a--the-blocking-defect-block-s3-01--s3-09) — is
blocking, and blocking is meant literally: no ticket from S3-10 onwards may be started
until every ticket in block A is committed and `make check`, `make cover`,
`make front-test` and `make guard` are all green at the end of it.** That is the user's
condition, in their words: *"if these details are resolved in the next steps, we are good
to go."*

**Block B — S3-10 … S3-29 — is the feature stage** from `PLAN.md` §5: the task detail
panel, the tree view, search and archive, plus **C6** and `README.md`.

### What the user's hand pass found — and why no gate here could have found it

Stage 2 closed on a Reviewer PASS that was **honest about its own limits**: it listed five
checks as *owed to a human at a real keyboard*, because there is no display on this machine
and `xvfb-run`, `scrot`, `import` and `grim` are all absent. The user then ran
`./build/bin/nexus` on real hardware, and **the gap Stage 2 named is exactly where nine
defects were sitting.**

They are **K6 – K14** in [`PLAN.md` §7](./PLAN.md), each traced to exact lines by a
read-only investigation, each ruled on as **D18 – D26**:

| | Defect | Ruling | Ticket |
|---|---|---|---|
| **K6** | The dragged card vanishes on grab — there is no `DragOverlay`, so the source `<li>` is translated in place, painted over by the next column's stacking context, and dropped to `null` at the `SortableContext` boundary | **D18** | [S3-05](#s3-05--fixfrontend-the-dragged-card-follows-the-pointer-k6-d18) |
| **K7** | Resizing breaks the layout and it does **not** recover without a restart. ~50 simultaneous `backdrop-filter` surfaces; **confirmed by the user's own discriminating experiment** — switching to Studio (`--blur: 0px`) repaired a stuck layout live | **D19** | [S3-04](#s3-04--fixfrontend-blur-the-few-large-surfaces-not-every-small-one-k7-d19) |
| **K8** | `shrink-0` on a max-content localised string overflows the column. **It overflows in English too.** The RU audit was green on it | **D20** | [S3-02](#s3-02--fix-no-localised-string-may-refuse-to-shrink-k8-d20) |
| **K9** | No `MinWidth`/`MinHeight`; the window can be dragged ~380px below the layout floor, and **the user changes display scaling often** | **D21** | [S3-03](#s3-03--fix-the-window-minimum-size-derived-from-the-layout-floor-k9-d21) |
| **K10** | No height chain, so the **document** scrolls; `overflow-x: auto` silently promotes `overflow-y`; every `.focus()` scrolls its ancestors | **D22** | [S3-01](#s3-01--fixfrontend-one-height-chain-the-board-owns-the-scroll-and-focus-never-scrolls-k10-d22) |
| **K11** | Neither UI font has a single Cyrillic glyph, so Russian drops the whole UI to the system `sans-serif`; `<html lang>` is frozen at `en` | **D23** — **the user's** | [S3-06](#s3-06--featfrontend-a-cyrillic-face-under-the-same-family-names-and-html-lang-k11-d23) |
| **K12** | Toasts stack without limit, and a domain *refusal* is rendered as "something went wrong" with no clue which action failed | **D24**, **D25** | [S3-07](#s3-07--fixfrontend-toasts-are-capped-de-duplicated-and-self-dismissing-k12-d24), [S3-08](#s3-08--feat-a-refusal-is-not-a-failure-and-every-toast-says-what-failed-k12-d25) |
| **K13** | Shortcuts match `event.key`, so a Cyrillic keyboard layout would kill Ctrl+N / Ctrl+K. **Latent** — the user keeps a Latin layout | ranked low, on purpose | [S3-28](#s3-28--fixfrontend-shortcuts-match-the-key-not-the-character-k13) |
| **K14** | The palette **selection** is not visibly indicated | **DEFERRED BY THE USER** | none — and that is the user's call |

**⚠️ The generalisable lesson, and it is D17 arriving in a second place.** `Card.tsx:36-43`
says in a comment that *"nothing on this card has a fixed width and nothing is
`whitespace-nowrap`"*, and `App.accept.test.tsx:344-355` asserts the absence of `truncate`
and `whitespace-nowrap`. **Both statements are true, and the card clipped anyway**, because
`shrink-0` on a localised max-content string clips identically and was not on the list.
**A mechanism-based audit is only as good as its enumeration of mechanisms, and an
enumeration is a guess about the future.** So S3-02 does not simply add a word to a grep —
it changes the audit's *shape* (see **D20**).

### ACCEPT

Two halves, and the first gates the second.

> **Block A.** Every defect in **K6 – K14** is closed, each with a **named mechanical
> check** where one is possible and an **explicit statement that only an eye can see it**
> where one is not — plus **S3-09**, the hand pass, run by a human on real hardware and
> reported item by item.

> **Block B.** *"Every field round-trips through Go; reparent in the tree shows on Kanban
> instantly"* — `PLAN.md` §5.

**How block B's half is demonstrated**, so it is not left to interpretation: a mechanical
test that, for **every editable field the detail panel exposes**, opens the panel on a
node, changes the field by keyboard, asserts the **exact Go call and argument**, and then
asserts the panel re-renders **Go's answer** and not the typed value — a field that renders
what the user typed rather than what Go stored is a field that has not round-tripped. Plus
a tree test that reparents a node and asserts the Kanban board reflects it without a manual
refresh. Plus S3-09's successor hand pass at the end of the stage.

### What Stage 3 must deliver

- **The nine defects, first, and completely.** Block A.
- **The detail slide-over**: Markdown editor and preview, inline subtasks, tags, due,
  priority, estimate, activity, the recurrence editor, attachments copied into the app
  data dir, the editable time log, the type switcher.
- **`COUNT` and `UNTIL` in the recurrence language** (**D27**, the user's answer to
  **OQ2**; **D28**, its consequences for **D5**). S3-19 is the domain half and S3-23 the
  editor half, in that order.
- **The tree view**: collapsible, inline rename, drag-to-reparent, arrow/Enter/Tab
  keyboard navigation.
- **Search** with tag/type/status/date filters, over the backend S1-04 chose.
- **The archive view**, with restore.
- **C6**, closed by **D26**: Go publishes every enum set.
- **`README.md`**, which has been deliberately unwritten since Stage 0 because it needed a
  screen to photograph. It has one now — but automated capture is **blocked on the user
  installing `xvfb` and `imagemagick`** (see S3-29).

### What Stage 3 must NOT do

Out of scope, and it must stay out: the **standalone frameless quick-add window** and the
natural-language parser (Stage 4 — Stage 3's quick-add stays the in-app one), **focus
mode**, the **tray**, autostart and the `.desktop`/`install.sh` work, **D-Bus sleep/lock**
timer handling, **backup, restore-from-file and export**, the **PMP timelog screen**, the
**calendar**, **stats** and **Gantt**.

**And two specific refusals, carried forward unchanged:**

- **Aurora's background drift** (**D16**, **K5**). No specification exists. The gate is
  built and **nothing draws a drift**; it returns when a spec does, and **it is not to be
  invented**. No document may be edited into implying the visual exists.
- **Making the palette selection more visibly indicated** (**K14**). The user raised it and
  then explicitly deferred it. It is not Stage 3 work, and adding it is scope creep with a
  user's sentence quoted out of context.

### Rules for the Dev agent, Stage 3 additions

The six rules at the top of this file and Stage 2's additions 7–15 **all still apply in
full**, unchanged. These are added:

16. **A green `make guard` is NOT a proof, and this is now true twice over.** Checks 3a/3b
    are name-based heuristics (**D17**, `wireDate`), and **K8** is the second instance of
    the same shape in a different tool: the Russian anti-clipping audit was green on a card
    that overflowed in English. **No commit body, ticket report or status line may cite a
    green mechanical check as evidence for a claim the check does not actually make.** Say
    what was checked and say what was not.
17. **Say which half of each claim is mechanical and which is by eye.** Several of block
    A's defects are invisible to jsdom, which has **no layout engine and no font engine**.
    A ticket that cannot mechanically prove its own fix must say so **in its acceptance
    criteria and in its commit body**, and the visual half goes on
    [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block)'s list. Pretending a
    test covers it is the one unrecoverable error here.
18. **Block A is blocking.** Do not start S3-10 until S3-01 … S3-09 are all committed and
    all four commands are green. If a block A ticket cannot be completed as written, stop
    and report — do not skip ahead.
19. **A rule still gets exactly one spelling, and block A is full of new candidates**:
    `preventScroll`, the blur allow-list, the layout floor, the refusal-code table, the
    Cyrillic `unicode-range`. Each of those is a rule. Each gets one home, named in its
    ticket. **Block B adds two more**: the **recurrence bound** (`COUNT`/`UNTIL` — it lives
    in `Matches`/`Previous`, not in `streak.go` and not in the editor) and **"has this
    series ended"** (`HabitView.Ended`, computed in Go and compared to nothing in
    TypeScript). See **D27** and **D28**.

### New dependencies, and why each one is justified

Same contract as Stage 2: a normal npm package, installed into `node_modules`, bundled by
Vite, **making no network call at runtime** — and that is an acceptance criterion on the
ticket that adds it, not a hope.

| Package | Ticket | Why it, and not hand-rolled |
|---|---|---|
| `@fontsource/inter` | S3-06 | **K11**: neither UI face has a Cyrillic glyph, so 81 of `ru.json`'s 82 leaves render in the system `sans-serif`. **Inter is the user's choice (OQ1, closed)**, one face registered under **both** family names. Vendored exactly as the three existing faces are — **woff2 files from `node_modules`, fingerprinted into `dist/assets`, no CDN, no `<link>`, no runtime fetch**, and that is an acceptance criterion and not a hope. **This is a new dependency and S3-06's commit body must say so explicitly.** The Dev verifies Inter's real coverage from the package's own `unicode.json` before committing — the name is not the evidence. |
| `react-markdown` + `remark-gfm` | S3-20 | The brief specifies a Markdown editor **and preview**. A hand-rolled renderer is a security surface and a correctness sink. **`react-markdown` does not render raw HTML by default, and `rehype-raw` must NOT be added** — a local-only app rendering arbitrary HTML from its own database is a needless hole. Both are pure-JS, bundled, no network. |

**Nothing else, and OQ2's answer does not change that.** In particular: **no RRULE
library.** **OQ2** came back as its middle option, not option C — `COUNT` and `UNTIL` are
added to the hand-rolled parser by **S3-19**, which is two bounded fields, while a library
would put a dependency into the one package that is **pure and at 100% coverage** in order
to accept a grammar the editor cannot express. `recurrence.go`'s own header comment already
records why no maintained pure-Go expander is usable here — a hidden clock read and zoned
instants — and that reasoning is now load-bearing twice. Also: **no e2e browser runner**, rejected in Stage 2 for reasons
that have not changed; **no date library**, because dates are Go's and `Intl` formats them.

### The gate commands, and the two non-gate targets

**`make check` is still exactly the five gates.** Unchanged in number and definition.

`make cover` and `make front-test` are unchanged. **`make guard` grows from six checks to
eight**, and each new check is added by the ticket whose own criterion requires it, with
the `Makefile` **named in that ticket's Scope** — the Scope-list omission that had to be
ratified for S2-22 is not repeated:

| Check | Added by | What it greps |
|---|---|---|
| 7 | **S3-01** | `.focus(` appears in exactly one module (`lib/focus.ts`). A focus call anywhere else is a second spelling of *"focus must not scroll"* (**D22**). |
| 8 | **S3-04** | `backdrop-blur-glass` appears **only** in the three files **D19** allow-lists. A hit anywhere else is a new compositing layer nobody decided on. |

Both are **exact greps, not heuristics** — which is why they are allowed to be guard checks
at all (**D17**). `GUARD_ALLOW_RE` **stays empty**: every hit so far has been a real bug
fixed at the source, and no Stage 3 ticket may be the first to allow-list one.

### Stage 3 ticket index

| ID | Title | Prefix |
|---|---|---|
| **Block A — blocking** | | |
| [S3-01](#s3-01--fixfrontend-one-height-chain-the-board-owns-the-scroll-and-focus-never-scrolls-k10-d22) | One height chain, the board owns the scroll, and focus never scrolls (**K10**, **D22**) | `fix:` |
| [S3-02](#s3-02--fix-no-localised-string-may-refuse-to-shrink-k8-d20) | No localised string may refuse to shrink (**K8**, **D20**) | `fix:` |
| [S3-03](#s3-03--fix-the-window-minimum-size-derived-from-the-layout-floor-k9-d21) | The window minimum size, derived from the layout floor (**K9**, **D21**) | `fix:` |
| [S3-04](#s3-04--fixfrontend-blur-the-few-large-surfaces-not-every-small-one-k7-d19) | Blur the few large surfaces, not every small one (**K7**, **D19**) | `fix:` |
| [S3-05](#s3-05--fixfrontend-the-dragged-card-follows-the-pointer-k6-d18) | The dragged card follows the pointer (**K6**, **D18**) | `fix:` |
| [S3-06](#s3-06--featfrontend-a-cyrillic-face-under-the-same-family-names-and-html-lang-k11-d23) | A Cyrillic face under the same family names, and `<html lang>` (**K11**, **D23**) | `feat:` |
| [S3-07](#s3-07--fixfrontend-toasts-are-capped-de-duplicated-and-self-dismissing-k12-d24) | Toasts are capped, de-duplicated and self-dismissing (**K12**, **D24**) | `fix:` |
| [S3-08](#s3-08--feat-a-refusal-is-not-a-failure-and-every-toast-says-what-failed-k12-d25) | A refusal is not a failure, and every toast says what failed (**K12**, **D25**) | `feat:` |
| [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) | **The hand pass that closes the blocking block** | `test:` |
| **Block B — the feature stage** | | |
| [S3-10](#s3-10--featservice-nodedetail--the-one-read-the-panel-renders) | `NodeDetail` — the one read the panel renders | `feat:` |
| [S3-11](#s3-11--featservice-the-field-writers--title-description-due-estimate-activity) | The field writers — title, description, due, estimate, activity | `feat:` |
| [S3-12](#s3-12--featservice-the-type-switcher) | The type switcher | `feat:` |
| [S3-13](#s3-13--featservice-tags) | Tags | `feat:` |
| [S3-14](#s3-14--featservice-attachments-copied-into-the-app-data-dir) | Attachments copied into the app data dir | `feat:` |
| [S3-15](#s3-15--featservice-the-editable-time-log) | The editable time log | `feat:` |
| [S3-16](#s3-16--featservice-search-with-tagtypestatusdate-filters) | Search with tag/type/status/date filters | `feat:` |
| [S3-17](#s3-17--featservice-the-archive-list-and-restore) | The archive list and restore | `feat:` |
| [S3-18](#s3-18--featservice-publish-every-enum-set-c6-d26) | Publish every enum set (**C6**, **D26**) | `feat:` |
| [S3-19](#s3-19--featdomain-count-and-until-enter-the-recurrence-language-d27-d28) | **`COUNT` and `UNTIL` enter the recurrence language** (**D27**, **D28**) — **S3-23 is blocked on it** | `feat:` |
| [S3-20](#s3-20--featfrontend-the-detail-slide-over--shell-keyboard-markdown) | **The detail slide-over** — shell, keyboard, Markdown | `feat:` |
| [S3-21](#s3-21--featfrontend-the-panels-field-editors) | The panel's field editors | `feat:` |
| [S3-22](#s3-22--featfrontend-inline-subtasks) | Inline subtasks | `feat:` |
| [S3-23](#s3-23--featfrontend-the-recurrence-editor-d27-d28) | The recurrence editor (**D27**, **D28**) — **blocked on S3-19** | `feat:` |
| [S3-24](#s3-24--featfrontend-attachments-the-editable-time-log-and-the-running-clock) | Attachments, the editable time log, and the running clock | `feat:` |
| [S3-25](#s3-25--featfrontend-the-tree-view) | **The tree view** | `feat:` |
| [S3-26](#s3-26--featfrontend-the-search-screen) | **The search screen** | `feat:` |
| [S3-27](#s3-27--featfrontend-the-archive-view) | **The archive view** | `feat:` |
| [S3-28](#s3-28--fixfrontend-shortcuts-match-the-key-not-the-character-k13) | Shortcuts match the key, not the character (**K13**) | `fix:` |
| [S3-29](#s3-29--docs-readmemd-and-the-screenshots--blocked-on-the-user) | `README.md` and the screenshots — **BLOCKED on the user** | `docs:` |

**Sequencing, and the three constraints that are real.**

- **Block A before block B, entirely.** Rule 18.
- **Inside block A: S3-01 → S3-02 → S3-03.** The height chain settles first, the widths
  settle second, and **S3-03 derives the window's minimum size from the layout floor — so
  it must come after the floor is final.** Running it earlier bakes in a number that S3-02
  then changes.
- **S3-04 before S3-05.** Both rewrite `Column.tsx`. The blur policy goes first because
  S3-05's `DragOverlay` has to work under whatever the policy turned out to be, and
  because S3-05 deletes the `z-10` that only existed to fight the stacking context S3-04
  is reducing.
- **S3-06, S3-07 and S3-08 touch none of those files** and could in principle move, but
  they are kept after the layout work so the hand pass in S3-09 sees one settled screen.
- **Inside block B: every Go ticket before the frontend ticket that calls it.**
  S3-10 … S3-19 are Go and touch no `frontend/src` file except the regenerated
  `frontend/wailsjs`. **S3-20 before S3-21 … S3-24**, which fill the panel it creates.
  **S3-25 introduces the view switch** and therefore precedes S3-26 and S3-27.
- **S3-19 before S3-23, and that dependency is hard.** The user's **OQ2** answer put
  `COUNT` and `UNTIL` into the recurrence language (**D27**, **D28**), so the recurrence
  editor cannot offer *"repeat N times"* or *"repeat until date"* until `internal/domain`
  accepts them. **S3-19 is the domain half and S3-23 is blocked on it**; S3-23 may not
  widen the parser itself, which is the split the recurrence-editor ticket already demanded
  if OQ2 came back B. **Note the renumber**: inserting S3-19 shifted every frontend ticket
  up by one, so the recurrence editor that used to be S3-22 is now **S3-23**, and
  `README.md` is **S3-29**.

### Composition — who mounts what, in Stage 3

**The rule from Stage 2 stands unchanged and is the reason that section exists:**

> **A component that is built and not mounted is an unfinished ticket.** The ticket that
> builds a top-level piece is the ticket that mounts it, in the same commit. "Mounted"
> means reachable in the running app from `main.tsx`.

`frontend/src/App.mount.test.tsx`'s orphan walk stays green throughout, and **each mounting
ticket keeps its own `render(<App />)` reachability criterion**, because the two halves do
different jobs and neither is sufficient alone — the Dev demonstrated the orphan walk's
blind spot twice in Stage 2.

| Region in `App.tsx` | Filled by | Mounted in |
|---|---|---|
| detail slide-over layer (a new region, beside the overlay layer) | `components/DetailPanel.tsx` | **S3-20** |
| region 3 becomes a **view switch** on a new `activeView` in `store/ui.ts` | `views/Tree.tsx` | **S3-25** |
| region 3 — the search screen | `views/Search.tsx` | **S3-26** |
| region 3 — the archive view | `views/Archive.tsx` | **S3-27** |

Everything else is mounted by its parent: the field editors of S3-21, the subtask list of
S3-22, the recurrence editor of S3-23 and the attachment/time-log sections of S3-24 are all
children of S3-20's panel and become reachable when it opens.

**Two consequences worth stating before they are discovered.**

1. **S3-25 owns the view switch, not just the tree.** Today `views/Kanban.tsx` is region 3
   and nothing else can be. S3-25 introduces `activeView`, the switch in `App.tsx`, and
   flips the command palette's `view:tree` row from *unavailable, coming in stage N* to
   available. S3-26 and S3-27 then each add one view and one palette row, and **no row may
   be left silently doing nothing** — the Stage 2 standard.
2. **The detail panel and the overlays must not be open at once by accident.**
   `store/ui.ts` holds **one** overlay, which is what makes that impossible today. The
   panel is not an overlay — it is a slide-over beside the board — so S3-20 must say
   explicitly what happens when `Ctrl+K` is pressed with the panel open, and test it.

---

### Block A — the blocking defect block (S3-01 … S3-09)

**Nine tickets. None of block B starts until all nine are committed and green.**

**Read this before starting any of them.** Most of what these tickets fix is **invisible to
every mechanical check in this repository**: jsdom has no layout engine, so every
`offsetWidth` is 0 and nothing ever overflows; jsdom has no font engine, so no test can see
which glyphs a face contains at paint time; there is no display, so nothing here can watch
a window resize or a compositing layer go stale. **Each ticket below therefore states, in
its own acceptance criteria, which half of its claim is mechanical and which half is
[S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block)'s.** A ticket that
claims more than it proved is the one failure this project has decided it will not accept
(rule 17, **D17**).

---

## S3-01 — fix(frontend): one height chain, the board owns the scroll, and focus never scrolls (K10, D22)

**This is K10, ruled by D22**, and it is first because everything else in block A is
measured against a settled layout.

Three symptoms, one missing decision about who owns the scroll. There is **no `#root` rule
at all** in the compiled CSS; `body` has only `min-height: 100vh`; `<main>` (`App.tsx:154`)
carries `min-w-0`, which is a **no-op in a column flex container**, and **not** the
`min-h-0` it needs, so its `min-height: auto` floors it at min-content and the document
grows past the viewport. `Kanban.tsx:316` sets `overflow-x-auto`, which per CSS Overflow 3
**promotes `overflow-y` to `auto`**, so the board already scrolls in an axis its own comment
says it does not. And all eleven `.focus()` calls omit `{ preventScroll: true }`, so each
scrolls every scrollable ancestor.

**Scope (may touch):** `frontend/src/style.css`, `frontend/src/App.tsx`,
`frontend/src/views/Kanban.tsx`, a new `frontend/src/lib/focus.ts` and its test, the call
sites `frontend/src/components/{HabitStrip,QuickAdd,CommandPalette,AppearanceControls}.tsx`,
the `Makefile` (guard check 7 — declared here deliberately, so it is not a Scope omission),
tests beside them.

Requirements:

- **A full-height chain**: `html`, `body` and `#root` are full height, the shell fills it,
  and `<main>` carries **`min-h-0`** so it may shrink below its content. The useless
  `min-w-0` on `<main>` goes, or is justified in a comment — a class that does nothing is a
  claim that misleads the next reader.
- **The board states both axes explicitly** and owns its own scrolling. The comment at
  `Kanban.tsx:307-310` is corrected to describe what the code now does; a comment that
  contradicts the CSS is worse than no comment.
- **One focus helper, and only one.** A single `focusWithoutScrolling(element)` in
  `frontend/src/lib/focus.ts`. **Every** existing `.focus()` call routes through it —
  `Kanban.tsx:152`, `HabitStrip.tsx:85`, `QuickAdd.tsx:121,164,168,201,225`,
  `CommandPalette.tsx:93,95`, `AppearanceControls.tsx:164`.
- **`make guard` check 7**: `.focus(` appears in exactly one module. An exact grep, not a
  heuristic.
- **The killed theory stays killed.** Do **not** add `scrollLeft`/`scrollTop` handling. The
  "latched `scrollLeft`" explanation was investigated and ruled out (**D22**): columns are
  `flex-1 basis-0`, so `scrollWidth` tracks `clientWidth` and the UA clamps the offset. A
  comment naming it as ruled out is welcome; code for it is not.

**Acceptance criteria**
- [ ] `git grep -n '\.focus(' frontend/src` returns hits in **exactly one** non-test
      module. **`make guard` check 7 enforces it** — negative control: add a bare `.focus()`
      to a component, watch check 7 fail naming the file and line, remove it.
- [ ] A test asserts `focusWithoutScrolling` passes `{ preventScroll: true }` — driven
      through a spy on a real element, not by reading the source.
- [ ] A test asserts `<main>` carries `min-h-0` and that the shell, `#root`, `body` and
      `html` form an unbroken height chain, read off the rendered DOM and the stylesheet.
- [ ] The S2-16 keyboard suite and the S2-22 ACCEPT flow pass **unmodified**. This is a
      layout change; a changed keyboard test means something else moved.
- [ ] **Stated honestly:** jsdom has no layout engine, so **none** of the above proves the
      document stops scrolling or that the board scrolls instead. That is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 1, and the
      commit body must say so.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `fix(frontend): give the shell one height chain and stop focus scrolling (S3-01)`

---

## S3-02 — fix: no localised string may refuse to shrink (K8, D20)

**This is K8, ruled by D20.** `html { font-size: 13px }` (`style.css:38`) makes every
Tailwind rem 13/16 of nominal, so **`min-w-36` is 117px, not 144px**, and `gap-2`/`p-2` are
6.5px. At the floor, column content is ~102px and the card interior ~87px.
`DueBadge.tsx:43` is `shrink-0 font-mono` and takes its max-content width — **~94px in
English, ~125px in Russian, against 87px available. It overflows in English.** Same defect
at `Column.tsx:175` (the card-count span) and `HabitChip.tsx:69`. The user's screenshot
shows the consequence: headings wrapping **one character per line**, because an 86px
`shrink-0` count span starves a heading down to ~16px.

**The audit changes shape, and that is the point of this ticket.** Adding `shrink-0` to the
existing grep fixes this bug and leaves the next one. See **D20**.

**Scope (may touch):** `frontend/src/components/{DueBadge,Column,HabitChip,Card}.tsx`,
`frontend/src/App.accept.test.tsx` (the Russian audit at ~line 344), a new shared helper
under `frontend/src/test/`, tests beside them. **Not** `frontend/src/style.css` — the 13px
base is `design/README.md`'s and is not being changed.

Requirements:

- **Remove `shrink-0` from every element whose text is localised** — i18n output,
  `formatDate`, `formatNumber`: `DueBadge.tsx:43`, `Column.tsx:175`, `HabitChip.tsx:93`,
  `Card.tsx:107`. Give them `min-w-0` and let them wrap, per **D20**.
- **`shrink-0` stays where it is correct** and the distinction is stated in a comment:
  a **fixed-size, non-text box** — `TypeIcon`, `TimerDot`, `HabitChip`'s icons,
  `ProgressBar`'s track — has a size that does not depend on the locale.
  `PriorityChip` (`P0`…`P3`) and `ProgressBar`'s ratio are bounded ASCII; **state the
  ruling either way in a comment rather than leaving it to the next reader to re-derive.**
- **`Card.tsx:36-43`'s comment is corrected.** It currently asserts a property the card did
  not have. A comment that is confidently wrong is how this defect survived a review.
- **The Russian audit becomes a DOM walk**: render each screen in Russian and, for **every
  element carrying localised text**, assert that no shrink-refusing utility is on it —
  `shrink-0`, `flex-none`, a fixed `w-*`, `whitespace-nowrap`, `truncate`. A walk catches a
  *new* utility with the same effect on the day it is used; a word list does not.
- The audit must **fail on the pre-fix tree**. That is the negative control and it is not
  optional here: this is the second time a green audit covered a real clipping bug.

**Acceptance criteria**
- [ ] The rewritten audit, run against the tree **before** this ticket's component changes,
      **fails**, naming `DueBadge`, the column count span and the habit streak. Restore,
      confirm green. Record both runs in the commit body.
- [ ] `git grep -n 'shrink-0' frontend/src` — every surviving hit is on a fixed-size
      non-text box or a bounded-ASCII string, and the commit body lists them one by one
      with which of the two it is.
- [ ] The audit walks the **rendered DOM in Russian**, not the source text, and covers the
      board, the card, the habit strip, both overlays, the header and a toast.
- [ ] `make guard` check 6 still passes over `App.accept.test.tsx` — the audit added **no
      mouse event**.
- [ ] **Stated honestly:** jsdom has no layout engine, so this proves the **mechanisms**
      and **never** the absence of clipping. The onset widths in **K8** — RU heading
      ~784px, RU badge ~814px, EN badge ~659px — are **±5% arithmetic, not measurements**.
      Confirming the fix at a real 1024×768, in both languages, is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 2.
- [ ] `make front-test`, `make check` green.

**Commit:** `fix(frontend): let every localised string shrink, and widen the ru audit (S3-02)`

---

## S3-03 — fix: the window minimum size, derived from the layout floor (K9, D21)

**This is K9, ruled by D21**, and it comes third because **the floor is not final until
S3-01 and S3-02 have landed.**

`main.go:137-139` sets `Width: 1024, Height: 768` and no minimum. Wails calls
`SetMinSize(0, 0)` unconditionally (`window.go:134-135` → `window.c:266`), so GTK is hinted
with `min_width = min_height = 0` and the window can be dragged ~380px below the layout's
own ~624px floor. **The user changes display scaling often and it varies**, so the startup
window is not reliably 1024 CSS px.

**The rule this collides with is the hardest one in this repository**, and **D21** is the
containment: a `MinWidth: 640` in Go is the CSS floor written down a second time. The
precedent is **D12** — `main.go` may not name a colour, so `background.go` derives it from
`design/tokens.css`. Do the same with the floor.

**Scope (may touch):** `main.go`, a new `layout.go` and `layout_test.go` at the repo root
beside `background.go`. **No file under `frontend/` is modified** — the test *reads* frontend
sources, which is not touching them.

Requirements:

- `MinWidth` and `MinHeight` on `options.App`, computed — **not typed**.
- **Each input has exactly one home**, and the Go side derives or verifies rather than
  restating: the root font size (`frontend/src/style.css`), Tailwind's spacing unit, the
  column `min-w-*` unit count (`components/Column.tsx`), the board `gap-*` and shell `p-*`
  unit counts, and the **column count taken from the set Go already uses to build the five
  columns** — never a `5` typed into `main.go`.
- **The recommended mechanism, and the Dev may substitute a better one and say so:** a Go
  **test** parses `frontend/src/style.css` and the relevant component sources and asserts
  the constants in `layout.go` still match. Parsing at test time and not at run time keeps
  the binary free of file reads and keeps the frontend sources out of the embedded assets.
- **`MinHeight` is a stated composition of named terms**, and the ticket says which terms
  are estimates. A card's height depends on chip count and title wrapping, which is layout
  and cannot be computed here — **disclose it, do not round it up quietly** (**D21**).
- `main.go` gains **no magic number and no hex literal**; D12's grep
  (`git grep -nE '#[0-9a-fA-F]{3,8}|RGBA\{' main.go`) still returns nothing.

**Acceptance criteria**
- [ ] **The drift test is the criterion.** Change `min-w-36` to `min-w-32` in
      `Column.tsx` → the named Go test **fails**. Change `html { font-size: 13px }` → it
      **fails**. Restore both, confirm green, and record all four runs in the commit body.
- [ ] `git grep -nE 'MinWidth:|MinHeight:' main.go` shows each fed by a **named identifier**,
      never an integer literal.
- [ ] The column count is read from Go's published set; a test adding a sixth status moves
      the computed floor without any edit to `main.go`.
- [ ] The commit body shows the **arithmetic**, term by term, for both dimensions, and
      marks each estimated term as estimated.
- [ ] `make check` green, `make cover` green (this adds Go code — its tests land in the
      same commit, rule 14).
- [ ] **Stated honestly:** that GTK actually enforces the hint, and that the window is
      usable at exactly the minimum under a non-default display scaling, is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 3.

**Commit:** `fix(app): derive the window minimum size from the layout floor (S3-03)`

---

## S3-04 — fix(frontend): blur the few large surfaces, not every small one (K7, D19)

**This is K7, ruled by D19.** `backdrop-blur-glass` is on **every** surface — roughly fifty
at once: `Column.tsx:165`, `Card.tsx:91`, `HabitChip.tsx:69`, `QuickAdd.tsx:239`,
`CommandPalette.tsx:163`, `Toast.tsx:54` — and each is its own compositing layer. Resizing
breaks the layout and **it does not recover without a restart**.

**The cause is confirmed by the user's own discriminating experiment**, not by inference:
with the layout stuck broken, **switching the palette to Studio — whose `--blur` is `0px` —
repaired it live, with no restart.** That is WebKitGTK compositing-layer staleness, and it
eliminates the competing scrollbar-hysteresis theory, which no palette switch could have
affected.

**Scope (may touch):** `frontend/src/components/{Column,Card,HabitChip,QuickAdd,CommandPalette,Toast}.tsx`,
the `Makefile` (guard check 8 — declared here), tests beside them. **`design/` is read-only
and is not touched.**

Requirements:

- Apply **D19**'s allow-list exactly: `backdrop-blur-glass` **stays** on `Column.tsx`
  (5 instances, large), `QuickAdd.tsx` and `CommandPalette.tsx` (full-screen scrims, at most
  one at a time). It is **removed** from `Card.tsx`, `HabitChip.tsx` and `Toast.tsx`.
- **Translucency is kept.** The three losing surfaces keep their `bg-surface` /
  `bg-elevated` tokens, so Aurora is still translucent — only the `backdrop-filter` goes.
  **No new token, no palette conditional, no hex.**
- **`make guard` check 8**: `backdrop-blur-glass` appears **only** in those three files. An
  exact grep. This is what stops the class quietly returning to the card in Stage 4.
- The comments at `Card.tsx:46`, `Column.tsx:39` and `Toast.tsx:28` explain the surface
  treatment and must be **corrected to describe the new policy and to name D19** — three
  stale comments would be three future arguments.
- **Do not add a resize workaround.** No `requestAnimationFrame` nudge, no forced reflow, no
  transform toggle. **D19** rejects them by name.

**Acceptance criteria**
- [ ] `git grep -n 'backdrop-blur-glass' frontend/src` returns hits in **exactly** the three
      allow-listed files. **`make guard` check 8 enforces it** — negative control: put the
      class back on `Card.tsx`, watch check 8 fail naming the file and line, remove it.
- [ ] A test asserts the card, the habit chip and the toast still carry their surface token,
      so the fix removed a filter and not the surface.
- [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` still returns nothing, and `design/`
      is byte-identical.
- [ ] The S2 component tests pass with only the class-name assertions that named
      `backdrop-blur-glass` updated — and the commit body lists each one changed, so nothing
      is weakened under cover of a class rename.
- [ ] **Stated honestly, and this is the important one:** **nothing here proves the resize
      bug is fixed.** There is no display on this machine. The real check is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 4 — resize
      repeatedly, at more than one display scaling, in both palettes. **D19**'s escalation
      ladder applies if it still reproduces: next drop the column's blur too, and after
      that it is a **user decision**, because the palette's specified appearance is what
      would be given up.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `fix(frontend): blur only the column and the overlays (S3-04)`

---

## S3-05 — fix(frontend): the dragged card follows the pointer (K6, D18)

**This is K6, ruled by D18**, and it is the most visible defect the user hit: **the card
disappears the instant it is grabbed.**

`views/Kanban.tsx:294-306` mounts `DndContext` with **no `DragOverlay`**. Confirmed against
the installed `@dnd-kit/sortable@10.0.0`: `sortable.cjs.development.js:314` computes
`useDragOverlay = Boolean(dragOverlay.rect !== null)` → **false**, so
`shouldDisplaceDragSource` is true and the source `<li>` is translated in place. Two
independent consequences: the next column's `backdrop-filter` stacking context **paints over
it** (and `z-10` at `Column.tsx:126` can only order siblings within one column), and at
`sortable.cjs.development.js:514-524` the per-column `SortableContext` (`Column.tsx:187`)
makes `overIndex` `-1` once the pointer leaves, so `finalTransform` becomes `null` and the
card **teleports home and stops following the pointer**.

**Scope (may touch):** `frontend/src/views/Kanban.tsx`,
`frontend/src/components/{Column,Card}.tsx`, `frontend/src/store/ui.ts`, tests beside them.

Requirements:

- A **`<DragOverlay>` as a direct child of `DndContext`**, rendering the **existing
  `<Card>`** for the active node. **Not** a second preview component — one card component,
  or the two will drift.
- The source `<li>` is **hidden while `isDragging`** (`opacity-0`) and **keeps its space**,
  so the list does not jump on grab.
- **Delete the now-dead `z-10`** at `Column.tsx:126`. It only ever existed to fight the
  stacking context, it never could, and leaving it is leaving a false explanation in the
  code.
- **`draggingNodeId` (`store/ui.ts:28`) gets its first consumer.** It has been written and
  read by nothing since S2-17. The overlay reads it. If the overlay ends up not needing it,
  **say so and delete it** — a field with no reader is a second spelling waiting for a
  caller.
- **Reduced motion routes through the existing helper**: `dropAnimation={null}` when
  `prefersReducedMotion()` is true, from the same `lib/appearance.ts` function
  `Column.tsx:116` already calls. Not a second media query.
- **S2-17's contract is unchanged**: cross-column drop still calls the coupled
  `MoveToColumn`, in-column reorder still calls `MoveNode`, optimistic-with-rollback still
  restores from Go's answer, and S2-16's keyboard map still works untouched.

**Acceptance criteria**
- [ ] With a drag in progress, the overlay renders the active card **once** and the source
      `<li>` is present, space-holding and hidden. Asserted on the rendered tree.
- [ ] `draggingNodeId` has a reader — `git grep -n 'draggingNodeId' frontend/src` shows a
      consumer outside `store/`, or the field is gone and the commit body says why.
- [ ] `git grep -n 'z-10' frontend/src/components/Column.tsx` returns nothing.
- [ ] Under `prefers-reduced-motion: reduce`, `dropAnimation` is `null`, decided by the
      **existing** helper — a test that stubs the media query both ways.
- [ ] The whole S2-17 drag suite and the S2-16 keyboard suite pass **unmodified**.
- [ ] `make guard` check 6 is untouched — this ticket does not go near
      `App.accept.test.tsx`; drag tests are allowed a pointer, because a pointer is what is
      under test.
- [ ] **Stated honestly:** jsdom renders no pixels, so *"the card visibly follows the
      pointer across a column boundary and is not painted over"* is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 5. What is
      mechanical here is the overlay's existence, the source's hidden state, and the
      unchanged call contract.
- [ ] `make front-test`, `make check` green.

**Commit:** `fix(frontend): render the dragged card in a DragOverlay (S3-05)`

---

## S3-06 — feat(frontend): a Cyrillic face under the same family names, and `<html lang>` (K11, D23)

**This is K11, and D23 is the USER'S ruling** — it is not a PM call and may not be
re-litigated by the Dev.

`design/tokens.css:18` sets Aurora's `--font-ui` to `'Space Grotesk'`, `:56` sets Studio's
to `'Figtree'`. Verified from `node_modules/@fontsource/*/unicode.json` **and from the
shipped `frontend/dist/assets/`**: Space Grotesk covers `[vietnamese, latin-ext, latin]`,
Figtree covers `[latin-ext, latin]`. **Neither contains a single Cyrillic glyph.** Only
JetBrains Mono does. **81 of `ru.json`'s 82 leaves are Cyrillic**, so Russian drops the
whole UI into the system `sans-serif` while numbers stay in JetBrains Mono — two typefaces
in one header, different metrics, and a `flex-wrap` header very plausibly gaining a row.
**This is the user's reported "theme and colour buttons break when I switch language".**

> **✅ UNBLOCKED — OQ1 is answered. The face is Inter, and that is the user's choice.**
> **One** face, vendored under **both** family names via `unicode-range`, so `design/` is
> not edited and Latin keeps its designed typeface. One face means one `unicode-range`
> block and one coverage test.
>
> **The name is not the evidence.** Before committing, the Dev **verifies Inter's real
> coverage from the `@fontsource` package's own `unicode.json`** — including `U+2116` (№)
> and the combining acute `U+0301` — rather than trusting the face's reputation or the
> suggested range below. If the package does not in fact cover a codepoint `ru.json` uses,
> **stop and report**; do not narrow the test to match the font.

**Scope (may touch):** `frontend/src/style.css`, `frontend/src/lib/appearance.ts` and its
test, `frontend/index.html` (comment only), `frontend/package.json` and the lockfile, a new
coverage test under `frontend/src/locales/`, tests beside them. **`design/` is read-only and
is NOT edited** — that constraint is the whole reason D23 has the shape it has.

Requirements:

- **`@font-face` blocks in `frontend/src/style.css` registering Inter under BOTH
  existing family names** — the same strings `design/tokens.css` names — scoped with
  `unicode-range`. Suggested range, to be confirmed against the package's real coverage:
  `U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1, U+2116`.
- **Vendored and local.** `@fontsource/inter`, bundled from `node_modules`, fingerprinted
  into `dist/assets` like the three existing faces, **loaded from disk**. **No CDN, no
  `<link>`, no runtime fetch** — rule 11, and it is a hard project rule, not a preference.
- **A new dependency, and it is declared as one.** `@fontsource/inter` is the stage's first
  new package; the commit body **names it explicitly** as a new dependency, with the
  bundle-size delta, per the dependency table above.
- **A coverage test**: every codepoint used in `ru.json` is covered by some **bundled**
  `unicode-range` for `--font-ui`. Read the ranges from what is actually bundled; a range
  retyped into the test proves the test agrees with itself.
- **`<html lang>` is written from settings, in one place.** `frontend/index.html:14`
  hard-codes `lang="en"` and **nothing ever updates it** — grepped, zero writers.
  `lib/appearance.ts` already receives the whole `SettingsView`, already runs at boot and on
  every settings write, and is already the only module that writes to `documentElement`.
  **`root.lang = appearance.language` belongs there.** The static `lang="en"` stays as the
  pre-boot default, exactly as `data-palette="aurora"` does, and the comment says so.
- **No second spelling of the language value.** `i18next` and `root.lang` both come from the
  same `SettingsView` field; neither derives the other.

**Acceptance criteria**
- [ ] The coverage test **fails** if the `unicode-range` is narrowed, and **fails** if a
      Cyrillic string is added to `ru.json` outside the covered ranges. Negative control
      both ways, recorded in the commit body.
- [ ] `document.documentElement.lang` is `ru` after the settings read reports Russian and
      `en` after it reports English — asserted through `render(<App />)`, and with **exactly
      one writer** in `frontend/src` (`git grep -n '\.lang *=' frontend/src`).
- [ ] `git diff -- design/` is empty, and `--font-ui` is **not** redefined anywhere in
      `frontend/src`.
- [ ] `git grep -nE 'https?://' frontend/src` still returns nothing, and
      `@fontsource/inter`'s files resolve from `node_modules` — the bundle is inspected and
      the commit body names the emitted woff2 files. **No network fetch at runtime**, which
      is the acceptance criterion the dependency table demands and not a hope.
- [ ] **Inter's real coverage was read from
      `node_modules/@fontsource/inter/unicode.json`** — not assumed from the face's name —
      and the commit body quotes the subsets it found, including `U+2116` and `U+0301`.
- [ ] `make front-test`, `make check` green; the build output size delta is recorded in the
      commit body.
- [ ] **Stated honestly:** jsdom has **no font engine**. Nothing here proves the Cyrillic UI
      actually renders in the designed family, or that the header stops wrapping. That is
      [S3-09](#s3-09--test-the-hand-pass-that-closes-the-blocking-block) item 6, in both
      palettes.

**Commit:** `feat(frontend): vendor a cyrillic ui face and set html lang from settings (S3-06)`

---

## S3-07 — fix(frontend): toasts are capped, de-duplicated and self-dismissing (K12, D24)

**This is the first half of K12, ruled by D24.** `store/toast.ts:35-47` appends
unconditionally — **no cap, no de-duplication, no auto-dismiss**. The user's screenshot
shows three identical toasts filling the lower half of the window and covering the board.

**The second half — what the message actually says — is
[S3-08](#s3-08--feat-a-refusal-is-not-a-failure-and-every-toast-says-what-failed-k12-d25),
deliberately kept separate.** One is a list policy, the other is a message contract.

**Scope (may touch):** `frontend/src/store/toast.ts` and its test,
`frontend/src/components/Toast.tsx` and its test, the locale files, tests beside them.

Requirements:

- **A cap of three.** A fourth push drops the oldest.
- **De-duplication of an identical consecutive `messageKey`**: increment a count on the
  existing toast rather than appending. The count is a **number**, so `font-mono`, and it is
  **pluralised through i18next** — Russian has three plural forms, and `ru.json` already
  carries `_few`/`_many` for the card count, which is the pattern to follow.
- **Auto-dismiss after a bounded interval**, **paused while the toast has focus or the
  pointer is over it** — a toast that vanishes while being read is a new defect.
- **The store still holds no text.** `store/toast.ts:5-12`'s design intent is preserved
  exactly: an **i18n key, never a sentence**, and the raw Go error to the console only.
- **Timers must be testable and must not leak.** Fake timers in the test; the dismiss timer
  is cleared on unmount and on manual dismiss.
- The interval and the cap are **named constants in one place**, not numbers sprinkled
  through the component and the store.

**Acceptance criteria**
- [ ] Four distinct pushes leave three toasts, oldest gone. Three identical pushes leave
      **one** toast showing a count of 3, in `font-mono`, pluralised correctly in both
      languages.
- [ ] A toast dismisses itself after the interval, under fake timers; the timer is **paused**
      while it holds focus, and resumes after blur.
- [ ] Manual dismiss and unmount both clear the timer — asserted, so the suite does not leak
      handles.
- [ ] `git grep -n 'pushToast' frontend/src` shows every call still passing a **key**;
      `store/toast.ts` contains no user-visible sentence.
- [ ] The toast region is still keyboard-reachable and still announces (S2-13's a11y
      behaviour is unchanged) — the existing `Toast.test.tsx` assertions survive.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `fix(frontend): cap, de-duplicate and auto-dismiss toasts (S3-07)`

---

## S3-08 — feat: a refusal is not a failure, and every toast says what failed (K12, D25)

**This is the second half of K12, ruled by D25.** Every rejection from Go raises the single
key `toast.error.body` — *"Nexus could not finish that. Nothing was changed."* So **D9**
declining to put a project into `doing` — **a rule working exactly as specified** — is
rendered as a malfunction. And after the fact, **we still cannot tell which three operations
failed in the user's screenshot.** Both halves are this ticket's.

**Scope (may touch):** a new `internal/service/refusal.go` and its test, `app.go`,
`frontend/wailsjs/**` (regenerated, committed in the same commit — rule 12),
`frontend/src/store/call.ts`, `frontend/src/store/**` (the call sites that name their
operation), `frontend/src/components/Toast.tsx`, the locale files, tests beside them.

Requirements:

- **One Go table, sentinel → stable code**, and nothing else classifies anything. The
  inventory exists already: `ErrProjectNeverDoing`, `ErrTypeHasNoColumn`, `ErrTypeHasNoDue`,
  `ErrTypeHasNoChildren`, `ErrCircularParent`, `ErrNoRecurrence`, `ErrUnsupportedRecurrence`,
  `ErrTimerNotAllowed`, `ErrNodeArchived`, `ErrNodeDone`, `ErrNotAHabit`,
  `ErrInvalidSetting`. An error in **no** table is a **failure**, not a refusal, and keeps
  today's key.
- **The transport, and the Dev may substitute a cleaner Wails-native channel if they
  disclose it:** Wails marshals `error` into a rejected promise carrying the message string,
  so the code travels **behind a fixed machine prefix** produced by **exactly one** Go
  function and parsed by **exactly one** TypeScript function. Every bound method still
  returns `(T, error)`.
- **The frontend classifies nothing and parses no prose.** It maps **code → i18n key** — a
  label table, the same shape the five Kanban columns already have and the shape **D26**
  moves the other sets to. **Matching on Go's English error text is forbidden**: a message
  is not an API.
- **Every toast names the operation.** Each store action passes an operation key
  (`move`, `create`, `check habit`, `set priority`, …) alongside the outcome, so a stack of
  three toasts is legible. This is the diagnostic half and it is a requirement, not a
  nicety.
- **Exhaustive or red.** A code Go can emit with no key in `en.json`/`ru.json` fails a test.
- **A refusal is styled as information, not as an error.** Token semantics: `danger` stays
  for failures; a refusal uses a non-alarming surface. No new hex, no new token.

**Acceptance criteria**
- [ ] A test drives a real refusal end to end — `MoveToColumn(project, doing)` (**D9**) —
      and asserts the toast shows the **specific** localised refusal message in both
      languages, and **not** `toast.error.body`.
- [ ] An error that is in no table still raises exactly one `toast.error.body`, unchanged.
      S2's *"every rejection reaches a toast"* criterion is not weakened.
- [ ] **Exhaustiveness test**: adding a sentinel to the Go table without adding a key turns
      a named test red. Negative control, recorded in the commit body.
- [ ] `git grep -n` finds **one** Go function producing the code and **one** TypeScript
      function parsing it; no component and no store action does either.
- [ ] Every toast raised anywhere in the app names its operation — asserted by driving one
      failure per store action and checking the rendered text differs per action.
- [ ] `frontend/wailsjs` is regenerated and committed in this commit;
      `git status --porcelain` is empty after `make check`.
- [ ] `make cover` green, `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(service): give refusals their own code and name the failed operation (S3-08)`

---

## S3-09 — test: the hand pass that closes the blocking block

**This ticket is a human at a real keyboard in front of a real window.** It is not
automatable here and pretending otherwise is the one failure this project will not accept.
There is **no display on this machine**, and `xvfb-run`, `scrot`, `import` and `grim` are
**all absent**.

**It is also the gate into block B.** Rule 18: nothing from S3-10 onwards starts until this
is reported.

**Scope (may touch):** `internal/service/task_test.go` and a new
`frontend/src/components/DueBadge.test.tsx` — **test files only, and only for the D8 item
below. No production file.** The rest of this ticket is a report.

### Part 1 — the D8 due-badge claim finally gets the mechanical halves it can have

Since Stage 2 this has been the honest gap: the ACCEPT flow's fake client **does not model
D8's due rewrite**, so *"the badge reads the upcoming Friday"*, *"the badge reads today"*
and *"the date is cleared on the way back"* have **no mechanical backing anywhere**. The
fake must **not** be taught the rule — a fake that implements D8 is D8 with a second
spelling, which is the defect this project keeps paying for. Instead, two halves that are
each honest:

- **Go's half**: a service-level test that moves a node to This week, then to Today, then
  back to Backlog, and asserts the `due` and `due_source` that come back from `Board()` —
  the upcoming Friday, today (with the Friday edge case from `PLAN.md` §4), and cleared.
- **The frontend's half**: `DueBadge` renders **the DTO's date**, formatted, and nothing
  else — reinforcing what `Card.test.tsx` already pins for `overdue`.

**What remains hand-only after both**: that the two halves meet across the real IPC bridge
in a real window. Say exactly that, and no more.

### Part 2 — the hand script

Build and launch, then report **every item, pass or fail, in the ticket's commit body or
report**:

```sh
make build && ./build/bin/nexus
```

| # | What to do | What must be true | Closes |
|---|---|---|---|
| 1 | Look at the window. Resize the board content past the bottom | The **board** scrolls; the **document** does not. No document-level scrollbar. Tab through every region — **the board never jumps or scrolls when focus moves** | **K10** / S3-01 |
| 2 | At a real 1024×768, in **English** and then in **Russian** | Nothing clips. Column headings do **not** wrap one character per line; the due badge and the card-count are fully readable; the habit chip fits | **K8** / S3-02 |
| 3 | Drag the window edge inward as far as it will go, at **more than one display scaling** | It stops at the minimum and the layout is still usable there — not broken, not clipped | **K9** / S3-03 |
| 4 | Resize the window repeatedly, in **both palettes**, at **more than one display scaling** | The layout never sticks broken, and never needs a restart. **If it still does**, report it and apply **D19**'s escalation ladder — do not quietly add a repaint hack | **K7** / S3-04 |
| 5 | Grab a card and drag it slowly across every column boundary | The card is **visible under the pointer the whole time**, is not painted over by the next column, and does not teleport home. Drop works; a rejected drop rolls back | **K6** / S3-05 |
| 6 | Switch to Russian, in **both palettes** | The whole UI is in the designed family — **not** the system `sans-serif`. Numbers and dates are still JetBrains Mono. The header does not gain a row. `document.documentElement.lang` is `ru` | **K11** / S3-06 |
| 7 | Force several failures in a row, and one real refusal (drag a project to Doing) | At most three toasts; identical ones collapse with a count; each dismisses itself; **each says which operation failed**; the refusal reads as a rule, not as a malfunction | **K12** / S3-07, S3-08 |
| 8 | **Steps 1–10 of [Stage 2's hand script](#half-2--by-hand-on-the-real-binary-mouse-untouched)**, mouse untouched, reported step by step | As written there — **including the three D8 due-badge assertions at steps 3, 4 and 7** | Stage 2 hand-pass item 1 |
| 9 | Quit and relaunch, in **both** themes | **No flash of the wrong background** on the first frame (**D12**, **K1**) | Stage 2 hand-pass item 2 |
| 10 | Tab around, in **both palettes** | The `:focus-visible` accent ring is **actually painted** and is visible against both | Stage 2 hand-pass item 4 |

**Acceptance criteria**
- [ ] Part 1's two tests are committed and green, and the commit body states **precisely
      what they do and do not prove** — specifically that the bridge is still hand-only.
- [ ] The fake client was **not** taught D8. Confirmed by diff.
- [ ] **All ten items above are reported individually, pass or fail.** A silent item is a
      failed ticket. A failure is a normal outcome and is reported as one, with what was
      seen.
- [ ] Any item that fails produces a **named follow-up ticket** before block B starts — or
      an explicit, recorded decision by the user to proceed anyway.
- [ ] `make check`, `make cover`, `make front-test`, `make guard` all green at this commit —
      **this is the point where block A is declared closed.**

**Commit:** `test: back the d8 badge claim and run the block a hand pass (S3-09)`

---

### Block B — the feature stage (S3-10 … S3-29)

**Do not start any of these until [block A](#block-a--the-blocking-defect-block-s3-01--s3-09)
is closed by S3-09.**

**S3-10 … S3-19 are Go** and touch no file under `frontend/src`; each commits the
regenerated `frontend/wailsjs` if it changes a bound signature (rule 12), and each lands
with its tests in the same commit (rule 14 — `make cover` has no headroom to spend).
**S3-20 … S3-28 are the screens.** **S3-29 is the README.**

**S3-19 is new, and it is why block B has nineteen numbers where it had eighteen.** The
user answered **OQ2** with the middle option, so `COUNT` and `UNTIL` enter the recurrence
language and `internal/domain/recurrence.go` has to accept them before
[S3-23](#s3-23--featfrontend-the-recurrence-editor-d27-d28) can offer them.

**The prime directive has not moved.** The frontend renders what Go returns and computes
nothing — and block B adds four new places that will be tempted: the Markdown preview
(rendering, not deriving), the recurrence editor (**it must not parse, validate or expand an
RRULE in TypeScript** — **D27**), **whether a bounded recurrence has ended** (that is Go's
`HabitView.Ended`, **D28**; a frontend that compares `UNTIL` to today is `wireDate` again),
and the running timer clock (it renders Go's
`elapsedSeconds`, and Stage 2 deleted the wall-clock helpers that sat on top of it for
exactly this reason).

---

## S3-10 — feat(service): `NodeDetail` — the one read the panel renders

**Scope (may touch):** `internal/service/dto.go`, `internal/service/read.go`,
`internal/service/read_test.go`, `app.go`, `frontend/wailsjs/**`.

Requirements:

- One read, `NodeDetail(id)`, returning everything the slide-over shows: the node's own
  fields, its **derived** status, progress and overdue flag, its tags, its children as the
  same view the board uses, its time entries with a total, its attachments, its recurrence
  as a **structure** (never a sentence), and its ancestor chain for a breadcrumb.
- **Every derived value is on the DTO.** If the panel will show it, Go computes it —
  including anything the tree/search screens will reuse.
- **JSON tags per S2-02's wire contract.** No field ships with Go's exported name.
- Built on the **existing** `Index` from S2-01: one index per call, no second traversal.

**Acceptance criteria**
- [ ] A table-driven test covers a task, a leaf project, an empty project
      (`progress.defined == false`, **D15**), a habit, a note and a bug.
- [ ] The derived status and progress come from the **same** `domain` entry points the board
      uses — `git grep` shows no second derivation.
- [ ] Every DTO field has a JSON tag; the generated `models.ts` is committed.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add the node detail read (S3-10)`

---

## S3-11 — feat(service): the field writers — title, description, due, estimate, activity

**Scope (may touch):** `internal/service/task.go`, `internal/service/task_test.go`,
`app.go`, `frontend/wailsjs/**`.

Requirements:

- `SetTitle`, `SetDescription`, `SetDue`, `SetEstimate`, `SetActivity`, each built to the
  shape `SetDue`/`SetPriority` already have: read, apply **one** field, let
  `domain.Node.Validate` decide, write, read back. A refused value writes nothing.
- **D1 holds:** any *user* edit of `due` sets `due_source = 'manual'`. That rule already has
  one spelling; do not add a second.
- **`HasDue` holds:** a `note` is refused a due date (`ErrTypeHasNoDue`). Ask the predicate;
  do not restate it.
- `activity` takes the fixed list from **D4**, which `domain` already owns.
- **One entry point per field, reachable from `app.go`.** Two setters for one field, one
  validating and one not, is a second spelling with a call-site-shaped fuse (**D13**'s
  lesson).

**Acceptance criteria**
- [ ] Each setter round-trips: write, read back, the stored value is Go's normalised one.
- [ ] A due date set by hand reports `due_source == 'manual'`, and a subsequent column move
      overwrites it and flips it to `auto` (**D8**) — asserted in one test, because that is
      the pair that confused Stage 1.
- [ ] `SetDue` on a `note` is refused with `ErrTypeHasNoDue` and writes nothing.
- [ ] Every new bound method returns `(T, error)`; `frontend/wailsjs` committed.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add the detail-panel field writers (S3-11)`

---

## S3-12 — feat(service): the type switcher

**Its own ticket because it is the rules-heaviest write in the stage.** Changing a node's
type can invalidate its status, its due date and its right to have children — three rules
that already exist and **must not be restated here**.

**Scope (may touch):** `internal/service/task.go`, `internal/service/task_test.go`,
`internal/domain/node.go` only if a predicate genuinely needs widening (**stop and report
before doing so**), `app.go`, `frontend/wailsjs/**`.

Requirements:

- `SetType(id, type)`, deciding every consequence through the **existing** predicates:
  `NodeType.HasColumn` (a type with no column may not hold a column status and may not have
  children — `ErrTypeHasNoChildren`), `NodeType.HasDue`, `DoingRefusal`/`CanBeDoing`
  (**D9**), and the habit-requires-recurrence check (`ErrNoRecurrence`).
- **No new predicate, and no new spelling of an old one.** If a rule is missing, that is a
  **stop-and-report**, not a local `if`.
- The refusals surface through **S3-08**'s refusal codes, so the panel can say *why*.

**Acceptance criteria**
- [ ] A matrix test over every from-type × to-type pair, asserting the outcome and the
      sentinel. Illegal transitions write **nothing**.
- [ ] Switching a node **with children** to a no-column type is refused with
      `ErrTypeHasNoChildren`; switching to `habit` without a recurrence is refused with
      `ErrNoRecurrence`; switching to `note` with a due date is refused or clears it —
      **whichever the existing predicates already imply, and the test says which and why**.
- [ ] `git grep` shows the predicates called, not re-implemented.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add the node type switcher (S3-12)`

---

## S3-13 — feat(service): tags

**Scope (may touch):** `internal/service/tag.go` (new) and its test, `internal/store`'s tag
repository (S1-14, already built), `app.go`, `frontend/wailsjs/**`.

Requirements:

- List, create, rename, recolour, delete, attach, detach — over the `tags` / `node_tags`
  schema Stage 1 already migrated.
- **Deleting a tag detaches it everywhere**, in one transaction.
- **A tag's colour is chosen from the accent set Go publishes** (**D26**, S3-18), not a free
  hex string. This is what keeps rule 9 — *no hex literal in `frontend/src`* — true when the
  tag picker ships. The stored column keeps whatever it stores; the **offered set** is Go's.
- Name uniqueness and trimming are decided in `domain`, once.

**Acceptance criteria**
- [ ] Attach/detach is idempotent; deleting a tag removes every `node_tags` row, asserted.
- [ ] A colour outside the published set is refused and writes nothing.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add tag management (S3-13)`

---

## S3-14 — feat(service): attachments copied into the app data dir

**Scope (may touch):** `internal/service/attachment.go` (new) and its test,
`internal/store`'s attachment repository (S1-16), `app.go`, `frontend/wailsjs/**`.

Requirements:

- An attachment is **copied into the app data directory**, never linked. The original may be
  moved or deleted by the user and the attachment must survive it.
- The stored path is **relative to the data dir**; a read that resolves outside the data dir
  is refused. Path traversal is a rule, and it has one spelling.
- MIME is detected from content, not from the extension.
- **Local-only:** no fetch, no upload, no preview service. A file picker comes from the
  Wails runtime; no new dependency.

**Acceptance criteria**
- [ ] Copying, then deleting the source, leaves the attachment readable.
- [ ] A crafted relative path escaping the data dir is refused, asserted.
- [ ] Removing an attachment removes both the row and the file, in that order, with the file
      removal failure not orphaning the row.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): copy attachments into the app data dir (S3-14)`

---

## S3-15 — feat(service): the editable time log

**Scope (may touch):** `internal/service/timer.go`, `internal/service/timer_test.go`,
`internal/store`'s time-entry repository (S1-15), `app.go`, `frontend/wailsjs/**`.

Requirements:

- List a node's entries; edit an entry's start and end; insert an entry by hand; delete one.
- **The single-active invariant is global and is not weakened** (§4, S1-19): an edit may not
  produce a second open entry, and a hand-inserted entry may not overlap an open one.
- **Minutes are stored as measured.** D4's rounding to the nearest 15 minutes is the **PMP
  output format** (Stage 6), not storage. Rounding here would make Stage 6's numbers
  unreconstructable.
- An entry ending before it starts is refused in `domain`, once.

**Acceptance criteria**
- [ ] A table-driven test over edit/insert/delete, including every way to attempt two open
      entries — each refused, nothing written.
- [ ] The S1-19 timer suite passes **unmodified**.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): make the time log editable (S3-15)`

---

## S3-16 — feat(service): search with tag/type/status/date filters

**Scope (may touch):** `internal/service/search.go` (S1-22) and its test, `internal/store`'s
search implementation (S1-17), `app.go`, `frontend/wailsjs/**`.

Requirements:

- Filters on tag, type, **derived** status, and a due-date range, composable with the text
  query.
- **Whatever backend S1-04 chose stands** — FTS5 if available, `LIKE` over
  `title` + `description_md` otherwise. **No cgo, ever**, and the fallback is pre-approved.
- **Filtering happens in SQL where it can and in `domain` where it must.** A *derived*
  status is not a column, so say explicitly which filters are pushed down and which are
  applied after derivation — and make the ordering stable either way.
- Archived nodes are excluded unless asked for (S3-17 is the view that asks).

**Acceptance criteria**
- [ ] A test asserts each filter alone and all four combined, including a derived-status
      filter over a parent whose stored status disagrees — the case that must not regress.
- [ ] Result ordering is deterministic, asserted over a fixture with ties.
- [ ] `go list -deps ./... | grep -i mattn` prints nothing; `CGO_ENABLED=0 go build ./...`
      succeeds.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add search filters (S3-16)`

---

## S3-17 — feat(service): the archive list and restore

**Scope (may touch):** `internal/service/read.go`, `internal/service/task.go` and their
tests, `app.go`, `frontend/wailsjs/**`.

Requirements:

- A read listing archived nodes with enough context to identify one — title, type, parent
  chain, `archived_at`.
- Restore already exists (S1-18) and **needs no new rule**: once a node has
  column-bearing children again, derivation takes over and the stored status stops being
  consulted (**D14**'s closing note).
- **D14 is not re-implemented here.** Archiving already re-inspects a node it turned into a
  leaf (S2-06). This ticket reads and restores; it does not touch that rule.

**Acceptance criteria**
- [ ] Archiving a subtree lists it once, by its root, not once per descendant — or lists
      every node, if that is the choice; **the ticket's test states which and why**.
- [ ] Restore → the node reappears on the board in its **derived** column, asserted.
- [ ] S2-06's D14 test passes **unmodified**.
- [ ] `make cover` green; `make check` green.

**Commit:** `feat(service): add the archive list (S3-17)`

---

## S3-18 — feat(service): publish every enum set (C6, D26)

**This closes C6, by D26** — the first of C6's two recorded options, because it **removes**
the second spelling instead of detecting it.

**An honest correction that the ticket must carry:** a fresh audit found current locale
parity **clean** — 82 RU leaves against 80 EN, the two extra being
`board.column.cardCount_few` and `_many`, the CLDR plural forms Russian requires, and
`locales.test.ts` already enforces parity. **C6 is a drift risk for future enum values, not
a present defect.** Nothing is wrong on screen today; what is missing is the thing that
would go red if it became wrong.

**Scope (may touch):** `internal/service/settings.go` and its test, `app.go`,
`frontend/wailsjs/**`, `frontend/src/lib/commands.ts`,
`frontend/src/components/AppearanceControls.tsx`, `frontend/src/locales/*.json`,
`frontend/src/locales/locales.test.ts`, tests beside them.

Requirements:

- Go publishes each set — **statuses, node types, themes, palettes, accents, priorities** —
  over the wire, the way the five Kanban columns already are.
- **The frontend enumerates what Go enumerates**; the locale files are reduced to **labels**.
  `Object.keys` of a locale table stops being an authoritative set anywhere in
  `frontend/src`.
- **The criterion C6 states is the criterion:** adding a value to a Go set must turn
  something red until the frontend offers it.
- The same shape serves **S3-08**'s refusal codes. One mechanism used twice, not two.

**Acceptance criteria**
- [ ] Negative control, per set: add a value to the Go set, watch the **named** test fail,
      remove it. Six sets, six demonstrations, recorded in the commit body.
- [ ] `git grep -n 'Object.keys' frontend/src` shows **no** locale table used as a set.
- [ ] The command palette and the appearance controls still offer exactly what they offered
      before — the existing S2-20 and S2-21 suites pass with only the source of the set
      changed, and the commit body says so file by file.
- [ ] `make cover`, `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(service): publish the enum sets and reduce the locales to labels (S3-18)`

---

## S3-19 — feat(domain): `COUNT` and `UNTIL` enter the recurrence language (D27, D28)

**This ticket exists because the user answered OQ2 with the middle option**, not with the
PM's recommendation. **D27 is the user's ruling** and is not re-litigated by the Dev: the
editor will offer *"repeat N times"* and *"repeat until date"*, so
`internal/domain/recurrence.go` must **accept** `COUNT` and `UNTIL` instead of rejecting
them. **D28** is the PM ruling that spells out the mechanics and the **D5** interaction;
read both before starting.

**[S3-23](#s3-23--featfrontend-the-recurrence-editor-d27-d28) is blocked on this ticket**
and may not widen the parser itself. This is the split the old recurrence-editor ticket
already demanded if OQ2 came back B: *"the domain half comes first."*

**Scope (may touch):** `internal/domain/recurrence.go`, `internal/domain/streak.go` **only
if a bound actually forces it**, `internal/domain/recurrence_test.go`,
`internal/domain/streak_test.go`, `internal/domain/sweep_test.go` or a sibling sweep file,
`internal/service/habit.go` and `internal/service/dto.go` for the `Ended` field, their
tests, `app.go`, `frontend/wailsjs/**`, and the locale files **for the one new strip
label only**.

Requirements:

- **`COUNT=n` is accepted**, `n` in `1..1000`. It bounds the series positionally: the
  series is the first `n` occurrences counting from `DTSTART` inclusive. The upper bound
  is there for the same reason `maxWindowDays` and `maxSearchDays` are — a positional
  bound is walked, so something has to stop the walk. `COUNT=0` and a non-numeric `COUNT`
  stay `ErrUnsupportedRecurrence`.
- **`UNTIL=YYYYMMDD` is accepted** and is an **inclusive calendar-day** upper bound.
  **The `DATE-TIME` form (`UNTIL=20261231T000000Z`) stays rejected at parse time** —
  Nexus has no time of day anywhere, `habit_checks` is keyed by a `Date`, and accepting a
  `Z` instant would need a timezone rule this project does not have. **The existing
  rejection-table row for `UNTIL` is the `DATE-TIME` form and therefore keeps passing
  unchanged**; it is relabelled to say *why* it is rejected, not deleted.
- **`COUNT` and `UNTIL` together are rejected.** RFC 5545 says they must not both appear,
  and two bounds on one series is two spellings of one rule.
- **The bound is part of the rule, not part of the clock.** `Matches`, `Expand`, `Next`
  and `Previous` all respect it with no reference to today; `internal/domain` stays pure
  and **no `time.Now()` appears** (rule: anything time-dependent takes an injected
  `now func() time.Time`). **`UNTIL` is a date bound, and it is compared against the
  candidate occurrence date, never against the current date.**
- **`BYSETPOS`, `BYMONTH`, `BYWEEKNO`, `BYYEARDAY`, `BYHOUR` and ordinal weekdays
  (`BYDAY=2MO`, `BYDAY=-1FR`) stay rejected at parse time.** That boundary has exactly one
  spelling — `TestParseRecurrenceRejectsWhatItCannotExpand` — and this ticket narrows it
  by exactly two rows and not one more.
- **The rejection assertions MOVE, they are not deleted.** `{"COUNT",
  "FREQ=DAILY;COUNT=10"}` moves from the rejection table into
  `TestParseRecurrenceAcceptsTheSupportedSubset` with its parsed value asserted; the
  `UNTIL` row stays where it is (it is the `DATE-TIME` form) and a new
  `UNTIL=YYYYMMDD` case joins the accept table. **A rejection case that simply disappears
  from the diff is a review failure.**
- **D28's D5 consequence is implemented by the bound and by nothing else.** Once the
  series has ended there are **no further scheduled occurrences**, so by **D5** no
  occurrence can pass unchecked and **the streak cannot break: it freezes at its final
  value**. `Streak`/`StreakOf` walk backwards through `Previous`, so this must fall out of
  bounding `Previous` and `Matches` — **if it needs a second code path in `streak.go`,
  that is the signal that the bound is in the wrong place.** D5 is **not amended** and no
  new streak rule is written down.
- **`HabitView` gains exactly one new Go-computed field** — `Ended bool` — and the strip
  renders a localised marker from it. `ScheduledToday` already goes `false` after the end
  and **is not duplicated**; `Ended` exists only because *"not today"* and *"never again"*
  are different facts and the strip would otherwise have to tell them apart in TypeScript.
  **The frontend compares no dates and reads no `COUNT`** — that is `wireDate` (**D17**)
  and `make guard` would not see it.
- **A finished habit stays in the strip**, marked finished, still showing the streak it
  ended on. It is removed by archiving, which already exists. A habit that silently
  vanishes is indistinguishable from data loss — the same argument **D15** made for the
  empty-project marker.
- **No new error sentinel.** A check written on a date the rule does not schedule is
  already specified as harmless — `streak.go` says so in its own doc comment: it is stored,
  contributes nothing and repairs nothing. An ended series is just *"every later date is
  unscheduled"*, so `Check`/`CheckToday` keep their current behaviour and **D25**'s
  refusal-code table does not grow. Offering or withholding the tick on a finished habit is
  presentation, which `HabitView`'s doc comment already assigns to the strip.
- **No RRULE library.** Unchanged, and the reasoning in `recurrence.go`'s own header
  comment still holds — it is now load-bearing for two decisions rather than one.

**Acceptance criteria**
- [ ] `COUNT` truncation, asserted through all four entry points — `Occurrences`,
      `NextOccurrence`, `PreviousOccurrence` and `Matches` — including the boundary pair
      (the `n`-th occurrence matches, the `n+1`-th does not) and `COUNT=1`.
- [ ] `UNTIL` inclusivity, asserted on the bound day itself and the day after, on all
      three frequencies, and with `INTERVAL` set so the bound falls between two
      occurrences.
- [ ] **Negative control**, recorded in the commit body: break the bound (make `UNTIL`
      exclusive, then let `COUNT` off by one), confirm the **named** test fails each time,
      restore. Two demonstrations.
- [ ] **A property sweep with an independent reference implementation**, in the shape of
      [`internal/domain/sweep_test.go`](./internal/domain/sweep_test.go) — random bounded
      rules × random windows, expanded by a reference that **shares no code** with
      `Recurrence.Expand` and derives the bound its own way (the reference counts, the
      implementation walks). That habit caught what four rounds of example-based tests
      missed in Stage 1; it is required here, not suggested.
- [ ] **Purity holds**: `internal/domain` still has no `time.Now()`, no `database/sql`, no
      `os`, no `net` — the existing purity test passes unmodified.
- [ ] The streak of an ended series is asserted to be **stable across three different
      injected `now` values**, months after the end — frozen, not decayed and not zeroed.
      And a series whose **final** occurrence went unchecked is asserted to be `0`, because
      that occurrence did pass unchecked.
- [ ] `HabitView.Ended` is asserted true past the bound and false before it, and
      `git grep -nE 'COUNT|UNTIL' frontend/src` finds no bound logic — only the label.
- [ ] `TestParseRecurrenceRejectsWhatItCannotExpand` still rejects `BYSETPOS`, `BYMONTH`,
      `BYWEEKNO`, `BYYEARDAY`, `BYHOUR`, ordinal `BYDAY`, the `DATE-TIME` `UNTIL`, and
      `COUNT`+`UNTIL` together. The diff shows **two rows moved**, none removed.
- [ ] `make cover` green with `internal/domain` still at **100.0%** — the bar it has held
      since Stage 1 — and `internal/service` still ≥ 90%. `make check` green.
- [ ] **Stated honestly in the commit body:** `make guard` cannot see any of this. The
      new field is a derivation with a name none of its heuristics match (**D17**), so the
      check is the Reviewer reading the diff.

**Commit:** `feat(domain): accept COUNT and UNTIL in the recurrence language (S3-19)`

---

## S3-20 — feat(frontend): the detail slide-over — shell, keyboard, Markdown

**This ticket creates the panel region and mounts it** — see
[Composition](#composition--who-mounts-what-in-stage-3). A panel that exists and is not in
the shell is an unfinished ticket.

**Scope (may touch):** `frontend/src/components/DetailPanel.tsx` (new),
`frontend/src/App.tsx`, `frontend/src/store/**`, `frontend/src/lib/keyboard.ts`,
`frontend/package.json`, the locale files, tests beside them.

Requirements:

- A slide-over **beside the board**, not a modal over it, hydrated from `NodeDetail`
  (S3-10). `Enter` on a focused card opens it; `Esc` closes it and **returns focus to the
  card it came from**.
- **Markdown editor and preview**, via `react-markdown` + `remark-gfm`. **`rehype-raw` must
  not be added** — no raw HTML from the database, ever.
- **It is not an overlay.** `store/ui.ts` holds one overlay and that is what keeps quick-add
  and the palette mutually exclusive. The panel is separate state, so this ticket must
  **state and test** what `Ctrl+N` and `Ctrl+K` do while the panel is open.
- **It has its own scroll container** (**D22**): the document does not scroll, so anything
  taller than the shell must say where it scrolls.
- Focus is trapped while open where that is correct, and every control is keyboard-reachable.

**Acceptance criteria**
- [ ] **Mounted and reachable**: `render(<App />)`, keyboard only, opens the panel on a card
      and closes it with `Esc`, focus restored. The test imports `App` and **does not import
      `DetailPanel`** — that restriction is the assertion.
- [ ] The preview renders GFM and **does not render raw HTML** — a fixture containing a
      `<script>` and an `<img onerror=…>` is asserted inert.
- [ ] `Ctrl+K` with the panel open behaves as specified, asserted.
- [ ] The panel scrolls internally; the document does not (the S3-01 height-chain test still
      passes).
- [ ] The orphan check (`App.mount.test.tsx`) is green; `make guard`, `make front-test`,
      `make check` green.

**Commit:** `feat(frontend): add the detail slide-over with markdown (S3-20)`

---

## S3-21 — feat(frontend): the panel's field editors

**Scope (may touch):** `frontend/src/components/DetailPanel.tsx` and new child components
under `frontend/src/components/`, `frontend/src/store/**`, the locale files, tests beside
them.

Requirements:

- Editors for **due, priority, estimate, activity, tags and type**, each calling the S3-11 /
  S3-12 / S3-13 method and **re-reading**. Nothing is computed locally — not the overdue
  flag, not whether a type may hold children, not whether a due date is allowed.
- **Every offered set comes from Go** (**D26**, S3-18): priorities, types, activities,
  tag colours.
- **A refusal is shown as a refusal** (**D25**, S3-08) — in place, next to the field, and
  the field reverts to Go's value.
- The tag editor offers only published colours, so **rule 9 stays true**: no hex in
  `frontend/src`.

**Acceptance criteria**
- [ ] **The round-trip criterion, per field**: change it by keyboard, assert the **exact Go
      call and argument**, then assert the panel renders **Go's answer** — a field that
      renders what the user typed has not round-tripped.
- [ ] A refused edit (a due date on a `note`; a type switch that `ErrTypeHasNoChildren`
      refuses) shows the specific refusal and leaves the field at Go's value.
- [ ] `make guard` finds no recomputed rule; `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src`
      returns nothing.
- [ ] Every editor is keyboard-reachable and operable, mouse untouched.
- [ ] `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the detail panel field editors (S3-21)`

---

## S3-22 — feat(frontend): inline subtasks

**Scope (may touch):** `frontend/src/components/DetailPanel.tsx` and a new subtask list
component, `frontend/src/store/**`, the locale files, tests beside them.

Requirements:

- Add, rename, reorder and complete children from inside the panel, through the **existing**
  `CreateNode` / `MoveNode` / `MoveToColumn` calls.
- **The parent's status and progress are not computed here.** They arrive re-derived from Go
  after each write (**D2**, **D10**, **D11**). A locally incremented counter is a second
  spelling of the progress rule.
- Keyboard: add with `Enter`, navigate with arrows, reorder with a modifier chord that does
  not collide with S2-16's map — **S2-16's map stays normative**.

**Acceptance criteria**
- [ ] Adding a child re-reads, and the parent's bar changes to **Go's** new value — asserted
      against a fixture where a locally computed value would differ.
- [ ] Reorder calls `MoveNode`, not `MoveToColumn`.
- [ ] The S2-16 keyboard suite passes unmodified.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add inline subtasks to the detail panel (S3-22)`

---

## S3-23 — feat(frontend): the recurrence editor (D27, D28)

> **⚠️ BLOCKED on [S3-19](#s3-19--featdomain-count-and-until-enter-the-recurrence-language-d27-d28).**
> **OQ2 is answered and closed**: the user took the middle option, so the editor offers an
> **end condition** — *"repeat N times"* (`COUNT`) and *"repeat until date"* (`UNTIL`) —
> and `internal/domain` gains them in S3-19. **Do not start this ticket until S3-19 is
> committed, and do not touch `internal/domain/recurrence.go` here**: a domain change
> smuggled into a UI ticket is exactly what the split exists to prevent.

**Scope (may touch):** a new recurrence editor component under `frontend/src/components/`,
`frontend/src/store/**`, the locale files, tests beside them. **`internal/domain` is NOT in
scope — S3-19 owns it.**

Requirements:

- **The editor does not parse, validate or expand an RRULE in TypeScript.** It collects
  structured choices; **Go** builds and validates the rule; the frontend renders the
  structure Go returns **through i18n**.
- **No human-readable sentence is returned from Go** — that is **D15**'s rejected option (c)
  and it cannot be translated.
- **The editor cannot compose a rule Go would reject**, which makes
  `ErrUnsupportedRecurrence` unreachable from the UI. That is the criterion, and it is a
  better one than a free-text box with an error under it. It still holds with the language
  widened — the editor offers `COUNT` **or** `UNTIL` **or** neither, never both (**D27**),
  the `COUNT` field is bounded to `1..1000`, and the date field can only emit the
  `YYYYMMDD` form (**D28**).
- **The end condition is offered, not typed.** Three mutually exclusive choices — *forever*,
  *N times*, *until a date* — because `COUNT`+`UNTIL` together is rejected by Go and an
  editor that can express a refusal is an editor with a bug in it.
- **Nothing here knows whether a series has ended.** That is Go's `HabitView.Ended` field
  (S3-19); the editor does not compare the bound to today, and neither does anything else
  in `frontend/src`.
- A habit requires a recurrence (`ErrNoRecurrence`); that rule is Go's and is not restated.

**Acceptance criteria**
- [ ] Every combination the editor can produce is accepted by Go — asserted by generating
      the editor's whole reachable space, **including both end conditions and neither**, and
      round-tripping each one.
- [ ] The editor cannot emit `COUNT` and `UNTIL` in the same rule — asserted over the same
      generated space, not by inspection.
- [ ] `git grep -niE 'FREQ=|BYDAY|RRULE|COUNT=|UNTIL=' frontend/src` finds **no parsing and
      no construction** of a rule string outside the one call that hands the structure to Go.
- [ ] The described rule renders correctly in **both** languages, from i18n keys and not
      from a Go string — including the two end conditions, with Russian's plural forms for
      *"N times"*.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the recurrence editor (S3-23)`

---

## S3-24 — feat(frontend): attachments, the editable time log, and the running clock

**Scope (may touch):** new components under `frontend/src/components/`,
`frontend/src/components/DetailPanel.tsx`, `frontend/src/lib/format.ts`,
`frontend/src/store/**`, the locale files, tests beside them.

Requirements:

- Attachments: add, list, open, remove — over S3-14. The file picker is the Wails runtime's;
  **no new dependency and no network.**
- The editable time log: list, edit, insert, delete — over S3-15. The single-active
  invariant is Go's and is not enforced a second time here.
- **The running clock.** `ae6befd` deleted `displayElapsedSeconds`, `timerStartedAt`,
  `formatTime` and `formatDuration` because they were **wall-clock arithmetic on top of Go's
  `elapsedSeconds`, rendered by nothing** — the exact shape of a second implementation
  waiting for its first caller. **This ticket adds back exactly what it renders and not one
  helper more.** It renders Go's `elapsedSeconds`; if a ticking display needs a local
  interval, that interval **re-reads from Go** rather than incrementing a local number.
- Durations are `font-mono`, formatted in `lib/format.ts` — **one** duration format in the
  application.

**Acceptance criteria**
- [ ] The clock renders Go's `elapsedSeconds`. A test supplying an `elapsedSeconds`
      inconsistent with the start time renders **Go's number** — proving TypeScript is not
      counting.
- [ ] `git grep -n 'Date.now\|new Date' frontend/src` shows no elapsed-time arithmetic.
- [ ] Editing an entry calls the S3-15 method and re-reads; an overlap is refused and shown
      as a refusal.
- [ ] Every helper added to `lib/format.ts` has a caller in this commit.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add attachments, the time log and the running clock (S3-24)`

---

## S3-25 — feat(frontend): the tree view

**This ticket introduces the view switch** — see
[Composition](#composition--who-mounts-what-in-stage-3) — and mounts the first non-Kanban
view.

**Scope (may touch):** `frontend/src/views/Tree.tsx` (new), `frontend/src/App.tsx`,
`frontend/src/store/ui.ts`, `frontend/src/lib/commands.ts`, `frontend/src/store/**`, the
locale files, tests beside them.

Requirements:

- `activeView` in `store/ui.ts`; region 3 of the shell switches on it; the command palette's
  **`view:tree` row flips from *unavailable, coming in stage N* to available**. **No palette
  row may silently do nothing** — the Stage 2 standard.
- Collapsible nodes, **inline rename**, **drag-to-reparent** (dnd-kit, already a dependency;
  the S3-05 `DragOverlay` policy applies here too), and arrow/`Enter`/`Tab` keyboard
  navigation.
- **Reparenting calls `MoveNode` and re-reads.** The frontend does **not** walk the subtree,
  and does not decide whether a parent is legal — `ErrTypeHasNoChildren` and `ErrCircularParent`
  are Go's and surface as refusals (**D25**).
- Expansion state is UI state and lives in `store/ui.ts`, per `ARCHITECTURE.md` §6.

**Acceptance criteria**
- [ ] **The ACCEPT half**: reparent a node in the tree, switch to Kanban, and the board shows
      it **without a manual refresh** — asserted through `render(<App />)`, keyboard only.
- [ ] A refused reparent (into its own subtree; onto a habit) rolls back and shows the
      specific refusal.
- [ ] **Mounted and reachable**: the test imports `App` and **does not import `Tree`**.
- [ ] The whole tree is navigable and operable by keyboard alone, mouse untouched.
- [ ] The orphan check is green; `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the tree view with drag-to-reparent (S3-25)`

---

## S3-26 — feat(frontend): the search screen

**Scope (may touch):** `frontend/src/views/Search.tsx` (new), `frontend/src/App.tsx`,
`frontend/src/store/**`, `frontend/src/lib/commands.ts`, the locale files, tests beside them.

Requirements:

- Query plus the four filters from S3-16 — tag, type, status, date range — each offered from
  **Go's published set** (**D26**).
- Results are Go's, in Go's order. **No client-side filtering, sorting or ranking**: that
  would be a second search.
- A result opens the detail panel; the view is fully keyboard-driven.
- Registers its palette row.

**Acceptance criteria**
- [ ] Each filter and the query produce exactly one `Search` call with the expected
      arguments; the rendered list is the response, unreordered.
- [ ] An empty result renders a localised empty state, not a blank region.
- [ ] **Mounted and reachable**: the test imports `App` and not `Search`.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the search screen (S3-26)`

---

## S3-27 — feat(frontend): the archive view

**Scope (may touch):** `frontend/src/views/Archive.tsx` (new), `frontend/src/App.tsx`,
`frontend/src/store/**`, `frontend/src/lib/commands.ts`, the locale files, tests beside them.

Requirements:

- Lists S3-17's archived nodes with enough context to identify one, and restores from there.
- **Restore re-reads the board**; the node reappears in its **derived** column, which the
  frontend does not compute.
- Registers its palette row.

**Acceptance criteria**
- [ ] Restoring from the archive view puts the node back on the board in the column **Go**
      reports, asserted through `render(<App />)`.
- [ ] An empty archive renders a localised empty state.
- [ ] **Mounted and reachable**: the test imports `App` and not `Archive`.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `feat(frontend): add the archive view (S3-27)`

---

## S3-28 — fix(frontend): shortcuts match the key, not the character (K13)

**This is K13, and it is ranked last on purpose.** `lib/keyboard.ts:107-115` matches on
`event.key`, and `event.code` appears **nowhere** in `frontend/src`. Under a Cyrillic
keyboard layout the N key emits `т` and K emits `л`, so **Ctrl+N and Ctrl+K would silently
stop working**. **The user has confirmed they keep a Latin layout**, so this is *not* the bug
they reported — it is a real latent defect in a Russian-language application, recorded and
scheduled rather than fixed under pressure.

**Scope (may touch):** `frontend/src/lib/keyboard.ts` and its test, tests beside them.

Requirements:

- Chords that name a **letter** match the physical key (`event.code`); chords that name a
  **named key** (`ArrowRight`, `Escape`, `Space`) keep matching `event.key`, because those
  are layout-independent already.
- **One matcher, not two.** The rule lives in `matches()`, which is already the single
  spelling — do not add a parallel path.
- Modifier matching stays **exact in both directions** (the existing comment explains why)
  and `Meta` stays required-up.

**Acceptance criteria**
- [ ] A synthetic event with `key: 'т'`, `code: 'KeyN'`, `ctrlKey: true` matches the
      quick-add chord; one with `key: 'n'`, `code: 'KeyN'` still does.
- [ ] Arrow, `Escape` and `Space` bindings are **unchanged** — the whole S2-16 suite passes
      unmodified.
- [ ] Negative control: revert the change, watch the new Cyrillic-layout test fail, restore.
- [ ] `make guard`, `make front-test`, `make check` green.

**Commit:** `fix(frontend): match shortcut chords on the physical key (S3-28)`

---

## S3-29 — docs: `README.md` and the screenshots — BLOCKED on the user

**`README.md` has been deliberately unwritten since Stage 0**, because every screenshot it
needs required a screen. The board exists now, so it is finally writable.

> **⚠️ BLOCKER the user must clear (sudo required), exactly as the Stage 0 blocker was:**
>
> ```sh
> sudo apt install xvfb imagemagick
> ```
>
> Without them there is **no way to capture a screenshot on this machine** — `xvfb-run`,
> `scrot`, `import` and `grim` are all absent. **Until it is cleared, this ticket produces
> the prose and leaves the image slots empty and marked as owed** — it does **not** ship a
> README with invented screenshots or borrowed ones, and it does **not** claim a screen
> nobody photographed.

**Scope (may touch):** `README.md` (new), and a captured-images directory if the blocker is
cleared. **No source file.**

Requirements:

- What Nexus is, **local-only and what that rules out**, the stack, how to build
  (**`-tags webkit2_41`**, **E1**, and why), how to run the gates, and where the data lives.
- Screenshots of the Kanban board and the habits strip, in **both palettes** and in **both
  languages** — the four-way grid is the point, because it is the claim that has been hardest
  to keep true.
- **No claim that is not verified.** If a screenshot is missing, the README says it is
  missing.

**Acceptance criteria**
- [ ] Every command in the README is run as written and works, from a clean checkout.
- [ ] The local-only rule is stated as a rule, not as a feature.
- [ ] If the blocker is uncleared: the image slots are explicitly marked as owed, and the
      commit body says so. **Nothing is faked.**
- [ ] `make check` green (the README changes no code, but the tree must still be clean).

**Commit:** `docs: add the readme (S3-29)`

---

## Stage 3 — DONE criteria

Stage 3 closes only once **all** of these hold, **verified by the Reviewer and not asserted
by the PM**. Per `PLAN.md` §5 a stage cannot close without a **PASS**; criterion 21 is that
PASS.

**Block A has its own gate, earlier**: criteria 1–7 must hold **before S3-10 starts**, and
S3-09 is where that is declared.

1. [ ] **Block A is closed before block B began.** S3-01 … S3-09 all committed, in order,
       and `make check` / `make cover` / `make front-test` / `make guard` all green at
       S3-09, before the first block B commit. Checkable from the git history.
2. [ ] **Every one of K6 – K12 is closed**, each by its named ticket, each with the
       mechanical check that ticket promised — **and each with an explicit written statement
       of what its mechanical check does NOT prove** (rule 17).
3. [ ] **`make guard` is eight of eight**, with **`GUARD_ALLOW_RE` still empty**. Check 7
       (one focus module) and check 8 (the blur allow-list) are both **exact greps**, both
       demonstrated by negative control.
4. [ ] **S3-09's hand pass is reported item by item, all ten items, pass or fail**, by a
       human on real hardware. A silent item fails this criterion. **Any failure produced a
       named follow-up ticket or a recorded user decision to proceed.**
5. [ ] **The three D8 due-badge assertions have their two mechanical halves** (S3-09 part 1),
       and the fake client was **not** taught D8. What remains hand-only is stated as
       hand-only.
6. [ ] **Russian does not clip at a real 1024×768, in Russian and in English**, confirmed by
       eye — the check that came back **failed** from the user's pass.
7. [ ] **The resize defect does not reproduce**, at more than one display scaling and in both
       palettes — or **D19**'s escalation ladder was followed and the outcome recorded,
       including escalation to the user if Aurora's blur must go.
8. [ ] **The ACCEPT criterion is MET**: *every field round-trips through Go* — demonstrated
       per field by asserting the exact Go call and then asserting the panel renders **Go's
       answer**, not the typed value — and *reparent in the tree shows on Kanban instantly*,
       demonstrated through `render(<App />)`, keyboard only.
9. [ ] **All twenty-nine tickets are committed**, one conventional commit each, in order,
       authored solely by `Ismat <mukhamejanov.ismat@gmail.com>`, with **no AI author, no
       co-author trailer and no "Generated with" line** (**D7**).
10. [ ] `make check` green — **all five gates, unchanged in number and definition**,
        including `wails build -tags webkit2_41`.
11. [ ] `make cover` green: `internal/domain` **and** `internal/service` both ≥90%, measured
        per package after `go clean -testcache`.
12. [ ] `make front-test` green, and the file/test count is recorded.
13. [ ] **No rule has a second spelling.** The Stage 1 inventory is still one definition each;
        the **new** Stage 3 candidates each have exactly one home — `preventScroll`, the blur
        allow-list, the layout floor, the refusal-code table, the Cyrillic `unicode-range`,
        the recurrence construction. **`make guard` passing is not the evidence; the Reviewer
        reading the diff is** (**D17**, rule 16).
14. [ ] `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src` returns nothing, and
        `git diff -- design/` is empty. **`design/` was not edited to add a font** — that is
        the constraint **D23** was shaped around.
15. [ ] **Local-only holds.** `git grep -nE 'https?://' frontend/src` returns nothing; the new
        font and Markdown packages are bundled from `node_modules`; no CDN, no fetch, no
        telemetry. `CGO_ENABLED=0 go build ./...` succeeds and
        `go list -deps ./... | grep -i mattn` prints nothing.
16. [ ] **No raw HTML is rendered from the database.** `rehype-raw` is absent, and the inert
        `<script>` fixture test is green.
17. [ ] **C6 is closed (D26)**: every enum set is published by Go, no locale table is used as
        a set, and **adding a value to a Go set turns something red** — six negative controls.
18. [ ] **The recurrence language is widened exactly as far as D27 says and no further.**
        `COUNT` and `UNTIL=YYYYMMDD` are accepted; **`BYSETPOS`, `BYMONTH`, `BYWEEKNO`,
        `BYYEARDAY`, `BYHOUR`, ordinal weekdays, the `DATE-TIME` `UNTIL` and
        `COUNT`+`UNTIL` together are still rejected at parse time**, with the rejection
        assertions **moved, not deleted**. `internal/domain` is still pure and still at
        **100.0%**. The **D28** consequence holds: an ended series **freezes** the streak
        rather than breaking or zeroing it, and that falls out of the bound rather than
        out of a second code path in `streak.go`. Negative control and an independent
        reference sweep both recorded.
19. [ ] **Every component Stage 3 built is reachable in the running app.** The orphan walk
        reports an empty difference and every mounting ticket's `render(<App />)`
        reachability test is green. **Neither half is sufficient alone** — Stage 2
        demonstrated the walk's blind spot twice.
20. [ ] **Nothing is claimed that was not verified.** Specifically: the Aurora drift is still
        **gated and not drawn** (**D16**, **K5**), **K14** is still **deferred by the user**
        and was not quietly implemented, and every jsdom-invisible claim in this stage names
        the eye that checked it.
21. [ ] **The Reviewer returns PASS.** Stage 1 closed on its fourth review and Stage 2 on its
        second; a first-round FAIL here is a normal outcome, not a failure of the process.

**Out of Stage 3 scope, and it must stay out**: the standalone frameless quick-add window
and the NL parser, focus mode, the tray, autostart, D-Bus sleep/lock, backup, export, the
PMP timelog screen, the calendar, stats, Gantt — **Aurora's background drift** (**D16**,
**K5**; not to be invented) — and **making the palette selection more visibly indicated**
(**K14**), which the user raised and then explicitly deferred.
