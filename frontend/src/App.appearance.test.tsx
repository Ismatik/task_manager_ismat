import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, SettingsView } from './lib/client';
import { createAppStore, type AppStore } from './store';
import { board, createFakeClient, node, nodeView, settingsView } from './test/fakeClient';
import { BOTH_LANGUAGES, renderIn, tabUntil } from './test/render';

// Nexus — the appearance and language controls, asserted THROUGH the shell.
//
// This file imports `App` and does NOT import AppearanceControls or
// AccentPicker. The import restriction is the assertion: S2-21 fills region 1,
// the last empty one, and a test that imported the component would prove the
// component works while the running app still had a blank header.
//
// Everything is driven by `user-event` key presses.
//
// # No hex literal anywhere in here either
//
// `make guard` check 1 covers the whole of frontend/src, tests included. So the
// accent values below are CSS colour keywords and nonsense strings, never
// `#rrggbb` — which costs nothing, because nothing in TypeScript has an opinion
// about accent syntax. `domain.IsHexColour` is the rule, it lives in Go, and
// the case that proves the frontend does not second-guess it is the one where
// the fake refuses a value and the UI reverts.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

/** A colour the frontend passes through untouched. Go decides if it is valid. */
const A_COLOUR = 'rebeccapurple';
const ANOTHER_COLOUR = 'papayawhip';

/**
 * A window whose getComputedStyle is the real one.
 *
 * The other test files stub a minimal view object, because nothing they touch
 * reads computed style. The accent picker does — that is where its preset comes
 * from — so this one hands over jsdom's own implementation.
 */
function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
    getComputedStyle: (element: Element) => window.getComputedStyle(element),
  } as unknown as Window;
}

function storeOver(client: Client): AppStore {
  return createAppStore(client, { view: testWindow() });
}

/**
 * A fake Go that remembers what it was told, so a "restart" can read it back.
 *
 * `settingsView()` is the seeded default (aurora / dark / "" / en), and every
 * setter stores and returns the whole resulting view, exactly as
 * SettingsService does.
 */
function settingsGo(initial: Partial<SettingsView> = {}) {
  const fake = createFakeClient({
    board: board(COLUMNS, [nodeView({ node: node({ id: 'a', title: 'a' }) })]),
    settings: settingsView(initial),
  });

  return fake;
}

/** Renders the shell the way main.tsx does: settings first, then the tree. */
async function openTheApp(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await store.getState().loadSettings();
  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('article');

  return user;
}

/** The control for one value of one setting, by the id it publishes. */
function control(setting: string): HTMLElement {
  return document.querySelector<HTMLElement>(`[data-setting="${setting}"]`)!;
}

/** The accent picker's button for `which` — 'default' or 'preset'. */
function accentButton(which: string): HTMLElement | null {
  return document.querySelector<HTMLElement>(`[data-accent="${which}"]`);
}

/** Tabs to an ordinary control and presses Enter on it. */
async function pressByKeyboard(user: ReturnType<typeof userEvent.setup>, element: HTMLElement) {
  await tabUntil(user, () => document.activeElement === element);
  await user.keyboard('{Enter}');
}

/**
 * Chooses one value of one setting, with the keyboard and nothing else.
 *
 * Tab reaches the GROUP — one stop, on whichever value is in force — and the
 * arrows reach the rest of it, because a roving tabindex is what keeps a header
 * of four settings from being nine tab stops. So this walks: Tab into the
 * group, arrow to the value, Enter. Right first and then left, because the
 * arrows clamp rather than wrap and the target may be on either side.
 */
async function chooseByKeyboard(user: ReturnType<typeof userEvent.setup>, setting: string) {
  const [group] = setting.split(':');
  const target = control(setting);

  await tabUntil(
    user,
    () => document.activeElement?.getAttribute('data-setting')?.startsWith(`${group}:`) === true,
  );

  for (const arrow of ['{ArrowRight}', '{ArrowLeft}']) {
    for (let step = 0; step < 8 && document.activeElement !== target; step += 1) {
      await user.keyboard(arrow);
    }
  }

  expect(document.activeElement, `${setting} was never focused`).toBe(target);
  await user.keyboard('{Enter}');
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
  // applyAppearance writes to the REAL document, so one test's palette would
  // otherwise be the next test's starting state.
  const root = document.documentElement;
  root.removeAttribute('style');
  root.removeAttribute('class');
  root.removeAttribute('lang');
  delete root.dataset.palette;
  delete root.dataset.drift;
});

describe('the appearance controls are in the application', () => {
  it('fills region 1: the shell now renders a header before the board', async () => {
    const fake = settingsGo();
    const { container } = await renderIn('en', <App store={storeOver(fake.client)} />);
    await screen.findByRole('region', { name: COLUMNS[0] });

    const shell = container.firstElementChild!;
    expect([...shell.children].map((child) => child.tagName)).toEqual(['HEADER', 'MAIN']);
    expect(screen.getByRole('banner')).toBeInTheDocument();
  });

  it('is reachable from the board with Tab, and does not open the palette', async () => {
    // The criterion in S2-21's own words: "Tab from the board reaches the
    // controls without opening the command palette". Tab is the browser's own,
    // over the shell's DOM order, so it gets there by walking — and Ctrl+K is
    // never pressed, so no overlay opens.
    const fake = settingsGo();
    const store = storeOver(fake.client);
    const user = await openTheApp(store);

    await tabUntil(user, () => document.activeElement?.hasAttribute('data-node-id') === true);
    expect(document.activeElement).toHaveAttribute('data-node-id', 'a');

    await tabUntil(user, () => document.activeElement?.hasAttribute('data-setting') === true);

    expect(document.activeElement).toHaveAttribute('data-setting');
    expect(store.getState().openOverlay).toBeNull();
    expect(screen.queryByRole('dialog')).toBeNull();
  });

  it('shows the values the service reported, not values of its own', async () => {
    const fake = settingsGo({ palette: 'studio', theme: 'light', language: 'en' });
    await openTheApp(storeOver(fake.client));

    expect(control('palette:studio')).toHaveAttribute('aria-pressed', 'true');
    expect(control('palette:aurora')).toHaveAttribute('aria-pressed', 'false');
    expect(control('theme:light')).toHaveAttribute('aria-pressed', 'true');
    expect(control('language:en')).toHaveAttribute('aria-pressed', 'true');
    expect(accentButton('default')).toHaveAttribute('aria-pressed', 'true');
  });

  it('is ONE tab stop per group, with the arrows moving inside it', async () => {
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    await tabUntil(user, () => document.activeElement === control('palette:aurora'));

    // The tabindex sits on the value in force; the other is reachable only
    // with the arrows.
    expect(control('palette:aurora')).toHaveAttribute('tabindex', '0');
    expect(control('palette:studio')).toHaveAttribute('tabindex', '-1');

    await user.keyboard('{ArrowRight}');
    expect(control('palette:studio')).toHaveFocus();

    // Clamped at the end, never wrapped — the board's rule and the strip's.
    await user.keyboard('{ArrowRight}');
    expect(control('palette:studio')).toHaveFocus();

    await user.keyboard('{ArrowLeft}{ArrowLeft}');
    expect(control('palette:aurora')).toHaveFocus();
  });
});

describe('changing a setting, by keyboard alone', () => {
  it('writes the palette through SetPalette and applies the answer', async () => {
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    await chooseByKeyboard(user, 'palette:studio');

    await waitFor(() => expect(fake.state.settings.palette).toBe('studio'));
    await waitFor(() => expect(document.documentElement.dataset.palette).toBe('studio'));
    expect(control('palette:studio')).toHaveAttribute('aria-pressed', 'true');
  });

  it('writes the theme through SetTheme and applies the answer', async () => {
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    expect(document.documentElement.classList.contains('dark')).toBe(true);

    await chooseByKeyboard(user, 'theme:light');

    await waitFor(() => expect(fake.state.settings.theme).toBe('light'));
    await waitFor(() => expect(document.documentElement.classList.contains('dark')).toBe(false));
  });

  it('writes the language through SetLanguage and switches the UI', async () => {
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    await chooseByKeyboard(user, 'language:ru');

    await waitFor(() => expect(fake.state.settings.language).toBe('ru'));
    // i18next followed the SERVICE'S answer, which is lib/i18n.ts's rule.
    await waitFor(() => expect(screen.getByRole('main')).toHaveAttribute('aria-label', 'Доска'));
    expect(screen.getByRole('banner')).toHaveAttribute('aria-label', 'Оформление и язык');
    // ...and so did the document itself (K11, D23). Before S3-06, <html lang>
    // stayed at the markup's "en" for the life of the process.
    expect(document.documentElement.lang).toBe('ru');
  });

  // K11 / D23, asserted at the shell. lib/appearance.test.ts asserts the write;
  // this asserts that the SETTINGS READ reaches it, which is the half that was
  // missing entirely — the function existed, nothing ever called it with a
  // language, and there was no writer of <html lang> in frontend/src at all.
  it.each([['en'], ['ru']])(
    'serves the document as %s, because that is what the settings read reported',
    async (language) => {
      const fake = settingsGo({ language });
      await openTheApp(storeOver(fake.client), language);

      expect(document.documentElement.lang).toBe(language);
    },
  );

  it('survives a restart: the controls come back reading what Go stored', async () => {
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    await chooseByKeyboard(user, 'palette:studio');
    await waitFor(() => expect(fake.state.settings.palette).toBe('studio'));

    // A restart is a new store over the same Go. Nothing is carried across in
    // the frontend; the value comes back because the SERVICE has it.
    screen.getByRole('banner').ownerDocument.body.innerHTML = '';
    const second = storeOver(fake.client);
    await openTheApp(second);

    expect(control('palette:studio')).toHaveAttribute('aria-pressed', 'true');
    expect(second.getState().settings?.palette).toBe('studio');
  });
});

describe('the accent picker', () => {
  it('sends a free value to Go untouched, validating nothing', async () => {
    // domain.IsHexColour is the rule and it is asked once, in Go. A pre-check
    // here would be the rule written a second time.
    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    const field = screen.getByLabelText('A colour of your own');
    await tabUntil(user, () => document.activeElement === field);
    await user.keyboard(`${A_COLOUR}{Enter}`);

    await waitFor(() => expect(fake.state.settings.accent).toBe(A_COLOUR));
    await waitFor(() =>
      expect(document.documentElement.style.getPropertyValue('--accent')).toBe(A_COLOUR),
    );
  });

  it('writes "" for "the palette’s own" and REMOVES the inline --accent', async () => {
    // D6: "" means "use the palette's own --accent", and the only way the
    // palette gets its value back is for the declaration to be gone. An
    // `--accent: ;` left behind would shadow the palette with nothing.
    const fake = settingsGo({ accent: A_COLOUR });
    const user = await openTheApp(storeOver(fake.client));

    expect(document.documentElement.style.getPropertyValue('--accent')).toBe(A_COLOUR);

    await pressByKeyboard(user, accentButton('default')!);

    await waitFor(() => expect(fake.state.settings.accent).toBe(''));
    await waitFor(() =>
      expect(document.documentElement.style.getPropertyValue('--accent')).toBe(''),
    );
    expect(document.documentElement.getAttribute('style')).not.toContain('--accent');
    expect(accentButton('default')).toHaveAttribute('aria-pressed', 'true');
  });

  it('reverts to the service’s answer and toasts when Go refuses a value', async () => {
    const fake = settingsGo({ accent: A_COLOUR });
    fake.reject('SetAccent', new Error('service: setting accent: invalid value "nope"'));
    const user = await openTheApp(storeOver(fake.client));

    const field = screen.getByLabelText('A colour of your own');
    expect(field).toHaveValue(A_COLOUR);

    await tabUntil(user, () => document.activeElement === field);
    await user.clear(field);
    await user.keyboard('nope{Enter}');

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');
    // The Go sentence is for the console, never the screen.
    expect(alert).not.toHaveTextContent('invalid value');

    // Reverted FROM GO'S ANSWER — the box and the document both show what the
    // service still says, not the rejected typing.
    await waitFor(() => expect(field).toHaveValue(A_COLOUR));
    expect(document.documentElement.style.getPropertyValue('--accent')).toBe(A_COLOUR);
  });

  it('offers no preset when the tokens cannot be read, rather than inventing one', async () => {
    // vitest runs with `css: false`, so design/tokens.css is not loaded and
    // --accent-2 resolves to nothing. The picker then degrades to exactly the
    // fallback TASKS.md names: the free value plus the palette default.
    const fake = settingsGo();
    await openTheApp(storeOver(fake.client));

    expect(accentButton('preset')).toBeNull();
    expect(accentButton('default')).toBeInTheDocument();
    expect(screen.getByLabelText('A colour of your own')).toBeInTheDocument();
  });

  it('offers the token’s value as a preset when the tokens CAN be read', async () => {
    // The stub stands in for design/tokens.css, which is what supplies this
    // value in the running app. What is asserted is that the picker sends
    // whatever the TOKEN says — not a value written down in frontend/src.
    document.documentElement.style.setProperty('--accent-2', ANOTHER_COLOUR);

    const fake = settingsGo();
    const user = await openTheApp(storeOver(fake.client));

    const preset = accentButton('preset');
    expect(preset).not.toBeNull();

    await pressByKeyboard(user, preset!);

    await waitFor(() => expect(fake.state.settings.accent).toBe(ANOTHER_COLOUR));
  });
});

describe('Russian', () => {
  it.each(BOTH_LANGUAGES)('renders in %s with nothing hard-coded', async (language) => {
    const fake = settingsGo();
    await openTheApp(storeOver(fake.client), language);

    const header = screen.getByRole('banner');

    expect(header.textContent).not.toContain('settings.');
    expect(header.getAttribute('aria-label')).not.toContain('settings.');
    expect(screen.getByLabelText(language === 'ru' ? 'Свой цвет' : 'A colour of your own'))
      .toBeInTheDocument();
  });

  it('wraps at every level rather than clipping the wider language', async () => {
    // A 30%-wider language across four groups on one strip: the header wraps
    // its groups, each group wraps its controls, and every label breaks on
    // word boundaries. There is no fixed width and nothing truncates.
    const fake = settingsGo();
    await openTheApp(storeOver(fake.client), 'ru');

    const header = screen.getByRole('banner');
    expect(header).toHaveClass('flex-wrap');
    expect(header).toHaveClass('min-w-0');

    for (const button of [control('palette:aurora'), control('theme:dark'), control('language:ru')]) {
      expect(button).toHaveClass('break-words');
      expect(button).toHaveClass('min-w-0');
      expect(button.className).not.toContain('truncate');
      expect(button.className).not.toContain('whitespace-nowrap');
    }

    for (const group of header.querySelectorAll(':scope > div')) {
      expect(group).toHaveClass('flex-wrap');
    }
  });
});
