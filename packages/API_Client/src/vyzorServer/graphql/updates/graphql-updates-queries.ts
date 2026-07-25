import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_UPDATES = gql`
  query GetUpdates($organizationId: ID!, $status: String, $limit: Int, $offset: Int) {
    updatesVersions(organizationId: $organizationId, status: $status, limit: $limit, offset: $offset) {
      versions {
        ...UpdateVersion
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

export const GET_UPDATES_STATUS = gql`
  query GetUpdatesStatus($organizationId: ID!, $deviceId: ID) {
    updatesStatus(organizationId: $organizationId, deviceId: $deviceId) {
      version
      sync {
        status
        lastSyncAt
        nextSyncAt
        versionsFound
        error
      }
      latest {
        id
        version
        apkFilename
        sha256
      }
      device {
        currentVersion
        needsUpdate
      }
      apkFilename
      sha256
    }
  }
`;

export const GET_UPDATES_CHANGELOG = gql`
  query GetUpdatesChangelog($organizationId: ID!, $version: String, $limit: Int) {
    updatesChangelog(organizationId: $organizationId, version: $version, limit: $limit) {
      ...ChangelogEntry
    }
  }
`;

export const GET_UPDATES_HISTORY = gql`
  query GetUpdatesHistory($organizationId: ID!, $status: String, $page: Int, $limit: Int) {
    updatesHistory(organizationId: $organizationId, status: $status, page: $page, limit: $limit) {
      pushes {
        ...PushHistoryEntry
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

export async function queryUpdates(params: { organizationId: string; status?: string; limit?: number; offset?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_UPDATES,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesStatus(params: { organizationId: string; deviceId?: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_STATUS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesChangelog(params: { organizationId: string; version?: string; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_CHANGELOG,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatesHistory(params: { organizationId: string; status?: string; page?: number; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_UPDATES_HISTORY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
