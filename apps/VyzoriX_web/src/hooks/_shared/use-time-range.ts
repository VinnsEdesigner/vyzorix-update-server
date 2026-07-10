/**
 * Use Time Range Hook
 * 
 * Hook for time range selection used in metrics and dashboard.
 * Supports presets: 1h, 6h, 24h, 7d
 */

import { useState, useCallback, useMemo } from "react";

// ============================================================================
// Types
// ============================================================================

/**
 * Available time range presets
 */
export const TIME_RANGES = {
  "1h": { label: "1h", value: "1h", hours: 1 },
  "6h": { label: "6h", value: "6h", hours: 6 },
  "24h": { label: "24h", value: "24h", hours: 24 },
  "7d": { label: "7d", value: "7d", hours: 168 },
} as const;

export type TimeRangeKey = keyof typeof TIME_RANGES;
export type TimeRangeValue = (typeof TIME_RANGES)[TimeRangeKey];

/**
 * Time range state
 */
export interface TimeRange {
  range: TimeRangeKey;
  startTime: Date;
  endTime: Date;
}

/**
 * Use time range options
 */
export interface UseTimeRangeOptions {
  initialRange?: TimeRangeKey;
  onRangeChange?: (range: TimeRangeKey) => void;
}

// ============================================================================
// Hook
// ============================================================================

/**
 * Time range selection hook
 */
export function useTimeRange(options: UseTimeRangeOptions = {}) {
  const { initialRange = "6h", onRangeChange } = options;

  // Calculate timestamps for a range
  const calculateRange = useCallback((range: TimeRangeKey): TimeRange => {
    const end = new Date();
    const start = new Date(end.getTime() - TIME_RANGES[range].hours * 60 * 60 * 1000);
    return { range, startTime: start, endTime: end };
  }, []);

  const [timeRange, setTimeRange] = useState<TimeRange>(() =>
    calculateRange(initialRange)
  );

  // Set range
  const setRange = useCallback(
    (range: TimeRangeKey) => {
      setTimeRange(calculateRange(range));
      onRangeChange?.(range);
    },
    [calculateRange, onRangeChange]
  );

  // Refresh to current time
  const refresh = useCallback(() => {
    setTimeRange(calculateRange(timeRange.range));
  }, [timeRange.range, calculateRange]);

  // Get time range for API params
  const getTimeParams = useCallback(
    (range?: TimeRangeKey) => {
      const r = range ?? timeRange.range;
      const end = new Date();
      const start = new Date(
        end.getTime() - TIME_RANGES[r].hours * 60 * 60 * 1000
      );

      return {
        startTime: start.toISOString(),
        endTime: end.toISOString(),
        range: r,
      };
    },
    [timeRange.range]
  );

  // Get timestamps in milliseconds
  const getTimestamps = useCallback(
    (range?: TimeRangeKey) => {
      const r = range ?? timeRange.range;
      const end = Date.now();
      const start = end - TIME_RANGES[r].hours * 60 * 60 * 1000;
      return { start, end };
    },
    [timeRange.range]
  );

  return {
    // State
    range: timeRange.range,
    startTime: timeRange.startTime,
    endTime: timeRange.endTime,

    // Setters
    setRange,

    // Helpers
    refresh,
    getTimeParams,
    getTimestamps,

    // Presets
    ranges: TIME_RANGES,
  };
}

// ============================================================================
// Resolution Calculator
// ============================================================================

/**
 * Calculate appropriate resolution for a time range
 */
export function calculateResolution(
  range: TimeRangeKey
): { resolution: string; maxPoints: number } {
  switch (range) {
    case "1h":
      return { resolution: "1m", maxPoints: 60 };
    case "6h":
      return { resolution: "5m", maxPoints: 72 };
    case "24h":
      return { resolution: "15m", maxPoints: 96 };
    case "7d":
      return { resolution: "1h", maxPoints: 168 };
    default:
      return { resolution: "5m", maxPoints: 72 };
  }
}

/**
 * Hook for resolution based on time range
 */
export function useResolution(range: TimeRangeKey) {
  return useMemo(() => calculateResolution(range), [range]);
}

// ============================================================================
// URL Sync Helper
// ============================================================================

/**
 * Parse time range from URL
 */
export function parseTimeRangeFromURL(searchParams: URLSearchParams): TimeRangeKey {
  const range = searchParams.get("range");
  if (range && range in TIME_RANGES) {
    return range as TimeRangeKey;
  }
  return "6h";
}

/**
 * Update time range in URL
 */
export function updateTimeRangeURL(range: TimeRangeKey): URLSearchParams {
  const params = new URLSearchParams();
  if (range !== "6h") {
    params.set("range", range);
  }
  return params;
}

// ============================================================================
// Formatters
// ============================================================================

/**
 * Format time range for display
 */
export function formatTimeRange(range: TimeRangeKey): string {
  return TIME_RANGES[range].label;
}

/**
 * Format timestamp for API
 */
export function formatTimestamp(date: Date): string {
  return date.toISOString();
}