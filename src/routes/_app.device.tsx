import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { getDeviceStatus, registerDevice } from "@/lib/vyzorix-api";
import { useStream } from "@/lib/device-stream-context";
import { StatusBadge, type DeviceHealth } from "@/components/status-badge";
import { formatRelative, formatUptime, shortHash } from "@/lib/format";

export const Route = createFileRoute("/_app/device")({
  head: () => ({ meta: [{ title: "Device — Vyzorix" }] }),
  component: DevicePage,
});

// Format device class for display (e.g., "nokia_c22" -> "Nokia C22")
function formatDeviceClass(deviceClass: string | undefined): string {
  if (!deviceClass) return "Unknown Device";
  return deviceClass
    .replace(/_/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

function DevicePage() {
  const { serverUrl, deviceId, thresholds } = useVyzorixConfig();
  const stream = useStream();
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
      : (t?.riskScore ?? 0) >= thresholds.riskCrit || (t?.thermalTemp ?? 0) >= thresholds.thermalCrit
      ? "critical"
      : (t?.riskScore ?? 0) >= thresholds.riskWarn || (t?.thermalTemp ?? 0) >= thresholds.thermalWarn
      ? "warning"
      : "online";

  const deviceDisplayName = formatDeviceClass(status.data?.deviceClass);

  return (
    <div className="space-y-4">
      {status.isLoading ? (
        <Card>
          <CardHeader>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div className="space-y-2">
                <div className="h-5 w-48 animate-pulse rounded-md bg-muted" />
                <div className="h-4 w-64 animate-pulse rounded-md bg-muted" />
              </div>
              <div className="h-6 w-20 animate-pulse rounded-full bg-muted" />
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              {Array.from({ length: 8 }).map((_, i) => (
                <div key={i} className="rounded-md border p-3">
                  <Skeleton className="h-3 w-16" />
                  <Skeleton className="mt-2 h-5 w-24" />
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      ) : (
      <Card>
        <CardHeader className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle>{deviceDisplayName} — primary</CardTitle>
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
      )}

      <Separator />

      <RegisterPanel />
    </div>
  );
}

function RegisterPanel() {
  const { serverUrl, deviceId } = useVyzorixConfig();
  // Persist the registration form across navigations — previously a fresh
  // random fid/token was generated every mount, which is why "nothing seemed
  // to save".
  const persist = (key: string, fallback: () => string) => {
    const k = `vyzorix.register.${key}`;
    const [v, setV] = useState<string>(() => {
      try { return localStorage.getItem(k) ?? fallback(); } catch { return fallback(); }
    });
    useEffect(() => { try { localStorage.setItem(k, v); } catch {} }, [v, k]);
    return [v, setV] as const;
  };
  const [firebaseInstallId, setFid] = persist("fid", () => "dev-fid-" + Math.random().toString(36).slice(2, 10));
  const [fcmToken, setFcm] = persist("fcm", () => "dev-fcm-token-" + Math.random().toString(36).slice(2, 14));
  const [appVersion, setAv] = persist("appVersion", () => "1.0.0-mock");
  const [deviceClass, setDc] = persist("deviceClass", () => "nokia_c22");
  const [secret, setSecret] = useState<string | null>(() => {
    try { return localStorage.getItem(`vyzorix.register.secret.${deviceId}`); } catch { return null; }
  });
  const [busy, setBusy] = useState(false);

  const submit = async () => {
    setBusy(true);
    try {
      const res = await registerDevice(serverUrl, { deviceId, firebaseInstallId, fcmToken, appVersion, deviceClass });
      setSecret(res.commandSecret);
      try { localStorage.setItem(`vyzorix.register.secret.${deviceId}`, res.commandSecret); } catch {}
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
