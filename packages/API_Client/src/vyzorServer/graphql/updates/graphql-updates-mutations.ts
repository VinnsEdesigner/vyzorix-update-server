import { graphqlClient } from '../_shared/graphql-client';

export const PUSH_UPDATE = `
  mutation PushUpdate($organizationId: ID!, $version: String!, $deviceIds: [ID!]!, $installType: String!, $scheduledAt: Int) {
    pushUpdate(organizationId: $organizationId, version: $version, deviceIds: $deviceIds, installType: $installType, scheduledAt: $scheduledAt) {
      pushId
      version
      installType
      scheduledAt
      status
      initiatedBy
      initiatedAt
      deviceCount
    }
  }
`;

export const CANCEL_UPDATE = `
  mutation CancelUpdate($organizationId: ID!, $id: ID!) {
    cancelUpdate(organizationId: $organizationId, id: $id) {
      id
      status
      cancelledAt
      cancelledBy
    }
  }
`;

export const SYNC_UPDATES = `
  mutation SyncUpdates {
    syncUpdates {
      status
      startedAt
      message
      versionsFound
    }
  }
`;

export async function mutatePushUpdate(params: { organizationId: string; version: string; deviceIds: string[]; installType: string; scheduledAt?: number }) {
  return graphqlClient.getClient().mutate({
    mutation: PUSH_UPDATE,
    variables: params,
  });
}

export async function mutateCancelUpdate(params: { organizationId: string; id: string }) {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_UPDATE,
    variables: params,
  });
}

export async function mutateSyncUpdates() {
  return graphqlClient.getClient().mutate({
    mutation: SYNC_UPDATES,
  });
}
