// Nexus — the priority -> chip mapping, and the ONLY place it is written down.
//
// ARCHITECTURE.md section 6 fixes it, so that Stage 2 does not re-decide it:
//
//     domain.Priority   chip   token
//     1                 P0     danger
//     2                 P1     danger
//     3                 P2     warning
//     4                 (none) —
//
// Priority 4 is the unremarkable default and renders NO CHIP AT ALL — not a grey
// "P3". A card with nothing to say about its priority says nothing.
//
// # Why the label is a literal here and not a locale key
//
// "P0" is a designator, not prose: it is the same three characters in English
// and in Russian, exactly like the product name. Routing it through
// en.json/ru.json would create a key whose Russian value must equal its English
// value — the one thing locales.test.ts is written to refuse — and would have to
// be excused in that test's UNTRANSLATED_ON_PURPOSE list to stay green. The
// chip's accessible NAME is translated (`card.priority.label` interpolates this
// label into "Priority P0" / "Приоритет P0"), which is the part a reader of a
// language actually needs.
//
// `make guard` check 5 greps frontend/src for a P0/P1/P2 literal and excludes
// exactly this file — the exclusion exists because this is where it lives.
//
// # Why the token comes back as a class name
//
// The alternative is to return a tone word ('danger') and have the component
// turn it into a class. That is the mapping written a second time, in the file
// that was supposed to be reading it — and `text-${tone}` would additionally be
// a class Tailwind's content scanner cannot see, so the colour would simply not
// be in the bundle. One value, one place, and the scanner sees a literal class.

/** One priority chip: what it reads, and the token colour it reads in. */
export interface PriorityChipSpec {
  /** The chip text. Locale-independent by nature — see above. */
  label: string;
  /** Tailwind token classes. `danger` for P0/P1, `warning` for P2. */
  className: string;
}

/**
 * The mapping itself. A record keyed by the stored 1..4 rather than a switch, so
 * that it reads as the table in ARCHITECTURE.md and so that 4 is absent rather
 * than handled: the absence IS the rule.
 */
const CHIPS: Readonly<Record<number, PriorityChipSpec>> = {
  1: { label: 'P0', className: 'border-danger text-danger' },
  2: { label: 'P1', className: 'border-danger text-danger' },
  3: { label: 'P2', className: 'border-warning text-warning' },
};

/**
 * The chip for a stored priority, or null when the card shows none.
 *
 * Null for 4, and null for anything outside 1..4 — a value the domain cannot
 * currently produce, and a card that renders no chip is a better answer to it
 * than a crash.
 */
export function priorityChip(priority: number): PriorityChipSpec | null {
  return CHIPS[priority] ?? null;
}
