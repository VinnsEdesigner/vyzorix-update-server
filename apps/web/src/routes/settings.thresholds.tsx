import { createFileRoute } from "@tanstack/react-router";
import AppLayout from "@/components/layout/AppLayout";
import { useState, useEffect, useRef, type ReactElement } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/hooks/use-auth";
import { DEFAULT_SETTINGS, useVyzorixConfig, type Thresholds } from "@/lib/vyzorix-config";

const NumField = ({
  label,
  value,
  onChange,
  disabled,
}: {
  label: string;
  value: number;
  onChange: (v: number) => void;
  disabled?: boolean;
}): ReactElement => {
  return (
    <div className="space-y-1.5">
      <Label className="text-xs">{label}</Label>
      <Input
        type="number"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        disabled={disabled}
      />
    </div>
  );
};

const ThresholdSettings = (): ReactElement => {
  const cfg = useVyzorixConfig();
  const cfgRef = useRef(cfg);
  cfgRef.current = cfg;
  const { operator } = useAuth();
  const [t, setT] = useState<Thresholds>(cfg.thresholds);
  const [loading, setLoading] = useState(false);

  const canEdit = operator?.role === "super_admin";

  // Load thresholds from server on mount
  useEffect(() => {
    const loadFromServer = async (): Promise<void> => {
      try {
        const res = await fetch("/v1/auth/me", {
          method: "GET",
          credentials: "include",
        });
        if (res.ok) {
          const op = await res.json();
          if (op.thresholds) {
            setT(op.thresholds);
            cfgRef.current.update({ thresholds: op.thresholds });
          }
        }
      } catch {
        // Use local defaults if server fetch fails
      }
    };
    loadFromServer();
  }, []);

  const save = async (): Promise<void> => {
    setLoading(true);
    try {
      const res = await fetch("/v1/auth/me/settings", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ thresholds: t }),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.message ?? "Failed to save thresholds");
      }
      cfg.update({ thresholds: t });
      toast.success("Thresholds saved to server");
    } catch (e) {
      toast.error("Failed to save thresholds", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setLoading(false);
    }
  };

  const reset = (): void => {
    setT(DEFAULT_SETTINGS.thresholds);
    toast.info("Thresholds reset to defaults — save to persist");
  };

  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>Signal thresholds</CardTitle>
          <CardDescription>
            Drive the dashboard status badge, alerts page and chart reference lines.{" "}
            {canEdit
              ? "Saved to server — persists across devices."
              : "Super admin role required to edit."}
          </CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 sm:grid-cols-2">
          <NumField
            label="Risk · warning ≥"
            value={t.riskWarn}
            onChange={(v) => setT({ ...t, riskWarn: v })}
            disabled={!canEdit}
          />
          <NumField
            label="Risk · critical ≥"
            value={t.riskCrit}
            onChange={(v) => setT({ ...t, riskCrit: v })}
            disabled={!canEdit}
          />
          <NumField
            label="Thermal · warning ≥ (°C)"
            value={t.thermalWarn}
            onChange={(v) => setT({ ...t, thermalWarn: v })}
            disabled={!canEdit}
          />
          <NumField
            label="Thermal · critical ≥ (°C)"
            value={t.thermalCrit}
            onChange={(v) => setT({ ...t, thermalCrit: v })}
            disabled={!canEdit}
          />
          <NumField
            label="Buffer · warn under (%)"
            value={t.bufferWarn}
            onChange={(v) => setT({ ...t, bufferWarn: v })}
            disabled={!canEdit}
          />
          <NumField
            label="Buffer · critical under (%)"
            value={t.bufferCrit}
            onChange={(v) => setT({ ...t, bufferCrit: v })}
            disabled={!canEdit}
          />
        </CardContent>
      </Card>

      <div className="flex justify-end gap-2">
        <Button variant="outline" onClick={reset} disabled={!canEdit || loading}>
          Reset to defaults
        </Button>
        <Button onClick={save} disabled={!canEdit || loading}>
          {loading ? "Saving..." : "Save thresholds"}
        </Button>
      </div>
    </div>
  );
};

export const Route = createFileRoute("/settings/thresholds")({
  component: () => (
    <AppLayout title="Settings">
      <ThresholdSettings />
    </AppLayout>
  ),
});
