import { createElement, type ComponentProps } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { screen } from '@testing-library/react';

import App from '../App';
import { createAppStore } from '../store';
import { board, createFakeClient } from '../test/fakeClient';
import { renderIn } from '../test/render';

// Nexus — the one prop of the DragOverlay that is a DECISION (S3-05, K6, D18).
//
// `dropAnimation` is `null` under prefers-reduced-motion and the library's
// default otherwise, and the answer comes from lib/appearance.ts's
// `prefersReducedMotion` — the SAME helper components/Column.tsx asks for its
// inline transition. Not a second media query: design/tokens.css already kills
// every CSS transition under the preference, but dnd-kit's drop animation is
// driven from JavaScript and is invisible to that rule, so it has to be asked
// for explicitly. Asking a second way would be the rule spelled twice.
//
// # Why this is asserted on the PROP
//
// jsdom runs no animation and composites nothing, so "the card did not fly back
// into place" is not observable here in any form. What IS observable is what the
// board asked the library for, which is the whole of the decision this file
// owns. The visible half is S3-09 item 5.
//
// The rest of the overlay — that the card is drawn once, outside every column,
// and that its seat stays behind holding its space — is asserted against the
// real rendered tree in App.dnd.test.tsx, where the drag machinery already
// lives. This file mocks one component and therefore stays out of that business.

/** Every props object the board has handed to DragOverlay, in order. */
const handed = vi.hoisted(() => [] as { dropAnimation?: unknown }[]);

vi.mock('@dnd-kit/core', async (importActual) => {
  const actual = await importActual<typeof import('@dnd-kit/core')>();

  return {
    ...actual,
    // A pass-through that records. Replacing the component outright would make
    // this a test of a stub; rendering the real one keeps the board's tree the
    // shipping tree, and a prop the real component rejected would still fail.
    DragOverlay: (props: ComponentProps<typeof actual.DragOverlay>) => {
      handed.push(props);
      return createElement(actual.DragOverlay, props);
    },
  };
});

/** Every media query anything asked this window about, in order. */
let asked: string[] = [];

function testWindow(reducedMotion: boolean): Window {
  return {
    document,
    matchMedia: (query: string) => {
      asked.push(query);
      return { matches: reducedMotion, media: query };
    },
  } as unknown as Window;
}

async function showTheBoard(reducedMotion: boolean) {
  const fake = createFakeClient({ board: board(['col-1', 'col-2']) });
  const store = createAppStore(fake.client, { view: testWindow(reducedMotion) });

  await renderIn('en', <App store={store} />);
  await screen.findByRole('region', { name: 'col-1' });
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
  handed.length = 0;
  asked = [];
});

describe('the drag overlay', () => {
  it('is mounted by the board at all', async () => {
    // Non-vacuity. Every assertion below reads the LAST props object handed
    // over, and `undefined` would satisfy a `toBeUndefined()` just as well if
    // the overlay were never rendered.
    await showTheBoard(false);

    expect(handed.length, 'the board rendered no DragOverlay').toBeGreaterThan(0);
  });

  it('refuses the drop animation when the user has asked for less motion', async () => {
    await showTheBoard(true);

    expect(handed.at(-1)?.dropAnimation).toBeNull();
  });

  it('leaves the library its own drop animation otherwise', async () => {
    await showTheBoard(false);

    // `undefined`, not a value of ours: the default belongs to dnd-kit, and
    // naming one here would be a second opinion about how a card falls.
    expect(handed.at(-1)?.dropAnimation).toBeUndefined();
  });

  it('asks the one media query the appearance helper owns, and no other', async () => {
    await showTheBoard(true);

    // If a second query ever appears in this list, the decision has been made
    // twice — which is exactly what routing through lib/appearance.ts prevents.
    expect([...new Set(asked)]).toEqual(['(prefers-reduced-motion: reduce)']);
  });
});
