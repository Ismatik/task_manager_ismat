import type {
  Client,
  ColumnView,
  HabitView,
  Node,
  NodeView,
  ProgressView,
  SettingsView,
  Tag,
  TimerView,
} from '../lib/client';

// Nexus — a fake Go client, for tests.
//
// It resolves with whatever it is told to, rejects on demand, and records every
// call. That is the whole point of lib/client.ts being an interface: "a
// rejected call raises exactly one toast" is asserted here, in milliseconds,
// instead of being a sentence somebody wrote after clicking around a window.
//
// The fixtures below are shaped exactly like internal/service/dto.go. They are
// not a model of it — a test that invented its own shape would pass while the
// real board crashed — so every field the DTO declares is present, including
// the ones whose value is the zero value.

/**
 * The stored status a fixture node carries by default.
 *
 * Deliberately NOT one of the five real column names. Nothing in frontend/src
 * knows what those are — that is domain.Status's answer, and `make guard`
 * check 2 refuses a quoted copy of it — and a fixture that spelt them out would
 * be that copy in the one place nobody looks. It being a nonsense value also
 * proves something: the store handles this node exactly as it handles a real
 * one, because it never interprets the string.
 */
export const PLACEHOLDER_STATUS = 'status-from-go';

export function tag(overrides: Partial<Tag> = {}): Tag {
  return { id: 'tag-1', name: 'home', color: 'success', ...overrides } as Tag;
}

export function node(overrides: Partial<Node> = {}): Node {
  return {
    id: 'node-1',
    parentId: undefined,
    type: 'task',
    title: 'Write the board',
    descriptionMd: '',
    status: PLACEHOLDER_STATUS,
    due: undefined,
    dueSource: 'none',
    priority: 4,
    estimateMin: undefined,
    recurrence: undefined,
    activity: undefined,
    sortOrder: 0,
    createdAt: '2026-09-21T09:00:00Z',
    updatedAt: '2026-09-21T09:00:00Z',
    completedAt: undefined,
    archivedAt: undefined,
    ...overrides,
  } as Node;
}

/**
 * A ProgressView with nothing to measure — the D7/D11 case in which neither 0%
 * nor 100% is true.
 *
 * A const rather than a factory on purpose: `make guard` check 3b refuses a
 * lower-case function whose name contains "progress", and it is right to. Spread
 * it to vary it: `{ ...undefinedProgress, defined: true, percent: 50 }`.
 */
export const undefinedProgress = {
  done: 0,
  total: 0,
  percent: 0,
  defined: false,
} as ProgressView;

export function timerView(overrides: Partial<TimerView> = {}): TimerView {
  return {
    running: false,
    entryId: '',
    startedAt: undefined,
    elapsedSeconds: 0,
    ...overrides,
  } as TimerView;
}

export function nodeView(overrides: Partial<NodeView> = {}): NodeView {
  return {
    node: node(),
    status: PLACEHOLDER_STATUS,
    progress: undefinedProgress,
    overdue: false,
    isLeaf: true,
    tags: [],
    timer: timerView(),
    children: [],
    ...overrides,
  } as NodeView;
}

export function columnView(status: string, nodes: NodeView[] = []): ColumnView {
  return { status, nodes } as ColumnView;
}

export function habitView(overrides: Partial<HabitView> = {}): HabitView {
  return {
    node: node({ id: 'habit-1', type: 'habit', recurrence: 'FREQ=DAILY' }),
    scheduledToday: true,
    checkedToday: false,
    streak: 0,
    ...overrides,
  } as HabitView;
}

export function settingsView(
  overrides: Partial<SettingsView> = {},
): SettingsView {
  return {
    palette: 'aurora',
    theme: 'dark',
    accent: '',
    language: 'en',
    ...overrides,
  } as SettingsView;
}

/**
 * A board of columns, named by whatever statuses the caller says Go returned.
 *
 * The names are a parameter and never a literal, for the reason on
 * PLACEHOLDER_STATUS above.
 */
export function board(statuses: string[], nodes: NodeView[] = []): ColumnView[] {
  return statuses.map((status, index) =>
    columnView(status, index === 0 ? nodes : []),
  );
}

export interface FakeClientState {
  board: ColumnView[];
  habits: HabitView[];
  settings: SettingsView;
  timer: TimerView;
}

export interface FakeClient {
  client: Client;
  state: FakeClientState;
  /** Method name -> how many times it was called. */
  calls: Record<string, number>;
  /** Makes the named method reject with `error` until it is cleared. */
  reject(method: keyof Client, error: unknown): void;
  /** Stops the named method rejecting. */
  resolve(method: keyof Client): void;
}

export function createFakeClient(initial: Partial<FakeClientState> = {}): FakeClient {
  const state: FakeClientState = {
    board: initial.board ?? [],
    habits: initial.habits ?? [],
    settings: initial.settings ?? settingsView(),
    timer: initial.timer ?? timerView(),
  };

  const calls: Record<string, number> = {};
  const rejections = new Map<string, unknown>();

  const record = <T>(method: string, produce: () => T): Promise<T> => {
    calls[method] = (calls[method] ?? 0) + 1;

    if (rejections.has(method)) {
      return Promise.reject(rejections.get(method));
    }
    return Promise.resolve(produce());
  };

  const client: Client = {
    Board: () => record('Board', () => state.board),
    Tree: () => record('Tree', () => []),
    Progress: () => record('Progress', () => undefinedProgress),

    CreateNode: () => record('CreateNode', () => node()),
    MoveToColumn: () => record('MoveToColumn', () => node()),
    MoveNode: () => record('MoveNode', () => node()),
    SetDue: () => record('SetDue', () => node()),
    // The only write in this fake that reads one of its arguments: the priority
    // is the whole claim of the palette's four rows, so the node that comes
    // back carries the value that was sent.
    SetPriority: (_nodeID, priority) => record('SetPriority', () => node({ priority })),
    ArchiveNode: () => record('ArchiveNode', () => 1),
    RestoreNode: () => record('RestoreNode', () => 1),

    Search: () => record('Search', () => []),
    HabitStrip: () => record('HabitStrip', () => state.habits),
    CheckHabit: () => record('CheckHabit', () => state.habits),
    UncheckHabit: () => record('UncheckHabit', () => state.habits),
    TimerStart: () => record('TimerStart', () => state.timer),
    TimerStop: () => record('TimerStop', () => state.timer),
    TimerCurrent: () => record('TimerCurrent', () => state.timer),

    Settings: () => record('Settings', () => state.settings),
    SetPalette: (value) =>
      record('SetPalette', () => {
        state.settings = { ...state.settings, palette: value } as SettingsView;
        return state.settings;
      }),
    SetTheme: (value) =>
      record('SetTheme', () => {
        state.settings = { ...state.settings, theme: value } as SettingsView;
        return state.settings;
      }),
    SetAccent: (value) =>
      record('SetAccent', () => {
        state.settings = { ...state.settings, accent: value } as SettingsView;
        return state.settings;
      }),
    SetLanguage: (value) =>
      record('SetLanguage', () => {
        state.settings = { ...state.settings, language: value } as SettingsView;
        return state.settings;
      }),
  };

  return {
    client,
    state,
    calls,
    reject: (method, error) => rejections.set(method as string, error),
    resolve: (method) => rejections.delete(method as string),
  };
}
