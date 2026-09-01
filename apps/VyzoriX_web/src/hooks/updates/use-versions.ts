import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { VersionListResult } from '@vyzorix/api-client';
import {
  getUpdatesVersions,
  useGetUpdatesVersions,
} from '@/generated-rq/updates/update-management';
import { fetchVersionsViaGraphQL, normalizeWireVersionList } from './_graphql-fallback';

export interface VersionParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useVersions(params?: VersionParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetUpdatesVersions<VersionListResult>(
    { status: params?.status, page: params?.page, limit: params?.limit },
    {
      query: {
        queryKey: ['updates', 'versions', { ...params, organizationId }] as const,
        enabled: organizationId !== null,
        queryFn: async () => {
          try {
            return normalizeWireVersionList(
              await getUpdatesVersions({ status: params?.status, page: params?.page, limit: params?.limit }),
            ) as unknown as Awaited<ReturnType<typeof getUpdatesVersions>>;
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchVersionsViaGraphQL(organizationId, {
              status: params?.status,
              limit: params?.limit,
              offset: params?.page ? (params.page - 1) * (params.limit ?? 20) : undefined,
            }) as unknown as Awaited<ReturnType<typeof getUpdatesVersions>>;
          }
        },
      },
    },
  );
}

export type { VersionListResult } from '@vyzorix/api-client';
