import { LOG_ENTRY_FRAGMENT } from "./graphql-logs-fragments";

export const GET_LOGS = /* GraphQL */ `
  query GetLogs($imei: String!, $type: String, $limit: Int, $cursor: String) {
    logs(imei: $imei, type: $type, limit: $limit, cursor: $cursor) {
      logs {
        ...LogEntry
      }
      pagination {
        limit
        hasMore
        nextCursor
      }
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;

export const GET_LOG_STATS = /* GraphQL */ `
  query GetLogStats($imei: String!, $startTime: Int, $endTime: Int) {
    logStats(imei: $imei, startTime: $startTime, endTime: $endTime) {
      total
      byType {
        connection
        command
        telemetry
        error
        warning
      }
    }
  }
`;

export const GET_LOG = /* GraphQL */ `
  query GetLog($id: String!) {
    log(id: $id) {
      ...LogEntry
    }
  }
  ${LOG_ENTRY_FRAGMENT}
`;