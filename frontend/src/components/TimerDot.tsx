import { useTranslation } from 'react-i18next';

// Nexus — the "a timer is running on this card" indicator.
//
// `running` is TimerView.Running, which Go sets on exactly one node in the whole
// application (the single-active invariant, S1-19, seen from the read side).
// This component asks no further question: not whether the node is timeable, not
// whether it is a leaf, not how long it has been going.
//
// # Why it shows no elapsed time
//
// It could: TimerView carries `elapsedSeconds`. But a number that was true when
// Board() answered and never moves again is worse than no number — it reads as a
// stopped clock. The store already owns the honest version
// (`displayElapsedSeconds`, which advances Go's value against an injected clock
// between reads), and wiring a per-second ticker into a card is not this
// ticket's. So the card says THAT the timer is running, in `accent`, which is
// the token design/README.md assigns to a running timer, and says it to screen
// readers too.

export interface TimerDotProps {
  /** `timer.running`, exactly as Go reported it. */
  running: boolean;
}

export function TimerDot({ running }: TimerDotProps) {
  const { t } = useTranslation();

  if (!running) {
    return null;
  }

  return (
    <span
      role="img"
      aria-label={t('card.timer.running')}
      className="inline-block h-2 w-2 shrink-0 rounded-sm bg-accent"
    />
  );
}

export default TimerDot;
