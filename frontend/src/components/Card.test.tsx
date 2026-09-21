import { describe, expect, it } from 'vitest';
import { screen } from '@testing-library/react';

import { Card } from './Card';
import { priorityChip } from '../lib/priority';
import { BOTH_LANGUAGES, renderIn } from '../test/render';
import {
  node,
  nodeView,
  PLACEHOLDER_STATUS,
  tag,
  timerView,
  undefinedProgress,
} from '../test/fakeClient';
import type { NodeView, ProgressView } from '../lib/client';

// Nexus — the card renders what Go sent, and decides nothing.
//
// Two conventions in this file are deliberate and worth stating up front.
//
// 1. NO COLUMN NAME IS WRITTEN DOWN. Which strings are Kanban columns is
//    domain.Status's answer and `make guard` check 2 refuses a quoted copy of
//    it in frontend/src. So "the same card in another column" is expressed as
//    two DIFFERENT opaque status strings, which is also the stronger assertion:
//    it proves the card's behaviour does not depend on the value at all, rather
//    than proving it for two particular names.
//
// 2. NO CHIP LABEL IS WRITTEN DOWN. The three designators live in
//    lib/priority.ts and nowhere else (guard check 5), so the chip assertions
//    ask that module what the label is instead of restating it.

/** Two opaque statuses. The card never reads either; that is the point. */
const ONE_COLUMN = PLACEHOLDER_STATUS;
const ANOTHER_COLUMN = 'a-different-status-from-go';

/** A progress Go measured: three of five work leaves finished. */
const MEASURED: ProgressView = {
  ...undefinedProgress,
  done: 3,
  total: 5,
  percent: 60,
  defined: true,
};

/**
 * An empty project: a leaf whose progress Go could not define (D11).
 *
 * It is a leaf because nothing beneath it has a Kanban column, and its progress
 * is undefined because there is no work in it to measure. Both facts come from
 * Go; this fixture only carries them.
 */
function emptyProject(status: string): NodeView {
  return nodeView({
    node: node({ id: 'p-1', type: 'project', title: 'Kitchen rebuild', status }),
    status,
    progress: undefinedProgress,
    isLeaf: true,
  });
}

/** Every element inside the card that actually shows text. */
function textCarriers(): HTMLElement[] {
  return [...screen.getByRole('article').querySelectorAll<HTMLElement>('*')].filter(
    (element) => (element.textContent ?? '').trim() !== '',
  );
}

describe('the card', () => {
  it('renders the title and the type icon', async () => {
    await renderIn('en', <Card view={nodeView()} />);

    expect(screen.getByRole('article')).toHaveTextContent('Write the board');
    expect(screen.getByRole('img', { name: 'Task' })).toBeInTheDocument();
  });

  // ---------------------------------------------------------------- D15 / K3

  it('renders the empty-project marker, and no bar, for an undefined progress', async () => {
    // K3, named: an empty project whose DERIVED status is the finished column
    // used to sit there with nothing on it at all — no bar, because D11 leaves
    // its progress genuinely undefined, and no other chrome. D15 closes it.
    await renderIn('en', <Card view={emptyProject(ONE_COLUMN)} />);

    expect(screen.getByTestId('progress-empty')).toHaveTextContent('Empty project');
    expect(screen.queryByRole('progressbar')).toBeNull();
    // And the card is not blank: the title is still on it.
    expect(screen.getByRole('article')).toHaveTextContent('Kitchen rebuild');
  });

  it('renders the same marker whatever column the card is in', async () => {
    // The rule is `defined`, not "is this the finished column". A rule that
    // fired in one column only would be a rule with a column in it (D15).
    const { unmount } = await renderIn('en', <Card view={emptyProject(ONE_COLUMN)} />);
    const first = screen.getByTestId('progress-empty').textContent;
    unmount();

    await renderIn('en', <Card view={emptyProject(ANOTHER_COLUMN)} />);

    expect(screen.getByTestId('progress-empty').textContent).toBe(first);
    expect(screen.queryByRole('progressbar')).toBeNull();
  });

  it('renders the marker in Russian too', async () => {
    await renderIn('ru', <Card view={emptyProject(ONE_COLUMN)} />);

    expect(screen.getByTestId('progress-empty')).toHaveTextContent('Пустой проект');
  });

  it('renders a bar at exactly the percent Go computed, and no marker', async () => {
    await renderIn(
      'en',
      <Card
        view={nodeView({
          node: node({ type: 'project', title: 'Kitchen rebuild' }),
          progress: MEASURED,
          isLeaf: false,
        })}
      />,
    );

    const bar = screen.getByRole('progressbar');

    expect(bar).toHaveAttribute('aria-valuenow', '60');
    expect(bar.firstElementChild).toHaveStyle({ width: '60%' });
    expect(screen.queryByTestId('progress-empty')).toBeNull();
    expect(screen.getByRole('article')).toHaveTextContent('3 of 5');
  });

  it('never renders a percentage the DTO did not carry', async () => {
    // 3 of 5 is 60%, and the card shows 60 because dto.go rounded it — not
    // because anything here divided. Handing it a percent that disagrees with
    // done/total proves which of the two it is reading.
    await renderIn(
      'en',
      <Card view={nodeView({ progress: { ...MEASURED, percent: 17 }, isLeaf: false })} />,
    );

    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '17');
  });

  it('renders no progress slot at all for a work leaf', async () => {
    // A leaf task is one work leaf, so Go measures it as 0 of 1 or 1 of 1 — a
    // bar that can only ever be empty or full, repeating what its column
    // already says. D15's table: a non-project renders nothing in this slot.
    await renderIn(
      'en',
      <Card view={nodeView({ progress: { ...undefinedProgress, total: 1, defined: true } })} />,
    );

    expect(screen.queryByRole('progressbar')).toBeNull();
    expect(screen.queryByTestId('progress-empty')).toBeNull();
  });

  // ------------------------------------------------------------------ chips

  it('renders no chip for priority 4', async () => {
    await renderIn('en', <Card view={nodeView({ node: node({ priority: 4 }) })} />);

    for (const priority of [1, 2, 3]) {
      expect(screen.queryByText(priorityChip(priority)!.label)).toBeNull();
    }
  });

  it.each([1, 2])('renders priority %i in danger', async (priority) => {
    await renderIn('en', <Card view={nodeView({ node: node({ priority }) })} />);

    expect(screen.getByText(priorityChip(priority)!.label)).toHaveClass('text-danger');
  });

  it('renders priority 3 in warning', async () => {
    await renderIn('en', <Card view={nodeView({ node: node({ priority: 3 }) })} />);

    expect(screen.getByText(priorityChip(3)!.label)).toHaveClass('text-warning');
  });

  // --------------------------------------------------------------- overdue

  it('reddens the due badge from the overdue FLAG, not from the date', async () => {
    // The assertion that proves TypeScript is not deciding: the flag says
    // overdue and the date is years away. A component that compared the date to
    // the clock would render this muted.
    await renderIn(
      'en',
      <Card view={nodeView({ node: node({ due: '2099-12-31' }), overdue: true })} />,
    );

    expect(screen.getByText(/2099/)).toHaveClass('text-danger');
  });

  it('leaves the due badge muted when the flag is false, whatever the date', async () => {
    // The mirror image: a date long past, and no flag. Still muted.
    await renderIn(
      'en',
      <Card view={nodeView({ node: node({ due: '2001-01-02' }), overdue: false })} />,
    );

    expect(screen.getByText(/2001/)).toHaveClass('text-muted');
  });

  it('renders no due badge when there is no due date', async () => {
    await renderIn('en', <Card view={nodeView()} />);

    expect(screen.queryByText(/\d{4}/)).toBeNull();
  });

  // ------------------------------------------------------- timer, tags, mono

  it('shows the timer indicator only while Go says the timer runs', async () => {
    const { unmount } = await renderIn('en', <Card view={nodeView()} />);
    expect(screen.queryByRole('img', { name: 'Timer running' })).toBeNull();
    unmount();

    await renderIn('en', <Card view={nodeView({ timer: timerView({ running: true }) })} />);

    expect(screen.getByRole('img', { name: 'Timer running' })).toBeInTheDocument();
  });

  it('renders the tags the service already resolved', async () => {
    await renderIn(
      'en',
      <Card
        view={nodeView({
          tags: [tag({ id: 't1', name: 'home' }), tag({ id: 't2', name: 'errand' })],
        })}
      />,
    );

    expect(screen.getByRole('list', { name: 'Tags' })).toBeInTheDocument();
    expect(screen.getAllByRole('listitem')).toHaveLength(2);
  });

  it('puts every number and date in JetBrains Mono, and no word in it', async () => {
    // PLAN.md section 3: JetBrains Mono for every number, date, timer and
    // shortcut hint. Asserted as a property of the whole card rather than
    // element by element, so a number added later is covered by this test
    // without anyone remembering to extend it.
    await renderIn(
      'en',
      <Card
        view={nodeView({
          node: node({ priority: 1, due: '2026-03-01', estimateMin: 45, title: 'Write the board' }),
          progress: MEASURED,
          isLeaf: false,
          tags: [tag({ name: 'home' })],
        })}
      />,
    );

    for (const element of textCarriers()) {
      if (element.children.length > 0) {
        continue;
      }
      const carriesDigits = /\d/.test(element.textContent ?? '');
      expect(
        element.classList.contains('font-mono'),
        `"${element.textContent}" font-mono=${element.classList.contains('font-mono')}`,
      ).toBe(carriesDigits);
    }
  });

  // -------------------------------------------------------------- RU width

  it.each(BOTH_LANGUAGES)('survives a long title and long tags in %s', async (language) => {
    // jsdom has no layout engine, so no test in this project can measure a
    // pixel of clipping — that half is on the hand pass. What CAN be asserted
    // mechanically is the contract that makes clipping impossible: every string
    // is present in full, and no element that carries text also carries a class
    // that would cut it off. A fixed width is set on the wrapper so the intent
    // of the case is visible even though jsdom will not honour it.
    const title = 'Переработать документацию по интеграции и согласовать её со смежной командой';
    const label = 'документация-и-согласование';

    await renderIn(
      language,
      <div style={{ width: '220px' }}>
        <Card
          view={nodeView({
            node: node({ title, priority: 1, due: '2026-03-01', estimateMin: 45 }),
            overdue: true,
            tags: [tag({ id: 't1', name: label }), tag({ id: 't2', name: label })],
            progress: undefinedProgress,
            isLeaf: true,
          })}
        />
        ,
      </div>,
    );

    expect(screen.getByRole('article')).toHaveTextContent(title);
    expect(screen.getAllByText(label)).toHaveLength(2);

    const clipping = /(^|\s)(truncate|whitespace-nowrap|text-ellipsis|overflow-hidden)(\s|$)/;
    for (const element of textCarriers()) {
      expect(element.className, `${element.tagName} clips its text`).not.toMatch(clipping);
    }
  });

  it('renders the whole card in Russian', async () => {
    await renderIn(
      'ru',
      <Card
        view={nodeView({
          node: node({ type: 'bug', priority: 3, estimateMin: 45, due: '2026-03-01' }),
          overdue: true,
        })}
      />,
    );

    expect(screen.getByRole('img', { name: 'Ошибка' })).toBeInTheDocument();
    expect(screen.getByRole('article')).toHaveTextContent('45 мин');
  });
});
