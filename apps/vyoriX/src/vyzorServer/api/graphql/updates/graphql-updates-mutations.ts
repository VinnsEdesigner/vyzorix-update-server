/**
 * Updates GraphQL Mutations
 * 
 * GraphQL mutation definitions for updates.
 * Based on SERVER_BACKEND_UPDATES_API.md.
 */

// ============================================================================
// Mutation Definitions
// ============================================================================

/**
 * Push update to devices
 */
export const PUSH_UPDATE = /* GraphQL */ `
  mutation PushUpdate($input: PushUpdateInput!) {
    pushUpdate(input: $input) {
      id
      version
      deviceIds
      installType
      status
      initiatedBy
      initiatedAt
      devices {
        total
        pending
        sent
        acknowledged
        failed
      }
    }
  }
`;

/**
 * Cancel an update push
 */
export const CANCEL_UPDATE = /* GraphQL */ `
  mutation CancelUpdate($pushId: String!) {
    cancelUpdate(pushId: $pushId) {
      id
      version
      status
      completedAt
      cancelledAt
    }
  }
`;

/**
 * Sync versions from GitHub
 */
export const SYNC_FROM_GITHUB = /* GraphQL */ `
  mutation SyncFromGitHub {
    syncFromGitHub {
      status
      startedAt
      versionsFound
      message
    }
  }
`;

// ============================================================================
// Input Types (for reference)
// ============================================================================

/**
 * PushUpdateInput:
 * {
 *   version: String!           // Version to push
 *   deviceIds: [String!]!     // List of device IMEIs
 *   installType: InstallType!  // IMMEDIATE or SCHEDULED
 *   scheduledAt: DateTime     // Required if SCHEDULED
 * }
 * 
 * InstallType: IMMEDIATE | SCHEDULED
 */