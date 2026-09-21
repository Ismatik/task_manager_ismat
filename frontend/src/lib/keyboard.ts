import type { ColumnView } from './client';

// Nexus — the keyboard map, and the ONE place it is written down.
//
// TASKS.md S2-16 prints this table and calls it normative: the command palette
// (S2-20) renders its hints from here rather than restating them, dnd-kit
// (S2-17) must not bind a key that collides with it, and the habits strip
// (S2-18) takes its Space from here. A second copy of a key binding is a second
// binding, and the one on screen would be the untested one.
//
//   Tab / Shift+Tab           move between regions (the shell's DOM order)
//   ArrowLeft / ArrowRight    the adjacent column, landing on the nearest card
//   ArrowUp / ArrowDown       the previous / next card in the column
//   Home / End                the first / last card in the column
//   Ctrl+Shift+ArrowRight     move the focused card one column right
//   Ctrl+Shift+ArrowLeft      move the focused card one column left
//   Ctrl+N                    quick add (S2-19)
//   Ctrl+K                    command palette (S2-20)
//   Escape                    close the topmost overlay
//   Enter on a card           RESERVED for Stage 3's detail slide-over
//   Space on a habit          toggle today's check (S2-18)
//
// # Tab is not in the table below, and that is deliberate
//
// Movement between regions is the browser's own Tab, over the shell's DOM order
// (App.tsx). Implementing it would mean holding a second copy of the region
// order in a focus manager, which is the defect this project keeps paying for.
// The board is ONE tab stop — a roving tabindex, not five columns times n cards
// of them — and that is what `rovingNodeId` below is for.
//
// # Enter has no entry, and that is also deliberate
//
// S2-16: "It must not be silently swallowed — no handler at all is correct." So
// `boardActionFor` returns null for Enter, nothing calls preventDefault, and the
// event reaches whatever Stage 3 eventually puts there.

/** What a key press does inside the board. */
export type BoardAction =
  | 'nextCard'
  | 'previousCard'
  | 'firstCard'
  | 'lastCard'
  | 'nextColumn'
  | 'previousColumn'
  | 'moveRight'
  | 'moveLeft';

/** What a key press does wherever focus happens to be. */
export type GlobalAction = 'quickAdd' | 'commandPalette' | 'closeOverlay';

/** One key combination. Absent modifiers mean the modifier must be UP. */
export interface Chord {
  /** `KeyboardEvent.key`, compared case-insensitively. */
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  /**
   * How the combination is written on screen.
   *
   * Not an i18n key: `Ctrl`, `Home` and `Esc` are what the keycaps say in both
   * languages, and a translated modifier name would be a modifier name that
   * does not match the keyboard. Whatever renders one of these does so in
   * `font-mono` (PLAN.md section 3).
   */
  hint: string;
}

const BOARD_KEYS: Readonly<Record<BoardAction, Chord>> = {
  nextCard: { key: 'ArrowDown', hint: '↓' },
  previousCard: { key: 'ArrowUp', hint: '↑' },
  firstCard: { key: 'Home', hint: 'Home' },
  lastCard: { key: 'End', hint: 'End' },
  nextColumn: { key: 'ArrowRight', hint: '→' },
  previousColumn: { key: 'ArrowLeft', hint: '←' },
  moveRight: { key: 'ArrowRight', ctrl: true, shift: true, hint: 'Ctrl+Shift+→' },
  moveLeft: { key: 'ArrowLeft', ctrl: true, shift: true, hint: 'Ctrl+Shift+←' },
};

const GLOBAL_KEYS: Readonly<Record<GlobalAction, Chord>> = {
  quickAdd: { key: 'n', ctrl: true, hint: 'Ctrl+N' },
  commandPalette: { key: 'k', ctrl: true, hint: 'Ctrl+K' },
  closeOverlay: { key: 'Escape', hint: 'Esc' },
};

/** Space on a focused habit, consumed by the strip (S2-18). */
const HABIT_KEYS = {
  toggleHabit: { key: ' ', hint: 'Space' } as Chord,
};

/** Every chord, by name. S2-20 reads its hints from here. */
export const KEYS = { ...BOARD_KEYS, ...GLOBAL_KEYS, ...HABIT_KEYS };

/** The name of every action this map binds. */
export type ActionName = keyof typeof KEYS;

/**
 * Does this event match this chord?
 *
 * Modifiers are matched EXACTLY, in both directions: `ArrowRight` alone does not
 * match `Ctrl+Shift+ArrowRight`, and `Ctrl+Shift+ArrowRight` does not match a
 * bare `ArrowRight`. Without that, the order the table is scanned in would
 * silently become part of the map.
 *
 * Meta is always required to be up. A binding that also fired on Super+N would
 * be a binding that fights the desktop.
 */
export function matches(event: KeyboardEvent, chord: Chord): boolean {
  return (
    event.key.toLowerCase() === chord.key.toLowerCase() &&
    event.ctrlKey === (chord.ctrl ?? false) &&
    event.shiftKey === (chord.shift ?? false) &&
    !event.altKey &&
    !event.metaKey
  );
}

/** The board action this event asks for, or null. */
export function boardActionFor(event: KeyboardEvent): BoardAction | null {
  for (const [action, chord] of Object.entries(BOARD_KEYS)) {
    if (matches(event, chord)) {
      return action as BoardAction;
    }
  }
  return null;
}

/** The global action this event asks for, or null. */
export function globalActionFor(event: KeyboardEvent): GlobalAction | null {
  for (const [action, chord] of Object.entries(GLOBAL_KEYS)) {
    if (matches(event, chord)) {
      return action as GlobalAction;
    }
  }
  return null;
}

// ---------------------------------------------------------------------------
// Where the focus goes. Pure functions over the board Go returned.
//
// None of this is a domain rule. It reads the ARRAY Go sent — its order, and
// which array each node is in — and returns an id or a column out of it. It
// never asks what a status means, never computes a "next status", and would
// behave identically if the five columns were called one through five.

/** Where a node sits on the board, in array indices. */
interface Seat {
  column: number;
  card: number;
}

function seatOf(board: ColumnView[], nodeId: string): Seat | null {
  for (let column = 0; column < board.length; column += 1) {
    const card = board[column].nodes.findIndex((view) => view.node.id === nodeId);
    if (card !== -1) {
      return { column, card };
    }
  }
  return null;
}

/** The first card on the board, in Go's order, or null on an empty board. */
function firstCardId(board: ColumnView[]): string | null {
  for (const column of board) {
    if (column.nodes.length > 0) {
      return column.nodes[0].node.id;
    }
  }
  return null;
}

/**
 * The one card in the board that carries `tabIndex=0`.
 *
 * The roving tabindex, and the reason the board is a single tab stop: five
 * columns times n cards as n+5 tab stops is unusable with a keyboard. The
 * selected card holds it; before anything has been selected — and after a
 * selected card leaves the board — the first card does.
 */
export function rovingNodeId(board: ColumnView[], selectedNodeId: string | null): string | null {
  if (selectedNodeId !== null && seatOf(board, selectedNodeId) !== null) {
    return selectedNodeId;
  }
  return firstCardId(board);
}

/**
 * The card a navigation action moves focus to, or null when there is nowhere to
 * go.
 *
 * Every edge CLAMPS. Nothing wraps: arriving back at Backlog by holding
 * ArrowRight is a surprise, and a surprise in a keyboard model is a bug report.
 */
export function nextFocusId(
  board: ColumnView[],
  nodeId: string,
  action: BoardAction,
): string | null {
  const seat = seatOf(board, nodeId);
  if (seat === null) {
    return null;
  }

  const here = board[seat.column].nodes;

  switch (action) {
    case 'nextCard':
      return here[Math.min(seat.card + 1, here.length - 1)].node.id;
    case 'previousCard':
      return here[Math.max(seat.card - 1, 0)].node.id;
    case 'firstCard':
      return here[0].node.id;
    case 'lastCard':
      return here[here.length - 1].node.id;
    case 'nextColumn':
      return nearestCardId(board, seat, 1);
    case 'previousColumn':
      return nearestCardId(board, seat, -1);
    default:
      // moveRight and moveLeft are not navigation; the caller handles them.
      return null;
  }
}

/**
 * The nearest card in the given direction, at the same height where possible.
 *
 * It steps PAST an empty column rather than stopping on it. An empty column has
 * no card to land on, so stopping there would leave the keyboard unable to
 * reach anything beyond it — the columns past the gap would simply be
 * unreachable without a mouse, on a screen whose acceptance criterion is that a
 * mouse is never needed.
 */
function nearestCardId(board: ColumnView[], seat: Seat, offset: number): string | null {
  for (let column = seat.column + offset; column >= 0 && column < board.length; column += offset) {
    const nodes = board[column].nodes;
    if (nodes.length > 0) {
      return nodes[Math.min(seat.card, nodes.length - 1)].node.id;
    }
  }
  return null;
}

/**
 * The column one step from the node's own, or null at either end.
 *
 * This is what a keyboard move targets, and the returned value is GO'S COLUMN —
 * the caller sends `column.status` back, a string Go produced. The frontend
 * knows the columns' ORDER, because Go returned them in order; it does not know
 * their names, and it never computes a "next status".
 *
 * Right from the last column and left from the first return null: a no-op, not
 * an error and not a wrap-around.
 */
export function adjacentColumn(
  board: ColumnView[],
  nodeId: string,
  offset: number,
): ColumnView | null {
  const seat = seatOf(board, nodeId);
  if (seat === null) {
    return null;
  }

  const column = seat.column + offset;
  if (column < 0 || column >= board.length) {
    return null;
  }
  return board[column];
}
