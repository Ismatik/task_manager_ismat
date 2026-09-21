import { describe, expect, it } from 'vitest';

import { optional, unwrappedBindingNames, wailsClient } from './client';

describe('the Go client', () => {
  it('wraps EVERY generated binding', () => {
    // The claim "lib/client.ts wraps every binding" is otherwise one nobody
    // checks again. A method added to app.go regenerates App.js on the next
    // `wails build`, and without this the frontend would simply not know it
    // existed — which is how a binding ends up called from three places
    // directly instead of one place here.
    expect(unwrappedBindingNames()).toEqual([]);
  });

  it('exposes the bindings by reference, so the coverage check means something', () => {
    // If the methods were arrow wrappers, unwrappedBindingNames() would report
    // every binding as unwrapped and the case above would be impossible to
    // pass. Asserting they are functions is the cheap proof that the shape the
    // check depends on is the shape that exists.
    for (const [name, method] of Object.entries(wailsClient)) {
      expect(typeof method, `${name} is not callable`).toBe('function');
    }
  });

  it('covers the whole surface app.go binds', () => {
    // Named explicitly rather than counted, so that a binding DELETED in Go is
    // as visible in the diff as one added.
    expect(Object.keys(wailsClient).sort()).toEqual([
      'ArchiveNode',
      'Board',
      'CheckHabit',
      'CreateNode',
      'HabitStrip',
      'MoveNode',
      'MoveToColumn',
      'Progress',
      'RestoreNode',
      'Search',
      'SetAccent',
      'SetDue',
      'SetLanguage',
      'SetPalette',
      'SetTheme',
      'Settings',
      'TimerCurrent',
      'TimerStart',
      'TimerStop',
      'Tree',
      'UncheckHabit',
    ]);
  });
});

describe('optional', () => {
  // Go sends null; the generator declares the field as `?`, which is
  // undefined. Both arrive in practice, and `x === null` is false for one of
  // them. This is why every consumer reads nullable fields through here.
  it('reads null and undefined the same way', () => {
    expect(optional(null)).toBeNull();
    expect(optional(undefined)).toBeNull();
  });

  it('passes a real value through untouched, including falsy ones', () => {
    expect(optional('2026-09-21')).toBe('2026-09-21');
    expect(optional(0)).toBe(0);
    expect(optional('')).toBe('');
    expect(optional(false)).toBe(false);
  });
});
