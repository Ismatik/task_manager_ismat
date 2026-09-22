import { afterEach, describe, expect, it } from 'vitest';

import { describeRefusals, shrinkRefusals } from './shrink';

// Nexus — the audit's own test, and it exists for one claim in particular.
//
// The rewritten Russian audit (S3-02, D20) says it catches "a NEW utility with
// the same effect on the day it is used". That is a strong claim and it is the
// entire reason the audit changed shape rather than gaining a fifth word, so it
// is asserted here directly — with a utility that does not exist in Tailwind and
// certainly does not exist in this repository.
//
// Running the audit over the pre-S3-02 component tree proved it catches
// `shrink-0`. That is one mechanism. This file covers the others and, more
// importantly, the absence of any mechanism at all.

function tree(html: string): HTMLElement {
  const host = document.createElement('div');
  host.innerHTML = html;
  document.body.append(host);
  return host;
}

afterEach(() => {
  document.body.innerHTML = '';
});

describe('shrinkRefusals', () => {
  it('catches a utility it has never heard of', () => {
    // `shrink-hard` is not a Tailwind class and never will be. The audit does
    // not need to know what it means: it is in the family that decides
    // shrinking and it is not on the permitted list, so the answer is no.
    const refusals = shrinkRefusals(tree('<span class="shrink-hard">Сегодня</span>'));

    expect(refusals).toHaveLength(1);
    expect(refusals[0].tokens).toEqual(['shrink-hard']);
    expect(describeRefusals(refusals)[0]).toContain('Сегодня');
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

  it('says nothing about an empty tree, and that is the dangerous answer', () => {
    // Recorded rather than relied on. A walk over a screen that never rendered
    // reports no refusals, which is indistinguishable from a clean screen — so
    // every caller asserts separately that it walked something. K8 got through a
    // green audit once; it will not be because the audit had nothing to look at.
    expect(shrinkRefusals(tree(''))).toEqual([]);
  });
});
