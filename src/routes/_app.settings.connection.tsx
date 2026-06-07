import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, useRef, type ReactElement } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { useServerHealth } from "@/hooks/use-server-health";
import { updateSettings, me, type ClientSettings } from "@/lib/vyzorix-auth";
import { DEFAULT_SERVER_URL, useVyzorixConfig } from "@/lib/vyzorix-config";

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

const getHealthStatus = (data: { ok?: boolean } | undefined, isError: boolean) => {
  if (data?.ok) return { variant: "default" as const, label: "ok" };
  if (isError) return { variant: "destructive" as const, label: "down" };
  return { variant: "secondary" as const, label: "checking" };
};

function ConnectionSettings(): ReactElement {
  const cfg = useVyzorixConfig();
  const cfgRef = useRef(cfg);
  cfgRef.current = cfg;
  const [serverUrl, setServerUrl] = useState(cfg.serverUrl);
  const [serverUrlError, setServerUrlError] = useState<string | null>(null);
  const [deviceId, setDeviceId] = useState(cfg.deviceId);
  const [timeout, setTimeout] = useState<number>(cfg.requestTimeoutMs);
  const [autoReconnect, setAutoReconnect] = useState(cfg.autoReconnect);
  const [strictHmac, setStrictHmac] = useState(cfg.strictHmac);
  const [dashboardToken, setDashboardToken] = useState(cfg.dashboardToken);
  const [saving, setSaving] = useState(false);

  const health = useServerHealth(cfg.serverUrl);

  // Load client settings from server on mount
  useEffect(() => {
    const loadFromServer = async () => {
      try {
        const op = await me(cfgRef.current.serverUrl);
        if (op.client) {
          setAutoReconnect(op.client.autoReconnect ?? true);
          setStrictHmac(op.client.strictHmac ?? false);
          cfgRef.current.update({
            autoReconnect: op.client.autoReconnect ?? true,
            strictHmac: op.client.strictHmac ?? false,
          });
        }
      } catch {
        // Use local defaults
      }
    };
    loadFromServer();
  }, []);

  const handleServerUrlChange = (value: string) => {
    setServerUrl(value);
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

  const save = async () => {
    if (!validateForm()) {
      toast.error("Please fix the server URL before saving");
      return;
    }

    setSaving(true);
    try {
      const client: ClientSettings = {
        autoReconnect,
        strictHmac,
      };
      await updateSettings(cfg.serverUrl, { client });
      cfg.update({
        serverUrl: serverUrl.trim(),
        deviceId: deviceId.trim(),
        requestTimeoutMs: Math.max(500, Math.min(60_000, timeout || 8000)),
        autoReconnect,
        strictHmac,
        dashboardToken: dashboardToken.trim(),
      });
      toast.success("Connection settings saved to server");
    } catch (e) {
      toast.error("Failed to save settings", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setSaving(false);
    }
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
            <Badge variant={getHealthStatus(health.data, health.isError).variant}>
              {getHealthStatus(health.data, health.isError).label}
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
          <CardDescription>
            How the dashboard talks to the server — saved to server, persists across devices
          </CardDescription>
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
              hint="When enabled, server requires HMAC signature on every command from this operator"
              checked={strictHmac}
              onChange={setStrictHmac}
            />
          </div>
          <div className="flex justify-end">
            <Button onClick={save} disabled={saving}>
              {saving ? "Saving..." : "Save connection"}
            </Button>
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
}): ReactElement {
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
