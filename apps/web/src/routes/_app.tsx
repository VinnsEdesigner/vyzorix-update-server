import { createFileRoute, Outlet, useNavigate, useRouterState } from "@tanstack/react-router";
import { useEffect, useState, type ReactElement } from "react";

import { AppSidebar } from "@/components/app-sidebar";
import SpinningBlocksLoader from "@/components/auth/SpinningBlocksLoader";
import { ConnectionBadge } from "@/components/connection-badge";
import { LogDock } from "@/components/logs/log-dock";
import { Separator } from "@/components/ui/separator";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/sonner";
import { useAuth } from "@/hooks/use-auth";
import { DeviceStreamProvider, useStream } from "@/lib/device-stream-context";

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/device": "Device",
  "/diagnostics": "Diagnostics",
  "/alerts": "System alerts",
  "/updates": "Updates",
  "/logs": "Logs",
  "/settings": "Settings",
};

const AppLayout = (): ReactElement => {
  return (
    <SidebarProvider>
      <DeviceStreamProvider>
        <AppShell />
      </DeviceStreamProvider>
    </SidebarProvider>
  );
};

const AppShell = (): ReactElement => {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const title = titles[pathname] ?? (pathname.startsWith("/settings") ? "Settings" : "Vyzorix");
  const { state } = useStream();
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  const [checked, setChecked] = useState(false);

  const isLogsPage = pathname === "/logs";

  // Client-side auth check - redirect to login if not authenticated
  useEffect(() => {
    if (checked) return;

    if (!isLoading) {
      setChecked(true);
      if (!isAuthenticated) {
        navigate({ to: "/auth/login", replace: true });
      }
    }
  }, [isAuthenticated, isLoading, checked, navigate]);

  // Show loading state while checking auth
  if (isLoading || !checked) {
    return (
      <SidebarProvider>
        <DeviceStreamProvider>
          <div className="flex h-screen items-center justify-center">
            <SpinningBlocksLoader />
          </div>
        </DeviceStreamProvider>
      </SidebarProvider>
    );
  }

  // Not authenticated - will redirect
  if (!isAuthenticated) {
    return (
      <SidebarProvider>
        <DeviceStreamProvider>
          <div className="flex h-screen items-center justify-center">
            <SpinningBlocksLoader />
          </div>
        </DeviceStreamProvider>
      </SidebarProvider>
    );
  }

  return (
    <>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <SidebarTrigger />
          <Separator orientation="vertical" className="mx-1 h-5" />
          <h1 className="text-sm font-semibold">{title}</h1>
          <div className="ml-auto flex items-center gap-2">
            <ConnectionBadge state={state} />
          </div>
        </header>
        <main className="flex-1 p-4 pb-14 md:p-6 md:pb-14">
          <Outlet />
          {!isLogsPage && <LogDock />}
        </main>
        <Toaster />
      </SidebarInset>
    </>
  );
};

/**
 * Server-side authentication check for protected routes
 *
 * This beforeLoad hook runs on both server and client:
 * - Server: Uses middleware-injected context (if available)
 * - Client: Checks server-injected state (SSR hydration) OR falls back to /v1/auth/me API call
 *
 * Based on Library's SSR pattern for auth checking
 */
export const Route = createFileRoute("/_app")({
  beforeLoad: () => {
    // Client-side check: verify session with server using cookie
    if (typeof window !== "undefined") {
      // Import dynamically to avoid SSR issues
      const globalState = (
        window as unknown as { __VYZORIX_PREFETCHED_STATE__?: { isAuthenticated?: boolean } }
      ).__VYZORIX_PREFETCHED_STATE__;
      if (globalState?.isAuthenticated) {
        // Server already validated the session via SSR - user is authenticated, don't redirect
        // eslint-disable-next-line no-useless-return
        return;
      }

      // No SSR state - let the component handle auth check via useAuth hook
      // This prevents the redirect here and allows the hook to show loading state
    }

    // Server-side: For now, allow access and let the route handle auth
    // The SSR state injection will provide the actual auth check
  },
  component: AppLayout,
});
