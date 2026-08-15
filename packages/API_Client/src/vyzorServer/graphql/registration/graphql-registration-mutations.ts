import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';
import type { AcknowledgeAction } from '../../../domain/registration';

export const ACK_INBOX = gql`
  mutation AckInbox($imei: String!, $action: AckAction!, $notes: String) {
    ackInbox(imei: $imei, action: $action, notes: $notes) {
      success
      status
      error
    }
  }
`;

export const DEREGISTER_DEVICE = gql`
  mutation DeregisterDevice($imei: String!, $hard: Boolean) {
    deregisterDevice(imei: $imei, hard: $hard) {
      success
      status
      error
    }
  }
`;

export async function mutateAckInbox(params: { imei: string; action: AcknowledgeAction; notes?: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: ACK_INBOX,
    variables: { ...params, action: params.action.toUpperCase() as 'APPROVE' | 'REJECT' | 'ACKNOWLEDGE' },
  });
}

export async function mutateDeregisterDevice(params: { imei: string; hard?: boolean }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: DEREGISTER_DEVICE,
    variables: params,
  });
}
