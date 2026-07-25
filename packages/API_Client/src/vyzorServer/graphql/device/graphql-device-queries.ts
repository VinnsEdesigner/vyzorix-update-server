import { graphqlClient } from '../_shared/graphql-client';

export const DEVICE_LIST_FRAGMENT = `
  fragment DeviceList on Device {
    id
    imei
    name
    deviceName
    model
    manufacturer
    status
    lastSeen
    online
  }
`;

export const DEVICE_FRAGMENT = `
  fragment Device on Device {
    id
    imei
    name
    deviceName
    model
    manufacturer
    appVersion
    osVersion
    securityPatch
    buildId
    status
    registeredAt
    lastSeen
    fcmTokenValid
    commandSecretSet
    connection {
      webSocketStatus
      connectedAt
      protocol
      clientIp
    }
  }
`;

export const GET_DEVICES = `
  query GetDevices($organizationId: ID!, $limit: Int, $offset: Int) {
    devices(organizationId: $organizationId, limit: $limit, offset: $offset) {
      id
      imei
      name
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
  query GetDevice($organizationId: ID!, $id: ID!) {
    device(organizationId: $organizationId, id: $id) {
      id
      imei
      name
      deviceName
      model
      manufacturer
      appVersion
      osVersion
      securityPatch
      buildId
      status
      registeredAt
      lastSeen
      fcmTokenValid
      commandSecretSet
      connection {
        webSocketStatus
        connectedAt
        protocol
        clientIp
      }
    }
  }
`;

export const GET_DEVICE_COUNT = `
  query GetDeviceCount($organizationId: ID!) {
    deviceCount(organizationId: $organizationId)
  }
`;

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

export async function queryDeviceCount(organizationId: string) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_COUNT,
    variables: { organizationId },
    fetchPolicy: 'network-only',
  });
}
