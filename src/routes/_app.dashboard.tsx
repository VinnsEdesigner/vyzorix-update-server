import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Activity, ThermometerSun, Volume2, Wifi, Clock, ShieldAlert, Cpu, AudioLines } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { ResponsiveContainer, LineChart, Line, CartesianGrid, XAxis, YAxis, Tooltip, ReferenceLine } from "recharts";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { useStream } from "@/lib/device-stream-context";
import { useServerHealth } from "@/hooks/use-server-health";
import { getDashboardDevices, getDeviceStatus, getVersion } from "@/lib/vyzorix-api";
import { StatusBadge, type DeviceHealth } from "@/components/status-badge";
import { formatRelative, formatUptime } from "@/lib/format";

export const Route = createFileRoute("/_app/dashboard")({
  head: () => ({ meta: [{ title: "Dashboard — Vyzorix" }] }),
  component: DashboardPage,
});

const tip = { background: "var(--popover)", border: "1px solid var(--border)", borderRadius: 8, fontSize: 12 };

function deriveHealth(
  online: boolean,
  riskScore: number | undefined,
  thermal: number | undefined,
  th: { riskWarn: number; riskCrit: number; thermalWarn: number; thermalCrit: number },
): DeviceHealth {
  if (!online) return "offline";
  if ((riskScore ?? 0) >= th.riskCrit || (thermal ?? 0) >= th.thermalCrit) return "critical";
  if ((riskScore ?? 0) >= th.riskWarn || (thermal ?? 0) >= th.thermalWarn) return "warning";
  return "online";
}

function DashboardPage() {
  const { serverUrl, deviceId, thresholds, dashboardToken } = useVyzorixConfig();
  const health = useServerHealth(serverUrl);
  const stream = useStream();
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

  const devices = useQuery({
    queryKey: ["vyzorix", "dashboard-devices", serverUrl, dashboardToken],
    queryFn: () => getDashboardDevices(serverUrl, dashboardToken),
    enabled: health.data?.ok === true,
    refetchInterval: 15_000,
    retry: false,
  });

  const online = status.data?.online ?? (stream.state === "connected");
  const deviceHealth = deriveHealth(online, t?.riskScore, t?.thermalTemp, thresholds);

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
                <code className="text-xs">go run .</code> or change the URL in Settings.
              </p>
            </div>
          </CardContent>
        </Card>
      ) : null}

      {/* Hero device card */}
      {status.isLoading || health.isLoading ? (
        <Card>
          <CardHeader>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div className="space-y-2">
                <div className="h-6 w-40 animate-pulse rounded-md bg-muted" />
                <div className="h-4 w-64 animate-pulse rounded-md bg-muted" />
              </div>
              <div className="h-6 w-20 animate-pulse rounded-full bg-muted" />
            </div>
          </CardHeader>
          <CardContent className="grid gap-3 md:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => <MetricSkeleton key={i} />)}
          </CardContent>
        </Card>
      ) : (
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
          <Metric icon={Activity} label="Risk score" value={t?.riskScore != null ? `${t.riskScore}` : "—"} hint={t?.riskScore != null ? (t.riskScore >= thresholds.riskCrit ? "Critical — soft reboot predicted" : t.riskScore >= thresholds.riskWarn ? "Investigate" : "Healthy") : "Awaiting signals"} />
          <Metric icon={ThermometerSun} label="Thermal" value={t?.thermalTemp != null ? `${t.thermalTemp.toFixed(1)}°C` : "—"} hint={t?.thermalTemp != null ? (t.thermalTemp >= thresholds.thermalCrit ? "THROTTLE_HEAVY" : t.thermalTemp >= thresholds.thermalWarn ? "THROTTLE_LIGHT" : "NONE") : "Awaiting signals"} />
          <Metric icon={Clock} label="Uptime" value={formatUptime(t?.uptime)} hint={`Last seen ${formatRelative(status.data?.lastSeen)}`} />
          <Metric icon={Volume2} label="Speaker" value={t?.speakerOn == null ? "—" : t.speakerOn ? "FORCED" : "OFF"} hint={t?.activeDevice ?? "—"} />
        </CardContent>
      </Card>
      )}

      {/* Live signals */}
      {status.isLoading || health.isLoading ? (
        <div className="grid gap-4 lg:grid-cols-2">
          {[0, 1].map(i => (
            <Card key={i}>
              <CardHeader className="pb-2">
                <div className="h-4 w-32 animate-pulse rounded-md bg-muted" />
              </CardHeader>
              <CardContent>
                <div className="h-48 w-full animate-pulse rounded-md bg-muted" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Risk score · live</CardTitle>
            <CardDescription>{stream.telemetryHistory.length} frames buffered</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartShell data={riskSeries} thresholds={[thresholds.riskWarn, thresholds.riskCrit]} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-base">Thermal · live (°C)</CardTitle>
            <CardDescription>From DeviceSignal.thermalTemp</CardDescription>
          </CardHeader>
          <CardContent>
            <ChartShell data={thermalSeries} thresholds={[thresholds.thermalWarn, thresholds.thermalCrit]} />
          </CardContent>
        </Card>
      </div>
      )}

      {/* Route + capture summary */}
      {status.isLoading || health.isLoading ? (
        <div className="grid gap-4 lg:grid-cols-3">
          {[0, 1, 2].map(i => (
            <Card key={i}>
              <CardHeader><div className="h-4 w-28 animate-pulse rounded-md bg-muted" /></CardHeader>
              <CardContent className="space-y-2">
                {Array.from({ length: 3 }).map((_, j) => <div key={j} className="h-3 w-full animate-pulse rounded bg-muted" />)}
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (
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
            <p className="text-xs text-muted-foreground">{t?.bufferLevel ?? 0}% fill · underrun threshold {thresholds.bufferWarn}%</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader><CardTitle className="text-base">Update server</CardTitle></CardHeader>
          <CardContent className="space-y-2 text-sm">
            <KV k="Latest version" v={version.data?.version ?? "—"} />
            <KV k="Version code" v={version.data?.version_code != null ? `${version.data.version_code}` : "—"} />
            <KV k="Health" v={health.data?.ok ? "ok" : "down"} />
            <KV k="Fleet devices" v={devices.data ? `${devices.data.length}` : "—"} />
          </CardContent>
        </Card>
      </div>
      )}
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
    return <div className="flex h-48 items-center justify-center text-xs text-muted-foreground">Waiting for live signals…</div>;
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
