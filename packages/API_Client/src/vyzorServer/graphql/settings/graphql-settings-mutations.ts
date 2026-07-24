






export const UPDATE_SETTINGS =  `
  mutation UpdateSettings($input: ClientSettingsInput!) {
    updateMySettings(input: $input) {
      client {
        serverUrl
        deviceId
        requestTimeoutMs
        autoReconnect
        strictHmac
        logBufferLimit
        signalHistoryLimit
      }
      thresholds {
        riskWarn
        riskCrit
        thermalWarn
        thermalCrit
        bufferWarn
        bufferCrit
      }
      notifications {
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
          secret
          types
        }
      }
    }
  }
`;


export const UPDATE_THRESHOLDS =  `
  mutation UpdateThresholds($input: ThresholdsInput!) {
    updateMyThresholds(input: $input) {
      riskWarn
      riskCrit
      thermalWarn
      thermalCrit
      bufferWarn
      bufferCrit
    }
  }
`;


export const UPDATE_NOTIFICATIONS =  `
  mutation UpdateNotifications($input: NotificationInput!) {
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
        secret
        types
      }
    }
  }
`;


export const TEST_WEBHOOK =  `
  mutation TestWebhook($url: String!) {
    testWebhook(url: $url) {
      success
      statusCode
      responseTime
      error
    }
  }
`;


export const ROTATE_WEBHOOK_SECRET =  `
  mutation RotateWebhookSecret {
    rotateWebhookSecret
  }
`;


export const RESET_SETTINGS =  `
  mutation ResetSettings {
    resetMySettings {
      client {
        serverUrl
        deviceId
        requestTimeoutMs
        autoReconnect
        strictHmac
        logBufferLimit
        signalHistoryLimit
      }
      thresholds {
        riskWarn
        riskCrit
        thermalWarn
        thermalCrit
        bufferWarn
        bufferCrit
      }
      notifications {
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
          secret
          types
        }
      }
    }
  }
`;


export const UPDATE_OPERATOR =  `
  mutation UpdateOperator($name: String, $email: String) {
    updateMe(name: $name, email: $email) {
      id
      email
      name
      role
      permissions
      createdAt
    }
  }
`;






import { graphqlClient } from '../_shared/graphql-client';

export interface ClientSettingsInput {
  serverUrl?: string;
  deviceId?: string;
  requestTimeoutMs?: number;
  autoReconnect?: boolean;
  strictHmac?: boolean;
  logBufferLimit?: number;
  signalHistoryLimit?: number;
}

export interface ThresholdsInput {
  riskWarn?: number;
  riskCrit?: number;
  thermalWarn?: number;
  thermalCrit?: number;
  bufferWarn?: number;
  bufferCrit?: number;
}

export interface NotificationInput {
  enabled?: boolean;
  email?: EmailNotificationInput;
  push?: PushNotificationInput;
  webhook?: WebhookNotificationInput;
}

export interface EmailNotificationInput {
  thresholdBreach?: boolean;
  deviceOffline?: boolean;
  deviceOnline?: boolean;
  updateAvailable?: boolean;
  commandFailed?: boolean;
  registrationRequest?: boolean;
}

export interface PushNotificationInput {
  thresholdBreach?: boolean;
  deviceOffline?: boolean;
  deviceOnline?: boolean;
  updateAvailable?: boolean;
  commandFailed?: boolean;
  registrationRequest?: boolean;
}

export interface WebhookNotificationInput {
  enabled?: boolean;
  url?: string;
  types?: string[];
}

export async function mutateUpdateSettings(input: ClientSettingsInput) {
  return graphqlClient.mutate({
    mutation: UPDATE_SETTINGS,
    variables: { input },
  });
}

export async function mutateUpdateThresholds(input: ThresholdsInput) {
  return graphqlClient.mutate({
    mutation: UPDATE_THRESHOLDS,
    variables: { input },
  });
}

export async function mutateUpdateNotifications(input: NotificationInput) {
  return graphqlClient.mutate({
    mutation: UPDATE_NOTIFICATIONS,
    variables: { input },
  });
}

export async function mutateTestWebhook(url: string) {
  return graphqlClient.mutate({
    mutation: TEST_WEBHOOK,
    variables: { url },
  });
}

export async function mutateRotateWebhookSecret() {
  return graphqlClient.mutate({
    mutation: ROTATE_WEBHOOK_SECRET,
  });
}

export async function mutateResetSettings() {
  return graphqlClient.mutate({
    mutation: RESET_SETTINGS,
  });
}

export async function mutateUpdateOperator(name?: string, email?: string) {
  return graphqlClient.mutate({
    mutation: UPDATE_OPERATOR,
    variables: { name, email },
  });
}
