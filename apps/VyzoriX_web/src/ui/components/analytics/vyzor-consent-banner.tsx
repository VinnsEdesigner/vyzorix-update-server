import { useVyzorConsent } from '@/hooks/analytics';

export function VyzorConsentBanner() {
  const { consent, bannerDismissed, grant, deny } = useVyzorConsent();

  if (consent !== 'pending' || bannerDismissed) return null;

  return (
    <div
      role="dialog"
      aria-label="Analytics consent"
      className="fixed bottom-0 left-0 right-0 z-50 flex items-center justify-between gap-4 border-t border-border bg-card p-4 shadow-lg"
    >
      <p className="text-sm text-card-foreground">
        We use anonymous analytics to improve VyzoriX. No personal data is collected.
      </p>
      <div className="flex shrink-0 gap-2">
        <button
          type="button"
          onClick={deny}
          className="rounded-md border border-border px-3 py-1.5 text-sm text-muted-foreground hover:bg-muted"
        >
          Decline
        </button>
        <button
          type="button"
          onClick={grant}
          className="rounded-md bg-primary px-3 py-1.5 text-sm text-primary-foreground hover:bg-primary/90"
        >
          Accept
        </button>
      </div>
    </div>
  );
}
