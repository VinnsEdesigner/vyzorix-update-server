import { createFileRoute, Outlet, useRouterState } from "@tanstack/react-router";

import { AppSidebar } from "@/components/app-sidebar";
import { ConnectionBadge } from "@/components/connection-badge";
import { LogDock } from "@/components/logs/log-dock";
import { Separator } from "@/components/ui/separator";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/sonner";
import { DeviceStreamProvider, useStream } from "@/lib/device-stream-context";

export const Route = createFileRoute("/_app")({
  ssr: false,
  // Auth temporarily disabled for local exploration. Re-enable by restoring the
  // beforeLoad guard below once Google sign-in is configured.
  component: AppLayout,
});

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/device": "Device",
  "/diagnostics": "Diagnostics",
  "/alerts": "System alerts",
  "/updates": "Updates",
  "/logs": "Logs",
  "/settings": "Settings",
};

// eslint-disable-next-line func-style
function AppLayout(): JSX.Element {
  return (
    <SidebarProvider>
      <DeviceStreamProvider>
        <AppShell />
      </DeviceStreamProvider>
    </SidebarProvider>
  );
}

// eslint-disable-next-line func-style
function AppShell(): JSX.Element {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const title = titles[pathname] ?? (pathname.startsWith("/settings") ? "Settings" : "Vyzorix");
  const { state } = useStream();
  const isLogsPage = pathname === "/logs";

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
}
