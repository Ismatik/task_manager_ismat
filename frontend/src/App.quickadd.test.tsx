import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

import App from './App';
import type { Client, ColumnView, NewNode, Node, NodeView } from './lib/client';
import { createAppStore, type AppStore } from './store';
import { columnView, createFakeClient, node, nodeView } from './test/fakeClient';
import { BOTH_LANGUAGES, renderIn, tabUntil } from './test/render';

// Nexus — quick add, asserted THROUGH the shell.
//
// This file imports `App` and does NOT import QuickAdd. The import restriction
// is the assertion: this is step 1 of the stage's ACCEPT flow, so it has to
// work where the user is and not where the test is. Everything below is driven
// by `user-event` key presses; there is not a click in the file.
//
// No status name is written down either (`make guard` check 2): the column the
// new card lands in is whichever one Go put it in.

const COLUMNS = ['col-1', 'col-2', 'col-3', 'col-4', 'col-5'];

/**
 * A fake Go that actually stores what it is asked to create.
 *
 * Local rather than shared, for the reason App.keyboard.test.tsx gives about
 * its own: the whole point here is the DRAFT and the card that appears because
 * of it, and the shared fake ignores its arguments and answers identically
 * every time.
 */
function creatingGo(seed: NodeView[] = []) {
  const base = createFakeClient();
  let board: ColumnView[] = COLUMNS.map((status, index) =>
    columnView(status, index === 0 ? seed : []),
  );

  const drafts: NewNode[] = [];
  const refusal = { error: null as Error | null };
  let nextId = 1;

  const client: Client = {
    ...base.client,
    Board: () => Promise.resolve(board),
    CreateNode: (draft: NewNode) => {
      drafts.push(draft);

      if (refusal.error !== null) {
        return Promise.reject(refusal.error);
      }

      const created: Node = node({ id: `new-${nextId}`, title: draft.title, type: draft.type });
      nextId += 1;

      board = board.map((column, index) =>
        index === 0 ? columnView(column.status, [...column.nodes, nodeView({ node: created })]) : column,
      );
      return Promise.resolve(created);
    },
  };

  return { client, drafts, refusal };
}

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

function storeOver(client: Client): AppStore {
  return createAppStore(client, { view: testWindow() });
}

/** Renders the shell and waits for the board. */
async function enterTheApp(store: AppStore, language = 'en') {
  const user = userEvent.setup();

  await renderIn(language, <App store={store} />);
  await screen.findAllByRole('region');

  return user;
}

const OPEN = '{Control>}n{/Control}';

/**
 * The quick add's own title field.
 *
 * Scoped to the overlay rather than `screen.getByRole('textbox')`: S2-21 filled
 * region 1 with the appearance controls, whose accent field is a textbox too,
 * so "the only textbox on the page" stopped being a way to say "this one".
 */
function titleField(): HTMLElement {
  return within(screen.getByRole('dialog')).getByRole('textbox');
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('quick add', () => {
  it('is in the application: Ctrl+N opens it with focus in the title field', async () => {
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client));

    expect(screen.queryByRole('dialog')).toBeNull();

    await user.keyboard(OPEN);

    const overlay = await screen.findByRole('dialog');
    expect(overlay).toHaveAttribute('aria-modal', 'true');
    expect(titleField()).toHaveFocus();
  });

  it('creates a node with the typed title, calling CreateNode exactly once', async () => {
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    await user.keyboard('Fix the tap{Enter}');

    await waitFor(() => expect(go.drafts).toHaveLength(1));
    expect(go.drafts[0].title).toBe('Fix the tap');
    // The type is the default the label table's order supplies, and the rest of
    // the draft is at its zero value: the defaults are Go's, and a frontend
    // that filled them in would be inventing them.
    expect(go.drafts[0].type).toBe('task');
    expect(go.drafts[0].status).toBe('');
    expect(go.drafts[0].priority).toBe(0);
    expect(go.drafts[0].descriptionMd).toBe('');
    expect(go.drafts[0].recurrence).toBeUndefined();
    expect(go.drafts[0].due).toBeUndefined();

    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
  });

  it('lands focus on the new card once the board has been re-read', async () => {
    // This is what makes step 2 of the ACCEPT script flow into step 3: the card
    // that was just created is the card Ctrl+Shift+Right will move.
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    await user.keyboard('Fix the tap{Enter}');

    await waitFor(() => {
      const active = document.activeElement as HTMLElement | null;
      expect(active?.dataset.nodeId).toBe('new-1');
    });
    expect(document.activeElement).toHaveTextContent('Fix the tap');
  });

  it('closes on Escape and gives focus back to where it came from', async () => {
    const go = creatingGo([nodeView({ node: node({ id: 'n-1', title: 'Already here' }) })]);
    const user = await enterTheApp(storeOver(go.client));

    // Park focus on a real element the board owns, so "where it came from" is
    // something other than the body. Tab UNTIL, because region 1's appearance
    // controls (S2-21) come before the board in the shell's DOM order.
    await tabUntil(user, () => document.activeElement?.hasAttribute('data-node-id') === true);
    const opener = document.activeElement;
    expect((opener as HTMLElement).dataset.nodeId).toBe('n-1');

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    expect(titleField()).toHaveFocus();

    await user.keyboard('{Escape}');

    await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull());
    expect(document.activeElement).toBe(opener);
    expect(go.drafts).toEqual([]);
  });

  it('traps focus: Tab cycles inside the overlay and never leaves it', async () => {
    const go = creatingGo([nodeView({ node: node({ id: 'n-1', title: 'Already here' }) })]);
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    const overlay = await screen.findByRole('dialog');

    const seen = new Set<string | null>();
    for (let step = 0; step < 6; step += 1) {
      await user.tab();
      const active = document.activeElement as HTMLElement;
      expect(overlay.contains(active)).toBe(true);
      seen.add(active.getAttribute('role') ?? active.tagName);
    }

    // It really cycles rather than sticking on one element.
    expect(seen.size).toBeGreaterThan(1);

    await user.tab({ shift: true });
    expect(overlay.contains(document.activeElement)).toBe(true);
  });

  it('chooses the type with the arrows, and sends the one that was chosen', async () => {
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    await user.keyboard('Ship it');

    // Tab from the title reaches the group's one tab stop — the selected radio.
    await user.tab();
    expect(document.activeElement).toHaveAttribute('role', 'radio');
    expect(document.activeElement).toHaveAttribute('aria-checked', 'true');

    await user.keyboard('{ArrowRight}');

    const chosen = (document.activeElement as HTMLElement).dataset.type;
    expect(chosen).not.toBe('task');
    expect(document.activeElement).toHaveAttribute('aria-checked', 'true');
    expect(document.querySelectorAll('[role="radio"][aria-checked="true"]')).toHaveLength(1);

    await user.tab();
    await user.keyboard('{Enter}');

    await waitFor(() => expect(go.drafts).toHaveLength(1));
    expect(go.drafts[0].type).toBe(chosen);
  });

  it('offers exactly the types the label table names, and no fallback entry', async () => {
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');

    const offered = screen.getAllByRole('radio').map((radio) => radio.getAttribute('data-type'));

    expect(offered).toEqual(['task', 'project', 'habit', 'note', 'bug']);
  });

  // ------------------------------------------------------------- refusals

  it('surfaces Go’s refusal of a habit with no recurrence, and stays open', async () => {
    // The rule is domain's — a habit must carry a recurrence rule — and it is
    // asked exactly once, in Go. There is no recurrence check in frontend/src.
    const go = creatingGo();
    go.refusal.error = new Error('service: creating a node: a habit needs a recurrence rule');
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    await user.keyboard('Run every morning');

    await user.tab();
    await user.keyboard('{ArrowRight}{ArrowRight}');
    expect((document.activeElement as HTMLElement).dataset.type).toBe('habit');

    await user.tab();
    await user.keyboard('{Enter}');

    const alert = await screen.findByRole('alert');
    expect(screen.getAllByRole('alert')).toHaveLength(1);
    expect(alert).toHaveTextContent('Nexus could not finish that.');
    // The Go sentence is for the console, never the screen.
    expect(alert).not.toHaveTextContent('recurrence');

    // Still open, still holding the title — a refusal must not cost the typing.
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    expect(titleField()).toHaveValue('Run every morning');
    expect(go.drafts[0].type).toBe('habit');
  });

  it('sends an empty title to Go and shows Go’s refusal', async () => {
    // Deliberately no guard in the component: `Node.Validate` refuses it, and a
    // second refusal in TypeScript would be the rule written twice.
    const go = creatingGo();
    go.refusal.error = new Error('domain: a node needs a title');
    const user = await enterTheApp(storeOver(go.client));

    await user.keyboard(OPEN);
    await screen.findByRole('dialog');
    await user.keyboard('   {Enter}');

    await waitFor(() => expect(go.drafts).toHaveLength(1));
    expect(go.drafts[0].title).toBe('   ');
    expect(await screen.findByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('dialog')).toBeInTheDocument();
  });

  // -------------------------------------------------------------- Russian

  it.each(BOTH_LANGUAGES)('renders in %s with nothing hard-coded', async (language) => {
    const go = creatingGo();
    const user = await enterTheApp(storeOver(go.client), language);

    await user.keyboard(OPEN);
    const overlay = await screen.findByRole('dialog');

    // Every visible string resolved to a translation, so none of them is the
    // raw key and none is English sitting in a Russian panel.
    expect(overlay.textContent).not.toContain('quickAdd.');
    expect(overlay.textContent).not.toContain('card.type.');
    expect(titleField().getAttribute('aria-label')).not.toContain('quickAdd.');

    // Nothing is held to one line or clipped: the type row wraps and the
    // buttons break on word boundaries, which is what survives a 30% wider
    // language at 1024px.
    expect(screen.getByRole('radiogroup')).toHaveClass('flex-wrap');
    for (const radio of screen.getAllByRole('radio')) {
      expect(radio).toHaveClass('break-words');
    }
  });
});
