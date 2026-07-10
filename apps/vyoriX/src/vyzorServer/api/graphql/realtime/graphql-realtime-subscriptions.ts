/**
 * Realtime GraphQL Subscriptions
 * 
 * GraphQL subscription definitions for real-time updates.
 * Based on REALTIME_WEBSOCKET_ARCHITECTURE.md.
 */

// ============================================================================
// Subscription Definitions
// ============================================================================

/**
 * Subscribe to device update events
 */
export const DEVICE_UPDATED_SUBSCRIPTION = /* GraphQL */ `
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

/**
 * Subscribe to real-time telemetry
 */
export const TELEMETRY_RECEIVED_SUBSCRIPTION = /* GraphQL */ `
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

/**
 * Subscribe to command status changes
 */
export const COMMAND_STATUS_SUBSCRIPTION = /* GraphQL */ `
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

/**
 * Subscribe to dashboard events (connection, alerts, etc.)
 */
export const DASHBOARD_EVENT_SUBSCRIPTION = /* GraphQL */ `
  subscription OnDashboardEvent {
    dashboardEvent {
      type
      deviceId
      timestamp
      data
    }
  }
`;