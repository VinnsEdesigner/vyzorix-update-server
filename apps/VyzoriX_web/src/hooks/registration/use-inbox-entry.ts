import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getInbox, type InboxEntry } from '@vyzorix/api-client';
import { fetchInboxEntryViaGraphQL, normalizeInboxEntry } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useInboxEntry(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<InboxEntry | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationInboxEntry(imei ?? ''),
    queryFn: async (): Promise<InboxEntry | null> => {
      if (!imei) return null;
      try {
        const result = await getInbox().getDeviceInboxImei(imei);
        return result ? normalizeInboxEntry(result) : null;
      } catch (restError) {
        if (!organizationId) throw restError;
        return fetchInboxEntryViaGraphQL(organizationId, imei);
      }
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export type { InboxEntry };
