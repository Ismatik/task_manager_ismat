# Nexus — Plan

My restatement of the brief, written before any code. It stays the source of truth for
the data model and the decisions; only the stage status below moves. **Stage 0 is
closed (PASS); Stage 1 is current** — see §5 and `TASKS.md`.

---

## 1. What we are building

A **local-only personal task manager** for Linux desktop. No server, no sync, no
account. All data lives in a single SQLite file on the user's machine.

Shipping as a **Wails v2** desktop app: a Go backend compiled into one binary with
a WebView frontend. Stack fixed by the brief:

| Layer | Choice | Note |
|---|---|---|
| Shell | Wails v2 | Go ↔ JS bindings, single binary |
| Backend | Go 1.26 | `go.mod` says `go 1.26` — see §7 D7 |
| DB | `modernc.org/sqlite` | pure Go, **no cgo** — keeps cross-compile simple |
| Frontend | React 18 + TypeScript | |
| Styling | Tailwind + tokens from `design/` | |
| Components | ReactBits | see §7 D6 — must not fight the design *tokens* |

The central architectural commitment:

> **All domain logic lives in Go. The frontend renders what Go returns.**

The frontend never computes a status, a streak, a progress percentage, an overdue
flag, or a due date. If a number appears on screen, Go computed it. This is not a
style preference — it is what makes the rules testable, and it is the thing I will
push back on hardest during review.

Corollaries:
- Domain logic sits in `internal/domain` and `internal/service`, unit-tested there.
- Every Wails-exposed method returns `(T, error)`. The frontend surfaces every
  error in a toast — no silent failures.

## 2. Non-negotiable cross-cutting rules

**i18n from day one.** No hard-coded UI strings anywhere. Every string is a key in
`frontend/src/locales/en.json` and `frontend/src/locales/ru.json`. Retrofitting i18n
is miserable, so this is enforced from the first component. Russian is longer than
English — the design brief requires ~30% extra width in buttons and labels.

**Accessibility.** Everything keyboard-reachable, focus always visible,
`prefers-reduced-motion` respected everywhere (including pausing Aurora's background
drift, not just CSS transitions).

**Scope discipline.** One stage at a time. I do not touch a later stage's files
early, and I do not "improve" scope. If the brief is contradictory or impossible,
I say so and stop rather than guess.

## 3. Design system — follow, do not invent

Source of truth is the handoff in `./design/` at the repo root (moved there — see §7 D6).
I read `README.md` first, as instructed.

- **Colors only via Tailwind token names**: `bg`, `surface`, `elevated`, `line`,
  `ink`, `muted`, `accent`, `accent-2`, `on-accent`, `danger`, `warning`, `success`.
  **Zero hex literals in components.** This is mechanically checkable, and I intend
  to have the Reviewer grep for it.
- **Two palettes × two themes**: `<html data-palette="aurora|studio">` plus a `.dark`
  class (`darkMode: 'class'`). Plus a user accent override written to `--accent`.
  All three persist in `settings`. Defaults: **aurora + dark**.
- **Surfaces differ per palette**: Aurora is translucent (`bg-surface` +
  `backdrop-blur-glass`); Studio is solid (`shadow-sm`). Radii only from
  `rounded-sm/md/lg`.
- **Fonts**: `--font-ui` switches per palette (Space Grotesk / Figtree).
  **JetBrains Mono for every number, date, timer and shortcut hint.** Base 13px,
  8px spacing grid.
- **Semantics**: `accent` = active view / running timer / focus ring / primary
  button. `danger` = overdue, P0–P1, bug. `warning` = P2, streak flames.
  `success` = done, habit checks, task icon.
- **Motion** ≤200ms via `duration-fast/base/slow`.

## 4. Data model

One table of `nodes` forming a **single tree** — task, project, habit, note and bug
are all rows in the same table, distinguished by `type`. This is the key modelling
decision: a subtask and a project differ by type and position, not by table.

```
nodes(id uuid, parent_id?, type ∈ {task,project,habit,note,bug}, title,
      description_md, status ∈ {backlog,week,today,doing,done}, due date?,
      due_source ∈ {manual,auto}, priority 1..4, estimate_min,
      recurrence RRULE?, activity?, sort_order,
      created_at, updated_at, completed_at?, archived_at?)

tags(id, name, color)            node_tags(node_id, tag_id)
time_entries(id, node_id, started_at, ended_at?)
attachments(id, node_id, path, mime)
habit_checks(node_id, date)      settings(key, value)
```

### Derived state — never stored

- A parent's status is the **least-advanced status among its non-done children**;
  `done` only when **all** children are done. `note` children are excluded.
- Progress % = **done leaves / total leaves**.

Because this is derived, it can never drift from the children. Dragging a parent is
therefore a **cascade**, not a write to the parent — see §7 D2.

### Column ↔ due coupling

Moving a card between columns writes a due date, and this is the rule I expect to
generate the most edge-case tests:

- → **Today**: `due = today`.
- → **This week**: `due = the upcoming Friday`; **if today is Friday, due = today**.
  (Today, 2026-09-18, *is* a Friday — the edge case is live on day one.)
- → **Backlog**: clears `due` **only if `due_source = 'auto'`**, i.e. only if it was
  auto-set by one of the moves above. A date the user typed by hand survives — see §7 D1.
- **Overdue** = `due < today AND status ≠ done`.

### Tree, timer, habits

- **Dragging moves the whole subtree** — `parent_id` + `sort_order` change; children
  keep their own status.
- **Exactly one active timer, globally.** Moving a card to Doing opens a
  `time_entry` and closes any open entry on any other node. On lock/sleep (D-Bus
  `org.freedesktop.login1` `PrepareForSleep`, plus screensaver `ActiveChanged`) the
  open entry is closed. On resume we ask "resume timer?" — we never silently
  re-open, and we never bill the user for time spent asleep.
- **Habits**: `recurrence` required, one `habit_check` per day, streak = consecutive
  **scheduled occurrences of the RRULE** that were checked — *not* calendar days, so a
  weekly habit checked four weeks running has a streak of 4. Today's still-pending
  occurrence does not break a live streak. See §7 D5. Habits **never** appear in
  Kanban columns.

### Behaviour by type

| type | behaviour |
|---|---|
| `task` | default |
| `project` | progress bar, **no timer** |
| `habit` | habit strip only, never in columns |
| `note` | no status, no due; excluded from parent derivation |
| `bug` | extra `repro_steps` + `severity`, stored as frontmatter in `description_md` |

## 5. Stages

Each stage ends with: Reviewer runs `go test ./...`, `go vet ./...`, `npm run lint`,
`npm run typecheck`, `wails build -tags webkit2_41` (see §7 E1), and checks
acceptance criteria → **PASS** or a
list of blocking issues. Dev fixes, Reviewer re-checks. **No stage closes without
PASS.** Then I print a summary (files touched, test count, how to verify by hand)
and stop for your "next".

| # | Stage | Acceptance |
|---|---|---|
| 0 | **Scaffold** — `wails init` react-ts, Tailwind, ESLint/Prettier, Go layout, embedded SQL migrations, `settings`, Makefile, `make check`, single-instance lock (`--quick` → quick-add on running instance; bare → focus main window) — **CLOSED, PASS** | `make check` green, empty window opens, second launch focuses the first |
| 1 | **Domain + store**, Go only, no UI — repos, tree ops (create/move subtree/reorder/archive/restore), derived status + progress, column↔due rules, timer with single-active invariant, habit streaks, FTS5 spike then search. Table-driven tests incl. **parent→Done cascades to every unfinished descendant**, circular parent, overlapping timers, `due_source` transitions | **≥90% coverage** on `internal/domain` + `internal/service` |
| 2 | **Kanban + Habits strip** (launch screen) — Wails bindings, Zustand hydrated from Go, 5 columns, dnd-kit drag of card+subtree, optimistic UI with rollback on error, full card chrome, habit strip w/ streaks, quick-add (Ctrl+N), command palette (Ctrl+K), theme/palette/accent in settings, EN/RU | **Create → move through every column → complete, keyboard only, no mouse** |
| 3 | **Detail + Tree + Search/Archive** — slide-over with Markdown editor/preview, inline subtasks, tags, due, priority, estimate, RRULE editor, attachments copied into app data dir, editable time log, type switcher; collapsible tree with inline rename, drag-to-reparent, arrow/Enter/Tab keyboard nav; archive view; FTS search with tag/type/status/date filters | Every field round-trips through Go; reparent in tree shows on Kanban instantly |
| 4 | **Quick-add + Focus mode** — frameless standalone window, Go-side NL parser (date, `!priority`, `#tag`, `>Project` fuzzy, `~estimate`, `@type`), live preview chips, Enter creates & closes, Esc closes; Focus mode (one card, large timer, Esc exits); sleep/lock timer handling | `deploy KA Avto fri 15:00 !high #work >KA Avto ~2h` parses correctly in tests **and** in the UI |
| 5 | **Platform integration** — tray via `energye/systray`, badge = overdue + due today, menu (Open / Quick add / Start-Stop timer / Quit); `.desktop` + `install.sh` → autostart, GNOME `gsettings` shortcut Super+Space → `nexus --quick`, warn if AppIndicator missing | Reboot → app opens; Super+Space → quick-add |
| 6 | **Backup, export, PMP timelog** — nightly JSON export of all tables to `~/Nexus/backups/YYYY-MM-DD.json`, keep 30; Restore-from-file with confirmation; Markdown export of a subtree; day timelog screen grouping `time_entries` by node with editable minutes, output in PMP KIT format with Copy | **Restore reproduces an identical Kanban** |
| 7 | **Calendar + Stats** — month/week by due, drag to reschedule; completed per day/week, time per project/tag, streaks, estimate vs actual | Numbers match raw SQL run directly |
| 8 | **Gantt** (projects only) — children as bars between created/due, drag to change due | Dragging a bar updates due everywhere |

### Stage 0 — CLOSED, PASS

The Reviewer returned **PASS**. All five gates green — including gate 5,
`wails build -tags webkit2_41`, now that the user has installed `pkg-config`,
`libgtk-3-dev` and `libwebkit2gtk-4.1-dev` — and all seven Stage 0 DONE criteria in
`TASKS.md` met, including the two that needed a real window: the app opens, and a
second launch exits immediately and focuses the first. Thirteen tickets, **S0-01 …
S0-13**, one conventional commit each, no AI author and no co-author trailer on any
of them.

What landed: the Wails v2 react-ts scaffold; Tailwind consuming `design/` without
duplicating a single token; ESLint/Prettier/`typecheck` with `--max-warnings=0`; the
five Go packages of **D3** with the core enums and a mechanical purity test; embedded
migrations and the SQLite connection on `modernc.org/sqlite` with WAL, foreign keys
and a busy timeout verified on open; the `settings` table with idempotent seeding; the
unix-socket single-instance lock with `--quick`; `make check`; `CLAUDE.md`; and
`design/pmp-timelog-format.md`.

What Stage 0 deliberately did **not** ship, and Stage 1 therefore owes: the `nodes`
schema and every rule in §4 — status derivation, progress, the column↔due coupling,
the single-active timer, streaks and search. Stage 0's only migration is
`0001_settings.sql`; Stage 1 appends `0002` onwards and never edits it.

**Stage 1 is broken into tickets S1-01 … S1-22 in `TASKS.md`.**

**Final review**: fresh clone → `make check` → `wails build -tags webkit2_41` → `install.sh` →
reboot checklist, executed and reported. `QA.md` with 25 manual scenarios covering
every rule in §4. Known gaps reported honestly — **nothing marked done that was not
actually verified.**

## 6. How we work

I am the **orchestrator**. Three sub-agents, delegated explicitly:

- **PM agent** — owns `PLAN.md` and `TASKS.md`. Breaks each stage into tickets with
  acceptance criteria. Decides when a stage is DONE. **Never writes code.**
  (Takes ownership of this file once you say "go".)
- **Dev agent** — implements **exactly one ticket at a time**, tests alongside, one
  conventional commit per ticket. **No AI listed as git author or co-author** —
  commits are authored as you.
- **Reviewer agent** — after each stage reads the diff, runs the five commands
  above, checks acceptance criteria, returns PASS or blocking issues.

---

## 7. Resolved decisions

The open questions are **closed**. Every answer below was given by the user and is
authoritative — it overrides anything earlier in this document that contradicts it.
Referenced as **D1–D7** and **E1–E3** from tickets in `TASKS.md`.

### D1 — `due_source` (was Q1: due-date provenance)
Add the column `due_source TEXT NOT NULL DEFAULT 'manual'`, values in
`{manual, auto}`.

- The column↔due rules (→ Today, → This week) set `due_source = 'auto'`.
- **Any user edit of `due` sets `due_source = 'manual'`.**
- Moving to **Backlog** clears `due` **only when `due_source = 'auto'`**.

This replaces the earlier `due_is_auto BOOLEAN` proposal.

### D2 — Parent status is derived; dragging a parent CASCADES (was Q2)
Parent status is **never stored**. Neither option (a) nor (b) was taken; the rule is:

- Dragging a parent to column **X** sets `status = X` on **every descendant that is
  not `done` and not a `note`**.
- The parent **renders in its derived column**: the least-advanced status among its
  non-done, non-note children; `done` only when **all** such children are done.
- **The old "reject move to Done" test is REPLACED.** The new test is:
  *drag parent to Done → all unfinished descendants become `done`, `completed_at` is
  set on each, and the parent derives `done`.*
- **Parents never enter `doing` on their own.** Only **leaves** start a timer.
- A node **with no children, or whose children are all notes, behaves as a leaf** —
  it has its own stored status and can be dragged and timed like a task.

### D3 — ARCHITECTURE.md (was Q3)
The layout below is decided and is written into `ARCHITECTURE.md` during **Stage 0**:

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

Note: `main.go` sits at the **repo root** (Wails convention), not under `cmd/nexus`.

### D4 — PMP timelog format (was Q5)
Create `design/pmp-timelog-format.md` in **Stage 0**, derived from `design/SKILL.md`
(the interactive skill stays as-is; the new file is the deterministic spec).

- Add node field **`activity TEXT`** with the fixed list from `SKILL.md`:
  `Разработка, Анализ, Тестирование, Документация, Совещание, Согласование,
  Управление проектом`.
- **Defaults by node type**, user-overridable at any time:

  | type | default activity |
  |---|---|
  | `task` | Разработка |
  | `bug` | Тестирование |
  | `project` | Управление проектом |
  | `note` | Документация |

- **Комментарий is GENERATED, then user-edited.** The generator produces it from the
  node; the user may rewrite it before copying.
- **Hours are rounded to the nearest 15 minutes.** Measured hours only — never
  invented, per `SKILL.md`.

### D5 — Streaks count scheduled occurrences, not days (was Q6)
A streak is the number of **consecutive scheduled occurrences of the node's RRULE**
that were checked.

- A weekly habit checked 4 weeks running has **streak = 4**.
- The streak **breaks when a scheduled occurrence passes unchecked**.
- **Today's still-pending occurrence does not break the streak.**

### D6 — design/ location and ReactBits (was Q4 + part of Q7)
- The design directory has **already been moved** to `./design/` at the repo root.
  No spaces in the path. All tooling references `./design/`.
- **ReactBits components are used as named in the design export.** The rule
  "do not invent styles" applies to **tokens, colours and typography** — not to
  component choice. A ReactBits component named in the export is in scope; its
  colours/fonts/radii must still resolve through the Tailwind token names.

### D7 — Smaller ambiguities, now settled (was Q7)
- **`note` leaves are excluded from the progress denominator**, consistent with
  status derivation.
- **Projects can never enter `doing`.** No timer on a project. (See also D2: parents
  in general never enter `doing`.)
- **git**: repo is initialised, branch `main`, no remote. Commits use the
  **configured git user** — `Ismat <mukhamejanov.ismat@gmail.com>`.
  **No AI author and no AI co-author trailer on any commit.**
- **`go.mod` says `go 1.26`.** The brief's 1.23 is superseded; 1.23 is not installed.

### FTS5 — spike it, do not guess (was Q8, unchanged)
**Spike FTS5 on `modernc.org/sqlite` in Stage 1**, first thing. If FTS5 is not
available in the pinned version, **fall back to LIKE-based search on `title` +
`description_md`**. **Never introduce cgo** — the no-cgo guarantee is not negotiable
and the fallback is pre-approved, so this is no longer a stop-and-ask fork.

---

### Environment constraints — verified on this machine

**E1 — `libwebkit2gtk-4.0-dev` DOES NOT EXIST on Ubuntu 24.04.** Only
`libwebkit2gtk-4.1-dev` is packaged. Therefore:

> **Every `wails build` and `wails dev` invocation must pass `-tags webkit2_41`.**

The Makefile bakes this in. A build command without the tag will fail to link, and
that failure is expected, not a bug to debug. This is a real, verified constraint.

**E2 — `wails` CLI v2.16.0 is installed** at `$(go env GOPATH)/bin/wails`
(`/home/ismat/go/bin/wails`). It is **not necessarily on `PATH` in non-login
shells**, so the Makefile must resolve it explicitly (e.g. via
`$(shell go env GOPATH)/bin/wails`) rather than assume `wails` is callable.

**E3 — Toolchain present**: Go 1.26.0, Node v22.22.2, npm 10.9.7.
`pkg-config`, `libgtk-3-dev` and `libwebkit2gtk-4.1-dev` were **missing** when Stage 0
was planned and have since been installed by the user, which is what unblocked gate 5
and closed Stage 0. `pkg-config --exists gtk+-3.0 webkit2gtk-4.1` now succeeds.

---

### Known issues — recorded, not scheduled

**K1 — `BackgroundColour` never reaches GTK under a comma-decimal locale.**
This machine runs `LC_NUMERIC=ru_RU.UTF-8`. Wails builds the window background as a
CSS string in C at `window.c:205` and formats the alpha with the **process locale**
rather than the C locale, so it emits

```
rgba(27, 38, 54, 0,0)
```

GTK's CSS parser rejects that declaration — a comma is an argument separator, not a
decimal point — and the background colour is therefore **silently never applied**.
Nothing is logged; the window just uses its default.

This is an **upstream Wails bug**, not ours, and it is **not fixed in Stage 1**:
Stage 1 is Go-only, has no window work in it, and a locale workaround bolted on now
would be untestable until there is a palette to compare against. **The decision
belongs to Stage 2**, where the palette/theme work makes the window background have
to match a token. The options to weigh there, none of them chosen here:

1. force `LC_NUMERIC=C` for the process before `wails.Run`;
2. leave the window background transparent and let the frontend paint it, which is
   where the palette lives anyway;
3. patch upstream and pin the fork.

Related but **separate**: `main.go` passed `&options.RGBA{R: 27, G: 38, B: 54, A: 1}`.
Alpha is 0–255, so `A: 1` is ~0.4% opacity — a real bug of ours, fixed in Stage 1 by
ticket **S1-02**. Fixing it does not fix K1, and K1 is why the fix cannot be verified
by eye on this machine.

---

**Status: decisions locked. Stage 0 is CLOSED (PASS). Stage 1 is broken into tickets
S1-01 … S1-22 in `TASKS.md`.**
