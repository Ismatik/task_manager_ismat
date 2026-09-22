# Nexus — Plan

My restatement of the brief, written before any code. It stays the source of truth for
the data model and the decisions; only the stage status below moves. **Stages 0, 1 and 2
are all CLOSED (PASS).** Stage 2 closed on the **second** review, at `8818db4`; round 1
returned FAIL on two blocking issues, both fixed (`ae6befd`, `4b1af9c`, and the ruling
they were missing is **D17**). **Stage 3 is PLANNED and not started.** See §5 and
`TASKS.md`.

**Stage 2's PASS was mechanical, and the user has since run the app by hand on real
hardware — the thing no gate on this machine can do. It found nine defects.** They are
recorded below as **K6 – K14**, each traced to an exact line by a read-only
investigation, and each ruled on in §7 as **D18 – D26**. Eight of them are invisible to
every gate this repository has, because jsdom has no layout engine and no font engine, and
because **input cannot be driven here** (**E4** — capture can, and the belief that it could
not is itself part of why these nine waited for the user's own laptop). **Stage 3 therefore opens with a blocking remediation block,
S3-01 … S3-09, which must be green before a single feature ticket starts** — that was the
user's condition for proceeding. This is **D17's lesson arriving in a second place**: the
Russian anti-clipping audit was green *on a real clipping bug*, because it asserts the
absence of `truncate` and `whitespace-nowrap` and `shrink-0` was not on the list.

**The two questions Stage 3's planning left open have since been answered by the user and
are closed.** **OQ1 → Inter** (which face to vendor; **D23** now names it and **S3-06** is
unblocked) and **OQ2 → the middle option** (`COUNT` and `UNTIL` enter the recurrence
language; **D27**, the user's, with **D28** ruling on what an ended series means for
**D5**). OQ2's answer added a ticket, the block A review added five more, and S3-34's own
report added a sixth: Stage 3 is now **thirty-five** tickets — **S3-19** widens
`internal/domain` before **S3-23** can offer an end condition, **S3-30 … S3-34** close
**K15**, **K16** and **D32**'s four observations and build the capture harness **D31**
rules on, and **S3-35** closes **K17** (**D34**).

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
- **A recurrence may end.** Since **D27** the rule may carry `COUNT` or `UNTIL`, and past
  that bound the series has **no further scheduled occurrences**. By D5 nothing can then
  break the streak, so it **freezes** at its final value; the habit stays in the strip,
  marked ended by a Go-computed field, until it is archived. See §7 **D28**.

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
| 2 | **Kanban + Habits strip** (launch screen) — **CLOSED, PASS** (PASS on the **second** review, at `8818db4`; round 1 returned FAIL on two blocking issues, fixed by `ae6befd` and `4b1af9c` — the history is kept below). S2-01 … S2-22 all committed, plus the gap-closing `7af4d1d` — Wails bindings, Zustand hydrated from Go, 5 columns, dnd-kit drag of card+subtree, optimistic UI with rollback on error, full card chrome, habit strip w/ streaks, quick-add (Ctrl+N), command palette (Ctrl+K), theme/palette/accent in settings, EN/RU | **Create → move through every column → complete, keyboard only, no mouse** — **MET**: `frontend/src/App.accept.test.tsx` drives the assembled `<App />` with an exact asserted call sequence (one `CreateNode`, then `MoveToColumn` with each of the four statuses Go supplied, in order), and **`make guard` check 6 fails if a mouse event ever enters that file** — **and moving a card to Doing itself opens a `time_entry`** (§4 coupling, **D13**; ticket S2-03), asserted from the UI. The **ten-step hand run on the real binary is still owed** and is carried into Stage 3 |
| 3 | **Defect remediation, then Detail + Tree + Search/Archive** — **IN PROGRESS**, S3-01 … S3-35. **Opens with a blocking block, S3-01 … S3-09 plus S3-30 … S3-35**, closing the nine defects the user's hand pass found (**K6–K14**, ruled by **D18–D26**): the height chain, the `shrink-0` overflow, the window minimum size, the blur policy, the missing `DragOverlay`, the Cyrillic UI font, the toast flood and the refusal channel — then the six tickets added after the block A review (**K15**'s column height, **K16**'s hand-applied `refuse()`, the `make shots` capture harness, **D32**'s audit corrections, and **K17**'s non-recursive component sweep) and the hand pass that confirms them. **No feature ticket starts until that block is green.** Then: slide-over with Markdown editor/preview, inline subtasks, tags, due, priority, estimate, RRULE editor including the `COUNT`/`UNTIL` end conditions the user's **D27** added, attachments copied into app data dir, editable time log, type switcher; collapsible tree with inline rename, drag-to-reparent, arrow/Enter/Tab keyboard nav; archive view; FTS search with tag/type/status/date filters | **Two halves.** Block A: every defect in **K6–K14** is closed, each with a named mechanical check *or* an honest statement that only an eye can see it, plus the hand pass of **S3-09**. Block B: **every field round-trips through Go; reparent in tree shows on Kanban instantly** |
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

### Stage 2 — CLOSED, PASS

The Reviewer returned **PASS** on the **second** review, verified at commit `8818db4`.

**Twenty-two tickets, S2-01 … S2-22, in `TASKS.md`. All twenty-two are committed**, one
conventional commit each, plus one gap-closing commit — `7af4d1d`, which added
`SetPriority` and wired the palette's priority rows (see *"The one plan correction"*
below) — and round 1's two fixes, `ae6befd` and `4b1af9c`. All 84 commits on `main` are
authored solely by `Ismat <mukhamejanov.ismat@gmail.com>`, with no AI author and no
co-author trailer on any of them.

Re-measured by the Reviewer at `8818db4`:

| | |
|---|---|
| `make check` | green — **all five gates**, including `wails build -tags webkit2_41` |
| `make cover` | `internal/domain` **100.0%**, `internal/service` **93.8%** (bar ≥90%) |
| `make front-test` | **22 files, 247 tests, 0 failures** |
| `make guard` | **six of six** checks pass, with `GUARD_ALLOW_RE` **empty** |

(The 247 supersedes the 251 this document carried before round 1: `ae6befd` deleted the
orphaned tests of the dead TypeScript time-derivations along with the code.)

The Reviewer further confirmed, independently: **zero hex literals** in `frontend/src`;
`design/` byte-identical to its Stage 1 state; no cgo and no `mattn` in the dependency
graph; the Stage 0 and Stage 1 migrations untouched; `TestDomainIsPure` and
`TestDomainReadsNoClock` passing **unmodified**; and that **no test was weakened by the
`CheckHabitToday`/`UncheckHabitToday` binding rename** — two of the changed tests are
strictly stronger than what they replaced.

**Both round-1 blockers were verified closed the hard way**: the Reviewer reverted each
fix individually and watched the specific test go red, then restored it. That is the
negative-control habit this project adopted in Stage 1, applied to a review rather than
to a commit.

**The ACCEPT criterion is MET**, on its mechanical half. The bar is *create → move
through every column → complete, keyboard only, no mouse*, and
`frontend/src/App.accept.test.tsx` demonstrates exactly that: it drives the **assembled
`<App />`** — the real composition root, with the store, the overlays, the strip and the
toast layer mounted — imports no component beneath it, and asserts the **exact sequence
of client calls**: one `CreateNode`, then `MoveToColumn` with each of the four statuses
Go supplied, in order. It presses keys and only keys, and **`make guard` check 6 fails if
a `click`, a `pointer` or a `fireEvent.mouse*` ever enters that file** — and fails too if
the file is deleted or emptied, so the cheapest way past it is closed. The Doing→timer
coupling (**C1**, **D13**) is asserted from the UI in the same flow: the running-timer
indicator appears on Doing and is gone on Done.

**What is still owed is the hand half**, and the PASS did not include it: the **ten-step
hand run on `./build/bin/nexus`** and four other checks that need a display. They are
listed under *"Owed to a hand pass"* below and are **carried into Stage 3**. Nothing in
this document may describe them as done.

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

#### The two-round review history — kept in full

**Stage 2 failed its first review and passed its second.** Both rounds are kept below,
because what round 1 caught is the most useful content in this section: neither blocker
was a missing feature, and both were the *same* defect class that cost Stage 1 three
rounds — a rule with a second spelling.

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

**Review round 2 — PASS**, at `8818db4`. The Reviewer re-checked both round-1 blockers
and, for each, **reverted the fix and watched the specific test fail before restoring
it** — `ae6befd`'s habit-day fix against the store test that drives the toggle at
`23:59:30` and again at `00:00:30`, and `4b1af9c`'s `-iE` against the planted
`view.status === 'Done'` line. A revert that leaves the suite green proves nothing; both
reverts went red. The Reviewer also planted a regression named outside
`overdue|derive|streak|progress|percent` and reproduced the round-1 `wireDate` blind spot
against a green `make guard`, confirming that **reading the diff is what caught it and
that the guard is a heuristic** (**D17**). All five gates green including
`wails build -tags webkit2_41`; coverage **100.0% / 93.8%**; `make front-test`
**22 files, 247 tests, 0 failures**; `make guard` six of six with `GUARD_ALLOW_RE`
empty; zero hex literals; `design/` byte-identical; no cgo; migrations untouched;
`internal/domain` purity tests unmodified and passing; all 84 commits authored solely by
`Ismat <mukhamejanov.ismat@gmail.com>` with no AI trailer; and **no test weakened by the
binding rename** — two are strictly stronger. **Stage 2 is CLOSED.**

The commits of the review cycle, in order: `ae6befd`, `4b1af9c` (round 1's fixes) ·
`8818db4` (the documentation of D17 and the round-1 corrections; the commit the PASS was
verified at). `7af4d1d` predates round 1 and closed the `set priority` planning gap.

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

#### Carried into Stage 3 — obligations, not suggestions

**Stage 2 closed owing the following. None of it may be dropped silently, and the
matching list in `TASKS.md`, "Carried into Stage 3", is where each one becomes a ticket
or an explicit refusal.**

**⚠️ Read this one first. A green `make guard` is NOT a proof, and no status line in this
repository may offer one as evidence that the frontend computes nothing.** `make guard`
checks **3a and 3b are name-based heuristics** over
`overdue|derive|streak|progress|percent`. A derivation named anything else is invisible
to them **by construction** — and this is not hypothetical: **`wireDate` was named after
none of those words and sat behind a green guard for an entire stage**, computing the
habit check day in TypeScript, until a human read the diff. The Reviewer reproduced it at
round 2 with a planted regression that walked straight past a clean guard. A green guard
is evidence that **five names are absent**. **Reading the diff is still the check** — see
**D17**.

**1. Five things owed to a hand pass, none of them verifiable without a display.** They
are enumerated under *"Owed to a hand pass"* above and repeated here so they cannot be
lost between stages: the **ten-step keyboard run on `./build/bin/nexus`** — including the
**three D8 due-badge assertions at steps 3, 4 and 7, which have no mechanical backing
whatsoever**; **no background flash on the first frame in either theme** (**D12**,
**K1**); **Russian not clipping at a real 1024×768**; the **`:focus-visible` ring
actually painted in both palettes**; and **S2-07's *"a window opens and the first
`Board()` returns five columns"***. Stage 2's PASS did **not** cover any of them.

> **AMENDED during Stage 3, and the amendment is a correction of this list's premise.**
> *"None of them verifiable without a display"* was right; *"and there is no display"* was
> **wrong** — `xvfb-run`, `import` and `convert` are installed and always were (**E4**).
> Two of the five are now **discharged by capture**: **Russian not clipping at a real
> 1024×768** (photographed; **K8** is fixed on real paint) and **S2-07's window-opens /
> five-columns**. The other three need **input** or a **first frame** and are structurally
> out of reach here, so they stay on **S3-09**. **D31** builds the capture target and
> states the split; it is deliberately **not** written as closing the hand pass.

**2. K1/D12, K2 and K3, carried through from Stage 1's handoff and unaffected by this
stage.** Stage 2 implemented the decisions (S2-08, S2-06, S2-14), but K1's no-flash
behaviour is item 1's second bullet and is still unwatched; K2 and K3 keep their entries
in §7 as the record of why the code looks the way it does.

**3. D16 / K5 — the Aurora drift gate is implemented and audited both ways, and
*nothing draws a drift*.** `data-drift` is `on` without reduced motion and `off` with it,
on every palette, and **no code reads it**. Nobody has watched a drift pause because
there is none to watch. **No document in this repository may be edited into implying the
visual exists**, and it is **not to be invented**: it returns when a drift specification
does.

**4. C6 — enum membership is spelled a second time, and it is not a defect.** Details
below; the Reviewer ruled it an acceptable judgement call for Stage 2, and Stage 3 picks
between the two recorded fixes.

##### C6 — enum membership is spelled twice, and nothing checks it

`commands.ts` and `AppearanceControls.tsx` take `Object.keys` of
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

**Three things Stage 2 was explicitly forbidden from doing — and did not.** They are the
shape of Stage 1's four review rounds, turned into rules up front, and they bind every
later stage exactly as they bound this one:

1. **No rule may gain a second spelling — in TypeScript any more than in Go.** The
   frontend renders what Go returns. `make guard` (S2-10) greps for the specific
   violations: a status string literal in a component, a recomputed overdue flag, a
   recomputed percentage, a re-derived column eligibility. **It is a grep, not a
   proof** — see the warning at the head of *"Carried into Stage 3"* and **D17**.
2. **No hex literal in `frontend/src`, ever**, and `design/` stays read-only. Confirmed
   at round 2: zero hits, and `design/` byte-identical.
3. **No hard-coded user-visible string**, from the very first component.

**Final review**: fresh clone → `make check` → `wails build -tags webkit2_41` → `install.sh` →
reboot checklist, executed and reported. `QA.md` with 25 manual scenarios covering
every rule in §4. Known gaps reported honestly — **nothing marked done that was not
actually verified.**

### Stage 3 — IN PROGRESS; block A open, block B not started

**Thirty-five tickets, S3-01 … S3-35, in `TASKS.md`, in two blocks.** The split is not
cosmetic: the user ran the app on real hardware after Stage 2 closed, found nine defects,
and said *"if these details are resolved in the next steps, we are good to go."* That is a
precondition, so it is a block and not a backlog.

**Block A — S3-01 … S3-09, blocking.** The defects, in an order chosen so that no ticket
has to reach outside its Scope:

| Ticket | Defect | Issue / ruling |
|---|---|---|
| S3-01 | No height chain; the *document* scrolls instead of the board, and every `.focus()` scrolls its ancestors | **K10** / **D22** |
| S3-02 | `shrink-0` on max-content localised strings overflows the column — **it overflows in English too** | **K8** / **D20** |
| S3-03 | The window has no minimum size and the user changes display scaling often | **K9** / **D21** |
| S3-04 | ~50 simultaneous `backdrop-filter` surfaces; resizing breaks the layout until restart | **K7** / **D19** |
| S3-05 | The dragged card vanishes on grab — there is no `DragOverlay` | **K6** / **D18** |
| S3-06 | Neither UI font contains a single Cyrillic glyph; `<html lang>` is frozen at `en` | **K11** / **D23** (the user's) |
| S3-07 | Error toasts stack without limit and cover the board | **K12** / **D24** |
| S3-08 | A domain *refusal* is rendered as "something went wrong", and nothing says **what** failed | **K12** / **D25** |
| S3-30 | `refuse()` is applied by hand 26 times and nothing forces a 23rd bound method to call it | **K16** / **D30** |
| S3-31 | The columns hug their content, so an empty column is barely a drop target and nothing scrolls inside one | **K15** / **D29** |
| S3-32 | Capture works and always did; make it a target, with a mandatory blankness check | **E4** / **D31** |
| S3-33 | The shrink audit's own comments state three different, all wrong, corpus sizes, and its non-vacuity guard counts SVGs it cannot judge | **D32** |
| S3-34 | The store sweep is not recursive, sweeps test files, and passes vacuously on an empty set | **D32** |
| S3-35 | The component sweep is not recursive, so **D19**'s resize-workaround ban can go silently partial | **K17** / **D34** |
| S3-09 | The hand pass that confirms all of the above, plus the checks still owed from Stage 2 | — |

**Block A is OPEN.** The Reviewer returned **PASS on the code** for `c32a1c9..b53f1a4`
(S3-01 … S3-05, S3-07, S3-08). **S3-30, S3-31, S3-33 and S3-34 have since shipped and are
green at `c3c954e`** — five gates, `make guard` 8/8 with `GUARD_ALLOW_RE` empty,
`make front-test` **374 tests across 29 files**, `make cover` domain **100.0%** / service
**94.0%**. What remains: **S3-06**, blocked on the user running
`npm install @fontsource/inter`; **S3-32**, the capture harness; **S3-35**, added after
S3-34's own report (**K17**/**D34**); and **S3-09**, the hand pass. **S3-30 and S3-34 were
the Reviewer's two conditions on block B, and S3-35 is a third the PM adds** — block B adds
bound methods, store surface **and components**, which is precisely what all three stop
decaying.

**Why that order.** S3-01 fixes the vertical and the scroll axes; S3-02 then fixes the
horizontal against a settled height chain; **S3-03 comes third because it derives the
window's minimum size from the layout floor, and the floor is not final until S3-01 and
S3-02 have both landed**. S3-04 and S3-05 both rewrite `Column.tsx`; blur goes first
because S3-05's overlay has to work whatever the blur policy turned out to be, and not the
other way round. S3-06 through S3-08 touch none of those files. **Then the five added
tickets, in the order S3-30 → S3-31 → S3-32 → S3-33 → S3-34: S3-31 settles the screen
before S3-32 photographs it, and everything else is independent.** S3-09 is a human
at a real keyboard and is still the gate into block B — **S3-32 makes its list shorter, not
empty** (**D31**).

**Block B — S3-10 … S3-29, the feature stage.** Go first (S3-10 … S3-19: the detail read,
the field writers, the type switcher, tags, attachments, the editable time log, search
filters, the archive list, the enum-set publication that closes **C6**, and **S3-19, the
`COUNT`/`UNTIL` widening of the recurrence parser** that the user's **OQ2** answer
requires), then the
frontend (S3-20 … S3-28: the slide-over, the field editors, inline subtasks, the
recurrence editor, attachments and the time log with the running clock, the tree view, the
search screen, the archive view, and the latent non-Latin-keyboard defect **K13**).
**S3-29 is `README.md`**, and it is **no longer blocked**: it was marked *blocked on the
user running `sudo apt install xvfb imagemagick`*, and **both are installed and always
were** (**E4**). Its four-way screenshot grid now comes out of **S3-32**'s `make shots`
rather than out of a hand capture, and S3-29 is re-pointed at that target (**D31**).

**Both open questions have been answered by the user and are CLOSED.** They are kept in
full at the end of §7, marked closed with the answer, because the reasoning is worth
keeping. **OQ1 → Inter**, one face under both family names, which unblocks **S3-06** and
fixes **D23**'s open half. **OQ2 → option B**, the middle option and *not* the
recommendation: the parser gains **`COUNT`** and **`UNTIL`**, so the editor may offer
"repeat N times" and "repeat until date". That is **D27** (the user's) and **D28** (the PM
ruling that spells out its consequences), and it is why block B gained a ticket: **S3-19**
does the domain widening and **S3-23 is blocked on it**. `BYSETPOS`, `BYMONTH` and ordinal
weekdays stay **rejected at parse time**.

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

Referenced as **D1–D34** and **E1–E4** from tickets in `TASKS.md`. **Every question is
closed.** The two that were open — **OQ1** and **OQ2** — were put to the user rather than
answered silently, because a PM who answers them silently is the defect this project keeps
finding; they are **kept at the end of this section, marked CLOSED with the answer**, not
deleted, because the reasoning is worth keeping and a decision without its rejected
alternatives is half a record.

**D1–D12 were given by the user and are authoritative** — they override anything
earlier in this document that contradicts them. **D1–D7** were settled before Stage 0.
**D8**, **D9**, **D10** and **D11** were confirmed by the user *during* Stage 1, when
implementation exposed questions the earlier decisions did not answer; they carry the
same authority. D10 and D11 came out of the second review, and **D10 generalises the
derivation rule in §4**, which has been amended accordingly. **D12** was confirmed by
the user while Stage 2 was being planned and closes **K1**.

**D23 was given by the user** during the Stage 3 planning pass, after they ran the app by
hand: vendor a Cyrillic-capable face and register it under the **same two family names**
via `unicode-range`, so `design/` is not edited and Latin keeps its designed typeface. It
carries the same authority as D1–D12. **Its one open half — *which* face — is now also the
user's and is Inter** (OQ1, closed). The user also supplied the **discriminating
experiment** behind **K7** — switching to Studio, whose `--blur` is `0px`, repaired a
stuck layout live with no restart — which is evidence, not a decision, and is recorded
under K7. **D19 was put back to the user for confirmation and the user confirmed it as
written.**

**D27 was given by the user**, answering **OQ2**, and it carries the same authority as
D1–D12: the recurrence language gains **`COUNT`** and **`UNTIL`**. The user chose the
**middle** option over the PM's recommendation, which was to keep the editor inside the
existing subset. **D28 is the PM ruling that spells out what that costs** — the parse
boundary, the purity constraint, and the **D5** interaction for a series that has ended —
and it is overturnable in the ordinary way, while **D27 itself is not the PM's to revisit**.

**D13, D14, D15, D16, D17 and D18–D22, D24, D25, D26, D28 are PM rulings.** D13, D14 and D15
were made during Stage 2
planning because `TASKS.md` **C1**, **C4** and **C5** demanded a decision and the user's
brief did not contain one; **D16** was made *during* Stage 2, when the Dev asked what
draws Aurora's background drift and correctly declined to invent it; **D17** was made
*after* Stage 2 was implemented, when the Reviewer's first round found the habit check
day being computed in TypeScript on a decision that had been *"made in a code comment"*
and never written down. **D18–D22 and D24–D26** were made while Stage 3 was being planned,
on the nine defects the user's hand pass found: each one is a rule that has to be written
somewhere, and a rule written only in a component is the defect class this project has
paid for five times. They are written in
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

### D18 — the dragged card is drawn in a `DragOverlay`, and the source item is hidden (PM ruling, Stage 3 planning; closes K6)

**The card the user is dragging is rendered once, in a `<DragOverlay>` mounted as a direct
child of `DndContext`. The card's own `<li>` stays in the list, keeps its space, and is
hidden while `isDragging`.**

- **Why an overlay and not a fix to the in-place translation.** Without an overlay,
  `@dnd-kit/sortable` sets `useDragOverlay = false` and translates the source item *inside
  its own column*, which produces two independent failures at once. The neighbouring
  column paints over it, because `backdrop-filter` creates a stacking context and the
  column's `z-10` can only order siblings within one column; and each column is its own
  `SortableContext`, so the moment the pointer crosses into another column `overIndex`
  goes to `-1` in the source context, the transform becomes `null`, and **the card
  teleports home and stops following the pointer**. An overlay is not a workaround for
  either of those — it is the mode the library documents for cross-container dragging, and
  it removes both causes rather than fighting them.
- **`draggingNodeId` finally gets a consumer.** `store/ui.ts` has held it since S2-17 and
  **nothing anywhere reads it**. The overlay reads it. A piece of state with no reader is a
  second spelling waiting to happen; this ruling either uses it or it should have been
  deleted.
- **One card component, not two.** The overlay renders the **existing `<Card>`**. A
  separate "drag preview" component would be a second rendering of the card, which is the
  defect class by another name — and it would drift the first time a chip is added.
- **Reduced motion routes through the existing helper.** `dropAnimation={null}` when
  `prefersReducedMotion()` is true, asked of the same `lib/appearance.ts` helper
  `Column.tsx` already calls. Not a second media query.
- **Rejected alternative: raise the source card's `z-index` and drop the blur.** It cannot
  work. Even if the stacking context were gone, the `SortableContext` boundary still makes
  the transform `null` across columns, so the card would still stop following the pointer —
  and it would trade a visible bug for an invisible one.
- **Accepted consequence:** the dragged card is painted outside the column's overflow and
  outside its surface, so it is *unblurred* while in flight even under Aurora. That is what
  a lifted object should look like, and it is a change a user will notice.

### D19 — blur the few large surfaces, not every small one (PM ruling, Stage 3 planning; **CONFIRMED BY THE USER**; closes K7)

> **The user was asked to confirm this one and did, as written.** Blur stays on the five
> columns and the two overlay scrims; the card, the habit chip and the toast lose it.
> Nothing in the ruling changed — the confirmation is recorded because the ruling gives
> away a documented property of a read-only design handoff, and a PM ruling was the wrong
> authority to do that on alone.
>
> **The escalation ladder below survives the confirmation and is not shortened by it.** If
> S3-09's hand pass still reproduces the stuck layout at seven blurred layers, step 2 (drop
> the blur from the columns too, leaving the two overlays) is still a PM step — but **step
> 3, that Aurora's `--blur` is unusable on this WebKitGTK, remains a USER decision**,
> because at that point the palette's specified appearance is what is being given up.
> Confirming step 1 did not pre-authorise step 3.

**`backdrop-blur-glass` is applied to the five column surfaces and the two overlay
scrims — and to nothing else.** It is removed from the card, the habit chip and the toast.
Those three keep their translucent `bg-surface`/`bg-elevated` tokens, so Aurora stays
translucent; they simply stop each creating a compositing layer.

The full allow-list, and it is meant to be exhaustive:

| Keeps `backdrop-blur-glass` | Loses it |
|---|---|
| `components/Column.tsx` — 5 instances, large | `components/Card.tsx` — one per card, unbounded |
| `components/QuickAdd.tsx` — the full-screen scrim, at most 1 | `components/HabitChip.tsx` — one per habit |
| `components/CommandPalette.tsx` — the full-screen scrim, at most 1 | `components/Toast.tsx` — one per toast |

- **The tension, stated rather than glossed.** `design/README.md` specifies Aurora's
  surfaces as translucent **with backdrop blur**, and `design/` is read-only. This ruling
  does not edit it and does not contradict its *semantics*: every surface is still
  translucent, and the blur is still what separates Aurora from Studio. What changes is
  **how many elements** carry the effect — a count the design handoff never states,
  because a handoff specifies a look and not a compositing budget.
- **The evidence is the user's, and it is a discriminating experiment rather than a
  theory.** With the layout stuck broken after a resize, **switching the palette to Studio
  — whose `--blur` is `0px` — repaired it live, with no restart.** That is WebKitGTK
  compositing-layer staleness and it eliminates the competing scrollbar-hysteresis
  explanation, which no palette switch could have touched.
- **Rejected alternative (a): drop `backdrop-filter` entirely.** It is the one property
  that makes Aurora Aurora, and `design/` is explicit about it. Removing it to fix a
  compositing bug is deciding a design question with an engineering hammer.
- **Rejected alternative (b): keep all ~50 and force a repaint on resize.** A
  `requestAnimationFrame` nudge, a forced reflow, a transform toggle — every one of them is
  a workaround for a symptom whose cause we have already identified, and each is a second
  rule about resizing that nothing tests.
- **Rejected alternative (c): make the blur conditional on element count.** A rule that
  reads "blur cards when there are fewer than N" is a rule nobody can see the boundary of.
- **Accepted consequence, and the escalation ladder if it is not enough.** Seven blurred
  surfaces may still be too many for this WebKitGTK build. **This cannot be proven on this
  machine** — a resize cannot be performed here at all, because there is **no window
manager** (**E4**), so S3-04's real check is S3-09's hand pass. If the hand
  pass still reproduces the stuck layout: step 2 is to drop the blur from the column as
  well, leaving it on the two overlays only; step 3 is that Aurora's `--blur` is unusable
  on this platform, which is a **user decision** and not a PM one, because at that point
  the palette's specified appearance is what is being given up.

### D20 — no localised string may refuse to shrink (PM ruling, Stage 3 planning; closes K8)

**An element whose text comes from i18n, from `formatDate`, or from `formatNumber` may not
carry `shrink-0`, a fixed width, `whitespace-nowrap` or `truncate`.** `shrink-0` stays
legal on a **fixed-size non-text box** — an icon, the timer dot, the progress track —
because those have a size that does not depend on the locale.

- **The failure it closes.** `html { font-size: 13px }` makes every Tailwind rem 13/16 of
  nominal, so `min-w-36` is **117px, not 144px** and `gap-2`/`p-2` are 6.5px. At the floor
  the usable card interior is ~87px, and `DueBadge`'s `shrink-0 font-mono` takes its
  max-content width: ~94px in English, ~125px in Russian. **It overflows in English.** The
  user's screenshot shows the consequence one level up: the `shrink-0` card-count span
  ("0 карточек", ~86px) starves the column heading down to ~16px, so it wraps **one
  character per line** — "Бэ / кл / ог".
- **This is D17 in a second place, and that is the part that generalises.** `Card.tsx`
  states in a comment that *"nothing on this card has a fixed width and nothing is
  `whitespace-nowrap`"*, and `App.accept.test.tsx` asserts the absence of `truncate` and
  `whitespace-nowrap`. **Both are true, and the card clipped anyway**, because `shrink-0`
  on a max-content localised string clips identically and was not on the list. A
  mechanism-based audit is only as good as its enumeration of mechanisms, and an
  enumeration is a guess about the future.
- **So the audit changes shape, not just its word list.** Adding `shrink-0` to the grep
  would fix this bug and leave the next one. The audit becomes a **walk over the rendered
  Russian DOM** that asks, of every element carrying localised text, whether any
  shrink-refusing utility is on it — so a *new* utility with the same effect is caught by
  the same test the day it is used.
- **Rejected alternative: let the column grow and let the board scroll.** The board already
  scrolls horizontally, so this looks free. It is not: it makes five columns unreachable
  without scrolling at the default window size, and it converts a text-fitting bug into a
  navigation one.
- **Accepted consequence:** a long localised date can now wrap inside a badge. A date on
  two lines is ugly; a date sliced off at the column edge is unreadable, and the user has
  a screenshot of the second one.
- **What this ruling cannot promise.** jsdom has **no layout engine**, so no test in this
  repository will ever assert "it does not clip". The audit asserts mechanisms. **The
  absence of clipping is S3-09's hand pass, at a real 1024×768 and in Russian**, and the
  estimated onset widths — RU heading ~784px, RU due badge ~814px, EN due badge ~659px,
  board floor 624px — are **±5% arithmetic, not measurements**.

### D21 — the window's minimum size is derived from the layout floor, never typed (PM ruling, Stage 3 planning; closes K9)

**`main.go` gains `MinWidth`/`MinHeight`, and the numbers are computed from named inputs
that are each written down exactly once. A literal `MinWidth: 640` is refused.**

- **The failure it closes.** `main.go` sets `Width: 1024, Height: 768` and no minimum, and
  Wails calls `SetMinSize(0, 0)` unconditionally, so GTK is hinted with
  `min_width = min_height = 0` and the window can be dragged ~380px below the layout's own
  ~624px floor. **The user changes display scaling often and it varies**, so the startup
  window is not reliably 1024 CSS px and may begin near or below the floor.
- **The rule this collides with is this project's hardest one.** A minimum width in Go is
  the CSS floor written down a second time, and *"a rule spelled twice is two rules"* has
  cost this project four review rounds. The containment is the same shape as **D12**'s:
  D12 forbade a hex literal in `main.go` and made the colour be *derived from
  `design/tokens.css`*, which Go already parses. The floor gets the same treatment.
- **The inputs, each with exactly one home:** the root font size (`frontend/src/style.css`,
  `html { font-size: … }`), Tailwind's spacing unit, the column's `min-w-*` unit count
  (`components/Column.tsx`), the board `gap-*` and the shell `p-*` unit counts, and the
  **column count, which must come from the set Go already publishes** — not from a `5`
  typed in `main.go`.
- **The criterion is drift, not correctness.** Changing `min-w-36` to `min-w-32`, or the
  root font size, must make a **named test fail** unless the Go-side floor changes in the
  same commit. Demonstrated by negative control, as Stage 1 and Stage 2 both required.
- **Rejected alternative: pick a round number like 800×600 and document it.** It is a
  magic number with a paragraph next to it, which is a magic number. It also goes stale
  silently the first time a column's minimum width changes — and nothing would notice.
- **Accepted consequence:** the vertical floor is softer than the horizontal one. The
  board's own minimum height is one column header plus one card, and a card's height
  depends on how many chips it carries and how the title wraps — which is layout, and
  layout is exactly what cannot be computed here. **`MinHeight` is therefore a stated
  composition of named terms, and the ticket must say which of its terms are estimates.**
  A term that cannot be derived is disclosed, not rounded up quietly.

### D22 — one height chain; the board scrolls, the document does not; focus never scrolls (PM ruling, Stage 3 planning; closes K10)

**Three rules, one ruling, because they are three symptoms of one missing decision about
who owns the scroll.**

1. **The shell is a full-height chain.** `html`, `body` and `#root` are full height; the
   shell fills it; `<main>` carries `min-h-0` so it may actually shrink. Today there is
   **no `#root` rule at all**, `body` has only `min-height: 100vh`, and `<main>` has
   `min-w-0` — which is a **no-op**, because the shell is a *column* flex container and the
   axis that needs releasing is the block axis. So `<main>`'s `min-height: auto` floors it
   at min-content, the shell grows past the viewport, and the **document** scrolls. The
   user's screenshot shows the document-level vertical scrollbar.
2. **The board owns both scroll axes, explicitly.** `Kanban.tsx` sets `overflow-x-auto`
   and nothing else, and per CSS Overflow 3 an `overflow-x` of `auto` against an
   `overflow-y` of `visible` **promotes `overflow-y` to `auto`**. So the board already
   clips and scrolls vertically, silently, contradicting its own comment. The axes are to
   be stated rather than inherited from a spec rule nobody reads.
3. **No `.focus()` call anywhere may scroll its ancestors.** Every one of the eleven
   `.focus()` calls in `frontend/src` omits `{ preventScroll: true }`, so each scrolls
   every scrollable ancestor including the board. **The rule gets one spelling**: a single
   focus helper, and a `make guard` check that `.focus(` appears in exactly one module. A
   rule applied eleven times by hand is a rule that will be applied ten times after the
   next ticket.

- **Ruled out, with the reason, so nobody rediscovers it: the "latched `scrollLeft`"
  theory.** It was investigated and is **wrong**. The columns are `flex-1 basis-0`, so
  `scrollWidth` tracks `clientWidth`, the user agent clamps the scroll offset on resize,
  and nothing in `frontend/src` ever reads or writes `scrollLeft`/`scrollTop` — grepped.
  A stale horizontal offset is not what the user saw.
- **Rejected alternative: set `overflow: hidden` on `body` and stop there.** It hides the
  document scrollbar without giving the board a height to scroll inside, so the bottom of
  the board becomes unreachable instead of scrollable. It is the fix that makes the bug
  quieter.
- **Accepted consequence:** the page no longer scrolls as a document, so anything the shell
  contains that does not fit must have its own scroll container. Today only the board
  qualifies; every later region — the detail slide-over, the tree, the archive list — must
  say which one it is.

### D23 — vendor **Inter** under the same two family names (THE USER'S, Stage 3 planning; closes K11)

**A Cyrillic-capable face is vendored and registered under the *same two family names* the
palettes already use, scoped with `unicode-range`, so Latin keeps its designed typeface and
`design/` is not edited.** Suggested range: `U+0301, U+0400-045F, U+0490-0491, U+04B0-04B1,
U+2116`.

- **The failure it closes.** `design/tokens.css` sets Aurora's `--font-ui` to
  `'Space Grotesk'` and Studio's to `'Figtree'`. Verified from
  `node_modules/@fontsource/*/unicode.json` **and from the shipped `frontend/dist/assets/`
  the user is actually running**: Space Grotesk covers `[vietnamese, latin-ext, latin]` and
  Figtree covers `[latin-ext, latin]`. **Neither contains a single Cyrillic glyph.** Only
  JetBrains Mono does. **81 of the 82 leaves in `ru.json` are Cyrillic**, so switching to
  Russian drops the entire UI out of its designed typeface into the system `sans-serif`
  while every number and date stays in JetBrains Mono — two unrelated typefaces in one
  header, different metrics, and a `flex-wrap` header very plausibly gaining a second row.
  **This is the user's reported "theme and colour buttons break when I switch language".**
- **Why `unicode-range` and not a family-list fallback.** A fallback list would work, but it
  would have to be written into `--font-ui`, which lives in `design/tokens.css`, which is
  **read-only**. Registering extra `@font-face` rules under the same family names in
  `frontend/src/style.css` adds coverage to a family without editing the declaration that
  names it — which is precisely what `unicode-range` is for.
- **Local-only is not negotiable here either.** The face is vendored through `@fontsource`
  and bundled from `node_modules`, exactly as the three existing faces are. **No CDN, no
  `<link>`, no runtime fetch.** A new npm dependency must be called out on its ticket.
- **Pair it with a mechanical check, or this recurs on the next locale.** A test asserting
  that **every codepoint used in `ru.json` is covered by some bundled `unicode-range` for
  `--font-ui`** — over the ranges actually bundled, not a range retyped into the test.
- **The face is Inter, and that is the user's choice too** (**OQ1**, closed). This bullet
  used to read *"which face is not decided here"*, and it was right not to decide it: the
  candidates were Inter and Manrope, and choosing silently would have been inventing a
  design decision and attributing it to the handoff, which is what **D6 as amended**
  forbids. **One** face, under **both** family names — so one `unicode-range` block and
  one coverage test, and Studio does not get a worse match than Aurora or the reverse.
- **The name is not the evidence.** Inter's reputation for wide Cyrillic coverage is not a
  substitute for reading `node_modules/@fontsource/inter/unicode.json`, which is exactly
  how Space Grotesk and Figtree were found to have none. The Dev verifies the real
  coverage — **including `U+2116` (№) and the combining acute `U+0301`** — before
  committing, and if a codepoint `ru.json` uses is not covered, **stops and reports**
  rather than narrowing the test to match the font.
- **Related and confirmed, and settled by the same ruling.** `frontend/index.html` hard-codes
  `<html lang="en">` and **nothing ever updates it** — grepped, zero writers. `lib/appearance.ts`
  already receives the whole `SettingsView`, already runs once at boot and once per settings
  write, and is already the one module that writes to `documentElement`. **`root.lang` is
  written there**: one site, one spelling. The static `lang="en"` in `index.html` stays as
  the pre-boot default, exactly as `data-palette="aurora"` does.

### D24 — toasts are capped, de-duplicated and self-dismissing (PM ruling, Stage 3 planning; part of K12)

**The toast list holds at most three toasts; an identical consecutive `messageKey`
increments a count on the existing toast instead of appending; every toast dismisses itself
after a bounded interval, and the interval is paused while the toast has focus or the
pointer is over it.**

- **The failure it closes.** `store/toast.ts` appends unconditionally — **no cap, no
  de-duplication, no auto-dismiss**. The user's screenshot shows three identical toasts
  filling the lower half of the window and covering the board.
- **Why a count and not silent suppression.** Three failures are not one failure. Collapsing
  them without saying so would hide a repeat, and a repeat is the most useful thing about the
  second one. The count is a number, so it is rendered in `font-mono` like every other
  number, and pluralised through i18next — Russian has three plural forms.
- **Why auto-dismiss at all, given the "no silent failure" rule.** The rule is that a failure
  must be *surfaced*, not that it must be *permanent*. The raw cause already goes to the
  console and stays there. A toast that never leaves converts one failure into a permanently
  smaller window.
- **The design intent at `store/toast.ts:5-12` is preserved exactly**: a toast carries an
  **i18n key, never a sentence**, and the raw Go error goes to the console only. Nothing in
  this ruling puts text in the store.
- **Rejected alternative: a cap with no de-duplication.** Three identical toasts capped at
  three is still three identical toasts. The cap bounds the damage; the de-duplication is
  what makes the list informative.
- **Accepted consequence:** a burst of more than three *distinct* failures loses the oldest
  from the screen. It is still in the console, which is where the detail has always lived.

### D25 — a refusal is not a failure, and every toast says what failed (PM ruling, Stage 3 planning; part of K12)

**A rule that correctly refuses an action gets its own message, distinct from "something
went wrong", and every toast — refusal or failure — names the operation the user attempted.**

- **The failure it closes.** Every rejection from Go raises the single key
  `toast.error.body` — *"Nexus could not finish that. Nothing was changed."* So **D9**
  refusing to put a project into `doing` (a rule working exactly as specified) is rendered
  as a malfunction, and the user has no way to tell which of three stacked toasts came from
  which action. **After the fact, we still do not know which three operations failed in the
  user's screenshot** — that is the diagnostic half, and it is why "name the operation" is
  part of this ruling and not a nicety.
- **The refusal code is Go's, and it has one spelling.** Go already owns a clean inventory
  of sentinels — `ErrProjectNeverDoing`, `ErrTypeHasNoColumn`, `ErrTypeHasNoDue`,
  `ErrTypeHasNoChildren`, `ErrCircularParent`, `ErrNoRecurrence`,
  `ErrUnsupportedRecurrence`, `ErrTimerNotAllowed`, `ErrNodeArchived`, `ErrNodeDone`,
  `ErrNotAHabit`, `ErrInvalidSetting`. **One table maps sentinel → stable code**, in Go, and
  nothing else classifies anything. An error that is in no table is a **failure**, not a
  refusal, and keeps today's key.
- **The frontend classifies nothing and parses no prose.** It maps a code to an i18n key.
  That is a **label table**, the same shape the five Kanban columns already have and the
  shape **D26** moves the other enum sets to — Go publishes the set, the locale file
  supplies the words. **The frontend must never match on Go's English error text**; a
  message is not an API.
- **The mapping must be exhaustive or red.** A code Go can emit with no key in
  `en.json`/`ru.json` fails a test. Otherwise the next sentinel ships as a blank toast.
- **Rejected alternative (a): let each call site choose its own key.** That is twenty
  try/catch blocks, nineteen of which are right — the exact reasoning `store/call.ts`
  already gives for existing in the first place.
- **Rejected alternative (b): have Go return the user-visible sentence.** Rejected for the
  same reason as **D15**'s option (c): a string returned from Go cannot be translated by the
  i18n layer, and Nexus ships in two languages.
- **Accepted consequence:** a refusal is no longer styled as an error. It is information —
  the rule announcing itself — and the user learns the rule instead of concluding the app is
  broken.

### D26 — Go publishes every enum set; the locale files supply labels only (PM ruling, Stage 3 planning; closes C6)

**Of C6's two recorded options, Stage 3 takes the first: bind a set-publishing method for
each set — statuses, node types, themes, palettes, accents, priorities — so the frontend
enumerates what Go enumerates and the locale file is reduced to labels.**

- **Why the first option and not the parity test.** The preferred fix was already named as
  preferred in `TASKS.md`, for the right reason: it **removes** the second spelling instead
  of detecting it. The board's five columns are the worked example inside this repository —
  the frontend takes the set from Go and `en.json` only names it — and C6 is the request to
  generalise the shape that already works.
- **An honest correction to C6's own text.** A fresh audit found current key parity
  **clean**: 82 RU leaves against 80 EN, the two extra being `board.column.cardCount_few`
  and `_many`, the CLDR plural forms Russian requires, and `locales.test.ts` already
  enforces parity. **C6 is a drift risk for future enum values, not a present defect.** It
  is still worth closing, on exactly the stated criterion: **adding a value to a Go set must
  turn something red** until the frontend offers it.
- **This is the same shape as D25's refusal codes**, deliberately. One mechanism, used
  twice, rather than two.

### D27 — `COUNT` and `UNTIL` enter the recurrence language (THE USER'S, closes OQ2)

**The recurrence parser is widened to accept `COUNT` and `UNTIL`, so the editor may offer
"repeat N times" and "repeat until date". `BYSETPOS`, `BYMONTH` and ordinal weekdays such
as "the 2nd Monday" stay rejected at parse time with `ErrUnsupportedRecurrence`.**

This is **OQ2's option B**, the middle one. **The PM recommended option A** — an editor
confined to the subset the parser already accepts — and **the user chose otherwise.** That
is recorded plainly because the recommendation is still in this document a few screens
below, and a reader who finds it must not mistake it for the decision.

- **What it buys.** *"Every weekday until 31 December"* and *"ten times and then stop"* are
  the two shapes a real habit reaches for and the two the brief's *"custom RRULE"* most
  plainly promises. Both **terminate a series**, which is why they were named together in
  OQ2 as the pair that carries real user value.
- **What it does not buy, and the boundary has exactly one spelling.** `BYSETPOS`,
  `BYMONTH`, `BYWEEKNO`, `BYYEARDAY`, `BYHOUR` and ordinal `BYDAY` remain **rejected at
  parse time**, in `ParseRecurrence`, refused at the point the rule is written rather than
  approximated at the point it is read — the reasoning `recurrence.go` already gives for
  `ErrUnsupportedRecurrence` being deliberately loud. The rejection boundary is asserted in
  **one place**, `TestParseRecurrenceRejectsWhatItCannotExpand`, and it narrows by exactly
  two rows.
- **Still no RRULE library** (OQ2's option C, unchanged). Two bounded fields are not a
  reason to put a dependency into the one package that is pure and at 100% coverage, and
  `recurrence.go`'s header already records why no maintained pure-Go expander is usable
  here: a hidden clock read when `DTSTART` is unset, and expansion to zoned instants when
  `habit_checks` is keyed by a date.
- **`ErrUnsupportedRecurrence` stays unreachable from the UI.** The editor offers what Go
  accepts and cannot compose what Go rejects — that criterion survives the widening intact,
  it simply now has two more fields inside it.
- **This is a domain change and it gets a domain ticket.** **S3-19** widens
  `internal/domain`; **S3-23**, the recurrence editor, is **blocked on it** and may not
  widen the parser itself. A domain change smuggled into a UI ticket is the thing the split
  exists to prevent.

### D28 — a bounded recurrence ends, and an ended series freezes the streak (PM ruling, implementing D27 against D5)

**D27 says which parts are accepted. This says what they mean, because a terminating series
changes what "a scheduled occurrence passed unchecked" refers to, and D5 is a rule that
must not acquire a second spelling.**

**1. The bound lives in the rule, never in the clock.**
`internal/domain` is pure: no `time.Now()`, and anything time-dependent takes an injected
`now func() time.Time`. `COUNT` is positional and `UNTIL` is an absolute date, so **both
are decidable from the rule and the candidate date alone**. `Matches`, `Expand`, `Next` and
`Previous` each respect the bound with no reference to today. **`UNTIL` is compared against
the occurrence being tested, never against the current date** — that comparison is what
`Today(now)` is for, and it belongs to the callers that already take a clock.

**2. `COUNT=n` means the first `n` occurrences from `DTSTART` inclusive**, with `n` in
`1..1000`. The cap exists for the same reason `maxWindowDays` and `maxSearchDays` do: a
positional bound has to be walked, so something must stop the walk. `COUNT=0` and a
non-numeric `COUNT` stay unsupported.

**3. `UNTIL` is accepted only in the `YYYYMMDD` DATE form, and is inclusive.** The
`DATE-TIME` form (`UNTIL=20261231T000000Z`) **stays rejected**. Nexus has no time of day
anywhere — `habit_checks` is keyed by a `Date`, `domain.Date` is a calendar day, and D17
already settled that the calendar day is Go's to name. Accepting a `Z` instant would
require a timezone rule this project does not have, and silently truncating one would be a
second, invisible rule about what a habit's day is. Pleasingly, the existing rejection test
row for `UNTIL` is already the `DATE-TIME` form, so **it keeps passing unchanged**.

**4. `COUNT` and `UNTIL` may not both appear.** RFC 5545 forbids it, and two bounds on one
series is two spellings of one rule. Rejected at parse time, and the editor cannot compose
it.

**5. The D5 consequence, and it is a derivation rather than a new rule.** **D5 is not
amended.** A streak is consecutive **scheduled** occurrences that were checked, and it
**breaks when a scheduled occurrence passes unchecked**. After the bound there are **no
scheduled occurrences at all**, therefore none can pass unchecked, therefore **nothing can
break the streak: it freezes at its final value and stays there.**

- A series whose **final** occurrence was checked keeps that streak forever. It is a
  finished habit with a finished streak, and that is the honest reading of D5.
- A series whose **final** occurrence passed unchecked has a streak of **0**, by the same
  sentence of D5 and with no special case.
- **It must fall out of bounding `Previous` and `Matches`, not out of new code in
  `streak.go`.** The backwards walk already starts above today and steps through
  `Previous`; once `Previous` respects the bound the walk is correct with no further
  change. **If an implementation needs a second code path for "the series ended", the bound
  is in the wrong place** — that is the review signal, and it is the same signal this
  project has now paid for five times.

**6. "Scheduled today" needs no new rule; "ended" needs one field.** `ScheduledToday`
already goes `false` past the bound, so the due-today flag is correct for free and is
**not** duplicated. But *"not today, Tuesday"* and *"never again"* are different facts, and
a strip that had to tell them apart would have to compare `UNTIL` to today **in
TypeScript** — which is `wireDate` (**D17**) with a different name and behind the same
permanently green guard. So **`HabitView` gains exactly one Go-computed field, `Ended`**,
and the strip renders a localised marker from it. This is **D15**'s shape: where a value is
in a terminal or undefined state, Go says so and the UI names it, rather than drawing
something misleading.

**7. A finished habit stays in the strip**, marked finished, still showing the streak it
ended on. It leaves by being **archived**, which already exists and is reversible. Hiding
it automatically was rejected for **D15**'s reason: a row that silently disappears is
indistinguishable from data loss, and the accumulation is at least visible and undoable.

**8. No new error sentinel, and D25's table does not grow.** A check written on a date the
rule does not schedule is **already** specified as harmless — `streak.go` says so in its
own doc comment: it is stored, contributes nothing to the streak and repairs no break. An
ended series is simply *"every later date is unscheduled"*, so `Check`/`CheckToday` keep
their behaviour unchanged.

- **Rejected alternative: a new `ErrRecurrenceEnded` refusal.** It would be one more
  sentinel in **D25**'s exhaustive-or-red table, and it would contradict a semantics
  `streak.go` has already written down. Whether the strip *offers* the tick on a finished
  habit is **presentation**, which `HabitView`'s own doc comment already assigns to the
  strip.
- **Rejected alternative: zero the streak when the series ends.** It reads as punishment for
  finishing, it is not what D5 says, and it would delete the one number that records the
  habit succeeded.
- **Rejected alternative: drop ended habits from `Strip()`.** See point 7.

### D29 — a Kanban column is a full-height drop target, and it scrolls inside itself (PM ruling, after the block A review; closes K15)

**The board's columns hug their content.** `views/Kanban.tsx:399` is
`flex h-full min-w-0 items-start gap-2 overflow-x-auto overflow-y-auto`, and `items-start`
overrides the default `stretch`, so each column is as tall as its cards. A 1024×768 capture
of the running binary shows it plainly: four of the five columns are short boxes sitting at
the top of a large empty area.

**The Reviewer logged this as non-blocking observation 8 and called it "purely visual". It
is not.** `components/Column.tsx`'s `useDroppable` exists precisely to catch a card dropped
on an **empty column** or on the **padding below the last card**. A column that ends where
its cards end has almost no such area — an empty column is a header-high strip, and the
large region below it belongs to the board, not to any column. That is a **drag-and-drop
defect wearing a layout bug's clothes**, and it lands on the surface **S3-05** just rebuilt.

**The ruling, in four parts.**

1. **Columns stretch to the board's height.** The board row goes back to the flex default
   `stretch` (or an explicit `items-stretch`), so every column occupies the full height the
   board has, empty or not. This is what makes `useDroppable`'s rectangle the whole column.

2. **`items-start` was not load-bearing and is not to be replaced by a height literal.**
   It is doing nothing that any ticket asked for — no decision in this file names it. It is
   removed, not neutralised with a `min-h-[…]`; a typed height would be **D21**'s defect in
   a second place (the layout floor is derived, never typed) and a second spelling of the
   height chain.

3. **The column owns its own vertical scroll; the board keeps the horizontal.** This is the
   half **D22** left unfinished. Today nothing scrolls inside a column, so a column with
   more cards than fit pushes against the board's `overflow-y-auto` — which **D22** only
   tolerated because `overflow-x: auto` had silently promoted it. After this ruling the
   **card list** inside each column is the `overflow-y-auto` element, with `min-h-0` on
   every flex ancestor between it and the board so it can actually shrink; the board keeps
   `overflow-x-auto` for the five columns and stops being the vertical scroller. **The
   document still never scrolls** — that part of D22 is unchanged and its test must stay
   green.

4. **The column header does not scroll away.** The heading and the card count stay put
   while the cards scroll under them, because a column whose heading scrolls out of view
   makes a drag across five columns unnavigable.

**Rejected alternative: leave it and make the board a drop target for "the empty area".**
That invents a second answer to *"which column did this land in"* and puts a coordinate
decision in TypeScript. `useDroppable` per column is already the single spelling; the fix
is to give it the rectangle it was always supposed to have.

**Rejected alternative: a fixed column height in `vh` or `px`.** See part 2. It reintroduces
a typed floor, and it is wrong the moment the habits strip or the header changes height.

**What is mechanical and what is not.** jsdom has no layout engine, so no test here can
observe that a column is 600px tall or that a drop landed in the padding. The mechanical
half is the **class contract** — the board is not `items-start`, the scroll container is
the card list and not the board's vertical axis, the `min-h-0` chain is unbroken — asserted
the way `App.keyboard.test.tsx` already asserts class contracts, plus **S3-32**'s captures,
which show a real paint at a real size. The half that is still an eye is the drag itself:
input cannot be driven here (**E4**).

### D30 — no bound method may refuse without going through `refuse()` (PM ruling, after the block A review; closes K16)

**S3-08 gave refusals their own code and named the failed operation, and it did so by hand:
`refuse()` is applied 26 times across the 22 bound methods in `app.go`, and nothing
mechanical forces a 23rd method to call it.** The Reviewer verified the current coverage is
complete and ruled the gap **not blocking for block A** — a miss degrades a message, it does
not corrupt data — but ruled that it **must be a named ticket before S3-10 starts**. This is
that ruling; the ticket is **S3-30**.

> **CORRECTED — the figures were "27 refusals across 25 bound methods" and both were
> wrong.** `app.go` has **25** `func (a *App)` declarations, but three of them —
> `startup`, `context` and `onIPCMessage` — are **unexported, are not bound by Wails, and
> do not return `(T, error)`**, so the bound surface is **22**. And `grep -c 'refuse('
> app.go` returns **27** because one of those lines is the **doc comment at `app.go:60`**;
> the call sites number **26**. Nothing is broken by the error — the committed test
> enumerates and passes **22**, and 22 clears **D32**'s floor of 20 — but **S3-30's
> acceptance criterion "All 25 current bound methods are found" was unsatisfiable as
> written**, and is corrected in `TASKS.md`.
>
> **This is recorded rather than quietly overwritten because it is precisely the defect
> class S3-33 exists to fix**: a count stated in prose in several places, restated from
> memory, corrected in one of them. S3-33 found a corpus size written four times with three
> values, none right, in `frontend/src`; this is the same failure one document up, in a
> file the PM owns. The lesson from **D32** applies without change: *a number in prose has
> no enforcement*. The enforcement here is `app_refusal_test.go`'s own enumeration, which is
> derived and not typed — so the only thing that can go stale is the prose, and the prose is
> now dated to this correction.

**The argument is not hypothetical, and it is an internal inconsistency rather than a
worry.** **S3-01**, in the same block, argued that an eleven-times-by-hand rule needed a
guard and built **check 7** for it. **S3-08 then shipped a twenty-six-times-by-hand rule
without one.** Two tickets in one block reached opposite conclusions about the same shape.
Block B is where new bound methods arrive — `NodeDetail`, six field writers, the type
switcher, tags, attachments, the time log, search, the archive, the enum sets — which is
exactly the condition under which a hand-applied rule decays.

**The rule gets a mechanical enforcer, and the enforcer reads the source.** The existing
shape is already in this repository twice: `layout_test.go:45` and `refusal_test.go:151`
read files across the package boundary rather than trusting a convention. A Go test that
parses `app.go`, enumerates every `func (a *App) …(…) (T, error)`, and asserts each body
routes its error through `refuse(` is the same technique, in the same language, with no new
dependency and no new tool.

**It is a Go test, not a ninth `make guard` check.** `make guard` greps `frontend/src`; this
is a Go file and belongs in `go test`, which is **gate 1** and therefore stronger than a
non-gate target. It also keeps the guard's eight checks at eight, and **D17** applies: a
name-based heuristic would not do — the assertion must be structural (every bound method,
enumerated from the source) rather than a grep for a word.

**The failure message must name the method.** A red test that says "some method does not
refuse" costs more than it saves; it says which one, and what to add.

### D31 — capture is a build target, and it discharges static paint only (PM ruling, after the block A review)

**For three stages this project recorded every visual criterion as unverifiable, on a
premise that was false.** `xvfb-run`, `import` and `convert` are installed and always were
(**E4**). A 1024×768 capture of the running binary in Russian has now been taken and read,
and it settled two open items and found one new defect (**K15**). That must stop being
something an orchestrator improvises and become a target anyone can run.

**1. There is a `make shots` target, and it is a non-gate target.** It joins `make cover`,
`make front-test` and `make guard`. **`make check` is still exactly the five gates** —
unchanged in number and definition, as it has been since S0-11.

**2. Two checks are mandatory, and they are the whole reason the target exists rather than
a shell snippet in a report.** The failure mode is that the binary runs healthily, exits 0,
and `:99` has **no window on it at all**, because this shell sets `WAYLAND_DISPLAY` and GTK
opens on the real compositor instead (**E4**). So the target does both:

- **Window presence first** — `xwininfo -root -children | grep nexus`. Cheaper, and
  specific: it distinguishes *"no window was ever created on this display"* from *"a window
  exists but painted blank"*, which have different fixes.
- **Then blankness** — `convert <png> -format "%k" info:` returning `1` means nothing
  painted, and the target **fails** on it.

A harness without both silently certifies blank images, which is the worst outcome available
here: a green capture pipeline producing evidence of nothing.

> **CORRECTED.** This point previously blamed `WEBKIT_DISABLE_DMABUF_RENDERER` and
> `WEBKIT_DISABLE_COMPOSITING_MODE`. **Those do nothing here**; `GDK_BACKEND=x11` is the
> cause and the cure. See **E4** item 1 for the isolated measurement, and note that the
> *check* survives the correction unchanged — it was always right that a blank PNG must be
> red; only the named cause was wrong.

**3. States are selected by seeding settings, never by clicking.** Input cannot be driven
(**E4**). The four palette/theme combinations and the two languages are persisted rows in
SQLite, so the matrix is reachable by writing settings before launch.

**4. The harness never touches the user's database.** `internal/store/db.go:41` already
honours `XDG_DATA_HOME` when it is absolute — that is the single spelling of where the data
lives, and the harness uses it rather than inventing a flag. `make shots` exports
`XDG_DATA_HOME` to a **throwaway directory under `build/`**, seeds settings there, and
`~/.local/share/nexus/nexus.db` is never opened. There is no `sqlite3` CLI on this machine,
so the seeding goes through **this project's own pure-Go driver** — which is correct
regardless, because a second writer to the schema would be a second spelling of it.

**5. What this converts, stated exactly, because overclaiming it is the one failure that
would make it worse than nothing.**

- **It converts, from "owed to the user's hardware" to "observable here":** the static
  paint at a real size, in every palette, theme and language — Russian not clipping, the
  column headings not wrapping one character per line, the due badge and card count
  readable, the habit chip fitting, the header not gaining a row, the UI rendering in the
  designed family rather than the system `sans-serif`, and the columns filling the window
  (**K15**). Anyone can now look, including the Reviewer, on this machine, from a committed
  image — where before only the user could, on their own laptop.
- **It does not convert "an eye" into "a machine".** The only *automated* assertions a PNG
  supports here are that it is not blank and that its dimensions are right. *"Nothing
  clips"* is still a judgement made by looking; what changed is **who can look and how
  cheaply**, not that a computer decides it.
- **It cannot touch anything needing input or a resize** (**E4**): the ten-step
  keyboard-only run, the `:focus-visible` ring (jsdom evaluates `:focus-visible` as false
  for programmatic focus, and a static capture has no focus to paint), the drag following
  the pointer (**K6**), the empty-column drop (**K15**), the no-flash-of-wrong-background
  first frame (**D12**, **K1** — a delayed capture cannot see the first frame), and
  **K7, the resize defect the user actually reported**. Those stay on **S3-09**'s hand
  script, and **S3-09 is still the gate into block B.**

**6. It unblocks `README.md`.** **S3-29** has been marked *BLOCKED on the user running
`sudo apt install xvfb imagemagick`* since Stage 0. Those packages are present; the blocker
was never real. S3-29 is **unblocked**, and its four-way screenshot grid is produced by
`make shots` rather than by hand.

### D32 — a check that checks nothing must be red, not green (PM ruling, after the block A review)

**Four of the block A review's non-blocking observations are one defect.** Recorded once,
here, so they are not re-argued four times:

- **Guard check 8b bypasses `GUARD_ALLOW_RE`, unlike every other check.** That asymmetry is
  **correct and is hereby deliberate**: an allow-list exists to permit a *forbidden* thing
  in a named place, and letting it also switch off a **presence** assertion would let one
  entry delete the only proof that Aurora still has its blur. It is **undocumented**, which
  is the actual finding, and the next ticket to touch the `guard:` recipe carries the
  comment (**S3-32**).
- **`new Set([]).size === 0` passes `"gives every operation a sentence of its own"`.** The
  derivation is non-empty today and the sibling test that proves it fires was confirmed —
  but a property that holds because a set is empty is a green that means nothing. The
  non-emptiness assertion belongs **at the derivation**, one line from it, not one test
  away.
- **`import.meta.glob('../store/*.ts')` is not recursive and sweeps the store's own test
  files.** Harmless today — the store is flat, and the five `*.test.ts` keys are a strict
  subset of the real ones — but a future `src/store/slices/x.ts` would be missed **in both
  directions of the comparison**, which is silence rather than a red. **Block B adds store
  surface**, so the window in which this is harmless is closing.
- **A non-vacuity guard that counts `element.className !== ''` counts SVG elements it
  cannot judge**, because `className` on an `SVGElement` is an `SVGAnimatedString` and never
  equals `''`. Every icon counts toward the `> 20`. Harmless today; weaker than it reads,
  and the thing it guards is the audit that **K8** walked past.

**The rule.** *An enumeration derived from source must fail when it derives nothing, must
not silently under-match, and must count only what it can actually judge.* **Absence of
evidence is a red.** This is **D17**'s lesson one level up: D17 says a green heuristic is
not a proof; D32 says a green **vacuity** is not even a heuristic. Tickets **S3-33** and
**S3-34** implement the parts that need code; the guard-8b half needs a comment and no
behaviour change.

**Not a defect, and deliberately not ticketed: the shrink audit's measured false-positive
surface.** Inverting S3-02's audit to deny-by-default leaves **951 of 11,290** neutral
Tailwind classes reportable — **8.4%**: negated spacing and z utilities, off-palette
colours (which are *intentionally* refused, since rule 9 forbids them anyway), `columns-*`,
`break-*`, `contain-*`, `inline-table`, `not-sr-only`. **Every one of them is loud and none
is silent**, which is the direction D20 chose on purpose: an omission from `IRRELEVANT` is
a false positive, where an omission from the old `DECIDES_*` families was a false negative.
**`not-sr-only` is the one true misclassification** — it sets `width: auto`, which is a
permission, not a refusal — and it gets a `PERMITTED` entry **the day it is first used**,
at which point the audit says so itself. No ticket: the defect reports itself.

### D33 — the board's vertical scroll is a dormant fallback, and a card list may not clip (PM ruling, after S3-31; ratifies two things S3-31 reported rather than decided)

**S3-31 shipped at `e7868c8` and the Dev deliberately did not decide two questions, reporting
them instead.** That was correct — both needed a ruling, not a silent acceptance. Both are
ruled here.

**1. The board still carries `overflow-y-auto`, and that is ACCEPTED.** D29 part 3 and
S3-31's criterion said the board *"stops being the vertical scroller"*. **As a class change
that is not achievable**, and the ticket's own next criterion is why: removing
`overflow-y-auto` is only *spellable* as `overflow-y-hidden`, and
`frontend/src/App.layout.test.tsx:154` asserts the class — a test S3-31 was forbidden to
modify, because needing to modify it was the agreed signal that the ticket had broken **D22**
rather than completed it.

The ruling rests on more than that procedural point, because a procedural point alone would
be a reason to re-open rather than to accept:

- **There is no "unset" utility.** The spellable values are `auto`, `hidden`, `scroll`,
  `visible` and `clip`. `overflow-y: visible` on an element whose x-axis is `auto`
  **computes to `auto` anyway** — CSS does not allow one axis to be `visible` while the
  other is not. So the real choice is only between **`auto` and `hidden`**.
- **Of those two, `auto` is the safe failure and `hidden` is the unsafe one.** `hidden` is a
  **clip**: content that somehow exceeded the board would be unreachable, silently, with no
  scrollbar to say so. `auto` degrades to a scrollbar — visible, recoverable, and loud. This
  project chooses loud over silent everywhere else (**D20**, **D32**); it does not get to
  choose silent here for tidiness.
- **After S3-31 the board's vertical axis should never engage at all.** The columns stretch
  to the board's height and each column's card list owns its own `overflow-y-auto` behind an
  unbroken `min-h-0` chain, so nothing inside the board can exceed it. **"Should never
  engage" is not "cannot", and the record must say the weaker true thing rather than the
  stronger false one.** That is the entire correction: D29 part 3's wording promised a
  removal and delivered a dormancy, and the criterion is rewritten in `TASKS.md` to state
  what is actually true and testable — *the vertical scroll container **in use** is the
  column's card list; the board retains `overflow-y-auto` as a dormant fallback*.
- **D22 is unchanged.** *The document never scrolls* was always the load-bearing half, and
  it is still green, unmodified.

**2. The column's card list is `overflow-x-auto`, not `overflow-x-hidden`. RATIFIED.** The
Dev chose `auto` because **S3-02's shrink audit fails `overflow-hidden` by name**, and that
is the right reason:

- **D20 refuses `overflow-hidden` because it is the canonical way a localised string is
  silently clipped**, and Russian runs ~30% wider. A card list is *full* of localised
  titles. Excepting the one container most densely packed with user text would invert the
  rule at exactly the point it exists for.
- **An exception would be the first allow-list-shaped carve-out** in a project whose
  `GUARD_ALLOW_RE` is empty and, by standing rule, stays empty — every hit so far has been a
  real bug fixed at the source.
- **The x-axis should never engage either**, for the same reason as the board's y-axis: the
  cards are `min-w-0` and wrap. If it ever does engage, a scrollbar appears and someone can
  see it, which is the direction D20 chose deliberately.

**No ticket for either half.** Part 1 is a wording correction in `TASKS.md`; part 2 is a
ratification of shipped code. Both are **PM rulings and the user may overturn them.**

### D34 — a source sweep enumerates recursively, or it is not an enumeration (PM ruling, after S3-31; opens K17, ticket S3-35)

**This is D32 applied to the one sweep S3-34 was told to report on and not touch.**
S3-34's criterion required the Dev to report on `frontend/src/components/surface.test.tsx:85`
rather than change it. The report came back, and it is a real finding with a real distinction
from the store sweep:

`surface.test.tsx:85` is `import.meta.glob('./*.tsx')` — **not recursive**, the same latent
problem S3-34 just fixed for the store. It is **mitigated differently and only partly**: it
filters test files with `path.endsWith('.test.tsx')` *inside the loop* rather than by
pattern, and it carries a floor (`'nothing was scanned'`, `> 5`). So it **cannot go silently
empty** — D32's vacuity half is already covered — but it **can go silently partial**, which
D32 forbids in the same sentence (*"must not silently under-match"*).

**Why that matters more here than it did for the store, and this is the whole argument.**
What that sweep enforces is **D19**'s prohibition on a `requestAnimationFrame` /
`offsetHeight` / `getBoundingClientRect` resize workaround. **A component the glob misses is
a resize workaround that nothing in this repository forbids.** And the resize defect is
**K7 — the one the user actually reported**, and the one that is **structurally unverifiable
on this machine** (**E4**: no window manager, the window cannot be resized at all). The only
enforcement K7 has here is this sweep. A partial sweep is the single worst place for silence
in the whole test suite.

`frontend/src/components/` is flat today — 18 files — so it is correct today, and the
Reviewer's reading that it was not blocking was right. **Block B ends that.** S3-20 … S3-27
add the detail slide-over, the field editors, inline subtasks, the recurrence editor,
attachments, the editable time log and the running clock, the tree view, the search screen
and the archive view. Any of those may land in a subdirectory, and **a slide-over panel is
precisely where a developer reaches for `getBoundingClientRect`.**

**The ruling.**

1. **The fix is the one already proven in this repository twice** — the negative-pattern
   form `App.mount.test.tsx:47` uses and `Toast.test.tsx` now uses after S3-34:
   `['./**/*.tsx', '!./**/*.test.tsx']`. The in-loop `endsWith` filter goes, because the
   pattern now does that job and two spellings of one rule is the defect this project keeps
   paying for.
2. **The floor stays**, and is re-checked against what the recursive glob now matches. A
   sweep that finds nothing must still be red (**D32**).
3. **It gates block B**, alongside **S3-30** and **S3-34**. This is a **PM ruling and not
   the Reviewer's**, so it is worth saying why rather than asserting it: the change is one
   line plus a deletion; the surface it protects grows in block B specifically; and it
   protects the enforcement of the defect the user reported and no machine here can see.
   The cost of doing it now is minutes. The cost of doing it after block B is auditing ten
   new components by hand for a rule that was supposed to be mechanical.
4. **No production file is touched, and no classification changes.** This is a test-file
   glob and a deleted `continue`.

**Rejected alternative: leave it, because `src/components/` is flat and a test asserts the
structure elsewhere.** Nothing asserts that `src/components/` stays flat. The floor proves
the sweep found *something*, never that it found *everything* — which is exactly the
distinction S3-34 was raised on, and rejecting it here would be the two-tickets-one-block
inconsistency **K16** was raised on.

**This is a PM ruling and the user may overturn it.**

### OQ1 — which Cyrillic face? **CLOSED — the user answered: Inter**

> **ANSWERED.** **Inter**, one face vendored under **both** family names via
> `unicode-range` — the recommended option below. **D23** now names it, **S3-06 is
> unblocked**, and the dependency is `@fontsource/inter`, which S3-06's commit body must
> call out as a new dependency. The Dev still verifies real coverage from the package's own
> `unicode.json`, and the file is loaded **from disk** — no CDN, no `<link>`, no runtime
> fetch, which is a hard project rule and not a preference. The table below is kept because
> the runner-up and the reason it lost are worth keeping.

**As asked:** D23 fixed the approach and left the face open. Manrope and Inter were both
named as candidates and neither had been chosen.

| Option | For | Against |
|---|---|---|
| **Inter** — *recommended* | Very wide Cyrillic coverage; screen-first design and hinting at 13px, which is Nexus' base size; x-height close to both Space Grotesk and Figtree, so line boxes do not jump between Latin and Cyrillic in the same string | Neutral to the point of anonymous; it will not look like Space Grotesk's geometry |
| **Manrope** | Geometric, visibly closer to Aurora's Space Grotesk; a good match under one of the two palettes | The same file also serves Studio, where it is a worse match than Inter is |
| **Two faces, one per palette** | Best visual match on both | Doubles the bundled font weight, and D23 as the user stated it says *"a" face registered under the **same** two family names* |

**Recommendation: Inter, one face, both family names.** One face means one `unicode-range`
block and one coverage test. **The Dev must verify the chosen package's real coverage from
its `unicode.json` before committing** — including `U+2116` (№) and the combining acute
`U+0301` — rather than trusting this table.

### OQ2 — what may "custom RRULE" mean? **CLOSED — the user answered: option B**

> **ANSWERED, and NOT with the recommendation.** The user chose **B**, the middle option:
> the parser is widened to accept **`COUNT`** and **`UNTIL`**, so the editor offers *"repeat
> N times"* and *"repeat until date"*. `BYSETPOS`, `BYMONTH` and ordinal weekdays stay
> **rejected at parse time**.
>
> That answer is **D27**, and the consequences the option's own "Cost" column predicted —
> new domain work, and the **D5** streak question a terminating series raises — are ruled on
> in **D28**. It added one ticket: **S3-19** widens `internal/domain`, and **S3-23**, the
> recurrence editor, is **blocked on it**.
>
> **The recommendation below is kept, and it is the recommendation, not the decision.** A
> reader who finds "A — *recommended*" three paragraphs down and stops reading has read the
> wrong thing.

The brief promises a *"recurrence editor (daily/weekly/custom RRULE)"* in the detail panel.
**Stage 1's hand-rolled parser does not support "custom RRULE" in the general sense.**
`internal/domain/recurrence.go` accepts `FREQ=DAILY`, `FREQ=WEEKLY` with optional `BYDAY`,
and `FREQ=MONTHLY` with optional `BYMONTHDAY` (1..31), plus `INTERVAL` and `WKST=MO`.
Everything else — **`COUNT`, `UNTIL`, `BYSETPOS`, `BYMONTH`, ordinal weekdays like "the 2nd
Monday"** — is rejected **at parse time** with `ErrUnsupportedRecurrence`.

So the promise and the implementation disagree, and somebody has to say which one moves.

| Option | What ships | Cost |
|---|---|---|
| **A — *recommended*. Custom means custom within the supported subset.** The editor offers exactly what the parser accepts: frequency, interval, weekdays, month-day. A rule outside it is not offerable, so `ErrUnsupportedRecurrence` becomes unreachable from the UI | A complete, honest editor in Stage 3, with no new domain work and no new dependency | "Custom RRULE" reads narrower than the brief's phrasing. `UNTIL`/`COUNT` — "every weekday until 31 Dec" — is the most plausible thing a user will reach for and will not find |
| **B — widen the parser first.** Add `COUNT` and `UNTIL` (the two that carry real user value), leave `BYSETPOS`/`BYMONTH`/ordinal weekdays rejected | Covers most real habits. Bounded: two fields, both of which terminate a series | New domain work in a stage that already opens with nine defects, and **streaks (D5) count scheduled occurrences** — a terminating series changes what "a scheduled occurrence passed unchecked" means at the end of the series, which is a rule that needs its own decision |
| **C — adopt an RRULE library.** | Full RFC 5545 | A dependency in the one package that is **pure and has 100% coverage**, to serve an editor whose UI cannot express most of what it would then accept |

**Recommendation was: A for Stage 3, with B carried as a named follow-up** if the user
wanted end dates. **The user took B directly** (**D27**). The criterion A was recommended
for survives the change and is not weakened by it: **S3-23's acceptance criterion is still
that the editor cannot compose a rule Go would reject** — a checkable claim, and a better
one than a free-text RRULE box with an error message under it. The reachable space the
editor must round-trip is simply larger by two fields.

Whichever was chosen, one rule holds: **the editor does not parse, validate or expand an
RRULE in TypeScript.** It collects structured choices, Go builds and validates the rule and
returns the structure back, and the frontend renders that structure through i18n. A
human-readable sentence **is not returned from Go** — that is D15's rejected option (c), and
it cannot be translated.

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

**E4 — screenshots ARE possible here; input is NOT.** This corrects a claim that stood
unchallenged from Stage 0 to Stage 3 and sent three stages of visual checks into the
hand-owed pile without cause. **Verified by running them:**

| Present | Absent |
|---|---|
| `Xvfb`, `xvfb-run`, `import`, `convert`, `xwininfo`, `xdpyinfo` | `scrot`, `grim`, `xdotool`, `xte`, `wmctrl`, python-Xlib, `sqlite3`, **any window manager** |

Four facts follow, and none of them is negotiable by a hopeful ticket:

1. **`GDK_BACKEND=x11` is the whole trick, and this shell sets `WAYLAND_DISPLAY=wayland-0`.**
   With `WAYLAND_DISPLAY` set, GTK prefers the Wayland backend and the window opens on the
   **real compositor**, leaving `:99` with **no window at all** while the process runs
   healthily and exits 0. The capture is then one flat colour, which reads exactly like a
   working capture of a broken app. The working invocation is:

   ```sh
   env -u WAYLAND_DISPLAY DISPLAY=:99 GDK_BACKEND=x11 LC_NUMERIC=C ./build/bin/nexus &
   ```

   > **CORRECTED — and the wrong claim is kept here on purpose, because how it was wrong is
   > the reusable part.** From the first writing of E4 until this correction, item 1 read:
   > *"WebKitGTK will not paint under Xvfb until `WEBKIT_DISABLE_DMABUF_RENDERER=1` **and**
   > `WEBKIT_DISABLE_COMPOSITING_MODE=1` are exported."* **Those two variables do nothing on
   > this machine.** The claim was produced by changing several variables in one step and
   > attributing the win to the wrong one — the same defect this project keeps catching in
   > comments and in `TASKS.md` prose, here committed against the environment itself.
   > Re-measured by isolating **one** variable at a time against the same running binary:
   >
   > | configuration | windows on `:99` | distinct colours |
   > |---|---|---|
   > | `WEBKIT_DISABLE_*` only | **0** | **1** |
   > | `GDK_BACKEND=x11` only | 2 | **961** |
   > | both | 2 | 961 |
   >
   > `GDK_BACKEND=x11` alone is sufficient; the `WEBKIT_DISABLE_*` pair is neither necessary
   > nor sufficient. It is harmless to export, but a ticket may not present it as the cause,
   > and `make shots` (**S3-32**) bakes in `GDK_BACKEND=x11` and `env -u WAYLAND_DISPLAY`.

2. **Two checks, in this order, and the first is the cheap one.**
   **(a) Window presence:** `xwininfo -root -children | grep nexus` — zero matches means no
   window was ever created **on this display**, which is a different failure from a window
   that painted blank and has a different fix (the backend, not the renderer).
   **(b) Blankness:** `convert <png> -format "%k" info:` returning `1` means nothing painted.
   A harness runs both; neither alone distinguishes the two failures, and the `%k` floor
   cannot tell you *why* it is `1`.
3. **Input cannot be driven.** No `xdotool`, no `xte`, no `wmctrl`, no python-Xlib, and
   installing any of them needs `sudo`. Keystrokes, clicks and drags are out of reach.
4. **The window cannot be resized**, because there is **no window manager** at all. So
   **K7**, the resize defect the user reported, is structurally unreachable here no matter
   what is installed short of one.

State selection therefore happens by **seeding persisted settings** before launch, which is
possible because the palette, theme and language are SQLite rows and `internal/store`
honours `XDG_DATA_HOME`. **D31** turns all of this into `make shots` and states precisely
what it does and does not discharge.

---

### Known issues — all decided, none deleted

**All seventeen are DECIDED, none is deleted.** **K6 – K14 are new**, and all nine come out
of the one thing no gate on this machine can do: **the user ran the app by hand on real
hardware.** Eight of the nine are invisible to `make check`, `make front-test` and
`make guard` alike, because jsdom has no layout engine and no font engine and this machine
has no display. Each was then traced to exact lines by a read-only investigation, and each
has a ruling — **K6 → D18, K7 → D19, K8 → D20, K9 → D21, K10 → D22, K11 → D23 (the
user's), K12 → D24 and D25, K13 → deferred with a ticket, K14 → deferred by the user.**
They become the blocking block **S3-01 … S3-09**.

K1, K2 and K3 were carried into Stage 2
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

#### K6 – K14 — found by the user's hand pass after Stage 2 closed

**None of these was findable here.** Stage 2's PASS was honest about what it covered; this
is the part it said it did not cover, arriving all at once. **The lesson is not that the
gates are bad — it is that the five checks Stage 2 listed as owed to a human were owed for
a reason, and eight of these nine live in exactly that gap.**

**K6 — the dragged card vanishes the moment it is grabbed. DECIDED by D18; ticket S3-05.**
`views/Kanban.tsx:294-306` mounts `DndContext` with **no `DragOverlay`**. Confirmed against
the installed `@dnd-kit/sortable@10.0.0`: `sortable.cjs.development.js:314` computes
`useDragOverlay = Boolean(dragOverlay.rect !== null)`, which is **false**, so
`shouldDisplaceDragSource` is true and the active card's own `<li>` is translated in place
inside its column. Two independent consequences follow from that one fact:

1. **The neighbouring column paints over it.** `components/Column.tsx:165` carries
   `backdrop-blur-glass` → `backdrop-filter`, which per Filter Effects L2 creates a
   **stacking context — including under Studio, because `blur(0px)` is not `none`**. The
   `z-10` at `Column.tsx:126` only orders siblings *within* the source column and cannot
   reach across one.
2. **The card teleports home and stops following the pointer.**
   `sortable.cjs.development.js:514-524`: each column is its own `SortableContext`
   (`Column.tsx:187`), so once the pointer is over another column `overIndex` is `-1` in
   the source context, `finalTransform` becomes `null`, and the translation is dropped.

`draggingNodeId` (`store/ui.ts:28`) is currently **write-only, with no consumer anywhere in
`frontend/src`**. D18 gives it one.

**K7 — resizing the window breaks the layout, and it does NOT recover without a restart.
DECIDED by D19; ticket S3-04. Cause CONFIRMED BY EXPERIMENT.**
The user ran the discriminating test: with the layout stuck broken, **switching the palette
to Studio — whose `--blur` is `0px` — repaired it live, with no restart.** That is
WebKitGTK compositing-layer staleness, and it **eliminates the competing
scrollbar-hysteresis theory**, which a palette switch could not have affected.
`backdrop-blur-glass` is currently on **every** surface — roughly fifty simultaneously:
`Column.tsx:165`, `Card.tsx:91`, `HabitChip.tsx:69`, `QuickAdd.tsx:239`,
`CommandPalette.tsx:163`, `Toast.tsx:54` — and each one gets its own compositing layer.
**D19 keeps the blur on the five columns and the two overlay scrims and removes it from the
card, the chip and the toast**, with a stated escalation ladder if seven is still too many.
Whether it is enough **cannot be established on this machine**; it is S3-09's hand pass.

**K8 — `shrink-0` on a max-content localised string overflows the column. DECIDED by D20;
ticket S3-02. CONFIRMED BY ARITHMETIC AND BY SCREENSHOT.**
`frontend/src/style.css:38` sets `html { font-size: 13px }`, so every Tailwind rem is 13/16
of nominal: **`min-w-36` is 117px, not 144px**, and `gap-2`/`p-2` are 6.5px, not 8px. At the
floor the usable column content width is ~102px and the card interior ~87px.
`DueBadge.tsx:43` is `shrink-0 font-mono`, so it takes its max-content width and refuses to
shrink: **~94px in English and ~125px in Russian against 87px available — it overflows even
in English.** The same defect is at `Column.tsx:175` (the card-count span) and
`HabitChip.tsx:69`. The user's screenshot proves the consequence: the column headings wrap
**one character per line** — "Бэ / кл / ог", "Се / го / дн / я" — because the `shrink-0`
count span "0 карточек" (~86px) starves the heading down to ~16px. Estimated onset widths,
**±5% and not measurements**: RU heading ~784px, RU due badge ~814px, EN due badge ~659px,
board floor 624px.

**The irony is the important part, and it is D17 in a new place.** `Card.tsx:36-43` states
*"nothing on this card has a fixed width and nothing is `whitespace-nowrap`"* and
`App.accept.test.tsx:344-355` asserts the absence of `truncate` and `whitespace-nowrap`.
Both are true. **`shrink-0` on a localised max-content string clips identically and was not
on the list, so the mechanism-based RU audit was green on a real clipping bug.**

**K9 — the window has no minimum size, and display scaling varies. DECIDED by D21; ticket
S3-03.**
`main.go:137-139` sets `Width: 1024, Height: 768` and **no `MinWidth`/`MinHeight`**. Wails
calls `SetMinSize(0, 0)` unconditionally (`window.go:134-135` → `window.c:266`), hinting
GTK with `min_width = min_height = 0`, so the window can be dragged roughly **380px below
the layout's own ~624px floor**. **The user has confirmed they change display scaling
frequently and that it varies**, so the startup window is not reliably 1024 CSS px and may
begin at or below the floor — which is why this is not merely a nicety about dragging an
edge.

**K10 — there is no height chain, so the document scrolls instead of the board. DECIDED by
D22; ticket S3-01.**
There is no `html`/`body` height rule and **no `#root` rule at all** — verified in the
compiled CSS. The only height declarations are `body { min-height: 100vh }`
(`style.css:44`) and `min-h-screen` on the shell (`App.tsx:142`). `<main className="min-w-0
flex-1">` (`App.tsx:154`) **lacks `min-h-0`**, so its `min-height: auto` floors it at
min-content and it cannot shrink — and the `min-w-0` that is there is a **no-op**, because
the shell is a *column* flex container. Separately, `Kanban.tsx:316` sets `overflow-x-auto`,
and per CSS Overflow 3 an `overflow-x` of `auto` with `overflow-y: visible` **promotes
`overflow-y` to `auto`**, so the board already clips and scrolls in **both** axes,
contradicting its own comment at `Kanban.tsx:307-310`. Finally, **every `.focus()` call in
the codebase omits `{ preventScroll: true }`** — `Kanban.tsx:152`, `HabitStrip.tsx:85`,
`QuickAdd.tsx:121,164,168,201,225`, `CommandPalette.tsx:93,95`,
`AppearanceControls.tsx:164` — so each one scrolls every scrollable ancestor, the board
included. The user's screenshot shows a document-level vertical scrollbar, which is this.

**Investigated and explicitly RULED OUT — do not let a ticket chase it:** the "latched
`scrollLeft`" theory. The columns are `flex-1 basis-0`, so `scrollWidth` tracks
`clientWidth` and the user agent clamps the offset; nothing in the frontend ever reads or
writes `scrollLeft`/`scrollTop` — grepped.

**K11 — neither UI font contains a single Cyrillic glyph. DECIDED by D23, which is the
USER'S ruling; ticket S3-06.**
`design/tokens.css:18` sets Aurora's `--font-ui` to `'Space Grotesk'` and `:56` sets
Studio's to `'Figtree'`. Verified from `node_modules/@fontsource/*/unicode.json` **and from
the shipped `frontend/dist/assets/` the user is running**: Space Grotesk covers
`[vietnamese, latin-ext, latin]`, Figtree covers `[latin-ext, latin]`. **Neither has
Cyrillic.** Only JetBrains Mono does. **81 of the 82 leaves in `ru.json` are Cyrillic**, so
Russian drops the whole UI into the system `sans-serif` while numbers and dates stay in
JetBrains Mono — two unrelated typefaces in one header, different metrics, and the
`flex-wrap` header very plausibly gaining a second row. **This is the user's reported
"theme and colour buttons break when I switch language".** Related and confirmed:
`frontend/index.html:14` hard-codes `<html lang="en">` and **nothing ever updates it** —
grepped, zero writers.

**K12 — error toasts stack without limit, and none of them says what failed. DECIDED by
D24 (the stacking) and D25 (the message); tickets S3-07 and S3-08.**
`store/toast.ts:35-47` appends unconditionally: **no cap, no de-duplication, no
auto-dismiss.** The user's screenshot shows three identical toasts — *"Что-то пошло не так
/ Nexus не смог выполнить это действие / Ничего не изменилось"* — filling the lower half of
the window and covering the board. Two separate problems, kept separate on purpose: the
**stacking**, and the fact that a domain **refusal** (D9 declining to put a project into
`doing`, say) is a rule working correctly and is being rendered as a malfunction.
**Unresolved and worth saying plainly: we still do not know which three operations failed
in that screenshot.** Whatever ships, the user must end up able to tell — that is a
criterion on S3-08, not a wish.

**K13 — keyboard shortcuts match the character, so a Cyrillic layout would kill Ctrl+N and
Ctrl+K. LATENT, NOT the reported bug; ticket S3-28, deliberately ranked low.**
`lib/keyboard.ts:107-115` matches on `event.key`, and `event.code` is used **nowhere** in
`frontend/src`. Under a Cyrillic keyboard layout the N key emits `т` and K emits `л`, so
both chords would silently stop working. **The user has confirmed they keep a Latin
layout**, so this is not what they hit. It is a real latent defect in a Russian-language
application and it is recorded rather than fixed under pressure — **it is not in the
blocking block.**

**K14 — the palette *selection* is not visibly indicated. DEFERRED BY THE USER; not
scheduled, and that is the user's call, not an omission.**
The user raised it in their own words and then explicitly parked it: record it as a later
design refinement, and it is fine as it stands. It is logged here so it is not lost, and it
is **not** in Stage 3's blocking block and not in its feature block. It returns when the
user asks for it.

---

#### K15 – K16 — found after the block A review returned PASS

**Both were found after `c32a1c9..b53f1a4` passed.** K15 came from the **first real
screenshot this project has ever taken of its own binary** (**E4**); K16 came from the
Reviewer's own non-blocking list and is the reason it named a condition on block B rather
than simply closing.

**K15 — the columns do not fill the window height, so an empty column is barely a drop
target. DECIDED by D29; ticket S3-31. CONFIRMED BY SCREENSHOT.**
`views/Kanban.tsx:399` is `flex h-full min-w-0 items-start gap-2 overflow-x-auto
overflow-y-auto`. `items-start` overrides flex's default `stretch`, so each column is only
as tall as its cards: in a 1024×768 capture of the running binary, **four of the five
columns are short boxes at the top of a large empty area**. The Reviewer logged it as
non-blocking observation 8 and called it *"purely visual"* — and on the pixels alone that
was a fair reading. **It is not purely visual.** `components/Column.tsx`'s `useDroppable`
exists to catch a card dropped on an **empty column** or on the **padding below the last
card**; a column that ends where its cards end has almost none of either, and the large
region below it belongs to the board rather than to any column. So this lands directly on
**S3-05**'s drag work and on **D18**'s overlay. A second, separate fact from the same
line: **nothing scrolls inside a column**, so a column with more cards than fit has nowhere
to put them — the half **D22** left unfinished, tolerable until now only because
`overflow-x: auto` had silently promoted the board's `overflow-y`.

**The same capture discharged two long-standing items, and that is the other half of the
news.** At a real 1024×768 in **Russian**: the due date `25 сент. 2026 г.` sits inside its
card, the column headings wrap without overlapping, and **there is no document scrollbar**.
So **K8 is confirmed fixed on real paint** and no longer owed to a hand pass, and Stage 2's
*"a window opens and the first `Board()` returns five columns"* is **discharged** — the
window opened and five columns were photographed.

**K15 is now FIXED, and a second 1024×768 capture taken after `e7868c8` confirms it.** All
five columns are full-height bordered boxes running from under the header to the bottom
edge, so `useDroppable`'s rectangle is the whole visible column — where before, four of the
five were header-high strips. **That closes the static half of K15 and nothing more.** The
empty-column drop and the in-column scroll need input, which cannot be driven here
(**E4**), and they stay on **S3-09**.

**The same capture showed two things that are known and not fixed**, recorded here so a
later reader does not re-report them as new:

- **The UI renders in the system `sans-serif`.** That is **K11**, and **S3-06** is blocked
  on the user running `npm install @fontsource/inter`. Expected, not a regression.
- **The header occupies two rows.** This is **not yet a finding either way.** The capture
  was taken under the wrong font, and *"the header does not gain a row"* (**K11**/S3-06,
  S3-09 item 6) is a claim about Russian **in the designed family** — a `flex-wrap` header
  measured with substituted metrics proves nothing about the shipped one. It is re-judged
  from the re-taken captures after S3-06 lands, which **S3-32** already requires.

**None of this shrinks the hand pass beyond those items.** **K7 — the resize defect the
user actually reported — remains structurally unreachable on this machine**: there is no
window manager, so the window cannot be resized at all, at any size, by any means available
here. No capture, present or future, moves it.

**K16 — `refuse()` is applied by hand 26 times across 22 bound methods, and nothing forces
a 23rd. LATENT, no symptom today. DECIDED by D30; ticket S3-30 — SHIPPED at `c50cf24`.**
**S3-08**'s coverage was verified complete by the Reviewer, and a miss would degrade a
message rather than corrupt data — which is why block A was not held for it. What makes it
worth a number is the **internal inconsistency**: **S3-01, in the same block, argued that an
eleven-times-by-hand rule needed a guard and built check 7 for it**, and **S3-08 then
shipped a twenty-six-times-by-hand rule without one**. Block B adds ten or more bound
methods, which is exactly when a hand-applied rule decays. The Reviewer's condition —
**a named ticket before S3-10 starts** — is met by **S3-30**, whose committed test
enumerates the bound surface from `app.go`'s own text and passes at **22**.

> **The figures in this entry were "27 across 25" and both were wrong.** The bound surface
> is **22**: `app.go` has 25 `func (a *App)` declarations, three of which — `startup`,
> `context`, `onIPCMessage` — are **unexported, are not bound by Wails, and do not return
> `(T, error)`**. The call sites number **26**: `grep -c 'refuse(' app.go` says 27 because
> `app.go:60` is a **doc comment**, not a call. Nothing is broken by the error, and 22
> clears **D32**'s floor of 20 — but **S3-30's acceptance criterion *"All 25 current bound
> methods are found"* was unsatisfiable as written**, and is corrected in `TASKS.md`. See
> **D30** for the full correction. It is kept rather than overwritten because a number
> restated in two places and corrected in one is exactly the defect **S3-33** was written to
> fix, in a smaller costume.

---

#### K17 — found by S3-34's own report, after S3-31 shipped

**K17 — the component sweep's glob is not recursive, so D19's resize-workaround ban can go
silently partial. LATENT, no symptom today. DECIDED by D34; ticket S3-35.**
`frontend/src/components/surface.test.tsx:85` is `import.meta.glob('./*.tsx')`. It is the
same non-recursive defect **S3-34** fixed for the store, mitigated differently: it filters
test files with `path.endsWith('.test.tsx')` *inside the loop* and carries a floor
(`'nothing was scanned'`, `> 5`), so it **cannot go silently empty** but **can go silently
partial**.

**It is ranked above the store sweep, not below it, for one reason**: what it enforces is
**D19**'s ban on a `requestAnimationFrame` / `offsetHeight` / `getBoundingClientRect` resize
workaround, so a component the glob misses is a resize workaround **nothing forbids** — and
the resize defect is **K7**, the one the user actually reported and the one this machine
**cannot** verify at all (**E4**: no window manager). `src/components/` is flat today (18
files) and the sweep is therefore correct today; block B's detail panel, field editors,
subtasks, recurrence editor, attachments, time log, tree, search and archive views are when
that ends. **D34 rules it a gate on block B** alongside S3-30 and S3-34 — a PM ruling, not
the Reviewer's, argued in D34 rather than asserted.

---

**Status: decisions locked — D1–D34, E1–E4. Stage 0 is CLOSED (PASS). Stage 1 is
CLOSED (PASS) — all twenty-two tickets, S1-01 … S1-22, ACCEPT met at 100.0% / 92.9%,
PASS returned on the fourth review at `a1f09b7` after three FAILs whose history is kept
in §5. **Stage 2 is CLOSED (PASS)** — twenty-two tickets, S2-01 … S2-22, in `TASKS.md`,
all committed plus the gap-closing `7af4d1d`, **PASS returned on the second review at
`8818db4`** after one FAIL whose history is kept in §5. **ACCEPT met**: the no-mouse
create → move across all five columns → complete flow is demonstrated by
`frontend/src/App.accept.test.tsx` driving the assembled `<App />` with an exact asserted
call sequence, and `make guard` check 6 fails if a mouse event ever enters that file.
`make check` green over all five gates, `make cover` 100.0% / 93.8%, `make front-test`
22 files / 247 tests / 0 failures, `make guard` six of six with an empty
`GUARD_ALLOW_RE`; both round-1 blockers verified closed by reverting each fix and
watching the specific test fail. All five Stage 2 obligations were absorbed into
tickets — C1 → S2-03, C2 → S2-01, C3 → S2-08, C4 → S2-06, C5 → S2-14 — and every known
issue has a decision: **K1 → D12** (user), **K2 → D14**, **K3 → D15** and
**K5 → D16** (PM rulings, overturnable), **K4** RESOLVED in `9664506`.
**Two planning defects were found mid-stage, both by the Dev refusing to widen scope
silently, and both corrected in `TASKS.md`:** nothing owned `App.tsx` or `main.tsx`, so
nothing mounted anything; and the out-of-scope line contradicted the brief on *set
priority*, resolved in the brief's favour. Three disclosed widenings (S2-13, S2-21,
S2-22) are **ratified**, and **D17** records the one rule the stage had never written
down. **Carried into Stage 3, and not scheduled**: the **five checks owed to a hand
pass** (the ten-step run on the real binary including the three unbacked D8 due-badge
assertions, no background flash in either theme, Russian not clipping at 1024×768, the
painted `:focus-visible` ring, and S2-07's window-opens/five-columns), **C6** (the five
enum sets spelled a second time in the locale files — a judgement call, not a defect),
**K1/D12, K2 and K3** from Stage 1's handoff, and **D16/K5** — the drift gate is built
and **nothing draws a drift**, which no document may imply otherwise.
**A green `make guard` is NOT proof that the frontend computes nothing**: checks 3a/3b
are name-based heuristics, `wireDate` sat behind a green guard for an entire stage, and
**D17** says so — reading the diff is still the check.
**Stage 3 is IN PROGRESS, block A open** — thirty-five tickets, **S3-01 … S3-35**, in
`TASKS.md`, in two blocks. **The user has since run the app by hand on real hardware and
found nine defects**, recorded as **K6 – K14** and ruled as **D18 – D26**: eight of the
nine are invisible to every gate here, because jsdom has no layout engine and no font
engine and this machine has no display. **Block A, S3-01 … S3-09, is blocking and must be
green before a single feature ticket starts** — the user's own condition. **D23 is the
user's** (vendor a Cyrillic face under the same two family names via `unicode-range`), as
is the discriminating Studio experiment behind **K7**; **D18–D22 and D24–D26 are PM
rulings and the user may overturn any of them**. **D26 closes C6** by publishing the enum
sets from Go — noting the honest correction that current locale parity is **clean**, so C6
is a drift risk and not a present defect. **Both open questions have since been answered
by the user and are CLOSED. OQ1 → Inter**, one face under both family names, which fixes
**D23**'s one open half and unblocks S3-06. **OQ2 → the middle option, not the PM's
recommendation**: `COUNT` and `UNTIL` enter the recurrence language (**D27**, the
user's), while `BYSETPOS`, `BYMONTH` and ordinal weekdays stay rejected at parse time.
**D28** is the PM ruling that spells out the consequences — an ended series has no further
scheduled occurrences, so by **D5** nothing can break the streak and it **freezes** — and
that answer added one ticket: **S3-19** widens `internal/domain`, and **S3-23**, the
recurrence editor, is blocked on it. **K13** (shortcuts match the character, not the
key) is latent and ranked low; **K14** (the palette selection is not visibly indicated) is
**deferred by the user**. The **three D8 due-badge assertions still have no mechanical
backing**, and S3-09 is where that is fixed and the rest is looked at by eye.
**The block A review returned PASS on the code** for `c32a1c9..b53f1a4`, and **S3-30,
S3-31, S3-33 and S3-34 have since shipped green at `c3c954e`** — five gates, `make guard`
8/8 with `GUARD_ALLOW_RE` empty, `make front-test` **374 across 29 files**, `make cover`
**100.0% / 94.0%**. Block A stays **open** on **S3-06** (blocked on the user running
`npm install @fontsource/inter`), **S3-32**, **S3-35** and **S3-09**. The five tickets added
after the PASS — **S3-30 … S3-34** — are ruled by **D29 – D32** and recorded as
**K15 – K16**; **S3-35** is a sixth, ruled by **D34** and recorded as **K17**, raised by
**S3-34's own report** on a glob it was told to report on and not touch. All of
**D29 – D34 are PM rulings and the user may overturn any of them**. **S3-30** and **S3-34**
are the Reviewer's explicit conditions on starting block B; **S3-35 is a third the PM
adds**, and D34 argues it rather than asserting it — a hand-applied `refuse()`, a
non-recursive store sweep and a non-recursive component sweep all decay exactly when block B
adds bound methods, store slices and components.
**E4 is new, it is a correction rather than a discovery, and it has since been corrected a
second time**: `xvfb-run`, `import` and `convert` are installed and always were, so
**capture is possible here** — but the **cause of the blank capture was named wrong**. It is
**`GDK_BACKEND=x11`**, not the `WEBKIT_DISABLE_*` pair, which does nothing on this machine;
this shell sets `WAYLAND_DISPLAY`, so GTK opens the window on the real compositor and `:99`
gets none. Isolated one variable at a time: `WEBKIT_DISABLE_*` alone → **0 windows, 1
colour**; `GDK_BACKEND=x11` alone → **2 windows, 961 colours**. A harness checks
**`xwininfo -root -children | grep nexus` first**, then the `%k` blankness floor. Capture has
now photographed the running binary twice: it **confirmed K8 fixed on real paint**,
**discharged** Stage 2's *"a window opens and `Board()` returns five columns"*, **found
K15**, and — after `e7868c8` — **confirmed K15 fixed**, all five columns full-height from
under the header to the bottom edge. The same frame shows the **system `sans-serif`**
(**K11**/S3-06, expected) and a **two-row header**, which is *not yet a finding either way*
because the font is wrong and is re-judged after S3-06.
**D31 makes capture `make shots` and states its limits plainly**: it converts static
paint from *owed to the user's hardware* to *observable here*, it does **not** turn an eye
into a machine, and it **cannot** reach anything needing input or a resize — the
keyboard run, the focus ring, the drag, the first-frame flash, and **K7, the defect the
user actually reported, which is structurally unreachable here because there is no window
manager and the window cannot be resized at all**. **S3-09 is still the gate into block B**;
capture shortens its list and does not close it, and **no document may write this as if the
hand pass is shrinking further than it is**. **S3-29 (`README.md`) is unblocked** and
re-pointed at `make shots`. **Two counts were corrected this pass and both corrections are
kept visible**: the bound surface of `app.go` is **22**, not 25 (three declarations are
unexported and unbound), and the `refuse()` call sites number **26**, not 27 (`app.go:60` is
a doc comment) — the same defect class **S3-33** fixed in `frontend/src`, one document up.
See §5, "Stage 2 — CLOSED, PASS", "Stage 3 — IN PROGRESS" and "Carried into
Stage 3", and `TASKS.md`.**
