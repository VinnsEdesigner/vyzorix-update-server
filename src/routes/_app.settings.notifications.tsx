import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/settings/notifications")({
  component: NotificationsSettings,
});

function NotificationsSettings() {
  const { notificationsEnabled, update } = useVyzorixConfig();

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
        <CardDescription>Alert sounds, browser pushes, and toast behaviour.</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Row label="Toast notifications" hint="In-app sonner toasts for command results and errors">
          <Switch checked={notificationsEnabled} onCheckedChange={(v) => update({ notificationsEnabled: v })} />
        </Row>
        <Row label="Browser notifications" hint="Native OS pushes for critical alerts (requires permission)">
          <Button variant="outline" size="sm" onClick={requestBrowser}>Request permission</Button>
        </Row>
      </CardContent>
    </Card>
  );
}

function Row({ label, hint, children }: { label: string; hint: string; children: React.ReactNode }) {
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