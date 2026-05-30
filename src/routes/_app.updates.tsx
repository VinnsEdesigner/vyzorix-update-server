import { createFileRoute } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Zap, Download } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { useVyzorixConfig } from "@/lib/vyzorix-config";
import { dispatchCommand, getVersion, headApk } from "@/lib/vyzorix-api";
import { formatBytes, shortHash } from "@/lib/format";

export const Route = createFileRoute("/_app/updates")({
  head: () => ({ meta: [{ title: "Updates — Vyzorix" }] }),
  component: UpdatesPage,
});

function UpdatesPage() {
  const { serverUrl, deviceId } = useVyzorixConfig();

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
    try {
      const res = await dispatchCommand(serverUrl, deviceId, "WAKE_UP_UPDATER");
      toast.success(`WAKE_UP_UPDATER → ${res.delivery}`, { description: `dispatch ${res.dispatchId}` });
    } catch (e) {
      toast.error("WAKE_UP_UPDATER failed", { description: e instanceof Error ? e.message : String(e) });
    }
  };

  return (
    <div className="space-y-4">
      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Current release</CardTitle>
            <CardDescription>
              Served from <code className="text-xs">{serverUrl}/api/v1/version</code>
            </CardDescription>
          </CardHeader>
          <CardContent>
            {version.isError ? (
              <p className="text-sm text-muted-foreground">Failed to load version.json — {(version.error as Error).message}</p>
            ) : version.isLoading || !v ? (
              <p className="text-sm text-muted-foreground">Loading…</p>
            ) : (
              <div className="grid gap-3 sm:grid-cols-3">
                <KV k="Version" v={v.version} />
                <KV k="Version code" v={`${v.version_code}`} />
                <KV k="APK file" v={v.apk_filename} />
                <KV k="APK size (manifest)" v={formatBytes(v.apk_size_bytes)} />
                <KV k="APK size (HEAD)" v={apkSize.data == null ? "—" : formatBytes(apkSize.data)} />
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
        </CardHeader>
        <CardContent>
          {v ? (
            <p className="text-sm text-muted-foreground whitespace-pre-wrap">{v.release_notes}</p>
          ) : (
            <p className="text-sm text-muted-foreground">—</p>
          )}
          <div className="mt-4 flex items-center gap-2">
            <Badge variant="outline">phase 1</Badge>
            <Badge variant="outline">mock server</Badge>
            <span className="text-xs text-muted-foreground">Per ADR-0009. Real server lands in Phase 1.5 with no Android changes.</span>
          </div>
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
