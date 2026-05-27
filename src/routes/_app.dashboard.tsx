import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Smartphone, Wifi, AlertTriangle, Activity, ThermometerSun, Cpu } from "lucide-react";
import { mockDevices, mockAlerts } from "@/lib/mock-data";
import { StatusBadge } from "@/components/status-badge";
import { ResponsiveContainer, AreaChart, Area, XAxis, YAxis, Tooltip, CartesianGrid } from "recharts";

export const Route = createFileRoute("/_app/dashboard")({
  head: () => ({ meta: [{ title: "Dashboard — Vyzorix" }] }),
  component: DashboardPage,
});

const trend = Array.from({ length: 24 }, (_, i) => ({
  t: `${i}:00`,
  online: 6 + Math.round(Math.sin(i / 3) * 2),
  risk: 20 + Math.round(Math.abs(Math.sin(i / 2)) * 40),
}));

function DashboardPage() {
  const online = mockDevices.filter((d) => d.status === "online").length;
  const warn = mockDevices.filter((d) => d.status === "warning").length;
  const crit = mockDevices.filter((d) => d.status === "critical").length;
  const avgRisk = Math.round(
    mockDevices.reduce((a, d) => a + d.riskScore, 0) / mockDevices.length,
  );
  const avgTemp = (
    mockDevices.filter((d) => d.status !== "offline").reduce((a, d) => a + d.thermalTemp, 0) /
    Math.max(1, mockDevices.filter((d) => d.status !== "offline").length)
  ).toFixed(1);

  const metrics = [
    { label: "Total devices", value: mockDevices.length, icon: Smartphone, hint: `${online} online` },
    { label: "Online", value: online, icon: Wifi, hint: `${mockDevices.length - online} not reachable` },
    { label: "Avg risk score", value: avgRisk, icon: Activity, hint: avgRisk > 60 ? "Above threshold" : "Healthy" },
    { label: "Avg thermal", value: `${avgTemp}°C`, icon: ThermometerSun, hint: "Last 5 min" },
    { label: "Warnings", value: warn, icon: AlertTriangle, hint: `${crit} critical` },
    { label: "Fleet CPU", value: "34%", icon: Cpu, hint: "Avg load" },
  ];

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
        {metrics.map((m) => (
          <Card key={m.label}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">{m.label}</CardTitle>
              <m.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold">{m.value}</div>
              <p className="text-xs text-muted-foreground">{m.hint}</p>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Fleet activity</CardTitle>
            <CardDescription>Online devices and average risk score over 24h</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-64 w-full">
              <ResponsiveContainer width="100%" height="100%">
                <AreaChart data={trend} margin={{ top: 10, right: 10, left: -10, bottom: 0 }}>
                  <defs>
                    <linearGradient id="g1" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="0%" stopColor="var(--primary)" stopOpacity={0.35} />
                      <stop offset="100%" stopColor="var(--primary)" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                  <XAxis dataKey="t" stroke="var(--muted-foreground)" fontSize={11} />
                  <YAxis stroke="var(--muted-foreground)" fontSize={11} />
                  <Tooltip contentStyle={{ background: "var(--popover)", border: "1px solid var(--border)", borderRadius: 8, fontSize: 12 }} />
                  <Area type="monotone" dataKey="online" stroke="var(--primary)" fill="url(#g1)" strokeWidth={2} />
                  <Area type="monotone" dataKey="risk" stroke="var(--muted-foreground)" fill="transparent" strokeWidth={1.5} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System alerts</CardTitle>
            <CardDescription>Recent events across fleet</CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-64 pr-3">
              <ul className="space-y-3">
                {mockAlerts.map((a) => (
                  <li key={a.id} className="flex items-start gap-3 text-sm">
                    <Badge
                      variant={a.severity === "critical" ? "destructive" : a.severity === "warning" ? "secondary" : "outline"}
                      className="mt-0.5 uppercase text-[10px]"
                    >
                      {a.severity}
                    </Badge>
                    <div className="flex-1">
                      <p className="leading-snug">{a.message}</p>
                      <p className="text-xs text-muted-foreground">{a.device} · {a.at}</p>
                    </div>
                  </li>
                ))}
              </ul>
            </ScrollArea>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Devices at a glance</CardTitle>
          <CardDescription>Top of the fleet — open Devices for the full list</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2">
          {mockDevices.slice(0, 5).map((d, i) => (
            <div key={d.id}>
              <div className="flex items-center justify-between gap-3 py-1.5 text-sm">
                <div className="min-w-0 flex-1">
                  <p className="truncate font-medium">{d.name}</p>
                  <p className="truncate text-xs text-muted-foreground">{d.model} · Android {d.androidVersion} · {d.appVersion}</p>
                </div>
                <div className="hidden gap-6 text-xs text-muted-foreground sm:flex">
                  <span>risk <span className="text-foreground">{d.riskScore}</span></span>
                  <span>{d.thermalTemp ? `${d.thermalTemp}°C` : "—"}</span>
                </div>
                <StatusBadge status={d.status} />
              </div>
              {i < 4 && <Separator />}
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}