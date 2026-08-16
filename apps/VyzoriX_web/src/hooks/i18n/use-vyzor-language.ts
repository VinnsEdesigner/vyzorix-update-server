import { useCallback } from 'react';
import {
  VYZOR_AVAILABLE_LANGUAGES,
  changeVyzorLanguage,
  type VyzorLanguageValue,
} from '@/lib/i18n';
import { useVyzorI18nStore } from '@/stores/vyzor-i18n-store';

export function useVyzorLanguage() {
  const locale = useVyzorI18nStore((s) => s.locale);
  const setLocaleStore = useVyzorI18nStore((s) => s.setLocale);

  const setLocale = useCallback(
    (next: VyzorLanguageValue) => {
      setLocaleStore(next);
    },
    [setLocaleStore],
  );

  const changeLanguage = useCallback(async (next: VyzorLanguageValue) => {
    const { getDefaultVyzorI18n } = await import('@/lib/i18n');
    const ok = await changeVyzorLanguage(getDefaultVyzorI18n(), next);
    if (ok) setLocaleStore(next);
    return ok;
  }, [setLocaleStore]);

  return {
    locale,
    setLocale,
    changeLanguage,
    available: VYZOR_AVAILABLE_LANGUAGES,
  };
}
