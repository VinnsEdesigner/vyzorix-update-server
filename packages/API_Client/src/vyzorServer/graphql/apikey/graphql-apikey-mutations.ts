import { API_KEY_FRAGMENT, API_KEY_WITH_SECRET_FRAGMENT } from "./graphql-apikey-fragments";
import type { ApiKeyScope } from "@/domain/apikey";

export const CREATE_API_KEY = /* GraphQL */ `
  mutation CreateApiKey($input: CreateApiKeyInput!) {
    createApiKey(input: $input) {
      success
      key {
        ...ApiKeyWithSecret
      }
      error
    }
  }
  ${API_KEY_WITH_SECRET_FRAGMENT}
`;

export const UPDATE_API_KEY = /* GraphQL */ `
  mutation UpdateApiKey($id: String!, $input: UpdateApiKeyInput!) {
    updateApiKey(id: $id, input: $input) {
      success
      key {
        ...ApiKey
      }
      error
    }
  }
  ${API_KEY_FRAGMENT}
`;

export const REVOKE_API_KEY = /* GraphQL */ `
  mutation RevokeApiKey($id: String!) {
    revokeApiKey(id: $id) {
      success
      error
    }
  }
`;

export const ROTATE_API_KEY = /* GraphQL */ `
  mutation RotateApiKey($id: String!) {
    rotateApiKey(id: $id) {
      success
      key {
        ...ApiKeyWithSecret
      }
      error
    }
  }
  ${API_KEY_WITH_SECRET_FRAGMENT}
`;

export interface CreateApiKeyInput {
  name: string;
  scope: ApiKeyScope;
  expiresInDays?: number;
}

export interface UpdateApiKeyInput {
  name?: string;
  scope?: ApiKeyScope;
}
