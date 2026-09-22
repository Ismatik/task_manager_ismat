import { useCallback, useRef, useState, type KeyboardEvent as ReactKeyboardEvent } from 'react';
import { useTranslation } from 'react-i18next';

import type { HabitView } from '../lib/client';
import { focusWithoutScrolling } from '../lib/focus';
import { boardActionFor, matches, KEYS, type BoardAction } from '../lib/keyboard';
import { useAppState, useAppStore } from '../store/context';
import { HabitChip } from './HabitChip';

// Nexus — region 2 of the shell: the habits strip.
//
// # Habits are not cards, and this file does not say so
//
// PLAN.md §4: a habit has no Kanban column, so `Board()` already excludes it and
// there is nothing here to filter. A `type === 'habit'` test in the frontend
// would be a second spelling of a rule domain.NodeType already owns — the rule
// would then live in two places, and TASKS.md's engineering note says what
// happens next. So the strip renders `HabitStrip()`'s answer and the board
// renders `Board()`'s, and neither one asks what the other is holding.
//
// # What IS a choice here
//
// Whether to show a habit that is not scheduled today. That is presentation
// over S2-04's `scheduledToday` flag, not a recomputation of the schedule: the
// strip shows every habit Go returned and DIMS the ones Go says are not due,
// so a Monday habit is still visible on a Tuesday and still says what it is.
// Hiding them would make the strip's contents change shape under the user
// overnight for reasons the screen never explains.
//
// # The keyboard, and where its keys come from
//
// The strip is ONE tab stop with a roving tabindex, like the board. Within it
// the arrows move between chips and Space toggles today's check. Every one of
// those chords is read out of lib/keyboard.ts — S2-16's normative map — through
// `boardActionFor` and `KEYS.toggleHabit`, and not one of them is written down
// again here. The strip is a single ROW, so the map's "adjacent column" chord
// (ArrowLeft / ArrowRight) is the same physical movement in it, and Home / End
// are its two ends; what those actions are CALLED is the board's vocabulary,
// which is a small price for having the keys themselves in one place.
//
// Focus is the truth, exactly as on the board: `document.activeElement` decides
// which chip is focused and the local `focusedId` only mirrors it, so that the
// roving tabindex can be a property of the render. It is component state rather
// than store state because nothing outside this strip has an opinion about it —
// the store's `selectedNodeId` is the BOARD's selection and giving it a second
// meaning would be the same defect in a different file.

/** The chip a navigation action moves focus to, or null when there is nowhere to go. */
function nextChipId(habits: HabitView[], habitId: string, action: BoardAction | null): string | null {
  const at = habits.findIndex((habit) => habit.node.id === habitId);
  if (at === -1) {
    return null;
  }

  switch (action) {
    case 'nextColumn':
      // Clamped, never wrapped — arriving back at the first habit by holding
      // ArrowRight is a surprise, and a surprise in a keyboard model is a bug
      // report. Same rule as the board's.
      return habits[Math.min(at + 1, habits.length - 1)].node.id;
    case 'previousColumn':
      return habits[Math.max(at - 1, 0)].node.id;
    case 'firstCard':
      return habits[0].node.id;
    case 'lastCard':
      return habits[habits.length - 1].node.id;
    default:
      return null;
  }
}

export function HabitStrip() {
  const { t } = useTranslation();
  const store = useAppStore();
  const habits = useAppState((state) => state.habits);
  const [focusedId, setFocusedId] = useState<string | null>(null);

  const stripRef = useRef<HTMLDivElement>(null);

  // A scan rather than a `[data-habit-id="..."]` selector: a node id is a uuid
  // today, and an attribute selector that has to be escaped is a bug waiting
  // for the first id with a quote in it. Kanban.tsx says the same of cards.
  const focusChip = useCallback((habitId: string) => {
    for (const chip of stripRef.current?.querySelectorAll<HTMLElement>('[data-habit-id]') ?? []) {
      if (chip.dataset.habitId === habitId) {
        focusWithoutScrolling(chip);
        return;
      }
    }
  }, []);

  // Null before the first read, empty when there are no habits. Either way the
  // region emits NO DOM AT ALL — S2-15's rule: an empty region is not a blank
  // bar across the screen and not a placeholder somebody has to remember to
  // delete. A failed read is the first case too: the store leaves `habits` null
  // and raises a toast, so the user sees the shell and an error.
  if (habits === null || habits.length === 0) {
    return null;
  }

  const roving = habits.some((habit) => habit.node.id === focusedId)
    ? focusedId
    : habits[0].node.id;

  /** The chip an event came from, by the id the chip publishes on itself. */
  const chipIdFrom = (target: EventTarget | null): string | null => {
    if (!(target instanceof Element)) {
      return null;
    }
    return target.closest<HTMLElement>('[data-habit-id]')?.dataset.habitId ?? null;
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    const habitId = chipIdFrom(event.target);
    if (habitId === null) {
      return;
    }

    if (matches(event.nativeEvent, KEYS.toggleHabit)) {
      // preventDefault because Space scrolls the page otherwise, which is the
      // one thing a toggle must not also do.
      event.preventDefault();
      void store.getState().toggleHabit(habitId);
      return;
    }

    const next = nextChipId(habits, habitId, boardActionFor(event.nativeEvent));
    if (next === null) {
      return;
    }

    event.preventDefault();
    focusChip(next);
  };

  return (
    // role="group" and deliberately not a <section> with a name, which would be
    // an ARIA `region` landmark. The five columns are already regions, and a
    // sixth one sitting above them would make "the board's columns, in order"
    // a query that quietly answers six. The strip is a group of controls, not a
    // landmark.
    <div
      ref={stripRef}
      role="group"
      aria-label={t('habits.label')}
      onKeyDown={handleKeyDown}
      onFocus={(event) => setFocusedId(chipIdFrom(event.target))}
      // `flex-wrap` rather than a scroller: in Russian these labels are about a
      // third wider, and a strip that overflows sideways hides habits behind an
      // edge the keyboard has already walked past.
      className="flex min-w-0 flex-wrap items-center gap-2"
    >
      {habits.map((habit) => (
        <HabitChip
          key={habit.node.id}
          habit={habit}
          tabIndex={habit.node.id === roving ? 0 : -1}
          onToggle={(nodeId) => void store.getState().toggleHabit(nodeId)}
        />
      ))}
    </div>
  );
}

export default HabitStrip;
