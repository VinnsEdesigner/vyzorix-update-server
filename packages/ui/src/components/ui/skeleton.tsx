import type { ReactElement } from "react";

import { cn } from "@/lib/utils";

/**
 * Skeleton with shimmer animation.
 * Uses bg-tertiary â bg-secondary â bg-tertiary gradient shift.
 */
function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>): ReactElement {
  return (
    <div
      className={cn("skeleton rounded", className)}
      {...props}
    />
  );
}

/**
 * SVG Skeleton Line - for grid lines and strokes
 */
function SkeletonLine({ className, ...props }: React.SVGLineElementAttributes<SVGLineElement>): ReactElement {
  return (
    <line
      className={cn("skeleton-line", className)}
      {...props}
    />
  );
}

/**
 * SVG Skeleton Path - for area fills
 */
function SkeletonPath({ className, ...props }: React.SVGPathElementAttributes<SVGPathElement>): ReactElement {
  return (
    <path
      className={cn("skeleton-path", className)}
      {...props}
    />
  );
}

/**
 * SVG Skeleton Stroke - for line charts
 */
function SkeletonStroke({ className, ...props }: React.SVGPathElementAttributes<SVGPathElement>): ReactElement {
  return (
    <path
      className={cn("skeleton-stroke", className)}
      {...props}
    />
  );
}

/**
 * SVG Skeleton Circle - for data points
 */
function SkeletonCircle({ className, ...props }: React.SVGCircleElementAttributes<SVGCircleElement>): ReactElement {
  return (
    <circle
      className={cn("skeleton-circle", className)}
      {...props}
    />
  );
}

/**
 * SVG Skeleton Rect - for labels
 */
function SkeletonRect({ className, ...props }: React.SVGRectElementAttributes<SVGRectElement>): ReactElement {
  return (
    <rect
      className={cn("skeleton rounded", className)}
      {...props}
    />
  );
}

export { Skeleton, SkeletonLine, SkeletonPath, SkeletonStroke, SkeletonCircle, SkeletonRect };
