import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, useRef } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { updateSettings, me } from "@/lib/vyzorix-auth";
import { DEFAULT_SETTINGS, useVyzorixConfig, type Thresholds } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/settings/thresholds")({
  component: ThresholdSettings,
});

// eslint-disable-next-line func-style
function ThresholdSettings(): JSX.Element {
  const cfg = useVyzorixConfig();
  const cfgRef = useRef(cfg);
  cfgRef.current = cfg;
  const [t, setT] = useState<Thresholds>(cfg.thresholds);
  const [loading, setLoading] = useState(false);

  const canEdit = cfg.operator.role === "super_admin";

  // Load thresholds from server on mount
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const loadFromServer = async () => {
      try {
        const op = await me(cfgRef.current.serverUrl);
        if (op.thresholds) {
          setT(op.thresholds);
          cfgRef.current.update({ thresholds: op.thresholds });
        }
      } catch {
        // Use local defaults if server fetch fails
      }
    };
    loadFromServer();
  }, []);

  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const save = async () => {
    setLoading(true);
    try {
      await updateSettings(cfg.serverUrl, { thresholds: t });
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

  // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
  const reset = () => {
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
}

// eslint-disable-next-line func-style
function NumField({
  label,
  value,
  onChange,
  disabled,
}: {
  label: string;
  value: number;
  onChange: (v: number) => void;
  disabled?: boolean;
}) {
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
}
