import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, HabitView, NodeView } from './lib/client';
import { KEYS } from './lib/keyboard';
import { createAppStore, type AppStore } from './store';
import { columnView, createFakeClient, habitView, node, nodeView } from './test/fakeClient';
import { renderIn, tabUntil } from './test/render';

// Nexus — drag and drop, asserted through the shell (S2-17).
//
// This file imports `App` and does NOT import Kanban, Column or Card. The
// import restriction is the assertion, exactly as in App.test.tsx and
// App.keyboard.test.tsx: importing the component you are looking for turns the
// test into a test of that component and proves nothing about the application.
//
// # The geometry, and why it has to be supplied
//
// jsdom performs no layout: every `getBoundingClientRect` is 0×0 at the origin.
// dnd-kit decides what a card is over by comparing RECTANGLES, so on an unstubbed
// jsdom every droppable is the same point and "which column did it land on" has
// no answer at all. `layOutTheBoard` below gives the board the geometry a browser
// would have measured — three columns side by side, cards stacked inside them —
// and every coordinate the drags use is computed from it rather than written down.
// That is the one piece of the environment these tests fake; the events
// themselves are real PointerEvents through the real React event pipeline, and
// the sensor, the collision detection and the store are the shipping ones.
//
// What this CANNOT show is stated where it belongs, in the commit body: that a
// physical pointer in a real WebKit window picks a card up. There is no display
// on this machine and no xvfb, so that half is a hand check.
//
// # No column name is written down
//
// The mocked statuses are nonsense strings — `make guard` check 2 — and using
// opaque ones is the stronger claim anyway: the board drags cards between
// columns while knowing nothing about any of them.

const COLUMNS = ['col-1', 'col-2', 'col-3'];

function card(id: string, children: NodeView[] = []): NodeView {
  return nodeView({ node: node({ id, title: id }), children });
}

interface FakeGo {
  client: Client;
  /** Every `MoveToColumn(nodeID, target)` in order. */
  toColumn: Array<{ nodeID: string; target: string }>;
  /** Every `MoveNode(nodeID, newParentID, toIndex)` in order. */
  reorder: Array<{ nodeID: string; parent: string; index: number }>;
  /** How many times the board has been read. */
  reads: () => number;
  /** How many times a habit was checked. */
  habitChecks: () => number;
  /** Set before a drag to make Go refuse the move. */
  refuse: { move: boolean };
  /** Set before a drag to make Go take its time; `release` finishes the call. */
  hold: { move: boolean };
  release: () => void;
}

/**
 * A fake Go that really moves the card, and records what it was asked for.
 *
 * Local to this file rather than added to test/fakeClient.ts for the reason
 * App.keyboard.test.tsx gives about its own: what is under test is a SEQUENCE —
 * pick up, hover, drop, re-read — and a fake that answered identically every
 * time would let the test pass while the real board went nowhere.
 */
function movingGo(seed: Record<number, NodeView[]> = {}, habits: HabitView[] = []): FakeGo {
  const base = createFakeClient({ habits });
  let board: ColumnView[] = COLUMNS.map((status, index) => columnView(status, seed[index] ?? []));

  const toColumn: FakeGo['toColumn'] = [];
  const reorder: FakeGo['reorder'] = [];
  const refuse = { move: false };
  const hold = { move: false };
  let reads = 0;
  let release = () => {};

  const client: Client = {
    ...base.client,
    Board: () => {
      reads += 1;
      return Promise.resolve(board);
    },
    MoveToColumn: (nodeID: string, target: string) => {
      toColumn.push({ nodeID, target });
      if (refuse.move) {
        return Promise.reject(new Error('service: a project can never be doing'));
      }

      const settle = () => {
        board = relocate(board, nodeID, target);
      };
      if (!hold.move) {
        settle();
        return Promise.resolve(node());
      }
      return new Promise((resolve) => {
        release = () => {
          settle();
          resolve(node());
        };
      });
    },
    MoveNode: (nodeID: string, parent: string, index: number) => {
      reorder.push({ nodeID, parent, index });
      return Promise.resolve(node());
    },
  };

  return {
    client,
    toColumn,
    reorder,
    refuse,
    hold,
    reads: () => reads,
    habitChecks: () => base.calls.CheckHabitToday ?? 0,
    release: () => release(),
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

function testWindow(reducedMotion = false): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: reducedMotion, media: query }),
  } as unknown as Window;
}

function storeOver(client: Client, reducedMotion = false): AppStore {
  return createAppStore(client, { view: testWindow(reducedMotion) });
}

async function showTheBoard(go: FakeGo, language = 'en', reducedMotion = false) {
  const store = storeOver(go.client, reducedMotion);

  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('article');

  return store;
}

// --------------------------------------------------------------- the geometry

const COLUMN_WIDTH = 200;
const COLUMN_HEIGHT = 400;
const HEADING_HEIGHT = 40;
const CARD_HEIGHT = 60;

function box(x: number, y: number, width: number, height: number): DOMRect {
  return {
    x,
    y,
    left: x,
    top: y,
    right: x + width,
    bottom: y + height,
    width,
    height,
    toJSON: () => ({}),
  } as DOMRect;
}

/** Three columns across, cards stacked down each — the layout jsdom will not do. */
function rectOf(element: Element): DOMRect {
  const section = element.closest('[data-column]');
  if (section === null) {
    // The DragOverlay (S3-05, D18) draws the active card OUTSIDE every column,
    // which is the whole point of it — so it has no column to take a position
    // from. A real browser measures it all the same, and what it measures is a
    // box the size and place of the card that was lifted. dnd-kit uses that rect
    // for collision detection from the moment the overlay mounts, so a fake that
    // returned 0x0 here would put the drag at the origin and no column would
    // ever be "over".
    //
    // This extends the GEOMETRY fake, which the note at the top of this file
    // already names as the one piece of the environment these tests supply. No
    // assertion in this file changed for S3-05.
    const lifted = element.closest<HTMLElement>('[data-node-id]');
    const source =
      lifted === null
        ? null
        : document.querySelector(`[data-column] [data-node-id="${lifted.dataset.nodeId}"]`);

    return source === null ? box(0, 0, 0, 0) : rectOf(source);
  }

  const columns = [...document.querySelectorAll('[data-column]')];
  const x = columns.indexOf(section) * COLUMN_WIDTH;

  const item = element.closest('li');
  if (item === null || item.parentElement === null) {
    return box(x, 0, COLUMN_WIDTH, COLUMN_HEIGHT);
  }

  const row = [...item.parentElement.children].indexOf(item);
  return box(x + 10, HEADING_HEIGHT + row * CARD_HEIGHT, COLUMN_WIDTH - 20, CARD_HEIGHT - 10);
}

function layOutTheBoard() {
  vi.spyOn(Element.prototype, 'getBoundingClientRect').mockImplementation(function (this: Element) {
    return rectOf(this);
  });
}

function centreOf(element: Element) {
  const rect = rectOf(element);
  return { clientX: rect.left + rect.width / 2, clientY: rect.top + rect.height / 2 };
}

// --------------------------------------------------------------- the gestures

function cardNamed(id: string): HTMLElement {
  return document.querySelector<HTMLElement>(`[data-node-id="${id}"]`)!;
}

function columnNamed(status: string): HTMLElement {
  return screen.getByRole('region', { name: status });
}

/**
 * Presses on a card and moves past the activation distance.
 *
 * The press lands on the `<article>`, which carries no drag handler of its own —
 * the listeners are on the `<li>` around it, and the event gets there by
 * bubbling. That is deliberate: it is the same path a real pointer takes, and it
 * is why the card keeps its tabindex and its data-node-id and nothing else.
 */
function pickUp(id: string) {
  const handle = cardNamed(id);
  const at = centreOf(handle);

  fireEvent.pointerDown(handle, { ...at, isPrimary: true, button: 0, buttons: 1, pointerId: 1 });
  fireEvent.pointerMove(document, {
    clientX: at.clientX + 20,
    clientY: at.clientY,
    isPrimary: true,
    pointerId: 1,
  });
}

function moveOver(target: Element) {
  fireEvent.pointerMove(document, { ...centreOf(target), isPrimary: true, pointerId: 1 });
}

function letGo(target: Element) {
  fireEvent.pointerUp(document, { ...centreOf(target), isPrimary: true, pointerId: 1 });
}

/** The whole gesture: pick the card up, carry it to `target`, let go. */
function dragOnto(id: string, target: Element) {
  pickUp(id);
  moveOver(target);
  letGo(target);
}

/** The titles in a column, top to bottom, read off the screen. */
function titlesIn(status: string): string[] {
  return within(columnNamed(status))
    .queryAllByRole('article')
    .map((article) => article.dataset.nodeId ?? '');
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
  layOutTheBoard();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('dragging a card to another column', () => {
  it('moves it there before the call resolves, then re-reads', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')], 1: [card('b1')] });
    go.hold.move = true;
    await showTheBoard(go);

    const readsBefore = go.reads();
    dragOnto('a1', cardNamed('b1'));

    // Optimistic: the card is under the other column's heading while Go is
    // still thinking about it.
    await waitFor(() => expect(titlesIn(COLUMNS[1])).toContain('a1'));
    expect(titlesIn(COLUMNS[0])).toEqual(['a2']);
    expect(go.toColumn).toEqual([{ nodeID: 'a1', target: COLUMNS[1] }]);
    expect(go.reads()).toBe(readsBefore);
    expect(screen.queryByRole('alert')).toBeNull();

    // And then Go answers, and the board is re-read rather than left as guessed.
    go.release();
    await waitFor(() => expect(go.reads()).toBe(readsBefore + 1));
    expect(titlesIn(COLUMNS[1])).toEqual(['b1', 'a1']);
  });

  it('calls MoveToColumn once, with the status Go supplied', async () => {
    const go = movingGo({ 0: [card('a1')], 2: [card('c1')] });
    await showTheBoard(go);

    dragOnto('a1', cardNamed('c1'));

    await waitFor(() => expect(go.toColumn).toEqual([{ nodeID: 'a1', target: COLUMNS[2] }]));
    expect(go.reorder).toEqual([]);
  });

  it('lands on an empty column', async () => {
    const go = movingGo({ 0: [card('a1')] });
    await showTheBoard(go);

    dragOnto('a1', columnNamed(COLUMNS[2]));

    await waitFor(() => expect(go.toColumn).toEqual([{ nodeID: 'a1', target: COLUMNS[2] }]));
  });

  it('sends the parent id and nothing else — the subtree is Go‘s to move', async () => {
    // MoveToColumn moves parent_id and sort_order and leaves every descendant's
    // own parent_id and status alone (domain.PlanMove). So the call names ONE
    // node, and there is no walk of `view.children` anywhere in the frontend to
    // get it wrong — `make guard` check 3 is the mechanical half of this.
    const go = movingGo({
      0: [card('parent', [card('child-1'), card('child-2')])],
      1: [card('b1')],
    });
    await showTheBoard(go);

    dragOnto('parent', cardNamed('b1'));

    await waitFor(() => expect(go.toColumn).toHaveLength(1));
    expect(go.toColumn).toEqual([{ nodeID: 'parent', target: COLUMNS[1] }]);
    expect(go.reorder).toEqual([]);
  });
});

describe('a drop Go refuses', () => {
  it('puts the card back, re-reads, and raises exactly one toast', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')], 1: [card('b1')] });
    go.refuse.move = true;
    await showTheBoard(go);

    const readsBefore = go.reads();
    dragOnto('a1', cardNamed('b1'));

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');
    expect(alert).not.toHaveTextContent('a project can never be doing');

    // The rollback is a RE-READ, not the array captured before the call: the
    // board came back from Go, and the card is where Go says it is.
    await waitFor(() => expect(go.reads()).toBe(readsBefore + 1));
    await waitFor(() => expect(titlesIn(COLUMNS[0])).toEqual(['a1', 'a2']));
    expect(titlesIn(COLUMNS[1])).toEqual(['b1']);
  });
});

describe('reordering inside one column', () => {
  it('calls MoveNode and never MoveToColumn', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2'), card('a3')] });
    await showTheBoard(go);

    dragOnto('a1', cardNamed('a3'));

    await waitFor(() => expect(go.reorder).toHaveLength(1));
    expect(go.reorder[0].nodeID).toBe('a1');
    expect(go.reorder[0].index).toBe(2);
    expect(go.toColumn).toEqual([]);
  });

  it('does nothing at all when the card is put back where it came from', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go);

    const readsBefore = go.reads();
    dragOnto('a1', cardNamed('a1'));

    await waitFor(() => expect(go.reorder).toEqual([]));
    expect(go.toColumn).toEqual([]);
    expect(go.reads()).toBe(readsBefore);
    expect(screen.queryByRole('alert')).toBeNull();
  });
});

describe('the keyboard model S2-16 fixed', () => {
  it('is not touched by dnd-kit: neither Space nor Enter lifts a card', async () => {
    // dnd-kit's DEFAULT sensor list is [PointerSensor, KeyboardSensor] and its
    // default start codes are Space and Enter. S2-16 has already spent both —
    // Space toggles a habit, Enter is reserved for Stage 3 — so the
    // KeyboardSensor is not registered at all and `useDraggable` therefore
    // builds no onKeyDown. This is that claim, asserted rather than reasoned:
    // if a KeyboardSensor were ever added, `draggingNodeId` would be set here.
    const go = movingGo({ 0: [card('a1')], 1: [card('b1')] });
    const store = await showTheBoard(go);
    const user = userEvent.setup();

    // Tab UNTIL, not once: S2-21 filled region 1 with the appearance controls,
    // which come before the board in the shell's DOM order.
    await tabUntil(user, () => document.activeElement === cardNamed('a1'));
    expect(document.activeElement).toBe(cardNamed('a1'));

    // One assertion per key, and in that order. The first version of this test
    // pressed both and asserted once at the end — and PASSED with a
    // KeyboardSensor deliberately registered, because Space lifted the card and
    // Enter, which is also one of dnd-kit's `end` codes, put it straight back
    // down. A test that stays green while the thing it guards is broken is
    // worse than no test at all.
    await user.keyboard(' ');
    expect(store.getState().draggingNodeId, 'Space must not lift a card').toBeNull();

    await user.keyboard('{Enter}');
    expect(store.getState().draggingNodeId, 'Enter must not lift a card').toBeNull();

    expect(go.toColumn).toEqual([]);
    expect(go.reorder).toEqual([]);
  });

  it('keeps the roving tabindex: no card is a button, and one card is the tab stop', async () => {
    // `useSortable`'s `attributes` force tabIndex=0 and role="button" on
    // whatever wears them. They are deliberately not spread. If they ever were,
    // every card would become a tab stop and this is where it would show.
    const go = movingGo({ 0: [card('a1'), card('a2')], 1: [card('b1')] });
    await showTheBoard(go);

    const cards = [...document.querySelectorAll<HTMLElement>('[data-node-id]')];
    expect(cards).toHaveLength(3);
    for (const each of cards) {
      expect(each.getAttribute('role')).toBeNull();
      expect(each.tagName).toBe('ARTICLE');
    }
    expect(document.querySelectorAll('[data-node-id][tabindex="0"]')).toHaveLength(1);
    expect(document.querySelectorAll('[role="button"]')).toHaveLength(0);
  });

  it('leaves Space on a habit doing what S2-16 says it does', async () => {
    // The collision the ticket names, from the other side: the strip and the
    // board are on screen together, and Space in the strip still checks a habit.
    const go = movingGo({ 0: [card('a1')] }, [
      habitView({ node: node({ id: 'h1', title: 'Read' }) }),
    ]);
    await showTheBoard(go);
    const user = userEvent.setup();

    const chip = await screen.findByRole('checkbox', { name: /Read/ });
    await tabUntil(user, () => document.activeElement === chip);
    expect(chip).toHaveFocus();

    await user.keyboard(' ');

    await waitFor(() => expect(go.habitChecks()).toBe(1));
  });
});

describe('what a screen reader is told', () => {
  it('reads the app‘s own instructions, in the app‘s language', async () => {
    // dnd-kit ships `defaultScreenReaderInstructions` — "To pick up a draggable
    // item, press the space bar" — hard-coded in English, and here it would
    // also be untrue. Both halves are replaced.
    const go = movingGo({ 0: [card('a1')] });
    await showTheBoard(go, 'ru');

    const description = cardNamed('a1').getAttribute('aria-describedby');
    expect(description).not.toBeNull();

    const instructions = document.getElementById(description!);
    expect(instructions).not.toBeNull();
    expect(instructions).toHaveTextContent('Перетащите карточку');
    // The chords come out of lib/keyboard.ts, which is where they are written.
    expect(instructions).toHaveTextContent(KEYS.moveLeft.hint);
    expect(instructions).toHaveTextContent(KEYS.moveRight.hint);

    expect(document.body.textContent).not.toContain('press the space bar');
  });

  it('announces the drop in the app‘s language, naming the card and the column', async () => {
    const go = movingGo({ 0: [card('a1')], 1: [card('b1')] });
    await showTheBoard(go, 'ru');

    dragOnto('a1', cardNamed('b1'));

    const live = await screen.findByRole('status');
    await waitFor(() => expect(live).toHaveTextContent('a1'));
    expect(live).toHaveTextContent('Карточка');
    expect(live.textContent).not.toContain('was dropped over droppable area');
  });
});

describe('the card in flight (S3-05, K6, D18)', () => {
  /** Every element publishing `id`, split by whether it sits inside a column. */
  function copiesOf(id: string) {
    const all = [...document.querySelectorAll<HTMLElement>(`[data-node-id="${id}"]`)];
    return {
      seated: all.filter((element) => element.closest('[data-column]') !== null),
      inFlight: all.filter((element) => element.closest('[data-column]') === null),
    };
  }

  it('draws the lifted card once, outside every column', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go);

    // Before the lift there is one card and no overlay. Without this the
    // assertion below could pass on a board that rendered the overlay always.
    expect(copiesOf('a1').inFlight).toHaveLength(0);

    pickUp('a1');

    await waitFor(() => expect(copiesOf('a1').inFlight).toHaveLength(1));
    // ONE copy in flight and ONE seat, not two of either: the overlay renders
    // the same <Card> the column does, and a second preview component would
    // show up here as a third.
    expect(copiesOf('a1').seated).toHaveLength(1);

    letGo(cardNamed('a1'));
    await waitFor(() => expect(copiesOf('a1').inFlight).toHaveLength(0));
  });

  it('hides the seat the card came out of without closing the gap', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go);

    const seat = () => copiesOf('a1').seated[0].closest('li')!;
    expect(seat().className, 'nothing is hidden before the lift').not.toContain('opacity-0');

    pickUp('a1');

    await waitFor(() => expect(seat().className).toContain('opacity-0'));
    // Still in the list, and still a list item: `hidden` or an unmount would
    // close the gap and make the column jump under the pointer on grab.
    expect(seat()).toBeInTheDocument();
    expect(seat().hasAttribute('hidden'), 'the seat gave up its space').toBe(false);
    expect(within(columnNamed(COLUMNS[0])).queryAllByRole('listitem')).toHaveLength(2);

    letGo(cardNamed('a1'));
  });

  it('left no z-index behind to explain what it never explained', async () => {
    // The `z-10` on the dragged <li> only ever existed to fight the neighbouring
    // column's stacking context, which z-index cannot cross. D18 deletes it, and
    // a stale false explanation in the code is worse than none.
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go);

    pickUp('a1');
    await waitFor(() => expect(cardNamed('a1').closest('li')!.className).toContain('opacity-0'));
    expect(cardNamed('a1').closest('li')!.className).not.toMatch(/(^|\s)z-\d/);

    letGo(cardNamed('a1'));
  });
});

describe('the drop-target highlight', () => {
  it('marks the hovered column, and only while it is hovered', async () => {
    const go = movingGo({ 0: [card('a1')], 1: [card('b1')] });
    await showTheBoard(go);

    expect(columnNamed(COLUMNS[1]).dataset.over).toBe('false');

    pickUp('a1');
    moveOver(cardNamed('b1'));

    await waitFor(() => expect(columnNamed(COLUMNS[1]).dataset.over).toBe('true'));
    // Token names only. design/ owns the values; `make guard` check 1 refuses a
    // hex literal anywhere in frontend/src.
    expect(columnNamed(COLUMNS[1]).className).toContain('border-accent');
    expect(columnNamed(COLUMNS[1]).className).not.toContain('border-line');
    expect(columnNamed(COLUMNS[0]).className).toContain('border-line');

    letGo(cardNamed('b1'));
    await waitFor(() => expect(columnNamed(COLUMNS[1]).dataset.over).toBe('false'));
  });
});

describe('prefers-reduced-motion', () => {
  /** The inline transition on the `<li>` wrapping a card, mid-drag. */
  function transitionOn(id: string): string {
    return cardNamed(id).closest('li')!.style.transition;
  }

  // What is asserted here is PRESENCE against ABSENCE of the declaration, and
  // not a duration. dnd-kit decides the duration — and under jsdom's stubbed,
  // never-changing geometry it picks its zero-duration layout variant rather
  // than the 200ms sorting one, which is its business and not this rule's. The
  // rule is that the declaration is not written at all when the user has asked
  // for less motion, so that design/tokens.css's `transition: none !important`
  // is not left arguing with an inline style.
  //
  // Both halves are needed. On its own, "no transition under reduced motion"
  // passes on a board that never had one, and dnd-kit only sets the property
  // mid-gesture — which is why both drags are asserted while the pointer is
  // still down.

  it('leaves dnd-kit its transition when motion is welcome', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go, 'en', false);

    pickUp('a1');
    moveOver(cardNamed('a2'));

    await waitFor(() => expect(transitionOn('a2')).toMatch(/transform/));
    letGo(cardNamed('a2'));
  });

  it('writes none at all when the user has asked for less of it', async () => {
    const go = movingGo({ 0: [card('a1'), card('a2')] });
    await showTheBoard(go, 'en', true);

    pickUp('a1');
    moveOver(cardNamed('a2'));

    // The transform proves the card really is mid-sort, so the empty transition
    // below is a decision and not just a card nothing has happened to yet.
    await waitFor(() => expect(cardNamed('a2').closest('li')!.style.transform).not.toBe(''));
    expect(transitionOn('a2')).toBe('');
    letGo(cardNamed('a2'));
  });
});
