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
