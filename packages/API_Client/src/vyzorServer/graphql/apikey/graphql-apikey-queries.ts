import { graphqlClient } from '../_shared/graphql-client';

export const GET_API_KEYS = `
  query GetApiKeys($organizationId: ID!) {
    apiKeys(organizationId: $organizationId) {
      keys {
        id
        name
        key_preview
        created_at
        last_used_at
        expires_at
      }
    }
  }
`;

export async function queryApiKeys(organizationId: string) {
  return graphqlClient.getClient().query({
    query: GET_API_KEYS,
    variables: { organizationId },
    fetchPolicy: 'network-only',
  });
}
