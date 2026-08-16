import { useState, useCallback } from 'react';
import {
  useInfiniteQuery,
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  apiKeys,
  validateCreateApiKeyRequest,
  validateUpdateApiKeyRequest,
  type ApiKey,
  type ApiKeyWithSecret,
  type ApiKeyListResult,
  type ApiKeyScope,
  type CreateApiKeyInput,
  type UpdateApiKeyInput,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

const PAGE_SIZE = 20;

interface ApiKeyListParams {
  page?: number;
  limit?: number;
}

/**
 * List API keys with cursor-style pagination (spec §5.2).
 *
 * Uses useInfiniteQuery so the UI can fetch successive pages via
 * fetchNextPage. The return shape flattens all loaded pages into a single
 * `keys` array and exposes the first page's pagination + monthly stats, so
 * consumers don't need to know about the infinite-query page structure.
 */
export function useApiKeys(params?: ApiKeyListParams) {
  const organizationId = useCurrentOrganizationId();
  const limit = params?.limit ?? PAGE_SIZE;

  const query = useInfiniteQuery({
    queryKey: queryKeys.apiKeysQueryKeys.list({ ...params, organizationId, limit }),
    queryFn: ({ pageParam }) =>
      apiKeys.list({
        page: pageParam,
        limit,
        organizationId: organizationId ?? undefined,
      }),
    initialPageParam: params?.page ?? 1,
    getNextPageParam: (lastPage) => {
      const { page, totalPages } = lastPage.pagination;
      return page < totalPages ? page + 1 : undefined;
    },
    enabled: organizationId !== null,
  });

  const pages = query.data?.pages ?? [];
  const firstPage = pages[0];

  return {
    keys: pages.flatMap((page: ApiKeyListResult) => page.keys),
    pagination: firstPage?.pagination ?? { page: 1, limit, total: 0, totalPages: 0 },
    monthlyLimit: firstPage?.stats.monthlyLimit ?? 0,
    keysCreatedThisMonth: firstPage?.stats.keysCreatedThisMonth ?? 0,
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    hasNextPage: query.hasNextPage ?? false,
    error: query.error,
    fetchNextPage: query.fetchNextPage,
    refetch: query.refetch,
  };
}

export function useApiKey(
  id: string | undefined,
  options?: Omit<UseQueryOptions<ApiKey>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.apiKeysQueryKeys.detail(id ?? ''),
    queryFn: () => apiKeys.get(id!, organizationId ?? undefined),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

/**
 * Create an API key with client-side validation (spec §5.3).
 *
 * Validates the request before sending; on validation failure the mutation is
 * short-circuited and `validationErrors` is populated instead of hitting the
 * server.
 */
export function useCreateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]>>({});

  const mutation = useMutation({
    mutationFn: (input: CreateApiKeyInput): Promise<ApiKeyWithSecret> => {
      const validation = validateCreateApiKeyRequest(
        input.name,
        input.scope,
        input.expiresInDays ?? null,
      );
      if (!validation.isValid) {
        throw { validationErrors: validation.errors };
      }
      return apiKeys.create(input, organizationId ?? undefined);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
  });

  const createKey = useCallback(
    async (input: CreateApiKeyInput): Promise<ApiKeyWithSecret> => {
      setValidationErrors({});
      try {
        return await mutation.mutateAsync(input);
      } catch (error) {
        if (isValidationRejection(error)) {
          setValidationErrors(error.validationErrors);
        }
        throw error;
      }
    },
    [mutation],
  );

  return {
    createKey,
    isCreating: mutation.isPending,
    createdKey: mutation.data ?? null,
    error: mutation.error,
    validationErrors,
    reset: () => {
      setValidationErrors({});
      mutation.reset();
    },
  };
}

/**
 * Update an API key (rename / change scope) with validation (spec §5.6).
 */
export function useUpdateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]>>({});

  const mutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApiKeyInput }): Promise<ApiKey> => {
      const validation = validateUpdateApiKeyRequest(input);
      if (!validation.isValid) {
        throw { validationErrors: validation.errors };
      }
      return apiKeys.update(id, input, organizationId ?? undefined);
    },
    onSuccess: (updated, { id }) => {
      queryClient.setQueryData(queryKeys.apiKeysQueryKeys.detail(id), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
  });

  const updateKey = useCallback(
    async (id: string, input: UpdateApiKeyInput): Promise<ApiKey> => {
      setValidationErrors({});
      try {
        return await mutation.mutateAsync({ id, input });
      } catch (error) {
        if (isValidationRejection(error)) {
          setValidationErrors(error.validationErrors);
        }
        throw error;
      }
    },
    [mutation],
  );

  return {
    updateKey,
    isUpdating: mutation.isPending,
    updatedKey: mutation.data ?? null,
    error: mutation.error,
    validationErrors,
    reset: () => {
      setValidationErrors({});
      mutation.reset();
    },
  };
}

/**
 * Revoke an API key with an optimistic update (spec §5.4).
 *
 * Removes the key from cached list pages immediately; rolls back on error.
 */
export function useRevokeApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  const [pendingRevoke, setPendingRevoke] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: (id: string) => apiKeys.revoke(id, organizationId ?? undefined),
    onMutate: async (keyId: string) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
      const previousData = queryClient.getQueriesData({
        queryKey: queryKeys.apiKeysQueryKeys.lists(),
      });

      // The list query is an useInfiniteQuery, so cached data has the shape
      // { pages: ApiKeyListResult[], pageParams: number[] }. Optimistically
      // remove the revoked key from every loaded page.
      queryClient.setQueriesData(
        { queryKey: queryKeys.apiKeysQueryKeys.lists() },
        (old: { pages: ApiKeyListResult[]; pageParams: number[] } | undefined) => {
          if (!old) return old;
          return {
            ...old,
            pages: old.pages.map((page) => ({
              ...page,
              keys: page.keys.filter((k) => k.id !== keyId),
            })),
          };
        },
      );

      setPendingRevoke(keyId);
      return { previousData };
    },
    onError: (_err, _keyId, context) => {
      if (context?.previousData) {
        for (const [queryKey, data] of context.previousData) {
          queryClient.setQueryData(queryKey, data);
        }
      }
    },
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: queryKeys.apiKeysQueryKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
    onSettled: () => {
      setPendingRevoke(null);
    },
  });

  const revokeKey = useCallback(
    async (id: string): Promise<void> => {
      return mutation.mutateAsync(id);
    },
    [mutation],
  );

  return {
    revokeKey,
    isRevoking: mutation.isPending,
    pendingRevoke,
    error: mutation.error,
  };
}

/**
 * Rotate an API key; the new full secret is returned once and surfaced via
 * `rotatedKey` so the UI can show it in a dialog (spec §5.5).
 */
export function useRotateApiKey() {
  const queryClient = useQueryClient();
  const organizationId = useCurrentOrganizationId();
  const [pendingRotate, setPendingRotate] = useState<string | null>(null);
  const [rotatedKey, setRotatedKey] = useState<ApiKeyWithSecret | null>(null);

  const mutation = useMutation({
    mutationFn: (id: string) => apiKeys.rotate(id, organizationId ?? undefined),
    onSuccess: (data) => {
      setRotatedKey(data);
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
    onSettled: () => {
      setPendingRotate(null);
    },
  });

  const rotateKey = useCallback(
    async (id: string): Promise<ApiKeyWithSecret> => {
      setPendingRotate(id);
      try {
        return await mutation.mutateAsync(id);
      } finally {
        setPendingRotate(null);
      }
    },
    [mutation],
  );

  return {
    rotateKey,
    isRotating: mutation.isPending,
    pendingRotate,
    rotatedKey,
    error: mutation.error,
    clearRotatedKey: () => setRotatedKey(null),
  };
}

/** Type guard for the synthetic validation-rejection error thrown by the create/update mutations. */
function isValidationRejection(
  error: unknown,
): error is { validationErrors: Record<string, string[]> } {
  return (
    typeof error === 'object' &&
    error !== null &&
    'validationErrors' in error &&
    typeof (error as { validationErrors: unknown }).validationErrors === 'object'
  );
}

export type { ApiKey, ApiKeyWithSecret, ApiKeyListResult, CreateApiKeyInput, UpdateApiKeyInput, ApiKeyScope };
