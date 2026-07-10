import type { ReactElement } from "react";

import { cn } from "@/lib/utils";

/**
 * Metrics Skeleton - shimmer placeholder for metrics page.
 * Uses muted â secondary â muted gradient shift.
 */
function MetricsSkeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>): ReactElement {
  return (
    <div
      className={cn("skeleton rounded", className)}
      {...props}
    />
  );
}

/**
 * Metrics SVG Skeleton Line - for grid lines
 */
function MetricsSkeletonLine({ className, ...props }: React.SVGLineElementAttributes<SVGLineElement>): ReactElement {
  return (
    <line
      className={cn("skeleton-line", className)}
      {...props}
    />
  );
}

/**
 * Metrics SVG Skeleton Path - for area fills
 */
function MetricsSkeletonPath({ className, ...props }: React.SVGPathElementAttributes<SVGPathElement>): ReactElement {
  return (
    <path
      className={cn("skeleton-path", className)}
      {...props}
    />
  );
}

/**
 * Metrics SVG Skeleton Stroke - for line charts
 */
function MetricsSkeletonStroke({ className, ...props }: React.SVGPathElementAttributes<SVGPathElement>): ReactElement {
  return (
    <path
      className={cn("skeleton-stroke", className)}
      {...props}
    />
  );
}

/**
 * Metrics SVG Skeleton Circle - for data points
 */
function MetricsSkeletonCircle({ className, ...props }: React.SVGCircleElementAttributes<SVGCircleElement>): ReactElement {
  return (
    <circle
      className={cn("skeleton-circle", className)}
      {...props}
    />
  );
}

/**
 * Metrics SVG Skeleton Rect - for labels
 */
function MetricsSkeletonRect({ className, ...props }: React.SVGRectElementAttributes<SVGRectElement>): ReactElement {
  return (
    <rect
      className={cn("skeleton rounded", className)}
      {...props}
    />
  );
}

export { 
  MetricsSkeleton, 
  MetricsSkeletonLine, 
  MetricsSkeletonPath, 
  MetricsSkeletonStroke, 
  MetricsSkeletonCircle, 
  MetricsSkeletonRect 
};