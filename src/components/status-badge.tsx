import { Badge } from "@/components/ui/badge";
import type { DeviceStatus } from "@/lib/mock-data";

export function StatusBadge({ status }: { status: DeviceStatus }) {
  const map: Record<DeviceStatus, { label: string; variant: "default" | "secondary" | "destructive" | "outline"; dot: string }> = {
    online: { label: "Online", variant: "default", dot: "bg-primary" },
    warning: { label: "Warning", variant: "secondary", dot: "bg-yellow-500" },
    critical: { label: "Critical", variant: "destructive", dot: "bg-destructive" },
    offline: { label: "Offline", variant: "outline", dot: "bg-muted-foreground" },
  };
  const cfg = map[status];
  return (
    <Badge variant={cfg.variant} className="gap-1.5">
      <span className={`h-1.5 w-1.5 rounded-full ${cfg.dot}`} />
      {cfg.label}
    </Badge>
  );
}