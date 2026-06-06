import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

import { useVyzorixConfig, DEFAULT_SETTINGS } from "@/lib/vyzorix-config";
import { resetSettings } from "@/lib/vyzorix-auth";

export const Route = createFileRoute("/_app/settings/advanced")({
  component: AdvancedSettings,
});

function AdvancedSettings() {
  const cfg = useVyzorixConfig();
  const [logLimit, setLogLimit] = useState(cfg.logBufferLimit);
  const [signalLimit, setSignalLimit] = useState(cfg.signalHistoryLimit);
  const [resetting, setResetting] = useState(false);

  const canDanger = cfg.operator.role === "super_admin";

  const save = () => {
    cfg.update({
      logBufferLimit: Math.max(50, Math.min(5000, logLimit || 500)),
      signalHistoryLimit: Math.max(30, Math.min(2000, signalLimit || 240)),
    });
    toast.success("Advanced settings saved · refresh to apply buffer sizes");
  };

  const handleReset = async () => {
    if (!canDanger) return;
    setResetting(true);
    try {
      const op = await resetSettings(cfg.serverUrl);
      // Update local config from server response
      cfg.update({
        thresholds: {
          riskWarn: op.thresholds?.riskWarn ?? DEFAULT_SETTINGS.thresholds.riskWarn,
          riskCrit: op.thresholds?.riskCrit ?? DEFAULT_SETTINGS.thresholds.riskCrit,
          thermalWarn: op.thresholds?.thermalWarn ?? DEFAULT_SETTINGS.thresholds.thermalWarn,
          thermalCrit: op.thresholds?.thermalCrit ?? DEFAULT_SETTINGS.thresholds.thermalCrit,
          bufferWarn: op.thresholds?.bufferWarn ?? DEFAULT_SETTINGS.thresholds.bufferWarn,
          bufferCrit: op.thresholds?.bufferCrit ?? DEFAULT_SETTINGS.thresholds.bufferCrit,
        },
        autoReconnect: op.client?.autoReconnect ?? true,
        strictHmac: op.client?.strictHmac ?? false,
        notificationsEnabled: op.client?.notificationsEnabled ?? true,
      });
      cfg.reset(); // Reset browser-only settings too
      toast.success("All settings reset to defaults on server");
    } catch (e) {
      toast.error("Reset failed", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setResetting(false);
    }
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Buffers</CardTitle>
          <CardDescription>Memory limits for the in-browser signal and log buffers</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label>Log retention (entries)</Label>
            <Input
              type="number"
              min={50}
              max={5000}
              step={50}
              value={logLimit}
              onChange={(e) => setLogLimit(Number(e.target.value))}
            />
            <p className="text-xs text-muted-foreground">
              Diagnostics terminal keeps the last N frames in a rolling buffer.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label>Signal history (frames)</Label>
            <Input
              type="number"
              min={30}
              max={2000}
              step={30}
              value={signalLimit}
              onChange={(e) => setSignalLimit(Number(e.target.value))}
            />
            <p className="text-xs text-muted-foreground">
              How many recent signals power the live charts.
            </p>
          </div>
        </CardContent>
      </Card>

      <div className="flex justify-end">
        <Button onClick={save}>Save advanced</Button>
      </div>

      <Separator />

      <Card className="border-destructive/40">
        <CardHeader>
          <CardTitle className="text-destructive">Danger zone</CardTitle>
          <CardDescription>
            Reset every server setting (thresholds, client preferences) to defaults. Requires
            super-admin,
          </CardDescription>
        </CardHeader>
        <CardContent className="flex items-center justify-between gap-3">
          <p className="text-xs text-muted-foreground">
            {canDanger
              ? "Super-admin role active — resets server-side operator settings."
              : "Switch to super-admin in Operator to enable."}
          </p>
          <Button variant="destructive" disabled={!canDanger || resetting} onClick={handleReset}>
            {resetting ? "Resetting..." : "Reset all settings"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
