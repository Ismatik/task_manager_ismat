import type { CSSProperties } from 'react';
import { useDndContext, useDroppable } from '@dnd-kit/core';
import { SortableContext, useSortable, verticalListSortingStrategy } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { useTranslation } from 'react-i18next';

import { prefersReducedMotion } from '../lib/appearance';
import type { ColumnView, NodeView } from '../lib/client';
import { useAppState } from '../store/context';
import { Card } from './Card';

// Nexus — one Kanban column.
//
// # It holds no list of statuses, and no opinion about order
//
// The column is handed a ColumnView that Board() produced: the status it stands
// for, and its cards already in sort_order. It does not sort, does not filter,
// does not know how many columns there are and does not know what the five are
// called. `column.status` is an opaque string that came from Go and is used for
// exactly one thing — looking up a LABEL.
//
// That lookup is the one authorised place a status name may be written in the
// frontend, and it is in locales/, not in code: a label table is presentation, a
// status list in code is a rule. `make guard` check 2 enforces the difference by
// refusing a quoted status anywhere in frontend/src except the locale files.
//
// The `defaultValue` is deliberate. If Go ever returns a column this build has
// no label for, the heading reads the raw status rather than going blank —
// honest, and impossible to miss.
//
// # Width, and Russian
//
// A fixed width would be the RU bug in its purest form — the Russian for "This
// week" is half again as long. So the columns SHARE the board's width
// (`flex-1 basis-0`) down to a legible floor (`min-w-36`) and the board scrolls
// only once even that will not fit; the heading wraps on word boundaries. There
// is no `w-` anywhere and nothing truncates.
//
// # Height, and why an empty column is a drop target at all (S3-31, K15, D29)
//
// There is NO height in this file. The section is a flex item of the board's
// row, and since S3-31 removed the board's `items-start` it stretches to the
// board's full height — derived from the shell's chain, never typed, which is
// D21's rule and D22's chain continued rather than re-spelt.
//
// That is not decoration, and the screenshot that found it makes the case
// better than a paragraph can: with `items-start` the column ended where its
// cards ended, so `useDroppable` below — which exists precisely to catch a card
// dropped on an EMPTY column or on the padding under the last card — had a
// header-high strip to work with, and the large area beneath belonged to the
// board, which is not a drop target. Stretching the section IS the drop target.
//
// The scroll goes with it. `min-h-0` here and on the `<ul>`, plus `flex-1` on
// the `<ul>`, make the CARD LIST the vertical scroller: a column with more
// cards than fit scrolls inside itself, under a heading that stays put because
// the heading is the list's sibling rather than its child. Before S3-31 nothing
// scrolled inside a column at all and the board absorbed it, which took every
// column's heading out of view at once.
//
// Surfaces are token-only and palette-blind: `bg-surface` + `backdrop-blur-glass`
// is translucent under Aurora, a no-op under Studio (whose `--blur` is 0px),
// where `shadow-sm` supplies the edge instead.
//
// The column is one of only THREE files allowed to say `backdrop-blur-glass` at
// all (K7, D19): here, and the two full-screen overlay scrims. It keeps it
// because it is a LARGE surface and there are exactly five of them — the card,
// the habit chip and the toast lost it because there is one per card, one per
// habit and one per toast, which is how ~50 compositing layers came to exist at
// once and how a resize came to break the layout until the app was restarted.
// `make guard` check 8 is an exact grep over that allow-list, in BOTH
// directions: 8a fails if the class appears anywhere else, and 8b fails if it
// stops appearing here. Only the first half existed until S3-04 was re-opened,
// and with only that half, deleting the blur from this file was green
// everywhere — all five gates, the guard and the whole vitest suite.
//
// # The drag half (S2-17), and the two places it deliberately stops
//
// The column is a DROP TARGET (`useDroppable`, keyed by the status Go sent) and
// its list is a sortable container (`SortableContext`, with the same key). Each
// card is wrapped in a `<li>` that is the draggable. Two things about that are
// decisions rather than defaults:
//
// 1. **`useSortable`'s `attributes` are NOT spread onto anything.** They force
//    `tabIndex=0` and `role="button"` on whatever wears them, and the board is
//    a ROVING TABINDEX: exactly one card in the whole board is `tabIndex=0` and
//    the rest are -1 (S2-16, normative, asserted). Spreading them would make
//    every card a tab stop and turn five columns of cards into n+5 of them —
//    the precise thing the roving tabindex exists to prevent. `aria-describedby`
//    is the one attribute taken out of that bag, by hand, because it points at
//    dnd-kit's hidden instruction text and that text is worth having.
//    `aria-roledescription` is left behind with the rest: its default value is
//    the hard-coded English string "draggable".
// 2. **The `<li>` is the draggable, not the `<article>` inside it.** The card
//    keeps `data-node-id` and its tabindex and gains no handlers, so the board's
//    delegated key and focus handlers are untouched by any of this.
//
// Motion is asked for explicitly rather than assumed. design/tokens.css already
// kills every CSS transition under `prefers-reduced-motion`, but dnd-kit's
// transition is an INLINE style, and the value here is read from the same
// `prefersReducedMotion` the drift gate uses, so the two cannot disagree.

export interface ColumnProps {
  /** One ColumnView from Board(), rendered exactly as it arrived. */
  column: ColumnView;
  /**
   * The one card on the WHOLE BOARD that carries `tabIndex=0`.
   *
   * The board is a single tab stop and the arrows move inside it, so every
   * other card is `-1`: focusable from script, invisible to Tab. The decision
   * is the board's (lib/keyboard.ts `rovingNodeId`); the column only passes it
   * on, which is why it is an id and not a per-column flag.
   */
  rovingNodeId: string | null;
}

/**
 * What a column publishes about itself on every drag target inside it.
 *
 * Both the column and each of its cards carry a copy, so whatever a drag is
 * over can be asked which column it belongs to without anybody searching the
 * board for it. `status` is what the column compares to light itself up;
 * `column` is the heading, which is what the announcements say out loud —
 * looked up once, here, by the component that already had to do it.
 */
interface ColumnIdentity {
  status: string;
  column: string;
}

interface SortableCardProps {
  view: NodeView;
  identity: ColumnIdentity;
  tabIndex: number;
}

/** One card, wrapped in the `<li>` that is the thing a pointer actually drags. */
function SortableCard({ view, identity, tabIndex }: SortableCardProps) {
  const appWindow = useAppState((state) => state.view);
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: view.node.id,
    data: { ...identity, title: view.node.title },
  });

  const style: CSSProperties = {
    transform: CSS.Translate.toString(transform),
    // `undefined` and not `'none'`: removing the declaration lets the stylesheet
    // have the last word, which is what design/tokens.css's reduced-motion rule
    // is for.
    transition: prefersReducedMotion(appWindow) ? undefined : (transition ?? undefined),
  };

  return (
    <li
      ref={setNodeRef}
      style={style}
      // `touch-none` because a pointer drag and a touch scroll are the same
      // gesture until one of them wins, and the browser wins by default.
      //
      // `opacity-0` while dragging, and deliberately NOT `hidden` (K6, D18): the
      // card is drawn in views/Kanban.tsx's DragOverlay from the moment it is
      // lifted, so two copies would otherwise be on screen at once. Opacity
      // keeps the item's SPACE, so the list does not jump on grab and the gap
      // the card came out of is the gap it drops back into.
      //
      // The z-index that used to sit here is gone, and the criterion is that a
      // grep for it over this file finds nothing — hence no literal in this
      // paragraph. It existed to lift the dragged card over the ones it passed,
      // and it could never do that: z-index orders siblings within ONE stacking
      // context, and the neighbouring column's backdrop-filter makes its own.
      // Leaving it would be leaving a false explanation in the code. The overlay
      // is what solves it, by painting outside every column.
      className={`min-w-0 touch-none ${isDragging ? 'opacity-0' : ''}`}
      {...listeners}
    >
      <Card view={view} tabIndex={tabIndex} describedBy={attributes['aria-describedby']} />
    </li>
  );
}

export function Column({ column, rovingNodeId }: ColumnProps) {
  const { t } = useTranslation();

  const heading = t(`board.column.name.${column.status}`, { defaultValue: column.status });
  const identity: ColumnIdentity = { status: column.status, column: heading };

  // The column itself is a drop target, keyed by the status Go sent. It is what
  // catches a card dropped on an EMPTY column, and on the padding of a full
  // one — the two places where there is no card underneath to drop onto.
  const { setNodeRef } = useDroppable({ id: column.status, data: identity });

  // The highlight follows the DRAG, not this droppable: a card hovering over
  // another card is over that card, and `useDroppable`'s own `isOver` would be
  // false for the column holding both — so the column under the pointer would
  // stay dark for the whole gesture and only light up in the gaps between its
  // cards. Asking what the drag is over, and reading the column off ITS OWN
  // published identity, is the same channel the announcements use and needs no
  // second way of turning a drop target into a column.
  const { over } = useDndContext();
  const isOver = over?.data.current?.status === column.status;

  return (
    <section
      ref={setNodeRef}
      aria-label={heading}
      data-column={column.status}
      data-over={isOver}
      // The drop-target highlight, and every colour in it is a token name:
      // `accent` is what design/README.md assigns to the active thing on
      // screen, and `elevated` is the surface one step up. Nothing here is a
      // hex literal and nothing here knows which column it is.
      // `min-h-0` is the column's link in the vertical chain (S3-31, D29). The
      // board stretches this section to its full height — that is what gives
      // `useDroppable` above a rectangle to catch a drop in — and this section
      // is itself a COLUMN flex container, so its card list below cannot shrink
      // under `min-height: auto` unless the release is written at every step
      // between the two. There is deliberately no height here: the height comes
      // from the board, derived, never typed (D21).
      className={`flex min-h-0 min-w-36 flex-1 basis-0 flex-col gap-2 rounded-lg border p-2 shadow-sm backdrop-blur-glass transition-colors duration-fast ${
        isOver ? 'border-accent bg-elevated' : 'border-line bg-surface'
      }`}
    >
      {/* A plain div, and deliberately not a header element: a header nested
          in a section maps to the `banner` landmark in some accessibility
          trees, and five banners on one page is five wrong landmarks. Region 1
          of the shell is the page header, and there is exactly one of it. */}
      <div className="flex min-w-0 items-baseline justify-between gap-2">
        <h2 className="min-w-0 break-words text-ink">{heading}</h2>
        {/* `min-w-0 break-words`, never `shrink-0` (K8, D20). This span is what
            the user photographed: "0 карточек" is about 86px at max-content, and
            holding it there starved the heading beside it down to roughly 16px,
            so "Бэклог" wrapped ONE CHARACTER PER LINE. Both children may shrink
            now, and the heading — which is the more important of the two — keeps
            its `break-words` so it wraps on word boundaries first. */}
        <span className="min-w-0 break-words font-mono text-muted">
          {t('board.column.cardCount', { count: column.nodes.length })}
        </span>
      </div>

      {column.nodes.length === 0 ? (
        <p className="min-w-0 break-words text-muted">{t('board.column.empty')}</p>
      ) : (
        // `id` is the status, which makes it the `containerId` every card in
        // this list reports on its drag data — so views/Kanban.tsx reads which
        // column a drag started in and ended over straight off the event,
        // rather than searching the board for the card.
        <SortableContext
          id={column.status}
          items={column.nodes.map((view) => view.node.id)}
          strategy={verticalListSortingStrategy}
        >
          {/* THE CARD LIST IS THE VERTICAL SCROLLER (S3-31, D29), and the
              heading above is its SIBLING, not its child — which is the whole
              of "the column header does not scroll away". A heading that
              scrolled out of view makes a drag across five columns
              unnavigable, and putting the scroll on the section instead would
              have taken the heading with it.

              `min-h-0 flex-1` together are the release and the claim: `flex-1`
              takes the height the heading did not, so the list — and with it
              the space a card can be dropped into — reaches the bottom of the
              column even when there are two cards; `min-h-0` lets it shrink
              below its content, without which `overflow-y-auto` has nothing to
              do because the list never becomes smaller than its cards.

              BOTH AXES ARE NAMED, and the horizontal one is `auto` rather
              than `hidden` because the S3-02 shrink audit argued it down.
              Per CSS Overflow 3 an `overflow-y` of `auto` against an
              `overflow-x` of `visible` promotes the horizontal axis to `auto`
              anyway — the exact promotion S3-01 found on the board, one
              element down — so the moment D29 makes this list the vertical
              scroller it becomes a horizontal scroll container too, whether or
              not anyone writes it down. The only spelling that would prevent
              that is `overflow-x-hidden`, which CLIPS: test/shrink.ts reported
              it here by name, as a mechanism that can cut Russian off mid-word
              with nothing on screen to say so, and permitting it would be a
              deliberate edit to another ticket's file. Silence is worse than
              either, so the computed truth is written out instead.

              Nothing can actually overflow it sideways — every card is
              `min-w-0` and wraps on word boundaries — so no horizontal
              scrollbar appears. What genuinely changed is that "a column never
              scrolls sideways" is now a statement about the SECTION, which has
              no overflow of its own, and not about this list. That follows
              from D29 rather than from a choice made here, and it is reported
              rather than buried. */}
          <ul className="flex min-h-0 min-w-0 flex-1 flex-col gap-2 overflow-x-auto overflow-y-auto">
            {column.nodes.map((view) => (
              <SortableCard
                key={view.node.id}
                view={view}
                identity={identity}
                tabIndex={view.node.id === rovingNodeId ? 0 : -1}
              />
            ))}
          </ul>
        </SortableContext>
      )}
    </section>
  );
}

export default Column;
