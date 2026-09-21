import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createAppStore } from './index';
import { createFakeClient } from '../test/fakeClient';

function newStore() {
  return createAppStore(createFakeClient().client);
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

    const first = store.getState().pushToast('toast.error.body');
    const second = store.getState().pushToast('toast.error.body');

    expect(first.id).not.toBe(second.id);
    expect(store.getState().toasts).toHaveLength(2);
  });

  it('dismisses one without touching the others', () => {
    const store = newStore();
    const first = store.getState().pushToast('toast.error.body');
    const second = store.getState().pushToast('toast.error.body');

    store.getState().dismissToast(first.id);

    expect(store.getState().toasts.map((toast) => toast.id)).toEqual([second.id]);
  });

  it('clears the whole queue', () => {
    const store = newStore();
    store.getState().pushToast('toast.error.body');
    store.getState().pushToast('toast.error.body');

    store.getState().clearToasts();

    expect(store.getState().toasts).toHaveLength(0);
  });

  it('is independent per store, so one test cannot see another test’s toasts', () => {
    const a = newStore();
    const b = newStore();

    a.getState().pushToast('toast.error.body');

    expect(b.getState().toasts).toHaveLength(0);
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
