import type { ReactElement } from "react";

import { Skeleton } from "@/components/ui/skeleton";

export function PageSkeleton({ rows = 3 }: { rows?: number }): ReactElement {
  return (
    <div className="space-y-4">
      <Skeleton className="h-24 w-full rounded-lg" />
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: rows * 3 }).map((_, i) => (
          <Skeleton key={i} className="h-28 rounded-lg" />
        ))}
      </div>
    </div>
  );
}

export function MetricSkeleton(): ReactElement {
  return (
    <div className="rounded-md border p-3">
      <Skeleton className="h-3 w-16" />
      <Skeleton className="mt-2 h-6 w-20" />
      <Skeleton className="mt-1 h-3 w-24" />
    </div>
  );
}
