export const API_KEY_FRAGMENT = /* GraphQL */ `
  fragment ApiKey on ApiKey {
    id
    operatorId
    name
    keyPrefix
    scope
    expiresAt
    isActive
    requestCount
    lastRequestAt
    createdAt
    updatedAt
    revokedAt
  }
`;

export const API_KEY_WITH_SECRET_FRAGMENT = /* GraphQL */ `
  fragment ApiKeyWithSecret on ApiKey {
    id
    operatorId
    name
    keyPrefix
    apiKey
    scope
    expiresAt
    isActive
    requestCount
    lastRequestAt
    createdAt
    updatedAt
    revokedAt
  }
`;
