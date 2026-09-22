import { useId, useRef, type KeyboardEvent as ReactKeyboardEvent } from 'react';
import { useTranslation } from 'react-i18next';

import { focusWithoutScrolling } from '../lib/focus';
import { changeLanguage, resources, type LanguagePort } from '../lib/i18n';
import { boardActionFor } from '../lib/keyboard';
import { useAppState, useAppStore } from '../store/context';
import { AccentPicker } from './AccentPicker';

// Nexus — region 1 of the shell: palette, theme, language and the accent.
//
// S2-12 made the three appearance values LIVE; this is what a user sets them
// with without knowing the command palette exists. It fills the last empty
// region, so after this commit every component Stage 2 built is reachable from
// main.tsx.
//
// # The frontend does not own any of these values
//
// Every control writes through SettingsService and then shows THE SERVICE'S
// ANSWER (store/settings.ts). Nothing is set optimistically, nothing is
// remembered locally, and a refusal is not "roll back to the previous value" —
// it is a re-read, because the service is the only thing that knows whether a
// write was refused, normalised or partly applied. `aria-checked` below is
// therefore a comparison against `settings`, which is the last thing Go said.
//
// # Where the three sets come from
//
// From the LABEL TABLES in locales/ — `settings.palette`, `settings.theme` and
// `settings.language` — read as key sets. That is S2-19's precedent for the
// quick-add type list and S2-15's principle for the column headings: a label
// table is presentation, a list in code is a rule. Nothing in Go publishes
// "which palettes exist" over the wire (`domain.Palettes()` exists but no
// binding returns it), and adding a binding is a Go change this ticket may not
// make. lib/commands.ts reads the same two tables for the palette's theme and
// language rows; two readers of one table is not two spellings of a rule.
//
// # Toggle buttons, and deliberately not a radio group
//
// Two things follow from `role="radio"` that are wrong here. The ARIA radio
// pattern selects whatever the arrows land on, and every selection here is a
// write to Go that repaints the whole window — so arrowing across a group would
// fire a write, a repaint and possibly a toast per keystroke. And a second
// `radiogroup` on the page collides with the quick add's type selector, which
// is a real one: `getAllByRole('radio')` would stop meaning "the types".
//
// So each control is a plain button carrying `aria-pressed`, described by the
// group's own visible label through `aria-describedby`, which is what makes
// "Dark" announce as the theme rather than as a word on its own. The arrows
// move FOCUS and Space or Enter selects. The chords come out of lib/keyboard.ts
// like every other chord in this project; the group is a single row, so the
// map's adjacent-column chord is the same physical movement in it, exactly as
// in the habits strip and the quick add's type row.
//
// # Russian
//
// "Appearance", "Palette", "Theme" and "Language" are all about a third longer
// in Russian, and there are four groups of them across one strip. So the header
// wraps at every level — the strip wraps its groups, each group wraps its
// buttons, and every label breaks on word boundaries. There is no fixed width
// anywhere except the accent field, whose content is a seven-character colour
// in both languages.

const LABELS = resources.en.translation.settings;

const PALETTES = Object.keys(LABELS.palette);
const THEMES = Object.keys(LABELS.theme);
const LANGUAGES = Object.keys(LABELS.language);

export function AppearanceControls() {
  const { t, i18n } = useTranslation();
  const store = useAppStore();
  const settings = useAppState((state) => state.settings);

  // The language goes through lib/i18n.ts's `changeLanguage`, which holds the
  // rule: the UI switches to the language the PORT REPORTS BACK, not the one it
  // was asked for. The port resolves rather than rejects on a refusal because
  // store/settings.ts has already raised the one toast and re-read, and a
  // second toast for one press is noise.
  const persistLanguage: LanguagePort = async (value) => {
    const accepted = await store.getState().setLanguage(value);
    return accepted?.language ?? i18n.language;
  };

  return (
    <header
      aria-label={t('settings.label')}
      className="flex min-w-0 flex-wrap items-center gap-x-4 gap-y-2"
    >
      <Choice
        group="palette"
        name={t('settings.group.palette')}
        options={PALETTES}
        current={settings?.palette}
        label={(value) => t(`settings.palette.${value}`)}
        onChoose={(value) => void store.getState().setPalette(value)}
      />

      <Choice
        group="theme"
        name={t('settings.group.theme')}
        options={THEMES}
        current={settings?.theme}
        label={(value) => t(`settings.theme.${value}`)}
        onChoose={(value) => void store.getState().setTheme(value)}
      />

      <Choice
        group="language"
        name={t('settings.group.language')}
        options={LANGUAGES}
        current={settings?.language}
        label={(value) => t(`settings.language.${value}`)}
        onChoose={(value) => void changeLanguage(i18n, value, persistLanguage)}
      />

      <AccentPicker />
    </header>
  );
}

interface ChoiceProps {
  /** Which setting this is, published on each button for tests. */
  group: string;
  /** The translated group name, shown and used as the group's accessible name. */
  name: string;
  /** The values, in the label table's own order. */
  options: string[];
  /** The value in force, as the service last reported it, or undefined. */
  current: string | undefined;
  label: (value: string) => string;
  onChoose: (value: string) => void;
}

/**
 * One setting as a row of toggle buttons: one tab stop, arrows inside it.
 *
 * A roving tabindex rather than n tab stops, for the reason the board and the
 * strip both give — a header of four groups would otherwise be nine tab stops
 * between the user and anything they came here to do.
 *
 * The tabindex sits on the value IN FORCE, so tabbing back into the group lands
 * on what is currently selected rather than wherever the arrows were left. When
 * the service has not answered yet, it sits on the first option; the group is
 * still usable, and nothing is drawn as pressed, because nothing is known to be.
 */
function Choice({ group, name, options, current, label, onChoose }: ChoiceProps) {
  const groupRef = useRef<HTMLDivElement>(null);
  const nameId = useId();

  const roving = current !== undefined && options.includes(current) ? current : options[0];

  const moveFocus = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    const action = boardActionFor(event.nativeEvent);
    if (action !== 'nextColumn' && action !== 'previousColumn') {
      return;
    }

    event.preventDefault();

    const buttons = [...(groupRef.current?.querySelectorAll<HTMLElement>('[data-setting]') ?? [])];
    const at = buttons.indexOf(document.activeElement as HTMLElement);
    const step = action === 'nextColumn' ? 1 : -1;

    // Clamped, never wrapped — the same rule the board and the strip follow.
    focusWithoutScrolling(buttons[Math.min(Math.max(at + step, 0), buttons.length - 1)]);
  };

  return (
    <div ref={groupRef} onKeyDown={moveFocus} className="flex min-w-0 flex-wrap items-center gap-2">
      <span id={nameId} className="min-w-0 break-words text-muted">
        {name}
      </span>

      {options.map((value) => (
        <button
          key={value}
          type="button"
          data-setting={`${group}:${value}`}
          aria-describedby={nameId}
          aria-pressed={value === current}
          tabIndex={value === roving ? 0 : -1}
          onClick={() => onChoose(value)}
          className={`min-w-0 break-words rounded-sm border px-2 py-1 transition-colors duration-fast ${
            value === current
              ? 'border-accent bg-accent text-on-accent'
              : 'border-line bg-surface text-muted'
          }`}
        >
          {label(value)}
        </button>
      ))}
    </div>
  );
}

export default AppearanceControls;
