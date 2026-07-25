import { graphqlClient } from '../_shared/graphql-client';
import { gql } from '@apollo/client';

export const GET_MEMBERSHIP_DETAILS = gql`
  query GetMembershipDetails($membershipId: ID!) {
    membership(membershipId: $membershipId) {
      ...MembershipFields
    }
  }
`;

export async function queryMembershipDetails(params: { membershipId: string }): Promise<unknown> {
  return graphqlClient.getClient().query({
    query: GET_MEMBERSHIP_DETAILS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}
