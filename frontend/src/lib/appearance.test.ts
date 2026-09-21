import { afterEach, describe, expect, it, vi } from 'vitest';

import {
  applyAppearance,
  auroraDriftEnabled,
  prefersReducedMotion,
  revealAfterBoot,
} from './appearance';

// jsdom implements no media queries at all, so every case that cares about
// prefers-reduced-motion installs its own matchMedia. Returning a window-shaped
// object rather than patching the global keeps the two branches independent and
// means no test can leak a media preference into the next one.
function windowWith(reducedMotion: boolean): Window {
  return {
    document,
    matchMedia: vi.fn((query: string) => ({
      matches: query.includes('prefers-reduced-motion') && reducedMotion,
      media: query,
    })),
  } as unknown as Window;
}

/** A window that, like jsdom's, has no matchMedia at all. */
function windowWithoutMediaQueries(): Window {
  return { document } as unknown as Window;
}

afterEach(() => {
  const root = document.documentElement;
  root.removeAttribute('data-palette');
  root.removeAttribute('data-drift');
  root.removeAttribute('data-appearance');
  root.removeAttribute('style');
  root.classList.remove('dark');
});

describe('applyAppearance', () => {
  // All four palette x theme combinations. The assertions are on the DOM
  // attributes and not on colours: which colour "aurora dark" is belongs to
  // design/tokens.css, and a test that knew it would be a second place the
  // value is written down.
  it.each([
    ['aurora', 'dark', true],
    ['aurora', 'light', false],
    ['studio', 'dark', true],
    ['studio', 'light', false],
  ])('puts %s / %s on the document', (palette, theme, expectDark) => {
    applyAppearance({ palette, theme, accent: '' }, windowWith(false));

    const root = document.documentElement;
    expect(root.dataset.palette).toBe(palette);
    expect(root.classList.contains('dark')).toBe(expectDark);
  });

  it('sets the inline accent override when the accent is non-empty', () => {
    applyAppearance({ palette: 'aurora', theme: 'dark', accent: 'oklch(70% 0.2 200)' }, windowWith(false));

    expect(document.documentElement.style.getPropertyValue('--accent')).toBe('oklch(70% 0.2 200)');
  });

  it('REMOVES the accent property when the accent is empty, rather than emptying it', () => {
    const root = document.documentElement;

    applyAppearance({ palette: 'aurora', theme: 'dark', accent: 'rebeccapurple' }, windowWith(false));
    expect(root.style.getPropertyValue('--accent')).toBe('rebeccapurple');

    applyAppearance({ palette: 'aurora', theme: 'dark', accent: '' }, windowWith(false));

    // The property must be GONE from the style attribute, not present and
    // empty (D6: "" means "use the palette's own").
    //
    // Honest note on what this case can and cannot catch: CSSOM defines
    // setProperty(p, "") as removeProperty(p), so rewriting the branch that way
    // is NOT a failure this can see — it is the same operation. What it does
    // catch, and was watched to catch, is the bug that actually gets written:
    // an `if (accent) setProperty(...)` with no else, which leaves the previous
    // accent in place and makes clearing a no-op.
    expect(root.style.getPropertyValue('--accent')).toBe('');
    expect(root.getAttribute('style') ?? '').not.toContain('--accent');
  });

  it('lifts the boot gate so the document paints', () => {
    document.documentElement.dataset.appearance = 'pending';

    applyAppearance({ palette: 'studio', theme: 'light', accent: '' }, windowWith(false));

    expect(document.documentElement.dataset.appearance).toBeUndefined();
  });
});

describe('the Aurora drift', () => {
  it('runs on aurora when motion is allowed', () => {
    expect(auroraDriftEnabled('aurora', windowWith(false))).toBe(true);

    applyAppearance({ palette: 'aurora', theme: 'dark', accent: '' }, windowWith(false));
    expect(document.documentElement.dataset.drift).toBe('on');
  });

  it('is NOT running under prefers-reduced-motion: reduce', () => {
    expect(auroraDriftEnabled('aurora', windowWith(true))).toBe(false);

    applyAppearance({ palette: 'aurora', theme: 'dark', accent: '' }, windowWith(true));

    // The gate is an attribute rather than a CSS-only rule on purpose.
    // tokens.css already kills every CSS animation under the media query, but
    // it cannot see a drift driven from JavaScript, and design/README.md asks
    // for the drift to pause and not merely to be un-animated.
    expect(document.documentElement.dataset.drift).toBe('off');
  });

  it('does not run on studio, which has no drift', () => {
    expect(auroraDriftEnabled('studio', windowWith(false))).toBe(false);

    applyAppearance({ palette: 'studio', theme: 'light', accent: '' }, windowWith(false));
    expect(document.documentElement.dataset.drift).toBe('off');
  });
});

describe('prefersReducedMotion', () => {
  it('reads the media query', () => {
    expect(prefersReducedMotion(windowWith(true))).toBe(true);
    expect(prefersReducedMotion(windowWith(false))).toBe(false);
  });

  it('answers false where matchMedia does not exist, rather than throwing', () => {
    expect(prefersReducedMotion(windowWithoutMediaQueries())).toBe(false);
  });
});

describe('revealAfterBoot', () => {
  it('is idempotent', () => {
    document.documentElement.dataset.appearance = 'pending';

    revealAfterBoot(document);
    revealAfterBoot(document);

    expect(document.documentElement.dataset.appearance).toBeUndefined();
  });
});
