import { SETTINGS_FRAGMENT, OPERATOR_FRAGMENT, THRESHOLDS_FRAGMENT } from "./graphql-settings-fragments";

export const GET_SETTINGS = /* GraphQL */ `
  query GetSettings {
    mySettings {
      ...Settings
    }
  }
  ${SETTINGS_FRAGMENT}
`;

export const GET_THRESHOLDS = /* GraphQL */ `
  query GetThresholds {
    myThresholds {
      ...Thresholds
    }
  }
  ${THRESHOLDS_FRAGMENT}
`;

export const GET_NOTIFICATIONS = /* GraphQL */ `
  query GetNotifications {
    myNotifications {
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

export const GET_OPERATOR = /* GraphQL */ `
  query GetOperator {
    me {
      ...Operator
    }
  }
  ${OPERATOR_FRAGMENT}
`;