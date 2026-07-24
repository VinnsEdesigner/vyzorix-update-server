import { graphqlClient } from '../_shared/graphql-client';

export interface ClientSettingsInput {
  theme?: string;
  language?: string;
  timezone?: string;
}

export interface ThresholdInput {
  temp_min?: number;
  temp_max?: number;
  battery_min?: number;
  battery_max?: number;
  speed_max?: number;
  distance_max?: number;
}

export interface NotificationInput {
  enabled?: boolean;
  email_enabled?: boolean;
  push_enabled?: boolean;
  webhook_enabled?: boolean;
  webhook_url?: string;
}

export const UPDATE_SETTINGS = `
  mutation UpdateSettings($input: ClientSettingsInput) {
    updateMySettings(input: $input) {
      operator {
        id
        email
        name
      }
      client {
        theme
        language
        timezone
      }
    }
  }
`;

export const UPDATE_THRESHOLDS = `
  mutation UpdateThresholds($organizationId: ID!, $input: ThresholdInput) {
    updateOrganizationThresholds(organizationId: $organizationId, input: $input) {
      thresholds {
        temp_min
        temp_max
        battery_min
        battery_max
        speed_max
        distance_max
      }
    }
  }
`;

export const UPDATE_NOTIFICATIONS = `
  mutation UpdateNotifications($input: NotificationInput) {
    updateMyNotifications(input: $input) {
      enabled
      email {
        enabled
      }
      push {
        enabled
      }
      webhook {
        enabled
        url
      }
    }
  }
`;

export const UPDATE_DEVICE_SETTINGS = `
  mutation UpdateDeviceSettings($organizationId: ID!, $deviceImei: String!, $input: ThresholdInput) {
    updateDeviceSettings(organizationId: $organizationId, deviceImei: $deviceImei, input: $input) {
      device_id
      effective_thresholds {
        temp_min
        temp_max
        battery_min
        battery_max
        speed_max
        distance_max
      }
    }
  }
`;

export async function mutateUpdateSettings(input: ClientSettingsInput) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_SETTINGS,
    variables: { input },
  });
}

export async function mutateUpdateThresholds(organizationId: string, input: ThresholdInput) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_THRESHOLDS,
    variables: { organizationId, input },
  });
}

export async function mutateUpdateNotifications(input: NotificationInput) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_NOTIFICATIONS,
    variables: { input },
  });
}

export async function mutateUpdateDeviceSettings(params: { organizationId: string; deviceImei: string; input: ThresholdInput }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_DEVICE_SETTINGS,
    variables: params,
  });
}
