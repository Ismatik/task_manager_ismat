import { useTranslation } from 'react-i18next';

import { Column } from '../components/Column';
import { useAppState } from '../store/context';

// Nexus — the board.
//
// # Five columns, in Go's order, and the frontend does not know what they are
//
// `Board()` returns a slice of ColumnView, and dto.go says why it is a slice and
// not a map: "the five columns have an order ... and a map has none, in Go or in
// JSON". So this renders the array as it arrived. There is no list of statuses
// here, no `.sort(`, no `.filter(`, and no five-element constant — the board is
// five columns long because Go sent five, and Kanban.test.tsx proves the point
// by mocking a client that returns them in a deliberately unusual order and
// asserting the UI shows THAT order.
//
// # Three states, and none of them is a blank window
//
//   board === null   the read has not answered yet — nothing, briefly
//   board is empty   Go answered with no columns: a localised line, not a void
//   otherwise        the columns
//
// A FAILED read is the second case: S2-13's store leaves `board` at null and
// raises a toast, App's effect does not retry, and the user sees an empty board
// and an error rather than a white rectangle. That is S2-13's rule unchanged —
// a failed read is a bad session, not a blank window.

export function Kanban() {
  const { t } = useTranslation();
  const board = useAppState((state) => state.board);

  if (board === null) {
    return null;
  }

  if (board.length === 0) {
    return <p className="min-w-0 break-words text-muted">{t('board.empty')}</p>;
  }

  return (
    // `overflow-x-auto` on the board and nowhere else: when five columns will
    // not fit even at their floor width, the BOARD scrolls. A column never
    // scrolls sideways inside itself, which is what "no horizontal scroll
    // inside a column" means.
    <div className="flex min-w-0 items-start gap-2 overflow-x-auto">
      {board.map((column) => (
        <Column key={column.status} column={column} />
      ))}
    </div>
  );
}

export default Kanban;
