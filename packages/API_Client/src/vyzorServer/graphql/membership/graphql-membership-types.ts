

import { gql } from '@apollo/client';

export const MEMBERSHIP_FRAGMENT = gql`
  fragment MembershipFields on OrganizationMembership {
    id
    role
    joinedAt
    organization {
      id
      name
    }
    operator {
      id
      email
      name
    }
  }
`;

export interface GQLMembership {
  id: string;
  role: string;
  joinedAt: string;
  organization: {
    id: string;
    name: string;
  };
  operator: {
    id: string;
    email: string;
    name: string;
  };
}
