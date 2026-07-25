import { graphqlClient } from '../_shared/graphql-client';

export const INBOX_ENTRY_FRAGMENT = `
  fragment InboxEntry on InboxEntry {
    id
    imei
    deviceName
    model
    manufacturer
    osVersion
    appVersion
    firmware
    securityPatch
    buildId
    fcmToken
    firebaseInstallId
    status
    receivedAt
    updatedAt
    acknowledgedAt
    approvedAt
    rejectedAt
    notes
  }
`;

export const GET_INBOX_ENTRIES = `
  query GetInboxEntries($organizationId: ID!, $status: String, $page: Int, $limit: Int) {
    inbox(organizationId: $organizationId, status: $status, page: $page, limit: $limit) {
      requests {
        ...InboxEntry
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${INBOX_ENTRY_FRAGMENT}
`;

export const GET_INBOX_ENTRY = `
  query GetInboxEntry($organizationId: ID!, $imei: String!) {
    inboxEntry(organizationId: $organizationId, imei: $imei) {
      ...InboxEntry
    }
  }
  ${INBOX_ENTRY_FRAGMENT}
`;

export async function queryInboxEntries(params: { organizationId: string; status?: string; page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRIES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInboxEntry(params: { organizationId: string; imei: string }) {
  return graphqlClient.getClient().query({
    query: GET_INBOX_ENTRY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
