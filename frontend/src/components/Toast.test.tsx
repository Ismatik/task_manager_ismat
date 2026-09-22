import { afterEach, describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nextProvider } from 'react-i18next';

import { ToastList } from './Toast';
import { createI18n } from '../lib/i18n';
import { TOAST_DISMISS_MS } from '../store/toast';
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

const failure: Toast = { id: 1, messageKey: 'toast.error.body', count: 1 };

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
    await renderToasts('en', [{ id: 1, messageKey: 'toast.error.body', cause, count: 1 }]);

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
      [failure, { id: 2, messageKey: 'toast.error.body', count: 1 }],
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

const repeated: Toast = { id: 1, messageKey: 'toast.error.body', count: 3 };

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
    await renderToasts('ru', [{ id: 1, messageKey: 'toast.error.body', count: 5 }]);

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
    await renderFaked([failure, { id: 2, messageKey: 'toast.error.body', count: 1 }], onDismiss);

    expect(vi.getTimerCount()).toBe(2);

    await vi.advanceTimersByTimeAsync(TOAST_DISMISS_MS);
    expect(onDismiss.mock.calls.map(([id]) => id)).toEqual([1, 2]);
  });
});
