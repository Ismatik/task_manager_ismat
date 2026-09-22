import { useTranslation } from 'react-i18next';

import type { ProgressView } from '../lib/client';

// Nexus — the progress slot, which has exactly three states (D15).
//
// # The three states, and the two DTO flags that choose between them
//
//   progress.defined === false     the localised "empty project" marker
//   isLeaf === true                nothing at all
//   otherwise                      the bar, at progress.percent
//
// Both conditions are plain fields on NodeView. Nothing is computed here: not
// the percentage (internal/service/dto.go rounds it in Go, once, "because two
// implementations of a rounding rule are two rounding rules"), not the leaf-ness
// (dto.go: IsLeaf "is what the UI needs to decide between a progress bar and a
// timer button"), and not the defined-ness.
//
// # Why `defined` is tested FIRST, and why the marker has no column in it
//
// D15 closes K3: an empty project whose derived status is the finished one sat
// in that column with nothing on it at all — no bar, because its progress is genuinely
// undefined (D11), and no other chrome. The fix is a marker, and the condition
// is `defined`, NOT `status === done`: the same project in Backlog is blank in
// exactly the same way, and a rule that fired only in one column would be a rule
// with a column in it. So the marker fires in every column, and this component
// never learns which column it is in.
//
// It is never a 0% bar and never a "0/0": `defined === false` exists precisely
// because neither 0% nor 100% is true (D7, D9). Drawing an empty track is
// rendering 0% with extra steps.
//
// # Why a work leaf gets nothing rather than a 1-unit bar
//
// Go measures a leaf task as one work leaf, so its progress IS defined — 0% or
// 100%, and nothing in between is reachable. A bar that can only ever be empty
// or full duplicates what the card's column already says, on every card on the
// board. D15's table says a non-project renders nothing in this slot, and
// `isLeaf` is the DTO's way of saying that without the frontend holding a list
// of types.
//
// (A note or a habit is also a leaf with undefined progress, and would take the
// marker branch — but neither ever reaches a card: Board() excludes every type
// with no Kanban column. When Stage 3's tree view renders them, that is the
// ticket that decides what they show, with a flag from Go rather than a guess
// here.)
//
// The width is an inline style because it is a value, not a class: Tailwind
// cannot emit a percentage it has never seen. No colour is inline — the track is
// `line`, the fill is `accent`, both tokens.

export interface ProgressBarProps {
  /** `progress`, exactly as Go computed it. */
  progress: ProgressView;
  /** `isLeaf`, the flag dto.go documents as the UI's bar-or-timer switch. */
  isLeaf: boolean;
}

export function ProgressBar({ progress, isLeaf }: ProgressBarProps) {
  const { t } = useTranslation();

  if (!progress.defined) {
    return (
      // `break-words`, never `truncate`: the Russian marker is longer than the
      // English one and a slot that clipped it would be the RU-width bug this
      // stage exists to avoid.
      <p className="min-w-0 break-words text-muted" data-testid="progress-empty">
        {t('card.progress.empty')}
      </p>
    );
  }

  if (isLeaf) {
    return null;
  }

  const ratio = t('card.progress.ratio', { done: progress.done, total: progress.total });

  return (
    <div className="flex items-center gap-2">
      <div
        role="progressbar"
        aria-label={ratio}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-valuenow={progress.percent}
        className="h-1 min-w-0 flex-1 overflow-hidden rounded-sm bg-line"
      >
        <div
          className="h-full rounded-sm bg-accent transition-[width] duration-base"
          style={{ width: `${progress.percent}%` }}
        />
      </div>
      {/* `min-w-0 break-words` (K8, D20). `ratio` is i18n output — "3 of 5",
          "3 из 5" — so it is a localised string and may not be held at
          max-content. It is NOT in S3-02's enumerated list of four; the rewritten
          audit found it, which is the point of changing the audit's shape. */}
      <span className="min-w-0 break-words font-mono text-muted">{ratio}</span>
    </div>
  );
}

export default ProgressBar;
