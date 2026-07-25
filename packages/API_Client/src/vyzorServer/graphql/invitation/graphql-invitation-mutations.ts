import { graphqlClient } from '../_shared/graphql-client';
import { INVITATION_FRAGMENT } from './graphql-invitation-types';

export const INVITE_MEMBER = `
  ${INVITATION_FRAGMENT}
  mutation InviteMember($organizationId: ID!, $email: String!, $role: OrgRole!, $notes: String) {
    inviteMember(organizationId: $organizationId, email: $email, role: $role, notes: $notes) {
      ...InvitationFields
    }
  }
`;

export const CANCEL_INVITATION = `
  mutation CancelInvitation($id: ID!) {
    cancelInvitation(id: $id)
  }
`;

export const ACCEPT_INVITATION = `
  ${INVITATION_FRAGMENT}
  mutation AcceptInvitation($token: String!, $notes: String) {
    acceptInvitation(token: $token, notes: $notes) {
      ...InvitationFields
    }
  }
`;

export const REJECT_INVITATION = `
  mutation RejectInvitation($token: String!, $notes: String) {
    rejectInvitation(token: $token, notes: $notes)
  }
`;

export async function mutateInviteMember(params: { organizationId: string; email: string; role: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: INVITE_MEMBER,
    variables: params,
  });
}

export async function mutateCancelInvitation(params: { id: string }) {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_INVITATION,
    variables: params,
  });
}

export async function mutateAcceptInvitation(params: { token: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: ACCEPT_INVITATION,
    variables: params,
  });
}

export async function mutateRejectInvitation(params: { token: string; notes?: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REJECT_INVITATION,
    variables: params,
  });
}
