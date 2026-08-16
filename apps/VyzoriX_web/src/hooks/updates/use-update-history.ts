import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  updates,
  type UpdateHistoryResult,
  type HistoryParams,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchUpdateHistoryViaGraphQL } from './_graphql-fallback';

export function useUpdateHistory(
  params?: HistoryParams,
  options?: Omit<UseQueryOptions<UpdateHistoryResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateHistory({ ...params, organizationId }),
    queryFn: async () => {
      try {
        return await updates.getHistory(params);
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
