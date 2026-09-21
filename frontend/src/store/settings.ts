import {
  SetAccent,
  SetLanguage,
  SetPalette,
  SetTheme,
  Settings,
} from '../../wailsjs/go/main/App';
import type { service } from '../../wailsjs/go/models';

import { applyAppearance, revealAfterBoot } from '../lib/appearance';
import type { ToastStore } from './toast';

// Nexus — the settings store: the four persisted preferences (D6), held exactly
// as Go returned them.
//
// # The store does not own these values
//
// Every write goes to the service and the store then holds THE SERVICE'S
// ANSWER. That is why SettingsService returns the whole resulting SettingsView
// from every setter (S2-05) and why nothing here is set optimistically: the
// service validates the value, and a UI holding its own opinion about the
// palette would be a second opinion about what the palette is.
//
// A refusal therefore does not "roll back to the previous value" — it RE-READS,
// and applies whatever the service reports, because the service is the only
// thing that knows. One toast is raised, carrying an i18n key; the Go error
// goes to the console.
//
// # The wailsjs import
//
// This is currently the only file in frontend/src that imports the generated
// client, and it imports it because S2-12 needs Settings() before there is a
// lib/client.ts to route through. S2-13 creates that module and moves this
// import behind it — store/** is in S2-13's scope for exactly this reason.

/**
 * The service surface this store needs, as an interface so a test can supply a
 * fake that refuses, and so S2-13 can swap the generated bindings for the
 * client wrapper without touching anything below.
 */
export interface SettingsPort {
  read(): Promise<service.SettingsView>;
  setPalette(value: string): Promise<service.SettingsView>;
  setTheme(value: string): Promise<service.SettingsView>;
  setAccent(value: string): Promise<service.SettingsView>;
  setLanguage(value: string): Promise<service.SettingsView>;
}

/** The real port: thin delegations to the generated bindings. */
export const wailsSettingsPort: SettingsPort = {
  read: () => Settings(),
  setPalette: (value) => SetPalette(value),
  setTheme: (value) => SetTheme(value),
  setAccent: (value) => SetAccent(value),
  setLanguage: (value) => SetLanguage(value),
};

/** The key every appearance failure raises. S2-11 put it in both locale files. */
export const SETTINGS_ERROR_KEY = 'toast.error.body';

export interface SettingsStore {
  /** The last view the service returned, or null before the first read. */
  current(): service.SettingsView | null;
  /** Reads settings and applies the appearance. Never throws. */
  hydrate(): Promise<service.SettingsView | null>;
  setPalette(value: string): Promise<service.SettingsView | null>;
  setTheme(value: string): Promise<service.SettingsView | null>;
  setAccent(value: string): Promise<service.SettingsView | null>;
  setLanguage(value: string): Promise<service.SettingsView | null>;
  subscribe(listener: () => void): () => void;
}

export function createSettingsStore(
  port: SettingsPort,
  toasts: ToastStore,
  view: Window = window,
): SettingsStore {
  let settings: service.SettingsView | null = null;
  const listeners = new Set<() => void>();

  const emit = () => {
    for (const listener of listeners) {
      listener();
    }
  };

  /** Stores a view the service produced and puts the document into its state. */
  const adopt = (next: service.SettingsView): service.SettingsView => {
    settings = next;
    applyAppearance(next, view);
    emit();
    return next;
  };

  /**
   * One toast, then the service's own answer.
   *
   * Re-reading rather than reverting to the local copy is the whole point: the
   * write may have been refused, or partly applied, or the value may have been
   * normalised. Only the service knows, so it is asked.
   */
  const refused = async (cause: unknown): Promise<service.SettingsView | null> => {
    toasts.push(SETTINGS_ERROR_KEY, cause);

    try {
      return adopt(await port.read());
    } catch (readCause) {
      // The service cannot even be read. The document keeps whatever it has —
      // which is the last thing the service DID say — and nothing is invented.
      // No second toast: the user already has one, and two toasts for one
      // failed click is noise, not information.
      console.error('[nexus] settings re-read failed', readCause);
      revealAfterBoot(view.document);
      return settings;
    }
  };

  const write = async (
    operation: (value: string) => Promise<service.SettingsView>,
    value: string,
  ): Promise<service.SettingsView | null> => {
    try {
      return adopt(await operation(value));
    } catch (cause) {
      return refused(cause);
    }
  };

  return {
    current: () => settings,

    async hydrate() {
      try {
        return adopt(await port.read());
      } catch (cause) {
        toasts.push(SETTINGS_ERROR_KEY, cause);
        // The markup's static default stays in force, and the boot gate comes
        // off regardless: a failed read must not leave a blank window.
        revealAfterBoot(view.document);
        emit();
        return null;
      }
    },

    setPalette: (value) => write(port.setPalette, value),
    setTheme: (value) => write(port.setTheme, value),
    setAccent: (value) => write(port.setAccent, value),
    setLanguage: (value) => write(port.setLanguage, value),

    subscribe(listener) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
  };
}
