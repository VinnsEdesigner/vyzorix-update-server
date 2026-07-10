import { API_KEY_FRAGMENT } from "./graphql-apikey-fragments";

export const GET_API_KEYS = /* GraphQL */ `
  query GetApiKeys($page: Int, $limit: Int) {
    apiKeys(page: $page, limit: $limit) {
      keys {
        ...ApiKey
      }
      pagination {
        page
        limit
        total
        totalPages
      }
      monthlyLimit
      keysCreatedThisMonth
    }
  }
  ${API_KEY_FRAGMENT}
`;

export const GET_API_KEY = /* GraphQL */ `
  query GetApiKey($id: String!) {
    apiKey(id: $id) {
      ...ApiKey
    }
  }
  ${API_KEY_FRAGMENT}
`;

export const GET_API_KEY_STATS = /* GraphQL */ `
  query GetApiKeyStats {
    apiKeyStats {
      monthlyLimit
      keysCreatedThisMonth
    }
  }
`;

import { graphqlClient } from '../_shared/graphql-client';

export async function queryApiKeys(params?: { page?: number; limit?: number }) {
  return graphqlClient.query({
    query: GET_API_KEYS,
    variables: params,
    fetchPolicy: 'network-only',
  });
}

export async function queryApiKey(id: string) {
  return graphqlClient.query({
    query: GET_API_KEY,
    variables: { id },
    fetchPolicy: 'network-only',
  });
}

export async function queryApiKeyStats() {
  return graphqlClient.query({
    query: GET_API_KEY_STATS,
    fetchPolicy: 'network-only',
  });
}
