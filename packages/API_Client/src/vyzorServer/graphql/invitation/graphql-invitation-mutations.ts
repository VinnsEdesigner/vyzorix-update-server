

import { gql } from '@apollo/client';
import { INVITATION_FRAGMENT } from './graphql-invitation-types';

export const CREATE_INVITATION = gql`
  mutation CreateInvitation($input: CreateInvitationInput!) {
    createInvitation(input: $input) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const CANCEL_INVITATION = gql`
  mutation CancelInvitation($organizationId: ID!, $invitationId: ID!) {
    cancelInvitation(input: { organizationId: $organizationId, invitationId: $invitationId }) {
      success
    }
  }
`;

export const RESEND_INVITATION = gql`
  mutation ResendInvitation($organizationId: ID!, $invitationId: ID!) {
    resendInvitation(input: { organizationId: $organizationId, invitationId: $invitationId }) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const APPROVE_INVITATION = gql`
  mutation ApproveInvitation($token: String!, $notes: String) {
    approveInvitation(input: { token: $token, notes: $notes }) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const REJECT_INVITATION = gql`
  mutation RejectInvitation($token: String!, $notes: String) {
    rejectInvitation(input: { token: $token, notes: $notes }) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;
