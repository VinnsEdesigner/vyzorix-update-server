import { Link, useNavigate, useRouterState } from "@tanstack/react-router";
import {
  LayoutDashboard,
  Smartphone,
  Activity,
  PackageCheck,
  Settings,
  Shield,
  Bell,
  Terminal,
  LogOut,
} from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { getStoredOperator, logout } from "@/lib/vyzorix-auth";
import { DEFAULT_SERVER_URL } from "@/lib/vyzorix-config";

const navItems = [
  { title: "Dashboard", url: "/dashboard", icon: LayoutDashboard },
  { title: "Device", url: "/device", icon: Smartphone },
  { title: "Diagnostics", url: "/diagnostics", icon: Activity },
  { title: "Alerts", url: "/alerts", icon: Bell },
  { title: "Updates", url: "/updates", icon: PackageCheck },
  { title: "Logs", url: "/logs", icon: Terminal },
  { title: "Settings", url: "/settings", icon: Settings },
];

const OPERATOR_UPDATE_EVENT = "vyz.operator.updated";

// eslint-disable-next-line func-style
export function AppSidebar(): JSX.Element {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  const navigate = useNavigate();
  const [operator, setOperator] = useState(getStoredOperator);

  // Sync operator from localStorage and listen for name updates from settings
  useEffect(() => {
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const sync = () => setOperator(getStoredOperator());
    // Initial sync
    sync();
    // Listen for name updates from the operator settings page
    window.addEventListener(OPERATOR_UPDATE_EVENT, sync);
    return () => window.removeEventListener(OPERATOR_UPDATE_EVENT, sync);
  }, []);

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const signOut = async () => {
    try {
      await logout(DEFAULT_SERVER_URL);
    } catch {
      // ignore logout error - sign out should still proceed
    }
    toast.success("Signed out");
    navigate({ to: "/login", replace: true });
  };

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2 px-2 py-1.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
            <Shield className="h-4 w-4" />
          </div>
          <div className="flex flex-col leading-tight group-data-[collapsible=icon]:hidden">
            <span className="text-sm font-semibold">Vyzorix</span>
            <span className="text-xs text-muted-foreground">Update Server</span>
          </div>
        </div>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>Operations</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {navItems.map((item) => (
                <SidebarMenuItem key={item.url}>
                  <SidebarMenuButton asChild isActive={pathname === item.url}>
                    <Link to={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter>
        <div className="space-y-1 px-2 py-1.5 group-data-[collapsible=icon]:hidden">
          {operator?.email && (
            <p className="truncate text-xs text-muted-foreground" title={operator.email}>
              {operator.email}
            </p>
          )}
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-full justify-start gap-2 px-2 text-xs"
            onClick={signOut}
          >
            <LogOut className="h-3.5 w-3.5" /> Sign out
          </Button>
        </div>
        <Button
          variant="ghost"
          size="icon"
          className="hidden h-8 w-8 group-data-[collapsible=icon]:flex"
          onClick={signOut}
          title="Sign out"
        >
          <LogOut className="h-4 w-4" />
        </Button>
      </SidebarFooter>
    </Sidebar>
  );
}
