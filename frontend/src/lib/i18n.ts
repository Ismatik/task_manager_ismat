import { createInstance, type i18n as I18n } from 'i18next';
import { initReactI18next } from 'react-i18next';

import en from '../locales/en.json';
import ru from '../locales/ru.json';

// Nexus — the i18n runtime.
//
// # Everything is bundled. Nothing is fetched.
//
// The resources below are imported as modules, so Vite inlines them into the
// bundle. There is no i18next-http-backend, no i18next-fs-backend and no
// language detector: Nexus is local-only, and a detector is also simply wrong
// here — the language is a persisted user setting (D6), read out of Go, not a
// guess made from navigator.language.
//
// # The frontend does not own the language
//
// `changeLanguage` below takes a LanguagePort and switches to the language the
// PORT REPORTS BACK, not to the one it was asked for. That is the same rule the
// appearance runtime follows and it exists for the same reason: the service
// validates the value and its answer is the truth. A UI that assumed its own
// optimistic value would be a second opinion about what the language is.
//
// The port is injected rather than imported because of a layering rule that is
// checked mechanically: `frontend/src/lib/client.ts` is the ONLY file in
// frontend/src allowed to import from wailsjs. This module is therefore
// deliberately ignorant of Wails, which is also what makes it testable against
// an in-memory fake that behaves like SettingsService.

export const resources = {
  en: { translation: en },
  ru: { translation: ru },
} as const;

// The language used when nothing has been chosen yet, and the one i18next falls
// back to for a key that is somehow missing. It is not a policy decision about
// which language the app opens in — that comes from settings.
export const fallbackLanguage = 'en';

// LanguagePort persists a language and returns the language that is now in
// force. Go's SetLanguage returns the whole resulting SettingsView for exactly
// this reason; the adapter narrows it to the field this module cares about.
//
// A rejected write rejects the promise. Nothing here catches it: the caller
// raises the toast, because "every rejection is surfaced" is one rule and it
// lives in one place.
export type LanguagePort = (language: string) => Promise<string>;

/**
 * Builds an isolated i18next instance for a language.
 *
 * `createInstance` rather than the default singleton: a singleton leaks the
 * language from one test into the next, and a test that passes only when it
 * runs first is not a test.
 */
export async function createI18n(language: string): Promise<I18n> {
  const instance = createInstance();

  await instance.use(initReactI18next).init({
    resources,
    lng: language,
    fallbackLng: fallbackLanguage,
    // The keys are already namespaced by their object shape ("toast.error.title"),
    // so the ':' namespace separator would only get in the way, and '.' is the
    // nesting separator i18next uses by default.
    ns: ['translation'],
    defaultNS: 'translation',
    interpolation: {
      // React escapes everything it renders already; leaving i18next's own
      // escaping on turns an apostrophe in a Russian string into &#39;.
      escapeValue: false,
    },
    // A missing key is a bug, and a bug that renders as the key itself is a bug
    // somebody notices. Returning the key rather than an empty string is what
    // makes the locale-parity test's failure mode visible on screen too.
    returnEmptyString: false,
    react: {
      // Nothing is loaded asynchronously — the resources are in the bundle and
      // init is awaited before the first render — so suspending would only ever
      // be a way to blank the screen for a tick.
      useSuspense: false,
    },
  });

  return instance;
}

/**
 * Switches the language, persisting it through the port first.
 *
 * Returns the language now in force, which is the port's answer and not
 * necessarily the requested one. Rejects — without changing the UI — when the
 * port rejects.
 */
export async function changeLanguage(
  instance: I18n,
  language: string,
  persist: LanguagePort,
): Promise<string> {
  const accepted = await persist(language);
  await instance.changeLanguage(accepted);
  return accepted;
}
