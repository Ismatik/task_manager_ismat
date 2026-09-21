import { describe, expect, it } from 'vitest';
import { render, screen } from '@testing-library/react';
import { I18nextProvider, useTranslation } from 'react-i18next';

import { changeLanguage, createI18n, fallbackLanguage, type LanguagePort } from './i18n';

// A component, not a bare `t()` call: the point of these cases is that the
// whole chain works — I18nextProvider, the react-i18next context, useTranslation
// and the bundled resources. Calling i18next directly would prove the library
// works, which was never in doubt.
function CardCount({ count }: { count: number }) {
  const { t } = useTranslation();

  return <p data-testid="count">{t('board.column.cardCount', { count })}</p>;
}

function AppName() {
  const { t } = useTranslation();

  return <h1 data-testid="name">{t('app.name')}</h1>;
}

/**
 * An in-memory stand-in for SettingsService: it stores what it is given and
 * reports the stored value back, which is the contract Go's SetLanguage has
 * (it returns the resulting SettingsView).
 */
function fakeSettings(initial: string) {
  const state = { language: initial };

  const port: LanguagePort = async (language) => {
    state.language = language;
    return state.language;
  };

  return { state, port };
}

describe('the i18n runtime', () => {
  it('renders Russian when the language is ru', async () => {
    const i18n = await createI18n('ru');

    render(
      <I18nextProvider i18n={i18n}>
        <CardCount count={3} />
      </I18nextProvider>,
    );

    expect(screen.getByTestId('count')).toHaveTextContent('3 карточки');
  });

  it('renders English when the language is en', async () => {
    const i18n = await createI18n('en');

    render(
      <I18nextProvider i18n={i18n}>
        <CardCount count={3} />
      </I18nextProvider>,
    );

    expect(screen.getByTestId('count')).toHaveTextContent('3 cards');
  });

  it('gives Russian three distinct plural forms across 1 / 3 / 5 / 21', async () => {
    const i18n = await createI18n('ru');

    const rendered = [1, 3, 5, 21].map((count) =>
      i18n.t('board.column.cardCount', { count }).replace(String(count), '').trim(),
    );

    const [one, few, many, twentyOne] = rendered;

    // Three distinct nouns, not four: 21 takes the SAME form as 1. That is the
    // case a hand-rolled `count === 1 ? a : b` gets wrong, and the concrete
    // reason i18next is a dependency rather than a lookup object.
    expect(new Set(rendered).size).toBe(3);
    expect(twentyOne).toBe(one);
    expect(few).not.toBe(one);
    expect(many).not.toBe(few);
    expect(many).not.toBe(one);
  });

  it('gives English two forms, and 21 is plural there', async () => {
    const i18n = await createI18n('en');

    expect(i18n.t('board.column.cardCount', { count: 1 })).toBe('1 card');
    expect(i18n.t('board.column.cardCount', { count: 3 })).toBe('3 cards');
    expect(i18n.t('board.column.cardCount', { count: 21 })).toBe('21 cards');
  });

  it('persists a language change, and survives a reload of the store', async () => {
    const settings = fakeSettings(fallbackLanguage);
    const i18n = await createI18n(settings.state.language);

    expect(i18n.language).toBe('en');

    await changeLanguage(i18n, 'ru', settings.port);

    expect(i18n.language).toBe('ru');
    expect(settings.state.language).toBe('ru');

    // The restart: throw the instance away and build a new one from what the
    // settings store now holds, exactly as startup does.
    const afterRestart = await createI18n(settings.state.language);

    expect(afterRestart.language).toBe('ru');

    render(
      <I18nextProvider i18n={afterRestart}>
        <CardCount count={5} />
      </I18nextProvider>,
    );

    expect(screen.getByTestId('count')).toHaveTextContent('5 карточек');
  });

  it("switches to the service's answer, not to the language it asked for", async () => {
    // A service that refuses to move: it keeps English whatever it is sent.
    const stubborn: LanguagePort = async () => 'en';
    const i18n = await createI18n('en');

    const accepted = await changeLanguage(i18n, 'ru', stubborn);

    expect(accepted).toBe('en');
    expect(i18n.language).toBe('en');
  });

  it('leaves the language alone when the write is rejected', async () => {
    const refusing: LanguagePort = async () => {
      throw new Error('unsupported language');
    };
    const i18n = await createI18n('en');

    await expect(changeLanguage(i18n, 'ru', refusing)).rejects.toThrow('unsupported language');
    expect(i18n.language).toBe('en');
  });

  it('is not configured with any loader that could reach the network', async () => {
    const i18n = await createI18n('en');

    // i18next only ever fetches through a backend module. There is none: the
    // resources are bundled, and `services.backendConnector.backend` stays null.
    const services = i18n.services as unknown as {
      backendConnector: { backend: unknown };
      languageDetector?: unknown;
    };

    expect(services.backendConnector.backend).toBeFalsy();
    expect(services.languageDetector).toBeFalsy();
  });

  it('keeps the two languages independent', async () => {
    // createInstance, not the default singleton: two instances at once, each on
    // its own language, is what a locale-parity test and an RU layout test need
    // in the same process.
    const enI18n = await createI18n('en');
    const ruI18n = await createI18n('ru');

    render(
      <I18nextProvider i18n={enI18n}>
        <AppName />
      </I18nextProvider>,
    );

    expect(enI18n.language).toBe('en');
    expect(ruI18n.language).toBe('ru');
    expect(screen.getByTestId('name')).toHaveTextContent('Nexus');
  });
});
