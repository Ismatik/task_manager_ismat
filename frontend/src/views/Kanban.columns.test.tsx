import { beforeEach, describe, expect, it, vi } from 'vitest';
import { screen } from '@testing-library/react';

import App from '../App';
import { createAppStore } from '../store';
import { board, createFakeClient, nodeView, node } from '../test/fakeClient';
import { renderIn } from '../test/render';

// Nexus — the columns fill the board and scroll inside themselves (S3-31, K15,
// D29).
//
// # WHAT THIS FILE CAN AND CANNOT PROVE — read this before trusting it
//
// jsdom has NO LAYOUT ENGINE, and the vitest config sets `css: false`, so no
// stylesheet is ever applied to the rendered tree. Every `offsetHeight` here is
// 0, nothing overflows and nothing scrolls. This file therefore **cannot**
// observe that a column is 600px tall, that a column with twelve cards scrolls
// inside itself, or that a card dropped on the empty space under the last card
// landed in the right column. It asserts the CLASS CONTRACT and the SHAPE of
// the tree, which is the half a machine can hold.
//
// The other half is split in two and both halves are owed elsewhere. The static
// picture — columns filling the window at a real 1024×768, in every palette,
// theme and language — is S3-32's `make shots`. The drop itself and the
// in-column scroll need INPUT, which cannot be driven on this machine at all
// (E4: no xdotool, no xte, no wmctrl, no window manager), and they are S3-09's.
//
// # Why the droppable node is asserted by capturing the ref
//
// "The `useDroppable` ref sits on the stretched element, not on an inner
// content box" is a claim about which DOM node dnd-kit was handed, and reading
// classes off a `data-column` element only shows that SOME element has them. So
// `useDroppable` is wrapped — the real hook still runs, and the real ref is
// still set — and the node it receives is recorded and compared against the
// board's own child. Same technique as Kanban.overlay.test.tsx, for the same
// reason: the shipping tree stays the shipping tree.

/** Every element dnd-kit's droppable ref was handed, in order. */
const droppableNodes = vi.hoisted(() => [] as { id: unknown; node: HTMLElement }[]);

vi.mock('@dnd-kit/core', async (importActual) => {
  const actual = await importActual<typeof import('@dnd-kit/core')>();

  return {
    ...actual,
    useDroppable: (args: Parameters<typeof actual.useDroppable>[0]) => {
      const result = actual.useDroppable(args);

      return {
        ...result,
        setNodeRef: (element: HTMLElement | null) => {
          if (element !== null) {
            droppableNodes.push({ id: args.id, node: element });
          }
          result.setNodeRef(element);
        },
      };
    },
  };
});

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

/**
 * A board whose first column holds three cards and whose second holds none.
 *
 * Both shapes matter here: the card list only exists where there are cards, and
 * the EMPTY column is the one K15 is actually about — it was a header-high
 * strip, and the region under it belonged to the board rather than to any
 * column.
 */
async function showTheBoard() {
  const cards = ['card-1', 'card-2', 'card-3'].map((id) =>
    nodeView({ node: node({ id, title: id }) }),
  );
  const fake = createFakeClient({ board: board(['col-1', 'col-2'], cards) });
  const store = createAppStore(fake.client, { view: testWindow() });

  const rendered = await renderIn('en', <App store={store} />);
  await screen.findByRole('region', { name: 'col-1' });
  return rendered;
}

/** The board row: the parent of the columns, found through a column. */
function boardRow(): HTMLElement {
  const column = document.querySelector('[data-column]');
  expect(column, 'no column was rendered, so nothing below checked anything').not.toBeNull();

  const row = column?.parentElement ?? null;
  expect(row, 'the board row was not found').not.toBeNull();
  return row as HTMLElement;
}

/** One column section, by the status Go sent. */
function columnFor(status: string): HTMLElement {
  const section = document.querySelector<HTMLElement>(`[data-column="${status}"]`);
  expect(section, `no column for ${status}`).not.toBeNull();
  return section as HTMLElement;
}

/** The card list inside a column, which is the element that scrolls. */
function cardListIn(section: HTMLElement): HTMLElement {
  const list = section.querySelector('ul');
  expect(list, 'the column has no card list').not.toBeNull();
  return list as HTMLElement;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
  droppableNodes.length = 0;
});

describe('the board row', () => {
  it('renders the columns it is supposed to, so the rest of this file is not vacuous', async () => {
    await showTheBoard();

    expect(document.querySelectorAll('[data-column]')).toHaveLength(2);
    expect(cardListIn(columnFor('col-1')).querySelectorAll('li')).toHaveLength(3);
  });

  it('does not hold its columns at the top', async () => {
    await showTheBoard();

    // THE defect (K15). `items-start` overrides flex's default `stretch`, so
    // every column was as tall as its cards — which on a real capture made four
    // of five columns short boxes at the top of a large empty area, and left
    // `useDroppable` almost no rectangle to catch a drop in.
    expect(boardRow()).not.toHaveClass('items-start');
  });

  it('keeps the horizontal scroll and a height to scroll inside', async () => {
    await showTheBoard();
    const row = boardRow();

    // D29 leaves this half of D22 alone: five columns that will not fit even at
    // their floor width still scroll sideways, here.
    expect(row).toHaveClass('overflow-x-auto');
    // And the columns can only stretch to a height this element actually has.
    expect(row).toHaveClass('h-full');
  });

  it('does not replace items-start with a typed height anywhere', async () => {
    await showTheBoard();

    // D29 part 2, and D21 behind it: the layout floor is DERIVED, never typed.
    // A `min-h-[…]`, an `h-[…]` or a `vh` on the row or on any column would be
    // a second spelling of the height chain, passing every other test in this
    // file while re-introducing the defect D21 was written for.
    const suspects = [boardRow(), columnFor('col-1'), columnFor('col-2')];

    for (const element of suspects) {
      expect(element.className).not.toMatch(/\b(min-)?h-\[/);
      expect(element.className).not.toMatch(/\bh-screen\b/);
    }
  });
});

describe('the column', () => {
  it('is the droppable, and the droppable is the element the board stretches', async () => {
    await showTheBoard();
    const row = boardRow();

    // Non-vacuity: the wrapper above must actually have seen dnd-kit hand over
    // a node, or every assertion here is about an empty list.
    expect(droppableNodes.length, 'no droppable ref was ever set').toBeGreaterThan(0);

    for (const status of ['col-1', 'col-2']) {
      const captured = droppableNodes.filter((entry) => entry.id === status).at(-1);
      expect(captured, `no droppable was registered for ${status}`).toBeDefined();

      // The node dnd-kit holds IS the section, and the section is a DIRECT
      // CHILD of the board row — which is what makes it the flex item that
      // stretches. A ref on an inner content box would satisfy a class check
      // and still be the wrong rectangle.
      expect(captured?.node).toBe(columnFor(status));
      expect(captured?.node.parentElement).toBe(row);
    }
  });

  it('may shrink below its content, so its card list can', async () => {
    await showTheBoard();

    // The middle link of the chain. The section is itself a COLUMN flex
    // container, so `min-height: auto` on it would stop the list below from
    // ever shrinking — and `overflow-y-auto` on a list that never shrinks below
    // its content has nothing to do.
    expect(columnFor('col-1')).toHaveClass('min-h-0');
    expect(columnFor('col-2')).toHaveClass('min-h-0');
  });

  it('is not itself the vertical scroller', async () => {
    await showTheBoard();

    // The scroll belongs to the card list, one element down. Putting it here
    // would take the heading with it, which is the thing D29 part 4 forbids.
    expect(columnFor('col-1')).not.toHaveClass('overflow-y-auto');
    expect(columnFor('col-1')).not.toHaveClass('overflow-y-scroll');
  });
});

describe('the card list', () => {
  it('is the vertical scroller, and states both axes', async () => {
    await showTheBoard();
    const list = cardListIn(columnFor('col-1'));

    expect(list).toHaveClass('overflow-y-auto');
    // Both axes NAMED, for S3-01's reason one element up: CSS Overflow 3
    // promotes a `visible` companion axis to `auto`, so the horizontal axis
    // becomes a scroll container here the moment the vertical one does —
    // written down rather than left to a spec rule nobody reads.
    //
    // `auto` and NOT `hidden`, and that is a decision the S3-02 shrink audit
    // made rather than a preference: `overflow-x-hidden` CLIPS, test/shrink.ts
    // reports it on an element carrying Russian, and permitting it would be an
    // edit to that module. Nothing can overflow this list sideways anyway —
    // every card is `min-w-0` — so no horizontal scrollbar can appear.
    expect(list).toHaveClass('overflow-x-auto');
    expect(list).not.toHaveClass('overflow-x-hidden');
  });

  it('leaves the column section itself with no overflow of its own', async () => {
    await showTheBoard();

    // The Stage 2 rule — "the board scrolls sideways; a COLUMN never does" — is
    // asserted on the section in App.accept.test.tsx, and stays true of the
    // section. Restated here as the thing this ticket must not have broken
    // while moving the scroll one element down.
    for (const status of ['col-1', 'col-2']) {
      expect(columnFor(status).className).not.toContain('overflow-');
    }
  });

  it('completes the min-h-0 chain from the board down to itself', async () => {
    await showTheBoard();
    const row = boardRow();
    const section = columnFor('col-1');
    const list = cardListIn(section);

    // Walk the real ancestors rather than trusting the two elements this file
    // happens to know about: every FLEX ancestor between the scroller and the
    // board must release `min-height: auto`, and a wrapper introduced later
    // without `min-h-0` would silently break the scroll while both endpoints
    // still looked right.
    const broken: string[] = [];
    for (let at = list.parentElement; at !== null && at !== row; at = at.parentElement) {
      if (at.className.includes('flex') && !at.className.includes('min-h-0')) {
        broken.push(`${at.tagName.toLowerCase()}.${at.className}`);
      }
    }

    expect(broken, 'a flex ancestor between the board and the card list cannot shrink').toEqual([]);
    expect(list).toHaveClass('min-h-0');
    // `flex-1` is the claim the release makes possible: the list takes the
    // height the heading did not, so the droppable padding under the last card
    // reaches the bottom of the column instead of stopping at the last card.
    expect(list).toHaveClass('flex-1');
  });

  it('does not contain the heading, so the heading cannot scroll away', async () => {
    await showTheBoard();
    const section = columnFor('col-1');
    const list = cardListIn(section);

    // D29 part 4, asserted as a SHAPE rather than as a class: a heading inside
    // the scroller scrolls with it, and a drag across five columns whose
    // headings have scrolled out of view is unnavigable.
    const heading = screen.getByRole('heading', { name: 'col-1' });
    expect(section).toContainElement(heading);
    expect(list).not.toContainElement(heading);

    // The card count rides with the heading for the same reason.
    const count = heading.parentElement;
    expect(count, 'the heading has no row of its own').not.toBeNull();
    expect(list.contains(count)).toBe(false);
  });

  it('does not refuse to shrink, in either axis (K8, D20)', async () => {
    await showTheBoard();
    const list = cardListIn(columnFor('col-1'));

    // The list now carries layout classes it did not before, and `shrink-0` is
    // the one that must never appear on anything in a column: it is what
    // starved the Russian heading down to one character per line.
    expect(list).not.toHaveClass('shrink-0');
    expect(list).toHaveClass('min-w-0');
  });
});
