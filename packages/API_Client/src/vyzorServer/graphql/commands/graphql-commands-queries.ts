import { COMMAND_LIST_FRAGMENT, COMMAND_FRAGMENT } from "./graphql-commands-fragments";

export const GET_COMMANDS = /* GraphQL */ `
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

export const GET_PENDING_COMMANDS = /* GraphQL */ `
  query GetPendingCommands($imei: String!) {
    pendingCommands(imei: $imei) {
      ...CommandList
    }
  }
  ${COMMAND_LIST_FRAGMENT}
`;

export const GET_COMMAND = /* GraphQL */ `
  query GetCommand($dispatchId: String!) {
    command(dispatchId: $dispatchId) {
      ...Command
    }
  }
  ${COMMAND_FRAGMENT}
`;

export const GET_COMMAND_STATUS = /* GraphQL */ `
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