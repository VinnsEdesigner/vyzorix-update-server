import { graphqlClient } from '../_shared/graphql-client';
import { MEMBERSHIP_FRAGMENT } from './graphql-membership-types';

export const UPDATE_MEMBER_ROLE = `
  ${MEMBERSHIP_FRAGMENT}
  mutation UpdateMemberRole($organizationId: ID!, $memberId: ID!, $role: OrgRole!) {
    updateMemberRole(organizationId: $organizationId, memberId: $memberId, role: $role) {
      ...MembershipFields
    }
  }
`;

export const REMOVE_MEMBER = `
  mutation RemoveMember($organizationId: ID!, $memberId: ID!) {
    removeMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export async function mutateUpdateMemberRole(params: { organizationId: string; memberId: string; role: string }) {
  return graphqlClient.getClient().mutate({
    mutation: UPDATE_MEMBER_ROLE,
    variables: params,
  });
}

export async function mutateRemoveMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REMOVE_MEMBER,
    variables: params,
  });
}
