import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ShieldAlert } from "lucide-react";
import { toast } from "sonner";

import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/settings/operator")({
  component: OperatorSettings,
});

function OperatorSettings() {
  const cfg = useVyzorixConfig();
  const [name, setName] = useState(cfg.operator.name);
  const [email, setEmail] = useState(cfg.operator.email);
  const [role, setRole] = useState(cfg.operator.role);
  const [notifications, setNotifications] = useState(cfg.notificationsEnabled);

  const save = () => {
    cfg.update({
      operator: { name: name.trim(), email: email.trim(), role },
      notificationsEnabled: notifications,
    });
    toast.success("Operator profile saved");
  };

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Operator identity</CardTitle>
          <CardDescription>Attached to every command for audit purposes (Phase 1.5+).</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="space-y-1.5">
            <Label>Display name</Label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. J. Mokoena" />
          </div>
          <div className="space-y-1.5">
            <Label>Email</Label>
            <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} placeholder="operator@example.com" />
          </div>
          <div className="space-y-1.5">
            <Label>Role</Label>
            <Select value={role} onValueChange={(v) => setRole(v as typeof role)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="viewer">Viewer — read-only</SelectItem>
                <SelectItem value="operator">Operator — dispatch routine commands</SelectItem>
                <SelectItem value="super_admin">Super admin — destructive + key rotation</SelectItem>
              </SelectContent>
            </Select>
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
            <Button onClick={save}>Save operator</Button>
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