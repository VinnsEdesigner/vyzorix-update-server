import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchInboxEntryViaGraphQL } from './_graphql-fallback';
import { useGetDeviceInboxImei, getDeviceInboxImei } from '@/generated-rq/inbox/device-inbox';

export function useInboxEntry(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceInboxImei(
    imei ?? '',
    {
      query: {
        queryKey: ['inbox', imei, organizationId] as const,
        enabled: imei !== undefined && organizationId !== null,
        // REST primary; fall back to GraphQL on REST failure.
        queryFn: async () => {
          try {
            return await getDeviceInboxImei(imei!);
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchInboxEntryViaGraphQL(organizationId, imei!) as unknown as Awaited<ReturnType<typeof getDeviceInboxImei>>;
          }
        },
      },
    },
  );
}
