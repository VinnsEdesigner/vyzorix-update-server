
import { gql } from '@apollo/client';
import { graphqlClient } from '../_shared/graphql-client';





export const UPDATE_VERSION_FRAGMENT = gql`
  fragment UpdateVersion on UpdateVersion {
    id
    version
    apkFilename
    apkSize
    sha256
    releaseDate
    releaseNotes
    releaseType
    isLatest
  }
`;

export const PUSH_DEVICE_FRAGMENT = gql`
  fragment PushDevice on PushDevice {
    id
    deviceId
    deviceName
    status
    sentAt
    acknowledgedAt
    error
  }
`;

export const UPDATE_PUSH_FRAGMENT = gql`
  fragment UpdatePush on UpdatePush {
    id
    version
    installType
    scheduledAt
    status
    initiatedBy
    initiatedAt
    completedAt
    cancelledAt
    deviceCount
    devices {
      ...PushDevice
    }
  }
  ${PUSH_DEVICE_FRAGMENT}
`;

export const SYNC_STATUS_FRAGMENT = gql`
  fragment SyncStatus on SyncStatus {
    status
    lastSyncAt
    nextSyncAt
    versionsFound
    error
  }
`;

export const CHANGELOG_ENTRY_FRAGMENT = gql`
  fragment ChangelogEntry on ChangelogEntry {
    version
    date
    type
    notes
  }
`;





export const GET_UPDATE_STATUS = gql`
  query GetUpdateStatus {
    updatesStatus {
      sync {
        ...SyncStatus
      }
      latest {
        ...UpdateVersion
      }
      device {
        currentVersion
        needsUpdate
      }
    }
  }
  ${SYNC_STATUS_FRAGMENT}
  ${UPDATE_VERSION_FRAGMENT}
`;

export const GET_VERSIONS = gql`
  query GetVersions($status: String, $page: Int, $limit: Int) {
    updatesVersions(status: $status, page: $page, limit: $limit) {
      versions {
        ...UpdateVersion
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${UPDATE_VERSION_FRAGMENT}
`;

export const GET_CHANGELOG = gql`
  query GetChangelog($version: String) {
    updatesChangelog(version: $version) {
      changelog {
        ...ChangelogEntry
      }
    }
  }
  ${CHANGELOG_ENTRY_FRAGMENT}
`;

export const GET_UPDATE_HISTORY = gql`
  query GetUpdateHistory($status: String, $page: Int, $limit: Int) {
    updatesHistory(status: $status, page: $page, limit: $limit) {
      pushes {
        ...UpdatePush
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${UPDATE_PUSH_FRAGMENT}
`;

export const GET_UPDATE_PUSH = gql`
  query GetUpdatePush($pushId: String!) {
    updatesHistoryDetail(pushId: $pushId) {
      ...UpdatePush
    }
  }
  ${UPDATE_PUSH_FRAGMENT}
`;

export const GET_SYNC_STATUS = gql`
  query GetSyncStatus {
    updatesSyncStatus {
      ...SyncStatus
    }
  }
  ${SYNC_STATUS_FRAGMENT}
`;





export async function queryUpdateStatus() {
  return graphqlClient.query({
    query: GET_UPDATE_STATUS,
    fetchPolicy: 'network-only',
  });
}

export async function queryVersions(params?: { status?: string; page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_VERSIONS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryChangelog(params?: { version?: string }) {
  return graphqlClient.query({
    query: GET_CHANGELOG,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdateHistory(params?: { status?: string; page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_UPDATE_HISTORY,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryUpdatePush(pushId: string) {
  return graphqlClient.query({
    query: GET_UPDATE_PUSH,
    variables: { pushId },
    fetchPolicy: 'network-only',
  });
}

export async function querySyncStatus() {
  return graphqlClient.query({
    query: GET_SYNC_STATUS,
    fetchPolicy: 'network-only',
  });
}
