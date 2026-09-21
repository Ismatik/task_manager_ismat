import { createContext, useContext } from 'react';
import { useStore } from 'zustand';

import type { AppState, AppStore } from './index';

// Nexus — the one route the store takes into the React tree.
//
// # Why a context and not a module-level singleton
//
// `createAppStore` is a FACTORY, and S2-13 says why: "a rejected call raises
// exactly one toast" is not an assertion you can make against state another test
// left behind. A `export const store = createAppStore(wailsClient)` here would
// quietly take that property away from every test written afterwards — and it
// would also import the real Wails client into every module that wanted state,
// which breaks the one-importer rule lib/client.ts is built around.
//
// So: `main.tsx` builds exactly one store and hands it to `<App store={...} />`,
// App puts it in this context, and everything below reads it from here. One
// construction site, one route down, and a test builds its own store over a fake
// client and passes it to the same prop.
//
// # Why the context is exported and there is no Provider COMPONENT
//
// `<StoreContext.Provider>` is used directly, in App.tsx. A one-line
// `StoreProvider` wrapper would be a second name for the same thing, and it
// would make this a file that exports both a component and two hooks — which
// `react-refresh/only-export-components` warns about and `npm run lint
// --max-warnings=0` turns into a failed gate.

/**
 * The store, or null when nothing has provided one. Null rather than a default
 * store: a component rendered outside the provider is a wiring bug, and the
 * hook below says so instead of silently reading from a store nobody can see.
 */
export const StoreContext = createContext<AppStore | null>(null);

/** The store itself — for actions, which do not need a subscription. */
export function useAppStore(): AppStore {
  const store = useContext(StoreContext);

  if (store === null) {
    throw new Error('nexus: no store in context — render this inside <App store={...} />');
  }
  return store;
}

/**
 * A slice of state, re-rendering only when the selected value changes.
 *
 * Selector-based rather than "give me the whole state": the board, the habit
 * strip and the toast list change at different times, and a component that
 * subscribed to all of state would re-render on every one of them.
 */
export function useAppState<T>(selector: (state: AppState) => T): T {
  return useStore(useAppStore(), selector);
}
