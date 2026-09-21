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
      className={`shrink-0 rounded-sm border px-1.5 font-mono ${chip.className}`}
    >
      {chip.label}
    </span>
  );
}

export default PriorityChip;
