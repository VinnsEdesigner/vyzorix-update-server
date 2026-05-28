import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { AlertTriangle, AlertCircle, Info, Search } from "lucide-react";
import { mockAlerts } from "@/lib/mock-data";

export const Route = createFileRoute("/_app/alerts")({
  head: () => ({ meta: [{ title: "System alerts — Vyzorix" }] }),
  component: AlertsPage,
});

type Severity = "critical" | "warning" | "info";

const severityIcon: Record<Severity, typeof AlertTriangle> = {
  critical: AlertCircle,
  warning: AlertTriangle,
  info: Info,
};

function AlertsPage() {
  const [query, setQuery] = useState("");
  const [severity, setSeverity] = useState<"all" | Severity>("all");

  const filtered = useMemo(() => {
    return mockAlerts.filter((a) => {
      if (severity !== "all" && a.severity !== severity) return false;
      if (!query.trim()) return true;
      const q = query.toLowerCase();
      return (
        a.message.toLowerCase().includes(q) ||
        a.device.toLowerCase().includes(q)
      );
    });
  }, [query, severity]);

  const counts = useMemo(
    () => ({
      critical: mockAlerts.filter((a) => a.severity === "critical").length,
      warning: mockAlerts.filter((a) => a.severity === "warning").length,
      info: mockAlerts.filter((a) => a.severity === "info").length,
    }),
    [],
  );

  return (
    <div className="space-y-6">
      <div className="grid gap-4 sm:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Critical</CardTitle>
            <AlertCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{counts.critical}</div>
            <p className="text-xs text-muted-foreground">Requires immediate action</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Warning</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{counts.warning}</div>
            <p className="text-xs text-muted-foreground">Investigate when possible</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Info</CardTitle>
            <Info className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{counts.info}</div>
            <p className="text-xs text-muted-foreground">Routine fleet events</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>System alerts</CardTitle>
          <CardDescription>All recent events across the fleet</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <div className="relative flex-1">
              <Search className="pointer-events-none absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search by device or message…"
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
            <Button variant="outline" onClick={() => { setQuery(""); setSeverity("all"); }}>
              Reset
            </Button>
          </div>

          <div>
            {filtered.length === 0 ? (
              <p className="py-10 text-center text-sm text-muted-foreground">
                No alerts match your filters.
              </p>
            ) : (
              filtered.map((a, i) => {
                const Icon = severityIcon[a.severity as Severity];
                return (
                  <div key={a.id}>
                    <div className="flex items-start gap-3 py-3">
                      <Icon className="mt-0.5 h-4 w-4 text-muted-foreground" />
                      <div className="min-w-0 flex-1">
                        <p className="text-sm leading-snug">{a.message}</p>
                        <p className="text-xs text-muted-foreground">
                          {a.device} · {a.at}
                        </p>
                      </div>
                      <Badge
                        variant={
                          a.severity === "critical"
                            ? "destructive"
                            : a.severity === "warning"
                              ? "secondary"
                              : "outline"
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