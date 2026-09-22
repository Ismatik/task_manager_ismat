import { describe, expect, it } from 'vitest';
import { screen } from '@testing-library/react';

import { DueBadge } from './DueBadge';
import { formatDate } from '../lib/format';
import { BOTH_LANGUAGES, renderIn } from '../test/render';

// Nexus — the badge renders the date Go sent, and nothing else (S3-09, Part 1).
//
// # Which half of the D8 claim this is
//
// Three sentences from the ten-step hand script have had no mechanical backing
// anywhere since Stage 2: "the badge reads the upcoming Friday", "the badge
// reads today" and "the date is cleared on the way back". They are claims about
// the COLUMN-TO-DUE COUPLING (PLAN.md §4, D8) seen through a component, and the
// automated ACCEPT flow cannot make them because its fake client does not model
// the rewrite — and must not be taught to. A fake that implements §4 is §4 with
// a second spelling, which is the defect this project has paid for more than
// any other.
//
// So the claim is split into the two halves that can each be honest alone:
//
//   * Go's half is TestBoardCarriesTheDueDateTheBadgeReads in
//     internal/service/task_test.go. It drives a card to This-week, to Today and
//     back, on a Wednesday clock and on a Friday clock, and asserts the date and
//     the provenance that come back from Board() — the read the board renders.
//   * THIS half is the other end: given that date, the badge shows it, formatted
//     for the active language, and shows nothing it was not given.
//
// # What neither half proves, and what is still owed
//
// That the two MEET. Nothing here crosses the Wails IPC bridge, and nothing in
// Go renders a span. A human at a real keyboard, dragging a real card in a real
// window, is still the only thing that can see Go's date arrive in this
// component — that is S3-09 Part 2, and it is not discharged by this file.
//
// # The dates are the Go half's own literals
//
// The three wire strings below are copied from the table in
// TestBoardCarriesTheDueDateTheBadgeReads, which asserts that Board() emits
// exactly these. Sharing the literals is the most the join can be pinned
// without the bridge: if Go's spelling of a calendar date ever changes, that
// test fails on its `wire` assertion and this file's fixtures are visibly the
// old format.

/** What Go's half asserts a drag to This-week writes: the upcoming Friday. */
const WHAT_THE_WEEK_COLUMN_WROTE = '2026-09-18';

/** What Go's half asserts a drag to Today writes, on its Wednesday clock. */
const WHAT_THE_DAY_COLUMN_WROTE = '2026-09-16';

/** And what a drag back to the first column leaves: no date at all (D1, §4). */
const WHAT_THE_FIRST_COLUMN_LEFT = null;

/**
 * The badge element itself, which is the only element this component renders.
 *
 * `container.firstElementChild` rather than a text query on purpose: the point
 * of several assertions below is what the badge does NOT contain, and a query
 * that found the element by its text could not fail that way.
 */
function badgeIn(container: HTMLElement): HTMLElement {
  const element = container.firstElementChild;

  expect(element, 'the badge rendered nothing at all').not.toBeNull();
  return element as HTMLElement;
}

describe('the due badge', () => {
  // ------------------------------------------------------- non-vacuity first

  it('is fed three states that are actually distinguishable', () => {
    // D32: a check that checks nothing must be red, and every assertion in this
    // file is "given THIS date, the badge shows THAT text". Two fixtures that
    // formatted to the same string would make the two column claims one claim
    // with two names, silently — exactly the shape of the Friday blind spot
    // Go's half exists to close, where the upcoming Friday and the current day
    // are the same date on the one clock every service test used to run on.
    expect(WHAT_THE_WEEK_COLUMN_WROTE).not.toBe(WHAT_THE_DAY_COLUMN_WROTE);
    expect(WHAT_THE_FIRST_COLUMN_LEFT).toBeNull();

    for (const language of BOTH_LANGUAGES) {
      const one = formatDate(WHAT_THE_WEEK_COLUMN_WROTE, language);
      const other = formatDate(WHAT_THE_DAY_COLUMN_WROTE, language);

      expect(one, `the fixtures format to nothing in ${language}`).not.toBe('');
      expect(one, `the two fixtures are the same text in ${language}`).not.toBe(other);
    }

    // And that the two languages are genuinely two: the language assertions
    // further down would pass on a formatter that ignored its locale.
    expect(formatDate(WHAT_THE_WEEK_COLUMN_WROTE, 'en')).not.toBe(
      formatDate(WHAT_THE_WEEK_COLUMN_WROTE, 'ru'),
    );
  });

  // ------------------------------------------------- the three D8 states

  it.each([
    ['the date the week column wrote', WHAT_THE_WEEK_COLUMN_WROTE],
    ['the date the day column wrote', WHAT_THE_DAY_COLUMN_WROTE],
  ])('renders %s, and only that', async (_what, wire) => {
    const { container } = await renderIn('en', <DueBadge due={wire} overdue={false} />);

    // "Formatted": the one date format in the application, asked of the module
    // that owns it rather than written out here.
    expect(badgeIn(container)).toHaveTextContent(formatDate(wire, 'en'));

    // "And only that": the whole render is the badge's own text. A badge that
    // added a word, a separator or a second element would fail here even if the
    // date within it were right.
    expect(container.textContent).toBe(formatDate(wire, 'en'));

    // And it is THAT date rather than whatever formatDate happens to produce —
    // the day and the year off the wire string, read straight out of the
    // fixture. Without this the two assertions above would agree with a broken
    // formatter, because they ask the same broken formatter what to expect.
    const [year, , day] = wire.split('-');
    expect(container.textContent).toContain(year);
    expect(container.textContent).toContain(day);
  });

  it('renders nothing at all once the date has been cleared', async () => {
    // The third claim. A card dragged back to the first column loses an auto
    // date (§4, D1), and Go sends null; the badge is absent, not empty and not
    // a placeholder.
    const { container } = await renderIn('en', <DueBadge due={WHAT_THE_FIRST_COLUMN_LEFT} overdue={false} />);

    expect(container.firstElementChild).toBeNull();
    expect(container.textContent).toBe('');
  });

  // ------------------------------------- it reads the field, it decides nothing

  it.each([true, false])('shows the same date whether the flag is %s', async (flag) => {
    // Card.test.tsx pins that the COLOUR comes from Go's flag and not from the
    // date. This is the same rule from the other side: the TEXT comes from the
    // date and not from the flag. Together they say the badge has two inputs
    // and derives neither.
    const { container } = await renderIn(
      'en',
      <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue={flag} />,
    );

    expect(container.textContent).toBe(formatDate(WHAT_THE_WEEK_COLUMN_WROTE, 'en'));
  });

  it('takes its colour from the flag even when the date contradicts it', async () => {
    // The inversion that proves no comparison happens here: a date long in the
    // past with the flag false, and a date far in the future with it true. A
    // component that compared either date to the clock would colour both the
    // other way round. PLAN.md §4 defines the rule as `due < today AND status !=
    // done`; `status` is not even on this component's props, which is the
    // second reason it could not evaluate it if it tried.
    const past = await renderIn('en', <DueBadge due="2001-01-02" overdue={false} />);
    expect(badgeIn(past.container)).toHaveClass('text-muted');
    past.unmount();

    const future = await renderIn('en', <DueBadge due="2099-12-31" overdue />);
    expect(badgeIn(future.container)).toHaveClass('text-danger');
  });

  // ------------------------------------------------------ language and a11y

  it.each(BOTH_LANGUAGES)('formats the date in %s, and in that language only', async (language) => {
    const { container } = await renderIn(
      language,
      <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue={false} />,
    );

    expect(container.textContent).toBe(formatDate(WHAT_THE_WEEK_COLUMN_WROTE, language));
  });

  it.each(BOTH_LANGUAGES)('says more to a screen reader than it shows in %s', async (language) => {
    // The sentence around the date is the accessible name's, not the visible
    // text's — which is what makes "and only that" above a real constraint
    // rather than an accident of there being nothing else to say. Asserted as a
    // relationship between the two strings so that neither locale file is
    // quoted here: the label contains the date and is longer than it.
    const { container } = await renderIn(
      language,
      <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue={false} />,
    );

    const shown = formatDate(WHAT_THE_WEEK_COLUMN_WROTE, language);
    const label = badgeIn(container).getAttribute('aria-label') ?? '';

    expect(label).toContain(shown);
    expect(label.length, `the ${language} label is only the date again`).toBeGreaterThan(
      shown.length,
    );
  });

  it('gives the overdue state a different sentence from the plain one', async () => {
    // Same reasoning, one level up: D25's "a refusal is not a failure" habit
    // applied to the badge — an overdue card has to SAY it is overdue, not only
    // be red, or the state is invisible to anyone not reading the colour.
    const plain = await renderIn('en', <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue={false} />);
    const first = badgeIn(plain.container).getAttribute('aria-label');
    plain.unmount();

    const late = await renderIn('en', <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue />);

    expect(first).not.toBeNull();
    expect(badgeIn(late.container).getAttribute('aria-label')).not.toBe(first);
  });

  // ------------------------------------------------------------------ mono

  it('puts the date in JetBrains Mono', async () => {
    // PLAN.md §3: JetBrains Mono for every number, date, timer and shortcut
    // hint. Card.test.tsx asserts it as a property of the whole card; here it
    // is asserted of the element itself, which is the one that has to carry it.
    const { container } = await renderIn('en', <DueBadge due={WHAT_THE_WEEK_COLUMN_WROTE} overdue={false} />);

    expect(badgeIn(container)).toHaveClass('font-mono');
    expect(screen.getByText(formatDate(WHAT_THE_WEEK_COLUMN_WROTE, 'en'))).toBeInTheDocument();
  });
});
