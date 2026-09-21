// Nexus — the toast queue.
//
// PLAN.md section 1: no silent failures. Every rejection that crosses the Go
// boundary ends up here, and the component that draws it arrives in S2-13.
//
// # What a toast carries is a KEY, not a sentence
//
// The store never holds user-visible text. It holds an i18n key, so the message
// is translated at render time in whatever language is current — a string
// captured here would be frozen in the language that happened to be active when
// the error occurred. The raw Go error travels alongside it for the console and
// is never shown to the user.

let nextId = 1;

export interface Toast {
  /** Stable per toast, so a list can key on it and a dismiss can name one. */
  id: number;
  /** An i18n key. Never a sentence. */
  messageKey: string;
  /** The raw failure, for the console. Never rendered. */
  cause?: unknown;
}

export interface ToastStore {
  list(): readonly Toast[];
  push(messageKey: string, cause?: unknown): Toast;
  dismiss(id: number): void;
  clear(): void;
  subscribe(listener: () => void): () => void;
}

/**
 * Creates an independent toast queue.
 *
 * A factory rather than a module-level singleton: tests need one queue per
 * case, and "exactly one toast was raised" is not an assertion you can make
 * against state another test left behind.
 */
export function createToastStore(): ToastStore {
  let toasts: readonly Toast[] = [];
  const listeners = new Set<() => void>();

  const emit = () => {
    for (const listener of listeners) {
      listener();
    }
  };

  return {
    list: () => toasts,

    push(messageKey, cause) {
      const toast: Toast = { id: nextId, messageKey, cause };
      nextId += 1;
      toasts = [...toasts, toast];

      // The detail goes to the console and stops there — a Go error string is
      // untranslated, often names a node id, and is not something a user can
      // act on. It is also never swallowed.
      if (cause !== undefined) {
        console.error('[nexus]', messageKey, cause);
      }

      emit();
      return toast;
    },

    dismiss(id) {
      toasts = toasts.filter((toast) => toast.id !== id);
      emit();
    },

    clear() {
      toasts = [];
      emit();
    },

    subscribe(listener) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
  };
}
