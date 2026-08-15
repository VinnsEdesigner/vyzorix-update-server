import {
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  apiKeys,
  type ApiKey,
  type ApiKeyWithSecret,
  type ApiKeyListResult,
  type CreateApiKeyInput,
  type UpdateApiKeyInput,
  type ApiKeyScope,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

interface ApiKeyListParams {
  page?: number;
  limit?: number;
}

export function useApiKeys(
  params?: ApiKeyListParams,
  options?: Omit<UseQueryOptions<ApiKeyListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.apiKeys({ ...params, organizationId }),
    queryFn: () => apiKeys.list({ ...params, organizationId: organizationId ?? undefined }),
    enabled: organizationId !== null,
    ...options,
  });
}

export function useApiKey(
  id: string | undefined,
  options?: Omit<UseQueryOptions<ApiKey>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.apiKey(id ?? ''),
    queryFn: () => apiKeys.get(id!, organizationId ?? undefined),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

export function useCreateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (input: CreateApiKeyInput) =>
      apiKeys.create(input, organizationId ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export function useUpdateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApiKeyInput }) =>
      apiKeys.update(id, input, organizationId ?? undefined),
    onSuccess: (updated, { id }) => {
      queryClient.setQueryData(queryKeys.apiKey(id), updated);
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export function useRevokeApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => apiKeys.revoke(id, organizationId ?? undefined),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      queryClient.removeQueries({ queryKey: queryKeys.apiKey(id) });
    },
  });
}

export function useRotateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  return useMutation({
    mutationFn: (id: string) => apiKeys.rotate(id, organizationId ?? undefined),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
}

export type { ApiKey, ApiKeyWithSecret, ApiKeyListResult, CreateApiKeyInput, UpdateApiKeyInput, ApiKeyScope };
