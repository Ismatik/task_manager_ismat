import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nextProvider } from 'react-i18next';

import { ToastList } from './Toast';
import { createI18n } from '../lib/i18n';
import type { Toast } from '../store';

async function renderToasts(
  language: string,
  toasts: readonly Toast[],
  onDismiss: (id: number) => void = () => {},
) {
  const i18n = await createI18n(language);

  render(
    <I18nextProvider i18n={i18n}>
      <ToastList toasts={toasts} onDismiss={onDismiss} />
    </I18nextProvider>,
  );
}

const failure: Toast = { id: 1, messageKey: 'toast.error.body' };

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
    await renderToasts('en', [{ id: 1, messageKey: 'toast.error.body', cause }]);

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
      [failure, { id: 2, messageKey: 'toast.error.body' }],
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
