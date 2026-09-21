import { describe, expect, it } from 'vitest';
import { render } from '@testing-library/react';

import App from './App';

// The runner is the deliverable of S2-10, not this test. It is a smoke test and
// it is deliberately the smallest thing that proves the whole chain works:
// vitest picks up src/**/*.test.tsx, the react plugin compiles TSX, jsdom
// supplies a document, @testing-library mounts into it, and the jest-dom
// matcher registered by src/test/setup.ts is both callable and typed.
//
// It also proves the two gates see this file: it is linted by gate 3 and
// type-checked by gate 4 like any other source file, because it lives beside
// its subject rather than in a parallel tree with its own tsconfig.
describe('App', () => {
  it('mounts the page shell', () => {
    const { container } = render(<App />);

    const shell = container.firstElementChild;

    expect(shell).toBeInTheDocument();
    // The shell carries the page background, and it carries it through the
    // design tokens. Asserting the token class rather than a colour is the
    // point: the resolved value is design/tokens.css's business and changes
    // with the palette, so a test that knew it would be a second place the
    // value is written down.
    expect(shell).toHaveClass('bg-bg');
  });
});
