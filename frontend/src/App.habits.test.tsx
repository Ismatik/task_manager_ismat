import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent, { type UserEvent } from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, HabitView } from './lib/client';
import { createAppStore, type AppStore } from './store';
import { board, columnView, createFakeClient, habitView, node, nodeView } from './test/fakeClient';
import { BOTH_LANGUAGES, renderIn, tabUntil } from './test/render';

// Nexus — the habits strip, asserted THROUGH the shell.
//
// This file imports `App` and does NOT import HabitStrip or HabitChip. That
// import restriction is the assertion, exactly as in App.test.tsx and
// App.keyboard.test.tsx: importing the component you are looking for turns the
// test into a test of that component and proves nothing about it being in the
// application. Every case below is driven by `user-event` key presses.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

/** A Monday at 09:30 local time. Nothing on the wire depends on it (S2-18). */
const CLOCK = new Date(2026, 8, 21, 9, 30).getTime();

function habit(id: string, title: string, overrides: Partial<HabitView> = {}): HabitView {
  return habitView({ node: node({ id, type: 'habit', title }), ...overrides });
}

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

/** A fake Go with a board and a strip, recording what the strip was asked. */
function stripGo(habits: HabitView[], cards: ColumnView[] = board(COLUMNS, [nodeView()])) {
  const base = createFakeClient();
  let strip = habits;

  const asked: { method: string; args: unknown[] }[] = [];
  const refuse = { check: false };

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
        ? ({ ...entry, checkedToday: method === 'CheckHabitToday' } as HabitView)
        : entry,
    );
    return Promise.resolve(strip);
  };

  const client: Client = {
    ...base.client,
    Board: () => Promise.resolve(cards),
    HabitStrip: () => Promise.resolve(strip),
    CheckHabitToday: (...args) => answer('CheckHabitToday', args),
    UncheckHabitToday: (...args) => answer('UncheckHabitToday', args),
  };

  return { client, asked, refuse };
}

function storeOver(client: Client): AppStore {
  return createAppStore(client, { view: testWindow(), now: () => CLOCK });
}

/** Renders the shell and waits for the strip to arrive. */
async function enterTheApp(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('checkbox');

  return user;
}

/**
 * Tabs forward until a habit chip has focus.
 *
 * Not a single Tab: S2-21 filled region 1 with the appearance controls, which
 * come before the strip in the shell's DOM order. How many stops they add is
 * not this file's business, and the strip being ONE stop — the claim below — is
 * asserted separately, on its tabindex.
 */
async function tabToTheStrip(user: UserEvent) {
  await tabUntil(user, () => focusedHabitId() !== null);
}

/** The id of the chip that currently has focus, or null. */
function focusedHabitId(): string | null {
  const active = document.activeElement;
  if (!(active instanceof HTMLElement)) {
    return null;
  }
  return active.dataset.habitId ?? null;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('the habits strip', () => {
  it('is in the application, hydrated from HabitStrip()', async () => {
    const go = stripGo([habit('h-1', 'Read'), habit('h-2', 'Stretch')]);

    await enterTheApp(storeOver(go.client));

    expect(screen.getByRole('checkbox', { name: /Read/ })).toBeInTheDocument();
    expect(screen.getByRole('checkbox', { name: /Stretch/ })).toBeInTheDocument();
  });

  it('renders the streak Go supplied, in font-mono, however inconsistent it looks', async () => {
    // Nothing is checked and the streak is already 4. A frontend that counted
    // anything would say 0 — and a streak is not a count of days anyway (D5:
    // consecutive SCHEDULED occurrences), which is the whole reason the number
    // is Go's.
    const go = stripGo([habit('h-1', 'Read', { checkedToday: false, streak: 4 })]);

    await enterTheApp(storeOver(go.client));

    const streak = screen.getByText('4');
    expect(streak).toBeInTheDocument();
    expect(streak).toHaveClass('font-mono');
  });

  it('shows a habit that is not scheduled today rather than hiding it', async () => {
    // Presentation over S2-04's flag, not a recomputation of the schedule: the
    // chip is dimmed and says so, and the flag it wears is the DTO's.
    const go = stripGo([
      habit('h-1', 'Read', { scheduledToday: true }),
      habit('h-2', 'Stretch', { scheduledToday: false }),
    ]);

    await enterTheApp(storeOver(go.client));

    expect(screen.getByRole('checkbox', { name: /Read/ })).toHaveAttribute(
      'data-scheduled',
      'true',
    );
    const off = screen.getByRole('checkbox', { name: /Stretch/ });
    expect(off).toHaveAttribute('data-scheduled', 'false');
    expect(off).toHaveTextContent('Not scheduled today');
  });

  // --------------------------------------------------------------- keyboard

  it('is reachable from the board with Shift+Tab, and the board from it with Tab', async () => {
    // The shell's DOM order is the region order and nothing else: the strip is
    // region 2 and the board is region 3, so the strip is one Shift+Tab back
    // from the board and the board is one Tab on from the strip.
    const go = stripGo([habit('h-1', 'Read')]);
    const user = await enterTheApp(storeOver(go.client));

    await tabToTheStrip(user);
    expect(focusedHabitId()).toBe('h-1');

    await user.tab();
    expect(focusedHabitId()).toBeNull();
    expect(document.activeElement).toHaveAttribute('data-node-id');

    await user.tab({ shift: true });
    expect(focusedHabitId()).toBe('h-1');
  });

  it('is ONE tab stop, with the arrows moving inside it', async () => {
    const go = stripGo([habit('h-1', 'Read'), habit('h-2', 'Stretch'), habit('h-3', 'Walk')]);
    const user = await enterTheApp(storeOver(go.client));

    const tabStops = () => document.querySelectorAll('[data-habit-id][tabindex="0"]');

    await tabToTheStrip(user);
    expect(focusedHabitId()).toBe('h-1');
    expect(tabStops()).toHaveLength(1);

    await user.keyboard('{ArrowRight}');
    expect(focusedHabitId()).toBe('h-2');
    expect(tabStops()).toHaveLength(1);
    expect(tabStops()[0]).toHaveFocus();

    await user.keyboard('{End}');
    expect(focusedHabitId()).toBe('h-3');

    // Clamped at the end, never wrapped.
    await user.keyboard('{ArrowRight}');
    expect(focusedHabitId()).toBe('h-3');

    await user.keyboard('{Home}');
    expect(focusedHabitId()).toBe('h-1');

    await user.keyboard('{ArrowLeft}');
    expect(focusedHabitId()).toBe('h-1');
  });

  it('toggles today’s check on Space, naming the node and no day', async () => {
    const go = stripGo([habit('h-1', 'Read')]);
    const user = await enterTheApp(storeOver(go.client));

    await tabToTheStrip(user);
    await user.keyboard(' ');

    await waitFor(() =>
      expect(go.asked).toEqual([{ method: 'CheckHabitToday', args: ['h-1'] }]),
    );
    await waitFor(() => expect(screen.getByRole('checkbox', { name: /Read/ })).toBeChecked());

    // And back again — the direction is Go's checkedToday, not a local flag.
    await user.keyboard(' ');

    await waitFor(() => expect(go.asked).toHaveLength(2));
    expect(go.asked[1].method).toBe('UncheckHabitToday');
    expect(go.asked[1].args).toEqual(['h-1']);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: /Read/ })).not.toBeChecked());
  });

  it('reverts and raises exactly one toast when Go refuses the check', async () => {
    const go = stripGo([habit('h-1', 'Read')]);
    go.refuse.check = true;
    const user = await enterTheApp(storeOver(go.client));

    await tabToTheStrip(user);
    await user.keyboard(' ');

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');

    // Reverted from Go's answer, so the box is back where Go says it is — and
    // focus never left the chip.
    await waitFor(() => expect(screen.getByRole('checkbox', { name: /Read/ })).not.toBeChecked());
    expect(focusedHabitId()).toBe('h-1');
  });

  it('keeps Space out of the board — a card is not a habit', async () => {
    const go = stripGo([habit('h-1', 'Read')]);
    const user = await enterTheApp(storeOver(go.client));

    await tabToTheStrip(user);
    await user.tab();
    expect(document.activeElement).toHaveAttribute('data-node-id');

    await user.keyboard(' ');

    expect(go.asked).toEqual([]);
  });

  // ------------------------------------------------------- empty and broken

  it('renders no DOM at all when Go returns no habits', async () => {
    const go = stripGo([]);
    const store = storeOver(go.client);

    const { container } = await renderIn('en', <App store={store} />);
    await screen.findByRole('region', { name: COLUMNS[0] });

    // An empty region is an empty slot, not a bar of nothing across the top.
    // Two children under the shell: region 1's header, which S2-21 filled, and
    // region 3's board. The strip is not one of them.
    expect(screen.queryByRole('group')).toBeNull();
    expect([...container.firstElementChild!.children].map((child) => child.tagName)).toEqual([
      'HEADER',
      'MAIN',
    ]);
  });

  it('shows the board and a toast when the strip cannot be read', async () => {
    const fake = createFakeClient({ board: board(COLUMNS) });
    fake.reject('HabitStrip', new Error('sql: database is closed'));
    const store = storeOver(fake.client);

    await renderIn('en', <App store={store} />);

    await screen.findByRole('alert');
    expect(await screen.findByRole('region', { name: COLUMNS[0] })).toBeInTheDocument();
    expect(screen.queryByRole('group')).toBeNull();
  });

  // --------------------------------------------------------------- Russian

  it.each(BOTH_LANGUAGES)('renders in %s with nothing hard-coded', async (language) => {
    const go = stripGo(
      [habit('h-1', 'Читать книгу каждый вечер', { streak: 12 })],
      [columnView(COLUMNS[0], [nodeView()])],
    );

    await enterTheApp(storeOver(go.client), language);

    const chip = screen.getByRole('checkbox');
    const group = screen.getByRole('group');

    expect(chip).toHaveTextContent('Читать книгу каждый вечер');
    expect(screen.getByText('12')).toHaveClass('font-mono');
    // The group's name is translated, so it is never the raw key.
    expect(group.getAttribute('aria-label')).not.toContain('habits.');
    // Nothing is clipped or held to one line: the strip wraps and the title
    // breaks on word boundaries, which is what survives a 30% wider language.
    expect(group).toHaveClass('flex-wrap');
    expect(chip.querySelector('.break-words')).not.toBeNull();
  });
});
