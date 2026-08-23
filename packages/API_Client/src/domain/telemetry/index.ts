// Telemetry domain — generated types + hand-rolled risk assessment algorithms.
// The threshold comparison functions (isRiskWarning, isRiskCritical, getRiskStatus)
// are genuine domain logic not expressible in OpenAPI.
import type {
  TelemetryEntry,
  TelemetryHistoryQueryResult,
  TelemetryHistoryEntry,
  TelemetryStatsResult,
  MetricAggregateResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';
import type { MetricThreshold } from '../_shared';

export type {
  TelemetryEntry,
  TelemetryHistoryQueryResult,
  TelemetryHistoryEntry,
  TelemetryStatsResult,
  MetricAggregateResult,
};
export type { MetricThreshold };

// ---- Domain types (not in OpenAPI, used by hooks/stores) ----

export interface TelemetryFrame {
  timestamp: Date;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  latencyMs?: number;
}

// ---- Risk assessment algorithms (hand-rolled) ----

export function isRiskWarning(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.warning && value < threshold.critical;
}

export function isRiskCritical(value: number, threshold: MetricThreshold): boolean {
  return value >= threshold.critical;
}

export function getRiskStatus(value: number, threshold: MetricThreshold): 'normal' | 'warning' | 'critical' {
  if (isRiskCritical(value, threshold)) return 'critical';
  if (isRiskWarning(value, threshold)) return 'warning';
  return 'normal';
}
