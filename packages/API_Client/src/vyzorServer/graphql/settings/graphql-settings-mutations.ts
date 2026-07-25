import { graphqlClient } from '../_shared/graphql-client';

export interface ThresholdInput {
  riskWarn?: number;
  riskCrit?: number;
  thermalWarn?: number;
  thermalCrit?: number;
  bufferWarn?: number;
  bufferCrit?: number;
}

export interface DeviceSettingsInput {
  customName?: string;
  location?: string;
  thresholds?: ThresholdInput;
}

export interface OrganizationSettingsInput {
  timezone?: string;
  dateFormat?: string;
  alertCooldownMinutes?: number;
  defaultThresholds?: ThresholdInput;
}

export const UPDATE_NOTIFICATIONS = `
  mutation UpdateNotifications($input: UpdateNotificationsInput!) {
    updateMyNotifications(input: $input) {
      enabled
      channels
      email {
        thresholdBreach
        deviceOffline
        deviceOnline
        updateAvailable
        commandFailed
        registrationRequest
      }
      push {
        thresholdBreach
        deviceOffline
        deviceOnline
        updateAvailable
        commandFailed
        registrationRequest
      }
      webhook {
        enabled
        url
        types
      }
    }
  }
`;

export const UPDATE_DEVICE_SETTINGS = `
  mutation UpdateDeviceSettings($organizationId: ID!, $deviceImei: String!, $input: UpdateDeviceSettingsInput!) {
    updateDeviceSettings(organizationId: $organizationId, deviceImei: $deviceImei, input: $input) {
      deviceImei
      customName
      location
      thresholds {
        riskWarn
        riskCrit
        thermalWarn
        thermalCrit
        bufferWarn
        bufferCrit
      }
      effectiveThresholds {
        riskWarn
        riskCrit
        thermalWarn
        thermalCrit
        bufferWarn
        bufferCrit
      }
    }
  }
`;

export const UPDATE_ORGANIZATION_SETTINGS = `
  mutation UpdateOrganizationSettings($organizationId: ID!, $input: UpdateOrganizationSettingsInput!) {
    updateOrganizationSettings(organizationId: $organizationId, input: $input) {
      organizationId
      timezone
      dateFormat
      alertCooldownMinutes
      defaultThresholds {
        riskWarn
        riskCrit
        thermalWarn
        thermalCrit
        bufferWarn
        bufferCrit
      }
    }
  }
`;

export async function mutateUpdateNotifications(input: any) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_NOTIFICATIONS,
    variables: { input },
  });
}

export async function mutateUpdateDeviceSettings(params: { organizationId: string; deviceImei: string; input: DeviceSettingsInput }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_DEVICE_SETTINGS,
    variables: params,
  });
}

export async function mutateUpdateOrganizationSettings(params: { organizationId: string; input: OrganizationSettingsInput }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_ORGANIZATION_SETTINGS,
    variables: params,
  });
}
