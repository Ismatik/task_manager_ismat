import type { StateCreator } from 'zustand';

import type { AppState } from './index';

// Nexus — the genuinely client-only state.
//
// ARCHITECTURE.md section 6: "frontend/src/store holds UI state. Domain state
// is owned by Go and fetched." This is the UI half — the things that exist only
// because there is a screen, and that Go has no opinion about: what is
// selected, which column has focus, which overlay is open, which card is under
// the pointer.
//
// Nothing here is a rule. `focusedColumn` is the status string of a column the
// BOARD returned, never a column this file knows the name of; that is why it is
// a plain string and why there is no list of the five anywhere in frontend/src.

/** Which modal surface is open, if any. */
export type Overlay = 'quickAdd' | 'commandPalette' | null;

export interface UiSlice {
  /** The card the keyboard model is on, by node id. */
  selectedNodeId: string | null;
  /** The column the keyboard model is in, by the status string Go returned. */
  focusedColumn: string | null;
  /** Which overlay is open. */
  openOverlay: Overlay;
  /** The card currently being dragged, by node id. */
  draggingNodeId: string | null;

  select(nodeId: string | null): void;
  focusColumn(status: string | null): void;
  openOverlayPanel(overlay: Overlay): void;
  closeOverlay(): void;
  setDragging(nodeId: string | null): void;
}

export const createUiSlice: StateCreator<AppState, [], [], UiSlice> = (set) => ({
  selectedNodeId: null,
  focusedColumn: null,
  openOverlay: null,
  draggingNodeId: null,

  select: (nodeId) => set({ selectedNodeId: nodeId }),
  focusColumn: (status) => set({ focusedColumn: status }),
  openOverlayPanel: (overlay) => set({ openOverlay: overlay }),
  closeOverlay: () => set({ openOverlay: null }),
  setDragging: (nodeId) => set({ draggingNodeId: nodeId }),
});
