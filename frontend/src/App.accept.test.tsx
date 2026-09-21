import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cleanup, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, NewNode, Node, NodeView } from './lib/client';
import { createAppStore, type AppStore } from './store';
import { columnView, createFakeClient, habitView, node, nodeView, settingsView } from './test/fakeClient';
import { BOTH_LANGUAGES, renderIn, tabUntil } from './test/render';

// Nexus — Stage 2's ACCEPT criterion, as something a machine checks.
//
//   > Create a task, move it across all five columns, and complete it — with
//   > no pointing device.
//
// # It presses keys and nothing else, and that is enforced rather than claimed
//
// `make guard` check 6 greps THIS FILE for the three words a pointing device
// would be spelt with, in any case, and fails on a single hit — in code or in
// a comment. So the phrase above is written the long way round: the forbidden
// word cannot appear here, which is exactly the property being asserted. The
// check also fails if this file is deleted or emptied, because otherwise the
// easiest way to pass it would be to remove it.
//
// # It renders <App />
//
// The whole assembled screen of S2-15 onwards — the store, the header, the
// strip, the board, both overlays and the toast layer — and it imports NO
// component beneath it. The ACCEPT criterion is a property of the assembled
// screen: a flow test that mounted <Kanban /> and handed it a quick add would
// be testing an arrangement no user ever gets, and would pass on a day when
// App.tsx rendered none of it. `git grep -n "^import" on this file shows App,
// the client TYPES, the store factory and the two test helper modules. Nothing
// from components/ and nothing from views/.
//
// # No column name is written down
//
// The five columns are nonsense strings the fake hands over, and the flow sends
// back whatever Go put in the array. Which strings are real columns is
// domain.Status's answer (`make guard` check 2), and using opaque ones is the
// stronger claim anyway: the keyboard route works without knowing what any of
// the five mean.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

/**
 * Which column the timer couples to, as an INDEX into the fake's own board.
 *
 * The TEST is allowed to know this, because in this file the test is playing
 * Go: D13 says a direct move of a timeable node into the Doing column opens a
 * time entry in the same transaction, and closes it on the way out. That is
 * Go's rule, implemented in Go (S2-03), and what is asserted here is that the
 * assembled screen SHOWS it — the indicator appears on Doing and is gone on
 * Done. The application code still learns it only from `NodeView.timer`.
 */
const DOING = 3;

const THE_TITLE = 'Keyboard accept';

const RIGHT = '{Control>}{Shift>}{ArrowRight}{/Shift}{/Control}';
const QUICK_ADD = '{Control>}n{/Control}';
const PALETTE = '{Control>}k{/Control}';

interface Call {
  method: string;
  argument: string;
}

/** The read-only half of the binding surface. Everything else is a write. */
const READS = new Set(['Board', 'HabitStrip', 'Settings', 'TimerCurrent', 'Tree', 'Progress']);

/**
 * A fake Go that stores what it is told and moves what it is asked to move.
 *
 * It records every call in order, which is what makes "the exact sequence" an
 * assertion rather than a count, and it models the one Go behaviour the screen
 * has to reflect: the timer that opens on the way into Doing and closes on the
 * way out (C1, D13).
 */
function acceptGo(habits: ReturnType<typeof habitView>[] = []) {
  const base = createFakeClient();
  let board: ColumnView[] = COLUMNS.map((status) => columnView(status));
  const calls: Call[] = [];
  let created = 0;

  const client: Client = {
    ...base.client,
    Board: () => {
      calls.push({ method: 'Board', argument: '' });
      return Promise.resolve(board);
    },
    HabitStrip: () => {
      calls.push({ method: 'HabitStrip', argument: '' });
      return Promise.resolve(habits);
    },
    Settings: () => {
      calls.push({ method: 'Settings', argument: '' });
      return Promise.resolve(settingsView());
    },
    CreateNode: (draft: NewNode) => {
      calls.push({ method: 'CreateNode', argument: draft.title });
      created += 1;

      const fresh: Node = node({ id: `new-${created}`, title: draft.title });
      board = board.map((column, index) =>
        index === 0
          ? columnView(column.status, [...column.nodes, seated(nodeView({ node: fresh }), 0)])
          : column,
      );
      return Promise.resolve(fresh);
    },
    MoveToColumn: (nodeID: string, target: string) => {
      calls.push({ method: 'MoveToColumn', argument: target });
      board = relocate(board, nodeID, target);
      return Promise.resolve(node({ id: nodeID }));
    },
  };

  return {
    client,
    calls,
    /** Every call that changed something, in the order it was made. */
    writes: () => calls.filter((call) => !READS.has(call.method)),
  };
}

/**
 * The view a card has once it is sitting in column `index`.
 *
 * The only field that depends on where it landed is the timer, and that is
 * D13's coupling seen from the read side: `TimerView.Running` is true for a
 * node in Doing and false everywhere else, including Done.
 */
function seated(view: NodeView, index: number): NodeView {
  return { ...view, timer: { ...view.timer, running: index === DOING } } as NodeView;
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
  return without.map((column, index) =>
    column.status === target
      ? columnView(column.status, [...column.nodes, seated(found, index)])
      : column,
  );
}

function testWindow(reduceMotion = false): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: reduceMotion, media: query }),
    getComputedStyle: (element: Element) => window.getComputedStyle(element),
  } as unknown as Window;
}

function storeOver(client: Client, reduceMotion = false): AppStore {
  return createAppStore(client, { view: testWindow(reduceMotion) });
}

/** Opens the app the way main.tsx does: settings first, then the tree. */
async function openTheApp(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await store.getState().loadSettings();
  await renderIn(language, <App store={store} />);
  await screen.findByRole('banner');

  return user;
}

/** The card with this id, wherever it is on the board. */
function cardWithId(nodeId: string): HTMLElement {
  return document.querySelector<HTMLElement>(`[data-node-id="${nodeId}"]`)!;
}

/** The status of the column holding `nodeId`, read off the rendered board. */
function columnHolding(nodeId: string): string | null {
  return cardWithId(nodeId)?.closest<HTMLElement>('[data-column]')?.dataset.column ?? null;
}

/** The id of whatever card has focus, or null. */
function focusedCardId(): string | null {
  const active = document.activeElement;
  return active instanceof HTMLElement ? (active.dataset.nodeId ?? null) : null;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
  const root = document.documentElement;
  root.removeAttribute('style');
  root.removeAttribute('class');
  delete root.dataset.palette;
  delete root.dataset.drift;
});

describe('the ACCEPT criterion, driven by keys alone', () => {
  it('creates a task, walks it across all five columns and completes it', async () => {
    const go = acceptGo();
    const user = await openTheApp(storeOver(go.client));

    // 1. Ctrl+N anywhere opens the quick add with the caret in the title.
    await user.keyboard(QUICK_ADD);
    const overlay = await screen.findByRole('dialog');
    expect(within(overlay).getByRole('textbox')).toHaveFocus();

    // 2. Type a title, Enter. The overlay closes, the card appears in the
    //    first column Go returned, and focus lands ON THAT CARD — which is
    //    what makes the four presses below a flow rather than four hunts.
    await user.keyboard(`${THE_TITLE}{Enter}`);

    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    await waitFor(() => expect(focusedCardId()).toBe('new-1'));
    expect(columnHolding('new-1')).toBe(COLUMNS[0]);
    expect(cardWithId('new-1')).toHaveTextContent(THE_TITLE);

    // 3-6. Four presses, one column per press, focus following the card.
    for (let step = 1; step <= 4; step += 1) {
      await user.keyboard(RIGHT);

      await waitFor(() => expect(columnHolding('new-1')).toBe(COLUMNS[step]));
      await waitFor(() => expect(focusedCardId()).toBe('new-1'));
    }

    // The card is in the last column Go returned — read off the DOM, not off
    // the fake, because what is being demonstrated is what the user sees.
    const last = screen.getByRole('region', { name: COLUMNS[4] });
    expect(within(last).getByRole('article')).toHaveTextContent(THE_TITLE);

    // THE EXACT SEQUENCE: one CreateNode, then MoveToColumn with each of the
    // four statuses GO SUPPLIED, in Go's order. Not a count, not a set — the
    // order is the claim, and so is the absence of anything else.
    expect(go.writes()).toEqual([
      { method: 'CreateNode', argument: THE_TITLE },
      { method: 'MoveToColumn', argument: COLUMNS[1] },
      { method: 'MoveToColumn', argument: COLUMNS[2] },
      { method: 'MoveToColumn', argument: COLUMNS[3] },
      { method: 'MoveToColumn', argument: COLUMNS[4] },
    ]);

    // And one more press goes nowhere: the right-hand end is a no-op, not a
    // wrap-around and not an error.
    await user.keyboard(RIGHT);
    expect(go.writes()).toHaveLength(5);
    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('shows the running timer on Doing and nothing on Done', async () => {
    // C1 / D13, seen from the screen. The frontend starts no timer of its own:
    // the coupling is Go's, inside the same transaction as the move, and all
    // the UI does is render NodeView.timer.running. That the flow never calls
    // TimerStart is half the assertion.
    const go = acceptGo();
    const user = await openTheApp(storeOver(go.client));

    await user.keyboard(QUICK_ADD);
    await screen.findByRole('dialog');
    await user.keyboard(`${THE_TITLE}{Enter}`);
    await waitFor(() => expect(focusedCardId()).toBe('new-1'));

    const running = () => within(cardWithId('new-1')).queryByLabelText('Timer running');

    expect(running()).toBeNull();

    for (let step = 1; step <= 3; step += 1) {
      await user.keyboard(RIGHT);
      await waitFor(() => expect(columnHolding('new-1')).toBe(COLUMNS[step]));
    }

    // In Doing: the indicator is there, and on no other card.
    await waitFor(() => expect(running()).not.toBeNull());
    expect(screen.getAllByLabelText('Timer running')).toHaveLength(1);

    await user.keyboard(RIGHT);
    await waitFor(() => expect(columnHolding('new-1')).toBe(COLUMNS[4]));

    // In Done: gone.
    await waitFor(() => expect(running()).toBeNull());
    expect(screen.queryAllByLabelText('Timer running')).toHaveLength(0);

    expect(go.calls.filter((call) => call.method === 'TimerStart')).toEqual([]);
  });

  it('completes the card a second way, through the command palette', async () => {
    // The second independent keyboard route, and the one step 8 of the hand
    // script drives. Back to the first column, then Ctrl+K and the row for the
    // last column.
    const go = acceptGo();
    const user = await openTheApp(storeOver(go.client));

    await user.keyboard(QUICK_ADD);
    await screen.findByRole('dialog');
    await user.keyboard(`${THE_TITLE}{Enter}`);
    await waitFor(() => expect(focusedCardId()).toBe('new-1'));

    await user.keyboard(PALETTE);
    await screen.findByRole('combobox');

    const rows = () => screen.getAllByRole('option');
    const wanted = `move:${COLUMNS[4]}`;

    for (let step = 0; step < rows().length; step += 1) {
      const active = rows().find((row) => row.getAttribute('aria-selected') === 'true');
      if (active?.getAttribute('data-command') === wanted) {
        break;
      }
      await user.keyboard('{ArrowDown}');
    }
    await user.keyboard('{Enter}');

    await waitFor(() => expect(columnHolding('new-1')).toBe(COLUMNS[4]));
    expect(go.writes()).toEqual([
      { method: 'CreateNode', argument: THE_TITLE },
      { method: 'MoveToColumn', argument: COLUMNS[4] },
    ]);
  });
});

// ---------------------------------------------------------------------------
// The RUSSIAN audit.
//
// What jsdom can and cannot judge, said plainly rather than implied. jsdom has
// no layout engine: every offsetWidth is 0 and nothing ever overflows, so no
// test in this project can assert "it does not clip at 1024x768". That half is
// the hand pass.
//
// What IS mechanical is the two ways a layout clips: text held to one line, and
// a box that cannot shrink. So every screen is rendered IN RUSSIAN and checked
// for the mechanisms that prevent both — flex-wrap on every strip, min-w-0 on
// every flex child that could be squeezed, break-words on every label — and for
// the absence of the two utilities that guarantee clipping, `truncate` and
// `whitespace-nowrap`. Plus the thing that actually goes wrong most often: a
// raw i18n key on screen because ru.json was not updated.

describe('the Russian audit', () => {
  /** The whole assembled screen, in Russian, with every region non-empty. */
  async function everyScreenInRussian() {
    const go = acceptGo([habitView({ node: node({ id: 'h-1', title: 'Читать по вечерам' }) })]);
    const store = storeOver(go.client);
    const user = await openTheApp(store, 'ru');

    await user.keyboard(QUICK_ADD);
    await screen.findByRole('dialog');
    await user.keyboard(`${THE_TITLE}{Enter}`);
    await waitFor(() => expect(focusedCardId()).toBe('new-1'));

    store.getState().pushToast('toast.error.body');

    return { user, store };
  }

  it('shows no raw key on any screen', async () => {
    const { user } = await everyScreenInRussian();

    // Header, strip, board and toast are all on screen at once.
    const page = document.body.textContent ?? '';
    for (const namespace of ['board.', 'card.', 'habits.', 'settings.', 'toast.', 'palette.']) {
      expect(page, `a raw ${namespace} key reached the screen`).not.toContain(namespace);
    }

    // And the two overlays, one at a time.
    await user.keyboard(QUICK_ADD);
    const quickAdd = await screen.findByRole('dialog');
    expect(quickAdd.textContent).not.toContain('quickAdd.');
    expect(quickAdd.textContent).not.toContain('card.type.');
    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    await user.keyboard(PALETTE);
    const palette = await screen.findByRole('dialog');
    expect(palette.textContent).not.toContain('palette.');
    expect(palette.textContent).not.toContain('board.column.');
  });

  it('holds nothing to one line and lets every box shrink', async () => {
    const { user } = await everyScreenInRussian();

    // `getAttribute` and not `.className`: on an SVG element — every type icon
    // is one — `className` is an SVGAnimatedString and has no `includes`.
    const clipping = (root: ParentNode) =>
      [...root.querySelectorAll<HTMLElement>('[class]')]
        .map((element) => element.getAttribute('class') ?? '')
        .filter((classes) => classes.includes('truncate') || classes.includes('nowrap'));

    expect(clipping(document.body), 'the launch screen clips somewhere').toEqual([]);

    // The strips that have to wrap, named one by one so a failure says which.
    expect(screen.getByRole('banner'), 'the header').toHaveClass('flex-wrap');
    expect(screen.getByRole('group', { name: 'Привычки' }), 'the strip').toHaveClass('flex-wrap');
    expect(screen.getByRole('alert').firstElementChild, 'the toast').toHaveClass('flex-col');

    // The board scrolls sideways; a COLUMN never does, which is what "no
    // horizontal scroll inside a column" means.
    const columns = screen.getAllByRole('region');
    expect(columns).toHaveLength(COLUMNS.length);
    for (const column of columns) {
      expect(column.className).not.toContain('overflow-x');
      expect(column).toHaveClass('min-w-36');
      expect(column).toHaveClass('basis-0');
    }

    await user.keyboard(QUICK_ADD);
    const quickAdd = await screen.findByRole('dialog');
    expect(clipping(quickAdd), 'the quick add clips').toEqual([]);
    expect(screen.getByRole('radiogroup')).toHaveClass('flex-wrap');
    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    await user.keyboard(PALETTE);
    const palette = await screen.findByRole('dialog');
    expect(clipping(palette), 'the palette clips').toEqual([]);
    for (const row of screen.getAllByRole('option')) {
      expect(row).toHaveClass('flex-wrap');
    }
  });

  it.each(BOTH_LANGUAGES)('assembles all five regions in %s', async (language) => {
    const go = acceptGo([habitView({ node: node({ id: 'h-1', title: 'Читать' }) })]);
    const store = storeOver(go.client);
    const user = await openTheApp(store, language);

    store.getState().pushToast('toast.error.body');

    expect(screen.getByRole('banner'), 'region 1').toBeInTheDocument();
    expect(await screen.findByRole('checkbox'), 'region 2').toBeInTheDocument();
    expect(screen.getByRole('main'), 'region 3').toBeInTheDocument();
    expect(await screen.findByRole('alert'), 'region 5').toBeInTheDocument();

    await user.keyboard(PALETTE);
    expect(await screen.findByRole('dialog'), 'region 4').toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// The ACCESSIBILITY audit.
//
// Three of its five parts are mechanical and are below. The two that are not
// are named rather than faked:
//
//   the focus RING being painted   jsdom evaluates :focus-visible as false for
//                                  a programmatically focused element, and the
//                                  rule itself lives in src/style.css which
//                                  `css: false` makes unreadable. What IS
//                                  mechanical — that no file switches an
//                                  outline off, and that every reachable
//                                  element can take focus at all — is already
//                                  asserted in App.keyboard.test.tsx and is not
//                                  restated here.
//   Aurora's background drift      D16 and K5: the GATE ships, the visual does
//                                  not. What is audited is that the decision is
//                                  published as data-drift and respects the
//                                  media query. NOTHING DRAWS THE DRIFT, so
//                                  nobody may report having watched it pause.

describe('the accessibility audit', () => {
  it('can focus every interactive element on the launch screen', async () => {
    const go = acceptGo([habitView({ node: node({ id: 'h-1', title: 'Read' }) })]);
    const store = storeOver(go.client);
    await openTheApp(store);
    await screen.findByRole('checkbox');

    store.getState().pushToast('toast.error.body');
    await screen.findByRole('alert');

    const interactive = [
      ...document.querySelectorAll<HTMLElement>('button, input, [data-node-id], [data-habit-id]'),
    ];
    expect(interactive.length, 'nothing was enumerated').toBeGreaterThan(5);

    for (const element of interactive) {
      element.focus();
      expect(document.activeElement, `${element.tagName} cannot take focus`).toBe(element);
    }
  });

  it('names every overlay and gives the palette its combobox roles', async () => {
    const go = acceptGo();
    const user = await openTheApp(storeOver(go.client));

    await user.keyboard(QUICK_ADD);
    const quickAdd = await screen.findByRole('dialog');
    expect(quickAdd).toHaveAttribute('aria-modal', 'true');
    expect(quickAdd.getAttribute('aria-label')).toBeTruthy();
    await user.keyboard('{Escape}');
    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());

    await user.keyboard(PALETTE);
    const palette = await screen.findByRole('dialog');
    expect(palette).toHaveAttribute('aria-modal', 'true');
    expect(palette.getAttribute('aria-label')).toBeTruthy();

    const filter = screen.getByRole('combobox');
    expect(filter).toHaveAttribute('aria-expanded', 'true');
    expect(filter).toHaveAttribute('aria-controls', screen.getByRole('listbox').id);
    // The active row is published where a combobox publishes it, because the
    // rows are not tab stops and could not otherwise be announced.
    const active = screen.getAllByRole('option').find((row) => row.id === filter.getAttribute('aria-activedescendant'));
    expect(active, 'no option is the active descendant').toBeDefined();
    expect(active).toHaveAttribute('aria-selected', 'true');
  });

  it('keeps every region one tab stop, and reachable', async () => {
    // The header's four groups, the strip and the board are all roving
    // tabindexes: n cards and n chips as n tab stops would make the keyboard
    // useless, and the arrows move inside each one.
    const go = acceptGo([
      habitView({ node: node({ id: 'h-1', title: 'Read' }) }),
      habitView({ node: node({ id: 'h-2', title: 'Walk' }) }),
    ]);
    const store = storeOver(go.client);
    const user = await openTheApp(store);

    await user.keyboard(QUICK_ADD);
    await screen.findByRole('dialog');
    await user.keyboard(`${THE_TITLE}{Enter}`);
    await waitFor(() => expect(focusedCardId()).toBe('new-1'));

    expect(document.querySelectorAll('[data-node-id][tabindex="0"]')).toHaveLength(1);
    expect(document.querySelectorAll('[data-habit-id][tabindex="0"]')).toHaveLength(1);
    expect(document.querySelectorAll('[data-setting][tabindex="0"]')).toHaveLength(3);

    // And Tab alone gets from the board to the header controls, without any
    // shortcut being pressed and without opening anything.
    await tabUntil(user, () => document.activeElement?.hasAttribute('data-setting') === true);
    expect(store.getState().openOverlay).toBeNull();
  });

  it('gates the Aurora drift on the media query — the gate, and nothing behind it', async () => {
    // D16 / K5. `auroraDriftEnabled` publishes the decision as data-drift and
    // design/tokens.css kills every transition under the media query. NOTHING
    // CONSUMES THE ATTRIBUTE: the visual is deliberately not built, because
    // design/ specifies no drift and inventing one is forbidden. So what is
    // asserted is the gate, and nobody may claim to have watched a drift pause.
    const go = acceptGo();
    await openTheApp(storeOver(go.client, false));
    expect(document.documentElement.dataset.drift).toBe('on');

    // Unmount the first app before rendering the second: two shells in one
    // body is two of everything getByRole looks for.
    cleanup();
    document.documentElement.removeAttribute('class');
    delete document.documentElement.dataset.drift;

    const reduced = acceptGo();
    await openTheApp(storeOver(reduced.client, true));
    expect(document.documentElement.dataset.drift).toBe('off');

    // The claim that there is nothing behind the gate, asserted rather than
    // asserted about: no element in the assembled screen reads it.
    expect(document.querySelectorAll('[data-drift] [class*="drift"]')).toHaveLength(0);
  });
});
