import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/settings/advanced")({
  component: AdvancedSettings,
});

function AdvancedSettings() {
  const cfg = useVyzorixConfig();
  const [logLimit, setLogLimit] = useState(cfg.logBufferLimit);
  const [signalLimit, setSignalLimit] = useState(cfg.signalHistoryLimit);

  const canDanger = cfg.operator.role === "super_admin";

  const save = () => {
    cfg.update({
      logBufferLimit: Math.max(50, Math.min(5000, logLimit || 500)),
      signalHistoryLimit: Math.max(30, Math.min(2000, signalLimit || 240)),
    });
    toast.success("Advanced settings saved · refresh to apply buffer sizes");
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
            Reset every dashboard configuration to defaults. Does not touch the device or server.
          </CardDescription>
        </CardHeader>
        <CardContent className="flex items-center justify-between gap-3">
          <p className="text-xs text-muted-foreground">
            {canDanger
              ? "Super-admin role required and active."
              : "Switch to super-admin in Operator to enable."}
          </p>
          <Button
            variant="destructive"
            disabled={!canDanger}
            onClick={() => {
              cfg.reset();
              toast.info("All dashboard settings reset");
            }}
          >
            Reset all settings
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
