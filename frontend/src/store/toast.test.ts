import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createAppStore } from './index';
import { TOAST_CAP } from './toast';
import { createFakeClient } from '../test/fakeClient';

function newStore() {
  return createAppStore(createFakeClient().client);
}

// Four DISTINCT keys, so nothing here collapses by accident: after D24 two
// pushes of one key are deliberately one toast, and a test that reached for the
// same key twice would be testing de-duplication while claiming to test a cap.
const KEYS = ['toast.error.body', 'board.empty', 'habits.label', 'palette.empty'];

/** One operation, so that a test varying the MESSAGE varies only that. */
const OP = 'toast.operation.move';

/** Pushes through the store, spelling only what a given test cares about. */
function push(
  store: ReturnType<typeof newStore>,
  messageKey: string,
  extra: { operationKey?: string; cause?: unknown; kind?: 'failure' | 'refusal' } = {},
) {
  return store.getState().pushToast({ operationKey: OP, messageKey, ...extra });
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('the toast slice', () => {
  it('holds an i18n key, never a sentence', () => {
    const store = newStore();

    push(store, 'toast.error.body');

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe('toast.error.body');
  });

  it('sends the raw cause to the console and never to the user', () => {
    const store = newStore();
    const cause = new Error('sql: no rows in result set');

    const toast = push(store, 'toast.error.body', { cause });

    expect(console.error).toHaveBeenCalledWith('[nexus]', OP, 'toast.error.body', cause);
    // The cause travels with the toast for the console, but the only thing
    // that can be rendered is the key.
    expect(toast.messageKey).toBe('toast.error.body');
    expect(toast.cause).toBe(cause);
  });

  it('gives every toast a distinct id, so two failures are two toasts', () => {
    const store = newStore();

    const first = push(store, KEYS[0]);
    const second = push(store, KEYS[1]);

    expect(first.id).not.toBe(second.id);
    expect(store.getState().toasts).toHaveLength(2);
  });

  it('dismisses one without touching the others', () => {
    const store = newStore();
    const first = push(store, KEYS[0]);
    const second = push(store, KEYS[1]);

    store.getState().dismissToast(first.id);

    expect(store.getState().toasts.map((toast) => toast.id)).toEqual([second.id]);
  });

  it('clears the whole queue', () => {
    const store = newStore();
    push(store, KEYS[0]);
    push(store, KEYS[1]);

    store.getState().clearToasts();

    expect(store.getState().toasts).toHaveLength(0);
  });

  it('is independent per store, so one test cannot see another test’s toasts', () => {
    const a = newStore();
    const b = newStore();

    push(a, 'toast.error.body');

    expect(b.getState().toasts).toHaveLength(0);
  });

  // --- D24: the list is capped and de-duplicated (S3-07) ---------------------

  it('holds at most three toasts — a fourth distinct push drops the oldest', () => {
    const store = newStore();

    const raised = KEYS.map((key) => push(store, key));

    expect(TOAST_CAP).toBe(3);
    const left = store.getState().toasts;
    expect(left).toHaveLength(TOAST_CAP);
    // The OLDEST went, not the newest: what is on screen is the most recent
    // three, in the order they happened.
    expect(left.map((toast) => toast.id)).toEqual([raised[1].id, raised[2].id, raised[3].id]);
    expect(left.map((toast) => toast.messageKey)).toEqual([KEYS[1], KEYS[2], KEYS[3]]);
  });

  it('collapses an identical consecutive key into one toast carrying a count', () => {
    const store = newStore();

    push(store, KEYS[0]);
    push(store, KEYS[0]);
    const third = push(store, KEYS[0]);

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].count).toBe(3);
    // The id survives the collapse, so React keeps the element and the dismiss
    // button does not move out from under the pointer.
    expect(third.id).toBe(store.getState().toasts[0].id);
  });

  it('starts a fresh toast at a count of one', () => {
    const store = newStore();

    expect(push(store, KEYS[0]).count).toBe(1);
  });

  // --- D25: the toast names the operation, and a refusal is not a failure ----

  it('keeps the same outcome under two different operations apart', () => {
    const store = newStore();

    push(store, KEYS[0], { operationKey: 'toast.operation.move' });
    push(store, KEYS[0], { operationKey: 'toast.operation.create' });

    // Collapsing these would say "that happened twice" about two DIFFERENT
    // things, which is exactly the illegibility D25 is closing.
    expect(store.getState().toasts).toHaveLength(2);
    expect(store.getState().toasts.map((toast) => toast.operationKey)).toEqual([
      'toast.operation.move',
      'toast.operation.create',
    ]);
  });

  it('collapses only when BOTH the operation and the outcome repeat', () => {
    const store = newStore();

    push(store, KEYS[0], { operationKey: 'toast.operation.move' });
    push(store, KEYS[0], { operationKey: 'toast.operation.move' });

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].count).toBe(2);
  });

  it('is a failure unless the caller says otherwise', () => {
    const store = newStore();

    // An error in no Go table is a failure and keeps today's key. The default
    // is the safe direction: an unclassified rejection never poses as a rule.
    expect(push(store, KEYS[0]).kind).toBe('failure');
    expect(push(store, KEYS[1], { kind: 'refusal' }).kind).toBe('refusal');
  });

  it('carries the operation key into the console line', () => {
    const store = newStore();
    const cause = new Error('boom');

    push(store, KEYS[0], { operationKey: 'toast.operation.startTimer', cause });

    expect(console.error).toHaveBeenCalledWith(
      '[nexus]',
      'toast.operation.startTimer',
      KEYS[0],
      cause,
    );
  });

  it('collapses only a CONSECUTIVE repeat — a different key in between is a new toast', () => {
    const store = newStore();

    push(store, KEYS[0]);
    push(store, KEYS[1]);
    push(store, KEYS[0]);

    // Three failures, the first and third alike but not adjacent. Merging them
    // would reorder the list and claim a repeat that did not happen.
    expect(store.getState().toasts.map((toast) => toast.messageKey)).toEqual([
      KEYS[0],
      KEYS[1],
      KEYS[0],
    ]);
    expect(store.getState().toasts.map((toast) => toast.count)).toEqual([1, 1, 1]);
  });

  it('logs every repeat to the console, even the ones the screen collapses', () => {
    const store = newStore();
    const first = new Error('first');
    const second = new Error('second');

    push(store, KEYS[0], { cause: first });
    const repeated = push(store, KEYS[0], { cause: second });

    // De-duplication is a screen policy, not a logging policy: the console is
    // where the detail has always lived and it loses nothing.
    expect(console.error).toHaveBeenCalledTimes(2);
    expect(console.error).toHaveBeenNthCalledWith(1, '[nexus]', OP, KEYS[0], first);
    expect(console.error).toHaveBeenNthCalledWith(2, '[nexus]', OP, KEYS[0], second);
    // The toast carries the newest cause, which is the one worth looking at.
    expect(repeated.cause).toBe(second);
  });

  it('holds no user-visible sentence — every toast is a dotted key', () => {
    const store = newStore();

    KEYS.forEach((key) => push(store, key));

    for (const toast of store.getState().toasts) {
      for (const key of [toast.messageKey, toast.operationKey]) {
        expect(key).toMatch(/^[a-zA-Z0-9]+(\.[a-zA-Z0-9_-]+)+$/);
        expect(key).not.toContain(' ');
      }
    }
  });

  it('notifies subscribers on push and on dismiss', () => {
    const store = newStore();
    let calls = 0;
    const unsubscribe = store.subscribe(() => {
      calls += 1;
    });

    const toast = push(store, 'toast.error.body');
    store.getState().dismissToast(toast.id);

    expect(calls).toBe(2);

    unsubscribe();
    push(store, 'toast.error.body');
    expect(calls).toBe(2);
  });
});
