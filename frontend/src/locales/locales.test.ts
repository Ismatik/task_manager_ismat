import { describe, expect, it } from 'vitest';

import en from './en.json';
import ru from './ru.json';

// Nexus — the locale files are complete, and stay complete.
//
// This test outlives the stage. Every ticket from here on adds keys, and the
// failure mode it exists for is the cheap one: a key added to en.json and
// forgotten in ru.json, which renders as the raw key on a Russian screen that
// the person who added it never looks at.
//
// # Plural keys are compared as ONE key, not as their suffixes
//
// A naive set comparison would fail by design: English needs `_one` and
// `_other`, Russian needs `_one`, `_few`, `_many` and `_other`. So the walk
// below splits a leaf ending in a CLDR plural suffix into a base key plus a
// form, compares the BASE keys across the two files, and then asserts that each
// file carries exactly the forms Intl.PluralRules says that language has. That
// second assertion is the stronger one: it is what catches a Russian plural
// written with only `_one`/`_other`, which is the mistake a hand-rolled lookup
// object makes and the reason i18next is a dependency at all.

const PLURAL_SUFFIX = /_(zero|one|two|few|many|other)$/;

type Leaf = string;
type Tree = { [key: string]: Tree | Leaf };

interface Collected {
  /** Dotted paths of every non-plural leaf. */
  plain: Map<string, string>;
  /** Dotted base path -> the CLDR forms present for it. */
  plurals: Map<string, Set<string>>;
}

function collect(tree: Tree, prefix = '', into?: Collected): Collected {
  const acc: Collected = into ?? { plain: new Map(), plurals: new Map() };

  for (const [key, value] of Object.entries(tree)) {
    const path = prefix === '' ? key : `${prefix}.${key}`;

    if (typeof value === 'object') {
      collect(value, path, acc);
      continue;
    }

    const suffix = PLURAL_SUFFIX.exec(key);
    if (suffix === null) {
      acc.plain.set(path, value);
      continue;
    }

    const base = path.slice(0, path.length - suffix[0].length);
    const forms = acc.plurals.get(base) ?? new Set<string>();
    forms.add(suffix[1]);
    acc.plurals.set(base, forms);
  }

  return acc;
}

function missing(from: Iterable<string>, present: Set<string>): string[] {
  return [...from].filter((key) => !present.has(key)).sort();
}

const EN = collect(en as Tree);
const RU = collect(ru as Tree);

// Keys whose Russian value is INTENTIONALLY the English text. Each one is a
// proper noun or a fixed label that is not translated in either language, and
// each needs a reason. Anything not listed here and identical across the two
// files is an English string sitting in ru.json.
const UNTRANSLATED_ON_PURPOSE = new Set<string>([
  // The product name. "Nexus" is the same word in Russian.
  'app.name',
]);

describe('locale files', () => {
  it('have identical plain key sets', () => {
    const enKeys = new Set(EN.plain.keys());
    const ruKeys = new Set(RU.plain.keys());

    expect(missing(enKeys, ruKeys), 'keys in en.json but missing from ru.json').toEqual([]);
    expect(missing(ruKeys, enKeys), 'keys in ru.json but missing from en.json').toEqual([]);
  });

  it('have identical plural key sets', () => {
    const enBases = new Set(EN.plurals.keys());
    const ruBases = new Set(RU.plurals.keys());

    expect(missing(enBases, ruBases), 'plural keys in en.json but missing from ru.json').toEqual(
      [],
    );
    expect(missing(ruBases, enBases), 'plural keys in ru.json but missing from en.json').toEqual(
      [],
    );
  });

  it.each([
    ['en', EN],
    ['ru', RU],
  ])('carries every CLDR plural form %s actually has', (language, collected) => {
    const required = new Intl.PluralRules(language, { type: 'cardinal' })
      .resolvedOptions()
      .pluralCategories.slice()
      .sort();

    for (const [base, forms] of collected.plurals) {
      expect([...forms].sort(), `${language}: ${base} is missing a plural form`).toEqual(required);
    }
  });

  it.each([
    ['en', EN],
    ['ru', RU],
  ])('has no empty value in %s', (language, collected) => {
    const empty = [...collected.plain.entries()]
      .filter(([, value]) => value.trim() === '')
      .map(([key]) => key);

    expect(empty, `${language}: empty values`).toEqual([]);
  });

  it('has no English text left sitting in ru.json', () => {
    const untranslated = [...EN.plain.entries()]
      .filter(([key, value]) => RU.plain.get(key) === value)
      .map(([key]) => key)
      .filter((key) => !UNTRANSLATED_ON_PURPOSE.has(key))
      .sort();

    expect(untranslated, 'ru.json values identical to en.json').toEqual([]);
  });

  it('is actually being checked — both files carry keys', () => {
    // A vacuous pass is the one way every assertion above can be green while
    // proving nothing: two empty objects have identical key sets.
    expect(EN.plain.size + EN.plurals.size).toBeGreaterThan(0);
    expect(RU.plain.size + RU.plurals.size).toBeGreaterThan(0);
  });
});
