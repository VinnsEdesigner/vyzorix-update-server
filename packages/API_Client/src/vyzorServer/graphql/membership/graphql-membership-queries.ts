import { graphqlClient } from '../_shared/graphql-client';
import { MEMBERSHIP_FRAGMENT } from './graphql-membership-types';

export const GET_MY_MEMBERSHIPS = `
  query GetMyMemberships($page: Int, $limit: Int) {
    myMemberships(page: $page, limit: $limit) {
      items {
        ...MembershipFields
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${MEMBERSHIP_FRAGMENT}
`;

export const GET_ORGANIZATION_MEMBERS = `
  query GetOrganizationMembers($organizationId: ID!, $page: Int, $limit: Int) {
    organizationMembers(organizationId: $organizationId, page: $page, limit: $limit) {
      items {
        ...MembershipFields
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
  ${MEMBERSHIP_FRAGMENT}
`;

export async function queryMyMemberships(params?: { page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_MY_MEMBERSHIPS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryOrganizationMembers(params: { organizationId: string; page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_ORGANIZATION_MEMBERS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
