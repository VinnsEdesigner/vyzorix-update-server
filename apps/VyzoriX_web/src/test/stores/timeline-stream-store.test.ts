import { describe, it, expect, beforeEach } from 'vitest';
import { useTimelineStreamStore } from '@/stores/timeline-stream-store';
import type { TimelineEvent } from '@vyzorix/api-client';

function makeEvent(overrides: Partial<TimelineEvent> = {}): TimelineEvent {
  return {
    id: `evt-${Math.random().toString(36).slice(2, 8)}`,
    deviceId: '123',
    type: 'TELEMETRY',
    timestamp: new Date(),
    data: {},
    ...overrides,
  };
}

describe('useTimelineStreamStore', () => {
  beforeEach(() => {
    useTimelineStreamStore.setState({
      byDevice: {},
      filters: {},
      autoScroll: true,
      activeOrganizationId: null,
    });
  });

  it('append prepends newest-first (matches server desc ordering)', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'evt-1', timestamp: new Date(1000) }));
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'evt-2', timestamp: new Date(2000) }));
    const events = useTimelineStreamStore.getState().byDevice['123']!;
    expect(events).toHaveLength(2);
    expect(events[0]?.id).toBe('evt-2');
  });

  it('ring buffer trims to cap on overflow', () => {
    for (let i = 0; i < 510; i++) {
      useTimelineStreamStore.getState().append('123', makeEvent({ id: `evt-${i}` }));
    }
    const events = useTimelineStreamStore.getState().byDevice['123']!;
    expect(events).toHaveLength(500);
  });

  it('appendBatch merges and dedupes by id, prepending the batch in given order', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'existing' }));
    useTimelineStreamStore.getState().appendBatch('123', [
      makeEvent({ id: 'existing' }),
      makeEvent({ id: 'new-a' }),
      makeEvent({ id: 'new-b' }),
    ]);
    const events = useTimelineStreamStore.getState().byDevice['123']!;
    // Deduped batch (in arrival order) prepended before the existing tail.
    expect(events.map((e) => e.id)).toEqual(['new-a', 'new-b', 'existing']);
  });

  it('appendBatch with empty array is a no-op', () => {
    useTimelineStreamStore.getState().appendBatch('123', []);
    expect(useTimelineStreamStore.getState().byDevice['123']).toBeUndefined();
  });

  it('setFilter category filters via getEventCategory', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'tel', type: 'TELEMETRY' }));
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'cmd', type: 'COMMAND_SENT' }));
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'err', type: 'ERROR' }));
    useTimelineStreamStore.getState().setFilter({ category: 'command' });
    const filtered = useTimelineStreamStore.getState().getEvents('123');
    expect(filtered).toHaveLength(1);
    expect(filtered[0]?.id).toBe('cmd');
  });

  it('REGISTERED events map to connection category (matches Go server)', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'reg', type: 'REGISTERED' }));
    useTimelineStreamStore.getState().setFilter({ category: 'connection' });
    const filtered = useTimelineStreamStore.getState().getEvents('123');
    expect(filtered).toHaveLength(1);
    expect(filtered[0]?.id).toBe('reg');
  });

  it('setFilter rangeMs drops events older than the window', () => {
    const old = new Date(Date.now() - 60 * 60 * 1000);
    const recent = new Date();
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'old', timestamp: old }));
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'recent', timestamp: recent }));
    useTimelineStreamStore.getState().setFilter({ rangeMs: 5 * 60 * 1000 });
    const filtered = useTimelineStreamStore.getState().getEvents('123');
    expect(filtered.map((e) => e.id)).toEqual(['recent']);
  });

  it('toggleAutoScroll flips the flag', () => {
    expect(useTimelineStreamStore.getState().autoScroll).toBe(true);
    useTimelineStreamStore.getState().toggleAutoScroll();
    expect(useTimelineStreamStore.getState().autoScroll).toBe(false);
  });

  it('clear(imei) removes a single device buffer', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'a' }));
    useTimelineStreamStore.getState().append('456', makeEvent({ id: 'b' }));
    useTimelineStreamStore.getState().clear('123');
    expect(useTimelineStreamStore.getState().byDevice['123']).toBeUndefined();
    expect(useTimelineStreamStore.getState().byDevice['456']).toBeDefined();
  });

  it('clear() with no arg wipes all buffers', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'a' }));
    useTimelineStreamStore.getState().clear();
    expect(useTimelineStreamStore.getState().byDevice).toEqual({});
  });

  it('setActiveOrganization clears buffers on org switch (no cross-org leak)', () => {
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'a' }));
    useTimelineStreamStore.getState().setActiveOrganization('org-1');
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'b' }));
    useTimelineStreamStore.getState().setActiveOrganization('org-2');
    expect(useTimelineStreamStore.getState().byDevice).toEqual({});
    expect(useTimelineStreamStore.getState().activeOrganizationId).toBe('org-2');
  });

  it('setActiveOrganization is a no-op when org unchanged', () => {
    useTimelineStreamStore.getState().setActiveOrganization('org-1');
    useTimelineStreamStore.getState().append('123', makeEvent({ id: 'a' }));
    useTimelineStreamStore.getState().setActiveOrganization('org-1');
    expect(useTimelineStreamStore.getState().byDevice['123']).toHaveLength(1);
  });
});
