import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getUpdates, type VersionListResult } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchVersionsViaGraphQL, normalizeWireVersionList } from './_graphql-fallback';

export interface VersionParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useVersions(
  params?: VersionParams,
  options?: Omit<UseQueryOptions<VersionListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateVersions({ ...params, organizationId }),
    queryFn: async (): Promise<VersionListResult> => {
      try {
        return normalizeWireVersionList(
          await getUpdates().getUpdatesVersions({
            status: params?.status,
            page: params?.page,
            limit: params?.limit,
          }),
        );
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return fetchVersionsViaGraphQL(organizationId, {
          status: params?.status,
          limit: params?.limit,
          offset: params?.page ? (params.page - 1) * (params.limit ?? 20) : undefined,
        });
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}

export type { VersionListResult };
