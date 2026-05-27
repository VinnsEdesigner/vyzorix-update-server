import { Badge } from "@/components/ui/badge";

export function ConnectionBadge({
  state = "connected",
}: {
  state?: "connected" | "reconnecting" | "disconnected";
}) {
  const map = {
    connected: { label: "WebSocket · connected", variant: "default" as const, dot: "bg-primary" },
    reconnecting: { label: "WebSocket · reconnecting", variant: "secondary" as const, dot: "bg-yellow-500" },
    disconnected: { label: "WebSocket · disconnected", variant: "destructive" as const, dot: "bg-destructive" },
  };
  const cfg = map[state];
  return (
    <Badge variant={cfg.variant} className="gap-1.5">
      <span className={`h-1.5 w-1.5 rounded-full ${cfg.dot} ${state === "connected" ? "animate-pulse" : ""}`} />
      {cfg.label}
    </Badge>
  );
}