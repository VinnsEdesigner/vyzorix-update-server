import { createVyzorStore } from '@/lib/state';
import type { VyzorAnalyticsConsent } from '@/lib/analytics';

export interface VyzorConsentStoreState {
  consent: VyzorAnalyticsConsent;
  bannerDismissed: boolean;
  setConsent: (consent: VyzorAnalyticsConsent) => void;
  dismissBanner: () => void;
}

export const useVyzorConsentStore = createVyzorStore<VyzorConsentStoreState>(
  'VyzorConsentStore',
  (set) => ({
    consent: 'pending',
    bannerDismissed: false,
    setConsent: (consent) => set({ consent, bannerDismissed: true }),
    dismissBanner: () => set({ bannerDismissed: true }),
  }),
  {
    persist: {
      name: 'vyzorix-analytics-consent',
      partialize: (state) => ({
        consent: state.consent,
        bannerDismissed: state.bannerDismissed,
      }),
    },
  },
);
