import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/settings")({
  head: () => ({ meta: [{ title: "Settings — Vyzorix" }] }),
  component: SettingsPage,
});

function SettingsPage() {
  const [endpoint, setEndpoint] = useState("https://vyzorix-update-server.onrender.com");
  const [interval, setInterval] = useState("3600");
  const [rate, setRate] = useState([60]);
  const [cooldown, setCooldown] = useState([30]);
  const [forced, setForced] = useState(false);
  const [verbose, setVerbose] = useState(true);

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Server</CardTitle>
          <CardDescription>Endpoints and credentials</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-1.5">
            <Label htmlFor="endpoint">Base endpoint</Label>
            <Input id="endpoint" value={endpoint} onChange={(e) => setEndpoint(e.target.value)} />
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="interval">Update check interval (s)</Label>
            <Input id="interval" type="number" value={interval} onChange={(e) => setInterval(e.target.value)} />
          </div>
          <Separator />
          <Row label="Force latest version" hint="Override device-side prompt">
            <Switch checked={forced} onCheckedChange={setForced} />
          </Row>
          <Row label="Verbose logging" hint="Stream INFO+ to dashboard terminal">
            <Switch checked={verbose} onCheckedChange={setVerbose} />
          </Row>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Thresholds</CardTitle>
          <CardDescription>Operational guardrails</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <div className="flex justify-between text-sm"><Label>Rate limit (req/min)</Label><span className="text-muted-foreground">{rate[0]}</span></div>
            <Slider value={rate} onValueChange={setRate} min={10} max={240} step={10} />
          </div>
          <div className="space-y-2">
            <div className="flex justify-between text-sm"><Label>Command cooldown (s)</Label><span className="text-muted-foreground">{cooldown[0]}</span></div>
            <Slider value={cooldown} onValueChange={setCooldown} min={5} max={120} step={5} />
          </div>
        </CardContent>
      </Card>

      <Card className="lg:col-span-2">
        <CardHeader>
          <CardTitle>Session</CardTitle>
          <CardDescription>Admin account · vyzorix-ops</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-wrap items-center justify-between gap-3">
          <div className="text-sm text-muted-foreground">JWT expires in 1h 24m. CORS origins: <code>android-app://com.vyzorix.audiorouter</code></div>
          <div className="flex gap-2">
            <Button variant="outline" onClick={() => toast.info("Token rotated")}>Rotate token</Button>
            <Button onClick={() => toast.success("Settings saved")}>Save changes</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function Row({ label, hint, children }: { label: string; hint?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <div>
        <p className="text-sm font-medium">{label}</p>
        {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
      </div>
      {children}
    </div>
  );
}