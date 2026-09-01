import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { getDeviceInboxImei, useGetDeviceInboxImei } from '@/generated-rq/inbox/device-inbox';
import { fetchInboxEntryViaGraphQL, normalizeInboxEntry } from './_graphql-fallback';
import type { InboxStatus, InboxEntry } from '@vyzorix/api-client';

export interface RegistrationStatus {
  imei: string;
  status: InboxStatus;
  entry: InboxEntry | null;
}

export function useRegistrationStatus(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceInboxImei<RegistrationStatus | null>(
    imei ?? '',
    {
      query: {
        queryKey: ['registration', 'status', imei] as const,
        enabled: imei !== undefined && imei !== '' && organizationId !== null,
        queryFn: async () => {
          if (!imei) return null as unknown as Awaited<ReturnType<typeof getDeviceInboxImei>>;
          let entry: InboxEntry | null;
          try {
            const raw = await getDeviceInboxImei(imei);
            entry = raw ? normalizeInboxEntry(raw) : null;
          } catch (restError) {
            if (!organizationId) throw restError;
            entry = await fetchInboxEntryViaGraphQL(organizationId, imei!);
          }
          if (!entry) return null as unknown as Awaited<ReturnType<typeof getDeviceInboxImei>>;
          return { imei, status: entry.status, entry } as unknown as Awaited<ReturnType<typeof getDeviceInboxImei>>;
        },
      },
    },
  );
}

export type { InboxStatus, InboxEntry } from '@vyzorix/api-client';
