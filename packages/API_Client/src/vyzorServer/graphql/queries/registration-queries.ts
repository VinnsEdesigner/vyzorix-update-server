import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_INBOX_ENTRIES = gql`
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

export const GET_INBOX_ENTRY = gql`
  query GetInboxEntry($organizationId: ID!, $imei: String!) {
    inboxEntry(organizationId: $organizationId, imei: $imei) {
      ...InboxEntry
    }
  }
`;

export const GET_DEVICES = gql`
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

export const GET_DEVICE = gql`
  query GetDevice($organizationId: ID!, $id: ID!) {
    device(organizationId: $organizationId, id: $id) {
      ...Device
    }
  }
`;

export const GET_DEVICE_TELEMETRY = gql`
  query GetDeviceTelemetry($organizationId: ID!, $deviceId: ID!, $limit: Int) {
    telemetryHistory(organizationId: $organizationId, deviceId: $deviceId, limit: $limit) {
      ...TelemetryFrame
    }
  }
`;

export async function queryInboxEntries(params: { organizationId: string; status?: string; page?: number; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRIES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInboxEntry(params: { organizationId: string; imei: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDevices(params: { organizationId: string; limit?: number; offset?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_DEVICES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDevice(params: { organizationId: string; id: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_DEVICE,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceTelemetry(params: { organizationId: string; deviceId: string; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_TELEMETRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
