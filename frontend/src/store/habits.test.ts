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

/** What `wireDate` must make of CLOCK — the local calendar day, never UTC's. */
const CLOCK_DAY = '2026-09-21';

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
 * A fake Go that records the day it was asked about and answers with a strip.
 *
 * Local to this file rather than added to test/fakeClient.ts because the point
 * of every case below is the ARGUMENT, and the shared fake deliberately ignores
 * its arguments.
 */
function habitGo(initial: HabitView[]) {
  const base = createFakeClient({ habits: initial });
  let strip = initial;

  const asked: { method: string; nodeId: string; day: string }[] = [];
  const refuse = { check: false };
  /** Resolves the pending CheckHabit, so the optimistic window can be inspected. */
  let release: (() => void) | null = null;

  const answer = (method: string, nodeId: string, day: string): Promise<HabitView[]> => {
    asked.push({ method, nodeId, day });

    if (refuse.check) {
      return Promise.reject(new Error('service: checking a habit: node is archived'));
    }

    // The cast is test/fakeClient.ts's, for the reason lib/client.ts gives:
    // models.ts declares HabitView as a class, the wire sends a plain object.
    strip = strip.map((entry) =>
      entry.node.id === nodeId
        ? ({
            ...entry,
            checkedToday: method === 'CheckHabit',
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
    CheckHabit: (nodeID, date) => answer('CheckHabit', nodeID, date),
    UncheckHabit: (nodeID, date) => answer('UncheckHabit', nodeID, date),
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
  it('checks an unchecked habit, on the local calendar day', async () => {
    const go = habitGo([habit('h-1')]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    expect(await store.getState().toggleHabit('h-1')).toBe(true);

    expect(go.asked).toEqual([{ method: 'CheckHabit', nodeId: 'h-1', day: CLOCK_DAY }]);
    expect(store.getState().habits?.[0].checkedToday).toBe(true);
  });

  it('unchecks a checked habit — the direction is Go’s checkedToday, not a local flag', async () => {
    const go = habitGo([habit('h-1', { checkedToday: true })]);
    const store = storeOver(go.client);
    await store.getState().loadHabits();

    await store.getState().toggleHabit('h-1');

    expect(go.asked).toEqual([{ method: 'UncheckHabit', nodeId: 'h-1', day: CLOCK_DAY }]);
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
