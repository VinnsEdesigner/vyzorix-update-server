import { graphqlClient } from '../_shared/graphql-client';

export const ACK_INBOX = `
  mutation AckInbox($imei: String!, $action: AckAction!, $notes: String) {
    ackInbox(imei: $imei, action: $action, notes: $notes) {
      success
      status
      error
    }
  }
`;

export const DEREGISTER_DEVICE = `
  mutation DeregisterDevice($imei: String!, $hard: Boolean) {
    deregisterDevice(imei: $imei, hard: $hard) {
      success
      status
      error
    }
  }
`;

export async function mutateAckInbox(params: { imei: string; action: 'APPROVE' | 'REJECT'; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: ACK_INBOX,
    variables: params,
  });
}

export async function mutateDeregisterDevice(params: { imei: string; hard?: boolean }) {
  return graphqlClient.getClient().mutate({
    mutation: DEREGISTER_DEVICE,
    variables: params,
  });
}
