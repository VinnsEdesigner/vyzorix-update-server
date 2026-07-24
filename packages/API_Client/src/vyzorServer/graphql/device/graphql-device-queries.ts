






export const DEVICE_LIST_FRAGMENT =  `
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


export const DEVICE_FRAGMENT =  `
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






export const GET_DEVICES =  `
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


export const GET_DEVICE =  `
  query GetDevice($imei: String!) {
    device(imei: $imei) {
      ...Device
    }
  }
  ${DEVICE_FRAGMENT}
`;


export const GET_DEVICE_STATUS =  `
  query GetDeviceStatus($imei: String!) {
    deviceStatus(imei: $imei) {
      imei
      online
      lastSeen
      connectionStatus
    }
  }
`;


export const GET_DEVICE_STATS =  `
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
