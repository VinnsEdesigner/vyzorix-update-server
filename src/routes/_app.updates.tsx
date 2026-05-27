import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";
import { mockUpdateHistory, mockChangelog, mockDevices } from "@/lib/mock-data";
import { CloudUpload, Download, Zap } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/_app/updates")({
  head: () => ({ meta: [{ title: "Updates — Vyzorix" }] }),
  component: UpdatesPage,
});

const updateStates = ["DOWNLOADED", "INSTALLING", "AVAILABLE", "NOT_CHECKED", "SUCCESS", "AVAILABLE", "DOWNLOADING", "SUCCESS"] as const;

function UpdatesPage() {
  return (
    <div className="space-y-4">
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Current release</CardTitle>
            <CardDescription>Served at <code className="text-xs">/api/v1/version</code></CardDescription>
          </CardHeader>
          <CardContent className="grid gap-3 sm:grid-cols-3">
            <KV k="Version" v="2.2.0" />
            <KV k="Version code" v="220" />
            <KV k="Build" v="220" />
            <KV k="Min SDK" v="26" />
            <KV k="Size" v="12.4 MB" />
            <KV k="Forced" v="No" />
            <KV k="Released" v="2026-05-20" />
            <KV k="Checksum" v="a93f…7b1c" />
            <KV k="Channel" v="stable" />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Distribution</CardTitle>
            <CardDescription>Fleet update adoption</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Adoption label="2.2.0" pct={62} />
            <Adoption label="2.1.0" pct={25} />
            <Adoption label="2.0.0" pct={13} />
            <Button className="w-full" onClick={() => toast.success("WAKE_UP_UPDATER broadcast queued for fleet")}>
              <Zap className="h-4 w-4" /> Wake updater on fleet
            </Button>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Release history</CardTitle>
            <CardDescription>Updated by GitHub Actions on tagged releases</CardDescription>
          </div>
          <Button variant="outline" size="sm" onClick={() => toast.info("Triggered manual version.json regen")}>
            <CloudUpload className="h-4 w-4" /> Regenerate version.json
          </Button>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Version</TableHead>
                <TableHead>Code</TableHead>
                <TableHead>Released</TableHead>
                <TableHead>Size</TableHead>
                <TableHead>Forced</TableHead>
                <TableHead>Downloads</TableHead>
                <TableHead className="w-[80px]"></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {mockUpdateHistory.map((u) => (
                <TableRow key={u.version}>
                  <TableCell className="font-medium">{u.version}</TableCell>
                  <TableCell className="font-mono text-xs">{u.versionCode}</TableCell>
                  <TableCell>{u.releaseDate}</TableCell>
                  <TableCell>{u.fileSize}</TableCell>
                  <TableCell>{u.forced ? <Badge variant="destructive">Forced</Badge> : <Badge variant="outline">No</Badge>}</TableCell>
                  <TableCell>{u.downloads}</TableCell>
                  <TableCell><Button variant="ghost" size="sm"><Download className="h-4 w-4" /></Button></TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader><CardTitle className="text-base">Changelog (2.2.0)</CardTitle></CardHeader>
          <CardContent>
            <ul className="list-inside list-disc space-y-1.5 text-sm text-muted-foreground">
              {mockChangelog.map((c, i) => <li key={i}>{c}</li>)}
            </ul>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Per-device update state</CardTitle>
            <CardDescription>Mock from TelemetryFrame update fields</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {mockDevices.slice(0, 6).map((d, i) => (
              <div key={d.id} className="flex items-center justify-between gap-3 rounded-md border p-2.5 text-sm">
                <div className="min-w-0">
                  <p className="truncate font-medium">{d.name}</p>
                  <p className="text-xs text-muted-foreground">app {d.appVersion}</p>
                </div>
                <Badge variant={updateStates[i] === "SUCCESS" ? "default" : updateStates[i] === "AVAILABLE" ? "secondary" : "outline"}>
                  {updateStates[i]}
                </Badge>
              </div>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium">{v}</p>
    </div>
  );
}

function Adoption({ label, pct }: { label: string; pct: number }) {
  return (
    <div>
      <div className="mb-1 flex items-center justify-between text-xs">
        <span className="font-medium">{label}</span>
        <span className="text-muted-foreground">{pct}%</span>
      </div>
      <Progress value={pct} />
    </div>
  );
}