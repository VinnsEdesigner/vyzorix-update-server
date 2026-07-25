export const TELEMETRY_FRAME_FRAGMENT = `
  fragment TelemetryFrame on TelemetryEntry {
    id
    deviceId
    receivedAt
    riskScore
    bufferLevel
    thermalTemp
    payload
  }
`;
