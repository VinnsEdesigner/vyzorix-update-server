import { graphqlClient } from '../_shared/graphql-client';

export const UPDATE_MEMBER_ROLE = `
  mutation UpdateMemberRole($organizationId: ID!, $memberId: ID!, $role: OrgRole!) {
    updateMemberRole(organizationId: $organizationId, memberId: $memberId, role: $role) {
      id
      organizationId
      operatorId
      role
      lifecycle
      invitedAt
      joinedAt
      removedAt
      suspendedAt
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

export const REMOVE_MEMBER = `
  mutation RemoveMember($organizationId: ID!, $memberId: ID!) {
    removeMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const SUSPEND_MEMBER = `
  mutation SuspendMember($organizationId: ID!, $memberId: ID!) {
    suspendMember(organizationId: $organizationId, memberId: $memberId)
  }
`;

export const REINSTATE_MEMBER = `
  mutation ReinstateMember($organizationId: ID!, $memberId: ID!) {
    reinstateMember(organizationId: $organizationId, memberId: $memberId)
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

export async function mutateSuspendMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: SUSPEND_MEMBER,
    variables: params,
  });
}

export async function mutateReinstateMember(params: { organizationId: string; memberId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REINSTATE_MEMBER,
    variables: params,
  });
}
