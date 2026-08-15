import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import {
  useDebounce,
  useDebouncedCallback,
  useThrottle,
  useMediaQuery,
  useLocalStorage,
} from '@/hooks/_shared/use-debounce';

describe('useDebounce', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('returns initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('initial', 300));
    expect(result.current).toBe('initial');
  });

  it('updates after delay', () => {
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: 'a' },
    });
    rerender({ value: 'b' });
    expect(result.current).toBe('a');
    act(() => vi.advanceTimersByTime(300));
    expect(result.current).toBe('b');
  });

  it('does not update before delay', () => {
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: 'a' },
    });
    rerender({ value: 'b' });
    act(() => vi.advanceTimersByTime(299));
    expect(result.current).toBe('a');
  });

  it('resets timer on rapid changes', () => {
    const { result, rerender } = renderHook(({ value }) => useDebounce(value, 300), {
      initialProps: { value: 'a' },
    });
    rerender({ value: 'b' });
    act(() => vi.advanceTimersByTime(200));
    rerender({ value: 'c' });
    act(() => vi.advanceTimersByTime(200));
    expect(result.current).toBe('a');
    act(() => vi.advanceTimersByTime(100));
    expect(result.current).toBe('c');
  });
});

describe('useDebouncedCallback', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('does not call immediately', () => {
    const fn = vi.fn();
    const { result } = renderHook(() => useDebouncedCallback(fn, 300));
    result.current('arg');
    expect(fn).not.toHaveBeenCalled();
  });

  it('calls after delay', () => {
    const fn = vi.fn();
    const { result } = renderHook(() => useDebouncedCallback(fn, 300));
    result.current('arg');
    act(() => vi.advanceTimersByTime(300));
    expect(fn).toHaveBeenCalledWith('arg');
  });

  it('only fires last call', () => {
    const fn = vi.fn();
    const { result } = renderHook(() => useDebouncedCallback(fn, 300));
    result.current('a');
    result.current('b');
    result.current('c');
    act(() => vi.advanceTimersByTime(300));
    expect(fn).toHaveBeenCalledTimes(1);
    expect(fn).toHaveBeenCalledWith('c');
  });
});

describe('useThrottle', () => {
  beforeEach(() => vi.useFakeTimers());
  afterEach(() => vi.useRealTimers());

  it('returns initial value', () => {
    const { result } = renderHook(() => useThrottle('init', 300));
    expect(result.current).toBe('init');
  });
});

describe('useMediaQuery', () => {
  it('returns boolean based on matchMedia', () => {
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({
      matches: true,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }));
    const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'));
    expect(result.current).toBe(true);
    vi.unstubAllGlobals();
  });

  it('updates when media query changes', () => {
    const listeners = new Set<(e: MediaQueryListEvent) => void>();
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({
      matches: false,
      addEventListener: (_: string, cb: (e: MediaQueryListEvent) => void) => listeners.add(cb),
      removeEventListener: (_: string, cb: (e: MediaQueryListEvent) => void) => listeners.delete(cb),
    }));
    const { result } = renderHook(() => useMediaQuery('(min-width: 768px)'));
    expect(result.current).toBe(false);
    act(() => {
      for (const cb of listeners) cb({ matches: true } as MediaQueryListEvent);
    });
    expect(result.current).toBe(true);
    vi.unstubAllGlobals();
  });
});

describe('useLocalStorage', () => {
  it('returns initial value when nothing stored', () => {
    const { result } = renderHook(() => useLocalStorage('key-init', 'default'));
    expect(result.current[0]).toBe('default');
  });

  it('reads from localStorage on init', () => {
    localStorage.setItem('key-stored', JSON.stringify('stored-val'));
    const { result } = renderHook(() => useLocalStorage('key-stored', 'default'));
    expect(result.current[0]).toBe('stored-val');
  });

  it('updates localStorage on set', () => {
    vi.useFakeTimers();
    const { result } = renderHook(() => useLocalStorage('key-set', 'initial'));
    act(() => {
      result.current[1]('new-val');
      vi.advanceTimersByTime(500);
    });
    expect(JSON.parse(localStorage.getItem('key-set')!)).toBe('new-val');
    vi.useRealTimers();
  });
});
