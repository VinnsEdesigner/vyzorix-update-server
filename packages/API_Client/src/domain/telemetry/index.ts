/**
 * Telemetry Domain Index
 * 
 * Re-exports all telemetry domain types and mappers.
 */

// Types
export type {
  TelemetryFrame,
  MetricChartPoint,
  MetricStats,
  MetricWithChart,
  MetricThreshold,
  MetricWithThreshold,
  MetricsTimeRange,
  MetricsDevice,
  RiskScoreMetric,
  ThermalMetric,
  BufferLevelMetric,
  UptimeMetric,
  MetricsCollection,
  DeviceMetrics,
  MetricEventType,
  MetricEvent,
  RawTelemetryFrame,
  RawMetricStats,
  RawChartPoint,
} from "./telemetry-entity";

export {
  isRiskWarning,
  isRiskCritical,
  getRiskStatus,
} from "./telemetry-entity";

// Mappers
export type {
  RawMetricsTimeRange,
  RawMetricsDevice,
  RawMetricWithThreshold,
  RawUptimeMetric,
  RawMetricsCollection,
  RawMetricEvent,
  RawDeviceMetrics,
} from "./telemetry-mappers";

export {
  telemetryFrameFromRaw,
  chartPointFromRaw,
  statsFromRaw,
  metricWithThresholdFromRaw,
  riskScoreFromRaw,
  thermalFromRaw,
  bufferLevelFromRaw,
  uptimeFromRaw,
  metricsCollectionFromRaw,
  timeRangeFromRaw,
  metricsDeviceFromRaw,
  metricEventFromRaw,
  deviceMetricsFromRaw,
  telemetryFramesFromRaw,
  chartPointsFromRaw,
  metricEventsFromRaw,
} from "./telemetry-mappers";
