import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, NodeView } from './lib/client';
import { createAppStore, type AppStore } from './store';
import { columnView, createFakeClient, node, nodeView } from './test/fakeClient';
import { renderIn } from './test/render';

// Nexus — the ACCEPT criterion's rehearsal.
//
// Every case here is driven through `render(<App store={...} />)` and by
// `user-event` key presses only. Never a click, never a `fireEvent` shortcut,
// and never `<Kanban />` on its own: S2-16 requires the real shell, because a
// key handler bound to a node the shell does not render passes every test that
// renders the board directly.
//
// This file imports `App`. It does NOT import Kanban, Column or Card — the
// import restriction is the assertion, exactly as in App.test.tsx.
//
// No column name is written down: `make guard` check 2, and the stronger claim
// that the map works without knowing what any of the five mean.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

function card(id: string): NodeView {
  return nodeView({ node: node({ id, title: id }) });
}

/**
 * A fake Go that actually moves a card, and records what it was asked for.
 *
 * It is local to this file rather than added to test/fakeClient.ts on purpose:
 * four consecutive Ctrl+Shift+ArrowRight presses are a SEQUENCE, and a fake
 * that answered identically every time would make the sequence untestable —
 * every press would target the same column and the test would pass while the
 * real board went nowhere. Relocating the node is the whole behaviour modelled,
 * because it is the whole behaviour the next press depends on.
 */
function movingGo(seed: Record<number, string[]> = { 0: ['a'] }) {
  const base = createFakeClient();
  let board: ColumnView[] = COLUMNS.map((status, index) =>
    columnView(
      status,
      (seed[index] ?? []).map((id) => card(id)),
    ),
  );

  const targets: string[] = [];
  const refuse = { move: false };

  const client: Client = {
    ...base.client,
    Board: () => Promise.resolve(board),
    MoveToColumn: (nodeID: string, target: string) => {
      targets.push(target);
      if (refuse.move) {
        return Promise.reject(new Error('service: a project can never be doing'));
      }
      board = relocate(board, nodeID, target);
      return Promise.resolve(node());
    },
  };

  return { client, targets, refuse };
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

/** Renders the shell and puts focus on the board's one tab stop. */
async function enterTheBoard(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('article');
  await user.tab();

  return user;
}

/** The id of the card that currently has focus, or null. */
function focusedCardId(): string | null {
  const active = document.activeElement;
  if (!(active instanceof HTMLElement)) {
    return null;
  }
  return active.dataset.nodeId ?? null;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('navigating the board with the keyboard', () => {
  it('reaches every card in every column with the arrows alone', async () => {
    const go = movingGo({ 0: ['a1', 'a2'], 1: ['b1', 'b2'], 2: ['c1', 'c2'] });
    const user = await enterTheBoard(storeOver(go.client));

    const visited = new Set<string>([focusedCardId()!]);

    // Along the top row, then down, then back along the bottom row.
    for (const keys of [
      '{ArrowRight}',
      '{ArrowRight}',
      '{ArrowDown}',
      '{ArrowLeft}',
      '{ArrowLeft}',
    ]) {
      await user.keyboard(keys);
      visited.add(focusedCardId()!);
    }
    // And up the first column again, which must land back where we started.
    await user.keyboard('{ArrowUp}');
    visited.add(focusedCardId()!);

    expect([...visited].sort()).toEqual(['a1', 'a2', 'b1', 'b2', 'c1', 'c2']);
    expect(focusedCardId()).toBe('a1');
  });

  it('jumps to the first and last card of a column with Home and End', async () => {
    const go = movingGo({ 0: ['a1', 'a2', 'a3'] });
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard('{End}');
    expect(focusedCardId()).toBe('a3');

    await user.keyboard('{Home}');
    expect(focusedCardId()).toBe('a1');
  });

  it('keeps exactly one card in the tab order at any moment', async () => {
    const go = movingGo({ 0: ['a1', 'a2'], 1: ['b1'] });
    const user = await enterTheBoard(storeOver(go.client));

    const tabStops = () => document.querySelectorAll('[data-node-id][tabindex="0"]');

    expect(tabStops()).toHaveLength(1);
    await user.keyboard('{ArrowDown}');
    expect(tabStops()).toHaveLength(1);
    expect(tabStops()[0]).toHaveFocus();

    await user.keyboard('{ArrowRight}');
    expect(tabStops()).toHaveLength(1);
    expect(tabStops()[0]).toHaveFocus();
  });

  it('treats the whole board as ONE tab stop', async () => {
    // Five columns times n cards as n+5 tab stops is unusable. Tab from the
    // card leaves the board entirely rather than walking to the next card.
    const store = storeOver(movingGo({ 0: ['a1', 'a2'], 1: ['b1'] }).client);
    const user = await enterTheBoard(store);

    expect(focusedCardId()).toBe('a1');

    await user.tab();

    expect(focusedCardId()).toBeNull();
  });

  it('moves between regions with Tab and back with Shift+Tab', async () => {
    // Region order is the shell's DOM order and nothing else. The toast layer
    // is region 5, the board is region 3, and today those are the two regions
    // with anything in them.
    const store = storeOver(movingGo().client);
    store.getState().pushToast('toast.error.body');
    const user = await enterTheBoard(store);

    expect(focusedCardId()).toBe('a');

    await user.tab();
    expect(screen.getByRole('button')).toHaveFocus();

    await user.tab({ shift: true });
    expect(focusedCardId()).toBe('a');
  });

  it('leaves Enter alone', async () => {
    // Reserved for Stage 3's detail slide-over. Not handled, and — the part
    // that matters — not intercepted either: nothing calls preventDefault, so
    // whatever is bound later actually receives it.
    const go = movingGo();
    const user = await enterTheBoard(storeOver(go.client));

    const seen: KeyboardEvent[] = [];
    const listener = (event: KeyboardEvent) => {
      if (event.key === 'Enter') {
        seen.push(event);
      }
    };
    document.addEventListener('keydown', listener);
    await user.keyboard('{Enter}');
    document.removeEventListener('keydown', listener);

    expect(seen).toHaveLength(1);
    expect(seen[0].defaultPrevented).toBe(false);
    expect(focusedCardId()).toBe('a');
    expect(go.targets).toEqual([]);
    expect(screen.queryByRole('alert')).toBeNull();
  });
});

describe('moving a card with the keyboard', () => {
  const RIGHT = '{Control>}{Shift>}{ArrowRight}{/Shift}{/Control}';
  const LEFT = '{Control>}{Shift>}{ArrowLeft}{/Shift}{/Control}';

  it('walks a card across all five columns, keeping focus on it', async () => {
    // The ACCEPT criterion's spine: four presses, four calls, the statuses Go
    // supplied in Go's order, and the same card focused at every step.
    const go = movingGo();
    const user = await enterTheBoard(storeOver(go.client));

    for (let step = 1; step <= 4; step += 1) {
      await user.keyboard(RIGHT);

      await waitFor(() => expect(go.targets).toHaveLength(step));
      await waitFor(() => expect(focusedCardId()).toBe('a'));
    }

    expect(go.targets).toEqual(COLUMNS.slice(1));
    // And the card really is in the last column now, not merely reported as
    // moved: the next press goes nowhere.
    await user.keyboard(RIGHT);
    expect(go.targets).toHaveLength(4);
  });

  it('does nothing at all off the right-hand end', async () => {
    const go = movingGo({ 4: ['z'] });
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(RIGHT);

    expect(go.targets).toEqual([]);
    expect(screen.queryByRole('alert')).toBeNull();
    expect(focusedCardId()).toBe('z');
  });

  it('does nothing at all off the left-hand end', async () => {
    const go = movingGo({ 0: ['a'] });
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(LEFT);

    expect(go.targets).toEqual([]);
    expect(screen.queryByRole('alert')).toBeNull();
    expect(focusedCardId()).toBe('a');
  });

  it('raises exactly one toast on a refusal and leaves focus on the card', async () => {
    // A project dragged to Doing (D9) is the real case. Go refuses, the store
    // raises the one toast, the board is untouched and the card keeps focus —
    // a refusal must not cost the user their place.
    const go = movingGo();
    go.refuse.move = true;
    const user = await enterTheBoard(storeOver(go.client));

    await user.keyboard(RIGHT);

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');
    expect(alert).not.toHaveTextContent('a project can never be doing');
    expect(focusedCardId()).toBe('a');
  });
});

describe('the global shortcuts', () => {
  it('open the overlays from outside the board', async () => {
    // S2-16 requires Ctrl+N and Ctrl+K to fire wherever focus happens to be.
    // Focus is deliberately parked on the toast's dismiss button — a region the
    // board does not own — before every press.
    const store = storeOver(movingGo().client);
    store.getState().pushToast('toast.error.body');
    const user = await enterTheBoard(store);

    await user.tab();
    expect(screen.getByRole('button')).toHaveFocus();

    await user.keyboard('{Control>}n{/Control}');
    expect(store.getState().openOverlay).toBe('quickAdd');

    await user.keyboard('{Escape}');
    expect(store.getState().openOverlay).toBeNull();

    await user.keyboard('{Control>}k{/Control}');
    expect(store.getState().openOverlay).toBe('commandPalette');

    await user.keyboard('{Escape}');
    expect(store.getState().openOverlay).toBeNull();
  });

  it('do not fire on the unmodified letters', async () => {
    const store = storeOver(movingGo().client);
    const user = await enterTheBoard(store);

    await user.keyboard('nk');

    expect(store.getState().openOverlay).toBeNull();
  });
});

describe('the focus ring', () => {
  // Three things are true about this criterion and all three are said out loud.
  //
  // 1. jsdom evaluates `:focus-visible` to FALSE for a programmatically focused
  //    element — correctly, since the selector is a heuristic about HOW focus
  //    arrived. So no test in this project can assert the ring is painted. That
  //    half is the hand pass, in both palettes.
  // 2. The rule itself lives in src/style.css (`:focus-visible { outline-accent }`,
  //    applied to every element, shipped in S2-09) and CANNOT be read from in
  //    here: the vitest config sets `css: false`, so a CSS import — raw, globbed
  //    or otherwise — resolves to the empty string, and reading it off disk
  //    needs @types/node, which this ticket's Scope does not include. An
  //    assertion against an empty string would have passed while proving
  //    nothing, which is worse than not making it.
  // 3. What IS mechanical is the two ways a ring that exists gets lost: an
  //    element that cannot take focus at all, and a class that switches the
  //    outline off. Both are asserted below.
  const SOURCES = import.meta.glob('./**/*.{ts,tsx}', {
    query: '?raw',
    import: 'default',
    eager: true,
  }) as Record<string, string>;

  // Every spelling of "take the outline away" that could plausibly be written
  // here: the Tailwind utility with or without a variant prefix, the CSS
  // property in an inline style object (where the value is QUOTED, which is the
  // form a narrower pattern misses), and the zero-width version.
  const OUTLINE_KILLERS = /outline-none|outline\s*:\s*['"`]?\s*none|outlineWidth\s*:\s*['"`]?0/;

  it('is switched off nowhere in frontend/src', () => {
    const offenders = Object.entries(SOURCES)
      .filter(
        ([path, source]) => path !== './App.keyboard.test.tsx' && OUTLINE_KILLERS.test(source),
      )
      .map(([path]) => path);

    expect(offenders, 'a focus ring that only appears sometimes is guesswork').toEqual([]);
  });

  it('actually recognises an outline being switched off', () => {
    // The scan above is a regex over source text, so it is only as good as the
    // spellings it knows. These are the ones it claims to catch, asserted
    // rather than hoped for — the first version of it missed
    // `style={{ outline: 'none' }}` entirely, which is how this case came to
    // exist.
    for (const written of [
      'className="outline-none"',
      'className="focus:outline-none"',
      "style={{ outline: 'none' }}",
      'style={{ outline: none }}',
      'style={{ outlineWidth: 0 }}',
    ]) {
      expect(OUTLINE_KILLERS.test(written), written).toBe(true);
    }
  });

  it('has something to attach to on every element the keyboard reaches', () => {
    // Asserted on the live shell rather than on the source: an element with no
    // tabindex and no native focusability never matches :focus-visible at all,
    // which is the other way to have an invisible focus ring.
    const store = storeOver(movingGo({ 0: ['a1', 'a2'], 1: ['b1'] }).client);
    store.getState().pushToast('toast.error.body');

    return enterTheBoard(store).then(() => {
      const reachable = [...document.querySelectorAll<HTMLElement>('[data-node-id], button')];

      expect(reachable.length).toBeGreaterThan(1);
      for (const element of reachable) {
        const focusable =
          element.tagName === 'BUTTON' || element.tabIndex === 0 || element.tabIndex === -1;
        expect(focusable, `${element.tagName} cannot take focus`).toBe(true);
      }
    });
  });
});
