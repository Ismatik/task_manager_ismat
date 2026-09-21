import {
  useCallback,
  useEffect,
  useRef,
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react';
import { useTranslation } from 'react-i18next';

import { Column } from '../components/Column';
import { boardActionFor, nextFocusId, rovingNodeId } from '../lib/keyboard';
import { useAppState, useAppStore } from '../store/context';

// Nexus — the board, and the keyboard model inside it.
//
// # Five columns, in Go's order, and the frontend does not know what they are
//
// `Board()` returns a slice of ColumnView, and dto.go says why it is a slice and
// not a map: "the five columns have an order ... and a map has none, in Go or in
// JSON". So this renders the array as it arrived. There is no list of statuses
// here, no `.sort(`, no `.filter(` and no five-element constant — the board is
// five columns long because Go sent five, and App.test.tsx proves the point by
// mocking a client that returns them in a deliberately unusual order and
// asserting the UI shows THAT order.
//
// # Three states, and none of them is a blank window
//
//   board === null   the read has not answered yet — nothing, briefly
//   board is empty   Go answered with no columns: a localised line, not a void
//   otherwise        the columns
//
// A FAILED read is the first case: S2-13's store leaves `board` at null and
// raises a toast. The user sees the shell and an error rather than a white
// rectangle — S2-13's rule unchanged, a failed read is a bad session, not a
// blank window.
//
// # Focus is the truth, and selection follows it
//
// There is one source of truth for "which card are we on", and it is the DOM's
// own `document.activeElement`. The store's `selectedNodeId` MIRRORS it: the
// delegated focus handler below writes it, and nothing else does. The
// alternative — selection in the store, focus chasing it through an effect —
// gives two answers to one question, and they disagree the first time the user
// clicks a card instead of arrowing to it.
//
// What the store's copy is for is the roving tabindex, which has to be a
// property of the render rather than of the live DOM.
//
// # The one effect, and why it is not focus-stealing
//
// A Ctrl+Shift+Arrow move re-reads the board, so the card unmounts from one
// column and mounts in another and the browser drops focus to <body>. The
// effect puts it back. It fires only when the board HAD focus — `ownedFocus`,
// which is cleared when focus leaves for a real element elsewhere and kept when
// focus is lost to nothing, because being lost to nothing is exactly the unmount
// case. So the board never grabs focus on startup, and never takes it back from
// a toast button the user has tabbed to.

export function Kanban() {
  const { t } = useTranslation();
  const store = useAppStore();
  const board = useAppState((state) => state.board);
  const selectedNodeId = useAppState((state) => state.selectedNodeId);

  const boardRef = useRef<HTMLDivElement>(null);
  const ownedFocus = useRef(false);

  const roving = board === null ? null : rovingNodeId(board, selectedNodeId);

  // A scan rather than a `[data-node-id="..."]` selector: a node id is a uuid
  // today, and an attribute selector that has to be escaped is a bug waiting
  // for the first id with a quote in it.
  const focusCard = useCallback((nodeId: string) => {
    for (const card of boardRef.current?.querySelectorAll<HTMLElement>('[data-node-id]') ?? []) {
      if (card.dataset.nodeId === nodeId) {
        card.focus();
        return;
      }
    }
  }, []);

  useEffect(() => {
    if (!ownedFocus.current || roving === null) {
      return;
    }
    focusCard(roving);
  }, [board, roving, focusCard]);

  if (board === null) {
    return null;
  }

  if (board.length === 0) {
    return <p className="min-w-0 break-words text-muted">{t('board.empty')}</p>;
  }

  /** The card an event came from, by the id the card publishes on itself. */
  const cardIdFrom = (target: EventTarget | null): string | null => {
    if (!(target instanceof Element)) {
      return null;
    }
    return target.closest<HTMLElement>('[data-node-id]')?.dataset.nodeId ?? null;
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    const nodeId = cardIdFrom(event.target);
    if (nodeId === null) {
      return;
    }

    const action = boardActionFor(event.nativeEvent);
    if (action === null) {
      // Enter lands here, and leaves untouched: no handler, nothing
      // preventDefault-ed, nothing swallowed. Stage 3's detail slide-over gets
      // the event exactly as the user pressed it.
      return;
    }

    event.preventDefault();

    if (action === 'moveRight' || action === 'moveLeft') {
      void store.getState().moveToAdjacentColumn(nodeId, action === 'moveRight' ? 1 : -1);
      return;
    }

    const next = nextFocusId(board, nodeId, action);
    if (next !== null) {
      focusCard(next);
    }
  };

  const handleFocus = (event: ReactFocusEvent<HTMLDivElement>) => {
    ownedFocus.current = true;

    const nodeId = cardIdFrom(event.target);
    if (nodeId !== null && nodeId !== selectedNodeId) {
      store.getState().select(nodeId);
    }
  };

  const handleBlur = (event: ReactFocusEvent<HTMLDivElement>) => {
    // relatedTarget null means focus went NOWHERE, which is what happens when
    // the focused card unmounts mid-move. Keeping ownership then is the whole
    // point; clearing it would make the move lose the card.
    if (event.relatedTarget !== null && !event.currentTarget.contains(event.relatedTarget)) {
      ownedFocus.current = false;
    }
  };

  return (
    // `overflow-x-auto` on the board and nowhere else: when five columns will
    // not fit even at their floor width, the BOARD scrolls. A column never
    // scrolls sideways inside itself, which is what "no horizontal scroll
    // inside a column" means.
    <div
      ref={boardRef}
      onKeyDown={handleKeyDown}
      onFocus={handleFocus}
      onBlur={handleBlur}
      className="flex min-w-0 items-start gap-2 overflow-x-auto"
    >
      {board.map((column) => (
        <Column key={column.status} column={column} rovingNodeId={roving} />
      ))}
    </div>
  );
}

export default Kanban;
