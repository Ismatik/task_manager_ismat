import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createAppStore, GO_ERROR_KEY, type AppStore } from './index';
import type { Client, HabitView } from '../lib/client';
import { createFakeClient, habitView, node } from '../test/fakeClient';

// Nexus — the habit toggle, at the store level.
//
// The component half is asserted through `render(<App ... />)` in
// App.habits.test.tsx. This half is the part a rendered test cannot pin down
// precisely: WHICH method was called, with WHICH day, in WHICH order, and what
// the store held in the window between the keystroke and Go's answer.

/** A Monday at 09:30 local time, as milliseconds. */
const CLOCK = new Date(2026, 8, 21, 9, 30).getTime();

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

function habit(id: string, overrides: Partial<HabitView> = {}): HabitView {
  return habitView({ node: node({ id, type: 'habit', title: id }), ...overrides });
}

/**
 * A fake Go that records EVERY argument it was called with, and answers with a
 * strip.
 *
 * Local to this file rather than added to test/fakeClient.ts because the point
 * of every case below is the ARGUMENT LIST, and the shared fake deliberately
 * ignores its arguments. The rest parameter is what makes "Go was told no date"
 * an assertion rather than a claim: it captures whatever was actually passed,
 * so an extra argument would show up in `asked` instead of being swallowed by a
 * fixed signature.
 */
function habitGo(initial: HabitView[]) {
  const base = createFakeClient({ habits: initial });
  let strip = initial;

  const asked: { method: string; args: unknown[] }[] = [];
  const refuse = { check: false };
  /** Resolves the pending check, so the optimistic window can be inspected. */
  let release: (() => void) | null = null;

  const answer = (method: string, args: unknown[]): Promise<HabitView[]> => {
    asked.push({ method, args });
    const nodeId = args[0] as string;

    if (refuse.check) {
      return Promise.reject(new Error('service: checking a habit: node is archived'));
    }

    // The cast is test/fakeClient.ts's, for the reason lib/client.ts gives:
    // models.ts declares HabitView as a class, the wire sends a plain object.
    strip = strip.map((entry) =>
      entry.node.id === nodeId
        ? ({
            ...entry,
            checkedToday: method === 'CheckHabitToday',
            streak: entry.streak + 1,
          } as HabitView)
        : entry,
    );

    if (release === null) {
      return Promise.resolve(strip);
    }
    return new Promise((resolve) => {
      release = () => resolve(strip);
    });
  };

  const client: Client = {
    ...base.client,
    HabitStrip: () => Promise.resolve(strip),
    CheckHabitToday: (...args) => answer('CheckHabitToday', args),
    UncheckHabitToday: (...args) => answer('UncheckHabitToday', args),
  };

  return {
    client,
    asked,
    refuse,
    /** Makes the next check hang until `finish()` is called. */
    hold: () => {
      release = () => {};
    },
    finish: () => {
      const resolve = release;
      release = null;
      resolve?.();
    },
    strip: () => strip,
  };
}

function storeOver(client: Client): AppStore {
  return createAppStore(client, { view: testWindow(), now: () => CLOCK });
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('toggling a habit', () => {
  it('checks an unchecked habit, naming the node and nothing else', async () => {
    const go = habitGo([habit('h-1')]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    expect(await store.getState().toggleHabit('h-1')).toBe(true);

    expect(go.asked).toEqual([{ method: 'CheckHabitToday', args: ['h-1'] }]);
    expect(store.getState().habits?.[0].checkedToday).toBe(true);
  });

  it('unchecks a checked habit — the direction is Go’s checkedToday, not a local flag', async () => {
    const go = habitGo([habit('h-1', { checkedToday: true })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    await store.getState().toggleHabit('h-1');

    expect(go.asked).toEqual([{ method: 'UncheckHabitToday', args: ['h-1'] }]);
  });

  it('takes the streak from Go’s answer and never from a local count', async () => {
    // The fixture is deliberately inconsistent with its own history: nothing is
    // checked and the streak is already 4. A frontend that counted would say 0.
    const go = habitGo([habit('h-1', { streak: 4 })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    expect(store.getState().habits?.[0].streak).toBe(4);

    await store.getState().toggleHabit('h-1');

    // 5 because the fake said 5 — the store did not add one, it re-read.
    expect(store.getState().habits?.[0].streak).toBe(5);
  });

  it('ticks the box before the call resolves, and leaves the streak alone', async () => {
    const go = habitGo([habit('h-1', { streak: 4 })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    go.hold();
    const pending = store.getState().toggleHabit('h-1');

    // The optimistic window: the box has moved, the streak has not, because the
    // streak is D5's answer and Go has not given it yet.
    expect(store.getState().habits?.[0].checkedToday).toBe(true);
    expect(store.getState().habits?.[0].streak).toBe(4);

    go.finish();
    await pending;

    expect(store.getState().habits?.[0].streak).toBe(5);
  });

  it('reverts from Go’s answer and raises exactly one toast when refused', async () => {
    const go = habitGo([habit('h-1', { streak: 4 })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    go.refuse.check = true;

    expect(await store.getState().toggleHabit('h-1')).toBe(false);

    // Back to what Go says, which is what a re-read returns — not what the
    // optimistic write put there.
    expect(store.getState().habits?.[0].checkedToday).toBe(false);
    expect(store.getState().habits?.[0].streak).toBe(4);
    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].messageKey).toBe(GO_ERROR_KEY);
  });

  // S2-18, and the reason the binding lost its date parameter.
  //
  // The store used to compute the day and send it. A keystroke at 23:59:30 with
  // a strip read a minute earlier wrote a check for YESTERDAY, Go answered with
  // a strip whose checkedToday was still false, the optimistic tick reverted,
  // and the user saw nothing happen while a check landed on a day they never
  // chose. There is no day to get wrong now, and this is what says so: the same
  // toggle either side of local midnight produces the IDENTICAL call.
  it('sends the same call either side of local midnight — no date, ever', async () => {
    const at = [
      new Date(2026, 8, 21, 23, 59, 30).getTime(),
      new Date(2026, 8, 22, 0, 0, 30).getTime(),
    ];

    const recorded: { method: string; args: unknown[] }[][] = [];
    for (const clock of at) {
      const go = habitGo([habit('h-1')]);
      const store = createAppStore(go.client, { view: testWindow(), now: () => clock });
      await store.getState().loadHabits();

      expect(await store.getState().toggleHabit('h-1')).toBe(true);
      recorded.push(go.asked);
    }

    expect(recorded[0]).toEqual(recorded[1]);
    expect(recorded[0]).toEqual([{ method: 'CheckHabitToday', args: ['h-1'] }]);
    // Spelled out separately from the toEqual above, because a date sent as a
    // SECOND argument is exactly the regression this guards and it has to fail
    // on the count rather than on a value somebody might later update.
    expect(recorded[0][0].args).toHaveLength(1);
  });

  it('does nothing at all for a habit the strip does not hold', async () => {
    const go = habitGo([habit('h-1')]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    expect(await store.getState().toggleHabit('h-nope')).toBe(false);

    expect(go.asked).toEqual([]);
    expect(store.getState().toasts).toEqual([]);
  });

  it('does nothing before the strip has been read', async () => {
    const go = habitGo([habit('h-1')]);
    const store = storeOver(go.client);

    expect(await store.getState().toggleHabit('h-1')).toBe(false);
    expect(go.asked).toEqual([]);
  });

  it('leaves every other habit untouched', async () => {
    const go = habitGo([habit('h-1'), habit('h-2', { checkedToday: true, streak: 9 })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    await store.getState().toggleHabit('h-1');

    const second = store.getState().habits?.[1];
    expect(second?.checkedToday).toBe(true);
    expect(second?.streak).toBe(9);
  });
});
