import type { StateCreator } from 'zustand';

import { optional, type ColumnView, type HabitView, type TimerView } from '../lib/client';
import { adjacentColumn } from '../lib/keyboard';
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
   * Moves a card one column left (-1) or right (+1), and re-reads the board.
   *
   * Returns whether the board moved, which is what the caller needs to know to
   * decide whether to put focus back on the card. False covers both the edges
   * — right from the last column, left from the first — and a refusal from Go,
   * and the two are deliberately not distinguished here: neither one changed
   * anything, and only one of them raised a toast, which callGo already did.
   *
   * The target is `column.status` off GO'S OWN BOARD. Nothing computes a "next
   * status": the frontend knows the columns' order because Go returned them in
   * order, and knows nothing at all about their names.
   */
  moveToAdjacentColumn(nodeId: string, offset: number): Promise<boolean>;

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

  async moveToAdjacentColumn(nodeId, offset) {
    const { board } = get();
    if (board === null) {
      return false;
    }

    const target = adjacentColumn(board, nodeId, offset);
    if (target === null) {
      // The edges: no call, no toast, nothing at all. Not an error and not a
      // wrap-around.
      return false;
    }

    if ((await callGo(get(), () => get().client.MoveToColumn(nodeId, target.status))) === null) {
      // Refused — a project to doing (D9), say. callGo already raised the one
      // toast, and the board is untouched, so the card has not moved and
      // focus is still on it.
      return false;
    }

    // Re-read rather than patch. Go's answer is the truth: the move rewrites a
    // due date (D8) and cascades to the subtree (D2), and a local edit that
    // tried to keep up would be a second implementation of both.
    await get().loadBoard();
    return true;
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
