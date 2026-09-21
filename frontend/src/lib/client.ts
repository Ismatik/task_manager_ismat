import * as App from '../../wailsjs/go/main/App';
import type { domain, service } from '../../wailsjs/go/models';

// Nexus — the single place TypeScript talks to Go.
//
// # Why one file
//
// This is the ONLY file in frontend/src that imports from wailsjs, and that is
// an acceptance criterion, not a convention:
//
//     git grep -rn "wailsjs" frontend/src --files-with-matches
//
// must list this file and nothing else. Two consequences follow, and both are
// the point. The entire call surface onto Go is greppable in one place, so
// "what does the frontend ask Go for" has an answer you can read in a minute.
// And every consumer takes a Client, so a test supplies a fake rather than
// standing up a window — which is what makes "a rejected call raises exactly
// one toast" a thing a machine can assert.
//
// # Nothing here decides anything
//
// Every method is a delegation. No branch, no default, no retry, no caching, no
// translation of a Go error into a user-facing sentence. app.go is already
// "every method here is a delegation and nothing else"; this is the mirror of
// that statement on the other side of the wire. Errors are not caught here
// either: a bound method returns (T, error), Wails turns the error into a
// rejected promise, and the STORE surfaces it in a toast. Catching here would
// put the no-silent-failures rule in two places.
//
// # The generator quirk, and why it is not "fixed"
//
// Every nullable Go field — Node.parentId, Node.due, TimerView.startedAt and
// the rest — is sent as `null` on the wire but generated into models.ts as an
// optional `?`, i.e. TypeScript `undefined`. The two do not meet: `x === null`
// is false for undefined and `x === undefined` is false for null. The fix is
// NOT to add omitempty in Go, which would change the wire contract from "the
// field is there and null" to "the field is absent" and break every consumer
// that reads it. The fix is to read tolerantly, which is what `optional()`
// below is for and why every consumer uses `??` or truthiness.

// The DTOs, re-exported.
//
// This is the same rule applied to TYPES, and it matters for the same reason:
// the criterion is that `git grep -rn "wailsjs" frontend/src` names this file
// and no other, and an `import type { service } from '../../wailsjs/...'` in a
// component is still a second place that knows where the generated code lives.
// Routing the types through here as well means one import path to change if the
// generator ever moves, and one list of what actually crosses the wire.
export type Node = domain.Node;
export type Tag = domain.Tag;

export type NodeView = service.NodeView;
export type ColumnView = service.ColumnView;
export type HabitView = service.HabitView;
export type ProgressView = service.ProgressView;
export type TimerView = service.TimerView;
export type SettingsView = service.SettingsView;
export type NewNode = service.NewNode;
export type SearchOptions = service.SearchOptions;

/**
 * The bound surface of app.go, as an interface.
 *
 * Every method corresponds one-to-one to a Wails binding, and
 * `unwrappedBindingNames()` below fails if one is ever added in Go and
 * forgotten here.
 */
export interface Client {
  // The board and the tree.
  Board(): Promise<service.ColumnView[]>;
  Tree(rootID: string): Promise<service.NodeView[]>;
  Progress(nodeID: string): Promise<service.ProgressView>;

  // Writes.
  CreateNode(draft: service.NewNode): Promise<domain.Node>;
  MoveToColumn(nodeID: string, target: string): Promise<domain.Node>;
  MoveNode(nodeID: string, newParentID: string, toIndex: number): Promise<domain.Node>;
  SetDue(nodeID: string, due: string): Promise<domain.Node>;
  // `number`, and deliberately NOT a union of 1 | 2 | 3 | 4. Which numbers are
  // priorities is domain.Priority.Valid's answer, asked in Go through
  // domain.Node.Validate; a union here would be that range written a second
  // time, in the one language that cannot see the domain, and it is the copy
  // that would go stale. The generator says the same thing —
  // wailsjs/go/main/App.d.ts declares `arg2:number` — and so does models.ts,
  // where domain.Node.priority has always been a `number`.
  SetPriority(nodeID: string, priority: number): Promise<domain.Node>;
  ArchiveNode(nodeID: string): Promise<number>;
  RestoreNode(nodeID: string): Promise<number>;

  // Search, habits and the timer.
  Search(query: string, opts: service.SearchOptions): Promise<service.NodeView[]>;
  HabitStrip(): Promise<service.HabitView[]>;
  // No date parameter, and that is the point (S2-18): which calendar day the
  // present one is is domain.Today(clock)'s answer, asked in Go by the same
  // call HabitStrip derives checkedToday against. A date here would be a day
  // this side had to compute, which is a second clock for one rule — and one
  // minute either side of local midnight the two disagree. app.go's
  // CheckHabitToday says the rest.
  CheckHabitToday(nodeID: string): Promise<service.HabitView[]>;
  UncheckHabitToday(nodeID: string): Promise<service.HabitView[]>;
  TimerStart(nodeID: string): Promise<service.TimerView>;
  TimerStop(): Promise<service.TimerView>;
  TimerCurrent(): Promise<service.TimerView>;

  // Settings.
  Settings(): Promise<service.SettingsView>;
  SetPalette(value: string): Promise<service.SettingsView>;
  SetTheme(value: string): Promise<service.SettingsView>;
  SetAccent(value: string): Promise<service.SettingsView>;
  SetLanguage(value: string): Promise<service.SettingsView>;
}

/**
 * The real client: the generated bindings, by reference.
 *
 * By reference and not wrapped in arrow functions, deliberately — it is what
 * lets unwrappedBindingNames() compare identities and so notice a binding that
 * exists in Go and is missing from the Client interface. A wrapper would make
 * that check impossible and the coverage claim unverifiable.
 */
export const wailsClient: Client = {
  Board: App.Board,
  Tree: App.Tree,
  Progress: App.Progress,

  CreateNode: App.CreateNode,
  MoveToColumn: App.MoveToColumn,
  MoveNode: App.MoveNode,
  SetDue: App.SetDue,
  SetPriority: App.SetPriority,
  ArchiveNode: App.ArchiveNode,
  RestoreNode: App.RestoreNode,

  Search: App.Search,
  HabitStrip: App.HabitStrip,
  CheckHabitToday: App.CheckHabitToday,
  UncheckHabitToday: App.UncheckHabitToday,
  TimerStart: App.TimerStart,
  TimerStop: App.TimerStop,
  TimerCurrent: App.TimerCurrent,

  Settings: App.Settings,
  SetPalette: App.SetPalette,
  SetTheme: App.SetTheme,
  SetAccent: App.SetAccent,
  SetLanguage: App.SetLanguage,
};

/**
 * Names of generated bindings that `wailsClient` does not expose. Empty by
 * contract, and asserted to be empty by client.test.ts.
 *
 * This exists because "wraps every binding" is otherwise a claim nobody checks
 * again. A method added to app.go regenerates App.js on the next `wails build`,
 * and without this the frontend simply would not know it existed. It is also
 * why the test does not import wailsjs itself: that would put a second importer
 * in frontend/src and break the criterion this file is built around.
 */
export function unwrappedBindingNames(): string[] {
  const exposed = new Set<unknown>(Object.values(wailsClient));

  return Object.entries(App)
    .filter(([, binding]) => typeof binding === 'function' && !exposed.has(binding))
    .map(([name]) => name)
    .sort();
}

/**
 * Reads a nullable wire field as `T | null`, whichever of null or undefined
 * actually arrived.
 *
 * Go sends `null`; the generator declares `?`, which is `undefined`. Rather
 * than have every call site remember which one it is dealing with, they call
 * this. It converts nothing and defaults nothing — an absent value stays
 * absent, it just has one spelling afterwards.
 */
export function optional<T>(value: T | null | undefined): T | null {
  return value ?? null;
}
