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
//     is every class you are wearing one that CANNOT decide whether this box
//     becomes narrower than its own text?
//
// # The default answer is NO, and that is the whole point
//
// The first draft of this module (S3-02) got this wrong in a way worth writing
// down, because it is D17's failure mode inside the ticket written to close the
// previous instance of it. It judged a token only if the token matched one of
// seven `DECIDES_*` regex families — so `text-nowrap` and `size-24`, both real
// Tailwind 3.4 utilities, both functionally identical to `whitespace-nowrap` and
// `w-24` which WERE caught, produced zero refusals. The header claimed "a
// utility nobody has used yet is caught the day it is first used". It was false:
// an unknown utility was invisible, which is exactly K8 again.
//
// It is now the other way round. A token is a refusal UNLESS it is:
//
//   1. on IRRELEVANT — a Tailwind property family that cannot touch inline size
//      (colour, spacing, radius, motion, filters, position, height, ...); or
//   2. on PERMITTED — a utility that does touch inline size and is a permission
//      rather than a refusal (`min-w-0`, `flex-1`, `break-words`), each with its
//      reason written next to it; or
//   3. a width on a CONTAINER, which is a legitimate cap (see below).
//
// Anything else — a Tailwind utility this file has never heard of, a utility
// from a version of Tailwind that does not exist yet, a typo, a class from a
// third-party component — is reported. Permitting one is a deliberate edit to a
// list in this file, in a diff a reviewer reads.
//
// IRRELEVANT is still an enumeration; that cannot be avoided. What changed is
// the DIRECTION IT FAILS IN. An omission from IRRELEVANT is a FALSE POSITIVE: a
// test goes red naming the class, someone looks at it, and either the utility is
// a genuine refusal or one line is added. An omission from the old `DECIDES_*`
// list was a FALSE NEGATIVE: silence, which is what a green audit over a
// clipping card looks like.
//
// Two rules keep IRRELEVANT honest, and they are the only rules it has:
//
//   * A prefix may be OPEN (`^shadow`, `^rounded`, `^gap-`) only when EVERY
//     utility Tailwind generates under it is inline-size-neutral.
//   * The two families that MIX — `text-` (sizes and colours, but also
//     `text-nowrap`, `text-balance`, `text-ellipsis`) and `bg-` (colours, but
//     also background-position and friends) — are enumerated CLOSED, member by
//     member. An open `^text-` is precisely the bug described above.
//
// The lists below were checked against the installed tailwindcss 3.4.19 rather
// than from memory: a 358-utility corpus — 95 + 263, the two arrays committed
// in shrink.test.ts — was compiled with the project's own config, every
// generated declaration read, and each utility labelled by whether
// it declares width / min-width / max-width / flex / flex-basis / flex-shrink /
// flex-grow / flex-wrap / white-space / text-wrap / text-overflow /
// overflow-wrap / word-break / hyphens / overflow / -webkit-line-clamp /
// aspect-ratio. Both halves of that corpus are committed in shrink.test.ts and
// asserted against this module, so a too-broad IRRELEVANT entry fails a test by
// name. `sr-only` (width: 1px), `line-clamp-*`, `aspect-*` and `size-*` were all
// found that way and none of them was on the original list.
//
// THAT NUMBER IS ENFORCED, and it is worth saying why (S3-33, D32). It was
// wrong here, and differently wrong in two other files, for a whole stage: the
// corpus was described as 266, 266 and 357 utilities in three places and is 358
// in fact. A number in prose has no enforcement, and this project has now paid
// for that twice — `e172592` was the first. So shrink.test.ts reads this file,
// finds every "<n>-utility" / "<n>-candidate" claim in it, and compares each
// against `WIDTH_OR_WRAPPING.length + NEUTRAL.length`. Growing a list and
// leaving this paragraph alone is now RED rather than merely untrue.
//
// # What it still cannot do
//
// jsdom has NO LAYOUT ENGINE. Every offsetWidth is 0 and nothing ever overflows,
// so this asserts MECHANISMS and never the absence of clipping. "Russian does
// not clip at a real 1024x768" is S3-09 item 2, by eye, and the onset widths in
// K8 are +/-5% arithmetic rather than measurements.

/**
 * The project's colour token names — CLAUDE.md fixes this list, and `design/`
 * owns the values. Used to close the `bg-` and `text-` families: `bg-surface`
 * and `text-muted` are colours, `bg-cover` is a background size, and
 * `text-nowrap` is neither.
 *
 * A palette colour from outside this list (`bg-red-500`) is therefore reported.
 * That is not a false positive: it is guard check 1's rule arriving by a second
 * road, and there is nowhere in this project it would be correct.
 */
const COLOUR =
  '(bg|surface|elevated|line|ink|muted|accent|accent-2|on-accent|danger|warning|success|transparent|current|inherit)';

/**
 * Tailwind property families that CANNOT make a box refuse to become narrower
 * than its own text, and cannot pin its width.
 *
 * Grouped by what the family sets, because that is the only question that
 * decides membership. Every open prefix here was checked against the generated
 * CSS; see the header for the two rules.
 */
const IRRELEVANT: RegExp[] = [
  // --- colour, in every slot that takes one -------------------------------
  new RegExp(`^(bg|text)-${COLOUR}$`),
  /^(border|divide|ring|outline|shadow|fill|stroke|caret|accent|decoration|placeholder|from|via|to)(-|$)/,
  /^(bg|text|border|divide|ring|placeholder)-opacity-/,

  // --- background, everything about it except the colour -------------------
  /^bg-(no-repeat|repeat|repeat-x|repeat-y|repeat-round|repeat-space)$/,
  /^bg-(auto|cover|contain)$/,
  /^bg-(bottom|center|left|left-bottom|left-top|right|right-bottom|right-top|top)$/,
  /^bg-(fixed|local|scroll|none)$/,
  /^bg-(clip|origin|gradient-to|blend)-/,

  // --- type: size, family, weight, spacing, decoration, numerals -----------
  // `text-` is CLOSED. The wrapping utilities live under the same prefix.
  /^text-(xs|sm|base|lg|xl|[2-9]xl)$/,
  /^text-(left|center|right|justify|start|end)$/,
  /^font-/,
  /^(leading|tracking|align|list|underline-offset)-/,
  /^-?indent-/,
  /^(underline|overline|line-through|no-underline)$/,
  /^(uppercase|lowercase|capitalize|normal-case)$/,
  /^(italic|not-italic|antialiased|subpixel-antialiased)$/,
  /^(normal-nums|ordinal|slashed-zero|lining-nums|oldstyle-nums|proportional-nums|tabular-nums|diagonal-fractions|stacked-fractions)$/,

  // --- the block axis, which is not this module's axis ----------------------
  /^(h|min-h|max-h)-/,

  // --- spacing: padding, margin, gaps, gutters ------------------------------
  /^-?[pm][trblxyse]?-/,
  /^gap(-[xy])?-/,
  /^space-[xy]-/,

  // --- position and stacking -------------------------------------------------
  /^(static|fixed|absolute|relative|sticky)$/,
  /^-?(inset|top|right|bottom|left|start|end)-/,
  /^z-/,
  /^(isolate|isolation-auto)$/,
  /^(float|clear)-/,

  // --- display, when it is display and nothing else --------------------------
  /^(block|inline|inline-block|hidden|contents|flow-root|grid|inline-grid|list-item)$/,
  // `flex` is display and `flex-col` is a main-axis direction. Neither declares
  // anything about whether THIS box may become narrower — `flex-1`, `basis-*`
  // and `shrink-*`, which do, are judged and are below.
  /^(inline-)?flex$/,
  /^flex-(row|col)(-reverse)?$/,
  /^table(-|$)/,
  /^(caption|border-collapse|border-separate)/,

  // --- how children are placed, which is not how this box is sized -----------
  /^(items|justify|content|self|place)-/,
  /^-?order-/,
  /^(grid-cols|grid-rows|grid-flow|auto-cols|auto-rows)-/,
  /^-?(col|row)-/,

  // --- borders, corners, rings, shadows, opacity ------------------------------
  /^rounded(-|$)/,
  /^opacity-/,

  // --- motion and transforms --------------------------------------------------
  /^(transition|duration|ease|delay|animate)(-|$)/,
  /^transform(-|$)/,
  /^-?(scale|rotate|translate|skew|origin)-/,
  /^-?perspective(-|$)/,

  // --- filters, including the one that started K7 ------------------------------
  /^-?(blur|brightness|contrast|drop-shadow|grayscale|hue-rotate|invert|saturate|sepia|filter)(-|$)/,
  /^backdrop-/,
  /^(mix-blend|bg-blend)-/,

  // --- interactivity, painting, and markers that emit no CSS at all -------------
  /^(cursor|select|pointer-events|resize|scroll|snap|touch|will-change|appearance|overscroll|object|box|forced-color-adjust)(-|$)/,
  /^(visible|invisible|collapse)$/,
  /^(group|peer)(\/|$)/,
];

/**
 * The class families that pin a box to a width.
 *
 * This is NOT what finds refusals any more — anything unrecognised is a refusal
 * without needing to match here. It survives for one job: EXEMPTING CONTAINERS.
 *
 * A container legitimately caps its own width against the viewport — the toast's
 * `w-[min(24rem,...)]` and the two overlay panels' `w-[min(32rem,100%)]` are
 * exactly that — and the column's `min-w-36` is D21's derived layout floor. On
 * an element holding a localised string, a width is a promise about how long
 * that string is, which is a promise no locale keeps.
 *
 * `size-*` is deliberately NOT here. It sets height as well as width, it is not
 * the shape of a viewport cap, and a container wearing one around a localised
 * string is making the same promise a text element would.
 */
const WIDTH_FAMILY: RegExp[] = [/^w-/, /^min-w-/, /^max-w-/];

/**
 * Utilities that DO decide inline size, and are a permission rather than a
 * refusal — each with its reason.
 *
 * Short on purpose. A layout that needs something not here is a layout making a
 * new claim about how a localised string behaves, and D20 wants that claim
 * written down rather than assumed.
 */
const PERMITTED = new Map<string, string>([
  ['flex-wrap', 'the opposite of a refusal: children move to a new line'],
  ['flex-1', 'grow and shrink from a zero basis — the shrinkable default'],
  ['flex-auto', 'grow and shrink'],
  ['shrink', 'flex-shrink: 1, which is the permission itself'],
  ['grow', 'flex-grow decides growth; it never raises the minimum'],
  ['grow-0', 'refuses to GROW, which is not refusing to shrink'],
  ['basis-0', 'a zero basis, so the box is sized by the flex line and not by its text'],
  ['min-w-0', 'releases min-width:auto — the thing that lets a flex item go below min-content'],
  ['w-full', 'tracks the parent, not the content'],
  ['max-w-full', 'a ceiling, never a floor'],
  ['break-words', 'lets an unbreakable token wrap rather than overflow'],
  ['break-normal', 'the initial value'],
  ['whitespace-normal', 'the initial value — text may wrap'],
  ['whitespace-pre-wrap', 'preserves runs of spaces and still wraps'],
  ['text-wrap', 'text-wrap: wrap, the initial value'],
  [
    'sr-only',
    'width:1px, but the box is visually hidden — its width is not a layout and no locale can overflow it',
  ],
  [
    'overflow-y-auto',
    'a scrollbar in the BLOCK axis; the inline axis, which is the one localised text overflows, is untouched',
  ],
  [
    'overflow-x-auto',
    "the board's own sideways scroll, which D20 keeps; a COLUMN is asserted separately to have no overflow-x",
  ],
]);

/**
 * The utilities on PERMITTED, for the corpus test in shrink.test.ts.
 *
 * Exported rather than restated there: "every utility that touches inline size
 * is either reported or deliberately permitted" is the property being asserted,
 * and a second hand-kept copy of this list in the test would make the assertion
 * pass by construction the moment the two drifted.
 */
export const PERMITTED_UTILITIES: readonly string[] = [...PERMITTED.keys()];

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
 * The utility inside a token, with its variants and its `!` peeled off.
 *
 * `hover:bg-surface` is `bg-surface`; `md:focus:w-24` and `[&>p]:w-24` are
 * `w-24`. This matters in the direction you would expect: a responsive or
 * stateful variant of a refusal is still a refusal, and before this the whole
 * token was compared against the lists and matched nothing.
 *
 * The leading `-` of a negative utility is NOT a variant and is left alone, and
 * an arbitrary value is bracket-aware, so `w-[min(24rem,calc(100vw-2rem))]`
 * survives intact and `supports-[display:grid]:flex` loses only its variant.
 */
function bareUtility(token: string): string {
  let rest = token;
  for (;;) {
    const variant = /^[^:[\]]*(?:\[[^\]]*\][^:[\]]*)*:/.exec(rest);
    if (!variant || variant[0] === ':') {
      break;
    }
    rest = rest.slice(variant[0].length);
  }
  return rest.replace(/^!/, '');
}

/**
 * Is this token, on this kind of element, a refusal to shrink?
 *
 * Deny by default: the three ways out are named, and everything else is a no.
 */
function isRefusal(token: string, carriesText: boolean): boolean {
  const utility = bareUtility(token);
  if (utility === '') {
    return false;
  }
  if (PERMITTED.has(utility)) {
    return false;
  }
  if (IRRELEVANT.some((family) => family.test(utility))) {
    return false;
  }
  return carriesText || !WIDTH_FAMILY.some((family) => family.test(utility));
}

/**
 * Every element under `root` that has text in it and wears a class forbidding
 * it to shrink.
 *
 * Two tiers, because two kinds of element are at risk in different ways:
 *
 *   1. An element CARRYING text — a direct text node with a non-whitespace
 *      character. Its width can depend on the language, so it is judged on
 *      everything: it may neither refuse to shrink nor be pinned to a width.
 *   2. An element merely CONTAINING text further down. Its width also follows
 *      the language, through its children — the habit chip is one, and it held
 *      `shrink-0` around a Russian habit title until S3-02. So it is judged the
 *      same way EXCEPT that a `w-*`/`min-w-*`/`max-w-*` is allowed: capping a
 *      panel against the viewport is a legitimate thing for a container to do.
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

    const tokens = classTokens(element).filter((token) => isRefusal(token, own !== ''));

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

/**
 * Every element under `root` that this module could form an opinion about —
 * i.e. that wears at least one class token.
 *
 * # Why this is exported rather than written at the call site (S3-33, D32)
 *
 * `shrinkRefusals` returning an empty array means either "nothing refuses to
 * shrink" or "nothing was walked", and only one of those is good news. The
 * caller in App.accept.test.tsx therefore guards it with a non-vacuity count —
 * and that count was `element.className !== ''`, which is wrong on an
 * `SVGElement`: `className` there is an `SVGAnimatedString`, an object, so it
 * never equals the string `''` and EVERY ICON ON SCREEN counted toward the
 * threshold. The guard read stronger than it was, over the audit K8 walked
 * straight past.
 *
 * It reads the class the way `classTokens` does — `getAttribute('class')`,
 * which is a string on HTML and SVG alike — so a class-less `<svg>` no longer
 * inflates the count and an icon that DOES carry a class still counts, because
 * the audit really can judge that one.
 *
 * It lives here, next to `classTokens`, because a copy of this filter at the
 * call site would be the same rule in two places — which is the defect this
 * project has failed review over more times than any other.
 */
export function judgeableElements(root: ParentNode): Element[] {
  return [...root.querySelectorAll('*')].filter((element) => classTokens(element).length > 0);
}

/** The refusals as lines, for an assertion message that says where to look. */
export function describeRefusals(refusals: ShrinkRefusal[]): string[] {
  return refusals.map(
    (refusal) => `${refusal.tokens.join(' ')} on <${refusal.where}> carrying "${refusal.text}"`,
  );
}
