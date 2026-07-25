import { graphqlClient } from '../_shared/graphql-client';
import { UPDATE_VERSION_FRAGMENT, CHANGELOG_ENTRY_FRAGMENT, PUSH_HISTORY_ENTRY_FRAGMENT } from './graphql-updates-fragments';

export const GET_UPDATES = `
  ${UPDATE_VERSION_FRAGMENT}
  query GetUpdates($organizationId: ID!, $status: String, $limit: Int, $offset: Int) {
    updatesVersions(organizationId: $organizationId, status: $status, limit: $limit, offset: $offset) {
      versions {
        ...UpdateVersion
      }
      pagination {
        limit
        offset
        hasMore
      }
    }
  }
`;

export const GET_UPDATES_STATUS = `
  query GetUpdatesStatus($organizationId: ID!, $deviceId: ID) {
    updatesStatus(organizationId: $organizationId, deviceId: $deviceId) {
      currentVersion
      needsUpdate
    }
  }
`;

export const GET_UPDATES_CHANGELOG = `
  ${CHANGELOG_ENTRY_FRAGMENT}
  query GetUpdatesChangelog($organizationId: ID!, $version: String, $limit: Int) {
    updatesChangelog(organizationId: $organizationId, version: $version, limit: $limit) {
      ...ChangelogEntry
    }
  }
`;

export const GET_UPDATES_HISTORY = `
  ${PUSH_HISTORY_ENTRY_FRAGMENT}
  query GetUpdatesHistory($organizationId: ID!, $status: String, $page: Int, $limit: Int) {
    updatesHistory(organizationId: $organizationId, status: $status, page: $page, limit: $limit) {
      pushes {
        ...PushHistoryEntry
      }
      pagination {
        page
        limit
        hasMore
      }
    }
  }
`;

export async function queryUpdates(params: { organizationId: string; status?: string; limit?: number; offset?: number }) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesStatus(params: { organizationId: string; deviceId?: string }) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_STATUS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesChangelog(params: { organizationId: string; version?: string; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_CHANGELOG,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesHistory(params: { organizationId: string; status?: string; page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_HISTORY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
