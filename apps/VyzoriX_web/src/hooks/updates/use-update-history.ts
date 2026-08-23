import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getUpdates, type UpdateHistoryResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchUpdateHistoryViaGraphQL, normalizeWireHistoryList } from './_graphql-fallback';

export interface HistoryParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useUpdateHistory(
  params?: HistoryParams,
  options?: Omit<UseQueryOptions<UpdateHistoryResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateHistory({ ...params, organizationId }),
    queryFn: async (): Promise<UpdateHistoryResult> => {
      try {
        return normalizeWireHistoryList(
          await getUpdates().getUpdatesHistory({ page: params?.page, limit: params?.limit }),
        );
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return fetchUpdateHistoryViaGraphQL(organizationId, {
          status: params?.status,
          page: params?.page,
          limit: params?.limit,
        });
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { UpdateHistoryResult };
