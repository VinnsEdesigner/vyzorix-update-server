import { Outlet, createRootRoute } from '@tanstack/react-router';
import { QueryProvider } from '@/providers/query-provider';
import { VyzorI18nProvider } from '@/providers/vyzor-i18n-provider';
import { VyzorAnalyticsProvider } from '@/providers/vyzor-analytics-provider';
import { useWebSocketConnection } from '@/hooks/realtime/use-realtime';
import { useInitConnectivity } from '@/hooks/connectivity/use-connectivity';
import { VyzorErrorBoundary } from '@/ui/components/error/vyzor-error-boundary';
import { VyzorRouteError } from '@/ui/components/error/vyzor-route-error';
import { VyzorConsentBanner } from '@/ui/components/analytics/vyzor-consent-banner';

function RootComponent() {
  useWebSocketConnection();
  useInitConnectivity();

  return (
    <QueryProvider>
      <VyzorI18nProvider>
        <VyzorAnalyticsProvider>
          <VyzorErrorBoundary>
            <Outlet />
          </VyzorErrorBoundary>
          <VyzorConsentBanner />
        </VyzorAnalyticsProvider>
      </VyzorI18nProvider>
    </QueryProvider>
  );
}

export const Route = createRootRoute({
  component: RootComponent,
  errorComponent: ({ error }) => <VyzorRouteError error={error} />,
});
