import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { getDeviceStatus, registerDevice } from "@/lib/vyzorix-api";
import { useDeviceStream } from "@/hooks/use-device-stream";
import { StatusBadge, type DeviceHealth } from "@/components/status-badge";
import { formatRelative, formatUptime, shortHash } from "@/lib/format";

export const Route = createFileRoute("/_app/device")({
  head: () => ({ meta: [{ title: "Device — Vyzorix" }] }),
  component: DevicePage,
});

function DevicePage() {
  const { serverUrl, deviceId } = useVyzorixConfig();
  const stream = useDeviceStream(serverUrl, deviceId);
  const t = stream.lastTelemetry;

  const status = useQuery({
    queryKey: ["vyzorix", "status", serverUrl, deviceId],
    queryFn: () => getDeviceStatus(serverUrl, deviceId),
    enabled: !!serverUrl && !!deviceId,
    refetchInterval: 10_000,
    retry: false,
  });

  const health: DeviceHealth =
    !status.data?.online && stream.state !== "connected"
      ? "offline"
      : (t?.riskScore ?? 0) >= 75 || (t?.thermalTemp ?? 0) >= 55
      ? "critical"
      : (t?.riskScore ?? 0) >= 50 || (t?.thermalTemp ?? 0) >= 45
      ? "warning"
      : "online";

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle>Nokia C22 — primary</CardTitle>
            <CardDescription>{deviceId}</CardDescription>
          </div>
          <StatusBadge status={health} />
        </CardHeader>
        <CardContent>
          {status.isError ? (
            <p className="text-sm text-muted-foreground">
              Device not registered yet. Use the registration panel below or run the Android daemon to call{" "}
              <code className="text-xs">POST /v1/device/register</code>.
            </p>
          ) : (
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <KV k="App version" v={status.data?.appVersion ?? "—"} />
              <KV k="Device class" v={status.data?.deviceClass ?? "—"} />
              <KV k="Server says online" v={status.data?.online ? "yes" : "no"} />
              <KV k="Last seen" v={formatRelative(status.data?.lastSeen)} />
              <KV k="Uptime" v={formatUptime(t?.uptime)} />
              <KV k="Risk score" v={t?.riskScore != null ? `${t.riskScore}` : "—"} />
              <KV k="Thermal" v={t?.thermalTemp != null ? `${t.thermalTemp.toFixed(1)}°C` : "—"} />
              <KV k="Buffer fill" v={t?.bufferLevel != null ? `${t.bufferLevel}%` : "—"} />
            </div>
          )}
        </CardContent>
      </Card>

      <Separator />

      <RegisterPanel />
    </div>
  );
}

function RegisterPanel() {
  const { serverUrl, deviceId } = useVyzorixConfig();
  const [firebaseInstallId, setFid] = useState("dev-fid-" + Math.random().toString(36).slice(2, 10));
  const [fcmToken, setFcm] = useState("dev-fcm-token-" + Math.random().toString(36).slice(2, 14));
  const [appVersion, setAv] = useState("1.0.0-mock");
  const [deviceClass, setDc] = useState("nokia_c22");
  const [secret, setSecret] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setBusy(true);
    try {
      const res = await registerDevice(serverUrl, { deviceId, firebaseInstallId, fcmToken, appVersion, deviceClass });
      setSecret(res.commandSecret);
      toast.success("Device registered", { description: `command_secret ${shortHash(res.commandSecret)}` });
    } catch (e) {
      toast.error("Registration failed", { description: e instanceof Error ? e.message : String(e) });
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Register device</CardTitle>
        <CardDescription>POST /v1/device/register · idempotent on (deviceId, firebaseInstallId)</CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="firebaseInstallId" value={firebaseInstallId} onChange={setFid} />
          <Field label="fcmToken" value={fcmToken} onChange={setFcm} />
          <Field label="appVersion" value={appVersion} onChange={setAv} />
          <Field label="deviceClass" value={deviceClass} onChange={setDc} />
        </div>
        <div className="flex items-center justify-between gap-3">
          <p className="text-xs text-muted-foreground">
            {secret ? <>command_secret returned: <code>{shortHash(secret)}</code></> : "Returned exactly once on success."}
          </p>
          <Button onClick={submit} disabled={busy}>{busy ? "Registering…" : "Register"}</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs">{label}</Label>
      <Input value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium break-all">{v}</p>
    </div>
  );
}
