import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { RegisteredDevice } from '@vyzorix/api-client';
import { getDevicesImei, useGetDevicesImei } from '@/generated-rq/devices/device-management';
import { fetchRegisteredDeviceViaGraphQL, normalizeRegisteredDevice } from './_graphql-fallback';

export function useRegisteredDevice(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDevicesImei<RegisteredDevice>(
    imei ?? '',
    {
      query: {
        queryKey: ['registration', 'device', imei] as const,
        enabled: imei !== undefined && imei !== '' && organizationId !== null,
        queryFn: async () => {
          try {
            return normalizeRegisteredDevice(await getDevicesImei(imei!)) as unknown as Awaited<ReturnType<typeof getDevicesImei>>;
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchRegisteredDeviceViaGraphQL(organizationId, imei!) as unknown as Awaited<ReturnType<typeof getDevicesImei>>;
          }
        },
      },
    },
  );
}

export type { RegisteredDevice } from '@vyzorix/api-client';
