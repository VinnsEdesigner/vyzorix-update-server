import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import type { RegisteredDeviceListResult } from '@vyzorix/api-client';
import { getDevices, useGetDevices } from '@/generated-rq/devices/device-management';
import { fetchRegisteredDevicesViaGraphQL, normalizeRegisteredDevice } from './_graphql-fallback';

export interface UseRegisteredDevicesParams {
  status?: string;
  page?: number;
  limit?: number;
}

export function useRegisteredDevices(params?: UseRegisteredDevicesParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetDevices<RegisteredDeviceListResult>(
    { page: params?.page, limit: params?.limit },
    {
      query: {
        queryKey: ['registration', 'devices', { ...params, organizationId }] as const,
        enabled: organizationId !== null,
        queryFn: async () => {
          try {
            const result = await getDevices({ page: params?.page, limit: params?.limit });
            const limit = params?.limit ?? 20;
            const total = result.total ?? result.devices?.length ?? 0;
            return {
              devices: (result.devices ?? []).map(normalizeRegisteredDevice),
              pagination: {
                page: params?.page ?? 1,
                limit,
                total,
                totalPages: Math.ceil(total / limit),
              },
            } as unknown as Awaited<ReturnType<typeof getDevices>>;
          } catch (restError) {
            if (!organizationId) throw restError;
            return fetchRegisteredDevicesViaGraphQL(organizationId, params)as unknown as Awaited<ReturnType<typeof getDevices>>;
          }
        },
      },
    },
  );
}

export type { RegisteredDeviceListResult } from '@vyzorix/api-client';
