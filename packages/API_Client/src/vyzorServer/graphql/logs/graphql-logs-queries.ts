import { graphqlClient } from '../_shared/graphql-client';

export const GET_LOGS = `
  query GetLogs($organizationId: ID!, $imei: ID!, $type: String, $startTime: Int, $endTime: Int, $limit: Int, $cursor: String) {
    deviceLogs(organizationId: $organizationId, imei: $imei, type: $type, startTime: $startTime, endTime: $endTime, limit: $limit, cursor: $cursor) {
      entries {
        id
        deviceImei
        level
        message
        metadata
        timestamp
      }
      hasMore
      nextCursor
    }
  }
`;

export const GET_LOG_ENTRY = `
  query GetLogEntry($organizationId: ID!, $imei: ID!, $logId: ID!) {
    deviceLog(organizationId: $organizationId, imei: $imei, logId: $logId) {
      id
      deviceImei
      level
      message
      metadata
      timestamp
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

export async function queryLogEntry(params: { organizationId: string; imei: string; logId: string }) {
  return graphqlClient.getClient().query({
    query: GET_LOG_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
