import { useTranslation } from 'react-i18next';

import type { Tag } from '../lib/client';

// Nexus — the card's tags.
//
// The service resolves them (dto.go: "Tags are the node's labels, already
// resolved; the UI never joins"), so this renders a list and nothing else.
//
// # Why domain.Tag.Color is deliberately NOT painted yet
//
// `color` is a free-form string on the tags table. Nothing validates it against
// the eleven token names, so painting with it would mean one of two things:
// interpolating it into a class (`bg-${tag.color}`), which Tailwind's content
// scanner cannot see and which would therefore emit no CSS at all; or writing it
// into an inline style, which is a colour arriving from data instead of from
// design/tokens.css and is how a hex gets back into the UI through the side
// door. Until Go constrains the value to a token name, a tag reads in `muted` on
// a `line` outline like every other quiet chip on the card. That is a gap, and
// it is a deliberate one.
//
// Long tags and long Russian words wrap rather than clip: the list wraps, every
// chip may shrink, and nothing on this card is given a fixed width.

export interface TagListProps {
  tags: Tag[];
}

export function TagList({ tags }: TagListProps) {
  const { t } = useTranslation();

  if (tags.length === 0) {
    return null;
  }

  return (
    <ul aria-label={t('card.tags.label')} className="flex min-w-0 flex-wrap gap-1">
      {tags.map((tag) => (
        <li
          key={tag.id}
          className="max-w-full break-words rounded-sm border border-line px-1.5 text-muted"
        >
          {tag.name}
        </li>
      ))}
    </ul>
  );
}

export default TagList;
