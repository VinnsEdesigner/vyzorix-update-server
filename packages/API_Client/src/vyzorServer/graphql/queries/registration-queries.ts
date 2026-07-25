import { INBOX_ENTRY_FRAGMENT } from "../fragments/inbox-entry.fragment";
import { DEVICE_FRAGMENT } from "../fragments/device.fragment";
import { TELEMETRY_FRAME_FRAGMENT } from "../fragments/telemetry.fragment";
import { graphqlClient } from '../_shared/graphql-client';

export const GET_INBOX_ENTRIES = `
  ${INBOX_ENTRY_FRAGMENT}
  query GetInboxEntries($organizationId: ID!, $status: String, $page: Int, $limit: Int) {
    inbox(organizationId: $organizationId, status: $status, page: $page, limit: $limit) {
      requests {
        ...InboxEntry
      }
      pagination {
        page
        limit
        hasMore
      }
    }
  }
`;

export const GET_INBOX_ENTRY = `
  ${INBOX_ENTRY_FRAGMENT}
  query GetInboxEntry($organizationId: ID!, $imei: String!) {
    inboxEntry(organizationId: $organizationId, imei: $imei) {
      ...InboxEntry
    }
  }
`;

export const GET_DEVICES = `
  ${DEVICE_FRAGMENT}
  query GetDevices($organizationId: ID!, $limit: Int, $offset: Int) {
    devices(organizationId: $organizationId, limit: $limit, offset: $offset) {
      id
      imei
      deviceName
      model
      manufacturer
      status
      lastSeen
      online
    }
  }
`;

export const GET_DEVICE = `
  ${DEVICE_FRAGMENT}
  query GetDevice($organizationId: ID!, $id: ID!) {
    device(organizationId: $organizationId, id: $id) {
      ...Device
    }
  }
`;

export const GET_DEVICE_TELEMETRY = `
  ${TELEMETRY_FRAME_FRAGMENT}
  query GetDeviceTelemetry($organizationId: ID!, $deviceId: ID!, $limit: Int) {
    telemetryHistory(organizationId: $organizationId, deviceId: $deviceId, limit: $limit) {
      ...TelemetryFrame
    }
  }
`;

export async function queryInboxEntries(params: { organizationId: string; status?: string; page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRIES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInboxEntry(params: { organizationId: string; imei: string }) {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDevices(params: { organizationId: string; limit?: number; offset?: number }) {
  return graphqlClient.getClient().query({
    query: GET_DEVICES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDevice(params: { organizationId: string; id: string }) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceTelemetry(params: { organizationId: string; deviceId: string; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_TELEMETRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
