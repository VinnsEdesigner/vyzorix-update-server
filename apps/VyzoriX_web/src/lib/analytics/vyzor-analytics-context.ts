import { createContext, useContext } from 'react';
import type { VyzorAnalyticsAdapter } from '@/lib/analytics';

interface VyzorAnalyticsContextValue {
  adapter: VyzorAnalyticsAdapter;
}

export const VyzorAnalyticsContext =
  createContext<VyzorAnalyticsContextValue | null>(null);

export function useVyzorAnalyticsContext(): VyzorAnalyticsContextValue {
  const ctx = useContext(VyzorAnalyticsContext);
  if (!ctx) {
    throw new Error('useVyzorAnalyticsContext must be used within VyzorAnalyticsProvider');
  }
  return ctx;
}
