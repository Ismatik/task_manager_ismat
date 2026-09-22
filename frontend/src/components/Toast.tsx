import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';

import type { Toast } from '../store';
import { TOAST_DISMISS_MS } from '../store/toast';

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

// # It dismisses itself, and stops doing so while it is being read (D24)
//
// A toast that never leaves converts one failure into a permanently smaller
// window, so each one runs a TOAST_DISMISS_MS timer. The timer is a genuine
// PAUSE, not a restart: focusing the panel or putting the pointer on it banks
// the time remaining, and blurring or leaving resumes from there. A toast that
// vanishes while it is being read would be a new defect, and a toast that
// restarts its full interval every time the pointer crosses it is one an
// impatient pointer could keep alive forever.
//
// The timer lives here rather than in the store because "has focus" and "the
// pointer is over it" are DOM facts. The interval itself lives in
// store/toast.ts beside the cap, because the two are one policy.

interface ToastItemProps {
  toast: Toast;
  onDismiss: (id: number) => void;
}

function ToastItem({ toast, onDismiss }: ToastItemProps) {
  const { t } = useTranslation();
  const [held, setHeld] = useState(0);
  const remaining = useRef(TOAST_DISMISS_MS);
  const { id, count } = toast;

  // The one thing this component does with `kind`: choose a title and a border.
  // It does not decide what a refusal IS — that verdict arrived on the toast,
  // from the code Go sent (store/call.ts).
  const refusal = toast.kind === 'refusal';

  // A repeat resets the clock: the toast just said something new (the count
  // went up), so the reader gets the whole interval again. This runs BEFORE the
  // timer effect below on the same render, and after that effect's cleanup has
  // banked the elapsed time — React runs every cleanup first, then every effect,
  // each in declaration order — so the reset is what the timer then reads.
  useEffect(() => {
    remaining.current = TOAST_DISMISS_MS;
  }, [count]);

  useEffect(() => {
    if (held > 0) {
      return;
    }

    const startedAt = Date.now();
    const handle = window.setTimeout(() => onDismiss(id), remaining.current);

    return () => {
      window.clearTimeout(handle);
      // Cleared on every path there is: a re-render that pauses it, a manual
      // dismiss (which unmounts this item), and the unmount of the whole list.
      // Nothing is left holding a handle to a dismiss that cannot happen.
      const spent = Date.now() - startedAt;
      remaining.current = Math.max(0, remaining.current - spent);
    };
  }, [held, id, count, onDismiss]);

  // Focus and pointer are counted, not flagged: React's onFocus/onBlur bubble,
  // so tabbing from the panel to its own button fires a blur and a focus, and a
  // boolean would go false between them and fire the timer mid-read.
  const hold = () => setHeld((n) => n + 1);
  const release = () => setHeld((n) => Math.max(0, n - 1));

  return (
    <div
      onFocus={hold}
      onBlur={release}
      onMouseEnter={hold}
      onMouseLeave={release}
      className={`flex flex-col gap-2 rounded-md border bg-elevated p-3 text-ink shadow-sm ${
        // `danger` is reserved for things that BROKE (D25). A rule declining an
        // action is information — the app working — so it gets the ordinary
        // `line` border every other surface has, and no alarm colour at all.
        refusal ? 'border-line' : 'border-danger'
      }`}
    >
      <p className={refusal ? 'text-ink' : 'text-danger'}>
        {t(refusal ? 'toast.refusal.title' : 'toast.error.title')}
      </p>
      {/* WHAT the user was doing, always — the diagnostic half of D25. Three
          stacked toasts are legible only if each names its own operation. */}
      <p className="text-muted">{t(toast.operationKey)}</p>
      {/* Russian runs ~30% wider than English, so the text wraps rather
          than sitting on one line, and the panel is width-capped against
          the viewport rather than given a fixed pixel width. */}
      <p className="text-muted">{t(toast.messageKey)}</p>
      {/* A repeat is information, not noise: the count says so rather than the
          list quietly collapsing. font-mono because it is a number, and
          pluralised through i18next because Russian has three plural forms. */}
      {count > 1 ? <p className="font-mono text-muted">{t('toast.error.repeated', { count })}</p> : null}
      <button
        type="button"
        onClick={() => onDismiss(toast.id)}
        className="self-end rounded-sm px-2 py-1 text-ink transition-colors duration-fast hover:bg-surface"
      >
        {t('toast.error.dismiss')}
      </button>
    </div>
  );
}

export interface ToastListProps {
  toasts: readonly Toast[];
  onDismiss: (id: number) => void;
}

export function ToastList({ toasts, onDismiss }: ToastListProps) {
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
        <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
      ))}
    </div>
  );
}

export default ToastList;
