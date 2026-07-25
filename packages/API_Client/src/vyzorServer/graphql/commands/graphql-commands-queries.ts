import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_PENDING_COMMANDS = gql`
  query GetPendingCommands($organizationId: ID!, $deviceId: ID!) {
    pendingCommands(organizationId: $organizationId, deviceId: $deviceId) {
      dispatchId
      commandId
      deviceId
      command
      args
      status
      createdAt
      deliveredAt
    }
  }
`;

export const GET_COMMAND = gql`
  query GetCommand($organizationId: ID!, $dispatchId: ID!) {
    command(organizationId: $organizationId, dispatchId: $dispatchId) {
      dispatchId
      commandId
      deviceId
      command
      args
      status
      createdAt
      deliveredAt
    }
  }
`;

export async function queryPendingCommands(params: { organizationId: string; deviceId: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_PENDING_COMMANDS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryCommand(params: { organizationId: string; dispatchId: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_COMMAND,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
