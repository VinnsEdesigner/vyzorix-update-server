import { graphqlClient } from '../_shared/graphql-client';

export const GET_PENDING_COMMANDS = `
  query GetPendingCommands($organizationId: ID!, $deviceId: ID!) {
    pendingCommands(organizationId: $organizationId, deviceId: $deviceId) {
      id
      device_id
      command_type
      payload
      status
      created_at
      dispatched_at
    }
  }
`;

export const GET_COMMAND = `
  query GetCommand($organizationId: ID!, $dispatchId: ID!) {
    command(organizationId: $organizationId, dispatchId: $dispatchId) {
      id
      device_id
      command_type
      payload
      status
      created_at
      dispatched_at
      completed_at
      result {
        success
        message
        error
      }
    }
  }
`;

export async function queryPendingCommands(params: { organizationId: string; deviceId: string }) {
  return graphqlClient.getClient().query({
    query: GET_PENDING_COMMANDS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryCommand(params: { organizationId: string; dispatchId: string }) {
  return graphqlClient.getClient().query({
    query: GET_COMMAND,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
