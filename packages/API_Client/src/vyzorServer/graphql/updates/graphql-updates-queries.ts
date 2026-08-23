import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';
import {
  UPDATE_VERSION_FRAGMENT,
  CHANGELOG_ENTRY_FRAGMENT,
  PUSH_HISTORY_ENTRY_FRAGMENT,
} from './graphql-updates-fragments';
import {
  versionListResultFromRaw,
  changelogResultFromRaw,
  updateHistoryResultFromRaw,
  updateStatusFromRaw,
  type GraphQLUpdateStatusResult,
} from './graphql-updates-mappers';
import type {
  RawVersionConnection,
  RawChangelogConnection,
  RawUpdateHistoryConnection,
  RawUpdateStatusResponse,
} from './graphql-updates-types';
import type { VersionListResult, UpdateHistoryResult, ChangelogEntry } from '../../../domain/updates';

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
  ${UPDATE_VERSION_FRAGMENT}
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
        ...UpdateVersion
      }
      device {
        currentVersion
        needsUpdate
      }
      apkFilename
      sha256
    }
  }
  ${UPDATE_VERSION_FRAGMENT}
`;

export const GET_UPDATES_CHANGELOG = gql`
  query GetUpdatesChangelog($organizationId: ID!, $version: String, $limit: Int) {
    updatesChangelog(organizationId: $organizationId, version: $version, limit: $limit) {
      ...ChangelogEntry
    }
  }
  ${CHANGELOG_ENTRY_FRAGMENT}
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
  ${PUSH_HISTORY_ENTRY_FRAGMENT}
`;

export async function queryUpdates(params: { organizationId: string; status?: string; limit?: number; offset?: number }): Promise<VersionListResult> {
  const result = await graphqlClient.getClient().query<{ updatesVersions: RawVersionConnection }>({
    query: GET_UPDATES,
    variables: params,
    fetchPolicy: 'network-only',
  });
  return versionListResultFromRaw(result.data.updatesVersions);
}

export async function queryUpdatesStatus(params: { organizationId: string; deviceId?: string }): Promise<GraphQLUpdateStatusResult> {
  const result = await graphqlClient.getClient().query<{ updatesStatus: RawUpdateStatusResponse }>({
    query: GET_UPDATES_STATUS,
    variables: params,
    fetchPolicy: 'network-only',
  });
  return updateStatusFromRaw(result.data.updatesStatus);
}

export async function queryUpdatesChangelog(params: { organizationId: string; version?: string; limit?: number }): Promise<ChangelogEntry[]> {
  const result = await graphqlClient.getClient().query<{ updatesChangelog: RawChangelogConnection['changelog'] }>({
    query: GET_UPDATES_CHANGELOG,
    variables: params,
    fetchPolicy: 'network-only',
  });
  return changelogResultFromRaw({ changelog: result.data.updatesChangelog }).changelog;
}

export async function queryUpdatesHistory(params: { organizationId: string; status?: string; page?: number; limit?: number }): Promise<UpdateHistoryResult> {
  const result = await graphqlClient.getClient().query<{ updatesHistory: RawUpdateHistoryConnection }>({
    query: GET_UPDATES_HISTORY,
    variables: params,
    fetchPolicy: 'network-only',
  });
  return updateHistoryResultFromRaw(result.data.updatesHistory);
}
