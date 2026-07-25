import { graphqlClient } from '../_shared/graphql-client';
import { OPERATOR_SETTINGS_FRAGMENT, DEVICE_SETTINGS_FRAGMENT, ORGANIZATION_SETTINGS_FRAGMENT } from "./graphql-settings-fragments";

export const GET_SETTINGS = `
  ${OPERATOR_SETTINGS_FRAGMENT}
  query GetSettings {
    mySettings {
      ...OperatorSettings
    }
  }
`;

export const GET_DEVICE_SETTINGS = `
  ${DEVICE_SETTINGS_FRAGMENT}
  query GetDeviceSettings($organizationId: ID!, $deviceImei: String!) {
    deviceSettings(organizationId: $organizationId, deviceImei: $deviceImei) {
      ...DeviceSettings
    }
  }
`;

export const GET_ORGANIZATION_SETTINGS = `
  ${ORGANIZATION_SETTINGS_FRAGMENT}
  query GetOrganizationSettings($organizationId: ID!) {
    organizationSettings(organizationId: $organizationId) {
      ...OrganizationSettings
    }
  }
`;

export async function querySettings() {
  return graphqlClient.getClient().query({
    query: GET_SETTINGS,
    fetchPolicy: 'network-only',
  });
}

export async function queryDeviceSettings(params: { organizationId: string; deviceImei: string }) {
  return graphqlClient.getClient().query({
    query: GET_DEVICE_SETTINGS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryOrganizationSettings(organizationId: string) {
  return graphqlClient.getClient().query({
    query: GET_ORGANIZATION_SETTINGS,
    variables: { organizationId },
    fetchPolicy: 'network-only',
  });
}
