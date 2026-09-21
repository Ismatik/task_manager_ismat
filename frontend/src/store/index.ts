import { createStore, type StoreApi } from 'zustand/vanilla';

import type { Client } from '../lib/client';
import { createDataSlice, type DataSlice } from './data';
import { createSettingsSlice, type SettingsSlice } from './settings';
import { createToastSlice, type ToastSlice } from './toast';
import { createUiSlice, type UiSlice } from './ui';

// Nexus — the store.
//
// One store, four slices, and the slices exist to keep the files readable and
// not because there is more than one source of truth:
//
//   settings  the four persisted preferences, as Go returned them (D6)
//   data      the board, the habit strip and the timer, as Go returned them
//   ui        selection, focused column, open overlay, dragged card
//   toasts    every rejection that crossed the Go boundary
//
// # The three things it will not do
//
//   1. It computes no derived value. Status, column, progress, percentage,
//      overdue, streak and due date are all fields Go already filled in.
//   2. It never writes a domain field to a value Go did not produce, with the
//      single bounded exception of S2-17's optimistic move (see data.ts).
//   3. It swallows nothing. Every rejection becomes exactly one toast — see
//      call.ts, which is the only place that decides that.
//
// # Why the client and the clock are in the state
//
// `client` is injected, which is what makes every case in store tests a fake
// that resolves or rejects on demand rather than a WebKit window. `now` and
// `view` are injected for the same reason: a timer that ticks against a
// controllable clock is testable, and one that calls Date.now() directly is
// not. None of the three is reactive — nothing subscribes to them and nothing
// sets them.

export interface Environment {
  /** The Go binding surface. */
  client: Client;
  /** Wall-clock milliseconds. Injected so the timer display is testable. */
  now: () => number;
  /** The window whose document the appearance is applied to. */
  view: Window;
}

export interface AppState extends Environment, SettingsSlice, DataSlice, UiSlice, ToastSlice {}

export type AppStore = StoreApi<AppState>;

export interface AppStoreOptions {
  now?: () => number;
  view?: Window;
}

/**
 * Builds a store over a Go client.
 *
 * A factory rather than a module-level singleton: "a rejected call raises
 * exactly one toast" is not an assertion you can make against state another
 * test left behind.
 */
export function createAppStore(client: Client, options: AppStoreOptions = {}): AppStore {
  const { now = () => Date.now(), view = window } = options;

  return createStore<AppState>()((...args) => ({
    client,
    now,
    view,

    ...createSettingsSlice(...args),
    ...createDataSlice(...args),
    ...createUiSlice(...args),
    ...createToastSlice(...args),
  }));
}

export type { Toast } from './toast';
export type { Overlay } from './ui';
export { GO_ERROR_KEY } from './call';
