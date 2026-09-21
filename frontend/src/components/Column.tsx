import { useTranslation } from 'react-i18next';

import type { ColumnView } from '../lib/client';
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

export function Column({ column, rovingNodeId }: ColumnProps) {
  const { t } = useTranslation();

  const heading = t(`board.column.name.${column.status}`, { defaultValue: column.status });

  return (
    <section
      aria-label={heading}
      data-column={column.status}
      className="flex min-w-36 flex-1 basis-0 flex-col gap-2 rounded-lg border border-line bg-surface p-2 shadow-sm backdrop-blur-glass"
    >
      {/* A plain div, and deliberately not a header element: a header nested
          in a section maps to the `banner` landmark in some accessibility
          trees, and five banners on one page is five wrong landmarks. Region 1
          of the shell is the page header, and there is exactly one of it. */}
      <div className="flex min-w-0 items-baseline justify-between gap-2">
        <h2 className="min-w-0 break-words text-ink">{heading}</h2>
        <span className="shrink-0 font-mono text-muted">
          {t('board.column.cardCount', { count: column.nodes.length })}
        </span>
      </div>

      {column.nodes.length === 0 ? (
        <p className="min-w-0 break-words text-muted">{t('board.column.empty')}</p>
      ) : (
        <ul className="flex min-w-0 flex-col gap-2">
          {column.nodes.map((view) => (
            <li key={view.node.id} className="min-w-0">
              <Card view={view} tabIndex={view.node.id === rovingNodeId ? 0 : -1} />
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

export default Column;
