/**
 * Inbox Entry Fragment
 * 
 * Reusable GraphQL fragment for InboxEntry type.
 */

export const INBOX_ENTRY_FRAGMENT = /* GraphQL */ `
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
