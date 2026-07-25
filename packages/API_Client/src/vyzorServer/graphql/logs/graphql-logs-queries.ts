import { graphqlClient } from '../_shared/graphql-client';

export const LOG_ENTRY_FRAGMENT = `
  fragment LogEntry on LogEntry {
    id
    type
    timestamp
    data
  }
`;

export const GET_LOGS = `
  ${LOG_ENTRY_FRAGMENT}
  query GetLogs($organizationId: ID!, $imei: ID!, $type: String, $startTime: Int, $endTime: Int, $limit: Int, $cursor: String) {
    deviceLogs(organizationId: $organizationId, imei: $imei, type: $type, startTime: $startTime, endTime: $endTime, limit: $limit, cursor: $cursor) {
      events {
        ...LogEntry
      }
      pagination {
        limit
        hasMore
        nextCursor
      }
    }
  }
`;

export async function queryLogs(params: { organizationId: string; imei: string; type?: string; startTime?: number; endTime?: number; limit?: number; cursor?: string }) {
  return graphqlClient.getClient().query({
    query: GET_LOGS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
