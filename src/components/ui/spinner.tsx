import { Loader2 } from "lucide-react";
import type { ReactElement } from "react";

import { cn } from "@/lib/utils";

// eslint-disable-next-line func-style
export function Spinner({
  className,
  size = 16,
}: {
  className?: string;
  size?: number;
}): ReactElement {
  return (
    <Loader2
      className={cn("animate-spin text-muted-foreground", className)}
      style={{ width: size, height: size }}
      aria-label="Loading"
    />
  );
}
