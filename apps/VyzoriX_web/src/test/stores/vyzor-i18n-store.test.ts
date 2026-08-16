import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useVyzorI18nStore } from '@/stores/vyzor-i18n-store';
import { VYZOR_I18N_FALLBACK_LNG } from '@/lib/i18n';

describe('vyzor-i18n-store', () => {
  beforeEach(() => {
    localStorage.clear();
    useVyzorI18nStore.setState({ locale: VYZOR_I18N_FALLBACK_LNG });
  });

  it('defaults to the fallback language', () => {
    useVyzorI18nStore.setState({ locale: VYZOR_I18N_FALLBACK_LNG });
    expect(useVyzorI18nStore.getState().locale).toBe('en');
  });

  it('sets locale', () => {
    useVyzorI18nStore.getState().setLocale('fr');
    expect(useVyzorI18nStore.getState().locale).toBe('fr');
  });

  it('persists locale to localStorage', () => {
    useVyzorI18nStore.getState().setLocale('fr');
    const stored = JSON.parse(localStorage.getItem('vyzorix-i18n-locale') ?? '{}');
    expect(stored.state.locale).toBe('fr');
  });

  it('hydrates persisted locale', async () => {
    localStorage.clear();
    localStorage.setItem(
      'vyzorix-i18n-locale',
      JSON.stringify({ state: { locale: 'fr' }, version: 0 }),
    );
    vi.resetModules();
    const { useVyzorI18nStore: freshStore } = await import('@/stores/vyzor-i18n-store');
    await vi.waitFor(() => {
      expect(freshStore.getState().locale).toBe('fr');
    });
    vi.resetModules();
  });

  it('can switch back to en', () => {
    useVyzorI18nStore.getState().setLocale('fr');
    useVyzorI18nStore.getState().setLocale('en');
    expect(useVyzorI18nStore.getState().locale).toBe('en');
  });
});
