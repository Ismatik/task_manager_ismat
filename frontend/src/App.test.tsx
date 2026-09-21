import { beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import { createAppStore } from './store';
import { board, columnView, createFakeClient, node, nodeView } from './test/fakeClient';
import { renderIn } from './test/render';

// Nexus — the shell, asserted THROUGH the shell.
//
// # The import restriction is the assertion
//
// This file imports `App` and does NOT import Kanban, Column, Card or ToastList.
// That is not tidiness: importing the component you are looking for turns the
// test into a test of that component and proves nothing about mounting. Every
// claim below is made about what a user would see after `render(<App ... />)`,
// which is the only thing that can tell "built" apart from "in the
// application".
//
// It replaces S2-10's smoke test, which asserted the shell was an empty `bg-bg`
// div — true then, and deliberately not true any more. Updating it is part of
// this ticket rather than a scope widening: TASKS.md lists this file in S2-15's
// Scope for exactly this reason.
//
// # No column name is written down
//
// The mocked board's statuses are nonsense strings. Which strings are real
// columns is domain.Status's answer (`make guard` check 2), and using opaque
// ones is the stronger assertion anyway: the shell renders whatever Go says, in
// whatever order Go says it, while knowing nothing about any of it.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

/** A board whose column order is deliberately not the order anybody expects. */
const UNUSUAL_ORDER = ['col-5', 'col-2', 'col-4', 'col-1', 'col-3'];

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

describe('the app shell', () => {
  it('shows the board it hydrated from Go', async () => {
    const fake = createFakeClient({ board: board(COLUMNS) });
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('en', <App store={store} />);

    // The headings are the only evidence a user has, and this file cannot see
    // the Column component to cheat with.
    for (const status of COLUMNS) {
      expect(await screen.findByRole('region', { name: status })).toBeInTheDocument();
    }
    expect(fake.calls.Board).toBe(1);
  });

  it('renders the columns in the order Go returned them', async () => {
    const fake = createFakeClient({ board: board(UNUSUAL_ORDER) });
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('en', <App store={store} />);
    await screen.findByRole('region', { name: UNUSUAL_ORDER[0] });

    const rendered = screen
      .getAllByRole('region')
      .map((region) => region.getAttribute('aria-label'));

    // Not `toContain`, not a set comparison: the ORDER is the claim.
    expect(rendered).toEqual(UNUSUAL_ORDER);
  });

  it('puts the cards Go supplied in the columns Go put them in', async () => {
    const fake = createFakeClient({
      board: [
        columnView(COLUMNS[0], [nodeView({ node: node({ id: 'n-1', title: 'Fix the tap' }) })]),
        columnView(COLUMNS[1]),
      ],
    });
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('en', <App store={store} />);
    const first = await screen.findByRole('region', { name: COLUMNS[0] });

    expect(first).toHaveTextContent('Fix the tap');
    expect(screen.getByRole('region', { name: COLUMNS[1] })).not.toHaveTextContent('Fix the tap');
  });

  it('gives an empty column a localised line rather than a blank rectangle', async () => {
    const fake = createFakeClient({ board: board(COLUMNS) });
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('ru', <App store={store} />);
    const column = await screen.findByRole('region', { name: COLUMNS[0] });

    expect(column).toHaveTextContent('В этой колонке пусто.');
    expect(column).toHaveTextContent('0 карточек');
  });

  // ------------------------------------------------------------------ toasts

  it('raises exactly one translated toast when a call is rejected', async () => {
    // S2-13 built ToastList and rendered it directly in its own test. This is
    // the assertion that was missing: that it is IN THE APPLICATION.
    const fake = createFakeClient({ board: board(COLUMNS) });
    fake.reject('Board', new Error('node 7f3: the database is gone'));
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('en', <App store={store} />);
    const alert = await screen.findByRole('alert');

    expect(alert).toHaveTextContent('Nexus could not finish that.');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    // The raw Go error is for the console, never for the user.
    expect(alert).not.toHaveTextContent('7f3');
  });

  it('lets the keyboard dismiss the toast', async () => {
    const fake = createFakeClient({ board: board(COLUMNS) });
    fake.reject('Board', new Error('nope'));
    const store = createAppStore(fake.client, { view: testWindow() });

    await renderIn('en', <App store={store} />);
    await screen.findByRole('alert');

    const user = userEvent.setup();
    await user.tab();
    expect(screen.getByRole('button')).toHaveFocus();
    await user.keyboard('{Enter}');

    await waitFor(() => expect(screen.queryByRole('alert')).toBeNull());
  });

  it('shows an empty board and an error rather than a blank window', async () => {
    const fake = createFakeClient({ board: board(COLUMNS) });
    fake.reject('Board', new Error('nope'));
    const store = createAppStore(fake.client, { view: testWindow() });

    const { container } = await renderIn('en', <App store={store} />);
    await screen.findByRole('alert');

    // The shell itself is still there: the page background, and the board
    // region a later ticket's Tab order walks through.
    expect(container.firstElementChild).toHaveClass('bg-bg');
    expect(screen.getByRole('main')).toBeInTheDocument();
  });

  // ------------------------------------------------------- the empty regions

  it('renders nothing at all for the regions it does not fill', async () => {
    // Regions 1 (header), 2 (habits strip) and 4 (overlay layer) are empty
    // slots in the source, and region 5 (toasts) renders nothing while nothing
    // has failed. Not a blank bar, not a placeholder, not a reserved box — ONE
    // child under the shell, the board, and nothing else.
    const fake = createFakeClient({ board: board(COLUMNS) });
    const store = createAppStore(fake.client, { view: testWindow() });

    const { container } = await renderIn('en', <App store={store} />);
    await screen.findByRole('region', { name: COLUMNS[0] });

    const shell = container.firstElementChild!;

    expect(shell.children).toHaveLength(1);
    expect(shell.children[0].tagName).toBe('MAIN');
    expect(screen.queryByRole('banner')).toBeNull();
    expect(screen.queryByRole('dialog')).toBeNull();
    expect(screen.queryByRole('alert')).toBeNull();
  });

  // -------------------------------------------------------------- the store

  it('gives two renders in one file independent stores', async () => {
    // `createAppStore` is a factory and must stay one. If it were a module
    // singleton the toast raised by the first store would still be on screen
    // for the second, and every "exactly one toast" assertion in the suite
    // would become order-dependent.
    const first = createFakeClient({ board: board(COLUMNS) });
    first.reject('Board', new Error('nope'));
    const firstStore = createAppStore(first.client, { view: testWindow() });

    const second = createFakeClient({ board: board(COLUMNS) });
    const secondStore = createAppStore(second.client, { view: testWindow() });

    expect(firstStore).not.toBe(secondStore);

    const { unmount } = await renderIn('en', <App store={firstStore} />);
    await screen.findByRole('alert');
    expect(firstStore.getState().toasts).toHaveLength(1);
    unmount();

    await renderIn('en', <App store={secondStore} />);
    await screen.findByRole('region', { name: COLUMNS[0] });

    expect(secondStore.getState().toasts).toEqual([]);
    expect(screen.queryByRole('alert')).toBeNull();
  });
});
