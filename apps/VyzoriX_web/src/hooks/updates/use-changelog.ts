import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  updates,
  type ChangelogEntry,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchChangelogViaGraphQL } from './_graphql-fallback';

export function useChangelog(
  version?: string,
  options?: Omit<UseQueryOptions<ChangelogEntry[]>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.updateChangelog(version),
    queryFn: async () => {
      try {
        return await updates.getChangelog(version, organizationId ?? undefined);
      } catch {
        if (!organizationId) throw new Error('No organization selected');
        return fetchChangelogViaGraphQL(organizationId, version);
      }
    },
    enabled: organizationId !== null,
    ...options,
  });
}
