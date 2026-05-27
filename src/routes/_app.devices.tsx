import { createFileRoute } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/status-badge";
import { formatUptime, mockDevices, type DeviceStatus } from "@/lib/mock-data";
import { Search, RefreshCw, MoreHorizontal } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/devices")({
  head: () => ({ meta: [{ title: "Devices — Vyzorix" }] }),
  component: DevicesPage,
});

function DevicesPage() {
  const [q, setQ] = useState("");
  const [status, setStatus] = useState<"all" | DeviceStatus>("all");

  const filtered = useMemo(() => {
    return mockDevices.filter((d) => {
      const matchQ = q.length === 0 || `${d.name} ${d.id} ${d.model}`.toLowerCase().includes(q.toLowerCase());
      const matchS = status === "all" || d.status === status;
      return matchQ && matchS;
    });
  }, [q, status]);

  return (
    <Card>
      <CardHeader className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <CardTitle>Device fleet</CardTitle>
            <p className="text-sm text-muted-foreground">{filtered.length} of {mockDevices.length} devices</p>
          </div>
          <Button variant="outline" size="sm" onClick={() => toast.success("Fleet refreshed")}>
            <RefreshCw className="h-4 w-4" /> Refresh
          </Button>
        </div>
        <div className="flex flex-wrap gap-2">
          <div className="relative min-w-[200px] flex-1">
            <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Search id, name, model…" className="pl-8" />
          </div>
          <Select value={status} onValueChange={(v) => setStatus(v as typeof status)}>
            <SelectTrigger className="w-[160px]"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              <SelectItem value="online">Online</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="critical">Critical</SelectItem>
              <SelectItem value="offline">Offline</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Device</TableHead>
              <TableHead>Status</TableHead>
              <TableHead className="hidden md:table-cell">Risk</TableHead>
              <TableHead className="hidden md:table-cell">Uptime</TableHead>
              <TableHead className="hidden lg:table-cell">Thermal</TableHead>
              <TableHead className="hidden lg:table-cell">Buffer</TableHead>
              <TableHead className="hidden lg:table-cell">App</TableHead>
              <TableHead className="hidden xl:table-cell">Last seen</TableHead>
              <TableHead className="w-[40px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((d) => (
              <TableRow key={d.id}>
                <TableCell>
                  <div className="font-medium">{d.name}</div>
                  <div className="text-xs text-muted-foreground">{d.id} · {d.model} · Android {d.androidVersion}</div>
                </TableCell>
                <TableCell><StatusBadge status={d.status} /></TableCell>
                <TableCell className="hidden md:table-cell font-mono">{d.riskScore}</TableCell>
                <TableCell className="hidden md:table-cell">{formatUptime(d.uptimeSec)}</TableCell>
                <TableCell className="hidden lg:table-cell">{d.thermalTemp ? `${d.thermalTemp}°C` : "—"}</TableCell>
                <TableCell className="hidden lg:table-cell">{d.status === "offline" ? "—" : `${d.bufferLevel}%`}</TableCell>
                <TableCell className="hidden lg:table-cell">{d.appVersion}</TableCell>
                <TableCell className="hidden xl:table-cell text-muted-foreground">{d.lastSeen}</TableCell>
                <TableCell>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuLabel>{d.id}</DropdownMenuLabel>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => toast.info("Sent FORCE_SPEAKER")}>Force speaker</DropdownMenuItem>
                      <DropdownMenuItem onClick={() => toast.info("Sent RESET_AUDIO_HAL")}>Reset audio HAL</DropdownMenuItem>
                      <DropdownMenuItem onClick={() => toast.info("Sent TOGGLE_CAPTURE")}>Toggle capture</DropdownMenuItem>
                      <DropdownMenuItem onClick={() => toast.info("Sent REINIT_PROJECTION")}>Reinit projection</DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem onClick={() => toast.info("Wake push queued")}>Wake daemon (FCM)</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={9} className="text-center text-sm text-muted-foreground py-8">No devices match the current filters.</TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  );
}