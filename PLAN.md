# Nexus — Plan

My restatement of the brief, written before any code. It stays the source of truth for
the data model and the decisions; only the stage status below moves. **Stage 0 and
Stage 1 are both CLOSED (PASS); Stage 2 is IMPLEMENTED, NOT CLOSED** — all twenty-two
tickets are committed, **review round 1 returned FAIL on two blocking issues that are now
fixed** (`ae6befd`, `4b1af9c`, and the ruling they were missing is **D17**), the Reviewer
has not returned PASS, and a stage closes only on a PASS. See §5 and `TASKS.md`.

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

**The design export names no components — stop looking for the list.** The brief speaks
of "component choices from the Claude Design export" and of "ReactBits components
exactly as named in the design export". There is no such list. `design/` contains
`README.md`, `tokens.css`, `tailwind.config.js`, `SKILL.md` and
`pmp-timelog-format.md`, and the README specifies **tokens, fonts, semantics and
motion only** — not one component name appears in any of them. This was checked
during Stage 2 planning and is recorded here so that nobody re-hunts for a file that
does not exist.

The consequence, which is **D6 as amended**: Stage 2 **hand-builds its components
against the tokens**. ReactBits is used only where it genuinely fits and can be fully
re-skinned to the token names — never as an excuse to bring in a colour, a radius or a
font the tokens do not define. **No component spec may be invented and attributed to
the design export.** A component that "the design export calls for" is a component
somebody made up.

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
| 1 | **Domain + store**, Go only, no UI — repos, tree ops (create/move subtree/reorder/archive/restore), derived status + progress, column↔due rules, timer with single-active invariant, habit streaks, FTS5 spike then search. Table-driven tests incl. **parent→Done cascades to every unfinished descendant**, circular parent, overlapping timers, `due_source` transitions — **CLOSED, PASS** (PASS on the **fourth** review, at `a1f09b7`; it failed the first three — the history is kept below) | **≥90% coverage** on `internal/domain` + `internal/service` — **MET: 100.0% / 92.9%** |
| 2 | **Kanban + Habits strip** (launch screen) — **IMPLEMENTED, NOT CLOSED — S2-01 … S2-22 all committed, plus the gap-closing `7af4d1d` and round 1's two fixes `ae6befd` / `4b1af9c`; awaiting the re-review** — Wails bindings, Zustand hydrated from Go, 5 columns, dnd-kit drag of card+subtree, optimistic UI with rollback on error, full card chrome, habit strip w/ streaks, quick-add (Ctrl+N), command palette (Ctrl+K), theme/palette/accent in settings, EN/RU | **Create → move through every column → complete, keyboard only, no mouse** — demonstrated twice, by the automated flow test of S2-22 and by the hand script in `TASKS.md` — **and moving a card to Doing must itself open a `time_entry`** (§4 coupling, **D13**; ticket S2-03) |
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

### Stage 1 — CLOSED, PASS

The Reviewer returned **PASS** on the **fourth** review, verified at commit `a1f09b7`.
All twenty-two tickets **S1-01 … S1-22** are implemented and committed, one
conventional commit each, all authored solely by `Ismat
<mukhamejanov.ismat@gmail.com>` with no AI author and no co-author trailer. All five
gates are green, including gate 5, `wails build -tags webkit2_41`.

The **ACCEPT criterion is MET**. The bar is **≥90%** statement coverage on
`internal/domain` and `internal/service`, measured per package; re-measured after
`go clean -testcache` the actuals are `internal/domain` **100.0%** and
`internal/service` **92.9%**. (`internal/store` is at **86.4%** and is deliberately not
gated.) The Reviewer also confirmed independently that §7 D2 and the D9 sub-point are
correctly generalised — no surviving bullet states a *current* rule in `note`-only
terms — that no `.go` file changed after `9664506`, and that the tree was clean.

**The stage nevertheless failed review three times before that**, every time on the
same family of defects: a type rule spelled in more than one place, with one copy
diverging. Each divergence was a door illegal rows could walk through. The four-round
history is kept in full below, because it is the most useful thing in this document.

**First review — FAIL**, three blocking issues:

1. illegal type/status combinations were accepted on the create path — **fixed by the
   Dev in `fd5e31d`**;
2. due dates were written onto types that have no Kanban column — **fixed by the Dev in
   `f1802d7`**;
3. **decisions D8 and D9 existed nowhere in the specification** although five source
   files cited them as authoritative — **recorded in §7 below by the PM in `e7d74cc`**.

**Second review — FAIL**, on the remaining doors of the same rule:

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

The two user decisions those last two commits required were recorded in §7 as **D10**
and **D11** by the PM in `0854ed5`, which also generalised §4's derivation rule to
match them.

**Third review — FAIL**, one blocking issue: the last surviving divergence, the one
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

**K4** was re-written from "harmless, not scheduled" to **RESOLVED** by the PM in
`e3d626f`, which also recorded the predicate consolidation described next.

**The structural fix — Stage 2 must not undo it.** `9664506` also folded the four
separate spellings of "a project never enters `doing`" (`CanEnterDoing`,
`CheckStatus`'s sentinel selection, `PlanCascade`, `canBeTimed`) onto a single
predicate, **`domain.DoingRefusal(NodeType) error`**, with `NodeType.CanBeDoing()`
defined in terms of it; the richer, caller-specific error messages were kept. The
resulting type-rule inventory has **no rule spelled twice**: `HasColumn`, `HasDue`,
`DoingRefusal`/`CanBeDoing`, `countsAsWork`, the habit-requires-recurrence check and
`DefaultActivity` are each the single definition of their rule. Three review failures
came from a rule written down twice and then edited once; a new duplicate spelling is
the one thing **Stage 2** must not gain.

One honest caveat, recorded so nobody "tidies" it later: `countsAsWork`
(`HasColumn() && != NodeTypeProject`) and `DoingRefusal` admit **the same set of types
today**, but they answer **different questions** — "is this a unit of work?" (**D7**,
**D11**) versus "may this node be doing?" (**D9**). They were kept separate
deliberately, so that a future change to one does not silently move the other.

**Fourth review — PASS**, at `a1f09b7`. The Reviewer re-checked every fix from the
three failed rounds and confirmed: all five gates green including
`wails build -tags webkit2_41`; coverage re-measured after `go clean -testcache` at
**100.0% / 92.9%**, both over the ≥90% bar; §7 **D2** and the **D9** sub-point
correctly generalised, with an independent grep finding no remaining bullet that
states a *current* rule in `note`-only terms — that last doc fix is `a1f09b7` itself;
no `.go` file changed since `9664506`; tree clean; and no AI author or co-author
trailer on any commit. **Stage 1 is CLOSED.**

The eleven commits of the review cycle, in order: `fd5e31d`, `f1802d7`, `e7d74cc`
(round 1) · `1cbe582`, `f266bf5`, `81fb6e4`, `d5a170b`, `0854ed5` (round 2) ·
`9664506`, `e3d626f` (round 3) · `a1f09b7` (round 4, PASS).

#### Carried into Stage 2 — obligations, not suggestions

Stage 1 closed with five things owed to Stage 2. They are listed here and as explicit
acceptance criteria in `TASKS.md`, "Carried into Stage 2"; none of them may be dropped
silently.

1. **Wire `MoveToColumn(doing)` to `TimerService.Start`.** §4 couples them — "moving a
   card to Doing opens a `time_entry`" — but **no Stage 1 ticket did the wiring**, and
   the two services simply sit there composable. The Reviewer's warning stands: unless
   this is an explicit **Stage 2 acceptance criterion**, the coupling silently never
   ships. It is one, on the §5 Stage 2 row.
2. **`Board()` is O(n²·log n)** — `snapshot.view` rebuilds its index once per node.
   Fine at 200 nodes, not at 10k. **Fix it before the board is on screen**, not after.
3. **K1** — the `BackgroundColour`/`LC_NUMERIC` locale bug. Decide it with the palette
   work; three options are recorded below and none is chosen.
4. **K2** and **K3** — the two stale-stored-status consequences of **D11**, below.
   Both are Stage 2 rulings: K2 the card-state rule, K3 the card chrome.
5. **No rule may gain a second spelling — in Go *or* in TypeScript.** See the note
   immediately below.

#### Engineering note for Stage 2 — one rule, one spelling

**Three of Stage 1's four review rounds failed on the same defect**: a type rule
written down in two places and then edited in one. The inventory is clean as of
`9664506` — `HasColumn`, `HasDue`, `DoingRefusal`/`CanBeDoing`, `countsAsWork`, the
habit-requires-recurrence check and `DefaultActivity` are each **the single definition
of their rule**. Stage 2 must not introduce a second spelling of any of them.

This applies to **TypeScript exactly as it applies to Go**. The frontend renders what
Go returns and re-implements none of these rules: not "can this card go to Doing", not
"does this type get a column", not "does this count as work", not progress, not a
streak, not an overdue flag. A rule re-derived in a component is a second spelling, and
a second spelling is the defect that cost this project three review rounds.

### Stage 2 — IMPLEMENTED, NOT CLOSED

**Twenty-two tickets, S2-01 … S2-22, in `TASKS.md`. All twenty-two are committed**, one
conventional commit each, plus one gap-closing commit — `7af4d1d`, which added
`SetPriority` and wired the palette's priority rows (see *"The one plan correction"*
below). The stage is **implemented and awaiting review**; it closes only on a Reviewer
**PASS**, which has not happened. Nothing below may be read as "closed".

Measured on the tree as committed:

| | |
|---|---|
| `make check` | green — all five gates |
| `make cover` | `internal/domain` **100.0%**, `internal/service` **93.8%** (bar ≥90%) |
| `make front-test` | **22 files, 251 tests** |
| `make guard` | all **six** checks pass, with `GUARD_ALLOW_RE` **empty** |

**Five** things the Reviewer cannot verify on this machine are listed under
*"Owed to a hand pass"* below. They are owed, not done.

**Review round 1 returned FAIL on two blocking issues; both are fixed and the stage is
back with the Reviewer.** See *"Review round 1"* below. It is still **not closed** — a
stage closes only on PASS, and none has been returned.

The shape of the stage, and why it is in that order:

- **S2-01 … S2-08 are Go.** They pay off the two carried couplings (**C1**, **C2**),
  settle the three known issues (**C3**/**K1**, **C4**/**K2**, **C5**/**K3**), give the
  DTOs a stable wire contract, add the two service reads the launch screen needs and
  nothing in Stage 1 provided — the habit strip and the settings surface — and only then
  wire `main.go`/`app.go`, which today open no database and bind nothing but `Greet`.
  **The wiring is Stage 2's first real job and it is not optional**: every frontend
  ticket is blocked on it.
- **S2-09 … S2-13 are the frontend's foundation**: delete the scaffold demo, stand up a
  test runner and a mechanical rules guard, i18n, the appearance runtime, then the Go
  client and the Zustand store.
- **S2-14 … S2-22 are the launch screen** — card, columns, keyboard model, drag and
  drop, habits strip, quick-add, command palette, appearance controls, and finally the
  ACCEPT flow itself.

Four new decisions come out of this planning pass and are recorded in §7: **D12** (the
locale/background fix, given by the user), and **D13**, **D14**, **D15** — PM rulings
on the Doing↔timer coupling, on K2 and on K3, each open to the user's override. A fifth,
**D16**, was added *during* the stage: Aurora's background drift is gated and not drawn.
A sixth, **D17**, was added *after* it was implemented, out of the Reviewer's first
round: Go decides the habit check day and the frontend never names a date Go will act on.

**One planning defect, found and corrected mid-stage.** After S2-13 the Dev reported that
**no remaining ticket owned `frontend/src/App.tsx` or `frontend/src/main.tsx`** — so the
toast list, the board and both overlays would each have been built, tested and mounted by
nobody, every ticket would have passed, and the app would have opened blank with the
ACCEPT criterion undemonstrable. The tickets S2-14 … S2-22 were corrected in place: the
ticket that builds a top-level piece now mounts it, S2-15 creates the shell, and an
orphan check walks the import closure of `main.tsx` so an unmounted component turns a test
red on the commit that builds it. `TASKS.md`, *"Composition — who mounts what"*, carries
the rule. Worth recording as process, not just as a fix: the Dev found it by **refusing to
widen scope silently** and reporting it instead, which is exactly what that rule is for.

**The one plan correction — the brief wins over the plan.** `TASKS.md` listed *"the
tag/due/priority/estimate **editors**"* among the things Stage 2 must not do, while §2's
restatement of the brief names the Stage 2 command palette's action set as *"new, move to
column, **set priority**, start/stop timer, switch view, toggle theme, switch language"*.
Those two sentences contradicted each other, and the contradiction had teeth: S2-20
registered its four `set priority` rows as **unavailable, with a reason**, because no
binding set a priority — which left S2-20's own criterion, *"every action in the table is
reachable and executable by keyboard alone"*, unmet. The Dev **refused to widen into Go**
and reported it, which is the second time in this stage that rule caught a planning
defect rather than a coding one.

**The ruling: the brief wins.** Setting a priority **from the palette** is Stage 2; the
**priority editor in a detail panel** is Stage 3, along with the tag, due and estimate
editors. The out-of-scope line in `TASKS.md` said *editors* and meant *editors*; it was
simply too coarse to say so. `7af4d1d` added `TaskService.SetPriority`, the `App`
binding, the regenerated client and the palette wiring, with the 1..4 range stated
**exactly once**, in `domain.Priority.Valid`. This is recorded as a **plan correction,
not a scope change by the Dev** — the Dev did the right thing by stopping.

**Three disclosed scope widenings, all ratified. The Reviewer does not need to
re-litigate them.** Each was minimal, each was stated in its commit body naming the file
and the reason, and each falls under the general rule set in S2-13's ratification: a
widening that is minimal, disclosed, and needed to avoid leaving a rule with two
spellings is acceptable; an undisclosed one, or one that adds behaviour, is not.

| Ticket | Widening | Ruling |
|---|---|---|
| **S2-13** (`97873d9`) | `frontend/src/main.tsx`, six lines | **RATIFIED** (already, in `TASKS.md`); the zustand store subsumed two hand-rolled stores, so enforcing Scope would have *preserved* a duplicate |
| **S2-21** (`642e677`) | six existing test files plus a new shared `frontend/src/test/render.tsx` | **RATIFIED.** Mounting the header put focusable elements ahead of the board, so six tests asserting *"the first `Tab` lands on the board"* were mechanically wrong. The fix introduces `tabUntil(user, arrived)` — **one spelling instead of six hard-coded tab-stop counts** — so it **removed** duplication rather than adding it. **No assertion was weakened**; every one still makes the same claim about the same element |
| **S2-22** (`58552bf`) | the `Makefile`, to add `make guard` check 6 | **RATIFIED.** The ticket's own criterion is *"`make guard` fails if a mouse event is introduced into the accept test"* — the rule was **required** to live in `make guard`, so the ticket's Scope list was simply incomplete. Not a widening in substance, a Scope-list omission. **The five gates are still five**; `guard` is not one of them |

**Review round 1 — FAIL on two blocking issues, both now fixed.** Neither was a missing
feature; both were a rule with a second spelling, which is the same family that cost
Stage 1 three rounds.

1. **The habit check day was computed in TypeScript** and sent to Go, while Go computed
   its own today for the flags the same strip renders. **Two clocks, one rule.** Fixed in
   `ae6befd` and, more importantly, *decided*: the rule had only ever been written in a
   code comment, and is now **D17** in §7 — with its rejected alternative (publishing
   Go's today for the frontend to hand back) and its accepted consequence (the dated
   `Check`/`Uncheck` are service-only until Stage 7 binds them) on the record.
2. **`make guard` check 2 was case-sensitive while its own comment claimed the check was
   EXACT**, so `'Done'`, `'Doing'` and `'Today'` walked past a grep that stopped
   `'done'`. The Reviewer proved it with a line spelling `domain.Status`'s rule a second
   time in TypeScript. Fixed in `4b1af9c` — `-E` became `-iE` — and verified both ways:
   the proof line passes the old grep and fails the new one, and `-iE` returns **zero
   hits** across `frontend/src` as it stands, so nothing is grandfathered and
   `GUARD_ALLOW_RE` **stays empty**. The same comment said *"Five checks"* while the
   recipe runs six; corrected to match.

**A guard is not a proof, and this is now written down.** Checks 3a/3b are **name-based
heuristics** over `overdue|derive|streak|progress|percent`, so a derivation named
anything else — `wireDate`, for instance — is invisible to them by construction. The
Reviewer walked a recomputed today, a recomputed overdue flag and an inline percentage
past a clean `make guard` to show it. The Makefile discloses it; **D17** records it in
the plan, because the failure mode is a reader treating a green `guard` as evidence that
the frontend computes nothing. It is evidence that five names are absent.

**Dead TypeScript time-derivations were deleted in the same commit** (`ae6befd`):
`displayElapsedSeconds`, its `timerReadAt` input, `timerStartedAt`, and `format.ts`'s
`formatTime` and `formatDuration`, with their orphaned tests. All of it was wall-clock
arithmetic sitting on top of Go's `elapsedSeconds`, **tested but rendered by no
component** — which is exactly how a second implementation of a rule waits for its first
caller, and exactly the shape that failed Stage 1 three times. It was removed before it
could be wired up. **No ticker was added**; drawing the running clock is Stage 3's, and
that ticket can add precisely what it renders.

**Owed to a hand pass — not verified, and no document may imply otherwise.** There is no
display on this machine and `xvfb-run`, `scrot`, `import` and `grim` are all absent, so
the Reviewer cannot execute these. They are owed to a human at a real keyboard in front
of a real window:

1. **Steps 1–10 of the hand script** in `TASKS.md` against `./build/bin/nexus`,
   including the **due-badge assertions at steps 3, 4 and 7** — the automated flow's fake
   does not model **D8**'s due rewrite, so the badge claims have no mechanical backing at
   all.
2. **No flash of the wrong background on the first frame, in *both* themes** (**D12**,
   **K1**). GTK-side; jsdom cannot see it.
3. **Russian not clipping at a real 1024×768.** jsdom has **no layout engine** — every
   `offsetWidth` is 0 and nothing ever overflows — so the suite asserts the *mechanisms*
   that prevent clipping (`flex-wrap`, `break-words`, no `truncate`, no `nowrap`, no
   fixed widths) and **never** the absence of clipping. The mechanisms being right is not
   the same claim as the text fitting.
4. **The `:focus-visible` accent ring actually painted, in both palettes.** jsdom
   evaluates `:focus-visible` as **false** for programmatic focus, so what is asserted is
   reachability and focusability, not a painted ring.
5. **S2-07's own last criterion** — *"the app opens a window and the frontend can call
   `Board()` and receive five columns. Verified by hand and reported."* It is a
   hand-verification written into a ticket, it is as unverifiable here as the other four,
   and it was **missing from this list** until the Reviewer enumerated it. `make build`
   proves the binary links; nothing on this machine proves a window opens or that the
   first `Board()` over the real IPC bridge returns five columns.

And, held to the same standard: **the Aurora drift gate is implemented and audited both
ways, but nothing draws a drift, so nobody has watched one pause** (**D16**, **K5**).
That is not on the list above, because it is not owed to a human either — there is
nothing to look at. See K5 in §7.

**Two known limits found during the work, recorded so nobody over-trusts them.**

- **The orphan check has a blind spot.** `App.mount.test.tsx` walks the **static import**
  closure from `main.tsx`, so **deleting an element while keeping its import leaves the
  check green** — the Dev demonstrated this deliberately, twice, by removing
  `<CommandPalette />` and later `<AppearanceControls />` from `App.tsx` and watching the
  orphan walk stay green while the per-ticket reachability tests went red. That is the two
  halves of the composition check doing **different** jobs, and it is exactly why the
  per-mounting-ticket `render(<App />)` criterion exists alongside the walk. Neither half
  is sufficient alone.
- **The palette's remaining unavailable rows are honest and transient.** After `7af4d1d`
  only `view:tree`, `view:calendar` and `view:kanban` are registered unavailable, and each
  carries a reason tied to a **later stage** ("you are looking at it", "coming in stage
  N") rather than a permanent limitation. No row silently does nothing.

#### Carried into Stage 3 — one item, and it is not a defect

**Enum membership is spelled a second time, in the locale files, with nothing checking
it.** `commands.ts` and `AppearanceControls.tsx` take `Object.keys` of
`settings.palette`, `settings.theme`, `settings.language`, `palette.priority` and
`card.type` from `en.json` as the **authoritative sets**. Go owns all five —
`domain.Palettes()`, `domain.Themes()`, `domain.Priorities()` and the node types — and
publishes **none** of them over the wire. `locales.test.ts` checks en/ru key parity and
nothing ties either file to Go, so if Go gains a palette or a priority the UI silently
will not offer it and **no test goes red**.

The Reviewer ruled this an **acceptable judgement call for Stage 2, not a defect**, and
the instructive contrast is inside the same stage: **where Go does publish a set — the
columns — the frontend takes it from Go and the locale file only names it.** That is the
shape the other five should converge on.

Stage 3 resolves it one of two ways, and picks one: **bind a set-publishing method** so
the frontend enumerates what Go enumerates, or **add a parity test** that fails when a Go
set and its locale table disagree. **Nothing is scheduled now** and no Stage 2 ticket is
reopened for it.

**Three things Stage 2 is explicitly forbidden from doing.** They are the shape of
Stage 1's four review rounds, turned into rules up front:

1. **No rule may gain a second spelling — in TypeScript any more than in Go.** The
   frontend renders what Go returns. `make guard` (S2-10) greps for the specific
   violations: a status string literal in a component, a recomputed overdue flag, a
   recomputed percentage, a re-derived column eligibility.
2. **No hex literal in `frontend/src`, ever**, and `design/` stays read-only.
3. **No hard-coded user-visible string**, from the very first component.

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

The open questions are **closed**. Referenced as **D1–D17** and **E1–E3** from tickets
in `TASKS.md`.

**D1–D12 were given by the user and are authoritative** — they override anything
earlier in this document that contradicts them. **D1–D7** were settled before Stage 0.
**D8**, **D9**, **D10** and **D11** were confirmed by the user *during* Stage 1, when
implementation exposed questions the earlier decisions did not answer; they carry the
same authority. D10 and D11 came out of the second review, and **D10 generalises the
derivation rule in §4**, which has been amended accordingly. **D12** was confirmed by
the user while Stage 2 was being planned and closes **K1**.

**D13, D14, D15, D16 and D17 are PM rulings.** D13, D14 and D15 were made during Stage 2
planning because `TASKS.md` **C1**, **C4** and **C5** demanded a decision and the user's
brief did not contain one; **D16** was made *during* Stage 2, when the Dev asked what
draws Aurora's background drift and correctly declined to invent it; **D17** was made
*after* Stage 2 was implemented, when the Reviewer's first round found the habit check
day being computed in TypeScript on a decision that had been *"made in a code comment"*
and never written down. They are written in
the same form and bind the Dev exactly as the rest do — a rule with no single written
spelling is the defect that cost Stage 1 three review rounds — but their provenance is
different and **the user may overturn any of them**. If one is overturned, the ticket
that implements it changes with it; nothing else in this document depends on them.

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

**Amended during Stage 2 planning — the design export names no components.** The second
bullet assumed a component list in the export. There is none: `design/` holds
`README.md`, `tokens.css`, `tailwind.config.js`, `SKILL.md` and
`pmp-timelog-format.md`, and between them they specify tokens, fonts, semantics and
motion and **not one component name**. See the note at the end of §3. The bullet
therefore reads, from Stage 2 on: **components are hand-built against the tokens**, and
ReactBits is used only where it genuinely fits and can be re-skinned entirely to the
token names. The clause "as named in the design export" is inoperative because the
antecedent does not exist — and **inventing a component spec and attributing it to the
export is forbidden**, which is the failure mode this amendment exists to prevent.

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
  there is nothing under it to derive from. See known issue **K2**, now settled by
  **D14**.

### D12 — Force `LC_NUMERIC=C`, and seed the window background from settings (confirmed during Stage 2 planning; closes K1)

**Two halves, one decision.** Before `wails.Run`, the process forces `LC_NUMERIC=C`.
And `options.App.BackgroundColour` is **seeded from the palette/theme the user last
chose**, read out of the `settings` table at startup, instead of a constant compiled
into `main.go`.

The bug is **K1**: Wails formats the window background in C at `window.c:205` using the
**process** locale, so under this machine's `LC_NUMERIC=ru_RU.UTF-8` it emits
`rgba(27, 38, 54, 0,0)` — a comma where GTK's CSS parser needs a decimal point — and GTK
**silently drops the whole declaration**. Nothing is logged. Forcing the C locale for
numeric formatting makes the colour actually reach the window, which is what makes the
second half worth doing at all.

- **Why seed from settings.** The window background is painted by GTK *before* the
  WebView has loaded anything, so it is the first colour the user sees. A hardcoded
  dark value flashes wrong for a user on the light theme, and a hardcoded light value
  flashes wrong on dark. Reading `palette` + `theme` out of `settings` — which already
  hold the two values and already default to `aurora` + `dark` (**D6**) — makes the
  first frame match the last session in **both** themes.
- **Rejected alternative (a): leave the window transparent and let the frontend paint
  it.** The frontend is exactly where the palette lives, so this is tempting — and it is
  the option that guarantees the flash, because "transparent" is a real colour for the
  hundreds of milliseconds before the bundle evaluates.
- **Rejected alternative (b): patch upstream Wails and pin the fork.** A one-line
  upstream formatting bug is not worth a fork's maintenance cost, a vendored build step
  and a divergence to re-apply on every Wails release.
- **Accepted consequence, stated by the user:** the process's numeric locale is changed
  out from under any Go code that might want the user's locale for number formatting.
  Nothing in Nexus formats numbers through the C locale — Go's own `strconv` and
  `fmt` are locale-independent by design — so the only thing affected is the cgo/GTK
  layer, which is the thing being fixed.

**The duplication risk, and the rule that contains it.** This makes `main.go` a
**second place that knows a background colour**, and a second spelling of a value is the
defect class §5 spends three review rounds on. The containment is absolute and is an
acceptance criterion on ticket **S2-08**:

> **`main.go` must not contain a hex literal or an RGB triple.** The colour is derived
> from the `palette` and `theme` values read out of `settings`, against the token values
> owned by `design/tokens.css`. `git grep -nE '#[0-9a-fA-F]{3,8}|RGBA\{' main.go` must
> return nothing.

Because `design/` is read-only and Go cannot import CSS, the mapping from
(palette, theme) → `--bg` is generated or parsed from `design/tokens.css` rather than
retyped. Retyping four hex values into Go is precisely the thing this paragraph forbids.

### D13 — What the Doing↔timer coupling actually does (PM ruling, Stage 2; implements C1)

`PLAN.md` §4 says *"Moving a card to Doing opens a `time_entry`"* and stops there. Three
questions it does not answer had to be settled before **S2-03** could be written, because
each of them is a rule and a rule with no written spelling gets invented twice.

1. **A direct move of a timeable node to `doing` opens a timer, in the same
   transaction as the move.** Not after it, not in a second call from `app.go`. If the
   move commits and the timer does not, §4's coupling is a lie for exactly as long as it
   takes the user to notice.
2. **A cascade opens no timer.** Dragging a parent to Doing cascades `doing` onto every
   unfinished descendant with a column (**D2**), which could be twelve nodes; the single
   active timer is global (§4), so at most one of them could have it and there is no
   principled way to choose. *Rejected alternative:* start the timer on the first
   leaf in sort order — it picks a card the user did not point at, and the user then
   has to notice and stop it. A cascade is a planning gesture, not a "start working now"
   gesture.
3. **Moving a node out of `doing` closes its open entry** — to any column, `done`
   included. *Rejected alternative:* leave the entry open, since only `Stop` closes
   timers. That leaves a running timer on a card sitting in Done, which is both wrong
   in the data and visibly wrong on screen: `TimerView.Running` would be true on a
   finished card.

- **A move that `domain.DoingRefusal` refuses — a project — opens nothing**, because the
  move itself fails. This is not a fourth rule; it follows from **D9** and must not be
  re-stated as one.
- **The existing single-active invariant is untouched.** Moving a second card to `doing`
  closes the first card's entry exactly as `TimerService.Start` already does. The
  coupling reuses that code path; it does not get its own.
- **Accepted consequence:** time is attributed to the card the user dragged, and only
  ever to that one. Time spent on a subtask the user never dragged is not recorded by
  dragging its parent, which is honest — §4 already says *"never billed for time you did
  not spend"* about sleep, and the same principle applies here.

**Implementation consequence — one entry point, no bypass (S2-03).** There must be
exactly **one** move-to-column method reachable from `app.go`, and it must be the
coupled one. Two methods, one coupled and one not, is a second spelling with a
call-site-shaped fuse.

### D14 — Archiving re-inspects a node that becomes a leaf (PM ruling, Stage 2; closes K2)

**When archiving leaves a node with no remaining child that has a Kanban column, that
node's stored status is rewritten to the status it derived immediately before the
archive**, with `completed_at` set or cleared to match.

This is the ruling **K2** asked for, generalised from "a project" to "any node", because
the situation is not special to projects — a project is only where **D11** gave it
teeth.

- **The failure it closes.** `P{C1:done, C2:backlog}` derives `backlog` and renders in
  the Backlog column. An earlier drag of `P` to Done wrote `done` into `P`'s *stored*
  status, where it has been dead weight ever since. Archive `C1` and `C2`: `P` is now a
  leaf, **D11** says a leaf project counts as one work leaf decided by its **stored**
  status, and `P` silently becomes a **done** unit in its parent's bar — on a status
  nobody set, minutes after the board showed it as Backlog. The state is
  self-consistent, which is exactly why it never announces itself.
- **The principle: continuity.** What the board showed before the archive is what it
  shows after. Archiving a child is a statement about the child, and it must not change
  what the parent claims about itself.
- **Rejected alternative (a): reset the stored status to `backlog`.** Simple, and wrong
  in the common case — archiving the finished children of a finished project would
  un-finish it.
- **Rejected alternative (b): never trust a leaf project's stored status; treat it as
  unfinished always.** This reverses **D11** for the case D11 was written for, and makes
  a genuinely completed project impossible to record.
- **Rejected alternative (c): leave it, K2 is self-consistent.** Self-consistent and
  wrong is the worst of the three states, because nothing will ever flag it.
- **Accepted consequence:** archiving is no longer a pure "hide these rows" operation —
  it can write a status onto a node it did not archive. That write is bounded, and the
  boundary is checkable: **only** onto an ancestor that the archive turned into a leaf,
  and **only** to the value that ancestor was already displaying.

**Implementation notes (S2-06) — the rule gets one spelling.** "Has no remaining child
with a column" is `NodeType.HasColumn` and `Node.IsLeaf` (**D10**), asked of the set as
it will be *after* the archive; the status written is `domain.DeriveStatus` of the set as
it was *before*. Both already exist. **No new predicate.** Restore needs no rule: once a
node has column-bearing children again, derivation takes over and the stored status stops
being consulted.

### D15 — A card whose progress is undefined says so (PM ruling, Stage 2; closes K3)

**A node whose `Progress.Defined` is false draws no bar and renders an explicit, localised
"empty project" marker in its place** — one line, `muted`, in the slot the bar would have
occupied. It never renders a percentage, a `0/0`, or an empty bar track.

- **The failure it closes — K3.** An empty project stored `done` sits in the Done column
  with **nothing on it at all**: no bar, because its own progress is undefined
  (**D11** part 1), and no other chrome that says anything. It is the one place a
  finished card is blank.
- **The rule is about `Defined`, not about `done`.** An empty project in Backlog has the
  same problem in a quieter way, and a rule that fires only in the Done column would be
  a rule with a column in it. One condition, one marker, every column.
- **Rejected alternative (a): render a 0% bar.** `Defined == false` exists precisely
  because **neither 0% nor 100% is true** (**D7**, **D9**). Drawing an empty track is
  rendering 0% with extra steps.
- **Rejected alternative (b): render nothing and rely on the card's type icon.** That is
  the current behaviour and it is what K3 reports as a defect.
- **Rejected alternative (c): have Go return a "display this instead of a bar" string.**
  Tempting given §1, and wrong: the *rule* ("undefined progress means no bar") is already
  in Go, on the `Defined` flag. What the absence looks like is presentation, and a
  user-visible string returned from Go is a string that cannot be translated by the i18n
  layer.
- **Accepted consequence:** an empty project is visibly empty in every column, including
  Done. A card that reads "done" and "empty" at once is the honest rendering of the state
  **D11** deliberately allows.

**Implementation notes (S2-14):** one branch, in the card's progress slot, on
`progress.defined`. The strings are i18n keys in `en.json`/`ru.json` like every other
string. The frontend still computes **nothing** — it branches on a flag Go set.

### D16 — Aurora's background drift is gated in Stage 2 and drawn in a later one (PM ruling, during Stage 2; opens K5)

**The question.** `design/README.md` says, in one sentence: *"Aurora's background drift
(~60s) must also pause"* under `prefers-reduced-motion`. S2-12 built the gate —
`auroraDriftEnabled()` decides it in one place, respects the media query, and publishes
`data-drift="on"|"off"` on `<html>`. **No ticket ever asked anyone to draw the drift**,
so the attribute has no consumer. The Dev raised it rather than inventing a visual, which
was correct: `design/` holds `README.md`, `tokens.css`, `tailwind.config.js`, `SKILL.md`
and `pmp-timelog-format.md`, and **not one of them specifies a drift** — and §3 and **D6
as amended** forbid inventing a spec and attributing it to the export.

**The ruling: the gate ships in Stage 2; the visual does not.**

That sentence is a **constraint on a drift, not a specification of one**. It fixes two
things — a period of about sixty seconds, and that it pauses under reduced motion — and
leaves everything needed to actually draw it unsaid: how many layers, what geometry, what
opacity, what path, whether it is a CSS animation or a canvas, whether it sits behind the
board or behind the whole page. Building it means inventing at least four properties and
shipping them under the design export's name. That is the precise failure D6's amendment
exists to prevent, and it is worse than shipping nothing, because an invented drift is
hard to tell from a specified one six months later.

Against that, the cost of not having it is small and bounded: the drift moves no card,
blocks no key, changes no contrast decision and has **zero bearing on the ACCEPT
criterion**. It is decoration.

**Rejected alternatives**, both of which were live:

1. *Invent a modest drift and call it minimal* — two translating radial gradients in
   `accent`/`accent-2`, say. Rejected: "modest" is still invented, and the token names
   would give it a legitimacy the spec never granted.
2. *Delete the gate as dead code.* Rejected: the gate is not dead, it is **early**. It is
   correct, tested, and it is the part that is genuinely hard to retrofit — a motion
   feature built first and made reduced-motion-safe afterwards is how a11y regressions
   ship. `data-drift` is a published contract waiting for a consumer, and
   `appearance.ts` says so in a comment.

**What this obliges, and it is the part that matters.** Nobody may claim the drift was
verified. S2-22's a11y audit reports that `data-drift` is `off` under
`prefers-reduced-motion` and `on` otherwise, and reports that **nothing is drawn behind
it** — it does not report having watched a drift pause. The ACCEPT verification table in
`TASKS.md` listed exactly that hand-verification and has been corrected.

**When it comes back.** It becomes a Stage 3 ticket the moment a drift specification
exists — from the user, or from an authorised PM ruling that invents one *openly and in
this document* rather than silently in a component. At that point the ticket is only the
drawing: the gate, the media query and the test are already built and already green. The
standing consequence is recorded as **K5** below.

### D17 — Go decides the habit check day; the frontend never names a date Go will act on (PM ruling, during Stage 2, on the Reviewer's first-round FAIL)

**The failure.** `frontend/src/store/data.ts` carried a helper called `wireDate` that
built the **local calendar day in TypeScript** and passed it to
`CheckHabit(nodeID, date)`. Go meanwhile computed **its own today** at
`internal/service/habit.go` — `domain.Today(s.clock)` — for the `checkedToday` and
`scheduledToday` flags **the same strip renders**. Two clocks, one rule: §1's central
commitment — *"all domain logic lives in Go; the frontend renders what Go returns"* —
broken verbatim, on the one screen where both answers are visible at once.

It is not theoretical. Across a local midnight the user presses `Space`, the frontend
writes a check for **yesterday**, Go answers `checkedToday: false`, the optimistic tick
reverts, and the user sees nothing happen while a check lands on a day they never chose.

**And nobody had ruled on it.** In the Reviewer's words: *"no ticket or decision ever
ruled on it — the decision was made in a code comment."* That is what this entry fixes;
the code fix is `ae6befd`.

**The ruling: the frontend never names a date that Go will act on.**

- `HabitService.CheckToday` / `UncheckToday` take a node id and nothing else and
  **delegate to the existing dated `Check`/`Uncheck`** with `domain.Today(s.clock)`.
  The rule keeps **one spelling** — the same `domain.Today(clock)` the strip already
  derives `checkedToday` against. A second `time.Now()`-shaped path would have been the
  same defect in Go.
- The bindings become `CheckHabitToday(nodeID)` / `UncheckHabitToday(nodeID)`. The dated
  pair is **no longer bound at all**, deliberately: an unbound method is a door a
  computed date cannot be walked through.
- **`wireDate` was deleted, not bypassed.** A helper that still builds a date is a helper
  the next ticket calls.
- The dated `Check`/`Uncheck` **survive on the service for Stage 7's calendar**, where
  the user *picks* a day rather than the frontend *computing* one. A date the user chose
  is data; a date the frontend worked out is a derivation.

**Rejected alternative: publish Go's today on a DTO and let the frontend hand it back.**
It looks like it honours the rule, because the value originates in Go — and it does not.
The frontend still names the date on the wire, the value is still stale by exactly the
interval between the read and the write, and the midnight bug survives with an extra
round trip in front of it. The fix for a value the frontend should not compute is to
**stop sending it**, not to source it more respectably.

**Accepted consequence:** `HabitService.Check`/`Uncheck` are **service-only until Stage 7
binds them** — reachable from Go, unreachable from TypeScript, and exercised only by
their tests. That is deliberate, it is the door the calendar will use, and it must not be
removed as dead code.

**The part that generalises — `make guard` cannot see this class of defect.** Its checks
**3a and 3b are name-based heuristics**: they fire on identifiers containing
`overdue|derive|streak|progress|percent`. `wireDate` is named after none of them, so the
guard stayed green over a recomputed today for the whole stage. The Reviewer demonstrated
the blind spot deliberately, walking **a recomputed today, a recomputed overdue flag and
an inline percentage** past a clean `make guard`. The Makefile already discloses that
checks 3 and 4 are heuristics; it is recorded here too so that **no report, review or
status line may offer a green `make guard` as evidence that the frontend computes
nothing**. It is evidence that the *named patterns* are absent. Reading the diff remains
the check, and a derivation renamed away from those five words is invisible to the grep
by construction.

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

### Known issues — all decided, none deleted

**All five are now DECIDED, none is deleted.** K1, K2 and K3 were carried into Stage 2
as things to rule on, and Stage 2 planning ruled on all three: **K1 → D12**,
**K2 → D14**, **K3 → D15**. K1 and K2 are **implemented** (`2fe9e84`, `bff9a82`); K3's
ticket is **S2-14** and is still ahead. Each stays on the record below with its decision
named, so the failure mode remains visible and so nobody re-opens a question that has an
answer. **K4 is RESOLVED** in `9664506`, likewise kept. **K5** was opened *during* Stage
2 and decided in the same pass by **D16**.

**K1 — `BackgroundColour` never reaches GTK under a comma-decimal locale.
DECIDED by D12; implemented by S2-08.**
This machine runs `LC_NUMERIC=ru_RU.UTF-8`. Wails builds the window background as a
CSS string in C at `window.c:205` and formats the alpha with the **process locale**
rather than the C locale, so it emits

```
rgba(27, 38, 54, 0,0)
```

GTK's CSS parser rejects that declaration — a comma is an argument separator, not a
decimal point — and the background colour is therefore **silently never applied**.
Nothing is logged; the window just uses its default.

This is an **upstream Wails bug**, not ours, and it was **not fixed in Stage 1**:
Stage 1 is Go-only, has no window work in it, and a locale workaround bolted on then
would have been untestable until there was a palette to compare against.

Three options were recorded here and carried into Stage 2 as **C3**:

1. force `LC_NUMERIC=C` for the process before `wails.Run`;
2. leave the window background transparent and let the frontend paint it, which is
   where the palette lives anyway;
3. patch upstream and pin the fork.

**Option 1 was chosen by the user during Stage 2 planning, together with a second
half the options above did not contain: seed the background colour from the persisted
palette/theme rather than from a constant.** That is **D12** in §7, which records the
reasoning, the two rejected options and the containment rule that keeps a second hex
literal out of `main.go`. It is implemented by ticket **S2-08**.

Related but **separate**: `main.go` passed `&options.RGBA{R: 27, G: 38, B: 54, A: 1}`.
Alpha is 0–255, so `A: 1` is ~0.4% opacity — a real bug of ours, fixed in Stage 1 by
ticket **S1-02**. Fixing it does not fix K1, and K1 is why the fix cannot be verified
by eye on this machine.

**K2 — a leaf project's stale stored status now has teeth. DECIDED by D14; implemented
by S2-06.**
Since **D11**, a leaf project's *stored* status decides whether it counts as done in
its parent's denominator. Archiving the last real child of a project that an earlier
cascade had written `done` leaves it counted as a **done** unit on a status nobody set
deliberately. The state is self-consistent — column and bar agree — so this is not a
contradiction, which is exactly why it will not announce itself. It wanted a rule about
**re-inspecting a project's stored status when its last child is archived**.

**The ruling is D14**, taken during Stage 2 planning: when archiving leaves a node with
no remaining child that has a Kanban column, that node's stored status is rewritten to
the status it *derived immediately before the archive*, with `completed_at` set or
cleared to match. The principle is continuity — what the board showed before the
archive is what it shows after. See D14 for the three rejected alternatives.

**K3 — an empty project stored `done` renders in the Done column with no bar at all.
DECIDED by D15; implemented by S2-14.**
Its own progress is undefined (**D11**, part 1), so no bar is drawn, while its status
puts the card in Done. Not a contradiction, but it is **the one place a finished card
shows nothing**.

**The ruling is D15**, taken during Stage 2 planning: a card whose `Progress.Defined`
is false draws **no bar** and renders an explicit, localised "empty project" marker in
the bar's place — in every column, not only in Done, because the condition is
`Defined == false` and not `status == done`. It never renders a percentage, a `0/0` or
an empty track: `Defined == false` exists precisely because neither 0% nor 100% is true.

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

**K5 — `data-drift` is a gate with nothing behind it. DECIDED by D16; deliberately out
of Stage 2, and no ticket is scheduled.**
S2-12 built the reduced-motion gate for Aurora's background drift, exactly as
`design/README.md` requires, and **nothing draws the drift**, because `design/` specifies
none and inventing one is forbidden (§3, **D6** as amended). So
`document.documentElement.dataset.drift` is set correctly, on every palette and under
both motion preferences, and no code reads it.

The failure mode this is written down against is not the missing decoration — it is the
**false green**. S2-12's criterion *"under `prefers-reduced-motion: reduce`, the Aurora
drift is not running"* is satisfied **vacuously**: nothing is running under any
preference. A criterion that passes for the wrong reason is worse than a red one, so the
vacuity is stated here, in S2-12's ticket, and in S2-22's audit instructions, and the
ACCEPT table's claim that a human would watch the drift pause has been removed.

**As implemented and audited (S2-22, `58552bf`), this still holds exactly.** The audit
checked the gate **both ways** — `data-drift` is `on` without reduced motion and `off`
with it, on every palette — and reported, in those words, that **nothing draws the
drift** and that nobody has watched one pause. That is the correct report, and no
document in this repository may be edited into implying the visual exists.

**The ruling is D16.** The gate is early rather than dead — retrofitting reduced-motion
safety onto a shipped animation is how a11y regressions happen — and it is exactly the
half that is hard to add later. This becomes a Stage 3 ticket when a drift specification
exists; until then `appearance.ts` carries a comment saying what `data-drift` is for and
who is expected to read it.

---

**Status: decisions locked — D1–D17, E1–E3. Stage 0 is CLOSED (PASS). Stage 1 is
CLOSED (PASS) — all twenty-two tickets, S1-01 … S1-22, ACCEPT met at 100.0% / 92.9%,
PASS returned on the fourth review at `a1f09b7` after three FAILs whose history is kept
in §5. **Stage 2 is IMPLEMENTED, NOT CLOSED**: twenty-two tickets, S2-01 … S2-22, in
`TASKS.md`, **all twenty-two committed**, plus the gap-closing `7af4d1d`. `make check`
green, `make cover` 100.0% / 93.8%, `make front-test` 22 files / 251 tests, `make guard`
clean over all six checks with an empty `GUARD_ALLOW_RE`. **Review round 1 returned FAIL
on two blocking issues — a habit check day computed in TypeScript, and a `make guard`
check that claimed to be exact and was case-sensitive — and both are fixed (`ae6befd`,
`4b1af9c`). The Reviewer has not returned PASS; a stage closes only on PASS.** All five
carried obligations are absorbed into
tickets — C1 → S2-03, C2 → S2-01, C3 → S2-08, C4 → S2-06, C5 → S2-14 — and every known
issue now has a decision: **K1 → D12** (user), **K2 → D14**, **K3 → D15** and
**K5 → D16** (PM rulings, overturnable), **K4** RESOLVED in `9664506`. Nothing is open.
**Two planning defects were found mid-stage, both by the Dev refusing to widen scope
silently, and both corrected in `TASKS.md`:** nothing owned `App.tsx` or `main.tsx`, so
nothing mounted anything; and the out-of-scope line contradicted the brief on *set
priority*, resolved in the brief's favour. Three disclosed widenings (S2-13, S2-21,
S2-22) are **ratified**; **D17** records the one rule the stage had never written down;
and **five** checks are recorded as **owed to a hand pass, not verified**. One item is
**carried into Stage 3** — the five enum sets spelled in the locale files with nothing
tying them to Go — as an acceptable judgement call, not a defect, and it is not
scheduled. A green `make guard` is **not** proof that the frontend computes nothing;
checks 3a/3b are name-based heuristics and **D17** says so. See §5,
"Stage 2 — IMPLEMENTED, NOT CLOSED", and `TASKS.md`,
"Composition — who mounts what".**
