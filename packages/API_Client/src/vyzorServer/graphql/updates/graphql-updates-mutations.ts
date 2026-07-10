// Updates GraphQL Mutations using Apollo Client
import { gql } from '@apollo/client';
import { graphqlClient } from '../_shared/graphql-client';

// ============================================================================
// Mutations
// ============================================================================

export const PUSH_UPDATE = gql`
  mutation PushUpdate($input: PushUpdateInput!) {
    pushUpdate(input: $input) {
      id
      version
      installType
      status
      initiatedBy
      initiatedAt
      deviceCount
    }
  }
`;

export const CANCEL_UPDATE = gql`
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

export const SYNC_FROM_GITHUB = gql`
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
// Mutation Functions
// ============================================================================

export interface PushUpdateInput {
  version: string;
  deviceIds: string[];
  installType: 'IMMEDIATE' | 'SCHEDULED';
  scheduledAt?: string;
}

export async function mutatePushUpdate(input: PushUpdateInput) {
  return graphqlClient.mutate({
    mutation: PUSH_UPDATE,
    variables: { input },
  });
}

export async function mutateCancelUpdate(pushId: string) {
  return graphqlClient.mutate({
    mutation: CANCEL_UPDATE,
    variables: { pushId },
  });
}

export async function mutateSyncFromGitHub() {
  return graphqlClient.mutate({
    mutation: SYNC_FROM_GITHUB,
  });
}
