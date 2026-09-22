import { readFileSync } from 'node:fs';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { screen } from '@testing-library/react';

import App from './App';
import { createAppStore } from './store';
import { board, createFakeClient } from './test/fakeClient';
import { renderIn } from './test/render';

// Nexus — the height chain (S3-01, K10, D22).
//
// # WHAT THIS FILE CAN AND CANNOT PROVE — read before trusting it
//
// jsdom has NO LAYOUT ENGINE. Every `offsetHeight` is 0, nothing overflows,
// nothing scrolls and no stylesheet is applied to the rendered tree (the vitest
// config sets `css: false`). So this file **cannot** show that the document
// stopped scrolling or that the board scrolls instead. That is S3-09 item 1, by
// eye, on real hardware.
//
// What it CAN do is assert that every link of the chain is still written down,
// because the defect it guards against is a link going missing: the chain broke
// in Stage 2 precisely because `#root` was never given a rule and nobody
// noticed. Each assertion below fails if one link is removed, which is the
// drift property that matters between now and the hand pass.
//
// The stylesheet is read as TEXT rather than applied, for the same reason:
// `css: false` means jsdom never sees it, so the only honest way to assert a
// rule exists is to read the file that declares it.

/**
 * Two sources, read as TEXT.
 *
 * Paths are relative to the package root, which is where npm runs a script from
 * and therefore where `make front-test` puts us. Why `node:fs` at all, and why
 * neither `?raw` nor `import.meta.url` can do this job, is written down once in
 * src/test/node-builtins.d.ts — both of those routes fail SILENTLY, which is the
 * dangerous kind.
 */
function sourceOf(relative: string): string {
  const text = readFileSync(relative, 'utf8');
  // Non-vacuity, in one place: a read that came back empty would make every
  // assertion below pass while checking nothing.
  expect(text.length, `${relative} was read as empty`).toBeGreaterThan(0);
  return text;
}

const STYLE_CSS = sourceOf('src/style.css');
const KANBAN_SOURCE = sourceOf('src/views/Kanban.tsx');

/**
 * The declarations a selector carries, gathered from every rule that names it.
 *
 * A deliberately small reader rather than a CSS parser dependency: it strips
 * comments, splits on braces, and matches a selector as one entry of a
 * comma-separated list. That is enough for "does this selector declare a
 * height", and anything cleverer would be a parser this project has to own.
 */
function declarationsFor(css: string, selector: string): string[] {
  const withoutComments = css.replace(/\/\*[\s\S]*?\*\//g, '');
  const found: string[] = [];

  for (const [, selectors, body] of withoutComments.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
    const names = selectors.split(',').map((name) => name.trim());
    if (names.includes(selector)) {
      found.push(...body.split(';').map((declaration) => declaration.trim().replace(/\s+/g, ' ')));
    }
  }

  return found.filter((declaration) => declaration !== '');
}

function testWindow(): Window {
  return {
    document,
    matchMedia: (query: string) => ({ matches: false, media: query }),
  } as unknown as Window;
}

async function renderShell() {
  const fake = createFakeClient({ board: board(['col-1', 'col-2']) });
  const store = createAppStore(fake.client, { view: testWindow() });
  const rendered = await renderIn('en', <App store={store} />);
  await screen.findByRole('region', { name: 'col-1' });
  return rendered;
}

beforeEach(() => {
  vi.spyOn(console, 'error').mockImplementation(() => {});
});

describe('the height chain', () => {
  // The three links that live in CSS. Each is asserted on its own so that a
  // failure names the one that went missing, which is the whole point: `#root`
  // is the link that was never there.
  it.each(['html', 'body', '#root'])('gives %s a definite height', (selector) => {
    expect(declarationsFor(STYLE_CSS, selector)).toContain('height: 100%');
  });

  it('declares all three in one rule, so they cannot drift apart', () => {
    // Three separate rules would pass the test above and still be three places
    // to edit. D22 is one decision and it gets one spelling.
    const withoutComments = STYLE_CSS.replace(/\/\*[\s\S]*?\*\//g, '');
    const groups: string[][] = [];

    for (const [, selectors, body] of withoutComments.matchAll(/([^{}]+)\{([^{}]*)\}/g)) {
      if (/height:\s*100%/.test(body)) {
        groups.push(selectors.split(',').map((name) => name.trim()));
      }
    }

    expect(groups).toHaveLength(1);
    expect(groups[0]).toEqual(expect.arrayContaining(['html', 'body', '#root']));
  });

  it('no longer floors the body at the viewport instead of sizing it', () => {
    // `min-height: 100vh` was the broken link: a floor the shell could grow
    // past. Keeping it beside `height: 100%` would be two answers to one
    // question.
    expect(declarationsFor(STYLE_CSS, 'body')).not.toContain('min-height: 100vh');
  });

  it('makes the shell fill that height rather than merely exceed it', async () => {
    const { container } = await renderShell();
    const shell = container.firstElementChild;

    expect(shell).toHaveClass('h-full');
    // `min-h-screen` is the floor this ticket replaced. If it comes back, the
    // shell can grow past the viewport again and the document scrolls again.
    expect(shell).not.toHaveClass('min-h-screen');
  });

  it('lets <main> shrink below its content, and drops the class that could not', async () => {
    await renderShell();
    const main = screen.getByRole('main');

    // `min-h-0` releases `min-height: auto` on a COLUMN flex item, which is the
    // axis the shell actually flexes in.
    expect(main).toHaveClass('min-h-0');
    // `min-w-0` released the cross axis, whose minimum was already 0. It did
    // nothing while looking load-bearing — D22 names it as the no-op.
    expect(main).not.toHaveClass('min-w-0');
  });

  it('hands the scroll to the board, in both axes, explicitly', async () => {
    const { container } = await renderShell();
    const boardElement = container.querySelector('[data-column]')?.parentElement;

    expect(boardElement, 'the board was not found').not.toBeNull();
    // Both axes NAMED. `overflow-x-auto` alone is not "horizontal only": CSS
    // Overflow 3 promotes a `visible` companion axis to `auto`, which is how
    // the board came to scroll vertically behind a comment saying it did not.
    expect(boardElement).toHaveClass('overflow-x-auto');
    expect(boardElement).toHaveClass('overflow-y-auto');
    // And a definite height to scroll inside, or neither axis can ever overflow.
    expect(boardElement).toHaveClass('h-full');
  });

  it('keeps the killed scrollLeft theory out of the source', () => {
    // D22 investigated and ruled out the "latched scrollLeft" explanation. A
    // future ticket adding scroll-offset handling here would be re-litigating a
    // closed question in code.
    // A property ACCESS — `.scrollLeft` / `.scrollTop`. The prose above the
    // board names both words on purpose, to record that the theory was killed,
    // and a test that banned the words would ban the record of the ruling.
    expect(KANBAN_SOURCE).not.toMatch(/\.scroll(Left|Top)\b/);
  });
});
