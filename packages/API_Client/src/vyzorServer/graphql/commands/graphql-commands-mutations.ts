import { graphqlClient } from '../_shared/graphql-client';

export const SEND_COMMAND = `
  mutation SendCommand($deviceId: ID!, $command: String!, $args: JSON) {
    sendCommand(deviceId: $deviceId, command: $command, args: $args) {
      dispatchId
      commandId
      deviceId
      command
      args
      status
      createdAt
    }
  }
`;

export const RETRY_COMMAND = `
  mutation RetryCommand($dispatchId: ID!) {
    retryCommand(dispatchId: $dispatchId) {
      dispatchId
      commandId
      deviceId
      command
      status
      createdAt
    }
  }
`;

export const CANCEL_COMMAND = `
  mutation CancelCommand($dispatchId: ID!) {
    cancelCommand(dispatchId: $dispatchId) {
      success
      error
    }
  }
`;

export async function mutateSendCommand(params: { deviceId: string; command: string; args?: Record<string, unknown> }) {
  return graphqlClient.getClient().mutate({
    mutation: SEND_COMMAND,
    variables: params,
  });
}

export async function mutateRetryCommand(params: { dispatchId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: RETRY_COMMAND,
    variables: params,
  });
}

export async function mutateCancelCommand(params: { dispatchId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_COMMAND,
    variables: params,
  });
}
