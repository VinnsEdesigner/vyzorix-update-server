/**
 * Telemetry Domain Types
 * 
 * Domain types for device telemetry and metrics.
 * Pure TypeScript - no external imports.
 */

// ============================================================================
// Telemetry Frame
// ============================================================================

/**
 * Single telemetry data point from device
 */
export interface TelemetryFrame {
  timestamp: Date;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs?: number;
}

// ============================================================================
// Metrics Data
// ============================================================================

/**
 * Metric chart data point
 */
export interface MetricChartPoint {
  timestamp: Date;
  value: number;
}

/**
 * Aggregated metric statistics
 */
export interface MetricStats {
  current: number;
  avg: number;
  min: number;
  max: number;
  unit: string;
}

/**
 * Metric with chart data
 */
export interface MetricWithChart extends MetricStats {
  chart: MetricChartPoint[];
}

/**
 * Metric thresholds
 */
export interface MetricThreshold {
  warning: number;
  critical: number;
}

/**
 * Metric with threshold
 */
export interface MetricWithThreshold extends MetricStats {
  threshold: MetricThreshold;
}

// ============================================================================
// Metrics Collection
// ============================================================================

/**
 * Time range for metrics
 */
export interface MetricsTimeRange {
  start: Date;
  end: Date;
  range: string;
  resolution: string;
}

/**
 * Device info in metrics response
 */
export interface MetricsDevice {
  imei: string;
  deviceName: string;
}

/**
 * Risk score metric
 */
export interface RiskScoreMetric extends MetricWithThreshold {
  unit: "%";
}

/**
 * Thermal temperature metric
 */
export interface ThermalMetric extends MetricWithThreshold {
  unit: "Â°C";
}

/**
 * Buffer level metric
 */
export interface BufferLevelMetric extends MetricWithThreshold {
  unit: "%";
}

/**
 * Uptime metric
 */
export interface UptimeMetric {
  current: number;
  unit: string;
}

/**
 * Metrics collection for a device
 */
export interface MetricsCollection {
  riskScore: RiskScoreMetric;
  thermalTemp: ThermalMetric;
  bufferLevel: BufferLevelMetric;
  uptime: UptimeMetric;
}

/**
 * Complete device metrics response
 */
export interface DeviceMetrics {
  device: MetricsDevice;
  timeRange: MetricsTimeRange;
  metrics: MetricsCollection;
}

// ============================================================================
// Metric Events
// ============================================================================

/**
 * Metric event types
 */
export type MetricEventType = 
  | "threshold_breach"
  | "device_offline"
  | "device_online"
  | "command_sent"
  | "command_failed";

/**
 * Metric event
 */
export interface MetricEvent {
  timestamp: Date;
  type: MetricEventType;
  metric?: string;
  value?: number;
  threshold?: number;
  message?: string;
}

// ============================================================================
// Raw Telemetry
// ============================================================================

/**
 * Raw telemetry frame from API
 */
export interface RawTelemetryFrame {
  timestamp: string | number;
  risk_score?: number;
  thermal_temp?: number;
  buffer_level?: number;
  latency_ms?: number;
}

/**
 * Raw stats from API
 */
export interface RawMetricStats {
  current?: number;
  avg?: number;
  min?: number;
  max?: number;
  unit?: string;
}

/**
 * Raw chart point from API
 */
export interface RawChartPoint {
  timestamp: string | number;
  value: number;
}

/**
 * Helper functions
 */

/**
 * Check if risk is in warning zone
 */
export function isRiskWarning(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.warning && value < threshold.critical;
}

/**
 * Check if risk is critical
 */
export function isRiskCritical(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.critical;
}

/**
 * Get risk status
 */
export function getRiskStatus(value: number, threshold: MetricThreshold): "normal" | "warning" | "critical" {
  if (isRiskCritical(value, threshold)) return "critical";
  if (isRiskWarning(value, threshold)) return "warning";
  return "normal";
}
