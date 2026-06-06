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

// Validate URL has proper protocol
function isValidServerUrl(url: string): boolean {
  if (!url.trim()) return false;
  try {
    const u = new URL(url);
    return u.protocol === "http:" || u.protocol === "https:";
  } catch {
    return false;
  }
}

function ConnectionSettings() {
  const cfg = useVyzorixConfig();
  const [serverUrl, setServerUrl] = useState(cfg.serverUrl);
  const [serverUrlError, setServerUrlError] = useState<string | null>(null);
  const [deviceId, setDeviceId] = useState(cfg.deviceId);
  const [timeout, setTimeout] = useState<number>(cfg.requestTimeoutMs);
  const [autoReconnect, setAutoReconnect] = useState(cfg.autoReconnect);
  const [strictHmac, setStrictHmac] = useState(cfg.strictHmac);
  const [dashboardToken, setDashboardToken] = useState(cfg.dashboardToken);

  const health = useServerHealth(cfg.serverUrl);

  const handleServerUrlChange = (value: string) => {
    setServerUrl(value);
    // Clear error when user starts typing
    if (serverUrlError) setServerUrlError(null);
  };

  const validateForm = (): boolean => {
    const trimmed = serverUrl.trim();

    if (!trimmed) {
      setServerUrlError("Server URL is required");
      return false;
    }

    if (!isValidServerUrl(trimmed)) {
      setServerUrlError("Invalid URL. Include http:// or https:// (e.g., http://localhost:3000)");
      return false;
    }

    return true;
  };

  const save = () => {
    if (!validateForm()) {
      toast.error("Please fix the server URL before saving");
      return;
    }

    cfg.update({
      serverUrl: serverUrl.trim(),
      deviceId: deviceId.trim(),
      requestTimeoutMs: Math.max(500, Math.min(60_000, timeout || 8000)),
      autoReconnect,
      strictHmac,
      dashboardToken: dashboardToken.trim(),
    });
    toast.success("Connection settings saved");
  };

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Update server</CardTitle>
          <CardDescription>Base URL of the Render-backed Go update server</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="endpoint">Base URL</Label>
            <Input
              id="endpoint"
              value={serverUrl}
              onChange={(e) => handleServerUrlChange(e.target.value)}
              placeholder={DEFAULT_SERVER_URL}
              className={serverUrlError ? "border-destructive" : ""}
            />
            {serverUrlError ? (
              <p className="text-xs text-destructive">{serverUrlError}</p>
            ) : (
              <p className="text-xs text-muted-foreground">
                REST + WSS endpoints are derived from this URL.
              </p>
            )}
          </div>
          <Separator />
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Health check</span>
            <Badge
              variant={health.data?.ok ? "default" : health.isError ? "destructive" : "secondary"}
            >
              {health.data?.ok ? "ok" : health.isError ? "down" : "checking"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Target device</CardTitle>
          <CardDescription>
            Device ID for the target Android device running VyzorixAudioRouter
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="device">deviceId</Label>
            <Input
              id="device"
              value={deviceId}
              onChange={(e) => setDeviceId(e.target.value)}
              placeholder="Enter device ID"
            />
            <p className="text-xs text-muted-foreground">
              Must match the value the Android daemon registers with.
            </p>
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
              <Input
                id="timeout"
                type="number"
                min={500}
                max={60000}
                step={500}
                value={timeout}
                onChange={(e) => setTimeout(Number(e.target.value))}
              />
            </div>
            <div className="space-y-1.5 sm:col-span-2">
              <Label htmlFor="dashboard-token">Dashboard token</Label>
              <Input
                id="dashboard-token"
                type="password"
                value={dashboardToken}
                onChange={(e) => setDashboardToken(e.target.value)}
                placeholder="TOKEN_SECRET from Render"
              />
              <p className="text-xs text-muted-foreground">
                Used as Authorization/X-Vyzorix-Token for production dashboard endpoints and
                commands.
              </p>
            </div>
            <ToggleRow
              label="Auto-reconnect WebSocket"
              hint="Re-open the stream with exponential backoff after drops"
              checked={autoReconnect}
              onChange={setAutoReconnect}
            />
            <ToggleRow
              label="Strict HMAC on commands"
              hint="Require X-Vyzorix-Signature on every command. Match the real server ENFORCE_HMAC setting."
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

function ToggleRow({
  label,
  hint,
  checked,
  onChange,
}: {
  label: string;
  hint: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
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
