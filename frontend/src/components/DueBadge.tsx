import { useTranslation } from 'react-i18next';

import { formatDate } from '../lib/format';

// Nexus — the due date, and the red that says it has passed.
//
// # The red comes from Go, and only from Go
//
// `overdue` is a field on NodeView. PLAN.md section 4 defines it as
// `due < today AND status != done`, it is computed in internal/domain against
// an injected clock, and NOTHING in this file compares a date to anything. That
// is not a stylistic preference: the client's idea of the current day is the
// WebView's idea of midnight in the WebView's timezone, the derived status is not on the row at
// all, and a second implementation of the rule would be a second rule.
//
// Card.test.tsx pins this the only way that actually proves it: it renders a
// card with `overdue: true` and a due date years in the FUTURE, and asserts the
// badge is still `danger`. A component that decided for itself would render it
// muted and fail.
//
// The date itself is formatted by lib/format.ts — one date format in the whole
// application — in `font-mono`, per PLAN.md section 3.

export interface DueBadgeProps {
  /** `node.due`, the "YYYY-MM-DD" string, or null when there is none. */
  due: string | null;
  /** `overdue`, the flag Go computed. Never recomputed here. */
  overdue: boolean;
}

export function DueBadge({ due, overdue }: DueBadgeProps) {
  const { t, i18n } = useTranslation();

  if (due === null) {
    return null;
  }

  const shown = formatDate(due, i18n.language);

  return (
    <span
      aria-label={t(overdue ? 'card.due.overdue' : 'card.due.label', { date: shown })}
      // `min-w-0 break-words`, and NOT `shrink-0` (K8, D20). This badge is the
      // defect's own example: `html { font-size: 13px }` makes every Tailwind
      // rem 13/16 of nominal, so a column at `min-w-36` is 117px and the card
      // interior is about 87px — against a max-content date of roughly 94px in
      // English and 125px in Russian. `shrink-0` held it at max-content, so it
      // ran past the card edge IN ENGLISH TOO, which is why the Russian audit
      // was green on it. `break-words` matters as much as `min-w-0` here: a
      // formatted date has no space to wrap at, so without it the box shrinks
      // and the text overflows anyway. A date on two lines is ugly; a date
      // sliced off at the column edge is unreadable.
      className={`min-w-0 break-words font-mono ${overdue ? 'text-danger' : 'text-muted'}`}
    >
      {shown}
    </span>
  );
}

export default DueBadge;
