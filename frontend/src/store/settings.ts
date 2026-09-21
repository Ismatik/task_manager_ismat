import type { StateCreator } from 'zustand';

import { applyAppearance, revealAfterBoot } from '../lib/appearance';
import type { SettingsView } from '../lib/client';
import { GO_ERROR_KEY } from './call';
import type { AppState } from './index';

// Nexus — the settings slice: the four persisted preferences (D6), held exactly
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
// A refusal therefore does not "roll back to the previous value" — it RE-READS
// and applies whatever the service reports, because the service is the only
// thing that knows whether the write was refused, partly applied or normalised.
// One toast; the Go error goes to the console.

export interface SettingsSlice {
  /** The last view the service returned, or null before the first read. */
  settings: SettingsView | null;

  /** Reads settings and applies the appearance. Never throws. */
  loadSettings(): Promise<SettingsView | null>;

  setPalette(value: string): Promise<SettingsView | null>;
  setTheme(value: string): Promise<SettingsView | null>;
  setAccent(value: string): Promise<SettingsView | null>;
  setLanguage(value: string): Promise<SettingsView | null>;
}

export const createSettingsSlice: StateCreator<AppState, [], [], SettingsSlice> = (set, get) => {
  /** Stores a view the service produced and puts the document into its state. */
  const adopt = (next: SettingsView): SettingsView => {
    set({ settings: next });
    applyAppearance(next, get().view);
    return next;
  };

  /** One toast, then the service's own answer. */
  const refused = async (cause: unknown): Promise<SettingsView | null> => {
    get().pushToast(GO_ERROR_KEY, cause);

    try {
      return adopt(await get().client.Settings());
    } catch (readCause) {
      // The service cannot even be read. The document keeps whatever it has —
      // which is the last thing the service DID say — and nothing is invented.
      // No second toast: two toasts for one failed click is noise, not
      // information.
      console.error('[nexus] settings re-read failed', readCause);
      revealAfterBoot(get().view.document);
      return get().settings;
    }
  };

  const write = async (
    operation: (value: string) => Promise<SettingsView>,
    value: string,
  ): Promise<SettingsView | null> => {
    try {
      return adopt(await operation(value));
    } catch (cause) {
      return refused(cause);
    }
  };

  return {
    settings: null,

    async loadSettings() {
      try {
        return adopt(await get().client.Settings());
      } catch (cause) {
        get().pushToast(GO_ERROR_KEY, cause);
        // The markup's static default stays in force, and the boot gate comes
        // off regardless: a failed read must not leave a blank window.
        revealAfterBoot(get().view.document);
        return null;
      }
    },

    setPalette: (value) => write(get().client.SetPalette, value),
    setTheme: (value) => write(get().client.SetTheme, value),
    setAccent: (value) => write(get().client.SetAccent, value),
    setLanguage: (value) => write(get().client.SetLanguage, value),
  };
};
