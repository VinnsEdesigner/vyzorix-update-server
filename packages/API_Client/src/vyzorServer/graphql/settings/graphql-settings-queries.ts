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
import { graphqlClient } from '../_shared/graphql-client';

export async function querySettings() {
  return graphqlClient.query({
    query: GET_SETTINGS,
    fetchPolicy: 'network-only',
  });
}

export async function queryThresholds() {
  return graphqlClient.query({
    query: GET_THRESHOLDS,
    fetchPolicy: 'network-only',
  });
}

export async function queryNotifications() {
  return graphqlClient.query({
    query: GET_NOTIFICATIONS,
    fetchPolicy: 'network-only',
  });
}

export async function queryOperator() {
  return graphqlClient.query({
    query: GET_OPERATOR,
    fetchPolicy: 'network-only',
  });
}
