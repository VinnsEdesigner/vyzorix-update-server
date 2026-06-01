import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

import { DEFAULT_DEVICE_ID, DEFAULT_SERVER_URL, useVyzorixConfig } from "@/lib/vyzorix-config";
import { useServerHealth } from "@/hooks/use-server-health";

export const Route = createFileRoute("/_app/settings/connection")({
  component: ConnectionSettings,
});

function ConnectionSettings() {
  const cfg = useVyzorixConfig();
  const [serverUrl, setServerUrl] = useState(cfg.serverUrl);
  const [deviceId, setDeviceId] = useState(cfg.deviceId);
  const [timeout, setTimeout] = useState<number>(cfg.requestTimeoutMs);
  const [autoReconnect, setAutoReconnect] = useState(cfg.autoReconnect);
  const [strictHmac, setStrictHmac] = useState(cfg.strictHmac);

  const health = useServerHealth(cfg.serverUrl);

  const save = () => {
    cfg.update({
      serverUrl: serverUrl.trim() || DEFAULT_SERVER_URL,
      deviceId: deviceId.trim() || DEFAULT_DEVICE_ID,
      requestTimeoutMs: Math.max(500, Math.min(60_000, timeout || 8000)),
      autoReconnect,
      strictHmac,
    });
    toast.success("Connection settings saved");
  };

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Update server</CardTitle>
          <CardDescription>Base URL of the Go binary in <code className="text-xs">cmd/mockserver</code></CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="endpoint">Base URL</Label>
            <Input id="endpoint" value={serverUrl} onChange={(e) => setServerUrl(e.target.value)} placeholder={DEFAULT_SERVER_URL} />
            <p className="text-xs text-muted-foreground">REST + WSS endpoints are derived from this URL.</p>
          </div>
          <Separator />
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Health check</span>
            <Badge variant={health.data?.ok ? "default" : health.isError ? "destructive" : "secondary"}>
              {health.data?.ok ? "ok" : health.isError ? "down" : "checking"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Target device</CardTitle>
          <CardDescription>Single Nokia C22 — Phase 1 is deliberately scoped to one device</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="device">deviceId</Label>
            <Input id="device" value={deviceId} onChange={(e) => setDeviceId(e.target.value)} placeholder={DEFAULT_DEVICE_ID} />
            <p className="text-xs text-muted-foreground">Must match the value the Android daemon registers with.</p>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle>Transport</CardTitle>
          <CardDescription>How the dashboard talks to the server</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="timeout">Request timeout (ms)</Label>
              <Input id="timeout" type="number" min={500} max={60000} step={500} value={timeout}
                onChange={(e) => setTimeout(Number(e.target.value))} />
            </div>
            <ToggleRow
              label="Auto-reconnect WebSocket"
              hint="Re-open the stream with exponential backoff after drops"
              checked={autoReconnect}
              onChange={setAutoReconnect}
            />
            <ToggleRow
              label="Strict HMAC on commands"
              hint="Require X-Vyzorix-Signature on every command. Match the mock server's -strict-hmac flag."
              checked={strictHmac}
              onChange={setStrictHmac}
            />
          </div>
          <div className="flex justify-end">
            <Button onClick={save}>Save connection</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function ToggleRow({ label, hint, checked, onChange }: { label: string; hint: string; checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <div className="flex items-start justify-between gap-3 rounded-md border p-3">
      <div className="space-y-0.5">
        <p className="text-sm font-medium">{label}</p>
        <p className="text-xs text-muted-foreground">{hint}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onChange} />
    </div>
  );
}