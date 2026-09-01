import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { UpdateHistoryResult } from '@vyzorix/api-client';
import {
  getUpdatesHistory,
  useGetUpdatesHistory,
} from '@/generated-rq/updates/update-management';
import { fetchUpdateHistoryViaGraphQL, normalizeWireHistoryList } from './_graphql-fallback';

export interface HistoryParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useUpdateHistory(params?: HistoryParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetUpdatesHistory<UpdateHistoryResult>(
    { page: params?.page, limit: params?.limit },
    {
      query: {
        queryKey: ['updates', 'history', { ...params, organizationId }] as const,
        enabled: organizationId !== null,
        queryFn: async () => {
          try {
            return normalizeWireHistoryList(
              await getUpdatesHistory({ page: params?.page, limit: params?.limit }),
            ) as unknown as Awaited<ReturnType<typeof getUpdatesHistory>>;
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchUpdateHistoryViaGraphQL(organizationId, {
              status: params?.status,
              page: params?.page,
              limit: params?.limit,
            }) as unknown as Awaited<ReturnType<typeof getUpdatesHistory>>;
          }
        },
      },
    },
  );
}

export type { UpdateHistoryResult } from '@vyzorix/api-client';
