import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Zap, Download, ChevronDown } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { dispatchCommand, getVersion, headApk } from "@/lib/vyzorix-api";
import { formatBytes, shortHash } from "@/lib/format";

export const Route = createFileRoute("/_app/updates")({
  head: () => ({ meta: [{ title: "Updates — Vyzorix" }] }),
  component: UpdatesPage,
});

function UpdatesPage() {
  const { serverUrl, deviceId, dashboardToken } = useVyzorixConfig();

  const version = useQuery({
    queryKey: ["vyzorix", "version", serverUrl],
    queryFn: () => getVersion(serverUrl),
    retry: false,
  });

  const apkSize = useQuery({
    queryKey: ["vyzorix", "apk", serverUrl, version.data?.apk_filename],
    queryFn: () => headApk(serverUrl, version.data!.apk_filename),
    enabled: !!version.data?.apk_filename,
    retry: false,
  });

  const v = version.data;
  const apkUrl = v ? `${serverUrl.replace(/\/+$/, "")}/api/v1/apk/${v.apk_filename}` : "#";

  const wake = async () => {
    if (!deviceId.trim()) {
      toast.error("WAKE_UP_UPDATER failed", {
        description: "No device registered — set deviceId in Settings → Connection",
      });
      return;
    }
    try {
      const res = await dispatchCommand(
        serverUrl,
        deviceId,
        "WAKE_UP_UPDATER",
        undefined,
        dashboardToken,
      );
      toast.success(`WAKE_UP_UPDATER → ${res.delivery}`, {
        description: `dispatch ${res.dispatchId}`,
      });
    } catch (e) {
      toast.error("WAKE_UP_UPDATER failed", {
        description: e instanceof Error ? e.message : String(e),
      });
    }
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              Current release
              {v ? (
                <Badge variant="default">live · {v.version}</Badge>
              ) : (
                <Badge variant="outline">loading</Badge>
              )}
            </CardTitle>
            <CardDescription>
              Live manifest from <code className="text-xs">{serverUrl}/api/v1/version</code> — no
              cache, fetched on render
            </CardDescription>
          </CardHeader>
          <CardContent>
            {version.isError ? (
              <div className="space-y-2">
                <p className="text-sm text-destructive">
                  Failed to load version.json — {(version.error as Error).message}
                </p>
                <p className="text-xs text-muted-foreground">
                  Start the Go server with <code>go run .</code>, or update the URL in Settings →
                  Connection.
                </p>
              </div>
            ) : version.isLoading || !v ? (
              <p className="text-sm text-muted-foreground">Loading…</p>
            ) : (
              <div className="grid gap-3 sm:grid-cols-3">
                <KV k="Version" v={v.version} />
                <KV k="Version code" v={`${v.version_code}`} />
                <KV k="APK file" v={v.apk_filename} />
                <KV k="APK size (manifest)" v={formatBytes(v.apk_size_bytes)} />
                <KV
                  k="APK size (HEAD)"
                  v={
                    apkSize.isLoading
                      ? "checking…"
                      : apkSize.data == null
                        ? "—"
                        : formatBytes(apkSize.data)
                  }
                />
                <KV k="SHA-256" v={shortHash(v.apk_sha256, 8, 8)} />
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Trigger update</CardTitle>
            <CardDescription>For device {deviceId}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <Button className="w-full" onClick={wake} disabled={!v}>
              <Zap className="h-4 w-4" /> Wake updater on device
            </Button>
            <a href={apkUrl} target="_blank" rel="noreferrer" className="block">
              <Button variant="outline" className="w-full" disabled={!v}>
                <Download className="h-4 w-4" /> Download APK
              </Button>
            </a>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Release notes</CardTitle>
          <CardDescription>Pulled live from version.json on the update server</CardDescription>
        </CardHeader>
        <CardContent>
          {v ? (
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">{v.release_notes}</p>
          ) : (
            <p className="text-sm text-muted-foreground">—</p>
          )}
          <div className="mt-4 flex items-center gap-2">
            <Badge variant="outline">phase 1.5</Badge>
            <Badge variant="outline">real server</Badge>
            <span className="text-xs text-muted-foreground">
              Render-backed server keeps the Android mock contract paths stable.
            </span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Reference: release update format</CardTitle>
          <CardDescription>
            What a release entry should look like when the production update server is online
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Collapsible>
            <CollapsibleTrigger asChild>
              <Button variant="outline" size="sm" className="gap-2">
                <ChevronDown className="h-4 w-4" /> Show example release
              </Button>
            </CollapsibleTrigger>
            <CollapsibleContent className="mt-3 space-y-3">
              <pre className="overflow-x-auto rounded-md border bg-muted/40 p-3 font-mono text-xs">{`{
  "version": "1.4.2",
  "version_code": 142,
  "apk_filename": "vyzorix-audiorouter-1.4.2.apk",
  "apk_sha256": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
  "apk_size_bytes": 18234567,
  "release_notes": "• Fix VoIP route loss after Bluetooth disconnect on Nokia C22\\n• Reduce thermal mitigation false positives at 44°C\\n• Audio HAL recycle now persists ProjectionToken state\\n• Updater wake-up FCM payload now versioned (schema v3)"
}`}</pre>
              <p className="text-xs text-muted-foreground">
                Live values you see above come from the same shape, just fetched from{" "}
                <code>{serverUrl}/api/v1/version</code>.
              </p>
            </CollapsibleContent>
          </Collapsible>
        </CardContent>
      </Card>
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div className="rounded-md border p-3">
      <p className="text-xs text-muted-foreground">{k}</p>
      <p className="text-sm font-medium break-all">{v}</p>
    </div>
  );
}
