import { graphqlClient } from '../_shared/graphql-client';

export const GENERATE_API_KEY = `
  mutation GenerateApiKey($organizationId: ID!, $name: String!) {
    generateApiKey(organizationId: $organizationId, name: $name) {
      id
      name
      key
      created_at
    }
  }
`;

export const REVOKE_API_KEY = `
  mutation RevokeApiKey($organizationId: ID!, $keyId: ID!) {
    revokeApiKey(organizationId: $organizationId, keyId: $keyId) {
      success
      message
    }
  }
`;

export async function mutateGenerateApiKey(params: { organizationId: string; name: string }) {
  return graphqlClient.getClient().mutate({
    mutation: GENERATE_API_KEY,
    variables: params,
  });
}

export async function mutateRevokeApiKey(params: { organizationId: string; keyId: string }) {
  return graphqlClient.getClient().mutate({
    mutation: REVOKE_API_KEY,
    variables: params,
  });
}
