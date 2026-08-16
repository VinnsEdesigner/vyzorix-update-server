import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useVyzorConsentStore } from '@/stores/vyzor-consent-store';

describe('vyzor-consent-store', () => {
  beforeEach(() => {
    localStorage.clear();
    useVyzorConsentStore.setState({ consent: 'pending', bannerDismissed: false });
  });

  it('starts pending with banner visible', () => {
    const state = useVyzorConsentStore.getState();
    expect(state.consent).toBe('pending');
    expect(state.bannerDismissed).toBe(false);
  });

  it('grants consent and dismisses banner', () => {
    useVyzorConsentStore.getState().setConsent('granted');
    const state = useVyzorConsentStore.getState();
    expect(state.consent).toBe('granted');
    expect(state.bannerDismissed).toBe(true);
  });

  it('denies consent and dismisses banner', () => {
    useVyzorConsentStore.getState().setConsent('denied');
    const state = useVyzorConsentStore.getState();
    expect(state.consent).toBe('denied');
    expect(state.bannerDismissed).toBe(true);
  });

  it('can dismiss banner without deciding', () => {
    useVyzorConsentStore.getState().dismissBanner();
    expect(useVyzorConsentStore.getState().bannerDismissed).toBe(true);
    expect(useVyzorConsentStore.getState().consent).toBe('pending');
  });

  it('persists consent to localStorage', () => {
    useVyzorConsentStore.getState().setConsent('granted');
    const stored = JSON.parse(localStorage.getItem('vyzorix-analytics-consent') ?? '{}');
    expect(stored.state.consent).toBe('granted');
    expect(stored.state.bannerDismissed).toBe(true);
  });

  it('hydrates persisted consent', async () => {
    localStorage.clear();
    localStorage.setItem(
      'vyzorix-analytics-consent',
      JSON.stringify({ state: { consent: 'denied', bannerDismissed: true }, version: 0 }),
    );
    vi.resetModules();
    const { useVyzorConsentStore: freshStore } = await import('@/stores/vyzor-consent-store');
    await vi.waitFor(() => {
      expect(freshStore.getState().consent).toBe('denied');
    });
    vi.resetModules();
  });
});
