import { describe, expect, it } from 'vitest';
import { screen } from '@testing-library/react';

import { Card } from './Card';
import { HabitChip } from './HabitChip';
import { ToastList } from './Toast';
import { habitView, node, nodeView } from '../test/fakeClient';
import { renderIn } from '../test/render';

// Nexus — the three surfaces D19 took the backdrop filter OFF, and the one
// property that had to survive it (S3-04, K7).
//
// # What this file is for
//
// D19 removes a compositing layer from the card, the habit chip and the toast.
// It does NOT remove their translucency: under Aurora the surface tokens are
// what let the background through, and the filter only blurred what came
// through. A fix that quietly took the token as well would look identical in
// jsdom and wrong on screen, so the token is asserted here, by name, on all
// three.
//
// The OTHER half of the policy — that the filter survives on the column and the
// two overlay scrims and appears nowhere else — belongs to `make guard` check 8
// and is deliberately not restated here. Check 8 is an exact grep over the whole
// of frontend/src with NO file excluded, which is strictly stronger than
// anything this file could assert; and a test that spelt the class out would
// have had to be excluded from it, weakening the grep to make room for a
// weaker copy of itself. Same reasoning as the keys-only ACCEPT test under
// check 6.
//
// # WHAT NEITHER HALF PROVES
//
// That the resize bug is fixed. There is no display on this machine, jsdom
// composites nothing, and the failure is a WebKitGTK compositing-layer
// staleness that only appears when a real window is dragged. That is S3-09
// item 4 — resize repeatedly, at more than one display scaling, in both
// palettes — and D19's escalation ladder applies if it still reproduces.

/**
 * WHERE THE COMPONENTS ARE, as a glob, and it is one line with two deliberate
 * halves (S3-35, K17, D34).
 *
 * `**` because the sweep below must stay exhaustive as `src/components/` grows.
 * The directory is flat today — 18 files, 3 of them `*.test.tsx` — so the old
 * `./*.tsx` matched every one of them and was green, which is the shape of
 * defect this project keeps paying for: a check that is correct by accident.
 * Block B adds the detail slide-over, the field editors, inline subtasks, the
 * recurrence editor, attachments, the editable time log, the tree, search and
 * archive (S3-20 … S3-27), and the day one of them lands in
 * `src/components/detail/` the old pattern would have missed it in SILENCE.
 *
 * That silence is worse here than anywhere else in the suite. What the sweep
 * enforces is D19's ban on a `requestAnimationFrame` / `offsetHeight` /
 * `getBoundingClientRect` resize workaround, so a component the glob misses is
 * a resize workaround NOTHING in this repository forbids — and the resize
 * defect is K7, the one the user actually reported and the one this machine
 * cannot verify at all (E4: no window manager, the window cannot be resized).
 * A slide-over panel is precisely where a developer reaches for
 * `getBoundingClientRect`.
 *
 * The negative pattern excludes this directory's own tests, in the form
 * `App.mount.test.tsx:47` uses and `Toast.test.tsx` uses since S3-34, so there
 * is one spelling of "a source file but not a test file" in this repository
 * rather than two. It replaces the `path.endsWith('.test.tsx')` filter that
 * used to sit inside the loop: same three files excluded, one rule, one place.
 *
 * `COMPONENT_GLOB` beside it is PROSE, for the failure messages, and it exists
 * only because Vite resolves `import.meta.glob` at build time and therefore
 * requires its patterns as literals — a constant cannot be interpolated into
 * the call. So the two cannot quietly disagree, `sweepTheComponents` asserts
 * the property that sentence claims — no match is a `*.test.tsx` — against what
 * the glob actually returned, rather than leaving the description to be
 * believed.
 */
const COMPONENT_GLOB = './**/*.tsx (excluding ./**/*.test.tsx)';

const COMPONENT_MODULES = import.meta.glob(['./**/*.tsx', '!./**/*.test.tsx'], {
  query: '?raw',
  import: 'default',
  eager: true,
}) as Record<string, string>;

/**
 * Every component's source text, keyed by a path relative to `src/components/`,
 * and red rather than empty if it swept nothing (D32).
 *
 * The scan below is a loop over this map: over the empty map it performs zero
 * assertions and passes, and over a partial map it passes on everything it did
 * not see. The floor is what stops the first and the recursive glob is what
 * stops the second, and both belong HERE, at the derivation — the same
 * module-scope shape `Toast.test.tsx`'s `sweepTheStore` uses since S3-34.
 *
 * THREE assertions, not one, because they are three different failures and read
 * differently. No match at all means `src/components/` moved or the pattern is
 * wrong. Too few matches means the pattern is under-matching a directory that
 * has been well past the floor since Stage 1. A `*.test.tsx` in the result means
 * the negative half of the pattern stopped working, and the sweep would then be
 * scanning tests that are allowed to name these APIs.
 */
function sweepTheComponents(): Record<string, string> {
  const files = Object.keys(COMPONENT_MODULES).sort();

  expect(
    files.length,
    `the sweep of ${COMPONENT_GLOB} matched no files at all: src/components/ moved, or the pattern is wrong`,
  ).toBeGreaterThan(0);

  expect(
    files.length,
    `the sweep of ${COMPONENT_GLOB} scanned only ${files.length} file(s) — src/components/ has been well past that floor since Stage 1, so the pattern is under-matching: ${files.join(' ')}`,
  ).toBeGreaterThan(5);

  const tests = files.filter((path) => path.endsWith('.test.tsx'));
  expect(
    tests,
    `the sweep of ${COMPONENT_GLOB} read this directory's own tests, which are not components: ${tests.join(' ')}`,
  ).toEqual([]);

  return COMPONENT_MODULES;
}

const COMPONENT_SOURCES = sweepTheComponents();

describe('the surfaces that lost the backdrop filter (D19)', () => {
  it('leaves the card translucent and unpromoted', async () => {
    await renderIn('en', <Card view={nodeView({ node: node({ title: 'A card' }) })} />);
    const card = screen.getByRole('article');

    expect(card, 'the card stopped being a surface').toHaveClass('bg-surface');
    expect(card, 'Studio gets its edge from the shadow').toHaveClass('shadow-sm');
  });

  it('leaves the habit chip translucent and unpromoted', async () => {
    await renderIn(
      'en',
      <HabitChip habit={habitView({ node: node({ id: 'h-1', title: 'Read' }) })} tabIndex={0} onToggle={() => {}} />,
    );
    const chip = screen.getByRole('checkbox');

    expect(chip, 'the chip stopped being a surface').toHaveClass('bg-surface');
    expect(chip, 'Studio gets its edge from the shadow').toHaveClass('shadow-sm');
  });

  it('leaves the toast translucent and unpromoted', async () => {
    await renderIn(
      'en',
      <ToastList toasts={[
          {
            id: 1,
            operationKey: 'toast.operation.move',
            messageKey: 'toast.error.body',
            kind: 'failure',
            count: 1,
          },
        ]} onDismiss={() => {}} />,
    );
    // The panel, not the positioning wrapper: the wrapper never had a surface.
    const panel = screen.getByRole('alert').firstElementChild;

    expect(panel, 'the toast stopped being a surface').toHaveClass('bg-elevated');
    expect(panel, 'Studio gets its edge from the shadow').toHaveClass('shadow-sm');
  });

  it('did not trade the filter for a resize workaround', async () => {
    // D19 rejects them by name: no requestAnimationFrame nudge, no forced
    // reflow, no transform toggle. Each is a workaround for a symptom whose
    // cause has already been identified, and each would be a second rule about
    // resizing that nothing tests. Asserted here because the temptation arrives
    // exactly when S3-09's hand pass comes back ambiguous.
    //
    // Recursively, and over components only: see COMPONENT_GLOB above, which
    // also carries the floor this scan used to end with. A component in a
    // subdirectory that this loop never reads is a resize workaround nothing
    // forbids (S3-35, K17, D34).
    for (const [path, source] of Object.entries(COMPONENT_SOURCES)) {
      expect(String(source), `${path} nudges the compositor`).not.toMatch(
        /requestAnimationFrame|offsetHeight|getBoundingClientRect/,
      );
    }
  });
});
