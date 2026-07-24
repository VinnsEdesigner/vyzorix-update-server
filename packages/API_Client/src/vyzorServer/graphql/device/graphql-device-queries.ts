import { graphqlClient } from '../_shared/graphql-client';

export const DEVICE_LIST_FRAGMENT = `
  fragment DeviceList on Device {
    id
    imei
    device_name
    model
    manufacturer
    status
    last_seen
  }
`;

export const DEVICE_FRAGMENT = `
  fragment Device on Device {
    id
    imei
    device_name
    model
    manufacturer
    app_version
    os_version
    security_patch
    build_id
    status
    registered_at
    last_seen
    fcm_token_valid
    command_secret_set
    connection {
      web_socket_status
      connected_at
      protocol
      client_ip
    }
  }
`;

export const GET_DEVICES = `
  query GetDevices($organizationId: ID!, $limit: Int, $offset: Int, $online: Boolean) {
    devices(organizationId: $organizationId, limit: $limit, offset: $offset, online: $online) {
      id
      imei
      app_version
      device_class
      last_seen
      online
    }
  }
`;

export const GET_DEVICE = `
  query GetDevice($organizationId: ID!, $id: ID!) {
    device(organizationId: $organizationId, id: $id) {
      id
      imei
      device_name
      model
      manufacturer
      app_version
      os_version
      security_patch
      build_id
      status
      registered_at
      last_seen
      fcm_token_valid
      command_secret_set
      connection {
        web_socket_status
        connected_at
        protocol
        client_ip
      }
    }
  }
`;

export const GET_DEVICE_COUNT = `
  query GetDeviceCount($organizationId: ID!) {
    deviceCount(organizationId: $organizationId)
  }
`;

export async function queryDevices(params: { organizationId: string; limit?: number; offset?: number; online?: boolean }) {
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

export async function queryDeviceCount(organizationId: string) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_COUNT,
    variables: { organizationId },
    fetchPolicy: 'network-only',
  });
}
