






export interface TelemetryFrame {
  timestamp: Date;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs?: number;
}






export interface MetricChartPoint {
  timestamp: Date;
  value: number;
}


export interface MetricStats {
  current: number;
  avg: number;
  min: number;
  max: number;
  unit: string;
}


export interface MetricWithChart extends MetricStats {
  chart: MetricChartPoint[];
}


export interface MetricThreshold {
  warning: number;
  critical: number;
}


export interface MetricWithThreshold extends MetricStats {
  threshold: MetricThreshold;
}






export interface MetricsTimeRange {
  start: Date;
  end: Date;
  range: string;
  resolution: string;
}


export interface MetricsDevice {
  imei: string;
  deviceName: string;
}


export interface RiskScoreMetric extends MetricWithThreshold {
  unit: "%";
}


export interface ThermalMetric extends MetricWithThreshold {
  unit: "Â°C";
}


export interface BufferLevelMetric extends MetricWithThreshold {
  unit: "%";
}


export interface UptimeMetric {
  current: number;
  unit: string;
}


export interface MetricsCollection {
  riskScore: RiskScoreMetric;
  thermalTemp: ThermalMetric;
  bufferLevel: BufferLevelMetric;
  uptime: UptimeMetric;
}


export interface DeviceMetrics {
  device: MetricsDevice;
  timeRange: MetricsTimeRange;
  metrics: MetricsCollection;
}






export type MetricEventType = 
  | "threshold_breach"
  | "device_offline"
  | "device_online"
  | "command_sent"
  | "command_failed";


export interface MetricEvent {
  timestamp: Date;
  type: MetricEventType;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}






export interface RawTelemetryFrame {
  timestamp: string | number;
  risk_score?: number;
  thermal_temp?: number;
  buffer_level?: number;
  latency_ms?: number;
}


export interface RawMetricStats {
  current?: number;
  avg?: number;
  min?: number;
  max?: number;
  unit?: string;
}


export interface RawChartPoint {
  timestamp: string | number;
  value: number;
}




export function isRiskWarning(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.warning && value < threshold.critical;
}


export function isRiskCritical(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.critical;
}


export function getRiskStatus(value: number, threshold: MetricThreshold): "normal" | "warning" | "critical" {
  if (isRiskCritical(value, threshold)) return "critical";
  if (isRiskWarning(value, threshold)) return "warning";
  return "normal";
}
