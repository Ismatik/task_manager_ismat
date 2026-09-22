import { useTranslation } from 'react-i18next';

import { priorityChip } from '../lib/priority';

// Nexus — the priority chip.
//
// It decides nothing. `lib/priority.ts` owns the whole 1..4 -> chip mapping,
// including the fact that 4 has no chip, and this component renders whatever
// that module hands back — or nothing at all when it hands back null.
//
// The label is `font-mono` because PLAN.md section 3 puts JetBrains Mono on
// every number and designator; the colour is the token class the mapping
// supplies, never a class chosen here.

export interface PriorityChipProps {
  /** `node.priority`, the stored 1..4, exactly as Go sent it. */
  priority: number;
}

export function PriorityChip({ priority }: PriorityChipProps) {
  const { t } = useTranslation();

  const chip = priorityChip(priority);

  if (chip === null) {
    return null;
  }

  return (
    <span
      // The chip reads its designator; the accessible name reads it back as a
      // sentence, in the language the card is rendered in.
      aria-label={t('card.priority.label', { chip: chip.label })}
      // THE RULING ON BOUNDED ASCII, stated rather than left to be re-derived
      // (D20). The designator lib/priority.ts hands back is two ASCII
      // characters, is not localised and cannot wrap, so `shrink-0` and the
      // default `min-width: auto` size this box IDENTICALLY — for a single
      // unbreakable token, min-content and max-content are the same width. The
      // class therefore bought nothing, and dropping it is what lets the audit
      // in test/shrink.ts ask ONE unconditional question of every element that
      // carries text. An audit with an exception list is an audit whose
      // exception list grows; this one has none, and that is worth more than a
      // class that did nothing.
      className={`rounded-sm border px-1.5 font-mono ${chip.className}`}
    >
      {chip.label}
    </span>
  );
}

export default PriorityChip;
