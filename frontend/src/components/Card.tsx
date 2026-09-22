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
// # Layout, and Russian — and the paragraph that used to be here was WRONG
//
// It said: "nothing on this card has a fixed width and nothing is
// `whitespace-nowrap` ... the short mono fragments — the chip, the date, the
// estimate — are `shrink-0` so they never break mid-designator." Every clause of
// that was TRUE, and the card clipped anyway. `shrink-0` on a max-content
// localised string clips exactly as `whitespace-nowrap` does: it holds the box
// at the width of the longest line the string can produce and lets it run past
// the card edge. `html { font-size: 13px }` makes every Tailwind rem 13/16 of
// nominal, so the card interior at the column floor is about 87px against a due
// badge of roughly 94px in English — IT OVERFLOWED IN ENGLISH, under a comment
// saying it could not, and under a Russian audit that was green on it. That is
// K8, and it is D17 arriving a second time in a second tool.
//
// A comment that is confidently wrong is how a defect survives a review, so this
// one states the RULE rather than a list of the mechanisms somebody happened to
// check: an element whose text comes from i18n, from `formatDate` or from
// `formatNumber` may not refuse to shrink (D20). The title wraps on word
// boundaries (`break-words`), the chip row wraps (`flex-wrap`), every flex child
// carries `min-w-0`, and the mono fragments — the due date and the estimate —
// wrap too rather than being held at max-content. `shrink-0` survives on this
// card only where the box is a FIXED SIZE and holds NO TEXT: TypeIcon and
// TimerDot, whose `h-4 w-4` and `h-2 w-2` are numbers and not strings.
//
// It is checked by a walk over the rendered Russian DOM (src/test/shrink.ts) and
// not by a list of forbidden words, because a list only ever knows about the
// mechanisms somebody already thought of. A layout that only works in English is
// a broken layout — and so is an audit that only knows the utilities we listed.
//
// # Surfaces, and the blur this card NO LONGER carries (K7, D19)
//
// `bg-surface` and `shadow-sm`, and there is no palette conditional anywhere:
// the token is translucent under Aurora and opaque under Studio, and `shadow-sm`
// — `none` in Aurora — is what gives Studio its edge. One markup, two palettes,
// unchanged.
//
// What is gone is the backdrop filter. Aurora's glass utility used to sit here
// too, which meant ONE COMPOSITING LAYER PER CARD and ~50 on a full board;
// resizing broke the layout and it did not recover without a restart. D19's
// allow-list keeps the blur on the five columns and the two overlay scrims and
// takes it off the card, the habit chip and the toast — seven surfaces instead
// of fifty. The card is still translucent under Aurora, because the token is
// what makes it translucent; it simply stops promoting itself.
//
// The evidence is the user's and it is a discriminating experiment rather than a
// theory: with the layout stuck broken, switching the palette to Studio — whose
// `--blur` is 0px — repaired it LIVE, with no restart. `make guard` check 8 is
// what stops the utility quietly coming back here; it is an exact grep, and it
// is the reason this paragraph does not spell the class out, exactly as the
// keys-only ACCEPT test does not spell out the pointing device check 6 forbids.

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
      className="flex flex-col gap-2 rounded-md border border-line bg-surface p-2 text-ink shadow-sm"
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
            className="min-w-0 break-words font-mono text-muted"
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
