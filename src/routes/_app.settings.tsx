import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";

import { DEFAULT_DEVICE_ID, DEFAULT_SERVER_URL, useVyzorixConfig } from "@/lib/vyzorix-config";
import { useServerHealth } from "@/hooks/use-server-health";

export const Route = createFileRoute("/_app/settings")({
  head: () => ({ meta: [{ title: "Settings — Vyzorix" }] }),
  component: SettingsPage,
});

function SettingsPage() {
  const { serverUrl, deviceId, setServerUrl, setDeviceId } = useVyzorixConfig();
  const [draftUrl, setDraftUrl] = useState(serverUrl);
  const [draftId, setDraftId] = useState(deviceId);
  const health = useServerHealth(serverUrl);

  const save = () => {
    setServerUrl(draftUrl.trim() || DEFAULT_SERVER_URL);
    setDeviceId(draftId.trim() || DEFAULT_DEVICE_ID);
    toast.success("Settings saved");
  };

  const reset = () => {
    setDraftUrl(DEFAULT_SERVER_URL);
    setDraftId(DEFAULT_DEVICE_ID);
    setServerUrl(DEFAULT_SERVER_URL);
    setDeviceId(DEFAULT_DEVICE_ID);
    toast.info("Reset to defaults");
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
            <Input id="endpoint" value={draftUrl} onChange={(e) => setDraftUrl(e.target.value)} placeholder={DEFAULT_SERVER_URL} />
            <p className="text-xs text-muted-foreground">REST and WSS endpoints are derived from this URL.</p>
          </div>
          <Separator />
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Health check</span>
            <Badge variant={health.data?.ok ? "default" : health.isError ? "destructive" : "secondary"}>
              {health.isLoading ? "checking…" : health.data?.ok ? "ok" : "down"}
            </Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Target device</CardTitle>
          <CardDescription>Single Nokia C22 — Phase 1 deliberately scoped to one device</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="device">deviceId</Label>
            <Input id="device" value={draftId} onChange={(e) => setDraftId(e.target.value)} placeholder={DEFAULT_DEVICE_ID} />
            <p className="text-xs text-muted-foreground">Must match the value the Android daemon registers with.</p>
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle>Mock server quick start</CardTitle>
          <CardDescription>From the repo root</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs">go run ./cmd/mockserver -addr=:8080 -data=./cmd/mockserver/testdata</pre>
          <p className="text-muted-foreground">
            The mock server defaults to <code className="text-xs">-strict-hmac=false</code>, so the dashboard can dispatch commands
            without the Android-side command_secret. Strict HMAC support will come with the real server in Phase 1.5.
          </p>
          <div className="flex flex-wrap items-center justify-between gap-3 pt-2">
            <p className="text-xs text-muted-foreground">
              Stored locally in your browser. No credentials are sent anywhere except the configured server.
            </p>
            <div className="flex gap-2">
              <Button variant="outline" onClick={reset}>Reset</Button>
              <Button onClick={save}>Save changes</Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
