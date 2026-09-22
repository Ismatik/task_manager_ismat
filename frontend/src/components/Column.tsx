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
// Surfaces are token-only and palette-blind: `bg-surface` + `backdrop-blur-glass`
// is translucent under Aurora, a no-op under Studio (whose `--blur` is 0px),
// where `shadow-sm` supplies the edge instead.
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
      // `z-10` lifts the card being dragged over the ones it passes.
      className={`min-w-0 touch-none ${isDragging ? 'z-10' : ''}`}
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
      className={`flex min-w-36 flex-1 basis-0 flex-col gap-2 rounded-lg border p-2 shadow-sm backdrop-blur-glass transition-colors duration-fast ${
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
          <ul className="flex min-w-0 flex-col gap-2">
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
