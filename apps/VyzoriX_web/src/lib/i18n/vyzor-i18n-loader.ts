import type { i18n } from 'i18next';
import type { VyzorLocaleResources } from './vyzor-i18n-types';

type LocaleLoader = () => Promise<VyzorLocaleResources>;

const localeLoaders: Record<string, LocaleLoader> = {
  en: async () => {
    const [common, updates] = await Promise.all([
      import('./locales/en/vyzor-common-en'),
      import('./locales/en/vyzor-updates-en'),
    ]);
    return {
      common: common.vyzorCommonEn as unknown as Record<string, string>,
      updates: updates.vyzorUpdatesEn as unknown as Record<string, string>,
    };
  },
  fr: async () => {
    const [common, updates] = await Promise.all([
      import('./locales/fr/vyzor-common-fr'),
      import('./locales/fr/vyzor-updates-fr'),
    ]);
    return {
      common: common.vyzorCommonFr as unknown as Record<string, string>,
      updates: updates.vyzorUpdatesFr as unknown as Record<string, string>,
    };
  },
};

export async function loadLocale(locale: string): Promise<VyzorLocaleResources | null> {
  const loader = localeLoaders[locale];
  if (!loader) return null;
  return loader();
}

export async function loadLocaleInto(
  i18nInstance: i18n,
  locale: string,
): Promise<boolean> {
  const resources = await loadLocale(locale);
  if (!resources) return false;

  for (const ns of Object.keys(resources) as Array<keyof VyzorLocaleResources>) {
    if (!i18nInstance.hasResourceBundle(locale, ns)) {
      i18nInstance.addResourceBundle(locale, ns, resources[ns]);
    }
  }
  return true;
}
