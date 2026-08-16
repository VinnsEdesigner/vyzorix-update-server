import { createVyzorStore } from '@/lib/state';
import type { CommandParams, PresetCommandType } from '@vyzorix/api-client';

export interface PendingCommand {
  dispatchId: string;
  imei: string;
  commandType: PresetCommandType;
  params?: CommandParams;
  createdAt: number;
}

export interface CommandDispatchState {
  pendingCommands: Record<string, PendingCommand>;
  pendingCount: number;
  addPending: (command: PendingCommand) => void;
  removePending: (dispatchId: string) => void;
  clearPending: () => void;
  getPending: (dispatchId: string) => PendingCommand | undefined;
}

export const useCommandDispatchStore = createVyzorStore<CommandDispatchState>('CommandDispatchStore', (set, get) => ({
  pendingCommands: {},
  pendingCount: 0,
  addPending: (command) =>
    set((state) => {
      const pendingCommands = { ...state.pendingCommands, [command.dispatchId]: command };
      return { pendingCommands, pendingCount: Object.keys(pendingCommands).length };
    }),
  removePending: (dispatchId) =>
    set((state) => {
      if (!state.pendingCommands[dispatchId]) return state;
      const pendingCommands = { ...state.pendingCommands };
      delete pendingCommands[dispatchId];
      return { pendingCommands, pendingCount: Object.keys(pendingCommands).length };
    }),
  clearPending: () => set({ pendingCommands: {}, pendingCount: 0 }),
  getPending: (dispatchId) => get().pendingCommands[dispatchId],
}));
