import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const PUSH_UPDATE = gql`
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

export const CANCEL_UPDATE = gql`
  mutation CancelUpdate($organizationId: ID!, $id: ID!) {
    cancelUpdate(organizationId: $organizationId, id: $id) {
      id
      status
      cancelledAt
      cancelledBy
    }
  }
`;

export const SYNC_UPDATES = gql`
  mutation SyncUpdates {
    syncUpdates {
      status
      startedAt
      message
      versionsFound
    }
  }
`;

export async function mutatePushUpdate(params: { organizationId: string; version: string; deviceIds: string[]; installType: string; scheduledAt?: number }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: PUSH_UPDATE,
    variables: params,
  });
}

export async function mutateCancelUpdate(params: { organizationId: string; id: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_UPDATE,
    variables: params,
  });
}

export async function mutateSyncUpdates(): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: SYNC_UPDATES,
  });
}
