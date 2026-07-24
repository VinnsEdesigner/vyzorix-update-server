






export const DEVICE_UPDATED_SUBSCRIPTION =  `
  subscription OnDeviceUpdated($deviceId: ID!) {
    deviceUpdated(deviceId: $deviceId) {
      id
      imei
      deviceName
      online
      lastSeen
    }
  }
`;


export const TELEMETRY_RECEIVED_SUBSCRIPTION =  `
  subscription OnTelemetryReceived($deviceId: ID!) {
    telemetryReceived(deviceId: $deviceId) {
      timestamp
      riskScore
      thermalTemp
      bufferLevel
      latencyMs
    }
  }
`;


export const COMMAND_STATUS_SUBSCRIPTION =  `
  subscription OnCommandStatusChanged($dispatchId: ID!) {
    commandStatusChanged(dispatchId: $dispatchId) {
      id
      dispatchId
      type
      deviceImei
      status
      result {
        success
        message
        error
      }
      updatedAt
    }
  }
`;


export const DASHBOARD_EVENT_SUBSCRIPTION =  `
  subscription OnDashboardEvent {
    dashboardEvent {
      type
      deviceId
      timestamp
      data
    }
  }
`;