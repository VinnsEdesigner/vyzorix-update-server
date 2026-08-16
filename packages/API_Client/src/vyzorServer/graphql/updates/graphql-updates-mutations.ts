import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';
import {
  pushUpdateResponseFromRaw,
  cancelUpdateResponseFromRaw,
  syncResponseFromRaw,
  type PushUpdateResult,
  type CancelUpdateResult,
  type SyncUpdatesResult,
} from './graphql-updates-mappers';
import type {
  RawPushUpdateResponse,
  RawCancelUpdateResponse,
  RawSyncResponse,
} from './graphql-updates-types';

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

export async function mutatePushUpdate(params: { organizationId: string; version: string; deviceIds: string[]; installType: string; scheduledAt?: number }): Promise<PushUpdateResult> {
  const result = await graphqlClient.getClient().mutate<{ pushUpdate: RawPushUpdateResponse }>({
    mutation: PUSH_UPDATE,
    variables: params,
  });
  const data = result.data?.pushUpdate;
  if (!data) throw new Error('pushUpdate mutation returned no data');
  return pushUpdateResponseFromRaw(data);
}

export async function mutateCancelUpdate(params: { organizationId: string; id: string }): Promise<CancelUpdateResult> {
  const result = await graphqlClient.getClient().mutate<{ cancelUpdate: RawCancelUpdateResponse }>({
    mutation: CANCEL_UPDATE,
    variables: params,
  });
  const data = result.data?.cancelUpdate;
  if (!data) throw new Error('cancelUpdate mutation returned no data');
  return cancelUpdateResponseFromRaw(data);
}

export async function mutateSyncUpdates(_params?: { organizationId?: string }): Promise<SyncUpdatesResult> {
  // The server `syncUpdates` mutation accepts no arguments; organization scoping
  // is enforced via the org-scoped GraphQL endpoint URL
  // (/:orgId/graphql) selected by `graphqlClient.setOrganization`.
  const result = await graphqlClient.getClient().mutate<{ syncUpdates: RawSyncResponse }>({
    mutation: SYNC_UPDATES,
  });
  const data = result.data?.syncUpdates;
  if (!data) throw new Error('syncUpdates mutation returned no data');
  return syncResponseFromRaw(data);
}
