import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { UpdateStatusResponse } from '@vyzorix/api-client';
import {
  getUpdatesStatus,
  useGetUpdatesStatus,
} from '@/generated-rq/updates/update-management';
import { fetchUpdateStatusViaGraphQL, normalizeWireUpdateStatus } from './_graphql-fallback';

export function useUpdateStatus() {
  const organizationId = useCurrentOrganizationId();
  return useGetUpdatesStatus<UpdateStatusResponse>({
    query: {
      queryKey: ['updates', 'status', organizationId ?? ''] as const,
      enabled: organizationId !== null,
      queryFn: async () => {
        try {
          return normalizeWireUpdateStatus(await getUpdatesStatus()) as unknown as Awaited<ReturnType<typeof getUpdatesStatus>>;
        } catch (restError) {
          if (!organizationId) throw restError;
          return fetchUpdateStatusViaGraphQL(organizationId) as unknown as Awaited<ReturnType<typeof getUpdatesStatus>>;
        }
      },
    },
  });
}
