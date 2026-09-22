import { readFileSync } from 'node:fs';

import { describe, expect, it } from 'vitest';

import ru from './ru.json';

// Nexus — every codepoint ru.json uses is covered by a BUNDLED font face of the
// UI family that will render it (K11, D23 — the user's ruling).
//
// # The failure it exists for
//
// design/tokens.css sets Aurora's --font-ui to 'Space Grotesk' and Studio's to
// 'Figtree'. Read from the packages' own unicode.json files, Space Grotesk
// covers [vietnamese, latin-ext, latin] and Figtree covers [latin-ext, latin]:
// NEITHER CONTAINS A SINGLE CYRILLIC GLYPH. Every Russian screen therefore fell
// back to the system sans-serif while numbers stayed in JetBrains Mono, and the
// header gained a row. Nothing in the suite noticed for two stages, because
// vitest runs with `css: false` and jsdom has no font engine.
//
// This test is the mechanical half D23 asks for. It cannot see a rendered
// glyph — nothing here can — but it CAN see that the stylesheet the browser
// will load declares a face for the codepoints the locale file actually uses.
// Without it this recurs the moment a third locale, or one more Russian string
// with a character outside the bundled subset, is added.
//
// # Nothing below is retyped
//
//   * The two UI family names come from design/tokens.css's own --font-ui
//     declarations. design/ is read-only and is where they are decided.
//   * The unicode-ranges come from the @font-face rules that are ACTUALLY
//     BUNDLED: src/style.css and every CSS file it @imports, followed into
//     node_modules. A range copied into this file would only prove the test
//     agrees with itself.
//   * The codepoints come from ru.json.
//
// So narrowing the range in style.css turns this red, and adding a Russian
// string outside the bundled subset turns this red. Both were provoked before
// this file was committed.
//
// # It also proves the faces are LOCAL
//
// Every `src` URL of every bundled face is read off disk here. A CDN URL is not
// a file, so it fails — which is rule 11 ("no network call, ever") asserted
// rather than hoped for.

/** Paths are relative to the vitest working directory, which is `frontend/`. */
const STYLESHEET = 'src/style.css';

/** Where a bare package specifier in an @import or a url() resolves. */
const NODE_MODULES = 'node_modules';

/** A codepoint no UI face may claim, used to prove `covers` can answer NO. */
const NOT_A_UI_CODEPOINT = 0x4e00; // CJK unified ideograph, in no bundled range.

/** The two codepoints D23 singles out because they are easy to lose. */
const COMBINING_ACUTE = 0x0301;
const NUMERO_SIGN = 0x2116;

/** An inclusive codepoint interval declared by a `unicode-range` descriptor. */
interface Range {
  from: number;
  to: number;
}

/** One @font-face rule, as bundled. */
interface Face {
  /** The file the rule was read from — named in every failure message. */
  source: string;
  family: string;
  /** Empty means the rule declared no unicode-range, i.e. it claims everything. */
  ranges: Range[];
  /** Every url() in the rule's `src`, resolved to a path on disk. */
  files: string[];
}

// ---------------------------------------------------------------------------
// Paths. Deliberately tiny: `node:path` is not declared for this project (see
// src/test/node-builtins.d.ts for why @types/node is not a dependency), and the
// two operations needed here are a fold and a join.

function normalise(path: string): string {
  const out: string[] = [];
  for (const part of path.split('/')) {
    if (part === '' || part === '.') continue;
    // A `..` that cannot be folded is KEPT, not dropped. design/tokens.css sits
    // above the working directory — `src/style.css` imports it as
    // `../../design/tokens.css` — so swallowing it would turn the real path
    // into `design/tokens.css`, which does not exist. That failed loudly here;
    // the shape to avoid is the one where it resolves to some other real file.
    if (part === '..' && out.length > 0 && out[out.length - 1] !== '..') {
      out.pop();
      continue;
    }
    out.push(part);
  }
  return out.join('/');
}

function directoryOf(file: string): string {
  const cut = file.lastIndexOf('/');
  return cut === -1 ? '' : file.slice(0, cut);
}

/**
 * Resolves a specifier the way the bundler does: relative to the importing
 * file, or out of node_modules when it names a package.
 */
function resolveFrom(importer: string, specifier: string): string {
  if (specifier.startsWith('.')) {
    return normalise(`${directoryOf(importer)}/${specifier}`);
  }
  return normalise(`${NODE_MODULES}/${specifier}`);
}

// ---------------------------------------------------------------------------
// CSS. A regex reader rather than a parser: @font-face bodies contain no nested
// braces, and anything this cannot read is thrown on rather than skipped —
// silence is the one outcome a coverage check may not have (D32).

const IMPORT_RE = /@import\s+(?:url\()?['"]([^'"]+)['"]/g;
const FONT_FACE_RE = /@font-face\s*\{([^}]*)\}/g;
const FAMILY_RE = /font-family\s*:\s*([^;]+)/;
const UNICODE_RANGE_RE = /unicode-range\s*:\s*([^;}]+)/;
const SRC_RE = /src\s*:\s*([^;}]+)/;
const URL_RE = /url\(\s*['"]?([^'")]+)['"]?\s*\)/g;

function read(file: string): string {
  return readFileSync(file, 'utf8');
}

/** Strips one layer of matching quotes from a CSS identifier. */
function unquote(value: string): string {
  const trimmed = value.trim();
  const quote = trimmed.slice(0, 1);
  if ((quote === "'" || quote === '"') && trimmed.endsWith(quote)) {
    return trimmed.slice(1, -1);
  }
  return trimmed;
}

/** The first family of a font-family list — the one that actually renders. */
function firstFamily(list: string): string {
  return unquote(list.split(',')[0]);
}

/**
 * Parses a `unicode-range` descriptor.
 *
 * Throws on anything it does not understand. A range it skipped would shrink
 * the covered set silently, which turns "uncovered" into a false alarm and,
 * worse, makes an unparsed range look like a deliberate omission.
 */
function parseRanges(descriptor: string): Range[] {
  return descriptor.split(',').map((part) => {
    const token = part.trim();
    const interval = /^[uU]\+([0-9a-fA-F]{1,6})-([0-9a-fA-F]{1,6})$/.exec(token);
    if (interval !== null) {
      return { from: parseInt(interval[1], 16), to: parseInt(interval[2], 16) };
    }
    const wildcard = /^[uU]\+([0-9a-fA-F]{0,5})(\?{1,6})$/.exec(token);
    if (wildcard !== null) {
      const digits = wildcard[1] + wildcard[2];
      return {
        from: parseInt(digits.replace(/\?/g, '0'), 16),
        to: parseInt(digits.replace(/\?/g, 'F'), 16),
      };
    }
    const single = /^[uU]\+([0-9a-fA-F]{1,6})$/.exec(token);
    if (single !== null) {
      const at = parseInt(single[1], 16);
      return { from: at, to: at };
    }
    throw new Error(`unicode-range token this test cannot parse: ${JSON.stringify(token)}`);
  });
}

function facesIn(file: string, css: string): Face[] {
  const faces: Face[] = [];

  for (const [, body] of css.matchAll(FONT_FACE_RE)) {
    const family = FAMILY_RE.exec(body);
    if (family === null) {
      throw new Error(`${file}: an @font-face rule with no font-family`);
    }

    const declaredRange = UNICODE_RANGE_RE.exec(body);
    const src = SRC_RE.exec(body);
    const files: string[] = [];
    if (src !== null) {
      for (const [, url] of src[1].matchAll(URL_RE)) {
        files.push(resolveFrom(file, url));
      }
    }

    faces.push({
      source: file,
      family: firstFamily(family[1]),
      ranges: declaredRange === null ? [] : parseRanges(declaredRange[1]),
      files,
    });
  }

  return faces;
}

/**
 * Every @font-face rule the browser will have, reached by following @import
 * from the entry stylesheet exactly as the bundler does.
 */
function bundledFaces(entry: string): Face[] {
  const seen = new Set<string>();
  const faces: Face[] = [];

  const visit = (file: string): void => {
    if (seen.has(file)) return;
    seen.add(file);

    const css = read(file);
    faces.push(...facesIn(file, css));

    for (const [, specifier] of css.matchAll(IMPORT_RE)) {
      visit(resolveFrom(file, specifier));
    }
  };

  visit(entry);
  return faces;
}

// ---------------------------------------------------------------------------
// The two inputs that are not CSS: the family names design/ chose, and the
// characters the Russian locale actually uses.

/** The `--font-ui` families design/tokens.css declares, one per palette. */
function uiFamilies(): string[] {
  const tokens = resolveFrom(
    STYLESHEET,
    // The tokens file is reached through the stylesheet's own @import, so this
    // test cannot drift from what is bundled even if design/ moves.
    (IMPORT_RE.exec(read(STYLESHEET)) ?? ['', ''])[1],
  );
  IMPORT_RE.lastIndex = 0;

  const families = [...read(tokens).matchAll(/--font-ui\s*:\s*([^;]+)/g)].map(([, list]) =>
    firstFamily(list),
  );
  return [...new Set(families)];
}

type Tree = { [key: string]: Tree | string };

/** Every codepoint that appears in a value of ru.json. */
function codepointsOf(tree: Tree, into = new Set<number>()): Set<number> {
  for (const value of Object.values(tree)) {
    if (typeof value === 'object') {
      codepointsOf(value, into);
      continue;
    }
    for (const character of value) {
      into.add(character.codePointAt(0) as number);
    }
  }
  return into;
}

function covers(ranges: Range[], codepoint: number): boolean {
  return ranges.some((range) => codepoint >= range.from && codepoint <= range.to);
}

function hex(codepoint: number): string {
  return `U+${codepoint.toString(16).toUpperCase().padStart(4, '0')}`;
}

// ---------------------------------------------------------------------------

const FACES = bundledFaces(STYLESHEET);
const FAMILIES = uiFamilies();
const RU_CODEPOINTS = [...codepointsOf(ru as Tree)].sort((a, b) => a - b);

/** The faces a UI family will be rendered from, and their combined coverage. */
function coverageOf(family: string): { faces: Face[]; ranges: Range[] } {
  const faces = FACES.filter((face) => face.family === family);
  const ranges = faces.flatMap((face) => (face.ranges.length === 0 ? EVERYTHING : face.ranges));
  return { faces, ranges };
}

/** What a rule with no unicode-range claims: the whole of Unicode. */
const EVERYTHING: Range[] = [{ from: 0, to: 0x10ffff }];

describe('the UI families cover ru.json', () => {
  it.each(FAMILIES.map((family) => [family] as const))(
    '%s has a bundled face for every codepoint ru.json uses',
    (family) => {
      const { ranges } = coverageOf(family);
      const uncovered = RU_CODEPOINTS.filter((codepoint) => !covers(ranges, codepoint)).map(
        (codepoint) => `${hex(codepoint)} ${JSON.stringify(String.fromCodePoint(codepoint))}`,
      );

      expect(
        uncovered,
        `${family}: ru.json uses codepoints no bundled @font-face claims, so they will render ` +
          `in the system fallback face. Widen the unicode-range in ${STYLESHEET}, or bundle the ` +
          `subset that contains them — do not narrow this test.`,
      ).toEqual([]);
    },
  );

  it.each(FAMILIES.map((family) => [family] as const))(
    '%s covers the two codepoints D23 singles out',
    (family) => {
      const { ranges } = coverageOf(family);
      // Both are easy to lose: the combining acute is how Russian marks stress
      // in a dictionary form, and No. is the numero sign every Russian UI uses
      // for "number". Neither appears in ru.json today, which is exactly why
      // they are asserted separately — a codepoint-driven test alone would go
      // green on a range narrowed to what today's strings happen to need.
      expect(covers(ranges, COMBINING_ACUTE), `${family}: ${hex(COMBINING_ACUTE)}`).toBe(true);
      expect(covers(ranges, NUMERO_SIGN), `${family}: ${hex(NUMERO_SIGN)}`).toBe(true);
    },
  );

  it('loads every bundled face from disk, never from a URL', () => {
    const unreadable: string[] = [];

    for (const face of FACES) {
      expect(face.files.length, `${face.source}: ${face.family} declares no src url`).toBeGreaterThan(
        0,
      );
      for (const file of face.files) {
        // A CDN reference is not a path, so this is also how "no network fetch
        // at runtime" is asserted rather than hoped for (rule 11).
        try {
          if (read(file).length === 0) unreadable.push(`${file} (empty)`);
        } catch {
          unreadable.push(`${file} (unreadable)`);
        }
      }
    }

    expect(unreadable, 'font files that are not on disk').toEqual([]);
  });
});

describe('the coverage check is actually checking something', () => {
  // D32: a check that checks nothing must be red, not green. Every assertion
  // above is of the shape "this list is empty", and every one of them passes
  // vacuously if the derivation feeding it produced nothing. So the derivation
  // asserts itself here.

  it('found both UI families, from design/tokens.css', () => {
    expect(FAMILIES.length, 'design/tokens.css declared no --font-ui').toBeGreaterThan(1);
  });

  it('found the bundled faces, including a Cyrillic one per UI family', () => {
    expect(FACES.length, `no @font-face rule was reached from ${STYLESHEET}`).toBeGreaterThan(0);

    for (const family of FAMILIES) {
      const { faces } = coverageOf(family);
      expect(faces.length, `${family}: no bundled @font-face rule at all`).toBeGreaterThan(0);

      // The Cyrillic coverage must come from a face declared in the project's
      // own stylesheet. If it ever came from somewhere else, the ranges asserted
      // above would no longer be the ones D23 put there.
      const cyrillic = faces.filter(
        (face) => face.source === STYLESHEET && covers(face.ranges, 'я'.codePointAt(0) as number),
      );
      expect(
        cyrillic.length,
        `${family}: ${STYLESHEET} declares no Cyrillic @font-face for it`,
      ).toBeGreaterThan(0);
    }
  });

  it('read a real, mostly Cyrillic, set of codepoints out of ru.json', () => {
    expect(RU_CODEPOINTS.length, 'ru.json yielded no codepoints').toBeGreaterThan(0);

    const cyrillic = RU_CODEPOINTS.filter(
      (codepoint) => codepoint >= 0x0400 && codepoint <= 0x04ff,
    );
    expect(cyrillic.length, 'ru.json yielded no Cyrillic at all').toBeGreaterThan(0);
  });

  it('can answer NO — the ranges do not claim the whole of Unicode', () => {
    // Without this, a bug that made `covers` return true unconditionally — or a
    // face that lost its unicode-range and so claimed everything — would leave
    // every assertion above green while proving nothing at all.
    for (const family of FAMILIES) {
      const { ranges } = coverageOf(family);
      expect(
        covers(ranges, NOT_A_UI_CODEPOINT),
        `${family} claims ${hex(NOT_A_UI_CODEPOINT)}, which no bundled subset contains — ` +
          `the coverage set is too wide to prove anything`,
      ).toBe(false);
    }
  });
});
