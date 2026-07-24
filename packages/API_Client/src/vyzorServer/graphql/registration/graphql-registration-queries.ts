






export const INBOX_ENTRY_FRAGMENT =  `
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


export const DEVICE_FRAGMENT =  `
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


export const TELEMETRY_FRAME_FRAGMENT =  `
  fragment TelemetryFrame on TelemetryFrame {
    timestamp
    riskScore
    thermalTemp
    bufferLevel
    uptime
  }
`;






export const GET_INBOX_ENTRIES =  `
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


export const GET_INBOX_ENTRY =  `
  query GetInboxEntry($imei: String!) {
    inboxEntry(imei: $imei) {
      ...InboxEntry
    }
  }
  ${INBOX_ENTRY_FRAGMENT}
`;


export const GET_DEVICES =  `
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


export const GET_DEVICE =  `
  query GetDevice($imei: String!) {
    device(imei: $imei) {
      ...Device
    }
  }
  ${DEVICE_FRAGMENT}
`;


export const GET_DEVICE_TELEMETRY =  `
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
