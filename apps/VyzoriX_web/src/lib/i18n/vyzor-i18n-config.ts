import { createInstance, type i18n as I18nInstance } from 'i18next';
import { initReactI18next } from 'react-i18next';
import { vyzorCommonEn } from './locales/en/vyzor-common-en';
import { vyzorUpdatesEn } from './locales/en/vyzor-updates-en';
import { loadLocaleInto } from './vyzor-i18n-loader';
import { VYZOR_I18N_NAMESPACES } from './vyzor-i18n-types';

export const VYZOR_I18N_DEFAULT_NS = 'common' as const;
export const VYZOR_I18N_FALLBACK_LNG = 'en' as const;

export const VYZOR_AVAILABLE_LANGUAGES = [
  { label: 'English', value: 'en' as const },
  { label: 'Français', value: 'fr' as const },
];

export type VyzorLanguageValue = (typeof VYZOR_AVAILABLE_LANGUAGES)[number]['value'];

export function createVyzorI18n(): I18nInstance {
  const instance = createInstance();
  instance.use(initReactI18next).init({
    resources: {
      en: {
        common: vyzorCommonEn as unknown as Record<string, string>,
        updates: vyzorUpdatesEn as unknown as Record<string, string>,
      },
    },
    lng: VYZOR_I18N_FALLBACK_LNG,
    fallbackLng: VYZOR_I18N_FALLBACK_LNG,
    supportedLngs: ['en', 'fr'],
    ns: [...VYZOR_I18N_NAMESPACES],
    defaultNS: VYZOR_I18N_DEFAULT_NS,
    fallbackNS: VYZOR_I18N_DEFAULT_NS,
    interpolation: { escapeValue: false },
    react: { useSuspense: false },
  });
  return instance;
}

let defaultI18n: I18nInstance | null = null;

export function getDefaultVyzorI18n(): I18nInstance {
  if (!defaultI18n) {
    defaultI18n = createVyzorI18n();
  }
  return defaultI18n;
}

export async function changeVyzorLanguage(
  i18nInstance: I18nInstance,
  locale: string,
): Promise<boolean> {
  const loaded = await loadLocaleInto(i18nInstance, locale);
  if (!loaded) return false;
  await i18nInstance.changeLanguage(locale);
  return true;
}
