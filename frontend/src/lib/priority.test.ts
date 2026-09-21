import { describe, expect, it } from 'vitest';

import { priorityChip } from './priority';

// The mapping's own test, and it deliberately does not restate the three
// labels.
//
// A test that asserted the first chip's exact designator would be the mapping
// written a second time — in the file whose job is to catch a second spelling —
// and it would also be the second place in frontend/src holding that literal,
// which `make guard` check 5 refuses outright (it excludes lib/priority.ts and
// nothing else). So what is asserted here is the RULE:
//
//   - 1 and 2 are the same colour, and it is `danger`;
//   - 3 is `warning`;
//   - 4 has no chip;
//   - the three labels are distinct and all of the P<digit> shape.
//
// Those are the claims ARCHITECTURE.md section 6 actually makes. The exact three
// characters are presentation and live in one file.

const LABEL_SHAPE = /^P\d$/;

describe('the priority -> chip mapping', () => {
  it('gives priority 4 no chip at all', () => {
    // Not a grey "P3": nothing. ARCHITECTURE.md section 6.
    expect(priorityChip(4)).toBeNull();
  });

  it('colours 1 and 2 with danger', () => {
    expect(priorityChip(1)?.className).toContain('text-danger');
    expect(priorityChip(2)?.className).toContain('text-danger');
  });

  it('colours 3 with warning', () => {
    expect(priorityChip(3)?.className).toContain('text-warning');
    expect(priorityChip(3)?.className).not.toContain('danger');
  });

  it('labels the three chips distinctly', () => {
    const labels = [1, 2, 3].map((priority) => priorityChip(priority)?.label);

    expect(new Set(labels).size).toBe(3);
    for (const label of labels) {
      expect(label).toMatch(LABEL_SHAPE);
    }
  });

  it('names no colour of its own', () => {
    // Every class it hands out is a token name. A hex would also be caught by
    // `make guard` check 1; this catches the subtler version, a Tailwind colour
    // that is not one of the eleven the design handoff defines.
    for (const priority of [1, 2, 3]) {
      for (const className of priorityChip(priority)!.className.split(' ')) {
        expect(className).toMatch(/^(border|text)-(danger|warning)$/);
      }
    }
  });

  it('renders no chip for a priority the domain cannot produce', () => {
    expect(priorityChip(0)).toBeNull();
    expect(priorityChip(9)).toBeNull();
  });
});
