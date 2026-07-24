export const DEVICE_UPDATED_SUBSCRIPTION = `
  subscription OnDeviceUpdated($organizationId: ID!, $deviceId: ID!) {
    deviceUpdated(organizationId: $organizationId, deviceId: $deviceId) {
      id
      imei
      device_name
      status
      last_seen
    }
  }
`;

export const TELEMETRY_RECEIVED_SUBSCRIPTION = `
  subscription OnTelemetryReceived($organizationId: ID!, $deviceId: ID!) {
    telemetryReceived(organizationId: $organizationId, deviceId: $deviceId) {
      timestamp
      risk_score
      thermal_temp
      buffer_level
      latency_ms
    }
  }
`;

export const COMMAND_STATUS_SUBSCRIPTION = `
  subscription OnCommandStatusChanged($organizationId: ID!, $dispatchId: ID!) {
    commandStatusChanged(organizationId: $organizationId, dispatchId: $dispatchId) {
      id
      device_id
      command_type
      status
      result {
        success
        message
        error
      }
      updated_at
    }
  }
`;
