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
  `done` only when **all** children are done. **Children whose type has no Kanban
  column (`note`, `habit`) are excluded entirely** — from the "least advanced" scan
  and from the "all done" test. A node whose children *all* lack a column is therefore
  a **leaf** and reports its own stored status. See §7 **D10**.
- Progress % = **done work leaves / total work leaves**. A type with no Kanban column
  is never work and never counts (§7 **D7**, **D10**); a `project` is not a unit of
  work either — **except a *leaf* project, which counts as one work leaf in its
  parent's denominator**, done iff its own stored status is `done` (§7 **D9**, **D11**).

Because this is derived, it can never drift from the children. Dragging a parent is
therefore a **cascade**, not a write to the parent — see §7 D2.

### Column ↔ due coupling

Moving a card between columns writes a due date, and this is the rule I expect to
generate the most edge-case tests:

- → **Today**: `due = today`.
- → **This week**: `due = the upcoming Friday`; **if today is Friday, due = today**.
  (Today, 2026-09-18, *is* a Friday — the edge case is live on day one.)
- Both of those moves **always overwrite an existing due date, including a manual
  one**, and set `due_source = 'auto'`. The column move wins — see §7 **D8**.
- → **Backlog**: clears `due` **only if `due_source = 'auto'`**, i.e. only if it was
  auto-set by one of the moves above. A date the user typed by hand survives — see §7 D1.
- **Overdue** = `due < today AND status ≠ done`.
- Types with **no Kanban column** (`note`, `habit`) have no legal column move and
  therefore never get a due date this way — §7 **D9**.

### Tree, timer, habits

- **Dragging moves the whole subtree** — `parent_id` + `sort_order` change; children
  keep their own status.
- **A type with no Kanban column cannot be a parent.** A `note` and a `habit` take no
  children at all: a move onto one is refused (`ErrTypeHasNoChildren`, predicate
  `NodeType.HasColumn`). They may still *be* children — a habit grouped under a project
  is legal, it just contributes nothing to that project's column or bar (**D10**).
- **Exactly one active timer, globally.** Moving a card to Doing opens a
  `time_entry` and closes any open entry on any other node. (Stage 1 built both halves
  — the move and the timer — but **did not wire them together**; that wiring is an
  explicit **Stage 2** acceptance criterion, see §5.) On lock/sleep (D-Bus
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
| `project` | progress bar, **no timer**; **never enters `doing`, even when empty** (D9) |
| `habit` | habit strip only, never in columns; excluded from parent derivation and from progress (D10) |
| `note` | no status, no due; excluded from parent derivation and from progress (D7, D10) |
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
| 1 | **Domain + store**, Go only, no UI — repos, tree ops (create/move subtree/reorder/archive/restore), derived status + progress, column↔due rules, timer with single-active invariant, habit streaks, FTS5 spike then search. Table-driven tests incl. **parent→Done cascades to every unfinished descendant**, circular parent, overlapping timers, `due_source` transitions — **IMPLEMENTED, NOT CLOSED** (failed review three times, all fixes landed, awaiting re-check; see below) | **≥90% coverage** on `internal/domain` + `internal/service` — **met: 100.0% / 92.9%** |
| 2 | **Kanban + Habits strip** (launch screen) — Wails bindings, Zustand hydrated from Go, 5 columns, dnd-kit drag of card+subtree, optimistic UI with rollback on error, full card chrome, habit strip w/ streaks, quick-add (Ctrl+N), command palette (Ctrl+K), theme/palette/accent in settings, EN/RU | **Create → move through every column → complete, keyboard only, no mouse** — **and moving a card to Doing must itself open a `time_entry`** (§4 coupling; see `TASKS.md`, "Carried into Stage 2") |
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

### Stage 1 — IMPLEMENTED, NOT CLOSED

All twenty-two tickets **S1-01 … S1-22** are implemented and committed, one
conventional commit each. The ACCEPT criterion is **met**: `internal/domain`
**100.0%**, `internal/service` **92.9%** statement coverage, measured per package.
(`internal/store` is at **86.4%** and is not gated.)

**The stage nevertheless failed review three times**, every time on the same family of
defects: a type rule spelled in more than one place, with one copy diverging. Each
divergence was a door illegal rows could walk through.

First review — **FAIL**, three blocking issues:

1. illegal type/status combinations were accepted on the create path — **fixed by the
   Dev in `fd5e31d`**;
2. due dates were written onto types that have no Kanban column — **fixed by the Dev in
   `f1802d7`**;
3. **decisions D8 and D9 existed nowhere in the specification** although five source
   files cited them as authoritative — **fixed in §7 above**.

Second review — **FAIL**, on the remaining doors of the same rule. All fixes have
landed:

- `1cbe582` — the no-column rule routed through `domain.NodeType.HasColumn()`
  everywhere, so it has exactly one spelling;
- `f266bf5` — a due date refused on a `note`, and D9's wording corrected (a `habit`
  may have a date; it is denied a *column*, not a *date*);
- `81fb6e4` — **every** no-column type excluded from status derivation, not just
  `note` — the generalisation now recorded as **D10**, which also closed a second,
  pre-existing bug: a task whose only children were habits derived `done` while
  stored at `backlog`;
- `d5a170b` — an empty project counted as one unfinished work leaf in its parent's
  denominator — recorded as **D11**.

The two user decisions those last two commits required are recorded in §7 as **D10**
and **D11**, and §4's derivation rule has been generalised to match them.

Third review — **FAIL**, one blocking issue: the last surviving divergence, the one
this plan had written off as harmless in **K4**. `ValidateMove` refused only a `note`
as a parent, so a habit could still take children. The Reviewer showed end-to-end that
the two claims K4 made were both false: a task parked under a habit made the **habit
itself render in a Kanban column** — §4 says habits never appear in columns — and made
its ancestor project read `done` behind a **100% bar** while the unfinished task was
visible on the board at the same time. A 200,000-forest sweep traced roughly **32,300**
column violations and **800** derivation/progress violations to that single door.

Fixed by the Dev in `9664506`:

- `ValidateMove` now refuses **every** parent whose type has no Kanban column, not just
  a `note`, through the existing `NodeType.HasColumn` predicate;
- `deriveStatus` gained a **self-cut**: a no-column node derives `backlog` even if
  corrupt data somehow gives it children, so the invariant no longer rests on the
  validator alone;
- the sentinel `ErrNoteParent` was **renamed to `ErrTypeHasNoChildren`** ("domain: this
  node type cannot have children") **with no alias left behind** — an alias would have
  been a second spelling of the rule, which is the defect class itself.

**The structural fix — Stage 2 must not undo it.** The same commit folded the four
separate spellings of "a project never enters `doing`" (`CanEnterDoing`,
`CheckStatus`'s sentinel selection, `PlanCascade`, `canBeTimed`) onto a single
predicate, **`domain.DoingRefusal(NodeType) error`**, with `NodeType.CanBeDoing()`
defined in terms of it; the richer, caller-specific error messages were kept. The
resulting type-rule inventory has **no rule spelled twice**: `HasColumn`, `HasDue`,
`DoingRefusal`/`CanBeDoing`, `countsAsWork`, the habit-requires-recurrence check and
`DefaultActivity` are each the single definition of their rule. Three review failures
came from a rule written down twice and then edited once; a new duplicate spelling is
the one thing this stage must not gain.

One honest caveat, recorded so nobody "tidies" it later: `countsAsWork`
(`HasColumn() && != NodeTypeProject`) and `DoingRefusal` admit **the same set of types
today**, but they answer **different questions** — "is this a unit of work?" (**D7**,
**D11**) versus "may this node be doing?" (**D9**). They were kept separate
deliberately, so that a future change to one does not silently move the other.

**Stage 1 is therefore NOT closed.** Per §5, no stage closes without a PASS, and the
Reviewer has not yet re-checked the fixes. The stage closes when — and only when —
that re-check returns **PASS**. Until then nothing in Stage 2 starts.

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
Referenced as **D1–D11** and **E1–E3** from tickets in `TASKS.md`.

**D1–D7** were settled before Stage 0. **D8**, **D9**, **D10** and **D11** were
confirmed by the user *during* Stage 1, when implementation exposed questions the
earlier decisions did not answer; they are recorded here in the same form and carry the
same authority. D10 and D11 came out of the second review, and **D10 generalises the
derivation rule in §4**, which has been amended accordingly.

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
  not `done` and whose type has a Kanban column**. Descendants with no column
  (`note`, `habit`) are skipped — see **D10**.
- The parent **renders in its derived column**: the least-advanced status among its
  non-done children **whose type has a Kanban column**; `done` only when **all** such
  children are done. Children with no column are excluded entirely — from the
  "least advanced" scan and from the "all done" test.
- **The old "reject move to Done" test is REPLACED.** The new test is:
  *drag parent to Done → all unfinished descendants become `done`, `completed_at` is
  set on each, and the parent derives `done`.*
- **Parents never enter `doing` on their own.** Only **leaves** start a timer.
- A node **with no children, or whose children all have no Kanban column, behaves as a
  leaf** — it has its own stored status and can be dragged and timed like a task.
  All-`note` children and all-`habit` children count alike (`IsLeaf`/`HasColumn`).
  **Except a `project`**: type beats leaf-ness, so an empty project is still never
  dragged to Doing and never timed — see **D9**, which settles this contradiction.

**Amended by D10.** As originally recorded, the three bullets above said `note` where
the rule is **any type with no Kanban column** (`note` **and** `habit`). They are
restated above in the generalised form; the single spelling is
`domain.NodeType.HasColumn()`. Read any surviving "non-note" in this decision the same
way — implementing this decision as `note`-only reproduces the defect D10 fixed.

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
  status derivation. (**Generalised by D10**: *every* type with no Kanban column is
  excluded, from both the denominator and the derivation. The habit was already
  excluded from the denominator in practice; D10 is what made derivation agree.)
- **Projects can never enter `doing`.** No timer on a project. (See also D2: parents
  in general never enter `doing`. And **D9**, which settles the case D7 and D2 between
  them left open: the **empty** project, which is a project *and* a leaf.)
- **git**: repo is initialised, branch `main`, no remote. Commits use the
  **configured git user** — `Ismat <mukhamejanov.ismat@gmail.com>`.
  **No AI author and no AI co-author trailer on any commit.**
- **`go.mod` says `go 1.26`.** The brief's 1.23 is superseded; 1.23 is not installed.

### D8 — A column move ALWAYS overwrites the due date (confirmed during Stage 1)
A move to **Today** or **This week** overwrites **any** existing due date — including
one the user typed by hand — and sets `due_source = 'auto'`.

- **The column move always wins.** The column and the date can therefore never
  disagree: a card sitting in Today always shows today's date.
- The move to **Backlog** is unchanged and still reads `due_source` (**D1**): it clears
  `due` only when the source is `auto`.
- **Rejected alternative:** "a manual due date survives the move". It was rejected
  because it permits a card to sit in the Today column showing a date that is not
  today — the one state the column↔due coupling exists to make impossible.
- **Accepted consequence, stated by the user:** a hand-typed date is *destroyed* by the
  drag. Worse, a later move to Backlog then clears the date **entirely**, because by
  then its source is `auto` and no longer `manual`. The user accepted this.

This is a reading of **D1**, not a replacement for it. D1 defines the provenance
column and the Backlog rule; D8 settles what the Today / This-week moves do to a date
that is already there.

**Implemented consequences (Stage 1 — do not re-derive these):**
- `domain` applies the overwrite unconditionally on → Today and → This week, and flips
  `due_source` to `auto` in the same operation.
- Types with **no Kanban column** (`note`, `habit` — see D9) **never receive a due date
  from a column move**, because for them no column move is legal in the first place.

### D9 — Type beats leaf-ness; a project is NEVER timeable (confirmed during Stage 1)
A `project` node **never enters `doing` and never starts a timer — even when it has no
children at all.**

This resolves a real contradiction found while implementing Stage 1: **D7** says
projects can never enter `doing`, while **D2** says a node with no children, or with
only `note` children, "behaves as a leaf" and can be dragged and timed (D2's leaf rule
is now generalised by **D10** to children with **no Kanban column**, which widens the
contradiction rather than changing it). An *empty project* satisfies both descriptions
at once.

- **Type wins.** The per-type rule ("`project` → progress bar, **no timer**", §4) is
  more specific than the general leaf rule, so it takes precedence. Leaf-ness decides
  what a *task* does; type decides whether the node is a unit of work at all.
- **Rejected alternative:** "leaf-ness wins" — an empty project would be timeable, and
  the timer would then **silently disappear** the moment the user added the first
  subtask. A control that vanishes on an unrelated action is worse than one that was
  never offered.
- **Accepted consequence, stated by the user:** a freshly created, still-empty project
  is **inert** — it cannot be dragged to Doing and cannot be timed. *Adding a child is
  how work becomes timeable.*

**Implemented consequences (Stage 1 — do not re-derive these):**
- The rule has **one spelling**: `domain.DoingRefusal(NodeType) error`, with
  `NodeType.CanBeDoing()` defined in terms of it. `CanEnterDoing`, `CheckStatus`'s
  choice of sentinel, `PlanCascade` and `canBeTimed` all ask it rather than restating
  it (see §5, Stage 1 history, for why).
- `MoveToColumn(project, doing)` is refused with `ErrProjectNeverDoing`, and so is
  creating a project directly in `doing`.
- `PlanCascade` **skips project descendants** when cascading `doing`, and refuses a
  project as the drag target outright.
- A project is **not counted as a unit of work in the progress denominator**.
  Consequently a project with nothing under it that counts as work — all notes, all
  habits, any mix of types with no Kanban column, or empty — reports
  `Defined() == false`: **neither 0% nor 100%**, but *no percentage at all*. A progress
  bar with nothing to measure must not claim it measured nothing.
  **Amended by D11** for one case only: a project's *own* progress is still undefined
  when nothing under it is work, but a **leaf** project now counts as **one work leaf
  in its parent's denominator**. The two statements answer different questions and are
  both true — see D11.
- Types **without a Kanban column** (`note`, `habit`) are refused a **column status** by
  every door: the create path refuses anything but the inert `backlog`, `MoveToColumn`
  refuses the drag outright, and `PlanCascade` skips them as descendants, so a drag on
  their parent cannot give them one either. The rule lives in one place,
  `NodeType.HasColumn`.
- **A `note` never receives a due date**, by any door — §4 says `note` = "no status, **no
  due**" — so `CreateNode` and `SetDue` refuse one (`ErrTypeHasNoDue`). **A `habit` may
  have one.** §4 denies a habit a *column*, not a *date*, and this bullet used to claim
  otherwise for both types; the code was never written that way and the restriction is
  not invented now. The predicate is `NodeType.HasDue`, deliberately separate from
  `HasColumn` because the two questions have different answers for a habit.

### D10 — A node with no Kanban column is excluded from parent derivation (confirmed during Stage 1)
**Nodes whose type has no Kanban column are excluded from parent status derivation,
exactly as notes already were.**

This **generalises §4's earlier wording**, which excluded only `note` children. The
principle is **D9**'s: a type that has no Kanban column cannot contribute to a
**column-derived** status. It also makes derivation consistent with the progress
denominator, which already excluded both notes and habits.

- **Motivating case:** `project{task:done, habit}` derived **backlog** while its
  progress bar read **100%**. The habit, parked for ever at its inert `backlog`, was
  scanned as real work and could never become done, so the parent could never derive
  `done` — a card sitting in the Backlog column with a full bar drawn on it. It now
  derives **done** with a 100% bar: the column and the bar agree.
- **Rejected alternatives:** (a) keep §4 literal and exclude only notes — this is the
  state that produced the defect, and it leaves the column and the bar visibly
  disagreeing on the same card; (b) forbid habits as children entirely — this removes
  the ability to group habits under a project, which is a legitimate way to organise
  them, in order to fix a derivation bug that has nothing to do with the tree shape.
- **Accepted consequence, stated by the user:** a subtree made only of no-column types
  contributes **nothing** to its parent — not to the column, not to the bar. The parent
  of such a subtree is a leaf and answers with its own stored status.

**Implemented consequences (Stage 1 — do not re-derive these):**
- `DeriveStatus` **skips any child where `!child.Type.HasColumn()`** — it is not
  scanned for "least advanced" and not descended into.
- `Node.IsLeaf` treats a node **whose children all lack a column** as a leaf, so it
  reports its own stored status.
- `walkProgress` **cuts at the same predicate**. This was forced, not optional: cutting
  derivation without cutting progress reintroduces the identical class of bug one level
  down, which is exactly how the original defect was written.
- Everything routes through **`domain.NodeType.HasColumn()`**. The rule has one
  spelling; a second spelling is a bug by construction.
- This also closed a **second, pre-existing bug**: a task whose only children were
  habits derived `done` while its stored status was `backlog`.

### D11 — An empty project is one unfinished work leaf in its parent (confirmed during Stage 1)
**An empty project counts as one unfinished work leaf in its parent's progress
denominator.** It is real work that has not been broken down yet.

The user accepted that this **reverses part of D9's "a project is never a unit of
work" for the empty case specifically**.

- **Motivating case:** `project{empty sub-project, task:done}` derived **backlog**
  while its bar read **100%** — the same defect shape as D10. Derivation was already
  right (the empty sub-project has a column and is not done, so it holds the parent
  back); the **denominator** was wrong, because the sub-project was skipped entirely.
  It now reads **1 of 2 = 50%**, and both halves say "not done". **Derivation needed no
  change.**

**The distinction that keeps D11 consistent with S1-07 — Stage 2 must not re-derive it:**

1. **A project's OWN progress**, when you ask about *that project itself*: if it
   contains no work beneath it, it stays **`Defined() == false`** — neither 0% nor
   100%, and no bar is drawn. **UNCHANGED.** S1-07's reasoning still binds: *"a project
   containing only notes has no work in it, and both 0% and 100% are lies the UI would
   render as a bar."*
2. **A project's contribution to its PARENT's denominator**: a project that is a
   **leaf** (no children with a column, per **D10** — so an empty one, an all-notes one
   and an all-habits one alike) counts as **one work leaf**, **done iff its own stored
   status is `done`**. **THIS is the change.**

So an empty project **shows no bar of its own while counting as one unfinished unit
inside its parent**. Both are true at once: the first question is "what is inside this
project?" (nothing), the second is "is this project finished?" (no).

- A project **with real children is still not counted as a unit itself**; its children
  are counted. Nesting adds levels, not units — a chain of empty projects is one leaf,
  at the bottom.
- **Rejected alternatives:** (a) exclude the empty project from derivation too, so it
  stops holding the parent back — the parent would then read fully done while
  containing an unplanned project, which hides unstarted work, the one thing a plan
  must not do; (b) leave the disagreement in place — a 100% bar on a card in Backlog is
  a defect whichever half you believe.
- **Accepted consequence, stated by the user:** part of **D9** is reversed for the
  empty case. "A project is never a unit of work" now holds only while the project has
  work under it.

**Implemented consequences (Stage 1 — do not re-derive these):**
- `ComputeProgress` threads a single "this is the node we were asked about" flag
  through `walkProgress`, applied at the one leaf site. `countsAsWork`, `HasColumn`
  and `IsLeaf` are untouched — **position in the walk is not a property of the type**,
  so it deliberately did not become a fourth spelling of the type rule.
- A leaf project is counted **done by its own stored status**, not by derivation —
  there is nothing under it to derive from. See known issue **K2**.

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

**K2 — a leaf project's stale stored status now has teeth.**
Since **D11**, a leaf project's *stored* status decides whether it counts as done in
its parent's denominator. Archiving the last real child of a project that an earlier
cascade had written `done` leaves it counted as a **done** unit on a status nobody set
deliberately. The state is self-consistent — column and bar agree — so this is not a
contradiction and **is not scheduled now**. It may want a rule about **re-inspecting a
project's stored status when its last child is archived**. That ruling belongs to
**Stage 2**.

**K3 — an empty project stored `done` renders in the Done column with no bar at all.**
Its own progress is undefined (**D11**, part 1), so no bar is drawn, while its status
puts the card in Done. Not a contradiction, but it is **the one place a finished card
shows nothing**. Recorded for the Stage 2 card-chrome work to decide what, if anything,
a done-but-unmeasurable card should render.

**K4 — RESOLVED in `9664506`, not an open issue. A no-column type cannot have
children at all.**
K4 used to read that `ValidateMove` refused only a `note` as a parent, so a habit could
still have children, and that the resulting subtree was "fully inert — nothing lies".
**Both claims were false.** The third review demonstrated it end-to-end: a task parked
under a habit made the **habit render in a Kanban column**, contradicting §4's "habits
never appear in Kanban columns", and made its ancestor project read `done` with a
**100% bar** while the unfinished task was still visible on the board. A
200,000-forest sweep attributed roughly **32,300** column violations and **800**
derivation/progress violations to this one door.

The fix: `ValidateMove` refuses **every** parent whose type has no Kanban column via
`NodeType.HasColumn`, and `deriveStatus` cuts at the node itself, so a no-column node
derives `backlog` even if corrupt data gives it children. The sentinel `ErrNoteParent`
became **`ErrTypeHasNoChildren`**, with no alias, because an alias is a second spelling
of the rule. Nothing is parked out of sight because nothing can be parked there.
Recorded here rather than deleted so the failure mode stays on the record; **K1–K3
remain the open ones.**

---

**Status: decisions locked — D1–D11, E1–E3. Stage 0 is CLOSED (PASS). Stage 1 is
implemented (all twenty-two tickets, S1-01 … S1-22) but is NOT closed: it failed review
three times, every fix has landed — including `9664506`, which resolved **K4** and
folded the "may be doing" rule onto one predicate — and it is awaiting a re-check.
See §5, "Stage 1 — IMPLEMENTED, NOT CLOSED".**
