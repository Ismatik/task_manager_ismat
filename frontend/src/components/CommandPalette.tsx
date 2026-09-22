import {
  useEffect,
  useId,
  useRef,
  useState,
  type KeyboardEvent as ReactKeyboardEvent,
} from 'react';
import { useTranslation } from 'react-i18next';

import { buildCommands, matchingCommands, type Command } from '../lib/commands';
import { focusWithoutScrolling } from '../lib/focus';
import { boardActionFor } from '../lib/keyboard';
import { useAppState, useAppStore } from '../store/context';

// Nexus — the command palette, on S2-16's Ctrl+K.
//
// # It decides nothing about what the palette can do
//
// The row list is `lib/commands.ts`'s answer and this file filters, renders and
// runs it. That is not decoration: the actions are a pure function of the state
// Go handed us, and keeping them pure is what lets a test hold the list still
// and read it. Everything visual is here; nothing else is.
//
// Every row's `run` is a store call — the SAME store call the board, the strip
// and the quick add already make. A palette that reached for lib/client.ts
// would be a second implementation of every action it offers, with its own copy
// of the re-read, the toast and the focus handling, and the two would disagree
// on the first change to either.
//
// # The keyboard, and where its keys come from
//
//   Ctrl+K      opens it — the shell's document listener (App.tsx), not here
//   typing      filters, on the TRANSLATED label, so it works in Russian
//   Up / Down   move the active row, through lib/keyboard.ts's own chords
//   Enter       runs the active row
//   Escape      closes it — again the shell's, and again not a second binding
//
// `Home` and `End` are deliberately NOT bound, and this is the one place the
// board's map is not followed to the letter. Focus lives in a text field here,
// where Home and End move the caret; taking them would break the ordinary
// editing of the query to save one keystroke of list navigation. Arrow Left and
// Right are left alone for exactly the same reason. What IS taken comes out of
// `boardActionFor`, so the chords themselves still have one spelling.
//
// The hints on the right of each row are `KEYS` values, rendered in `font-mono`
// (PLAN.md §3). Not one of them is written down in this file, and a test
// compares every hint the palette renders against lib/keyboard.ts's table.
//
// # Two components, one file — the same shape as QuickAdd
//
// The outer component is what the shell mounts; the panel is mounted only while
// the overlay is open, which is what makes the query field START EMPTY every
// time without an effect that clears it. A fresh mount is a fresh useState, and
// "reopen it and yesterday's half-typed query is still there" is a bug nobody
// would have written on purpose.

export function CommandPalette() {
  const open = useAppState((state) => state.openOverlay) === 'commandPalette';

  if (!open) {
    // Region 4 of the shell emits no DOM while nothing is open. An overlay that
    // rendered an invisible box would still be in the tab order.
    return null;
  }
  return <CommandPalettePanel />;
}

function CommandPalettePanel() {
  const { t, i18n } = useTranslation();
  const store = useAppStore();

  const board = useAppState((state) => state.board);
  const selectedNodeId = useAppState((state) => state.selectedNodeId);
  const settings = useAppState((state) => state.settings);

  const [query, setQuery] = useState('');
  const [active, setActive] = useState(0);
  const filterRef = useRef<HTMLInputElement>(null);

  // Restore focus to whatever opened the palette, on every close.
  //
  // Escape is not handled here: the shell listens on the document for the whole
  // global map (S2-16) and closes the topmost overlay, so a second Escape
  // handler in this file would be a second binding for one key. The cleanup
  // below runs whichever way the palette closed.
  //
  // Running "new task" is the one close that ends somewhere else, and it needs
  // no exception: React flushes this cleanup before the quick-add panel's own
  // effect, which then takes the focus for its title field. Last writer wins,
  // and the last writer is the overlay that is still on screen.
  useEffect(() => {
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;

    focusWithoutScrolling(filterRef.current);

    return () => focusWithoutScrolling(opener);
  }, []);

  const commands = buildCommands({ board, selectedNodeId, settings, store, t, i18n });
  const visible = matchingCommands(commands, query, i18n.language);

  // Clamped at render rather than corrected in an effect: the list shrinks as
  // the user types and grows again when they backspace, and an effect chasing
  // it would render one frame with an active row that is not there.
  const activeRow = Math.min(active, Math.max(visible.length - 1, 0));

  const listId = useId();
  const optionId = (index: number) => `${listId}-${index}`;

  const run = (command: Command) => {
    if (command.unavailable !== undefined) {
      // Registered, and it says why it cannot run. Nothing happens, and
      // nothing pretends to.
      return;
    }
    command.run();

    // Only close on an action that did not open something else. `run` above may
    // already have replaced this overlay (new task does), and closing after it
    // would shut the panel it just opened.
    if (store.getState().openOverlay === 'commandPalette') {
      store.getState().closeOverlay();
    }
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (event.key === 'Tab') {
      // The overlay is modal and the query field is its only tab stop, so Tab
      // has nowhere to go. Swallowing it is the focus trap.
      event.preventDefault();
      return;
    }

    if (event.key === 'Enter') {
      event.preventDefault();
      if (visible.length > 0) {
        run(visible[activeRow]);
      }
      return;
    }

    const action = boardActionFor(event.nativeEvent);
    if (action !== 'nextCard' && action !== 'previousCard') {
      return;
    }

    event.preventDefault();
    const step = action === 'nextCard' ? 1 : -1;
    // Clamped, never wrapped — the same rule the board follows, for the same
    // reason: a list that jumps from the bottom back to the top is a surprise.
    setActive(Math.min(Math.max(activeRow + step, 0), Math.max(visible.length - 1, 0)));
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-label={t('palette.label')}
      onKeyDown={handleKeyDown}
      // `bg-bg` plain rather than a token with an opacity modifier: the palette
      // colours resolve through bare CSS variables (design/tailwind.config.js)
      // that carry no <alpha-value> placeholder, so `bg-bg/70` would render
      // fully opaque anyway. QuickAdd makes the same point.
      className="fixed inset-0 z-40 flex items-start justify-center bg-bg p-4 backdrop-blur-glass"
    >
      <div className="flex w-[min(36rem,100%)] flex-col gap-2 rounded-lg border border-line bg-elevated p-4 text-ink shadow-sm">
        <input
          ref={filterRef}
          role="combobox"
          aria-expanded="true"
          aria-controls={listId}
          aria-activedescendant={visible.length > 0 ? optionId(activeRow) : undefined}
          aria-autocomplete="list"
          aria-label={t('palette.filter')}
          value={query}
          onChange={(event) => {
            setQuery(event.target.value);
            setActive(0);
          }}
          className="min-w-0 rounded-sm border border-line bg-surface px-2 py-1 text-ink"
        />

        {visible.length === 0 ? (
          <p className="min-w-0 break-words text-muted">{t('palette.empty')}</p>
        ) : (
          <ul
            id={listId}
            role="listbox"
            aria-label={t('palette.results')}
            // A capped, scrolling list rather than one that grows past the
            // bottom of the window: there are two dozen rows and Russian wraps
            // every one of them onto a second line.
            className="flex max-h-80 min-w-0 flex-col gap-2 overflow-y-auto"
          >
            {visible.map((command, index) => (
              <li
                key={command.id}
                id={optionId(index)}
                role="option"
                data-command={command.id}
                aria-selected={index === activeRow}
                aria-disabled={command.unavailable !== undefined}
                onClick={() => run(command)}
                // Wrapping, never truncating: the Russian labels are about a
                // third wider and a row that ends in an ellipsis is a row the
                // user cannot read.
                className={`flex min-w-0 flex-wrap items-baseline gap-2 rounded-sm px-2 py-1 transition-colors duration-fast ${
                  index === activeRow ? 'bg-accent text-on-accent' : 'bg-surface text-ink'
                } ${command.unavailable === undefined ? '' : 'opacity-60'}`}
              >
                <span className="min-w-0 flex-1 break-words">{command.label}</span>

                {command.unavailable !== undefined && (
                  <span
                    className={`min-w-0 break-words ${
                      index === activeRow ? 'text-on-accent' : 'text-muted'
                    }`}
                  >
                    {command.unavailable}
                  </span>
                )}

                {command.hint !== undefined && (
                  // font-mono for every shortcut hint (PLAN.md §3), and the
                  // string itself comes from lib/keyboard.ts.
                  <span
                    className={`font-mono ${
                      index === activeRow ? 'text-on-accent' : 'text-muted'
                    }`}
                  >
                    {command.hint}
                  </span>
                )}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}

export default CommandPalette;
