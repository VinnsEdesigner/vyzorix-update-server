/**
 * Registration GraphQL Queries
 * 
 * GraphQL query definitions for device registration/inbox.
 * Based on DEVICE_REGISTRATION_SYSTEM.md specification (frontend spec).
 */

// ============================================================================
// Fragments
// ============================================================================

/**
 * Inbox entry fragment (matches GraphQL type exactly)
 */
export const INBOX_ENTRY_FRAGMENT = /* GraphQL */ `
  fragment InboxEntry on InboxEntry {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    firmware
    securityPatch
    buildId
    fcmToken
    firebaseInstallId
    status
    receivedAt
    updatedAt
    acknowledgedAt
    approvedAt
    rejectedAt
    notes
  }
`;

/**
 * Device fragment (matches GraphQL type exactly)
 */
export const DEVICE_FRAGMENT = /* GraphQL */ `
  fragment Device on Device {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    fcmToken
    status
    registeredAt
    lastSeen
  }
`;

/**
 * Telemetry frame fragment
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

// ============================================================================
// Query Definitions
// ============================================================================

/**
 * Get inbox entries (paginated)
 * Query: inboxEntries
 */
export const GET_INBOX_ENTRIES = /* GraphQL */ `
  query GetInboxEntries($status: InboxStatus, $page: Int, $limit: Int) {
    inboxEntries(status: $status, page: $page, limit: $limit) {
      entries {
        ...InboxEntry
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${INBOX_ENTRY_FRAGMENT}
`;

/**
 * Get single inbox entry by IMEI
 * Query: inboxEntry
 */
export const GET_INBOX_ENTRY = /* GraphQL */ `
  query GetInboxEntry($imei: String!) {
    inboxEntry(imei: $imei) {
      ...InboxEntry
    }
  }
  ${INBOX_ENTRY_FRAGMENT}
`;

/**
 * Get all registered devices (paginated)
 * Query: devices
 */
export const GET_DEVICES = /* GraphQL */ `
  query GetDevices($status: DeviceStatus, $page: Int, $limit: Int) {
    devices(status: $status, page: $page, limit: $limit) {
      devices {
        ...Device
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${DEVICE_FRAGMENT}
`;

/**
 * Get single device by IMEI
 * Query: device
 */
export const GET_DEVICE = /* GraphQL */ `
  query GetDevice($imei: String!) {
    device(imei: $imei) {
      ...Device
    }
  }
  ${DEVICE_FRAGMENT}
`;

/**
 * Get device telemetry history
 * Query: deviceTelemetry
 */
export const GET_DEVICE_TELEMETRY = /* GraphQL */ `
  query GetDeviceTelemetry(
    $imei: String!
    $startTime: Int
    $endTime: Int
    $limit: Int
  ) {
    deviceTelemetry(
      imei: $imei
      startTime: $startTime
      endTime: $endTime
      limit: $limit
    ) {
      frames {
        ...TelemetryFrame
      }
      pagination {
        limit
        total
      }
    }
  }
  ${TELEMETRY_FRAME_FRAGMENT}
`;
import { graphqlClient } from '../_shared/graphql-client';

export async function queryInboxEntries(params?: { status?: string; page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_INBOX_ENTRIES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInboxEntry(imei: string) {
  return graphqlClient.query({
    query: GET_INBOX_ENTRY,
    variables: { imei },
    fetchPolicy: 'network-only',
  });
}

export async function queryDevices(params?: { status?: string; page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_DEVICES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDevice(imei: string) {
  return graphqlClient.query({
    query: GET_DEVICE,
    variables: { imei },
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceTelemetry(params: {
  imei: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
}) {
  return graphqlClient.query({
    query: GET_DEVICE_TELEMETRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
