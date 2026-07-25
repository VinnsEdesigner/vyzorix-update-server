import { gql } from '@apollo/client';

export const UPDATE_VERSION_FRAGMENT = gql`
  fragment UpdateVersion on UpdateVersion {
    id
    version
    releaseType
    releaseNotes
    apkFilename
    apkSize
    sha256
    releasedAt
    createdAt
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

export const PUSH_HISTORY_ENTRY_FRAGMENT = gql`
  fragment PushHistoryEntry on PushHistoryEntry {
    id
    version
    installType
    status
    initiatedBy
    initiatedAt
    completedAt
    deviceCount
    pending
    acknowledged
    failed
  }
`;
