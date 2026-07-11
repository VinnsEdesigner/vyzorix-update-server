/**
 * Telemetry Mappers
 * 
 * Transformations from raw API response to domain types.
 * Raw API uses snake_case, domain uses camelCase.
 */

import type {
  TelemetryFrame,
  MetricChartPoint,
  MetricStats,
  MetricWithChart,
  MetricWithThreshold,
  MetricsTimeRange,
  MetricsDevice,
  MetricsCollection,
  DeviceMetrics,
  MetricEvent,
  RawTelemetryFrame,
  RawMetricStats,
  RawChartPoint,
  RiskScoreMetric,
  ThermalMetric,
  BufferLevelMetric,
  UptimeMetric,
} from "./telemetry-entity";

// ============================================================================
// Raw API Types (snake_case)
// ============================================================================

/**
 * Raw metrics time range from API
 */
export interface RawMetricsTimeRange {
  start?: string | number;
  end?: string | number;
  range?: string;
  resolution?: string;
}

/**
 * Raw metrics device from API
 */
export interface RawMetricsDevice {
  imei?: string;
  device_name?: string;
}

/**
 * Raw metric with threshold from API
 */
export interface RawMetricWithThreshold {
  current?: number;
  avg?: number;
  min?: number;
  max?: number;
  unit?: string;
  threshold?: {
    warning?: number;
    critical?: number;
  };
  chart?: RawChartPoint[];
}

/**
 * Raw uptime metric from API
 */
export interface RawUptimeMetric {
  current?: number;
  unit?: string;
}

/**
 * Raw metrics collection from API
 */
export interface RawMetricsCollection {
  risk_score?: RawMetricWithThreshold;
  thermal_temp?: RawMetricWithThreshold;
  buffer_level?: RawMetricWithThreshold;
  uptime?: RawUptimeMetric;
}

/**
 * Raw metric event from API
 */
export interface RawMetricEvent {
  timestamp?: string | number;
  type?: string;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}

/**
 * Raw device metrics from API
 */
export interface RawDeviceMetrics {
  device?: RawMetricsDevice;
  time_range?: RawMetricsTimeRange;
  metrics?: RawMetricsCollection;
  events?: RawMetricEvent[];
}

// ============================================================================
// Transform Helpers
// ============================================================================

/**
 * Parse timestamp from various formats
 */
function parseTimestamp(value?: string | number | null): Date {
  if (!value) return new Date();
  
  if (typeof value === "number") {
    // Unix timestamp in seconds or milliseconds
    return new Date(value > 1e12 ? value : value * 1000);
  }
  
  return new Date(value);
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform raw telemetry frame to domain
 */
export function telemetryFrameFromRaw(raw: RawTelemetryFrame): TelemetryFrame {
  return {
    timestamp: parseTimestamp(raw.timestamp),
    riskScore: raw.risk_score ?? 0,
    thermalTemp: raw.thermal_temp ?? 0,
    bufferLevel: raw.buffer_level ?? 0,
    latencyMs: raw.latency_ms,
  };
}

/**
 * Transform raw chart point to domain
 */
export function chartPointFromRaw(raw: RawChartPoint): MetricChartPoint {
  return {
    timestamp: parseTimestamp(raw.timestamp),
    value: raw.value,
  };
}

/**
 * Transform raw stats to domain
 */
export function statsFromRaw(raw?: RawMetricStats | null): MetricStats | null {
  if (!raw) return null;
  
  return {
    current: raw.current ?? 0,
    avg: raw.avg ?? 0,
    min: raw.min ?? 0,
    max: raw.max ?? 0,
    unit: raw.unit ?? "",
  };
}

/**
 * Transform raw metric with threshold to domain
 */
export function metricWithThresholdFromRaw(raw?: RawMetricWithThreshold | null): MetricWithThreshold | null {
  if (!raw) return null;
  
  return {
    current: raw.current ?? 0,
    avg: raw.avg ?? 0,
    min: raw.min ?? 0,
    max: raw.max ?? 0,
    unit: (raw.unit as "%" | "Â°C") ?? "",
    threshold: {
      warning: raw.threshold?.warning ?? 0,
      critical: raw.threshold?.critical ?? 0,
    },
  };
}

/**
 * Transform raw risk score metric
 */
export function riskScoreFromRaw(raw?: RawMetricWithThreshold | null): RiskScoreMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "%",
  } as RiskScoreMetric;
}

/**
 * Transform raw thermal metric
 */
export function thermalFromRaw(raw?: RawMetricWithThreshold | null): ThermalMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "Â°C",
  } as ThermalMetric;
}

/**
 * Transform raw buffer level metric
 */
export function bufferLevelFromRaw(raw?: RawMetricWithThreshold | null): BufferLevelMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "%",
  } as BufferLevelMetric;
}

/**
 * Transform raw uptime metric
 */
export function uptimeFromRaw(raw?: RawUptimeMetric | null): UptimeMetric | null {
  if (!raw) return null;
  
  return {
    current: raw.current ?? 0,
    unit: raw.unit ?? "",
  };
}

/**
 * Transform raw metrics collection
 * Note: Thresholds should be fetched from the server via /v1/auth/me/settings or /v1/auth/me/thresholds
 * and applied at the presentation layer, not hardcoded here.
 */
export function metricsCollectionFromRaw(raw?: RawMetricsCollection | null): MetricsCollection | null {
  if (!raw) return null;
  
  const riskScore = riskScoreFromRaw(raw.risk_score);
  const thermalTemp = thermalFromRaw(raw.thermal_temp);
  const bufferLevel = bufferLevelFromRaw(raw.buffer_level);
  const uptime = uptimeFromRaw(raw.uptime);
  
  // If we have at least some data, return it with server-provided thresholds or null
  // Thresholds should be provided by the server or fetched separately
  if (!riskScore && !thermalTemp && !bufferLevel && !uptime) return null;
  
  return {
    riskScore: riskScore ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "%",
      threshold: { warning: 0, critical: 0 }, // Thresholds should come from server
      chart: [],
    },
    thermalTemp: thermalTemp ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "Â°C",
      threshold: { warning: 0, critical: 0 }, // Thresholds should come from server
      chart: [],
    },
    bufferLevel: bufferLevel ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "%",
      threshold: { warning: 0, critical: 0 }, // Thresholds should come from server
      chart: [],
    },
    uptime: uptime ?? {
      current: 0,
      unit: "%",
    },
  };
}

/**
 * Transform raw time range
 */
export function timeRangeFromRaw(raw?: RawMetricsTimeRange | null): MetricsTimeRange | null {
  if (!raw) return null;
  
  return {
    start: parseTimestamp(raw.start),
    end: parseTimestamp(raw.end),
    range: raw.range ?? "",
    resolution: raw.resolution ?? "",
  };
}

/**
 * Transform raw metrics device
 */
export function metricsDeviceFromRaw(raw?: RawMetricsDevice | null): MetricsDevice | null {
  if (!raw) return null;
  
  return {
    imei: raw.imei ?? "",
    deviceName: raw.device_name ?? "",
  };
}

/**
 * Transform raw metric event
 */
export function metricEventFromRaw(raw: RawMetricEvent): MetricEvent {
  return {
    timestamp: parseTimestamp(raw.timestamp),
    type: (raw.type as MetricEvent["type"]) ?? "threshold_breach",
    metric: raw.metric,
    value: raw.value,
    threshold: raw.threshold,
    message: raw.message,
  };
}

/**
 * Transform raw device metrics
 * Note: Thresholds should be fetched from the server via /v1/auth/me/settings or /v1/auth/me/thresholds
 * and applied at the presentation layer, not hardcoded here.
 */
export function deviceMetricsFromRaw(raw?: RawDeviceMetrics | null): DeviceMetrics | null {
  if (!raw) return null;
  
  return {
    device: metricsDeviceFromRaw(raw.device) ?? { imei: "", deviceName: "" },
    timeRange: timeRangeFromRaw(raw.time_range) ?? {
      start: new Date(),
      end: new Date(),
      range: "",
      resolution: "",
    },
    metrics: metricsCollectionFromRaw(raw.metrics) ?? {
      riskScore: { current: 0, avg: 0, min: 0, max: 0, unit: "%", threshold: { warning: 0, critical: 0 }, chart: [] },
      thermalTemp: { current: 0, avg: 0, min: 0, max: 0, unit: "Â°C", threshold: { warning: 0, critical: 0 }, chart: [] },
      bufferLevel: { current: 0, avg: 0, min: 0, max: 0, unit: "%", threshold: { warning: 0, critical: 0 }, chart: [] },
      uptime: { current: 0, unit: "%" },
    },
  };
}

// ============================================================================
// Array Transformers
// ============================================================================

/**
 * Transform array of raw telemetry frames
 */
export function telemetryFramesFromRaw(raw: RawTelemetryFrame[]): TelemetryFrame[] {
  return raw.map(telemetryFrameFromRaw);
}

/**
 * Transform array of raw chart points
 */
export function chartPointsFromRaw(raw: RawChartPoint[]): MetricChartPoint[] {
  return raw.map(chartPointFromRaw);
}

/**
 * Transform array of raw metric events
 */
export function metricEventsFromRaw(raw: RawMetricEvent[]): MetricEvent[] {
  return raw.map(metricEventFromRaw);
}
