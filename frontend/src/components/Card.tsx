import { useTranslation } from 'react-i18next';

import { optional, type NodeView } from '../lib/client';
import { formatNumber } from '../lib/format';
import { DueBadge } from './DueBadge';
import { PriorityChip } from './PriorityChip';
import { ProgressBar } from './ProgressBar';
import { TagList } from './TagList';
import { TimerDot } from './TimerDot';
import { TypeIcon } from './TypeIcon';

// Nexus — one card, and the clearest statement in the codebase of the rule the
// whole project is built on: THE FRONTEND COMPUTES NOTHING.
//
// Everything this card shows is a field internal/service/dto.go already filled
// in, against an injected clock, in Go:
//
//   view.status            the DERIVED status (D2) — the column it belongs in
//   view.progress          done/total/percent/defined (D7, D11)
//   view.overdue           due < today AND status != done (D1)
//   view.isLeaf            bar-or-timer, dto.go's own words
//   view.timer.running     the single global timer (S1-19)
//   view.tags              already joined
//
// Not one of them is recomputed, and `make guard` greps for the attempt. If a
// card ever needs something that is not on the DTO, the fix is a Go change.
//
// # Why the stored status is never read here
//
// `view.node.status` is on the row and is MEANINGLESS FOR A PARENT — dto.go says
// so in as many words. The card never reads it; the board places the card using
// `view.status`, the derived one. K2 is the standing reminder of what a stale
// stored status does when something trusts it.
//
// # Layout, and Russian
//
// Russian runs about 30% wider than English, so nothing on this card has a fixed
// width and nothing is `whitespace-nowrap`. The title wraps on word boundaries
// (`break-words`), the chip row wraps (`flex-wrap`), every flex child that could
// be squeezed carries `min-w-0` so it may actually shrink instead of forcing the
// column wider, and the short mono fragments — the chip, the date, the estimate
// — are `shrink-0` so they never break mid-designator. A layout that only works
// in English is a broken layout.
//
// Surfaces come from the tokens and there is no palette conditional anywhere:
// `bg-surface` + `backdrop-blur-glass` is translucent under Aurora, and under
// Studio `--blur` is 0px, which makes the same class a no-op while `shadow-sm`
// — `none` in Aurora — gives Studio its edge. One markup, two palettes.

export interface CardProps {
  /** One NodeView from Board(), rendered exactly as it arrived. */
  view: NodeView;
  /**
   * 0 for the one card the board's roving tabindex is on, -1 for every other.
   *
   * Undefined leaves the card out of the focus model entirely, which is what a
   * caller that is not the board wants.
   */
  tabIndex?: number;
  /**
   * The id of the element that says how this card is dragged, or undefined when
   * nothing has made it draggable.
   *
   * dnd-kit renders that element — one hidden node per DndContext, holding the
   * TRANSLATED instructions views/Kanban.tsx hands it — and publishes its id.
   * It arrives as a prop rather than out of a dnd-kit hook here because this
   * card is not the draggable: components/Column.tsx is, and this is only the
   * focusable element inside it, which is the one an assistive technology can
   * read a description off at all.
   */
  describedBy?: string;
}

export function Card({ view, tabIndex, describedBy }: CardProps) {
  const { t, i18n } = useTranslation();

  // `optional` because the wire sends null where the generator declares `?`.
  // See lib/client.ts; reading it any other way works until the day it does not.
  const due = optional(view.node.due);
  const estimate = optional(view.node.estimateMin);

  return (
    <article
      // The card publishes its id on itself, and that is the ONLY channel the
      // board's delegated key and focus handlers use to find out which card an
      // event came from. No per-card handler, no callback prop threaded through
      // the column, and nothing for a later ticket to forget to pass down.
      data-node-id={view.node.id}
      tabIndex={tabIndex}
      aria-describedby={describedBy}
      className="flex flex-col gap-2 rounded-md border border-line bg-surface p-2 text-ink shadow-sm backdrop-blur-glass"
    >
      <div className="flex min-w-0 items-start gap-2">
        <TypeIcon type={view.node.type} />
        <h3 className="min-w-0 flex-1 break-words">{view.node.title}</h3>
        <TimerDot running={view.timer.running} />
      </div>

      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <PriorityChip priority={view.node.priority} />
        <DueBadge due={due} overdue={view.overdue} />
        {estimate !== null && (
          <span
            aria-label={t('card.estimate.label', {
              value: t('card.estimate.value', { minutes: formatNumber(estimate, i18n.language) }),
            })}
            className="shrink-0 font-mono text-muted"
          >
            {t('card.estimate.value', { minutes: formatNumber(estimate, i18n.language) })}
          </span>
        )}
      </div>

      <TagList tags={view.tags} />

      <ProgressBar progress={view.progress} isLeaf={view.isLeaf} />
    </article>
  );
}

export default Card;
