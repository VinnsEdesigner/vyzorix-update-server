import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const INBOX_ENTRY_FRAGMENT = gql`
  fragment InboxEntry on InboxEntry {
    id
    imei
    model
    manufacturer
    osVersion
    appVersion
    firebaseInstallId
    status
    notes
    operatorId
    createdAt
    approvedAt
    rejectedAt
  }
`;

export const GET_INBOX_ENTRIES = gql`
  query GetInboxEntries($organizationId: ID!, $status: String, $page: Int, $limit: Int) {
    inbox(organizationId: $organizationId, status: $status, page: $page, limit: $limit) {
      requests {
        ...InboxEntry
      }
      pagination {
        total
        limit
        offset
        hasMore
      }
    }
  }
`;

export const GET_INBOX_ENTRY = gql`
  query GetInboxEntry($organizationId: ID!, $imei: String!) {
    inboxEntry(organizationId: $organizationId, imei: $imei) {
      ...InboxEntry
    }
  }
`;

export async function queryInboxEntries(params: { organizationId: string; status?: string; page?: number; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRIES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInboxEntry(params: { organizationId: string; imei: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
