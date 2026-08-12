import type { MetricThreshold } from "../_shared";


// Raw telemetry frame from device - real-time sensor data
export interface TelemetryFrame {
  timestamp: Date;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs?: number;
}

// Raw format from REST API (snake_case)
export interface RawTelemetryFrame {
  timestamp: string | number;
  risk_score?: number;
  thermal_temp?: number;
  buffer_level?: number;
  latency_ms?: number;
}

// Metric event types from the metrics API (snake_case for consistency with server)
export type MetricEventType = 
  | "threshold_breach"
  | "device_offline"
  | "device_online"
  | "command_sent"
  | "command_failed";

// Metric event from DeviceMetrics response
export interface MetricEvent {
  timestamp: Date;
  type: MetricEventType;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}

// Raw metric event from REST API
export interface RawMetricEvent {
  timestamp?: string | number;
  type?: string;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}

// Threshold helpers for risk assessment
export type { MetricThreshold } from "../_shared";

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
