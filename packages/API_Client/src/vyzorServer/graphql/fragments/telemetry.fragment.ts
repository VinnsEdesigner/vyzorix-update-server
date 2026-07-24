export const TELEMETRY_FRAME_FRAGMENT = `
  fragment TelemetryFrame on TelemetryEntry {
    timestamp
    risk_score
    thermal_temp
    buffer_level
    uptime
    battery_voltage
    signal_strength
    speed
    distance
  }
`;
