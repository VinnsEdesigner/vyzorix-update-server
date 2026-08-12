// Cross-context domain primitives shared by multiple bounded contexts.
// Centralizing these here removes per-context redefinitions that collided
// in the domain barrel; each context now imports from _shared instead.

/** Device connectivity state, owned by no single context. */
export type DeviceStatus = "online" | "offline" | "deregistered";

/** Warning/critical band shared by metrics + telemetry contexts. */
export interface MetricThreshold {
  warning: number;
  critical: number;
}
