import { useEffect, useMemo, useState, type ReactNode } from 'react';
import {
  VyzorNoopAnalyticsAdapter,
  readVyzorAnalyticsConfig,
  type VyzorAnalyticsAdapter,
} from '@/lib/analytics';
import { VyzorAnalyticsContext } from '@/lib/analytics/vyzor-analytics-context';
import { useVyzorConsentStore } from '@/stores/vyzor-consent-store';

interface VyzorAnalyticsProviderProps {
  children: ReactNode;
}

async function createPostHogAdapter(): Promise<VyzorAnalyticsAdapter> {
  const { VyzorPostHogAnalyticsAdapter } = await import(
    '@/lib/analytics/vyzor-posthog-analytics-adapter'
  );
  return new VyzorPostHogAnalyticsAdapter();
}

export function VyzorAnalyticsProvider({ children }: VyzorAnalyticsProviderProps) {
  const consent = useVyzorConsentStore((s) => s.consent);
  const [adapter, setAdapter] = useState<VyzorAnalyticsAdapter>(() => {
    const config = readVyzorAnalyticsConfig();
    if (!config.enabled) return new VyzorNoopAnalyticsAdapter();
    return new VyzorNoopAnalyticsAdapter();
  });

  useEffect(() => {
    const config = readVyzorAnalyticsConfig();
    if (!config.enabled) return;
    let cancelled = false;
    createPostHogAdapter().then((real) => {
      if (!cancelled) {
        real.setConsent(consent);
        setAdapter(real);
      }
    });
    return () => { cancelled = true; };
  }, []);

  useEffect(() => {
    adapter.setConsent(consent);
  }, [adapter, consent]);

  const value = useMemo(() => ({ adapter }), [adapter]);

  return (
    <VyzorAnalyticsContext.Provider value={value}>
      {children}
    </VyzorAnalyticsContext.Provider>
  );
}
