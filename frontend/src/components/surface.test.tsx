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
      <ToastList toasts={[{ id: 1, messageKey: 'toast.error.body' }]} onDismiss={() => {}} />,
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
    const sources = import.meta.glob('./*.tsx', { query: '?raw', import: 'default', eager: true });

    for (const [path, source] of Object.entries(sources)) {
      if (path.endsWith('.test.tsx')) {
        continue;
      }
      expect(String(source), `${path} nudges the compositor`).not.toMatch(
        /requestAnimationFrame|offsetHeight|getBoundingClientRect/,
      );
    }

    expect(Object.keys(sources).length, 'nothing was scanned').toBeGreaterThan(5);
  });
});
