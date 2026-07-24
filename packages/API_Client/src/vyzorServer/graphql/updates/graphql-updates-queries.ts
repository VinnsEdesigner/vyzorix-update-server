import { gql } from '@apollo/client';
import { graphqlClient } from '../_shared/graphql-client';

export const GET_UPDATES = gql`
  query GetUpdates($organizationId: ID!) {
    updates(organizationId: $organizationId) {
      versions {
        id
        version
        apk_filename
        apk_size
        sha256
        release_date
        release_notes
        release_type
        is_latest
      }
    }
  }
`;

export async function queryUpdates(organizationId: string) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES,
    variables: { organizationId },
    fetchPolicy: 'network-only',
  });
}
