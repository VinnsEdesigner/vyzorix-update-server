import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useTimelineStreamStore } from '@/stores/timeline-stream-store';
import { useTimelineFilter } from '@/hooks/diagnostics/use-timeline-filter';

describe('useTimelineFilter', () => {
  beforeEach(() => {
    useTimelineStreamStore.setState({
      byDevice: {},
      filters: {},
      autoScroll: true,
      activeOrganizationId: null,
    });
  });

  it('defaults to all categories, no range, autoScroll on', () => {
    const { result } = renderHook(() => useTimelineFilter());
    expect(result.current.category).toBe('all');
    expect(result.current.rangeMs).toBeNull();
    expect(result.current.autoScroll).toBe(true);
  });

  it('setCategory updates the filter state', () => {
    const { result } = renderHook(() => useTimelineFilter());
    act(() => result.current.setCategory('error'));
    expect(result.current.category).toBe('error');
    act(() => result.current.setCategory('all'));
    expect(result.current.category).toBe('all');
  });

  it('setRangeMs updates the range', () => {
    const { result } = renderHook(() => useTimelineFilter());
    act(() => result.current.setRangeMs(60_000));
    expect(result.current.rangeMs).toBe(60_000);
    act(() => result.current.setRangeMs(null));
    expect(result.current.rangeMs).toBeNull();
  });

  it('toggleAutoScroll flips the flag', () => {
    const { result } = renderHook(() => useTimelineFilter());
    expect(result.current.autoScroll).toBe(true);
    act(() => result.current.toggleAutoScroll());
    expect(result.current.autoScroll).toBe(false);
  });

  it('toQueryParams produces a 24h window when no range is set', () => {
    const { result } = renderHook(() => useTimelineFilter());
    const params = result.current.toQueryParams();
    const now = Date.now();
    expect(params.startTime).toBeGreaterThan(now - 25 * 60 * 60 * 1000);
    expect(params.startTime).toBeLessThan(now - 23 * 60 * 60 * 1000);
    expect(params.endTime).toBeLessThanOrEqual(now);
    expect(params.limit).toBe(50);
    expect(params.eventType).toBeUndefined();
  });

  it('toQueryParams honors a custom range', () => {
    const { result } = renderHook(() => useTimelineFilter());
    act(() => result.current.setRangeMs(5 * 60 * 1000));
    const now = Date.now();
    const params = result.current.toQueryParams();
    expect(params.startTime).toBeGreaterThan(now - 6 * 60 * 1000);
    expect(params.startTime).toBeLessThan(now - 4 * 60 * 1000);
  });

  it('clear resets filters and clears the stream', () => {
    const { result } = renderHook(() => useTimelineFilter());
    act(() => {
      result.current.setCategory('command');
      result.current.setRangeMs(10_000);
    });
    useTimelineStreamStore.getState().append('123', {
      id: 'evt-1',
      deviceId: '123',
      type: 'TELEMETRY',
      timestamp: new Date(),
      data: {},
    });
    expect(useTimelineStreamStore.getState().byDevice['123']).toHaveLength(1);
    act(() => result.current.clear());
    expect(result.current.category).toBe('all');
    expect(result.current.rangeMs).toBeNull();
    expect(useTimelineStreamStore.getState().byDevice).toEqual({});
  });
});
