import type { ReactElement } from 'react';
import { render, type RenderResult } from '@testing-library/react';
import { I18nextProvider } from 'react-i18next';

import { createI18n } from '../lib/i18n';

// Nexus — render a tree in a real language, for tests.
//
// Every component from S2-14 onwards calls `t()`, so every component test needs
// an i18next instance in context. Copying six lines of I18nextProvider setup
// into each test file would be six places to edit when the runtime changes, and
// this project's prime directive is that nothing is spelt twice.
//
// `createI18n` builds an ISOLATED instance per call rather than reusing the
// default singleton, so a test that renders in Russian cannot leak the language
// into the next file. That is the property BOTH_LANGUAGES below relies on.

/**
 * The two languages that ship. A test that renders in both uses this rather
 * than writing the pair out, so adding a third locale is one edit.
 */
export const BOTH_LANGUAGES = ['en', 'ru'] as const;

/** Renders `ui` with i18next initialised in `language`. */
export async function renderIn(language: string, ui: ReactElement): Promise<RenderResult> {
  const i18n = await createI18n(language);

  return render(<I18nextProvider i18n={i18n}>{ui}</I18nextProvider>);
}
