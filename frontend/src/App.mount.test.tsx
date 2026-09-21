import { describe, expect, it } from 'vitest';

// Nexus — the orphan check. TASKS.md, "Composition — who mounts what":
//
//   > A component that is built and not mounted is an unfinished ticket.
//
// This is the mechanical half of that rule, and it is an acceptance criterion on
// every ticket from S2-15 to S2-22. It enumerates every non-test component under
// components/ and views/, computes the import closure of main.tsx by walking
// relative import specifiers, and FAILS NAMING any module in the first set and
// not in the second.
//
// # Why it walks from main.tsx and not from this file
//
// Because "imported by something" is not the property that matters. A component
// imported only by its own test is still an orphan: it is built, it is green,
// and it is not in the application. The walk therefore starts at the real entry
// point — the file index.html loads — and follows only relative specifiers,
// which are the ones that stay inside this source tree.
//
// This test deliberately imports NO component. It reads source text through
// `import.meta.glob(..., { query: '?raw' })`, so importing it does not put a
// single module into anybody's graph; if it imported the components it is
// looking for, it would be the thing making them reachable.
//
// # What it cannot see, stated plainly
//
// It is a static import walk, so it answers "can main.tsx reach this file",
// not "does the user ever see it". A component mounted behind a condition that
// is never true passes here. That gap is why TASKS.md pairs this with a
// per-ticket REACHABILITY criterion — each mounting ticket asserts its piece
// through `render(<App ... />)`, driven by keyboard, in a test that does not
// import the component it is checking for. The two halves are not redundant.

/** Every module's source text, keyed by a path relative to src/. */
const SOURCES = import.meta.glob('./**/*.{ts,tsx}', {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>;

/** The components that must be reachable: non-test .tsx under components/ and views/. */
// Lazy on purpose: only the KEYS are wanted, and an eager glob would import
// every component into this file — which is the one thing this test must not
// do, because it would make them reachable from the test itself.
const MOUNTABLE = Object.keys(
  import.meta.glob(['./components/**/*.tsx', './views/**/*.tsx', '!./**/*.test.tsx']),
).sort();

const ENTRY = './main.tsx';

/** The relative specifiers `source` imports or re-exports. */
function relativeSpecifiers(source: string): string[] {
  // `from '...'`, a bare side-effect `import '...'`, and `import('...')`. Good
  // enough because it is checked against reality: a specifier this misses shows
  // up as a component reported unreachable, which is a loud failure, never a
  // quiet pass.
  const pattern = /(?:from|import)\s*\(?\s*['"](\.[^'"]*)['"]/g;

  return [...source.matchAll(pattern)].map((match) => match[1]);
}

/** Resolves a relative specifier against the importing module's path. */
function resolve(fromPath: string, specifier: string): string | null {
  const base = new URL(specifier, `file:///${fromPath.slice(2)}`).pathname.slice(1);

  for (const candidate of [base, `${base}.ts`, `${base}.tsx`, `${base}/index.ts`]) {
    const key = `./${candidate}`;
    if (key in SOURCES) {
      return key;
    }
  }
  return null;
}

/** Every module reachable from the entry point by relative imports. */
function importClosure(): Set<string> {
  const seen = new Set<string>();
  const queue = [ENTRY];

  while (queue.length > 0) {
    const path = queue.pop()!;
    if (seen.has(path)) {
      continue;
    }
    seen.add(path);

    for (const specifier of relativeSpecifiers(SOURCES[path])) {
      const next = resolve(path, specifier);
      if (next !== null) {
        queue.push(next);
      }
    }
  }

  return seen;
}

describe('every component is mounted', () => {
  it('can actually see the source tree', () => {
    // The vacuous pass this check could otherwise have: a glob that matched
    // nothing would report no orphans for ever. Both sets must be non-empty and
    // the entry point must exist.
    expect(SOURCES[ENTRY], 'main.tsx was not found by the glob').toBeTypeOf('string');
    expect(MOUNTABLE.length, 'no components were enumerated').toBeGreaterThan(0);
  });

  it('reaches the entry point and its first hop', () => {
    // A second guard against a silently broken walk: if `relativeSpecifiers` or
    // `resolve` stopped working, the closure would collapse to {main.tsx} and
    // every component would be reported — which is loud. This catches the
    // subtler version where the closure is empty or missing App itself.
    const closure = importClosure();

    expect(closure.has(ENTRY)).toBe(true);
    expect(closure.has('./App.tsx')).toBe(true);
    expect(closure.size).toBeGreaterThan(2);
  });

  it('leaves nothing under components/ or views/ unreachable from main.tsx', () => {
    const closure = importClosure();

    const orphans = MOUNTABLE.filter((path) => !closure.has(path));

    expect(
      orphans,
      'built and mounted by nobody — see TASKS.md "Composition — who mounts what"',
    ).toEqual([]);
  });
});
