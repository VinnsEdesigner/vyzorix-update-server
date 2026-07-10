/**
 * Device GraphQL Queries
 * 
 * GraphQL query definitions for devices.
 * Based on SERVER_BACKEND_DASHBOARD_COMMANDS_API.md and SERVER_BACKEND_DEVICE_REGISTRATION_API.md.
 */

// ============================================================================
// Fragments
// ============================================================================

/**
 * Device list item fragment
 */
export const DEVICE_LIST_FRAGMENT = /* GraphQL */ `
  fragment DeviceList on Device {
    id
    imei
    deviceName
    model
    manufacturer
    online
    lastSeen
  }
`;

/**
 * Full device fragment
 */
export const DEVICE_FRAGMENT = /* GraphQL */ `
  fragment Device on Device {
    id
    imei
    deviceName
    model
    manufacturer
    appVersion
    osVersion
    securityPatch
    buildId
    online
    registeredAt
    lastSeen
    fcmTokenValid
    commandSecretSet
  }
`;

// ============================================================================
// Query Definitions
// ============================================================================

/**
 * Get all devices (paginated)
 */
export const GET_DEVICES = /* GraphQL */ `
  query GetDevices($page: Int, $limit: Int, $status: String) {
    devices(page: $page, limit: $limit, status: $status) {
      devices {
        ...DeviceList
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
  ${DEVICE_LIST_FRAGMENT}
`;

/**
 * Get single device by IMEI
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
 * Get device status (online/offline)
 */
export const GET_DEVICE_STATUS = /* GraphQL */ `
  query GetDeviceStatus($imei: String!) {
    deviceStatus(imei: $imei) {
      imei
      online
      lastSeen
      connectionStatus
    }
  }
`;

/**
 * Get device stats for dashboard
 */
export const GET_DEVICE_STATS = /* GraphQL */ `
  query GetDeviceStats {
    dashboardStats {
      devices {
        total
        online
        offline
      }
      commands {
        totalToday
        pending
        failed
      }
      activity {
        last24h {
          commands
          registrations
          deregistrations
        }
      }
    }
  }
`;
// ============================================================================
// Query Functions (using Apollo Client)
// ============================================================================

import { graphqlClient } from '../_shared/graphql-client';

export async function queryDevices(params?: { page?: number; limit?: number; status?: string }) {
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

export async function queryDeviceStatus(imei: string) {
  return graphqlClient.query({
    query: GET_DEVICE_STATUS,
    variables: { imei },
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceStats() {
  return graphqlClient.query({
    query: GET_DEVICE_STATS,
    fetchPolicy: 'network-only',
  });
}
