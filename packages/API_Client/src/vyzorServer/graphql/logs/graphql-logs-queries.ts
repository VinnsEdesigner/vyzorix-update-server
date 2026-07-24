import { graphqlClient } from '../_shared/graphql-client';

export const GET_LOGS = `
  query GetLogs($organizationId: ID!, $deviceId: ID!, $limit: Int, $offset: Int) {
    logs(organizationId: $organizationId, deviceId: $deviceId, limit: $limit, offset: $offset) {
      entries {
        id
        device_id
        level
        message
        metadata
        timestamp
      }
      total
    }
  }
`;

export const GET_LOG_ENTRY = `
  query GetLogEntry($organizationId: ID!, $deviceId: ID!, $logId: ID!) {
    logEntry(organizationId: $organizationId, deviceId: $deviceId, logId: $logId) {
      id
      device_id
      level
      message
      metadata
      timestamp
    }
  }
`;

export async function queryLogs(params: { organizationId: string; deviceId: string; limit?: number; offset?: number }) {
  return graphqlClient.getClient().query({
    query: GET_LOGS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryLogEntry(params: { organizationId: string; deviceId: string; logId: string }) {
  return graphqlClient.getClient().query({
    query: GET_LOG_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
