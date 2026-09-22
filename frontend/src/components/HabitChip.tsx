import { Check, Flame } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { HabitView } from '../lib/client';
import { formatNumber } from '../lib/format';

// Nexus — one habit in the strip: a checkbox and a streak.
//
// # Every value here is Go's, including the one that looks like arithmetic
//
// `streak` is D5's answer — CONSECUTIVE SCHEDULED OCCURRENCES of the recurrence
// rule that were checked, ending at the present. Not calendar days, and not
// something a count of ticks in TypeScript would arrive at: a habit due on
// Mondays with three Mondays checked has a streak of 3 whatever the calendar
// says in between. So the number is rendered, never computed, and the same goes
// for `checkedToday` and `scheduledToday`. If the strip needs a value that is
// not on HabitView, the fix is a Go change.
//
// # Why a div with role="checkbox" and not <input type="checkbox">
//
// Because Space is in a map that already exists. S2-16's keyboard model is
// normative and lib/keyboard.ts is its one spelling; a native checkbox would
// toggle on Space by itself, which means the binding would live both in the map
// and in the browser, and only one of them would be tested. A non-native
// element has no default Space behaviour, so the strip's delegated handler —
// which asks lib/keyboard.ts — is the only thing that can toggle it.
//
// The chip publishes its id on itself and carries no handler of its own for the
// keyboard, exactly as Card.tsx does: the strip's delegated handlers find out
// which chip an event came from by reading `data-habit-id` off the closest
// ancestor that has one.
//
// # Colours, and the two the handoff actually assigns
//
// design/README.md: `warning` for streak flames, `success` for habit checks.
// Those are the two, and they are the two used. Everything else is the neutral
// `muted`/`ink` pair, because inventing a colour and attributing it to the
// handoff is the thing D6-as-amended forbids.

export interface HabitChipProps {
  /** One HabitView from HabitStrip(), rendered exactly as it arrived. */
  habit: HabitView;
  /**
   * 0 for the one chip in the WHOLE STRIP that is in the tab order, -1 for
   * every other.
   *
   * The strip is a single tab stop and the arrows move inside it, for the same
   * reason the board is: n tab stops across a row of habits is a keyboard
   * obstacle course. The decision is the strip's; the chip only wears it.
   */
  tabIndex: number;
  /** The pointer route to the same toggle the keyboard reaches through Space. */
  onToggle: (nodeId: string) => void;
}

export function HabitChip({ habit, tabIndex, onToggle }: HabitChipProps) {
  const { t, i18n } = useTranslation();

  return (
    <div
      data-habit-id={habit.node.id}
      // `scheduledToday` is published rather than acted on, so a test can prove
      // the flag came from the DTO instead of being worked out here.
      data-scheduled={habit.scheduledToday}
      role="checkbox"
      aria-checked={habit.checkedToday}
      tabIndex={tabIndex}
      onClick={() => onToggle(habit.node.id)}
      // No backdrop filter either (K7, D19): Aurora's glass utility here was one
      // compositing layer per habit, and D19's allow-list is the column and the
      // two overlay scrims and nothing else. `bg-surface` stays, so the chip is
      // still translucent under Aurora — only the filter went. `make guard`
      // check 8 is an exact grep, which is why the class is not spelt out here.
      //
      // No `shrink-0` on the chip (K8, D20). Its width is the width of a
      // habit's TITLE, which is localised, so holding it at max-content is the
      // same defect as the due badge's one level up: in Russian the strip's
      // chips would each demand their full title and push the row past the
      // window. `min-w-0` lets a chip give ground, and the strip's `flex-wrap`
      // means it moves to a new line before it ever has to.
      className={`flex min-w-0 items-center gap-2 rounded-md border border-line bg-surface px-2 py-1 shadow-sm transition-colors duration-fast ${
        habit.scheduledToday ? 'text-ink' : 'text-muted'
      }`}
    >
      {/* Both icons keep `shrink-0`, and D20 says why: a fixed-size, NON-TEXT
          box has a width that is a number rather than a string, so it cannot
          change with the locale. 8px grid — a 16px box is two steps. */}
      <Check
        aria-hidden
        className={`h-4 w-4 shrink-0 ${habit.checkedToday ? 'text-success' : 'text-muted'}`}
      />

      {/* Russian runs ~30% wider, so the title wraps on word boundaries rather
          than being truncated, and every flex child that could be squeezed
          carries min-w-0 so it may actually shrink. */}
      <span className="min-w-0 break-words">{habit.node.title}</span>

      {!habit.scheduledToday && <span className="sr-only">{t('habits.notScheduledToday')}</span>}

      <Flame aria-hidden className="h-4 w-4 shrink-0 text-warning" />
      <span
        // The visible text is a bare number, which reads as nothing out loud.
        // The label says what the number is; the number itself stays a number.
        aria-label={t('habits.streak.label', {
          value: formatNumber(habit.streak, i18n.language),
        })}
        className="min-w-0 break-words font-mono"
      >
        {formatNumber(habit.streak, i18n.language)}
      </span>
    </div>
  );
}

export default HabitChip;
