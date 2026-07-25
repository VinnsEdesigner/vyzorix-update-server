export const INBOX_ENTRY_FRAGMENT = `
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
