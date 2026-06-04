import { createFileRoute, Outlet, redirect, useRouterState } from "@tanstack/react-router";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { ConnectionBadge } from "@/components/connection-badge";
import { Separator } from "@/components/ui/separator";
import { Toaster } from "@/components/ui/sonner";
import { DeviceStreamProvider, useStream } from "@/lib/device-stream-context";
import { LogDock } from "@/components/logs/log-dock";
import { supabase } from "@/integrations/supabase/client";
import { ensureAdminAccess } from "@/lib/admin.functions";
import { toast } from "sonner";

export const Route = createFileRoute("/_app")({
  ssr: false,
  beforeLoad: async ({ location }) => {
    const { data, error } = await supabase.auth.getUser();
    if (error || !data.user) {
      throw redirect({ to: "/login", search: { redirect: location.href } });
    }
    try {
      const res = await ensureAdminAccess();
      if (!res.allowed) {
        await supabase.auth.signOut();
        toast.error("This account is not authorized.");
        throw redirect({ to: "/login" });
      }
    } catch (e) {
      if (e && typeof e === "object" && "to" in e) throw e;
      await supabase.auth.signOut();
      throw redirect({ to: "/login" });
    }
  },
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

function AppLayout() {
  return (
    <SidebarProvider>
      <DeviceStreamProvider>
        <AppShell />
      </DeviceStreamProvider>
    </SidebarProvider>
  );
}

function AppShell() {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const title =
    titles[pathname] ??
    (pathname.startsWith("/settings") ? "Settings" : "Vyzorix");
  const { state } = useStream();

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
          <LogDock />
        </main>
        <Toaster />
      </SidebarInset>
    </>
  );
}
