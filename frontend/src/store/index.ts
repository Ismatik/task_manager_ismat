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
// `view` are injected for the same reason: behaviour that depends on the wall
// clock or the document is testable when the two are handed in and not when
// they are reached for. None of the three is reactive — nothing subscribes to
// them and nothing sets them.
//
// `now` is read by NO production code and that is deliberate (S2-18): the day a
// habit check lands on is Go's, the elapsed seconds on a timer are Go's, and
// the helper that used to advance the second of those between reads was deleted
// because nothing rendered it. The seam stays because store/habits.test.ts uses
// it to prove the point — the same toggle either side of local midnight sends
// the identical call — and because Stage 3's running clock will need it.

export interface Environment {
  /** The Go binding surface. */
  client: Client;
  /** Wall-clock milliseconds. Injected; read by tests, not by the store. */
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
