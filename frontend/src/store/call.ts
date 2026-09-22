// Nexus — the one place a Go rejection becomes a toast.
//
// PLAN.md section 1: no silent failures. ARCHITECTURE.md section 4: every bound
// method returns (T, error), and Wails turns the error into a rejected promise.
// Put those together and every single call into Go has the same two outcomes,
// so the handling belongs in one function rather than in a try/catch copied
// into twenty actions — nineteen of which would be right.
//
// # A refusal is not a failure (D25, S3-08)
//
// Some of what Go rejects is a malfunction; the rest is a RULE doing its job —
// a project may not enter doing (D9), a note has no due date. Until S3-08 both
// raised the single key `toast.error.body`, so a rule working exactly as
// specified read as "something went wrong".
//
// Go classifies, and only Go: internal/service/refusal.go holds one table of
// sentinel -> stable code, and tags the message with a fixed machine token on
// the way out. This module finds that token and turns the code into an i18n
// key. It reads NO prose: matching on Go's English message is forbidden,
// because a message is not an API, and because the message is untranslated
// anyway. An error carrying no token is a failure and keeps today's key — the
// classification is Go's whitelist, so an unknown error stays a failure, which
// is the safe direction.

import type { ToastDraft } from './toast';

/**
 * The i18n key a Go FAILURE raises — a malfunction, not a rule.
 *
 * One key, not one per call site: a Go error message is untranslated, usually
 * names a node id, and is not something a user can act on. The specific failure
 * goes to the console. A refusal, which IS something the user can act on, gets
 * its own key through the code below rather than by accident.
 */
export const GO_ERROR_KEY = 'toast.error.body';

/**
 * The dotted key each refusal code is labelled under.
 *
 * A convention and not a second table: `code -> 'toast.refusal.body.' + code`
 * is one rule with one spelling, and Go's
 * TestEveryRefusalCodeHasALabelInBothLanguages is what proves every code Go can
 * emit has a sentence here and in ru.json. A hand-written map would be a second
 * inventory of the codes, which is the defect this whole ticket is about.
 */
export const REFUSAL_KEY_PREFIX = 'toast.refusal.body.';

/**
 * The machine token service.Refuse writes, and the only thing read off a Go
 * error anywhere in the frontend.
 *
 * Spelled identically in internal/service/refusal.go; the Go test
 * TestTheWireMarkerIsSpelledTheSameInTypeScript reads THIS FILE and fails if
 * the two drift apart, because a TypeScript regex cannot import a Go constant.
 */
const REFUSAL_TOKEN = /\[nexus-refusal:([a-z0-9-]+)\]/;

/** What a toast says: the outcome half. The operation half is the caller's. */
export type ToastKind = 'failure' | 'refusal';

export interface ToastOutcome {
  kind: ToastKind;
  messageKey: string;
}

/**
 * THE parser. The only function in frontend/src that looks inside a Go error.
 *
 * Deliberately private: exporting it would invite a component to classify
 * something, and classification is Go's. `outcomeOf` below is the whole public
 * surface.
 */
function refusalCodeOf(cause: unknown): string | null {
  // Wails rejects with an Error whose message is Go's err.Error() (see the
  // generated runtime: `t.error instanceof Error ? t.error : new Error(t.error)`).
  // A string is accepted too, because a fake client in a test is allowed to be
  // simpler than the runtime, and because `throw 'x'` is legal JavaScript.
  const text = cause instanceof Error ? cause.message : typeof cause === 'string' ? cause : '';

  const found = REFUSAL_TOKEN.exec(text);
  return found === null ? null : found[1];
}

/**
 * Classifies one rejection: a refusal with its own sentence, or a failure.
 *
 * Used by `callGo` and by the two settings writes that cannot use it. Nothing
 * else decides what a rejection means.
 */
export function outcomeOf(cause: unknown): ToastOutcome {
  const code = refusalCodeOf(cause);

  if (code === null) {
    return { kind: 'failure', messageKey: GO_ERROR_KEY };
  }
  return { kind: 'refusal', messageKey: REFUSAL_KEY_PREFIX + code };
}

/** What `callGo` needs of the store: somewhere to put the toast. */
export interface ToastSink {
  pushToast(draft: ToastDraft): unknown;
}

/**
 * Calls Go, and turns a rejection into exactly one toast and a null.
 *
 * `operationKey` is the i18n key naming what the user was doing — "moving the
 * card", "ticking the habit". It is a requirement and not a nicety (D25): three
 * stacked toasts that all say "something went wrong" cannot be told apart, and
 * the user's screenshot that opened K12 is still unexplained for exactly that
 * reason.
 *
 * Returns null rather than rethrowing so that callers read as
 * `const board = await callGo(...); if (board === null) return;` — a refusal is
 * an ordinary outcome of a user action, not an exception to be handled three
 * frames up. The store is not touched on failure: nothing is cleared, nothing
 * is guessed, and the last thing Go actually said stays on screen.
 */
export async function callGo<T>(
  sink: ToastSink,
  operationKey: string,
  operation: () => Promise<T>,
): Promise<T | null> {
  try {
    return await operation();
  } catch (cause) {
    sink.pushToast({ operationKey, ...outcomeOf(cause), cause });
    return null;
  }
}
