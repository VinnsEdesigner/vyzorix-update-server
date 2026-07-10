import type { ReactElement } from "react";

import { cn } from "@/lib/utils";

/**
 * Mini block spinner - rotating outer ring only.
 * Standard loading indicator for all UI components.
 */
export function Spinner({
  className,
  size = 16,
}: {
  className?: string;
  size?: number;
}): ReactElement {
  const borderWidth = Math.max(2, Math.round(size / 8));

  return (
    <div
      className={cn("relative inline-block", className)}
      style={{ width: size, height: size }}
      role="status"
      aria-label="Loading"
    >
      <style>{`
        @keyframes block-spin {
          0% { transform: rotate(0deg); }
          100% { transform: rotate(360deg); }
        }
        .animate-block-spin {
          animation: block-spin 2s linear infinite;
        }
      `}</style>
      {/* Outer ring - rotating */}
      <div
        className="absolute inset-0 border-white animate-block-spin rounded-[3px]"
        style={{ borderWidth }}
      />
    </div>
  );
}