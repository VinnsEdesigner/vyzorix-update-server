/**
 * Telemetry Frame Fragment
 * 
 * Reusable GraphQL fragment for TelemetryFrame type.
 */

export const TELEMETRY_FRAME_FRAGMENT = /* GraphQL */ `
  fragment TelemetryFrame on TelemetryFrame {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
    uptime
  }
`;
