import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { ChangelogEntry } from '@vyzorix/api-client';
import {
  getUpdatesChangelog,
  useGetUpdatesChangelog,
  getGetUpdatesChangelogQueryKey,
} from '@/generated-rq/updates/update-management';
import { fetchChangelogViaGraphQL, normalizeWireChangelog } from './_graphql-fallback';

export function useChangelog(version?: string) {
  const organizationId = useCurrentOrganizationId();
  const params = version ? { version } : undefined;
  return useGetUpdatesChangelog<ChangelogEntry[]>(
    params,
    {
      query: {
        queryKey: getGetUpdatesChangelogQueryKey(params),
        enabled: organizationId !== null,
        queryFn: async () => {
          try {
            return normalizeWireChangelog(
              await getUpdatesChangelog(version ? { version } : undefined),
            ) as unknown as Awaited<ReturnType<typeof getUpdatesChangelog>>;
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchChangelogViaGraphQL(organizationId, version) as unknown as Awaited<ReturnType<typeof getUpdatesChangelog>>;
          }
        },
      },
    },
  );
}

export type { ChangelogEntry } from '@vyzorix/api-client';
