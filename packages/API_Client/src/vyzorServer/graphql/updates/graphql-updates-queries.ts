import { gql } from '@apollo/client';
import { graphqlClient } from '../_shared/graphql-client';

export const GET_UPDATES = gql`
  query GetUpdates($organizationId: ID!, $status: String, $limit: Int, $offset: Int) {
    updatesVersions(organizationId: $organizationId, status: $status, limit: $limit, offset: $offset) {
      versions {
        id
        version
        apkFilename
        apkSize
        sha256
        releasedAt
        releaseNotes
        releaseType
        isLatest
      }
    }
  }
`;

export async function queryUpdates(organizationId: string, status?: string, limit?: number, offset?: number) {
  return graphqlClient.getClient().query({
    query: GET_UPDATES,
    variables: { organizationId, status, limit, offset },
    fetchPolicy: 'network-only',
  });
}
