import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, useRef } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { ShieldAlert } from "lucide-react";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

import { useVyzorixConfig, DEFAULT_SERVER_URL } from "@/lib/vyzorix-config";
import { getStoredOperator, updateName } from "@/lib/vyzorix-auth";

export const Route = createFileRoute("/_app/settings/operator")({
  ssr: false,
  component: OperatorSettings,
});

function OperatorSettings() {
  const cfg = useVyzorixConfig();
  const stored = getStoredOperator();

  // Email and role come from the database — read-only, pre-filled from stored operator
  const [email, setEmail] = useState(stored?.email ?? "");
  const [role, setRole] = useState(stored?.role ?? "operator");
  const [notifications, setNotifications] = useState(cfg.notificationsEnabled);

  // Name is editable — syncs to the Go server on every change
  const [name, setName] = useState(stored?.name ?? "");
  const [savingName, setSavingName] = useState(false);

  // Auto-save name after 1 second of no typing
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    // Sync from stored operator if it changed (e.g. after login)
    const op = getStoredOperator();
    if (op) {
      setName(op.name);
      setEmail(op.email);
      setRole(op.role);
    }
  }, []);

  useEffect(() => {
    if (saveTimer.current) clearTimeout(saveTimer.current);
    if (!stored) return;

    // Only save if name actually changed from what we have stored
    if (name.trim() && name.trim() !== stored.name) {
      saveTimer.current = setTimeout(async () => {
        setSavingName(true);
        try {
          const updated = await updateName(DEFAULT_SERVER_URL, name.trim());
          toast.success("Display name saved");
        } catch (e) {
          toast.error("Failed to save name", { description: e instanceof Error ? e.message : "try again" });
        } finally {
          setSavingName(false);
        }
      }, 1000);
    }

    return () => {
      if (saveTimer.current) clearTimeout(saveTimer.current);
    };
  }, [name]);

  // Emit event so sidebar and other components see the updated name
  useEffect(() => {
    if (savingName) return; // don't fire while still saving
    window.dispatchEvent(new Event("vyz.operator.updated"));
  }, [name, savingName]);

  const saveNotifications = () => {
    cfg.update({ notificationsEnabled: notifications });
    toast.success("Notification settings saved");
  };

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Operator identity</CardTitle>
          <CardDescription>
            Set during sign-up. Name is editable; email and role are server-controlled.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-1.5">
            <Label>Display name</Label>
            <div className="relative">
              <Input
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. J. Mokoena"
              />
              {savingName && (
                <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
              )}
            </div>
            {name.trim() !== stored?.name && !savingName && (
              <p className="text-xs text-rose-500">Saving…</p>
            )}
            {name.trim() === stored?.name && name.trim() !== "" && (
              <p className="text-xs text-muted-foreground">Saved</p>
            )}
          </div>
          <div className="space-y-1.5">
            <Label>Email</Label>
            <Input value={email} readOnly disabled className="bg-muted/50 cursor-not-allowed" />
            <p className="text-xs text-muted-foreground">Set during registration. Cannot be changed.</p>
          </div>
          <div className="space-y-1.5">
            <Label>Role</Label>
            <Input value={role} readOnly disabled className="bg-muted/50 cursor-not-allowed capitalize" />
            <p className="text-xs text-muted-foreground">Server-controlled. Promotions require a super_admin.</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Privileges</CardTitle>
          <CardDescription>What the current role can do</CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <RoleRow allowed={true} label="View dashboard, alerts, signals" />
          <RoleRow allowed={role !== "viewer"} label="Dispatch FORCE_SPEAKER, RESET_AUDIO_HAL, REQUEST_STATUS" />
          <RoleRow allowed={role === "super_admin"} label="DUMP_FLIGHT_DATA, ROTATE_KEYS, edit thresholds" />
          <RoleRow allowed={role === "super_admin"} label="Reset all dashboard configuration" />
        </CardContent>
      </Card>

      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle>Notifications</CardTitle>
          <CardDescription>How the dashboard surfaces threshold breaches</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-start justify-between gap-3 rounded-md border p-3">
            <div>
              <p className="text-sm font-medium">Toast notifications on threshold breach</p>
              <p className="text-xs text-muted-foreground">Pop a toast every time risk or thermal crosses the critical threshold</p>
            </div>
            <Switch checked={notifications} onCheckedChange={setNotifications} />
          </div>
          {role === "super_admin" && (
            <div className="flex items-start gap-3 rounded-md border border-destructive/40 bg-destructive/5 p-3">
              <ShieldAlert className="mt-0.5 h-4 w-4 text-destructive" />
              <p className="text-xs text-muted-foreground">
                Super-admin role is active. Destructive commands are unlocked across the dashboard.
              </p>
            </div>
          )}
          <div className="flex justify-end">
            <Button onClick={saveNotifications}>Save notification settings</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function RoleRow({ allowed, label }: { allowed: boolean; label: string }) {
  return (
    <div className="flex items-center justify-between rounded-md border p-2.5">
      <span className={allowed ? "" : "text-muted-foreground"}>{label}</span>
      <Badge variant={allowed ? "default" : "outline"}>{allowed ? "allowed" : "blocked"}</Badge>
    </div>
  );
}