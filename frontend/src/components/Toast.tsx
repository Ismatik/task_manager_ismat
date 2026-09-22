import { useTranslation } from 'react-i18next';

import type { Toast } from '../store';

// Nexus — the error toast.
//
// PLAN.md section 1: no silent failures. This is where a rejection that crossed
// the Go boundary becomes something the user can see.
//
// # It renders a key, not a message
//
// The store holds an i18n key (store/toast.ts) and the translation happens
// here, at render time, in whatever language is current. The raw Go error never
// reaches this component: it went to the console when the toast was raised. A
// Go error string is untranslated, usually names a node id, and is not
// something a user can act on.
//
// # Presentational on purpose
//
// It takes its list and its dismiss handler as props rather than reaching into
// the store. That keeps it renderable in both languages, at any width, with any
// number of toasts, without standing up a store — and it is what lets the
// mounting point decide where it lives.
//
// Colours come from the token names and nowhere else: `danger` because
// design/README.md assigns danger to failure, `elevated`/`line`/`ink`/`muted`
// for the surface. Aurora's surfaces are translucent and `bg-elevated` is what
// makes this one translucent; `shadow-sm` is what gives Studio its edge.
//
// It does NOT carry Aurora's backdrop-filter utility (K7, D19). A toast is
// small, there can be several at once, and each one was its own compositing
// layer — the same reason the card and the habit chip lost it. D19's allow-list
// is the five columns and the two overlay scrims, and `make guard` check 8
// enforces exactly that list with an exact grep, which is why the class name is
// not written here. The translucency is unaffected: the token, not the filter,
// is what lets the background through.

export interface ToastListProps {
  toasts: readonly Toast[];
  onDismiss: (id: number) => void;
}

export function ToastList({ toasts, onDismiss }: ToastListProps) {
  const { t } = useTranslation();

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div
      // role="alert" and aria-live="assertive": a failure is the one thing a
      // screen reader user must not have to go looking for.
      role="alert"
      aria-live="assertive"
      className="fixed bottom-4 right-4 z-50 flex w-[min(24rem,calc(100vw-2rem))] flex-col gap-2"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="flex flex-col gap-2 rounded-md border border-danger bg-elevated p-3 text-ink shadow-sm"
        >
          <p className="text-danger">{t('toast.error.title')}</p>
          {/* Russian runs ~30% wider than English, so the text wraps rather
              than sitting on one line, and the panel is width-capped against
              the viewport rather than given a fixed pixel width. */}
          <p className="text-muted">{t(toast.messageKey)}</p>
          <button
            type="button"
            onClick={() => onDismiss(toast.id)}
            className="self-end rounded-sm px-2 py-1 text-ink transition-colors duration-fast hover:bg-surface"
          >
            {t('toast.error.dismiss')}
          </button>
        </div>
      ))}
    </div>
  );
}

export default ToastList;
