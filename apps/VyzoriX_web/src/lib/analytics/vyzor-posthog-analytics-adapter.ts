import type posthog from 'posthog-js';
import type { VyzorAnalyticsAdapter, VyzorAnalyticsConsent, VyzorAnalyticsUserProperties } from './vyzor-analytics-adapter';
import { readVyzorAnalyticsConfig } from './vyzor-analytics-config';

export class VyzorPostHogAnalyticsAdapter implements VyzorAnalyticsAdapter {
  private client: typeof posthog | null = null;
  private consent: VyzorAnalyticsConsent = 'pending';

  async ensureClient(): Promise<typeof posthog | null> {
    if (this.client) return this.client;
    const config = readVyzorAnalyticsConfig();
    if (!config.apiKey) return null;

    const mod = await import('posthog-js');
    this.client = mod.default;
    this.client.init(config.apiKey, {
      api_host: config.apiHost ?? 'https://us.i.posthog.com',
      autocapture: false,
      opt_out_capturing_by_default: true,
      loaded: (ph) => {
        if (this.consent === 'granted') ph.opt_in_capturing();
        else ph.opt_out_capturing();
      },
    });
    return this.client;
  }

  identify(distinctId: string, properties?: VyzorAnalyticsUserProperties): void {
    if (this.consent !== 'granted') return;
    this.ensureClient().then((ph) => ph?.identify(distinctId, properties));
  }

  track(event: string, properties?: Record<string, unknown>): void {
    if (this.consent !== 'granted') return;
    this.ensureClient().then((ph) => ph?.capture(event, properties));
  }

  page(name: string, properties?: Record<string, unknown>): void {
    if (this.consent !== 'granted') return;
    this.ensureClient().then((ph) =>
      ph?.capture('$pageview', { ...properties, $page_name: name }),
    );
  }

  flush(): void {
    if (this.consent !== 'granted') return;
    this.ensureClient().then((ph) => {
      if (!ph) return;
      // PostHog auto-flushes; this is a no-op for compatibility
    });
  }

  reset(): void {
    this.ensureClient().then((ph) => ph?.reset());
  }

  setConsent(consent: VyzorAnalyticsConsent): void {
    this.consent = consent;
    this.ensureClient().then((ph) => {
      if (!ph) return;
      if (consent === 'granted') ph.opt_in_capturing();
      else ph.opt_out_capturing();
    });
  }
}
