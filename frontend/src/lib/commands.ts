import type { i18n as I18nInstance, TFunction } from 'i18next';

import type { AppStore } from '../store';
import type { ColumnView, SettingsView } from './client';
import { changeLanguage, resources, type LanguagePort } from './i18n';
import { adjacentColumn, KEYS } from './keyboard';

// Nexus — what the command palette can do, and the ONE place that list is built.
//
// The component below it (components/CommandPalette.tsx) filters, renders and
// runs; it decides nothing about what exists. That split is the whole reason
// this module is a separate file: the action list is the interesting part, it is
// a pure function of the state Go handed us, and a pure function is a thing a
// test can hold still and read.
//
// # Every action goes through the store
//
// Not one `run` below touches `lib/client.ts`. They call the same store actions
// the board, the strip and the overlays already call, so "move a card" has ONE
// implementation and the palette is a second doorway to it rather than a second
// copy of it. A palette that called MoveToColumn itself would re-spell the
// re-read, the toast and the focus handling, and the two spellings would drift
// on the first change to either.
//
// # Nothing here is a domain rule, and no set is typed in as a literal
//
//   columns     `board` — the array Go returned, in Go's order. This module
//               never writes a status down and never computes a "next" one.
//   priorities  the keys of the `palette.priority` LABEL TABLE in locales/.
//   views       the keys of the `palette.view` label table.
//   themes      the keys of the `settings.theme` label table.
//   languages   the keys of the `settings.language` label table.
//
// The label-table precedent is S2-19's, which took the quick-add type list the
// same way, and S2-15 settled the principle for the column headings: "a label
// table is presentation; a status list in code is a rule". Where Go publishes a
// set — the columns — the set comes from Go and the table only names it.
//
// # One group is registered UNAVAILABLE, with a reason, and why
//
// TASKS.md S2-20 rules on the view switcher in so many words: the views that do
// not exist yet are "registered as disabled with a localised 'coming in stage
// N', or omitted. Pick one and be consistent — a dead entry that silently does
// nothing is the worse option." Registered-with-a-reason is the choice:
//
//   switch view    Kanban is the only view Stage 2 builds. Its own entry says
//                  so; tree and calendar name the stage they arrive in.
//
// An entry with a reason is honest in a way that both alternatives are not: an
// enabled entry that does nothing lies, and an omitted entry makes the user
// wonder whether they mistyped.
//
// The PRIORITY rows used to be the second such group, because there was no
// binding to call. There is one now — App.SetPriority, over
// TaskService.SetPriority — and the rows are live. That was not scope creep:
// the brief names "set priority" among the palette's actions, and S2-20's own
// acceptance criterion is that EVERY action in its table is reachable and
// executable by keyboard alone, which four permanently-disabled rows do not
// satisfy. What is still out of Stage 2 is the priority EDITOR in a detail
// panel; a palette row is not that.
//
// A temporary reason on a row is therefore exactly what it claims to be — the
// row is disabled until the thing exists, and then it is not.

/** One row of the palette. */
export interface Command {
  /** Stable identity, for React keys and for tests. Never shown. */
  id: string;
  /** The translated text the user reads, and the text the filter matches on. */
  label: string;
  /**
   * How the same action is reached without the palette, or undefined.
   *
   * Always a `hint` off lib/keyboard.ts — S2-16's normative map — and never a
   * string written here. A palette that spelt "Ctrl+N" itself would be the
   * second copy of a binding, and the one on screen would be the untested one.
   */
  hint?: string;
  /** Why this cannot run right now, translated, or undefined when it can. */
  unavailable?: string;
  /** Does the thing. Fire-and-forget: every rejection is already a toast. */
  run: () => void;
}

/** Everything the list is built out of. All of it is state Go produced. */
export interface CommandContext {
  /** The columns, in Go's order, or null before the first read. */
  board: ColumnView[] | null;
  /** The card the board's keyboard model is on, or null. */
  selectedNodeId: string | null;
  /** The four persisted preferences, as the service last reported them. */
  settings: SettingsView | null;
  store: AppStore;
  t: TFunction;
  /** The live i18next instance — the language switch drives it. */
  i18n: I18nInstance;
}

/**
 * The stage each view that does not exist yet arrives in (`PLAN.md` §5).
 *
 * A roadmap, not a rule: nothing in Go has an opinion about it, and it is here
 * rather than in the locale files so that the sentence around it is translated
 * once instead of once per view.
 */
const VIEW_STAGE: Readonly<Record<string, number>> = { tree: 3, calendar: 7 };

/** The view this stage actually builds. */
const CURRENT_VIEW = 'kanban';

// The label tables, read as SETS. `resources.en` and not the current language:
// the keys are identical in both files — locales.test.ts is what keeps them
// that way — and the English copy is the one that is always loaded.
const LABELS = resources.en.translation;

const PRIORITIES = Object.keys(LABELS.palette.priority);
const VIEWS = Object.keys(LABELS.palette.view);
const THEMES = Object.keys(LABELS.settings.theme);
const LANGUAGES = Object.keys(LABELS.settings.language);

/**
 * The entry after `current` in a label table, wrapping at the end.
 *
 * Used by "toggle the theme" and "switch the language", which are single
 * entries rather than one entry per value because that is what TASKS.md asks
 * for and what the hand ACCEPT script drives: type `theme`, press Enter, watch
 * it flip. With two entries a cycle and a toggle are the same thing, and the
 * cycle is what keeps working when a third palette or a third language lands.
 *
 * A value the table does not name — which is what a stale or unreadable setting
 * looks like — starts from the beginning rather than throwing.
 */
function nextInOrder(order: string[], current: string | undefined): string {
  const at = current === undefined ? -1 : order.indexOf(current);
  return order[(at + 1) % order.length];
}

/**
 * Builds the palette's rows for the state it is handed.
 *
 * Pure: it reads the context, returns an array, and touches nothing. The `run`
 * closures are the only part that acts, and each one is a single store call.
 */
export function buildCommands(context: CommandContext): Command[] {
  return [
    newTaskCommand(context),
    ...moveCommands(context),
    ...priorityCommands(context),
    ...timerCommands(context),
    ...viewCommands(context),
    themeCommand(context),
    languageCommand(context),
  ];
}

function newTaskCommand({ t, store }: CommandContext): Command {
  return {
    id: 'newTask',
    label: t('palette.action.newTask'),
    hint: KEYS.quickAdd.hint,
    // Setting the overlay to quickAdd REPLACES the palette: store/ui.ts holds
    // one open overlay, so the palette closes by the same act that opens the
    // quick add, and there is never a moment with two of them on screen.
    run: () => store.getState().openOverlayPanel('quickAdd'),
  };
}

/**
 * One entry per column GO RETURNED, in Go's order.
 *
 * No column is named here and none is counted: the array is the array off
 * `Board()`, the label comes from the same `locales/` table the column headings
 * use, and the target sent back is `column.status` — a string Go produced.
 *
 * The two columns either side of the focused card carry the keyboard hints for
 * the chords that would reach them, read out of S2-16's map. Any other column
 * has no chord and therefore shows no hint, which is the honest answer rather
 * than a hint that does not work.
 */
function moveCommands({ board, selectedNodeId, t, store }: CommandContext): Command[] {
  if (board === null) {
    return [];
  }

  const left = selectedNodeId === null ? null : adjacentColumn(board, selectedNodeId, -1);
  const right = selectedNodeId === null ? null : adjacentColumn(board, selectedNodeId, 1);

  return board.map((column) => ({
    id: `move:${column.status}`,
    label: t('palette.action.moveToColumn', {
      column: t(`board.column.name.${column.status}`, { defaultValue: column.status }),
    }),
    hint: hintForColumn(column, left, right),
    unavailable: selectedNodeId === null ? t('palette.reason.noCard') : undefined,
    run: () => {
      if (selectedNodeId !== null) {
        void store.getState().moveToColumn(selectedNodeId, column.status);
      }
    },
  }));
}

function hintForColumn(
  column: ColumnView,
  left: ColumnView | null,
  right: ColumnView | null,
): string | undefined {
  if (left !== null && left.status === column.status) {
    return KEYS.moveLeft.hint;
  }
  if (right !== null && right.status === column.status) {
    return KEYS.moveRight.hint;
  }
  return undefined;
}

/**
 * Set the focused card's priority, one row per entry in the label table.
 *
 * # Where the four come from, and where they do not
 *
 * `PRIORITIES` is `Object.keys` of the `palette.priority` label table — S2-19's
 * precedent, kept rather than improved on, because the alternative is worse in
 * both directions. A literal `[1, 2, 3, 4]` here would be
 * `domain.Priority.Valid`'s range written a second time; a TypeScript union
 * `1 | 2 | 3 | 4` would be the same copy with a compiler enforcing it, which
 * makes it harder to notice rather than easier. The label table has to name
 * every priority anyway — a row with no wording is not a row — so it is already
 * the one list, and reading it as a set adds nothing new to the project.
 *
 * `Number(value)` is the wire encoding of a key that IS the priority digit, not
 * a computation and not a validation: a key the table mis-spelt becomes NaN,
 * crosses as null, arrives in Go as 0 and is refused there with a message
 * naming the field. Nothing in TypeScript decides what a priority is.
 *
 * Unavailable only when there is no card — the same reason, from the same
 * table, that the move and timer-start rows use. Whether a particular node may
 * take a particular priority is Go's question and is never pre-empted here.
 */
function priorityCommands({ selectedNodeId, t, store }: CommandContext): Command[] {
  return PRIORITIES.map((value) => ({
    id: `priority:${value}`,
    label: t('palette.action.setPriority', { priority: t(`palette.priority.${value}`) }),
    unavailable: selectedNodeId === null ? t('palette.reason.noCard') : undefined,
    run: () => {
      if (selectedNodeId !== null) {
        void store.getState().setPriority(selectedNodeId, Number(value));
      }
    },
  }));
}

/**
 * Start and stop the one global timer.
 *
 * The start entry is NOT hidden for a node Go would refuse. Whether a node may
 * be timed is `domain.DoingRefusal`'s answer (D9) — a project never can — and
 * asking that question here would be the rule written a second time in the one
 * language that cannot see the domain. So the entry is offered, Go refuses, and
 * the refusal arrives as a toast like every other one.
 */
function timerCommands({ selectedNodeId, t, store }: CommandContext): Command[] {
  return [
    {
      id: 'timerStart',
      label: t('palette.action.startTimer'),
      unavailable: selectedNodeId === null ? t('palette.reason.noCard') : undefined,
      run: () => {
        if (selectedNodeId !== null) {
          void store.getState().startTimer(selectedNodeId);
        }
      },
    },
    {
      id: 'timerStop',
      label: t('palette.action.stopTimer'),
      run: () => void store.getState().stopTimer(),
    },
  ];
}

/** The views, with the one that exists saying so and the rest naming their stage. */
function viewCommands({ t }: CommandContext): Command[] {
  return VIEWS.map((view) => ({
    id: `view:${view}`,
    label: t('palette.action.switchView', { view: t(`palette.view.${view}`) }),
    unavailable:
      view === CURRENT_VIEW
        ? t('palette.reason.currentView')
        : t('palette.reason.stage', { stage: VIEW_STAGE[view] }),
    run: () => {},
  }));
}

/**
 * Flip to the next theme, through SetTheme.
 *
 * The store holds the SERVICE'S answer and applies it (store/settings.ts), so
 * there is nothing to write locally and nothing to roll back: a refusal
 * re-reads and the document follows whatever the service says.
 */
function themeCommand({ settings, t, store }: CommandContext): Command {
  return {
    id: 'toggleTheme',
    label: t('palette.action.toggleTheme'),
    unavailable: settings === null ? t('palette.reason.noSettings') : undefined,
    run: () => void store.getState().setTheme(nextInOrder(THEMES, settings?.theme)),
  };
}

/**
 * Move to the next language, through SetLanguage and then i18next.
 *
 * It goes through `changeLanguage` (lib/i18n.ts) rather than calling
 * `i18n.changeLanguage` directly, because that function holds the rule: the UI
 * switches to the language the PORT REPORTS BACK, not the one it was asked for.
 * A refusal therefore leaves the UI in the language the service still says is
 * in force, and store/settings.ts has already raised the one toast — which is
 * why the port below resolves instead of rejecting.
 */
function languageCommand({ settings, t, store, i18n }: CommandContext): Command {
  const persist: LanguagePort = async (value) => {
    const accepted = await store.getState().setLanguage(value);
    // null is "refused, and the re-read failed too". Nothing is known to have
    // changed, so nothing changes on screen either.
    return accepted?.language ?? i18n.language;
  };

  return {
    id: 'switchLanguage',
    label: t('palette.action.switchLanguage'),
    unavailable: settings === null ? t('palette.reason.noSettings') : undefined,
    run: () => void changeLanguage(i18n, nextInOrder(LANGUAGES, settings?.language), persist),
  };
}

/**
 * The rows whose label contains `query`, compared case-insensitively.
 *
 * It matches the TRANSLATED label, which is what makes the palette work in
 * Russian: a filter over ids or over English keys would answer only to English
 * words on a Russian screen. Locale-aware lower-casing, because the Turkish
 * dotted I is a real language and `toLowerCase()` is not always its answer.
 */
export function matchingCommands(
  commands: Command[],
  query: string,
  language: string,
): Command[] {
  const needle = query.trim().toLocaleLowerCase(language);
  if (needle === '') {
    return commands;
  }
  return commands.filter((command) => command.label.toLocaleLowerCase(language).includes(needle));
}
