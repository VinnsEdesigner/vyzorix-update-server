import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_LOGS = gql`
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

export async function queryLogs(params: { organizationId: string; imei: string; type?: string; startTime?: number; endTime?: number; limit?: number; cursor?: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_LOGS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
