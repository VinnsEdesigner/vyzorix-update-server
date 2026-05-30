import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Activity, ThermometerSun, Volume2, Wifi, Clock, ShieldAlert, Cpu, AudioLines } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { ResponsiveContainer, LineChart, Line, CartesianGrid, XAxis, YAxis, Tooltip, ReferenceLine } from "recharts";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { useDeviceStream } from "@/hooks/use-device-stream";
import { useServerHealth } from "@/hooks/use-server-health";
import { getDeviceStatus, getVersion } from "@/lib/vyzorix-api";
import { StatusBadge, type DeviceHealth } from "@/components/status-badge";
import { formatRelative, formatUptime } from "@/lib/format";

export const Route = createFileRoute("/_app/dashboard")({
  head: () => ({ meta: [{ title: "Dashboard — Vyzorix" }] }),
  component: DashboardPage,
});

const tip = { background: "var(--popover)", border: "1px solid var(--border)", borderRadius: 8, fontSize: 12 };

function deriveHealth(online: boolean, riskScore?: number, thermal?: number): DeviceHealth {
  if (!online) return "offline";
  if ((riskScore ?? 0) >= 75 || (thermal ?? 0) >= 55) return "critical";
  if ((riskScore ?? 0) >= 50 || (thermal ?? 0) >= 45) return "warning";
  return "online";
}

function DashboardPage() {
  const { serverUrl, deviceId } = useVyzorixConfig();
  const health = useServerHealth(serverUrl);
  const stream = useDeviceStream(serverUrl, deviceId);
  const t = stream.lastTelemetry;

  const status = useQuery({
    queryKey: ["vyzorix", "status", serverUrl, deviceId],
    queryFn: () => getDeviceStatus(serverUrl, deviceId),
    enabled: !!serverUrl && !!deviceId && health.data?.ok === true,
    refetchInterval: 15_000,
    retry: false,
  });

  const version = useQuery({
    queryKey: ["vyzorix", "version", serverUrl],
    queryFn: () => getVersion(serverUrl),
    enabled: health.data?.ok === true,
    retry: false,
  });

  const online = status.data?.online ?? (stream.state === "connected");
  const deviceHealth = deriveHealth(online, t?.riskScore, t?.thermalTemp);

  const riskSeries = stream.telemetryHistory.map((f, i) => ({ i, v: f.riskScore ?? 0 }));
  const thermalSeries = stream.telemetryHistory.map((f, i) => ({ i, v: f.thermalTemp ?? 0 }));

  return (
    <div className="space-y-4">
      {/* Connection banner */}
      {health.data?.ok === false || health.isError ? (
        <Card className="border-destructive/40">
          <CardContent className="flex items-start gap-3 py-4 text-sm">
            <ShieldAlert className="mt-0.5 h-4 w-4 text-destructive" />
            <div className="space-y-1">
              <p className="font-medium">Cannot reach update server</p>
              <p className="text-muted-foreground">
                Tried <code className="text-xs">{serverUrl}/healthz</code>. Start it with{" "}
                <code className="text-xs">go run ./cmd/mockserver</code> or change the URL in Settings.
              </p>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {/* Hero device card */}
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              Nokia C22 <span className="text-xs font-normal text-muted-foreground">· {deviceId}</span>
            </CardTitle>
            <CardDescription>
              VyzorixAudioRouter daemon · {status.data?.appVersion ?? version.data?.version ?? "unknown build"} · {status.data?.deviceClass ?? "nokia_c22"}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <StatusBadge status={deviceHealth} />
            <Badge variant="outline" className="gap-1.5">
              <Wifi className="h-3 w-3" />
              {stream.state}
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-4">
          <Metric icon={Activity} label="Risk score" value={t?.riskScore != null ? `${t.riskScore}` : "—"} hint={t?.riskScore != null ? (t.riskScore >= 75 ? "Critical — soft reboot predicted" : t.riskScore >= 50 ? "Investigate" : "Healthy") : "Awaiting telemetry"} />
          <Metric icon={ThermometerSun} label="Thermal" value={t?.thermalTemp != null ? `${t.thermalTemp.toFixed(1)}°C` : "—"} hint={t?.thermalTemp != null ? (t.thermalTemp >= 55 ? "THROTTLE_HEAVY" : t.thermalTemp >= 45 ? "THROTTLE_LIGHT" : "NONE") : "Awaiting telemetry"} />
          <Metric icon={Clock} label="Uptime" value={formatUptime(t?.uptime)} hint={`Last seen ${formatRelative(status.data?.lastSeen)}`} />
          <Metric icon={Volume2} label="Speaker" value={t?.speakerOn == null ? "—" : t.speakerOn ? "FORCED" : "OFF"} hint={t?.activeDevice ?? "—"} />
        </CardContent>
      </Card>

      {/* Live charts */}
      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Risk score · live</CardTitle>
            <CardDescription>{stream.telemetryHistory.length} frames buffered</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartShell data={riskSeries} thresholds={[50, 75]} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Thermal · live (°C)</CardTitle>
            <CardDescription>From TelemetryFrame.thermalTemp</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartShell data={thermalSeries} thresholds={[45, 55]} />
          </CardContent>
        </Card>
      </div>

      {/* Route + capture summary */}
      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><AudioLines className="h-4 w-4" /> Route state</CardTitle></CardHeader>
          <CardContent className="space-y-3 text-sm">
            <KV k="Active device" v={t?.activeDevice ?? "—"} />
            <KV k="Audio mode" v={t?.audioMode != null ? `${t.audioMode}` : "—"} />
            <KV k="Speaker" v={t?.speakerOn == null ? "—" : t.speakerOn ? "FORCED" : "OFF"} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-base flex items-center gap-2"><Cpu className="h-4 w-4" /> Capture buffer</CardTitle></CardHeader>
          <CardContent className="space-y-3">
            <Progress value={t?.bufferLevel ?? 0} />
            <p className="text-xs text-muted-foreground">{t?.bufferLevel ?? 0}% fill · underrun threshold 50%</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-base">Update server</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <KV k="Latest version" v={version.data?.version ?? "—"} />
            <KV k="Version code" v={version.data?.version_code != null ? `${version.data.version_code}` : "—"} />
            <KV k="Health" v={health.data?.ok ? "ok" : "down"} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function Metric({ icon: Icon, label, value, hint }: { icon: typeof Activity; label: string; value: string; hint?: string }) {
  return (
    <div className="rounded-md border p-3">
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <span>{label}</span>
        <Icon className="h-3.5 w-3.5" />
      </div>
      <p className="mt-1 text-xl font-semibold">{value}</p>
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-xs text-muted-foreground">{k}</span>
      <span className="font-mono text-xs">{v}</span>
    </div>
  );
}

function ChartShell({ data, thresholds }: { data: { i: number; v: number }[]; thresholds?: number[] }) {
  if (data.length === 0) {
    return <div className="flex h-48 items-center justify-center text-xs text-muted-foreground">Waiting for live telemetry…</div>;
  }
  return (
    <div className="h-48 w-full">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
          <XAxis dataKey="i" hide />
          <YAxis stroke="var(--muted-foreground)" fontSize={10} />
          <Tooltip contentStyle={tip} />
          {thresholds?.map((y) => (
            <ReferenceLine key={y} y={y} stroke="var(--muted-foreground)" strokeDasharray="3 3" />
          ))}
          <Line type="monotone" dataKey="v" stroke="var(--primary)" dot={false} strokeWidth={2} isAnimationActive={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
