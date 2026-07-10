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
import { graphqlClient } from '../_shared/graphql-client';

export async function queryLogs(params: { imei: string; type?: string; limit?: number; cursor?: string }) {
  return graphqlClient.query({
    query: GET_LOGS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryLogStats(params: { imei: string; startTime?: number; endTime?: number }) {
  return graphqlClient.query({
    query: GET_LOG_STATS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryLog(id: string) {
  return graphqlClient.query({
    query: GET_LOG,
    variables: { id },
    fetchPolicy: 'network-only',
  });
}
