import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { registration, type RegisteredDevice } from '@vyzorix/api-client';
import { fetchRegisteredDeviceViaGraphQL } from './_graphql-fallback';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useRegisteredDevice(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<RegisteredDevice | null>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.registrationDevice(imei ?? ''),
    queryFn: async () => {
      if (!imei) return null;
      const org = organizationId ?? undefined;
      try {
        return await registration.getDevice(imei, org);
      } catch (restError) {
        if (!organizationId) throw restError;
        return fetchRegisteredDeviceViaGraphQL(organizationId, imei);
      }
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export type { RegisteredDevice };
