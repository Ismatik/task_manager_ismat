import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createToastStore } from './toast';

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('the toast queue', () => {
  it('holds an i18n key, never a sentence', () => {
    const toasts = createToastStore();

    toasts.push('toast.error.body');

    expect(toasts.list()).toHaveLength(1);
    expect(toasts.list()[0].messageKey).toBe('toast.error.body');
  });

  it('sends the raw cause to the console and never to the user', () => {
    const toasts = createToastStore();
    const cause = new Error('sql: no rows in result set');

    const toast = toasts.push('toast.error.body', cause);

    expect(console.error).toHaveBeenCalledWith('[nexus]', 'toast.error.body', cause);
    // The cause travels with the toast for the console, but the only thing that
    // can be rendered is the key.
    expect(toast.messageKey).toBe('toast.error.body');
    expect(toast.cause).toBe(cause);
  });

  it('gives every toast a distinct id, so two failures are two toasts', () => {
    const toasts = createToastStore();

    const first = toasts.push('toast.error.body');
    const second = toasts.push('toast.error.body');

    expect(first.id).not.toBe(second.id);
    expect(toasts.list()).toHaveLength(2);
  });

  it('dismisses one without touching the others', () => {
    const toasts = createToastStore();
    const first = toasts.push('toast.error.body');
    const second = toasts.push('toast.error.body');

    toasts.dismiss(first.id);

    expect(toasts.list().map((toast) => toast.id)).toEqual([second.id]);
  });

  it('is independent per store, so one test cannot see another test’s toasts', () => {
    const a = createToastStore();
    const b = createToastStore();

    a.push('toast.error.body');

    expect(b.list()).toHaveLength(0);
  });

  it('notifies subscribers on push, dismiss and clear', () => {
    const toasts = createToastStore();
    let calls = 0;
    const unsubscribe = toasts.subscribe(() => {
      calls += 1;
    });

    const toast = toasts.push('toast.error.body');
    toasts.dismiss(toast.id);
    toasts.clear();

    expect(calls).toBe(3);

    unsubscribe();
    toasts.push('toast.error.body');
    expect(calls).toBe(3);
  });
});
