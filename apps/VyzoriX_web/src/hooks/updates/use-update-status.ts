import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getUpdates, type UpdateStatusResponse } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchUpdateStatusViaGraphQL, normalizeWireUpdateStatus } from './_graphql-fallback';

export function useUpdateStatus(
  options?: Omit<UseQueryOptions<UpdateStatusResponse>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updatesStatus(organizationId ?? ''),
    queryFn: async (): Promise<UpdateStatusResponse> => {
      try {
        return normalizeWireUpdateStatus(await getUpdates().getUpdatesStatus());
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return fetchUpdateStatusViaGraphQL(organizationId);
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { UpdateStatusResponse };
