import { graphqlClient } from '../_shared/graphql-client';
import { INVITATION_FRAGMENT } from './graphql-invitation-types';

export const GET_MY_INVITATIONS = `
  query GetMyInvitations {
    myInvitations {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;

export const GET_ORGANIZATION_INVITATIONS = `
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
  ${INVITATION_FRAGMENT}
`;

export const GET_INVITATION_BY_TOKEN = gql`
  query GetInvitationByToken($token: String!) {
    invitationByToken(token: $token) {
      ...InvitationFields
    }
  }
  ${INVITATION_FRAGMENT}
`;

export async function queryMyInvitations() {
  return graphqlClient.getClient().query({
    query: GET_MY_INVITATIONS,
    fetchPolicy: 'network-only',
  });
}

export async function queryOrganizationInvitations(params: { organizationId: string; page?: number; limit?: number }) {
  return graphqlClient.getClient().query({
    query: GET_ORGANIZATION_INVITATIONS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryInvitationByToken(token: string) {
  return graphqlClient.getClient().query({
    query: GET_INVITATION_BY_TOKEN,
    variables: { token },
    fetchPolicy: 'network-only',
  });
}
