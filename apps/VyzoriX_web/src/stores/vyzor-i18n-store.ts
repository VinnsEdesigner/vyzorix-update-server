import { createVyzorStore } from '@/lib/state';
import { VYZOR_I18N_FALLBACK_LNG, type VyzorLanguageValue } from '@/lib/i18n';

export interface VyzorI18nStoreState {
  locale: VyzorLanguageValue;
  setLocale: (locale: VyzorLanguageValue) => void;
}

export const useVyzorI18nStore = createVyzorStore<VyzorI18nStoreState>(
  'VyzorI18nStore',
  (set) => ({
    locale: VYZOR_I18N_FALLBACK_LNG,
    setLocale: (locale) => set({ locale }),
  }),
  {
    persist: {
      name: 'vyzorix-i18n-locale',
      partialize: (state) => ({ locale: state.locale }),
    },
  },
);
