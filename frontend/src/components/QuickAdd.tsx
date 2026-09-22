import {
  useEffect,
  useRef,
  useState,
  type FormEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type MutableRefObject,
} from 'react';
import { useTranslation } from 'react-i18next';

import { focusWithoutScrolling } from '../lib/focus';
import { resources } from '../lib/i18n';
import { boardActionFor } from '../lib/keyboard';
import { useAppState, useAppStore } from '../store/context';

// Nexus — the in-app quick add, on S2-16's Ctrl+N.
//
// IN-APP ONLY. The frameless standalone quick-add WINDOW and the
// natural-language parser are Stage 4 (internal/parse is a skeleton), and
// neither is started here. This takes a title and a type and creates a node.
//
// # It validates nothing, and that is the feature
//
// An empty title, a habit with no recurrence rule, a type that does not exist:
// every one of those is refused by domain.Node.Validate or domain.CheckStatus,
// in Go, in the same transaction that would have written the row. So the
// overlay sends the draft and SURFACES THE REFUSAL IN A TOAST. A pre-check here
// would be the rule written a second time — and the second copy is the one that
// goes stale, because nobody edits it when the first one changes. The overlay
// stays open on a refusal, with the title still in the box, which is what makes
// the toast actionable.
//
// The defaults are Go's too. store/data.ts sends the rest of NewNode at its
// zero value, and TaskService.CreateNode reads an empty status as backlog, a
// zero priority as 4 and a nil activity as D4's default for the type.
//
// # Where the list of types comes from
//
// From the LABEL TABLE in locales/, and nowhere else. S2-15 settled the
// principle for the column headings and it is the same principle here: "a label
// table is presentation; a status list in code is a rule". No binding
// enumerates domain.NodeType — that would be a Go change, and Go files are
// outside this ticket's scope — so the offered list is the keys of `card.type`,
// in the order they are written there, which is the order the ticket names:
// task, project, habit, note, bug. The first is the default, so the type a
// quick add creates by default is a fact about that table rather than a string
// somebody typed into a component. Anything Go does not recognise is refused on
// arrival, so this list decides what is OFFERED and never what is valid.
//
// # Two components, one file, and why the panel is separate
//
// The outer component is always mounted, because it owns the one thing that has
// to outlive the overlay: putting focus on the card that was just created,
// which happens after the panel is gone. The panel is mounted only while the
// overlay is open, which is what makes its title and type fields START EMPTY
// every time without an effect that resets them — a fresh mount is a fresh
// useState, and "reopen it and yesterday's half-typed title is still there" is
// a bug nobody would have written on purpose.

/**
 * The one entry of `card.type` that is not a type.
 *
 * TypeIcon.tsx renders it for a node whose type this build has no name for. It
 * is a presentation fallback, so it is not something a user may choose.
 */
const FALLBACK_LABEL_KEY = 'unknown';

const OFFERED_TYPES = Object.keys(resources.en.translation.card.type).filter(
  (key) => key !== FALLBACK_LABEL_KEY,
);

const DEFAULT_TYPE = OFFERED_TYPES[0];

/**
 * Everything inside `root` that Tab would stop on.
 *
 * `tabIndex >= 0` rather than a selector listing every focusable tag: the type
 * buttons are a radio group with a roving tabindex, so four of the five are
 * deliberately -1 and must not be cycled to.
 */
function focusStopsWithin(root: HTMLElement): HTMLElement[] {
  return [...root.querySelectorAll<HTMLElement>('input, button, [tabindex]')].filter(
    (element) => element.tabIndex >= 0,
  );
}

export function QuickAdd() {
  const store = useAppStore();
  const open = useAppState((state) => state.openOverlay) === 'quickAdd';
  const board = useAppState((state) => state.board);

  const [createdId, setCreatedId] = useState<string | null>(null);
  const alreadyFocused = useRef<string | null>(null);
  // False for exactly one close: the one that follows a successful create,
  // after which focus belongs on the NEW CARD and not back where it came from.
  const restoreFocus = useRef(true);

  // Focus the card that was just created.
  //
  // `board` is in the dependency list because that is what makes this
  // deterministic rather than a guess about timing: `createNode` re-reads
  // before it resolves, so the new card is already in the store — and therefore
  // already committed — by the time this runs.
  //
  // The scan goes through `data-node-id`, which Card.tsx publishes as "the ONLY
  // channel" anything uses to address a card. Using the published channel is
  // not the same as reimplementing the board's focus model: this asks for one
  // card by id and has no opinion about which card is next, which is the
  // board's business and stays there. A habit has no card, so finding nothing
  // is a legitimate outcome and not an error.
  //
  // `alreadyFocused` is a ref rather than a second piece of state so that this
  // runs once per created node without writing state from inside an effect.
  useEffect(() => {
    if (createdId === null || alreadyFocused.current === createdId) {
      return;
    }
    alreadyFocused.current = createdId;

    for (const card of document.querySelectorAll<HTMLElement>('[data-node-id]')) {
      if (card.dataset.nodeId === createdId) {
        focusWithoutScrolling(card);
        return;
      }
    }
  }, [createdId, board]);

  if (!open) {
    // Region 4 of the shell emits no DOM while nothing is open. An overlay that
    // rendered an invisible box would still be in the tab order.
    return null;
  }

  const finish = (nodeId: string) => {
    restoreFocus.current = false;
    store.getState().closeOverlay();
    setCreatedId(nodeId);
  };

  return <QuickAddPanel restoreFocus={restoreFocus} onCreated={finish} />;
}

interface QuickAddPanelProps {
  /** Cleared by the parent for the one close that must not restore focus. */
  restoreFocus: MutableRefObject<boolean>;
  onCreated: (nodeId: string) => void;
}

function QuickAddPanel({ restoreFocus, onCreated }: QuickAddPanelProps) {
  const { t } = useTranslation();
  const store = useAppStore();

  const [title, setTitle] = useState('');
  const [nodeType, setNodeType] = useState(DEFAULT_TYPE);
  const titleRef = useRef<HTMLInputElement>(null);

  // Open and close, as one effect and its cleanup.
  //
  // Escape is not handled here. The shell already listens on the document for
  // the whole global map (S2-16) and closes the topmost overlay, so a second
  // Escape handler in this file would be a second binding for one key.
  useEffect(() => {
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;

    focusWithoutScrolling(titleRef.current);

    return () => {
      if (restoreFocus.current) {
        focusWithoutScrolling(opener);
      }
      restoreFocus.current = true;
    };
  }, [restoreFocus]);

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const created = await store.getState().createNode(title, nodeType);
    if (created === null) {
      // Refused. The toast is already up, the overlay stays open and the title
      // stays in the box, so the user can fix it rather than retype it.
      return;
    }
    onCreated(created.id);
  };

  /** Tab and Shift+Tab, wrapped — the overlay is modal, so focus cannot leave it. */
  const trapFocus = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key !== 'Tab') {
      return;
    }

    const stops = focusStopsWithin(event.currentTarget);
    if (stops.length === 0) {
      return;
    }

    event.preventDefault();

    const at = stops.indexOf(document.activeElement as HTMLElement);
    const step = event.shiftKey ? -1 : 1;
    focusWithoutScrolling(stops[(at + step + stops.length) % stops.length]);
  };

  /**
   * The arrows inside the type group.
   *
   * The chords are read out of lib/keyboard.ts rather than written again: the
   * group is a single row, so the map's adjacent-column chord is the same
   * physical movement in it, exactly as in the habits strip. Selection follows
   * focus, which is the ARIA radio-group pattern.
   */
  const chooseTypeWithArrows = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    const action = boardActionFor(event.nativeEvent);
    if (action !== 'nextColumn' && action !== 'previousColumn') {
      return;
    }

    event.preventDefault();

    const at = OFFERED_TYPES.indexOf(nodeType);
    const step = action === 'nextColumn' ? 1 : -1;
    const next = OFFERED_TYPES[(at + step + OFFERED_TYPES.length) % OFFERED_TYPES.length];

    setNodeType(next);
    focusWithoutScrolling(event.currentTarget.querySelector<HTMLElement>(`[data-type="${next}"]`));
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('quickAdd.label')}
      onKeyDown={trapFocus}
      // `bg-bg` plain and not a token with an opacity modifier: the palette
      // colours resolve through bare CSS variables (design/tailwind.config.js)
      // that carry no <alpha-value> placeholder, so `bg-bg/70` would render
      // fully opaque anyway. An honest solid scrim beats a transparency that
      // only exists in the class name.
      className="fixed inset-0 z-40 flex items-start justify-center bg-bg p-4 backdrop-blur-glass"
    >
      <form
        onSubmit={(event) => void submit(event)}
        className="flex w-[min(32rem,100%)] flex-col gap-2 rounded-lg border border-line bg-elevated p-4 text-ink shadow-sm"
      >
        <input
          ref={titleRef}
          value={title}
          onChange={(event) => setTitle(event.target.value)}
          aria-label={t('quickAdd.title')}
          // No `required`, no minlength, no trim-and-refuse. An empty title is
          // domain.Node.Validate's refusal and arrives as a toast.
          className="min-w-0 rounded-sm border border-line bg-surface px-2 py-1 text-ink"
        />

        <div
          role="radiogroup"
          aria-label={t('quickAdd.type')}
          onKeyDown={chooseTypeWithArrows}
          // Wrapping rather than a fixed row: the Russian labels are about a
          // third wider and must not be pushed off the edge of the panel.
          className="flex min-w-0 flex-wrap items-center gap-2"
        >
          {OFFERED_TYPES.map((value) => (
            <button
              key={value}
              type="button"
              role="radio"
              data-type={value}
              aria-checked={value === nodeType}
              // Roving tabindex: the group is one tab stop and the arrows move
              // inside it. Five extra tab stops in a two-field overlay would
              // make Tab useless.
              tabIndex={value === nodeType ? 0 : -1}
              onClick={() => setNodeType(value)}
              className={`min-w-0 break-words rounded-sm border px-2 py-1 transition-colors duration-fast ${
                value === nodeType
                  ? 'border-accent bg-accent text-on-accent'
                  : 'border-line bg-surface text-muted'
              }`}
            >
              {t(`card.type.${value}`)}
            </button>
          ))}
        </div>

        <button
          type="submit"
          className="min-w-0 self-end break-words rounded-sm border border-accent bg-accent px-2 py-1 text-on-accent transition-colors duration-fast"
        >
          {t('quickAdd.submit')}
        </button>
      </form>
    </div>
  );
}

export default QuickAdd;
