

import { gql } from '@apollo/client';
import { MEMBERSHIP_FRAGMENT } from './graphql-membership-types';

export const GET_MY_MEMBERSHIPS = gql`
  query GetMyMemberships {
    me {
      memberships {
        ...MembershipFields
      }
    }
  }
  ${MEMBERSHIP_FRAGMENT}
`;

export const GET_ORGANIZATION_MEMBERS = gql`
  query GetOrganizationMembers($organizationId: ID!) {
    organization(id: $organizationId) {
      members {
        ...MembershipFields
      }
    }
  }
  ${MEMBERSHIP_FRAGMENT}
`;
