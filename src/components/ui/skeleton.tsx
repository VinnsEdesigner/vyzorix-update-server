import type { ReactElement } from "react";

import { cn } from "@/lib/utils";

// eslint-disable-next-line func-style
function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>): ReactElement {
  return <div className={cn("animate-pulse rounded-md bg-primary/10", className)} {...props} />;
}

export { Skeleton };
