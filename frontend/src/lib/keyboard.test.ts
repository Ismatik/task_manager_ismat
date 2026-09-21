import { describe, expect, it } from 'vitest';

import {
  adjacentColumn,
  boardActionFor,
  globalActionFor,
  KEYS,
  matches,
  nextFocusId,
  rovingNodeId,
  type Chord,
} from './keyboard';
import { columnView, node, nodeView } from '../test/fakeClient';
import type { ColumnView } from './client';

// The keyboard map and the focus arithmetic, tested without React.
//
// The interaction half — that these bindings actually reach a card, move it and
// keep focus on it — is asserted through `render(<App ... />)` in
// App.keyboard.test.tsx, because a key handler attached to a node the shell
// never renders passes every test in this file.
//
// Column names are opaque here for the usual reason: `make guard` check 2
// refuses a quoted status anywhere in frontend/src, and nonsense names make the
// stronger claim — the map has no idea what a column means.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

function card(id: string) {
  return nodeView({ node: node({ id, title: id }) });
}

/** A board with the named cards in the columns their index names. */
function board(cards: Record<number, string[]>): ColumnView[] {
  return COLUMNS.map((status, index) => columnView(status, (cards[index] ?? []).map(card)));
}

function press(key: string, modifiers: Partial<KeyboardEvent> = {}): KeyboardEvent {
  return new KeyboardEvent('keydown', { key, ...modifiers });
}

describe('the keyboard map', () => {
  it('matches a chord only when every modifier agrees', () => {
    const chord: Chord = { key: 'ArrowRight', ctrl: true, shift: true, hint: 'x' };

    expect(matches(press('ArrowRight', { ctrlKey: true, shiftKey: true }), chord)).toBe(true);
    expect(matches(press('ArrowRight', { ctrlKey: true }), chord)).toBe(false);
    expect(matches(press('ArrowRight'), chord)).toBe(false);
    // Alt and Meta must be up even though the chord does not mention them: a
    // binding that also fired on Super+Arrow is a binding that fights the
    // desktop.
    expect(
      matches(press('ArrowRight', { ctrlKey: true, shiftKey: true, metaKey: true }), chord),
    ).toBe(false);
  });

  it('tells a bare arrow apart from the move chord, in both directions', () => {
    // The assertion that keeps the SCAN ORDER out of the map. If modifiers were
    // matched loosely, whichever entry came first in the table would win.
    expect(boardActionFor(press('ArrowRight'))).toBe('nextColumn');
    expect(boardActionFor(press('ArrowRight', { ctrlKey: true, shiftKey: true }))).toBe(
      'moveRight',
    );
    expect(boardActionFor(press('ArrowLeft'))).toBe('previousColumn');
    expect(boardActionFor(press('ArrowLeft', { ctrlKey: true, shiftKey: true }))).toBe('moveLeft');
  });

  it('binds nothing at all to Enter', () => {
    // S2-16: reserved for Stage 3's detail slide-over, and "it must not be
    // silently swallowed — no handler at all is correct".
    expect(boardActionFor(press('Enter'))).toBeNull();
    expect(globalActionFor(press('Enter'))).toBeNull();
  });

  it('keeps the global shortcuts out of the board map and the other way round', () => {
    expect(globalActionFor(press('n', { ctrlKey: true }))).toBe('quickAdd');
    expect(globalActionFor(press('k', { ctrlKey: true }))).toBe('commandPalette');
    expect(globalActionFor(press('Escape'))).toBe('closeOverlay');

    expect(boardActionFor(press('n', { ctrlKey: true }))).toBeNull();
    expect(globalActionFor(press('ArrowDown'))).toBeNull();
    // Unmodified letters belong to whatever has focus — a text field, one day.
    expect(globalActionFor(press('n'))).toBeNull();
  });

  it('gives every action a distinct hint for S2-20 to render', () => {
    const hints = Object.values(KEYS).map((chord) => chord.hint);

    expect(hints.every((hint) => hint.length > 0)).toBe(true);
    expect(new Set(hints).size).toBe(hints.length);
  });
});

describe('where the focus goes', () => {
  const three = board({ 0: ['a', 'b', 'c'] });

  it('clamps at the top and the bottom of a column rather than wrapping', () => {
    expect(nextFocusId(three, 'a', 'previousCard')).toBe('a');
    expect(nextFocusId(three, 'c', 'nextCard')).toBe('c');
    expect(nextFocusId(three, 'a', 'nextCard')).toBe('b');
    expect(nextFocusId(three, 'c', 'previousCard')).toBe('b');
  });

  it('jumps to the first and last card of the column', () => {
    expect(nextFocusId(three, 'b', 'firstCard')).toBe('a');
    expect(nextFocusId(three, 'b', 'lastCard')).toBe('c');
  });

  it('lands on the card at the same height in the next column', () => {
    const wide = board({ 0: ['a', 'b', 'c'], 1: ['d', 'e', 'f'] });

    expect(nextFocusId(wide, 'b', 'nextColumn')).toBe('e');
    expect(nextFocusId(wide, 'e', 'previousColumn')).toBe('b');
  });

  it('lands on the last card when the next column is shorter', () => {
    const ragged = board({ 0: ['a', 'b', 'c'], 1: ['d'] });

    expect(nextFocusId(ragged, 'c', 'nextColumn')).toBe('d');
  });

  it('steps past an empty column instead of stopping on it', () => {
    // Stopping on an empty column would leave every column beyond the gap
    // unreachable without a mouse — on the one screen whose acceptance
    // criterion is that a mouse is never needed.
    const gapped = board({ 0: ['a'], 3: ['d'] });

    expect(nextFocusId(gapped, 'a', 'nextColumn')).toBe('d');
    expect(nextFocusId(gapped, 'd', 'previousColumn')).toBe('a');
  });

  it('goes nowhere off either end of the board', () => {
    const edges = board({ 0: ['a'], 4: ['e'] });

    expect(nextFocusId(edges, 'e', 'nextColumn')).toBeNull();
    expect(nextFocusId(edges, 'a', 'previousColumn')).toBeNull();
  });

  it('knows nothing about a card that is not on the board', () => {
    expect(nextFocusId(three, 'ghost', 'nextCard')).toBeNull();
  });
});

describe('the roving tabindex', () => {
  const two = board({ 0: ['a'], 1: ['b'] });

  it('is on the selected card', () => {
    expect(rovingNodeId(two, 'b')).toBe('b');
  });

  it('falls back to the first card when nothing is selected', () => {
    expect(rovingNodeId(two, null)).toBe('a');
  });

  it('falls back when the selected card has left the board', () => {
    // Archived, or moved out from under us. Leaving the tabindex on a card that
    // is not rendered is a board with no tab stop at all.
    expect(rovingNodeId(two, 'ghost')).toBe('a');
  });

  it('has nowhere to be on a board with no cards', () => {
    expect(rovingNodeId(board({}), null)).toBeNull();
  });
});

describe('the move target', () => {
  const spread = board({ 0: ['a'], 2: ['c'], 4: ['e'] });

  it("is GO'S column, one step along", () => {
    // The value handed back to MoveToColumn is a status string Go produced.
    // Nothing computes a "next status"; the order is Go's and so is the name.
    expect(adjacentColumn(spread, 'c', 1)?.status).toBe(COLUMNS[3]);
    expect(adjacentColumn(spread, 'c', -1)?.status).toBe(COLUMNS[1]);
  });

  it('is null at either end — a no-op, not a wrap-around', () => {
    expect(adjacentColumn(spread, 'e', 1)).toBeNull();
    expect(adjacentColumn(spread, 'a', -1)).toBeNull();
  });

  it('is null for a card that is not on the board', () => {
    expect(adjacentColumn(spread, 'ghost', 1)).toBeNull();
  });
});
