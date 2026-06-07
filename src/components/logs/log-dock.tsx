import { Card } from "@/components/ui/card";

interface LogDockProps {
  className?: string;
}

// eslint-disable-next-line func-style
export function LogDock({ className }: LogDockProps): JSX.Element {
  return (
    <Card className={className}>
      <div className="p-4 text-muted-foreground text-sm">Log dock coming soon</div>
    </Card>
  );
}
