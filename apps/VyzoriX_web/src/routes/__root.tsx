import { Outlet, createRootRoute } from '@tanstack/react-router';
import { QueryProvider } from '@/providers/query-provider';
import { useWebSocketConnection } from '@/hooks/realtime/use-realtime';
import { useInitConnectivity } from '@/hooks/connectivity/use-connectivity';

function RootComponent() {
  useWebSocketConnection();
  useInitConnectivity();

  return (
    <QueryProvider>
      <Outlet />
    </QueryProvider>
  );
}

export const Route = createRootRoute({
  component: RootComponent,
});
