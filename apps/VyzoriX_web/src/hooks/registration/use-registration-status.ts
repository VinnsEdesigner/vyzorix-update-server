import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getInbox, type InboxEntry, type InboxStatus } from '@vyzorix/api-client';
import { fetchInboxEntryViaGraphQL, normalizeInboxEntry } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface RegistrationStatus {
  imei: string;
  status: InboxStatus;
  entry: InboxEntry | null;
}

export function useRegistrationStatus(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<RegistrationStatus | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationStatus(imei ?? ''),
    queryFn: async (): Promise<RegistrationStatus | null> => {
      if (!imei) return null;
      let entry: InboxEntry | null;
      try {
        const raw = await getInbox().getDeviceInboxImei(imei);
        entry = raw ? normalizeInboxEntry(raw) : null;
      } catch (restError) {
        if (!organizationId) throw restError;
        entry = await fetchInboxEntryViaGraphQL(organizationId, imei);
      }
      if (!entry) return null;
      return { imei, status: entry.status, entry };
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export type { InboxStatus, InboxEntry };
