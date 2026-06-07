import { createFileRoute } from "@tanstack/react-router";
import { AlertTriangle, AlertCircle, Info, Search } from "lucide-react";
import { useMemo, useState, type ReactElement } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Separator } from "@/components/ui/separator";
import { useStream } from "@/lib/device-stream-context";
import type { TelemetryFrame } from "@/lib/vyzorix-api";
import { useVyzorixConfig } from "@/lib/vyzorix-config";
import type { Thresholds } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/alerts")({
  head: () => ({ meta: [{ title: "System alerts — Vyzorix" }] }),
  component: AlertsPage,
});

type Severity = "critical" | "warning" | "info";
interface DerivedAlert {
  id: string;
  severity: Severity;
  message: string;
  at: number;
}

const severityIcon: Record<Severity, typeof AlertTriangle> = {
  critical: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

function deriveAlerts(history: TelemetryFrame[], th: Thresholds): DerivedAlert[] {
  const out: DerivedAlert[] = [];
  history.forEach((f, i) => {
    const at =
      typeof f.timestamp === "number" ? f.timestamp : Date.now() - (history.length - i) * 1000;
    if ((f.riskScore ?? 0) >= th.riskCrit) {
      out.push({
        id: `risk-${i}`,
        severity: "critical",
        message: `Risk score ${f.riskScore} — soft reboot predicted`,
        at,
      });
    } else if ((f.riskScore ?? 0) >= th.riskWarn) {
      out.push({
        id: `risk-${i}`,
        severity: "warning",
        message: `Risk score ${f.riskScore} — elevated`,
        at,
      });
    }
    if ((f.thermalTemp ?? 0) >= th.thermalCrit) {
      out.push({
        id: `thermal-${i}`,
        severity: "critical",
        message: `Thermal ${f.thermalTemp?.toFixed(1)}°C — THROTTLE_HEAVY`,
        at,
      });
    } else if ((f.thermalTemp ?? 0) >= th.thermalWarn) {
      out.push({
        id: `thermal-${i}`,
        severity: "warning",
        message: `Thermal ${f.thermalTemp?.toFixed(1)}°C — THROTTLE_LIGHT`,
        at,
      });
    }
    if (f.bufferLevel != null && f.bufferLevel < th.bufferWarn) {
      out.push({
        id: `buf-${i}`,
        severity: "warning",
        message: `Buffer fill ${f.bufferLevel}% — approaching underrun`,
        at,
      });
    }
    if (f.speakerOn === false) {
      out.push({
        id: `spk-${i}`,
        severity: "info",
        message: `Speaker route lost — active=${f.activeDevice ?? "unknown"}`,
        at,
      });
    }
  });
  return out.slice(-100).reverse();
}

function AlertsPage(): ReactElement {
  const { thresholds } = useVyzorixConfig();
  const stream = useStream();
  const [query, setQuery] = useState("");
  const [severity, setSeverity] = useState<"all" | Severity>("all");

  const alerts = useMemo(
    () => deriveAlerts(stream.telemetryHistory, thresholds),
    [stream.telemetryHistory, thresholds],
  );
  const filtered = alerts.filter((a) => {
    if (severity !== "all" && a.severity !== severity) return false;
    if (!query.trim()) return true;
    return a.message.toLowerCase().includes(query.toLowerCase());
  });

  const counts = {
    critical: alerts.filter((a) => a.severity === "critical").length,
    warning: alerts.filter((a) => a.severity === "warning").length,
    info: alerts.filter((a) => a.severity === "info").length,
  };

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-3">
        <SummaryCard
          label="Critical"
          count={counts.critical}
          hint="Requires immediate action"
          Icon={AlertCircle}
        />
        <SummaryCard
          label="Warning"
          count={counts.warning}
          hint="Investigate when possible"
          Icon={AlertTriangle}
        />
        <SummaryCard label="Info" count={counts.info} hint="Routine route events" Icon={Info} />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>System alerts</CardTitle>
          <CardDescription>
            Derived from live DeviceSignal thresholds · {alerts.length} total
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search message…"
                className="pl-8"
              />
            </div>
            <Select value={severity} onValueChange={(v) => setSeverity(v as typeof severity)}>
              <SelectTrigger className="sm:w-44">
                <SelectValue placeholder="Severity" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">All severities</SelectItem>
                <SelectItem value="critical">Critical</SelectItem>
                <SelectItem value="warning">Warning</SelectItem>
                <SelectItem value="info">Info</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={() => {
                setQuery("");
                setSeverity("all");
              }}
            >
              Reset
            </Button>
          </div>

          <div>
            {filtered.length === 0 ? (
              <p className="py-10 text-center text-sm text-muted-foreground">
                {alerts.length === 0
                  ? "No telemetry yet — waiting for the device."
                  : "No alerts match your filters."}
              </p>
            ) : (
              filtered.map((a, i) => {
                const Icon = severityIcon[a.severity];
                return (
                  <div key={a.id}>
                    <div className="flex items-start gap-3 py-3">
                      <Icon className="mt-0.5 h-4 w-4 text-muted-foreground" />
                      <div className="min-w-0 flex-1">
                        <p className="text-sm leading-snug">{a.message}</p>
                        <p className="text-xs text-muted-foreground">
                          {new Date(a.at).toLocaleTimeString()}
                        </p>
                      </div>
                      <Badge
                        variant={
                          (
                            Object.fromEntries([
                              ["critical", "destructive"],
                              ["warning", "secondary"],
                            ]) as Record<string, "destructive" | "secondary">
                          )[a.severity] ?? ("outline" as const)
                        }
                        className="uppercase text-[10px]"
                      >
                        {a.severity}
                      </Badge>
                    </div>
                    {i < filtered.length - 1 && <Separator />}
                  </div>
                );
              })
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function SummaryCard({
  label,
  count,
  hint,
  Icon,
}: {
  label: string;
  count: number;
  hint: string;
  Icon: typeof AlertTriangle;
}): ReactElement {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{label}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-semibold">{count}</div>
        <p className="text-xs text-muted-foreground">{hint}</p>
      </CardContent>
    </Card>
  );
}
