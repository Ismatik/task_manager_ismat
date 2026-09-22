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

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('the toast slice', () => {
  it('holds an i18n key, never a sentence', () => {
    const store = newStore();

    store.getState().pushToast('toast.error.body');

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe('toast.error.body');
  });

  it('sends the raw cause to the console and never to the user', () => {
    const store = newStore();
    const cause = new Error('sql: no rows in result set');

    const toast = store.getState().pushToast('toast.error.body', cause);

    expect(console.error).toHaveBeenCalledWith('[nexus]', 'toast.error.body', cause);
    // The cause travels with the toast for the console, but the only thing
    // that can be rendered is the key.
    expect(toast.messageKey).toBe('toast.error.body');
    expect(toast.cause).toBe(cause);
  });

  it('gives every toast a distinct id, so two failures are two toasts', () => {
    const store = newStore();

    const first = store.getState().pushToast(KEYS[0]);
    const second = store.getState().pushToast(KEYS[1]);

    expect(first.id).not.toBe(second.id);
    expect(store.getState().toasts).toHaveLength(2);
  });

  it('dismisses one without touching the others', () => {
    const store = newStore();
    const first = store.getState().pushToast(KEYS[0]);
    const second = store.getState().pushToast(KEYS[1]);

    store.getState().dismissToast(first.id);

    expect(store.getState().toasts.map((toast) => toast.id)).toEqual([second.id]);
  });

  it('clears the whole queue', () => {
    const store = newStore();
    store.getState().pushToast(KEYS[0]);
    store.getState().pushToast(KEYS[1]);

    store.getState().clearToasts();

    expect(store.getState().toasts).toHaveLength(0);
  });

  it('is independent per store, so one test cannot see another test’s toasts', () => {
    const a = newStore();
    const b = newStore();

    a.getState().pushToast('toast.error.body');

    expect(b.getState().toasts).toHaveLength(0);
  });

  // --- D24: the list is capped and de-duplicated (S3-07) ---------------------

  it('holds at most three toasts — a fourth distinct push drops the oldest', () => {
    const store = newStore();

    const raised = KEYS.map((key) => store.getState().pushToast(key));

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

    store.getState().pushToast(KEYS[0]);
    store.getState().pushToast(KEYS[0]);
    const third = store.getState().pushToast(KEYS[0]);

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].count).toBe(3);
    // The id survives the collapse, so React keeps the element and the dismiss
    // button does not move out from under the pointer.
    expect(third.id).toBe(store.getState().toasts[0].id);
  });

  it('starts a fresh toast at a count of one', () => {
    const store = newStore();

    expect(store.getState().pushToast(KEYS[0]).count).toBe(1);
  });

  it('collapses only a CONSECUTIVE repeat — a different key in between is a new toast', () => {
    const store = newStore();

    store.getState().pushToast(KEYS[0]);
    store.getState().pushToast(KEYS[1]);
    store.getState().pushToast(KEYS[0]);

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

    store.getState().pushToast(KEYS[0], first);
    const repeated = store.getState().pushToast(KEYS[0], second);

    // De-duplication is a screen policy, not a logging policy: the console is
    // where the detail has always lived and it loses nothing.
    expect(console.error).toHaveBeenCalledTimes(2);
    expect(console.error).toHaveBeenNthCalledWith(1, '[nexus]', KEYS[0], first);
    expect(console.error).toHaveBeenNthCalledWith(2, '[nexus]', KEYS[0], second);
    // The toast carries the newest cause, which is the one worth looking at.
    expect(repeated.cause).toBe(second);
  });

  it('holds no user-visible sentence — every toast is a dotted key', () => {
    const store = newStore();

    KEYS.forEach((key) => store.getState().pushToast(key));

    for (const toast of store.getState().toasts) {
      expect(toast.messageKey).toMatch(/^[a-zA-Z0-9]+(\.[a-zA-Z0-9_]+)+$/);
      expect(toast.messageKey).not.toContain(' ');
    }
  });

  it('notifies subscribers on push and on dismiss', () => {
    const store = newStore();
    let calls = 0;
    const unsubscribe = store.subscribe(() => {
      calls += 1;
    });

    const toast = store.getState().pushToast('toast.error.body');
    store.getState().dismissToast(toast.id);

    expect(calls).toBe(2);

    unsubscribe();
    store.getState().pushToast('toast.error.body');
    expect(calls).toBe(2);
  });
});
