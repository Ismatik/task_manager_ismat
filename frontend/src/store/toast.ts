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
//
// # The list is capped and de-duplicated (D24, S3-07)
//
// Until S3-07 `pushToast` appended unconditionally: no cap, no de-duplication,
// no auto-dismiss. Three identical failures were three identical panels filling
// the lower half of the window and covering the board, and they stayed there.
//
// So: at most TOAST_CAP toasts, and a push whose messageKey equals the newest
// toast's increments a count on that toast instead of appending. The count is
// deliberately NOT silent suppression — three failures are not one failure, and
// a repeat is the most useful thing about the second one, so the number is
// carried here and rendered (pluralised, in font-mono) by the component.
//
// Auto-dismiss is NOT here. "Has focus" and "the pointer is over it" are DOM
// facts, so the timer lives with the DOM, in components/Toast.tsx, and this
// module stays a plain reducer over a list. The interval it uses is
// TOAST_DISMISS_MS below, because the cap and the interval are one policy and
// belong in one file.

let nextId = 1;

/**
 * The most toasts that may be on screen at once (D24).
 *
 * A fourth distinct push drops the oldest. The dropped one is not lost: its
 * cause went to the console when it was raised, which is where the detail has
 * always lived.
 */
export const TOAST_CAP = 3;

/**
 * How long a toast stays up before dismissing itself, in milliseconds (D24).
 *
 * "No silent failure" is a rule about a failure being SURFACED, not about it
 * being permanent — a toast that never leaves converts one failure into a
 * permanently smaller window.
 */
export const TOAST_DISMISS_MS = 6000;

export interface Toast {
  /** Stable per toast, so a list can key on it and a dismiss can name one. */
  id: number;
  /** An i18n key. Never a sentence. */
  messageKey: string;
  /** The raw failure, for the console. Never rendered. */
  cause?: unknown;
  /**
   * How many times this same messageKey arrived in a row. 1 for a fresh toast.
   * A number, so the component renders it in font-mono and pluralises it.
   */
  count: number;
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
    const toasts = get().toasts;
    const newest = toasts[toasts.length - 1];

    // The detail goes to the console and stops there — once per push, including
    // the pushes that collapse into a count, so de-duplication on screen never
    // costs a line in the log. It is also never swallowed: something that
    // failed always leaves a trace somebody can read.
    if (cause !== undefined) {
      console.error('[nexus]', messageKey, cause);
    }

    if (newest !== undefined && newest.messageKey === messageKey) {
      // The same thing again: bump the count, keep the id (so React keeps the
      // element and the dismiss button the user may be pointing at), and carry
      // the newest cause. The new object is what tells the component to restart
      // its dismiss timer.
      const repeated: Toast = { ...newest, cause, count: newest.count + 1 };
      set({ toasts: [...toasts.slice(0, -1), repeated] });
      return repeated;
    }

    const toast: Toast = { id: nextId, messageKey, cause, count: 1 };
    nextId += 1;

    // slice(-TOAST_CAP) rather than shift-while-too-long: one expression, and
    // it is also correct if the cap is ever lowered below a longer list.
    set({ toasts: [...toasts, toast].slice(-TOAST_CAP) });

    return toast;
  },

  dismissToast(id) {
    set({ toasts: get().toasts.filter((toast) => toast.id !== id) });
  },

  clearToasts() {
    set({ toasts: [] });
  },
});
