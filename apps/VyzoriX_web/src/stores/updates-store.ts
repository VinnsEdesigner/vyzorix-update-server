import { createVyzorStore } from '@/lib/state';
import type { InstallType, SyncStatus } from '@vyzorix/api-client';

export interface PushDraft {
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt: Date | null;
}

export interface UpdatesStoreState {
  draft: PushDraft;
  lastSyncStatus: SyncStatus;
  lastSyncAt: number | null;
  isSyncing: boolean;
  setDraftVersion: (version: string) => void;
  setDraftDeviceIds: (deviceIds: string[]) => void;
  toggleDraftDevice: (deviceId: string) => void;
  setDraftInstallType: (installType: InstallType) => void;
  setDraftScheduledAt: (scheduledAt: Date | null) => void;
  resetDraft: () => void;
  setSyncStatus: (status: SyncStatus) => void;
  setLastSyncAt: (timestamp: number | null) => void;
  setSyncing: (isSyncing: boolean) => void;
  clear: () => void;
}

const DEFAULT_DRAFT: PushDraft = {
  version: '',
  deviceIds: [],
  installType: 'immediate',
  scheduledAt: null,
};

export const useUpdatesStore = createVyzorStore<UpdatesStoreState>('UpdatesStore', (set) => ({
  draft: { ...DEFAULT_DRAFT },
  lastSyncStatus: 'idle',
  lastSyncAt: null,
  isSyncing: false,

  setDraftVersion: (version) =>
    set((state) => ({ draft: { ...state.draft, version } })),

  setDraftDeviceIds: (deviceIds) =>
    set((state) => ({ draft: { ...state.draft, deviceIds } })),

  toggleDraftDevice: (deviceId) =>
    set((state) => {
      const exists = state.draft.deviceIds.includes(deviceId);
      const deviceIds = exists
        ? state.draft.deviceIds.filter((id) => id !== deviceId)
        : [...state.draft.deviceIds, deviceId];
      return { draft: { ...state.draft, deviceIds } };
    }),

  setDraftInstallType: (installType) =>
    set((state) => ({ draft: { ...state.draft, installType } })),

  setDraftScheduledAt: (scheduledAt) =>
    set((state) => ({ draft: { ...state.draft, scheduledAt } })),

  resetDraft: () => set({ draft: { ...DEFAULT_DRAFT } }),

  setSyncStatus: (lastSyncStatus) => set({ lastSyncStatus }),

  setLastSyncAt: (lastSyncAt) => set({ lastSyncAt }),

  setSyncing: (isSyncing) => set({ isSyncing }),

  clear: () =>
    set({
      draft: { ...DEFAULT_DRAFT },
      lastSyncStatus: 'idle',
      lastSyncAt: null,
      isSyncing: false,
    }),
}));
