export type VyzorAnalyticsConsent = 'pending' | 'granted' | 'denied';

export interface VyzorAnalyticsUserProperties {
  [key: string]: string | number | boolean | null;
}

export interface VyzorAnalyticsAdapter {
  identify(
    distinctId: string,
    properties?: VyzorAnalyticsUserProperties,
  ): void;
  track(
    event: string,
    properties?: Record<string, unknown>,
  ): void;
  page(name: string, properties?: Record<string, unknown>): void;
  flush(): void;
  reset(): void;
  setConsent(consent: VyzorAnalyticsConsent): void;
}
