import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const INVITE_MEMBER = gql`
  mutation InviteMember($organizationId: ID!, $email: String!, $role: OrgRole!, $notes: String) {
    inviteMember(organizationId: $organizationId, email: $email, role: $role, notes: $notes) {
      id
      organizationId
      email
      role
      status
      token
      notes
      invitedAt
      respondedAt
      expiresAt
      inviter {
        id
        name
      }
      organization {
        id
        name
      }
    }
  }
`;

export const CANCEL_INVITATION = gql`
  mutation CancelInvitation($id: ID!) {
    cancelInvitation(id: $id)
  }
`;

export const ACCEPT_INVITATION = gql`
  mutation AcceptInvitation($token: String!, $notes: String) {
    acceptInvitation(token: $token, notes: $notes) {
      id
      organizationId
      operatorId
      role
      lifecycle
      invitedAt
      joinedAt
      operator {
        id
        email
        name
      }
      organization {
        id
        name
      }
    }
  }
`;

export const REJECT_INVITATION = gql`
  mutation RejectInvitation($token: String!, $notes: String) {
    rejectInvitation(token: $token, notes: $notes)
  }
`;

export async function mutateInviteMember(params: { organizationId: string; email: string; role: string; notes?: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: INVITE_MEMBER,
    variables: params,
  });
}

export async function mutateCancelInvitation(params: { id: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: CANCEL_INVITATION,
    variables: params,
  });
}

export async function mutateAcceptInvitation(params: { token: string; notes?: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: ACCEPT_INVITATION,
    variables: params,
  });
}

export async function mutateRejectInvitation(params: { token: string; notes?: string }): Promise<unknown> {
  return graphqlClient.getClient().mutate({
    mutation: REJECT_INVITATION,
    variables: params,
  });
}
