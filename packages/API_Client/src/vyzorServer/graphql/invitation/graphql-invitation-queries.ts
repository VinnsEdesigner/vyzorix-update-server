

import { gql } from '@apollo/client';
import { INVITATION_FRAGMENT } from './graphql-invitation-types';

export const GET_MY_INVITATIONS = gql`
  query GetMyInvitations {
    me {
      invitations {
        ...InvitationFields
      }
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const GET_ORGANIZATION_INVITATIONS = gql`
  query GetOrganizationInvitations($organizationId: ID!) {
    organization(id: $organizationId) {
      invitations {
        ...InvitationFields
      }
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const GET_INVITATION_BY_TOKEN = gql`
  query GetInvitationByToken($token: String!) {
    invitation(token: $token) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;
