import { createFileRoute, Outlet, useRouterState } from "@tanstack/react-router";
import { SidebarProvider, SidebarTrigger, SidebarInset } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { ConnectionBadge } from "@/components/connection-badge";
import { Separator } from "@/components/ui/separator";
import { Toaster } from "@/components/ui/sonner";

export const Route = createFileRoute("/_app")({
  component: AppLayout,
});

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/devices": "Devices",
  "/diagnostics": "Diagnostics",
  "/updates": "Updates",
  "/settings": "Settings",
};

function AppLayout() {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const title = titles[pathname] ?? "Vyzorix";

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
          <SidebarTrigger />
          <Separator orientation="vertical" className="mx-1 h-5" />
          <h1 className="text-sm font-semibold">{title}</h1>
          <div className="ml-auto flex items-center gap-2">
            <ConnectionBadge state="connected" />
          </div>
        </header>
        <main className="flex-1 p-4 md:p-6">
          <Outlet />
        </main>
        <Toaster />
      </SidebarInset>
    </SidebarProvider>
  );
}