import { createFileRoute } from "@tanstack/react-router";
import { useState, useEffect, useRef, type ReactElement } from "react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { updateSettings, me, type ClientSettings } from "@/lib/vyzorix-auth";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

export const Route = createFileRoute("/_app/settings/notifications")({
  component: NotificationsSettings,
});

function NotificationsSettings(): ReactElement {
  const { notificationsEnabled, update, serverUrl } = useVyzorixConfig();
  const serverUrlRef = useRef(serverUrl);
  const updateRef = useRef(update);
  serverUrlRef.current = serverUrl;
  updateRef.current = update;
  const [enabled, setEnabled] = useState(notificationsEnabled);
  const [saving, setSaving] = useState(false);

  // Load from server on mount
  useEffect(() => {
    const loadFromServer = async () => {
      try {
        const op = await me(serverUrlRef.current);
        if (op.client) {
          setEnabled(op.client.notificationsEnabled ?? true);
          updateRef.current({ notificationsEnabled: op.client.notificationsEnabled ?? true });
        }
      } catch {
        // Use local defaults
      }
    };
    loadFromServer();
  }, []);

  const handleToggle = async (v: boolean) => {
    setEnabled(v);
    setSaving(true);
    try {
      const client: ClientSettings = { notificationsEnabled: v };
      await updateSettings(serverUrl, { client });
      update({ notificationsEnabled: v });
      toast.success("Notification settings saved to server");
    } catch (e) {
      setEnabled(!v); // revert on error
      toast.error("Failed to save", {
        description: e instanceof Error ? e.message : String(e),
      });
    } finally {
      setSaving(false);
    }
  };

  const requestBrowser = async () => {
    if (!("Notification" in window)) {
      toast.error("Browser notifications unsupported");
      return;
    }
    const p = await Notification.requestPermission();
    toast.message(`Browser permission: ${p}`);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Notifications</CardTitle>
        <CardDescription>
          Alert sounds, browser pushes, and toast behaviour. Saved to server — persists across
          devices.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Row label="Toast notifications" hint="In-app sonner toasts for command results and errors">
          <Switch checked={enabled} onCheckedChange={handleToggle} disabled={saving} />
        </Row>
        <Row
          label="Browser notifications"
          hint="Native OS pushes for critical alerts (requires permission)"
        >
          <Button variant="outline" size="sm" onClick={requestBrowser}>
            Request permission
          </Button>
        </Row>
      </CardContent>
    </Card>
  );
}

function Row({
  label,
  hint,
  children,
}: {
  label: string;
  hint: string;
  children: React.ReactNode;
}): ReactElement {
  return (
    <div className="flex items-center justify-between gap-4 rounded-md border p-3">
      <div className="space-y-0.5">
        <Label className="text-sm">{label}</Label>
        <p className="text-xs text-muted-foreground">{hint}</p>
      </div>
      {children}
    </div>
  );
}
