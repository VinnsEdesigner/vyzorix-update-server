import { API_KEY_FRAGMENT, API_KEY_WITH_SECRET_FRAGMENT } from "./graphql-apikey-fragments";
import type { ApiKeyScope } from "@/domain/apikey";

export const CREATE_API_KEY =  `
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

export const UPDATE_API_KEY =  `
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

export const REVOKE_API_KEY =  `
  mutation RevokeApiKey($id: String!) {
    revokeApiKey(id: $id) {
      success
      error
    }
  }
`;

export const ROTATE_API_KEY =  `
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

import { graphqlClient } from '../_shared/graphql-client';

export async function mutateCreateApiKey(input: CreateApiKeyInput) {
  return graphqlClient.mutate({
    mutation: CREATE_API_KEY,
    variables: { input },
  });
}

export async function mutateUpdateApiKey(id: string, input: UpdateApiKeyInput) {
  return graphqlClient.mutate({
    mutation: UPDATE_API_KEY,
    variables: { id, input },
  });
}

export async function mutateRevokeApiKey(id: string) {
  return graphqlClient.mutate({
    mutation: REVOKE_API_KEY,
    variables: { id },
  });
}

export async function mutateRotateApiKey(id: string) {
  return graphqlClient.mutate({
    mutation: ROTATE_API_KEY,
    variables: { id },
  });
}
