import type { StateCreator } from 'zustand';

import {
  optional,
  type ColumnView,
  type HabitView,
  type NewNode,
  type Node as StoredNode,
  type TimerView,
} from '../lib/client';
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
// There are exactly two bounded exceptions, both written down in TASKS.md so
// nobody has to guess at them: S2-17's optimistic move, which relocates a card
// between columns locally, and S2-18's optimistic habit check, which ticks
// `checkedToday` locally. Both revert FROM GO'S ANSWER on error — a re-read,
// never a remembered snapshot, which can be stale if anything else changed.
// That is optimism about a value Go is ABOUT to compute, not a recomputation of
// it. Everything else re-reads.
//
// Note what the optimistic tick does NOT touch: `streak`. The streak is D5's
// answer — consecutive scheduled RRULE occurrences, not calendar days — and
// guessing at it locally would be the rule written a second time. The checkbox
// moves at once; the number moves when Go says so.
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
   * Ticks or un-ticks today's check for a habit, and re-reads the strip.
   *
   * Which way round it goes is `checkedToday` off GO'S OWN STRIP — the frontend
   * does not track a check of its own, and does not decide whether the current
   * day is one the schedule names. Returns whether the strip changed, false
   * for an unknown habit and for a refusal; callGo has already raised the one
   * toast in the second case.
   */
  toggleHabit(nodeId: string): Promise<boolean>;

  /**
   * Creates a root node from a title and a type, and re-reads what it lands in.
   *
   * Returns the node Go stored, or null when Go refused — an empty title, a
   * habit with no recurrence rule, a type that does not exist. Every one of
   * those is a rule domain.Node.Validate and domain.CheckStatus already own, so
   * NOTHING IS PRE-VALIDATED HERE: pre-validating is writing the rule a second
   * time, and the second copy is the one that goes stale. callGo has raised the
   * one toast by the time this returns null.
   *
   * Everything the draft does not carry is left at its ZERO VALUE on purpose,
   * because the defaults are Go's: TaskService.CreateNode reads an empty status
   * as backlog, a zero priority as 4, and a nil activity as D4's default for
   * the type. A frontend that filled those in would be inventing defaults and
   * would disagree with Go the first time one of them changed.
   */
  createNode(title: string, type: string): Promise<StoredNode | null>;

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

/**
 * The calendar day at `at`, written the way domain.Date marshals — "YYYY-MM-DD".
 *
 * CheckHabit and UncheckHabit take the day as a parameter, so somebody has to
 * name it, and no binding reports Go's idea of today. This is that name and
 * nothing more: it is the LOCAL calendar day off the store's injected clock,
 * the same wall clock Go's own `time.Now()` reads, and it decides nothing.
 * Whether that day is an occurrence of the habit's recurrence, whether the
 * check breaks a streak and whether the day may be checked at all are all Go's
 * answers, asked by sending it this string.
 *
 * Built out of the local getters rather than `toISOString()`, which is UTC and
 * would tick over to tomorrow at 03:00 for a user in Almaty — a day the user
 * never chose, produced by a timezone they never mentioned. lib/format.ts makes
 * the same point about reading one.
 */
/**
 * A copy of `habit` with today's check flipped, and nothing else touched.
 *
 * The cast is the generator quirk lib/client.ts documents, in its second form.
 * models.ts declares HabitView as a CLASS carrying a `convertValues` method,
 * while the wire sends — and the store therefore only ever holds — a plain JSON
 * object with no such method. A spread of one is structurally everything a
 * HabitView is and is still not assignable to it. test/fakeClient.ts casts for
 * the same reason, and the answer is not to stop using the generated types:
 * every field name written here is still checked against Go's DTO by gate 4.
 */
function withCheckFlipped(habit: HabitView): HabitView {
  return { ...habit, checkedToday: !habit.checkedToday } as HabitView;
}

function wireDate(at: number): string {
  const local = new Date(at);
  const pad = (part: number) => String(part).padStart(2, '0');

  return `${local.getFullYear()}-${pad(local.getMonth() + 1)}-${pad(local.getDate())}`;
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

  async toggleHabit(nodeId) {
    const { habits } = get();
    if (habits === null) {
      return false;
    }

    const before = habits.find((habit) => habit.node.id === nodeId);
    if (before === undefined) {
      // A habit the strip does not hold. No call, no toast: nothing happened.
      return false;
    }

    const day = wireDate(get().now());

    // The bounded optimism (TASKS.md S2-18). The box ticks on the keystroke
    // rather than a round trip later, because a checkbox that lags is a
    // checkbox the user presses twice. Only `checkedToday` moves — the streak
    // is D5's and waits for Go.
    set({
      habits: habits.map((habit) =>
        habit.node.id === nodeId ? withCheckFlipped(habit) : habit,
      ),
    });

    const answer = await callGo(get(), () =>
      before.checkedToday
        ? get().client.UncheckHabit(nodeId, day)
        : get().client.CheckHabit(nodeId, day),
    );

    if (answer === null) {
      // Refused. Revert FROM GO'S ANSWER rather than from the array captured
      // above: a snapshot is only right if nothing else changed in the
      // meantime, and "nothing else changed" is not something this can know.
      // callGo has already raised the one toast.
      await get().loadHabits();
      return false;
    }

    // CheckHabit and UncheckHabit both return the whole strip as it now is, so
    // this IS the re-read — the streak arrives recomputed with it.
    set({ habits: answer });
    return true;
  },

  async createNode(title, type) {
    // The zero values are the point — see the interface above. The cast is the
    // same generator quirk withCheckFlipped explains: NewNode is generated as a
    // class, the wire takes a plain object, and gate 4 still checks every field
    // name here against Go's struct.
    const draft = {
      parentId: undefined,
      type,
      title,
      descriptionMd: '',
      status: '',
      due: undefined,
      priority: 0,
      estimateMin: undefined,
      recurrence: undefined,
      activity: undefined,
    } as NewNode;

    const created = await callGo(get(), () => get().client.CreateNode(draft));
    if (created === null) {
      return null;
    }

    // Selection first, so that when the board arrives the roving tabindex is
    // already on the new card rather than a frame behind it.
    get().select(created.id);

    // Both reads, because which of the two the node lands in is Go's answer,
    // not one this could work out: a habit goes to the strip and never to a
    // column (PLAN.md section 4), everything else goes to a column. Asking for
    // both is one extra call; deciding between them here would be the rule.
    await Promise.all([get().loadBoard(), get().loadHabits()]);
    return created;
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
