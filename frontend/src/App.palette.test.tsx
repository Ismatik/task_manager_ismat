import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, NodeView, SettingsView } from './lib/client';
import { KEYS } from './lib/keyboard';
import { createAppStore, type AppStore } from './store';
import {
  columnView,
  createFakeClient,
  node,
  nodeView,
  settingsView,
  timerView,
} from './test/fakeClient';
import { BOTH_LANGUAGES, renderIn, tabUntil } from './test/render';

// Nexus — the command palette, asserted THROUGH the shell.
//
// This file imports `App` and does NOT import CommandPalette. The import
// restriction is the assertion: S2-20 mounts its own overlay into S2-15's
// overlay layer, and a test that imported the component would prove the
// component works while the running app had no palette in it at all.
//
// It imports `lib/keyboard` on purpose — that is the point of one of the cases
// below, which compares every hint the palette renders against S2-16's map.
// That is a library, not a component, and the map having ONE spelling is
// exactly what is being asserted.
//
// Everything is driven by `user-event` key presses. No column name is written
// down (`make guard` check 2): the columns are whatever the fake Go returned.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

const OPEN_PALETTE = '{Control>}k{/Control}';
const OPEN_QUICK_ADD = '{Control>}n{/Control}';

/**
 * A fake Go that answers the whole surface the palette touches, and records
 * what it was asked for.
 *
 * Local rather than shared, for the reason App.keyboard.test.tsx gives about
 * its own: the arguments ARE the assertion here — which column, which theme,
 * which language — and the shared fake ignores them.
 */
function paletteGo(seed: Record<number, string[]> = { 0: ['a'] }) {
  const base = createFakeClient();
  let board: ColumnView[] = COLUMNS.map((status, index) =>
    columnView(
      status,
      (seed[index] ?? []).map((id) => nodeView({ node: node({ id, title: id }) })),
    ),
  );
  let settings: SettingsView = settingsView();

  const moves: string[] = [];
  const themes: string[] = [];
  const languages: string[] = [];
  const timerStarts: string[] = [];
  const refuse = { timerStart: false };
  let stops = 0;

  const client: Client = {
    ...base.client,
    Board: () => Promise.resolve(board),
    Settings: () => Promise.resolve(settings),
    MoveToColumn: (nodeID: string, target: string) => {
      moves.push(target);
      board = relocate(board, nodeID, target);
      return Promise.resolve(node());
    },
    TimerStart: (nodeID: string) => {
      timerStarts.push(nodeID);
      if (refuse.timerStart) {
        return Promise.reject(new Error('service: a project can never be doing'));
      }
      return Promise.resolve(timerView({ running: true }));
    },
    TimerStop: () => {
      stops += 1;
      return Promise.resolve(timerView());
    },
    SetTheme: (value: string) => {
      themes.push(value);
      settings = { ...settings, theme: value } as SettingsView;
      return Promise.resolve(settings);
    },
    SetLanguage: (value: string) => {
      languages.push(value);
      settings = { ...settings, language: value } as SettingsView;
      return Promise.resolve(settings);
    },
  };

  return {
    client,
    moves,
    themes,
    languages,
    timerStarts,
    refuse,
    stopCount: () => stops,
  };
}

/** Takes `nodeId` out of whatever column holds it and appends it to `target`. */
function relocate(board: ColumnView[], nodeId: string, target: string): ColumnView[] {
  let moving: NodeView | null = null;

  const without = board.map((column) =>
    columnView(
      column.status,
      column.nodes.filter((view) => {
        if (view.node.id !== nodeId) {
          return true;
        }
        moving = view;
        return false;
      }),
    ),
  );

  const found: NodeView | null = moving;
  if (found === null) {
    return board;
  }
  return without.map((column) =>
    column.status === target ? columnView(column.status, [...column.nodes, found]) : column,
  );
}

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

function storeOver(client: Client): AppStore {
  return createAppStore(client, { view: testWindow() });
}

/**
 * Renders the shell, waits for the board and puts focus on a card.
 *
 * The settings read is done FIRST, before the render, because that is what
 * main.tsx does: palette, theme, accent and language are pre-paint work, they
 * decide what the first frame looks like, and the shell deliberately does not
 * hydrate them a second time. A test that skipped it would be testing a state
 * the running app never opens in.
 */
async function enterTheBoard(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await store.getState().loadSettings();
  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('article');
  // Tab until a card has focus: region 1's appearance controls (S2-21) come
  // first in the shell's DOM order, and how many stops they add is not this
  // file's business.
  await tabUntil(user, () => document.activeElement?.hasAttribute('data-node-id') === true);

  return user;
}

/** Renders the shell and waits for the board WITHOUT selecting a card. */
async function enterTheApp(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await store.getState().loadSettings();
  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('region');

  return user;
}

/** The palette's rows, in the order it renders them. */
function rows(): HTMLElement[] {
  return screen.getAllByRole('option');
}

/** The row whose `data-command` starts with `prefix`, or undefined. */
function rowFor(prefix: string): HTMLElement | undefined {
  return rows().find((row) => row.getAttribute('data-command')?.startsWith(prefix));
}

/** Presses ArrowDown until the named row is the active one, then Enter. */
async function runRow(user: ReturnType<typeof userEvent.setup>, command: string) {
  for (let step = 0; step < rows().length; step += 1) {
    const active = rows().find((row) => row.getAttribute('aria-selected') === 'true');
    if (active?.getAttribute('data-command') === command) {
      await user.keyboard('{Enter}');
      return;
    }
    await user.keyboard('{ArrowDown}');
  }
  throw new Error(`no row ${command} was ever active`);
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('the command palette is in the application', () => {
  it('opens on Ctrl+K with focus in the filter field', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    expect(screen.queryByRole('dialog')).toBeNull();

    await user.keyboard(OPEN_PALETTE);

    const overlay = await screen.findByRole('dialog');
    expect(overlay).toHaveAttribute('aria-modal', 'true');
    expect(screen.getByRole('combobox')).toHaveFocus();
    expect(screen.getByRole('listbox')).toBeInTheDocument();
  });

  it('lets Ctrl+N and Ctrl+K each open their own overlay, one at a time', async () => {
    // Two overlays coexist in the shell and never on screen together:
    // store/ui.ts holds ONE open overlay, so opening either closes the other.
    const go = paletteGo();
    const store = storeOver(go.client);
    const user = await enterTheBoard(store);

    await user.keyboard(OPEN_QUICK_ADD);
    await screen.findByRole('dialog');
    expect(store.getState().openOverlay).toBe('quickAdd');
    expect(screen.queryByRole('combobox')).toBeNull();

    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    expect(store.getState().openOverlay).toBe('commandPalette');
    expect(screen.getAllByRole('dialog')).toHaveLength(1);
    expect(screen.getByRole('combobox')).toBeInTheDocument();

    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });

  it('closes on Escape and gives focus back to where it came from', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    const opener = document.activeElement;
    expect((opener as HTMLElement).dataset.nodeId).toBe('a');

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    expect(screen.getByRole('combobox')).toHaveFocus();

    await user.keyboard('{Escape}');

    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(document.activeElement).toBe(opener);
  });

  it('traps Tab inside itself', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    await user.tab();
    expect(screen.getByRole('combobox')).toHaveFocus();

    await user.tab({ shift: true });
    expect(screen.getByRole('combobox')).toHaveFocus();
  });

  it('starts empty every time it is reopened', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await user.keyboard('timer');
    expect(screen.getByRole('combobox')).toHaveValue('timer');

    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    expect(screen.getByRole('combobox')).toHaveValue('');
  });
});

describe('what the palette offers', () => {
  it('lists exactly the columns Go returned, in Go’s order', async () => {
    const unusual = ['col-5', 'col-2', 'col-4', 'col-1', 'col-3'];
    const go = paletteGo();
    const client: Client = {
      ...go.client,
      Board: () =>
        Promise.resolve(
          unusual.map((status, index) =>
            columnView(status, index === 0 ? [nodeView({ node: node({ id: 'a' }) })] : []),
          ),
        ),
    };

    const user = await enterTheBoard(storeOver(client));
    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const offered = rows()
      .map((row) => row.getAttribute('data-command'))
      .filter((id): id is string => id !== null && id.startsWith('move:'))
      .map((id) => id.slice('move:'.length));

    // Not a set comparison: the ORDER is the claim, and it is Go's.
    expect(offered).toEqual(unusual);
  });

  it('registers every action the brief names', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    for (const command of [
      'newTask',
      'move:',
      'priority:',
      'timerStart',
      'timerStop',
      'view:',
      'toggleTheme',
      'switchLanguage',
    ]) {
      expect(rowFor(command), `no row for ${command}`).toBeDefined();
    }
  });

  it('shows only hints that lib/keyboard.ts actually binds', async () => {
    // S2-16's map is normative and the palette reads it rather than restating
    // it. Every hint on screen must be a value in KEYS — a hint the map does
    // not bind is a hint that lies.
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const bound = new Set(Object.values(KEYS).map((chord) => chord.hint));

    const shown = rows()
      .map((row) => row.querySelector('.font-mono')?.textContent ?? null)
      .filter((hint): hint is string => hint !== null);

    expect(shown.length, 'no hint was rendered at all').toBeGreaterThan(0);
    for (const hint of shown) {
      expect(bound.has(hint), `${hint} is not in lib/keyboard.ts`).toBe(true);
    }

    // And the three the palette claims are there, by the map's own values.
    expect(shown).toContain(KEYS.quickAdd.hint);
    expect(shown).toContain(KEYS.moveRight.hint);
  });

  it('disables what needs a card when no card is focused, with a reason', async () => {
    const go = paletteGo();
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const start = rowFor('timerStart')!;
    expect(start).toHaveAttribute('aria-disabled', 'true');
    expect(start).toHaveTextContent('Focus a card first.');

    // Running it does nothing at all — no call, no toast.
    await runRow(user, 'timerStart');
    expect(go.timerStarts).toEqual([]);
    expect(screen.queryByRole('alert')).toBeNull();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('registers the priority and future-view rows as unavailable rather than dead', async () => {
    // TASKS.md S2-20: "a dead entry that silently does nothing is the worse
    // option". Neither group can act in Stage 2 — no binding sets a priority,
    // and Kanban is the only view — so both say so.
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const priority = rowFor('priority:')!;
    expect(priority).toHaveAttribute('aria-disabled', 'true');
    expect(priority).toHaveTextContent('stage 3');

    const kanban = rowFor('view:kanban')!;
    expect(kanban).toHaveAttribute('aria-disabled', 'true');
    expect(kanban).toHaveTextContent('You are looking at it.');

    const later = rows().find((row) => row.getAttribute('data-command') === 'view:tree')!;
    expect(later).toHaveAttribute('aria-disabled', 'true');
    expect(later).toHaveTextContent('stage 3');
  });
});

describe('running an action', () => {
  it('moves the focused card to the column the row names', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    await runRow(user, `move:${COLUMNS[4]}`);

    await waitFor(() => expect(go.moves).toEqual([COLUMNS[4]]));
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    const last = await screen.findByRole('region', { name: COLUMNS[4] });
    expect(within(last).getByRole('article')).toHaveTextContent('a');
  });

  it('opens the quick add, replacing itself', async () => {
    const go = paletteGo();
    const store = storeOver(go.client);
    const user = await enterTheBoard(store);

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    await runRow(user, 'newTask');

    await waitFor(() => expect(store.getState().openOverlay).toBe('quickAdd'));
    expect(screen.getAllByRole('dialog')).toHaveLength(1);
    expect(screen.queryByRole('combobox')).toBeNull();
    // Scoped to the overlay: S2-21's accent field is a textbox too, so "the
    // only textbox on the page" stopped meaning "the quick add's title".
    await waitFor(() =>
      expect(within(screen.getByRole('dialog')).getByRole('textbox')).toHaveFocus(),
    );
  });

  it('starts and stops the timer through the store', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await runRow(user, 'timerStart');

    await waitFor(() => expect(go.timerStarts).toEqual(['a']));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await runRow(user, 'timerStop');

    await waitFor(() => expect(go.stopCount()).toBe(1));
  });

  it('surfaces Go’s refusal of a timer and keeps the entry', async () => {
    // D9: a project is never timeable, and that rule is asked exactly once, in
    // Go. Hiding the row here would be DoingRefusal written a second time.
    const go = paletteGo();
    go.refuse.timerStart = true;
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await runRow(user, 'timerStart');

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');
    expect(alert).not.toHaveTextContent('a project can never be doing');

    // Still offered, and still not disabled: nothing here knows what Go
    // refuses.
    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    expect(rowFor('timerStart')).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('flips the theme through SetTheme', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await user.keyboard('theme');
    await runRow(user, 'toggleTheme');

    await waitFor(() => expect(go.themes).toHaveLength(1));
    expect(go.themes[0]).not.toBe(settingsView().theme);

    // And the document followed the SERVICE'S answer.
    await waitFor(() =>
      expect(document.documentElement.classList.contains('dark')).toBe(false),
    );
  });

  it('switches the language through SetLanguage and i18next', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await user.keyboard('language');
    await runRow(user, 'switchLanguage');

    await waitFor(() => expect(go.languages).toEqual(['ru']));
    // The UI really moved, and it moved to the language the SERVICE reported.
    await waitFor(() => expect(screen.getByRole('main')).toHaveAttribute('aria-label', 'Доска'));
  });
});

describe('filtering', () => {
  it('narrows the list as the user types and runs what is left', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const all = rows().length;
    await user.keyboard('timer');

    await waitFor(() => expect(rows().length).toBeLessThan(all));
    for (const row of rows()) {
      expect(row.textContent?.toLowerCase()).toContain('timer');
    }
  });

  it('says so when nothing matches, rather than showing an empty box', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');
    await user.keyboard('zzzzzz');

    await waitFor(() => expect(screen.queryByRole('listbox')).toBeNull());
    expect(screen.getByRole('dialog')).toHaveTextContent('Nothing matches that.');

    // Enter on nothing does nothing, and does not crash.
    await user.keyboard('{Enter}');
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  it('finds the RUSSIAN labels when the UI is Russian', async () => {
    // A palette that only answers to English words is broken in Russian, and
    // the filter matching the TRANSLATED label is what prevents it.
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client), 'ru');

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    await user.keyboard('таймер');

    await waitFor(() => expect(rows().length).toBeGreaterThan(0));
    for (const row of rows()) {
      expect(row.textContent?.toLowerCase()).toContain('таймер');
    }
    expect(rowFor('timerStart')).toBeDefined();

    // And the English word finds nothing, which is the other half of the claim.
    await user.clear(screen.getByRole('combobox'));
    await user.keyboard('timer');
    await waitFor(() => expect(screen.queryByRole('listbox')).toBeNull());
  });

  it.each(BOTH_LANGUAGES)('renders in %s with nothing hard-coded', async (language) => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client), language);

    await user.keyboard(OPEN_PALETTE);
    const overlay = await screen.findByRole('dialog');

    expect(overlay.textContent).not.toContain('palette.');
    expect(screen.getByRole('combobox').getAttribute('aria-label')).not.toContain('palette.');

    // Nothing is held to one line: Russian is about a third wider and every
    // row wraps rather than truncating.
    for (const row of rows()) {
      expect(row).toHaveClass('flex-wrap');
    }
  });
});

describe('moving the active row', () => {
  it('walks the list with the arrows and clamps at both ends', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    const selected = () => rows().findIndex((row) => row.getAttribute('aria-selected') === 'true');

    expect(selected()).toBe(0);

    await user.keyboard('{ArrowUp}');
    expect(selected()).toBe(0);

    await user.keyboard('{ArrowDown}{ArrowDown}');
    expect(selected()).toBe(2);

    await user.keyboard('{ArrowUp}');
    expect(selected()).toBe(1);

    for (let step = 0; step < rows().length + 2; step += 1) {
      await user.keyboard('{ArrowDown}');
    }
    expect(selected()).toBe(rows().length - 1);
  });

  it('publishes the active row through aria-activedescendant', async () => {
    const go = paletteGo();
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(OPEN_PALETTE);
    await screen.findByRole('dialog');

    await user.keyboard('{ArrowDown}');

    const active = rows().find((row) => row.getAttribute('aria-selected') === 'true')!;
    expect(screen.getByRole('combobox')).toHaveAttribute('aria-activedescendant', active.id);
  });
});
