import { createFileRoute } from "@tanstack/react-router";
import { ShieldAlert, Loader2 } from "lucide-react";
import { useState, useEffect, useRef, useCallback } from "react";
import { toast } from "sonner";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  getStoredOperator,
  updateName,
  updateSettings,
  type ClientSettings,
} from "@/lib/vyzorix-auth";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

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
  const [lastSavedName, setLastSavedName] = useState(stored?.name ?? "");

  // Auto-save name after 1 second of no typing
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const nameRef = useRef(name); // Keep ref in sync with latest name
  nameRef.current = name;

  // Sync from stored operator if it changes (e.g. after login)
  useEffect(() => {
    const op = getStoredOperator();
    if (op) {
      setName(op.name);
      setLastSavedName(op.name); // Also update lastSavedName to match
      setEmail(op.email);
      setRole(op.role);
    }
  }, []);

  // Debounced save function - reads fresh values from refs
  const saveName = useCallback(
    async (nameToSave: string) => {
      setSavingName(true);
      try {
        await updateName(cfg.serverUrl, nameToSave.trim());
        setLastSavedName(nameToSave.trim());
        toast.success("Display name saved");
      } catch (e) {
        toast.error("Failed to save name", {
          description: e instanceof Error ? e.message : "try again",
        });
      } finally {
        setSavingName(false);
      }
    },
    [cfg.serverUrl],
  );

  // Auto-save effect - properly handles race conditions
  useEffect(() => {
    // Clear any existing timer
    if (saveTimer.current) {
      clearTimeout(saveTimer.current);
      saveTimer.current = null;
    }

    const trimmedName = name.trim();

    // Don't save if empty or unchanged from last saved
    if (!trimmedName || trimmedName === lastSavedName) {
      return;
    }

    // Debounce: save after 1 second of no typing
    saveTimer.current = setTimeout(() => {
      const currentName = nameRef.current.trim();
      // Double-check before saving
      if (currentName && currentName !== lastSavedName) {
        saveName(currentName);
      }
    }, 1000);

    // Cleanup on unmount or before next effect run
    return () => {
      if (saveTimer.current) {
        clearTimeout(saveTimer.current);
        saveTimer.current = null;
      }
    };
  }, [name, lastSavedName, saveName]);

  // Emit event so sidebar and other components see the updated name
  useEffect(() => {
    if (savingName) return; // don't fire while still saving
    window.dispatchEvent(new Event("vyz.operator.updated"));
  }, [name, savingName]);

  const saveNotifications = async () => {
    setSavingName(true);
    try {
      const client: ClientSettings = { notificationsEnabled: notifications };
      await updateSettings(cfg.serverUrl, { client });
      cfg.update({ notificationsEnabled: notifications });
      toast.success("Notification settings saved to server");
    } catch (e) {
      toast.error("Failed to save", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setSavingName(false);
    }
  };

  // Determine save status for UI
  const trimmedName = name.trim();
  const isSaving = savingName;
  const isSaved = trimmedName === lastSavedName && trimmedName !== "";
  const isUnsaved = trimmedName !== lastSavedName && !isSaving;

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
              {isSaving && (
                <Loader2 className="absolute right-3 top-1/2 -translate-y-1/2 h-4 w-4 animate-spin text-muted-foreground" />
              )}
            </div>
            {isUnsaved && <p className="text-xs text-rose-500">Saving…</p>}
            {isSaved && <p className="text-xs text-muted-foreground">Saved</p>}
          </div>
          <div className="space-y-1.5">
            <Label>Email</Label>
            <Input value={email} readOnly disabled className="bg-muted/50 cursor-not-allowed" />
            <p className="text-xs text-muted-foreground">
              Set during registration. Cannot be changed.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label>Role</Label>
            <Input
              value={role}
              readOnly
              disabled
              className="bg-muted/50 cursor-not-allowed capitalize"
            />
            <p className="text-xs text-muted-foreground">
              Server-controlled. Promotions require a super_admin.
            </p>
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
          <RoleRow
            allowed={role !== "viewer"}
            label="Dispatch FORCE_SPEAKER, RESET_AUDIO_HAL, REQUEST_STATUS"
          />
          <RoleRow
            allowed={role === "super_admin"}
            label="DUMP_FLIGHT_DATA, ROTATE_KEYS, edit thresholds"
          />
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
              <p className="text-xs text-muted-foreground">
                Pop a toast every time risk or thermal crosses the critical threshold
              </p>
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
