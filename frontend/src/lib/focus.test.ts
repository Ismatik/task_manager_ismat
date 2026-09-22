import { afterEach, describe, expect, it, vi } from 'vitest';

import { focusWithoutScrolling } from './focus';

// Nexus — the focus helper, asserted through a real element.
//
// The assertion that matters is the ARGUMENT: `{ preventScroll: true }` reaches
// the DOM on every call. It is made with a spy on a real HTMLElement's own
// `focus`, not by reading this repository's source for the string — a source
// grep would pass on a file that never ran, and `make guard` check 7 already
// covers the grep half of the claim from a different angle.
//
// What this file CANNOT assert: that focusing stops scrolling. jsdom has no
// layout engine, so nothing in it ever scrolls and `preventScroll` is inert
// there. That half is S3-09 item 1.

const created: HTMLElement[] = [];

/** A real, focusable element in the document — not a stub object. */
function focusable(): HTMLElement {
  const element = document.createElement('button');
  document.body.append(element);
  created.push(element);
  return element;
}

afterEach(() => {
  for (const element of created.splice(0)) {
    element.remove();
  }
  vi.restoreAllMocks();
});

describe('focusWithoutScrolling', () => {
  it('asks the DOM not to scroll', () => {
    const element = focusable();
    const spy = vi.spyOn(element, 'focus');

    focusWithoutScrolling(element);

    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy).toHaveBeenCalledWith({ preventScroll: true });
  });

  it('really does move focus', () => {
    // Without this, the test above would pass over a helper that called a spy
    // and nothing else. The point of the helper is still to focus.
    const element = focusable();

    focusWithoutScrolling(element);

    expect(document.activeElement).toBe(element);
  });

  it('does nothing when there is nothing to focus', () => {
    // Every call site is a ref, an opener or a query result, and each of those
    // is legitimately absent sometimes. Throwing there would turn "the card had
    // already unmounted" into a crash.
    const element = focusable();
    element.focus();

    expect(() => focusWithoutScrolling(null)).not.toThrow();
    expect(() => focusWithoutScrolling(undefined)).not.toThrow();
    expect(document.activeElement).toBe(element);
  });
});
