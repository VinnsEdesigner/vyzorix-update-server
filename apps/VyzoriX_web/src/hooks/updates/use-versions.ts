import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  updates,
  type VersionListResult,
  type VersionParams,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchVersionsViaGraphQL } from './_graphql-fallback';

export function useVersions(
  params?: VersionParams,
  options?: Omit<UseQueryOptions<VersionListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateVersions({ ...params, organizationId }),
    queryFn: async () => {
      try {
        return await updates.getVersions(params);
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
