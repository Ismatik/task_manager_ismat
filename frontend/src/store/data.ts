import type { StateCreator } from 'zustand';

import { optional, type ColumnView, type HabitView, type TimerView } from '../lib/client';
import { callGo } from './call';
import type { AppState } from './index';

// Nexus — the domain state: the board, the habit strip and the timer, held
// EXACTLY as Go returned them.
//
// # Nothing in this file computes a derived value
//
// Not a status, not a column, not progress, not a percentage, not an overdue
// flag, not a streak, not a due date. Every one of those is already a field on
// the DTO (internal/service/dto.go), computed once, in Go, against the injected
// clock. If a screen needs a value that is not here, the answer is a Go change.
//
// The store also never writes a domain field to a value Go did not produce.
// There is exactly one bounded exception, written down in TASKS.md S2-13 so
// nobody has to guess at it: S2-17's optimistic move, which relocates a card
// between columns locally and reverts from Go's answer on error. That is
// optimism about a value Go is ABOUT to compute, not a recomputation of it.
// Everything else re-reads.
//
// # The one thing that is allowed to tick locally
//
// dto.go authorises it in so many words: "the UI ticks its own display between
// reads rather than polling Go every second". So the DISPLAY advances between
// reads — but the number it starts from is Go's elapsedSeconds, and
// TimerView.elapsedSeconds itself is never written to. See
// displayElapsedSeconds below.

export interface DataSlice {
  /** The five columns, in order, or null before the first read. */
  board: ColumnView[] | null;
  /** Every non-archived habit with its schedule, today's check and its streak. */
  habits: HabitView[] | null;
  /** The single global timer, as Go last reported it. */
  timer: TimerView | null;
  /** Local wall-clock milliseconds at which `timer` was read. */
  timerReadAt: number | null;

  loadBoard(): Promise<ColumnView[] | null>;
  loadHabits(): Promise<HabitView[] | null>;
  loadTimer(): Promise<TimerView | null>;

  /** Reads settings, board, strip and timer. Never throws. */
  hydrate(): Promise<void>;

  /**
   * The elapsed seconds to PUT ON SCREEN at wall-clock time `now`.
   *
   * Go's number plus the time since it was read, and only while the timer is
   * running. Stopped, it is Go's number unchanged. This is a display value and
   * is never written back into `timer`.
   */
  displayElapsedSeconds(now?: number): number;

  /** The instant the running timer started, or null. */
  timerStartedAt(): string | null;
}

export const createDataSlice: StateCreator<AppState, [], [], DataSlice> = (set, get) => ({
  board: null,
  habits: null,
  timer: null,
  timerReadAt: null,

  async loadBoard() {
    const board = await callGo(get(), () => get().client.Board());
    if (board === null) {
      return null;
    }

    set({ board });
    return board;
  },

  async loadHabits() {
    const habits = await callGo(get(), () => get().client.HabitStrip());
    if (habits === null) {
      return null;
    }

    set({ habits });
    return habits;
  },

  async loadTimer() {
    const timer = await callGo(get(), () => get().client.TimerCurrent());
    if (timer === null) {
      return null;
    }

    set({ timer, timerReadAt: get().now() });
    return timer;
  },

  async hydrate() {
    await get().loadSettings();
    await Promise.all([get().loadBoard(), get().loadHabits(), get().loadTimer()]);
  },

  displayElapsedSeconds(now = get().now()) {
    const { timer, timerReadAt } = get();

    if (timer === null) {
      return 0;
    }
    if (!timer.running || timerReadAt === null) {
      return timer.elapsedSeconds;
    }

    return timer.elapsedSeconds + Math.max(0, Math.floor((now - timerReadAt) / 1000));
  },

  timerStartedAt() {
    // `optional` because the generator declares this field as `?` while the
    // wire sends null — see lib/client.ts. Reading it any other way works until
    // the one day it does not.
    return optional(get().timer?.startedAt);
  },
});
