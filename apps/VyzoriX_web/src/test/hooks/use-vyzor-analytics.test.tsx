import { describe, it, expect, vi } from 'vitest';
import { renderHook } from '@testing-library/react';
import { VyzorAnalyticsProvider } from '@/providers/vyzor-analytics-provider';
import { useVyzorAnalytics } from '@/hooks/analytics/use-vyzor-analytics';
import type { ReactNode } from 'react';

function createWrapper() {
  return function Wrapper({ children }: { children: ReactNode }) {
    return <VyzorAnalyticsProvider>{children}</VyzorAnalyticsProvider>;
  };
}

describe('useVyzorAnalytics', () => {
  it('throws when used outside provider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    expect(() => renderHook(() => useVyzorAnalytics())).toThrow(
      'useVyzorAnalyticsContext',
    );
    spy.mockRestore();
  });

  it('returns track/identify/page/flush/reset within provider', () => {
    const { result } = renderHook(() => useVyzorAnalytics(), {
      wrapper: createWrapper(),
    });
    expect(typeof result.current.track).toBe('function');
    expect(typeof result.current.identify).toBe('function');
    expect(typeof result.current.page).toBe('function');
    expect(typeof result.current.flush).toBe('function');
    expect(typeof result.current.reset).toBe('function');
    expect(result.current.events).toBeDefined();
  });

  it('track calls adapter.track with event + properties', () => {
    const { result } = renderHook(() => useVyzorAnalytics(), {
      wrapper: createWrapper(),
    });
    const spy = vi.spyOn(result.current, 'track');
    result.current.track('test_event', { foo: 'bar' });
    expect(spy).toHaveBeenCalledWith('test_event', { foo: 'bar' });
    spy.mockRestore();
  });
});
