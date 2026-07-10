/**
 * Commands GraphQL Mutations
 * 
 * GraphQL mutation definitions for commands.
 * Based on SERVER_BACKEND_DASHBOARD_COMMANDS_API.md.
 */

// ============================================================================
// Mutation Definitions
// ============================================================================

/**
 * Send a command to a device
 */
export const SEND_COMMAND = /* GraphQL */ `
  mutation SendCommand($imei: String!, $commandType: String!, $params: JSON) {
    sendCommand(imei: $imei, commandType: $commandType, params: $params) {
      id
      dispatchId
      type
      deviceImei
      status
      createdAt
    }
  }
`;

/**
 * Cancel a pending command
 */
export const CANCEL_COMMAND = /* GraphQL */ `
  mutation CancelCommand($imei: String!, $dispatchId: String!) {
    cancelCommand(imei: $imei, dispatchId: $dispatchId) {
      success
    }
  }
`;

/**
 * Retry a failed command
 */
export const RETRY_COMMAND = /* GraphQL */ `
  mutation RetryCommand($dispatchId: String!) {
    retryCommand(dispatchId: $dispatchId) {
      id
      dispatchId
      type
      deviceImei
      status
      createdAt
    }
  }
`;
import { graphqlClient } from '../_shared/graphql-client';

export async function mutateSendCommand(imei: string, commandType: string, params?: Record<string, unknown>) {
  return graphqlClient.mutate({
    mutation: SEND_COMMAND,
    variables: { imei, commandType, params },
  });
}

export async function mutateCancelCommand(imei: string, dispatchId: string) {
  return graphqlClient.mutate({
    mutation: CANCEL_COMMAND,
    variables: { imei, dispatchId },
  });
}

export async function mutateRetryCommand(dispatchId: string) {
  return graphqlClient.mutate({
    mutation: RETRY_COMMAND,
    variables: { dispatchId },
  });
}
