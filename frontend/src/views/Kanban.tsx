import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  type FocusEvent as ReactFocusEvent,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react';
import {
  closestCorners,
  DndContext,
  PointerSensor,
  useSensor,
  useSensors,
  type Active,
  type Announcements,
  type DragEndEvent,
  type DragStartEvent,
  type Over,
  type ScreenReaderInstructions,
} from '@dnd-kit/core';
import { hasSortableData } from '@dnd-kit/sortable';
import { useTranslation } from 'react-i18next';

import { Column } from '../components/Column';
import type { ColumnView } from '../lib/client';
import { focusWithoutScrolling } from '../lib/focus';
import { boardActionFor, KEYS, nextFocusId, rovingNodeId } from '../lib/keyboard';
import type { DropTarget } from '../store/data';
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
//
// # Drag and drop (S2-17), and how the Space collision was resolved
//
// **dnd-kit's KeyboardSensor is not registered.** That is the resolution, and
// TASKS.md S2-17 names the direction: "S2-16's map is normative ... if dnd-kit's
// default `Space`-to-lift conflicts with the habit-strip `Space`, dnd-kit
// yields." Its default sensor list is `[PointerSensor, KeyboardSensor]` and its
// default keyboard codes are `start: [Space, Enter]` — both of which S2-16 has
// already spent: `Space` toggles a habit, and `Enter` on a card is RESERVED for
// Stage 3 and must not be intercepted. So `sensors` below is passed explicitly
// with PointerSensor alone. `useDraggable` builds its listeners out of the
// registered sensors' activators, so with no KeyboardSensor there is no
// `onKeyDown` on a card at all — nothing competes, rather than something
// competing and losing. The keyboard route to the same outcome already exists
// and is better: Ctrl+Shift+Arrow, S2-16, which needs no lift and no drop.
//
// The accessibility strings are ours. dnd-kit's defaults are hard-coded English
// ("Picked up draggable item ...", "To pick up a draggable item, press the space
// bar") — the second of which would also be a lie here — so both the
// announcements and the screen-reader instructions are `t()` keys, and the
// instructions name the keyboard chords by reading KEYS out of lib/keyboard.ts
// rather than spelling them a second time in a locale file.
//
// What a drop MEANS is not decided here. The handler reads the two ends of the
// gesture off the event and hands them to the store; which call that becomes,
// and what happens to the due date and the subtree, is store/data.ts and Go.

/**
 * How far a pointer must travel before a press becomes a drag.
 *
 * Without it every click on a card is a zero-distance drag, and Stage 3's
 * click-to-open would never fire. Half a step of the 8px grid.
 */
const DRAG_THRESHOLD_PX = 4;

/**
 * Where one end of a drag is, as a column Go named and a position in it.
 *
 * A card reports its own seat through `SortableContext` — `containerId` is the
 * column's status because components/Column.tsx keys the context with it. A
 * drop on the column itself has no seat, and means the end of that column: an
 * empty column, or the padding below the last card.
 */
function placeOf(board: ColumnView[], item: Active | Over): DropTarget | null {
  if (hasSortableData(item)) {
    const { containerId, index } = item.data.current.sortable;
    return { status: String(containerId), index };
  }

  const status = String(item.id);
  const column = board.find((each) => each.status === status);
  return column === undefined ? null : { status, index: column.nodes.length };
}

/** A string a drag event published about itself, or the id as a last resort. */
function published(item: Active | Over | null, key: string): string {
  const value = item?.data.current?.[key];
  return typeof value === 'string' ? value : String(item?.id ?? '');
}

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
        focusWithoutScrolling(card);
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

  // PointerSensor, and nothing else. See the note at the top of this file: the
  // default list also installs a KeyboardSensor on Space and Enter, both of
  // which S2-16 has already assigned.
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: DRAG_THRESHOLD_PX } }),
  );

  const announcements: Announcements = useMemo(
    () => ({
      onDragStart: ({ active }) => t('board.dnd.lifted', { card: published(active, 'title') }),
      onDragOver: ({ active, over }) =>
        over === null
          ? t('board.dnd.outside', { card: published(active, 'title') })
          : t('board.dnd.over', {
              card: published(active, 'title'),
              column: published(over, 'column'),
            }),
      onDragEnd: ({ active, over }) =>
        over === null
          ? t('board.dnd.cancelled', { card: published(active, 'title') })
          : t('board.dnd.dropped', {
              card: published(active, 'title'),
              column: published(over, 'column'),
            }),
      onDragCancel: ({ active }) => t('board.dnd.cancelled', { card: published(active, 'title') }),
    }),
    [t],
  );

  const screenReaderInstructions: ScreenReaderInstructions = useMemo(
    () => ({
      // The chords are READ from lib/keyboard.ts, never written here. S2-16's
      // map is the one spelling of them, and an instruction that named a key
      // the map did not bind would be an instruction that lies.
      draggable: t('board.dnd.instructions', {
        left: KEYS.moveLeft.hint,
        right: KEYS.moveRight.hint,
      }),
    }),
    [t],
  );

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

  const handleDragStart = (event: DragStartEvent) => {
    store.getState().setDragging(String(event.active.id));
  };

  const handleDragEnd = ({ active, over }: DragEndEvent) => {
    store.getState().setDragging(null);
    if (over === null) {
      // Let go over nothing. Not an error, and nothing to undo: the card was
      // never moved anywhere but on its own transform.
      return;
    }

    const from = placeOf(board, active);
    const to = placeOf(board, over);
    if (from === null || to === null) {
      return;
    }

    void store.getState().dropCard(String(active.id), from, to);
  };

  const handleDragCancel = () => {
    store.getState().setDragging(null);
  };

  return (
    <DndContext
      sensors={sensors}
      // The columns and the cards are both drop targets, nested, so the
      // question "which one is the pointer nearest" has to be answered by
      // corner distance rather than by containment — `pointerWithin` would
      // report the column every time the pointer was inside it, which is
      // always.
      collisionDetection={closestCorners}
      accessibility={{ announcements, screenReaderInstructions }}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
      onDragCancel={handleDragCancel}
    >
      {/* THE BOARD IS THE SCROLLER, in BOTH axes, and both are stated (D22).
          When five columns will not fit even at their floor width, the board
          scrolls sideways; when a column is taller than the window, the board
          scrolls down. A column never scrolls inside itself, which is what "no
          horizontal scroll inside a column" means.

          Until S3-01 this element said `overflow-x-auto` and nothing else, and
          the comment here claimed the vertical axis was untouched. It was not:
          per CSS Overflow 3 an `overflow-x` of `auto` against an `overflow-y` of
          `visible` PROMOTES `overflow-y` to `auto`, so the board was already
          clipping and scrolling vertically — silently, and contrary to its own
          documentation. Both axes are written down now, so the next reader is
          told the truth by the code rather than by a spec rule nobody reads.

          `h-full` is what makes either axis able to scroll at all: without a
          definite height the board is as tall as its tallest column and there is
          no overflow to scroll. It is a real height because App.tsx's <main>
          carries `min-h-0` over style.css's html/body/#root chain.

          RULED OUT, so nobody rediscovers it (D22): the "latched scrollLeft"
          theory. The columns are `flex-1 basis-0`, so `scrollWidth` tracks
          `clientWidth` and the user agent clamps the offset on resize. There is
          deliberately no scrollLeft/scrollTop handling here and none is wanted. */}
      <div
        ref={boardRef}
        onKeyDown={handleKeyDown}
        onFocus={handleFocus}
        onBlur={handleBlur}
        className="flex h-full min-w-0 items-start gap-2 overflow-x-auto overflow-y-auto"
      >
        {board.map((column) => (
          <Column key={column.status} column={column} rovingNodeId={roving} />
        ))}
      </div>
    </DndContext>
  );
}

export default Kanban;
