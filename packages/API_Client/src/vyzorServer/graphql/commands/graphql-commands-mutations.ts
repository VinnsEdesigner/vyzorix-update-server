import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const SEND_COMMAND = gql`
  mutation SendCommand($deviceId: ID!, $command: String!, $args: JSON) {
    sendCommand(deviceId: $deviceId, command: $command, args: $args) {
      dispatchId
      commandId
      status
      deviceOnline
    }
  }
`;

export const RETRY_COMMAND = gql`
  mutation RetryCommand($dispatchId: ID!) {
    retryCommand(dispatchId: $dispatchId) {
      dispatchId
      commandId
      deviceId
      command
      status
      createdAt
      deliveredAt
    }
  }
`;

export const CANCEL_COMMAND = gql`
  mutation CancelCommand($dispatchId: ID!) {
    cancelCommand(dispatchId: $dispatchId) {
      dispatchId
      cancelledAt
      status
    }
  }
`;

export async function mutateSendCommand(params: { deviceId: string; command: string; args?: Record<string, unknown> }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: SEND_COMMAND,
    variables: params,
  });
}

export async function mutateRetryCommand(params: { dispatchId: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: RETRY_COMMAND,
    variables: params,
  });
}

export async function mutateCancelCommand(params: { dispatchId: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_COMMAND,
    variables: params,
  });
}
