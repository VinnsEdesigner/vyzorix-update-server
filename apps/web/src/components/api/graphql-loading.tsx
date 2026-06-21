import { Loader2, Sparkles } from "lucide-react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";

/**
 * Skeleton for dashboard device cards.
 */
export function DeviceCardSkeleton({ className }: { className?: string }) {
  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <Skeleton className="h-5 w-32" />
          <Skeleton className="h-5 w-16 rounded-full" />
        </div>
        <CardDescription>
          <Skeleton className="mt-1 h-3 w-24" />
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-2">
          <div className="flex justify-between text-sm">
            <Skeleton className="h-3 w-20" />
            <Skeleton className="h-3 w-12" />
          </div>
          <Skeleton className="h-2 w-full rounded-full" />
          <div className="flex justify-between text-xs text-muted-foreground">
            <Skeleton className="h-3 w-16" />
            <Skeleton className="h-3 w-20" />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

/**
 * Skeleton for device detail page.
 */
export function DeviceDetailSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("space-y-4", className)}>
      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <Skeleton className="h-5 w-24 mb-2" />
            <Skeleton className="h-3 w-32" />
          </CardHeader>
          <CardContent className="space-y-3">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </CardContent>
        </Card>
        <Card className="lg:col-span-2">
          <CardHeader>
            <Skeleton className="h-5 w-32 mb-2" />
            <Skeleton className="h-3 w-48" />
          </CardHeader>
          <CardContent>
            <div className="grid gap-3 sm:grid-cols-2">
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          </CardContent>
        </Card>
      </div>
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-40 mb-2" />
          <Skeleton className="h-3 w-56" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-48 w-full" />
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Skeleton for dashboard overview.
 */
export function DashboardOverviewSkeleton({ className }: { className?: string }) {
  return (
    <div className={cn("space-y-4", className)}>
      <div className="grid gap-4 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Card key={i}>
            <CardHeader className="pb-2">
              <Skeleton className="h-4 w-24 mb-2" />
              <Skeleton className="h-8 w-16" />
            </CardHeader>
          </Card>
        ))}
      </div>
      <Card>
        <CardHeader>
          <Skeleton className="h-5 w-32 mb-2" />
          <Skeleton className="h-3 w-48" />
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <DeviceCardSkeleton key={i} />
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

/**
 * Spinner with optional label.
 */
export function GraphQLSpinner({
  label,
  className,
}: {
  label?: string;
  className?: string;
}) {
  return (
    <div className={cn("flex items-center gap-2 text-muted-foreground", className)}>
      <Loader2 className="h-4 w-4 animate-spin" />
      {label && <span className="text-sm">{label}</span>}
    </div>
  );
}

/**
 * Full-page loading state for GraphQL queries.
 */
export function GraphQLPageLoading({
  title = "Loading...",
  description,
}: {
  title?: string;
  description?: string;
}) {
  return (
    <div className="flex h-[50vh] flex-col items-center justify-center gap-4">
      <div className="relative">
        <Loader2 className="h-12 w-12 animate-spin text-primary" />
        <Sparkles className="absolute -right-1 -top-1 h-4 w-4 text-primary" />
      </div>
      <div className="text-center">
        <p className="text-lg font-medium">{title}</p>
        {description && (
          <p className="text-sm text-muted-foreground">{description}</p>
        )}
      </div>
    </div>
  );
}

/**
 * Inline loading indicator for small spaces.
 */
export function GraphQLInlineLoading({ className }: { className?: string }) {
  return (
    <span className={cn("inline-flex items-center gap-1 text-xs text-muted-foreground", className)}>
      <Loader2 className="h-3 w-3 animate-spin" />
      <span>Loading...</span>
    </span>
  );
}

/**
 * Empty state for when GraphQL returns no data.
 */
export function GraphQLEmptyState({
  title = "No Data",
  description = "No data available for this query.",
  action,
  icon: Icon = Sparkles,
}: {
  title?: string;
  description?: string;
  action?: React.ReactNode;
  icon?: React.ComponentType<{ className?: string }>;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="rounded-full bg-muted p-4">
        <Icon className="h-8 w-8 text-muted-foreground" />
      </div>
      <h3 className="mt-4 text-lg font-medium">{title}</h3>
      <p className="mt-1 text-sm text-muted-foreground">{description}</p>
      {action && <div className="mt-4">{action}</div>}
    </div>
  );
}

/**
 * Combined GraphQL query state component.
 * Shows appropriate UI based on query state (loading, error, empty, data).
 */
export function GraphQLQueryState<TData>({
  isLoading,
  isError,
  error,
  data,
  isEmpty,
  loadingComponent,
  errorComponent,
  emptyComponent,
  children,
  className,
}: {
  isLoading?: boolean;
  isError?: boolean;
  error?: Error | null;
  data?: TData | null;
  isEmpty?: boolean;
  loadingComponent?: React.ReactNode;
  errorComponent?: React.ReactNode;
  emptyComponent?: React.ReactNode;
  children?: React.ReactNode;
  className?: string;
}) {
  if (isLoading) {
    return loadingComponent ? (
      <>{loadingComponent}</>
    ) : (
      <GraphQLSpinner label="Loading GraphQL data..." className={className} />
    );
  }

  if (isError) {
    return errorComponent ? (
      <>{errorComponent}</>
    ) : (
      <div className={className}>
        <p className="text-sm text-destructive">
          Error: {error?.message || "Failed to load data"}
        </p>
      </div>
    );
  }

  if (isEmpty || !data) {
    return emptyComponent ? (
      <>{emptyComponent}</>
    ) : (
      <GraphQLEmptyState className={className} />
    );
  }

  return <div className={className}>{children}</div>;
}
