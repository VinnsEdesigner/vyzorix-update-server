import { describe, it, expect, beforeEach } from 'vitest';
import { useUpdatesStore } from '@/stores/updates-store';

describe('useUpdatesStore', () => {
  beforeEach(() => {
    useUpdatesStore.getState().clear();
  });

  it('has the default draft', () => {
    const { draft } = useUpdatesStore.getState();
    expect(draft.version).toBe('');
    expect(draft.deviceIds).toEqual([]);
    expect(draft.installType).toBe('immediate');
    expect(draft.scheduledAt).toBeNull();
  });

  it('setDraftVersion updates the version', () => {
    useUpdatesStore.getState().setDraftVersion('v1.2.0');
    expect(useUpdatesStore.getState().draft.version).toBe('v1.2.0');
  });

  it('setDraftDeviceIds replaces the list', () => {
    useUpdatesStore.getState().setDraftDeviceIds(['dev-1', 'dev-2']);
    expect(useUpdatesStore.getState().draft.deviceIds).toEqual(['dev-1', 'dev-2']);
  });

  it('toggleDraftDevice adds when absent', () => {
    useUpdatesStore.getState().toggleDraftDevice('dev-1');
    expect(useUpdatesStore.getState().draft.deviceIds).toEqual(['dev-1']);
  });

  it('toggleDraftDevice removes when present', () => {
    useUpdatesStore.getState().setDraftDeviceIds(['dev-1', 'dev-2']);
    useUpdatesStore.getState().toggleDraftDevice('dev-1');
    expect(useUpdatesStore.getState().draft.deviceIds).toEqual(['dev-2']);
  });

  it('setDraftInstallType updates the install type', () => {
    useUpdatesStore.getState().setDraftInstallType('scheduled');
    expect(useUpdatesStore.getState().draft.installType).toBe('scheduled');
  });

  it('setDraftScheduledAt updates the scheduled date', () => {
    const date = new Date('2026-09-01T10:00:00Z');
    useUpdatesStore.getState().setDraftScheduledAt(date);
    expect(useUpdatesStore.getState().draft.scheduledAt).toEqual(date);
  });

  it('resetDraft restores defaults', () => {
    useUpdatesStore.getState().setDraftVersion('v1.2.0');
    useUpdatesStore.getState().setDraftDeviceIds(['dev-1']);
    useUpdatesStore.getState().setDraftInstallType('scheduled');
    useUpdatesStore.getState().resetDraft();
    const { draft } = useUpdatesStore.getState();
    expect(draft.version).toBe('');
    expect(draft.deviceIds).toEqual([]);
    expect(draft.installType).toBe('immediate');
  });

  it('setSyncStatus updates lastSyncStatus', () => {
    useUpdatesStore.getState().setSyncStatus('synced');
    expect(useUpdatesStore.getState().lastSyncStatus).toBe('synced');
  });

  it('setLastSyncAt updates the timestamp', () => {
    useUpdatesStore.getState().setLastSyncAt(1000);
    expect(useUpdatesStore.getState().lastSyncAt).toBe(1000);
  });

  it('setSyncing toggles the syncing flag', () => {
    useUpdatesStore.getState().setSyncing(true);
    expect(useUpdatesStore.getState().isSyncing).toBe(true);
  });

  it('clear resets all state', () => {
    useUpdatesStore.getState().setDraftVersion('v1.2.0');
    useUpdatesStore.getState().setSyncStatus('error');
    useUpdatesStore.getState().setLastSyncAt(1000);
    useUpdatesStore.getState().setSyncing(true);
    useUpdatesStore.getState().clear();
    const state = useUpdatesStore.getState();
    expect(state.draft.version).toBe('');
    expect(state.lastSyncStatus).toBe('idle');
    expect(state.lastSyncAt).toBeNull();
    expect(state.isSyncing).toBe(false);
  });
});
