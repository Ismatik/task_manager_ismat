import { afterEach, describe, expect, it } from 'vitest';

import { describeRefusals, PERMITTED_UTILITIES, shrinkRefusals } from './shrink';

// Nexus — the audit's own test, and it exists for one claim in particular.
//
// The rewritten Russian audit (S3-02, D20) says it catches "a NEW utility with
// the same effect on the day it is used". That is a strong claim and it is the
// entire reason the audit changed shape rather than gaining a fifth word.
//
// # The first version of this file did not test that claim, and the claim was false
//
// It used `shrink-hard` — which matched the `/^shrink(-.*)?$/` family the module
// already enumerated, so it demonstrated only that an unpermitted member of a
// KNOWN family is caught. Measured against the same code, a `text-nowrap` and a
// `size-24` on a Russian string produced ZERO refusals, while the functionally
// identical `whitespace-nowrap` and `w-24` produced one each. Both are shipped
// Tailwind 3.4 utilities. The audit was an allow-list inside an enumerated set
// of families, and the header said otherwise — which is K8's own shape: a green
// audit asserting a property it did not have.
//
// So the first test below now uses utilities that are genuinely outside every
// family the module names, and the two corpus tests underneath it are what keeps
// the deny-by-default classification honest as Tailwind grows.

/**
 * # The corpus, and how it was obtained
 *
 * Not from memory. A 266-candidate list covering every Tailwind 3.4 utility
 * family was compiled through postcss with THIS PROJECT'S tailwind config
 * (`design/tailwind.config.js`, so the colour names are the real token names),
 * `@tailwind components; @tailwind utilities;`, and every generated declaration
 * was read. A candidate landed in WIDTH_OR_WRAPPING if any rule it generates
 * declares one of:
 *
 *   width min-width max-width inline-size min-inline-size max-inline-size
 *   flex flex-grow flex-shrink flex-basis flex-wrap
 *   white-space text-wrap text-overflow overflow-wrap word-break hyphens
 *   overflow overflow-x overflow-y -webkit-line-clamp aspect-ratio
 *
 * and in NEUTRAL otherwise. `sr-only` (width: 1px), `line-clamp-*`, `aspect-*`,
 * `container` (max-width) and the whole `size-*` family turned up that way and
 * none of them was on the module's original list.
 *
 * `group` and `peer` generate no CSS at all and are appended to NEUTRAL by hand.
 *
 * ONE measured entry is deliberately not spelled here: the `^outline` utility
 * that removes the ring. It came back NEUTRAL, like every other `^outline` one,
 * and `outline`, `outline-2` and `outline-offset-2` stand for it. It is left out
 * because App.keyboard.test.tsx scans every file under src/ for exactly that
 * string — "a focus ring that only appears sometimes is guesswork" — and a
 * corpus entry would be a false offender in a check that has no exclusions worth
 * spending. Same precedent as the guard's checks 6 and 8: a file that must not
 * name a literal does not name it, and says so instead.
 *
 * The corpus is a snapshot of tailwindcss 3.4.19, which is what is installed. It
 * is not a list of everything Tailwind will ever ship — nothing is, which is the
 * reason the module denies by default. What it proves is the other half: that no
 * entry on the module's IRRELEVANT list is too broad, because a too-broad entry
 * swallows a member of WIDTH_OR_WRAPPING and the first test below goes red
 * naming it.
 */
const WIDTH_OR_WRAPPING = [
  'w-0', 'w-24', 'w-px', 'w-full', 'w-screen', 'w-min', 'w-max', 'w-fit', 'w-auto', 'w-1/2',
  'w-[10rem]', 'w-2', 'w-4', 'w-28', 'min-w-0', 'min-w-24', 'min-w-36', 'min-w-full',
  'min-w-min', 'min-w-max', 'min-w-fit', 'max-w-0', 'max-w-24', 'max-w-full', 'max-w-min',
  'max-w-max', 'max-w-fit', 'max-w-none', 'max-w-prose', 'max-w-screen-sm', 'max-w-xs', 'size-0',
  'size-24', 'size-px', 'size-full', 'size-min', 'size-max', 'size-fit', 'size-auto', 'size-1/2',
  'flex-wrap', 'flex-nowrap', 'flex-wrap-reverse', 'flex-1', 'flex-auto', 'flex-initial',
  'flex-none', 'flex-[2_2_0%]', 'grow', 'grow-0', 'shrink', 'shrink-0', 'basis-0', 'basis-24',
  'basis-auto', 'basis-full', 'whitespace-normal', 'whitespace-nowrap', 'whitespace-pre',
  'whitespace-pre-line', 'whitespace-pre-wrap', 'whitespace-break-spaces', 'text-wrap',
  'text-nowrap', 'text-balance', 'text-pretty', 'text-ellipsis', 'text-clip', 'truncate',
  'break-normal', 'break-words', 'break-all', 'break-keep', 'hyphens-none', 'hyphens-manual',
  'hyphens-auto', 'line-clamp-2', 'line-clamp-none', 'aspect-square', 'aspect-video',
  'aspect-auto', 'overflow-auto', 'overflow-hidden', 'overflow-clip', 'overflow-visible',
  'overflow-scroll', 'overflow-x-auto', 'overflow-y-auto', 'overflow-x-hidden',
  'overflow-y-hidden', 'overflow-x-scroll', 'overflow-y-scroll', 'sr-only', 'not-sr-only',
  'container',];

/** The other half of the same run: utilities that decide nothing about inline size. */
const NEUTRAL = [
  'h-1', 'h-2', 'h-4', 'h-24', 'h-full', 'h-screen', 'min-h-0', 'min-h-24', 'min-h-screen',
  'max-h-24', 'max-h-80', 'max-h-full', 'flex', 'inline-flex', 'flex-row', 'flex-row-reverse',
  'flex-col', 'flex-col-reverse', 'text-xs', 'text-sm', 'text-base', 'text-lg', 'text-xl',
  'text-2xl', 'text-left', 'text-center', 'text-right', 'text-justify', 'text-ink', 'text-muted',
  'text-accent', 'text-accent-2', 'text-on-accent', 'text-danger', 'text-warning',
  'text-success', 'text-transparent', 'font-mono', 'font-sans', 'font-bold', 'font-medium',
  'leading-5', 'leading-tight', 'tracking-wide', 'indent-4', '-indent-4', 'align-middle',
  'list-inside', 'list-disc', 'list-none', 'underline', 'overline', 'line-through',
  'no-underline', 'uppercase', 'lowercase', 'capitalize', 'normal-case', 'italic', 'not-italic',
  'antialiased', 'subpixel-antialiased', 'tabular-nums', 'ordinal', 'slashed-zero',
  'normal-nums', 'decoration-accent', 'decoration-2', 'underline-offset-2', 'placeholder-muted',
  'box-border', 'box-content', 'block', 'inline', 'inline-block', 'hidden', 'contents',
  'flow-root', 'grid', 'inline-grid', 'list-item', 'table', 'table-cell', 'table-auto',
  'table-fixed', 'caption-bottom', 'border-collapse', 'border-separate', 'border-spacing-2',
  'float-left', 'float-none', 'clear-both', 'object-cover', 'object-contain', 'isolate',
  'isolation-auto', 'overscroll-auto', 'overscroll-contain', 'static', 'fixed', 'absolute',
  'relative', 'sticky', 'inset-0', '-inset-1', 'top-4', 'right-4', 'bottom-4', 'left-4',
  'start-4', 'end-4', 'z-40', 'z-50', 'grid-cols-3', 'grid-rows-2', 'col-span-2', 'col-start-1',
  'row-span-2', 'row-start-1', 'auto-cols-max', 'auto-rows-min', 'grid-flow-row', 'gap-1',
  'gap-2', 'gap-x-4', 'gap-y-2', 'space-x-2', 'space-y-2', 'justify-between', 'justify-center',
  'justify-items-center', 'justify-self-end', 'items-center', 'items-baseline', 'items-start',
  'content-center', 'self-end', 'self-start', 'place-items-center', 'place-content-center',
  'order-1', '-order-1', 'p-2', 'p-3', 'p-4', 'px-1.5', 'px-2', 'py-1', 'pt-3', 'pb-2', 'ps-2',
  'pe-2', 'm-2', 'mx-auto', 'mt-1', '-mt-2', '-mx-1', 'rounded', 'rounded-sm', 'rounded-md',
  'rounded-lg', 'rounded-t-lg', 'border', 'border-2', 'border-t', 'border-t-2', 'border-solid',
  'border-dashed', 'border-line', 'border-accent', 'border-danger', 'divide-x', 'divide-line',
  'shadow', 'shadow-sm', 'opacity-0', 'opacity-60', 'ring', 'ring-2', 'ring-accent',
  'ring-offset-2', 'outline', 'outline-2', 'outline-offset-2', 'bg-bg',
  'bg-surface', 'bg-elevated', 'bg-line', 'bg-accent', 'bg-danger', 'bg-transparent',
  'bg-no-repeat', 'bg-repeat-x', 'bg-cover', 'bg-contain', 'bg-center', 'bg-top', 'bg-fixed',
  'bg-local', 'bg-scroll', 'bg-none', 'bg-clip-border', 'bg-origin-border', 'bg-gradient-to-r',
  'from-accent', 'via-accent', 'to-accent', 'transition', 'transition-colors',
  'transition-[width]', 'duration-fast', 'duration-base', 'duration-slow', 'ease-in',
  'delay-100', 'animate-spin', 'transform', 'transform-gpu', 'transform-none', 'scale-95',
  'rotate-45', '-rotate-45', 'translate-x-2', '-translate-y-1', 'skew-y-3', 'origin-center',
  'blur-sm', 'brightness-50', 'contrast-50', 'drop-shadow', 'grayscale', 'hue-rotate-15',
  '-hue-rotate-15', 'invert', 'saturate-50', 'sepia', 'filter', 'backdrop-blur-sm',
  'backdrop-filter', 'backdrop-opacity-50', 'mix-blend-multiply', 'bg-blend-multiply',
  'cursor-pointer', 'select-none', 'pointer-events-none', 'resize', 'resize-none', 'touch-none',
  'will-change-transform', 'appearance-none', 'scroll-smooth', 'scroll-mt-2', 'snap-x',
  'visible', 'invisible', 'collapse', 'forced-color-adjust-none', 'caret-accent',
  'accent-accent', 'fill-accent', 'stroke-accent', 'group', 'peer',];

function tree(html: string): HTMLElement {
  const host = document.createElement('div');
  host.innerHTML = html;
  document.body.append(host);
  return host;
}

/** The refusing tokens the audit reports for one class on one Russian string. */
function judge(utility: string): string[] {
  const host = tree(`<span class="${utility}">Сегодня</span>`);
  const refusals = shrinkRefusals(host);
  host.remove();
  return refusals.flatMap((refusal) => refusal.tokens);
}

afterEach(() => {
  document.body.innerHTML = '';
});

describe('shrinkRefusals', () => {
  it('catches a utility it has never heard of', () => {
    // THE claim, and the reason the audit has this shape at all. None of these
    // three is inside any family the module enumerates:
    //
    //   * `text-nowrap` sets text-wrap: nowrap and clips exactly as
    //     `whitespace-nowrap` does. It is a Tailwind 3.4 utility available in
    //     this build today, and the previous version of this module returned
    //     zero refusals for it.
    //   * `size-24` sets width AND height. Same.
    //   * `quango-42` is not a Tailwind class and never will be. The audit does
    //     not need to know what it means: it cannot show that it is harmless, so
    //     the answer is no.
    for (const utility of ['text-nowrap', 'size-24', 'quango-42']) {
      expect(judge(utility), `${utility} was invisible to the audit`).toEqual([utility]);
    }

    expect(describeRefusals(shrinkRefusals(tree('<span class="size-24">Сегодня</span>')))[0]).toContain(
      'Сегодня',
    );
  });

  it('reports or deliberately permits every utility that touches inline size', () => {
    // The corpus's own reason for existing. A single too-broad entry on the
    // module's IRRELEVANT list — an open `^text-`, say, which would swallow
    // text-nowrap, text-balance, text-ellipsis and text-clip at once — shows up
    // here as a named utility the audit went quiet about.
    const invisible = WIDTH_OR_WRAPPING.filter(
      (utility) => judge(utility).length === 0 && !PERMITTED_UTILITIES.includes(utility),
    );

    expect(
      invisible,
      `these declare a width or wrapping property and the audit said nothing: ${invisible.join(' ')}`,
    ).toEqual([]);
  });

  it('stays quiet about every utility that decides nothing about inline size', () => {
    // The other direction, and the one that decides whether anybody leaves the
    // audit switched on. Deny-by-default only pays if the denials are real: an
    // audit that reports `px-2` and `duration-fast` gets deleted in a week.
    const noisy = NEUTRAL.filter((utility) => judge(utility).length > 0);

    expect(noisy, `false positives on utilities that cannot change a width: ${noisy.join(' ')}`).toEqual(
      [],
    );
  });

  it.each([
    'shrink-0',
    'flex-none',
    'flex-initial',
    'whitespace-nowrap',
    'whitespace-pre',
    'truncate',
    'text-ellipsis',
    'break-keep',
    'basis-32',
    'w-24',
    'min-w-36',
  ])('catches %s on an element carrying text', (utility) => {
    expect(shrinkRefusals(tree(`<span class="${utility}">Готово</span>`))).toHaveLength(1);
  });

  it.each([
    'min-w-0',
    'break-words',
    'flex-1',
    'basis-0',
    'shrink',
    'w-full',
    'max-w-full',
    'whitespace-normal',
    'flex-wrap',
  ])('permits %s, which is how a box gives ground', (utility) => {
    expect(shrinkRefusals(tree(`<span class="${utility}">Готово</span>`))).toEqual([]);
  });

  it('judges a responsive or stateful variant of a refusal as a refusal', () => {
    // `md:w-24` is a width at one breakpoint, which is a width. Before the
    // variant was peeled off, the whole token was compared against the lists,
    // matched nothing on either side, and — under deny-by-default — would now be
    // reported for the wrong reason. Both halves are asserted: the refusal is
    // still reported, and `hover:bg-surface` is still silent.
    expect(judge('md:w-24')).toEqual(['md:w-24']);
    expect(judge('lg:focus:text-nowrap')).toEqual(['lg:focus:text-nowrap']);
    expect(judge('hover:bg-surface')).toEqual([]);
    expect(judge('focus:ring-2')).toEqual([]);
  });

  it('ignores a class that decides nothing about width', () => {
    // A colour, a radius and a padding are not this module's business, and an
    // audit that flagged them would be turned off within a week.
    expect(
      shrinkRefusals(tree('<span class="rounded-sm border-line bg-surface p-2 font-mono">5</span>')),
    ).toEqual([]);
  });

  it('leaves a fixed-size box alone when it holds no text', () => {
    // D20's other half: an icon's `h-4 w-4` is a number, not a string, so its
    // width cannot change with the locale and `shrink-0` on it holds nothing
    // hostage. If this ever starts failing, every icon in the app lights up.
    expect(shrinkRefusals(tree('<svg class="h-4 w-4 shrink-0"></svg>'))).toEqual([]);
  });

  it('judges a container on shrinking but not on its own width', () => {
    // The habit chip is the first case: no text of its own, a localised title
    // below it, and `shrink-0` around the pair. The toast is the second: it caps
    // itself against the viewport, which is a ceiling and not a floor.
    const chip = tree('<div class="shrink-0"><span class="min-w-0">Читать</span></div>');
    expect(shrinkRefusals(chip)).toHaveLength(1);

    const toast = tree('<div class="w-[min(24rem,100%)]"><p class="min-w-0">Ошибка</p></div>');
    expect(shrinkRefusals(toast)).toEqual([]);
  });

  it('does not let a container hide behind the width exemption', () => {
    // The exemption is for `w-*`/`min-w-*`/`max-w-*` and nothing else, because
    // those are the shape of a viewport cap. `size-24` is not: it pins both axes
    // and it is not a cap, so a container wearing one around a Russian string is
    // judged like any other refusal.
    const boxed = tree('<div class="size-24"><span class="min-w-0">Читать</span></div>');
    expect(shrinkRefusals(boxed)[0].tokens).toEqual(['size-24']);
  });

  it('says nothing about an empty tree, and that is the dangerous answer', () => {
    // Recorded rather than relied on. A walk over a screen that never rendered
    // reports no refusals, which is indistinguishable from a clean screen — so
    // every caller asserts separately that it walked something. K8 got through a
    // green audit once; it will not be because the audit had nothing to look at.
    expect(shrinkRefusals(tree(''))).toEqual([]);
  });
});
