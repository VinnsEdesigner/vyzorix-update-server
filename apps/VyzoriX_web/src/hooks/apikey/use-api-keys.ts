import { useQueryClient } from '@tanstack/react-query';
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
  return useGetAuthApiKeys(
    { page: params?.page, limit: params?.limit },
    { query: { queryKey: ['api-keys', params] as const } },
  );
}

export function useApiKey(id: string | undefined) {
  return useGetAuthApiKeysKeyId(
    id ?? '',
    { query: { queryKey: ['api-keys', id] as const, enabled: id !== undefined } },
  );
}

export function useCreateApiKey() {
  const queryClient = useQueryClient();
  return usePostAuthApiKeys({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }) },
  });
}

export function useUpdateApiKey() {
  const queryClient = useQueryClient();
  return usePatchAuthApiKeysKeyId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }) },
  });
}

export function useRevokeApiKey() {
  const queryClient = useQueryClient();
  return useDeleteAuthApiKeysKeyId({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }) },
  });
}

export function useRotateApiKey() {
  const queryClient = useQueryClient();
  return usePostAuthApiKeysKeyIdRotate({
    mutation: { onSuccess: () => queryClient.invalidateQueries({ queryKey: ['api-keys'] }) },
  });
}
