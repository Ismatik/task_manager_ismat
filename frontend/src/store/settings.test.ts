import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createSettingsStore, SETTINGS_ERROR_KEY, type SettingsPort } from './settings';
import { createToastStore } from './toast';
import type { service } from '../../wailsjs/go/models';

// A window with working media queries, so the drift gate has something to read.
function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

function view(overrides: Partial<service.SettingsView> = {}): service.SettingsView {
  return {
    palette: 'aurora',
    theme: 'dark',
    accent: '',
    language: 'en',
    ...overrides,
  } as service.SettingsView;
}

/**
 * A fake SettingsService. It stores what it is given and reports the stored
 * value back, which is the contract every setter in app.go has: it returns the
 * whole resulting SettingsView.
 */
function fakeService(initial: service.SettingsView = view()) {
  let state = initial;
  const port: SettingsPort = {
    read: async () => state,
    setPalette: async (value) => {
      state = view({ ...state, palette: value });
      return state;
    },
    setTheme: async (value) => {
      state = view({ ...state, theme: value });
      return state;
    },
    setAccent: async (value) => {
      state = view({ ...state, accent: value });
      return state;
    },
    setLanguage: async (value) => {
      state = view({ ...state, language: value });
      return state;
    },
  };

  return {
    port,
    stored: () => state,
    set: (next: service.SettingsView) => {
      state = next;
    },
  };
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

describe('the settings store', () => {
  it('hydrates from the service and applies the appearance', async () => {
    const service = fakeService(view({ palette: 'studio', theme: 'light', accent: 'teal' }));
    const store = createSettingsStore(service.port, createToastStore(), testWindow());

    const hydrated = await store.hydrate();

    expect(hydrated?.palette).toBe('studio');
    expect(store.current()).toEqual(service.stored());

    const root = document.documentElement;
    expect(root.dataset.palette).toBe('studio');
    expect(root.classList.contains('dark')).toBe(false);
    expect(root.style.getPropertyValue('--accent')).toBe('teal');
    expect(root.dataset.appearance).toBeUndefined();
  });

  it('writes through the service and adopts its answer', async () => {
    const service = fakeService();
    const store = createSettingsStore(service.port, createToastStore(), testWindow());
    await store.hydrate();

    await store.setTheme('light');
    expect(document.documentElement.classList.contains('dark')).toBe(false);

    await store.setPalette('studio');
    expect(document.documentElement.dataset.palette).toBe('studio');

    await store.setAccent('crimson');
    expect(document.documentElement.style.getPropertyValue('--accent')).toBe('crimson');

    await store.setAccent('');
    expect(document.documentElement.getAttribute('style') ?? '').not.toContain('--accent');

    expect(service.stored()).toEqual(store.current());
  });

  it('survives a restart: the next hydrate restores palette, theme and accent', async () => {
    const service = fakeService();

    const first = createSettingsStore(service.port, createToastStore(), testWindow());
    await first.hydrate();
    await first.setPalette('studio');
    await first.setTheme('light');
    await first.setAccent('rebeccapurple');

    // The restart. A brand-new store over the SAME service, exactly as a second
    // launch of the binary reads the same `settings` rows.
    document.documentElement.dataset.appearance = 'pending';
    document.documentElement.removeAttribute('style');
    document.documentElement.removeAttribute('data-palette');

    const second = createSettingsStore(service.port, createToastStore(), testWindow());
    await second.hydrate();

    const root = document.documentElement;
    expect(root.dataset.palette).toBe('studio');
    expect(root.classList.contains('dark')).toBe(false);
    expect(root.style.getPropertyValue('--accent')).toBe('rebeccapurple');
  });

  it('on a rejected SetTheme, leaves the DOM on the service’s value and raises ONE toast', async () => {
    const service = fakeService(view({ theme: 'dark' }));
    const toasts = createToastStore();
    const store = createSettingsStore(service.port, toasts, testWindow());
    await store.hydrate();

    service.port.setTheme = async () => {
      throw new Error('unknown theme: chartreuse');
    };

    await store.setTheme('chartreuse');

    // The service still says dark, so the document says dark. Note this is NOT
    // "revert to what we had" — the store re-read, and would have adopted a
    // different value had the service moved.
    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(store.current()?.theme).toBe('dark');

    expect(toasts.list()).toHaveLength(1);
    expect(toasts.list()[0].messageKey).toBe(SETTINGS_ERROR_KEY);
  });

  it('adopts a value the service normalised rather than the one it was sent', async () => {
    const service = fakeService();
    const toasts = createToastStore();
    const store = createSettingsStore(service.port, toasts, testWindow());
    await store.hydrate();

    // A service that refuses the write but has meanwhile moved on its own.
    service.port.setPalette = async () => {
      throw new Error('refused');
    };
    service.set(view({ palette: 'studio', theme: 'light', accent: '', language: 'ru' }));

    await store.setPalette('aurora');

    // The truth is what the service reports NOW, not the value the store held
    // a moment ago. A plain rollback would have put aurora/dark back.
    expect(document.documentElement.dataset.palette).toBe('studio');
    expect(document.documentElement.classList.contains('dark')).toBe(false);
    expect(toasts.list()).toHaveLength(1);
  });

  it('raises one toast and still reveals the document when the read itself fails', async () => {
    const toasts = createToastStore();
    const store = createSettingsStore(
      {
        read: async () => {
          throw new Error('database is locked');
        },
      } as unknown as SettingsPort,
      toasts,
      testWindow(),
    );

    const hydrated = await store.hydrate();

    expect(hydrated).toBeNull();
    expect(toasts.list()).toHaveLength(1);
    // A Go error must not leave the user looking at a permanently blank window.
    expect(document.documentElement.dataset.appearance).toBeUndefined();
  });

  it('does not raise a second toast when the re-read also fails', async () => {
    const service = fakeService();
    const toasts = createToastStore();
    const store = createSettingsStore(service.port, toasts, testWindow());
    await store.hydrate();

    service.port.setTheme = async () => {
      throw new Error('refused');
    };
    service.port.read = async () => {
      throw new Error('and the re-read failed too');
    };

    await store.setTheme('light');

    expect(toasts.list()).toHaveLength(1);
  });

  it('notifies subscribers, and stops when unsubscribed', async () => {
    const service = fakeService();
    const store = createSettingsStore(service.port, createToastStore(), testWindow());

    let calls = 0;
    const unsubscribe = store.subscribe(() => {
      calls += 1;
    });

    await store.hydrate();
    expect(calls).toBe(1);

    await store.setTheme('light');
    expect(calls).toBe(2);

    unsubscribe();
    await store.setTheme('dark');
    expect(calls).toBe(2);
  });

  it('never invents a value Go did not produce', async () => {
    const service = fakeService(view({ palette: 'studio', theme: 'light', accent: 'teal', language: 'ru' }));
    const store = createSettingsStore(service.port, createToastStore(), testWindow());

    await store.hydrate();

    // Field for field, what the store holds is what the service returned.
    expect(store.current()).toStrictEqual(service.stored());
  });
});
