import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Progress } from "@/components/ui/progress";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { mockDevices, mockLogLines, makeSeries } from "@/lib/mock-data";
import { ResponsiveContainer, LineChart, Line, CartesianGrid, XAxis, YAxis, Tooltip, ReferenceLine } from "recharts";
import { toast } from "sonner";
import { Volume2, Cpu, Power, Camera, RefreshCcw } from "lucide-react";

export const Route = createFileRoute("/_app/diagnostics")({
  head: () => ({ meta: [{ title: "Diagnostics — Vyzorix" }] }),
  component: DiagnosticsPage,
});

const cpuSeries = makeSeries(60, 30, 30);
const riskSeries = makeSeries(60, 35, 50);
const thermalSeries = makeSeries(60, 42, 12).map((p) => ({ ...p, value: +(p.value).toFixed(1) }));
const bufferSeries = makeSeries(60, 80, 30);

const tip = { background: "var(--popover)", border: "1px solid var(--border)", borderRadius: 8, fontSize: 12 };

function DiagnosticsPage() {
  const [selected, setSelected] = useState(mockDevices[0].id);
  const device = mockDevices.find((d) => d.id === selected)!;

  const cmd = (name: string) => {
    toast.success(`Dispatched ${name} → ${device.id}`);
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle>Live diagnostics</CardTitle>
            <CardDescription>Real-time telemetry and remote control for the selected device.</CardDescription>
          </div>
          <Select value={selected} onValueChange={setSelected}>
            <SelectTrigger className="w-[260px]"><SelectValue /></SelectTrigger>
            <SelectContent>
              {mockDevices.map((d) => (
                <SelectItem key={d.id} value={d.id}>{d.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardHeader>
        <CardContent className="grid gap-4 md:grid-cols-3">
          <Stat label="Uptime" value={device.status === "offline" ? "—" : "2d 03h"} />
          <Stat label="Risk score" value={`${device.riskScore}`} hint={device.riskScore > 75 ? "Critical" : device.riskScore > 50 ? "Warn" : "OK"} />
          <Stat label="Thermal" value={device.thermalTemp ? `${device.thermalTemp}°C` : "—"} />
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Control panel</CardTitle>
            <CardDescription>Remote C2 commands</CardDescription>
          </CardHeader>
          <CardContent className="grid grid-cols-2 gap-2">
            <Button variant="outline" onClick={() => cmd("FORCE_SPEAKER")}><Volume2 className="h-4 w-4" />Force speaker</Button>
            <Button variant="outline" onClick={() => cmd("RESET_AUDIO_HAL")}><RefreshCcw className="h-4 w-4" />Reset HAL</Button>
            <Button variant="outline" onClick={() => cmd("TOGGLE_CAPTURE")}><Camera className="h-4 w-4" />Toggle capture</Button>
            <Button variant="outline" onClick={() => cmd("REINIT_PROJECTION")}><Cpu className="h-4 w-4" />Reinit projection</Button>
            <Button variant="destructive" className="col-span-2" onClick={() => cmd("DUMP_FLIGHT_DATA")}>
              <Power className="h-4 w-4" />Dump flight data
            </Button>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader><CardTitle className="text-base">Route state</CardTitle></CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-2">
            <div className="rounded-md border p-3">
              <p className="text-xs text-muted-foreground">Current route</p>
              <p className="text-sm font-medium">Speaker forced</p>
              <p className="mt-1 text-xs text-muted-foreground">Corrected 14s ago</p>
            </div>
            <div className="rounded-md border p-3">
              <p className="text-xs text-muted-foreground">Active device</p>
              <p className="text-sm font-medium">builtin_speaker</p>
              <p className="mt-1 text-xs text-muted-foreground">audio mode: 3 (COMMUNICATION)</p>
            </div>
            <div className="rounded-md border p-3">
              <p className="text-xs text-muted-foreground">Buffer health</p>
              <Progress value={device.bufferLevel || 0} className="mt-2" />
              <p className="mt-1 text-xs text-muted-foreground">{device.bufferLevel || 0}% fill</p>
            </div>
            <div className="rounded-md border p-3">
              <p className="text-xs text-muted-foreground">Thermal mitigation</p>
              <p className="text-sm font-medium">{device.thermalTemp > 50 ? "THROTTLE_HEAVY" : device.thermalTemp > 45 ? "THROTTLE_LIGHT" : "NONE"}</p>
              <p className="mt-1 text-xs text-muted-foreground">sample rate: {device.thermalTemp > 50 ? "24kHz" : "48kHz"}</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <Tabs defaultValue="charts">
        <TabsList>
          <TabsTrigger value="charts">Charts</TabsTrigger>
          <TabsTrigger value="logs">Log terminal</TabsTrigger>
        </TabsList>
        <TabsContent value="charts" className="grid gap-4 lg:grid-cols-2">
          <ChartCard title="CPU load" data={cpuSeries} stroke="var(--primary)" unit="%" />
          <ChartCard title="Risk score" data={riskSeries} stroke="var(--primary)" thresholds={[50, 75]} />
          <ChartCard title="Thermal" data={thermalSeries} stroke="var(--primary)" unit="°C" thresholds={[45, 55]} />
          <ChartCard title="Buffer health" data={bufferSeries} stroke="var(--primary)" unit="%" thresholds={[50]} />
        </TabsContent>
        <TabsContent value="logs">
          <Card>
            <CardHeader><CardTitle className="text-base">Live log stream</CardTitle></CardHeader>
            <CardContent>
              <ScrollArea className="h-80 rounded-md border bg-muted/40 p-3 font-mono text-xs leading-relaxed">
                {mockLogLines.map((l, i) => (
                  <div key={i} className="whitespace-pre-wrap">{l}</div>
                ))}
              </ScrollArea>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

function Stat({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{label}</p>
      <p className="text-xl font-semibold">{value}</p>
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function ChartCard({
  title,
  data,
  stroke,
  unit,
  thresholds,
}: {
  title: string;
  data: { t: number; value: number }[];
  stroke: string;
  unit?: string;
  thresholds?: number[];
}) {
  return (
    <Card>
      <CardHeader className="pb-2"><CardTitle className="text-base">{title}{unit ? ` (${unit})` : ""}</CardTitle></CardHeader>
      <CardContent>
        <div className="h-48 w-full">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={data} margin={{ top: 5, right: 10, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis dataKey="t" hide />
              <YAxis stroke="var(--muted-foreground)" fontSize={10} />
              <Tooltip contentStyle={tip} />
              {thresholds?.map((v) => (
                <ReferenceLine key={v} y={v} stroke="var(--muted-foreground)" strokeDasharray="3 3" />
              ))}
              <Line type="monotone" dataKey="value" stroke={stroke} dot={false} strokeWidth={2} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </CardContent>
    </Card>
  );
}