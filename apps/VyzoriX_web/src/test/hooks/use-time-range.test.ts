import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useTimeRange,
  calculateResolution,
  useResolution,
  parseTimeRangeFromURL,
  updateTimeRangeURL,
  formatTimeRange,
  TIME_RANGES,
} from '@/hooks/_shared/use-time-range';

describe('useTimeRange', () => {
  it('defaults to 6h range', () => {
    const { result } = renderHook(() => useTimeRange());
    expect(result.current.range).toBe('6h');
  });

  it('respects initialRange', () => {
    const { result } = renderHook(() => useTimeRange({ initialRange: '1h' }));
    expect(result.current.range).toBe('1h');
  });

  it('setRange updates range and timestamps', () => {
    const { result } = renderHook(() => useTimeRange());
    act(() => result.current.setRange('24h'));
    expect(result.current.range).toBe('24h');
  });

  it('setRange calls onRangeChange', () => {
    const onRangeChange = vi.fn();
    const { result } = renderHook(() => useTimeRange({ onRangeChange }));
    act(() => result.current.setRange('7d'));
    expect(onRangeChange).toHaveBeenCalledWith('7d');
  });

  it('startTime is before endTime', () => {
    const { result } = renderHook(() => useTimeRange());
    expect(result.current.startTime.getTime()).toBeLessThan(result.current.endTime.getTime());
  });

  it('refresh updates timestamps to current time', () => {
    const { result } = renderHook(() => useTimeRange());
    const before = result.current.endTime.getTime();
    vi.useFakeTimers();
    vi.setSystemTime(new Date(before + 60_000));
    act(() => result.current.refresh());
    vi.useRealTimers();
    expect(result.current.endTime.getTime()).toBeGreaterThan(before);
  });

  it('getTimeParams returns ISO strings and range', () => {
    const { result } = renderHook(() => useTimeRange({ initialRange: '1h' }));
    const params = result.current.getTimeParams();
    expect(typeof params.startTime).toBe('string');
    expect(typeof params.endTime).toBe('string');
    expect(params.range).toBe('1h');
  });

  it('getTimestamps returns ms numbers', () => {
    const { result } = renderHook(() => useTimeRange());
    const ts = result.current.getTimestamps();
    expect(typeof ts.start).toBe('number');
    expect(typeof ts.end).toBe('number');
    expect(ts.start).toBeLessThan(ts.end);
  });

  it('exposes all preset ranges', () => {
    const { result } = renderHook(() => useTimeRange());
    expect(Object.keys(result.current.ranges)).toEqual(['1h', '6h', '24h', '7d']);
  });
});

describe('calculateResolution', () => {
  it('1h returns 1m resolution with 60 points', () => {
    expect(calculateResolution('1h')).toEqual({ resolution: '1m', maxPoints: 60 });
  });

  it('6h returns 5m resolution with 72 points', () => {
    expect(calculateResolution('6h')).toEqual({ resolution: '5m', maxPoints: 72 });
  });

  it('24h returns 15m resolution with 96 points', () => {
    expect(calculateResolution('24h')).toEqual({ resolution: '15m', maxPoints: 96 });
  });

  it('7d returns 1h resolution with 168 points', () => {
    expect(calculateResolution('7d')).toEqual({ resolution: '1h', maxPoints: 168 });
  });
});

describe('useResolution', () => {
  it('returns resolution for the given range', () => {
    const { result } = renderHook(() => useResolution('1h'));
    expect(result.current.resolution).toBe('1m');
  });

  it('memoizes on range', () => {
    const { result, rerender } = renderHook(({ range }) => useResolution(range), {
      initialProps: { range: '1h' as '1h' | '6h' | '24h' | '7d' },
    });
    const first = result.current;
    rerender({ range: '1h' });
    expect(result.current).toBe(first);
    rerender({ range: '6h' });
    expect(result.current).not.toBe(first);
  });
});

describe('parseTimeRangeFromURL', () => {
  it('defaults to 6h when no param', () => {
    expect(parseTimeRangeFromURL(new URLSearchParams())).toBe('6h');
  });

  it('parses valid range', () => {
    expect(parseTimeRangeFromURL(new URLSearchParams('range=24h'))).toBe('24h');
  });

  it('falls back to 6h for invalid range', () => {
    expect(parseTimeRangeFromURL(new URLSearchParams('range=99h'))).toBe('6h');
  });
});

describe('updateTimeRangeURL', () => {
  it('omits default 6h', () => {
    expect(updateTimeRangeURL('6h').toString()).toBe('');
  });

  it('includes non-default range', () => {
    expect(updateTimeRangeURL('1h').get('range')).toBe('1h');
  });
});

describe('formatTimeRange', () => {
  it('returns the label', () => {
    expect(formatTimeRange('1h')).toBe('1h');
    expect(formatTimeRange('7d')).toBe('7d');
  });
});

describe('TIME_RANGES constant', () => {
  it('has 4 presets with correct hours', () => {
    expect(TIME_RANGES['1h'].hours).toBe(1);
    expect(TIME_RANGES['6h'].hours).toBe(6);
    expect(TIME_RANGES['24h'].hours).toBe(24);
    expect(TIME_RANGES['7d'].hours).toBe(168);
  });
});
