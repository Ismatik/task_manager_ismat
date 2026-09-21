import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createAppStore, GO_ERROR_KEY } from './index';
import { createFakeClient, settingsView } from '../test/fakeClient';

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

beforeEach(() => {
  document.documentElement.dataset.appearance = 'pending';
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  const root = document.documentElement;
  root.removeAttribute('data-palette');
  root.removeAttribute('data-drift');
  root.removeAttribute('data-appearance');
  root.removeAttribute('style');
  root.classList.remove('dark');
  vi.restoreAllMocks();
});

describe('the settings slice', () => {
  it('hydrates from the service and applies the appearance', async () => {
    const go = createFakeClient({
      settings: settingsView({ palette: 'studio', theme: 'light', accent: 'teal' }),
    });
    const store = createAppStore(go.client, { view: testWindow() });

    const loaded = await store.getState().loadSettings();

    expect(loaded?.palette).toBe('studio');
    expect(store.getState().settings).toStrictEqual(go.state.settings);

    const root = document.documentElement;
    expect(root.dataset.palette).toBe('studio');
    expect(root.classList.contains('dark')).toBe(false);
    expect(root.style.getPropertyValue('--accent')).toBe('teal');
    expect(root.dataset.appearance).toBeUndefined();
  });

  it('writes through the service and adopts its answer', async () => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().loadSettings();

    await store.getState().setTheme('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);

    await store.getState().setPalette('studio');
    expect(document.documentElement.dataset.palette).toBe('studio');

    await store.getState().setAccent('crimson');
    expect(document.documentElement.style.getPropertyValue('--accent')).toBe('crimson');

    await store.getState().setAccent('');
    expect(document.documentElement.getAttribute('style') ?? '').not.toContain('--accent');

    await store.getState().setLanguage('ru');
    expect(store.getState().settings?.language).toBe('ru');

    expect(store.getState().settings).toStrictEqual(go.state.settings);
  });

  it('survives a restart: the next store restores palette, theme and accent', async () => {
    const go = createFakeClient();

    const first = createAppStore(go.client, { view: testWindow() });
    await first.getState().loadSettings();
    await first.getState().setPalette('studio');
    await first.getState().setTheme('light');
    await first.getState().setAccent('rebeccapurple');

    // The restart. A brand-new store over the SAME service, exactly as a second
    // launch of the binary reads the same `settings` rows.
    document.documentElement.dataset.appearance = 'pending';
    document.documentElement.removeAttribute('style');
    document.documentElement.removeAttribute('data-palette');

    const second = createAppStore(go.client, { view: testWindow() });
    await second.getState().loadSettings();

    const root = document.documentElement;
    expect(root.dataset.palette).toBe('studio');
    expect(root.classList.contains('dark')).toBe(false);
    expect(root.style.getPropertyValue('--accent')).toBe('rebeccapurple');
  });

  it('on a rejected SetTheme, leaves the DOM on the service’s value and raises ONE toast', async () => {
    const go = createFakeClient({ settings: settingsView({ theme: 'dark' }) });
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().loadSettings();

    go.reject('SetTheme', new Error('unknown theme: chartreuse'));

    await store.getState().setTheme('chartreuse');

    // The service still says dark, so the document says dark. This is NOT
    // "revert to what we had": the store re-read, and would have adopted a
    // different value had the service moved.
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(store.getState().settings?.theme).toBe('dark');

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe(GO_ERROR_KEY);
  });

  it('adopts a value the service normalised rather than the one it was sent', async () => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().loadSettings();

    // A service that refuses the write but has meanwhile moved on its own.
    go.reject('SetPalette', new Error('refused'));
    go.state.settings = settingsView({ palette: 'studio', theme: 'light', language: 'ru' });

    await store.getState().setPalette('aurora');

    // The truth is what the service reports NOW, not the value the store held
    // a moment ago. A plain rollback would have put aurora/dark back.
    expect(document.documentElement.dataset.palette).toBe('studio');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
    expect(store.getState().toasts).toHaveLength(1);
  });

  it('raises one toast and still reveals the document when the read itself fails', async () => {
    const go = createFakeClient();
    go.reject('Settings', new Error('database is locked'));
    const store = createAppStore(go.client, { view: testWindow() });

    const loaded = await store.getState().loadSettings();

    expect(loaded).toBeNull();
    expect(store.getState().settings).toBeNull();
    expect(store.getState().toasts).toHaveLength(1);
    // A Go error must not leave the user looking at a permanently blank window.
    expect(document.documentElement.dataset.appearance).toBeUndefined();
  });

  it('does not raise a second toast when the re-read also fails', async () => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().loadSettings();

    go.reject('SetTheme', new Error('refused'));
    go.reject('Settings', new Error('and the re-read failed too'));

    await store.getState().setTheme('light');

    expect(store.getState().toasts).toHaveLength(1);
  });

  it('notifies subscribers, and stops when unsubscribed', async () => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });

    let calls = 0;
    const unsubscribe = store.subscribe(() => {
      calls += 1;
    });

    await store.getState().loadSettings();
    expect(calls).toBeGreaterThan(0);

    const afterLoad = calls;
    await store.getState().setTheme('light');
    expect(calls).toBeGreaterThan(afterLoad);

    unsubscribe();
    const afterUnsubscribe = calls;
    await store.getState().setPalette('studio');
    expect(calls).toBe(afterUnsubscribe);
  });

  it('never invents a value Go did not produce', async () => {
    const go = createFakeClient({
      settings: settingsView({ palette: 'studio', theme: 'light', accent: 'teal', language: 'ru' }),
    });
    const store = createAppStore(go.client, { view: testWindow() });

    await store.getState().loadSettings();

    // Field for field, what the store holds is what the service returned.
    expect(store.getState().settings).toStrictEqual(go.state.settings);
  });
});
