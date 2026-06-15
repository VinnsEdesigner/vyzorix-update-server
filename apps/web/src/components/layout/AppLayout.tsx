import { useNavigate, useRouterState } from "@tanstack/react-router";
import { useEffect, useState, type ReactElement, type ReactNode } from "react";

import { AppSidebar } from "@/components/app-sidebar";
import SpinningBlocksLoader from "@/components/auth/SpinningBlocksLoader";
import { ConnectionBadge } from "@/components/connection-badge";
import { LogDock } from "@/components/logs/log-dock";
import { Separator } from "@/components/ui/separator";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/sonner";
import { useAuth } from "@/hooks/use-auth";
import { DeviceStreamProvider, useStream } from "@/lib/device-stream-context";

interface AppLayoutProps {
  children: ReactNode;
  title?: string;
}

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/device": "Device",
  "/diagnostics": "Diagnostics",
  "/alerts": "System alerts",
  "/updates": "Updates",
  "/logs": "Logs",
  "/settings": "Settings",
};

/**
 * AppLayout - Shared layout component for protected routes
 * 
 * Features:
 * - Sidebar navigation
 * - Auth check with redirect to login
 * - SpinningBlocksLoader during auth verification
 * - Connection status badge
 * - Log dock (except on logs page)
 */
export const AppLayout = ({ children, title }: AppLayoutProps): ReactElement => {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const pageTitle = title ?? titles[pathname] ?? (pathname.startsWith("/settings") ? "Settings" : "Vyzorix");
  const { state } = useStream();
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  const [checked, setChecked] = useState(false);
  const [showSpinner, setShowSpinner] = useState(true);

  const isLogsPage = pathname === "/logs";

  // Minimum spinner display time for smooth UX
  useEffect(() => {
    const timer = setTimeout(() => {
      setShowSpinner(false);
    }, 1500);
    return () => clearTimeout(timer);
  }, []);

  // Client-side auth check - redirect to login if not authenticated
  useEffect(() => {
    if (checked) return;

    if (!isLoading) {
      setChecked(true);
      if (!isAuthenticated) {
        navigate({ to: "/login", replace: true });
      }
    }
  }, [isAuthenticated, isLoading, checked, navigate]);

  // Show loading spinner during minimum display time
  if (showSpinner) {
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

  // Still loading auth check - show spinner
  if (!checked && isLoading) {
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
    <SidebarProvider>
      <DeviceStreamProvider>
        <AppSidebar />
        <SidebarInset>
          <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
            <SidebarTrigger />
            <Separator orientation="vertical" className="mx-1 h-5" />
            <h1 className="text-sm font-semibold">{pageTitle}</h1>
            <div className="ml-auto flex items-center gap-2">
              <ConnectionBadge state={state} />
            </div>
          </header>
          <main className="flex-1 p-4 pb-14 md:p-6 md:pb-14">
            {children}
            {!isLogsPage && <LogDock />}
          </main>
          <Toaster />
        </SidebarInset>
      </DeviceStreamProvider>
    </SidebarProvider>
  );
};

export default AppLayout;
