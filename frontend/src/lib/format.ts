// Nexus — the one module that turns a value into the text of a number, a date
// or a duration.
//
// # Formatting is presentation. The values are still Go's.
//
// Nothing here decides anything. It receives a number Go computed, a
// "YYYY-MM-DD" a domain.Date marshalled, an RFC 3339 instant or an
// elapsedSeconds off a TimerView, and produces the string a human reads. No
// arithmetic on dates, no comparison against the current day, no percentage —
// those are derivations, and derivations live in internal/domain.
//
// # Why one module
//
// Two places that format a date are two date formats, and the one on screen
// would be the untested one. Every component imports from here; nothing calls
// Intl directly. Everything produced here is rendered in `font-mono`
// (PLAN.md section 3: JetBrains Mono for every number, date, timer and shortcut
// hint) — that is the caller's class name, but it is the reason this module
// exists as a set of small, uniform helpers.
//
// The locale is passed in rather than read from a module-level i18next
// singleton, so a test can format both languages in one process and a component
// can only ever format in the language it is actually rendering in.

/** Formats an integer or a decimal in the active locale (1234 -> "1,234" / "1 234"). */
export function formatNumber(value: number, locale: string): string {
  return new Intl.NumberFormat(locale).format(value);
}

/**
 * Formats a calendar date that arrived as the "YYYY-MM-DD" string domain.Date
 * marshals to.
 *
 * The parts are split and handed to the Date constructor as LOCAL year, month
 * and day. `new Date("2026-03-01")` would parse as UTC midnight and then render
 * as the 28th of February for anybody west of Greenwich — a date the user never
 * typed, produced by a timezone they never mentioned.
 */
export function formatDate(isoDate: string, locale: string): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  const local = new Date(year, month - 1, day);

  return new Intl.DateTimeFormat(locale, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  }).format(local);
}

// A time-of-day and a duration formatter used to live here, for a running timer
// nothing renders yet. They were deleted with the store helpers that fed them
// (S2-18): an unused formatter is harmless, but the elapsed-seconds arithmetic
// that called one was not, and a formatter kept "for when it is wired up" is an
// invitation to wire it up to a locally computed number. Stage 3 draws the
// timer; it can add exactly the formatter it renders.
