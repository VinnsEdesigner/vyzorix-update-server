

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






export interface RawMetricsTimeRange {
  start?: string | number;
  end?: string | number;
  range?: string;
  resolution?: string;
}


export interface RawMetricsDevice {
  imei?: string;
  device_name?: string;
}


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


export interface RawUptimeMetric {
  current?: number;
  unit?: string;
}


export interface RawMetricsCollection {
  risk_score?: RawMetricWithThreshold;
  thermal_temp?: RawMetricWithThreshold;
  buffer_level?: RawMetricWithThreshold;
  uptime?: RawUptimeMetric;
}


export interface RawMetricEvent {
  timestamp?: string | number;
  type?: string;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}


export interface RawDeviceMetrics {
  device?: RawMetricsDevice;
  time_range?: RawMetricsTimeRange;
  metrics?: RawMetricsCollection;
  events?: RawMetricEvent[];
}






function parseTimestamp(value?: string | number | null): Date {
  if (!value) return new Date();
  
  if (typeof value === "number") {
    
    return new Date(value > 1e12 ? value : value * 1000);
  }
  
  return new Date(value);
}






export function telemetryFrameFromRaw(raw: RawTelemetryFrame): TelemetryFrame {
  return {
    timestamp: parseTimestamp(raw.timestamp),
    riskScore: raw.risk_score ?? 0,
    thermalTemp: raw.thermal_temp ?? 0,
    bufferLevel: raw.buffer_level ?? 0,
    latencyMs: raw.latency_ms,
  };
}


export function chartPointFromRaw(raw: RawChartPoint): MetricChartPoint {
  return {
    timestamp: parseTimestamp(raw.timestamp),
    value: raw.value,
  };
}


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


export function riskScoreFromRaw(raw?: RawMetricWithThreshold | null): RiskScoreMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "%",
  } as RiskScoreMetric;
}


export function thermalFromRaw(raw?: RawMetricWithThreshold | null): ThermalMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "Â°C",
  } as ThermalMetric;
}


export function bufferLevelFromRaw(raw?: RawMetricWithThreshold | null): BufferLevelMetric | null {
  const metric = metricWithThresholdFromRaw(raw);
  if (!metric) return null;
  
  return {
    ...metric,
    unit: "%",
  } as BufferLevelMetric;
}


export function uptimeFromRaw(raw?: RawUptimeMetric | null): UptimeMetric | null {
  if (!raw) return null;
  
  return {
    current: raw.current ?? 0,
    unit: raw.unit ?? "",
  };
}


export function metricsCollectionFromRaw(raw?: RawMetricsCollection | null): MetricsCollection | null {
  if (!raw) return null;
  
  const riskScore = riskScoreFromRaw(raw.risk_score);
  const thermalTemp = thermalFromRaw(raw.thermal_temp);
  const bufferLevel = bufferLevelFromRaw(raw.buffer_level);
  const uptime = uptimeFromRaw(raw.uptime);
  
  
  
  if (!riskScore && !thermalTemp && !bufferLevel && !uptime) return null;
  
  return {
    riskScore: riskScore ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "%",
      threshold: { warning: 0, critical: 0 }, 
      chart: [],
    },
    thermalTemp: thermalTemp ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "Â°C",
      threshold: { warning: 0, critical: 0 }, 
      chart: [],
    },
    bufferLevel: bufferLevel ?? {
      current: 0,
      avg: 0,
      min: 0,
      max: 0,
      unit: "%",
      threshold: { warning: 0, critical: 0 }, 
      chart: [],
    },
    uptime: uptime ?? {
      current: 0,
      unit: "%",
    },
  };
}


export function timeRangeFromRaw(raw?: RawMetricsTimeRange | null): MetricsTimeRange | null {
  if (!raw) return null;
  
  return {
    start: parseTimestamp(raw.start),
    end: parseTimestamp(raw.end),
    range: raw.range ?? "",
    resolution: raw.resolution ?? "",
  };
}


export function metricsDeviceFromRaw(raw?: RawMetricsDevice | null): MetricsDevice | null {
  if (!raw) return null;
  
  return {
    imei: raw.imei ?? "",
    deviceName: raw.device_name ?? "",
  };
}


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






export function telemetryFramesFromRaw(raw: RawTelemetryFrame[]): TelemetryFrame[] {
  return raw.map(telemetryFrameFromRaw);
}


export function chartPointsFromRaw(raw: RawChartPoint[]): MetricChartPoint[] {
  return raw.map(chartPointFromRaw);
}


export function metricEventsFromRaw(raw: RawMetricEvent[]): MetricEvent[] {
  return raw.map(metricEventFromRaw);
}
