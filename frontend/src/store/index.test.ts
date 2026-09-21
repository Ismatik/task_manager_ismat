import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createAppStore, GO_ERROR_KEY } from './index';
import {
  board,
  createFakeClient,
  habitView,
  nodeView,
  settingsView,
  timerView,
} from '../test/fakeClient';

// Five columns, named with values that are deliberately NOT the real ones.
//
// The real names are domain.Status's answer and are never spelt out in
// frontend/src — `make guard` check 2 refuses a quoted copy, including in a
// test. Using nonsense here is not a workaround, it is the stronger assertion:
// the store orders, indexes and focuses these columns correctly while having no
// idea what they mean, which is exactly the property being claimed.
const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

beforeEach(() => {
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

describe('hydration', () => {
  it('populates board, strip and settings from the client', async () => {
    const go = createFakeClient({
      board: board(COLUMNS, [nodeView()]),
      habits: [habitView()],
      settings: settingsView({ language: 'ru' }),
      timer: timerView({ running: true, entryId: 'entry-1', elapsedSeconds: 42 }),
    });
    const store = createAppStore(go.client, { view: testWindow(), now: () => 1_000 });

    await store.getState().hydrate();

    const state = store.getState();
    expect(state.board).toStrictEqual(go.state.board);
    expect(state.habits).toStrictEqual(go.state.habits);
    expect(state.settings).toStrictEqual(go.state.settings);
    expect(state.timer).toStrictEqual(go.state.timer);
    expect(state.toasts).toHaveLength(0);
  });

  it('holds the DTOs field for field, with nothing added and nothing dropped', async () => {
    const view = nodeView({ overdue: true, isLeaf: false });
    const go = createFakeClient({
      board: board(COLUMNS, [view]),
      habits: [habitView()],
    });
    const store = createAppStore(go.client, { view: testWindow() });

    await store.getState().hydrate();

    const columns = store.getState().board ?? [];
    expect(columns.map((column) => column.status)).toEqual(COLUMNS);

    // The DTO fields are pinned by DESTRUCTURING rather than by a list of
    // quoted names. Two reasons, and the second is the better one. The obvious
    // spelling would have to quote the name of ProgressView's completed-count
    // field, which is also one of the five status words; `make guard` check 2
    // cannot tell the two apart and should not have to. And a destructured
    // name is checked by gate 4: rename
    // a field in Go, regenerate models.ts, and this stops COMPILING rather than
    // failing at runtime. The `Object.keys(...).length` beside each one is what
    // catches a field that was added and forgotten here.
    const card = columns[0].nodes[0];
    const { node, status, progress, overdue, isLeaf, tags, timer, children } = card;
    expect({ node, status, progress, overdue, isLeaf, tags, timer, children }).toStrictEqual(card);
    expect(Object.keys(card)).toHaveLength(8);
    expect(Object.keys(columns[0])).toHaveLength(2);

    const { done, total, percent, defined } = progress;
    expect({ done, total, percent, defined }).toStrictEqual(progress);
    expect(Object.keys(progress)).toHaveLength(4);

    const habit = store.getState().habits![0];
    const { node: habitNode, scheduledToday, checkedToday, streak } = habit;
    expect({ node: habitNode, scheduledToday, checkedToday, streak }).toStrictEqual(habit);
    expect(Object.keys(habit)).toHaveLength(4);

    const preferences = store.getState().settings!;
    const { palette, theme, accent, language } = preferences;
    expect({ palette, theme, accent, language }).toStrictEqual(preferences);
    expect(Object.keys(preferences)).toHaveLength(4);

    // The card is the object Go produced, not a copy the store reshaped.
    expect(columns[0].nodes[0]).toBe(view);
  });
});

describe('a rejected call', () => {
  it('raises exactly one toast and leaves the store unchanged', async () => {
    const go = createFakeClient({ board: board(COLUMNS, [nodeView()]) });
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().hydrate();

    const before = store.getState().board;
    go.reject('Board', new Error('sql: database is closed'));

    const result = await store.getState().loadBoard();

    expect(result).toBeNull();
    expect(store.getState().board).toBe(before);
    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe(GO_ERROR_KEY);
  });

  it.each([
    ['Board', (store: ReturnType<typeof createAppStore>) => store.getState().loadBoard()],
    ['HabitStrip', (store: ReturnType<typeof createAppStore>) => store.getState().loadHabits()],
    ['TimerCurrent', (store: ReturnType<typeof createAppStore>) => store.getState().loadTimer()],
  ])('is surfaced for %s — nothing fails silently', async (method, load) => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });

    go.reject(method as 'Board', new Error('refused'));
    await load(store);

    expect(store.getState().toasts).toHaveLength(1);
  });

  it('sends the Go error to the console and only the key to the user', async () => {
    const go = createFakeClient();
    const cause = new Error('node 7f3: a project can never be doing');
    go.reject('Board', cause);
    const store = createAppStore(go.client, { view: testWindow() });

    await store.getState().loadBoard();

    expect(console.error).toHaveBeenCalledWith('[nexus]', GO_ERROR_KEY, cause);
    expect(store.getState().toasts[0].messageKey).toBe(GO_ERROR_KEY);
  });
});

describe('setPriority', () => {
  it('sends the number through untouched and re-reads the board', async () => {
    const go = createFakeClient({ board: board(COLUMNS, [nodeView()]) });
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().hydrate();

    const reads = go.calls.Board;

    expect(await store.getState().setPriority('node-1', 1)).toBe(true);
    expect(go.calls.SetPriority).toBe(1);
    // The board is re-read rather than patched: the priority chip is drawn
    // from what Go returns, and a local edit would be a second writer.
    expect(go.calls.Board).toBe(reads + 1);
  });

  it('does not pre-validate the range — an impossible value still goes to Go', async () => {
    // domain.Priority.Valid owns 1..4. A guard here would be that rule written
    // a second time, so the call must be MADE and the refusal must be Go's.
    const go = createFakeClient({ board: board(COLUMNS, [nodeView()]) });
    go.reject('SetPriority', new Error('node "node-1": priority: 9 is outside 1..4'));
    const store = createAppStore(go.client, { view: testWindow() });
    await store.getState().hydrate();

    const before = store.getState().board;

    expect(await store.getState().setPriority('node-1', 9)).toBe(false);
    expect(go.calls.SetPriority, 'the value was filtered out locally').toBe(1);
    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe(GO_ERROR_KEY);
    // No optimism here at all, so there is nothing to roll back and the board
    // is the very same object it was before the call.
    expect(store.getState().board).toBe(before);
  });
});

describe('the timer', () => {
  it('holds exactly what Go returned, and derives nothing from it', async () => {
    let clock = 10_000;
    const go = createFakeClient({
      timer: timerView({
        running: true,
        entryId: 'entry-1',
        elapsedSeconds: 42,
        startedAt: '2026-09-21T14:00:00Z',
      }),
    });
    const store = createAppStore(go.client, { view: testWindow(), now: () => clock });

    await store.getState().loadTimer();

    expect(store.getState().timer).toEqual(go.state.timer);

    // Time passes and the store's number does not move. Advancing it is a
    // DISPLAY concern that belongs to the component drawing a running clock
    // (Stage 3), and until that component exists the wall-clock arithmetic
    // that used to live here has no consumer — which is why S2-18 deleted it
    // rather than leaving it to be wired up to the wrong number later.
    clock += 120_000;
    expect(store.getState().timer?.elapsedSeconds).toBe(42);
  });

  it('is null before anything has been read', () => {
    const store = createAppStore(createFakeClient().client, { view: testWindow() });

    expect(store.getState().timer).toBeNull();
  });
});

describe('the UI slice', () => {
  it('holds selection, focus, the open overlay and the dragged card', () => {
    const store = createAppStore(createFakeClient().client, { view: testWindow() });

    expect(store.getState().selectedNodeId).toBeNull();
    expect(store.getState().focusedColumn).toBeNull();
    expect(store.getState().openOverlay).toBeNull();
    expect(store.getState().draggingNodeId).toBeNull();

    store.getState().select('node-1');
    // The column is named by the status string the BOARD returned — the store
    // does not know the five names and does not decide them.
    store.getState().focusColumn(COLUMNS[3]);
    store.getState().openOverlayPanel('quickAdd');
    store.getState().setDragging('node-1');

    expect(store.getState().selectedNodeId).toBe('node-1');
    expect(store.getState().focusedColumn).toBe(COLUMNS[3]);
    expect(store.getState().openOverlay).toBe('quickAdd');
    expect(store.getState().draggingNodeId).toBe('node-1');

    store.getState().closeOverlay();
    expect(store.getState().openOverlay).toBeNull();
  });

  it('is client-only: none of it reaches Go', () => {
    const go = createFakeClient();
    const store = createAppStore(go.client, { view: testWindow() });

    store.getState().select('node-1');
    store.getState().focusColumn(COLUMNS[0]);
    store.getState().openOverlayPanel('commandPalette');
    store.getState().setDragging('node-2');

    expect(go.calls).toStrictEqual({});
  });
});
