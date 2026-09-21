import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import { ToastList } from './components/Toast';
import type { AppStore } from './store';
import { StoreContext, useAppState, useAppStore } from './store/context';
import { Kanban } from './views/Kanban';

// Nexus — the composition root.
//
// Read TASKS.md "Composition — who mounts what" before changing this file. The
// rule it states is the reason this shell exists at all:
//
//   > A component that is built and not mounted is an unfinished ticket.
//
// Until now the app mounted an empty <div>: the store was built in main.tsx and
// handed to nobody, and S2-13's ToastList was rendered by nothing. Every ticket
// passed, every gate was green, and the window opened on an empty page. So the
// shell is here, with the five regions named, and each later ticket mounts its
// own piece into the slot carrying its number. `App.mount.test.tsx` turns red on
// the commit that adds a component under components/ or views/ which main.tsx
// cannot reach.
//
// # The regions, in DOM order, and why the order IS the markup
//
//   1  header        S2-21 — appearance and language controls
//   2  habits strip  S2-18
//   3  board         THIS TICKET — views/Kanban.tsx
//   4  overlay layer S2-19 quick add, S2-20 command palette
//   5  toast layer   THIS TICKET — S2-13's ToastList, orphaned until now
//
// S2-16 specifies Tab/Shift+Tab as movement between regions. That order is this
// file's DOM order and nothing else: no tabIndex ladder, no ordering constant,
// no focus manager holding a second copy of the sequence. An empty region is an
// EMPTY SLOT IN THE SOURCE, named in a comment with its ticket — not a
// placeholder component, not localised filler and not a reserved blank box. A
// placeholder is a thing somebody has to remember to delete; an empty box is a
// bar of nothing across the screen. Regions 1, 2 and 4 emit no DOM at all, and
// App.test.tsx asserts it.
//
// # The store's one route down
//
// The store arrives as a PROP and is published on a context (store/context.ts).
// One construction site — main.tsx for the app, the test itself for a test —
// one route down, and `createAppStore` stays a factory so that two tests in one
// file get independent stores. A module-level singleton here would take that
// away, and would drag the real Wails client into every module that wanted
// state.

export interface AppProps {
  /** The one store, built by main.tsx over the real client, or by a test. */
  store: AppStore;
}

function App({ store }: AppProps) {
  return (
    <StoreContext.Provider value={store}>
      <Shell />
    </StoreContext.Provider>
  );
}

/**
 * The shell proper, rendered inside the provider so that it may use the hooks.
 *
 * Splitting it out is not decoration: a component cannot consume a context it
 * provides itself, and the alternative — putting the provider in main.tsx —
 * would mean `render(<App ... />)` in a test had no store at all, which is
 * precisely the mounting property every ticket from here on must assert.
 */
function Shell() {
  const { t } = useTranslation();
  const store = useAppStore();
  const toasts = useAppState((state) => state.toasts);
  const dismissToast = useAppState((state) => state.dismissToast);

  // The board read. It lives here rather than in main.tsx because main.tsx's
  // pre-paint work is the work that decides what the first frame LOOKS like —
  // palette, theme, accent, language — and the board is not that. Hydrating in
  // both places would be two hydration sites, which is one too many.
  //
  // The ref is a latch for React.StrictMode, which mounts every component twice
  // in development to surface exactly this kind of effect. Without it startup
  // would issue two Board() calls: harmless, and still noise nobody would ever
  // get round to explaining.
  const hydrated = useRef(false);

  useEffect(() => {
    if (hydrated.current) {
      return;
    }
    hydrated.current = true;

    // No catch here: the store turns a rejection into exactly one toast and
    // leaves the board null (S2-13's callGo). A failed read is an empty board
    // and an error message — never a blank window.
    void store.getState().loadBoard();
  }, [store]);

  return (
    <div className="flex min-h-screen min-w-0 flex-col gap-2 bg-bg p-2 text-ink">
      {/* Region 1 — header. S2-21 mounts the appearance and language controls. */}

      {/* Region 2 — habits strip. S2-18 mounts it. */}

      {/* Region 3 — board. */}
      <main aria-label={t('board.label')} className="min-w-0 flex-1">
        <Kanban />
      </main>

      {/* Region 4 — overlay layer. S2-19 quick add, S2-20 command palette. */}

      {/* Region 5 — toast layer. Renders nothing while there is nothing wrong. */}
      <ToastList toasts={toasts} onDismiss={dismissToast} />
    </div>
  );
}

export default App;
export { App };
