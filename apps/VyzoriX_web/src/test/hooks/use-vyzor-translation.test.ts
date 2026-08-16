import { describe, it, expect } from 'vitest';
import { createVyzorI18n } from '@/lib/i18n';

describe('vyzor-i18n-config', () => {
  it('creates an initialized i18n instance', () => {
    const instance = createVyzorI18n();
    expect(instance.isInitialized).toBe(true);
  });

  it('defaults to English', () => {
    const instance = createVyzorI18n();
    expect(instance.language).toBe('en');
  });

  it('has common namespace loaded', () => {
    const instance = createVyzorI18n();
    expect(instance.t('common:ok')).toBe('OK');
    expect(instance.t('common:cancel')).toBe('Cancel');
  });

  it('has updates namespace loaded', () => {
    const instance = createVyzorI18n();
    expect(instance.t('updates:title')).toBe('Updates');
    expect(instance.t('updates:pushUpdate')).toBe('Push Update');
  });

  it('supports interpolation', () => {
    const instance = createVyzorI18n();
    expect(instance.t('updates:deviceCount', { count: 5 })).toBe('5 devices');
  });

  it('returns key when translation missing', () => {
    const instance = createVyzorI18n();
    expect(instance.t('updates:nonexistent_key')).toBe('nonexistent_key');
  });

  it('supports supportedLngs', () => {
    const instance = createVyzorI18n();
    expect(instance.options.supportedLngs).toContain('en');
    expect(instance.options.supportedLngs).toContain('fr');
  });
});

describe('vyzor-i18n-loader', () => {
  it('loads English locale resources', async () => {
    const { loadLocale } = await import('@/lib/i18n');
    const resources = await loadLocale('en');
    expect(resources).not.toBeNull();
    expect(resources?.common.ok).toBe('OK');
    expect(resources?.updates.title).toBe('Updates');
  });

  it('loads French locale resources', async () => {
    const { loadLocale } = await import('@/lib/i18n');
    const resources = await loadLocale('fr');
    expect(resources).not.toBeNull();
    expect(resources?.common.cancel).toBe('Annuler');
    expect(resources?.updates.title).toBe('Mises à jour');
  });

  it('returns null for unsupported locale', async () => {
    const { loadLocale } = await import('@/lib/i18n');
    const resources = await loadLocale('de');
    expect(resources).toBeNull();
  });

  it('loads locale into i18n instance and changes language', async () => {
    const { changeVyzorLanguage, createVyzorI18n } = await import('@/lib/i18n');
    const instance = createVyzorI18n();
    const ok = await changeVyzorLanguage(instance, 'fr');
    expect(ok).toBe(true);
    expect(instance.language).toBe('fr');
    expect(instance.t('common:cancel')).toBe('Annuler');
  });

  it('returns false for unsupported locale in changeVyzorLanguage', async () => {
    const { changeVyzorLanguage, createVyzorI18n } = await import('@/lib/i18n');
    const instance = createVyzorI18n();
    const ok = await changeVyzorLanguage(instance, 'de');
    expect(ok).toBe(false);
    expect(instance.language).toBe('en');
  });
});
