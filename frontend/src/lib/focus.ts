// Nexus — the one way anything in this application moves focus.
//
// # The rule, and why it gets exactly one spelling
//
// D22: **no `.focus()` call anywhere may scroll its ancestors.** Before S3-01
// there were eleven of them, in five files, and every single one omitted
// `{ preventScroll: true }` — so every roving-tabindex step, every overlay open
// and every focus restore dragged the board (and, before the height chain, the
// whole document) to wherever the browser felt the element ought to be visible.
//
// A rule applied eleven times by hand is a rule that will be applied ten times
// after the next ticket. So it is applied once, here, and `make guard` check 7
// refuses `.focus(` in any non-test module but this one. That check is an EXACT
// grep rather than one of guard's name-based heuristics: there is no judgement
// in "does this line contain `.focus(`", which is what makes it fit to be a
// guard check at all (D17).
//
// # Why it takes a nullable
//
// Most call sites are `ref.current?.focus()`, `opener?.focus()` or
// `querySelector(...)?.focus()`. If the helper demanded a non-null element,
// each of those would keep its own `?.` — and a `?.focus(` is still a `.focus(`,
// so check 7 would have nothing left to catch. Accepting null here is what lets
// the call sites be plain calls.
//
// # What this does NOT do
//
// It does not decide WHICH element gets focus, does not remember a previous
// one, and does not restore anything. Those are the callers' business and they
// each have their own reason; this module owns one property of the act itself.

/**
 * Moves focus to `element` without scrolling any ancestor to reveal it.
 *
 * A nullish `element` is a no-op: "there was nothing to focus" is a normal
 * outcome for a ref that has not attached yet, an opener that has gone away, or
 * a query that matched nothing.
 */
export function focusWithoutScrolling(element: HTMLElement | null | undefined): void {
  // `preventScroll` is honoured by WebKit and ignored by jsdom, which has no
  // layout and therefore never scrolls anything. That asymmetry is why the test
  // beside this file asserts the ARGUMENT rather than an absence of scrolling:
  // what can be proven here is that the option is passed on every call, and
  // that the browser then obeys it is the hand pass's (S3-09 item 1).
  element?.focus({ preventScroll: true });
}

export default focusWithoutScrolling;
