import type { StateCreator } from 'zustand';

import type { AppState } from './index';

// Nexus — the toast slice.
//
// # A toast carries a KEY, not a sentence
//
// The store never holds user-visible text. It holds an i18n key, translated at
// render time in whatever language is current; a string captured here would be
// frozen in the language that happened to be active when the error occurred.
// The raw Go error rides along for the console and is never rendered.

let nextId = 1;

export interface Toast {
  /** Stable per toast, so a list can key on it and a dismiss can name one. */
  id: number;
  /** An i18n key. Never a sentence. */
  messageKey: string;
  /** The raw failure, for the console. Never rendered. */
  cause?: unknown;
}

export interface ToastSlice {
  toasts: readonly Toast[];
  pushToast(messageKey: string, cause?: unknown): Toast;
  dismissToast(id: number): void;
  clearToasts(): void;
}

export const createToastSlice: StateCreator<AppState, [], [], ToastSlice> = (set, get) => ({
  toasts: [],

  pushToast(messageKey, cause) {
    const toast: Toast = { id: nextId, messageKey, cause };
    nextId += 1;

    set({ toasts: [...get().toasts, toast] });

    // The detail goes to the console and stops there. It is also never
    // swallowed: something that failed always leaves a trace somebody can read.
    if (cause !== undefined) {
      console.error('[nexus]', messageKey, cause);
    }

    return toast;
  },

  dismissToast(id) {
    set({ toasts: get().toasts.filter((toast) => toast.id !== id) });
  },

  clearToasts() {
    set({ toasts: [] });
  },
});
