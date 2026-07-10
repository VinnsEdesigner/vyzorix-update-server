/**
 * Settings GraphQL Mutations
 * 
 * GraphQL mutation definitions for settings.
 * Based on SERVER_BACKEND_SETTINGS_API.md specification.
 */

// ============================================================================
// Mutation Definitions
// ============================================================================

/**
 * Update client settings
 */
export const UPDATE_SETTINGS = /* GraphQL */ `
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

/**
 * Update thresholds
 */
export const UPDATE_THRESHOLDS = /* GraphQL */ `
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

/**
 * Update notifications
 */
export const UPDATE_NOTIFICATIONS = /* GraphQL */ `
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

/**
 * Test webhook
 */
export const TEST_WEBHOOK = /* GraphQL */ `
  mutation TestWebhook($url: String!) {
    testWebhook(url: $url) {
      success
      statusCode
      responseTime
      error
    }
  }
`;

/**
 * Rotate webhook secret
 */
export const ROTATE_WEBHOOK_SECRET = /* GraphQL */ `
  mutation RotateWebhookSecret {
    rotateWebhookSecret
  }
`;

/**
 * Reset settings to defaults
 */
export const RESET_SETTINGS = /* GraphQL */ `
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

/**
 * Update operator profile
 */
export const UPDATE_OPERATOR = /* GraphQL */ `
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

// ============================================================================
// Input Types (for reference)
// ============================================================================

/**
 * ClientSettingsInput:
 * {
 *   serverUrl?: string
 *   deviceId?: string
 *   requestTimeoutMs?: number
 *   autoReconnect?: boolean
 *   strictHmac?: boolean
 *   logBufferLimit?: number
 *   signalHistoryLimit?: number
 * }
 * 
 * ThresholdsInput:
 * {
 *   riskWarn?: number
 *   riskCrit?: number
 *   thermalWarn?: number
 *   thermalCrit?: number
 *   bufferWarn?: number
 *   bufferCrit?: number
 * }
 * 
 * NotificationInput:
 * {
 *   enabled?: boolean
 *   email?: EmailNotificationInput
 *   push?: PushNotificationInput
 *   webhook?: WebhookNotificationInput
 * }
 * 
 * EmailNotificationInput / PushNotificationInput:
 * {
 *   thresholdBreach?: boolean
 *   deviceOffline?: boolean
 *   deviceOnline?: boolean
 *   updateAvailable?: boolean
 *   commandFailed?: boolean
 *   registrationRequest?: boolean
 * }
 * 
 * WebhookNotificationInput:
 * {
 *   enabled?: boolean
 *   url?: string
 *   types?: [String!]
 * }
 */
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
