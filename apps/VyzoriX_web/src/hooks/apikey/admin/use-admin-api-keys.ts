import {
  useInfiniteQuery,
  useQuery,
  useMutation,
  useQueryClient,
  type UseQueryOptions,
} from '@tanstack/react-query';
import {
  getAdmin,
  type AdminAPIKeyListResult,
  type Pagination,
  type GlobalAPIKeyStatsResult,
  type OperatorAPIKeyStatsResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';

const ADMIN_PAGE_SIZE = 20;

interface AdminKeyFilters {
  operatorId?: string;
  search?: string;
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
      getAdmin().getAdminApiKeys({
        page: pageParam,
        limit: ADMIN_PAGE_SIZE,
        operator_id: filters?.operatorId,
        search: filters?.search,
      }),
    initialPageParam: 1,
    getNextPageParam: (lastPage) => {
      const page = lastPage.pagination?.page ?? 1;
      const totalPages = lastPage.pagination?.total_pages ?? 0;
      return page < totalPages ? page + 1 : undefined;
    },
  });

  const pages = query.data?.pages ?? [];
  const firstPage = pages[0];

  return {
    keys: pages.flatMap((page: AdminAPIKeyListResult) => page.keys ?? []),
    pagination: normalizePagination(firstPage?.pagination, ADMIN_PAGE_SIZE),
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
  options?: Omit<UseQueryOptions<AdminAPIKeyListResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.operatorKeys(operatorId ?? '', page, limit),
    queryFn: () => getAdmin().getAdminApiKeysOperatorOperatorId(operatorId!, { page, limit }),
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
  options?: Omit<UseQueryOptions<GlobalAPIKeyStatsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.stats(),
    queryFn: () => getAdmin().getAdminApiKeysStats(),
    staleTime: 1000 * 60 * 5,
    ...options,
  });
}

/**
 * Per-operator API key statistics (super admin).
 */
export function useAdminOperatorKeyStats(
  operatorId: string | undefined,
  options?: Omit<UseQueryOptions<OperatorAPIKeyStatsResult>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: queryKeys.adminApiKeysQueryKeys.operatorStats(operatorId ?? ''),
    queryFn: () => getAdmin().getAdminApiKeysStatsOperatorOperatorId(operatorId!),
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
    mutationFn: async (keyId: string): Promise<void> => {
      await getAdmin().deleteAdminApiKeysKeyId(keyId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.adminApiKeysQueryKeys.all });
    },
  });
}

export type { AdminAPIKeyListResult, GlobalAPIKeyStatsResult, OperatorAPIKeyStatsResult };
