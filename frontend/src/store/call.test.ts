import { describe, expect, it, vi } from 'vitest';

import { GO_ERROR_KEY, REFUSAL_KEY_PREFIX, callGo, outcomeOf } from './call';
import type { ToastDraft } from './toast';

// Nexus — the one boundary that reads a Go error (D25, S3-08).
//
// # The fixtures are the real wire format
//
// The strings below are what internal/service/refusal.go's Refuse() actually
// produces: the Go message, then the bracketed machine token. Go's own
// TestTheWireMarkerIsSpelledTheSameInTypeScript reads call.ts and fails if the
// marker here and the marker there stop being the same string, which is the
// half of this contract a TypeScript test cannot check by itself.
//
// Note what the prose in these fixtures is for: it is there to be IGNORED.
// Matching on "a project never enters doing" would be a second implementation
// of D9 living in TypeScript, and it would be untranslatable besides.

/** Exactly what Go sends when D9 declines to put a project into doing. */
const PROJECT_NEVER_DOING =
  'service: move node "p-1": domain: a project never enters doing [nexus-refusal:project-never-doing]';

function sink() {
  const drafts: ToastDraft[] = [];
  return { drafts, pushToast: (draft: ToastDraft) => drafts.push(draft) };
}

describe('classifying a Go rejection', () => {
  it('reads the code out of a refusal and turns it into a key', () => {
    expect(outcomeOf(new Error(PROJECT_NEVER_DOING))).toEqual({
      kind: 'refusal',
      messageKey: `${REFUSAL_KEY_PREFIX}project-never-doing`,
    });
  });

  it('calls anything with no code a failure, and keeps today’s key', () => {
    // S2's criterion is not weakened: every rejection still reaches a toast.
    expect(outcomeOf(new Error('store: database is locked'))).toEqual({
      kind: 'failure',
      messageKey: GO_ERROR_KEY,
    });
  });

  it('reads a rejection that is a bare string, not an Error', () => {
    expect(outcomeOf(PROJECT_NEVER_DOING).kind).toBe('refusal');
  });

  it('is not fooled by a rejection that is neither', () => {
    for (const cause of [undefined, null, 42, {}, [], new Date()]) {
      expect(outcomeOf(cause)).toEqual({ kind: 'failure', messageKey: GO_ERROR_KEY });
    }
  });

  it('refuses a token that is not shaped like a code', () => {
    // Half a token, or one with a space in it, is not a code. Better a generic
    // failure than a key built out of something Go did not say.
    expect(outcomeOf(new Error('boom [nexus-refusal:Project Never Doing]')).kind).toBe('failure');
    expect(outcomeOf(new Error('boom nexus-refusal:project-never-doing')).kind).toBe('failure');
  });

  it('reads the code and not the prose', () => {
    // The same code behind completely different English. If this passed by
    // matching the message, changing the message would break it — and Go's
    // messages are not an API.
    const other = new Error('totally different wording [nexus-refusal:project-never-doing]');

    expect(outcomeOf(other).messageKey).toBe(`${REFUSAL_KEY_PREFIX}project-never-doing`);
  });
});

describe('callGo', () => {
  it('returns what Go returned, and raises nothing', async () => {
    const got = sink();

    await expect(callGo(got, 'toast.operation.move', () => Promise.resolve(7))).resolves.toBe(7);
    expect(got.drafts).toEqual([]);
  });

  it('raises exactly one toast, naming the operation and the refusal', async () => {
    const got = sink();
    const cause = new Error(PROJECT_NEVER_DOING);

    const answer = await callGo(got, 'toast.operation.move', () => Promise.reject(cause));

    expect(answer).toBeNull();
    expect(got.drafts).toEqual([
      {
        operationKey: 'toast.operation.move',
        kind: 'refusal',
        messageKey: `${REFUSAL_KEY_PREFIX}project-never-doing`,
        cause,
      },
    ]);
  });

  it('raises exactly one toast for a failure too, with the operation on it', async () => {
    const got = sink();
    const cause = new Error('store: disk I/O error');

    await callGo(got, 'toast.operation.startTimer', () => Promise.reject(cause));

    expect(got.drafts).toEqual([
      {
        operationKey: 'toast.operation.startTimer',
        kind: 'failure',
        messageKey: GO_ERROR_KEY,
        cause,
      },
    ]);
  });

  it('never lets a rejection escape to the caller', async () => {
    const got = sink();
    const thrown = vi.fn();

    await callGo(got, 'toast.operation.create', () => Promise.reject(new Error('x'))).catch(thrown);

    expect(thrown).not.toHaveBeenCalled();
  });
});
