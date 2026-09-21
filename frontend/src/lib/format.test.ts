import { describe, expect, it } from 'vitest';

import { formatDate, formatDuration, formatNumber, formatTime } from './format';

describe('format', () => {
  it('formats a number in the active locale', () => {
    // Russian groups with a non-breaking space rather than a comma; asserting
    // the two are DIFFERENT is the locale-awareness claim, and asserting the
    // digits survive is the correctness one.
    const english = formatNumber(1234567, 'en');
    const russian = formatNumber(1234567, 'ru');

    expect(english).toBe('1,234,567');
    expect(russian).not.toBe(english);
    expect(russian.replace(/\D/g, '')).toBe('1234567');
  });

  it('formats a YYYY-MM-DD date without shifting the day', () => {
    // The whole reason formatDate splits the string by hand: new Date(
    // "2026-03-01") is UTC midnight, and rendering that in a timezone behind
    // Greenwich yields the 28th of February — a date nobody entered.
    expect(formatDate('2026-03-01', 'en')).toContain('01');
    expect(formatDate('2026-03-01', 'en')).toContain('2026');
    expect(formatDate('2026-03-01', 'en')).toMatch(/Mar/);

    const russian = formatDate('2026-03-01', 'ru');
    expect(russian).toContain('01');
    expect(russian).toContain('2026');
    expect(russian).not.toBe(formatDate('2026-03-01', 'en'));
  });

  it('formats the first and last day of a month without drifting', () => {
    expect(formatDate('2026-01-01', 'en')).toContain('2026');
    expect(formatDate('2025-12-31', 'en')).toContain('2025');
    expect(formatDate('2025-12-31', 'en')).toContain('31');
  });

  it('formats an RFC 3339 instant as a wall clock time', () => {
    const at = formatTime('2026-09-21T14:05:00Z', 'en');

    expect(at).toMatch(/\d{1,2}[:.]\d{2}/);
  });

  it('formats elapsed seconds as a clock readout', () => {
    expect(formatDuration(0, 'en')).toBe('0:00:00');
    expect(formatDuration(59, 'en')).toBe('0:00:59');
    expect(formatDuration(60, 'en')).toBe('0:01:00');
    expect(formatDuration(3661, 'en')).toBe('1:01:01');
    expect(formatDuration(86400 + 1, 'en')).toBe('24:00:01');
  });

  it('clamps a nonsensical duration rather than rendering a negative clock', () => {
    expect(formatDuration(-5, 'en')).toBe('0:00:00');
    expect(formatDuration(12.7, 'en')).toBe('0:00:12');
  });
});
