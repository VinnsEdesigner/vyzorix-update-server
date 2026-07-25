

import { gql } from '@apollo/client';
import { MEMBERSHIP_FRAGMENT } from './graphql-membership-types';

export const UPDATE_MEMBER_ROLE = gql`
  mutation UpdateMemberRole($organizationId: ID!, $memberId: ID!, $role: OrganizationRole!) {
    updateOrganizationMember(input: { organizationId: $organizationId, memberId: $memberId, role: $role }) {
      ...MembershipFields
    }
  }
  ${MEMBERSHIP_FRAGMENT}
`;

export const REMOVE_MEMBER = gql`
  mutation RemoveMember($organizationId: ID!, $memberId: ID!) {
    removeOrganizationMember(input: { organizationId: $organizationId, memberId: $memberId }) {
      success
    }
  }
`;
