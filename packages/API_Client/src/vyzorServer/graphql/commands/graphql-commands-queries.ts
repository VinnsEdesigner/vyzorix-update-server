import { COMMAND_LIST_FRAGMENT, COMMAND_FRAGMENT } from "./graphql-commands-fragments";

export const GET_COMMANDS =  `
  query GetCommands($imei: String!, $status: String, $page: Int, $limit: Int) {
    commands(imei: $imei, status: $status, page: $page, limit: $limit) {
      commands {
        ...CommandList
      }
      pagination {
        page
        limit
        total
        totalPages
        hasMore
      }
    }
  }
  ${COMMAND_LIST_FRAGMENT}
`;

export const GET_PENDING_COMMANDS =  `
  query GetPendingCommands($imei: String!) {
    pendingCommands(imei: $imei) {
      ...CommandList
    }
  }
  ${COMMAND_LIST_FRAGMENT}
`;

export const GET_COMMAND =  `
  query GetCommand($dispatchId: String!) {
    command(dispatchId: $dispatchId) {
      ...Command
    }
  }
  ${COMMAND_FRAGMENT}
`;

export const GET_COMMAND_STATUS =  `
  query GetCommandStatus($dispatchId: String!) {
    commandStatus(dispatchId: $dispatchId) {
      dispatchId
      status
      updatedAt
      result {
        success
        message
        error
      }
    }
  }
`;
import { graphqlClient } from '../_shared/graphql-client';

export async function queryCommands(params: { imei: string; status?: string; page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_COMMANDS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryPendingCommands(imei: string) {
  return graphqlClient.query({
    query: GET_PENDING_COMMANDS,
    variables: { imei },
    fetchPolicy: 'network-only',
  });
}

export async function queryCommand(dispatchId: string) {
  return graphqlClient.query({
    query: GET_COMMAND,
    variables: { dispatchId },
    fetchPolicy: 'network-only',
  });
}

export async function queryCommandStatus(dispatchId: string) {
  return graphqlClient.query({
    query: GET_COMMAND_STATUS,
    variables: { dispatchId },
    fetchPolicy: 'network-only',
  });
}
