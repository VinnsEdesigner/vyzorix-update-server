import { useState, useCallback } from 'react';
import {
  useInfiniteQuery,
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  getApiKeys,
  validateCreateApiKeyRequest,
  validateUpdateApiKeyRequest,
  type APIKey,
  type APIKeyWithSecret,
  type APIKeyListResult,
  type Pagination,
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

function normalizePagination(raw: { page?: number; limit?: number; total?: number; total_pages?: number } | undefined, fallbackLimit: number): Pagination {
  return {
    page: raw?.page ?? 1,
    limit: raw?.limit ?? fallbackLimit,
    total: raw?.total ?? 0,
    totalPages: raw?.total_pages ?? 0,
  };
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
  const limit = params?.limit ?? PAGE_SIZE;
  const organizationId = useCurrentOrganizationId();

  const query = useInfiniteQuery({
    queryKey: queryKeys.apiKeysQueryKeys.list({ ...params, organizationId, limit }),
    queryFn: ({ pageParam }) =>
      getApiKeys().getAuthApiKeys({ page: pageParam, limit }),
    initialPageParam: params?.page ?? 1,
    getNextPageParam: (lastPage) => {
      const page = lastPage.pagination?.page ?? 1;
      const totalPages = lastPage.pagination?.total_pages ?? 0;
      return page < totalPages ? page + 1 : undefined;
    },
    enabled: organizationId !== null,
  });

  const pages = query.data?.pages ?? [];
  const firstPage = pages[0];

  return {
    keys: pages.flatMap((page: APIKeyListResult) => page.keys ?? []),
    pagination: normalizePagination(firstPage?.pagination, limit),
    monthlyLimit: firstPage?.monthly_limit ?? 0,
    keysCreatedThisMonth: firstPage?.keys_created_this_month ?? 0,
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
  options?: Omit<UseQueryOptions<APIKey>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.apiKeysQueryKeys.detail(id ?? ''),
    queryFn: () => getApiKeys().getAuthApiKeysKeyId(id!),
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
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]>>({});

  const mutation = useMutation({
    mutationFn: (input: CreateApiKeyInput): Promise<APIKeyWithSecret> => {
      const validation = validateCreateApiKeyRequest({
        name: input.name,
        scope: input.scope,
        expires_in_days: input.expiresInDays ?? undefined,
      });
      if (!validation.success) {
        throw { validationErrors: zodToValidationErrors(validation.error) };
      }
      return getApiKeys().postAuthApiKeys({
        name: input.name,
        scope: input.scope,
        expires_in_days: input.expiresInDays ?? undefined,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
  });

  const createKey = useCallback(
    async (input: CreateApiKeyInput): Promise<APIKeyWithSecret> => {
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
  const [validationErrors, setValidationErrors] = useState<Record<string, string[]>>({});

  const mutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateApiKeyInput }): Promise<APIKey> => {
      const validation = validateUpdateApiKeyRequest(input);
      if (!validation.success) {
        throw { validationErrors: zodToValidationErrors(validation.error) };
      }
      return getApiKeys().patchAuthApiKeysKeyId(id, {
        name: input.name,
        scope: input.scope,
      });
    },
    onSuccess: (updated, { id }) => {
      queryClient.setQueryData(queryKeys.apiKeysQueryKeys.detail(id), updated);
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
  });

  const updateKey = useCallback(
    async (id: string, input: UpdateApiKeyInput): Promise<APIKey> => {
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
  const [pendingRevoke, setPendingRevoke] = useState<string | null>(null);

  const mutation = useMutation({
    mutationFn: async (id: string): Promise<void> => {
      await getApiKeys().deleteAuthApiKeysKeyId(id);
    },
    onMutate: async (keyId: string) => {
      await queryClient.cancelQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
      const previousData = queryClient.getQueriesData({
        queryKey: queryKeys.apiKeysQueryKeys.lists(),
      });

      // The list query is an useInfiniteQuery, so cached data has the shape
      // { pages: APIKeyListResult[], pageParams: number[] }. Optimistically
      // remove the revoked key from every loaded page.
      queryClient.setQueriesData(
        { queryKey: queryKeys.apiKeysQueryKeys.lists() },
        (old: { pages: APIKeyListResult[]; pageParams: number[] } | undefined) => {
          if (!old) return old;
          return {
            ...old,
            pages: old.pages.map((page) => ({
              ...page,
              keys: (page.keys ?? []).filter((k) => k.id !== keyId),
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
      await mutation.mutateAsync(id);
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
  const [pendingRotate, setPendingRotate] = useState<string | null>(null);
  const [rotatedKey, setRotatedKey] = useState<APIKeyWithSecret | null>(null);

  const mutation = useMutation({
    mutationFn: (id: string) => getApiKeys().postAuthApiKeysKeyIdRotate(id),
    onSuccess: (data) => {
      setRotatedKey(data);
      queryClient.invalidateQueries({ queryKey: queryKeys.apiKeysQueryKeys.lists() });
    },
    onSettled: () => {
      setPendingRotate(null);
    },
  });

  const rotateKey = useCallback(
    async (id: string): Promise<APIKeyWithSecret> => {
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

/** Converts a zod safeParse failure into the field-error map the hooks expose. */
function zodToValidationErrors(error: { issues: { path: (string | number)[]; message: string }[] }): Record<string, string[]> {
  const out: Record<string, string[]> = {};
  for (const issue of error.issues) {
    const field = String(issue.path[0] ?? '_');
    (out[field] ??= []).push(issue.message);
  }
  return out;
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

export type { APIKey, APIKeyWithSecret, APIKeyListResult, CreateApiKeyInput, UpdateApiKeyInput, ApiKeyScope };
