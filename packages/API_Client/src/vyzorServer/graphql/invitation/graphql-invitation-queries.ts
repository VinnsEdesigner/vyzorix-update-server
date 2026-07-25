import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_MY_INVITATIONS = gql`
  query GetMyInvitations {
    myInvitations {
      ...InvitationFields
    }
  }
`;

export const GET_ORGANIZATION_INVITATIONS = gql`
  query GetOrganizationInvitations($organizationId: ID!, $page: Int, $limit: Int) {
    organizationInvitations(organizationId: $organizationId, page: $page, limit: $limit) {
      items {
        ...InvitationFields
      }
      pagination {
        page
        limit
        total
        totalPages
      }
    }
  }
`;

export async function queryMyInvitations(): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_MY_INVITATIONS,
    fetchPolicy: 'network-only',
  });
}

export async function queryOrganizationInvitations(params: { organizationId: string; page?: number; limit?: number }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_ORGANIZATION_INVITATIONS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
