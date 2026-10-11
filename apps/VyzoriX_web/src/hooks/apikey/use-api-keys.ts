import {
  useGetAuthApiKeys,
  usePostAuthApiKeys,
  useGetAuthApiKeysKeyId,
  usePatchAuthApiKeysKeyId,
  useDeleteAuthApiKeysKeyId,
  usePostAuthApiKeysKeyIdRotate,
} from '@/generated-rq/api-keys/operator-api-keys';

export interface ApiKeyParams {
  page?: number;
  limit?: number;
}

export function useApiKeys(params?: ApiKeyParams) {
  return useGetAuthApiKeys({ page: params?.page, limit: params?.limit });
}

export function useApiKey(id: string | undefined) {
  return useGetAuthApiKeysKeyId(id ?? '', { query: { enabled: id !== undefined } });
}

export function useCreateApiKey() {
  return usePostAuthApiKeys();
}

export function useUpdateApiKey() {
  return usePatchAuthApiKeysKeyId();
}

export function useRevokeApiKey() {
  return useDeleteAuthApiKeysKeyId();
}

export function useRotateApiKey() {
  return usePostAuthApiKeysKeyIdRotate();
}
