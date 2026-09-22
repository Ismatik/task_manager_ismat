// Nexus — "does anything on this screen refuse to shrink?", asked of the
// rendered DOM rather than of a list of components.
//
// # Why this is not another word list (D20, and it is D17 arriving twice)
//
// Card.tsx used to state in a comment that "nothing on this card has a fixed
// width and nothing is `whitespace-nowrap`", and App.accept.test.tsx asserted
// the absence of `truncate` and `whitespace-nowrap`. **Both were true, and the
// card clipped anyway**, because `shrink-0` on a max-content localised string
// clips identically and was not on the list. Adding `shrink-0` to the list
// would have fixed that bug and left the next one: a mechanism-based audit is
// only as good as its enumeration of mechanisms, and an enumeration is a guess
// about the future.
//
// So this inverts the question. Instead of asking "is one of the four bad
// utilities present", it asks, of every element that CARRIES TEXT:
//
//     of the classes you are wearing that decide whether this box may become
//     narrower than its own text, is every single one on the permitted list?
//
// Anything else — including a utility nobody has used yet — is a refusal, and
// is reported on the day it is first used. Permitting a new one is then a
// deliberate edit to the list below, made by whoever needs it, in a diff a
// reviewer reads.
//
// # What it still cannot do
//
// jsdom has NO LAYOUT ENGINE. Every offsetWidth is 0 and nothing ever overflows,
// so this asserts MECHANISMS and never the absence of clipping. "Russian does
// not clip at a real 1024x768" is S3-09 item 2, by eye, and the onset widths in
// K8 are +/-5% arithmetic rather than measurements.

/**
 * The class families that decide whether a box may give ground, and whether its
 * text may wrap when it does.
 *
 * A token matching one of these is a decision about shrinking — so it must be
 * justified. A token matching none of them (a colour, a radius, a padding) is
 * none of this module's business.
 */
const DECIDES_SHRINKING = [
  /^shrink(-.*)?$/,
  /^basis-.*$/,
  /^flex(-.*)?$/,
  /^whitespace-.*$/,
  /^break-.*$/,
  /^truncate$/,
  /^text-(ellipsis|clip)$/,
];

/**
 * The class families that pin a box to a width.
 *
 * Applied ONLY to an element carrying text of its own. A container legitimately
 * caps its own width against the viewport — the toast's `w-[min(24rem,...)]` and
 * the two overlay panels' `w-[min(32rem,100%)]` are exactly that, and they are
 * caps rather than floors. On an element holding a localised string, a width is
 * a promise about how long that string is, which is a promise no locale keeps.
 */
const DECIDES_WIDTH = [/^w-.*$/, /^min-w-.*$/, /^max-w-.*$/];

/**
 * The tokens a text-carrying element is allowed to wear, each with its reason.
 *
 * Short on purpose. A layout that needs something not here is a layout making a
 * new claim about how a localised string behaves, and D20 wants that claim
 * written down rather than assumed.
 */
const PERMITTED = new Map<string, string>([
  ['flex', 'display only — says nothing about this box shrinking'],
  ['inline-flex', 'display only'],
  ['flex-col', 'direction only'],
  ['flex-row', 'direction only'],
  ['flex-wrap', 'the opposite of a refusal: children move to a new line'],
  ['flex-1', 'grow and shrink from a zero basis — the shrinkable default'],
  ['flex-auto', 'grow and shrink'],
  ['shrink', 'flex-shrink: 1, which is the permission itself'],
  ['basis-0', 'a zero basis, so the box is sized by the flex line and not by its text'],
  ['min-w-0', 'releases min-width:auto — the thing that lets a flex item go below min-content'],
  ['w-full', 'tracks the parent, not the content'],
  ['max-w-full', 'a ceiling, never a floor'],
  ['break-words', 'lets an unbreakable token wrap rather than overflow'],
  ['break-normal', 'the initial value'],
  ['whitespace-normal', 'the initial value — text may wrap'],
  ['whitespace-pre-wrap', 'preserves runs of spaces and still wraps'],
]);

/** One element that will not shrink, described well enough to find it. */
export interface ShrinkRefusal {
  /** A CSS-ish path, e.g. `span.shrink-0.font-mono`. */
  where: string;
  /** The text it carries, trimmed and capped. */
  text: string;
  /** The offending class tokens. */
  tokens: string[];
}

/** The element's own text — not its descendants'. A container carries nothing. */
function ownText(element: Element): string {
  let text = '';
  for (const child of element.childNodes) {
    if (child.nodeType === Node.TEXT_NODE) {
      text += child.nodeValue ?? '';
    }
  }
  return text.trim();
}

/**
 * `getAttribute` and not `.className`: on an SVG element — every icon is one —
 * `className` is an SVGAnimatedString and has no `split`.
 */
function classTokens(element: Element): string[] {
  return (element.getAttribute('class') ?? '').split(/\s+/).filter((token) => token !== '');
}

/**
 * Every element under `root` that has text in it and wears a class forbidding
 * it to shrink.
 *
 * Two tiers, because two kinds of element are at risk in different ways:
 *
 *   1. An element CARRYING text — a direct text node with a non-whitespace
 *      character. Its width can depend on the language, so it is judged on both
 *      families: it may neither refuse to shrink nor be pinned to a width.
 *   2. An element merely CONTAINING text further down. Its width also follows
 *      the language, through its children — the habit chip is one, and it held
 *      `shrink-0` around a Russian habit title until S3-02. So it is judged on
 *      the shrink family, but NOT on widths: capping a panel against the
 *      viewport is a legitimate thing for a container to do.
 *
 * An element with no text anywhere below it is out of scope entirely. That is
 * the icon case, and D20 is explicit that a fixed-size non-text box keeps
 * `shrink-0` because its width is a number rather than a string.
 */
export function shrinkRefusals(root: ParentNode): ShrinkRefusal[] {
  const refusals: ShrinkRefusal[] = [];

  for (const element of root.querySelectorAll('*')) {
    const own = ownText(element);
    const text = own === '' ? (element.textContent ?? '').trim() : own;
    if (text === '') {
      continue;
    }

    const families = own === '' ? DECIDES_SHRINKING : [...DECIDES_SHRINKING, ...DECIDES_WIDTH];
    const tokens = classTokens(element).filter(
      (token) => families.some((family) => family.test(token)) && !PERMITTED.has(token),
    );

    if (tokens.length > 0) {
      refusals.push({
        where: `${element.tagName.toLowerCase()}.${classTokens(element).join('.')}`,
        text: text.length > 40 ? `${text.slice(0, 40)}...` : text,
        tokens,
      });
    }
  }

  return refusals;
}

/** The refusals as lines, for an assertion message that says where to look. */
export function describeRefusals(refusals: ShrinkRefusal[]): string[] {
  return refusals.map(
    (refusal) => `${refusal.tokens.join(' ')} on <${refusal.where}> carrying "${refusal.text}"`,
  );
}
