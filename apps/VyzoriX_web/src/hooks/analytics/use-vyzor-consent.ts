import { useCallback } from 'react';
import { useVyzorConsentStore } from '@/stores/vyzor-consent-store';
import { useVyzorAnalytics } from './use-vyzor-analytics';

export function useVyzorConsent() {
  const consent = useVyzorConsentStore((s) => s.consent);
  const setConsent = useVyzorConsentStore((s) => s.setConsent);
  const dismissBanner = useVyzorConsentStore((s) => s.dismissBanner);
  const bannerDismissed = useVyzorConsentStore((s) => s.bannerDismissed);
  const { track, events } = useVyzorAnalytics();

  const grant = useCallback(() => {
    setConsent('granted');
    track(events.consentDecision, { decision: 'granted' });
  }, [setConsent, track, events]);

  const deny = useCallback(() => {
    setConsent('denied');
    track(events.consentDecision, { decision: 'denied' });
  }, [setConsent, track, events]);

  return {
    consent,
    bannerDismissed,
    grant,
    deny,
    dismissBanner,
  };
}
