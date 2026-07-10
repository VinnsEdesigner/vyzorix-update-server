import { UPDATE_VERSION_FRAGMENT, UPDATE_PUSH_FRAGMENT, SYNC_STATUS_FRAGMENT, CHANGELOG_ENTRY_FRAGMENT } from "./graphql-updates-fragments";

export const GET_UPDATE_STATUS = /* GraphQL */ `
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

export const GET_VERSIONS = /* GraphQL */ `
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

export const GET_CHANGELOG = /* GraphQL */ `
  query GetChangelog($version: String) {
    updatesChangelog(version: $version) {
      changelog {
        ...ChangelogEntry
      }
    }
  }
  ${CHANGELOG_ENTRY_FRAGMENT}
`;

export const GET_UPDATE_HISTORY = /* GraphQL */ `
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

export const GET_UPDATE_PUSH = /* GraphQL */ `
  query GetUpdatePush($pushId: String!) {
    updatesHistoryDetail(pushId: $pushId) {
      ...UpdatePush
    }
  }
  ${UPDATE_PUSH_FRAGMENT}
`;

export const GET_SYNC_STATUS = /* GraphQL */ `
  query GetSyncStatus {
    updatesSyncStatus {
      ...SyncStatus
    }
  }
  ${SYNC_STATUS_FRAGMENT}
`;