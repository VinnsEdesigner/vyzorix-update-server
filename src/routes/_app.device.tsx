import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, type ReactElement, type JSX } from "react";
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

function formatDeviceClass(deviceClass: string | undefined): string {
  if (!deviceClass) return "Unknown Device";
  return deviceClass.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

function KV({ k, v }: { k: string; v: string }): ReactElement {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium break-all">{v}</p>
    </div>
  );
}

function computeDeviceHealth(
  online: boolean,
  streamConnected: string,
  riskScore: number | undefined,
  thermal: number | undefined,
  thresholds: { riskCrit: number; riskWarn: number; thermalCrit: number; thermalWarn: number },
): DeviceHealth {
  if (!online && streamConnected !== "connected") return "offline";
  const risk = riskScore ?? 0;
  const thermalVal = thermal ?? 0;
  if (risk >= thresholds.riskCrit || thermalVal >= thresholds.thermalCrit) return "critical";
  if (risk >= thresholds.riskWarn || thermalVal >= thresholds.thermalWarn) return "warning";
  return "online";
}

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

function DevicePage(): JSX.Element {
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

  const health: DeviceHealth = computeDeviceHealth(
    Boolean(status.data?.online),
    stream.state,
    t?.riskScore,
    t?.thermalTemp,
    thresholds,
  );

  const deviceDisplayName = formatDeviceClass(status.data?.deviceClass);

  // No device configured
  if (!deviceId) {
    return (
      <div className="space-y-4">
        <Card>
          <CardContent className="py-4">
            <p className="text-sm text-muted-foreground">
              No device configured. Set deviceId in Settings → Connection, then use the registration
              panel below to register your device.
            </p>
          </CardContent>
        </Card>
        <Separator />
        <RegisterPanel deviceStatus={null} />
      </div>
    );
  }

  // Loading state
  if (status.isLoading) {
    return (
      <div className="space-y-4">
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
        <Separator />
        <RegisterPanel deviceStatus={null} />
      </div>
    );
  }

  // Loaded state
  return (
    <div className="space-y-4">
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
              Device not registered yet. Use the registration panel below or run the Android daemon
              to call <code className="text-xs">POST /v1/device/register</code>.
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
      <RegisterPanel deviceStatus={status.data ?? null} />
    </div>
  );
}

function RegisterPanel({ deviceStatus }: { deviceStatus: DeviceStatus | null }): JSX.Element {
  const { serverUrl, deviceId } = useVyzorixConfig();

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

  useEffect(() => {
    if (deviceStatus) {
      setFid(deviceStatus.firebaseInstallId ?? "");
      setFcm(deviceStatus.fcmToken ?? "");
      setAv(deviceStatus.appVersion ?? "");
      setDc(deviceStatus.deviceClass ?? "");
    }
  }, [deviceStatus]);

  const submit = async (): Promise<void> => {
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

export const Route = createFileRoute("/_app/device")({
  head: () => ({ meta: [{ title: "Device — Vyzorix" }] }),
  component: DevicePage,
});
