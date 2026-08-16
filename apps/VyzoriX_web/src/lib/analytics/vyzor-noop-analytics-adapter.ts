import type { VyzorAnalyticsAdapter, VyzorAnalyticsConsent, VyzorAnalyticsUserProperties } from './vyzor-analytics-adapter';

export class VyzorNoopAnalyticsAdapter implements VyzorAnalyticsAdapter {
  identify(_distinctId: string, _properties?: VyzorAnalyticsUserProperties): void {}
  track(_event: string, _properties?: Record<string, unknown>): void {}
  page(_name: string, _properties?: Record<string, unknown>): void {}
  flush(): void {}
  reset(): void {}
  setConsent(_consent: VyzorAnalyticsConsent): void {}
}
