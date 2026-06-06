import { Card } from "@/components/ui/card";

interface LogDockProps {
  className?: string;
}

export function LogDock({ className }: LogDockProps) {
  return (
    <Card className={className}>
      <div className="p-4 text-muted-foreground text-sm">Log dock coming soon</div>
    </Card>
  );
}
