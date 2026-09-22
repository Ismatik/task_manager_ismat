import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nextProvider } from 'react-i18next';

import { ToastList } from './Toast';
import { createI18n } from '../lib/i18n';
import { createAppStore } from '../store';
import { TOAST_DISMISS_MS } from '../store/toast';
import { columnView, createFakeClient, habitView, nodeView } from '../test/fakeClient';
import type { Toast } from '../store';

async function renderToasts(
  language: string,
  toasts: readonly Toast[],
  onDismiss: (id: number) => void = () => {},
) {
  const i18n = await createI18n(language);

  return render(
    <I18nextProvider i18n={i18n}>
      <ToastList toasts={toasts} onDismiss={onDismiss} />
    </I18nextProvider>,
  );
}

const failure: Toast = {
  id: 1,
  operationKey: 'toast.operation.move',
  messageKey: 'toast.error.body',
  kind: 'failure',
  count: 1,
};

afterEach(() => {
  vi.useRealTimers();
});

describe('the toast', () => {
  it('renders nothing when there is nothing to say', async () => {
    await renderToasts('en', []);

    expect(screen.queryByRole('alert')).toBeNull();
  });

  it('renders the translated message in English', async () => {
    await renderToasts('en', [failure]);

    expect(screen.getByRole('alert')).toHaveTextContent('Nexus could not finish that.');
  });

  it('renders the translated message in Russian', async () => {
    await renderToasts('ru', [failure]);

    // The store held a KEY, not a sentence. This is the whole reason: the same
    // toast reads in whichever language is current at render time.
    expect(screen.getByRole('alert')).toHaveTextContent('Nexus не смог выполнить это действие.');
  });

  it('never shows the raw Go error', async () => {
    const cause = new Error('node 7f3: a project can never be doing');
    await renderToasts('en', [{ ...failure, cause }]);

    expect(screen.getByRole('alert').textContent).not.toContain('7f3');
    expect(screen.getByRole('alert').textContent).not.toContain('project');
  });

  it('announces itself to assistive technology', async () => {
    await renderToasts('en', [failure]);

    expect(screen.getByRole('alert')).toHaveAttribute('aria-live', 'assertive');
  });

  it('is dismissable from the keyboard alone', async () => {
    const dismissed: number[] = [];
    const user = userEvent.setup();
    await renderToasts('en', [failure], (id) => dismissed.push(id));

    // No click, no pointer: tab to the button and press it. "Everything
    // keyboard-reachable" (PLAN.md section 2) is an interaction property, and
    // this is what asserting it looks like.
    await user.tab();
    expect(screen.getByRole('button')).toHaveFocus();

    await user.keyboard('{Enter}');
    expect(dismissed).toEqual([1]);
  });

  it('dismisses the toast the user asked for, not the first one', async () => {
    const onDismiss = vi.fn();
    const user = userEvent.setup();
    await renderToasts(
      'en',
      [failure, { ...failure, id: 2 }],
      onDismiss,
    );

    const buttons = screen.getAllByRole('button');
    expect(buttons).toHaveLength(2);

    await user.tab();
    await user.tab();
    await user.keyboard('{Enter}');

    expect(onDismiss).toHaveBeenCalledTimes(1);
    expect(onDismiss).toHaveBeenCalledWith(2);
  });

  it('carries no hard-coded text — the Russian panel differs from the English one', async () => {
    const english = await createI18n('en');
    const russian = await createI18n('ru');

    const { container: en } = render(
      <I18nextProvider i18n={english}>
        <ToastList toasts={[failure]} onDismiss={() => {}} />
      </I18nextProvider>,
    );
    const { container: ru } = render(
      <I18nextProvider i18n={russian}>
        <ToastList toasts={[failure]} onDismiss={() => {}} />
      </I18nextProvider>,
    );

    expect(en.textContent).not.toBe('');
    expect(ru.textContent).not.toBe('');
    expect(ru.textContent).not.toBe(en.textContent);
  });
});

// --- D24: the repeat count, and the self-dismissing timer (S3-07) -----------

const repeated: Toast = { ...failure, count: 3 };

describe('the repeat count', () => {
  it('is not shown at all for a toast that happened once', async () => {
    await renderToasts('en', [failure]);

    expect(screen.getByRole('alert').textContent).not.toContain('Happened');
  });

  it('is shown, pluralised, in English', async () => {
    await renderToasts('en', [repeated]);

    expect(screen.getByRole('alert')).toHaveTextContent('Happened 3 times');
  });

  it('is shown, pluralised with the Russian _few form, in Russian', async () => {
    await renderToasts('ru', [repeated]);

    // Russian has three plural forms and 3 takes `_few`, not `_many`. This is
    // why the count goes through i18next rather than through a template.
    expect(screen.getByRole('alert')).toHaveTextContent('Произошло 3 раза');
  });

  it('takes the Russian _many form at 5, which a naive one/other table gets wrong', async () => {
    await renderToasts('ru', [{ ...failure, count: 5 }]);

    expect(screen.getByRole('alert')).toHaveTextContent('Произошло 5 раз');
  });

  it('renders the number in font-mono, like every other number in the app', async () => {
    await renderToasts('en', [repeated]);

    const line = screen.getByText(/Happened 3 times/);
    expect(line.className).toContain('font-mono');
  });
});

describe('the dismiss timer', () => {
  /**
   * Builds the i18n instance BEFORE the clock is faked.
   *
   * i18next's init returns a promise, and a promise awaited under a fake clock
   * is a promise nothing is advancing. Everything after this line is timer
   * behaviour, which is exactly what the fake clock is for.
   */
  async function renderFaked(toasts: readonly Toast[], onDismiss: (id: number) => void) {
    const i18n = await createI18n('en');
    vi.useFakeTimers();

    // fireEvent and the DOM's own focus(), NOT user-event: user-event schedules
    // its own delays on the same clock these tests drive by hand, and the two
    // deadlock. The events dispatched below are the ones the browser dispatches
    // — focusin bubbles from the button to the panel, and React synthesises
    // onMouseEnter/onMouseLeave from mouseover/mouseout — so nothing is faked
    // except the clock.
    const view = render(
      <I18nextProvider i18n={i18n}>
        <ToastList toasts={toasts} onDismiss={onDismiss} />
      </I18nextProvider>,
    );

    return { view, i18n };
  }

  it('dismisses the toast by itself after the interval, and not one tick before', async () => {
    const onDismiss = vi.fn();
    await renderFaked([failure], onDismiss);

    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS - 1);
    expect(onDismiss).not.toHaveBeenCalled();

    await vi.advanceTimersByTimeAsync(1);
    expect(onDismiss).toHaveBeenCalledExactlyOnceWith(failure.id);
  });

  it('pauses while the toast holds focus, and resumes from where it paused on blur', async () => {
    const onDismiss = vi.fn();
    await renderFaked([failure], onDismiss);
    const dismiss = screen.getByRole('button');

    // Spend a third of the interval, then take focus — really: focus() is the
    // DOM's own call and the focusin it fires is what bubbles to the panel.
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS / 3);
    dismiss.focus();
    expect(dismiss).toHaveFocus();

    // Ten intervals with focus held. A toast that vanishes while it is being
    // read is a new defect, so nothing may happen here at all.
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS * 10);
    expect(onDismiss).not.toHaveBeenCalled();

    dismiss.blur();
    expect(dismiss).not.toHaveFocus();

    // It RESUMES, it does not restart: two thirds of the interval were banked
    // before the pause, so two thirds is all that is left.
    await vi.advanceTimersByTimeAsync((TOAST_DISMISS_MS * 2) / 3 - 1);
    expect(onDismiss).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(onDismiss).toHaveBeenCalledExactlyOnceWith(failure.id);
  });

  it('pauses while the pointer is over it', async () => {
    const onDismiss = vi.fn();
    await renderFaked([failure], onDismiss);
    const panel = screen.getByRole('button').parentElement as HTMLElement;

    fireEvent.mouseOver(panel);
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS * 10);
    expect(onDismiss).not.toHaveBeenCalled();

    fireEvent.mouseOut(panel);
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS);
    expect(onDismiss).toHaveBeenCalledExactlyOnceWith(failure.id);
  });

  it('gives the reader the whole interval again when the toast repeats', async () => {
    const onDismiss = vi.fn();
    const { view } = await renderFaked([failure], onDismiss);

    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS - 1);

    // The store collapsed a second identical failure into this toast: same id,
    // count now 2. The panel just changed under the reader, so the clock starts
    // over rather than firing a millisecond later.
    view.rerender(
      <I18nextProvider i18n={await createI18n('en')}>
        <ToastList toasts={[{ ...failure, count: 2 }]} onDismiss={onDismiss} />
      </I18nextProvider>,
    );

    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS - 1);
    expect(onDismiss).not.toHaveBeenCalled();
    await vi.advanceTimersByTimeAsync(1);
    expect(onDismiss).toHaveBeenCalledExactlyOnceWith(failure.id);
  });

  it('leaves no timer behind when it is unmounted', async () => {
    const onDismiss = vi.fn();
    const { view } = await renderFaked([failure], onDismiss);

    expect(vi.getTimerCount()).toBe(1);
    view.unmount();

    // Not "it happens not to fire": there is no handle left at all.
    expect(vi.getTimerCount()).toBe(0);
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS * 10);
    expect(onDismiss).not.toHaveBeenCalled();
  });

  it('leaves no timer behind when the user dismisses it by hand', async () => {
    const onDismiss = vi.fn();
    const { view } = await renderFaked([failure], onDismiss);

    // The keyboard path is asserted above under real timers ("dismissable from
    // the keyboard alone"); this is the same activation, and what it is here to
    // assert is the handle afterwards.
    fireEvent.click(screen.getByRole('button'));
    expect(onDismiss).toHaveBeenCalledExactlyOnceWith(failure.id);

    // This component is presentational: the store is what removes the toast, so
    // the removal is what the rerender represents.
    view.rerender(
      <I18nextProvider i18n={await createI18n('en')}>
        <ToastList toasts={[]} onDismiss={onDismiss} />
      </I18nextProvider>,
    );

    expect(vi.getTimerCount()).toBe(0);
    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS * 10);
    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it('runs one independent timer per toast', async () => {
    const onDismiss = vi.fn();
    await renderFaked([failure, { ...failure, id: 2 }], onDismiss);

    expect(vi.getTimerCount()).toBe(2);

    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS);
    expect(onDismiss.mock.calls.map(([id]) => id)).toEqual([1, 2]);
  });
});

// --- D25: a refusal is not a failure, and every toast says what failed -------

type FakeGo = ReturnType<typeof createFakeClient>;
type FakeStore = ReturnType<typeof createAppStore>;

/** An error in no Go table: a failure, whichever action raised it. */
const boom = new Error('store: database is locked');

/** What each action's toast actually reads as. Filled by the sweep below. */
const seen = new Map<string, string>();

/**
 * Exactly the string Go sends when D9 declines to put a project into doing.
 *
 * Produced by service.Refuse; the message in front of the token is prose that
 * this side never reads. internal/service/refusal_test.go owns the other end of
 * this contract — it fails if the marker in store/call.ts stops matching the
 * one Go writes.
 */
const D9_REFUSAL = new Error(
  'service: move node "p-1": domain: a project never enters doing [nexus-refusal:project-never-doing]',
);

/** Drives a real store action against a rejecting client and renders the result. */
async function toastsFrom(
  language: string,
  arrange: (go: ReturnType<typeof createFakeClient>) => void,
  drive: (store: ReturnType<typeof createAppStore>) => Promise<unknown>,
) {
  const go = createFakeClient();
  arrange(go);
  const store = createAppStore(go.client, { view: window });

  await drive(store);

  const view = await renderToasts(language, store.getState().toasts);
  return { store, view };
}

describe('a refusal that crossed the boundary', () => {
  it('reads as the rule it is, in English, and not as a malfunction', async () => {
    const { store } = await toastsFrom(
      'en',
      (go) => go.reject('MoveToColumn', D9_REFUSAL),
      (store) => store.getState().moveToColumn('p-1', 'doing-from-go'),
    );

    const panel = screen.getByRole('alert');
    expect(panel).toHaveTextContent('A project never moves to Doing');
    expect(panel).toHaveTextContent('Moving the card');
    expect(panel).toHaveTextContent('A rule stopped this');
    // The point of the ticket: NOT the generic failure sentence.
    expect(panel.textContent).not.toContain('Nexus could not finish that');
    expect(panel.textContent).not.toContain('Something went wrong');
    expect(store.getState().toasts[0].messageKey).not.toBe('toast.error.body');
  });

  it('reads as the rule it is in Russian too, from the same stored toast', async () => {
    await toastsFrom(
      'ru',
      (go) => go.reject('MoveToColumn', D9_REFUSAL),
      (store) => store.getState().moveToColumn('p-1', 'doing-from-go'),
    );

    const panel = screen.getByRole('alert');
    expect(panel).toHaveTextContent('Проект не переходит');
    expect(panel).toHaveTextContent('Перемещение карточки');
    expect(panel.textContent).not.toContain('Nexus не смог выполнить это действие');
  });

  it('is not styled as an error — danger is reserved for things that broke', async () => {
    await toastsFrom(
      'en',
      (go) => go.reject('MoveToColumn', D9_REFUSAL),
      (store) => store.getState().moveToColumn('p-1', 'doing-from-go'),
    );

    const panel = screen.getByRole('alert').firstElementChild as HTMLElement;
    expect(panel.className).toContain('border-line');
    expect(panel.className).not.toContain('border-danger');
  });

  it('leaves an unclassified error exactly as it was: one generic failure', async () => {
    // S2's "every rejection reaches a toast" is not weakened, and a Go error in
    // no table does not get to pose as a rule.
    const { store } = await toastsFrom(
      'en',
      (go) => go.reject('MoveToColumn', new Error('store: database is locked')),
      (store) => store.getState().moveToColumn('p-1', 'doing-from-go'),
    );

    expect(store.getState().toasts).toHaveLength(1);
    expect(store.getState().toasts[0].kind).toBe('failure');
    const panel = screen.getByRole('alert');
    expect(panel).toHaveTextContent('Nexus could not finish that');
    expect(panel).toHaveTextContent('Moving the card');
    expect((panel.firstElementChild as HTMLElement).className).toContain('border-danger');
  });

  it('never renders Go’s own words, marker and all', async () => {
    await toastsFrom(
      'en',
      (go) => go.reject('MoveToColumn', D9_REFUSAL),
      (store) => store.getState().moveToColumn('p-1', 'doing-from-go'),
    );

    const text = screen.getByRole('alert').textContent ?? '';
    expect(text).not.toContain('nexus-refusal');
    expect(text).not.toContain('p-1');
    expect(text).not.toContain('service:');
  });
});

describe('every toast names the operation it came from', () => {
  // One failure per store action. The assertion is that the RENDERED sentences
  // are all different: a stack of three toasts is only legible if each says
  // which of the three things the user did went wrong, and after K12's
  // screenshot nobody could say which three they were.
  const actions: [string, (go: FakeGo) => void, (store: FakeStore) => Promise<unknown>][] = [
    ['loadBoard', (go) => go.reject('Board', boom), (s) => s.getState().loadBoard()],
    ['loadHabits', (go) => go.reject('HabitStrip', boom), (s) => s.getState().loadHabits()],
    ['loadTimer', (go) => go.reject('TimerCurrent', boom), (s) => s.getState().loadTimer()],
    ['loadSettings', (go) => go.reject('Settings', boom), (s) => s.getState().loadSettings()],
    [
      'move',
      (go) => go.reject('MoveToColumn', boom),
      (s) => s.getState().moveToColumn('n-1', 'from-go'),
    ],
    [
      'setPriority',
      (go) => go.reject('SetPriority', boom),
      (s) => s.getState().setPriority('n-1', 2),
    ],
    ['create', (go) => go.reject('CreateNode', boom), (s) => s.getState().createNode('t', 'task')],
    ['startTimer', (go) => go.reject('TimerStart', boom), (s) => s.getState().startTimer('n-1')],
    ['stopTimer', (go) => go.reject('TimerStop', boom), (s) => s.getState().stopTimer()],
    ['setPalette', (go) => go.reject('SetPalette', boom), (s) => s.getState().setPalette('x')],
    ['setTheme', (go) => go.reject('SetTheme', boom), (s) => s.getState().setTheme('x')],
    ['setAccent', (go) => go.reject('SetAccent', boom), (s) => s.getState().setAccent('x')],
    ['setLanguage', (go) => go.reject('SetLanguage', boom), (s) => s.getState().setLanguage('x')],
    [
      'checkHabit',
      (go) => go.reject('CheckHabitToday', boom),
      async (s) => {
        await s.getState().loadHabits();
        return s.getState().toggleHabit('habit-1');
      },
    ],
    [
      'reorder',
      (go) => go.reject('MoveNode', boom),
      async (s) => {
        await s.getState().loadBoard();
        const status = s.getState().board?.[0].status ?? '';
        return s
          .getState()
          .dropCard('node-1', { status, index: 0 }, { status, index: 1 });
      },
    ],
  ];

  it.each(actions)('%s says what it was doing', async (name, arrange, drive) => {
    const go = createFakeClient({
      habits: [habitView()],
      board: [columnView('first-from-go', [nodeView()]), columnView('second-from-go')],
    });
    arrange(go);
    const store = createAppStore(go.client, { view: window });

    await drive(store);

    const raised = store.getState().toasts.filter((toast) => toast.cause === boom);
    expect(raised.length, `${name} raised no toast`).toBeGreaterThan(0);
    await renderToasts('en', [raised[0]]);

    seen.set(name, screen.getByRole('alert').textContent ?? '');
    expect(seen.get(name)).not.toContain('toast.operation.');
  });

  it('gave every action a sentence of its own', () => {
    // Runs last, over what the sweep above recorded. Two actions sharing one
    // operation key would be two toasts nobody can tell apart.
    expect(seen.size).toBe(actions.length);
    expect(new Set(seen.values()).size, `duplicate wording: ${[...seen.values()].join(' | ')}`).toBe(
      actions.length,
    );
  });
});
