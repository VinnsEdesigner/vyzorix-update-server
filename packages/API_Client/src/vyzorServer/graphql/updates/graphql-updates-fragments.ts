export const UPDATE_VERSION_FRAGMENT =  `
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

export const PUSH_DEVICE_FRAGMENT =  `
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

export const UPDATE_PUSH_FRAGMENT =  `
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
  ${PUSH_DEVICE_FRAGMENT}
`;

export const SYNC_STATUS_FRAGMENT =  `
  fragment SyncStatus on SyncStatus {
    status
    lastSyncAt
    nextSyncAt
    versionsFound
    error
  }
`;

export const CHANGELOG_ENTRY_FRAGMENT =  `
  fragment ChangelogEntry on ChangelogEntry {
    version
    date
    type
    notes
  }
`;
