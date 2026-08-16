import {
  useInfiniteQuery,
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  admin,
  type AdminApiKeyListResult,
  type GlobalApiKeyStats,
  type OperatorApiKeyStats,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

const ADMIN_PAGE_SIZE = 20;

interface AdminKeyFilters {
  operatorId?: string;
  search?: string;
}

/**
 * List all API keys across operators (super admin) with infinite pagination
 * and optional filters (spec §12.5).
 *
 * Returns a flattened `keys` array across all loaded pages plus the first
 * page's pagination, so the UI doesn't deal with the infinite-query page
 * structure directly.
 */
export function useAdminApiKeys(filters?: AdminKeyFilters) {
  const query = useInfiniteQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.list(
      filters ? { ...filters } : {},
    ),
    queryFn: ({ pageParam }) =>
      admin.listAllKeys({
        page: pageParam,
        limit: ADMIN_PAGE_SIZE,
        operatorId: filters?.operatorId,
        search: filters?.search,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const { page, totalPages } = lastPage.pagination;
      return page < totalPages ? page + 1 : undefined;
    },
  });

  const pages = query.data?.pages ?? [];
  const firstPage = pages[0];

  return {
    keys: pages.flatMap((page: AdminApiKeyListResult) => page.keys),
    pagination: firstPage?.pagination ?? {
      page: 1,
      limit: ADMIN_PAGE_SIZE,
      total: 0,
      totalPages: 0,
    },
    isLoading: query.isLoading,
    isFetchingNextPage: query.isFetchingNextPage,
    hasNextPage: query.hasNextPage ?? false,
    error: query.error,
    fetchNextPage: query.fetchNextPage,
    refetch: query.refetch,
  };
}

/**
 * List all API keys for a specific operator (super admin).
 */
export function useAdminOperatorKeys(
  operatorId: string | undefined,
  page?: number,
  limit?: number,
  options?: Omit<UseQueryOptions<AdminApiKeyListResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.operatorKeys(operatorId ?? '', page, limit),
    queryFn: () => admin.getOperatorKeys(operatorId!, page, limit),
    enabled: operatorId !== undefined && operatorId !== '',
    ...options,
  });
}

/**
 * Global API key statistics (super admin). Cached for 5 minutes since the
 * aggregate is expensive to compute server-side and changes infrequently
 * (spec §12.5).
 */
export function useGlobalStats(
  options?: Omit<UseQueryOptions<GlobalApiKeyStats>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.stats(),
    queryFn: () => admin.getGlobalKeyStats(),
    staleTime: 1000 * 60 * 5,
    ...options,
  });
}

/**
 * Per-operator API key statistics (super admin).
 */
export function useAdminOperatorKeyStats(
  operatorId: string | undefined,
  options?: Omit<UseQueryOptions<OperatorApiKeyStats>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.operatorStats(operatorId ?? ''),
    queryFn: () => admin.getOperatorKeyStats(operatorId!),
    enabled: operatorId !== undefined && operatorId !== '',
    ...options,
  });
}

/**
 * Force-revoke any operator's API key (super admin). Invalidates both the
 * admin list and stats caches on success (spec §12.5).
 */
export function useForceRevokeKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (keyId: string) => admin.forceRevokeKey(keyId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminApiKeysQueryKeys.all });
    },
  });
}

export type { AdminApiKeyListResult, GlobalApiKeyStats, OperatorApiKeyStats };
