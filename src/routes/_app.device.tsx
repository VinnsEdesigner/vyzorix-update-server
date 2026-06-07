import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, type ReactElement } from "react";
import { toast } from "sonner";

import { StatusBadge, type DeviceHealth } from "@/components/status-badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { useStream } from "@/lib/device-stream-context";
import { formatRelative, formatUptime, shortHash } from "@/lib/format";
import { getDeviceStatus, registerDevice, type DeviceStatus } from "@/lib/vyzorix-api";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/device")({
  head: () => ({ meta: [{ title: "Device — Vyzorix" }] }),
  component: DevicePage,
});

// Format device class for display (e.g., "nokia_c22" -> "Nokia C22")
// eslint-disable-next-line func-style
function formatDeviceClass(deviceClass: string | undefined): string {
  if (!deviceClass) return "Unknown Device";
  return deviceClass.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

// eslint-disable-next-line func-style
function DevicePage() {
  const { serverUrl, deviceId, thresholds } = useVyzorixConfig();
  const stream = useStream();
  const t = stream.lastTelemetry;

  const status = useQuery({
    queryKey: ["vyzorix", "status", serverUrl, deviceId],
    queryFn: () => getDeviceStatus(serverUrl, deviceId),
    enabled: Boolean(serverUrl) && Boolean(deviceId),
    refetchInterval: 10_000,
    retry: false,
  });

  const health: DeviceHealth =
    // eslint-disable-next-line no-nested-ternary
    !status.data?.online && stream.state !== "connected"
      ? "offline"
      : // eslint-disable-next-line no-nested-ternary
        (t?.riskScore ?? 0) >= thresholds.riskCrit ||
          (t?.thermalTemp ?? 0) >= thresholds.thermalCrit
        ? "critical"
        : (t?.riskScore ?? 0) >= thresholds.riskWarn ||
            (t?.thermalTemp ?? 0) >= thresholds.thermalWarn
          ? "warning"
          : "online";

  const deviceDisplayName = formatDeviceClass(status.data?.deviceClass);

  return (
    <div className="space-y-4">
      {/* eslint-disable-next-line no-nested-ternary */}
      {!deviceId ? (
        <Card>
          <CardContent className="py-4">
            <p className="text-sm text-muted-foreground">
              No device configured. Set deviceId in Settings → Connection, then use the registration
              panel below to register your device.
            </p>
          </CardContent>
        </Card>
      ) : status.isLoading ? (
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
                Device not registered yet. Use the registration panel below or run the Android
                daemon to call <code className="text-xs">POST /v1/device/register</code>.
              </p>
            ) : (
              <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                <KV k="App version" v={status.data?.appVersion ?? "—"} />
                <KV k="Device class" v={status.data?.deviceClass ?? "—"} />
                <KV k="Server says online" v={status.data?.online ? "yes" : "no"} />
                <KV k="Last seen" v={formatRelative(status.data?.lastSeen)} />
                <KV k="Uptime" v={formatUptime(t?.uptime)} />
                <KV k="Risk score" v={t?.riskScore != null ? `${t.riskScore}` : "—"} />
                <KV
                  k="Thermal"
                  v={t?.thermalTemp != null ? `${t.thermalTemp.toFixed(1)}°C` : "—"}
                />
                <KV k="Buffer fill" v={t?.bufferLevel != null ? `${t.bufferLevel}%` : "—"} />
              </div>
            )}
          </CardContent>
        </Card>
      )}
      <Separator />
      <RegisterPanel deviceStatus={status.data ?? null} />
    </div>
  );
}

// eslint-disable-next-line func-style
function RegisterPanel({ deviceStatus }: { deviceStatus: DeviceStatus | null }) {
  const { serverUrl, deviceId } = useVyzorixConfig();

  // Load defaults from server status (persisted in DB) instead of localStorage.
  // Only keep command_secret in localStorage since it's returned once on registration.
  const [firebaseInstallId, setFid] = useState(deviceStatus?.firebaseInstallId ?? "");
  const [fcmToken, setFcm] = useState(deviceStatus?.fcmToken ?? "");
  const [appVersion, setAv] = useState(deviceStatus?.appVersion ?? "");
  const [deviceClass, setDc] = useState(deviceStatus?.deviceClass ?? "");
  const [secret, setSecret] = useState<string | null>(() => {
    try {
      return localStorage.getItem(`vyzorix.register.secret.${deviceId}`);
    } catch {
      return null;
    }
  });
  const [busy, setBusy] = useState(false);

  // Sync form fields when server status loads (e.g. after re-registration)
  useEffect(() => {
    if (deviceStatus) {
      setFid(deviceStatus.firebaseInstallId ?? "");
      setFcm(deviceStatus.fcmToken ?? "");
      setAv(deviceStatus.appVersion ?? "");
      setDc(deviceStatus.deviceClass ?? "");
    }
  }, [deviceStatus]);

  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const submit = async () => {
    if (!deviceId.trim()) {
      toast.error("Registration failed", {
        description: "deviceId is required — set it in Settings → Connection",
      });
      return;
    }
    if (!firebaseInstallId.trim()) {
      toast.error("Registration failed", { description: "firebaseInstallId is required" });
      return;
    }
    setBusy(true);
    try {
      const res = await registerDevice(serverUrl, {
        deviceId,
        firebaseInstallId: firebaseInstallId.trim(),
        fcmToken: fcmToken.trim(),
        appVersion: appVersion.trim(),
        deviceClass: deviceClass.trim(),
      });
      setSecret(res.commandSecret);
      try {
        localStorage.setItem(`vyzorix.register.secret.${deviceId}`, res.commandSecret);
      } catch {
        // ignore storage error
      }
      toast.success("Device registered", {
        description: `command_secret ${shortHash(res.commandSecret)}`,
      });
    } catch (e) {
      toast.error("Registration failed", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setBusy(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Register device</CardTitle>
        <CardDescription>
          POST /v1/device/register · idempotent on (deviceId, firebaseInstallId) · fields pre-filled
          from server when device is registered
        </CardDescription>
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
            {secret ? (
              <>
                command_secret returned: <code>{shortHash(secret)}</code>
              </>
            ) : (
              "Returned exactly once on success."
            )}
          </p>
          <Button onClick={submit} disabled={busy}>
            {busy ? "Registering…" : "Register"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

// eslint-disable-next-line func-style
function Field({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
}): ReactElement {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs">{label}</Label>
      <Input value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}

// eslint-disable-next-line func-style
function KV({ k, v }: { k: string; v: string }): ReactElement {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium break-all">{v}</p>
    </div>
  );
}
