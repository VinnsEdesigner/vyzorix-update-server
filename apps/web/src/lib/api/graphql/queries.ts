// GraphQL Queries for Vyzorix API
// These mirror the REST endpoints but allow fetching related data in a single request

import { gql } from "graphql-request";

// ============================================================
// FRAGMENTS - Reusable field selections
// ============================================================

export const DEVICE_FRAGMENT = gql`
  fragment DeviceFields on Device {
    id
    name
    online
    lastSeen
    fcmToken
    version
    deviceClass
    createdAt
  }
`;

export const COMMAND_FRAGMENT = gql`
  fragment CommandFields on Command {
    dispatchId
    commandId
    deviceId
    command
    args
    status
    createdAt
    deliveredAt
  }
`;

export const TELEMETRY_STATS_FRAGMENT = gql`
  fragment TelemetryStatsFields on TelemetryStats {
    riskScore {
      avg
      min
      max
    }
    bufferLevel {
      avg
      min
      max
    }
    thermalTemp {
      avg
      min
      max
    }
  }
`;

export const CONNECTION_STATUS_FRAGMENT = gql`
  fragment ConnectionStatusFields on ConnectionStatus {
    deviceId
    connected
    connectedAt
    lastMessageAt
    uptimeSeconds
  }
`;

// ============================================================
// DEVICE QUERIES
// ============================================================

export const GET_DEVICE = gql`
  query GetDevice($id: ID!) {
    device(id: $id) {
      id
      name
      online
      lastSeen
      fcmToken
      version
      deviceClass
      createdAt
    }
  }
`;

export const GET_DEVICES = gql`
  query GetDevices($limit: Int, $offset: Int) {
    devices(limit: $limit, offset: $offset) {
      id
      name
      online
      lastSeen
      version
      deviceClass
    }
  }
`;

export const GET_DEVICE_COUNT = gql`
  query GetDeviceCount {
    deviceCount
  }
`;

export const GET_DEVICES_WITH_STATS = gql`
  query GetDevicesWithStats($limit: Int, $offset: Int) {
    devices(limit: $limit, offset: $offset) {
      id
      name
      online
      lastSeen
      version
    }
    deviceCount
  }
`;

// ============================================================
// COMMAND QUERIES
// ============================================================

export const GET_COMMAND = gql`
  query GetCommand($dispatchId: ID!) {
    command(dispatchId: $dispatchId) {
      dispatchId
      commandId
      deviceId
      command
      args
      status
      createdAt
      deliveredAt
    }
  }
`;

export const GET_PENDING_COMMANDS = gql`
  query GetPendingCommands($deviceId: ID!) {
    pendingCommands(deviceId: $deviceId) {
      dispatchId
      commandId
      command
      status
      createdAt
    }
  }
`;

// ============================================================
// TELEMETRY QUERIES
// ============================================================

export const GET_TELEMETRY_HISTORY = gql`
  query GetTelemetryHistory($deviceId: ID!, $startTime: Int, $endTime: Int, $limit: Int) {
    telemetryHistory(deviceId: $deviceId, startTime: $startTime, endTime: $endTime, limit: $limit) {
      id
      deviceId
      receivedAt
      riskScore
      bufferLevel
      thermalTemp
      payload
    }
  }
`;

export const GET_LATEST_TELEMETRY = gql`
  query GetLatestTelemetry($deviceId: ID!) {
    latestTelemetry(deviceId: $deviceId) {
      id
      deviceId
      receivedAt
      riskScore
      bufferLevel
      thermalTemp
      payload
    }
  }
`;

export const GET_TELEMETRY_STATS = gql`
  query GetTelemetryStats($deviceId: ID!) {
    telemetryStats(deviceId: $deviceId) {
      riskScore {
        avg
        min
        max
      }
      bufferLevel {
        avg
        min
        max
      }
      thermalTemp {
        avg
        min
        max
      }
    }
  }
`;

// ============================================================
// CONNECTION QUERIES
// ============================================================

export const GET_CONNECTION_STATUS = gql`
  query GetConnectionStatus($deviceId: ID!) {
    connectionStatus(deviceId: $deviceId) {
      deviceId
      connected
      connectedAt
      lastMessageAt
      uptimeSeconds
    }
  }
`;

export const GET_ALL_CONNECTIONS = gql`
  query GetAllConnections {
    allConnections {
      deviceId
      connected
      connectedAt
      lastMessageAt
      uptimeSeconds
    }
  }
`;

// ============================================================
// HEALTH QUERY
// ============================================================

export const GET_HEALTH = gql`
  query GetHealth {
    health {
      ok
      serverTime
      connectedDevices
      version
    }
  }
`;

// ============================================================
// DASHBOARD QUERY - Get all dashboard data in one request!
// ============================================================

export const GET_DASHBOARD_DATA = gql`
  query GetDashboardData($deviceLimit: Int, $connectionLimit: Int) {
    # Get devices with basic info
    devices(limit: $deviceLimit) {
      id
      name
      online
      lastSeen
      version
      deviceClass
    }
    deviceCount

    # Get all connection statuses
    allConnections {
      deviceId
      connected
      connectedAt
      lastMessageAt
      uptimeSeconds
    }
  }
`;

// ============================================================
// DEVICE DETAIL QUERY - Get device with telemetry stats
// ============================================================

export const GET_DEVICE_DETAIL = gql`
  query GetDeviceDetail($deviceId: ID!) {
    device(id: $deviceId) {
      id
      name
      online
      lastSeen
      fcmToken
      version
      deviceClass
      createdAt
    }
    connectionStatus(deviceId: $deviceId) {
      deviceId
      connected
      connectedAt
      lastMessageAt
      uptimeSeconds
    }
    telemetryStats(deviceId: $deviceId) {
      riskScore {
        avg
        min
        max
      }
      bufferLevel {
        avg
        min
        max
      }
      thermalTemp {
        avg
        min
        max
      }
    }
    pendingCommands(deviceId: $deviceId) {
      dispatchId
      commandId
      command
      status
      createdAt
    }
  }
`;
