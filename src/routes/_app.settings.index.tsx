import { createFileRoute, Link } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ArrowRight, Bell, Cable, Palette, ShieldCheck, SlidersHorizontal, Wrench } from "lucide-react";
import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { useServerHealth } from "@/hooks/use-server-health";

export const Route = createFileRoute("/_app/settings/")({
  component: SettingsOverview,
});

const sections = [
  {
    to: "/settings/connection",
    icon: Cable,
    title: "Connection",
    description: "Update server URL, target device ID, request timeouts, auto-reconnect.",
  },
  {
    to: "/settings/operator",
    icon: ShieldCheck,
    title: "Operator profile",
    description: "Your identity, role and super-user controls. Audit-trail metadata.",
  },
  {
    to: "/settings/thresholds",
    icon: SlidersHorizontal,
    title: "Thresholds",
    description: "Risk, thermal and buffer thresholds that drive alerts and the dashboard.",
  },
  {
    to: "/settings/notifications",
    icon: Bell,
    title: "Notifications",
    description: "Toast and browser-push behaviour for alerts and command results.",
  },
  {
    to: "/settings/appearance",
    icon: Palette,
    title: "Appearance",
    description: "Theme (system / light / dark) for this browser.",
  },
  {
    to: "/settings/advanced",
    icon: Wrench,
    title: "Advanced",
    description: "Signal history depth, log retention, strict HMAC, reset everything.",
  },
] as const;

function SettingsOverview() {
  const { serverUrl, deviceId, operator } = useVyzorixConfig();
  const health = useServerHealth(serverUrl);

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-base">Active configuration</CardTitle>
          <CardDescription>What the rest of the dashboard is currently using</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <KV k="Server URL" v={serverUrl} />
          <KV k="Device ID" v={deviceId} />
          <KV k="Operator" v={operator.name || "—"} />
          <KV k="Role" v={operator.role} />
          <KV k="Health" v={health.data?.ok ? "ok" : health.isError ? "down" : "checking"} />
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-2">
        {sections.map((s) => (
          <Link key={s.to} to={s.to as "/settings/connection"} className="group">
            <Card className="h-full transition-colors group-hover:border-primary">
              <CardHeader className="flex flex-row items-start justify-between gap-3 space-y-0">
                <div className="flex items-start gap-3">
                  <s.icon className="mt-0.5 h-5 w-5 text-primary" />
                  <div>
                    <CardTitle className="text-base">{s.title}</CardTitle>
                    <CardDescription>{s.description}</CardDescription>
                  </div>
                </div>
                <ArrowRight className="h-4 w-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
              </CardHeader>
            </Card>
          </Link>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Mock server quick start</CardTitle>
          <CardDescription>From the repo root</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs">go run ./cmd/mockserver -addr=:8080 -data=./cmd/mockserver/testdata</pre>
          <div className="flex items-center gap-2 pt-1">
            <Badge variant="outline">Phase 1</Badge>
            <Badge variant="outline">single device</Badge>
            <Badge variant="outline">Nokia C22</Badge>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium break-all">{v}</p>
    </div>
  );
}