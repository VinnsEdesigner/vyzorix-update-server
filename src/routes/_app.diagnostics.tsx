import { createFileRoute, Link } from "@tanstack/react-router";
import { Terminal } from "lucide-react";
import { useState } from "react";
import {
  ResponsiveContainer,
  LineChart,
  Line,
  CartesianGrid,
  XAxis,
  YAxis,
  Tooltip,
  ReferenceLine,
} from "recharts";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useStream } from "@/lib/device-stream-context";
import { formatRelative } from "@/lib/format";
import { COMMANDS, dispatchCommand } from "@/lib/vyzorix-api";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/diagnostics")({
  head: () => ({ meta: [{ title: "Diagnostics — Vyzorix" }] }),
  component: DiagnosticsPage,
});

const tip = {
  background: "var(--popover)",
  border: "1px solid var(--border)",
  borderRadius: 8,
  fontSize: 12,
};

// eslint-disable-next-line func-style
function DiagnosticsPage(): JSX.Element {
  const { serverUrl, deviceId, thresholds, dashboardToken, strictHmac } = useVyzorixConfig();
  const stream = useStream();
  const [pending, setPending] = useState<string | null>(null);

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const send = async (cmd: string) => {
    if (!deviceId.trim()) {
      toast.error(`${cmd} failed`, {
        description: "No device registered — set deviceId in Settings → Connection",
      });
      return;
    }
    setPending(cmd);
    try {
      const res = await dispatchCommand(
        serverUrl,
        deviceId,
        cmd,
        undefined,
        dashboardToken,
        strictHmac,
      );
      toast.success(`${cmd} → ${res.delivery}`, { description: `dispatch ${res.dispatchId}` });
    } catch (e) {
      toast.error(`${cmd} failed`, { description: e instanceof Error ? e.message : String(e) });
    } finally {
      setPending(null);
    }
  };

  const risk = stream.telemetryHistory.map((f, i) => ({ i, v: f.riskScore ?? 0 }));
  const thermal = stream.telemetryHistory.map((f, i) => ({ i, v: f.thermalTemp ?? 0 }));
  const buffer = stream.telemetryHistory.map((f, i) => ({ i, v: f.bufferLevel ?? 0 }));
  const audioMode = stream.telemetryHistory.map((f, i) => ({ i, v: f.audioMode ?? 0 }));

  return (
    <div className="space-y-4">
      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Command panel</CardTitle>
            <CardDescription>POST /v1/device/{deviceId}/command</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-2">
            {COMMANDS.map((c) => (
              <Button
                key={c.id}
                variant={c.danger ? "destructive" : "outline"}
                size="sm"
                disabled={pending === c.id || stream.state === "disconnected"}
                onClick={() => send(c.id)}
                title={c.description}
              >
                {pending === c.id ? "Sending…" : c.label}
              </Button>
            ))}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Stream state</CardTitle>
            <CardDescription>WSS /v1/device/{deviceId}/stream</CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            <Stat label="Connection" value={stream.state} />
            <Stat label="Frames buffered" value={`${stream.telemetryHistory.length}`} />
            <Stat
              label="Last signal"
              value={formatRelative(stream.lastTelemetry?.timestamp ?? undefined)}
            />
            <Stat label="Last error" value={stream.error ?? "—"} />
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <ChartCard
          title="Risk score"
          data={risk}
          thresholds={[thresholds.riskWarn, thresholds.riskCrit]}
        />
        <ChartCard
          title="Thermal (°C)"
          data={thermal}
          thresholds={[thresholds.thermalWarn, thresholds.thermalCrit]}
        />
        <ChartCard title="Buffer level (%)" data={buffer} thresholds={[thresholds.bufferWarn]} />
        <ChartCard title="Audio mode" data={audioMode} />
      </div>

      <Card>
        <CardContent className="flex items-center justify-between gap-3 py-4 text-sm">
          <div className="flex items-center gap-2">
            <Terminal className="h-4 w-4 text-muted-foreground" />
            <span>
              All service logs now stream into the docked log console at the bottom of every page.
            </span>
          </div>
          <Link to="/logs">
            <Button variant="outline" size="sm">
              Open full page logs
            </Button>
          </Link>
        </CardContent>
      </Card>
    </div>
  );
}

// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="text-sm font-medium break-all">{value}</p>
    </div>
  );
}

// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
// eslint-disable-next-line func-style
function ChartCard({
  title,
  data,
  thresholds,
}: {
  title: string;
  data: { i: number; v: number }[];
  thresholds?: number[];
}) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        {data.length === 0 ? (
          <div className="flex h-48 items-center justify-center text-xs text-muted-foreground">
            Waiting for signals…
          </div>
        ) : (
          <div className="h-48 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={data} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
                <XAxis dataKey="i" hide />
                <YAxis stroke="var(--muted-foreground)" fontSize={10} />
                <Tooltip contentStyle={tip} />
                {thresholds?.map((y) => (
                  <ReferenceLine
                    key={y}
                    y={y}
                    stroke="var(--muted-foreground)"
                    strokeDasharray="3 3"
                  />
                ))}
                <Line
                  type="monotone"
                  dataKey="v"
                  stroke="var(--primary)"
                  dot={false}
                  strokeWidth={2}
                  isAnimationActive={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
