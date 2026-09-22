import { Bug, FileText, FolderKanban, Repeat, SquareCheck, type LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

// Nexus — the icon that says what kind of row a card is.
//
// # This table is presentation, not a rule
//
// Which types exist, which of them have a Kanban column, which may be a parent
// and which count as work are all domain.NodeType's answers and are never asked
// here. What is here is the only thing Go cannot answer, because it is not a
// value: which glyph and which token colour a type is DRAWN with. That is the
// same category as the column heading labels in locales/ — a label table is
// presentation; a type list used to decide something would be a rule, and there
// is no decision in this file.
//
// An unrecognised type therefore renders the neutral icon instead of throwing.
// The frontend does not hold the list of types, so it must not fall over when
// Go sends one it has not seen.
//
// # The two colours that are specified, and the one that is not
//
// design/README.md assigns exactly two icon colours: `success` for the task icon
// and `danger` for the bug icon. It says nothing about project, habit or note,
// and D6-as-amended forbids inventing a spec and attributing it to the export —
// so the rest take the neutral `muted`, which is the absence of a semantic
// rather than a new one. If the handoff ever names them, this table is the one
// place to change.
//
// Colour reaches the SVG through `currentColor`: lucide draws with it, so the
// icon inherits whichever token class this component carries and no hex ever
// enters frontend/src.

interface TypeGlyph {
  Icon: LucideIcon;
  /** A Tailwind token class. `currentColor` in the SVG picks it up. */
  className: string;
}

const NEUTRAL: TypeGlyph = { Icon: FileText, className: 'text-muted' };

const GLYPHS: Readonly<Record<string, TypeGlyph>> = {
  task: { Icon: SquareCheck, className: 'text-success' },
  bug: { Icon: Bug, className: 'text-danger' },
  project: { Icon: FolderKanban, className: 'text-muted' },
  habit: { Icon: Repeat, className: 'text-muted' },
  note: NEUTRAL,
};

export interface TypeIconProps {
  /** `node.type`, the stored string, exactly as Go sent it. */
  type: string;
}

export function TypeIcon({ type }: TypeIconProps) {
  const { t } = useTranslation();

  const glyph = GLYPHS[type] ?? NEUTRAL;

  return (
    <glyph.Icon
      role="img"
      // The accessible name is translated and falls back to a generic word for
      // a type this build has no name for, rather than reading the raw enum out
      // loud. i18next takes the first key that resolves.
      aria-label={t([`card.type.${type}`, 'card.type.unknown'])}
      // `shrink-0` SURVIVES here, and D20 says why: this is a FIXED-SIZE,
      // NON-TEXT box. Its `h-4 w-4` is a decision about the 8px grid, not
      // about a string, so its width cannot change with the locale and holding
      // it is holding nothing hostage. That is the whole distinction: a box
      // whose size depends on WORDS may not refuse to shrink; a box whose size
      // is a number may.
      // 8px grid: 16px box, which is two steps.
      className={`h-4 w-4 shrink-0 ${glyph.className}`}
    />
  );
}

export default TypeIcon;
