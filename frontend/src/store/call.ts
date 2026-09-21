// Nexus — the one place a Go rejection becomes a toast.
//
// PLAN.md section 1: no silent failures. ARCHITECTURE.md section 4: every bound
// method returns (T, error), and Wails turns the error into a rejected promise.
// Put those together and every single call into Go has the same two outcomes,
// so the handling belongs in one function rather than in a try/catch copied
// into twenty actions — nineteen of which would be right.

/**
 * The i18n key every Go rejection raises.
 *
 * One key, not one per call site: a Go error message is untranslated, usually
 * names a node id, and is not something a user can act on. The specific failure
 * goes to the console. If a particular refusal ever deserves its own sentence
 * it gets its own key here, deliberately, rather than by accident.
 */
export const GO_ERROR_KEY = 'toast.error.body';

/** What `callGo` needs of the store: somewhere to put the toast. */
export interface ToastSink {
  pushToast(messageKey: string, cause?: unknown): unknown;
}

/**
 * Calls Go, and turns a rejection into exactly one toast and a null.
 *
 * Returns null rather than rethrowing so that callers read as
 * `const board = await callGo(...); if (board === null) return;` — a refusal is
 * an ordinary outcome of a user action, not an exception to be handled three
 * frames up. The store is not touched on failure: nothing is cleared, nothing
 * is guessed, and the last thing Go actually said stays on screen.
 */
export async function callGo<T>(sink: ToastSink, operation: () => Promise<T>): Promise<T | null> {
  try {
    return await operation();
  } catch (cause) {
    sink.pushToast(GO_ERROR_KEY, cause);
    return null;
  }
}
