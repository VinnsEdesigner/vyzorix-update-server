import { describe, it, expect, beforeEach } from 'vitest';
import { useLogStreamStore } from '@/stores/log-stream-store';
import type { LogEntry } from '@vyzorix/api-client';

function makeEntry(overrides: Partial<LogEntry> = {}): LogEntry {
  return {
    id: `log-${Math.random().toString(36).slice(2, 8)}`,
    deviceId: '123',
    eventType: 'info',
    timestamp: new Date(),
    ...overrides,
  };
}

describe('useLogStreamStore', () => {
  beforeEach(() => {
    useLogStreamStore.setState({ byDevice: {}, filters: {}, autoScroll: true, activeOrganizationId: null });
  });

  it('append adds entries to a device buffer', () => {
    useLogStreamStore.getState().append('123', makeEntry({ id: 'log-1' }));
    useLogStreamStore.getState().append('123', makeEntry({ id: 'log-2' }));
    expect(useLogStreamStore.getState().byDevice['123']).toHaveLength(2);
  });

  it('ring buffer trims to cap on overflow', () => {
    for (let i = 0; i < 510; i++) {
      useLogStreamStore.getState().append('123', makeEntry({ id: `log-${i}` }));
    }
    const entries = useLogStreamStore.getState().byDevice['123']!;
    expect(entries).toHaveLength(500);
    expect(entries[0]?.id).toBe('log-10');
  });

  it('appendBatch appends multiple entries', () => {
    useLogStreamStore.getState().appendBatch('123', [makeEntry({ id: 'a' }), makeEntry({ id: 'b' })]);
    expect(useLogStreamStore.getState().byDevice['123']).toHaveLength(2);
  });

  it('setFilter filters entries by type', () => {
    useLogStreamStore.getState().append('123', makeEntry({ id: 'a', eventType: 'error' }));
    useLogStreamStore.getState().append('123', makeEntry({ id: 'b', eventType: 'info' }));
    useLogStreamStore.getState().setFilter({ type: 'error' });
    expect(useLogStreamStore.getState().getEntries('123')).toHaveLength(1);
    expect(useLogStreamStore.getState().getEntries('123')[0]?.id).toBe('a');
  });

  it('toggleAutoScroll flips the flag', () => {
    expect(useLogStreamStore.getState().autoScroll).toBe(true);
    useLogStreamStore.getState().toggleAutoScroll();
    expect(useLogStreamStore.getState().autoScroll).toBe(false);
  });

  it('clear removes a single device buffer', () => {
    useLogStreamStore.getState().append('123', makeEntry());
    useLogStreamStore.getState().append('456', makeEntry());
    useLogStreamStore.getState().clear('123');
    expect(useLogStreamStore.getState().byDevice['123']).toBeUndefined();
    expect(useLogStreamStore.getState().byDevice['456']).toHaveLength(1);
  });

  it('clear with no arg removes all buffers', () => {
    useLogStreamStore.getState().append('123', makeEntry());
    useLogStreamStore.getState().append('456', makeEntry());
    useLogStreamStore.getState().clear();
    expect(Object.keys(useLogStreamStore.getState().byDevice)).toHaveLength(0);
  });

  it('setActiveOrganization clears buffers on org switch', () => {
    useLogStreamStore.getState().setActiveOrganization('org-1');
    useLogStreamStore.getState().append('123', makeEntry());
    useLogStreamStore.getState().setActiveOrganization('org-2');
    expect(Object.keys(useLogStreamStore.getState().byDevice)).toHaveLength(0);
  });

  it('setActiveOrganization does NOT clear when org unchanged', () => {
    useLogStreamStore.getState().setActiveOrganization('org-1');
    useLogStreamStore.getState().append('123', makeEntry());
    useLogStreamStore.getState().setActiveOrganization('org-1');
    expect(useLogStreamStore.getState().byDevice['123']).toHaveLength(1);
  });
});
