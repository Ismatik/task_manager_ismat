import { useId, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';

import { useAppState, useAppStore } from '../store/context';

// Nexus — the accent override, region 1's last control.
//
// # Where the preset comes from, and why there is only one
//
// TASKS.md S2-21: "the presets must not be hex literals in frontend/src. They
// come from tokens or from Go. If that forces the preset list into design/, it
// does not go there - design/ is read-only - so it comes from Go, or the picker
// offers only the free value plus the palette default. Decide it in the ticket
// and say which."
//
// THE DECISION: from the tokens, at runtime. `--accent-2` is read off the live
// document with getComputedStyle, so its value is whatever design/tokens.css
// declares for the palette and theme currently in force, and not one character
// of it is written down here. `git grep -nE '#[0-9a-fA-F]{3,8}' frontend/src`
// stays empty, design/ stays untouched, and the preset follows the palette
// instead of being a snapshot of one.
//
// It is ONE preset and not four because `design/README.md` assigns meanings:
// `danger` is overdue and the two urgent priority chips, `warning` is the third
// chip and streak flames, `success` is done and habit checks. Offering them as
// accents would put a semantic colour somewhere it means something else.
// `accent-2` is the only other token that IS an accent, so it is the only one
// offered. Under Studio it happens to equal `--accent`; that is the palette's
// own choice and not this file's business.
//
// When the value cannot be read - no stylesheet, or a host with no
// getComputedStyle, which is every jsdom test that does not stub one - the
// preset is simply not offered, and the picker degrades to exactly the fallback
// the ticket names: the free value plus the palette default. Nothing is
// invented to fill the gap.
//
// # "Use the palette's accent" is a choice, not the absence of one
//
// D6: "" means "use the palette's own --accent". It is the seeded default, so
// it must be reachable and not a state you can only leave - hence an explicit
// button that writes "". lib/appearance.ts then REMOVES the inline declaration
// rather than setting it to empty, which is the only way the palette gets its
// value back.
//
// # Nothing is validated here
//
// `domain.IsHexColour` is the rule and it lives in Go, asked in the same call
// that would have stored the value. So the free field sends whatever was typed
// and SURFACES THE REFUSAL: store/settings.ts raises the one toast and re-reads,
// and `choose` below puts the service's answer back in the box. A pre-check
// here would be that rule written a second time, and the second copy is the one
// that goes stale.

/** The token the preset is read from. A NAME; its value lives in design/. */
const PRESET_TOKEN = '--accent-2';

/**
 * The value a custom property resolves to on the live document, or ''.
 *
 * Guarded the way lib/appearance.ts guards matchMedia, and for the same reason:
 * a host that cannot answer says nothing rather than guessing. vitest runs with
 * `css: false`, so no stylesheet is loaded and this is '' unless a test sets
 * the property itself - which is what the preset case does.
 */
function tokenValue(view: Window, property: string): string {
  if (typeof view.getComputedStyle !== 'function') {
    return '';
  }
  return view.getComputedStyle(view.document.documentElement).getPropertyValue(property).trim();
}

export function AccentPicker() {
  const { t } = useTranslation();
  const store = useAppStore();
  const settings = useAppState((state) => state.settings);
  const view = useAppState((state) => state.view);

  const accent = settings?.accent ?? '';
  const [draft, setDraft] = useState(accent);
  const [shown, setShown] = useState(accent);
  const nameId = useId();

  // The box follows the SERVICE, always, and by two routes because there are
  // two ways it can fall behind.
  //
  //   here             the accent changed somewhere else — a late settings
  //                    read, or another control writing it.
  //   `choose` below   this picker asked for a write, and the answer came back.
  //
  // The second route cannot be folded into the first. A REFUSAL leaves the
  // stored accent exactly as it was, so `accent` does not change, nothing here
  // fires, and the rejected typing would sit in the box looking accepted.
  // Reverting from the ANSWER — what setAccent resolves with once
  // store/settings.ts has re-read — is the only version of this that is right
  // in both cases.
  //
  // Adjusted DURING RENDER rather than in an effect. React documents this as
  // the way to react to a value from outside changing; an effect would paint
  // the stale box first and then correct it, and ESLint's
  // react-hooks/set-state-in-effect refuses it outright.
  if (shown !== accent) {
    setShown(accent);
    setDraft(accent);
  }

  // Read during render rather than cached: the value depends on the palette and
  // the theme, both of which change under us, and a cached copy would be the
  // previous palette's accent one render later.
  const preset = tokenValue(view, PRESET_TOKEN);

  const choose = async (value: string) => {
    const answer = await store.getState().setAccent(value);
    // Null is "refused, and the re-read failed too" — nothing is known, so the
    // box keeps what the user typed rather than inventing a value.
    setDraft(answer?.accent ?? value);
  };

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    void choose(draft);
  };

  return (
    // No `role="group"`: the habits strip is one, and a second would make
    // `getByRole('group')` stop meaning "the strip". The visible label carries
    // the naming instead, through aria-describedby, exactly as the three
    // choices above it do.
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <span id={nameId} className="min-w-0 break-words text-muted">
        {t('settings.accent.group')}
      </span>

      <button
        type="button"
        data-accent="default"
        aria-describedby={nameId}
        aria-pressed={accent === ''}
        onClick={() => void choose('')}
        className={chipClass(accent === '')}
      >
        {t('settings.accent.default')}
      </button>

      {preset !== '' && (
        <button
          type="button"
          data-accent="preset"
          aria-describedby={nameId}
          aria-pressed={accent === preset}
          onClick={() => void choose(preset)}
          className={chipClass(accent === preset)}
        >
          {t('settings.accent.preset')}
        </button>
      )}

      <form onSubmit={submit} className="flex min-w-0 flex-wrap items-center gap-2">
        <input
          value={draft}
          onChange={(event) => setDraft(event.target.value)}
          aria-label={t('settings.accent.custom')}
          // The one fixed width in the header, and it is fixed because its
          // content is: a colour value is seven characters in every language,
          // so this field does not grow with Russian the way a label does.
          className="w-28 min-w-0 rounded-sm border border-line bg-surface px-2 py-1 font-mono text-ink"
        />
        <button type="submit" className={chipClass(false)}>
          {t('settings.accent.apply')}
        </button>
      </form>
    </div>
  );
}

/** One control's classes. Token names only; `accent` marks the one in force. */
function chipClass(active: boolean): string {
  return `min-w-0 break-words rounded-sm border px-2 py-1 transition-colors duration-fast ${
    active ? 'border-accent bg-accent text-on-accent' : 'border-line bg-surface text-muted'
  }`;
}

export default AccentPicker;
