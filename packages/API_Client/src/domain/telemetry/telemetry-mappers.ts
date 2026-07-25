

import type { TelemetryFrame, RawTelemetryFrame, MetricEvent, RawMetricEvent } from "./telemetry-entity";

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

export function telemetryFramesFromRaw(raw: RawTelemetryFrame[]): TelemetryFrame[] {
  return raw.map(telemetryFrameFromRaw);
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

export function metricEventsFromRaw(raw: RawMetricEvent[]): MetricEvent[] {
  return raw.map(metricEventFromRaw);
}
