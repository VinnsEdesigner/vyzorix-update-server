import { useEffect, useState, type ReactNode } from 'react';
import { I18nextProvider } from 'react-i18next';
import type { i18n as I18nInstance } from 'i18next';
import {
  getDefaultVyzorI18n,
  changeVyzorLanguage,
} from '@/lib/i18n';
import { useVyzorI18nStore } from '@/stores/vyzor-i18n-store';

interface VyzorI18nProviderProps {
  children: ReactNode;
}

export function VyzorI18nProvider({ children }: VyzorI18nProviderProps) {
  const [instance] = useState<I18nInstance>(() => getDefaultVyzorI18n());
  const locale = useVyzorI18nStore((s) => s.locale);
  const [ready, setReady] = useState(instance.isInitialized);

  useEffect(() => {
    if (!instance.isInitialized) return;
    if (instance.language === locale) {
      setReady(true);
      return;
    }
    let cancelled = false;
    changeVyzorLanguage(instance, locale).then(() => {
      if (!cancelled) setReady(true);
    });
    return () => {
      cancelled = true;
    };
  }, [instance, locale]);

  return (
    <I18nextProvider i18n={instance}>
      {ready ? children : null}
    </I18nextProvider>
  );
}
