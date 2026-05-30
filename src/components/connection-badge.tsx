import { Badge } from "@/components/ui/badge";

export type ConnectionState =
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "idle";

export function ConnectionBadge({ state = "idle" }: { state?: ConnectionState }) {
  const map: Record<ConnectionState, { label: string; variant: "default" | "secondary" | "destructive" | "outline"; dot: string }> = {
    connected: { label: "WS · connected", variant: "default", dot: "bg-primary" },
    connecting: { label: "WS · connecting", variant: "secondary", dot: "bg-yellow-500" },
    reconnecting: { label: "WS · reconnecting", variant: "secondary", dot: "bg-yellow-500" },
    disconnected: { label: "WS · disconnected", variant: "destructive", dot: "bg-destructive" },
    idle: { label: "WS · idle", variant: "outline", dot: "bg-muted-foreground" },
  };
  const cfg = map[state];
  return (
    <Badge variant={cfg.variant} className="gap-1.5">
      <span className={`h-1.5 w-1.5 rounded-full ${cfg.dot} ${state === "connected" ? "animate-pulse" : ""}`} />
      {cfg.label}
    </Badge>
  );
}