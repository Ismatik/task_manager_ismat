// Nexus — vitest setup, loaded before every test file.
//
// The jest-dom entry point is the `/vitest` one: it registers the matchers on
// vitest's `expect` AND augments vitest's Assertion type, which is what makes
// `toBeInTheDocument()` type-check under gate 4. The plain
// '@testing-library/jest-dom' entry point registers against Jest's global
// expect and would compile but never run here.
import '@testing-library/jest-dom/vitest';

import { afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';

// Unmount everything between tests. Without this, a component from a previous
// test is still in document.body and getByRole finds two of whatever the next
// test is looking for — a failure mode that only appears once there is more
// than one test, i.e. after the person who wrote the first one has moved on.
afterEach(() => {
  cleanup();
});
