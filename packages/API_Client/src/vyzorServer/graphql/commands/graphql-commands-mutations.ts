import { graphqlClient } from '../_shared/graphql-client';

export const SEND_COMMAND = `
  mutation SendCommand($organizationId: ID!, $deviceId: ID!, $commandType: String!, $params: JSON) {
    sendCommand(organizationId: $organizationId, deviceId: $deviceId, commandType: $commandType, params: $params) {
      id
      device_id
      command_type
      payload
      status
      created_at
    }
  }
`;

export const DISMISS_COMMAND = `
  mutation DismissCommand($organizationId: ID!, $dispatchId: ID!) {
    dismissCommand(organizationId: $organizationId, dispatchId: $dispatchId) {
      success
    }
  }
`;

export async function mutateSendCommand(params: { organizationId: string; deviceId: string; commandType: string; params?: Record<string, unknown> }) {
  return graphqlClient.getClient().mutate({
    mutation: SEND_COMMAND,
    variables: params,
  });
}

export async function mutateDismissCommand(params: { organizationId: string; dispatchId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: DISMISS_COMMAND,
    variables: params,
  });
}
