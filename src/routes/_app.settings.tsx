import { createFileRoute, Link, Outlet, useRouterState } from "@tanstack/react-router";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/_app/settings")({
  head: () => ({ meta: [{ title: "Settings — Vyzorix" }] }),
  component: SettingsLayout,
});

const tabs = [
  { to: "/settings", label: "Overview", exact: true },
  { to: "/settings/connection", label: "Connection" },
  { to: "/settings/operator", label: "Operator" },
  { to: "/settings/thresholds", label: "Thresholds" },
  { to: "/settings/advanced", label: "Advanced" },
] as const;

function SettingsLayout() {
  const pathname = useRouterState({ select: (r) => r.location.pathname });
  return (
    <div className="space-y-4">
      <nav className="flex flex-wrap gap-1 rounded-md border bg-card p-1 text-sm">
        {tabs.map((t) => {
          const active = t.exact ? pathname === t.to : pathname === t.to || pathname.startsWith(t.to + "/");
          return (
            <Link
              key={t.to}
              to={t.to}
              className={cn(
                "rounded px-3 py-1.5 transition-colors",
                active ? "bg-primary text-primary-foreground" : "text-muted-foreground hover:bg-muted hover:text-foreground",
              )}
            >
              {t.label}
            </Link>
          );
        })}
      </nav>
      <Outlet />
    </div>
  );
}