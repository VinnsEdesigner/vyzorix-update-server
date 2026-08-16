import { useCallback } from 'react';
import { useVyzorAnalyticsContext } from '@/lib/analytics';
import { VYZOR_ANALYTICS_EVENTS } from '@/lib/analytics';

export function useVyzorAnalytics() {
  const { adapter } = useVyzorAnalyticsContext();

  const track = useCallback(
    (event: string, properties?: Record<string, unknown>) => {
      adapter.track(event, properties);
    },
    [adapter],
  );

  const identify = useCallback(
    (distinctId: string, properties?: Record<string, string | number | boolean | null>) => {
      adapter.identify(distinctId, properties);
    },
    [adapter],
  );

  const page = useCallback(
    (name: string, properties?: Record<string, unknown>) => {
      adapter.page(name, properties);
    },
    [adapter],
  );

  const flush = useCallback(() => adapter.flush(), [adapter]);

  const reset = useCallback(() => adapter.reset(), [adapter]);

  return {
    track,
    identify,
    page,
    flush,
    reset,
    events: VYZOR_ANALYTICS_EVENTS,
  };
}
